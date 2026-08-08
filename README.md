# Tayara API — Travel Marketplace Backend

Go + Gin + GORM (gin-vue-admin) backend for the Tayara travel marketplace (inbound UAE /
outbound international trips). Source of truth for requirements is the PRD; the build plan is in
[docs/ROADMAP.md](docs/ROADMAP.md). This is a **travel** marketplace (not cars) — see
[CLAUDE.md](CLAUDE.md).

## Modules (in build order)

| Phase | Module | Purpose |
|-------|--------|---------|
| 0 | foundation | `constants/`, `config/travel.go`, utils (codegen, minor-unit money, SLA clock), base reference data, seed initializer |
| 1 | `catalog` | packages + children (hotels/activities/addons/availability/required-fields), custom-trip builder, agent CRUD, admin moderation, public browse |
| 2 | `booking` | lifecycle state machine (red-locked), final offers, SLA timers, notifications, timeline |
| 3 | `payment` | Mongo-ledger payment lifecycle + Stripe (Visa/Mastercard/Apple Pay/Google Pay) + Tabby/Tamara BNPL, fees/commission, refunds |
| 4 | docs/KYC | per-traveler documents + agent review, KYC approve/reject, issuance → completed, audit logging |
| 5 | payouts | manual admin payouts, eligibility gates, agent earnings |
| 6 | `engage` | reviews & ratings, notification templates + resend, ad_creatives scheduler + clicks, wishlist/shares/coupons |
| 7 | `analytics` | admin Users/Agents/Business cards, agent dashboard card, RBAC row-scoping |
| 8 | verification | end-to-end smoke tests, `go vet`, README |

## Architecture notes

- **Booking lifecycle** is a single canonical state machine in
  `constants/` (e.g. `requested → agent_reviewing → price_confirmed → awaiting_payment →
  payment_authorized → paid_confirmed → docs_pending → docs_submitted ⇄ docs_pending_correction →
  processing → issued_ready → completed`; terminals `cancelled`/`refunded`). All transitions route
  through one guarded `transition()` that is **red-locked** (`utils.TryLock`; Redis `SET NX`,
  in-process fallback) so a concurrent op on the same booking is rejected.
- **Money** is `int64` minor units (AED fils); **percent** is `int64` basis points (10000 = 100%).
  No floats for money. See `utils/money.go`.
- **Payment runs through MongoDB** (qmgo): an append-only `payment_events` ledger + current
  `payment_states`, guarded by the same per-booking red lock. The authoritative MySQL row
  (`payment_transactions`) is written alongside. **Card numbers never reach the server** —
  Stripe.js/Elements tokenise client-side; we only hold payment_intent ids + client secrets.
- **Fees are never skipped**: on capture the service computes `fee = MoneyPercentOf(net, fee-bps)`,
  `commission = MoneyPercentOf(net, booking.commissionPercent)`, `agentShare = net − fee − commission`,
  and records all three to both Mongo and MySQL.
- **KYC (user identity)** is a separate axis from **traveler documents** (agent-reviewed per traveler).
- **RBAC row-scoping** is enforced in every service: the caller's id (`utils.GetAgentID`/`GetUserID`/
  `GetBxUserID`) is baked into list/aggregate queries.

## Configuration

Copy `config.yaml` and set the following via environment (never commit real secrets):

| Variable | Purpose |
|----------|---------|
| `STRIPE_SECRET_KEY` / `STRIPE_PUBLISHABLE_KEY` / `STRIPE_WEBHOOK_SECRET` | Stripe test keys |
| `TABBY_SECRET_KEY` / `TABBY_BASE_URL` / `TABBY_WEBHOOK_SECRET` | Tabby BNPL |
| `TAMARA_SECRET_KEY` / `TAMARA_BASE_URL` / `TAMARA_WEBHOOK_SECRET` | Tamara BNPL |

Active Stripe methods are `stripe.payment-methods` (visa/mastercard/apple_pay/google_pay);
platform fee is `stripe.fee-bps`. Business defaults (SLA hours, hold hours, payout gates) live in
`travel.*` and are seeded into `base_system_setting` on first boot. Payments orchestration config
(enabled methods + BNPL) is under `payment.*`.

MySQL, Redis, and MongoDB connection settings are all under `config.yaml` (`mysql`, `redis`, `mongo`).

### Release configuration (host & domain)

Production ships with **two** ready-made configs. The app always listens on `:8888` on all
interfaces — TLS + public domains are terminated by whatever fronts it (Nginx / load balancer),
so the files below only set the backing-store hosts and the CORS whitelist.

| File | Scale | Selected by | Backing stores |
|------|-------|-------------|----------------|
| [`config.release.yaml`](config.release.yaml) | **Small** — single VPS, one domain | auto in `gin.ReleaseMode` | MySQL/Redis/Mongo on the same box (`127.0.0.1`), uploads served by the app (`oss-type: local`) |
| [`config.release.large.yaml`](config.release.large.yaml) | **Large** — distributed cloud | `GVA_CONFIG=config.release.large.yaml go run .` | managed RDS/Aurora (`db.tayara.internal`), ElastiCache, Atlas Mongo (payment ledger on), media on **Cloudflare R2** (`oss-type: cloudflare-r2`, custom domain) |

