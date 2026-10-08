---
phase: 01-foundation-stability-system-boundaries
reviewed: 2026-10-08T00:05:01Z
depth: standard
files_reviewed: 124
files_reviewed_list:
  - .github/workflows/ci.yml
  - .gitleaks.toml
  - .tool-versions
  - apps/backend/.golangci.yml
  - apps/backend/cmd/api/main.go
  - apps/backend/cmd/flux/main.go
  - apps/backend/cmd/migrator/main.go
  - apps/backend/cmd/migrator/main_test.go
  - apps/backend/cmd/redirector/main.go
  - apps/backend/cmd/worker/main.go
  - apps/backend/go.mod
  - apps/backend/go.sum
  - apps/backend/internal/app/api.go
  - apps/backend/internal/app/cleanup.go
  - apps/backend/internal/app/cleanup_test.go
  - apps/backend/internal/app/lifecycle.go
  - apps/backend/internal/app/lifecycle_test.go
  - apps/backend/internal/app/observability_test.go
  - apps/backend/internal/app/redirector.go
  - apps/backend/internal/app/roles_test.go
  - apps/backend/internal/app/vendor_test.go
  - apps/backend/internal/app/worker.go
  - apps/backend/internal/config/config.go
  - apps/backend/internal/config/config_test.go
  - apps/backend/internal/config/observability.go
  - apps/backend/internal/database/database.go
  - apps/backend/internal/database/database_test.go
  - apps/backend/internal/database/migrations/001_setup.sql
  - apps/backend/internal/database/migrator.go
  - apps/backend/internal/database/migrator_test.go
  - apps/backend/internal/errs/http.go
  - apps/backend/internal/errs/types.go
  - apps/backend/internal/handler/base.go
  - apps/backend/internal/handler/base_test.go
  - apps/backend/internal/handler/handlers.go
  - apps/backend/internal/handler/health.go
  - apps/backend/internal/handler/health_test.go
  - apps/backend/internal/handler/openapi.go
  - apps/backend/internal/handler/openapi_test.go
  - apps/backend/internal/lib/email/client.go
  - apps/backend/internal/lib/email/client_test.go
  - apps/backend/internal/lib/email/emails.go
  - apps/backend/internal/lib/email/preview.go
  - apps/backend/internal/lib/email/templates.go
  - apps/backend/internal/lib/job/correlation_test.go
  - apps/backend/internal/lib/job/email_tasks.go
  - apps/backend/internal/lib/job/handlers.go
  - apps/backend/internal/lib/job/job.go
  - apps/backend/internal/lib/job/job_test.go
  - apps/backend/internal/lib/utils/utils.go
  - apps/backend/internal/lifecycle/cleanup.go
  - apps/backend/internal/logger/logger.go
  - apps/backend/internal/logger/logger_test.go
  - apps/backend/internal/middleware/auth.go
  - apps/backend/internal/middleware/context.go
  - apps/backend/internal/middleware/global.go
  - apps/backend/internal/middleware/middlewares.go
  - apps/backend/internal/middleware/rate_limit.go
  - apps/backend/internal/middleware/request_id.go
  - apps/backend/internal/middleware/tracing.go
  - apps/backend/internal/middleware/tracing_test.go
  - apps/backend/internal/model/base.go
  - apps/backend/internal/observability/correlation.go
  - apps/backend/internal/observability/redaction.go
  - apps/backend/internal/observability/redaction_test.go
  - apps/backend/internal/observability/telemetry.go
  - apps/backend/internal/observability/telemetry_test.go
  - apps/backend/internal/repository/repositories.go
  - apps/backend/internal/router/router.go
  - apps/backend/internal/router/system.go
  - apps/backend/internal/server/server.go
  - apps/backend/internal/service/auth.go
  - apps/backend/internal/service/services.go
  - apps/backend/internal/sqlerr/err.go
  - apps/backend/internal/sqlerr/handler.go
  - apps/backend/internal/testing/assertions.go
  - apps/backend/internal/testing/container.go
  - apps/backend/internal/testing/helpers.go
  - apps/backend/internal/testing/server.go
  - apps/backend/internal/testing/transaction.go
  - apps/backend/internal/testing/transaction_test.go
  - apps/backend/internal/transport/contract_test.go
  - apps/backend/internal/transport/health.gen.go
  - apps/backend/internal/transport/oapi-codegen.yaml
  - apps/backend/internal/validation/utils.go
  - apps/backend/internal/validation/utils_test.go
  - apps/backend/static/assets.go
  - apps/backend/static/openapi.html
  - apps/backend/static/openapi.json
  - apps/backend/taskfile.yml
  - apps/backend/templates/assets.go
  - apps/backend/templates/emails/welcome.html
  - biome.json
  - bun.lock
  - compose.yaml
  - deploy/otel-collector.yaml
  - docs/development.md
  - docs/observability.md
  - package.json
  - packages/emails/package.json
  - packages/emails/src/templates/welcome.test.tsx
  - packages/emails/src/templates/welcome.tsx
  - packages/openapi/openapi.json
  - packages/openapi/package.json
  - packages/openapi/src/contracts/health.ts
  - packages/openapi/src/gen.test.ts
  - packages/openapi/src/gen.ts
  - packages/openapi/src/index.ts
  - packages/zod/package.json
  - packages/zod/src/health.test.ts
  - packages/zod/src/health.ts
  - packages/zod/src/index.ts
  - packages/zod/src/utils.ts
  - patches/ts-deepmerge@8.0.0.patch
  - scripts/check.test.ts
  - scripts/check.ts
  - scripts/generate.ts
  - scripts/generated.test.ts
  - scripts/install-tools.ts
  - scripts/scan.test.ts
  - scripts/scan.ts
  - scripts/tsconfig.json
  - tools.lock.json
  - turbo.json
