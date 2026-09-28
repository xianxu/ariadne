---
id: '000203'
status: done
started: 2026-08-22T10:04:13-07:00
created: 2026-08-22
updated: 2026-08-22
estimate_hours: 1.38
actual_hours: 1.57
---

# gate refusals tell the fixer to address findings, not to fix the class they belong to

## Problem

The `family:` mechanism (#194) is fully built on the *reviewer's* side and
absent on the *fixer's*. It has three specified consumers:

1. **the reviewer** — `judge/prompts/milestone-review.md` tells it to slug the
   underlying RULE not the symptom, reuse prior slugs verbatim, and (per
   `helptext/close.md`) escalate on a repeat from "fix this instance" to "state
   the rule that covers all of them";
2. **the ledger** — persists families and computes repeat counts;
3. **the operator** — the convergence line, `Not converging: fix rules, not
   instances.`

Nobody tells the agent *receiving* the findings to do anything differently. What
it actually reads at the moment it starts fixing is:

    changecode.go:554  address the findings above and re-run …
    close.go:1194      address them, or have the review dispose them explicitly …
    close.go:1237      address the findings, then re-run `sdlc close`
    close.go:1809      1. Fix the findings NOW, before committing this close.

"Address them" reads as "address each of them", which is precisely the per-site
patching the family machinery exists to detect. The convergence line is a
*diagnosis* emitted after the round, not a *procedure* offered before it — by
the time you read "Not converging" you have already spent the round.

**Evidence (parley.nvim#202).** Four boundary rounds against a cap of three,
with two families each surviving three separate rounds:

| family | findings | rounds |
|---|---|---|
| `invariant-without-regression-guard` | BR-2, BR-8, BR-13, BR-17 | 1, 1, 3, 4 |
| `stale-restatement-of-moved-source` | BR-4, BR-5, BR-9, BR-16, BR-18 | 1, 1, 1, 3, 4 |

Each round the agent fixed the site the finding named and stopped. Round 1's
BR-2 ("this invariant has no executable guard") would have retired BR-13 and
BR-17 too, had the agent enumerated *every invariant it had asserted in prose*
instead of guarding the one named. The rule was available — it was sitting in
the `family:` field — and was read as a label rather than as a worklist.

The tendency also shows one stage earlier: #202's first plan was the narrow
per-spec patch, inherited uncritically from the issue's own Done-when. The
general fix came from the plan gate (PQ-5), not from the planning agent.
