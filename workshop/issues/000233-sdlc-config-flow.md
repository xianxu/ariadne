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

The third of three SDLC flows. `full` is today's flow; `quick` (#231) drops the
plan, the plan-quality judge and the structured estimate and keeps one close gate
with a small-diff review. `config` asks whether even that close review can go.
As in #231, a flow is recorded in frontmatter (`flow: {kind: config,
provenance: …}`) and acted on by the gates; it is not a verb. Split out of #231 §5 at the operator's review on
2026-09-17.

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

Like #231's hard shell, all three are facts about the diff, so `sdlc close`
checks them. Meeting all three, the close review is skipped and `sdlc close`
records which criteria admitted it.

Failing one upgrades the issue, as #231's gates do, naming the criterion in the
Log; `sdlc close` then runs that flow's review. For (1) or (2) the upgrade is to
`quick`, one flow up, not to the full flow. For (3) it is to `full` instead: #231
narrowed the quick recipe to ARCH-DRY, ARCH-PURE and ARCH-PURPOSE, so a moved
authority surface would get no security lens on the quick flow.

There is precedent for content-scaled gating: #177 already auto-satisfies the
atlas gate on docs-only windows. This is the same move, one step further, with
the predicate written down.

## Done when

- `#Flow.kind` in `construct/vocabulary/issue.cue` gains `config`, and `sdlc
  change-code --flow config` pins it (`provenance: operator`).
- `sdlc close` accepts a config-flow issue without a review only when all three
  criteria hold, and records which ones admitted it.
- Failing a criterion upgrades the issue — to `quick` for (1) or (2), to `full`
  for (3) — naming the criterion, and `sdlc close` runs that flow's review.
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
- [ ] Decide whether `config` can be inferred. #231 infers at `change-code`,
      before the diff exists, and its gates only ever upgrade, so inferring
      `config` from the diff at close would be the one downward move. Until
      decided, `config` is operator-pinned only, and the three criteria still
      bind at close.
- [ ] Implement admission: declarative-only diff, the revert-must-fail mutation
      check, and the authority-surface check.
- [ ] Add `config` to `#Flow.kind`; let `change-code --flow config` pin it.
- [ ] Wire `sdlc close` to check the criteria on `kind: config`, skip the review
      when they hold, record the admitting criteria, and upgrade the issue when
      one fails.
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

## Revisions

### 2026-09-17 — flows are frontmatter, set by existing verbs

Reason: #231 decided that a flow is a `flow:` frontmatter field set by `sdlc
change-code --flow`, modeled in the cue schema and read by every later gate,
with no new verbs.

Delta: `sdlc config` is now `flow: config`, extending #231's `#Flow`. Admission
is checked at close, like #231's hard shell. A failed criterion makes `close`
refuse with a re-route (`--flow quick` or `--flow full`) instead of routing
silently. Replaced the operator-override question with the matching #231 rule:
the operator may declare the flow, and the criteria still bind at close.

### 2026-09-17 — the gates infer the flow; config is pinned

Reason: #231 now has the gates infer the flow, record it as `{kind,
provenance}`, and only ever upgrade it, with no refusals on crossing a limit.

Delta: `flow: config` is now `kind: config`, pinned by the operator. A failed
criterion upgrades the issue (to `quick` or `full`) instead of refusing. Whether
`config` can be inferred at all is open (Plan): it would be the one downward
move against #231's upgrade-only rule.
