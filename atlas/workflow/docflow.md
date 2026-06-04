# docflow — branch-scoped prose review

A thin git wrapper that turns an `xx-fix` co-authoring session (operator + agent
trading 🤖 markers in a markdown doc) into durable, attributable history. Each
review **round** is journaled as commits on a `review/<slug>` branch; on finish
the branch `--no-ff` merges back to its base. Introduced in #79; companion to the
`xx-fix` skill. Lives at `scripts/docflow.sh` (propagated via `base.manifest`).

## Why

`xx-fix` works well, but a heavier multi-round review had no memory: each round
overwrote the last, and the agent's *rationale* lived only in the chat transcript,
which is gone within days. `docflow` rescues both — the per-round diffs **and** the
reasoning (in commit bodies) — without a bespoke versioning scheme. Git already
does this; the script is guards + orchestration over git, not a state machine.

## The split (ARCH-PURE)

- **`xx-fix` skill** = the prose *mechanic* (parse/apply/reply to markers). Pure.
- **`docflow.sh`** = the *envelope* that owns all git side effects (branch / commit
  / merge). Pure string→string helpers (`slugify`, `round_subject`,
  `review_branch_name`, `marker_count`) are factored from the git seam and unit-
  tested directly in `scripts/docflow.test.sh`; the e2e in that file drives a real
  throwaway repo (no mocks).

## Verbs

| Verb | What |
|---|---|
| `start <file>...` | Create `review/<slug>` from the current branch (or add files to the current review branch — default-add, never silently guess). Records the base branch in `git config branch.<rb>.docflowBase`. Tracks an untracked draft as **round 0**. |
| `round --side human\|agent [-m sum] [--body …] [files…]` | Journal one side of a round. Two commits/round → attributable log: **human** rounds use the operator's git identity, **agent** rounds use `--author=$DOCFLOW_AGENT_AUTHOR` (default `Claude <noreply@anthropic.com>`) + a `Co-Authored-By` trailer. Stages tracked changes only (`git add -u`) so it never sweeps untracked junk. Skips a no-op side. |
| `status` | Current branch, base, round counts, in-scope files + 🤖 count each. |
| `finish [--force]` | **Guard:** refuses while any 🤖 remains in an in-scope file (so markers never ship). Then `--no-ff` merge to base + delete the review branch. `--force` merges as-is — the "abandon" path. |

Commit subject convention (greppable): `review(<slug>): <side> r<N> — <summary>`.

## History model — why no squash, yet no branch sprawl

`finish` does a **`--no-ff`** merge and **deletes** the review branch. That single
choice gives both views with zero cost:

- `git log --first-parent <base>` → **one clean line per reviewed batch** (the
  merge commits). The "what shipped" view.
- plain `git log <base>` → **every round** (the forensic view), plus each round's
  rationale in its body.

Deleting the branch loses nothing: the round commits stay reachable as the merge
commit's second parent. No squash (which would destroy per-round revert points and
the bodies), no hundreds of stale branches.

**One review branch = one ship-batch** = the set of docs merged together (e.g.
several follow-up posts reviewed in one sitting). `finish` requires all in-scope
docs marker-clean.

## Where it fits

Opt-in. Plain `xx-fix` marker-processing needs none of this; reach for `docflow`
on heavier, multi-round co-authoring where the trail is worth keeping. The `xx-fix`
SKILL.md "Round journaling" section tells the agent when to call each verb.
