# Running the PAYE backend locally

Verified end-to-end in a clean sandbox: MariaDB 10.11, migrations applied,
API server started, and `/api/v1/calculator/paye` returning correct results
for a real gross salary against the seeded 2026 Kenyan payroll rules.

## 1. Prerequisites

- **Go 1.24+** (`go.mod` requires it — see the note at the bottom if you're
  stuck on an older toolchain).
- **MySQL 8.0+ or MariaDB 10.6+**, reachable from your machine.
- The `paye/migrations/` folder must sit alongside `paye/backend/` (that's
  already how the zip is laid out — don't move `backend/` on its own).

## 2. Fetch dependencies

```bash
cd backend
go mod tidy
```

Do this once before anything else — `go.sum` as shipped is missing entries
for `golang.org/x/crypto` (a transitive dependency pulled in by the MySQL
driver), so `go build`/`go run` will fail with a "missing go.sum entry"
error until this has been run at least once with normal internet access.

## 3. Create a database user

The migrations create the `budget254_paye` database themselves, so you only
need a MySQL user with rights to create it:

```sql
CREATE USER 'budget254'@'127.0.0.1' IDENTIFIED BY 'change_me';
GRANT ALL PRIVILEGES ON budget254_paye.* TO 'budget254'@'127.0.0.1';
FLUSH PRIVILEGES;
```

Use a real password locally too — the seeded rules include live 2022-2026
Kenyan tax bands, not throwaway data.

## 4. Configure environment variables

```bash
cd backend
cp .env.example .env
```

Edit `.env` and set `DB_PASSWORD` to match what you used above. `DB_NAME`
is fully respected now — set it to whatever you want and both `cmd/migrate`
and `cmd/api` will create/use that exact database (see
`docs/ERRORS_AND_FIXES.md` for what changed here).

`.env` is loaded automatically by both `cmd/api` and `cmd/migrate` — no
need to export anything into your shell first, on any OS. (A real
environment variable, if one happens to already be set, always takes
priority over `.env` — normal convention, so CI/production env vars are
never silently overridden by a stray `.env` file.)

## 5. Run the database migrations

A `cmd/migrate` tool has been added, following the same pattern used across
your other Go projects (numeric-prefix `.sql` files applied in order,
tracked in a `schema_migrations` table).

```bash
go run ./cmd/migrate status   # see what's pending
go run ./cmd/migrate up       # apply everything pending
```

Expected output on a fresh database:

```
applied 001_budget254_paye.sql
applied 002_seed_kenya_payroll_rules.sql
applied 003_improve_budget254_paye_schema.sql
applied 004_correct_and_refine_kenya_payroll_rules.sql
applied 005_audit_checks.sql
applied 006_users_and_saved_calculations.sql
applied 007_admin_rule_management.sql
applied 008_phase9_rule_workflow.sql
applied 009_seed_initial_admin.sql
applied 010_auth_hardening.sql
applied 011_fix_admin_audit_logs.sql
```

It's safe to re-run `up` any time — already-applied files are skipped.

## 6. Run the API server

```bash
go run ./cmd/api
```

You should see:

```
2026/08/30 18:05:46 budget254-paye-api listening on :8080
```

If it hangs with no output at all and doesn't crash, something is wrong —
that used to be silent even on success, but now a clean start always
prints the line above, and a bad `DB_HOST`/`DB_USER`/`DB_PASSWORD` now
fails immediately with a clear error instead of only failing on the first
real request.

## 7. Smoke-test it

```bash
curl http://127.0.0.1:8080/health
# {"status":"ok","version":"0.5.0"}

curl "http://127.0.0.1:8080/api/v1/payroll/rules?date=2026-01-01"
# lists the published rules in effect on that date

curl -X POST http://127.0.0.1:8080/api/v1/calculator/paye \
  -H "Content-Type: application/json" \
  -d '{"gross_salary":"100000","calculation_date":"2026-01-01"}'
```

For a gross salary of KES 100,000 on 2026-01-01, this should return:

```json
{
  "taxable_income": "91430.00",
  "paye_before_relief": "22212.35",
  "relief": "2400.00",
  "paye": "19812.35",
  "total_deductions": "28382.35",
  "net_salary": "71617.65",
  ...
}
```
(Hand-checked against the seeded PAYE bands and NSSF/SHIF/AHL rates — see
`ERRORS_AND_FIXES.md` for the full breakdown.)

`gross_salary` and `calculation_date` must be **JSON strings**, not numbers
— the decoder rejects unknown/mistyped fields on purpose.

## 8. Test the app as a regular user

The calculator itself (`/api/v1/calculator/paye`, `/api/v1/payroll/rules`)
needs **no login at all** — that's the public API, already covered above.

To test the customer account side (register, log in, save a calculation,
view history):

```bash
# Register
curl -s -X POST http://127.0.0.1:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"Email":"jane@example.com","Password":"correcthorsebattery","FirstName":"Jane","LastName":"Doe"}'
# -> returns { "user": {...}, "tokens": { "access_token": "...", ... } }
```

Password must be 12+ characters. Copy the `access_token` from the
response, then:

```bash
TOKEN="paste the access_token here"

# Save a calculation
curl -s -X POST http://127.0.0.1:8080/api/v1/calculations \
  -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" \
  -d '{"calculation_date":"2026-01-01","gross_salary":"100000","taxable_income":"91430.00","paye_before_relief":"22212.35","relief":"2400.00","paye":"19812.35","total_deductions":"28382.35","net_salary":"71617.65"}'

# List saved calculations
curl -s http://127.0.0.1:8080/api/v1/calculations -H "Authorization: Bearer $TOKEN"

# Log in again later with the same account
curl -s -X POST http://127.0.0.1:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"Email":"jane@example.com","Password":"correcthorsebattery"}'

# Refresh an expired access token (grab refresh_token from the login/register response)
curl -s -X POST http://127.0.0.1:8080/api/v1/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{"refresh_token":"paste the refresh_token here"}'

# Change password (needs the access token)
curl -s -X POST http://127.0.0.1:8080/api/v1/auth/change-password \
  -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" \
  -d '{"old_password":"correcthorsebattery","new_password":"aNewPassword123"}'
```

Ten wrong login attempts locks an account for 15 minutes
(`AUTH_MAX_FAILED_REQUESTS` in `.env`) — the API returns `423
ACCOUNT_LOCKED` even for the correct password until the lockout expires.

A customer token only works on customer routes (`/api/v1/calculations`) —
it's rejected with `401` on any `/api/v1/admin/...` route, and vice versa.
There's no customer-facing UI yet, only the API — `frontends/web` in this
repo is an empty placeholder (just a `.gitkeep`).

## 9. Run the admin panel

The admin panel is a separate React app (`frontends/admin`). An admin
account is seeded automatically by the migrations:

- **Email:** `alumasinde@gmail.com`
- **Password:** `21082108`

Change this after first login via the **Change Password** page in the app
itself (see below).

With the API still running (step 6 above), in a **new terminal**:

```bash
cd frontends/admin
cp .env.example .env
npm install
npm run dev
```

Open the URL Vite prints (typically `http://localhost:5173`). Visiting
any page while logged out redirects to `/login` — sign in with the
credentials above. From there:

- **Dashboard** — real stats: how many published rules are in effect
  today, read live from the database.
- **Live Payroll Rules** (`/live-rules`) — the rules the calculator is
  actually using right now (`rule_definitions`/`rule_versions`) — **this
  is where your seeded data shows up**. Each row has a **New version**
  button that pre-fills the rule editor with that rule's current values
  (rate, cap, bands) — adjust what changed, set a new effective date, and
  submit it for review. This is the fastest way to update SHIF, NSSF, or
  any other statutory rate without retyping everything.
- **Rule Set Drafts** (`/rules`) — a governance workspace
  (`payroll_rule_sets`) for drafting new rule versions through a
  draft → review → publish workflow. Each row shows its ID — copy it into
  the Publishing page to move it through the workflow. **Publishing a
  rule set here now actually updates Live Payroll Rules** — the two used
  to be disconnected; publishing only flipped a status flag. See
  `docs/ERRORS_AND_FIXES.md` for the fix.
- **New Rule Set** (`/rules/new`) — the rule editor. Each component's
  calculation is a real structured editor now, not raw JSON, covering
  every method the live calculator actually understands: **Fixed amount**
  (one KES value), **Percentage** (a rate, with a live "= X%" preview),
  **Percentage with minimum** (SHIF's method — a rate plus a KES floor),
  **Capped percentage** (NSSF's method — a rate, an earnings ceiling, and
  a hard KES cap), **Progressive bands** (PAYE's method — an
  add/remove-row from/to/rate table), or **Tiered fixed amount** (a
  from/to/fixed-amount table). A live sandbox preview runs your draft
  against a test gross salary before you save. An "Advanced (raw JSON)"
  option is still
  there as a fallback but can't be published — the app tells you so.
- **Publishing** (`/workflow`) — paste a rule set's ID (from Rule Set
  Drafts) to submit it for review, approve/reject it, publish it, or
  archive it. Publishing runs the same transaction that updates Live
  Payroll Rules.
- **Audit Log** (`/audit`) — every rule-set create and publish, with who
  did it and when; click "Details" on a row for the full before/after
  JSON.
- **Admin Users** (`/admins`) — create new admin accounts (pick a role:
  `SUPER_ADMIN`, `RULE_EDITOR`, `RULE_APPROVER`, or `AUDITOR`), and
  enable/disable existing ones.
- **Change Password** (`/change-password`) — for your own account.

Sessions no longer just die after 15 minutes — an expired access token is
refreshed automatically using the stored refresh token, once, before
falling back to asking you to log in again. Ten wrong password attempts
locks an account for 15 minutes, customer and admin accounts alike
(`AUTH_MAX_FAILED_REQUESTS` in `.env`).

If login fails with a network error instead of a clean "invalid email or
password", check that the API's `CORS_ALLOWED_ORIGINS` includes
`http://localhost:5173` (it does by default in `.env.example`) and that
the port Vite actually started on matches.

### Changing the admin panel's colors

Colors live in `frontends/admin/public/config.txt` as plain `key=value`
lines — edit it and refresh the browser, no rebuild needed:

```
primary=#15171a
accent=#2f6feb
background=#f6f7f9
surface=#ffffff
text=#1b1d21
mutedText=#6b7280
border=#e4e6ea
danger=#a72828
success=#1f8a4c
```

Under the hood, `frontends/admin/src/styles/theme.css` defines these as
CSS custom properties on `:root` (with the values above as defaults), and
`frontends/admin/src/theme.ts` fetches `config.txt` at page load and
overwrites whichever ones it finds. `frontends/admin/src/styles/app.css`
holds all the layout/component rules and references only these variables
— no color is hardcoded there, so editing `config.txt` reskins the whole
app.

## 10. Run the test suite

```bash
go build ./...
go vet ./...
go test ./...
```

All three should be clean (`go vet` reports a few pre-existing "unkeyed
struct literal" style nits — harmless, not errors).

## 11. Docker (production-style)

`docker-compose.production.yml` and `backend/Dockerfile` are already set up
for a containerized run (Go 1.24 Alpine build → distroless runtime, plus a
MySQL 8.4 service with a healthcheck). For local dev, running natively as
above is simpler and faster to iterate on; reach for Compose when you want
to test the actual container image.

---

## Caveats worth knowing

- **`QUALIFYING_PENSION`, `QUALIFYING_MORTGAGE_INTEREST`, and
  `POST_RETIREMENT_MEDICAL_FUND` are ceilings, not automatic deductions.**
  They cap what an employee can *declare* (actual pension contributions,
  actual mortgage interest, etc.) — there's no percentage rate to compute
  automatically, so they're intentionally excluded from the automatic
  calculation and only surfaced via `GET /api/v1/payroll/rules` as
  reference limits. There's currently no endpoint that lets a caller
  declare one of these and have it validated against its cap — that's a
  product gap, not a bug, and worth deciding on separately.
- **Sandbox note on Go 1.24**: this environment only had Go 1.22 available
  and no access to `proxy.golang.org`, so verifying the build here required
  temporarily lowering `go.mod`'s `go 1.24` directive and adding local
  `replace` directives for two transitive dependencies. None of that
  shipped — the delivered `go.mod`/`go.sum` are byte-identical to what you
  uploaded. On a normal machine with Go 1.24 and internet access, none of
  this is needed; just `go build ./...` directly.
