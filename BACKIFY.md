# Backify — Ý tưởng dự án

> **Backend-as-a-Service cho developer muốn ship nhanh, không muốn config dài dòng.**
> Tạo backend hoàn chỉnh (auth, CRUD, storage) trong 5 phút qua wizard flow.

**Version:** 0.4.0
**Ngày:** 2026-09-15
**Trạng thái:** Pre-MVP / Architecture phase

---

## 1. Vấn đề

Developer muốn làm MVP / side project / app nhỏ phải setup backend từ đầu:
- Viết auth từ scratch (hoặc copy từ project cũ)
- Setup database, migration, ORM
- Build CRUD endpoints
- Wire file upload
- Deploy + config reverse proxy
- Rồi làm lại từ đầu cho project tiếp theo

→ Mất **3-7 ngày** chỉ cho boilerplate, trước khi viết dòng logic sản phẩm nào.

Các BaaS hiện có (Supabase, Appwrite, Firebase) giải quyết được — nhưng được build cho dev đã hiểu RLS, migration, multi-tenant. Với người mới, learning curve quá cao.

---

## 2. Giải pháp

**Backify dùng wizard flow thay vì dashboard.**

```
Tạo project → Chọn plan → Define Field Pool (entity + fields)
→ Chọn module (Auth, CRUD, Storage)
→ Chọn function (signup, signin, create, update)
→ Toggle field nào function đó dùng
→ Generate
→ Nhận URL + API key trong 5 phút
```

**Triết lý cốt lõi:** Bạn không config backend. Bạn lắp ráp nó từ các phần có sẵn.

### Điểm khác biệt — Field Pool

Mỗi entity (User, Product, Order) có **field pool** — danh sách tất cả field nó có thể có. Mỗi function (signup, create...) chỉ **toggle ON/OFF** field từ pool đó.

**Ví dụ — Entity "User" pool:**
```
🔒 id uuid
🔒 email email, unique
🔒 password password
fullName string
dob date
address string
phone phone
gender enum [male, female, other]
```

**User A (web giao hàng) toggle cho signup:**
```
✓ fullName ✓ dob ✓ address ✓ email ✓ phone ✓ password
✗ gender
```

**User B (web social) toggle cho signup:**
```
✓ fullName ✓ dob ✓ email ✓ phone ✓ gender ✓ password
✗ address
```

→ Cùng module. Cùng function. Config khác. Backend behavior khác. Không viết dòng code nào.

### 2.1. Field Pool là gì trong chiến lược — acquisition wedge, không phải moat kỹ thuật

Field Pool về bản chất là JSON schema với `enabled: bool` per field per function. Đối thủ (Supabase, Appwrite) copy được trong ~2 tuần nếu muốn — đây không phải moat kỹ thuật.

Vai trò thật của Field Pool: **cửa vào**, không phải **thành trì**.

| Tầng | Thứ giữ user | Vai trò |
|---|---|---|
| **Acquisition** | Field Pool UX — dễ demo, dễ hiểu, dễ viral | Kéo user vào |
| **Retention** | Tốc độ config-propagation + zero-downtime schema migration | Giữ user ở lại khi họ đã build trên Backify |
| **Defense** | Cộng đồng VN + payment VN + support tiếng Việt | Chống đối thủ quốc tế nhảy vào |

**Moat thật nằm ở 2 metric phải đo và public:**
- **Config-propagation latency** — target < 100ms từ lúc user bấm Save đến lúc request mới apply config. Cơ chế: RabbitMQ publish `project.config.updated` → Runtime subscribe → invalidate cache ngay lập tức (push-based). Redis cache TTL 30s **chỉ là fallback** (phòng khi mất message), không phải cơ chế chính — nếu benchmark cho ra >100ms, nghĩa là hệ thống đang rơi về TTL thay vì invalidation, cần điều tra.
- **Zero-downtime migration** — target 0 request fail khi user đổi field trong pool.

Marketing xoay quanh 2 con số này, không xoay quanh "Field Pool là moat".

---

## 3. Target User

**Primary persona:**
- Developer cá nhân / indie hacker / freelancer
- Biết code frontend (React, Vue, Flutter)
- Không muốn / không có thời gian setup backend
- Budget < $25/tháng
- **MVP: VN-only.** UI, docs, support bằng tiếng Việt. Không build i18n ở giai đoạn này — xem lý do ở mục 12. Global mở rộng là quyết định của Phase 2, đi cùng lúc với Stripe (mục 8, 10).

