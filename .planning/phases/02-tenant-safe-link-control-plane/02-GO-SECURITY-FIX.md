# Phase 2 Go Security Repair

Completed 2026-10-09 as a separate atomic blocker repair after plan 02-05, before
02-06. This repair does not complete or advance a phase plan or requirement.

## Problem and scope

The existing package-exposure scan correctly blocked on newly published
GO-2026-6599, GO-2026-6600, GO-2026-6603, GO-2026-6604, GO-2026-6605,
GO-2026-6607, GO-2026-6608, GO-2026-6609, GO-2026-6610, GO-2026-6611,
GO-2026-6612, GO-2026-6613 and GO-2026-6617. Upgrade the existing supported
Go patch from 1.26.8 to 1.26.9 and `golang.org/x/net` from v0.59.0 to v0.60.0.
No other backend module versions changed. Bun 1.3.14, Node 22.23.3, React,
provider libraries and standalone tool versions remain pinned as before.

Primary verification sources:

- [Go release history](https://go.dev/doc/devel/release) confirms Go 1.26.9,
  released 2026-10-08, contains the supported-branch security fixes.
- [GO-2026-6610](https://pkg.go.dev/vuln/GO-2026-6610) identifies the HTTP/2
  `x/net` correction in v0.60.0.
- [GO-2026-6617](https://pkg.go.dev/vuln/GO-2026-6617) documents another affected
  standard-library surface covered by the patched runtime.

## Files and integrity provenance

- `.tool-versions`, `apps/backend/go.mod`, `scripts/scan.ts` and scanner test
  protocol/SBOM fixtures agree on Go 1.26.9. `docs/development.md` documents
  the exact launcher/toolchain selection.
- `apps/backend/go.sum` records the sumdb-verified `x/net` v0.60.0 module sum
  `h1:79p50tfZlm0J9YfoDsSi639qSXNGVwEzOPLCxM2FsYU=` and go.mod sum
  `h1:2DA/G1UfVbCpQPeWTmMPGY7Cs2PkBkwu743bVX5PIVg=`.
- The downloaded toolchain was verified by Go's checksum database. Independently,
  the official `go1.26.9.linux-amd64.tar.gz` distribution matched SHA256
  `42d158b4d8f7b61ac0a830567c940a86098fb7aac52e467a5ebec03ef5cc2f8d`
  from [official release metadata](https://go.dev/dl/?mode=json). Its `go/bin/go`
  was byte-identical to the cached toolchain executable.
- `tools.lock.json` retains govulncheck v1.8.0, source archive SHA256
  `396b780e8e9cf35cd5b0c481fed8e841cff89a21bf4f1b61556d1694f42fb572`,
  module/go.mod sums, and every transitive source sum. Only its compiler,
  temporary module Go directive, verification date and executable hashes changed.
- All four binaries were built twice from the verified source with
  `CGO_ENABLED=0`, `GOAMD64=v1`, `GOARM64=v8.0`, `-mod=readonly`, `-trimpath`,
  `-buildvcs=false`, and `-ldflags=-buildid=`. Repeated outputs matched exactly.
  The ordinary installer independently rebuilt the host binary in its own
  isolated stage and verified it against the new lock before execution/publication.

| Platform | govulncheck executable SHA256 |
|----------|------------------------------|
| linux-x64 | `cd3cb4bb5fb83a54f27f8bfee66f4b93dd9677c166089170b7ab2f3e5ea99b64` |
| linux-arm64 | `fb652a05669792284189d045d54bd6e37895b110b90f36f4073db3b97c696b9b` |
| darwin-x64 | `05a520745c06363982f3c4239ceff82ed2393ce861ef0dcd5bfde8074d99dfff` |
| darwin-arm64 | `c2d4c9ee9a497e3e8b4d33c5a13f4369c0905190fe3b7b2fd8ed9cc0b6546a54` |

Initial unbounded upstream HTTP/compiler downloads stalled and were terminated.
Bounded retries verified the official archive successfully; builds then used
the already verified cached compiler explicitly, avoiding repeated downloads.
No source/executable integrity checks were bypassed.

## Verification

- Backend `go mod verify`: passed. Isolated govulncheck build modules: verified.
- Verified tool installation and `--verify-config`: passed.
- Actual CI runtime/manifest agreement code: passed with Go 1.26.9,
  Node 22.23.3 and Bun 1.3.14.
- Full `bun run check`: **passed, exit 0**, under Go 1.26.9. Included uncached
  generation without artifact drift; tool checks; format/lint/vet/staticcheck;
  all workspace typechecks, tests and production builds; all script regressions;
  installer self-test; every registered race-enabled Go unit/integration suite;
  explicit migration checks; all role builds; and final dependency, complete
  worktree and full Git history scans. The browser gate validated **13 completed
  cases**, with production Next output cleanup before secrets scanning.
- Scanner regressions included the real temporary vulnerable OpenPGP import,
  which remained a required failure, plus protocol/scope/toolchain rejection,
  bounded private reporting, closed public sanitation and secret-history cases.
- Independent final `bun run scan:dependencies`: **passed, exit 0**. No
  vulnerable imported packages across all backend roles and tests; Bun clean.
  GO-2026-5932 remains visibly reported as unused `x/crypto` v0.57.0 inventory,
  with no vulnerable package imported. No exception, suppression or downgrade.
- `git diff --check`: passed. No production stubs, new trust-boundary surfaces,
  generated artifact changes or tracked file deletions were introduced.

Use the verified compiler for subsequent root commands:

```sh
export PATH="/home/logan78/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.26.9.linux-amd64/bin:$PATH"
export GOTOOLCHAIN=go1.26.9
```

The system Go installation remains untouched. Hosted CI for this repair is not
yet observed. Live provider acceptance remains pending plan 02-35.

## Workflow tracking and self-check

SDK handlers resolved only the new Go dependency blocker, recorded the repair
decision and session handoff, and retained the current phase/plan position.
ROADMAP and REQUIREMENTS are unchanged; 02-06 remains the next incomplete plan.
All seven implementation/documentation paths and this evidence file exist;
integrity, exact runtime agreement and the complete local gate passed.
