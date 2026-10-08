# Phase 2: Tenant-Safe Link Control Plane - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in 02-CONTEXT.md; this log preserves the alternatives considered.

**Date:** 2026-10-08
**Phase:** 2-Tenant-Safe Link Control Plane
**Areas discussed:** Sign-in and onboarding

## Area Selection

The user selected area 1, “Sign-in and onboarding,” from sign-in/onboarding, team permissions/invitations, short-link creation, and link library/lifecycle. Existing requirements and prior foundation decisions were carried forward.

## Email Sign-in

| Option | Description | Selected |
|--------|-------------|----------|
| Email + password | Email verification and password recovery | Yes |
| Email magic link | Open an emailed link to sign in | |
| You decide | Leave the email option to research and planning | |

**User's choice:** `1`
**Notes:** Google and GitHub sign-in were carried forward from `spec.md` §22. Current Clerk documentation confirms both email options; the user chose the password flow.

## First Workspace

| Option | Description | Selected |
|--------|-------------|----------|
| Ask for a workspace name | Explicit creation with owner membership | Yes |
| Create automatically | Default name that can be changed later | |
| You decide | Choose during planning | |

**User's choice:** `1. Ask for a workspace name (recommended)`
**Notes:** Invitation entry was presented as joining the invited workspace rather than forcing unrelated workspace creation. Valid acceptance and authorization still apply.

## Returning User Destination

| Option | Description | Selected |
|--------|-------------|----------|
| Last-used workspace | Only if still a member; otherwise show available workspaces | Yes |
| Workspace chooser every time | Select before entering the dashboard | |
| You decide | Choose during planning | |

**User's choice:** `1. Last-used workspace (recommended`
**Notes:** The missing closing punctuation did not change the selected option. Membership checks remain authoritative.

## New Workspace Landing Page

| Option | Description | Selected |
|--------|-------------|----------|
| Links page | Empty state with prominent “Create your first link” | Yes |
| Team setup | Invite teammates first, with an option to skip to links | |
| You decide | Choose during planning | |

**User's choice:** `1. Links page (recommended)`

## Completion

| Option | Description | Selected |
|--------|-------------|----------|
| Capture context | Leave other areas to research and planning | Yes |
| Discuss sign-in further | Continue the selected area | |
| Discuss another area | Teams, link creation or lifecycle | |

**User's choice:** `1`

## the agent's Discretion

The user explicitly approved capturing the four decisions with the other areas left to research and planning. No additional product policies are represented as user-selected locks. Required workspace/team/link/security capabilities remain mandatory.

## Deferred Ideas

None — discussion stayed within phase scope.
