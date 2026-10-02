# Slot Readiness in Fleet Inventory Implementation Plan

> **For agentic workers:** Consult AGENTS.md Section 3 (Subagent Strategy) to determine the appropriate execution approach: use superpowers-subagent-driven-development (if subagents are suitable per AGENTS.md) or superpowers-executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** `sdlc fleet inventory` reports one row per local slot: its address,
its member checkouts (the host plus the substrate clones weave declares for
it) and one readiness verdict (ready / holds-work / needs-recovery / missing /
unknown), with the paths behind a needs-recovery verdict.

**Architecture:** Two milestones. **M1** adds the per-checkout facts readiness
needs — the dirty paths the status scan already reads, and an active Git
operation from one shared detector (`gitx.OperationInProgress`, replacing two
duplicated marker lists) — plus the pure per-checkout and per-slot verdict.
**M2** discovers slots (each `primary` / `slot` host from `pkg/workspace`),
derives membership by walking the host's declared substrate dependencies with
`pkg/layergraph.ParseRows`, joins members to inventory rows, and emits a
versioned `slots` section; registers the recovery entry; updates help and atlas.

**Tech Stack:** Go; `pkg/workspace` (identity: kind, address, resting branch,
environment root), `pkg/layergraph` (the `construct/deps` grammar),
`internal/fleet` (inventory rows, claims from #288), `internal/gitx`.

---

## Decisions

- **What a slot is.** Every host checkout `pkg/workspace` classifies as
  `primary` (`repo:0`) or `slot` (`repo:N`). Ordinary feature worktrees and
  dependency clones are not slots (they stay inventory rows). A repository
  without the slot layout has no slots.
- **Membership is declared, never listed.** A numbered slot's members are its
  host plus every substrate dependency reached transitively from the host's
  `construct/deps` (`substrate <path> [url]` rows, `layergraph.ParseRows`;
  paths resolve relative to the declaring checkout, as weave does). In a
  numbered environment weave places each one directly under the environment
  root. An undeclared sibling directory is not a member. A declared path that
  does not exist is a `missing` member. `data` rows are out of scope: the issue
  asks for substrate dependencies, and weave's slot policy governs only those.
- **`:0` slots are their host alone** (operator-confirmed: `:0` peers are
  shared, not isolated; each `:1+` slot clones its own dependencies). A primary's substrate paths
  (`../ariadne`) resolve to other repositories' own primaries — themselves
  `:0` slots with their own rows. Counting them as members would report one
  checkout under two slots.
- **Resting branch per member.** Host: its identity's resting branch (`main`
  for `:0`, `main-slotN` for `:N`). Dependency clone: `main` (weave clones and
  refreshes dependencies on `main`, `cmd/weave/internal/acquire/acquire.go:153`).
