---
phase: 2
slug: tenant-safe-link-control-plane
status: approved
reviewed_at: "2026-10-08T11:11:59Z"
shadcn_initialized: false
preset: none
created: 2026-10-08
---

# Phase 2 — UI Design Contract

> Visual and interaction source of truth for planning and execution. Independent UI checking approved all six dimensions after one copywriting revision.

## Sources and Scope

Locked decisions D-01–D-04 come from `02-CONTEXT.md`. Role, invitation, short-key, lifecycle, pagination and conflict policies come from `02-RESEARCH.md`. All visual choices below are defaults selected under explicit delegated discretion and `--auto`; they are not additional user locks. Scope covers TEN-01–TEN-08, LINK-01–LINK-03, LINK-05–LINK-09 and SAFE-02–SAFE-04, and all five Phase 2 roadmap success criteria.

Deliver authentication, named workspace onboarding, selection/switching, Links create/detail/library/lifecycle/suspension and owner/admin Team management. Omit navigation or controls for analytics, redirect testing, custom domains, billing, API keys and enterprise capabilities. No destination/title editing (LINK-04), preview scraping, hard deletion or compulsory team wizard. A configured short URL is a stable identifier; this phase does not establish redirect availability.

## Design System

| Property | Value |
|----------|-------|
| Tool | none; manual reusable React components and CSS custom properties |
| Preset | not applicable |
| Component library | none; semantic HTML, native select and native dialog |
| Icon library | none; original local 16/20px SVG outlines, currentColor, 2px stroke |
| Font | system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif |
| Presentation | Next App Router/TypeScript; exact dependencies follow research and frozen compatibility checks |
| Authentication | Clerk prebuilt SignIn/SignUp/recovery; supported appearance customization only |

Scout on 2026-10-08 found empty `apps/frontend`, no `components.json`, frontend tokens, Tailwind/PostCSS configuration or local project skills. shadcn initialization gate disposition: **choose no initialization**, under auto delegation and the research's explicit plain-CSS recommendation. No scaffolder, registry blocks or new UI dependencies enter this contract. Preserve root React graph and existing generation/privacy/tool gates. Go owns product authorization and transactions; provider organizations never supply workspace permissions.

Original identity: restrained neutral surfaces, deep blue action accent, compact bordered lists, typographic hierarchy and the temporary text wordmark “Flux.” No borrowed product layouts/assets, hero imagery, gradients, decorative illustrations or fabricated metrics. Light theme only for this slice; native forced-colors support is mandatory.

Reusable inventory: AppShell, PageHeader, Button, Field, Select, StatusBadge, InlineNotice, EmptyState, LoadingRows, ConfirmDialog, WorkspaceChooser, LinkList, LinkDetail, TeamList and InvitationPanel. Share validation, busy, focus and error patterns across screens. Four button styles only: primary, secondary, text, destructive. Two surface styles: flat bordered content and modal. One badge style with text; no per-screen aesthetic variants.

## Spacing Scale

| Token | Value | Usage |
|-------|-------|-------|
| xs | 4px | Inline icon/text gap |
| sm | 8px | Label/help gap, row vertical inset |
| md | 16px | Form gaps, compact section inset |
| lg | 24px | Desktop page inset, panel padding |
| xl | 32px | Section separation |
| 2xl | 48px | Auth and empty-state breathing room |
| 3xl | 64px | Header height, major layout spacing |

Spacing exceptions: none. Dimensions are distinct from spacing: interactive targets at least 44×44px, input/button height 44px, desktop row minimum 64px, border 1px, focus outline 2px with 2px offset. Allowed radii: 4px controls/badges and 8px panels/dialogs only. One modal shadow: `0 8px 24px rgb(15 23 42 / 16%)`; ordinary cards have no shadow.

## Typography

| Role | Size | Weight | Line Height |
|------|------|--------|-------------|
| Body | 16px | 400 | 1.5 |
| Label/metadata | 14px | 600 labels, 400 metadata | 1.5 |
| Heading | 20px | 600 | 1.2 |
| Display/page title | 28px | 600 | 1.2 |

Exactly four sizes and two weights (400/600) for original product UI. No 12px metadata or uppercase tracking. Inputs remain 16px on mobile. Long URLs wrap in details; list truncation retains the full accessible text and does not hide the only action. Full destination/title are available in detail without hover. Provider internals use supported token overrides where available; document any unavoidable provider variation and verify accessibility independently. Do not rewrite authentication to satisfy a cosmetic token count.

