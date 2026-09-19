# Why Documentation Comes Before Code

> "Documentation is the source of truth. Code must follow the docs, not the other way around."

## Principles

1. **Docs define reality**  
   Every architectural decision, every permission rule, every scope boundary lives in documentation first.

2. **Code is an implementation of the docs**  
   If code and docs diverge, the docs win. Code must be changed to match the docs.

3. **Contracts protect the team**  
   Clear contracts (especially Permission Contract) prevent misunderstanding, reduce support load, and protect against future scope creep.

4. **Onboarding & future contributors**  
   New people (including future-you) should be able to understand the system by reading docs, not by reverse-engineering code.

## Order of work in this repository

1. Core documentation & contracts
2. Architecture Decision Records (ADRs)
3. Data model & API design
4. Implementation
5. Tests that verify the contracts

This order is non-negotiable for the MVP phase.
