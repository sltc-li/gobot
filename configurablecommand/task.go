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
	Killed
	Running
	Succeeded
	Failed
)

var (
	ErrNoKillPermission = errors.New("no kill permission")
)

type Task struct {
	ID  int
	Msg gobot.Message

	bot gobot.Bot
	cmd Command

	runAt    time.Time
	killAt   *time.Time
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
	if t.killAt != nil {
		return Killed
	}
	if t.startAt != nil {
		return Running
	}
	return Pending
}

func (t *Task) Active() bool {
	return t.Status() == Pending || t.Status() == Running
}

func (t *Task) Start() {
	now1 := time.Now()
	t.startAt = &now1
	saveTask(t)

	err := t.execute()

	if t.executor.IsStopped() {
		return
	}

	now2 := time.Now()
	t.finishAt = &now2
	t.err = err
	saveTask(t)
}

func (t *Task) Kill(userID string) error {
	if userID != t.Msg.UserID {
		return ErrNoKillPermission
	}

	if t.Status() == Running && t.executor != nil {
		t.err = t.executor.Stop()
	}
	now := time.Now()
	t.killAt = &now
	saveTask(t)
	return nil
}

func (t *Task) Duration() time.Duration {
	switch t.Status() {
	case Pending:
		return 0
	case Killed:
		if t.startAt != nil {
			return t.killAt.Sub(*t.startAt)
		}
		return 0
	case Running:
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
	bot.GetLogger().Printf("%s is executing `%s` in %s - #%d", user, executor.Command(), channel, t.ID)
	if err := executor.Start(); err != nil {
		bot.SendMessage(fmt.Sprintf(errMsgFmt, err.Error()), msg.ChannelID)
		return err
	}
	if err := executor.Wait(); err != nil {
		return err
	}
	if !executor.IsStopped() {
		bot.SendMessage(fmt.Sprintf("<@%s> *succeeded* - `%s` :open_mouth:", msg.UserID, msg.Text), msg.ChannelID)
	}
	return nil
}
