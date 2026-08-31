# Errors found and fixed

Everything below was found by actually building, testing, and running the
project against a real MariaDB instance — not just reading the code.

## Critical — the backend didn't compile at all

**`internal/response/response.go` and `internal/response/json.go`** both
declared a `JSON()` function and a symbol named `Error` (one a struct type,
one a function) in the same package. That's a package-level redeclaration —
`go build` failed outright. This package is imported by `payroll/handler`
and `rules/handler`, which **are** wired into `cmd/api/main.go`, so this
wasn't dead code — the API literally could not build.

*Fix*: renamed the struct in `response.go` to `FailError`, kept one `JSON`
function. Both `response.Fail(...)` and `response.Error(...)` call sites
across the codebase are unaffected.

**`internal/middleware/middleware.go` and `internal/middleware/request.go`**
had the identical problem: both declared `type key string` and
`func RequestID`. Same fix pattern — consolidated so `ID()` now delegates to
the canonical `RequestIDFromContext()`, and the extra `Recover`/`Security`
helpers that other handlers depend on were kept.

## Critical — the PAYE calculator produced no correct results

This was the big one, and it only showed up once I ran the server against a
real, migrated database rather than just reading the code:

1. **`internal/rules/repository.Applicable()` never loaded rule parameters
   or bands.** Its own comment said *"Additional parameters/bands are loaded
   by code-specific queries in the service layer"* — but no such queries
   existed anywhere. Every calculation that needed a rate, a cap, or a tax
   band (which is nearly every calculation) failed with `CALCULATION_ERROR`.

2. Even with parameters loaded, the **production engine
   (`internal/payroll/engine.Engine.Percentage`) reads parameter keys
   `"minimum"`/`"maximum"`**, but the actual seeded data uses
   `upper_earnings_limit`, `maximum_contribution`, `minimum_amount`, etc.
   It also divides the rate by 100 (`base.Mul(rate).Div(100)`), but the
   seeded `rate` values are already fractions (`0.06` = 6%, not `6`). So
   even a "successful" calculation on old data would have been silently
   wrong by a factor of 100.

3. **`PERCENTAGE_WITH_MINIMUM` (used by SHIF) had no handler at all** in
   the calculation logic — it wasn't in the method whitelist.

*Fix*: added `Repository.ResolvedApplicable()`, which properly loads every
rule's parameters and bands from `rule_parameters`/`rule_bands`, matching
the actual seeded schema and parameter names. Rewrote
`internal/payroll/service.Service` to use this plus the new generic engine
(see below) instead of the old, data-incompatible one. The original
`Repository.Applicable()` is untouched and still backs the
`GET /api/v1/payroll/rules` listing endpoint.

4. **Three rules (`QUALIFYING_PENSION`, `QUALIFYING_MORTGAGE_INTEREST`,
   `POST_RETIREMENT_MEDICAL_FUND`) have no `rate` parameter at all** — by
   design. They're ceilings on a contribution an employee *declares*
   (`calculation_base = ACTUAL_QUALIFYING_CONTRIBUTION`), not an automatic
   percentage-of-gross deduction. Feeding them through the engine like an
   ordinary rule fails with "missing parameter rate".

*Fix*: these are now filtered out of the automatic calculation
(`category = 'ALLOWABLE_DEDUCTION'`). They still show up in the rules
listing endpoint as reference limits. There's currently no way for a caller
to declare one of these and have it capped automatically — a product
decision to make, not something I invented an answer for.

**Verified live**: gross salary KES 100,000 on 2026-01-01 →
`taxable_income=91430.00`, `paye=19812.35`, `net_salary=71617.65`. Hand-checked
against the seeded PAYE bands (10%/25%/30%/32.5%/35%), NSSF cap (4,320),
SHIF (2.75%), and AHL (1.5%) — matches exactly.

## `cmd/api` was silent on both success and failure

