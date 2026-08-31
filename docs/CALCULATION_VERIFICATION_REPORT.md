# Budget254 PAYE Calculation Verification — Phase 4

## Scope
Phase 4 adds deterministic engine verification, progressive-band boundary tests,
relief floor protection, taxable/net deduction separation, capped percentage and
tiered calculations, decimal precision checks, and an independent reconciliation
validator.

## Important boundary
These tests validate engine behaviour. Kenya-specific historical expected figures
must remain sourced from the versioned MySQL migrations and official notices; they
must not be silently duplicated into Go production code.

## Required CI command
`cd backend && go test ./...`

## Regression policy
For each published rule change, add a database-backed golden case containing:
- requested payroll date
- expected resolved version codes
- expected official calculation result
- source/reference used to verify the expected result

A failure must block release until the rule data or expected official case is
reviewed.
