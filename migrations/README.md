# Existing migrations are preserved conceptually

This Phase 5 package does not replace or rewrite your audited migrations.
Copy this overlay into your existing `paye/` repository and keep your existing:
- 001_budget254_paye.sql
- 002_seed_kenya_payroll_rules.sql
- later schema/rule audit migrations

as the database source of truth.
