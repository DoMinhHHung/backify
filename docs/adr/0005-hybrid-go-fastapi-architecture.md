# ADR 0005: Hybrid Go + FastAPI Architecture (Initial Direction)

## Status

Proposed (to be confirmed after deeper design)

## Context

We need a backend that is:

- Fast and efficient for the permission & CRUD core
- Easy to iterate on for the wizard / control plane / API gateway
- Able to support schema-per-project isolation

## Decision (Direction)

We will explore a hybrid approach:

- **Go** for the performance-critical core (permission engine, CRUD execution, multi-tenant schema handling)
- **FastAPI (Python)** for the control plane, wizard API, and developer-facing orchestration layer

Final service boundaries will be decided after the data model and permission engine design are complete.

## Consequences

- Two languages increase operational complexity
- Allows using the best tool for each job
- Requires clear API contracts between the Go core and the FastAPI layer

This ADR will be updated or superseded once the concrete service boundaries are decided.
