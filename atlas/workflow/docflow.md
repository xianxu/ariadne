# docflow — branch-scoped prose review

A thin git wrapper that turns an `xx-fix` co-authoring session (operator + agent
trading 🤖 markers in a markdown doc) into durable, attributable history. Each
review **round** is journaled as commits on a `review/<slug>` branch; on ship
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
| `start <file>...` | Create `review/<slug>` from the current branch (or add files to the current review branch — default-add, never silently guess). Records the base branch + each in-scope doc (deduped) as plain files under `.git/docflow/<slug>/` (`base`, `files`) — the source of truth for what `round` stages. Tracks an untracked draft as **round 0**. |
| `round --side human\|agent [-m sum] [--body …] [files…]` | Journal one side of a round. Two commits/round → attributable log: **human** rounds use the operator's git identity, **agent** rounds use `--author=$DOCFLOW_AGENT_AUTHOR` (default `Claude <noreply@anthropic.com>`) + a `Co-Authored-By` trailer. With no explicit files, stages **only the recorded in-scope docs** (`.git/docflow/<slug>/files`), never `git add -u` — so unrelated tracked WIP is never swept into the review (#81). Skips a no-op side. |
| `status` | Current branch, base, round counts, in-scope files + 🤖 count each. |
| `ship [--force]` | The explicit "land on main" act — *not* fired by marker-zero alone. **Guard:** refuses while any 🤖 remains in an in-scope file (so markers never ship). Then `--no-ff` merge to base + delete the review branch. `--force` merges as-is — the "abandon" path. Alias: `finish` (deprecated, warns then calls `ship`). |

Commit subject convention (greppable): `review(<slug>): <side> r<N> — <summary>`.

State lives in plain files under `.git/docflow/<slug>/`, **not** `.git/config`: the
Claude Code Bash sandbox denies `.git/config` (and `.git/hooks/`) writes — they can
execute code — while plain files under `.git/` are writable. Storing state there
keeps docflow fully sandbox-compatible (no `dangerouslyDisableSandbox` per call) and
leaves the `.git/config` security boundary intact (#84). The dir is resolved via
`git rev-parse --git-dir` (worktree-local).

## History model — why no squash, yet no branch sprawl

`ship` does a **`--no-ff`** merge and **deletes** the review branch. That single
choice gives both views with zero cost:

- `git log --first-parent <base>` → **one clean line per reviewed batch** (the
  merge commits). The "what shipped" view.
- plain `git log <base>` → **every round** (the forensic view), plus each round's
  rationale in its body.

Deleting the branch loses nothing: the round commits stay reachable as the merge
commit's second parent. No squash (which would destroy per-round revert points and
the bodies), no hundreds of stale branches.

**One review branch = one ship-batch** = the set of docs merged together (e.g.
several follow-up posts reviewed in one sitting). `ship` requires all in-scope
docs marker-clean.

## Where it fits

Opt-in. Plain `xx-fix` marker-processing needs none of this; reach for `docflow`
on heavier, multi-round co-authoring where the trail is worth keeping. The `xx-fix`
SKILL.md "Round journaling" section tells the agent when to call each verb.

## Pair Review Workbench

In a hosted pair review pane, `xx-fix` is the agent half of pair's document
workbench. The agent still owns docflow branch/round/ship operations, but it does
not edit the file directly: it writes `{old, occurrence, new, explain}` records
to pair's handoff seam, waits for the pane to apply the records undo-ably, then
commits the landed artifact with `docflow round`. The mode vocabulary for that
hosted flow is Generate / Edit / Proofread; fact-check remains an instruction
that dispatches `doc-review` and folds accepted findings back through records.
