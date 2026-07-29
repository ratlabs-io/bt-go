# bt-go

Composable behavior-tree library for Go agents (games AI, robotics-style control, scripts). Latest release at time of writing: **v1.6.0**. Assumed **zero external consumers** — prefer clean breaks over compatibility shims.

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
Hierarchical, mutex-protected key-value store for mutable agent/world state. Sole data plane for `Set`/`Get`/`GetAs` / typed **Key[T]**.
_Avoid_: parallel maps, dual stores, request-scoped context values for domain state

**Key[T]**:
Optional typed blackboard key (`type Key[T any] string`) used with `SetKey`/`GetKey`/`MustGetKey`. Same store as string keys; type param documents value type only.
_Avoid_: a second store or replacing string keys

**Composite**:
Multi-child control node (Sequence, Selector, Parallel, …).

**Decorator**:
Single-child wrapper (Inverter, Repeater, Named, Observing, AbortHook, …).

**Reactive composite**:
Restarts evaluation from the first child every Tick (`NewSequence`, `NewSelector`). Higher-priority Selector children can preempt.

**Memory composite**:
Resumes the last Running child (`NewMemorySequence`, `NewMemorySelector`). No preemption while stuck on a child. Idle resume index is **-1**.

**Halt**:
Abort signal when a parent abandons a Running child (preemption, branch switch, policy completion, cancel) without a terminal Success/Failure tick on that child. Implemented via **Haltable**.
_Avoid_: treating Reset as Halt (Reset does not Halt; call Halt when aborting mid-run)

**AbortHook**:
Decorator that runs a user callback when Halted while active (last status was Running). Primary cleanup seam for preempted branches.

**Parallel (sequential)**:
Default `NewParallel`: tick every child this Tick, same stack, deterministic order. Classic BT “parallel.” Residual Running children are Halted when the policy returns Success or Failure.

**Parallel (concurrent)**:
`NewConcurrentParallel`: one goroutine per child; shared Env/blackboard must stay race-safe. Same residual-Halt rule after join.

**Observing**:
Per-node tick decorator. Reports only that node’s child result.

**Instrument**:
Deep observation: rebuilds a tree so every node tick reports to a callback (or **StatusRecorder** via `InstrumentRecorder`). Does not mutate the original tree.

**TreeRunner**:
Schedules ticks until `env.Context()` is done; Halts the tree on cancel. Does not preempt an in-flight Tick body.

## Relationships

- A **Behavior** tree is ticked with one **Env** per agent/run
- **Env** owns one **Blackboard** and wraps one `context.Context`
- **Reactive** and **Memory** composites are different control-flow products, not aliases
- **Halt** propagates through every control-flow node that can abandon Running work (Sequence, Selector, Memory*, Parallel, BinarySelector, Switch, Conditional, decorators)
- **Named** / **Observing** / **AbortHook** are **Decorators** (pass-through or thin wrappers)
- **Instrument** is an opt-in tree-wide observation adapter; **Observing** remains the per-node form

## Settled decisions (do not re-litigate without cause)

1. **Env ≠ context.Context** — has-a, not is-a. See `docs/adr/0001-env-not-context.md`.
2. **No PrioritySelector** — it was identical to Selector; use `NewSelector`.
3. **Single data plane** — blackboard only; no ContextData map.
4. **v2 module path** only if we want hard major coexistence; with zero users, breaking v1.x tags are fine.
5. **Import path** is module root: `github.com/ratlabs-io/bt-go` (package name `bt`). Stay flat until non-library root noise or clear subdomains appear (`docs/adr/0005-module-root-package.md`).
6. **Complete Halt graph** — all abandoning control-flow nodes track Running children and Halt them; Parallel Halts residual Running on terminal policy results. See `docs/adr/0006-halt-across-control-flow.md`.
7. **Typed keys are additive** — `Key[T]` + helpers; string keys remain first-class. See `docs/adr/0007-typed-keys-additive.md`.
8. **Deep observation is Instrument** — not a second Env channel; per-node Observing stays. See `docs/adr/0008-instrument-deep-observation.md`.

## Example dialogue

> **Dev:** “Should agent health live on the context?”  
> **Domain:** “On the **Blackboard** via **Env**. `env.Context()` is only for cancel/timeout. Prefer `Key[int]` or `GetAs`.”  
>
> **Dev:** “Selector vs MemorySelector when flee must interrupt attack?”  
> **Domain:** “**Reactive** **Selector** so flee is re-checked every **Tick**. Wrap attack in **AbortHook** so **Halt** clears wind-up.”  
>
> **Dev:** “Is Parallel multi-threaded?”  
> **Domain:** “Only **NewConcurrentParallel**. Default **Parallel** means all children tick this frame, sequentially.”  
>
> **Dev:** “How do I log every node’s status?”  
> **Domain:** “**Instrument** the tree (or **InstrumentRecorder**). Use **Observing** only for a few hotspots.”

## Flagged ambiguities

- **“Context”** — always disambiguate: stdlib `context.Context` vs **Env** (formerly BehaviorContext).
- **“Parallel”** — sequential-by-default in this library; concurrent is opt-in by name.
- **“Priority”** — means child order on **Selector**, not a separate type.
- **“Observe”** — per-node **Observing** vs tree-wide **Instrument**.

## Open / next evaluation targets

Closed in **v1.6.0**: complete Halt graph, typed keys, Instrument, godoc examples, flat package reconfirm.

Worth a fresh pass (not decided as wrong—just next scrutiny):

- Whether **Instrument** should deep-support custom/third-party node types beyond wrap-as-is
- Subtree blackboard scopes as a first-class API (hierarchy exists; patterns are user-side)
- Concurrent Parallel cancel of in-flight child Ticks (today: join then Halt residual only)
- Further composites/decorators (decorators catalog, time/cooldown nodes, subtree include) if product needs them
- Whether TreeRunner should expose tick middleware / shared observation without Instrument rebuild
