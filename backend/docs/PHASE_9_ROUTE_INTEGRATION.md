# Phase 9 Route Integration

Mount behind the existing separate admin authentication middleware.

## Editor
- GET /api/v1/admin/rule-sets
- POST /api/v1/admin/rule-sets
- PUT /api/v1/admin/rule-sets/{id}
- POST /api/v1/admin/rule-sets/validate
- POST /api/v1/admin/rule-sets/preview

## Workflow
- POST /api/v1/admin/rule-sets/{id}/submit-review
- POST /api/v1/admin/rule-sets/{id}/approve
- POST /api/v1/admin/rule-sets/{id}/reject
- POST /api/v1/admin/rule-sets/{id}/publish
- POST /api/v1/admin/rule-sets/{id}/archive

Permission mapping:
- rules.read -> list/view/preview
- rules.write -> create/edit/validate/submit
- rules.review -> approve/reject
- rules.publish -> publish/archive

Before publishing:
1. Run server-side validation.
2. Require an APPROVED change request.
3. Reject effective-date overlaps.
4. Publish in one database transaction.
5. Write audit logs.
6. Calculator resolver reads PUBLISHED only.