- **Per-checkout verdict**, from that checkout's facts, first match wins:
  1. `unknown` — a probe the verdict depends on returned an error instead of
     an answer: the facts' git reads (`rev-parse HEAD`, `status`), the base
     lookup (no `origin/main`/`main` to count unlanded commits against), the
     operation-marker lookup, reading a member's `construct/deps`, or a member
     directory that exists but is not a Git checkout. (An absent declared
     dependency directory is `missing`, not `unknown`.)
  2. `needs-recovery` — any dirty path (modified or untracked), an active Git
     operation, or a detached HEAD. Reasons and paths are listed.
  3. `holds-work` — on a branch other than its resting branch that has commits
     not on main (`ahead > 0`), or whose name prefix names a non-terminal issue
     (even with zero commits); or, on any branch, this machine holds a tracker
     claim on the checkout (#288 `claims`, the claim-awareness the issue names).
  4. `ready` — otherwise: clean, no operation, on its resting branch or on a
     branch with nothing unlanded and no open issue.
  Merged-ness is ancestry (`ahead == 0`): a squash-landed branch reads as
  holds-work, which is conservative (never wrongly ready).
- **Slot verdict** is the worst member verdict, in the order ready <
  holds-work < unknown < missing < needs-recovery: something to recover
  outranks something absent, which outranks something unread. Unknown is never
  ready. Every member keeps its own verdict and reasons, so the worst one never
  hides the others.
- **Paths are bounded** (dirty *files* per checkout — the number of members is
  not capped). A needs-recovery member lists at most 20 dirty paths
  plus `dirty_paths_truncated` (the remainder count); `dirty_count` already
  gives the total. Paths are kept on the in-memory facts only (`json:"-"`), so
  rows' JSON does not grow.
- **Versioned.** The inventory gains `schema_version: 1`; consumers reject other
  versions (strict decoding already rejects unknown keys). `slots` is
  non-null.
- **Observation, not reservation.** Help text states that an action reusing a
  slot (pair#363's reboot, then `weave refresh`) re-checks it at action time.
- **ARCH-ORDER:** holds no state between events; a single-shot read. A slot that
  changes while being read is reported as observed (the verdict is re-checked
  at action time).
- **ARCH-FUNERAL:** creates nothing durable.
- **ARCH-CONSTRAINTS:** adds one `git rev-parse --git-path …` (all markers in
  one call) and at most a handful of `lstat`s per row; per slot, one
  `workspace.Resolve` and one bounded `construct/deps` read per member
  (`acquire`'s read bound). No network: membership and verdicts are local; the
  only network read is #288's existing tracker fetch per tracked repository.

## Core concepts

### Pure entities

| Name | Lives in | Status |
|------|----------|--------|
| `MeasuredFacts.DirtyPaths` / `.Operation` / `.OperationError` | `cmd/sdlc/internal/fleet/types.go` | modified |
| `Verdict` (+ `VerdictReady`… constants, `verdictRank`) | `cmd/sdlc/internal/fleet/slots.go` | new |
| `MemberVerdict` | `cmd/sdlc/internal/fleet/slots.go` | new |
| `JudgeCheckout` | `cmd/sdlc/internal/fleet/slots.go` | new |
| `SlotDecl` / `MemberDecl` | `cmd/sdlc/internal/fleet/slots.go` | new |
| `Slot` / `SlotMember` | `cmd/sdlc/internal/fleet/slots.go` | new |
| `AssembleSlots` | `cmd/sdlc/internal/fleet/slots.go` | new |
| `DeclaredMembers` | `cmd/sdlc/internal/fleet/membership.go` | new |
| `Inventory` (`schema_version`, `slots`) | `cmd/sdlc/internal/fleet/types.go` | modified |
| `gitx.OperationMarkers` | `cmd/sdlc/internal/gitx/operation.go` | new |

- **JudgeCheckout(row TreeRow, resting string, issueOpen func) MemberVerdict** —
  the per-checkout verdict and its reasons (pure; reads facts, issues, claims).
  - **DRY rationale:** one place defines ready/holds-work/needs-recovery; pair
    stops re-deriving git state.
- **DeclaredMembers(host string, readDeps func(dir) (content, found, err))
  ([]MemberDecl, error)** — the transitive substrate walk over
  `layergraph.ParseRows`, with the file read injected. Cycles terminate (a
  visited set); a dependency's own missing `construct/deps` is a leaf. Absent
  dependency directories are still members (`missing` downstream).
  - **Future extensions:** `data` rows, if a slot ever needs them.
- **AssembleSlots(decls []SlotDecl, rows []TreeRow, issueOpen) []Slot** — joins
  declared members to rows by canonical path, judges each, folds the worst.
  Pure, table-tested.
- **gitx.OperationMarkers** — the one list of in-progress-operation markers
  (union of today's two: MERGE_HEAD, CHERRY_PICK_HEAD, REVERT_HEAD,
  REBASE_HEAD, rebase-merge, rebase-apply, sequencer, BISECT_LOG, BISECT_START)
  plus `ActiveOperation(gitPaths []string, exists func(string) (bool, error))`
  — pure over resolved paths. `move-detail` and landing switch to it.

### Integration points

| Name | Lives in | Status | Wraps |
|------|----------|--------|-------|
| `CollectFacts` | `cmd/sdlc/internal/fleet/facts.go` | modified | `git status`, `git rev-parse --git-path` |
| `collectSlots` | `cmd/sdlc/internal/fleet/inventory.go` | new | `workspace.Resolve`, `construct/deps` reads |
| `gitOperationInProgress` / `landingNoOperation` | `cmd/sdlc/issuemovedetail.go`, `cmd/sdlc/landing.go` | modified | — (now call `gitx`) |

- **CollectFacts** keeps the status paths it already parses and resolves every
  operation marker with one `rev-parse --git-path` call, then stats each. The
  fleet fake git (`fakegit_test.go`) learns the new command.
- **collectSlots** — for each row, `workspace.Resolve` (an `InventoryOptions`
  seam, `ResolveIdentity`, so fleet unit tests stay fake-backed); for each
  `primary`/`slot` host, `DeclaredMembers` with a real bounded file read.

## Chunk 1 — M1: per-checkout readiness

- [ ] M1 — per-checkout readiness facts and verdict

### Task 1: shared operation detector
**Files:** create `internal/gitx/operation.go` + test; modify
`issuemovedetail.go`, `landing.go`.
- [ ] Failing test `TestActiveOperation` (pure: marker present / absent / stat
  error) and a real-git test starting a conflicted rebase → `rebase-merge`.
- [ ] Implement; switch both callers; their existing tests stay green.

### Task 2: facts carry dirty paths and the active operation
**Files:** `internal/fleet/facts.go`, `types.go`, `fakegit_test.go`, `facts_test.go`.
- [ ] Failing tests: dirty paths parsed from `-z` status (renames, untracked);
  operation detected through the fake; a failed operation probe is recorded,
  never "none".
- [ ] Implement; JSON of rows unchanged (goldens stay green).

### Task 3: per-checkout verdict
**Files:** create `internal/fleet/slots.go`, `slots_test.go`.
- [ ] Failing table test `TestJudgeCheckout` over {facts ok / unavailable /
  base unavailable / operation probe failed} × {clean / dirty / operation /
  detached} × {resting / merged branch / ahead>0 / zero-commit open-issue
  branch / closed-issue branch} × {claim held / none}: verdict, reasons, path
  cap at 20 with truncation count.
- [ ] Implement `JudgeCheckout`, `Verdict`, ranks.
- [ ] `sdlc milestone-close --issue 289 --milestone M1`.

## Chunk 2 — M2: slots in inventory

- [ ] M2 — fleet inventory reports one readiness row per slot

### Task 4: declared membership
**Files:** create `internal/fleet/membership.go`, `membership_test.go`.
- [ ] Failing tests (pure, injected reads): host with no deps; direct dep;
  transitive dep; cycle; malformed `construct/deps` → error; an undeclared
  sibling is not a member; absent declared dep is returned (to be `missing`).
- [ ] Implement `DeclaredMembers` on `layergraph.ParseRows`.

### Task 5: assemble slots + contract
**Files:** `slots.go`, `types.go` (Inventory `schema_version`, `slots`;
validation; strict decoding), `render.go`, tests.
- [ ] Failing tests: `TestAssembleSlots` (worst-of fold; missing member;
  member not in rows → unknown; `:0` host-only); contract round trip and one
  rejection per invariant (unknown verdict, slot verdict not the worst member,
  needs-recovery member without reasons, more than 20 paths, wrong
  schema_version); render snapshot.
- [ ] Implement.

### Task 6: collection wiring + real-git fixtures
**Files:** `inventory.go` (`collectSlots`, `ResolveIdentity` seam),
`cmd/sdlc/fleet.go`, new `cmd/sdlc/fleetslots_test.go`.
- [ ] Failing integration test `TestFleetInventorySlotReadiness` on a real
  fleet fixture (`<fleet>/{prod,dep}` primaries + `worktree/prod-slotN/{prod,dep}`
  environments, prod declaring `substrate ../dep`), one slot per case: ready;
  holds-work by unlanded commits; holds-work by a zero-commit issue branch;
  needs-recovery by modified file, untracked file, active rebase, detached
  HEAD — each once in the host and once in the dependency clone; missing
  dependency clone; an undeclared sibling directory that is dirty yet does not
  change the verdict; a failed probe (unreadable `construct/deps`) → unknown.
- [ ] Implement wiring.

### Task 7: recovery entry, help, atlas
**Files:** `internal/recovery/catalog.go` (extend the `fleet inventory` entry's
evidence and proofs), `helptext/fleet-inventory.md`,
`atlas/workflow/sdlc-binary.md`.
- [ ] `go run ./cmd/sdlc help recovery` shows `fleet inventory [read-only]` with
  the slot proof; `make test` green.
- [ ] `sdlc close --issue 289 --verified '…'`.

## Verification

- `make test` green.
- Manual: `go run ./cmd/sdlc fleet inventory --json | jq '.slots[] | {address,
  verdict, members: [.members[] | {path, verdict, reasons}]}'` on this machine
  lists ariadne:0–2 and pair:0–1, pair:1 with members `pair` and `ariadne`.

## Revisions

- 2026-10-02 — plan-quality PQ-1 + operator review, before implementation
  (the tasks above are to be re-cut to this at the next design pass):
  - **Dependency clones become inventory rows.** They are independent clones,
    so today's inventory never lists them (`FleetRepoDirs` excludes
    `worktree/`); joining slot members to rows would make every `:N` slot
    unknown. Each numbered slot's declared dependency clone is collected as a
    row (`collectInventoryRepo` on its path). Its tracker reads (claims, branch
    issue) reuse the read of the fleet repository with the same publication
    repository, so a clone adds no network fetch; with none in the fleet it is
    read like any tracked checkout.
  - **No new git processes for existing rows.** The active-operation check
    reads the worktree's `.git` pointer (file or directory) and `lstat`s the
    markers; slot discovery (`repo:0`/`repo:N`, environment root, resting
    branch) derives from row paths and the slot layout, not `workspace.Resolve`
    per row. Added cost: about 10 local git commands per dependency clone row.
  - **No dirty-file listing** (operator): a needs-recovery member gives its
    reasons (dirty count, active operation, detached HEAD), not paths. The
    `json:"-"` dirty-paths field and the 20-path cap are dropped.
  - Plan-quality Minors folded in: the unified operation-marker list is an
    intended behavior change for landing (gains REBASE_HEAD, BISECT_LOG) and
    move-detail (gains sequencer, BISECT_START), all checked with `lstat`;
    dependency clones rest on `main` because weave's environment policy is
    scoped (`acquire.go:87,152`); substrate paths from a member's
    `construct/deps` must resolve inside the environment root (else that
    member is `unknown`); "open issue" is `!vocab.IsTerminal(DeclaredStatus)`
    on the row's existing issue association; tests stated as one strategy line
    per risky function.
  - **Sequencing:** #288's tracker reads made inventory ~3x slower (15
    sequential `ls-remote`+`fetch`+auto-maintenance, 10.5s of 17s measured with
    `GIT_TRACE`). Fixed first in its own issue; #289 resumes on top of it.
