# Reactive composites by default; memory is opt-in

Default **Sequence** and **Selector** restart from the first child every Tick (reactive / priority-recheck). Resume-from-Running lives only on **MemorySequence** / **MemorySelector**. Half-implemented “running index” fields that are written but never read for control flow are forbidden — memory is either real or absent.
