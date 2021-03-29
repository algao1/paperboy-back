package tasker

import (
	"fmt"
	"log"
	"paperboy-back"
	"reflect"
	"runtime"
	"time"
)

// Factory is responsible for creating Taskers assigned with tasks.
type Factory struct{}

var _ paperboy.TaskerFactory = (*Factory)(nil)

// Tasker is responsible for executing the assigned task according to
// the given configurations.
type Tasker struct {
	Config *paperboy.TaskConfig
	Task   paperboy.Task
	Params *[]paperboy.Parameter
}

// CreateTasker returns a Tasker with the given configuration, task, and function parameters.
func (f *Factory) CreateTasker(conf paperboy.TaskConfig, task paperboy.Task, params ...paperboy.Parameter) (paperboy.Tasker, error) {
	taskValue := reflect.ValueOf(task)
	if taskValue.Kind() != reflect.Func {
		return &Tasker{}, fmt.Errorf("[%s] failed to create tasker: provided task is not a function", conf.Name)
	}

	errorInterface := reflect.TypeOf((*error)(nil)).Elem()
	taskReturn := taskValue.Type().Out(0)
	if !taskReturn.Implements(errorInterface) {
		return &Tasker{}, fmt.Errorf("[%s] failed to create tasker: function %s does not return error", conf.Name,
			runtime.FuncForPC(taskValue.Pointer()).Name())
	}

	return &Tasker{Config: &conf, Task: task, Params: &params}, nil
}

// Start begins executing the task at assigned intervals.
func (t *Tasker) Start() {
	log.Printf("[%v] starting task...\n", t.Config.Name)

	// Executes the task in a separate goroutine.
	go func() {
		recovering := false
		ticker := time.NewTicker(t.Config.Period)
		defer ticker.Stop()

		// Process the parameters.
		task := reflect.ValueOf(t.Task)
		params := make([]reflect.Value, len(*t.Params))
		for idx, param := range *t.Params {
			params[idx] = reflect.ValueOf(param)
		}

		// Executes task immediately upon startup.
		for ; true; <-ticker.C {
			err := task.Call(params)
			if _, ok := err[0].Interface().(error); ok {
				log.Printf("[%s] an error has occured: %s\n", t.Config.Name, err)
				ticker.Reset(t.Config.RecoverPeriod)
				recovering = true
				continue
			}

			// If Tasker has recovered from task, reset to normal ticker.
			if recovering {
				ticker.Reset(t.Config.Period)
				recovering = false
			}
		}
	}()
}
