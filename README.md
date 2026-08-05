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