findings:
  critical: 5
  warning: 2
  info: 0
  total: 7
status: issues_found
---

# Phase 01: Code Review Report

**Reviewed:** 2026-10-08T00:05:01Z  
**Depth:** standard  
**Files Reviewed:** 124  
**Status:** issues_found

## Narrative Findings (AI reviewer)

### Summary

Reviewed the explicit Phase 01 source/configuration scope and its existing error, database, generated-contract, test-runner and observability boundaries. Five BLOCKER findings and two WARNING findings remain. The strongest defects are a reproduced collector privacy leak, a reproduced database credential mismatch between migrator and API, and quality gates that either require an uninstalled executable or repair the drift they must reject.

Evidence below distinguishes direct reproductions from source-traced failures. The reported clean local full check does not exercise these cases. Hosted CI execution was not observed during this review. Review probes used synthetic values and disposable loopback containers; no production credentials or raw scanner reports were exposed. No implementation files were changed and no commits were made.

### Critical Issues

### CR-01: Collector exports private resource and scope schema URLs

**Severity:** BLOCKER  
**File:** `/home/logan78/Desktop/flux/deploy/otel-collector.yaml:19-23`  
**Also affected:** `/home/logan78/Desktop/flux/deploy/otel-collector.yaml:52-56`, `/home/logan78/Desktop/flux/deploy/otel-collector.yaml:81-85`, `/home/logan78/Desktop/flux/apps/backend/internal/app/observability_test.go:721-788`

**Issue:** The collector sanitizes resource/scope attributes, scope name and version, but never sanitizes the separate OTLP resource-level and scope-level `schemaUrl` fields. Both fields accept arbitrary strings and bypass the allowlist in all three signals. A dirty or independently instrumented sender can therefore put private URLs, query tokens or personal data directly into exported telemetry. This violates the advertised collector defense boundary and SAFE-01 even though the application SDK reconstructs its own resources safely.

**Evidence:** Started the exact digest-pinned Collector 0.162.0 with the unchanged checked configuration, then submitted synthetic OTLP trace, log and metric requests containing a private-marker URL in each resource and scope `schemaUrl`. All three requests returned HTTP 200; the debug exporter retained the marker six times. The existing dirty-input test omits these six fields, so its passing result does not cover this surface.

**Fix:** Clear resource and scope schema URLs in each trace/log/metric transform context using the accessors supported by the pinned collector, or reconstruct those envelopes without incoming schema URLs. Add dirty-input coverage for all six fields and inspect both outgoing OTLP and collector debug output. Validate the transform configuration against the pinned collector image before accepting the repair.

