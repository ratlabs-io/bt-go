# bt-go

Composable behavior-tree library for Go agents (games AI, robotics-style control, scripts). Latest release at time of writing: **v1.5.0**. Assumed **zero external consumers** — prefer clean breaks over compatibility shims.

## Language

**Behavior**:
A tree node that can be ticked. Returns a **RunStatus**.
_Avoid_: task, step (unless speaking informally about leaves)

**RunStatus**:
One of Success, Failure, Running — the only result of a Tick.
_Avoid_: bool outcomes for composites; state enums beyond these three

**Tick**:
One execution step of a **Behavior** with an **Env**. Not a game frame by itself; a **TreeRunner** may tick on a timer.

**Env**:
Per-tick / per-agent environment passed to every node. **Has-a** `context.Context` for cancel/deadline; **is not** a `context.Context`. Agent data lives on the **Blackboard**.
_Avoid_: BehaviorContext (removed), “context” alone when meaning Env, stuffing agent state into `context.WithValue`

**Blackboard**:
Hierarchical, mutex-protected key-value store for mutable agent/world state. Sole data plane for `Set`/`Get`/`GetAs`.
_Avoid_: parallel maps, dual stores, request-scoped context values for domain state

**Composite**:
Multi-child control node (Sequence, Selector, Parallel, …).

**Decorator**:
Single-child wrapper (Inverter, Repeater, Named, Observing, AbortHook, …).

**Reactive composite**:
Restarts evaluation from the first child every Tick (`NewSequence`, `NewSelector`). Higher-priority Selector children can preempt.

**Memory composite**:
Resumes the last Running child (`NewMemorySequence`, `NewMemorySelector`). No preemption while stuck on a child.

**Halt**:
Abort signal when a parent abandons a Running child (preemption, cancel) without a terminal Success/Failure tick on that child. Implemented via **Haltable**.
_Avoid_: treating Reset as Halt (Reset does not Halt; call Halt when aborting mid-run)

**AbortHook**:
Decorator that runs a user callback when Halted while active (last status was Running). Primary cleanup seam for preempted branches.

**Parallel (sequential)**:
Default `NewParallel`: tick every child this Tick, same stack, deterministic order. Classic BT “parallel.”

**Parallel (concurrent)**:
`NewConcurrentParallel`: one goroutine per child; shared Env/blackboard must stay race-safe.

**TreeRunner**:
Schedules ticks until `env.Context()` is done; Halts the tree on cancel. Does not preempt an in-flight Tick body.

## Relationships

- A **Behavior** tree is ticked with one **Env** per agent/run
- **Env** owns one **Blackboard** and wraps one `context.Context`
- **Reactive** and **Memory** composites are different control-flow products, not aliases
- **Halt** propagates through composites/decorators that track Running children
- **Named** / **Observing** / **AbortHook** are **Decorators** (pass-through or thin wrappers)

## Settled decisions (do not re-litigate without cause)

1. **Env ≠ context.Context** — has-a, not is-a. See `docs/adr/0001-env-not-context.md`.
2. **No PrioritySelector** — it was identical to Selector; use `NewSelector`.
3. **Single data plane** — blackboard only; no ContextData map.
4. **v2 module path** only if we want hard major coexistence; with zero users, breaking v1.x tags are fine.
5. **Import path** is module root: `github.com/ratlabs-io/bt-go` (package name `bt`).

## Example dialogue

> **Dev:** “Should agent health live on the context?”  
> **Domain:** “On the **Blackboard** via **Env**. `env.Context()` is only for cancel/timeout.”  
>
> **Dev:** “Selector vs MemorySelector when flee must interrupt attack?”  
> **Domain:** “**Reactive** **Selector** so flee is re-checked every **Tick**. Wrap attack in **AbortHook** so **Halt** clears wind-up.”  
>
> **Dev:** “Is Parallel multi-threaded?”  
> **Domain:** “Only **NewConcurrentParallel**. Default **Parallel** means all children tick this frame, sequentially.”

## Flagged ambiguities

- **“Context”** — always disambiguate: stdlib `context.Context` vs **Env** (formerly BehaviorContext).
- **“Parallel”** — sequential-by-default in this library; concurrent is opt-in by name.
- **“Priority”** — means child order on **Selector**, not a separate type.

## Open / next evaluation targets

Worth a fresh pass (not decided as wrong—just next scrutiny):

- Whether **Halt** tracking on reactive Sequence is complete enough
- Stringly blackboard keys vs richer typed keys
- Deep observation (tree-wide) vs per-node **Observing**
- Whether module should stay flat root package as surface grows
- More examples / godoc examples on exported symbols
