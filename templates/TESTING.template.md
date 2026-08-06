---
id: TEMPLATE-TESTING
title: Testing Strategy & Quality Assurance
layer: Level 2 (Quality & Policy)
status: Active
owner: QA / Test Architect
references:
  - ai/CODING_STANDARDS.md
---

# [System Name] — Testing Strategy & Guidelines

## Purpose
Define testing standards across Unit, Integration, End-to-End (E2E), and Benchmark testing, enforcing Test-Driven Development (TDD).

## Scope
Repository-wide test suite.

## Sections
- **1. Test Pyramid & Coverage Goals**: Unit (80%+), Integration, E2E targets.
- **2. Test-Driven Development (TDD) Workflow**: Failing test ➔ Minimal code ➔ Pass ➔ Refactor.
- **3. Mocking & Fixtures Strategy**: Rules for external dependency mocking.
- **4. Integration & Database Testing**: Real DB testing with testcontainers.
- **5. Benchmark Testing**: Performance regression testing rules.

## Cross References
- [Coding Standards](file:///home/logan78/Desktop/plan/ai/CODING_STANDARDS.md)
- [Performance Specs](file:///home/logan78/Desktop/plan/ai/PERFORMANCE.md)

## Acceptance Criteria
- [ ] All code changes must include corresponding unit/integration tests.

## Navigation
[TDD Workflow](#purpose) | [Integration](#sections) | [Coverage](#sections)
