# Changelog

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
