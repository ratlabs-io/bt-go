# bt-go — agent notes

Go behavior-tree library. Module: `github.com/ratlabs-io/bt-go`, package `bt`, latest **v1.6.1**.

## Before changing architecture

Read **CONTEXT.md** (domain language + settled decisions) and **docs/adr/** (0001–0008). Do not re-suggest:

- PrioritySelector, dual data stores, Env implementing `context.Context`
- Splitting the root package without crowding
- Dual blackboard / replacing string keys (typed `Key[T]` is additive only)
- Observer channel on Env (use Observing / Instrument)
- Incomplete Halt — every control-flow abandon path must reach `Haltable`

## Commands

```bash
go test ./...
go test ./... -race
go run ./examples/hello
go run ./examples/agent
```

## Layout

- Library sources at **repo root** (`package bt`), not `bt/` subfolder
- Examples under `examples/`
- Project memory: `CONTEXT.md`, `docs/adr/`, this file
- Public story: README + CHANGELOG

## Patterns

- Tick signature: `Tick(env Env) RunStatus`
- Agent state: blackboard / `GetAs[T]` / `Key[T]` — never `context.WithValue`
- Cancel: `env.Context().Done()`; long actions must poll it themselves
- Preemption cleanup: `Haltable` + `AbortHook` (works under Sequence, Selector, Memory*, Parallel, BinarySelector, Switch, Conditional, decorators)
- Parallel: sequential default; concurrent = `NewConcurrentParallel`; residual Running Halted on terminal policy result
- Reactive vs memory: explicit constructors (`NewSequence` vs `NewMemorySequence`); memory idle index **-1**
- Observation: per-node `NewObserving`; tree-wide `Instrument` / `InstrumentRecorder` (does not mutate original)
- Sequence Halt: only abandon later Running siblings (`lastRunning > i`); never false-Halt after Success progress
- Cancel: cooperative only — long Actions poll `env.Context().Done()`; no mid-Tick kill

## Gotchas

- `AbortHook` only fires `OnAbort` when last Tick was Running — masks some false Halts; test with real Haltable state, not only hooks
- `Instrument` rebuilds known types; custom third-party nodes wrap-as-is (not deep-cloned)
- Concurrent Parallel does not cancel in-flight child Tick bodies (join, then residual Halt)
- `Reset` ≠ `Halt` — always call Halt when aborting mid-run
- `TreeVisualizer` collapses `Observing` (Instrument dumps show real nodes + statuses)

## Do not expand without product pain

See CONTEXT settled #10 and “Deferred” open targets. Skip: runner middleware, scoped-blackboard decorator, catalog Timeout/Cooldown, Instrument third-party deep-clone, hard Tick interrupt.

## Release

Semver tags with `v` prefix (`v1.6.1`). No users assumed — breaking changes OK with a tag bump and CHANGELOG entry. Go major ≥2 needs `module .../v2` if ever used.

Architecture cadence: use the library in real agents; reopen CONTEXT deferred rows only when the same workaround appears twice. Do not grow surface for BT-checklist parity.
