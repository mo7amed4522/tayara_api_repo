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
| [`config.release.large.yaml`](config.release.large.yaml) | **Large** — distributed cloud | `GVA_CONFIG=config.release.large.yaml go run .` | managed RDS/Aurora (`db.tayara.internal`), ElastiCache, Atlas Mongo (payment ledger on), media on **Cloudflare R2** (`oss-type: cloudflare-r2`) |

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
media.tayara.com      -> Cloudflare R2 (oss-type: cloudflare-r2)
   free plan: R2 custom domains are paid. Serve publicly via the bucket's r2.dev URL
   (set R2_PUBLIC_URL=<bucket>.<account>.r2.dev) or a free Worker bound to R2.
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

### Public response caching

Public home, explore, package, agent-detail, and search responses use a shared Redis cache
with a local Ristretto hot tier. The cache model is **“one loader per key per hour; everyone
else reads Redis”**: a request loads a missing key once, and concurrent requests share that
result. The nominal TTL is one hour, with the cache's existing jitter applied to avoid
synchronized expiry. Request parameters are part of each key; user-specific wishlist flags
are read fresh and are never cached with the public recommendation pool. Agent-scoped explore
data uses a 60-second TTL.

When admin changes require immediate invalidation, remove the key from both tiers:

```go
cache.Delete(key)
if err := cache.DeleteRedis(ctx, key); err != nil {
	// Log or report the invalidation failure.
}
```

`cache.GetCacheStats()` exposes the existing L1 hit, miss, and eviction counters.

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

## Kafka testing from Postman

The support chat does not expose Kafka over HTTP. An HTTP `POST /api/v1/support/conversations/:id/messages`
only writes MySQL; the service then **produces** one record to a Kafka topic, and every replica's
**consumer** fans it out to the SSE clients of that conversation:

```
POST /messages ──> INSERT support_message ──> produce (topic, key = conversation id, value = envelope)
                                                   │
             GET /conversations/:id/stream (SSE) <── fanout consumer group
```

So a Kafka "request" is a **produce** (record: topic + key + value + headers + partition) and a
**consume** (topics + start position + stop condition + auth), not a GET/POST. Postman speaks
Kafka natively (protocol picker → **Kafka**), so both legs can be tested there.

### 0. One-time broker setup

```bash
docker compose up -d redpanda redpanda-console   # broker on localhost:9092, UI on :8083
docker compose exec redpanda rpk topic create \
  tayara.support.chat.messages.v1 tayara.support.chat.events.v1 \
  tayara.support.chat.messages.dlq.v1 tayara.support.presence.v1
```

The app never auto-creates topics (`support.kafka.auto-create-topics: false`), so create them
explicitly. Bootstrap address from the host is `localhost:9092`; from another container it is
`redpanda:29092` (`rpk`/Console list every topic and record for inspection).

| Topic | Carries | Record key |
|-------|---------|------------|
| `tayara.support.chat.messages.v1` | `message.created` | conversation id (`"1"`) |
| `tayara.support.chat.events.v1` | `ticket.opened`, `ticket.assigned`, `ticket.reviewed`, `faq.voted` | ticket or conversation id |
| `tayara.support.presence.v1` | `presence.changed` (transitions only) | actor key (`customer:7`, `support:9`) |
| `tayara.support.chat.messages.dlq.v1` | envelopes the consumer rejected | — |

**Import the collection folder.** Postman 12 stores collections in *Collection Format v3*, which is
a **folder** of `*.request.yaml` files — so `postman/tayara-support-kafka/` is the artifact to
import, not a single file inside it. Drag that whole folder onto the Import window (or
**Import → Folder**). A lone `.json`/`.yaml` file that is not a v2.1 collection, spec or cURL is
what produces *"This is not a collection, spec or cURL, so there is nothing to import."*

The folder holds six Kafka requests, already pointed at `localhost:9092`:

| Request | Mode | Topic |
|---------|------|-------|
| `01 Produce - message.created` | Producer | `tayara.support.chat.messages.v1` |
| `02 Consume - chat messages` | Consumer | `tayara.support.chat.messages.v1` |
| `03 Produce - ticket.assigned event` | Producer | `tayara.support.chat.events.v1` |
| `04 Consume - ticket and FAQ events` | Consumer | `tayara.support.chat.events.v1` |
| `05 Produce - presence.changed` | Producer | `tayara.support.presence.v1` |
| `06 Consume - presence changes` | Consumer | `tayara.support.presence.v1` |

Producers send one record with the record key and the `message.created` / `ticket.assigned` /
`presence.changed` JSON payload pre-filled (bodies live in each request's
`.resources/<request>.resources/messages/` directory). Consumers start at the beginning of the log
and stop after 10 records or 5 seconds idle, so each is a bounded request rather than a stream.

`postman/support-kafka.asyncapi.yaml` is the wire-contract reference — envelope schemas, topics,
server and record-key rules — kept in sync with `service/support/kafka/schema.go`. Read it; there is
nothing to import from it.

The manual steps below describe the same request field by field, for when you build one by hand.

### 1. Produce — create the request

> Fast path: import the folder as described above and skip to **Send**; the steps below are what
> each imported field means.

1. Click **+** in the sidebar → choose **Kafka** in the protocol picker, save it in a `Kafka`
   collection, and name it e.g. `Kafka · produce message.created`.
2. **Broker connection**
   - **Bootstrap URLs**: `localhost:9092`
   - **Secure connection**: off (local Redpanda is PLAINTEXT)
   - **Compression**: `none`; **Strict SSL**: off
3. **Authorization**: **No Auth** for the compose broker. Use **SASL** (PLAIN / SCRAM-SHA-256 /
   SCRAM-SHA-512) for Confluent Cloud, **AWS Signature** for MSK, **OAuth 2.0** where required.
4. **Mode**: **Producer**.
5. **Topic name**: `tayara.support.chat.messages.v1`.
6. **Message value**: content type **JSON**, then an envelope shaped like the schema:

   ```json
   {
     "schema_version": 1,
     "event_id": "018f4c1e-8a3b-7c2d-9e41-2f6b7c8d9e01",
     "event_type": "message.created",
     "occurred_at": "2026-11-26T10:00:00Z",
     "conversation_id": 1,
     "ticket_id": 1,
     "data": {
       "message_uuid": "018f4c1e-8a3b-7c2d-9e41-2f6b7c8d9e02",
       "seq": 1,
       "sender_type": "support",
       "sender_id": 1,
       "body": "hello from Postman over Kafka",
       "body_hash": "0000000000000000000000000000000000000000000000000000000000000000",
       "prev_hash": "0000000000000000000000000000000000000000000000000000000000000000",
       "created_at": "2026-11-26T10:00:00Z",
       "attachments": []
     }
   }
   ```

   `schema_version` must be `1` and `event_type` must be one of the six known values, otherwise the
   fanout consumer drops the record. `conversation_id` must match the conversation whose SSE
   clients should receive it.
7. **Message key**: the conversation id, e.g. `1`. With the producer's hash balancer the key decides
   the partition, so one conversation always lands on one partition and keeps its order.
8. **Message headers**: optional — the service itself sends none and the consumer ignores them.
   Add `content-type: application/json` and `event_type: message.created` for readability in rpk.
9. **Partition**: leave **Auto**. Set it only when you deliberately want to bypass key hashing.
10. **Ack**: check it, so the response returns the record's **partition** and **offset**.
11. **Send**. Then confirm the record exists:

    ```bash
    docker compose exec redpanda rpk topic consume tayara.support.chat.messages.v1 --num 1
    ```

### 2. Consume — create the second request

1. **+** → **Kafka**, name it `Kafka · consume chat messages`, same **Bootstrap URLs** and auth.
2. **Mode**: **Consumer**.
3. **Topics**: `tayara.support.chat.messages.v1`. Add more rows for the events/presence topics, or a
   regex row `^tayara\.support\..*$` to watch all support topics at once.
4. **Start From**: `Earliest` to replay the backlog, `Latest` (default) to wait for new records,
   `Offset` to resume from a known offset, `Timestamp` to catch up to a moment.