**Small scale layout** — one VPS, Nginx TLS in front:
```
api.tayara.com        -> 127.0.0.1:8888        (API)
admin.tayara.com      -> dashboard SPA         (cors origin)
agent.tayara.com      -> dashboard SPA         (cors origin)
tayara.com/uploads/   -> local static files    (oss-type: local)
```
**Large scale layout** — load balancer + autoscaled replicas, CDN-hosted dashboards:
```
api.tayara.com        -> LB/WAF -> fleet on :8888   (API + Stripe webhook ingress)
admin.tayara.com      -> static build on CDN
agent.tayara.com      -> static build on CDN
media.tayara.com      -> Cloudflare R2 public bucket / custom domain (oss-type: cloudflare-r2)
db.tayara.internal    -> RDS/Aurora multi-AZ (TLS)
redis.tayara.internal -> ElastiCache / Redis Cloud
cluster0.….mongodb.net-> Atlas (payment ledger)
```

**Environment variables** (all secrets; never commit real keys). The Stripe/Tabby/Tamara vars below
are used by both release configs; the host vars are populated per scale:

| Variable | Purpose |
|----------|---------|
| `JWT_SIGNING_KEY` | JWT signing key (release) |
| `MYSQL_HOST` / `MYSQL_USERNAME` / `MYSQL_PASSWORD` / `MYSQL_DB` | MySQL/RDS connection |
| `REDIS_ADDR` / `REDIS_PASSWORD` | Redis connection |
| `MONGO_USERNAME` / `MONGO_PASSWORD` / `MONGO_OPTIONS` | MongoDB/Atlas |
| `SMTP_HOST` / `SMTP_PASSWORD` | Email SMTP |
| `R2_ACCOUNT_ID` / `R2_ACCESS_KEY_ID` / `R2_SECRET_ACCESS_KEY` / `R2_BUCKET` / `R2_PUBLIC_URL` / `R2_PATH` | Cloudflare R2 object storage (large) |
| `AWS_ACCESS_KEY_ID` / `AWS_SECRET_ACCESS_KEY` / `AWS_S3_BUCKET` | AWS S3 (large) |
| `FIREBASE_API_KEY` / `FIREBASE_CREDENTIALS_FILE` | Firebase Auth |
| `STRIPE_SECRET_KEY` / `STRIPE_PUBLISHABLE_KEY` / `STRIPE_WEBHOOK_SECRET` | Stripe test/live keys |
| `TABBY_SECRET_KEY` / `TABBY_BASE_URL` / `TABBY_WEBHOOK_SECRET` | Tabby BNPL |
| `TAMARA_SECRET_KEY` / `TAMARA_BASE_URL` / `TAMARA_WEBHOOK_SECRET` | Tamara BNPL |

> `config.yaml` (dev default) carries real-looking DB passwords and a Firebase API key in plain text.
> Keep it out of git or replace with `${VAR}` before release.

## Run

```bash
cp config.yaml config-local.yaml   # then edit / export secrets
go run .
```

Boot order (see `main.go`): config → logger → GVA_DB (MySQL) → `RegisterTables`/`InitializeData`
(migrate + seed) → Redis → MongoDB (`initialize.Mongo.Initialization`) → routes → timer
(SLA watchdog + ad scheduler). The app starts even if Mongo is down (payment will fail closed).

## Verification (Phase 8)

```bash
go build ./...
go vet ./...
go test ./...                 # note: utils/ast/TestAst expects a non-existent external path
                              # (/Users/mo7/projects/server/...) and fails on this machine; unrelated
                              # to the marketplace modules.
go test ./service/booking/ -v # end-to-end lifecycle smoke + terminal-state tests
```

The end-to-end smoke test (`service/booking/lifecycle_smoke_test.go`) drives the real services
over an in-memory sqlite DB: seed → book → offer → pay(authorize/capture) → docs → KYC →
processing → issued → completed → payout read, plus a terminal-state enforcement check.

## Security (hardening + scan)

A defensive security review found and fixed the following (see `security/attack.sh`):

| Area | Finding | Fix |
|------|---------|-----|
| Admin registration | `/system/register` was open — anyone could mint an admin S-token (`super-admin-key` empty) | Fail closed: registration rejected unless `system.super-admin-key` is set (`api/v1/system/sys_admin_profile.go`) |
| Brute force | No rate limiting on login/register/firebase-login | In-memory IP limiter `middleware.AuthRateLimit()` wired into `publicGroup` (`initialize/router_biz.go`) |
| JWT key | Weak/guessable default signing keys | Boot guard `validateJWTSigningKey()` (≥32 chars outside local) in `main.go`; strong defaults |
| JWT alg | Parsers did not pin the algorithm (alg-confusion risk) | Both `ParseToken` implementations enforce HS256 |
| File upload | No extension/size/path validation; traversal via `customPath` | Allowlist + 10 MB cap + path-traversal guard in `utils/upload/local.go` |
| Enumeration | System login returned distinct "record not found" vs "password error" | Unified generic error + timing-uniform dummy hash (`service/system/sys_admin_profile.go`) |
| SQL injection | MSSQL autocode interpolated identifiers into SQL | `sanitizeIdent()` allowlist in `service/system/sys_auto_code_mssql.go` |
| Secrets | Live Firebase service-account private key + DB passwords committed | Untracked + `.gitignore`d; rotate the key |

**Run the scanner** (probes a live instance; static checks run without a server):

```bash
./security/attack.sh                         # http://localhost:8888
./security/attack.sh http://192.168.1.10:8888
```

Each check prints `[ PASS ]` (fixed) or `[ FAIL ]` (vulnerable); the exit code is 0 only when
nothing is vulnerable. **Note:** the Firebase service-account key is still in git history — rotate
the credential in the Firebase console and, for a public repo, rewrite history.
