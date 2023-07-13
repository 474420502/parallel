package parallel

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"
)

func TestNewParallel(t *testing.T) {
	ctx := context.Background()
	executeHandler := func(target int) (int, error) {
		return target, nil
	}
	resultHandler := func(result int, err error, cancel context.CancelFunc) {
		// Do nothing
	}

	p := NewParallel(ctx, executeHandler, resultHandler)
	if p == nil {
		t.Fatal("NewParallel should return a non-nil instance")
	}
}

func TestParallel2(t *testing.T) {
	ctx := context.Background()
	executeHandler := func(target int) (int, error) {
		time.Sleep(100 * time.Millisecond)
		if target == 5 {
			return -1, errors.New("error at 5")
		}
		return target * 2, nil
	}

	resultHandler := func(result int, err error, cancel context.CancelFunc) {
		if err != nil {
			t.Logf("Received error: %v", err)
			cancel()
		} else {
			t.Logf("Received result: %d", result)
		}
	}

	p := NewParallel(ctx, executeHandler, resultHandler)
	p.ReadyExecute()

	for i := 0; i < 10; i++ {
		p.Execute(i)
	}

	p.Wait()
}

func TestParallelLoopCancel(t *testing.T) {
	ctx := context.Background()

	executeHandler := func(target int) (int, error) {
		time.Sleep(100 * time.Millisecond)
		return target * 2, nil
	}

	resultHandler := func(result int, err error, cancel context.CancelFunc) {
		if result >= 10 {
			cancel()
		} else {
			t.Logf("Received result: %d", result)
		}
	}

	p := NewParallel(ctx, executeHandler, resultHandler)
	p.ReadyExecute()

	for i := 0; i < 10; i++ {
		p.Execute(i)
	}

	p.Wait()
}

func TestParallelMultipleReadyExecute(t *testing.T) {
	ctx := context.Background()

	executeHandler := func(target int) (int, error) {
		time.Sleep(100 * time.Millisecond)
		return target * 2, nil
	}

	resultHandler := func(result int, err error, cancel context.CancelFunc) {
		t.Logf("Received result: %d", result)
	}

	p := NewParallel(ctx, executeHandler, resultHandler)

	// Call ReadyExecute multiple times
	p.ReadyExecute()
	p.ReadyExecute()

	for i := 0; i < 10; i++ {
		p.Execute(i)
	}

	p.Wait()
}

func TestParallelCancel(t *testing.T) {
	ctx := context.Background()

	executeHandler := func(target int) (int, error) {
		time.Sleep(100 * time.Millisecond)
		return target * 2, nil
	}

	resultHandler := func(result int, err error, cancel context.CancelFunc) {
		t.Logf("Received result: %d", result)
		if result >= 10 {
			cancel()
		}
	}

	p := NewParallel(ctx, executeHandler, resultHandler)
	p.ReadyExecute()

	for i := 0; i < 10; i++ {
		p.Execute(i)
	}

	p.Wait()

	// Cancel after all tasks are done
	p.Cancel()

	select {
	case <-p.ctx.Done():
		t.Log("Context is canceled")
	default:
		t.Error("Context should be canceled")
	}
}

func TestReadyExecute(t *testing.T) {
	ctx := context.Background()
	executeHandler := func(target int) (int, error) {
		return target, nil
	}
	resultHandler := func(result int, err error, cancel context.CancelFunc) {
		// Do nothing
	}

	p := NewParallel(ctx, executeHandler, resultHandler)
	p.ReadyExecute()

	if p.executeChan == nil {
		t.Fatal("executeChan should be initialized after calling ReadyExecute")
	}

	if p.resultChan == nil {
		t.Fatal("resultChan should be initialized after calling ReadyExecute")
	}
}

func TestExecuteLoop(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	executeHandler := func(target int) (int, error) {
		return target * 2, nil
	}
	resultHandler := func(result int, err error, cancel context.CancelFunc) {
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		expected := result / 2
		if result != expected*2 {
			t.Errorf("Expected result: %d, got: %d", expected*2, result)
		}
	}

	p := NewParallel(ctx, executeHandler, resultHandler)
	p.ReadyExecute()

	// Test if tasks execute correctly
	for i := 1; i <= 10; i++ {
		p.Execute(i)
	}
	p.Wait()

	// Test if panic is handled correctly

	executeHandlerWithPanic := func(target int) (int, error) {
		panic(fmt.Errorf("executeHandlerWithPanic test panic"))
	}

	panicResultHandler := func(result int, err error, cancel context.CancelFunc) {
		if err != nil {
			t.Error("executeHandlerWithPanic is not panic")
		}
	}

	pPanic := NewParallel(ctx, executeHandlerWithPanic, panicResultHandler)
	pPanic.ReadyExecute()

	pPanic.Execute(1)
	pPanic.Wait()

}

func TestResultLoop(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	executeHandler := func(target int) (int, error) {
		return target * 2, nil
	}
	resultHandler := func(result int, err error, cancel context.CancelFunc) {
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		expected := result / 2
		if result != expected*2 {
			t.Errorf("Expected result: %d, got: %d", expected*2, result)
		}
	}

	p := NewParallel(ctx, executeHandler, resultHandler)
	p.ReadyExecute()

	// Test if results are processed correctly
	for i := 1; i <= 10; i++ {
		p.Execute(i)
	}
	p.Wait()

	// Test if panic in resultHandler is handled correctly
	resultHandlerWithPanic := func(result int, err error, cancel context.CancelFunc) {
		panic(fmt.Errorf("test panic in resultHandler"))
	}

	pPanic := NewParallel(ctx, executeHandler, resultHandlerWithPanic)
	pPanic.ReadyExecute()

	pPanic.Execute(1)
	pPanic.Wait()
}

func TestExecute(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	executeHandler := func(target int) (int, error) {
		return target * 2, nil
	}
	resultHandler := func(result int, err error, cancel context.CancelFunc) {
		// Do nothing
	}

	p := NewParallel(ctx, executeHandler, resultHandler)
	p.ReadyExecute()

	// Test if tasks are added to the execution channel
	for i := 1; i <= 10; i++ {
		p.Execute(i)
	}

	p.Wait()
}

func TestWait(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	executeHandler := func(target int) (int, error) {
		time.Sleep(500 * time.Millisecond)
		return target * 2, nil
	}
	resultHandler := func(result int, err error, cancel context.CancelFunc) {
		// Do nothing
	}

	p := NewParallel(ctx, executeHandler, resultHandler)
	p.ReadyExecute()

	// Test if Wait() blocks until all tasks are completed
	startTime := time.Now()

	for i := 1; i <= 4; i++ {
		p.Execute(i)
	}

	p.Wait()

	duration := time.Since(startTime)
	if duration > 1*time.Second {
		t.Errorf("Wait() took too long, expected less than 1s, but got: %v", duration)
	}
}
