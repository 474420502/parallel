package parallel

import (
	"errors"
	"fmt"
	"testing"
	"time"
)

// Test for ExecuteHandler
func TestExecuteHandler(t *testing.T) {

	// Define the execute handler
	executeHandler := func(ctx *ExecuteContext[int, int]) (*int, error) {

		target := ctx.Target()
		if target == -1 {
			return nil, errors.New("execution error")
		}
		result := target * 2
		return &result, nil
	}

	// Define the result handler
	resultHandler := func(ctx *ResultContext[int, int]) {
		// We do nothing in the result handler for this test
		if ctx.Target() == -1 && ctx.err == nil {
			t.Error("should err is not nil", ctx.err)
		}
	}

	// Create a new Parallel executor
	p := NewParallel[int, int](executeHandler, resultHandler)

	// Set the maximum number of concurrent execute handlers
	p.SetMaxExecuteNum(4)

	// Initialize the execution channels and start the background goroutines
	p.ReadyExecute()

	// Normal execution
	p.Execute(2)
	p.Wait()
	// There should be no panic or error in the normal execution

	// Execution with error
	defer func() {
		if r := recover(); r != nil {
			t.Error("The code did not panic", r)
		}
	}()
	p.Execute(-1)
	p.Wait()
	// The execution should panic because of the error returned by the execute handler
}

// Test for ResultHandler
func TestResultHandler(t *testing.T) {
	// Define the execute handler
	executeHandler := func(ctx *ExecuteContext[int, int]) (*int, error) {
		target := ctx.Target()
		result := target * 2
		return &result, nil
	}

	// Define the result handler
	resultHandler := func(ctx *ResultContext[int, int]) {
		target := ctx.Target()
		result := ctx.Result()

		if result != target*2 {
			t.Errorf("Unexpected result for target %d: got %d, want %d", target, result, target*2)
		}
	}

	// Create a new Parallel executor
	p := NewParallel[int, int](executeHandler, resultHandler)

	// Set the maximum number of concurrent execute handlers
	p.SetMaxExecuteNum(4)

	// Initialize the execution channels and start the background goroutines
	p.ReadyExecute()

	// Add some targets for execution
	p.Execute(2)
	p.Wait()
	// The result handler should not raise any error because the result is as expected

	p.Execute(3)
	p.Wait()
	// The result handler should not raise any error because the result is as expected
}

// Test for ReadyExecute
func TestReadyExecute(t *testing.T) {
	// Define the execute handler
	executeHandler := func(ctx *ExecuteContext[int, int]) (*int, error) {
		target := ctx.Target()
		result := target * 2
		return &result, nil
	}

	// Define the result handler
	resultHandler := func(ctx *ResultContext[int, int]) {
		// We do nothing in the result handler for this test
	}

	// Create a new Parallel executor
	p := NewParallel[int, int](executeHandler, resultHandler)

	// Set the maximum number of concurrent execute handlers
	p.SetMaxExecuteNum(4)

	// Initialize the execution channels and start the background goroutines
	p.ReadyExecute()

	if p.IsExecuting() == false {
		t.Errorf("ReadyExecute failed to start execution")
	}

	// It should not start execution again if it is already executing
	p.ReadyExecute()

	if p.IsExecuting() == false {
		t.Errorf("ReadyExecute stopped execution")
	}
}

// Test for Execute
func TestExecute(t *testing.T) {
	// Define the execute handler
	executeHandler := func(ctx *ExecuteContext[int, int]) (*int, error) {
		target := ctx.Target()
		result := target * 2
		return &result, nil
	}

	// Define the result handler
	resultHandler := func(ctx *ResultContext[int, int]) {
		// We do nothing in the result handler for this test
	}

	// Create a new Parallel executor
	p := NewParallel[int, int](executeHandler, resultHandler)

	// Set the maximum number of concurrent execute handlers
	p.SetMaxExecuteNum(4)

	// Initialize the execution channels and start the background goroutines
	p.ReadyExecute()

	// Add some targets for execution
	p.Execute(2)
	p.Execute(3)

	time.Sleep(100 * time.Millisecond) // Sleep a little for the execution to start

	if p.IsExecuting() == false {
		t.Errorf("Execute failed to start execution")
	}
}

