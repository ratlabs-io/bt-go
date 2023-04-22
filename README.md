# Go Behavior Tree Library

`go-bt` is a library for creating and executing behavior trees in Go. It provides a set of interfaces and implementations for creating and running behavior trees, as well as a set of utility functions for creating and manipulating behavior tree nodes.

## Installation

To use this package in your Go project, simply import it as follows:

```
import "github.com/ratlabs-io/go-bt"
```

## Usage

Here's an example of how to use `go-bt` to create a simple behavior tree:

```go
package main

import (
	"context"
	"fmt"

	"github.com/ratlabs-io/go-bt"
)

func main() {
	// Create a new sequence with two actions
	sequence := bt.NewSequence(
		bt.NewAction(func(ctx *bt.BehaviorContext) bt.RunStatus {
			fmt.Println("Action 1")
			return bt.Success
		}),
		bt.NewAction(func(ctx *bt.BehaviorContext) bt.RunStatus {
			fmt.Println("Action 2")
			return bt.Success
		}),
	)

	// Create a new behavior tree with the sequence as the root
	tree := bt.NewBehaviorTree(sequence)

	// Create a new behavior context
	ctx := bt.NewBehaviorContext(context.Background())

	// Tick the behavior tree
	status := tree.Tick(ctx)

	// Print the status
	fmt.Println(status)
}
```
