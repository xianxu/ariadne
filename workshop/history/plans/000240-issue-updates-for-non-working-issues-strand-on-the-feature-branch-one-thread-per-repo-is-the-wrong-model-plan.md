# Non-issue workshop trunk publication implementation plan

> **For agentic workers:** Consult AGENTS.md Section 3 (Subagent Strategy) to determine the appropriate execution approach: use superpowers-subagent-driven-development (if subagents are suitable per AGENTS.md) or superpowers-executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Publish one explicitly selected, dirty non-issue workshop prose file to main while another issue branch is active, without carrying that prose in the issue branch.

**Architecture:** Add `sdlc trunk publish --path P` around the existing `gitx.TrunkFile.UpdateMany` CAS transaction. A pure path/three-body decision selects only approved prose and refuses a main path that moved from the branch merge base; a thin command reads local Git/worktree state, publishes exact bytes, and retains the local file. Retry observes main and recognizes a prior successful push without another write. The approved cleanup-policy revision below supersedes the earlier automatic-cleanup design.

**Tech Stack:** Go, Cobra, `cmd/sdlc/internal/gitx`, real bare-Git test fixtures.

---

## Core concepts

| PURE entity | Lives in | Status |
|---|---|---|
| `trunkProsePath` eligibility | `cmd/sdlc/trunkpublish.go` | new |
| `prosePublishDecision` | `cmd/sdlc/trunkpublish.go` | new |

`trunkProsePath` is lexical and accepts only `.md` paths below `workshop/parley/`, `workshop/pensive/`, `workshop/targets/`, `workshop/projects/`, plus exact `workshop/lessons.md`. It rejects issue, plan, atlas, code, base-layer, asset, traversal, and absolute paths. Filesystem inspection separately verifies regular files, including no symlink in the path. The explicit allowlist is intentional: code, settings and atlas require issue review. A path must be modified only in the working tree, not already committed on this issue branch or staged in the index; this keeps branch history isolated while the published dirty copy remains available for manual cleanup (ARCH-PURPOSE).

`prosePublishDecision(local, branchBase, main, presence)` yields already-present, write, or refuse. It extends the existing body comparison in `cmd/sdlc/republish.go:37` only for presence-aware prose semantics; share the equal-body rule rather than duplicate it (ARCH-DRY). It permits a new local file only if absent at the branch merge base and main. It permits an existing local edit only if main still has the branch-base bytes; identical local/main bytes are already-present. A main path that moved refuses even when only another line changed; the user must merge/reconcile deliberately. Tests are pure (ARCH-PURE).

| Integration point | Lives in | Status | Wraps |
|---|---|---|---|
| `readLocalProse` | `cmd/sdlc/trunkpublish.go` | modified | regular-file validation and bounded local bytes |
| `runTrunkPublish` | `cmd/sdlc/trunkpublish.go` | new | working tree, Git queries, `gitx.TrunkFile` |
| `trunkRestoreCommand` | `cmd/sdlc/trunkpublish.go` | new | literal-path manual cleanup guidance |
| `gitx.TrunkFile.UpdateMany` | `cmd/sdlc/internal/gitx/updatemany.go` | reused | Git commit-tree, fetch, CAS push, bounded retry |

The command pins the selected path and local bytes before publishing. It verifies a regular file and rejects staged or branch-committed changes to it. On each CAS attempt it compares the pinned branch-base tree and current trunk view, constructs one `TrunkWrite` with `ExactBytes`. It uses the same stateful Git seam and real bare-Git fixtures as other SDLC publication tests (`cmd/sdlc/internal/gitx/updatemany.go:75`, ARCH-MOCK). An uncertain push returns the existing uncertainty error and leaves local bytes; rerunning first checks whether main carries the local bytes and repeats cleanup guidance while retaining the local file. A rejected CAS retries only if unrelated main paths moved; a selected-path conflict refuses. The command serializes within the existing SDLC transaction lock, while the remote CAS governs other workspaces (ARCH-ORDER).

