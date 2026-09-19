# ADR 0001: Record Architecture Decisions

## Status

Accepted

## Context

We need a lightweight way to record important architectural decisions so that the reasoning is not lost over time.

## Decision

We will use Architecture Decision Records (ADRs) stored in `docs/adr/`.

Each ADR follows a simple template:

- Title
- Status (Proposed / Accepted / Deprecated / Superseded)
- Context
- Decision
- Consequences

## Consequences

- Future contributors (including future-us) can understand *why* a decision was made.
- Changing a major decision requires a new ADR that supersedes the old one.
