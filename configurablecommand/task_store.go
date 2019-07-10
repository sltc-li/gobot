package configurablecommand

import (
	"encoding/json"
	"errors"
	"github.com/li-go/gobot/gobot"
	"github.com/li-go/gobot/localrepo"
	"time"
)

type TaskEntity struct {
	ID int `db:"id" gorm:"primary_key"`

	MsgType      gobot.MsgType `db:"msg_type"`
	MsgText      string        `db:"msg_text"`
	MsgChannelID string        `db:"msg_channel_id"`
	MsgUserID    string        `db:"msg_user_id"`

	CmdJson string `db:"cmd_json" gorm:"type:text"`

	RunAt    time.Time  `db:"run_at" gorm:"not null"`
	KillAt   *time.Time `db:"kill_at"`
	StartAt  *time.Time `db:"start_at"`
	FinishAt *time.Time `db:"finish_at"`

	ErrMsg *string `db:"err_msg"`
}

func NewTaskEntity(task *Task) (*TaskEntity, error) {
	buf, err := json.Marshal(task.cmd)
	if err != nil {
		return nil, err
	}
	var errMsg *string
	if task.err != nil {
		s := task.err.Error()
		errMsg = &s
	}
	return &TaskEntity{
		ID:           task.ID,
		MsgType:      task.Msg.Type,
		MsgText:      task.Msg.Text,
		MsgChannelID: task.Msg.ChannelID,
		MsgUserID:    task.Msg.UserID,
		CmdJson:      string(buf),
		RunAt:        task.runAt,
		KillAt:       task.killAt,
		StartAt:      task.startAt,
		FinishAt:     task.finishAt,
		ErrMsg:       errMsg,
	}, nil
}

func (entity *TaskEntity) Task() (*Task, error) {
	var cmd Command
	if err := json.Unmarshal([]byte(entity.CmdJson), &cmd); err != nil {
		return nil, err
	}
	var err error
	if entity.ErrMsg != nil {
		err = errors.New(*entity.ErrMsg)
	}
	return &Task{
		ID: entity.ID,
		Msg: gobot.Message{
			Type:      entity.MsgType,
			Text:      entity.MsgText,
			ChannelID: entity.MsgChannelID,
			UserID:    entity.MsgUserID,
		},
		cmd:      cmd,
		runAt:    entity.RunAt,
		killAt:   entity.KillAt,
		startAt:  entity.StartAt,
		finishAt: entity.FinishAt,
		err:      err,
	}, nil
}

type taskStore struct {
	repo localrepo.Repository
}

func newTaskStore() (*taskStore, error) {
	repo, err := localrepo.New()
	if err != nil {
		return nil, err
	}
	if err = repo.Migrate(TaskEntity{}); err != nil {
		return nil, err
	}
	return &taskStore{repo: repo}, nil
}

func (store *taskStore) Close() error {
	return store.repo.Close()
}

func (store *taskStore) Save(entity TaskEntity) error {
	// remove stored entity for update
	if err := store.repo.Del(TaskEntity{ID: entity.ID}); err != nil {
		return err
	}
	return store.repo.Put(entity)
}

func (store *taskStore) All() ([]TaskEntity, error) {
	var ee []TaskEntity
	err := store.repo.GetAll(TaskEntity{}, &ee)
	return ee, err
}
