# Phase 2 Architecture — Dynamic Historical Rules

The database is the source of truth for statutory values and effective dates.

```
HTTP Handler
    -> Rules Service (validation and business boundary)
        -> Rules Repository (SQL only)
            -> MySQL rule_definitions
            -> rule_versions
            -> calculation_methods
            -> rule_parameters
            -> rule_bands
            -> rule_dependencies
            -> rule_sources
```

## Historical resolution

A requested date selects only a rule version where:

- `status = PUBLISHED`
- `effective_from <= requested date`
- `effective_to IS NULL OR effective_to >= requested date`

The newest matching effective version wins. Published historical versions are
therefore immutable calculation snapshots for future engine work.

## Phase boundary

Phase 2 does not calculate PAYE yet. It safely resolves the data required by the
calculation engine. Phase 3 will consume these resolved rules and perform exact
formula calculations using decimal-safe arithmetic.
