---
collaboration-mode: discussion
---

# Resume an affected attempt after a registered flake: implementation plan

> **For agentic workers:** see AGENTS.md §3 (Subagent Strategy) to choose superpowers-subagent-driven-development
> or superpowers-executing-plans for this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** `sdlc local-check resume --name affected` reruns an affected attempt's registered flakes, then runs the
selected script groups, and records `sdlc-local/affected` for the same head. The inline known-flake step of a fresh
affected run stops recording a pass whose groups never ran.

**Architecture:** One pure helper turns "rerun these tests" plus an affected selection into the child's command batch:
the flake reruns, then the repository's affected command with `SDLC_AFFECTED_GO_PKGS=` empty and every selected group.
Both known-flake paths share that helper: the inline step in `runLocalCheck` and the saved resume in
`resumeSavedSuite`. The saved resume recomputes the selection the way a fresh affected run would, and fences it in a
**version 5** resume link (affected only; suite links stay version 3) so older sdlc builds fence that one head instead
of losing the whole store. The existing resume state machine, child record and proof validator are reused unchanged
in shape; the validator learns the affected parent/link pairing.

**Tech stack:** Go, `cmd/sdlc` (local-check runner, resume evidence store, `internal/affected`, `internal/flakes`).

---

## Background (what exists, what is missing)