**Secondary persona:**
- Agency nhỏ làm nhiều project cho khách
- Cần backend nhanh, có thể clone template
- Cần self-host hoặc multi-tenant

**Không target:**
- Enterprise (cần compliance, SLA)
- Non-tech user hoàn toàn (họ cần Bubble, không phải Backify)

---

## 4. Backify KHÔNG phải là gì

- **Không phải no-code builder.** Bạn vẫn viết frontend. Backify chỉ lo backend.
- **Không phải Supabase clone.** Supabase cho dev đã biết làm gì. Backify cho dev không muốn biết.
- **Không phải enterprise product.** Không SLA, không compliance, không B2B sales. Đây là tool cho indie dev và team nhỏ.
- **Chưa hoàn thiện.** Đây là dự án ở giai đoạn kiến trúc. Chưa có sản phẩm public.

---

## 5. Modules có sẵn

| Module | Functions |
|--------|-----------|
| 🔐 Auth | signup, signin, forgotPassword, oauth |
| 📦 CRUD | create, read, update, delete |
| 📁 Storage | upload, download, presign |
| 🔔 Notification | sendEmail, push |
| 💳 Payment | createOrder, webhook (phase 2 — xem mục 8) |

---

## 6. Compute Plans

| Plan | CPU | RAM | Storage | Price |
|---|---:|---:|---:|---:|
| Free | 0.1 | 512 MB | 1 GB | $0 |
| 0.5c-512mb | 0.5 | 512 MB | 10 GB | $7/tháng |
| 1c-2g | 1 | 2 GB | 50 GB | $25/tháng |
| 2c-4g | 2 | 4 GB | 100 GB | $85/tháng |

**Capabilities (add-on):** Images, Videos, Audio — tính phí theo dung lượng.

**Free tier — MVP (v0.2.0, đã sửa lại cho đúng thực tế):**
- Chạy trên **shared runtime instance** (multi-tenant, đúng nguyên tắc "1 binary serve mọi project" ở mục 7)
- **Không có cold start / container shutdown thật** — vì không có container riêng cho free project ở MVP
- Chỉ có "cache miss" khi config chưa load vào Redis — vài ms, không phải vài giây
- Giới hạn: 5 GB bandwidth/tháng, 10.000 request/ngày
- Gia hạn được theo tháng

> ⚠️ **Cold shutdown thật + slot eviction là tính năng của compute-plan-thật (container-per-project), nằm ở Phase 2 (xem mục 10 "Beyond MVP"). Không quảng cáo tính năng này cho MVP.**

---

## 7. Kiến trúc

### 7.1. Nguyên tắc đã chốt

1. **Không generate code** — 1 runtime duy nhất đọc config JSON, tự route. Update config = update 1 row DB. Không rebuild, không downtime.
2. **Microservices** — 5 services độc lập, scale riêng.
3. **Clean Architecture** — domain / usecase / port / adapter / handler. Domain không import infra.
4. **Field Pool per Entity** — không fixed field set, không free-form per function.
5. **Shared Postgres + schema-per-project** — không DB-per-project.
6. **RabbitMQ** cho async, **gRPC** cho sync service-to-service.
7. **Config-driven runtime** — 1 binary serve mọi project (áp dụng cho MVP; container-per-project chỉ dành cho compute-plan-thật ở Phase 2).

### 7.2. Sơ đồ hệ thống

```
                 ┌──────────────────┐
                 │   API Gateway     │
                 │   (Traefik)       │
                 └────────┬──────────┘
                          │
   ┌────────────┬─────────┼─────────────────┬────────────┐
   ▼             ▼         ▼                 ▼            ▼
┌────────┐  ┌────────┐ ┌────────┐      ┌────────┐   ┌────────┐
│Control │  │Runtime │ │ Auth   │      │Storage │   │Billing │
│Plane   │  │Service │ │Service │      │Service │   │Service │
└───┬────┘  └───┬────┘ └───┬────┘      └───┬────┘   └───┬────┘
    │           │           │               │             │
    └───────────┴───────────┼───────────────┴─────────────┘
                             │
                      ┌──────▼──────┐
                      │  RabbitMQ   │
                      └──────┬──────┘
                             │
   ┌────────────┬───────────┼───────────────┬────────────┐
   ▼             ▼           ▼               ▼            ▼
┌────────┐  ┌────────┐  ┌────────┐     ┌────────┐   ┌────────┐
│Control │  │  Data  │  │  Auth  │     │ MinIO  │   │Billing │
│  DB    │  │  DB    │  │ Redis  │     │  /R2   │   │  DB    │
└────────┘  └────────┘  └────────┘     └────────┘   └────────┘
```