After confirmed publication, the command retains the local file and prints
literal-path cleanup guidance if its bytes still match the pinned read. A
tracked file uses Git restore, while an untracked file uses targeted Git clean
before merging main. If
the local path changed, it warns to preserve and reconcile the newer edit.
A retry with main already matching the retained bytes makes no new commit and
repeats the guidance. Other dirty paths and the issue branch's commits/index
are untouched. No durable record family is added: the main commit is the
artifact, and normal Git history retains it (ARCH-FUNERAL). Work is a
human-invoked, one-path transaction, not a watcher. The command accepts files
up to 1 MiB (prose envelope; larger files refuse before reading fully), uses
`gitx.BoundRemote` for a 2-minute deadline and command cancellation, and the
existing three-attempt remote CAS budget. An exceeded limit fails before push
and leaves the local file untouched (ARCH-CONSTRAINTS).

## Task 1: pin the path and decision contract

**Files:** `cmd/sdlc/trunkpublish.go`, `cmd/sdlc/trunkpublish_test.go`

- [x] Write `trunkProsePath` tests with arbitrary path components (including traversal and separators) and `prosePublishDecision` tests over adversarial present/absent/value combinations; assert the specific refusal reason.
- [x] Run targeted tests to prove they fail before implementation.
- [x] Implement the pure allowlist and decision; run the targeted tests green.
- [x] Commit the tested core with the command as `#240: publish selected workshop prose to main`.

## Task 2: publish and finish one explicit set

**Files:** `cmd/sdlc/trunkpublish.go`, `cmd/sdlc/trunkpublish_test.go`, `cmd/sdlc/main.go`, `cmd/sdlc/internal/recovery/catalog.go`

- [x] Write `runTrunkPublish` real bare-Git tests that mutate main and the working tree independently; assert no branch/index effects and rejection before write when the selected path moved. The reused `gitx.UpdateMany` seam's existing tests cover CAS races and unknown push responses.
- [x] Write real-Git tests for interrupted post-push response, retained local
  copies, and edits through an open descriptor; rerun the command to prove a
  matching main body makes no second commit and changed local bytes get no
  destructive cleanup guidance.
- [x] Implement the Cobra verb with the existing transaction lock and `TrunkFile.UpdateMany`. Keep Git/process calls in the thin command; inject only the publication seam needed to reproduce outcomes. Register recovery semantics with concrete proof names.
- [x] Run targeted package tests, `go test ./cmd/sdlc -run 'TestRecoveryContractsAreProven|TestTrunkProsePath|TestProsePublishDecision|TestTrunkPublish|TestReadLocalProse' -count=1 -timeout 10m`; commit as `#240: publish selected workshop prose to main`.

## Task 3: document, verify and review

**Files:** `README.md`, `atlas/workflow/sdlc-binary.md`, `workshop/issues/000240-*.md`, `workshop/lessons.md` if review finds a transferable rule.

- [x] Document `sdlc issue publish --issue` for issue details, `sdlc trunk publish --path` for non-issue workshop prose, and the branch-reviewed home for atlas/code/settings in README and atlas. Keep `atlas/index.md` current if adding a file (none planned).
- [x] Run `gofmt`, targeted tests, `make -k`, and the local suite via `sdlc local-check run --name suite` after the tree is stable; inspect logs and compare the patch with main.
- [x] Implement the approved retention and bounded-read repairs, rerun affected checks, and record passing local acceptance for renewed close review.

Next workflow checkpoints: run `sdlc close --issue 240 --verified <evidence>`; publish via `sdlc pr` and `sdlc merge` after SHIP and green CI.

## Revisions

### 2026-10-14 — plan-quality round 1

PQ-1: one path per invocation removes partial batch cleanup; the retry state is
explicit above. PQ-2: named functions now carry adversarial-input strategy, and
lexical eligibility is separated from filesystem inspection. PQ-3: the issue
has one close boundary. The review's minor envelope note is addressed too.

### 2026-10-14 — implementation validation scope

The command's real-Git tests cover its path-level decision and local cleanup;
the reused `gitx.UpdateMany` suite remains the proof for CAS retry and unknown
remote responses. This avoids duplicating the lower seam's stateful fake tests.

