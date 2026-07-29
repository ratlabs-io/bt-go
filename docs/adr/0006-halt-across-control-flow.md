# Halt propagates through all abandoning control flow

Any composite that can leave a child Running without a further Tick must implement **Haltable** and Halt that child: Sequence, Selector, MemorySequence, MemorySelector, Parallel, BinarySelector, Switch, Conditional, and decorators.

- **Sequence** only Halts a previously Running later sibling when an earlier sibling fails or becomes Running; children that already returned Success this tick are not Halted.
- **Memory\*** idle resume index is **-1** so idle Halt is a true no-op.
- **Parallel** Halts residual Running children when the policy returns Success or Failure this tick (e.g. RequireOne with one Success).
- **Reset** never Halts; callers abort mid-run with Halt explicitly.

Rationale: **AbortHook** is only useful if every control-flow path that abandons work reaches it. Incomplete Halt was the main architecture gap after v1.5.0.