### 7.3. Services

| Service | Trách nhiệm | DB | Scale |
|---|---|---|---|
| **Control Plane** | Dashboard API — quản lý project/entity/field/module | Control DB | 1-2 |
| **Runtime** | Serve end-user API (hot path) | Data DB | 10-100 |
| **Auth** | JWT, OAuth, session, multi-tenant | Redis + Auth DB | 3-5 |
| **Storage** | Upload/download, presigned URL | MinIO + metadata | 2-5 |
| **Billing** | Usage tracking (MVP: không tích hợp cổng thanh toán, xem mục 10) | Billing DB | 1-2 |

### 7.4. Giao tiếp giữa services

| From → To | Cách | Lý do |
|---|---|---|
| Gateway → Any | HTTP | External API |
| Runtime → Auth | gRPC | Verify token, latency thấp |
| Control → Runtime | RabbitMQ | Invalidate cache (async) |
| Runtime → Billing | RabbitMQ | Track usage (async) |
| Control → Storage | HTTP/gRPC | Tạo bucket khi tạo project |

**Event bus (RabbitMQ):**
```
project.created           → Runtime (spawn schema), Billing (init usage)
project.config.updated    → Runtime (invalidate cache ngay — primary mechanism, xem mục 2.1)
project.deleted           → Runtime (drop schema), Storage (delete bucket)
user.signup.completed     → Notification (welcome email)
usage.request              → Billing (aggregate)
file.uploaded              → Billing (track egress)
```

### 7.5. Runtime — điểm mấu chốt

Runtime là service serve mọi project (MVP: shared instance cho mọi plan; container riêng chỉ áp dụng khi compute-plan-thật ra mắt ở Phase 2).

```
Request → POST /auth/signup (shop-app.api.backify.io)
↓
Traefik route theo subdomain → Runtime
↓
Runtime:
  1. Đọc config từ cache (Redis, TTL 30s = fallback, không phải cơ chế chính)
  2. Lấy signup.enabledFields
  3. Lấy pool User để validate
  4. Validate request body
  5. INSERT INTO proj_shop_app.users
  6. Return JWT + user info
```

**Cache invalidation (cơ chế chính, không phải TTL):**
- Control publish `project.config.updated` ngay khi user bấm Save
- Runtime subscribe → invalidate cache ngay
- TTL 30s chỉ là lưới an toàn khi message bị mất
- Target đo: latency từ lúc publish event tới lúc cache bị xoá phải < 100ms (xem mục 2.1)

### 7.6. Data isolation

- **Control DB:** schema `control` — platform users, projects, entities, fields, modules config
- **Data DB:** schema-per-project — `proj_abc123.users`, `proj_abc123.products`
- Custom field → JSONB column `data`
- System field → typed column
- **Auth DB:** Postgres riêng cho credentials + Redis cho session
- **Billing DB:** Postgres riêng (isolated vì security)

**Rule cứng:** Không share DB giữa services.

### 7.7. Scaling schema-per-project — benchmark trước, shard bằng lookup table

Schema-per-project trên 1 Postgres instance sẽ gặp vấn đề thật ở scale lớn:
- `pg_dump`/backup toàn instance chậm dần theo số schema
- `search_path` switching per-request cần PgBouncer cấu hình cẩn thận (transaction mode không giữ session state)
- `pg_catalog` bloat khi có hàng nghìn schema

**Kế hoạch:**
1. Tuần 1-2: viết script benchmark — tạo N schema, đo `pg_dump` time, đo `search_path` switching overhead, đo catalog size. Tìm ngưỡng X = số schema tối đa 1 instance chịu được với target p99 < 100ms.
2. Ghi ngưỡng X vào ADR-004.
3. Khi vượt ngưỡng X → **shard bằng lookup table** (`project_id → shard_id`, lưu trong Control DB).

> ⚠️ **Không dùng `hash(project_id) % N`.** Khi N tăng (thêm shard mới), hash mod đổi kết quả cho phần lớn project hiện có → phải di chuyển gần hết data giữa các instance chỉ vì thêm 1 shard. Lookup table cho phép chỉ assign project **mới** vào shard mới, project cũ không bị động tới.

