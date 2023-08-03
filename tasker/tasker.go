package tasker

import (
	"context"
	"time"

	"github.com/Nigel2392/jsext/v2/errs"
)

type contextKey struct{}

var TASKER_CONTEXT_KEY = contextKey{}

// Tasker interface.
type Tasker interface {
	// Enqueue a task periodically by name.
	// If there is an error, it will be of type:
	// - ErrNoNameSpecified
	// - ErrDurationLTEZero
	Enqueue(task Task) error

	// Execute a task after the duration has passed, or immediately if the duration is 0.
	// If duration is zero, the task will be executed immediately in a goroutine, but the function will return ErrDurationLTEZero.
	// If the task name is provided, the task will be reset to the new duration.
	// If the task name is not provided, the task will be executed once after the duration has passed.
	After(task Task) error

	// Dequeue a task by name.
	// If there is an error, it will be of types:
	// - ErrNoNameSpecified
	// - ErrNotFound
	Dequeue(taskName string) error
}

var (
	ErrNoNameSpecified = errs.Error("no name specified")
	ErrDurationLTEZero = errs.Error("duration less than or equal to zero")
	ErrNotFound        = errs.Error("task not found")
)

// Public facing task structure.
//
// This can be used to configure the task to run.
type Task struct {
	Name      string
	Func      func(ctx context.Context) error
	OnError   func(error)
	Duration  time.Duration
	OnDequeue func(ctx context.Context) error
}

type tasker struct {
	taskQueue map[string]*task
}

func New() Tasker {
	return &tasker{
		taskQueue: make(map[string]*task),
	}
}

func (t *tasker) defaultCtx() context.Context {
	return context.WithValue(context.Background(), TASKER_CONTEXT_KEY, t)
}

func (t *tasker) Enqueue(tsk Task) error {
	if tsk.Name == "" {
		return ErrNoNameSpecified
	}
	if tsk.Duration <= 0 {
		return ErrDurationLTEZero
	}
	if _, ok := t.taskQueue[tsk.Name]; ok {
		var err = t.Dequeue(tsk.Name)
		if err != nil {
			return err
		}
	}
	var executor = &task{
		T:   &tsk,
		ctx: t.defaultCtx(),
	}
	go executor.exec()
	t.taskQueue[tsk.Name] = executor
	return nil
}

func (t *tasker) After(task Task) error {
	var err error
	if task.Name != "" {
		// Reset the task if it already exists.
		if tsk, ok := t.taskQueue[task.Name]; ok {
			tsk.T.Duration = task.Duration
			go tsk.exec()
			return nil
		}
	}
	if task.Duration <= 0 {
		// Execute immediately, in a goroutine in case it is a long running task.
		go func() {
			err = task.Func(t.defaultCtx())
			if err != nil && task.OnError != nil {
				task.OnError(err)
			}
		}()
		return ErrDurationLTEZero
	}
	go func() {
		<-time.After(task.Duration)
		err = task.Func(t.defaultCtx())
		if err != nil && task.OnError != nil {
			task.OnError(err)
		}
	}()
	return nil
}

func (t *tasker) Dequeue(taskName string) error {
	if taskName == "" {
		return ErrNoNameSpecified
	}
	if tsk, ok := t.taskQueue[taskName]; ok {
		var err error
		if tsk.ticker != nil {
			tsk.ticker.Stop()
		}
		if tsk.T.OnDequeue != nil {
			err = tsk.T.OnDequeue(tsk.ctx)
		}
		delete(t.taskQueue, taskName)
		return err
	}
	return ErrNotFound
}

type task struct {
	T      *Task
	ticker *time.Ticker
	ctx    context.Context
}

func (t *task) exec() {
	t.reset()
	var err error
	for range t.ticker.C {
		err = t.executeFunc()
		if err != nil {
			break
		}
	}
}

func (t *task) executeFunc() error {
	var err = t.T.Func(t.ctx)
	if err != nil {
		if t.T.OnError != nil {
			t.T.OnError(err)
		}
	}
	return err
}

func (t *task) reset() {
	if t.ticker != nil {
		t.ticker.Stop()
	}
	t.ticker = time.NewTicker(t.T.Duration)
}
