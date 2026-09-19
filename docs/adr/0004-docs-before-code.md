# ADR 0004: Documentation is the Source of Truth

## Status

Accepted

## Context

In early-stage projects it is common for code to drift away from original intent. When the team is small (or solo), memory of “why” decisions were made disappears quickly.

## Decision

All significant product and technical decisions are written down **before** implementation.

Order of work:

1. Product vision & MVP scope
2. Permission Contract
3. Architecture Decision Records
4. Data model & API design
5. Code + tests that verify the contracts

If code and documentation diverge, **documentation wins**. Code must be updated to match the docs.

## Consequences

- Higher initial documentation overhead
- Much lower long-term confusion and rewriting
- Easier onboarding of future contributors
- Forces clarity of thinking before coding