### 7.8. Xoá field khỏi pool — Hybrid: chặn system field, auto-tắt custom field

**Quy tắc:**
- **System field** (`id`, `email`, `password`) → **chặn xoá hoàn toàn**. Trả lỗi `ErrSystemFieldCannotDelete`.
- **Custom field** (user tự thêm) → cho xoá, nhưng phải qua 2 bước:
  1. Check field có đang được toggle ON ở function nào không (`FindFunctionsUsingField`)
  2. Nếu có → yêu cầu client gửi lại request với `force=true` sau khi user xác nhận dialog. Nếu không có usage → xoá thẳng.

```go
func (uc *DeleteFieldUseCase) Execute(ctx context.Context, in DeleteFieldInput) error {
    field, err := uc.fields.FindByID(ctx, in.FieldID)
    if err != nil {
        return err
    }
    if field.IsSystem {
        return domain.ErrSystemFieldCannotDelete
    }

    usages, err := uc.modules.FindFunctionsUsingField(ctx, in.FieldID)
    if err != nil {
        return err
    }
    if len(usages) > 0 && !in.Force {
        return domain.ErrFieldInUse{FieldName: field.Name, Usages: usages}
    }

    if err := uc.modules.DisableFieldInAllFunctions(ctx, in.FieldID); err != nil {
        return err
    }
    if err := uc.fields.Delete(ctx, in.FieldID); err != nil {
        return err
    }

    return uc.publisher.Publish(ctx, "field.deleted", events.FieldDeleted{
        ProjectID: field.ProjectID,
        EntityID:  field.EntityID,
        FieldID:   in.FieldID,
    })
}
```

**Xử lý data — KHÔNG active-purge JSONB ở MVP:**

Custom field lưu trong cột `data` (JSONB, mục 7.6). Xoá field khỏi pool = xoá metadata, **không** chạy `UPDATE ... SET data = data - 'phone'` trên toàn bộ bảng — việc này quét full table, tốn I/O không cần thiết cho MVP. Key `phone` cũ vẫn nằm im trong JSONB của các row hiện có, nhưng Runtime chỉ đọc/validate field có trong pool hiện tại nên key đó trở thành vô hại và không truy cập được qua API.

Vì vậy **sửa lại nội dung confirm dialog** — không nói "xoá toàn bộ dữ liệu", vì thực tế MVP không purge:

> "Field 'phone' đang dùng ở 2 function (signup, signin). Xoá sẽ tắt toggle ở các function này và field sẽ không còn truy cập được qua API. Dữ liệu cũ có thể vẫn tồn tại ẩn trong hệ thống nhưng sẽ không dùng được. Tiếp tục?"

Cleanup JSONB thật (background job quét và strip key) là việc **optional, đẩy sang Phase 2** nếu cần giảm dung lượng — không phải yêu cầu MVP.

### 7.9. Auth Service — Multi-tenant auth cho end-user

Auth Service phục vụ **end-user của khách hàng** (user của project `shop-app`, `social-app`), không phải platform user (chủ project).

**Quyết định đã chốt:**

| # | Quyết định | Giá trị |
|---|---|---|
| 1 | JWT algorithm | HS256 |
| 2 | JWT expiry | 1 giờ |
| 3 | Refresh token expiry | 2 lựa chọn: 7 ngày / 30 ngày |
| 4 | Multi-tenant DB | **Database-per-project** (khác Control Plane dùng schema-per-project) |
| 5 | OAuth | Phase 2 (không có trong MVP) |
| 6 | Custom field validation | Không — chỉ system fields (email, password, fullName, phone) |
| 7 | Scope | 4 tuần |

**Multi-tenant design:**
- Mỗi project có database riêng: `auth_proj_<project_id>`
- JWT payload chứa `pid` (project_id)
- VerifyToken check `pid` khớp với expected project → chặn cross-tenant token reuse
- Auth Service subscribe `project.created` → tạo database cho project mới

**JWT (HS256):**
- Access token: HS256, expiry 1 giờ, payload `{sub, pid, email, jti, iat, exp}`
- Refresh token: random 32 bytes, hash SHA-256 lưu DB, expiry 7 hoặc 30 ngày (user chọn)
- Refresh token rotate mỗi lần refresh (dùng 1 lần)
- Token reuse detection: nếu refresh token cũ bị dùng lại → revoke toàn bộ session
- Blacklist: Redis, key `blacklist:<jti>`, TTL = JWT exp

