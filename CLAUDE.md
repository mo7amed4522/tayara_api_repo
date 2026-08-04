# Tayara API — Project Rules (conventions to follow)

This file is the **must-follow** contract for writing code in this repo. It mirrors the existing
gin-vue-admin / GVA architecture already established (see `main.go`, `initialize/`, existing
`model/{bxapi,dxapi,system}`, `service/`, `api/v1/`, `router/`, `middleware/`, `utils/`, `global/GVA_MODEL`).

The product is a **Travel Marketplace** (Admin Dashboard + Agent Dashboard + mobile users). There are
three authenticated actors → three already-existing domain groups:

| Actor        | Group    | Model              | Auth middleware          |
|--------------|----------|--------------------|--------------------------|
| Platform Admin | `system` | `SysAdminProfile` | `SystemJWTAuth()`        |
| Travel Agent   | `dxapi` (Manage) | `DxAgentProfile` | `ManageJWTAuth()`        |
| End User (mobile) | `bxapi` | `BxsUserProfile`  | `BxDomainJWTAuth()`      |

New marketplace modules are added as new groups (see `docs/ROADMAP.md`) wired the same way.
`router_biz.go` builds the public/private guard groups; `gorm_biz.go` registers models for migration.

---

## 1. Layer rules

### `model/`  (one entity per file, package per module)
- Embed `global.GVA_MODEL` (`CreatedAt`, `UpdatedAt`, `DeletedAt`) for any auditable/soft-delete row.
  For simple reference/join tables use a bare `ID uint gorm:"primarykey"`.
- Always define `func (M) TableName() string`. Table = snake_case, entity name singular.
- GORM tags: `gorm:"size:191;index;comment:..."`, plus `json:"..."` and `form:"..."` for binding.
- **Money = `int64` in minor units** (AED fils). **Percentages = `int64` basis points**, 10000 == 100%
  (already the convention: `DxAgentProfile.CommissionPercent`).
- Use `model/common.JSONMap` for JSON columns (gallery, selected activities, requirement summaries).
- Request DTOs → `model/<module>/request/*.go` (`XxxReq`). Response DTOs → `model/<module>/response/*.go`.
- Enums (booking/payment/KYC/status) live in `constants/` as typed string consts — **single source of truth**;
  never scatter status strings as literals.

### `service/`  (all business logic lives here)
- Package `service/<module>/enter.go` declares the default instance, e.g. `var XxxServiceApp = new(XxxServices)`.
- Methods: value receiver on the `XxxServices` struct; access DB via `global.GVA_DB`; guard `if global.GVA_DB == nil`.
- Return domain object(s) + `err`. Never return `gin.Context`.
- **Row-scoping is enforced in the service every time**: agent-facing lists/aggregates are filtered by the
  authenticated `agent_id`; admins see all. Pass the caller identity from the API via `utils.GetUserID(c)`-style helpers.
- Status transitions happen ONLY through service methods that (1) write `booking_status_history`, (2) emit a
  notification, (3) clear the SLA timer. Never mutate a status column directly from the API layer.

### `api/` (thin transport layer only)
- Package `api/v1/<module>` with an `XxxApi struct{}` + method receivers.
- API does exactly: bind (`ShouldBindJSON` / form), pull caller identity, call the service, emit a response.
- Responses via `model/common/response` helpers only:
  `response.OkWithData(data,c)`, `response.OkWithMessage(msg,c)`, `response.OkWithDetailed(data,msg,c)`,
  `response.FailWithMessage(msg,c)`. **No business logic in the API layer.**

### `router/`
- `router/<module>/enter.go`: `type XxxRouter struct{}` and `var xxxApi = api.ApiGroupApp.XxxApiGroup.XxxApi`.
- `InitXxxRouter(publicGroup, privateGroup, [authGroup])` registers route blocks grouped `{ }`.
- Register the router struct in `router/<module>/enter.go` group, then in `router/enter.go` `RouterGroupApp`.
- Auth use: public endpoints on `publicGroup`; actor-protected use the correct guard group passed from `router_biz.go`.

### `middleware/`
- Two existing domain JWT families already cover actors: `BxDomainJWTAuth()`, `ManageJWTAuth()`, `SystemJWTAuth()`.
- Add a new middleware only when there is a genuinely distinct concern (RBAC on an actor, rate limit, audit).
  Prefer reusing/recomposing the existing ones over duplicating.

### `initialize/`
- Register **models** for `AutoMigrate` in `initialize/gorm_biz.go` → `bizModel()`.
- Wire **routes** in `initialize/router_biz.go` → `initBizRouter(...)`, attaching the correct middleware group.
- Seed **reference data** (dictionaries, countries, plan tiers, system settings defaults) via the
  `service/system` Initializer chain (`register_init.go` + `data_init.go`): each initializer implements
  `InitializerName()`, `InitializeData(ctx)`, `DataInserted(ctx)`, `MigrateTable(ctx)`, `TableCreated(ctx)`.
- Add **cron/SLA jobs** in `initialize/timer.go` (or the existing timer package), gated by config.

### `config/` + `config.yaml`
- Add config sections as structs under `config/` and mirror them in `config.yaml`.
- Business numbers (SLA hours, commission %, hold duration, currencies) are **runtime System Settings,
  never hardcoded**; default them in config/Seed data. Stripe keys live in config, masked in logs.

### `utils/`
- Reusable pure helpers only (existing: bcrypt, paginate, upload/OSs, jwt, claims, sign, dict, flame/str).
- Add generic helpers here when reused across modules, e.g. a code generator for the PRD code prefixes
  (`BK-`, `CT-`, `PK-`, `AG-`, `US-`, `NT-`, `AD-`) and minor-unit money math.

### `constants/`
- Typed enums: booking lifecycle, payment status, KYC status, agent status, package status, ad placement/status,
  notification event types, add-on types, code prefixes. This is the single source of truth.

---

## 2. Rules that always apply
1. **Money**: `int64` minor units. **Percent**: `int64` basis points. Do the math in the service and
   **snapshot** commission/price values on the booking so later recalcs are auditable.
2. **Status changes happen in services only**, and every transition writes history + notifies + touches SLA.
3. **Agent data scoping** (`agent_id` filter) is mandatory on every agent-facing query/aggregate.
4. **No floats for money**, no hardcoded business numbers in code, no direct DB writes from the API layer.
5. Log via `global.GVA_LOG` (zap). Mask secrets (Stripe/firebase) in logs.
6. Keep Swagger annotations on API methods (docs are generated).
7. Naming: file `<module>_<entity>.go`; service type `XxxServices`; api type `XxxApi`; router `XxxRouter`;
   request `XxxReq`. Package name = module slug (e.g. `booking`, `catalog`, `payment`).
8. Commit only when asked. Follow `docs/ROADMAP.md` ordering; each section is a self-contained slice
   (model + service + api + router + initialize wiring) so it can be merged/worktree independently.