## Color

| Role | Value | Usage |
|------|-------|-------|
| Dominant (60%) | #FFFFFF | Main surface, form controls, dialogs |
| Secondary (30%) | #F1F5F9 | Shell, table headers, neutral notices |
| Accent (10% maximum) | #1D4ED8 | Primary CTA, selected navigation marker, focus indicator |
| Destructive | #B91C1C | Destructive action and error text/border only |
| Main text | #0F172A | Titles and body |
| Muted text | #475569 | Metadata/help; never opacity-reduced |
| Interactive border | #64748B | Input and secondary-button boundaries |
| Decorative divider | #CBD5E1 | Noninteractive row separators only |

Accent reserved for primary CTA, current-page marker and keyboard focus. Secondary actions/ordinary links use main text; links are underlined. White text on accent is approximately 6.7:1; white on destructive 6.5:1; main text on white 17.9:1; muted on white 7.6:1 and secondary 6.9:1. Interactive border on white/secondary exceeds 4:1. Verify computed states with WCAG AA targets: normal text ≥4.5:1, controls/focus ≥3:1. Hover primary #1E40AF; hover destructive #991B1B. Disabled controls retain readable text and a textual reason. Status colors remain neutral: Active, Disabled, Archived, Deleted and Suspended always appear as words. No green/red-only encoding. The 60/30/10 split describes approximate visible area, not data state proportions.

## Layout and Navigation

Desktop ≥1024px: 224px sidebar with workspace selector, Links and authorized Team entry; account/signout control at bottom. Main content max-width 1200px, 24px inset; 64px header. At 768–1023px use a top workspace/account bar and horizontal Links/Team navigation. Below 768px use stacked header/navigation, 16px page inset, full-width CTA/forms and single-column detail. Workspace names wrap; labels never rely on tooltips.

Auth/onboarding/chooser use a centered max-width 448px panel with 24px inset; panel fills small screens. Create link is a dedicated page with max-width 640px form, not a cramped popover. Desktop library is a semantic table with Link/title, Destination, State, Created and Actions; mobile renders the same authorized records as labeled cards with a detail anchor. Do not make an entire row a nested button. Details use a definition list and action group. Team tables likewise become labeled mobile cards. Only wide data tables may scroll within a named region; the page itself must fit 320px. At 200% text and 400% zoom content reflows and dialogs remain scrollable above mobile keyboards.

## Authentication and Onboarding (D-01–D-04)

**D-01:** Login/register embed Clerk prebuilt email/password components with Google and GitHub visible. Email verification and “Forgot password?” recovery remain reachable by keyboard; preserve provider validation/CAPTCHA/verification sequencing. Original wrapper headings: “Sign in to Flux” / “Create your account.” Use provider-supported localization for visible labels without bypassing its flow. Google/GitHub branding stays inside the provider's authorized controls. No email-link-only substitute, fake test login or account existence disclosure.

**D-02:** After verified authentication, a new user without invitations/workspaces sees “Create your workspace,” required “Workspace name,” helper “Choose a name your team will recognize,” and CTA “Create workspace.” Submit once with durable bootstrap idempotency; loading label “Creating workspace…”; navigate only after the API commits both workspace and owner. Validation: “Enter a workspace name.” Preserve typed name on safe failure. Never invent a personal workspace. A valid invitation instead leads to explicit review/acceptance; no unrelated workspace creation is required.

**D-03:** Bootstrap restores last-used workspace only after current Go authorization. Otherwise show “Choose a workspace” with only authorized names, roles and “Open workspace” actions, plus “Create workspace.” No workspaces: “You don't have a workspace yet. Create one or accept an invitation.” Switching clears tenant caches, selections, search, cursors, open dialogs and draft forms after an unsaved-change prompt; cancel outstanding requests and ignore late responses. Commit authorized selection before rendering new data. On membership loss, immediately clear old content and show chooser with “Your workspace access changed. Choose an available workspace.” Revalidate on tab focus and handle API denials; browser persistence cannot authorize access.

**D-04:** Successful creation opens Links directly, heading “No links yet,” body “Create a managed-domain link to organize its destination and status,” prominent CTA **“Create your first link.”** Invitation setup stays optional. Persistent unobtrusive helper on create/detail: “Link management is available. Redirects and analytics are not available yet.” Active means stored lifecycle, not demonstrated routing health.