**HS256 secret management:**
- Secret load từ env, không commit git
- 1 secret cho toàn bộ Auth Service (không phải per-project)
- Runtime cần cùng secret để verify → **vấn đề bảo mật**: nếu Runtime bị compromise → secret lộ → toàn bộ JWT giả mạo được
- **Mitigation:** Runtime không verify JWT trực tiếp, mà gọi gRPC `AuthService.VerifyToken`. Auth Service giữ secret, Runtime không cần biết.

---

## 8. Tech Stack

| Layer | Công nghệ |
|---|---|
| Ngôn ngữ | Go 1.26+ |
| HTTP | Chi (control) / Fiber (runtime) |
| RPC | gRPC + Protocol Buffers |
| Database | PostgreSQL 16 |
| Cache | Redis 7 |
| Message queue | RabbitMQ 3.13 |
| Object storage | MinIO (dev) → Cloudflare R2 (prod) |
| Reverse proxy | Traefik v3 |
| Container | Docker + Docker Compose |
| Migration | golang-migrate |
| Logging | zerolog (structured JSON) |
| Tracing | OpenTelemetry → Jaeger (**Phase 2**, không phải MVP — xem mục 10) |
| Metrics | Prometheus + Grafana (1 dashboard cơ bản ở MVP) |
| Auth | JWT + argon2id |
| Billing | MVP: chuyển khoản thủ công. Phase 2: Stripe + payOS (VN) |

---

## 9. Repo Structure

```
backify/
├── services/
│   ├── control-plane/
│   ├── runtime/
│   ├── auth/
│   ├── storage/
│   └── billing/
├── pkg/              # Shared libs (logger, config, tracing, proto)
├── deployments/      # Docker Compose, Traefik config
├── api/              # OpenAPI specs
├── docs/
│   ├── BACKIFY.md
│   └── adr/
├── go.work
└── Makefile
```

Mỗi service có `go.mod` riêng. `pkg/` chỉ chứa infrastructure, không chứa domain.

---

## 10. Roadmap (v0.2.0 — 13 tuần, realistic hơn bản 12 tuần cũ)

| Phase | Tuần | Nội dung |
|---|---|---|
| 1 | 1-4 | Foundation: monorepo, `go.work`, shared packages, Docker Compose full stack, Control Plane (domain, migration, CRUD API, auth cho platform users). **Song song: benchmark schema-per-project (mục 7.7).** |
| 2 | 5-9 | Auth Service (JWT, OAuth Google/GitHub, multi-tenant) + Runtime Service (config-driven routing, generic CRUD handler, cache invalidation qua RabbitMQ). Runtime là phần khó nhất — cho 5 tuần thay vì gộp chung với Auth. |
| 3 | 10-11 | Storage Service: presigned upload, MinIO. |
| 4 | 12-13 | Structured logging (zerolog) + 1 dashboard Grafana cơ bản (health, request rate, error rate) + deploy production (VPS + Docker Compose). |
| 5 | 14+ | Post-MVP: OpenTelemetry tracing, Loki centralized logging, Billing (Stripe + payOS), load test đầy đủ. |

**MVP không có Billing tự động** — bán thủ công qua chuyển khoản trước khi launch, tích hợp cổng thanh toán sau khi có traction.

### Beyond MVP (Phase 5+)
- Compute plan thật (container-per-project cho paid) — cold shutdown + slot eviction thật chỉ áp dụng từ đây
- Edge functions cho custom logic
- Frontend hosting (optional)
- AI-assisted config generation
- Multi-region
- Marketplace module

---

## 11. Tại sao Microservices từ đầu?

Đây là lựa chọn có chủ đích, không phải accidental complexity.

**Lý do:**
- Runtime cần scale độc lập (traffic gấp 100x Control Plane)
- Billing phải isolated vì security (giữ credentials thanh toán)
- Deploy độc lập — fix Billing không ảnh hưởng Runtime
- Auth là cross-cutting concern, cần service riêng
- Team growth — service ownership map với team structure

**Trade-off chấp nhận:** ~13 tuần tới MVP thay vì ~8 tuần với monolith. Distributed tracing, centralized logging, message idempotency là bắt buộc (nhưng tracing/logging tập trung bị đẩy sang Phase 2 để không block MVP — xem mục 10).

