# Parallel Module

This is a Go module that provides a way to execute tasks concurrently and handle their results in parallel. 

## Introduction

The module defines a `Parallel` struct, which is a parallel executor that receives tasks, executes them concurrently, and handles their results. The tasks and the results are typed using Go's generics feature, allowing you to use any type for them.

This module also provides two context structs: `ExecuteContext` and `ResultContext`, which carry the relevant information and provide utility methods to the handler functions during execution and result handling, respectively.

## Features

- Task execution and result handling in parallel.
- Customizable task execution and result handling logic.
- Shared value map for communication between tasks.
- Execution cancellation support.

## Usage

First, import the module:

```go
import "github.com/yourusername/parallel"
```

### Initialization

Create a new `Parallel` executor by providing execute and result handler functions:

```go
p := parallel.NewParallel(executeHandler, resultHandler)
```

Here, `executeHandler` is a function that receives an `ExecuteContext` and returns a result and an error. And `resultHandler` is a function that receives a `ResultContext`.

### Customization

You can set the maximum number of concurrent execute handlers that can run in parallel using `SetMaxExecuteNum` method:

```go
p.SetMaxExecuteNum(10) // Set to 10 concurrent tasks.
```

### Execution

Before executing tasks, you need to call `ReadyExecute` to initialize the executor:

```go
p.ReadyExecute()
```

Then, you can add tasks to be executed using `Execute` method:

```go
p.Execute(target)
```

To wait for all tasks to complete, use:

```go
p.Wait()
```

To cancel the execution, use:

```go
p.Cancel()
```

## Example

Here is a simple example of how to use this module:

```go
executeHandler := func(ctx *parallel.ExecuteContext[int, string]) (*string, error) {
	// Do some work with ctx.Target().
	// ...
	result := fmt.Sprintf("Result for target: %v", ctx.Target())
	return &result, nil
}

resultHandler := func(ctx *parallel.ResultContext[int, string]) {
	if ctx.Error() != nil {
		fmt.Println("An error occurred:", ctx.Error())
	} else {
		fmt.Println("Result for target", ctx.Target(), "is", ctx.Result())
	}
}

p := parallel.NewParallel(executeHandler, resultHandler)
p.SetMaxExecuteNum(10)
p.ReadyExecute()

for i := 0; i < 100; i++ {
	p.Execute(i)
}

p.Wait()
```

In this example, we create a new `Parallel` executor that performs some work on integers and produces a string as a result. It then waits for all tasks to complete before exiting.

## License

This project is licensed under the MIT License - see the [LICENSE.md](LICENSE.md) file for details.

## Contributing

Please read [CONTRIBUTING.md](CONTRIBUTING.md) for details on our code of conduct, and the process for submitting pull requests.