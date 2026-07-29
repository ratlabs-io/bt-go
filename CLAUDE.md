# bt-go — agent notes

Go behavior-tree library. Module: `github.com/ratlabs-io/bt-go`, package `bt`, latest **v1.6.0**.

## Before changing architecture

Read **CONTEXT.md** (domain language + settled decisions) and **docs/adr/**. Do not re-suggest PrioritySelector, dual data stores, implementing `context.Context` on Env, splitting the root package without crowding, dual blackboard stores for typed keys, or observer channels on Env without a strong new reason.

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
- Public story: README + CHANGELOG

## Patterns

- Tick signature: `Tick(env Env) RunStatus`
- Agent state: blackboard / `GetAs[T]` — never `context.WithValue`
- Cancel: `env.Context().Done()`; long actions must poll it themselves
- Preemption cleanup: `Haltable` + `AbortHook`
- Parallel: sequential default; concurrent = `NewConcurrentParallel`
- Reactive vs memory: explicit constructors (`NewSequence` vs `NewMemorySequence`)

## Release

Semver tags with `v` prefix (`v1.6.0`). No users assumed — breaking changes OK with a tag bump and CHANGELOG entry. Go major ≥2 needs `module .../v2` if ever used.
