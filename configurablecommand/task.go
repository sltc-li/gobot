package configurablecommand

import (
	"github.com/li-go/gobot/gobot"
	"sync"
	"time"
)

type Task struct {
	Cmd     Command
	Msg     gobot.Message
	StartAt time.Time
	EndAt   *time.Time
	Err     error
}

const (
	maxTasks = 100
)

var (
	tasks []*Task
	mutex sync.RWMutex
)

func GetTasks() []Task {
	mutex.RLock()
	defer mutex.RUnlock()
	var tt []Task
	for _, t := range tasks {
		tt = append(tt, *t)
	}
	return tt
}

func addTask(cmd Command, msg gobot.Message) *Task {
	mutex.Lock()
	defer mutex.Unlock()
	for len(tasks) >= maxTasks {
		removeTask()
	}
	task := &Task{
		Cmd:     cmd,
		Msg:     msg,
		Err:     nil,
		StartAt: time.Now(),
		EndAt:   nil,
	}
	tasks = append(tasks, task)
	return task
}

func finishTask(task *Task, err error) {
	t := time.Now()
	task.EndAt = &t
	task.Err = err
}

func removeTask() {
	if len(tasks) == 0 {
		return
	}

	removeIdx := 0
	for i, task := range tasks {
		if LessFunc(*task, *tasks[removeIdx]) {
			removeIdx = i
		}
	}

	var newTasks []*Task
	newTasks = append(newTasks, tasks[:removeIdx]...)
	newTasks = append(newTasks, tasks[removeIdx+1:]...)
	tasks = newTasks
}

func LessFunc(task1, task2 Task) bool {
	if task1.EndAt != nil && task2.EndAt == nil {
		return true
	}
	if task1.EndAt == nil && task2.EndAt != nil {
		return false
	}
	if task1.EndAt != nil && task2.EndAt != nil {
		return task1.EndAt.Before(*task2.EndAt)
	}
	return task1.StartAt.Before(task2.StartAt)
}