- Resume (#394 follow-ups, `localcheck_resume*.go`) is suite-only in these places: `localcheck_resume_cmd.go`
  (`--name` check, `declaredCheck(suite)`, output line), `localcheck_resume.go` (`resumeInputs` policy,
  `resumeSavedSuite` retry command, `ResumeLink{Version: 3, Check: suite}`, `loadSavedAttemptCurrent(…suite…)`,
  `postResumeRecord`'s republish hint), `localcheck_resume_evidence.go:76` (`resumeEnvelope` requires version 3 and
  check suite) and `:233` (parent `Check == suite`), `localcheck_resume_saved.go:187`, `localcheck_resume_store.go:249`,
  and the hint in `localcheck_run.go` (printed only when `c.Name == suite`).
- An affected run's command is the repository's script under `env SDLC_AFFECTED_GO_PKGS=… SDLC_AFFECTED_GROUPS=…`
  (`affectedLocalCheck`). In pair the Go packages run first; a Go failure stops the script before its groups.
- Nothing records which groups ran, and the selection is not persisted (`localRunRecord.Scope` is `json:"-"`;
  `SuiteAttempt` has no selection field). So "run only the unreached groups" cannot be computed exactly. Rerunning
  every selected group is the simplest correct choice. A group that already passed costs its time again, and is never
  trusted without running.
- **Pre-existing gap (same class, fixed here):** the inline known-flake step (`knownFlakeStep`, reached from
  `runLocalCheck` for any test check) reruns only the named Go tests and records `KnownFlakePassed`, so an affected
  run that stopped at a flaky Go test lands without its groups. `TestAffectedKnownFlakeReruns` currently asserts that
  outcome. ARCH-PURPOSE: the issue's point is "an affected pass means the selection ran", so both paths are fixed.

## Decisions

| # | Decision | Why |
|---|----------|-----|
| D1 | The child batch is `flake reruns` + one affected command with `SDLC_AFFECTED_GO_PKGS=` and `SDLC_AFFECTED_GROUPS=<all selected>`, appended only when the selection has groups. | One child, one log, one exit: the existing child record and proof validator apply unchanged. The env contract already allows empty Go packages with groups (`never with both empty`). |
| D2 | Go packages are not rerun beyond the flaky tests. | `go test` over a package list reports every package even after a failure; the parent's facts name every Go failure, and the classifier refuses any unregistered one or any unattributed failure (a build failure, a timeout). |
| D3 | The saved resume recomputes the selection on the original head against main's current `.sdlc/affected.json`, exactly as a fresh affected run would. A full selection refuses ("run --name suite"). An empty selection refuses. | Main's config can move, but a fresh run would use the current selection too. Persisting the original selection on the parent would add a format change for no extra safety. |
| D4 | The selection is fenced in a version 5 `ResumeLink` (`check: affected`, new required `selection: {go_pkgs, groups}`). The matching receipt is also version 5. Suite links and receipts stay version 3. | ARCH-EVOLVE: an older sdlc decodes a v3 link with `check: affected` as an invalid record, which marks the whole snapshot incomplete and blocks every resume and `sdlc actual` in the repository. Version 5 > `suiteAttemptReadVersion` (4) makes the old reader decode it as `UnsupportedProof` for that head only (`teststats_store.go:569-586`). Ignoring the new field would also be unsafe: an old reader would accept a Go-only rerun as the affected proof. |
| D5 | The posted record's `Scope` comes from the link's selection (`affectedScope`). | The status description stays truthful on resume and republish without persisting `Scope`. |
| D6 | The batch's weight and timeout come from the affected check when groups are appended (`c.Weight`, `c.TimeoutMinutes`), else the current rerun envelope (weight 1, `min(15m, timeout)`). | The groups are the affected check's own workload; the declared envelope already bounds them. |

## Core concepts

### Pure entities

| Name | Lives in | Status |
|------|----------|--------|
| `affectedRerun` | `cmd/sdlc/localcheck_affected_rerun.go` | new |
| `affectedRerunBatch` | `cmd/sdlc/localcheck_affected_rerun.go` | new |
| `ResumeSelection` | `cmd/sdlc/localcheck_resume_evidence.go` | new |
| `ResumeLink` | `cmd/sdlc/localcheck_resume_evidence.go` | modified |
| `resumeEnvelope` | `cmd/sdlc/localcheck_resume_evidence.go` | modified |
| `resolveResumeProof` | `cmd/sdlc/localcheck_resume_evidence.go` | modified |
| `localCheck.Rerun` | `cmd/sdlc/localchecks.go` | modified |

- **affectedRerun**: what a known-flake child must run besides the flaky tests: the base affected command (from
  `affected.json`) and the selected groups and Go packages. `nil` for the suite.
  - **Relationships:** 0..1 per `localCheck` (set by `affectedLocalCheck`); 1:1 with a v5 link's `ResumeSelection`.
  - **DRY rationale:** the inline step and the saved resume build the same batch from it; neither rebuilds the env
    contract (`affectedLocalCheck` already owns it; the helper reuses its env-building function).
- **affectedRerunBatch(rerun [][]string, a *affectedRerun) [][]string**: appends the groups command when
  `a != nil && len(a.Groups) > 0`; otherwise returns `rerun`. Unit-tested without IO.
- **ResumeSelection** `{GoPkgs []string json:"go_pkgs"; Groups []string json:"groups"}`: strict on decode. Each
  list is sorted and unique, and each entry is nonempty with no whitespace, NUL or control character (they become one
  env argv entry, never shell text; ARCH-SECURE). Present and required iff `version == 5`. Absent on v3.
- **resumeEnvelope**: accepts `(3, suite)` and `(5, affected)`; nothing else.
- **resolveResumeProof**: the parent may be `suite` or `affected`. The link's check must equal the parent's. An
  affected parent requires a v5 link and a suite parent a v3 link. The child stays `suite-rerun`.

### Integration points

| Name | Lives in | Status | Wraps |
|------|----------|--------|-------|
| `planAffectedSelection` | `cmd/sdlc/localcheck_affected.go` | new (extracted) | git diff + package graph (`affectedDeps`) |
| `resumeSavedSuite` → `resumeSaved` | `cmd/sdlc/localcheck_resume.go` | modified | resume store, slot runner |
| `knownFlakeStep` | `cmd/sdlc/localcheck_flakes.go` | modified | slot runner |
| `newLocalCheckResumeCmd` | `cmd/sdlc/localcheck_resume_cmd.go` | modified | CLI, `readAffectedConfig` |

- **planAffectedSelection(ctx, d, ad, cfg, sha) (affected.Selection, error)**: the changed-paths → graph → `Plan`
  half of `runAffectedCheck`, extracted so the resume plans exactly as a fresh run (ARCH-DRY). `runAffectedCheck`
  calls it, with no behavior change.
  - **Injected into:** `resumeSaved` takes an `affectedDeps` (production: `productionAffectedDeps`; tests: the
    existing fake `affectedDeps` used by `localcheck_affected_test.go`).
- **resumeSaved**: the renamed `resumeSavedSuite`, which takes the check name. For affected it calls
  `planAffectedSelection` before publishing the link, writes the v5 link, and launches
  `affectedRerunBatch(flakeReruns, sel)`. The state machine (`resumeTransition`) and every effect boundary are
  unchanged.
- **knownFlakeStep**: launches `affectedRerunBatch(commands, g.check.Rerun)` with D6's envelope.
- **Fakes:** the existing ones. `localRunRepo` (fake gh statuses and a real git store), `scriptRerun` (records the
  argv a rerun ran), `resumeSavedFixture` and `resumeCoreFixture`, and the fake `affectedDeps`. Nothing new external.

## Operating envelope (ARCH-CONSTRAINTS)

- Workload: a batch CLI, run once per failed attempt. The resume adds one selection plan (a git diff, and loading the
  graph in a worktree: seconds to a minute on pair) and the selected groups' own runtime.
- CPU and slots: when groups run, the child holds the affected check's declared weight for the batch, the same
  envelope as the original run. Otherwise it keeps weight 1.
- Timeout: the affected check's `timeout_minutes` when groups run; it stops and records the run as unattributed (the
  resume is then refused) like any timed-out rerun.
