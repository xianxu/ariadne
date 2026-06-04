---
id: 000079
status: working
deps: []
github_issue:
created: 2026-06-03
updated: 2026-06-03
estimate_hours: 2.5
---

# doc-review-flow: branch-scoped prose review with per-round git journaling

## Problem

`xx-fix` co-authoring (operator + agent trading 🤖 markers in a markdown doc)
works well, but a heavier multi-round review has no memory: each round
overwrites the last and the only record of how a paragraph evolved — and *why* —
lives in the ephemeral chat transcript, which is gone within days. We want the
full back-and-forth retained as durable, attributable, forensically-greppable
history, without inventing a bespoke versioning scheme (git already does this)
and without leaving review scaffolding (markers) in shipped prose.

Observed during a real review of a xianxu.dev blog post (the first heavy
`xx-fix` session). The agent's verbose reasoning — the actual rationale behind
each edit — is currently lost entirely; only the terse in-doc marker survives.

## Spec

A thin, **non-issue-bound** flow that wraps an `xx-fix` review in a git branch
and journals each round. Git is the only state store; the tool is guards +
orchestration over git, not a state machine. Decided to implement as a **shell
script** (sibling of `scripts/lib.sh`), NOT an `sdlc` verb: sdlc verbs are
400–900 lines hard-bound to `workshop/issues/` archiving with no precedent for a
file/branch-scoped flow (`ARCH-Simplicity` — don't drag issue/judge machinery
into a prose tool). The `xx-fix` skill gains a small "round journaling" section
that calls the script. Clean split: **xx-fix = mechanic (prose/markers), script
= side-effecting envelope (branch/commit/merge)** (`ARCH-PURE` — keep the prose
transform pure; isolate IO/git in the script).

### Verbs

- `start <file>...` — ensure a review branch exists (create `review/<slug>` from
  current branch, or **add the file(s) to the current review branch** if already
  on one — default-add, never silently guess; a new branch only on explicit
  request). `git add` + commit any untracked file so a brand-new draft (not yet
  in the repo) becomes tracked at round zero.
- `round --side human|agent [summary]` — internal, called by `xx-fix`. The
  per-round ping-pong: commit the working tree **before** the agent edits
  (author = human, captures incoming edits + markers), and **after** (author =
  agent, captures the agent's round). Two commits per round → clean attributable
  `git log`. Commit-message convention `review(<slug>): <side> rN — <summary>`
  so the whole thread is greppable; the agent's verbose rationale goes in the
  **commit body** (rescues the reasoning that today dies with the chat).
- `status` — loud about current branch, in-scope files, 🤖 count per file, rounds
  so far. Must make scope obvious before a `start` adds a doc to the wrong batch.
- `finish [--force]` — guard: **no 🤖 marker remains in any in-scope file**
  (`--force` skips the guard, mirroring sdlc's `--force`/`--no-<gate>`
  convention, and prints `merging with N outstanding 🤖 (published:false, won't
  render)`). Then `--no-ff` merge to the base branch and **delete the review
  branch**. `abandon` == `finish --force` (merge as-is, take the file however it
  stands); not a separate verb.

### History model (the resolved design)

- **No squash.** Squashing destroys the per-round revert points and the rationale
  in commit bodies. `finish` does a `--no-ff` merge so a single merge commit folds
  the round.
- **`--no-ff` + delete branch** gives both views with zero branch sprawl:
  `git log --first-parent main` = one clean line per reviewed doc (the merge
  commits); plain `git log` = every round (forensics). Deleting the branch loses
  nothing — round commits stay reachable as the merge commit's second parent.
- **One review branch = one ship-batch** = the set of docs merged together (e.g.
  several follow-up blog posts reviewed in one sitting). `finish` requires all
  in-scope docs clean.

### Scope / propagation

- Base-layer feature in ariadne (per operator). Skills are NOT auto-propagated
  (base.manifest only scaffolds empty `.claude/skills`; local skills sync from
  `construct/local/` via `sync-local-skills.sh`). A **script** can be added to
  `base.manifest` and propagate downstream (e.g. to xianxu.dev). xx-fix canonical
  source lives at `construct/local/fix/SKILL.md`.

## Done when

- A shell script implements `start` / `round` / `status` / `finish [--force]`
  over git, reusing `scripts/lib.sh` conventions (`set -euo pipefail`, color/output
  helpers).
- `xx-fix` SKILL.md (canonical source `construct/local/fix/SKILL.md`) gains a
  "round journaling" section instructing the agent to call the script before/after
  each round.
- `finish` refuses to merge a doc that still contains 🤖 markers unless `--force`.
- `git log --first-parent` shows one merge commit per reviewed doc; full log shows
  every round; no review branches left dangling after `finish`.
- An end-to-end dry run on a throwaway markdown file demonstrates start → 2 rounds
  → finish, verified by inspecting `git log --first-parent` vs full `git log`.
- atlas updated for the new flow + terminology; script registered in base.manifest
  for propagation.

## Plan

Single-pass atomic feature (one review boundary) → plain checkboxes, closes in
one `sdlc close`. ARCH cites: reuse `scripts/lib.sh` (`ARCH-DRY`); keep
string→string helpers (slug, marker-count, commit-msg, branch-name) separable
from the git IO seam, and keep the prose transform in `xx-fix` pure while the
script owns all git side effects (`ARCH-PURE`).

- [ ] `scripts/docflow.sh` — verbs `start` / `round --side` / `status` /
  `finish [--force]` over git; source `scripts/lib.sh` for colors + git helpers.
  Pure-ish helpers (slug-from-path, marker count via `rg '🤖'`, commit-message
  format, `review/<slug>` branch name) factored from the git-effecting parts.
- [ ] `start`: detect/create `review/<slug>`; default-add files to current review
  branch; `git add`+commit untracked drafts (round zero).
- [ ] `round --side human|agent [summary]`: two commits/round with author
  attribution, `review(<slug>): <side> rN — <summary>` subject + verbose body.
- [ ] `status`: branch, in-scope files, 🤖 count/file, rounds so far.
- [ ] `finish [--force]`: 🤖-marker guard (skip on `--force` with warning) →
  `--no-ff` merge to base → delete review branch.
- [ ] xx-fix `SKILL.md` (`construct/local/fix/SKILL.md`) — add "Round journaling"
  section instructing the agent to call `docflow start`/`round`/`finish`; sync to
  `.claude/skills/xx-fix/SKILL.md`.
- [ ] Register `scripts/docflow.sh` in `construct/base.manifest` for downstream
  propagation.
- [ ] e2e test: `scripts/docflow-test.sh` builds a throwaway git repo in
  `$TMPDIR`, runs start → 2 rounds → finish, asserts `git log --first-parent`
  shows one merge commit, full log shows every round, finish guard blocks on a
  lingering 🤖, and no review branch remains. (Real-git process-level test, no
  mocks — per AGENTS.md.)
- [ ] atlas: document the flow + terminology; link from `atlas/index.md`.

## Log

### 2026-06-03

Filed from a design brainstorm in the xianxu.dev session (heavy `xx-fix` review of
a blog post). Architecture fork (sdlc verb vs script vs skill) resolved toward a
thin script + xx-fix skill addition after an Explore pass showed sdlc verbs are
issue-coupled with no non-issue-bound precedent. Design (history model, verbs,
guards) settled with the operator before filing — see ## Spec.