---

## 12. Open Questions

### Đã trả lời (v0.2.0)

- **Field type migration khi user đổi type field (vd. string → number)?**
  → **MVP: cấm đổi type.** Field immutable sau khi tạo — chỉ được thêm field mới hoặc xoá field cũ (có warning mất data). Đổi type = xoá + tạo lại.
  → **Phase 2: Versioned field.** Field có version, data cũ đọc bằng version cũ, data mới ghi bằng version mới, convert on-the-fly khi đọc, background job migrate dần.
  → Ảnh hưởng thiết kế: field pool schema ở MVP phải coi field là immutable record, không có update-in-place cho `type`.

- **Field Pool có phải moat đủ mạnh?**
  → Không phải moat kỹ thuật. Là acquisition wedge. Moat thật = tốc độ config-propagation + zero-downtime migration (xem mục 2.1).

- **Free tier economics có ổn với cold shutdown + shared runtime?**
  → MVP không có cold shutdown thật (xem mục 6). Economics MVP dựa trên giới hạn bandwidth/request, không dựa trên container eviction.

- **K8s từ đầu hay Docker Compose đủ cho 1000 project đầu?**
  → Docker Compose đủ cho MVP. Ngưỡng chuyển sang K8s gắn với ngưỡng sharding Postgres ở mục 7.7 — quyết định cùng lúc, không tách riêng.

### Đã trả lời (v0.3.0)

- **Xoá field đang được function dùng thì xử lý sao?**
  → **Hybrid (xem mục 7.8):** system field chặn xoá hoàn toàn; custom field cho xoá nhưng phải confirm nếu đang được dùng (`force=true`), tự động tắt toggle ở mọi function liên quan.
  → **Không active-purge JSONB.** Xoá field = xoá metadata + tắt toggle, không quét UPDATE toàn bảng để strip key khỏi JSONB — key cũ nằm im, vô hại vì Runtime chỉ đọc field có trong pool hiện tại. Background cleanup job (nếu cần) là việc Phase 2, không phải MVP.
  → Lý do chọn Hybrid thay vì soft-delete: soft-delete cần thêm flag, cron job, restore UI — over-engineering cho MVP khi mà đằng nào cũng không active-purge data.

