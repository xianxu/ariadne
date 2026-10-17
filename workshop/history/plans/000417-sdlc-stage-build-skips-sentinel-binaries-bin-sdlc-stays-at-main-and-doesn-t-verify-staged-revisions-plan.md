# #417 — Stage builds and verifies sentinel binaries

**Goal:** `sdlc stage --build` and `sdlc stage --clear` prepare every Go command declared by the resulting checkout and refuse any stale, modified, missing, or unverifiable Go binary it would present for a live check.

**Architecture:** Keep `make build` as the generic build owner. A sentinel can name an additional target in a narrow machine-readable comment (`# stage-build: sdlc-build`); stage runs each declared target after the generic build. A final inspection of `bin/*` compares every Go binary's `vcs.revision` with the staged HEAD and connects each sentinel command to a binary by Go build-info package path, so output names may differ from source directory names.

**Tech stack:** Go `debug/buildinfo`, `os/exec`; existing Make targets and test fixtures.

## Concepts and integration points

| Name | Kind | Lives in | Status |
|------|------|----------|--------|
| Sentinel declaration reader and parser | INTEGRATION | `cmd/<name>/.skip-make-build`, `cmd/sdlc/stagebuild.go` | modified |
| Stage build orchestration | INTEGRATION | `cmd/sdlc/stagebuild.go` | modified |
| Build revision validation | PURE | `cmd/sdlc/stagebuild.go` | modified |

The sentinel is already the opt-out from the generic scanner. A `stage-build` declaration adds only the Make target that owns its special build; it does not duplicate a build recipe. An older sentinel with no declaration remains valid when the generic `make build` itself produced a matching, current binary (for example a Makefile.local target with a differently named output). For the legacy `cmd/sdlc` sentinel specifically, use its established `sdlc-build` target if the generic build left `bin/sdlc` stale; this lets clear rebuild a resting main commit that predates the new declaration, including an already-resting retry. Other missing outputs refuse and name the source command, asking for an explicit target. No shell is used to execute a declaration: accept exactly one Make target token and pass it as an argv element. This keeps the marker data, not code.

The final scan treats scripts and non-executable data as outside the VCS-stamp contract. A native executable with unreadable build info refuses, because corrupt Go metadata cannot safely be distinguished from a native non-Go binary at this boundary. A Go binary in `bin/*` must have `vcs.revision == staged HEAD` and an explicit `vcs.modified == false`. A sentinel's main package must occur as one verified binary's build-info `Path`; `bin/pair` can thus satisfy `cmd/pair-go`. If no Go binaries exist and no sentinel main packages exist, preserve the existing informational no-op for non-Go repositories.

The stage switch and record remain intact on build or verification failure. Clear first switches to the resting branch, then uses the same build-and-verify path against its resting HEAD; a failed clear can be retried even though the target is already resting. The return error names the offending binary or sentinel and expected commit. No new durable files or ledgers are created (ARCH-FUNERAL). This changes only the stage gate; it does not alter protected close or merge gates (ARCH-PURPOSE, ARCH-ORDER). Target execution is bounded by the finite checked-out `cmd/*` set and the Make process, as existing `make build` is (ARCH-CONSTRAINTS).

## Tasks

### 1. Tests first: stale and differently named binaries

**Files:** `cmd/sdlc/stage_test.go`, `cmd/sdlc/stagebuild_test.go`

- [x] Replace the old mismatch-warning assertion with a refusal that names `bin/tool`; add missing revision and `vcs.modified=true` cases.
- [x] In a staged fixture with `cmd/sdlc/.skip-make-build`, make the generic build leave an old `bin/sdlc` and a `sdlc-build` target create a new one. Assert that the declared target runs and the final revision matches the staged commit.
- [x] Add a sentinel command with a different output name and build-info `Path` matching its source (`cmd/pair-go` → `bin/pair`); assert that it passes after generic build. Make that output absent and assert a named refusal.
- [x] Add a clear fixture with staged-version binaries, switch to resting HEAD, and assert clear rebuilds and verifies resting-version binaries. Make the first clear build fail and assert a repeat repairs it while already resting.
- [x] Run `go test ./cmd/sdlc -run '^TestStage(Build|Sentinel)' -count=1` and confirm the intended red results before implementation.

### 2. Build each declared sentinel and verify final contents

**Files:** `cmd/sdlc/stagebuild.go`, `cmd/sdlc/stage.go`, `cmd/sdlc/.skip-make-build`, `Makefile.workflow`, `cmd/sdlc/helptext/stage.md`

- [x] Add `# stage-build: sdlc-build` to the sdlc sentinel. Document the optional directive beside `.skip-make-build` in `Makefile.workflow`.
- [x] After `make build`, enumerate `cmd/*/.skip-make-build` with `main.go`; parse at most one target token per marker and invoke `make <target>` for each declared target. A malformed or duplicate declaration refuses, naming the marker.
- [x] Scan final `bin/*` once. For each Go binary, inspect build-info; refuse on missing/mismatched revision or modified stamp and include its path. Match build-info `Path` to each sentinel package and refuse if any sentinel has no verified output.
- [x] Call the same build-and-verify path from clear after the target returns to its resting branch. Make a repeated clear perform the build check too, so interrupted or failed post-switch work converges.
- [x] Update stage help to say that `--build` verifies and refuses, rather than merely reporting revision lines.
- [x] Re-run targeted tests; run `git diff --check` and `make -k` (full suite), then close through the SDLC gate.

## Live check and landing

The TL requested the close review before the live check while #418 holds ariadne:0. Once the TL grants :0, build this branch's sdlc binary, run a live `sdlc stage --build` into :0, inspect every printed `bin/*` revision, clear, and inspect every restored revision. Do not stage without a fresh grant. Land after the live check passes.

## Revisions

- 2026-10-16 19:45 PDT — TL approved the sentinel and refusal design and added the reverse transition: `stage --clear` must rebuild the resting binaries and reuse the same revision check, including an already-resting retry. Add the clear fixture and call site; no separate verifier.
- 2026-10-16 — Full suite passed; TL requested the close review while #418 still holds ariadne:0. Move the live check after the review and before landing, with a fresh :0 grant.
- 2026-10-16 — Boundary review found incomplete evidence validation and untrusted Make arguments. Require an explicit `vcs.modified=false`, fail closed for native executable files whose Go metadata is unreadable, and accept only simple Make target names in sentinel declarations. Extract the Make invocation as an injected boundary; test the generic→owner→verify sequence, failures and retry with a stateful fake, while retaining a real-Make conformance fixture. Add README usage and recovery guidance. These corrections preserve the same stage/clear contract (ARCH-PURPOSE, ARCH-MOCK, ARCH-SECURE).
- 2026-10-16 — Close round 2 found that the native-format guard omitted XCOFF, Plan 9 and a Mach-O variant supported by Go's build-info reader; cover the reader's complete executable-format set with regression fixtures. The table also incorrectly called the file-reading sentinel parser PURE; classify that combined reader/parser as INTEGRATION instead (ARCH-PURE, ARCH-PURPOSE).
- 2026-10-16 — Live stage/clear against resting main f5927dfe showed that main predates the new `sdlc` sentinel declaration: generic `make build` skips `bin/sdlc`, so clear correctly refused a stale binary but did not finish. Preserve the shared build-and-verify path and add a narrow fallback to the existing `sdlc-build` target only for a legacy `cmd/sdlc` marker and only when `bin/sdlc` is not already clean at the resting HEAD. Cover the fallback and already-current generic build with a red-first clear-rebuild test; no new persistent state (ARCH-ORDER, ARCH-DRY).
