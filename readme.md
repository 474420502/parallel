Here's an example of a README.md file in English for the `parallel` package:

```markdown
# Parallel

`parallel` is a simple Go generics library for parallel task execution. It allows users to define custom execute handlers and result handlers for performing tasks in a parallel environment and processing results when tasks are completed.

## Features

- Implemented using generics, capable of handling tasks and results of any type
- Customizable execute and result handlers
- Supports limiting the number of concurrent execute handlers
- Cancellation and timeout support through `context`
- Automatically handles panics in concurrent execute and result handlers

## Installation

Using Go Modules:

```
go get github.com/474420502/parallel
```

## Usage

First, import the `parallel` package:

```go
import "github.com/474420502/parallel"
```

Next, create a `Parallel` instance and provide custom execute and result handlers:

```go
executeHandler := func(target int) (int, error) {
    return target * 2, nil
}
resultHandler := func(result int, err error) {
    if err != nil {
        log.Println("Error:", err)
    } else {
        log.Println("Result:", result)
    }
}

p := parallel.NewParallel(context.Background(), executeHandler, resultHandler)
```

Initialize the execution channels and start the background goroutines:

```go
p.ReadyExecute()
```

Add tasks to be executed in parallel:

```go
for i := 1; i <= 10; i++ {
    p.Execute(i)
}
```

Wait for all tasks to complete and for the results to be processed:

```go
p.Wait()
```

Cancel the context (optional):

```go
p.Cancel()
```

## Example

Here's a complete example showing how to use the `parallel` package to execute tasks in parallel and process their results:

```go
package main

import (
	"context"
	"log"
	"github.com/474420502/parallel"
)

func main() {
    executeHandler := func(target int) (int, error) {
        return target * 2, nil
    }
    resultHandler := func(result int, err error) {
        if err != nil {
            log.Println("Error:", err)
        } else {
            log.Println("Result:", result)
        }
    }
    
    p := parallel.NewParallel(context.Background(), executeHandler, resultHandler)
    p.SetMaxExecuteNum(4) // Optional: Set max number of concurrent execute handlers
    p.ReadyExecute()
    
    for i := 1; i <= 10; i++ {
        p.Execute(i)
    }
    
    p.Wait()
}
```

## License

The `parallel` package is licensed under the [MIT License](LICENSE).
```

This README.md file provides an overview of the `parallel` package, its features, installation, usage, and a complete example. Replace `yourusername` with your GitHub username in the import paths.