// Test for Wait
func TestWait(t *testing.T) {
	// Define the execute handler
	executeHandler := func(ctx *ExecuteContext[int, int]) (*int, error) {
		target := ctx.Target()
		result := target * 2
		return &result, nil
	}

	// Define the result handler
	resultHandler := func(ctx *ResultContext[int, int]) {
		// We do nothing in the result handler for this test
	}

	// Create a new Parallel executor
	p := NewParallel[int, int](executeHandler, resultHandler)

	// Set the maximum number of concurrent execute handlers
	p.SetMaxExecuteNum(4)

	// Initialize the execution channels and start the background goroutines
	p.ReadyExecute()

	// Add some targets for execution
	p.Execute(2)
	p.Execute(3)

	p.Wait()
	p.Cancel()

	if p.IsExecuting() == true {
		t.Errorf("Wait failed to stop execution")
	}
}

type IntTarget int
type IntResult int

func TestCancel(t *testing.T) {

	executeHandler := func(ctx *ExecuteContext[IntTarget, IntResult]) (*IntResult, error) {
		// In this example, we just multiply the target by 2 as the result.
		// In a real-world scenario, this could be a more complex operation that takes some time.
		target := ctx.Target()
		result := IntResult(target * 2)
		return &result, nil
	}

	resultHandler := func(ctx *ResultContext[IntTarget, IntResult]) {
		// We don't do anything with the result in this test.
		// In a real-world scenario, we might want to check that the result is as expected.
	}

	// Test cancellation before execution.
	p := NewParallel(executeHandler, resultHandler)
	p.Cancel()
	if p.IsExecuting() {
		t.Error("Expected executor to have stopped executing before starting")
	}
	p.Execute(IntTarget(1))
	if p.IsExecuting() {
		t.Error("Expected executor to not start executing after being cancelled")
	}

	// Test cancellation during execution.
	p = NewParallel(executeHandler, resultHandler)
	p.ReadyExecute()
	p.Execute(IntTarget(1))
	time.Sleep(10 * time.Millisecond) // Give the executor some time to start.
	p.Cancel()
	if p.IsExecuting() {
		t.Error("Expected executor to stop executing after being cancelled")
	}

	// Test cancellation after execution.
	p = NewParallel(executeHandler, resultHandler)
	p.ReadyExecute()
	p.Execute(IntTarget(1))
	p.Wait()
	p.Cancel()
	if p.IsExecuting() {
		t.Error("Expected executor to stay stopped after being cancelled")
	}
}

func TestExecuteAndResultHandling(t *testing.T) {
	executeHandler := func(ctx *ExecuteContext[IntTarget, IntResult]) (*IntResult, error) {
		target := ctx.Target()
		result := IntResult(target * 2)
		return &result, nil
	}

	resultHandler := func(ctx *ResultContext[IntTarget, IntResult]) {
		result := ctx.Result()
		expected := IntResult(ctx.Target() * 2)
		if result != expected {
			t.Errorf("Expected result %d, got %d", expected, result)
		}
	}

	p := NewParallel(executeHandler, resultHandler)
	p.ReadyExecute()

	// Test single task execution.
	p.Execute(IntTarget(1))
	p.Wait()

	// Test multiple tasks execution.
	for i := 2; i <= 10; i++ {
		p.Execute(IntTarget(i))
	}
	p.Wait()
}

func TestErrorHandlingInExecutionAndResult(t *testing.T) {
	executeHandler := func(ctx *ExecuteContext[IntTarget, IntResult]) (*IntResult, error) {
		return nil, errors.New("execute error")
	}

	resultHandler := func(ctx *ResultContext[IntTarget, IntResult]) {
		if ctx.Error() == nil || ctx.Error().Error() != "execute error" {
			t.Error("Expected an execute error, but got none")
		}
	}

	p := NewParallel(executeHandler, resultHandler)
	p.ReadyExecute()
	p.Execute(IntTarget(1))
	p.Wait()
}

