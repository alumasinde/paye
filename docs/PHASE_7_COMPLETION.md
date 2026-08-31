# Phase 7 Completion

## Backend
- User registration
- BCrypt password hashing
- Access JWTs
- Opaque refresh tokens stored only as SHA-256 hashes
- User-owned saved calculations
- History listing
- Deletion scoped to authenticated owner

## Android
- Login
- Registration
- Secure token storage
- Saved calculation history

## Required integration step
Merge the new authenticated routes into the existing Phase 5 app router:
- POST /api/v1/auth/register
- POST /api/v1/auth/login
- POST /api/v1/calculations (RequireAuth)
- GET /api/v1/calculations (RequireAuth)
- DELETE /api/v1/calculations/{id} (RequireAuth)

Do not remove the public calculator endpoint.
