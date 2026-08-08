---
id: HLD-XXX
title: "[Subsystem Name] High-Level Design"
layer: Level 4 (Subsystem)
status: Approved
owner: Lead Subsystem Architect
references:
  - ARCHITECTURE.md
---

# HLD: [Subsystem Name]

## Purpose
Document the high-level architecture, module boundaries, sequence diagrams, and storage schemas for a subsystem.

## Scope
Entire subsystem component.

## Sections
- **1. Subsystem Architecture Overview**: High-level structural overview.
- **2. Component Interactions & Sequence Flows**: Visual sequence diagrams for key flows.
- **3. Interface & Contract Summary**: Summary of REST/gRPC APIs exposed.
- **4. Data Model & Storage Strategy**: DB tables, KV caching, message topics used.
- **5. Failure Modes & Resilience**: Fallback behavior during dependency outages.

## Cross References
- [System Architecture](file:///home/logan78/Desktop/flux/docs/ARCHITECTURE.md)
- [Database Schema](file:///home/logan78/Desktop/flux/database/)

## Acceptance Criteria
- [ ] Interfaces match machine-readable contracts in `api/` and `database/`.

## Navigation
[Overview](#purpose) | [Sequences](#sections) | [Storage](#sections)
