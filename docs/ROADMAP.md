# Tayara Travel Marketplace — Build Roadmap & Section Plan

Source: `Travel Marketplace Platform (PRD).pdf` (120pp: Admin/Agent PRD + KYC/Payment doc).
Plan is **travel**, not cars (see `CLAUDE.md` opening note). Structure/order mirrors the existing
Go + Gin + GORM (gin-vue-admin) repo: model → service → api → router → initialize wiring per module.

## Canonical decisions (resolving PRD gaps — single source of truth)

1. **Booking lifecycle** (one canonical enum in `constants/`):
   `BOOKING_REQUESTED → AGENT_REVIEWING → PRICE_CONFIRMED → AWAITING_PAYMENT → PAYMENT_AUTHORIZED →
   PAID_CONFIRMED → DOCS_PENDING → DOCS_SUBMITTED (⇄ DOCS_PENDING_CORRECTION) → PROCESSING →
   ISSUED_READY → COMPLETED`, plus terminal `CANCELLED` / `REFUNDED` (reachable from any live state).
   The PRD's intermediate dashboard states **`Agent Reviewing`** and **`Awaiting User Payment`** are
   modeled explicitly as `agent_reviewing` and `awaiting_payment` (added during the Phase-2 build),
   so the admin filter cards match the PRD's 12-state set.
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

---

# Gap plan (post-Phase-8 audit, Phases 9–12)

Audit of the PRD schema inventory vs the build above. See README/chat for the full diff; the
missing pieces that remain are grouped into four self-contained phases.

**Phase 9 — Identity & RBAC (biggest gap: dxapi + bxapi are empty shells)**
- Agent registration (public) → admin approval flow: approve / reject / suspend / activate, set
  per-agent commission % + SLA hours at approval; `agent_documents` upload/verify; `agent_notes`;
  `dx_agent_plan` CRUD. Models exist; services/APIs do not.
- User (`bxapi`) registration/login/profile + admin block/unblock; `total_spend` update on capture.
- RBAC: wire the existing System/Manage/BxDomain JWT middleware into a route-authority matrix so
  admin vs agent vs user scoping is enforced centrally, not ad-hoc per service.
- **Exit:** agents/users can register, admins approve, and every guarded route enforces its role.

**Phase 10 — Financial audit completeness (missing tables)**
- `disputes` (Resolve Dispute modal: open → resolved, admin + reason).
- `commission_records` per booking (plan snapshot, override % + reason + by-admin, audited) so the
  commission history is queryable, not just columns.
- `setting_change_history` (audit trail on every system-settings edit).
- `package_categories` lookup (honeymoon/adventure) and `custom_trip_requirement_profiles`
  (dynamic per-country required documents for the custom-trip builder).
- **Exit:** all four tables exist with CRUD + they feed existing aggregates.

**Phase 11 — Engagement depth**
- Coupon apply-to-booking: validate code/validity, apply discount at payment, mark used.
- Notification provider dispatchers (FCM push + SMTP email) behind the existing Emit interface;
  mark sent/failed on delivery.
- "Secure Price Now" reservation-hold semantics (optional hold window before price confirmation).
- Admin document-audit read endpoints (who viewed/downloaded which document).
- **Exit:** coupons affect pricing, notifications actually deliver, holds + admin audit available.

**Phase 12 — Polish & hardening**
- Regenerate Swagger (install `swag`), normalize API error codes/responses.
- Seed full demo dataset; integration tests (concurrent capture/refund under the red lock).
- Deploy/run notes for MySQL + Redis + Mongo in production.
- **Exit:** docs fresh, error contract consistent, concurrency proven, production-ready notes.

---

# System-gap plan (PRD vs roadmap audit — added after Phase 12)

PRD↔roadmap comparison (full 120pp PRD read). The domain build (Phases 0–12) is covered; the gaps
below are **system-administration** features from the PRD's Admin Dashboard that belong in the
`system` (admin) and `base` groups. Ordered by priority.

**Phase 13 — System configuration & settings (must-have, MVP)**
- Full **System Settings** API in the `system` group: `GET /api/v1/system/settings` (grouped by
  section — Commission / SLA / Payment-Stripe / Cancellation / Payout / Currency & Contact),
  `PUT settings/update` (upsert key + audit to `setting_change_history`), and reset-section.
  `SystemSetting` model + seed already exist; wire the missing routes (only `settings/history` is
  currently wired).
