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

Add a **quick flow** and give its one gate a review recipe matched to small
diffs. It is one of three flows: `full` (today's), `quick` (this issue), and
`config` for declarative changes (#233).

A flow is not a verb. It is a frontmatter field, `flow:`, set by an existing
verb. Every later gate reads it and acts on it.

### 1. Declared at entry, enforced at close

**Declaration.** `sdlc change-code --flow quick` sets `flow: quick`. The flows
diverge at `change-code`: the full flow's plan-quality and estimate gates live
there, and it already owns branching, so no new verb is needed. `claim` is too
early — claiming happens before the design exists (#113), when size is not yet
knowable. `change-code` always writes the field, with `full` as the default, so
from implementation onward every issue states its flow. An issue with no `flow:`
(every issue filed before this one) reads as `full`.

The agent may declare quick on its own judgment, or the operator can instruct it
("use sdlc quick path"). Both take the same verb and face the same checks.
`change-code --flow quick` refuses a Plan with `Mx` milestones, because the
quick flow has a single boundary.

**The hard shell.** "Is this small?" is exactly the call that gets made wrong
(#263 looked like one keybinding), so the declaration is not trusted. `sdlc
close` measures the real diff. A quick-flow issue must stay inside all three
limits:

- **At most 2 code files changed.**
- **At most 100 changed lines.**
- **No declared shared surface touched.** A repo declares its shared surfaces
  (for parley: the keybinding registry, `config.lua`'s option schema,
  `construct/vocabulary/*.cue`, any cross-module seam). This is the criterion
  that would have refused #263 — the *feature* was one chord, but it touched the
  registry, and registry-shaped work grows guard generalizations and
  cross-module extractions whatever its headline size.

Code files exclude docs and the process trees (`workshop/`, `atlas/`, `*.md`).
The limits are single-sourced in the binary: the check and the help text read
the same constants.

Crossing the shell means the work is not quick, whoever declared it. `sdlc
close` refuses, naming the limit crossed, and the refusal is the next-action
spec: `sdlc change-code --flow full` re-routes the issue through the full flow's
gates (plan-quality, estimate), and the close then runs the full review. The
shell is checked where the diff exists, it cannot be crossed silently, and the
operator route does not bypass it.

The shell should rarely be discovered at close. `sdlc state` reports a
quick-flow issue whose diff has already crossed it, so the re-route usually
happens mid-work, while a plan and an estimate can still cover most of it.

### 2. What the quick flow drops

- The durable plan and the plan-quality judge.
- The structured estimate: `estimate_hours` and the `## Estimate` block.
  `actual_hours` is still measured at close, because measuring it costs nothing.
  With no estimate to pair it with, though, quick-flow rows leave est/actual
  velocity calibration rather than entering it half-filled.
- `change-code`'s structural gate.
- Milestones: `milestone-close` refuses a quick-flow issue.

What remains is **one review gate: `sdlc close`**, with its evidence
(`--verified`, the measured actual), the hard shell (§1) and the quick review
recipe (§4). `merge` and `push` keep their deterministic publish gate (#160),
which only checks that nothing changed after that review.

### 3. What must survive — the part that is not optional

**The acceptance contract.** On the quick flow the close review is the *only*
review, and `sdlc close` judges the diff against `## Done when`. In #263 the one
blocking plan-gate finding was that a mid-stream reframe changed what got built
without restating Done-when — three of five clauses became unsatisfiable *by
design*. On the full flow the plan gate caught that. On the quick flow nothing
would, and the single remaining review would judge against stale criteria.

So `sdlc close` runs two **deterministic** (no-LLM) checks on a quick-flow issue
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

`sdlc close` selects the recipe from `flow:`, so an issue re-routed to `full`
gets the full review with no special case.

### 5. Split out: the declarative change

The declarative tier — a config change admitted on mechanical facts, with no
close review at all — is now its own flow, `flow: config`, tracked in #233. This
issue does not depend on it. #233 builds on §1's shared-surface declaration and
extends `#Flow`.

## Done when

- `construct/vocabulary/issue.cue` models `flow?: #Flow` with
  `#Flow: "full" | "quick"`, so a mistyped value fails validation. An absent
  field reads as `full`, and `sdlc issue --help` documents the field.
- `sdlc change-code --flow quick|full` writes `flow:` (default `full`), refuses
  `quick` for a Plan with `Mx` milestones, and works as a re-route on an issue
  already mid-implementation.
- On the quick flow, no durable plan, plan-quality judge, structured estimate or
  `change-code` structural gate runs, and `milestone-close` refuses. `sdlc close`
  is the only review gate.
- A repo can declare its shared surfaces.
- `sdlc close` refuses a quick-flow issue whose diff crosses the hard shell (more
  than 2 code files, more than 100 changed lines, or a declared shared surface),
  naming the limit crossed and pointing at `sdlc change-code --flow full`. Tests
  pin each limit: exactly at it passes, one past it refuses.
- `sdlc state` reports a quick-flow issue whose diff has crossed the shell.
- `sdlc close` refuses a quick-flow issue with an empty `## Done when`, and runs
  the Done-when-freshness check deterministically.
- `small-diff-review.md` exists as its own recipe, and `sdlc close` selects the
  recipe from `flow:`.
- The recipe's architecture section is ARCH-DRY, ARCH-PURE and ARCH-PURPOSE,
  selected by marker from `judge/architecture.md`. A test asserts the rendered
  recipe carries exactly those three markers, so a registry edit cannot silently
  widen or drift it.
- The recipe requires family enumeration on the first finding, and the gate
  ledger shows repeat families dropping on quick-flow issues.
- Calibration: quick-flow rows are tagged in the ledger and excluded from
  est/actual calibration, since they carry no estimate. The trade-off is judged
  on what they do carry — gate rounds per issue, measured actual hours, and
  escaped defects (follow-up fixes citing a quick-flow issue) — against
  comparable full-flow rows. If those do not improve, the trade-off was wrong
  and this gets reverted on evidence.

## Plan

- [ ] Confirm the shell's combinator: this Spec reads the operator's "> 2 files
      and > 100 lines" as *crossing either limit* leaves the shell. Also decide
      whether test files count toward the file and line limits.
- [ ] Add `#Flow` and `flow?:` to `construct/vocabulary/issue.cue`; document the
      field in `sdlc issue --help`.
- [ ] `change-code --flow`: write the field; for `quick`, skip the plan-quality,
      estimate and structural gates and refuse `Mx` milestones; support the
      mid-implementation re-route. Make `milestone-close` refuse `quick`.
- [ ] Decide the shared-surface declaration format.
- [ ] Implement the hard shell at close with its re-route refusal, and the
      `sdlc state` drift line.
- [ ] Implement the two deterministic Done-when checks at close (§3).
- [ ] Add a marker-subset selector beside `ArchitectureBlock`; write
      `small-diff-review.md` on it; select the recipe at close from `flow:`.
- [ ] Tag quick-flow rows in the calibration ledger and exclude them from
      est/actual calibration. Decide whether a row re-routed to `full` mid-work
      enters calibration, given its estimate postdates the first code commit.
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

### 2026-09-17 — the flow lives in frontmatter; a hard shell at close

Reason: the operator decided on a hard shell — past 2 code files or 100 changed
lines, the normal flow runs — and on keeping the existing verbs, with the
issue's frontmatter recording its flow for later gates to act on, modeled in
the cue schema.

Delta:

- The quick path is now the quick *flow*: `flow: quick`, set by `sdlc
  change-code --flow quick`. No `sdlc quick` verb.
- §1 rewritten as declaration at entry and enforcement at close. The hard shell
  (2 code files, 100 changed lines, no shared surface) replaces the separate
  size budget and shared-surface escalation. It binds the operator route too,
  which supersedes the previous revision's "operator-admitted issues only warn".
  Crossing it makes `close` refuse and point at the re-route, rather than
  silently switching recipes. The admitting route is no longer recorded: both
  routes now face identical checks, so nothing would read it.
- Added the `sdlc state` drift line, so crossing the shell usually surfaces
  mid-work.
- §2: `milestone-close` refuses quick-flow issues; `merge`/`push` keep their
  deterministic publish gate.
- §4: the recipe is selected from `flow:`.
- Done when and Plan updated to match: the cue field, `change-code --flow`,
  tests at each shell limit, the drift line, and the re-route.
