# No PrioritySelector type

**Selector** already evaluates children in priority order (left-to-right, reactive). A separate **PrioritySelector** alias duplicated the API without different semantics and invited false distinctions. Removed in v1.5.0; use **NewSelector**.
