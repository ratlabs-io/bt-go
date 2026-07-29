# Parallel means sequential multi-child tick by default

In BT literature, “parallel” usually means all children get a tick this frame, not OS threads. **NewParallel** ticks children sequentially (deterministic). **NewConcurrentParallel** is the explicit goroutine variant. Sharing one Env under concurrent mode requires thread-safe blackboard use (already mutexed) and user care for other shared state.