## Link Library, Creation and Detail

Library heading “Links”; allowed users get “Create link,” viewers get “You have view-only access. Ask an owner or admin to change your role.” Search label “Search links,” helper “Search short keys, titles, or destinations,” explicit “Search links” submit and “Clear search.” No global character-key shortcut that intercepts typing. Filter native select: All nondeleted (default), Active, Disabled, Archived, Deleted. Reset cursors on changed search/filter. Retain stable newest-first ordering from research. Load 25 rows per page, use opaque cursor “Next page” and locally remembered “Previous page”; no guessed totals/page numbers. Disable Next only on authoritative exhaustion, announce “No more links.” Cursor failure: “This page is no longer available. Return to the first page.” CTA “Return to first page.”

Create fields: required “Destination URL” (`type=url`), optional “Title” (maximum 200 characters), optional “Custom short key.” Show server-configured managed hostname as immutable prefix, never an invented domain. Helper: “Leave blank to generate a key. Custom keys use 3–64 lowercase letters, numbers, hyphens or underscores and start with a letter or number.” Lowercase candidate visibly before submit; reject invalid characters rather than silently removing them. Client checks aid feedback; server validation wins. No availability ownership clues. Creation does not fetch the destination or render remote favicons.

Submit “Create link” → “Creating link…”; preserve one idempotency key/payload for uncertain retries, rotate only for deliberate changed submissions. Success opens committed detail and announces “Link created.” Detail includes canonical short URL (plain text with “Copy short URL”), destination, title or “Untitled link,” creator, created/updated timestamps, lifecycle, suspension/effective state and version. Display timestamps in local timezone with full date/time and timezone available as text; preserve exact machine value in `time` elements. No destination/title Edit control or “Visit/test redirect” CTA. Render all values as escaped text; external destination anchors use safe HTTP(S), `rel=noreferrer noopener`, and a visible “opens in a new tab” label if applicable.

## Lifecycle, Suspension and Permissions

| Stored lifecycle | Allowed ordinary actions | Result |
|------------------|--------------------------|--------|
| Active | Disable link, Archive link, Delete link | Disabled, Archived, Deleted respectively |
| Disabled | Enable link, Archive link, Delete link | Active, Archived, Deleted respectively |
| Archived | Restore link, Delete link | Disabled, Deleted respectively |
| Deleted | Restore link | Disabled; key stays reserved |

All roles read links/switch workspace; owner/admin/member create and perform ordinary lifecycle; viewer cannot mutate. Owner/admin alone suspend/clear suspension and list/manage Team. Members/viewers receive safe forbidden direct-route states. Never use hidden controls as authorization evidence.

Suspension is an independent restriction. Show both “Lifecycle: Active” and “Effective state: Suspended” when applicable; precedence is Deleted → Suspended → Archived → Disabled → Active, while all source state remains visible. Suspended detail shows recorded actor, reason and timestamp to authorized link readers. Disable “Enable link” with helper “An owner or admin must clear suspension first.” Other ordinary transitions never clear it. Owner/admin “Suspend link” requires nonblank reason; “Clear suspension” requires reason and leaves lifecycle Disabled before any explicit “Enable link.” Deleted records have no suspension mutation controls. Do not claim the control plane routes traffic.

Every mutation submits the displayed expected version and waits for committed response. On conflict refresh detail, announce stale-state copy, preserve draft reason, and require explicit review/resubmit with new version; never automatically repeat a destructive action. If a network result is uncertain, retry identical request safely rather than showing success. Refresh permission affordances after every committed change or authorization denial.

## Team and Invitations

Team is visible to owners/admins only; show members with identity, role and allowed actions plus Invitations section. Admins invite/change/remove members/viewers only; owners manage all roles, invite admins, and promote an existing member to owner. Invitations offer admin/member/viewer, never owner; default member. Role select options reflect actor/target capability. Final owner's demote/remove controls show “Promote another owner first”; server race denial uses the same copy. Ownership transfer is promote then demote, not an unimplemented single-step flow.

