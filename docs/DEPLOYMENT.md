# Production Deployment

## Recommended layout
Internet
  -> HTTPS reverse proxy/load balancer
  -> Go PAYE API
  -> MySQL private network

Do not expose MySQL publicly.

## Before first deployment
- Set production secrets outside Git.
- Generate a unique JWT secret.
- Create least-privilege MySQL user.
- Run migrations in order.
- Configure CORS to exact Budget254 origins.
- Verify backups and restore procedure.
- Verify `/health/live` and `/health/ready`.
- Run payroll regression suite.
- Verify admin permissions.
- Confirm published historical rules cannot be overwritten.

## Deployment procedure
1. Backup database.
2. Run CI successfully.
3. Apply migrations.
4. Build immutable application image.
5. Deploy.
6. Check readiness endpoint.
7. Run smoke calculation against a known fixture.
8. Monitor logs.
9. Roll back application if smoke checks fail.

## Rollback
Application rollback does not automatically roll back database migrations. Every migration must have a deliberate rollback/recovery plan.
