# Runtime Contract

> This is a **Systems Contract** for the Go runtime service (`services/runtime`).  
> It defines how clients and operators must talk to the runtime, and what the runtime guarantees in the current MVP.  
> Code must implement this contract. Any deviation is a bug.

**Status:** Active for MVP  
**Last updated:** 2026-09-23  
**Related:** [PERMISSION_CONTRACT.md](./PERMISSION_CONTRACT.md)

Ownership semantics (who may read/update/delete which rows) are defined only in **PERMISSION_CONTRACT**. This document does not restate those rules in full; if the two documents ever conflict, **PERMISSION_CONTRACT wins** for ownership behavior.

---

## 1. Identify project

All routes under `/v1` require the project context.

| Item | Value |
|------|--------|
| Header | `X-Project-Id` |
| Value | Control-plane project ID (string) |

Behavior:

1. Middleware reads `X-Project-Id`.
2. Missing header → **400** with body roughly `{"error":"missing X-Project-Id header"}`.
3. Runtime loads project config from control-plane (gRPC, with in-memory cache TTL from `CONFIG_CACHE_TTL`).
4. Project unknown or control-plane unavailable → **404** with body roughly `{"error":"project not found or unavailable"}`.
5. On success, `ProjectConfig` is attached to the request context for handlers and usecases.

`GET /health` and routes under `/internal/*` do **not** use `X-Project-Id` (internal routes take `projectID` from the path where needed).

---

## 2. End-user JWT claims

Protected `/v1` routes expect:

```http
Authorization: Bearer <access_token>
```

Access token claims used by the runtime (JWT payload):

| Claim JSON key | Meaning |
|----------------|---------|
| `uid` | End-user id |
| `pid` | Project id |
| `role` | `user` or `admin` |

Also standard registered claims (expiry, issued-at, token type for access vs refresh) as implemented by the JWT service.

Rules:

- Unauthenticated callers have **no** access to protected resources (CRUD, `/v1/auth/me`).
- Invalid, expired, or missing access token → **401**.
- Public under `/v1` (still require `X-Project-Id`): `POST /v1/auth/signup`, `POST /v1/auth/signin`, `POST /v1/auth/refresh`.

---

## 3. Ownership rules

Ownership is derived from entity relations (n-1 → `User`), not from a separate policy language.

**Do not duplicate the full rule set here.** See [PERMISSION_CONTRACT.md](./PERMISSION_CONTRACT.md).

Summary only:

- Normal **user**: create becomes owner; read/update/delete only own rows; otherwise **403**.
- **admin**: read-all for list/get; write/delete of others' rows remains limited per PERMISSION_CONTRACT MVP.
- Hard delete only.
- No field-level ACL, no n-n ownership, no soft-delete in MVP.

---

## 4. Module field gating (`enabledFields`)

This is **config-driven field acceptance**, not field-level permission in the ownership sense.

### Auth module

Path in config:

```text
modules.auth.functions.<fn>.enabledFields
```

Examples of `<fn>`: `signup`, `signin`, and other auth functions as implemented. Auth only consumes fields listed in `enabledFields` for that function.

### CRUD module

Path in config:

```text
modules.crud.entityFunctions.<EntityName>.<fn>.enabledFields
```

| Function | When |
|----------|------|
| `create` | `POST /v1/{entity}` |
| `update` | `PATCH /v1/{entity}/{id}` |

Policy (MVP):

- Only keys present in `enabledFields` are kept from the client body.
- Keys not enabled are **stripped** (ignored), not rejected with 400 solely for being extra.
- If the crud module is disabled, missing, or `enabledFields` is empty → client-supplied map is treated as empty after filter.
- System fields are never accepted from the client even if listed: `id`, `createdAt`, `created_at`, `updatedAt`, `updated_at`.
- Owner relation column is assigned by the permission engine on create; client cannot set or change it on update (server strips owner keys from the update payload).

---

## 5. Internal endpoints and `X-Internal-Key`

Internal routes are for operators / control-plane style callers, not end users.

| Method | Path | `X-Internal-Key` required |
|--------|------|---------------------------|
| `POST` | `/internal/bootstrap/{projectID}` | Yes |
| `POST` | `/internal/projects/{projectID}/users/{userID}/role` | Yes |
| `GET` | `/internal/config/{projectID}` | No (MVP gap: unauthenticated relative to the key; treat as trusted network only) |