- **Stripe fee % setting** (`stripe_fee_percent`) feeding commission math + the commission-record
  snapshot so `commission amount = gross − strip-fee − % of gross` is correct end-to-end.
- **Admin change-password** (profile security) for the system actor.
- **Exit:** settings readable/editable/resettable with full audit; commission math includes strip fee.

**Phase 14 — Auth & notification system gaps**
- **Password reset** email flow for admin + agent (one-time reset token w/ expiry, email dispatch
  via the existing SMTP dispatcher in `service/engage/dispatch`).
- **Notification template management**: admin edit/design `notification_templates` content per
  trigger + channel, per-trigger active toggle, and a **delivery-status** view
  (Sent/Failed/Pending per channel) alongside resend.
- **Exit:** forgotten password recoverable by email; notifications fully manageable in-admin.

**Phase 15 — Scheduling, RBAC admin, & the two booking states**
- Reconcile the PRD's extra dashboard states **`Agent Reviewing`** and **`Awaiting User Payment`**
  (PRD §6.3.1) against the canonical enum (Decision 1) so admin filter cards match the PRD's
  12-state set.
- Ad placement policy: enforce max-3-active/placement + duplicate display-order block and the
  Scheduled→Active→Expired **auto-job** (models/Phase6 start this; formalize + a cron-registry view).
- RBAC: admin role/menu/permission **management** surface on top of the existing `sys_casbin.go`
  enforcement (Phase 9 covered enforcement only).
- Multi-currency: track `supported_currencies` as a system setting (keep AED-only runtime; flag the
  PRD-implied AED+USD as a recorded deviation).
- **Exit:** admin states + ad policy enforced; role/menu management available; currency config recorded.

**Phase 16 — Model-access CSV (`ir.model.access.csv`-style)** (extends the Phase-13/15 CSV feature)
- Modeled on **Odoo's `ir.model.access.csv`**: a small declarative CSV that is the *single source of
  truth* for role→model→CRUD access. One row per (role, model) grant, with 4 perm bit-flags:
  `read /write /create /unlink`. Format (the standard Odoo header):
  ```
  id,name,model_id:id,group_id:id,perm_read,perm_write,perm_create,perm_unlink
  access_pkg_user,Package.user,catalog.package,group.user,       1,0,0,0
  access_pkg_agent,Package.agent,catalog.package,group.agent,    1,1,1,0
  access_pkg_admin,Package.admin,catalog.package,group.admin,    1,1,1,1
  ```
- **Export** (`GET /api/v1/admin/export/permissions`) — a new `service/export` method dumps every
  (role, model) grant as one CSV row. Columns `id,name,model_id:id,group_id:id,perm_read,
  perm_write,perm_create,perm_unlink` via the existing `csvFromRows` helper. A maintained
  `permission_matrix.csv` ships in `settings/` (or mirrored as a `system_setting`) as the
  human-reviewable, diff-able grant sheet.
- **Import / enforce** (`POST /admin/export/permissions/import`): parse the CSV, validate each row
  (known `model_id` mapped to a GORM model/table, known `group_id` → an Authority, boolean flags),
  translate into **Casbin policies** (subject = group authority, object = model table, action =
  `create`/`read`/`update`/`delete`) and apply through the existing `SyncPolicy`/`AddPolicies`
  path in `service/system/sys_casbin.go`, then `FreshCasbin()` + `LoadPolicy()` for immediate
  enforcement. Reject malformed rows with a row-numbered error report (no partial write).
- **Audit**: each import stamps `sys_operation_record` with the importing admin, row count, and a
  before/after diff — matching the repo's audit-first rule (decision 6, CLAUDE.md §2).
- **Route guard**: all under the System (admin) guard like the rest of the `export` module; import
  is admin-only (SystemJWTAuth + Casbin `POST`).
- **Exit:** admin can download the full role→model→CRUD matrix, edit it in a spreadsheet, and
  re-upload the sheet to reprovision RBAC across every model — with immediate Casbin enforcement
  and a full audit trail.

**Nice-to-have / v2** (recorded, not blocking): Global Search, ad impression/CTR
grouped analytics, generated-ticket PDF + email distribution (partially folded into Phase 4).

Each phase is a self-contained slice; proceed to the next only when the previous phase's exit criteria pass.

Each phase is a self-contained slice; proceed to the next only when the previous phase's exit criteria pass.