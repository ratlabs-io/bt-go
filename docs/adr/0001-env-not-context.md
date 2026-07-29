# Env is not a context.Context

Agent state and cancel/deadline have different lifetimes and mutability. **Env** holds a `context.Context` (has-a) for cancellation only and puts mutable agent/world state on a **Blackboard**. Env does not implement `context.Context`, so APIs cannot treat the tick environment as a request-scoped value bag (`WithValue`). Callers use `env.Context()` when they need stdlib cancel/timeout.
