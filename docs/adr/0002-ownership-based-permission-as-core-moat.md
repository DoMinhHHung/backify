# ADR 0002: Ownership-based Permission as Core Moat

## Status

Accepted

## Context

Most BaaS platforms (Supabase, Firebase, Appwrite) require developers to write security policies (RLS, Security Rules, etc.). This is the single biggest source of bugs and friction for indie developers.

Backify needs a clear technical differentiator that is hard to copy quickly.

## Decision

The core technical moat of Backify is **Ownership-based permission derived automatically from relations**.

- Developer only declares a relation (`Order.userId → User.id`).
- The system automatically enforces that a user can only access records they own.
- No policy language is exposed to the developer in MVP.

This feature is treated as **priority #1** and must be correct before any other modules (Storage, Notification, Payment) are added.

## Consequences

### Positive
- Strong, defensible differentiator
- Dramatically lower cognitive load for target users
- Clear success metrics (see Permission Contract)

### Negative / Risks
- Permission engine is complex to implement correctly
- Edge cases (soft-delete, shared resources, multi-level ownership) are deferred
- Must maintain a strict contract and test it rigorously

### Mitigation
- Extremely limited scope in MVP (see ADR 0003)
- Permission Contract is the source of truth
- Comprehensive automated tests for ownership rules before any release
