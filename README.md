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

The Selector is a behavior tree node that represents a selector. It iterates over its child nodes and returns the first non-failure status encountered. If a child node succeeds, the selector returns success. If a child node is running, the selector returns running. If all child nodes fail, the selector returns failure. The Selector struct extends the Composite struct and maintains an index to keep track of the last running child node.

```go

// Create a selector with the child nodes
selector := NewSelector(
  NewAction(func(ctx *BehaviorContext) RunStatus {
    // Perform some action
    return Success
  }),
  NewAction(func(ctx *BehaviorContext) RunStatus {
      // Perform some action
      return Success
  }),
  NewAction(func(ctx *BehaviorContext) RunStatus {
      // Perform some action
      return Success
  })
)

// Create a behavior context
ctx := NewBehaviorContext(context.Background())

// Execute the selector
status := selector.Tick(ctx)

fmt.Println(status)
```

}

````

## Conditional

The Conditional is a behavior tree node that conditionally executes an action based on a given condition. It consists of a Condition and an Action. The Condition represents a condition node that checks a condition using a provided check function. The Action represents the action node to be executed if the condition evaluates to true. If the condition is true, the Conditional node returns the result of executing the action; otherwise, it returns failure.

```go
// Create a conditional node with the condition and action
conditionalNode := NewConditional(
  NewCondition(func(ctx *BehaviorContext) bool {
    // Perform some condition check
    return ctx.Get("is_ready").(bool)
  }),
  NewAction(func(ctx *BehaviorContext) RunStatus {
      // Perform some action
      return Success
  })
)

// Create a behavior context
ctx := NewBehaviorContext(context.Background())

// Set a property in the behavior context
ctx.Set("is_ready", true)

// Execute the node
status := selector.Tick(ctx)

fmt.Println(status)
````

## BinarySelector

## Sequence

## PrioritySelector

## Switch

## TreeRunner
