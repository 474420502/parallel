package parallel

import (
	"context"
	"fmt"
	"log"
	"runtime"
	"sync"
)

var plog = log.Default()

type Result[RESULT any] struct {
	Value RESULT
	Err   error
}

type Parallel[TARGET any, RESULT any] struct {
	zeroResult     RESULT
	executeChan    chan TARGET
	resultChan     chan Result[RESULT]
	executeMax     int
	executeHandler func(target TARGET) (RESULT, error)
	resultHandler  func(result RESULT, err error, cancel context.CancelFunc)

	wg     sync.WaitGroup
	ctx    context.Context
	cancel context.CancelFunc
	once   sync.Once
}

// NewParallel creates a new parallel executor with custom executeHandler, resultHandler, and errorHandler.
func NewParallel[TARGET any, RESULT any](ctx context.Context, executeHandler func(target TARGET) (RESULT, error), resultHandler func(result RESULT, err error, cancalFunc context.CancelFunc)) *Parallel[TARGET, RESULT] {
	ctx, cancel := context.WithCancel(ctx)
	return &Parallel[TARGET, RESULT]{ctx: ctx, cancel: cancel, executeMax: runtime.NumCPU(), executeHandler: executeHandler, resultHandler: resultHandler}
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

// ReadyExecute initializes the execution channels and starts the background goroutines.
// It will only be executed once.
func (p *Parallel[TARGET, RESULT]) ReadyExecute() {
	p.once.Do(func() {
		p.executeChan = make(chan TARGET, p.executeMax)
		p.resultChan = make(chan Result[RESULT], p.executeMax)

		go p.executeLoop()
		go p.resultLoop()
	})
}

func (p *Parallel[TARGET, RESULT]) executeLoop() {
	defer close(p.executeChan)
	defer close(p.resultChan)

	var limitCount chan bool = make(chan bool, p.executeMax)

	for {
		select {
		case <-p.ctx.Done():
			return
		case target, ok := <-p.executeChan:
			if !ok {
				return
			}
			limitCount <- true
			go func(target TARGET) {

				var (
					result RESULT
					err    error
				)

				defer func() {
					if r := recover(); r != nil {
						rerr, ok := r.(error)
						if !ok {
							err = fmt.Errorf("unknown panic: %v", r)
						} else {
							err = rerr
						}
						p.resultChan <- Result[RESULT]{p.zeroResult, err}
					} else {
						p.resultChan <- Result[RESULT]{result, nil}
					}
					<-limitCount
					p.wg.Done()
				}()

				result, err = p.executeHandler(target)

			}(target)
		}
	}
}

func (p *Parallel[TARGET, RESULT]) resultLoop() {
	for result := range p.resultChan {
		go func(r Result[RESULT]) {
			defer func() {
				if err := recover(); err != nil {
					plog.Println("Recovered from panic in resultHandler:", err)
				}
			}()
			p.resultHandler(r.Value, r.Err, p.cancel)
		}(result)
	}
}

// Execute adds a target to be executed by the parallel executor.
func (p *Parallel[TARGET, RESULT]) Execute(target TARGET) {
	p.wg.Add(1)
	p.executeChan <- target
}

// Wait waits for all the tasks to be executed and results to be processed.
func (p *Parallel[TARGET, RESULT]) Wait() {
	p.wg.Wait()
}

// Cancel cancels the context
func (p *Parallel[TARGET, RESULT]) Cancel() {
	p.cancel()
}
