# bt-go

[![Go Reference](https://pkg.go.dev/badge/github.com/ratlabs-io/bt-go.svg)](https://pkg.go.dev/github.com/ratlabs-io/bt-go)

A small, composable **behavior tree** library for Go.

## Install

```bash
go get github.com/ratlabs-io/bt-go@v1.6.0
```

```go
import "github.com/ratlabs-io/bt-go"
// package name is bt → bt.NewEnv, bt.NewSequence, …
```

## Core model

| Concept | Role |
|--------|------|
| **Behavior** | Any node: `Tick(env) → Success \| Failure \| Running` |
| **Env** | Per-tick environment (**not** a `context.Context`) |
| **Blackboard** | Hierarchical KV store for agent/world state |
| **context.Context** | Cancel/deadline only — `env.Context()` |
| **Halt** | Abort cleanup when a parent abandons Running work |

### Env vs context.Context

| Concern | API |
|---------|-----|
| Cancel / deadline | `env.Context()` |
| Agent state | `env.Blackboard()` or `Set`/`Get`/`GetAs[T]` / `Key[T]` |

Do **not** put health/targets in `context.WithValue`. Use the blackboard.

```go
env := bt.NewEnv(parentCtx)
env.Set("health", 100)
h, ok := bt.GetAs[int](env, "health")

const Health bt.Key[int] = "health"
bt.SetKey(env, Health, 100)
```

### Reactive vs memory

| Constructor | Style |
|-------------|--------|
| `NewSequence` / `NewSelector` | Reactive — restart at child 0 each tick |
| `NewMemorySequence` / `NewMemorySelector` | Resume last `Running` child |

`Selector` is priority order (earlier = higher). Preempted Running children are **Halted**.

### Parallel

| Constructor | Execution |
|-------------|-----------|
| `NewParallel` | Sequential ticks (default, deterministic) |
| `NewConcurrentParallel` | One goroutine per child |

Policies: `RequireOne`, `RequireAll`, `SuccessOnOne`, `SuccessOnAll`.  
When a policy returns Success/Failure while some children are still Running, those residual children are **Halted**.

### Halt and abort hooks

```go
branch := bt.NewAbortHook(longRunningSubtree, func(env bt.Env) {
    // stop pathing, clear target, …
})
// When a higher-priority Selector branch wins, Halt runs OnAbort.
// Same for Sequence earlier-sibling failure, BinarySelector/Switch flips,
// Conditional condition fail, Parallel residual Running, runner cancel.
```

`TreeRunner.Run` Halts the tree when `env.Context()` is cancelled. In-flight `Tick` bodies are not interrupted — long actions should watch `env.Context().Done()`.

### Observation

| API | Scope |
|-----|--------|
| `NewObserving` | One node |
| `Instrument` / `InstrumentRecorder` | Whole tree (rebuilds wrappers; original unchanged) |

## Quick start

```go
tree := bt.NewSequence(
    bt.NewNamed("Greet", bt.NewAction(func(env bt.Env) bt.RunStatus {
        fmt.Println(bt.MustGet[string](env, "greeting"))
        return bt.Success
    })),
)
env := bt.NewEnv(context.Background())
env.Set("greeting", "Hello")
tree.Tick(env)
```

Examples:

- [`examples/hello`](./examples/hello) — minimal Sequence  
- [`examples/agent`](./examples/agent) — flee / memory combat / patrol  
- Godoc `Example*` tests in the package (Sequence, Memory, AbortHook, Keys, Instrument, …)

## Node catalog

**Leaves:** `NewAction`, `NewCondition`  

**Composites:** `NewSequence`, `NewMemorySequence`, `NewSelector`, `NewMemorySelector`, `NewParallel`, `NewConcurrentParallel`, `NewBinarySelector`, `NewSwitch`, `NewConditional`  

**Decorators:** `NewInverter`, `NewRepeater`, `NewUntilSuccess`, `NewUntilFailure`, `NewNamed`, `NewObserving`, `NewAbortHook`  

**Runner:** `NewTreeRunner` + `WithTickRate` / `WithCallbacks`  

**Typed data:** `GetAs[T]`, `MustGet[T]`, `Key[T]`, `SetKey` / `GetKey` / `MustGetKey`  

**Debug:** `NewTreeVisualizer`, `NewStatusRecorder`, `NewObserving`, `Instrument`, `InstrumentRecorder`  

## Development

```bash
go test ./...
go test ./... -race
go run ./examples/hello
go run ./examples/agent
```

See [CHANGELOG.md](CHANGELOG.md) for release history.

## License

MIT — see [LICENSE](LICENSE).
