package configurablecommand

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"os/user"
	"strings"
	"time"
)

type Executor struct {
	action Command
	params []param

	cmd *exec.Cmd

	file *os.File

	slackMsgCh <-chan string
	errMsgCh   <-chan string

	stopped bool
}

func NewExecutor(action Command, params []param) (*Executor, error) {
	args := strings.Split(action.Command, " ")
	for _, p := range params {
		args = append(args, "--"+p.Name, p.Value)
	}
	cmd := exec.Command(args[0], args[1:]...)

	logFilename := action.LogFilename
	if len(logFilename) == 0 {
		logFilename = "/dev/null"
	}
	logFilename, err := fullpath(logFilename)
	if err != nil {
		return nil, err
	}

	executor := &Executor{
		action: action,
		params: params,
		cmd:    cmd,
	}

	file, err := os.OpenFile(logFilename, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, fmt.Errorf("fail to open log file: %v", err)
	}
	logger := log.New(file, "", log.LstdFlags)
	slackMsgCh, err := executor.initStdoutPipe(cmd, logger)
	if err != nil {
		file.Close()
		return nil, fmt.Errorf("fail to open stdout pipe: %v", err)
	}
	errMsgCh, err := executor.initStderrPipe(cmd, logger)
	if err != nil {
		file.Close()
		return nil, fmt.Errorf("fail to open stderr pipe: %v", err)
	}

	executor.file = file
	executor.slackMsgCh = slackMsgCh
	executor.errMsgCh = errMsgCh
	return executor, nil
}

func (e *Executor) Close() error {
	e.stopped = true
	return e.file.Close()
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

			if text == "post_slack_begin" {
				texts = []string{}
				continue
			}
			if text == "post_slack_end" {
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
	return strings.Join(e.cmd.Args, " ")
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