### CR-02: Root check rewrites stale artifacts before its drift gate

**Severity:** BLOCKER  
**File:** `/home/logan78/Desktop/flux/scripts/check.ts:227-248`  
**Also affected:** `/home/logan78/Desktop/flux/scripts/check.ts:315-317`, `/home/logan78/Desktop/flux/packages/emails/package.json:8-13`, `/home/logan78/Desktop/flux/packages/openapi/src/gen.test.ts:205-213`, `/home/logan78/Desktop/flux/packages/openapi/src/gen.ts:46-49`

**Issue:** Both full and fast root checks run workspace builds and tests before `generate:check`. The emails build exports directly over the checked backend HTML. The OpenAPI CLI test invokes the generator with no explicit output arguments, so it overwrites both checked JSON documents before asserting that they match. Consequently a stale or missing welcome HTML, canonical OpenAPI JSON, or served OpenAPI JSON is repaired by earlier stages and the later byte comparison sees fresh bytes. CI can pass a commit whose published checked artifacts are stale. This defeats PLAT-05/D-10 and makes the ordinary check command mutate tracked assets.

**Evidence:** The stage construction is ordered: workspace builds at lines 228-229, workspace tests at 243-245, script tests at 247-248, and the drift check at 315-317. The email build invokes the export with the tracked backend output directory. The OpenAPI test spawns `node [generatorPath]`; the generator defaults point at the two tracked JSON files. That test compares the files only after the write. The isolated drift-check implementation itself is sound; its callers erase the evidence beforehand.

**Fix:** Make workspace builds/tests leave tracked artifacts untouched; use explicit temporary outputs for the CLI test and keep publishing email assets in the explicit generation command. Run the isolated drift gate before any stage that can write checked artifacts, or preserve and verify the initial artifact bytes for the whole root check. Add root-level mutation regressions for each artifact category that prove full/fast check reject initial drift without repairing it.

### CR-03: Fresh CI requires Task but does not install or pin it

**Severity:** BLOCKER  
**File:** `/home/logan78/Desktop/flux/.github/workflows/ci.yml:92-121`  
**Also affected:** `/home/logan78/Desktop/flux/apps/backend/internal/app/roles_test.go:751-764`, `/home/logan78/Desktop/flux/apps/backend/internal/database/migrator_test.go:258`, `/home/logan78/Desktop/flux/scripts/install-tools.ts:40`

**Issue:** Mandatory integration tests execute an external `task` binary for real builds, dry-run role targets and the migration check. CI installs Go, Node, Bun and the manifest's quality tools, but neither installs Task nor includes it in the verified tool inventory. A fresh runner without a host-installed Task fails the required tests with executable-not-found, regardless of the successful developer-machine check. The stated fresh-checkout prerequisite documentation also omits this dependency.

