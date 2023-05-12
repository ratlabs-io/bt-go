# Go Behavior Tree Library

`bt-go` is a library for creating and executing behavior trees in Go. It provides a set of interfaces and implementations for creating and running behavior trees, as well as a set of utility functions for creating and manipulating behavior tree nodes.

## Installation

To use this package in your Go project, simply import it as follows:

```
import "github.com/ratlabs-io/bt-go"
```

## Usage

Here's an example of how to use `bt-go` to create a simple behavior tree:

```go
package main

import (
	"context"
	"fmt"

	. "github.com/ratlabs-io/bt-go"
)

func main() {
	// Create a new sequence with two actions
	root := NewSequence(
		NewAction(func(ctx *BehaviorContext) RunStatus {
			hello, ok := ctx.Get("my_prop_a")
			if !ok {
				return Failure
			}

			fmt.Printf("%s ", hello.(string))

			return Success
		}),
		NewAction(func(ctx *BehaviorContext) RunStatus {
			world, ok := ctx.Get("my_prop_b")
			if !ok {
				return Failure
			}

			fmt.Printf("%s\n", world.(string))
			return Success
		}),
	)

	// Create a new behavior context
	ctx := NewBehaviorContext(context.Background())

	ctx.Set("my_prop_a", "Hello")
	ctx.Set("my_prop_b", "World")

	root.Tick(ctx)
	// Output: Hello World
}
```

## BehaviorContext

The BehaviorContext is a synchronized context used for sharing data between actions in a behavior tree. It provides a way to store and retrieve values associated with specific keys in a thread-safe manner.

```go
// Create a new behavior context
ctx := NewBehaviorContext(context.Background())

// Set properties in the behavior context
ctx.Set("my_prop_a", "Hello")
ctx.Set("my_prop_b", "World")

// Retrieve properties from the behavior context
valueA, ok := ctx.Get("my_prop_a")
if ok {
    fmt.Println(valueA.(string)) // Output: Hello
}

valueB, ok := ctx.Get("my_prop_b")
if ok {
    fmt.Println(valueB.(string)) // Output: World
}
```

## Action

In the context of a behavior tree, an Action represents a node that performs a specific action. It is a struct that encapsulates an action function, which takes a BehaviorContext as input and returns a RunStatus. When the Tick method is called on an Action instance with a provided BehaviorContext, it executes the action function and returns the resulting status. Actions are fundamental building blocks of behavior trees, allowing you to define custom logic and behaviors for specific nodes in the tree.

```go
// Create an action that increments a counter in the behavior context
incrementCounter := NewAction(func(ctx *BehaviorContext) RunStatus {
    counter, ok := ctx.Get("counter")
    if !ok {
        return Failure
    }

    // Increment the counter
    ctx.Set("counter", counter.(int)+1)

    return Success
})
```

## Selector

## Conditional

## BinarySelector

## Sequence

## PrioritySelector

## Switch

## TreeRunner