Invite form: “Email address,” “Role,” CTA “Invite member”; confirmation says “Invitation queued” until actual delivery. Seven-day expiry appears as full timestamp. States: Queued, Delivered, Delivery failed, Accepted, Expired, Revoked. Authorized admins/owners can resend permitted invitations; resend rotates token and says “A new invitation will replace the previous one.” Never display/copy raw invitation token in Team. Duplicate pending email: “An invitation is already pending for this email. Resend or revoke it.” Failed delivery: “Invitation delivery failed. Resend the invitation to try again.” Queued remains truthful during worker/provider delays.

Invitation entry keeps the secret out of logs/referrers, scrubs the arrival URL promptly and uses only a sanitized internal continuation through auth. No token in localStorage or arbitrary return URLs. After authentication, show only server-authorized workspace/role and CTA “Accept invitation”; acceptance requires a verified matching recipient and unexpired/unrevoked invitation. Wrong recipient receives no workspace/member details. Repeated acceptance by the same recipient opens authorized Links without changing their role. Never auto-accept on page load.

## Copywriting Contract

| Element/state | Exact product copy / action |
|---------------|----------------------------|
| Primary CTA | “Create link”; first workspace Links empty CTA “Create your first link” |
| No search results | “No matching links” / “Try a different search or clear your filters.” / “Clear filters” |
| Deleted list empty | “No deleted links” / “Deleted links appear here and can be restored.” |
| Team invitations empty | “No pending invitations” / “Invite a member to collaborate in this workspace.” |
| Loading | “Loading links…” / “Loading workspaces…” / “Loading team…” |
| List failure | “We couldn't load links. Try again.” / “Retry loading links”; team/workspaces use “We couldn't load team. Try again.” / “Retry loading team” and “We couldn't load workspaces. Try again.” / “Retry loading workspaces” |
| Mutation failure | “We couldn't save this change. Try again.” / “Retry change” |
| Uncertain create | “We couldn't confirm link creation. Retry to check the same request.” / “Retry creation” |
| Invalid destination | “Enter a valid HTTP or HTTPS URL.” |
| Blocked destination | “This destination isn't allowed. Choose a different public destination.” |
| Key conflict | “This short key is unavailable. Choose another key or generate one.” |
| Version conflict | “This link changed since you opened it. Review the latest details before trying again.” / “Review latest details” |
| Request reuse conflict | “This request changed. Review your fields and submit a new request.” |
| Missing precondition | “We couldn't verify the latest version. Reload this link before trying again.” / “Reload link” |
| Session expiry | “Your session ended. Sign in to continue.” / “Sign in” |
| Auth unavailable | “Sign-in is temporarily unavailable. Try again shortly.” / “Retry sign-in” |
| Recovery wrapper | “Reset your password” / “If this address is eligible, follow the recovery instructions to continue.” |
| Verification wrapper | “Verify your email” / “Complete email verification to continue.” |
| Forbidden | “You don't have permission to do this. Ask a workspace owner or admin for help.” / “Back to links” when authorized |
| Foreign/missing link | “Link not found” / “Choose a link from your workspace.” / “Back to links” |
| Expired/revoked invitation | “This invitation is no longer valid. Ask an owner or admin for a new invitation.” |
| Wrong recipient | “This invitation can't be accepted with this account. Sign in with the invited, verified email address.” / “Switch account” |
| Already accepted | “Invitation already accepted.” / “Open workspace” after authorization |
| Rate limit | “Too many requests. Wait before trying again.”; honor server retry interval |
| Copy success/failure | “Short URL copied.” / “Couldn't copy. Select and copy the short URL below.” |

Provider-owned verification/recovery inner text remains supported provider localization, with these original headings and safe wrapper states. Never interpolate raw errors, account existence, foreign identifiers or provider diagnostics. An optional support reference uses only the safe request ID.

Destructive confirmations use native modal dialog, named link/member/invitation from authorized data, contextual safe-action button first and specific final verb. Delete: “Delete link?” / “This link moves to Deleted. Its short key stays reserved. You can restore it as disabled.” / “Keep link” / “Delete link.” Archive: “Archive link?” / “Restoring this link will leave it disabled.” / “Keep link” / “Archive link.” Suspend: “Suspend link?” / “This restriction remains until an owner or admin clears it. Enter a reason.” / “Keep link” / “Suspend link.” Clear: “Clear suspension?” / “This link will remain disabled until explicitly enabled. Enter a reason.” / “Keep suspension” / “Clear suspension.” Remove: “Remove member?” / “This member will lose access to this workspace.” / “Keep member” / “Remove member.” Demote owner/admin: “Change role?” / “This change removes the member's current management permissions.” / “Keep current role” / “Change role.” Revoke: “Revoke invitation?” / “The recipient will no longer be able to accept this invitation.” / “Keep invitation” / “Revoke invitation.” Unsaved switch: “Discard unsaved changes?” / “Your current form will be cleared when you switch workspaces.” / “Stay in workspace” / “Discard and switch.” No typed-name ritual; safe action never commits.