### 2026-10-14 — close review round 1

BR-5 exposed that Git pathspecs can expand a selected wildcard filename, so
every selected-path Git operation must use literal syntax. BR-6 exposed a race
between validation and destructive cleanup. Cleanup now atomically moves the
selected source to a nearby backup, rechecks its bytes and HEAD/index state,
then uses exclusive creation for a tracked replacement; an untracked file is
already removed from its original name by that move. On a changed moved inode
it links the source back only if the original name remains free; otherwise it
retains the backup and reports its path. Tests inject edits to both the moved
inode and the original path, and HEAD/index changes after the move. BR-7 adds
the command's usage and recovery guidance to README. This is one cleanup
transition per invocation; the prior one-path recovery policy remains.

### 2026-10-14 — close review round 2

The second boundary review disproved the claim above that one selected path
implies one recoverable cleanup transition. A stop after rename leaves a backup
that retry does not recognize (BR-1). Rechecking bytes before unlink cannot
protect writes through an already-open descriptor after the check (BR-6).
Automatic unlink and an unbounded no-lost-writes promise cannot coexist while
another process may retain that descriptor. The cleanup policy is under design
review before code changes; the next revision must define each interrupted
state and the lifetime of any retained inode, or explicitly narrow cleanup.

### 2026-10-14 — approved cleanup policy

The hub approved the narrower policy: publish the pinned bytes to main and
leave the selected local file untouched. No backup, restore, remove, receipt,
or working-tree effect follows the CAS. A retry compares the retained local
bytes with main; if they match, it makes no new commit and repeats exact-path
cleanup guidance. After publication, re-read the local path: if its bytes
changed, warn to preserve and reconcile those edits rather than suggest
`git restore`. Otherwise print `git restore -- <literal path>` after merging
main into this branch, or before a merge that would otherwise conflict. Tests
must keep the source inode open and write through it after publication, proving
the file and edit remain visible. Existing cleanup tests are replaced with
retention, retry, and changed-local guidance tests (ARCH-ORDER, ARCH-FUNERAL).

### 2026-10-14 — tracked versus untracked guidance

The first recorded suite attempt was stopped while queued, before any tests
ran, because the printed Git restore command cannot remove a newly created
untracked file before main is merged. The revised output uses literal-path Git
restore for tracked files and targeted literal-path `git clean -f` for untracked
files before a conflicting merge. Real-Git tests execute each printed command,
including a wildcard filename with another matching file beside it.

### 2026-10-14 — resume in ariadne:5

The moved branch is owned by ariadne:5. Current-plan summary wording now agrees with the approved retention policy: no automatic cleanup after publication or retry, and the pure path predicate uses its actual lowercase Go name. Historical cleanup designs remain in the preceding Revisions. Current main, including #369, #370 and #378, is integrated before renewed acceptance.

### 2026-10-14 — renewed local acceptance

After the independent #379 readiness repair landed, the renewed recorded full suite passed on 1e559e8d (594.5s, all 12 SDLC shards and 59 other packages). The final checklist row now names verified pre-close readiness; close and landing remain subsequent workflow transitions, matching the issue's earlier close-gate wording revision. The approved implementation scope is unchanged.

### 2026-10-14 — close round 3 read-time bound

BR-1 through BR-7 are disposed after the approved retention repair; BR-8 shows metadata size checks do not enforce the promised read-time envelope. Every local prose read, both before and after publication, must consume at most maxTrunkProseBytes+1 bytes and reject overflow without consuming the remainder. Add deterministic exact-limit, static-oversize and post-validation-growth controls that measure consumed bytes through production wiring, plus initial-refusal and post-publication-retention checks. This is the existing 1 MiB contract, not a new publication policy. The current CAS summary now states the actual pinned-byte behavior: callback revalidation concerns main, while local bytes are reread only for post-publication guidance (ARCH-CONSTRAINTS, ARCH-ORDER). Repeat focused checks and the recorded full suite after the code fix, before renewed close.
