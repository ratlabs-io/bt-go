# bt-go: Go Behavior Tree Library

[![Go Reference](https://pkg.go.dev/badge/github.com/ratlabs-io/bt-go.svg)](https://pkg.go.dev/github.com/ratlabs-io/bt-go)

A modern, concurrent, and flexible behavior tree implementation in Go.

## Table of Contents

- [Overview](#overview)
- [Installation](#installation)
- [Core Concepts](#core-concepts)
- [Usage](#usage)
- [API Reference](#api-reference)
  - [BehaviorContext](#behaviorcontext)
  - [Action](#action)
  - [Sequence](#sequence)
  - [Selector](#selector)
  - [Parallel](#parallel)
  - [Conditional](#conditional)
  - [BinarySelector](#binaryselector)
  - [PrioritySelector](#priorityselector)
  - [Switch](#switch)
  - [TreeRunner](#treerunner)
- [Best Practices](#best-practices)
- [Contributing](#contributing)
- [License](#license)

## Overview

Behavior trees are a technique used in artificial intelligence to model and manage complex decision-making logic. They
provide a structured approach to designing the behavior of agents in video games, robotics, and other systems requiring
sophisticated and maintainable decision processes.

`bt-go` offers a comprehensive implementation of behavior trees in Go with:

- **Thread-safe** operations for concurrent systems
- **Composable** nodes for building complex behavior hierarchies
- **Stateful** execution for maintaining context between runs
- **Flexible** node types for various decision-making patterns
- **Visualization** tools for debugging and analysis

## Installation

```bash
go get github.com/ratlabs-io/bt-go
```

## Core Concepts

Behavior trees operate on a few fundamental concepts:

- **Nodes**: The building blocks of a behavior tree, each representing a specific behavior or decision
- **Tick**: The process of executing a node, which returns one of three states:
  - `Success`: The node has completed successfully
  - `Failure`: The node has failed to complete
  - `Running`: The node is still in progress and needs more time to complete
- **Context**: Shared data that nodes can access and modify during execution

## Usage

Here's a simple example of a behavior tree that prints "Hello World":

```go
package main

import (
	"context"
	"fmt"

	bt "github.com/ratlabs-io/bt-go"
)

func main() {
	// Create a tree with a sequence of two actions
	tree := bt.NewSequence(
		// First action: Get and print "Hello"
		bt.NewAction(func(ctx bt.BehaviorContext) bt.RunStatus {
			hello, ok := ctx.Get("greeting")
			if !ok {
				return bt.Failure
			}
			fmt.Printf("%s ", hello.(string))
			return bt.Success
		}),

		// Second action: Get and print "World"
		bt.NewAction(func(ctx bt.BehaviorContext) bt.RunStatus {
			world, ok := ctx.Get("subject")
			if !ok {
				return bt.Failure
			}
			fmt.Printf("%s\n", world.(string))
			return bt.Success
		}),
	)

	// Create and populate the behavior context
	ctx := bt.NewBehaviorContext(context.Background())
	ctx.Set("greeting", "Hello")
	ctx.Set("subject", "World")

	// Execute the tree
	status := tree.Tick(ctx)

	fmt.Printf("Tree execution result: %v\n", status)
	// Output:
	// Hello World
	// Tree execution result: Success
}
```

For more complex examples, see the [examples](./examples) directory.

## API Reference

### BehaviorContext

`BehaviorContext` provides a thread-safe key-value store for sharing data between nodes in a behavior tree.

```go
// Create a new context with parent context for cancellation
ctx := bt.NewBehaviorContext(context.Background())

// Store data
ctx.Set("key", "value")
ctx.Set("number", 42)
ctx.Set("enabled", true)

// Retrieve data with type assertions
value, exists := ctx.Get("key")
if exists {
    fmt.Println(value.(string)) // "value"
}

// Check if a key exists
if ctx.Has("number") {
    num, _ := ctx.Get("number")
    fmt.Println(num.(int)) // 42
}

// Delete a key
ctx.Delete("key")
```

### Action

`Action` is a leaf node that performs a specific task and returns a status.

```go
// Create an action that checks if a user is authenticated
isAuthenticated := bt.NewAction(func(ctx bt.BehaviorContext) bt.RunStatus {
    // Try to get the user from context
    user, ok := ctx.Get("user")
    if !ok || user == nil {
        fmt.Println("Authentication failed: No user found")
        return bt.Failure
    }

    // User exists, mark as authenticated
    fmt.Println("User authenticated:", user)
    ctx.Set("authenticated", true)
    return bt.Success
})

// Execute the action
status := isAuthenticated.Tick(ctx)
fmt.Println("Authentication status:", status)
```

### Sequence

`Sequence` executes child nodes in order until one fails or all succeed.

```go
// Create a sequence that opens a door
openDoor := bt.NewSequence(
    // First, check if user has key
    bt.NewAction(func(ctx bt.BehaviorContext) bt.RunStatus {
        hasKey, ok := ctx.Get("has_key")
        if !ok || !hasKey.(bool) {
            fmt.Println("Cannot open door: No key available")
            return bt.Failure
        }
        fmt.Println("Key found")
        return bt.Success
    }),

    // Next, unlock the door
    bt.NewAction(func(ctx bt.BehaviorContext) bt.RunStatus {
        fmt.Println("Unlocking the door...")
        // Here we could add logic to handle unlocking failures
        return bt.Success
    }),

    // Finally, open the door
    bt.NewAction(func(ctx bt.BehaviorContext) bt.RunStatus {
        fmt.Println("Opening the door...")
        ctx.Set("door_open", true)
        return bt.Success
    }),
)

// Set up context and execute
ctx := bt.NewBehaviorContext(context.Background())
ctx.Set("has_key", true)

// This will run all three actions in sequence and succeed
status := openDoor.Tick(ctx)
fmt.Println("Door operation status:", status)
```

### Selector

`Selector` executes child nodes in order until one succeeds or all fail.

```go
// Create a selector that tries different ways to open a door
tryOpenDoor := bt.NewSelector(
    // Try using a key
    bt.NewSequence(
        // Check if we have a key
        bt.NewAction(func(ctx bt.BehaviorContext) bt.RunStatus {
            hasKey, ok := ctx.Get("has_key")
            if !ok || !hasKey.(bool) {
                fmt.Println("No key available")
                return bt.Failure
            }
            return bt.Success
        }),
        // Use the key
        bt.NewAction(func(ctx bt.BehaviorContext) bt.RunStatus {
            fmt.Println("Opening door with key...")
            ctx.Set("door_open", true)
            return bt.Success
        }),
    ),

    // Try picking the lock
    bt.NewSequence(
        // Check if we have a lockpick
        bt.NewAction(func(ctx bt.BehaviorContext) bt.RunStatus {
            hasLockpick, ok := ctx.Get("has_lockpick")
            if !ok || !hasLockpick.(bool) {
                fmt.Println("No lockpick available")
                return bt.Failure
            }
            return bt.Success
        }),
        // Use the lockpick
        bt.NewAction(func(ctx bt.BehaviorContext) bt.RunStatus {
            fmt.Println("Picking the lock...")
            ctx.Set("door_open", true)
            return bt.Success
        }),
    ),

    // Last resort: Break down the door
    bt.NewAction(func(ctx bt.BehaviorContext) bt.RunStatus {
        fmt.Println("Breaking down the door...")
        ctx.Set("door_open", true)
        ctx.Set("door_damaged", true)
        return bt.Success
    }),
)

// Set up context and execute
ctx := bt.NewBehaviorContext(context.Background())
// No key or lockpick, so it should break down the door
status := tryOpenDoor.Tick(ctx)
fmt.Println("Door status:", status)
```

### Parallel

`Parallel` executes all child nodes concurrently with configurable success policies.

```go
// Create a parallel node that monitors multiple sensors
monitorSensors := bt.NewParallel(bt.RequireAll,
    // Monitor temperature
    bt.NewAction(func(ctx bt.BehaviorContext) bt.RunStatus {
        // Simulate reading a temperature sensor
        temp, ok := ctx.Get("temperature")
        if !ok {
            fmt.Println("Temperature sensor unavailable")
            return bt.Failure
        }

        if temp.(float64) > 30.0 {
            fmt.Println("Temperature too high:", temp)
            ctx.Set("temp_alert", true)
            return bt.Failure
        }

        fmt.Println("Temperature normal:", temp)
        return bt.Success
    }),

    // Monitor humidity
    bt.NewAction(func(ctx bt.BehaviorContext) bt.RunStatus {
        // Simulate reading a humidity sensor
        humidity, ok := ctx.Get("humidity")
        if !ok {
            fmt.Println("Humidity sensor unavailable")
            return bt.Failure
        }

        if humidity.(float64) > 80.0 {
            fmt.Println("Humidity too high:", humidity)
            ctx.Set("humidity_alert", true)
            return bt.Failure
        }

        fmt.Println("Humidity normal:", humidity)
        return bt.Success
    }),
)

// Set up context and execute
ctx := bt.NewBehaviorContext(context.Background())
ctx.Set("temperature", 25.0)
ctx.Set("humidity", 60.0)

// Both sensors report normal values
status := monitorSensors.Tick(ctx)
fmt.Println("Sensor monitoring status:", status)
```

Available policies:

- `RequireOne`: Succeeds if at least one child succeeds, fails if all fail
- `RequireAll`: Succeeds if all children succeed, fails if any fail
- `SuccessOnOne`: Succeeds if at least one child succeeds, returns Running otherwise
- `SuccessOnAll`: Succeeds if all children succeed, returns Running otherwise

### Conditional

`Conditional` executes an action only if a condition is met.

```go
// Create a conditional node that heals if health is low
healIfNeeded := bt.NewConditional(
    // Condition: Check if health is below threshold
    bt.NewCondition(func(ctx bt.BehaviorContext) bool {
        health, ok := ctx.Get("health")
        if !ok {
            fmt.Println("No health data available")
            return false
        }

        threshold := 50
        currentHealth := health.(int)
        needsHealing := currentHealth < threshold

        if needsHealing {
            fmt.Printf("Health low (%d/%d)\n", currentHealth, 100)
        } else {
            fmt.Printf("Health sufficient (%d/%d)\n", currentHealth, 100)
        }

        return needsHealing
    }),

    // Action: Use healing potion if condition is true
    bt.NewAction(func(ctx bt.BehaviorContext) bt.RunStatus {
        // Check if we have any potions left
        potions, ok := ctx.Get("healing_potions")
        if !ok || potions.(int) <= 0 {
            fmt.Println("No healing potions available")
            return bt.Failure
        }

        // Use a potion to restore health
        ctx.Set("healing_potions", potions.(int)-1)
        ctx.Set("health", 100)
        fmt.Println("Used healing potion! Health restored to full.")
        return bt.Success
    }),
)

// Set up context and execute
ctx := bt.NewBehaviorContext(context.Background())
ctx.Set("health", 30)
ctx.Set("healing_potions", 2)

// Health is low, should use a potion
status := healIfNeeded.Tick(ctx)
fmt.Println("Healing status:", status)
```

### BinarySelector

`BinarySelector` selects between two behaviors based on a condition.

```go
// Create a binary selector for combat strategy
combatStrategy := bt.NewBinarySelector(
    // Condition: Check if enemy is close
    bt.NewCondition(func(ctx bt.BehaviorContext) bool {
        distance, ok := ctx.Get("enemy_distance")
        if !ok {
            return false
        }

        closeThreshold := 5.0
        isClose := distance.(float64) < closeThreshold

        if isClose {
            fmt.Printf("Enemy is close (%.1f units)\n", distance.(float64))
        } else {
            fmt.Printf("Enemy is far (%.1f units)\n", distance.(float64))
        }

        return isClose
    }),

    // If true: Use melee attack
    bt.NewAction(func(ctx bt.BehaviorContext) bt.RunStatus {
        fmt.Println("Performing melee attack!")

        // Apply melee damage
        ctx.Set("attack_type", "melee")
        ctx.Set("damage_dealt", 15)

        return bt.Success
    }),

    // If false: Use ranged attack
    bt.NewAction(func(ctx bt.BehaviorContext) bt.RunStatus {
        fmt.Println("Firing ranged weapon!")

        // Apply ranged damage
        ctx.Set("attack_type", "ranged")
        ctx.Set("damage_dealt", 10)

        return bt.Success
    }),
)

// Set up context and execute
ctx := bt.NewBehaviorContext(context.Background())
ctx.Set("enemy_distance", 8.0)

// Enemy is far, should use ranged attack
status := combatStrategy.Tick(ctx)
fmt.Println("Combat status:", status)
```

### PrioritySelector

`PrioritySelector` selects the first child that succeeds based on priority order.

```go
// Create a priority selector for AI behavior
aiBehavior := bt.NewPrioritySelector(
    // Highest priority: Run away if low health
    bt.NewSequence(
        // Check if health is low
        bt.NewCondition(func(ctx bt.BehaviorContext) bool {
            health, ok := ctx.Get("health")
            if !ok {
                return false
            }
            return health.(int) < 20
        }),
        // Run away action
        bt.NewAction(func(ctx bt.BehaviorContext) bt.RunStatus {
            fmt.Println("Health critical! Running away to safety!")
            ctx.Set("ai_state", "fleeing")
            return bt.Success
        }),
    ),

    // Medium priority: Attack if enemy is visible
    bt.NewSequence(
        // Check if enemy is visible
        bt.NewCondition(func(ctx bt.BehaviorContext) bool {
            enemyVisible, ok := ctx.Get("enemy_visible")
            if !ok {
                return false
            }
            return enemyVisible.(bool)
        }),
        // Attack action
        bt.NewAction(func(ctx bt.BehaviorContext) bt.RunStatus {
            fmt.Println("Enemy spotted! Attacking!")
            ctx.Set("ai_state", "attacking")
            return bt.Success
        }),
    ),

    // Lowest priority: Patrol the area
    bt.NewAction(func(ctx bt.BehaviorContext) bt.RunStatus {
        fmt.Println("No enemies in sight. Patrolling the area...")
        ctx.Set("ai_state", "patrolling")
        return bt.Success
    }),
)

// Set up context and execute
ctx := bt.NewBehaviorContext(context.Background())
ctx.Set("health", 50)
ctx.Set("enemy_visible", false)

// Health is ok and no enemy visible, should patrol
status := aiBehavior.Tick(ctx)
fmt.Println("AI status:", status)
```

### Switch

`Switch` selects a behavior based on a dynamic key value.

```go
// Create a switch for different character states
characterState := bt.NewSwitch(
    // Key function: Get the current animation state
    func(ctx bt.BehaviorContext) string {
        state, ok := ctx.Get("state")
        if !ok {
            return "unknown"
        }
        return state.(string)
    },

    // Define behaviors for each state
    map[string]bt.Behavior{
        "idle": bt.NewAction(func(ctx bt.BehaviorContext) bt.RunStatus {
            fmt.Println("Playing idle animation...")
            // Execute idle behavior
            ctx.Set("stamina_recovery", true)
            return bt.Success
        }),

        "walking": bt.NewAction(func(ctx bt.BehaviorContext) bt.RunStatus {
            fmt.Println("Playing walking animation...")
            // Execute walking behavior
            ctx.Set("position_changing", true)
            return bt.Success
        }),

        "running": bt.NewAction(func(ctx bt.BehaviorContext) bt.RunStatus {
            fmt.Println("Playing running animation...")
            // Execute running behavior
            ctx.Set("position_changing", true)
            ctx.Set("stamina_draining", true)
            return bt.Success
        }),
    },

    // Default behavior for unknown states
    bt.NewAction(func(ctx bt.BehaviorContext) bt.RunStatus {
        fmt.Println("Unknown state, playing default animation...")
        // Handle unknown state
        ctx.Set("state", "idle")  // Reset to idle
        return bt.Success
    }),
)

// Set up context and execute
ctx := bt.NewBehaviorContext(context.Background())
ctx.Set("state", "running")

// Should play running animation
status := characterState.Tick(ctx)
fmt.Println("Animation status:", status)
```

### TreeRunner

`TreeRunner` executes a behavior tree at a specified tick rate.

```go
// Create a simple behavior tree for an NPC
npcBehavior := bt.NewSelector(
    // Check if player is nearby
    bt.NewSequence(
        bt.NewCondition(func(ctx bt.BehaviorContext) bool {
            playerDistance, ok := ctx.Get("player_distance")
            return ok && playerDistance.(float64) < 10.0
        }),
        bt.NewAction(func(ctx bt.BehaviorContext) bt.RunStatus {
            fmt.Println("Player detected! Greeting player...")
            return bt.Success
        }),
    ),

    // Default: Wander around
    bt.NewAction(func(ctx bt.BehaviorContext) bt.RunStatus {
        fmt.Println("Wandering around the area...")
        return bt.Success
    }),
)

// Create a tree runner with the behavior tree
// The tree will be ticked at regular intervals
runner := bt.NewTreeRunner(npcBehavior)

// Create a context with a cancelable parent context
parentCtx, cancel := context.WithCancel(context.Background())
ctx := bt.NewBehaviorContext(parentCtx)
ctx.Set("player_distance", 20.0)

// Run the tree in a goroutine
go func() {
    // This will run until the context is canceled
    runner.Run(ctx)
    fmt.Println("NPC behavior tree stopped")
}()

fmt.Println("NPC behavior tree started")

// Simulate the player moving closer
time.Sleep(1 * time.Second)
ctx.Set("player_distance", 5.0)
fmt.Println("Player moved closer to NPC")

// Let it run for a bit
time.Sleep(2 * time.Second)

// Stop the tree
fmt.Println("Stopping NPC behavior tree...")
cancel()

// Wait for the goroutine to finish
time.Sleep(500 * time.Millisecond)
```

## Best Practices

### Tree Design

1. **Keep nodes focused**: Each node should have a single responsibility.
2. **Use appropriate composite nodes**: Choose the right composite for your logic flow.
3. **Avoid deep nesting**: Deep trees can be hard to understand and debug.
4. **Make actions reusable**: Design action nodes to be reusable across different trees.

### Performance

1. **Limit context data**: Store only necessary data in the behavior context.
2. **Use parallel nodes wisely**: Parallel execution can improve performance but adds complexity.
3. **Avoid expensive operations in conditions**: Conditions are checked frequently, so keep them lightweight.

### Debugging

1. **Log node status**: Consider wrapping important nodes with logging decorators.
2. **Visualize your trees**: Use the visualization tools to debug complex behaviors.
3. **Test subtrees independently**: Test smaller components before integrating them.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
