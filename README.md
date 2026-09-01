# Budget254 PAYE — Phase 10: Production Hardening, Testing & Deployment

Production-hardening overlay for the existing PAYE monorepo.

This phase adds:
- configuration validation
- security middleware
- request IDs and structured logging foundation
- health/readiness endpoints
- rate limiting foundation
- graceful shutdown
- Docker deployment files
- GitHub Actions CI
- MySQL backup script
- test strategy and test fixtures
- production environment template
- deployment checklist

Integrate these packages into the existing backend composition root rather than replacing prior phases.



# Terminal 1
cd backend
git pull origin main
go run ./cmd/migrate up
go run ./cmd/api

# Terminal 2
cd frontends/admin
cp .env.example .env
npm install
npm run dev

# Web Frontend
C:\xampp\htdocs\paye\frontends\web>
$env:PAYROLL_API_BASE_URL="http://192.168.100.11:8080/api/v1"

php -S 0.0.0.0:8088