func TestErrorHandling(t *testing.T) {
	executeHandler := func(ctx *ExecuteContext[IntTarget, IntResult]) (*IntResult, error) {
		// An error occurs in the execution
		return nil, errors.New("execution error")
	}

	resultHandler := func(ctx *ResultContext[IntTarget, IntResult]) {
		// An error should be received in the result
		if ctx.Error() == nil || ctx.Error().Error() != "execution error" {
			t.Error("Expected to get execution error, but got none")
		}
	}

	p := NewParallel(executeHandler, resultHandler)
	p.ReadyExecute()
	p.Execute(IntTarget(1))
	p.Wait()
}

func TestHandlersLifecycle(t *testing.T) {
	executeHandler := func(ctx *ExecuteContext[IntTarget, IntResult]) (*IntResult, error) {
		// Set a value in the context
		ctx.SetValue("key", "value")
		return nil, nil
	}

	resultHandler := func(ctx *ResultContext[IntTarget, IntResult]) {
		// Get the value from the context
		value := ctx.Value("key")
		if value != "value" {
			t.Errorf("Expected context value to be 'value', but got %v", value)
		}
	}

	p := NewParallel(executeHandler, resultHandler)
	p.ReadyExecute()
	p.Execute(IntTarget(1))
	p.Wait()
}

func Test_ExecuteHandler_Panic(t *testing.T) {
	// 创建一个新的并行执行器
	p := NewParallel(func(ctx *ExecuteContext[string, string]) (*string, error) {
		// 在这个 handler 中，我们直接触发一个 panic
		panic("test panic")
	}, func(ctx *ResultContext[string, string]) {
		// 在此处，我们期待该 panic 已经被捕获并转化为 error
		if err := ctx.Error(); err == nil {
			t.Fatalf("Expected an error, but got nil")
		} else if err.Error() != "unknown panic: test panic" {
			t.Fatalf("Expected 'unknown panic: test panic', but got '%s'", err.Error())
		}
	})

	p.ReadyExecute()
	p.Execute("test")
	p.Wait()
}

func TestParallel_Execute(t *testing.T) {
	p := NewParallel[int, int](func(ctx *ExecuteContext[int, int]) (*int, error) {
		v := ctx.Target() + 1
		return &v, nil
	}, func(ctx *ResultContext[int, int]) {
		if ctx.Result() != 2 {
			t.Error("result error")
		}
	})
	p.Execute(1)
	p.Wait()
}

func TestParallel_Cancel(t *testing.T) {
	p := NewParallel[int, int](func(ctx *ExecuteContext[int, int]) (*int, error) {
		time.Sleep(time.Second)
		v := ctx.Target() + 1
		return &v, nil
	}, func(ctx *ResultContext[int, int]) {
		if ctx.Result() != 2 {
			t.Error("result error")
		}
	})
	p.Execute(1)
	time.Sleep(time.Millisecond * 500)
	p.Cancel()
	p.Wait()
}

func TestParallel_Execute_MultipleTasks(t *testing.T) {
	p := NewParallel[int, int](func(ctx *ExecuteContext[int, int]) (*int, error) {
		v := ctx.Target() + 1
		return &v, nil
	}, func(ctx *ResultContext[int, int]) {
		if ctx.Result() != 2 {
			t.Error("")
		}
	})
	p.Execute(1)
	p.Execute(2)
	p.Wait()
}

func TestParallel_Execute_WithSharedContext(t *testing.T) {
	p := NewParallel[int, int](func(ctx *ExecuteContext[int, int]) (*int, error) {
		ctx.SetValue("key", "value")
		v := ctx.Target() + 1
		return &v, nil
	}, func(ctx *ResultContext[int, int]) {
		if ctx.Result() != 2 {
			t.Error("")
		}

		if ctx.Value("key") != "value" {
			t.Error("")
		}

	})
	p.Execute(1)
	p.Wait()
}

func TestParallel_Execute_WithError(t *testing.T) {
	p := NewParallel[int, int](func(ctx *ExecuteContext[int, int]) (*int, error) {
		v := ctx.Target() + 1
		return &v, fmt.Errorf("error")
	}, func(ctx *ResultContext[int, int]) {
		if ctx.Error() == nil {
			t.Error("")
		}
	})
	p.Execute(1)
	p.Wait()
}
