# Phase 3 — Dynamic Payroll Engine

The engine is independent from HTTP and MySQL. HTTP validates transport data, the service resolves date-effective rules, and the engine executes generic calculation methods.

Supported dynamic methods: FIXED_AMOUNT, PERCENTAGE, PERCENTAGE_WITH_MINIMUM, CAPPED_PERCENTAGE, PROGRESSIVE_BANDS and TIERED_FIXED_AMOUNT.

No Kenya statutory rates are hardcoded into Go. The calculation result records rule versions used for auditability.
