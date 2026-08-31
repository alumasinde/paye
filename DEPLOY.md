# Deploying the Budget254 PAYE API to production

This gets `https://api.budget254.co.ke` running on a VPS, with automatic
HTTPS, behind rate limiting and security headers, deployed automatically
on every push to `main`.

You need a VPS - **not** shared cPanel hosting (it won't run a Go
binary as a long-running service; see the earlier conversation for why).
A $4-6/month box from any provider (DigitalOcean, Hetzner, Linode, a
Kenyan provider) with 1GB RAM is enough for this app's scale.

## 1. Point DNS at the server

Create an **A record** for `api.budget254.co.ke` pointing at your VPS's
IP address, in whatever panel manages `budget254.co.ke`'s DNS (this does
NOT need to be the same host as your shared plan - only the DNS record
needs to point here). If the VPS has an IPv6 address, add an **AAAA**
record too.

Wait for it to propagate before step 5 (`dig api.budget254.co.ke` should
return the VPS IP) - Caddy needs the domain to already resolve correctly
before it can get a certificate from Let's Encrypt.

## 2. Install Docker on the VPS

```bash
curl -fsSL https://get.docker.com | sh
sudo usermod -aG docker $USER
# log out and back in for the group change to apply
```

## 3. Clone the repo

```bash
sudo mkdir -p /opt/budget254-paye
sudo chown $USER /opt/budget254-paye
git clone <your-repo-url> /opt/budget254-paye
cd /opt/budget254-paye/paye
```

(The path `/opt/budget254-paye` is what the CI deploy job assumes - see
step 7. Use a different path if you want, but update `.github/workflows/ci.yml`
to match.)

## 4. Configure secrets

Two separate env files, for two separate purposes - see the comments in
each `.example` file for why they're split this way:

```bash
cp .env.production.example .env
cp backend/.env.production.example backend/.env.production
```

Edit **`.env`** (root level - feeds `docker-compose.production.yml`):
- `DOMAIN` - should already be `api.budget254.co.ke`
- `ACME_EMAIL` - your real email (Let's Encrypt uses this for expiry warnings)
- `DB_PASSWORD` / `DB_ROOT_PASSWORD` - generate with `openssl rand -hex 24` each

Edit **`backend/.env.production`** (the app's full config):
- `JWT_SECRET` - generate with `openssl rand -hex 32`
- `CORS_ALLOWED_ORIGINS` - your admin frontend's real production URL
- Leave `DB_HOST`/`DB_NAME`/`DB_USER`/`DB_PASSWORD` as-is - Compose
  overrides them from the root `.env` automatically (see step 4's note
  in the file itself)

## 5. First deploy

```bash
docker compose -f docker-compose.production.yml up -d --build
```

This builds and starts the API, MySQL, and Caddy. Caddy will request a
Let's Encrypt certificate for `api.budget254.co.ke` automatically on
first request - this only works if DNS from step 1 has already
propagated.

## 6. Run migrations

The `migrate` service is excluded from the normal `up` (it's a one-off
tool, not a long-running service):

```bash
docker compose -f docker-compose.production.yml --profile tools run --rm migrate up
```

Run `... run --rm migrate status` any time to see which migrations have
applied.

## 7. Verify

```bash
curl https://api.budget254.co.ke/health
curl https://api.budget254.co.ke/api/v1/ready
```

Both should return `200`. `/ready` also pings the database - if it
fails, check `docker compose -f docker-compose.production.yml logs api`.

## 8. Wire up automatic deploys

The `deploy` job in `.github/workflows/ci.yml` SSHes into the VPS and
re-deploys on every push to `main`, after tests pass. Add these
repository secrets (GitHub repo -> Settings -> Secrets and variables ->
Actions -> New repository secret):

| Secret | Value |
|---|---|
| `DEPLOY_HOST` | the VPS's IP or hostname |
| `DEPLOY_USER` | the SSH user from step 2/3 |
| `DEPLOY_SSH_KEY` | a private key whose matching public key is in that user's `~/.ssh/authorized_keys` on the VPS - generate a dedicated deploy key with `ssh-keygen -t ed25519 -f deploy_key -N ""`, add `deploy_key.pub` to the VPS, paste the contents of `deploy_key` (the private half) into this secret |
| `DEPLOY_PORT` | only needed if SSH isn't on port 22 |

After that, every push to `main` that passes `backend` and `admin` CI
runs `git pull` + `docker compose up -d --build` on the VPS, then
verifies `/health` responds before marking the workflow green. **Note:**
this does not re-run migrations automatically - run step 6 manually
after a deploy that adds a new migration file.

## 9. Point the Android app at it

Once `https://api.budget254.co.ke` is live, update
`frontends/android/.env`:

```
EXPO_PUBLIC_API_URL=https://api.budget254.co.ke
```

This is the change that finally makes the offline detection, retry
button, and request timeouts from earlier phases exercise real network
conditions instead of a LAN IP that only works from your laptop.

## Day-to-day operations

```bash
# view logs
docker compose -f docker-compose.production.yml logs -f api

# restart just the api (e.g. after manually editing backend/.env.production)
docker compose -f docker-compose.production.yml up -d --force-recreate api

# check what's running
docker compose -f docker-compose.production.yml ps
```

## What's deliberately not covered here

- **Staging environment.** This runbook is single-environment
  (production only). If you want a staging environment later, the
  cleanest path is a second VPS (or a second Docker Compose project on
  the same VPS with different ports/domain) running the same compose
  file against `staging-api.budget254.co.ke` with its own `.env`.
- **Database backups.** `mysql_data` is a Docker named volume, which
  survives container restarts/rebuilds but not VPS loss. Set up a cron
  job running `docker exec` + `mysqldump` to somewhere off the VPS
  (S3-compatible storage, or even just scheduled `scp` to your laptop)
  before this holds real user data you can't afford to lose.
