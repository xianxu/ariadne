---
id: 000248
status: done
deps: []
github_issue:
created: 2026-09-23
updated: 2026-09-24
estimate_hours:
started: 2026-09-23T23:34:13-07:00
flow: {kind: quick, provenance: inferred, spec: "50bdd523", done: "34aba0ca"}
actual_hours: 0.83
---

# Agent procedure: move this branch to :N

## Problem

Operators say "move this branch to :0" (or `:N`): take the issue branch
checked out in the current slot and check it out in slot `:N` instead, usually
`:0`, because that is the checkout whose built binary and runtime files the
operator's live sessions run from, so a smoke test needs the branch there.
Git allows a branch in only one worktree at a time, so the source slot must
first go back to its resting branch.

No agent procedure names this. `atlas/workflow/workspace-branching.md` covers
"In :2, create an issue branch from :1", refreshing a resting slot, and
independent work, but not moving an existing branch. On 2026-09-23 (pair#316)
an agent did not recognise the phrase until the operator spelled out the
steps, then improvised them: it guessed checkout paths instead of using
`sdlc workspace :N`, and it switched a `:0` holding untracked operator files
without a stated rule.

## Spec

Add a **"Move this branch to :N"** section to
`atlas/workflow/workspace-branching.md`, alongside the other procedures and
reusing its "Before switching or refreshing" preflight (ARCH-DRY). Like the rest
of that doc, it is an agent procedure over existing Git/SDLC commands, not a new
CLI verb.

1. Resolve source (current) and destination with `sdlc workspace --json` /
   `sdlc workspace :N --json`. Same `repo_identity`; destination on its
   `resting_branch`; source on a non-resting issue branch.
2. Preflight both trees as the doc already requires, except for untracked
   files in the destination. `:0` routinely holds operator scratch files, so
   adopt #246's rule: keep untracked files that don't collide with the branch,
   refuse collisions, never stash/reset/auto-commit. State this exception
   explicitly rather than silently relaxing the empty-status rule.
3. Before switching, compare destination-resting commits with both the feature
   branch and its upstream. If destination rest has commits absent from the
   feature branch, stop for an explicit ordering decision. The operator may
   publish those commits through their normal review path and then rebase the
   feature on updated remote main, deliberately include unpublished commits in
   the feature, or choose a temporary smoke test with a later reconciliation.
   Never push or rebase as part of the move. Then switch the source to its
   resting branch and `git switch <branch>` in the destination. Report which
   resting commits remain parked and absent from the build.
4. Repo-specific post-move step: a repo whose runtime is a built artifact (pair:
   `make build` in `:0`) declares it in its `AGENTS.local.md`; the procedure
   says "run the destination repo's declared post-move build, if any" and
   verifies the branch head is what was built.
5. Tell the operator that already-running sessions keep the old binary; a fresh
   thread/relaunch picks up the build.
6. Reverse ("move it back" / after merge): destination returns to its resting
   branch with the same preflight; `sdlc merge`'s durable-slot landing (#246)
   owns the post-merge return where it applies.

Discoverability: the base-layer AGENTS.md "Peer/slot" guidance names the
phrase ("move this branch to :N") and links the section, so a fresh session
maps the operator's words to the procedure.

## Done when

- `workspace-branching.md` has the "Move this branch to :N" section covering
  steps 1–6, including the untracked-file rule and the local-only-commits
  decision before switching. Its history check explicitly compares destination
  rest to both the feature and the configured upstream; the real-Git fixture
  executes the guide's history commands and exercises both comparisons.
- The AGENTS.md base layer points the phrase at it; propagated downstream.
- README names the move phrase and links the shared procedure for operators.
- Pair's `AGENTS.local.md` declares its post-move build (`make build` in :0),
  delivered through Pair-owned issue pair#318 and its own close/merge boundary.
- Dry run: a fresh agent session told "move this branch to :0" from a pair slot
  performs the procedure without further explanation.

## Plan

- [x] Document the move and return procedure beside slot branching, using the existing identity and readiness preflight. Preserve harmless destination untracked files, refuse incoming-path collisions, and report resting-only commits.
- [x] Add the phrase-to-procedure pointer to exported `AGENTS.base.md`; add Pair's `:0` post-move `make build` declaration on an isolated Pair branch.
- [x] Exercise source/destination switching, preserved refs and scratch files, collision refusal, and local-only resting commits with real Git fixtures; dry-run the written steps from a fresh agent context.

## Revisions

### 2026-09-23 — implementation shape

**Reason:** the source constitution is `AGENTS.base.md`, exported by `construct/base.manifest`; the root `AGENTS.md` is generated. Pair's primary checkout currently has an unrelated issue branch and untracked operator files.

**Delta:** keep the change as an agent procedure over existing `sdlc workspace` and Git. Edit Pair's local declaration on a separate checkout so its active work remains untouched. The procedure treats the target's existing tracked paths and incoming branch paths as the collision boundary, with Git's non-forcing switch as the final guard (ARCH-DRY, ARCH-SECURE). No new CLI, persistent state, or background work is introduced (ARCH-PURE, ARCH-ORDER, ARCH-FUNERAL).

### 2026-09-23 — local resting history is an ordering decision

**Reason:** the operator pointed out that moving a feature to `:0` parks local main commits; later PR integration can leave local and remote main on different lineages. A mere report at the end does not protect the intended history order.

**Delta:** compare the destination resting branch with the feature and its configured upstream before switching. When rest contains commits absent from the feature, stop for a choice: publish rest first and rebase the feature on the reviewed remote tip; deliberately include those unpublished commits in the feature; or accept a temporary test and reconcile rest before the feature ships. Moving branches itself never publishes or rewrites commits (ARCH-ORDER, ARCH-PURPOSE).

### 2026-09-23 — review boundary and upstream evidence

**Reason:** the first close review found that the guide described the configured
upstream without an explicit comparison. It also could not inspect Pair's
separate repository in Ariadne's pinned review range.

**Delta:** require a concrete comparison of destination rest with both the
feature and its configured upstream, with a real-Git fixture for each. Pair's
local build declaration is delivered and reviewed under pair#318 on its own
isolated branch; its composed `AGENTS.md` was verified with this Ariadne branch.
The pair#318 boundary and merge must complete before #248 is considered fully
shipped.

### 2026-09-24 — exercise the written history check

**Reason:** the second close review found that independently repeating the
upstream Git command in a test would still pass if the agent procedure omitted
that command.

**Delta:** the real-Git fixture now executes the move guide's fenced history
commands and requires the destination-only commit to appear in both comparisons.
Removing either command from the guide makes the fixture fail. This tests the
agent-facing contract itself (ARCH-PURPOSE).

### 2026-09-24 — operator discoverability

**Reason:** the third close review found README's slot guidance still named
only branching and refreshing even though the move is a user-invoked phrase.

**Delta:** add the move phrase to README's existing slot-procedure link and
update that link title. AGENTS.base.md remains the agent entry point.

## Log



- 2026-09-24: closed — Remote claim history merged without changing reviewed files; local main and origin/main are both ancestors of this feature; real-Git guide and Pair composition evidence remain valid.; review verdict: SHIP
- 2026-09-24: closed — Guide history commands run against real-Git fixture and identify parked commits against feature and configured upstream; fresh agent moved/built clean case and stopped on local-main divergence; README and AGENTS link procedure; Pair declaration closed in pair#318 and composed by weave.; review verdict: SHIP
### 2026-09-23

- Filed from pair#316's session, where the move was done by hand: pair-slot1
  → `main-slot1`; `~/workspace/pair` → `000316-couch-altn-binding-latency`
  with untracked `.nvimlog`, `.qoder/` and workshop drafts preserved and its
  4 local-only `main` commits left behind; `make build`; the fix confirmed in
  `bin/pair` via `go tool nm`.
- Related: #244 (slots v2 workflows), #246 (durable-slot landing: return-to-rest,
  untracked-collision rule to reuse), #247.
- `go test ./cmd/sdlc -run '^TestWorkspaceProcedure' -count=1` passed with
  real-Git move and collision fixtures. A disposable Pair worktree composed the
  exported Ariadne guide and Pair's post-move `make build` declaration using
  `weave compile`; Pair's active `:0` checkout was untouched.
- An independent fresh-agent dry run moved a feature from `pair:1` to `pair:0`,
  preserved `operator-scratch`, returned `:1` to `main-slot1`, and ran its
  declared `make build` without changing the selected HEAD. With a local-only
  `main` commit added to `:0`, a second run stopped before switching and asked
  for an explicit ordering choice.
- Pair's authored declaration is committed on `000318-declare-pair-post-move-build`
  at `defbbac5` and will cross its own Pair close/merge boundary. Ariadne's
  review range cannot contain a file owned by the Pair repository (ARCH-PURPOSE).
- The follow-up review required regression evidence tied to the agent guide.
  The fixture now runs the guide's history command block against its real Git
  repository, rather than duplicating those commands in Go.
- The third review accepted that evidence and found the operator README link
  needed the new move phrase; README now points at the same shared procedure.
