---
id: 000471
status: open
deps: []
github_issue:
created: 2026-10-19
updated: 2026-10-19
estimate_hours:
collaboration-mode: collaboration
card_mirror: '535587098bc39dd23e9cc86ce849b69eedc64edf' # card fields mirrored from issue-cards; edit via sdlc
---

# abandon: --issue N acted on the current (other issue's) branch and deleted it on origin

## Problem

2026-10-19, pair:1 (pair repo): `sdlc abandon --issue 694 --as punt --reason …` was run from the checkout of a *different*
issue's branch (#695). It acted on the current branch instead of #694's:
- committed #694's abandon log on top of #695's tip;
- saved that commit as `refs/ariadne/abandoned/000694` on origin;
- **deleted the #695 branch locally and on origin**;
- left #694's real branch untouched.

pair:1 repaired it by hand (recreated #695 at its tip, force-with-lease'd the archive ref to #694's real tip 45baebbbd, deleted
#694's branch). Nothing was lost only because the slot noticed. An abandon from the wrong checkout can delete another issue's
unmerged work on origin.

## Spec

`sdlc abandon --issue N` operates on issue N's branch, never on "the current branch" by assumption:
- If the checkout is on N's branch: today's behavior.
- If the checkout is on another issue's branch (or any branch that isn't N's): refuse, naming both, unless N's branch can be
  acted on without touching the checkout (archive its tip from the ref, delete it); prefer refusing in the first version.
- If N has no branch (planned but never started): abandon the card only.
- Never delete a branch whose tip isn't the archived commit. Check this before any push or delete (compare-and-swap on the ref).

Add the same guard to any other verb that takes `--issue N` but acts on HEAD's branch (audit and list them in the plan).

## Done when

- Fixture: on #B's branch, `abandon --issue A` refuses and changes nothing (local and origin refs unchanged).
- Fixture: on A's branch, today's behavior is unchanged.
- The audit of `--issue N` verbs that act on HEAD is in the plan, each either guarded or shown safe.

## Plan

- [ ]

## Log

### 2026-10-19