## Accessibility, Feedback and Motion

Use header/nav/main landmarks, one h1, skip link and `aria-current=page`. Buttons perform actions; anchors navigate. Forms have persistent labels, correct autocomplete, described hints, `aria-invalid` and error summary focused after failed submission. Tab order follows visual order; no positive tabindex. Native selects use native keyboard semantics; dialogs use `showModal`, contain focus, support Escape dismissal, focus the contextual safe-action button initially for destructive actions and restore trigger focus on close. Route transitions focus main heading; switching after revoked access focuses chooser heading.

Use `role=status`/polite announcements for committed changes, copy, page results and loading completion; assertive error summary only for blocking submission failure. Mark loading region `aria-busy`; skeleton rows are inert/aria-hidden and accompanied by textual status. Do not replace an unconfirmed operation with a success toast. Preserve existing same-workspace data on refresh failure with an inline error, but immediately remove data after signout/access loss. Loading skeletons use stable geometry and no shimmer. Delayed loading shows “This is taking longer than expected. You can retry.” with cancellation available and contextual “Retry loading links,” “Retry loading team” or “Retry loading workspaces” action; no automatic duplicate submission.

Focus outline remains visible above sticky headers. Honor reduced motion: no positional transitions or smooth scrolling. Default feedback may use 120ms color/opacity transition only, disabled under reduced motion. No custom drag, swipe, animation or keyboard interception is required. Controls remain usable by touch, keyboard and screen reader; do not rely on hover or color alone.

## Registry Safety

| Registry | Blocks Used | Safety Gate |
|----------|-------------|-------------|
| shadcn official | none | Not applicable; no initialization or imported blocks; checked 2026-10-08 |
| Third-party registries | none | Not applicable; no blocks declared or installed; checked 2026-10-08 |

Clerk is the selected authenticated provider SDK, not a component registry. Use only research-audited exact dependencies and normal dependency scans; preserve provider security UI. Context7 official Clerk docs checked 2026-10-08 confirm supported appearance customization for prebuilt components: [customization overview](https://clerk.com/docs/guides/customizing-clerk/overview). Strategy/instance configuration and real Google/GitHub/email/recovery acceptance remain execution prerequisites, not reasons to block this contract.

## Testable Observables and Quality Gates

Browser acceptance covers D-01 real provider signup/verification/recovery/social auth; D-02 one named workspace plus owner committed, safe retry creates one workspace; D-03 authorized restore/switch and removal across tabs without stale data/identifiers; D-04 first Links empty CTA without mandatory invitations. Provider setup shortcuts do not count as verification/recovery proof.

With real Go/PostgreSQL, prove role-dependent controls and direct-route denials, invitation recipient/expiry/revocation/repeat/delivery states, last-owner race denial, unsafe URL and normalized key conflict feedback, safe create retry, search/reset/cursor boundaries, every lifecycle edge and suspension precedence, clear-to-disabled and stale-version review. Two-workspace direct IDs/cursors and late responses never expose foreign values. No executable analytics behavior is claimed.

Rendered checks at 320/768/1280px, keyboard-only, zoom, reduced motion and screen reader announcements verify labels, focus return, blocked actions, no layout overflow and computed contrast. Screen-level tests use observable semantics/text, not copied CSS snapshots. Execute actual frontend format/lint/typecheck/test/build and register it in root gates; preserve deterministic authored/generated API agreement and existing backend/tool/scanner checks. No UI logs, traces, analytics or console dumps may contain form destinations, titles, names, emails, invitation secrets, credentials, raw errors or response bodies; audit actor/reason data stays in protected authorized records.

## Checker Sign-Off

- [ ] Dimension 1 Copywriting: PASS
- [ ] Dimension 2 Visuals: PASS
- [ ] Dimension 3 Color: PASS
- [ ] Dimension 4 Typography: PASS
- [ ] Dimension 5 Spacing: PASS
- [ ] Dimension 6 Registry Safety: PASS

**Approval:** pending
