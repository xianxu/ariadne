---
id: 000231
status: open
deps: []
github_issue:
created: 2026-09-16
updated: 2026-09-16
estimate_hours:
---

# A quick path for small diffs: scale the gate set and the review recipe to size

## Problem

The SDLC applies one gate set to every issue regardless of size. For a
well-defined ~100-line change that is not merely *expensive* — it is aimed at
the wrong bugs.

The observation that motivates this, from the operator:

> It is a trade-off. If there are only 100 lines of code, then certain bugs
> would be rare. This means the quick path should change **the type of review
> we do**, and potentially whether we write a plan down.

That is the sharp version, and it is stronger than "skip some steps". **Diff
size changes the bug distribution**, so it should change the review *recipe*,
not just the review *budget*. Architectural drift, cross-module inconsistency,
state-machine ordering faults and single-source sweeps all need size or spread
to exist. Boundary errors, a clause asserted in one of the two modes it names,
a doc sentence that contradicts the code, and a paste of an existing preamble
do not — they are as likely in 100 lines as in 1000, arguably likelier, because
a small change is where people skip the sweep.

### Evidence: parley.nvim#263 (measured, whole issue)

A single keybinding feature. Shipped 966 lines of prod+test; the feature module
itself is 86 lines.

| | |
|---|---|
| `workshop/` + `atlas/` churn | 1,516 lines |
| Durable plan | 1,282 lines, for an 86-line module |
| LLM gate rounds | 7 (2 plan-quality + 4 close + 1 re-close) |
| est / actual | 2.04 / 4.31 (**0.5×**) |

**Every ARCH-\* lens passed, every round.** DRY was flagged; PURE, PURPOSE,
MOCK, CONSTRAINTS, SECURE, ORDER and FUNERAL passed from round 1 to round 4.
The architecture judging was correct and found nothing, four times, because at
this size there was nothing of that kind to find.

What the boundary review *did* find, across four rounds:

| Family | Kind |
|---|---|
| `acceptance-clause-untested` | a clause naming two modes asserted in one |
| `doc-claim-contradicts-code` ×2 | a count asserted from memory; then the corrected count still unenforced |
| orphaned `@param` block (Critical) | a code hoist separated a doc block from its function |
| `duplicated-command-preamble` ×3 | copy-paste of an existing sequence, then two half-sweeps |
| `stale-close-evidence`, `prose-continuity`, `undocumented-branch` | doc-truth |

All local-correctness, doc-truth, and duplication-adjacency. Not one of them
required a plan to find, and not one is an architectural finding.

Meanwhile the **plan-quality** review's findings were largely defects *in the
plan artifact itself*: an assertion in the plan's literal test code that could
never pass, a cited "model this on X" file that did not use the technique
attributed to it, a step whose named oracle does not run under the spec key it
names. Those defects existed **because** a 1,282-line plan existed. Remove the
artifact and the defect class goes with it.

That asymmetry is the case: the plan review mostly policed the plan; the close
review found everything that mattered about the code.

## Spec

Add a **direct path** to `sdlc change-code`, and give the close boundary a
review recipe matched to small diffs.

### 1. Admission — objective, not judgment

"Is this small?" is exactly the call that gets made wrong (#263 looked like one
keybinding). Gate on facts the binary can check:

- The issue has a `## Spec` and `## Done when` a fresh reader could test against
  — the structural check `change-code` already runs.
- No `Mx` milestones (single boundary).
- Estimate below a threshold.
- **No shared-surface change.** A repo declares its shared surfaces (for
  parley: the keybinding registry, `config.lua`'s option schema,
  `construct/vocabulary/*.cue`, any cross-module seam). This is the criterion
  that would have refused #263 — the *feature* was one chord, but it touched the
  registry, and registry-shaped work grows guard generalizations and
  cross-module extractions whatever its headline size.

Failing admission is not a refusal, it is a routing decision: take the full path.

### 2. What the direct path drops

