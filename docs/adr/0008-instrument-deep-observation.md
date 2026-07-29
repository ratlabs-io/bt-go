# Deep observation via Instrument, not Env

Per-node **Observing** stays the light decorator. Tree-wide status/logging uses **Instrument** / **InstrumentRecorder**, which rebuild a parallel tree of Observing wrappers without mutating the original and without adding an observer channel to **Env** (Env stays cancel + blackboard only). Unknown third-party node types are wrapped as leaves (not deep-cloned).
