package parallel

import (
	"fmt"
	"log"
	"runtime"
	"sync"
	"sync/atomic"
)

var plog = log.Default()

type Result[RESULT any] struct {
	Value RESULT
	Err   error
}

// ExecuteContext is the context provided to the executeHandler function. It contains
// the target to be processed, a shared value map, and a boolean indicating if the execution is running.
type ExecuteContext[TARGET any, RESULT any] struct {
	target      *TARGET        // The target to be processed.
	value       map[string]any // Shared value map for communication between handlers.
	isExecuting *atomic.Bool   // Indicates if the execution is running.
}

// Target
func (ctx *ExecuteContext[TARGET, RESULT]) Target() TARGET {
	return *ctx.target
}

// SetValue sets a shared value to the context.
func (ctx *ExecuteContext[TARGET, RESULT]) SetValue(key string, value any) {
	ctx.value[key] = value
}

// Cancel cancels the execution.
func (ctx *ExecuteContext[TARGET, RESULT]) Cancel() {
	(*ctx.isExecuting).CompareAndSwap(true, false)
}

// ResultContext is the context provided to the resultHandler function. It contains
// the processed target, the result, a shared value map, an error if one occurred,
// and a boolean indicating if the execution is running.
type ResultContext[TARGET any, RESULT any] struct {
	target      *TARGET        // The processed target.
	result      *RESULT        // The result of processing the target.
	value       map[string]any // Shared value map for communication between handlers.
	err         error          // Error if one occurred during processing.
	isExecuting *atomic.Bool   // Indicates if the execution is running.
}

// Target returns the processed target from the result context.
func (ctx *ResultContext[TARGET, RESULT]) Target() TARGET {
	return *ctx.target
}

// Result returns the result of processing the target.
func (ctx *ResultContext[TARGET, RESULT]) Result() RESULT {
	return *ctx.result
}

// Value returns the shared value associated with the provided key after execution.
// If the key is not found in the shared value map, it returns nil.
func (ctx *ResultContext[TARGET, RESULT]) Value(key string) any {
	return ctx.value[key]
}

// Error returns the error that occurred during processing, if any.
func (ctx *ResultContext[TARGET, RESULT]) Error() error {
	return ctx.err
}

// Cancel cancels the execution. It switches the isExecuting flag to false.
func (ctx *ResultContext[TARGET, RESULT]) Cancel() {
	(*ctx.isExecuting).CompareAndSwap(true, false)
}

// Parallel
type Parallel[TARGET any, RESULT any] struct {
	executeChan chan TARGET

	executeMax     int
	executeHandler func(*ExecuteContext[TARGET, RESULT]) (*RESULT, error)
	resultHandler  func(*ResultContext[TARGET, RESULT])

	isExecuting atomic.Bool

	wg sync.WaitGroup
}

// NewParallel creates a new parallel executor with custom executeHandler and resultHandler.
func NewParallel[TARGET any, RESULT any](executeHandler func(ctx *ExecuteContext[TARGET, RESULT]) (*RESULT, error), resultHandler func(ctx *ResultContext[TARGET, RESULT])) *Parallel[TARGET, RESULT] {
	return &Parallel[TARGET, RESULT]{executeMax: runtime.NumCPU(), executeHandler: executeHandler, resultHandler: resultHandler}
}

// SetMaxExecuteNum sets the maximum number of concurrent execute handlers
// that can run in parallel. If the provided number is less than 1, it defaults
// to the number of CPUs available on the system (runtime.NumCPU()).
// This function panics if the provided number is less than 0.
func (p *Parallel[TARGET, RESULT]) SetMaxExecuteNum(num int) {

	if num < 0 {
		panic("execute num is not allowed to be less than 0")
	} else if num < 1 {
		num = runtime.NumCPU()
	}
	p.executeMax = num
}

// IsExecuting returns a boolean indicating whether the executor is currently executing tasks.
// It uses the atomic Load method to safely check the isExecuting flag.
func (p *Parallel[TARGET, RESULT]) IsExecuting() bool {
	return p.isExecuting.Load()
}

// ReadyExecute initializes the execution channels and starts the background goroutines.
// It will only be executed once. If it's called multiple times, the function will simply return after the first call.
func (p *Parallel[TARGET, RESULT]) ReadyExecute() {
	if !p.isExecuting.Load() {
		if p.executeChan != nil {
			close(p.executeChan)
		}
		p.isExecuting.Store(true)
		p.executeChan = make(chan TARGET, p.executeMax)
		go p.executeLoop()
	}
}

// executeLoop is the main execution loop. It reads from the executeChan and calls the
// executeHandler and resultHandler for each target in a separate goroutine.
func (p *Parallel[TARGET, RESULT]) executeLoop() {
	var limitCount chan bool = make(chan bool, p.executeMax)

	defer p.isExecuting.CompareAndSwap(true, false)

	for target := range p.executeChan {

		limitCount <- true
		go func(target TARGET) {

			var (
				ectx *ExecuteContext[TARGET, RESULT] = &ExecuteContext[TARGET, RESULT]{
					isExecuting: &p.isExecuting,
					value:       make(map[string]any),
				}

				rctx *ResultContext[TARGET, RESULT] = &ResultContext[TARGET, RESULT]{
					isExecuting: &p.isExecuting,
				}
			)

			defer func() {
				<-limitCount
				p.wg.Done()
			}()

			func() {
				defer func() {
					// 单元测试还没覆盖
					if r := recover(); r != nil {
						rerr, ok := r.(error)
						if ok {
							rctx.err = rerr
						} else {
							rctx.err = fmt.Errorf("unknown panic: %v", r)
						}
					}
					rctx.value = ectx.value
				}()
				ectx.target = &target
				rctx.result, rctx.err = p.executeHandler(ectx)
			}()

			func() {
				if !p.isExecuting.Load() {
					return
				}

				rctx.target = ectx.target
				rctx.value = ectx.value
				defer func() {
					if r := recover(); r != nil {
						plog.Println("Recovered from panic in resultHandler:", r)
					}
				}()
				p.resultHandler(rctx)
			}()

		}(target)

		if !p.isExecuting.Load() {
			break
		}
	}
}

// Execute adds a target to be executed by the parallel executor.
// If the execution is not running, the function will simply return without adding the target.
func (p *Parallel[TARGET, RESULT]) Execute(target TARGET) {
	if p.isExecuting.Load() {
		p.wg.Add(1)
		p.executeChan <- target
	}
}

// Wait waits for all the tasks to be executed and results to be processed.
func (p *Parallel[TARGET, RESULT]) Wait() {
	p.wg.Wait()
}

// Cancel cancels the execution.
// It switches the isExecuting flag to false.
func (p *Parallel[TARGET, RESULT]) Cancel() {
	p.isExecuting.CompareAndSwap(true, false)
}
