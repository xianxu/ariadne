---
id: 000252
status: open
deps: []
github_issue:
created: 2026-09-25
updated: 2026-09-25
estimate_hours:
---

# Issue cards: card fields on a tracker ref, details on the branch

## Problem

An issue file mixes two kinds of content with different audiences, and one
file cannot live in two places:

- **Global, slow-changing:** id, status, claim, dates, hours, title, the
  original problem. Every agent, slot and peer needs the latest.
- **Branch-local, fast-changing:** Spec, Done when, Plan, Log. Only the agent
  doing the work needs them, and they belong next to the code (keeping the
  implementation plan with the code keeps agent context together).

Today both live in `workshop/issues/NNN-slug.md` on `main`. sdlc therefore
publishes *copies* of issue commits to `origin/main` (`issue new`, `claim`,
`change-code`'s `PublishCommit`) while the resting or code branch keeps its own
commits. Local and remote diverge, and every issue verb adds to it.

Evidence from pair on 2026-09-24/25 (on top of #251's list):

- `main-slot1` reached **3 ahead / 19 behind**. The 3 were #332 issue syncs
  (also on the #332 branch); the 19 were mostly peers' issue-sync commits.
  Reconciled by hand three times in one session (backup ref, rebase, verify
  tree, drop backup).
- `change-code` printed `Source commit 6f184dc2…: published (79982e7c…)`: the
  published copy has a different SHA, so the original stays "ahead" forever.
- Both landing PRs (pair#168, pair#169) were "not mergeable": `issue new` or
  `claim` had put a blank template or status-only edit on `origin/main`, the
  branch had the filled file with no shared ancestor, so add/add conflict,
  resolved by hand ("keep ours").
- `sdlc merge` ends with "workspace retained on unchanged main-slot1. Refresh is
  separate". Nothing refreshes it, so behind-counts only grow.
- Filing a spin-off issue mid-branch forces a choice between rebasing the code
  branch onto everyone's changes (disruptive) or making a local copy
  (divergence).

## Spec

Designed with the operator in pair session, 2026-09-25.

### Split the issue file by audience

- **Card:** `workshop/issue-cards/NNN-slug.md` on a **dedicated tracker ref**
  (not `main`, so filings and claims stop moving `main`). Holds the global
  frontmatter (status, claim/`started`, `estimate_hours`, `actual_hours`,
  `github_issue`, dates), `# Title` and `## Problem`.
  A card is just the slower-changing part of today's issue file.
- **Details:** stays at today's path `workshop/issues/NNN-slug.md`, on the
  branch, landing on `main` with the code. It is a **superset of the card**
  (so the file looks unchanged to agents) plus Spec, Done when, Plan, Log,
  Revisions and branch-local frontmatter (`deps`, `flow:`, review anchors).

### Ownership rules

- **Only sdlc writes card fields.** `issue new`, `claim`, `close`, `merge` and
  similar verbs commit directly to the tracker ref, one file per commit, by
  their own SHA (no copies). Agents never free-edit card fields.
- **The card fields inside details are a read-only copy that sdlc owns.** sdlc
  refreshes them from the card whenever it touches details (`issue sync`,
  `start-plan`, `change-code`, `close`), under a marker such as
  `# card fields mirrored from issue-cards/…; edit via sdlc`.
- **Drift is refused, not ignored.** If a copied field in details disagrees with
  the card, `issue sync` and the close gate refuse with the next action
  ("status is owned by the card; use `sdlc …`"). The card is the source of truth
  for the fields it owns.
- **Precedence covers frontmatter, not `## Problem`.** The card keeps the
  original report; details start with a copy of it that the branch may revise.
- **Work can begin without a details file; close cannot finish without one.**
  Details are created on first need (first `start-plan` or Log entry); `close`
  requires Done when + verification there.

### Flows

- **Filing, including a spin-off from a code branch:** write the card on the
  tracker ref (id allocation stays a compare-and-swap: the push must
  fast-forward). The code branch is untouched.
- **Card writes after a rejected push:** a rejection means another card
  changed; sdlc fetches, replays its one-file commit on the new tip and pushes
  again, never conflicting on content. The only real contention is two writes
  to the *same* card (e.g. two claims); refusing is correct there.
- **Close / merge:** `close` writes `codecomplete` to the card; `merge` writes
  `done` after landing. If the `done` write fails (network), the card stays at
  `codecomplete`; the next sdlc run detects "PR merged, card not done" and
  finishes it. This removes today's "close before committing the issue file"
  ordering trap.
- **Reading other issues mid-branch:** sdlc reads cards from the tracker
  checkout, so the latest state is visible without pulling code.
- **Resting branches** only fast-forward; with issue churn off `main`, they
  stop accumulating local-only issue commits.

### Unchanged

The status vocabulary, lifecycle and gate logic. What changes is where sdlc
reads and writes the card fields (tracker checkout vs branch file), plus a
one-time migration that splits existing issue files. Archiving keeps today's
convention.

### Relation to #251

Supersedes #251's storage and landing design (its Problem, the
own-SHA/snapshot + merge-back landing mechanics, clean resting branches):
those mechanics existed to fix what the card split removes. #251's **phased
lifecycle** (spec/plan/code phases, per-phase claims, `parked`,
verdict-backed `complete`, sufficiency judge) is not addressed here and stays
an open question for the operator.

### Tracker, dependencies and migration decisions

1. Use a normal branch named `issue-tracker`. How sdlc and agents locate its
   checkout remains to be designed.
2. `deps` belongs to details, not the card. Dependency checks must read the
   appropriate details; a card-only listing cannot supply dependency metadata.
3. Split existing issue files in one migration, not lazily on first touch.
   Move issue-number allocation to the tracker at that cutover. Preserve
   existing details archiving semantics (`workshop/issues/` to
   `workshop/history/`).
4. Freeze SDLC writers across the participating downstream repos for migration;
   a mixed-version compatibility window is not required. The durable plan must
   specify rollout order, treatment of outstanding branches, verification and
   recovery before writers resume.

For cards and numbering, the recommendation is to retain cards in
`workshop/issue-cards/` on `issue-tracker`, including cards for archived issues.
Allocate `max(id) + 1` from those cards, not the number of files: gaps must not
cause reuse. Migration must account for historical IDs as well as active ones.
On a rejected allocation push, fetch and recompute the ID and path against the
new tracker tip; merely replaying a commit can duplicate an ID under a different
slug. Retention costs one small card per issue and a scan proportional to the
number of cards; the durable plan should validate that cost at fleet scale.

A separate next-ID file is possible, but is not needed with retained cards. It
would introduce another authoritative value and require atomically committing
the counter update with each new card, revising the one-file-per-commit rule.
Retaining cards keeps allocation derived from the issue records. (`ARCH-DRY`)
Card retention and the allocation mechanism remain recommendations for the
durable plan, rather than settled operator choices.

## Done when

- `sdlc issue new`/`claim`/`close`/`merge` write only cards on the tracker ref,
  by their own SHA; no verb publishes a copy of a branch commit.
- A full issue cycle in a slot (new → claim → design → change-code → close →
  merge, plus a mid-branch spin-off `issue new`) leaves the resting branch
  0 ahead / 0 behind after refresh, and the PR merges with no issue-file
  conflict (e2e test).
- Hand-editing a copied card field in details is refused by `issue sync` and
  the close gate with an actionable message (test).
- Existing issue files are migrated; gates read card fields from the card.
- Atlas documents the card/details split and ownership rules.

## Plan

- [ ] Resolve open questions with the operator
- [ ] Durable plan (outside the quick-flow shell)

## Log

### 2026-09-25

- Filed from pair session after reconciling `main-slot1` by hand for the third
  time. Design converged with the operator; see Spec. Supersedes #251's storage
  and landing half.

## Revisions

### 2026-09-25 14:00 PDT — Creation completes when details land; explicit handoff

Reason: a visible card can belong to a process still writing the initial issue.
Allowing another agent to fabricate missing details would give two agents
ownership of the same file. The operator clarified the creation boundary and
agreed an explicit handoff for issues whose containing code branch cannot ship
yet. These decisions supersede the conflicting Spec and Done when clauses above.

#### Creation and eligibility

- `issue new` creates both files: the card on the tracker ref and details on
  the creator's current local branch. Filing a spin-off therefore does touch
  that branch; it does not publish the branch's code.
- Card-only visibility means creation is still in progress. The creator may
  continue authoring locally, but another thread cannot claim the issue or
  independently create details to begin work.
- The issue is treated as fully created only when its details land on `main`.
  Claim must verify that landing; a card or a local details file alone is not
  sufficient. This is a readiness condition, not a new status enum decision.
- Remove lazy details creation from `start-plan` or first Log entry. Close
  still requires the details and their Done when + verification evidence.

#### `sdlc issue move-detail --issue N`

An explicit initial-details handoff unblocks another thread without waiting
for the current code branch to ship:

- If local details exist, publish their current contents to `main`, preserving
  the work already written, then remove the local source after confirmed
  publication. The creator no longer needs that local file after handing off.
- If local details do not exist, create initial details from the card on
  `main` and publish them. This is the explicit completion of creation, not
  automatic fabrication by a claim attempt. No separate `make-detail` command
  is needed.
- Refuse if details already exist on `main`: creation is complete and another
  thread may own further work. Do not overwrite them.
- A read error is not file absence. Failed or uncertain publication must not
  discard the local source; retry must reconcile any already-published result.
- Reconcile the transfer in Git history so the original branch's eventual PR
  neither deletes the published details nor reintroduces its old copy. A plain
  copy followed by a tracked deletion is insufficient. The concrete transfer
  mechanism remains for the durable plan. (`ARCH-ORDER`)

#### Synchronization and acceptance changes

- The proposed workflow no longer needs `issue sync` as a special issue
  synchronization operation. Card writes commit/publish through their owning
  SDLC verbs; details use ordinary branch commits and publication. Local
  checkpointing remains necessary. `move-detail` handles the explicit early
  handoff, rather than broadly synchronizing issue files.
- Replace the original `issue sync` mirror-refresh/refusal references with
  the remaining operations that consume or update details; retain the close
  gate's ownership check. Distinguishing a stale mirror from a hand edit still
  needs a concrete design.
- Amend Done when: `issue new` writes a tracker card AND local details; claim
  refuses until details land on `main`. The lifecycle e2e must include that
  landing before claim, and a spin-off handoff while code remains unshipped.
- Add tests for both `move-detail` paths, existing destination refusal, read
  errors, publication failure/uncertainty and retry, and the original branch's
  later merge preserving the handed-off details and subsequent edits.
- These are requirements for #252, not implemented commands. The current
  `issue sync` remains the repository's checkpoint mechanism until replaced.

### 2026-09-25 14:14 PDT — Tracker and migration comments resolved

Reason: the operator resolved the storage and rollout choices while reading
the issue and asked how card retention should support issue-number allocation.

- Selected `issue-tracker` as the branch name and kept `deps` in details;
  updated the field ownership descriptions accordingly.
- Selected a one-pass split and a coordinated freeze of downstream SDLC
  writers, preserving existing details archiving semantics. Checkout location,
  outstanding-branch handling and migration recovery remain design work.
- Recorded the recommendation to retain cards and derive `max(id) + 1` from
  the tracker, including historical IDs and recomputation after contention.
  Explained why a separate counter is unnecessary unless later scale evidence
  warrants it; this recommendation is not an additional operator decision.
