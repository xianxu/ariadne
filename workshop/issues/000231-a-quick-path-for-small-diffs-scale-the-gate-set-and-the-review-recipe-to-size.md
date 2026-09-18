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

Add a **quick path**, `sdlc quick`, and give its one gate a review recipe matched
to small diffs. It is one of three flows: `sdlc` (the full flow), `sdlc quick`
(this issue), and `sdlc config` for declarative changes (#233).

### 1. Admission — objective facts, or the operator's word

There are two ways onto the quick path.

**Automated.** "Is this small?" is exactly the call that gets made wrong (#263
looked like one keybinding), so the automated route gates on facts the binary can
check. Two are facts about the issue, checked at entry:

- The issue has a `## Spec` and `## Done when` a fresh reader could test against.
- No `Mx` milestones (single boundary).

Two are facts about the diff, which does not exist at entry, so `sdlc close`
checks them against the real diff:

- **No shared-surface change.** A repo declares its shared surfaces (for parley:
  the keybinding registry, `config.lua`'s option schema,
  `construct/vocabulary/*.cue`, any cross-module seam). This is the criterion
  that would have refused #263 — the *feature* was one chord, but it touched the
  registry, and registry-shaped work grows guard generalizations and
  cross-module extractions whatever its headline size.
- **Diff within a size budget.** This replaces an estimate threshold, which the
  quick path cannot use because it no longer estimates (§2).

Failing either check is a routing decision, not a refusal. At entry, the issue
takes the full path. At close, the review escalates from the quick recipe (§4)
to the full one.

**Operator instruction.** The operator can put an issue on the quick path
directly ("use sdlc quick path"). That is a decision, not a guess for the binary
to second-guess: entry skips the automated checks, and at close the diff checks
still run but only *name* a touched shared surface or an exceeded budget in the
close output. They do not escalate.

Either way, the issue records that it took the quick path and which route
admitted it (`criteria` or `operator`), so the ledger can tell the two apart.

### 2. What the quick path drops

- The durable plan and the plan-quality judge.
- The structured estimate: `estimate_hours` and the `## Estimate` block.
  `actual_hours` is still measured at close, because measuring it costs nothing.
  With no estimate to pair it with, though, quick-path rows leave est/actual
  velocity calibration rather than entering it half-filled.
- `change-code`'s structural and plan-quality gates. Whatever verb enters the
  quick path still owns the branching decision `change-code` owns today.

What remains is **one gate: `sdlc close`**, with its evidence (`--verified`, the
measured actual) and the quick review recipe (§4).

### 3. What must survive — the part that is not optional

**The acceptance contract.** On the quick path the close review is the *only*
review, and `sdlc close` judges the diff against `## Done when`. In #263 the one
blocking plan-gate finding was that a mid-stream reframe changed what got built
without restating Done-when — three of five clauses became unsatisfiable *by
design*. On the full path the plan gate caught that. On the quick path nothing
would, and the single remaining review would judge against stale criteria.

So `sdlc close` runs two **deterministic** (no-LLM) checks on a quick-path issue
before its review:

- `## Done when` is non-empty. The structural gate that used to guarantee this
  is gone (§2), and it is the review's only oracle.
- If the issue body changed after `claim`, `## Done when` must have changed too,
  or the agent acknowledges it did not need to. Cheap, and it protects the one
  oracle left.

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
- **Architecture: exactly ARCH-DRY, ARCH-PURE and ARCH-PURPOSE.** Every other
  ARCH-\* lens is off, with no tripwires. The three are *selected by marker* from
  the single-source registry `cmd/sdlc/internal/judge/architecture.md` (a subset
  selector beside `ArchitectureBlock`), never restated in the recipe; restating
  them would be the very ARCH-DRY defect the recipe hunts.
- **Cap the tail.** #263's four rounds were mostly *repeat families* caused by
  fixing the site a finding named. The recipe should require, on the **first**
  finding in a family, the enumeration of that family before any fix. A small
  diff has a small enumeration — this is cheap exactly where it is most often
  skipped.

When the close-time diff checks (§1) escalate an automatically admitted issue,
`sdlc close` runs the full recipe instead.

### 5. Split out: the declarative change

The declarative tier — a config change admitted on mechanical facts, with no
close review at all — is now its own flow, `sdlc config`, tracked in #233. This
issue does not depend on it. #233 builds on §1's shared-surface declaration.

## Done when

- `sdlc quick` exists and can be entered two ways: automatically, when the
  issue passes §1's entry checks, or by operator instruction, which skips them.
  The issue records that it took the quick path and which route admitted it.
- On the quick path, no durable plan, plan-quality judge, structured estimate or
  `change-code` structural gate runs; `sdlc close` is the only gate.
- A failed automated entry check routes to the full path, naming the failing
  criterion.
- A repo can declare its shared surfaces. At close, a quick-path diff that
  touched one or exceeded the size budget escalates to the full review recipe if
  it was admitted automatically, or is named in the close output if the
  operator admitted it.
- `sdlc close` refuses a quick-path issue with an empty `## Done when`, and runs
  the Done-when-freshness check deterministically.
- `small-diff-review.md` exists as its own recipe, and `sdlc close` selects it
  for quick-path issues.
- The recipe's architecture section is ARCH-DRY, ARCH-PURE and ARCH-PURPOSE,
  selected by marker from `judge/architecture.md`. A test asserts the rendered
  recipe carries exactly those three markers, so a registry edit cannot silently
  widen or drift it.
- The recipe requires family enumeration on the first finding, and the gate
  ledger shows repeat families dropping on quick-path issues.
- Calibration: quick-path rows are tagged in the ledger and excluded from
  est/actual calibration, since they carry no estimate. The trade-off is judged
  on what they do carry — gate rounds per issue, measured actual hours, and
  escaped defects (follow-up fixes citing a quick-path issue) — against
  comparable full-path rows. If those do not improve, the trade-off was wrong
  and this gets reverted on evidence.

## Plan

- [ ] Decide the entry surface for `sdlc quick`: its own verb, or a mode on
      `claim`/`change-code`. Either way it owns the branching decision.
- [ ] Decide where the quick-path marker and admitting route live: a Log line or
      frontmatter. Frontmatter is a `construct/vocabulary/issue.cue` change.
- [ ] Decide the shared-surface declaration format and the diff size budget (§1).
- [ ] Implement quick-path entry (automated checks + operator route) and the
      close-time diff checks with their escalation.
- [ ] Implement the two deterministic Done-when checks at close (§3).
- [ ] Add a marker-subset selector beside `ArchitectureBlock`; write
      `small-diff-review.md` on it; wire recipe selection at close.
- [ ] Tag quick-path rows in the calibration ledger and exclude them from
      est/actual calibration.
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

## Revisions

### 2026-09-17 — operator review: three flows, a narrower quick path

Reason: the operator reviewed the Spec. Their decisions: the quick path skips the
structured estimate too; its only gate is `sdlc close`; that review checks
ARCH-DRY, ARCH-PURE and ARCH-PURPOSE and nothing else; the operator can route an
issue onto the quick path by instruction, beside the automated admission; and
the declarative tier becomes its own flow in a separate issue. That makes three
flows: `sdlc`, `sdlc quick`, `sdlc config`.

Delta:

- Renamed the "direct path" (`change-code --direct`) to the quick path,
  `sdlc quick`.
- §1: added the operator route. Replaced "estimate below a threshold" with a diff
  size budget, and moved both diff facts (shared surface, size) to close, where
  the diff exists. Failing them there escalates the review recipe for
  automatically admitted issues, and only warns for operator-admitted ones.
- §2: the quick path now also drops the structured estimate and `change-code`'s
  structural gate. This reverses the earlier "not the estimate (calibration needs
  every row)": quick-path rows leave est/actual calibration instead.
- §3: close now also refuses an empty `## Done when`, since the structural gate
  that guaranteed one is gone.
- §4: the architecture section is exactly DRY, PURE and PURPOSE, selected from
  the registry. Dropped the SECURE/ORDER tripwires and their escalation.
- §5, its Done-when bullets and its two Plan items moved to #233.
- Done when and Plan rewritten to match. The calibration criterion now judges
  gate rounds, actual hours and escaped defects instead of the est/actual ratio.
