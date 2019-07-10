package configurablecommand

import (
	"errors"
	"sort"
	"sync"
	"time"

	"github.com/li-go/gobot/gobot"
)

const (
	maxTasks = 10
)

var (
	lastTaskID int
	tasks      []*Task
	mutex      sync.RWMutex

	ErrTooManyTasks = errors.New("too many tasks")
	ErrTaskNotFound = errors.New("task not found")
)

func LoadPendingTasks(bot gobot.Bot) {
	// ignore errors
	store, err := newTaskStore()
	if err != nil {
		return
	}
	taskEntities, err := store.All()
	if err != nil {
		return
	}
	for _, e := range taskEntities {
		task, err := e.Task()
		if err != nil {
			continue
		}
		status := task.Status()
		if status == Killed || status == Succeeded || status == Failed {
			continue
		}
		// restart running task
		if task.Status() == Running {
			task.startAt = nil
		}
		// use current bot
		task.bot = bot
		tasks = append(tasks, task)
	}
	sort.Slice(tasks, func(i, j int) bool {
		return tasks[i].ID < tasks[j].ID
	})
}

func saveTask(t *Task) {
	// ignore errors
	store, err := newTaskStore()
	if err != nil {
		return
	}
	entity, err := NewTaskEntity(t)
	if err != nil {
		return
	}
	_ = store.Save(*entity)
}

func GetTasks() []Task {
	mutex.RLock()
	defer mutex.RUnlock()
	var tt []Task
	for _, t := range tasks {
		tt = append(tt, *t)
	}
	return tt
}

func FindTask(id int) (*Task, error) {
	mutex.RLock()
	defer mutex.RUnlock()
	for _, t := range tasks {
		if t.ID == id {
			return t, nil
		}
	}
	return nil, ErrTaskNotFound
}

func addTask(bot gobot.Bot, msg gobot.Message, cmd Command) error {
	mutex.Lock()
	defer mutex.Unlock()
	if len(tasks) >= maxTasks {
		removeTask()
	}
	if len(tasks) >= maxTasks {
		return ErrTooManyTasks
	}

	task := &Task{
		ID:    lastTaskID + 1,
		Msg:   msg,
		bot:   bot,
		cmd:   cmd,
		runAt: time.Now(),
	}
	lastTaskID++
	tasks = append(tasks, task)
	saveTask(task)
	return nil
}

func removeTask() *Task {
	if len(tasks) == 0 {
		return nil
	}

	for i, task := range tasks {
		status := task.Status()
		if status == Killed || status == Succeeded || status == Failed {
			var newTasks []*Task
			newTasks = append(newTasks, tasks[:i]...)
			newTasks = append(newTasks, tasks[i+1:]...)
			tasks = newTasks
			return task
		}
	}
	return nil
}

// nextExecutableTask looks for pending task,
//  there should be no running task with same type (same name) of command
func nextExecutableTask() *Task {
	var runningTasks []*Task
	var pendingTasks []*Task
	for _, t := range tasks {
		status := t.Status()
		if status == Running {
			runningTasks = append(runningTasks, t)
		} else if status == Pending {
			pendingTasks = append(pendingTasks, t)
		}
	}
	for _, t := range pendingTasks {
		var found bool
		for _, running := range runningTasks {
			if running.cmd.Name == t.cmd.Name {
				found = true
			}
		}
		if !found {
			return t
		}
	}
	return nil
}

func init() {
	// how to stop
	go func() {
		tick := time.NewTicker(time.Second)
		defer tick.Stop()
		for {
			<-tick.C
			t := nextExecutableTask()
			if t != nil {
				go t.Start()
			}
		}
	}()
}
