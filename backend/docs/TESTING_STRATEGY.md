# Production Test Strategy

## Required CI gates
1. `go vet ./...`
2. `go test ./...`
3. `go test -race ./...`
4. Build API binary
5. Build admin frontend
6. Build Android app in its existing pipeline

## Payroll regression suite
For every verified historical rule period:
- gross salary boundary below each PAYE band
- exact band boundary
- one currency unit above boundary
- statutory deduction minimum/cap boundaries
- relief boundaries
- custom deductions
- expected net pay

Fixtures should be based on verified statutory source data already governed in the rule database.

## Integration tests
Run against a disposable MySQL instance:
- migrations
- public calculation endpoint
- saved calculations authorization
- admin authentication
- draft creation
- validation
- review
- publish
- overlap rejection
- audit records

## Security tests
- unauthorized admin route access
- permission denial
- malformed JSON
- oversized request body
- invalid token
- expired token
- refresh-token reuse/revocation
- rate-limit behavior
