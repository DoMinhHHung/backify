# Product Vision — Backify

## Positioning

**Backify = Rent-to-own Backend** for Vietnamese indie developers and freelancers who know frontend but do not want to set up and operate a backend.

It is **not**:
- A pure CRUD BaaS (Supabase clone)
- A pure code generator
- A pure managed hosting platform (Railway clone)

It **is**:
- A wizard that lets developers **assemble** a backend from ready-made modules
- Managed hosting so they don’t have to deal with ops
- Plain-text export path so they can fully own the code later (no lock-in)

## Core Value Proposition

1. **Wizard + Field Pool + Modules** → Assemble backend in minutes instead of days
2. **Ownership-based permission automatically derived from relations** → No need to write policies (the real moat)
3. **Managed hosting + plain-text export** → Rent now, own later

## Target User

- Vietnamese indie developers & freelancers
- Strong on frontend, weak/unwilling on backend ops & security rules
- Building MVPs and early-stage commercial products
- Price sensitivity: $7–25/month range

## Long-term Module Roadmap

| Phase   | Modules                                      |
|---------|----------------------------------------------|
| MVP     | Auth + CRUD + Relation + Ownership Permission + Admin override |
| Phase 2 | Storage, Notification, OAuth, n-n relations, field-level permission |
| Phase 3 | Payment (payOS, VNPay, MoMo, Stripe)         |

## Non-goals (for now)

- Competing with Supabase on every feature
- Supporting complex multi-tenant enterprise permission models in MVP
- Building a full no-code visual builder
