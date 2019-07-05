package configurablecommand

import (
	"errors"
	"fmt"
	"time"

	"github.com/li-go/gobot/gobot"
)

//go:generate stringer -type=TaskStatus
type TaskStatus int

const (
	Pending TaskStatus = iota
	Canceled
	Started
	Succeeded
	Failed
)

var (
	ErrNoCancelPermission = errors.New("no cancel permission")
)

type Task struct {
	ID  int
	Msg gobot.Message

	bot gobot.Bot
	cmd Command

	runAt    time.Time
	cancelAt *time.Time
	startAt  *time.Time
	finishAt *time.Time

	executor *Executor

	err error
}

func (t *Task) Status() TaskStatus {
	if t.finishAt != nil {
		if t.err == nil {
			return Succeeded
		}
		return Failed
	}
	if t.startAt != nil {
		return Started
	}
	return Pending
}

func (t *Task) Start() {
	now1 := time.Now()
	t.startAt = &now1

	err := t.execute()

	now2 := time.Now()
	t.finishAt = &now2
	t.err = err
}

func (t *Task) Cancel(userID string) error {
	if userID != t.Msg.UserID {
		return ErrNoCancelPermission
	}

	if t.Status() == Started && t.executor != nil {
		t.err = t.executor.Stop()
	}
	now := time.Now()
	t.cancelAt = &now
	return nil
}

func (t *Task) Duration() time.Duration {
	switch t.Status() {
	case Pending:
		return 0
	case Canceled:
		if t.startAt != nil {
			return time.Since(*t.startAt)
		}
		return 0
	case Started:
		return time.Since(*t.startAt)
	case Succeeded, Failed:
		return t.finishAt.Sub(*t.startAt)
	default:
		return 0
	}
}

func (t *Task) execute() error {
	bot := t.bot
	msg := t.Msg

	executor, err := t.cmd.newExecutor(bot, msg)
	if err != nil {
		return err
	}
	defer executor.Close()

	t.executor = executor

	// execute
	channel, err := bot.LoadChannel(msg.ChannelID)
	if err != nil {
		return err
	}
	user, err := bot.LoadUser(msg.UserID)
	if err != nil {
		return err
	}
	bot.GetLogger().Printf("%s is executing `%s` in %s", user, executor.Command(), channel)
	if err := executor.Start(); err != nil {
		return err
	}
	if err := executor.Wait(); err != nil {
		return err
	}
	bot.SendMessage(fmt.Sprintf("<@%s> *succeeded* - `%s` :open_mouth:", msg.UserID, msg.Text), msg.ChannelID)
	return nil
}