- **Target VN hay global từ đầu?**
  → **VN-first, KHÔNG "global-ready" ngay.** Chỉ build UI/docs tiếng Việt, payment chuyển khoản thủ công (đã chốt ở mục 8, 10). **Không build i18n framework ở MVP** — đây là điểm khác với đề xuất "VN-first, global-ready" ban đầu.
  → Lý do: dự án đã nhiều lần cắt scope MVP theo nguyên tắc "ship lean, defer polish" (Billing thủ công, observability rút gọn — mục 10). Set up i18n (`go-i18n`/`x/text`), dịch UI + docs song ngữ ngay từ đầu là thêm việc chưa cần thiết khi chưa có 1 user trả tiền nào — đi ngược nguyên tắc đó.
  → Global expansion (i18n + Stripe) gộp chung thành **1 quyết định ở Phase 2**, làm cùng lúc sau khi có traction VN — không tách i18n ra làm sớm riêng lẻ.
  → Không cần tránh né việc code sạch: giữ chuỗi hiển thị (labels, error message cho user) ở tầng handler/presentation theo đúng Clean Architecture sẵn có (mục 7.1 #3) — việc này không tốn thêm effort vì là kỷ luật kiến trúc đã có, không phải build thêm i18n infra.

### Còn mở (v0.4.0)

- **HS256 secret cho Runtime:** Runtime verify JWT thế nào nếu không có secret? → Giải pháp: Runtime gọi gRPC `VerifyToken`. Auth Service giữ secret, verify, trả user info. Runtime không bao giờ thấy secret.
- **Database-per-project benchmark:** 1 Postgres instance chịu được bao nhiêu database? Cần benchmark như mục 7.7.
- **Refresh token duration UI:** User chọn 7d/30d ở đâu? Trong signup form hay settings sau?
- **Custom field trong Auth:** Nếu user muốn `address` trong signup → phải qua Runtime API. Có cần thêm endpoint `PATCH /auth/me` để update metadata không?

### Còn mở

- Relation field (userId → User) — UI support thế nào? (đã đẩy sang Phase 2, nhưng UI concept chưa thiết kế)
- Có nên expose code cho user custom logic, hay chỉ config? (đẩy sang Phase 3 — Beyond MVP)
- Go-to-market: Product Hunt, Hacker News, hay VN communities? (không urgent — quyết định gần ngày launch)

---

## 13. Non-goals for MVP

- ❌ Billing tự động / cổng thanh toán (Phase 5)
- ❌ OpenTelemetry full tracing (Phase 5)
- ❌ Loki centralized logging (Phase 5)
- ❌ OAuth (Phase 2 — Auth Service MVP chỉ JWT)
- ❌ i18n / multi-language UI (Phase 2, gộp với global expansion)
- ❌ Active-purge JSONB khi xoá custom field (Phase 2, optional)
- ❌ Compute plan thật / container-per-project (Phase 5+)
- ❌ Cold shutdown + slot eviction (Phase 5+)
- ❌ Edge functions (Phase 4)
- ❌ Frontend hosting (Phase 4)
- ❌ AI-assisted config (Phase 4)
- ❌ Marketplace module (Phase 4)

---

## 14. Glossary

| Thuật ngữ | Định nghĩa |
|---|---|
| **Project** | 1 backend instance user tạo. Có subdomain riêng, schema DB riêng. |
| **Entity** | Concept trong project (User, Product, Order). Có field pool riêng. |
| **Field Pool** | Danh sách tất cả field entity có thể có. Source of truth. |
| **System Field** | Field mặc định không xóa được (id, email, password). |
| **Custom Field** | Field user tự thêm vào pool. |
| **Module** | Nhóm chức năng (Auth, CRUD, Storage). |
| **Function** | Hành động trong module (signup, signin, create). |
| **Runtime** | Service serve API cho end-user. Đọc config, validate, route. |
| **Control Plane** | Service cho dashboard (user quản lý project). |
| **Cold Start** | Thời gian khởi động container khi có request đầu tiên sau sleep. Chỉ áp dụng từ Phase 2 (compute-plan-thật). |
| **Slot Eviction** | Đuổi free project khỏi node để nhường cho paid project. Chỉ áp dụng từ Phase 2. |
| **Shard (lookup table)** | Cách chia Postgres instance khi vượt ngưỡng số schema — route bằng bảng `project_id → shard_id`, không dùng hash mod. |

---

## 15. Changelog

| Ngày | Version | Thay đổi |
|---|---|---|
| 2026-09-13 | 0.1.0 | Initial draft. Chốt microservices, field pool, config-driven runtime. |
| 2026-09-14 | 0.1.0 | Restructure thành bản copy-ready. |
| 2026-09-14 | 0.2.0 | Sửa inconsistency free tier vs config-driven runtime (mục 6, 7). Định vị lại Field Pool là acquisition wedge, không phải moat — thêm metric config-propagation & zero-downtime migration (mục 2.1). Roadmap 12→13 tuần, cắt observability MVP xuống logging + 1 dashboard, bỏ Billing tự động khỏi MVP (mục 10). Chốt field type immutable ở MVP, versioned field ở Phase 2 (mục 12). Sửa kế hoạch sharding từ hash%N sang lookup table (mục 7.7). Thêm mục "Non-goals for MVP" (mục 13). Thêm open question mới: cascade behavior khi xoá field đang được dùng (mục 12). |
| 2026-09-14 | 0.3.0 | Chốt xoá field: Hybrid — chặn system field, auto-tắt toggle + confirm cho custom field, KHÔNG active-purge JSONB ở MVP (mục 7.8, 12). Chốt target thị trường: VN-only cho MVP, KHÔNG build i18n ngay — global expansion gộp chung 1 quyết định với Stripe ở Phase 2, khác với đề xuất "VN-first, global-ready" ban đầu (mục 3, 12). Thêm 2 dòng vào Non-goals: i18n và active-purge JSONB (mục 13). |
| 2026-09-15 | 0.4.0 | Control Plane MVP hoàn thành (7 bước + IDOR fix + race condition fix + transaction fix). Chốt plan Auth Service: HS256, JWT 1h, refresh token 7d/30d, database-per-project, không OAuth ở MVP, không custom field validation, scope 4 tuần. Thêm mục 7.9 (Auth Service design). Thêm open questions mới về HS256 secret, database benchmark, refresh token UI. |

---

*Đây là single source of truth cho context dự án Backify. Mọi thay đổi lớn phải update file này.*
