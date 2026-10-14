---
phase: 3
slug: fast-managed-domain-redirects
status: approved
review_method: inline
reviewed_at: 2026-10-14
shadcn_initialized: false
preset: none
created: 2026-10-14
---

# Phase 3 UI design contract

Scope: LINK-04 edit destination/title and honest redirect availability copy. Preserve Phase2's shell, workspace behavior, native controls and all existing accessibility requirements. No redesign, new UI dependency, analytics page or click-delivery claim. Defaults selected under --auto; no new user-locked decisions.

## Design system

Reuse existing semantic React/Next components and plain CSS. No component registry, shadcn initialization, icons, media, new fonts or assets. Font system-ui. Existing light palette and forced-colors support remain.

## Spacing scale

Use existing4/8/16/24/32/48/64px tokens only. Inline edit form max640px, 16px field gaps and24px section inset; full width on mobile. Inputs/buttons >=44px tall. Existing4px control and8px panel radii. No exceptions.

## Typography

Existing body16px/400/1.5, labels14px/600/1.5, section heading20px/600/1.2, page heading28px/600/1.2. Inputs stay16px. Full destination wraps and is readable without hover; title is escaped text. Keep four sizes and two weights.

## Color

Existing surface #FFFFFF, secondary #F1F5F9, accent #1D4ED8, main text #0F172A, muted #475569, interactive border #64748B, destructive/error #B91C1C. Accent only primary action/current navigation/focus. Status and errors always include text. Focus2px outline with2px offset. Contrast text >=4.5:1 and controls/focus >=3:1.

## Interaction and copywriting

| Element/state | Contract |
|---|---|
| Entry | Detail action `Edit link`; available only when current authorized write capability permits and link is active/disabled, unsuspended. Hide forbidden actions, keep read-only destination/title. |
| Form | Inline region heading `Edit link`; required `Destination URL`, optional `Title`; URL helper `Use a public HTTP or HTTPS URL.` and title limit200. Read-only short URL remains visible. No host/key editor. |
| Initial focus | Destination input; on Cancel or successful Save restore focus to Edit link. No modal for ordinary editing. |
| Actions | `Save changes`, `Cancel`; submit busy `Saving changes…`, region aria-busy. No double submit. Cancel never silently discards an unknown commit result. |
| Success | Polite status `Link updated. Redirects may take up to 5 seconds to reflect changes.` Render the returned canonical version/title/destination. Never claim tracking occurred. |
| Syntax/safety | Reuse `Enter a valid HTTP or HTTPS URL.` / `This destination isn't allowed. Choose a different public destination.` Associate field errors via aria-describedby, focus first invalid field. |
| Conflict | Alert `This link changed since you opened it. Review the latest details before trying again.` Action `Review latest details`; preserve draft separately, require explicit review and a new intent/version before resubmit. |
| Unknown result | `We couldn't confirm whether your changes were saved. Retry the same changes to check the result.` Action `Retry save`; exact original body, version and idempotency key. Do not silently mutate intent while retrying. |
| Access loss | Existing workspace chooser/access-changed behavior; immediately clear protected details, drafts and pending attempts. Late responses cannot render in another session/workspace. |
| Discard/switch | Reuse unsaved-change handling; cancel preserves current workspace/draft, confirm clears it before new tenant data renders. |
| Availability helper | Replace stale Phase2 copy in create/library/detail with `Managed-domain redirects are available. Click tracking and analytics are not available yet.` Only ship after actual redirect acceptance; Active remains lifecycle, not an uptime promise. |
| Empty state | Preserve `No links yet` and `Create your first link`; no added marketing/analytics CTA. |

Enter submits the form, Escape cancels only a known-unsaved draft via existing navigation rules. At320px, 200% text and400% zoom fields/actions reflow, long URLs wrap, no page horizontal scroll. All controls keyboard accessible with visible focus. Browser tests cover success, conflict, lost response/retry, access loss, unsafe URL, text escaping and mobile reflow. Two workspace tests use real Go authorization.

## Registry safety

No registries or blocks used; no network scaffolder. Existing components only.

## Checker sign-off

Inline six-dimension review, not an independent-agent review: copywriting PASS (all required states/actions), visuals PASS (existing design and320px reflow), color PASS (inherited explicit contrast contract), typography PASS (four sizes/two weights), spacing PASS (existing scale), registry safety PASS (none). These check the written contract; rendered accessibility remains an execution test. Approved for planning2026-10-14.
