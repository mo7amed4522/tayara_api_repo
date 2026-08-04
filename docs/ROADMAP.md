# Tayara Travel Marketplace — Build Roadmap & Section Plan

Source: `Travel Marketplace Platform (PRD).pdf` (120pp: Admin/Agent PRD + KYC/Payment doc).
Plan is **travel**, not cars (see `CLAUDE.md` opening note). Structure/order mirrors the existing
Go + Gin + GORM (gin-vue-admin) repo: model → service → api → router → initialize wiring per module.

## Canonical decisions (resolving PRD gaps — single source of truth)

1. **Booking lifecycle** (one canonical enum in `constants/`):
   `BOOKING_REQUESTED → PRICE_PENDING → PRICE_CONFIRMED → PAID_CONFIRMED → DOCS_PENDING → DOCS_SUBMITTED
   (⇄ DOCS_PENDING_CORRECTION) → PROCESSING → ISSUED_READY → COMPLETED`, plus terminal `CANCELLED` / `REFUNDED`.
   Cancel/refund reachable from any live state.
2. **Independent axes** (three separate columns, not one compound status):
   - `payment_status`: `PENDING → AUTHORIZED → CAPTURED → PARTIALLY_REFUNDED / REFUNDED / FAILED`
   - `kyc_status`: `NOT_SUBMITTED → PENDING → APPROVED / REJECTED` (per user identity)
   - per-traveler doc `doc_status`: `MISSING → PENDING → APPROVED / NOT_OK / REQUEST_CORRECTION`
3. **KYC = user identity** (admin-reviewed) is separate from **traveler documents** (agent-reviewed, per traveler).
   Docs submission unlocks after payment; agent advances PROCESSING once docs are OK.
4. **Money** = `int64` minor units (AED fils). **Percent** = `int64` basis points (10000 = 100%). MVP currency AED.
5. **Payout** = manual admin action, per booking, after payment captured. MVP method `manual_bank_transfer`
   (`stripe_transfer_later` reserved). Eligibility gate: captured payment AND agent confirmation file uploaded
   AND admin verified AND not cancelled/refunded.
6. **Refunds** = admin-only, post-capture, via Stripe refund; amount ≤ captured; reason required.
7. **Custom trips**: MVP single-country. Config lives in `custom_trip_templates` (+ requirement profiles).
8. **Out of scope (RED / v2 — do not build)**: loyalty, AI recs, advanced analytics, real-time inventory,
   internal "employee" role, Stripe payouts, doc-abuse monitoring automation (log only).
9. **Agent "own data" scoping** is enforced at the service layer on every agent-facing query.

## Modules (new groups wired same as `system`/`bxapi`/`dxapi`)

| # | Module | Package(s) | Contents |
|---|--------|-----------|----------|
| A | Base/dictionaries | `base` | `countries`, `cities`, `nationalities`, `document_types`, `package_categories`, `system_settings`, shared dict |
| B | Catalog | `catalog` | `packages`, `package_availability`, `package_hotels`, `package_activities`, `package_addons`, `package_required_fields`, `package_document_requirements`, `package_views`, custom-trip templates + requirement profiles |
| C | Booking | `booking` | `bookings`, `booking_travelers`, `traveler_documents`, `custom_trip_requests`, `booking_offers` (final offer), `booking_status_history`, `booking_documents`, `disputes` |
| D | Payment | `payment` | `payment_transactions`, `commission_records`, `refunds`, `payouts` (+ Stripe integration) |
| E | Engage | `engage` | `reviews_ratings`, `notifications`, `notification_templates`, `ad_creatives` (+clicks), `wishlist`, `shares`, `coupons` |
| F | Analytics | adds read-agg endpoints across B–E | Admin (Users/Agents/Business) + Agent cards (bookings, revenue, confirmation time, SLA score, package/profile views) |

Identity layer already in the repo and reused/extended: `system.SysAdminProfile`, `bxapi.BxsUserProfile` (users),
`dxapi.DxAgentProfile` (agents — already has plan_id, commission_percent, sla_hours, approved_by). Add:
`agent_plans` (commission tiers), `agent_documents`, `agent_notes`; extend agent approval/commission/SLA flows.

## Execution order (strict — build exactly this, in this order)

**Phase 0 — Foundation & configuration**
- `config.yaml` + `config/`: `stripe`, `notification`, travel business defaults (SLA hours, hold duration,
  auto-capture, commission plan defaults, currency). Wire into config structs.
- `constants/`: booking lifecycle, payment_status, kyc_status, doc_status, agent/package/ad/notification
  enums + code prefixes (BK-, CT-, PK-, AG-, US-, NT-, AD-).
- `utils/`: code generator (per-prefix sequences), minor-unit money helpers, SLA clock helper.
- Base models `A` (countries/cities/nationalities/document_types/settings) + seed via initializer chain.
- Extend identity: `agent_plans`, agent documents/notes, KYC columns on `bx_user_profile`.
- **Exit:** app boots, tables migrate, base ref data seeded, enums/constants compile-clean.

**Phase 1 — Catalog (Agent side, no money yet)**
- `packages` + all child entities + package views; agent CRUD w/ status draft→pending_approval→active;
  admin moderate (accept/hide/flag); discount min-3% enforced; custom-trip template config.
- Wire router `Manage` (agent) + `System` (admin moder).
- **Exit:** agent can create/manage packages; admin can moderate; public read endpoints return catalog.

**Phase 2 — Booking lifecycle (core state machine)**
- `bookings` + travelers + documents + `custom_trip_requests`; lifecycle transitions as services;
  status history writer; `booking_offers` (final offer → PRICE_CONFIRMED); SLA timers (agent + user pay window)
  in `initialize/timer.go`; notifications on transitions.
- **Exit:** end-to-end booking request → offer → price confirmed, with history + SLA + notifications; no payment.

**Phase 3 — Payment (Stripe) & commission**
- Stripe integration (payment intent, auth hold, capture, capture-deadline UI support); `payment_transactions`;
  commission calc (basic/premium defaults; admin override w/ audit); snapshot on booking; refunds (admin).
- **Exit:** user pays, payment authorized→captured; booking auto-advances to DOCS_PENDING; commission computed.

**Phase 4 — Docs / KYC workflow**
- Per-traveler document submission + agent review (approve / request correction); process/issuance;
  KYC approve/reject (admin); booking_documents (agent confirmation file).
- **Exit:** full DOCS flow to COMPLETED with audit logging on document access.

**Phase 5 — Payouts**
- `payouts` (manual admin), eligibility gates, Pay-Agent modal, status paid + reference; agent payout views.
- **Exit:** admin can pay an eligible agent; payout reflected in agent earnings.

**Phase 6 — Content & engagement**
- `reviews_ratings` (post-completion), `notifications` + templates + resend, `ad_creatives` scheduler + clicks,
  `wishlist`/`shares`, `coupons` (display-scope MVP unless decided otherwise).
- **Exit:** reviews, notifications, ads, coupons/wishlist read/write endpoints land.

**Phase 7 — Analytics & hardening**
- Aggregate read endpoints: Admin Users/Agents/Business cards; Agent cards (avg confirmation time, SLA score,
  revenue, package/profile views). RBAC row-scoping pass on every list/agg. Swagger/docs regen.
- **Exit:** dashboards' data contracts available; agent scoping verified; docs regenerate clean.

**Phase 8 — Verification & wrap**
- End-to-end smoke (seed → booking → pay → docs → complete → payout), edge cases (cancel, SLA breach, refund,
  commission override), `go vet`/tests, README/run notes.

Each phase is a self-contained slice; proceed to the next only when the previous phase's exit criteria pass.