- Memory, disk and network: N/A beyond the existing run, with one log per child as today.

## Other principles

- **ARCH-ORDER:** no new states or events. `resumeTransition`'s enumeration (Eligible → link published → launched →
  completion confirmed → child published → proof validated/rejected; Linked/Unknown → reconcile) is unchanged.
  Selection planning is a pure pre-step before `ResumeRequested`, with no external effect, and a failure there leaves
  the parent Eligible. The event most likely to be mishandled is an interrupted child after link publication. That is
  today's reconcile path, and the v5 receipt binds the v5 link by digest exactly as v3 does (test).
- **ARCH-FUNERAL:** creates nothing durable beyond today's resume bundle (link, child, receipt). They live in the
  same store under the same retention (`localcheck_resume_retention.go`), whose predicates gain the v5/affected
  pairing (test).
- **ARCH-EVOLVE:** one persisted format changes, `ResumeLink`/`ResumeReconciliation`. The change is a version bump,
  justified in D4. Unknown fields keep today's handling: `resumeObject` retains them raw, and they are never
  rewritten because evidence is immutable. Known fields stay strict: `selection` is required and non-null on v5, and
  rejected on v3. `SuiteAttempt`, child records and status records do not change.
- **ARCH-MOCK:** no new external dependency; the git and graph seams already have fakes.

## Plan

### Task 1: `affectedRerun` and the shared batch

**Files:** create `cmd/sdlc/localcheck_affected_rerun.go` and `_test.go`; modify `cmd/sdlc/localchecks.go` (field
`Rerun *affectedRerun` on `localCheck`) and `cmd/sdlc/localcheck_affected.go` (`affectedLocalCheck` sets it and
reuses one env builder).

- [x] Write `TestAffectedRerunBatch`: nil → unchanged; no groups → unchanged; groups → one appended
  `env SDLC_AFFECTED_GO_PKGS= SDLC_AFFECTED_GROUPS=a b <cmd…>` argv. Also assert `affectedLocalCheck` builds its
  command through the same env builder, so the two cannot drift.
- [x] Run `go test ./cmd/sdlc -run TestAffectedRerunBatch`: it fails (undefined).
- [x] Implement it:
  ```go
  type affectedRerun struct{ Command, GoPkgs, Groups []string }
  func affectedEnvCommand(cmd, goPkgs, groups []string) []string // env argv; used by affectedLocalCheck too
  func affectedRerunBatch(rerun [][]string, a *affectedRerun) [][]string
  ```