The durable plan and the plan-quality judge. Not the estimate (calibration needs
every row), not the structural checks, not the close gates.

### 3. What must survive — the part that is not optional

**The acceptance contract.** On the direct path the close review is the *only*
review, and `sdlc close` judges the diff against `## Done when`. In #263 the one
blocking plan-gate finding was that a mid-stream reframe changed what got built
without restating Done-when — three of five clauses became unsatisfiable *by
design*. On the full path the plan gate caught that. On a direct path nothing
would, and the single remaining review would judge against stale criteria.

So the direct path keeps a **deterministic** (no-LLM) check: if the issue body
changed after `claim`, `## Done when` must have changed too, or the agent
acknowledges it did not need to. Cheap, and it protects the one oracle left.

### 4. The review recipe — the heart of this issue

Add `cmd/sdlc/internal/judge/prompts/small-diff-review.md` beside the existing
`plan.md` / `milestone-review.md` / `specs.md` recipes. It is **not** a lighter
milestone-review; it is aimed at a different distribution:

- **Amplify:** boundary and off-by-one conditions; every clause of Done-when
  that names two states/modes asserted in *both*; doc and comment claims checked
  against the code they describe (counts, "the only", "always"); the diff's
  immediate neighbourhood — the arch-spec classes a small edit disturbs, such as
  a hoist separating a doc block from its function; and whether any block in the
  diff already exists elsewhere in the tree.
- **Attenuate:** the full ARCH-\* sweep. Keep SECURE and ORDER as tripwires with
  a stated escalation ("if this diff crosses a trust or state boundary, stop and
  request the full recipe"); drop the rest to a one-line N/A unless the diff
  adds a module or a seam.
- **Cap the tail.** #263's four rounds were mostly *repeat families* caused by
  fixing the site a finding named. The recipe should require, on the **first**
  finding in a family, the enumeration of that family before any fix. A small
  diff has a small enumeration — this is cheap exactly where it is most often
  skipped.

### 5. A second tier below the direct path: the declarative change

Operator, 2026-09-17:

> small config change involves: 1/ update config; 2/ run test; 3/ update docs.
> really no need for even closing gate as structure of program didn't change and
> config is not Turing complete language typically, and thus much lower risk

This is a THIRD tier — the direct path in §1-4 still keeps the close gates; this
one questions them. The structural half of the argument holds cleanly: with no
control flow and no new state carried between events, ARCH-PURE, ARCH-ORDER and
the structural half of ARCH-DRY have nothing to bite on. Nothing was added that
could drift. §3's Done-when-freshness check is also usually trivially satisfied,
because a declarative change rarely reframes mid-stream.

**But "config" must not be the criterion, for this issue's own stated reason.**
§1 gates on facts rather than judgment precisely because "is this small?" gets
answered wrong — #263 *looked like* one keybinding. "Is this just config?" fails
the same way, and worse, because a config line's consequence is uncorrelated with
its size. Three counterexamples that are all config and none low-risk:

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

**So the admission predicate, in the same objective spirit as §1:**

1. **No structural delta** — the diff touches only files in a repo-declared
   declarative set. Reuses §1's shared-surface declaration format rather than
   inventing a second one.
2. **The changed value is asserted.** Reverting it must fail the suite — a
   one-value mutation check, cheap for a single knob. This is the criterion that
   turns step 2 from an unverified ritual into the thing that earns the tier, and
   it puts the incentive in the right place: an unasserted knob does not qualify,
   which is a reason to assert it.
3. **No declared authority surface moved** — remotes, credentials, permissions,
   paths escaping the repo. Repo-declared, same mechanism as (1). This is the
   ARCH-SECURE tripwire from §4 promoted to an admission criterion, because on
   this tier there is no reviewer left to trip it.

Meeting all three, the close review is skipped and `sdlc close` records which
criteria admitted it. Failing any one routes UP to the direct path (not to the
full path) with the failing criterion named.

There is precedent for content-scaled gating: #177 already auto-satisfies the
atlas gate on docs-only windows. This is the same move, one tier further, with
the predicate written down.

## Done when

- `sdlc change-code --direct` exists, admits only on the objective criteria
  above, and records in the issue that the direct path was taken and why it
  qualified.
- Admission failure routes to the full path with the failing criterion named.
- A repo can declare its shared surfaces; a diff touching one is refused the
  direct path.
- The Done-when-freshness check runs deterministically on the direct path.
- `small-diff-review.md` exists as its own recipe, and `sdlc close` selects it
  when the issue was admitted to the direct path.
- The recipe requires family enumeration on the first finding, and the gate
  ledger shows repeat families dropping on direct-path issues.
- A **declarative tier** exists below `--direct`, admitting only on all three
  §5 criteria, recording which ones admitted it, and routing UP to `--direct`
  (naming the failing criterion) rather than refusing.
- The assertion criterion is mechanical, not attested: the tier verifies that
  reverting the changed value fails the suite.
- A declarative-tier diff that moves a declared authority surface is refused,
  with a test fixture per surface kind (remote, credential, permission, escaping
  path).
- Declarative-tier rows are tagged in the ledger alongside direct-path rows, and
  are the FIRST thing reverted if escaped-defect rate rises on them — a tier with
  no review has no backstop but the ledger.
- Calibration: direct-path rows are tagged in the ledger so their est/actual
  ratio can be compared against full-path rows. If the ratio does not improve,
  the trade-off was wrong and this gets reverted on evidence.

## Plan

- [ ] Decide the admission predicate and the shared-surface declaration format
      (§1), and whether the declarative set (§5) is a second list or a facet of it.
- [ ] Decide whether the declarative tier ships with §1-4 or after evidence from it.
- [ ] Implement `--direct` + the deterministic Done-when-freshness check.
- [ ] Write `small-diff-review.md`; wire recipe selection at close.
- [ ] Tag direct-path rows in the calibration ledger.
- [ ] Re-run the recipe against parley.nvim#263's actual diff as a fixture: it
      should surface BR-1, BR-2 and BR-9 (the classes it is built for) without
      four rounds of ARCH-\* passes.

## Log

### 2026-09-16

Filed from a parley.nvim#263 retrospective. The operator's framing is the
issue's thesis and is quoted verbatim above: fewer lines means certain bug
classes become rare, so the quick path changes **the type of review**, not just
its cost.

Adjacent, not duplicate: **#228** (align local superpowers with SDLC review
gates) removes *duplicate review authority* between the skills and the gates.
This one scales *which gates run and what the reviewer looks for* to the size of
the change. #228 could land first without blocking this.

