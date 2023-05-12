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

## Conditional

The Conditional is a behavior tree node that conditionally executes an action based on a given condition. It consists of a Condition and an Action. The Condition represents a condition node that checks a condition using a provided check function. The Action represents the action node to be executed if the condition evaluates to true. If the condition is true, the Conditional node returns the result of executing the action; otherwise, it returns failure.

```go
// Create a conditional node with the condition and action
conditionalNode := NewConditional(
  NewCondition(func(ctx *BehaviorContext) bool {
    // Perform some condition check
    result, _ := ctx.Get("is_ready")
    return result.(bool)
  }),
  NewAction(func(ctx *BehaviorContext) RunStatus {
    // Perform some action
    return Success
  }),
)

// Create a behavior context
ctx := NewBehaviorContext(context.Background())

// Set a property in the behavior context
ctx.Set("is_ready", true)

// Execute the node
status := conditionalNode.Tick(ctx)

fmt.Println(status)
```

## BinarySelector

The BinarySelector is a behavior tree node that conditionally executes one of two child nodes based on a condition. It consists of a Condition node, an IfTrue node, and an IfFalse node. The Condition node is evaluated, and if it returns success, the IfTrue node is executed. Otherwise, the IfFalse node is executed. The BinarySelector allows for branching behavior in the behavior tree based on the outcome of the condition.

```go
// Create a condition node
conditionNode := NewCondition(func(ctx *BehaviorContext) bool {
  // Perform some condition check
  result, _ := ctx.Get("is_ready")
  return result.(bool)
})

// Create an action node for the true branch
trueActionNode := NewAction(func(ctx *BehaviorContext) RunStatus {
  // Perform some action for the true branch
  return Success
})

// Create an action node for the false branch
falseActionNode := NewAction(func(ctx *BehaviorContext) RunStatus {
  // Perform some action for the false branch
  return Failure
})

// Create a binary selector node with the condition and action nodes
binarySelectorNode := NewBinarySelector(
  conditionNode,
  trueActionNode,
  falseActionNode,
)

// Create a behavior context
ctx := NewBehaviorContext(context.Background())

// Set a property in the behavior context
ctx.Set("is_ready", true)

// Execute the binary selector node
status := binarySelectorNode.Tick(ctx)

fmt.Println(status)
```

## Sequence

The Sequence is a behavior tree node that processes its child nodes in sequence until one fails or is running. It inherits from the Composite struct and maintains an index to track the last running child node. During execution, the Sequence node iterates over its child nodes, calling their Tick method with the provided BehaviorContext. If any child node fails, the Sequence node returns failure. If any child node is running, it returns running. Only when all child nodes succeed, the Sequence node returns success.

```go
// Create child nodes
node1 := NewAction(func(ctx *BehaviorContext) RunStatus {
  // Perform some action
  return Success
})
node2 := NewAction(func(ctx *BehaviorContext) RunStatus {
  // Perform some action
  return Success
})
node3 := NewAction(func(ctx *BehaviorContext) RunStatus {
  // Perform some action
  return Success
})

// Create a sequence with the child nodes
sequence := NewSequence(node1, node2, node3)

// Create a behavior context
ctx := NewBehaviorContext(context.Background())

// Execute the sequence
status := sequence.Tick(ctx)

fmt.Println(status)
```

## PrioritySelector

The PrioritySelector is a behavior tree node that selects the first child node that succeeds and returns failure if none of the children succeed. It inherits from the Composite struct and maintains an index to track the last running child node. During execution, the PrioritySelector node iterates over its child nodes, calling their Tick method with the provided BehaviorContext. If any child node succeeds, the PrioritySelector node returns that status. If a child node is running, it returns running. If none of the child nodes succeed, it returns failure.

```go
// Create child nodes
node1 := NewAction(func(ctx *BehaviorContext) RunStatus {
    // Perform some action
    return Failure
})
node2 := NewAction(func(ctx *BehaviorContext) RunStatus {
    // Perform some action
    return Success
})
node3 := NewAction(func(ctx *BehaviorContext) RunStatus {
    // Perform some action
    return Success
})

// Create a priority selector with the child nodes
prioritySelector := NewPrioritySelector(node1, node2, node3)

// Create a behavior context
ctx := NewBehaviorContext(context.Background())

// Execute the priority selector
status := prioritySelector.Tick(ctx)

fmt.Println(status)
```

## Switch

The Switch is a behavior tree node that selects one of multiple child nodes based on a key returned by a key function. It consists of a KeyFunc type, which is a function that takes a BehaviorContext and returns a string key. The Switch node also has a map of cases, where each key in the map corresponds to a specific behavior node. If the key returned by the KeyFunc matches a key in the cases map, the corresponding behavior node is executed. If no match is found and a default behavior is provided, it is executed. If no match is found and no default behavior is provided, the Switch node returns failure.

```go
// Define a key function
keyFunc := func(ctx *BehaviorContext) string {
  // Perform some logic to determine the key
  key, _ := ctx.Get("condition_key")
  return key.(string)
}

// Create child behavior nodes
case1Behavior := NewAction(func(ctx *BehaviorContext) RunStatus {
  // Perform action for case 1
  return Success
})
case2Behavior := NewAction(func(ctx *BehaviorContext) RunStatus {
  // Perform action for case 2
  return Success
})
defaultBehavior := NewAction(func(ctx *BehaviorContext) RunStatus {
  // Perform default action
  return Success
})

// Create cases map with key-behavior pairs
cases := map[string]Behavior{
    "case1": case1Behavior,
    "case2": case2Behavior,
}

// Create a switch node with the key function, cases, and default behavior
switchNode := NewSwitch(keyFunc, cases, defaultBehavior)

// Create a behavior context
ctx := NewBehaviorContext(context.Background())

// Set the condition key in the context
ctx.Set("condition_key", "case2")

// Execute the switch node
status := switchNode.Tick(ctx)

fmt.Println(status)
```

## TreeRunner

The TreeRunner is a utility that runs a behavior tree with a specified tick rate. It takes a behavior tree as input and provides a Run method to execute the tree continuously until the provided behavior context's context is done. The tree is ticked repeatedly within a loop, allowing the behavior tree to make progress over time.

```go
// Create the behavior tree
rootNode := NewSequence(
  NewAction(func(ctx *BehaviorContext) RunStatus {
    // Perform some action
    return Success
  }),
  NewAction(func(ctx *BehaviorContext) RunStatus {
    // Perform some other action
    return Success
  }),
)

// Create a behavior context
ctx := NewBehaviorContext(context.Background())

// Create a tree runner with the behavior tree
treeRunner := NewTreeRunner(rootNode)

// Run the behavior tree until the context is done
treeRunner.Run(ctx)
```
