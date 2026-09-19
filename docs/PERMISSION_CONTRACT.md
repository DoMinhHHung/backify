# Permission Contract

> This is a **Systems Contract**.  
> It defines exactly what the permission system does and does **not** do in the current version.  
> Code must implement this contract. Any deviation is a bug.

**Status:** Active for MVP  
**Last updated:** 2026-09-20

---

## 1. Core Principle

Ownership is derived **automatically** from relations.

A developer only needs to declare:

```text
Order.userId  →  User.id   (relation, n-1)
```

The system then understands:

- An `Order` belongs to a `User`
- A normal user may only read / update / delete **their own** Orders
- An `admin` may read **all** Orders

**No policy language. No RLS. No security rules written by the developer.**

---

## 2. Supported Relation Types (MVP)

| Type   | Supported | Example                     |
|--------|-----------|-----------------------------|
| 1-n    | ✅ Yes    | User has many Orders        |
| n-1    | ✅ Yes    | Order belongs to one User   |
| n-n    | ❌ No     | (Phase 2)                   |
| Self   | ❌ No     | (not supported yet)         |

Only **one level** of ownership is supported in MVP.

---

## 3. Ownership Rules

### 3.1 Normal User

- Can `CREATE` a record and becomes its owner (via the relation field).
- Can `READ` only records they own.
- Can `UPDATE` only records they own.
- Can `DELETE` only records they own.
- Attempting to access a record they do not own → **403 Forbidden**.

### 3.2 Admin Role

- Can `READ` **all** records of any entity.
- Cannot `UPDATE` or `DELETE` other users’ records in MVP (simple override only).
- Admin capabilities beyond read-all are deferred to Phase 2.

### 3.3 Unauthenticated

- No access to any protected resource.

---

## 4. Explicit Non-Goals (MVP)

The following are **intentionally not supported**:

| Feature                    | Status   | Reason                              |
|---------------------------|----------|-------------------------------------|
| Soft-delete               | ❌       | Complexity + edge cases             |
| Shared ownership          | ❌       | Requires n-n + clear semantics      |
| Field-level permission    | ❌       | Significant added complexity        |
| Multi-level ownership     | ❌       | User → Order → OrderItem cascading  |
| Custom permission rules   | ❌       | Defeats the purpose of auto model   |
| Public / anonymous access | ❌       | Out of scope for first version      |

---

## 5. Behavior Guarantees

These must always hold:

1. Declaring a relation is sufficient to activate ownership rules.
2. Developer never writes a single line of policy / RLS / security rule.
3. `GET /orders` for User A returns only Orders owned by A.
4. `GET /orders/{id}` where the Order belongs to User B returns `403` for User A.
5. Admin `GET /orders` returns every Order.
6. Hard delete only (record is removed from the database).

---

## 6. How Relation is Declared (Conceptual)

```text
Entity: Order
Fields:
  - id          uuid          (system)
  - userId      relation      → User.id     ← this activates ownership
  - amount      number
  - status      enum
  - createdAt   timestamp     (system)
```

When the `userId` relation field is present, the permission engine treats the value of `userId` as the owner.

---

## 7. Future Extensions (Not in this Contract)

The following may be added later. They will receive their own contract amendments:

- Soft-delete + visibility rules
- n-n relations & shared resources
- Field-level permission
- Multi-level ownership cascading
- Admin write/delete powers
- Custom roles beyond `user` / `admin`

Until an official amendment is published, the rules in this document remain the only valid behavior.

---

## 8. Enforcement

Any code that violates this contract is considered a **bug**, even if the code “works”.

Permission tests must cover the guarantees in section 5 before any release.