- [x] Run it: it passes. Commit `#586: affected rerun batch shared by both flake paths`.

### Task 2: The inline step runs the groups (pre-existing gap)

**Files:** modify `cmd/sdlc/localcheck_flakes.go` (`knownFlakeStep`: batch + D6 envelope) and
`cmd/sdlc/localcheck_affected_flakes_test.go`.

- [x] Change `TestAffectedKnownFlakeReruns` so an affected check with groups `a b` expects the rerun argv to include
  the groups command. Add `TestAffectedKnownFlakeGroupsFailRefuses`: the flake passes, a group fails, and the outcome
  is not `KnownFlakePassed` and posts no success. Add a no-groups case where the batch is unchanged.
- [x] Run them: they fail.
- [x] Implement `commands = affectedRerunBatch(commands, g.check.Rerun)`, with weight and timeout from D6.
- [x] Run them: they pass. Commit `#586: an affected known-flake rerun also runs its groups`.

### Task 3: `planAffectedSelection` extraction

**Files:** modify `cmd/sdlc/localcheck_affected.go`.

- [x] Run the existing `localcheck_affected_test.go` and `landingcar_affected_test.go` as the guard (no behavior
  change). Extract the selection planning; `runAffectedCheck` calls it.
- [x] Re-run them: they pass. Commit `#586: extract affected selection planning for resume`.

### Task 4: The v5 link, envelope and proof

**Files:** modify `cmd/sdlc/localcheck_resume_evidence.go`, `localcheck_resume_retention.go:141,190` and
`localcheck_resume_store.go:249`; tests in `localcheck_resume_evidence_test.go` and
`localcheck_typed_oldreader_test.go`.

- [x] Write the tests:
  - `TestResumeEnvelopeVersionsByCheck`: (3, suite) and (5, affected) pass; (3, affected), (5, suite) and (4, any) are
    refused.
  - `TestResumeSelectionStrict`: v5 without `selection`, with null, unsorted, duplicate or space/control entries is
    refused; v3 with `selection` is refused.
  - `TestResolveResumeProofAffected`: an affected parent with a v5 link and a passing child gives Passed; with a v3
    link it is refused; check mismatch is refused.
  - `TestOldReaderFencesAffectedLinkPerHead`: decoding a v5 link through the pre-#586 path (decodeResumeEvidence
    fails → `decodeUnsupportedProof`) yields an `UnsupportedProof` for that head. A suite attempt on another head in
    the same snapshot still resolves, and the snapshot carries no warning.
  - A v5 receipt binds a v5 link by digest.
- [x] Run them: they fail.
- [x] Implement `ResumeSelection`, the version/check table in `resumeEnvelope`, `selection` required on v5 links in
  `decodeResumeEvidence`, the parent check relaxed to `isTestCheck` with the link pairing, and retention predicates
  that compare against the parent's check. `resume_store.go:249` uses the locked parent's check.
- [x] Run them: they pass. Commit `#586: version 5 resume link fences an affected selection`.

### Task 5: The saved resume for affected

**Files:** modify `cmd/sdlc/localcheck_resume.go`, `localcheck_resume_saved.go:116-131,187` and
`localcheck_resume_cmd.go`; tests in `localcheck_resume_saved_test.go` and `localcheck_resume_test.go`.

- [x] Write the tests. These are the Done-when tests:
  - `TestResumeAffectedRunsFlakesThenGroups`: a saved affected attempt failed only on a registered flake (log shows
    the Go failure, no groups). Resume reruns the test, then runs groups `a b`, and posts `sdlc-local/affected`
    success with `KnownFlakePassed`, `ResumeParent`/`ResumeChild` and scope "N packages, groups a b". The link on disk
    is version 5 with that selection.
  - `TestResumeAffectedUnregisteredFailureRefuses`: an unregistered failure is refused before any link or child, and
    nothing is posted.
  - `TestResumeAffectedGroupFailureRecordsFailure`: the flake passes, a group fails, the proof is Failed and nothing
    is posted.
  - `TestResumeAffectedFullSelectionRefuses`: a full or empty recomputed selection is refused, naming
    `--name suite`, with no link.
  - A repeat resume (or republish) of a passed affected bundle reposts from proof without rerunning.
  - In `TestResumeCLIValidation`, `--name affected` is accepted, `--name lint` is refused with "only suite or
    affected", and `--name` is still required.
