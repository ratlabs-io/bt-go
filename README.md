# bt-go

[![Go Reference](https://pkg.go.dev/badge/github.com/ratlabs-io/bt-go.svg)](https://pkg.go.dev/github.com/ratlabs-io/bt-go)

A small, composable **behavior tree** library for Go.

## Install

```bash
go get github.com/ratlabs-io/bt-go
```

Import the package:

```go
import "github.com/ratlabs-io/bt-go/bt"
```

## Core model

| Concept | Role |
|--------|------|
| **Behavior** | Any node: `Tick(env) → Success \| Failure \| Running` |
| **Env** | Per-tick environment (not a `context.Context`) |
| **Blackboard** | Hierarchical, mutex-protected key-value store for agent state |
| **context.Context** | Cancellation / deadlines only — via `env.Context()` |
| **Composite** | Multi-child control flow (`Sequence`, `Selector`, `Parallel`, …) |
| **Decorator** | Single-child wrapper (`Inverter`, `Repeater`, …) |
| **TreeRunner** | Ticks a root on a fixed interval until cancelled |

### Status

- `Success` — work finished successfully  
- `Failure` — work failed  
- `Running` — still in progress; tick again later  

### Env vs context.Context

`Env` is **has-a**, not **is-a**, relative to stdlib context:

| Concern | Where it lives |
|---------|----------------|
| Cancel, deadline, timeout | `env.Context()` → `context.Context` |
| Mutable agent/world state | `env.Blackboard()` (or `Set` / `Get` / `Delete` / `Has`) |

Do **not** put health, targets, inventory, or AI flags in `context.WithValue`. That is the classic Go anti-pattern; the blackboard is the right place for BT working memory.

```go
parent, cancel := context.WithCancel(context.Background())
defer cancel()

env := bt.NewEnv(parent)
env.Set("health", 100)

bb := env.Blackboard()
// bb.Get("health") == 100 — same store as env.Set/Get

// Long-running actions may watch:
//   select { case <-env.Context().Done(): ... }
```

`Env` deliberately does **not** implement `context.Context`.

### Reactive vs memory composites

Default `Sequence` and `Selector` are **reactive**: every `Tick` starts at the **first** child. They do not remember which child was `Running`.

| Constructor | Style | Resume behavior |
|-------------|--------|-----------------|
| `NewSequence` | reactive | Always restart at child 0 |
| `NewMemorySequence` | memory | Resume at last `Running` child; skip already-succeeded prefix |
| `NewSelector` | reactive | Always restart at child 0 (higher priority can preempt) |
| `NewMemorySelector` | memory | Stick with last `Running` child until it finishes (no preemption) |
| `NewPrioritySelector` | reactive | Alias of `NewSelector` |

Memory variants expose `Reset()` and `RunningIndex()`. Memory clears automatically on terminal `Success` / `Failure`.

Use **reactive** when priorities must be re-checked every tick (e.g. “abort patrol if under attack”). Use **memory** when a multi-step branch should finish without re-running expensive earlier steps.

### Parallel

`Parallel` ticks **every** child on **every** call, each in its own goroutine, then aggregates with a policy:

| Policy | Success | Failure | Otherwise |
|--------|---------|---------|-----------|
| `RequireOne` | ≥1 success | all fail | `Running` |
| `RequireAll` | all success | any fail | `Running` |
| `SuccessOnOne` | ≥1 success | — | `Running` |
| `SuccessOnAll` | all success | — | `Running` |

Children share the **same** `Env`. The blackboard is concurrent-safe; other shared mutable state in actions must be synchronized by the caller.

## Quick start

```go
package main

import (
	"context"
	"fmt"

	"github.com/ratlabs-io/bt-go/bt"
)

func main() {
	tree := bt.NewSequence(
		bt.NewAction(func(env bt.Env) bt.RunStatus {
			hello, ok := env.Get("greeting")
			if !ok {
				return bt.Failure
			}
			fmt.Printf("%s ", hello.(string))
			return bt.Success
		}),
		bt.NewAction(func(env bt.Env) bt.RunStatus {
			world, ok := env.Get("subject")
			if !ok {
				return bt.Failure
			}
			fmt.Printf("%s\n", world.(string))
			return bt.Success
		}),
	)

	env := bt.NewEnv(context.Background())
	env.Set("greeting", "Hello")
	env.Set("subject", "World")

	fmt.Println(tree.Tick(env)) // Success
}
```

Runnable copy: [`examples/hello`](./examples/hello).

## Node catalog

### Leaves

| Constructor | Behavior |
|-------------|----------|
| `NewAction(fn)` | Runs `fn`; returns its `RunStatus` |
| `NewCondition(fn)` | `true` → `Success`, `false` → `Failure` |

### Composites

| Constructor | Behavior |
|-------------|----------|
| `NewSequence(children...)` | All succeed, in order (reactive) |
| `NewMemorySequence(children...)` | Sequence with resume memory |
| `NewSelector(children...)` | First non-failure wins (reactive priority) |
| `NewMemorySelector(children...)` | Selector that sticks to Running child |
| `NewPrioritySelector(...)` | Alias of `NewSelector` |
| `NewParallel(policy, children...)` | Concurrent tick + policy |
| `NewBinarySelector(cond, ifTrue, ifFalse)` | Branch on condition success |
| `NewSwitch(keyFn, cases, default)` | Branch on string key |
| `NewConditional(cond, action)` | Run action only if condition succeeds |

### Decorators

| Constructor | Behavior |
|-------------|----------|
| `NewInverter(child)` | Swap Success ↔ Failure |
| `NewRepeater(child, n)` | Succeed after `n` successful child ticks (`n≤0` = forever) |
| `NewUntilSuccess(child)` | Keep ticking until child succeeds |
| `NewUntilFailure(child)` | Keep ticking until child fails (decorator then returns Success) |

### Runner

```go
runner := bt.NewTreeRunner(tree,
	bt.WithTickRate(50*time.Millisecond),
	bt.WithCallbacks(onSuccess, onFailure, onRunning),
)

go runner.Run(env)   // until env.Context() is cancelled
status := runner.RunOnce(env)
```

### Visualization

```go
fmt.Print(bt.NewTreeVisualizer(tree).Visualize())

rec := bt.NewStatusRecorder()
rec.Tick(env, someNode)
fmt.Print(rec.Visualize(tree))
```

Custom nodes can implement `NodeVisualizer` (`VisualizeNode() string`). Composites that implement `ChildrenProvider` (`GetChildren() []Behavior`) and decorators that implement `Decorator` render without special-casing in the visualizer.

## Architecture notes

1. **Env ≠ context.Context** — lifecycle and agent state are separate types and APIs.
2. **Single data plane** — `Set`/`Get` write through to the blackboard; no parallel maps.
3. **No private type assertions** — runners and parallel use only the `Env` interface.
4. **Reactive by default** — memory is opt-in (`NewMemorySequence` / `NewMemorySelector`).
5. **One reactive selector** — `PrioritySelector` is an alias for `Selector`.
6. **Parallel always re-ticks** — every child every tick; policies like `SuccessOnAll` can recover.

## Best practices

- Keep leaves small; compose with Sequence/Selector rather than giant actions.
- Store agent state on the blackboard; never in `context.WithValue`.
- Conditions should be cheap — they may run every tick on reactive parents.
- Under `Parallel`, only rely on the blackboard (or your own locks) for shared data.
- Blocking actions should honor `env.Context().Done()` if they might run long.

## Development

```bash
go test ./...
go test ./... -race
go run ./examples/hello
```

## License

MIT — see [LICENSE](LICENSE).
