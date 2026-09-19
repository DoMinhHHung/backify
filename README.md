# Backify

**Rent-to-own Backend** for Vietnamese indie developers & freelancers.

Backify is a Backend-as-a-Product that lets frontend-focused developers assemble a complete backend through a wizard, get managed hosting, and later export plain-text code when they want full ownership.

## Core Promise

- **Wizard-based assembly** of backend modules (Auth, CRUD, …)
- **Ownership-based permission** derived automatically from relations — no need to write policies
- **Managed hosting** + **plain-text export** (no lock-in)

## Current Status

🚧 **Pre-MVP** — Documentation & architecture first.

### MVP Scope (locked)

- Auth (signup, signin, refresh, me)
- CRUD
- Relation field (1-n / n-1 only)
- Ownership-based permission (automatic)
- Simple Admin override (read-all)

**Explicitly out of MVP:** Storage, Notification, Payment, OAuth, soft-delete, n-n relations, field-level permission.

## Why docs before code?

Documentation is the source of truth. Code must follow the docs, never the other way around.

## Tech Stack (planned)

- Go + FastAPI (hybrid)
- PostgreSQL (schema-per-project)
- Redis

## License

Proprietary (for now).