`app.NewFromEnv()`/`Run()` had no log output anywhere. `sql.Open` never
actually connects (it just validates the DSN and builds a lazy pool), and
nothing pinged the database at startup, so a wrong `DB_HOST`,
`DB_USER`, or `DB_PASSWORD` wouldn't surface until the first real request
hit the database - it would just sit there with zero output the whole
time, indistinguishable from a successful, silent start.

*Fix*: added `db.PingContext` (5s timeout) right after `sql.Open` in
`NewFromEnv`, returning a clear wrapped error if the database is
unreachable - verified live with a wrong `DB_PASSWORD`: fails in under a
second with `connect to database: Error 1045 (28000): Access denied...`
instead of hanging silently. Also added a `budget254-paye-api listening on
:8080` log line on successful startup and a `shutting down` line on
graceful shutdown, so a clean run is never silent either.

## Migrations ignored `DB_NAME`

Migrations 001-005 each hardcoded `CREATE DATABASE IF NOT EXISTS
budget254_paye` / `USE budget254_paye;`. Since a raw `.sql` file's own `USE`
statement runs *after* whatever database `cmd/migrate` had already
selected, this silently overrode `DB_NAME` from `.env` — set it to
anything else and the schema still landed in `budget254_paye`.

*Fix*: removed every `CREATE DATABASE`/`USE budget254_paye;` line from the
five migration files. `cmd/migrate` already creates and selects `DB_NAME`
itself before running any file, so the migrations are now fully
database-name-agnostic. Verified live: set `DB_NAME=paye_custom_name`,
ran `migrate up`, confirmed all 46 tables and 24 published rules landed in
`paye_custom_name` and `budget254_paye` was never created.

## `.env` wasn't loaded automatically

Neither `cmd/api` nor `cmd/migrate` read `.env` — both call
`config.Load()`, which only reads real process environment variables
(`os.Getenv`). On Linux/macOS this is easy to paper over with
`export $(grep -v '^#' .env | xargs)`; on Windows it isn't, and running
either command without first loading `.env` into the shell fails with
`config: DB_NAME and DB_USER are required`.

*Fix*: added `internal/envfile`, a small dependency-free loader that reads
`.env` from the working directory and calls `os.Setenv` for anything not
already set — real environment variables always win, so this can't
silently override a value set by a process manager, container, or CI.
Wired into both `cmd/api/main.go` and `cmd/migrate/main.go` before
`config.Load()` runs. Verified live: ran `migrate up` and `cmd/api` with
zero manually-exported environment variables and both worked; separately
confirmed a real `DB_NAME` env var still overrides `.env`'s value.

## `go.sum` was missing entries for a direct dependency

