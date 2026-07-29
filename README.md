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
| **Behavior** | Any node: `Tick(ctx) → Success \| Failure \| Running` |
| **BehaviorContext** | Per-run environment: cancellation + blackboard data |
| **Blackboard** | Hierarchical, mutex-protected key-value store |
| **Composite** | Multi-child control flow (`Sequence`, `Selector`, `Parallel`, …) |
| **Decorator** | Single-child wrapper (`Inverter`, `Repeater`, …) |
| **TreeRunner** | Ticks a root on a fixed interval until cancelled |

### Status

- `Success` — work finished successfully  
- `Failure` — work failed  
- `Running` — still in progress; tick again later  

### Data and cancellation

There is **one** data store: the **Blackboard**.

`BehaviorContext.Set` / `Get` / `Delete` / `Has` all operate on that blackboard (optionally hierarchical via `NewBlackboardWithParent`). `Context()` exposes the underlying `context.Context` for cancellation and deadlines — used by `TreeRunner.Run`.

```go
parent, cancel := context.WithCancel(context.Background())
defer cancel()

ctx := bt.NewBehaviorContext(parent)
ctx.Set("health", 100)

bb := ctx.GetBlackboard()
// bb.Get("health") == 100 — same store
```

### Reactive vs memory composites

Default `Sequence` and `Selector` are **reactive**: every `Tick` starts at the **first** child. They do not remember which child was `Running`.

| Constructor | Style | Resume behavior |
|-------------|--------|-----------------|
| `NewSequence` | reactive | Always restart at child 0 |
| `NewMemorySequence` | memory | Resume at last `Running` child; skip already-succeeded prefix |
| `NewSelector` | reactive | Always restart at child 0 (higher priority can preempt) |
| `NewMemorySelector` | memory | Stick with last `Running` child until it finishes (no preemption) |
| `NewPrioritySelector` | reactive | Alias of `NewSelector` |

Memory variants expose `Reset()` and `RunningIndex()` for tests and tooling. Memory clears automatically on terminal `Success` / `Failure`.

Use **reactive** when priorities must be re-checked every tick (e.g. “abort patrol if under attack”). Use **memory** when a multi-step branch should finish without re-running expensive earlier steps.

### Parallel

`Parallel` ticks **every** child on **every** call, each in its own goroutine, then aggregates with a policy:

| Policy | Success | Failure | Otherwise |
|--------|---------|---------|-----------|
| `RequireOne` | ≥1 success | all fail | `Running` |
| `RequireAll` | all success | any fail | `Running` |
| `SuccessOnOne` | ≥1 success | — | `Running` |
| `SuccessOnAll` | all success | — | `Running` |

Children share the **same** `BehaviorContext`. The blackboard is concurrent-safe; other shared mutable state in actions must be synchronized by the caller.

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
		bt.NewAction(func(ctx bt.BehaviorContext) bt.RunStatus {
			hello, ok := ctx.Get("greeting")
			if !ok {
				return bt.Failure
			}
			fmt.Printf("%s ", hello.(string))
			return bt.Success
		}),
		bt.NewAction(func(ctx bt.BehaviorContext) bt.RunStatus {
			world, ok := ctx.Get("subject")
			if !ok {
				return bt.Failure
			}
			fmt.Printf("%s\n", world.(string))
			return bt.Success
		}),
	)

	ctx := bt.NewBehaviorContext(context.Background())
	ctx.Set("greeting", "Hello")
	ctx.Set("subject", "World")

	fmt.Println(tree.Tick(ctx)) // Success
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

go runner.Run(ctx)   // until ctx.Context() is cancelled
status := runner.RunOnce(ctx)
```

### Visualization

```go
fmt.Print(bt.NewTreeVisualizer(tree).Visualize())

rec := bt.NewStatusRecorder()
rec.Tick(ctx, someNode)
fmt.Print(rec.Visualize(tree))
```

Custom nodes can implement `NodeVisualizer` (`VisualizeNode() string`). Composites that implement `ChildrenProvider` (`GetChildren() []Behavior`) and decorators that implement `Decorator` render without special-casing in the visualizer.

## Architecture notes

These design choices keep the surface small and avoid common BT-library traps:

1. **Single data plane** — context KV API is the blackboard; no parallel maps that can diverge (especially under `Parallel`).
2. **No private type assertions** — `TreeRunner` and `Parallel` use only the `BehaviorContext` interface (`Context()`, shared blackboard).
3. **Reactive by default** — `Sequence` / `Selector` restart from the first child; memory is an explicit opt-in (`NewMemorySequence` / `NewMemorySelector`), not a half-written field on the reactive types.
4. **One reactive selector** — `PrioritySelector` is not a second algorithm; it is `Selector`.
5. **Parallel always re-ticks** — sticky per-child completion memory made policies like `SuccessOnAll` unable to recover after a transient failure.

## Best practices

- Keep leaves small; compose with Sequence/Selector rather than giant actions.
- Store agent state on the blackboard; avoid hidden globals.
- Conditions should be cheap — they may run every tick on reactive parents.
- Under `Parallel`, only rely on the blackboard (or your own locks) for shared data.
- Use `TreeVisualizer` when debugging structure; use `StatusRecorder` for light status capture.

## Development

```bash
go test ./...
go test ./... -race
go run ./examples/hello
```

## License

MIT — see [LICENSE](LICENSE).
