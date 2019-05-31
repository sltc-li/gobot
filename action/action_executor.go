package action

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"
)

type Executor struct {
	action Action
	params []Param

	cmd *exec.Cmd

	file *os.File

	slackMsgCh <-chan string
	errMsgCh   <-chan string
}

func NewExecutor(action Action, params []Param) (*Executor, error) {
	args := strings.Split(action.Command, " ")
	for _, p := range params {
		args = append(args, "--"+p.Name, p.Value)
	}
	cmd := exec.Command(args[0], args[1:]...)
	file, err := os.OpenFile(action.LogFilename, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, fmt.Errorf("fail to open log file: %v", err)
	}
	logger := log.New(file, "", log.LstdFlags)
	slackMsgCh, err := initStdoutPipe(cmd, logger)
	if err != nil {
		file.Close()
		return nil, fmt.Errorf("fail to open stdout pipe: %v", err)
	}
	errMsgCh, err := initStderrPipe(cmd, logger)
	if err != nil {
		file.Close()
		return nil, fmt.Errorf("fail to open stderr pipe: %v", err)
	}
	return &Executor{
		action:     action,
		params:     params,
		cmd:        cmd,
		file:       file,
		slackMsgCh: slackMsgCh,
		errMsgCh:   errMsgCh,
	}, nil
}

func (e *Executor) Close() error {
	return e.file.Close()
}

func initStdoutPipe(cmd *exec.Cmd, logger *log.Logger) (<-chan string, error) {
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
		for scanner.Scan() {
			text := scanner.Text()
			logger.Print(text)

			if text == "post_slack_begin" {
				texts = []string{}
				continue
			}
			if text == "post_slack_end" {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				send(ctx, ch, strings.Join(texts, "\n"))
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

func initStderrPipe(cmd *exec.Cmd, logger *log.Logger) (<-chan string, error) {
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
		for {
			n, err := r.Read(p)
			if err == io.EOF {
				break
			}
			text := string(p[:n])
			logger.Print(text)

			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			send(ctx, ch, text)
			cancel()
		}
	}(stderr, ch)

	return ch, nil
}

func send(ctx context.Context, ch chan<- string, msg string) {
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

func (e *Executor) Exec() error {
	return e.cmd.Run()
}