`go.mod` requires `golang.org/x/crypto v0.36.0` directly (pulled in by the
MySQL driver for auth), but `go.sum` had zero entries for it — not even a
`go.mod` hash line. Any `go build`/`go run` fails with *"missing go.sum
entry for go.mod file"* until `go mod tidy` (or `go mod download
golang.org/x/crypto`) is run once with real internet access. This was
already broken in the uploaded project; I couldn't fix `go.sum` myself
since my sandbox has no route to `proxy.golang.org` and I won't fabricate
checksum hashes I can't verify — a wrong hash fails harder than a missing
one. Documented as a required first step in `docs/LOCAL_SETUP.md`.

## Real data bug: migration 002 couldn't apply

Running the migration against a real database (not just reading it) turned
up **20 `INSERT INTO rule_version_sources` statements with no column
list**. That table has 3 columns (`rule_version_id`, `rule_source_id`,
`created_at` — the last with a default). Without an explicit column list,
MySQL requires a value for every column, so all 20 threw "Column count
doesn't match value count" and migration 002 (which seeds every Kenyan
payroll rule from 2022 onward) could never apply.

*Fix*: added the column list to all 20 statements in
`migrations/002_seed_kenya_payroll_rules.sql`. Verified: full migration
chain now applies cleanly on a fresh database, and re-running `migrate up`
correctly reports "nothing to apply."

## Dead code, deleted or fixed

None of these affect the running server (`main.go` only wires the minimal
path through `internal/app/app.go`), but they broke `go build ./...` /
`go vet ./...` / `go test ./...`, which matters for CI and for anyone
running the full test suite.

- **`internal/rules/service/service.go`** called `repo.Resolve()` and
  `repo.ResolveMany()`, which don't exist on the repository, and had zero
  importers anywhere in the codebase. Deleted — implementing invented SQL
  logic for a package nothing calls didn't seem worth guessing at.
- **`internal/rules/routes.go`** (only reachable through the unused
  `internal/server/router.go`) wired `h.Resolve`/`h.ResolveMany`, which
  don't exist on `rules/handler.Handler` — same orphaned API. Repointed it
  at the `Applicable` method that actually exists.
- **`internal/server/router.go`** imported `internal/payroll` (not a real
  package — no `.go` files declare it) instead of
  `internal/payroll/routes` (which declares `package payroll`, confusingly,
  from inside a `routes/` directory). Fixed the import path and alias.
- **`internal/payroll/verification/validator.go` and two test files**
  (`tests/payroll/engine_test.go`, `tests/payroll/precision_test.go`)
  referenced a generic `engine.Calculate(Input, []rules.ResolvedRule) (Result, error)`
  API that was fully specified by the tests but never implemented anywhere.
  Implemented it from the test contract — all 6 test cases pass now, and
  it's the engine that now actually powers the calculator (see above).

## Admin panel and login — built from scratch

Following up on the earlier "still open" note below: you asked for the
admin panel and login UI to be built, with `internal/server/router.go`
and its dead packages wired in. Here's everything that took, including
two more real bugs found along the way.

**Admin auth didn't exist at all**, not just unwired. `internal/admin`
only had a `ByEmail()` lookup (fetches an admin's roles/permissions) — no
password verification, no JWT issuance, no HTTP handler, no route.
Built `internal/admin/service` and `internal/admin/handler` mirroring the
existing customer auth pattern, added `StoreRefresh` to the admin
repository, and wired `POST /api/v1/admin/auth/login`.

**`admin_roles` and `admin_permissions` were never linked.** Migration 007
seeded both tables but never populated `admin_role_permissions` — meaning
no admin role, including `SUPER_ADMIN`, could ever pass a permission
check. Added the missing grants (`SUPER_ADMIN` gets everything;
`RULE_EDITOR`/`RULE_APPROVER`/`AUDITOR` get what their stated descriptions
imply).

**Seeded the first admin account** (migration 009):
`alumasinde@gmail.com` / `21082108`, role `SUPER_ADMIN`. The bcrypt hash
was generated and independently verified (a standalone
`bcrypt.CompareHashAndPassword` check) before being committed to the
migration, rather than trusted blind.

**Real bug: saving a calculation always failed.** The customer
`saved_calculations` INSERT had 15 `?` placeholders in its `SELECT` list
for only 14 declared `INSERT` columns and only 14 Go arguments supplied —
a column-count mismatch that made every save attempt fail with a generic
500. Fixed and verified live: register → save → list now all work.

**Real security bug: admin and customer JWTs were interchangeable.**
Neither `RequireAuth` (customer) nor the new `RequireAdmin` checked what
*kind* of token they were given — both just trusted any valid signature
and read `sub` out of it. An admin's token could be used on customer-only
routes and vice versa. Fixed by tagging tokens with `typ:"customer"` /
`typ:"admin"` at issue time and having each middleware require its own
type. Verified live in both directions: a customer token now gets `401`
on `/api/v1/admin/...`, and an admin token now gets `401` on
`/api/v1/calculations`.

**`payrollrules/handler.Create` used a hardcoded admin ID (`1`)** for the
`created_by` foreign key, with a comment noting route integration should
fix it. Now pulled from the authenticated admin's JWT (`uid` claim) via
`middleware.AdminDBID`. Verified live: a rule set created while logged in
correctly recorded the real admin's database ID, not the placeholder.

**`frontends/admin` didn't actually build.** Three separate problems,
found by actually running `npm install && npm run build`, not just
reading the code:
- `index.html` was a bare fragment with no `<!DOCTYPE html>`/`<html>`/
  `<body>` wrapper at all.
- `package.json` never listed `@types/react`/`@types/react-dom` as dev
  dependencies, so every `.tsx` file failed to type-check.
- No `vite-env.d.ts` referencing `vite/client`, so `import.meta.env` and
  the CSS import in `main.tsx` both failed to type-check.

All three fixed; `npm run build` now succeeds with a clean production
bundle. Also verified `npm run dev` serves correctly and a real
browser-style request (with `Origin: http://localhost:5173` and a CORS
preflight) reaches `POST /api/v1/admin/auth/login` and gets a `200` back
- the full path a real browser would take, not just a same-origin curl.