Counter-evidence to weigh honestly before building: in #263 the plan was **not**
the dominant cost — the review tail was (4 close rounds, 3 of them repeat
families from partial fixes). A direct path alone would have saved the 1,282-line
plan and its 2 gate rounds; it would **not** have saved the tail. That is why
the family-enumeration rule is part of this issue rather than a separate one,
and why the Done-when criterion above is the acceptance test that matters.

### 2026-09-17 — second use case: the declarative change, and a third tier

Added §5 from an operator observation (quoted there): a small config change is
update-config / run-test / update-docs, with no structural change and therefore,
the argument goes, no need for a closing gate at all.

Recorded the structural half as sound and the file-type half as unsafe. The
correction that made it into §5: **"is it config" repeats exactly the judgment
error §1 exists to avoid.** A config line's consequence is uncorrelated with its
size — a re-pointed remote is one line and unbounded. So the tier is admitted on
three mechanical facts (declarative-only diff, the value is actually asserted, no
declared authority surface moved) rather than on the file's extension.

The second criterion is the one worth keeping: it converts the operator's step 2
from an unverified ritual into the admission test itself, and an unasserted knob
failing to qualify is the correct incentive rather than an obstacle.

Open, and deliberately not decided here: whether this tier ships with §1-4 or
waits for calibration evidence from the direct path. A tier with no review has
only the ledger as a backstop, which argues for sequencing it second.