5. **Stop After**: `Records` = `1` (or a small number). A live topic is unbounded, so Postman always
   runs a **bounded** consume; `Duration`, `Idle` and `Timestamp` are the other stop conditions, and
   an unset bound falls back to 1000 records. A bounded consume is what makes the request usable in
   the Collection Runner/CI.
6. **Partitions / start offset / end offset**: optional per-topic narrowing — leave empty normally.
7. **Authorization**: as in step 1 (No Auth locally, SASL/AWS/OAuth 2.0 otherwise).
8. **Scripts**:
   - *Before consume* — nothing needed for a plain read.
   - *On message* (runs once per received record):

     ```js
     const env = JSON.parse(pm.message.value);
     pm.test("envelope parses", () => pm.expect(env).to.be.an("object"));
     pm.test("schema_version 1", () => pm.expect(env.schema_version).to.eql(1));
     pm.test("known event_type", () => pm.expect([
       "message.created", "presence.changed", "ticket.opened",
       "ticket.assigned", "ticket.reviewed", "faq.voted"
     ]).to.include(env.event_type));
     pm.collectionVariables.set("lastSeq", env.data.seq);
     console.log("key=" + pm.message.key, env);
     ```

   - *After consume*: `pm.test("consume completed", () => pm.response.to.not.have.property("error"))`.

   For the producer request the hooks are *Before produce* (set
   `pm.request.key`/`pm.request.topic`/`pm.request.value`/`pm.request.headers` from variables) and
   *After produce* (assert the returned partition/offset).

### 3. End-to-end check with a real subscriber

Kafka records only reach a browser when somebody is subscribed to that conversation. Start a
subscriber first, then produce:

```bash
# x-token = the customer token from POST /api/v1/user/login
curl -N -H "x-token: $CUSTOMER_TOKEN" \
  "http://localhost:8888/api/v1/support/conversations/1/stream?sinceSeq=0"
```

Keep that stream open (or the browser page at `http://localhost:8888/api/v1/support/sse-client`,
which sets the `x-token` cookie from a form), **Send** the producer request, and the frame
`event: message.created` with your value appears immediately.

> Records produced by hand are **transport-only**: they are not inserted into MySQL, so they never
> show up in message history or `sinceSeq` replay. To test persistence, use the HTTP
> `POST .../messages` request — that path also produces the same envelope.

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

## Deploy & endpoints

The app always listens on `:{{system.addr}}` (= `8888`) on all interfaces. Put a reverse proxy /
load balancer in front for TLS + the public domain. Config reads `${VAR}` / `${VAR:-default}` from
the environment (`core/viper.go`), so set secrets as env vars — never edit them into config files.

**Required env vars** (deploy): `JWT_SIGNING_KEY` (≥32 chars), `MYSQL_HOST`/`MYSQL_USERNAME`/
`MYSQL_PASSWORD`, `REDIS_PASSWORD`, `STRIPE_SECRET_KEY`/`STRIPE_WEBHOOK_SECRET`, and for R2:
`R2_ACCOUNT_ID`, `R2_ACCESS_KEY_ID`, `R2_SECRET_ACCESS_KEY`, `R2_BUCKET`, `R2_PUBLIC_URL`,
`R2_PATH`. Enable admin registration with `SUPER_ADMIN_KEY`.

```bash
go build -o tayara-api .          # or build the release binary
GVA_CONFIG=config.release.large.yaml ./tayara-api
```

**Endpoints you get from config.yaml** (`system.addr: 8888`):

| Client | Base URL | Auth header | Login |
|--------|----------|-------------|-------|
| API base | `http://localhost:8888` (prod: your domain) | — | — |
| Mobile user | `…/api/v1/user/login` | `X-token` | `POST /api/v1/user/login` |
| Agent dashboard | `…/api/v1/agent/login` | `A-token` | `POST /api/v1/agent/login` |
| Admin dashboard | `…/api/v1/system/login` | `S-token` | `POST /api/v1/system/login` |
| Media (R2) | `https://${R2_PUBLIC_URL}/${R2_PATH}/…` | — | — |
| Stripe webhook | `…/api/v1/payment/webhook/stripe` | — | — |
