---
id: TEMPLATE-CODING-STANDARDS
title: Code Quality & Engineering Conventions
layer: Level 2 (Quality & Policy)
status: Active
owner: Staff Engineer
references:
  - ai/TESTING.md
---

# [System Name] — Coding Standards & Guidelines

## Purpose
Enforce uniform coding conventions, type-safety standards, error-handling patterns, and linter rules across the entire codebase.

## Scope
All source code files in all languages used in the repository.

## Sections
- **1. Core Engineering Principles**: Simplicity, readability, immutability, DRY.
- **2. Language-Specific Guidelines**: Conventions for Go, TypeScript, SQL, etc.
- **3. Naming Conventions**: Files, functions, variables, interfaces, packages.
- **4. Error Handling & Logging**: Error Wrapping, logging context, exception rules.
- **5. Automated Linter Configurations**: Tooling rules (`golangci-lint`, `eslint`).

## Cross References
- [Testing Strategy](file:///home/logan78/Desktop/plan/ai/TESTING.md)
- [Architecture](file:///home/logan78/Desktop/plan/docs/ARCHITECTURE.md)

## Acceptance Criteria
- [ ] Code passes all automated linters with 0 warnings.

## Navigation
[Principles](#purpose) | [Naming](#sections) | [Errors](#sections)
