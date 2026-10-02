# Slot Readiness in Fleet Inventory Implementation Plan

> **For agentic workers:** Consult AGENTS.md Section 3 (Subagent Strategy) to determine the appropriate execution approach: use superpowers-subagent-driven-development (if subagents are suitable per AGENTS.md) or superpowers-executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** `sdlc fleet inventory` reports one row per local slot: its address,
its member checkouts (the host plus the substrate clones weave declares for
it) and one readiness verdict (ready / holds-work / needs-recovery / missing /
unknown), with the reasons behind each member's verdict (no file lists).

**Architecture:** Two milestones, both local (no new network reads).
**M1** gives each checkout the one readiness fact inventory lacks — an active
Git operation — from a single shared detector that reads the worktree's `.git`
pointer and `lstat`s the markers (no git process), and adds the pure
per-checkout verdict. **M2** discovers slots from row paths (`repo:0` = a
fleet primary, `repo:N` = `<fleet>/worktree/<repo>-slotN/<repo>`), walks each
numbered host's declared substrate dependencies (`pkg/layergraph.ParseRows`),
adds the dependency clones as inventory rows (they are independent clones that
the fleet walk never sees), and folds members into one versioned `slots`
section; registers the recovery proof; updates help and atlas.

**Tech Stack:** Go; `pkg/workspace` (`SlotPath`, resting-branch rule),
`pkg/layergraph` (the `construct/deps` grammar), `internal/fleet` (rows,
#288 claims, #290 records cache), `internal/gitx`.

---

## Decisions

- **What a slot is.** Every fleet primary row is `repo:0` (resting `main`);
  every row whose tree is `workspace.SlotPath(fleet, repo, N)` for its own
  repository is `repo:N` (resting `main-slotN`). Both come from row paths; no
  `workspace.Resolve` per row. The resting-branch rule is exported once from
  `pkg/workspace` (`RestingBranch(n)`) and `Classify` uses it (ARCH-DRY).
  Ordinary feature worktrees and dependency clones are rows, not slots.
- **Membership is declared, never listed.** A `:N` slot's members are its host
  plus every substrate dependency reached transitively from the host's
  `construct/deps` (`substrate <path> [url]` rows via `layergraph.ParseRows`,
  each path relative to the checkout declaring it — weave's rule). Every
  resolved path must lie directly under the environment root (weave's slot
  policy); one that does not makes that member `unknown`. An undeclared sibling
  directory is not a member; a declared path that does not exist is a
  `missing` member. `data` rows are out of scope (the issue asks for substrate
  dependencies).
- **`:0` slots are their host alone** (operator-confirmed: `:0` peers are
  shared, not isolated; each `:1+` slot clones its own dependencies).
- **Dependency clones become rows.** Each existing member clone not already a
  row is collected with the same per-repository walk (`collectInventoryRepo` on
  its path), so it gets facts, branch-issue association and claims like any
  row, and a claim made in it is placed, not dangling. Its tracker reads reuse
  the records of the fleet primary of the same name when both have the same
  `origin` URL (one local `git config` read), so a clone adds no network read;
  otherwise it is read like any tracked checkout (#290's cache and deadline).
- **Resting branch per member.** Host: per the slot rule. Dependency clone:
  `main` — weave's environment policy is scoped, so it clones with `--branch
  main` and refreshes to `origin/main` (`cmd/weave/internal/acquire/acquire.go:87,152`).
- **Per-checkout verdict**, first match wins:
  1. `unknown` — a probe the verdict depends on returned an error instead of an
     answer: the facts' git reads (`rev-parse HEAD`, `status`), the base lookup
     (no `origin/main`/`main` to count unlanded commits against), the operation
     detector, reading a member's `construct/deps`, a member outside the
     environment root, or a member directory that is not a Git checkout.
  2. `needs-recovery` — dirty files (modified or untracked; `dirty_count`), an
     active Git operation, or a detached HEAD.
  3. `holds-work` — on a branch other than its resting branch with commits not
     on main (`ahead > 0`), or whose issue association names a status that is
     not terminal (`!vocab.Issue().IsTerminal(DeclaredStatus)`, even with zero
     commits); or this machine holds a tracker claim on the checkout (#288).
  4. `ready` — otherwise.
  Merged-ness is ancestry (`ahead == 0`): a squash-landed branch reads as
  holds-work — conservative, never wrongly ready.
- **Reasons, not paths** (operator). Each member carries its verdict and short
  reason codes (`dirty`, `operation:<marker>`, `detached`, `unlanded-commits`,
  `open-issue:<ref>`, `claimed:<ref>`, `probe:<what>`) with the error text for
  unknown; `dirty_count` is already on the row. No file lists.
- **Slot verdict** is the worst member verdict: ready < holds-work < unknown <
  missing < needs-recovery (operator-confirmed); every member keeps its own.
- **One operation detector.** `gitx.OperationMarkers` is the union of today's
  two lists (MERGE_HEAD, CHERRY_PICK_HEAD, REVERT_HEAD, REBASE_HEAD,
  rebase-merge, rebase-apply, sequencer, BISECT_LOG, BISECT_START), checked with
  `lstat` in the worktree's own git directory. `move-detail` and landing switch
  to it, obtaining that directory with one `git rev-parse --absolute-git-dir`
  (instead of a `--git-path` call per marker). Intended behavior change:
  landing now also refuses during REBASE_HEAD/BISECT_LOG, move-detail during
  sequencer/BISECT_START.
- **Versioned.** The inventory gains `schema_version: 1` and a non-null
  `slots`; strict decoding rejects other versions and unknown keys.
- **Observation, not reservation.** Help states that an action reusing a slot
  (pair#363's reboot, then `weave refresh`) re-checks it at action time.
- **ARCH-ORDER:** a single-shot read holding no state between events; a slot
  that changes mid-read is reported as observed.
- **ARCH-FUNERAL:** creates nothing durable.
- **ARCH-CONSTRAINTS:** no new git process for existing rows (the detector is
  file reads); per slot, a bounded `construct/deps` read per member; per
  dependency clone, the ordinary row cost (~10 local git commands, ~0.15s) plus
  one `git config` read. No new network reads.

## Core concepts

### Pure entities

| Name | Lives in | Status |
|------|----------|--------|
| `gitx.OperationMarkers` / `gitx.ActiveOperation` | `cmd/sdlc/internal/gitx/operation.go` | new |
| `gitx.WorktreeGitDir` | `cmd/sdlc/internal/gitx/operation.go` | new |
| `MeasuredFacts.Operation` / `.OperationError` | `cmd/sdlc/internal/fleet/types.go` | modified |
| `Verdict` (+ constants, rank) / `MemberVerdict` | `cmd/sdlc/internal/fleet/slots.go` | new |
| `JudgeCheckout` | `cmd/sdlc/internal/fleet/slots.go` | new |
| `DeclaredMembers` | `cmd/sdlc/internal/fleet/membership.go` | new |
| `SlotHost` / `discoverSlots` | `cmd/sdlc/internal/fleet/slots.go` | new |
| `Slot` / `SlotMember` / `AssembleSlots` | `cmd/sdlc/internal/fleet/slots.go` | new |
| `Inventory` (`schema_version`, `slots`) | `cmd/sdlc/internal/fleet/types.go` | modified |
| `workspace.RestingBranch` | `pkg/workspace/identity.go` | new |

- **ActiveOperation(gitDir, lstat)** — the first present marker, or an error
  for any other `lstat` failure; pure over an injected `lstat`.
- **WorktreeGitDir(root, readFile)** — `.git` directory, or the `gitdir:`
  target of a `.git` file (relative targets resolved against the worktree).
- **JudgeCheckout(row, resting)** — verdict + reason codes from the row's facts,
  issue association and claims. Pure.
- **DeclaredMembers(host, envRoot, readDeps)** — the transitive substrate walk
  with the read injected; cycles end on a visited set; a dependency with no
  `construct/deps` is a leaf; absent and out-of-environment members are
  returned, marked.
- **discoverSlots(rows, fleetRoot)** — hosts and addresses from paths. Pure.
- **AssembleSlots(hosts, members, rows)** — joins members to rows by canonical
  path, judges, folds the worst. Pure.

### Integration points

| Name | Lives in | Status | Wraps |
|------|----------|--------|-------|
| `CollectFacts` | `cmd/sdlc/internal/fleet/facts.go` | modified | `.git` pointer + marker `lstat`s |
| `collectSlots` | `cmd/sdlc/internal/fleet/inventory.go` | new | `construct/deps` reads, dependency-clone rows |
| `aliasRecords` | `cmd/sdlc/internal/fleet/issues.go` | new | `git config --get remote.origin.url` |
| `gitOperationInProgress` / `landingNoOperation` | `cmd/sdlc/issuemovedetail.go`, `cmd/sdlc/landing.go` | modified | `git rev-parse --absolute-git-dir` |

## Chunk 1 — M1: per-checkout readiness

- [ ] M1 — per-checkout readiness facts and verdict

- [ ] **Operation detector.** Tests: `ActiveOperation` table over markers ×
  {present, absent, lstat error}; `WorktreeGitDir` over {dir, absolute gitdir
  file, relative gitdir file, malformed file}; a real-git conflicted rebase
  in a linked worktree reports `rebase-merge`. Switch `move-detail` and landing
  (their tests stay green).
- [ ] **Facts.** `CollectFacts` records `Operation` / `OperationError` with no
  added git process (fake-git command log unchanged); row JSON unchanged.
- [ ] **Verdict.** `TestJudgeCheckout`: one table over the cross product of
  probe outcome, working-tree state, branch state and claim, asserting verdict
  and reason codes against the precedence in Decisions.
- [ ] `sdlc milestone-close --issue 289 --milestone M1`.

## Chunk 2 — M2: slots in inventory

- [ ] M2 — fleet inventory reports one readiness row per slot

- [ ] **Membership.** `TestDeclaredMembers`: one table over {no deps, direct,
  transitive, cycle, malformed deps, absent member, member outside the
  environment, undeclared sibling}.
- [ ] **Slots + contract.** `TestDiscoverSlots` (primary, numbered host, a
  look-alike path of another repository, feature worktree); `TestAssembleSlots`
  (worst-of fold over the verdict order; missing; member not a row → unknown;
  `:0` host-only); contract round trip + one rejection per invariant
  (version, unknown verdict, slot verdict ≠ worst member, reason-less
  non-ready member); render snapshot.
- [ ] **Wiring.** Dependency clones collected as rows; tracker aliasing.
  `TestFleetInventorySlotReadiness`: a real fleet fixture (`prod` declaring
  `substrate ../dep`, one numbered slot per verdict cause, causes placed in
  host and clone), asserting each slot's verdict, that an undeclared sibling
  never counts, that a clone's claim is placed, and that a clone adds no
  tracker read.
- [ ] **Recovery, help, atlas.** Extend the `fleet inventory` entry's proofs;
  help (verdicts, observation-not-reservation); atlas (`sdlc-binary.md`).
- [ ] `make test`; measure inventory time before/after (target: ≤ +1s here);
  `sdlc close --issue 289`.

## Verification

- `make test` green; inventory timing in the Log.
- Manual: `go run ./cmd/sdlc fleet inventory --json | jq '.slots[] | {address,
  verdict, members: [.members[] | {path, verdict, reasons}]}'` lists every
  `repo:0` and `repo:N` here, pair:N with members `pair` and `ariadne`.

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
- 2026-10-02 — plan body rewritten to the revised design above (supersedes the
  first draft's tasks, Core-concepts table and dirty-path decisions; the draft
  is in commit 3324197a). Resumed after #290 landed.
