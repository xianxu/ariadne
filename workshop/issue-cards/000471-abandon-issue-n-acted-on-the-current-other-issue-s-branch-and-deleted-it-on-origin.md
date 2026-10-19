---
id: 000471
status: open
created: 2026-10-19
updated: 2026-10-19
estimate_hours:
github_issue:
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