Key source:

- Process env `INTERNAL_API_KEY` (required at process start; empty key prevents startup).

Header:

```http
X-Internal-Key: <same value as INTERNAL_API_KEY>
```

When required and missing or wrong:

- Status **401**
- Body: `{"error":"unauthorized"}` (via `http.Error`; Content-Type may be text/plain)

Promote role body:

```json
{ "role": "admin" }
```

Success for promote: **204** No Content.

Bootstrap success: **200** JSON with schema name, version, and whether migration was applied.

---

## 6. Bootstrap and schema version

`POST /internal/bootstrap/{projectID}` (with valid internal key):

1. Load project config from control-plane.
2. Read applied schema version from `{schema}._schema_meta` (missing schema/table → treated as version 0).
3. If applied version ≥ config version → no-op response (`applied: false`).
4. Otherwise `EnsureSchema`:
   - `CREATE SCHEMA IF NOT EXISTS`
   - `CREATE TABLE IF NOT EXISTS` per entity (columns from entity pool + system `created_at` / `updated_at`; `id` UUID if not in pool)
   - Upsert `_schema_meta.version` to config version

**MVP limits:**

- Table creation is **CREATE IF NOT EXISTS**, not a full additive `ALTER` migration engine.
- Adding new fields to an **existing** table is limited; do not assume arbitrary column adds are applied on every bootstrap.
- Schema name is per project (`SchemaName` from control-plane config).

Response shape (conceptual):

```json
{
  "SchemaName": "proj_...",
  "Version": 1,
  "Applied": true
}
```

(Exact JSON field casing follows the Go struct encoding used by the handler.)

---

## 7. Public HTTP surface (MVP)

| Method | Path | Auth |
|--------|------|------|
| `GET` | `/health` | None |
| `POST` | `/v1/auth/signup` | `X-Project-Id` |
| `POST` | `/v1/auth/signin` | `X-Project-Id` |
| `POST` | `/v1/auth/refresh` | `X-Project-Id` |
| `GET` | `/v1/auth/me` | `X-Project-Id` + Bearer |
| `POST` | `/v1/{entity}` | `X-Project-Id` + Bearer |
| `GET` | `/v1/{entity}` | `X-Project-Id` + Bearer |
| `GET` | `/v1/{entity}/{id}` | `X-Project-Id` + Bearer |
| `PATCH` | `/v1/{entity}/{id}` | `X-Project-Id` + Bearer |
| `DELETE` | `/v1/{entity}/{id}` | `X-Project-Id` + Bearer |
| `GET` | `/v1/me/config` | `X-Project-Id` (project config from context) |

Create success → **201**. Delete success → **204**.

---

## 8. Primary error codes

| HTTP | Typical cause |
|------|----------------|
| **400** | Missing `X-Project-Id`; invalid JSON; domain validation (`validation_error`) |
| **401** | Missing/invalid JWT; missing/wrong `X-Internal-Key` on protected internal routes; invalid credentials where mapped to unauthorized |
| **403** | Ownership denial (`forbidden`) |
| **404** | Unknown project (middleware); entity/record not found (`not_found`) |
| **409** | Conflict where mapped (e.g. email taken → `email_taken`) |
| **500** | Unhandled internal errors |
| **502** | Used on some internal config proxy failures when talking to control-plane |

Domain error codes (JSON `code` where handlers use structured domain errors) include: `unauthorized`, `forbidden`, `not_found`, `validation_error`, `invalid_credentials`, `email_taken`.

---

## 9. Explicit non-goals (MVP)

Out of scope for this contract version:

- Soft-delete and related visibility rules  
- n-n relations and shared ownership  
- Field-level permission beyond module `enabledFields` stripping  
- Multi-level ownership cascading  
- Anonymous public CRUD  
- Full additive schema migration / arbitrary `ALTER COLUMN` on every bootstrap  
- Runtime event consumers (e.g. RabbitMQ) as part of this contract  

---

## 10. Enforcement

- Implementations that violate this document or PERMISSION_CONTRACT are bugs, even if "it works".
- CI must keep a dedicated **runtime** job: `go vet` and unit tests under `services/runtime` (including permission and usecase coverage for ownership and CRUD field filtering).
- Integration ownership tests may run with `DATABASE_URL` and build tag `integration`; they are not required for every unit-only CI path.