**Evidence:** The test call sites explicitly invoke `task`. No Task installation exists in the workflow or shared installer. The published [Ubuntu 24.04 runner inventory](https://raw.githubusercontent.com/actions/runner-images/main/images/ubuntu/Ubuntu2404-Readme.md), inspected for image 20260927.320.1, does not list Task among installed tools. Failure on that runner is a source/inventory inference; this review does not claim to have observed the pushed workflow run.

**Fix:** Add a verified Task version and checksums to the shared tool manifest/installer, make the installed binary available to all test subprocesses and install it in CI. Include it in prerequisites and tool verification. Alternatively remove the external Task requirement from mandatory runtime tests and add an explicitly provisioned separate Task parity gate. Verify a fresh environment whose PATH initially contains no Task.

### CR-04: API changes database passwords containing spaces

**Severity:** BLOCKER  
**File:** `/home/logan78/Desktop/flux/apps/backend/internal/database/database.go:94-102`  
**Also affected:** `/home/logan78/Desktop/flux/apps/backend/internal/database/migrator.go:68-74`

**Issue:** The API pool DSN applies `url.QueryEscape` to a password placed in URL userinfo. Query encoding turns a space into `+`; userinfo decoding preserves `+` literally. A valid PostgreSQL password containing a space therefore becomes a different password and the API cannot start. The migrator uses `url.UserPassword` correctly, so the same accepted configuration can migrate successfully and then fail at API startup. User and database names are also interpolated without component-appropriate URL encoding.

**Evidence:** Against a disposable pinned PostgreSQL 17.11 container with a synthetic password containing one space, the built migrator exited 0 and the built API exited 1 with the identical database settings. This is a retained preexisting bug in a reviewed foundation path; role extraction does not resolve it.

**Fix:** Share the migrator's URL builder with the pool constructor, using `url.URL`, `url.UserPassword`, properly encoded path and `url.Values`; or assign connection fields directly on pgx configuration. Add a real database test covering spaces and reserved characters that exercises both migrator and API against the same credential.

### CR-05: Scanner and installer timeouts leave descendants running and can wait indefinitely

**Severity:** BLOCKER  
**File:** `/home/logan78/Desktop/flux/scripts/scan.ts:22-55`  
**Also affected:** `/home/logan78/Desktop/flux/scripts/install-tools.ts:87-109`

**Issue:** These subprocess wrappers kill only the direct child on timeout/output overflow and settle only on its `close` event. Compiler/scanner descendants can inherit stdout/stderr. Killing their parent leaves the descendants alive and the pipes open, preventing `close` and keeping the promised bounded scan/install operation pending until those descendants exit. Standalone commands can hang indefinitely; an outer root/CI timeout merely imposes a different, much larger deadline.

**Evidence:** Called the exported scanner `capture` with `sh -c 'sleep 2 & wait'` and a 30 ms timeout. It rejected only after approximately 2005 ms when the surviving child released the inherited pipes. A nonterminating descendant makes the same failure unbounded. The installer's spawn/kill/close sequence has the same defect. The root runner already demonstrates process-group termination at `scripts/check.ts:329-341`.

**Fix:** Use a shared subprocess helper that starts and terminates the entire process group on the supported platform, bounds final pipe draining/settlement, and preserves closed diagnostics. Apply it to scanner and installer wrappers. Test nested children that hold stdout/stderr through both timeout and output overflow, and assert bounded completion plus descendant termination.

### Warnings

### WR-01: Exported OpenAPI document contains unresolved schema references

**Severity:** WARNING  
**File:** `/home/logan78/Desktop/flux/packages/openapi/src/index.ts:31-69`  
**Also affected:** `/home/logan78/Desktop/flux/packages/openapi/src/gen.ts:9-17`

**Issue:** `Object.assign` replaces the generated `components` object with only `securitySchemes`, deleting the named health schemas while operations still reference them. The private generator recovers schemas by generating the contract a second time, so checked/served JSON is currently valid, but the exported `OpenAPI` value is invalid and disagrees with the canonical document. Any reuse of that value bypasses the repair and produces broken references. Keeping a knowingly invalid builder plus a private repair undermines the single document boundary.

**Evidence:** Importing the exported value yielded no component schema keys and three unresolved health response references. `serializeOpenAPI()` yielded the repaired schemas. This warning concerns the exported builder; it does not claim the currently served checked JSON is invalid.

**Fix:** Preserve the initial generated `components` when adding security schemes, export the resulting complete document, and serialize that same value. Remove the duplicate generation/recovery path. Add a reference-resolution test for the exported document and an equality test against the serialized canonical document.

### WR-02: HTTPError.Is treats every HTTP error code as equal

**Severity:** WARNING  
**File:** `/home/logan78/Desktop/flux/apps/backend/internal/errs/http.go:43-47`

**Issue:** The method's documented contract is matching stable error codes, but it returns true for any `*HTTPError` target. For example, `errors.Is(&HTTPError{Code:"NOT_FOUND"}, &HTTPError{Code:"UNAUTHORIZED"})` returns true. This reusable foundation helper cannot distinguish errors and will route code-specific recovery into the wrong branch when used. Current middleware uses `errors.As`, so no active routed authorization bypass is asserted here.

**Fix:** Compare the actual stable codes after a nil-safe type check, or remove the custom `Is` method if code matching is not an intended API:
```go
func (e *HTTPError) Is(target error) bool {
    other, ok := target.(*HTTPError)
    return ok && e != nil && other != nil && e.Code == other.Code
}
```
Add tests for matching codes, differing codes, wrapped errors and a typed-nil target.

---

_Reviewer: gsd-code-reviewer_  
_Depth: standard_

