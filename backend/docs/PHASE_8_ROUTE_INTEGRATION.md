# Phase 8 Route Integration

Mount under the existing API v1 router.

## Admin authentication
- POST /api/v1/admin/auth/login
- POST /api/v1/admin/auth/refresh
- POST /api/v1/admin/auth/logout

## Rule management
- GET /api/v1/admin/rule-sets
- POST /api/v1/admin/rule-sets
- GET /api/v1/admin/rule-sets/{id}
- PUT /api/v1/admin/rule-sets/{id}
- POST /api/v1/admin/rule-sets/{id}/submit-review
- POST /api/v1/admin/rule-sets/{id}/publish
- POST /api/v1/admin/rule-sets/{id}/archive

All admin routes except login/refresh must:
1. Require a separate admin JWT.
2. Resolve roles and permissions.
3. Write an audit event for mutations.
4. Never expose password hashes or refresh tokens.

The calculator resolver must read only PUBLISHED rule sets.
