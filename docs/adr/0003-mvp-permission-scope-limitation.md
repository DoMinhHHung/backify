# ADR 0003: Strict Limitation of Permission Scope in MVP

## Status

Accepted

## Context

A full-featured permission system (soft-delete, n-n, field-level, multi-level ownership, shared resources, rich admin powers) is extremely complex and easy to get wrong. Security bugs in the permission layer are catastrophic.

## Decision

In MVP we deliberately support only:

- 1-n and n-1 relations
- Single-level ownership
- Hard delete
- Automatic ownership for normal users
- Admin read-all override only

We explicitly **do not** support in MVP:

- Soft-delete
- n-n relations
- Field-level permission
- Multi-level ownership cascading
- Shared ownership
- Admin write/delete on others’ data

## Consequences

### Positive
- Much higher chance of shipping a correct and trustworthy permission system
- Clear contract that users can understand
- Faster time-to-MVP

### Negative
- Some real-world apps will hit the ceiling quickly
- We must communicate the limitations very clearly (Permission Contract + onboarding)

### Follow-up
- Phase 2 will carefully extend the model with new ADRs and contract amendments
