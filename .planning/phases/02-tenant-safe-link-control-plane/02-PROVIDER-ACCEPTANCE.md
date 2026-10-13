# Phase 2 Provider Acceptance — Live-Provider Checkpoint

Status: **prerequisites verified; automated acceptance in progress**. Phase 2 remains incomplete at plan 35, task 2.

Plans 01–34 are complete and committed. This checkpoint is not a native successful provider execution record. It does not satisfy the automated or strict acceptance validator.

## Verified setup

- The real Clerk development instance accepted the supplied credentials. Backend and issuer JWKS matched.
- Real password sign-in completed in the operator's browser after its Device Trust email challenge. The configured instance requires signup email verification.
- Google was already enabled. GitHub sign-up/sign-in was enabled through the authorized Clerk Dashboard with shared development credentials and appeared in the real Flux UI.
- Resend accepted the authorized development delivery probe. The message was observed in the authorized inbox. The development sender is configured privately; production domain/sender validation remains separate.
- `provider:check` passed with coherent private configuration, authorized distinct account identities, invitation encryption keys, link settings and an HTTPS development invitation authority.
- Disposable pinned PostgreSQL/Redis and the production Go API started. API readiness returned HTTP 200. Run-owned resources were cleaned up afterward.

## Corrections and current execution

The live attempt exposed Clerk UI selector differences, Docker internal-network port-publishing behavior, and a mismatch between the browser origin precheck and Go's invitation hostname policy. Corrections retain production authentication and tenant isolation.

The focused native password-login test completed the real Device Trust challenge and received HTTP 200 from the production Go identity endpoint. Its authorized private inbox handoff consumed an actual emailed code through Clerk's UI. Clerk automatically submits the complete code; the redundant Continue click was removed. The focused run correctly failed full-proof promotion because it selected only one test. This result cannot satisfy the complete provider acceptance gate.

The full provider suite has no verified passing result. The last full run was interrupted or lost across the environment restart; its temporary evidence directory and running browser are unavailable. No successful provider proof has been promoted. The delivery probe is not workspace invitation delivery evidence. Signup completion, password recovery, Google/GitHub OAuth completion, invitation acceptance, session controls, quota/latency, HTTPS cookie inspection and signing-key rotation remain unverified for acceptance.

## Resume

Resumed on 2026-10-14 (local date). The provider setup correction is committed as `8cf69ba`. Its targeted verification recorded 46 passing frontend unit tests plus frontend/script type, format and lint checks. The full script suite is not verified: its existing root generated-drift regression exceeded its 180-second deadline, including in an isolated retry. `bun run generate:check` passed on resumption; the full drift regression remains required verification work.


On resumption, the ignored private Clerk/Resend configuration survived. The temporary test passwords, encryption ring and browser session did not. Private test configuration has been restored in an owned mode-0700 persistent directory outside the repository, with mode-0600 credential storage. Only the two authorized demo test accounts had their generated test passwords restored; existing verification state was preserved and the fresh signup address remains unused. `provider:check` passed again. The operator completed Google two-step verification in the headed persistent browser. The authorized inbox handoff has been restored, and a new complete provider run is in progress. No completion claim depends on the lost run. Follow `docs/development.md` for the inbox handoff. Complete the native provider project and promote only its generated sanitized record and matching proof. Run `provider:automated:check` after all automated rows pass. Complete the exact remaining human-only rows before running `provider:acceptance:check`.

Do not mark plan 35 or Phase 2 complete, and do not start Phase 3, while required provider evidence remains pending. No secret values, inbox codes, invitation credentials, raw cookies or account addresses are recorded here.
