---
type: side-doc
parent: ../000390-test-economy-pattern-plan.md
process-instruction: "[free-form] answer the operator's 🤖 comments in place"
---

# To land, must tests pass once? And flaky tests are tracked as separate issues?

_Side doc of [the plan](../000390-test-economy-pattern-plan.md) §3 (flake policy)._

**Yes to both, with one rule that keeps flakes from hiding real failures.** Proposed:

1. **Landing needs one passing full suite on the exact head** being landed. That doesn't change.
2. **A failure in a test that already has an open flake issue** (a known flake): the failed attempt is recorded against
   that issue (with its log; test-stats keeps it), and **a rerun on the same head that passes counts** as landing
   evidence. The landing isn't blocked by someone else's known flake.
3. **A failure that isn't a known flake blocks the landing until it's classified:**
   - if the branch caused it (it reproduces on the branch, not on main), it's the branch's bug: fix it;
   - if it is demonstrably intermittent (the same exact head fails then passes, on the branch or on main), the TL
     files a flake issue that day; from then on rule 2 applies, so a passing same-head rerun may admit the landing.
     A failure also seen on main **without** same-head fail→pass evidence is only a shared or suspected fault; a
     persistent main regression is not a flake and does not gain known-flake eligibility.
4. **Flakes are fixed as their own issues** (root cause, never a retry loop or a longer timeout), on the TL's daily
   standing list, like today's #423, #554, #562, #564, #568.

**Why not just "rerun until green":** a rerun that passes says nothing about why the first run failed. Requiring a named
flake issue first means every intermittent failure gets owned and fixed, while landings don't wait for the fix.

**What the plan text says now** is stricter ("a retry cannot convert an unexplained failure into passing evidence").
Rules 2–3 keep that ("unexplained" means not yet classified) but make the classified case explicit. If you agree, I'll
have ariadne:1 fold these four rules into §3.

**Decided (operator): approved.** The four rules go into §3.
