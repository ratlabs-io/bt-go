# Typed blackboard keys are additive

Blackboard keys remain strings. Optional `Key[T]` plus `SetKey`/`GetKey`/`MustGetKey` document the expected value type at compile time without a second store or breaking `Env.Set`/`Get`/`GetAs`. Rejected: replacing string keys, or a parallel typed map.
