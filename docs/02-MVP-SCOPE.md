# MVP Scope (Locked)

This document is the **definition of done** for the first shippable version.

## In Scope

### 1. Auth Module
- signup
- signin
- refresh token
- me (current user)

### 2. CRUD Module
- create / read / update / delete for user-defined entities

### 3. Relation Field
- Support only **1-n** and **n-1**
- One level only (no deep nested ownership yet)
- Declared as a field type inside the Field Pool

### 4. Ownership-based Permission (Automatic)
- When a relation `Order.userId → User.id` is declared, the system automatically understands ownership
- A user can only read/update/delete records they own
- No policy writing required from the developer

### 5. Admin Override (Simple)
- Role `admin` can **read all** records
- Admin write/delete capabilities are **out of scope** for MVP

## Explicitly Out of Scope (MVP)

- Storage module
- Notification module
- Payment module
- OAuth / social login
- Soft-delete
- n-n relations
- Field-level permission
- Multi-level ownership (User → Order → OrderItem automatic cascading)
- Shared ownership / collaborative resources
- Custom functions / raw SQL escape hatch (will come later)

## Success Metrics for MVP

The following must be true and demonstrable:

1. Developer creates `Order` entity with relation `userId → User.id`
2. System automatically enforces ownership — developer writes **zero** policy lines
3. User A creates an Order → User B calling `GET /orders` only sees their own
4. User A trying to access User B’s order receives `403 Forbidden`
5. Admin calling `GET /orders` sees all orders

If any of the above fails, MVP is not done.