**CORS didn't exist on the API at all.** `cmd/api` had zero CORS handling
- fine when the only client was `curl`, not fine for a browser-based
admin panel. Added a CORS middleware defaulting to allow
`http://localhost:5173`/`http://127.0.0.1:5173` (the Vite dev origin),
overridable via `CORS_ALLOWED_ORIGINS`.

**`JWT_SECRET` in `.env.example` was too short.** The placeholder text
`replace_with_32+_random_bytes` is only 29 characters — once `cmd/api`
started requiring a real 32+ byte secret at startup (see below), a fresh
`cp .env.example .env` would have failed immediately. Replaced it with an
actual 64-hex-char (32 byte) generated secret so the example file works
out of the box; still meant to be replaced for anything beyond local dev.

**Wired the previously-dead customer auth and saved-calculations
routes** too, since testing "as a user" needs them:
`POST /api/v1/auth/register`, `POST /api/v1/auth/login`,
`POST /api/v1/calculations`, `GET /api/v1/calculations`,
`DELETE /api/v1/calculations/{id}`.

**Not wired**: `internal/health`, `internal/periods`, and the
`internal/server/router.go` composition root itself remain unused —
`cmd/api` still assembles routes directly in `internal/app/app.go` rather
than switching to that alternate router, since the alternate one predates
(and duplicates, with drift) everything above. The audit log has no read
endpoint yet (writes work, nothing lists them back) - the admin panel's
Audit Log page is a placeholder until that's built.

## Admin panel: missing login guard, wrong data, no theming system

Three more issues, reported after using the panel for real.

**No login guard at all.** `main.tsx`'s routing rendered `Dashboard`,
`Rules`, etc. unconditionally — visiting `/` directly skipped `/login`
entirely regardless of whether a token existed. Added a `ProtectedRoute`
wrapper that checks `isAuthenticated()` (a real check against the stored
token, not just "did the login form run this session") and redirects to
`/login`; `/login` itself redirects away if already signed in. `api()`
also now clears a token that comes back `401` (expired/invalid), so a
stale session correctly bounces back to the login screen instead of every
request failing silently.

**"Payroll Rules" showed nothing despite real seeded data existing.**
The page queried `GET /admin/rule-sets`, which reads `payroll_rule_sets`
— a *separate* admin-governance table for drafting new rule versions
through a review/publish workflow. It was empty because nothing has ever
been created there, not because anything was broken. The actual live data
— the 24 published rule versions the calculator uses — lives in
`rule_definitions`/`rule_versions`/`rule_bands`/`rule_parameters`,
exposed by the already-public `GET /payroll/rules` endpoint. Added a new
**Live Payroll Rules** page that reads from that endpoint, so the real
data is now visible, and relabeled the old page **Rule Set Drafts** with
an explanation of why it's a different, intentionally-empty workspace.
Dashboard now shows a real count of live published rules instead of
static placeholder text.

**No theming system, styles inline in one file.** Moved everything from
a single `src/styles.css` into `src/styles/theme.css` (CSS custom
properties on `:root` — the color palette) and `src/styles/app.css`
(layout/component rules, referencing only those variables). Added
`public/config.txt` — a plain-text `key=value` file served as a static
asset and fetched at page load by `src/theme.ts`, which overwrites the
matching CSS variables. Editing `config.txt` and refreshing the browser
now reskins the whole app with no rebuild. Verified live: `config.txt`
served correctly by the Vite dev server, and a clean production build
(`npm run build`) still succeeds with the new structure.

