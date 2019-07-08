package configurablecommand

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"os/user"
	"strings"
	"time"
)

const (
	postSlackBegin = "post_slack_begin"
	postSlackEnd   = "post_slack_end"
	postSlack      = "#!/bin/sh\necho " + postSlackBegin + "\necho \"$@\"\necho " + postSlackEnd + "\n"
	postSlackPath  = "/usr/local/bin/post_slack"
)

type Executor struct {
	command Command
	params  []param

	cmd *exec.Cmd

	logFile *os.File

	slackMsgCh <-chan string
	errMsgCh   <-chan string

	stopped bool
}

func NewExecutor(c Command, params []param) (*Executor, error) {
	executor := &Executor{command: c, params: params}

	// create command
	if err := ioutil.WriteFile(postSlackPath, []byte(postSlack), 0777); err != nil {
		// ignore error
		log.Printf("warning: unable to create " + postSlackPath)
	}
	command := c.Command
	for _, p := range params {
		command += " --" + p.Name + " " + p.Value
	}
	cmd := exec.Command("bash", "-c", command)
	executor.cmd = cmd

	// create log file
	logFilename := c.LogFilename
	if len(logFilename) == 0 {
		logFilename = "/dev/null"
	}
	logFilename, err := fullpath(logFilename)
	if err != nil {
		executor.clean()
		return nil, err
	}
	logFile, err := os.OpenFile(logFilename, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		executor.clean()
		return nil, fmt.Errorf("fail to open log logFile: %v", err)
	}
	executor.logFile = logFile

	// create logger
	logger := log.New(logFile, "", log.LstdFlags)
	slackMsgCh, err := executor.initStdoutPipe(cmd, logger)
	if err != nil {
		executor.clean()
		return nil, fmt.Errorf("fail to open stdout pipe: %v", err)
	}
	errMsgCh, err := executor.initStderrPipe(cmd, logger)
	if err != nil {
		executor.clean()
		return nil, fmt.Errorf("fail to open stderr pipe: %v", err)
	}
	executor.slackMsgCh = slackMsgCh
	executor.errMsgCh = errMsgCh

	return executor, nil
}

func (e *Executor) clean() {
	if e.logFile != nil {
		e.logFile.Close()
	}
	os.Remove(postSlackPath)
}

func (e *Executor) Close() {
	e.stopped = true
	e.clean()
}

func fullpath(path string) (string, error) {
	if !strings.HasPrefix(path, "~") {
		return path, nil
	}
	u, err := user.Current()
	if err != nil {
		return "", err
	}
	return strings.Replace(path, "~", u.HomeDir, 1), nil
}

func (e *Executor) initStdoutPipe(cmd *exec.Cmd, logger *log.Logger) (<-chan string, error) {
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	ch := make(chan string)
	go func(r io.ReadCloser, ch chan<- string) {
		defer func() {
			r.Close()
			close(ch)
		}()

		scanner := bufio.NewScanner(r)
		var texts []string
		for !e.stopped && scanner.Scan() {
			text := scanner.Text()
			logger.Print(text)

			if text == postSlackBegin {
				texts = []string{}
				continue
			}
			if text == postSlackEnd {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				e.send(ctx, ch, strings.Join(texts, "\n"))
				cancel()
				texts = nil
				continue
			}
			if texts != nil {
				texts = append(texts, text)
			}
		}
	}(stdout, ch)

	return ch, nil
}

func (e *Executor) initStderrPipe(cmd *exec.Cmd, logger *log.Logger) (<-chan string, error) {
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, err
	}

	ch := make(chan string)
	go func(r io.ReadCloser, ch chan<- string) {
		defer func() {
			r.Close()
			close(ch)
		}()

		p := make([]byte, 10240)
		for !e.stopped {
			n, err := r.Read(p)
			if err == io.EOF {
				break
			}
			text := string(p[:n])
			logger.Print(text)

			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			e.send(ctx, ch, text)
			cancel()
		}
	}(stderr, ch)

	return ch, nil
}

func (e *Executor) send(ctx context.Context, ch chan<- string, msg string) {
	if e.stopped {
		return
	}

	select {
	case <-ctx.Done():
		return
	case ch <- msg:
		return
	}
}

func (e *Executor) Command() string {
	return e.cmd.Args[2]
}

func (e *Executor) NextSlackMessage() (string, bool) {
	msg, ok := <-e.slackMsgCh
	return msg, ok
}

func (e *Executor) NextErrorMessage() (string, bool) {
	msg, ok := <-e.errMsgCh
	return msg, ok
}

func (e *Executor) Start() error {
	return e.cmd.Start()
}

func (e *Executor) Wait() error {
	err := e.cmd.Wait()
	if e.stopped {
		return nil
	}
	return err
}

func (e *Executor) Stop() error {
	e.stopped = true
	return e.cmd.Process.Kill()
}

func (e *Executor) IsStopped() bool {
	return e.stopped
}