- [x] Run them: they fail.
- [x] Implement it:
  - `resumeSaved(ctx, d, ad, c, sha, parent)`: parent check == `c.Name`. For affected, plan the selection (D3) and
    write the link `{Version: 5, Check: affected, Selection}`. The batch is `affectedRerunBatch` with D6.
  - `loadSavedAttemptCurrent(…, c.Name, …)`, and the retry and republish hints name `c.Name`.
  - `resolvedSavedRecord` sets `rec.Scope` from the link selection (D5).
  - `resumeInputs` and `resume_saved.go:187` keep the suite's policy for both checks (`affectedCheckDecl` takes the
    suite's `KnownFlakes`, D4 of #473).
  - In the CLI, `--name suite|affected`. Affected reads `readAffectedConfig(main)`; an undeclared config refuses. It
    builds `productionAffectedDeps`, and its output line names the check.
  - The `localcheck_run.go` hint prints for affected too: `sdlc local-check resume --name affected --attempt …`.
- [x] Run them: they pass. Commit `#586: resume an affected attempt: flakes, then its groups`.

### Task 6: Help, atlas, verification

- [x] `cmd/sdlc/helptext/local-check.md`: the resume synopsis takes `--name suite|affected`, with one paragraph on
  "reruns the flakes, then every selected group; a full selection refuses". The resume command's `Short` changes to
  match. `atlas/workflow/ci-merge-check.md` gets the same note next to the resume section and the affected section.
- [x] Targeted run: `go test ./cmd/sdlc -run 'Resume|Affected|KnownFlake|OldReader|Retention'`. Then the pre-close
  check: `sdlc local-check run --pre-close` on the pushed head.
- [x] Commit `#586: docs: affected resume`.

## Done when (from the issue, with the inline fix)

- An affected attempt that failed only on registered flakes and stopped before its groups resumes. It reruns those
  tests, runs the groups, and records a pass for the head (`TestResumeAffectedRunsFlakesThenGroups`).
- An unregistered failure still refuses the resume (`TestResumeAffectedUnregisteredFailureRefuses`).
- A fresh affected run's known-flake step no longer lands without its groups
  (`TestAffectedKnownFlakeGroupsFailRefuses`).
- An older sdlc reading a v5 link fences only that head (`TestOldReaderFencesAffectedLinkPerHead`).

## Revisions

- 2026-10-24 (close review round 1): what shipped differs from the plan in
  three places.
  - `resumeSaved(ctx, d, c, sel resumeSelectionFunc, sha, parent)` takes a
    planner closure, not an `affectedDeps`. `affectedResumePlan` in
    `localcheck_resume_cmd.go` builds the closure from the pinned origin/main
    and an injected `affectedDeps` constructor (production:
    `productionAffectedDeps`). `TestAffectedResumePlanWiring` drives that
    wiring through a fake `affectedDeps` (BR-2).
  - Test names as shipped: `TestAffectedKnownFlakeRunsGroups` (for
    `TestAffectedKnownFlakeGroupsFailRefuses`),
    `TestResumeAffectedGroupFailureRecordsNoPass` and
    `TestResumeAffectedRefusesFullOrEmptySelection`;
    `TestAffectedKnownFlakeReruns` is unchanged, and
    `TestOldReaderFencesAffectedReceiptPerHead` was added for the v5 receipt.
    `localcheck_resume_retention.go` needed no change.
  - The resumed record's scope is the replanned selection (D3); the help text
    now says it can name packages neither run tested when main moved.
