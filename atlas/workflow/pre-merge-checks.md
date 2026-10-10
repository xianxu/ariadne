# Pre-merge Checks (two-gate model, #160)

## Purpose

Automated constitution enforcement before code lands on main. Since #160 the
enforcement is split into **two gates** on different verbs:

- **`sdlc close` — the LOCAL acceptance gate (all LLM review).** The fresh-context
  boundary review (`code-review.md`) runs here: code quality, requirements
  traceability, architecture, and the **Docs update gate (atlas + README)**. On a
  finalizing verdict it flips the issue `working → codecomplete`. This is the *only*
  place LLM review runs. The fresh reviewer receives a pinned repository manifest
  with immutable refs and read-only Git commands, not an inline unified patch;
  this keeps dispatch bounded while preserving repository-based inspection.
- **`sdlc merge` / `push` — the deterministic PUBLISH gate (no LLM).** They enforce
  the **reviewed-HEAD-unchanged invariant** (`runPublishGate`, `cmd/sdlc/publishgate.go`):
  refuse unless HEAD is unchanged since the codecomplete issues' `sdlc close` (i.e.
  nothing drifted after the review), then flip `codecomplete → done` and archive.

This folds in #142 (pre-merge judges should run at the earliest useful gate): the
old merge-time `plan`/`specs` LLM judges duplicated the close boundary review and
fired late — they are **removed** from merge/push. `lessons` (a no-LLM reminder)
moved to close.

## The reviewed-HEAD-unchanged invariant

`codecomplete ⟹ the close boundary review covered HEAD`. The anchor is the newest
commit that leaves the issue at `status: codecomplete` (a content read of the
issue file's git history). Because `sdlc close` is the **sole writer** of
`codecomplete` (`set-status` refuses it), that commit is a trustworthy anchor; a
re-close after drift produces a newer such commit, so the anchor advances. In a
tracker repository the anchor is the card binding's evidence commit. Deterministic —
it *forces* a real re-review rather than silently re-judging a delta (the elegance
that replaced #142's proposed post-close LLM delta-review).

**On the branch patch, not a commit count (#304, A4).** The gate replays each issue's
reviewed patch (the anchor, against its own merge-base with main) onto today's main
with `gitx.RebasedReviewedBase` and diffs that against HEAD (`classifyPublishDelta`).
Merging or rebasing main after close therefore needs no re-close. Only the branch's
own post-close code refuses, and the refusal names those paths, never main's. A
conflict refuses with its paths, because the resolution is unreviewed. On a branch
carrying several closes, any reviewed patch that covers HEAD passes: the newest
close reviewed the whole branch patch. That generalizes the old newest-anchor rule.
Refusals share the `publishGateRefusal` prefix, which `gatesig` keys on.

**The close binding survives a rebase (#304, B3).** A tracker close's evidence
message carries `Close-Token: <binding token>`. A rebase rewrites the evidence commit
but never its message, so `ownedCompletions` still owns the close: its evidence is in
the head, or a commit in range carries the card's *current* token. One rule serves
publish, PR landing and settle. A reopen-and-close mints a new token, so an older
close never claims a newer one. A codecomplete card this branch's patch touches, but
that no rule owns, refuses loudly (`refuseUnownedCompletions`), for example after a
squash dropped the token. Before #304 a rebase made the card unowned and the gate
reported "nothing to verify", which failed open. The reviewed commit stays resolvable
through `refs/sdlc/reviewed/<id>/<boundary>`, which is removed at done, at abandon and
by both legacy archives, and swept by settle. A clone without it refuses with a
re-close. **Known limit:** a legacy (no-tracker) repository has no sweep, because pin
refs are shared across worktrees and no single checkout can tell whether one is live.
A hand-archived legacy issue therefore keeps its (milestones + 1) pins until someone
deletes them with `git update-ref -d`. A sweep would need liveness across every
worktree (`git worktree list`) and an age bound (#304 close review, round 4).

**Doc-only tolerance (#174).** A post-anchor delta with **no code surface**
(`publishGateHasCodeSurface`: the #177 `hasCodePath` docs classifier — `*.md`,
`workshop/`, `atlas/`, `docs/` — tightened so anything under `cmd/` is code
even when `*.md`, because helptext is embedded binary surface) passes with an
info line instead of refusing — post-close bookkeeping
(lessons, plan ticks, atlas notes) is the friction shape #172 measured at 6/6
publish-gate refusals, and docs don't weaken the review's claims about code
behavior. Code deltas still refuse. The sanctioned protocol after a
FIX-THEN-SHIP verdict is printed by close itself (#174): fix findings before
committing, bundle everything into the one close commit, don't re-run close
(the quick flow's one exception is below);
the reclose guard (fires at `done` only) points post-publish follow-ups at a
new issue.

**Quick-flow re-measure (#231).** Those bundled fixes are never measured by
close, which sized the head its small-diff review saw. So for a codecomplete
issue on the quick flow the gate measures the final diff over close's own window
(`boundaryWindowBase` → HEAD) and, past `flow.MaxAddedLinesAfterReview` (twice
the shell's line limit), refuses and sends the issue back to `sdlc close` —
which finds the shell crossed, upgrades the issue and runs the full review. It
runs on both pass paths (HEAD at the anchor, and the doc-only delta) and stays
LLM-free; the FIX-THEN-SHIP protocol close prints names the exception on the
quick flow.

## The judge categories (now ad-hoc / close-time)

The LLM judge categories still exist as **standalone** `sdlc judge <cat>` commands
(ad-hoc use), and the review-shaped ones are embedded in the close boundary review —
they are no longer auto-dispatched at merge/push:

| Name | Where it runs now |
|------|-------------------|
| dry / pure | ad-hoc `sdlc judge`; the boundary review covers architecture at close |
| plan | folded into the close boundary review (requirements traceability) + the #124 conformance gate |
| specs | folded into the close boundary review's Docs update gate (atlas + README) |
| lessons | the no-LLM reminder ping, emitted at `sdlc close` (#160 Q4) |
| small-diff-review | close-time only (#231): the quick flow's one review, dispatched by `sdlc close` for a quick issue inside the shell — not a standalone `sdlc judge` check |

## Complementary tier: the CI merge-check

Separately, a **deterministic, server-side, repo-specific** CI merge-check
(`scripts/merge-checks.d/*`) gates the PR — see `ci-merge-check.md`. That is NOT an
LLM and is orthogonal to the publish gate above (a substrate gate / lint / test each
derivative plugs in).
