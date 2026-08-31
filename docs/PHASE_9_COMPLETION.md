# Phase 9 Completion

## Complete workflow delivered

DRAFT
→ validate
→ submit review
→ APPROVED change request
→ publish checks
→ PUBLISHED
→ archive when superseded

Rejected requests return the rule set to DRAFT.

## Rule editor
- Metadata
- Effective dates
- Official source reference
- Notes
- Multiple dynamic components
- Formula type
- Structured payload
- Validation feedback

## Validation
Server-side checks include:
- required fields
- duplicate component codes
- PAYE band structure
- PAYE rate range
- percentage-rate presence
- warnings for missing PAYE bands

## Preview
Draft rules can be tested with a sandbox gross salary before publication.

## Publishing protection
Publishing:
- requires an approved review
- checks for published effective-date overlap
- runs transactionally
- records publish events

## Final production wiring
The existing admin JWT middleware must provide the real admin database ID to workflow services. All mutation endpoints should call the audit writer. Published rule components should be immutable through the normal editor endpoint.
