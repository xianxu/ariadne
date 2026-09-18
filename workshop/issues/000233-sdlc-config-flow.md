---
id: 000233
status: open
deps: []
github_issue:
created: 2026-09-17
updated: 2026-09-17
estimate_hours:
---

# sdlc config: a review-free flow for declarative changes

## Problem

The third of three SDLC flows. `sdlc` is the full flow; `sdlc quick` (#231)
drops the plan, the plan-quality judge and the structured estimate and keeps one
close gate with a small-diff review. `sdlc config` asks whether even that close
review can go. Split out of #231 §5 at the operator's review on 2026-09-17.

Operator, 2026-09-17:

> small config change involves: 1/ update config; 2/ run test; 3/ update docs.
> really no need for even closing gate as structure of program didn't change and
> config is not Turing complete language typically, and thus much lower risk

The structural half of the argument holds cleanly: with no control flow and no
new state carried between events, ARCH-PURE, ARCH-ORDER and the structural half
of ARCH-DRY have nothing to bite on. Nothing was added that could drift. #231
§3's Done-when-freshness check is also usually trivially satisfied, because a
declarative change rarely reframes mid-stream.

**But "config" must not be the criterion, for #231's own stated reason.** #231 §1
gates on facts rather than judgment precisely because "is this small?" gets
answered wrong — parley.nvim#263 *looked like* one keybinding. "Is this just
config?" fails the same way, and worse, because a config line's consequence is
uncorrelated with its size. Three counterexamples that are all config and none
low-risk:

- **A moved external or authority surface.** A re-pointed remote, a widened
  permission, a path reaching outside the repo, a credential source. brain's own
  charter carries one as a standing prohibition — *"The gcrypt+GPG remote is not
  to be switched or re-pointed"* — and that is a single config line. Non-Turing-
  complete bounds STRUCTURAL risk; it says nothing about operational blast radius.
- **A value nothing asserts.** Step 2 ("run test") is the step carrying all the
  weight here, and it is the one most likely to be vacuous: a config value read
  once at startup and never asserted passes the suite whatever you set it to.
- **A restated single source.** Config is the classic ARCH-PURPOSE shadow-sweep
  case — the value changed in one file while a sibling still restates it. weave's
  settings-merge machinery exists because this keeps happening.

## Spec

**The admission predicate, in the same objective spirit as #231 §1:**

1. **No structural delta** — the diff touches only files in a repo-declared
   declarative set. Reuses #231's shared-surface declaration format rather than
   inventing a second one.
2. **The changed value is asserted.** Reverting it must fail the suite — a
   one-value mutation check, cheap for a single knob. This is the criterion that
   turns step 2 from an unverified ritual into the thing that earns the flow, and
   it puts the incentive in the right place: an unasserted knob does not qualify,
   which is a reason to assert it.
3. **No declared authority surface moved** — remotes, credentials, permissions,
   paths escaping the repo. Repo-declared, same mechanism as (1). This carries
   ARCH-SECURE's concern as an admission criterion, because on this flow there is
   no reviewer left to raise it.

Meeting all three, the close review is skipped and `sdlc close` records which
criteria admitted it.

Failing (1) or (2) routes UP to `sdlc quick` with the failing criterion named —
not to the full flow, and not a refusal. Failing (3) routes to the **full** flow
instead: #231 narrowed the quick recipe to ARCH-DRY, ARCH-PURE and ARCH-PURPOSE,
so a moved authority surface would get no security lens on the quick path.

There is precedent for content-scaled gating: #177 already auto-satisfies the
atlas gate on docs-only windows. This is the same move, one step further, with
the predicate written down.

## Done when

- `sdlc config` exists, admits only on all three criteria, and records which
  ones admitted it.
- Failing criterion (1) or (2) routes to `sdlc quick`, and failing (3) routes to
  the full flow; each names the failing criterion rather than refusing.
- The assertion criterion is mechanical, not attested: the flow verifies that
  reverting the changed value fails the suite.
- A diff that moves a declared authority surface is refused this flow, with a
  test fixture per surface kind (remote, credential, permission, escaping path).
- Config-flow rows are tagged in the ledger alongside quick-path rows, and are
  the FIRST thing reverted if escaped-defect rate rises on them — a flow with no
  review has no backstop but the ledger.

## Plan

- [ ] Decide whether the declarative set is a second list or a facet of #231's
      shared-surface declaration.
- [ ] Decide whether this ships with #231 or after calibration evidence from it
      (the Log below argues for after).
- [ ] Decide whether #231's operator route ("use sdlc quick path") has a
      counterpart here. With no reviewer, the three criteria are this flow's
      only backstop, which argues against an override that skips them.
- [ ] Implement admission: declarative-only diff, the revert-must-fail mutation
      check, and the authority-surface check.
- [ ] Wire `sdlc close` to skip the review for admitted issues and record the
      admitting criteria.
- [ ] Tag config-flow rows in the ledger.

## Log

### 2026-09-17

Split from #231 §5 at the operator's review, which named the three flows `sdlc`,
`sdlc quick` and `sdlc config`. The reasoning below is carried over from #231's
Log.

Recorded the structural half of the operator's argument as sound and the
file-type half as unsafe. **"Is it config" repeats exactly the judgment error
#231 §1 exists to avoid.** A config line's consequence is uncorrelated with its
size — a re-pointed remote is one line and unbounded. So the flow is admitted on
three mechanical facts (declarative-only diff, the value is actually asserted, no
declared authority surface moved) rather than on the file's extension.

The second criterion is the one worth keeping: it converts the operator's step 2
from an unverified ritual into the admission test itself, and an unasserted knob
failing to qualify is the correct incentive rather than an obstacle.

Open: whether this ships with #231 or waits for calibration evidence from the
quick path. A flow with no review has only the ledger as a backstop, which
argues for sequencing it second.

One consequence of the split: failing the authority-surface criterion now routes
to the full flow, not to `sdlc quick`, because #231's quick recipe no longer
checks ARCH-SECURE.