## Refresh tokens, lockout, admin management, audit log, and the publish bridge

Building out "what else to add to Admin" surfaced two more real, previously
hidden bugs — both found by actually running the feature, not by reading
the code.

**New features, all verified live:**
- **Refresh tokens** (customer + admin) — both logins already issued one
  and stored it, but there was no endpoint to use it; sessions just died
  after 15 minutes. Added `/auth/refresh` and `/admin/auth/refresh`, with
  rotation (the old refresh token is revoked the moment a new one is
  issued). Verified: refreshed successfully, then confirmed the *old*
  refresh token is rejected afterward.
- **Change password** (customer + admin). Verified live: old password
  correctly rejected after a change, new one accepted.
- **Failed-login lockout**, using `AUTH_MAX_FAILED_REQUESTS` (already in
  `.env.example`, previously unused). Verified live: 10 wrong passwords
  locked the account for 15 minutes — the 11th attempt was rejected with
  `423 ACCOUNT_LOCKED` even using the *correct* password.
- **Admin user management** — list/create/enable/disable other admin
  accounts, gated on the `admin.users` permission. Verified live:
  created a `RULE_EDITOR` admin, listed it back.
- **Audit log, readable** — `GET /admin/audit-logs`, plus writes wired
  into rule-set create and publish so there's something real to see.
- **The Publishing workflow, wired to real HTTP routes** —
  submit-review/approve/reject/publish/archive. The Go service logic
  already existed; it just had no routes. Verified the full lifecycle
  live: draft → submit → approve → publish, end to end.

**The big one — draft rule sets now actually reach the live calculator.**
Previously, "publishing" a rule set only flipped a status flag in
`payroll_rule_sets` — nothing you drafted there ever affected what the
calculator computes, because that's a completely separate schema from
`rule_definitions`/`rule_versions` (see the "Payroll Rules" data-mismatch
fix above). `Workflow.Publish` now materializes a rule set's components
into the live schema, in the same transaction as the publish itself, so
it's all-or-nothing. Supports `FIXED`, `PERCENTAGE`, and `BANDS` formula
types (the same three the generic engine understands); `JSON` formula
type is rejected at publish time with a clear error, since there's no
live-engine equivalent for arbitrary JSON.

**Verified end-to-end with real numbers**: published a rule set with a
fixed deduction, a percentage deduction, and a two-band progressive
deduction. Confirmed all three appeared correctly in
`rule_definitions`/`rule_versions`/`rule_parameters`/`rule_bands`, then
ran an actual calculation and hand-checked every number: taxable income,
statutory deduction amounts, total deductions, and net salary all matched
exactly what the published rules should produce.

**Two more real bugs found while building this:**

1. **`rule_versions.created_by`/`reviewed_by`/`approved_by`/`published_by`
   had foreign keys pointing at the wrong table.** They referenced
   `users` (originally meant, in migration 001, for internal
   rule-governance staff) — but migration 006 repurposed `users` for
   customer accounts, and migration 007 introduced `admin_users` as the
   real admin-identity table. These four foreign keys were never
   updated. Publishing a rule set (attributing it to the admin who
   published it) failed outright with a foreign key violation the moment
   it was actually tried. Fixed in migration 010.

2. **`admin_audit_logs` had the exact same collision bug as the `users`
   table** (see "Migrations ignored DB_NAME" section context above for
   the pattern): migration 001 already created a table by that name, with
   a completely different, incompatible schema (`actor_user_id`,
   `entity_id` as a bigint, no `public_id`). Migration 007's
   `CREATE TABLE IF NOT EXISTS admin_audit_logs` was silently skipped as
   a result. Every audit write had been failing since I first wired
   audit logging - hidden because the write's error was being discarded
   rather than surfaced. Fixed in migration 011 (drop and recreate with
   the schema migration 007 always intended - safe, since nothing else
   references this table and it held no real data). Also stopped
   discarding audit-write errors: they're logged now instead of silently
   swallowed, so a failure like this surfaces immediately next time
   rather than staying hidden until someone reads the log back.

