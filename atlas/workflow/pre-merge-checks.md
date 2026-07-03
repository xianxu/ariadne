# Pre-merge Checks (two-gate model, #160)

## Purpose

Automated constitution enforcement before code lands on main. Since #160 the
enforcement is split into **two gates** on different verbs:

- **`sdlc close` — the LOCAL acceptance gate (all LLM review).** The fresh-context
  boundary review (`code-review.md`) runs here: code quality, requirements
  traceability, architecture, and the **Docs update gate (atlas + README)**. On a
  finalizing verdict it flips the issue `working → codecomplete`. This is the *only*
  place LLM review runs.
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
re-close after drift produces a newer such commit, so the anchor advances. merge/push
refuse if any commit landed after the latest anchor, pointing you to re-run `sdlc
close`. Deterministic — it *forces* a real re-review rather than silently re-judging
a delta (the elegance that replaced #142's proposed post-close LLM delta-review).

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

## Complementary tier: the CI merge-check

Separately, a **deterministic, server-side, repo-specific** CI merge-check
(`scripts/merge-checks.d/*`) gates the PR — see `ci-merge-check.md`. That is NOT an
LLM and is orthogonal to the publish gate above (a substrate gate / lint / test each
derivative plugs in).
