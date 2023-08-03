package tasker

import (
	"errors"
	"time"
)

// Tasker interface.
type Tasker interface {
	// Enqueue a task periodically by name.
	Enqueue(task Task) error

	// Execute a task after the duration has passed, or immediately if the duration is 0.
	// If the task name is provided, the task will be reset to the new duration.
	// If the task name is not provided, the task will be executed once after the duration has passed.
	After(task Task) error

	// Dequeue a task by name.
	Dequeue(task Task) error
}

// Public facing task structure.
//
// This can be used to configure the task to run.
type Task struct {
	Name     string
	Func     func() error
	OnError  func(error)
	Duration time.Duration
}

type tasker struct {
	taskQueue map[string]*task
}

func New() Tasker {
	return &tasker{
		taskQueue: make(map[string]*task),
	}
}

func (t *tasker) Enqueue(tsk Task) error {
	if _, ok := t.taskQueue[tsk.Name]; ok {
		var err = t.Dequeue(tsk)
		if err != nil {
			return err
		}
	}
	var executor = &task{T: &tsk}
	go executor.exec()
	t.taskQueue[tsk.Name] = executor
	return nil
}

func (t *tasker) After(task Task) error {
	if task.Name != "" {
		if tsk, ok := t.taskQueue[task.Name]; ok {
			tsk.T.Duration = task.Duration
			tsk.reset()
			go tsk.exec()
			return nil
		}
	}
	if task.Duration <= 0 {
		task.Func()
	}
	go func() {
		<-time.After(task.Duration)
		task.Func()
	}()
	return nil
}

var (
	ErrNotFound = errors.New("task not found")
)

func (t *tasker) Dequeue(task Task) error {
	if tsk, ok := t.taskQueue[task.Name]; ok {
		tsk.ticker.Stop()
		delete(t.taskQueue, task.Name)
	}
	return ErrNotFound
}

type task struct {
	T      *Task
	ticker *time.Ticker
}

func (t *task) exec() {
	t.ticker.Stop()
	t.ticker = time.NewTicker(t.T.Duration)
	var err error
	for range t.ticker.C {
		err = t.executeFunc()
		if err != nil {
			break
		}
	}
}

func (t *task) executeFunc() error {
	var err = t.T.Func()
	if err != nil {
		if t.T.OnError != nil {
			t.T.OnError(err)
		}
	}
	return err
}

func (t *task) reset() {
	t.ticker.Stop()
	t.ticker = time.NewTicker(t.T.Duration)
}