## Real Calculations and Bands editor UI

The rule editor's component payload was a raw JSON textarea — functional,
but not what "real working UI" means. Replaced it with a calculation-type
switcher (Fixed amount / Percentage / Progressive bands / Advanced JSON)
that renders the right structured inputs for each:
- **Fixed amount** — a single KES amount field.
- **Percentage** — a rate field (as a fraction) with a live "= X%"
  preview.
- **Progressive bands** — a real row-by-row table (From / To / Rate) with
  add/remove-row buttons, editing the exact JSON shape the publish bridge
  above expects.
- **Advanced (raw JSON)** stays available as a fallback, with a visible
  warning that it can't be published (matches the server-side rejection).

Verified the exact payload shapes these editors produce against the live
publish bridge - a `BANDS` payload built through the row editor published
and calculated correctly.

Also: `RulePreview` (a working sandbox-preview component) existed but was
never actually rendered anywhere in the app. Wired it into the rule
editor, with proper loading/error states added.

## New admin panel pages

- **Admin Users** (`/admins`) — create/list/enable/disable admins.
- **Audit Log** (`/audit`) — now reads real data instead of being a
  placeholder; expandable before/after detail per entry.
- **Change Password** (`/change-password`).
- **Rule Set Drafts** table now shows each rule set's ID, needed to use
  the Publishing page's submit/approve/publish actions (which take an ID
  - previously there was no way to see one without inspecting network
  requests).
- `api.ts` now auto-refreshes an expired access token once (using the
  stored refresh token) before giving up and forcing a re-login, instead
  of failing every request the moment the 15-minute access token expires.

## Managing SHIF, NSSF, and other real statutory rates

The rule editor only supported `FIXED`/`PERCENTAGE`/`BANDS` — but SHIF
actually uses `PERCENTAGE_WITH_MINIMUM` and NSSF uses `CAPPED_PERCENTAGE`.
Neither could be drafted through the admin panel at all; the governance
schema's `formula_type` enum didn't even have those values.

*Fix*: migration 014 extends `payroll_rule_components.formula_type` to
cover every method the live engine (`internal/payroll/engine`) actually
understands, and the publish bridge
(`internal/payrollrules/service.Workflow`) and the rule editor UI were
extended to match: `Percentage with minimum`, `Capped percentage`, and
`Tiered fixed amount` now have real structured inputs alongside the
existing three.

Also added `GET /api/v1/admin/live-rules` (full parameter/band detail,
not just the summary the public listing returns) and a **New version**
button on the Live Payroll Rules page that pre-fills a draft from an
existing rule's current values - so updating SHIF's rate or NSSF's cap
means editing a couple of fields, not retyping the whole rule from
scratch.

**Verified live end-to-end**: pulled NSSF's real current values via the
new endpoint, converted them into a draft with an intentionally-changed
cap, published it through the actual submit → approve → publish
workflow, then confirmed a real PAYE calculation on the new effective
date enforced the new cap exactly.

## Still open — minor, worth knowing about

- **`internal/server/router.go` remains unused.** It's an alternate,
  older composition root (predates and duplicates the routes now wired in
  `internal/app/app.go`, with drift between the two). Not wired in, and
  no need to - everything worth keeping from it (health, periods) is
  either already covered or minor.
- **`internal/periods`** (historical payroll period lookups) is still not
  wired into any route. Nothing currently depends on it.
- **The `Review` step of the workflow doesn't have a UI beyond the raw
  ID + comment form on the Publishing page.** It works (verified live),
  it's just not integrated with the Rule Set Drafts table the way create
  is - you still need to copy an ID across pages by hand.
- **No customer-facing UI at all** — `frontends/web` is an empty
  placeholder. Everything customer-side (register/login/save
  calculations) is API-only, documented with `curl` examples in
  `docs/LOCAL_SETUP.md`.
