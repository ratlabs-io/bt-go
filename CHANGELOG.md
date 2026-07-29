# Changelog

## [v1.6.1] — 2026-07-29

### Fixed
- `TreeVisualizer` collapses `Observing` wrappers so `InstrumentRecorder` dumps show real node statuses (Named/composites/leaves) without Observing noise.

### Docs
- CONTEXT/CLAUDE: post-v1.6.0 architecture freeze; open targets deferred until real agent pain; cooperative cancel settled.

## [v1.6.0] — 2026-07-29

### Breaking
- `MemorySequence.RunningIndex()` is **-1** when idle (was `0`). `Reset` also idles to `-1`. Matches `MemorySelector`.

### Fixed
- Reactive `Sequence` no longer Halts a child that already returned Success this tick when a later sibling becomes Running.
- `BinarySelector`, `Switch`, and `Conditional` track Running branches and Halt on abandon / parent Halt.
- `Parallel` Halts residual Running children when the policy returns Success or Failure.
- `halt.go` docs: Reset does not Halt.

### Added
- `Instrument` / `InstrumentRecorder` — tree-wide observation without mutating the original tree.
- `Key[T]`, `SetKey`, `GetKey`, `MustGetKey` — typed keys on the same blackboard.
- Godoc `Example*` tests for Sequence, MemorySequence, AbortHook, GetAs, SetKey, Parallel, Instrument.
- Halt matrix tests (`halt_test.go`).
- ADRs 0006–0008; CONTEXT updated.

## [v1.5.0] — 2026-07-29

### Breaking
- Package lives at module root: import `github.com/ratlabs-io/bt-go` (was `.../bt-go/bt`).
- Removed `PrioritySelector` / `NewPrioritySelector` (identical to `Selector`).
- `NewParallel` is **sequential** by default (classic BT “tick all this frame”). Use `NewConcurrentParallel` for goroutines.

### Added
- `GetAs[T]` / `MustGet[T]` typed blackboard helpers.
- `Haltable` / `Halt` — abort cleanup; wired through Selector, Sequence, memory nodes, Parallel, decorators.
- `AbortHook` — user callback when a Running branch is preempted or cancelled.
- `Named` decorator for visualizer-friendly labels.
- `Observing` decorator for tick hooks.
- `Repeater.Reset`.
- `TreeRunner.Run` Halts the tree on context cancel.
- Example: `examples/agent` (reactive flee vs memory combat).
- This changelog.

### Changed
- README and package docs updated for Env, Halt, parallel modes, root import.

## [v1.4.0] — 2026-07-29

### Breaking
- Renamed `BehaviorContext` → `Env` (`NewEnv`). Not a `context.Context` (has-a cancel context + blackboard).

### Added
- `MemorySequence` / `MemorySelector` with `Reset` / `RunningIndex`.
- Unified blackboard as sole data plane; `Context()` for cancellation only.

## [v1.3.0] and earlier

See git tags `v0.0.0`–`v1.3.0` for prior history. Tag `1.0.0` (no `v` prefix) was removed as invalid for Go modules.
