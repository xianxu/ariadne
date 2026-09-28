---
id: 000233
status: open
created: 2026-09-17
updated: 2026-09-17
estimate_hours:
github_issue:
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
