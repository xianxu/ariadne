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
- **Per-checkout verdict** (`JudgeCheckout`), in this precedence:
  1. `needs-recovery` — whenever the facts that show it were read: dirty files
     (`dirty_count > 0`), an active Git operation, or a detached HEAD. Probes
     that failed alongside are listed in its reasons (a known problem is never
     downgraded to unknown; needs-recovery also outranks unknown in the fold).
  2. `unknown` — otherwise, a probe the verdict depends on returned an error
     instead of an answer: the facts' git reads, the base lookup (no
     `origin/main`/`main` to count unlanded commits against), the operation
     detector, the issue named by an issue-prefixed branch, this machine's
     claims, reading a member's `construct/deps`, a member outside the
     environment root, or a member directory that is not a Git checkout. One
     exception: if only the claims are unread and the checkout already holds
     work, it is holds-work (a claim could only add holds-work).
  3. `holds-work` — commits not on main (`ahead > 0`, on any branch including
     the resting one), an issue association on a non-resting branch whose
     status is not terminal (`!vocab.Issue().IsTerminal`, even with zero
     commits), or a claim this machine holds on the checkout (#288).
  4. `ready` — otherwise.
  Merged-ness is ancestry (`ahead == 0`): a squash-landed branch reads as
  holds-work — conservative, never wrongly ready.
- **Reasons, not paths** (operator). Each member carries its verdict and short
  reason codes (`dirty`, `operation:<marker>`, `detached`, `unlanded-commits`,
  `open-issue:<ref>`, `claimed:<ref>`, `probe:<what>`) with the error text for
  unknown; `dirty_count` is already on the row. No file lists.
- **Slot verdict** is the worst member verdict: ready < holds-work < unknown <
  missing < needs-recovery (operator-confirmed); every member keeps its own.
- **One operation detector.** `workspace.OperationMarkers` (in `pkg/` so weave
  shares it) is the union of today's three lists (MERGE_HEAD, CHERRY_PICK_HEAD, REVERT_HEAD, REBASE_HEAD,
  rebase-merge, rebase-apply, sequencer, BISECT_LOG, BISECT_START), checked with
  `lstat` in the worktree's own git directory. `move-detail`, `move`, landing
  and weave's refresh switch to it, obtaining that directory with one `git rev-parse --absolute-git-dir`
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
| `workspace.OperationMarkers` / `ActiveOperation` / `Lstat` | `pkg/workspace/operation.go` | new |
| `workspace.WorktreeGitDir` / `ReadGitPointer` | `pkg/workspace/operation.go` | new |
| `MeasuredFacts.Operation` / `.OperationError` | `cmd/sdlc/internal/fleet/types.go` | modified |
| `TreeRow.IssuesError` | `cmd/sdlc/internal/fleet/types.go` | modified |
| `Verdict` (+ constants, rank) / `MemberVerdict` | `cmd/sdlc/internal/fleet/slots.go` | new |
| `JudgeCheckout` | `cmd/sdlc/internal/fleet/slots.go` | new |
| `DeclaredMembers` / `MemberDecl` / `MemberState` | `cmd/sdlc/internal/fleet/membership.go` | new |
| `SlotHost` / `discoverSlots` | `cmd/sdlc/internal/fleet/slots.go` | new |
| `Slot` / `SlotMember` / `SlotDeclaration` / `AssembleSlots` / `withProbe` | `cmd/sdlc/internal/fleet/slots.go` | new |
| `ErrAmbiguousIssue` (in `AssociateBranchIssue`) | `cmd/sdlc/internal/fleet/issues.go` | new |
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
| `collectDependencyRows` (+ slot assembly in `CollectInventory`) | `cmd/sdlc/internal/fleet/inventory.go` | new | `construct/deps` reads, dependency-clone rows |
| `readDeclaration` / `statPath` | `cmd/sdlc/internal/fleet/membership.go` | new | `layergraph.ReadDeclaration`, `stat` |
| `layergraph.ReadDeclaration` / `DeclarationLimit` | `pkg/layergraph/read.go` | new (moved from weave's acquire) | bounded, no-follow, no-FIFO file read |
| `workspace.ValidSlotDependency` | `pkg/workspace/environment.go` | new (moved from weave's `Policy.validate`) | — (pure placement rule) |
| `canonicalMember` | `cmd/sdlc/internal/fleet/membership.go` | new | `CanonicalProspectivePath` |
| `sameOrigin` / `aliasOf` | `cmd/sdlc/internal/fleet/inventory.go` | new | `git config --get remote.origin.url` |
| `gitx.PublicationRepository` | `cmd/sdlc/internal/gitx/publicationtarget.go` | modified (exported) | remote URL → identity |
| `gitOperationInProgress` / `landingNoOperation` / weave `refresh` checkout check | `cmd/sdlc/issuemovedetail.go`, `cmd/sdlc/landing.go`, `cmd/weave/internal/refresh/git.go` | modified | `git rev-parse --absolute-git-dir` |

## Chunk 1 — M1: per-checkout readiness

- [ ] M1 — per-checkout readiness facts and verdict

- [ ] **Operation detector.** Tests: `ActiveOperation` table over markers ×
  {present, absent, lstat error}; `WorktreeGitDir` over {dir, absolute gitdir
  file, relative gitdir file, malformed file}; a real-git conflicted rebase
  in a linked worktree reports a rebase marker (REBASE_HEAD comes first). Switch `move-detail` and landing
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
- 2026-10-02 — M1 implementation: verdict precedence puts needs-recovery first
  whenever its facts were read (the previous text put unknown first, which
  would report a dirty checkout with an unreachable base as unknown). Unlanded
  commits count on the resting branch too. A failed claims read makes an
  otherwise-ready checkout unknown, not a holding one. Decisions rewritten in
  place.
- 2026-10-02 — M1 review BR-1: the verdict no longer infers a failed issue
  lookup from an empty association; the row records the lookup's error
  (`TreeRow.IssuesError`, `json:"-"`, the same failure already in
  diagnostics), and an issue-prefixed branch whose lookup answered with no
  match is ready. Minors: the detector moved to `pkg/workspace` and weave's
  refresh uses it (weave gains REBASE_HEAD/BISECT_LOG); the unused `root`
  parameter of `gitOperationInProgress` is gone. Core-concepts rows rewritten.
- 2026-10-02 — M2 implementation: integration names as built (table rows
  rewritten: `collectDependencyRows`, `sameOrigin`/`aliasOf` in
  `inventory.go`, `readDeclaration` in `membership.go`). Origins are compared
  as publication identities (`gitx.PublicationRepository`): this fleet's
  primary uses the SSH spelling and weave's clones the HTTPS one, and a raw
  URL comparison gave each clone its own tracker read (inventory 13.6s; with
  identities 6.3s, 15 `ls-remote` as before). A declaration error on a
  checkout already needing recovery stays needs-recovery with `probe:deps`
  listed.
- 2026-10-02 — M2 review BR-8: the `construct/deps` reader is weave's, moved to
  `pkg/layergraph.ReadDeclaration` (ordinary file only, `O_NOFOLLOW|O_NONBLOCK`,
  1 MiB bound); weave's acquire delegates to it and fleet uses it. Minors:
  present members are canonicalized before matching rows and aliases; the
  membership read/stat are injected into `collectDependencyRows`; contract
  "null"/"missing" rejection tests edit exactly one field (`editJSON`).
- 2026-10-02 — M2 review round 2 (BR-8, walk half): membership is judged by
  weave's own placement rule, moved to `pkg/workspace.ValidSlotDependency`
  (weave's `Policy.validate` calls it), on the declared and canonical paths
  together, so a member symlinked to a checkout elsewhere is `outside`, never
  that checkout. The transitive walk itself stays separate from weave's
  `acquire.Restore` on purpose: Restore performs acquisition (clones, staging,
  policy-gated IO); readiness needs a read-only walk over what exists. What
  they share — grammar (`layergraph.ParseRows`), reader
  (`layergraph.ReadDeclaration`) and placement rule — lives once in `pkg/`.
- 2026-10-02 — M2 close (FIX-THEN-SHIP) fixes bundled with the close: the
  pending-diagnostic flush runs after dependency-clone collection
  (`TestBrokenDependencyCloneIsReported`); `pkg/layergraph/read_test.go` tests
  the shared reader (ordinary, symlink, FIFO, oversize, missing); sdlc's
  `substrateChain` reads through it. `layergraph.Walk` and weave's `link` keep
  reading `construct/deps` through their injected FS abstractions on purpose
  (their tests depend on those seams); only direct OS reads were switched.
