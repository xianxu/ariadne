---
id: 000231
status: working
deps: []
github_issue:
created: 2026-09-16
updated: 2026-09-17
estimate_hours: 6.33
started: 2026-09-17T18:59:16-07:00
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

A flow is not a verb, and the agent does not choose it. The agent runs the same
verbs in the same order on every issue — claim, change-code, implement, close —
and the gates infer the flow and behave accordingly.

### 1. Inferred by the gates, pinned by the operator

**The record.** The issue's frontmatter carries the flow and who decided it, on
one line, because the issue frontmatter helpers are line-based
(`cmd/sdlc/internal/issue/frontmatter.go`):

```yaml
flow: {kind: quick, provenance: inferred, spec: "1a2b3c4d", done: "5e6f7a8b"}
```

`kind` is `full` or `quick` (#233 adds `config`). `provenance` is `inferred` when
a gate decided, `operator` when the operator pinned it. A quick record also
carries the contract hashes §3 uses (`spec:`, `done:`), always quoted: an
all-digit hex hash would otherwise read as a YAML number. An issue with no `flow:`
(every issue filed before this one) reads as `full`.

**Inference at `change-code`.** Two facts about the issue, both artifacts the
agent already produces under the constitution:

- `Mx` milestone rows in `## Plan` → `full`.
- A durable plan for the issue in `workshop/plans/` → `full`. The constitution
  already requires one for non-trivial work.

Neither → `quick`. Having no milestones is necessary for quick, not sufficient;
the durable plan is the second signal. `full` runs the plan-quality and estimate
gates as today; `quick` skips both.

**Operator pin.** The operator can set the flow ("use sdlc quick path", or
`full` for a small change that should get the whole process). The agent passes
`sdlc change-code --flow quick|full`, which writes `provenance: operator`. The
flag attests that the operator asked, in the spirit of the `--no-<gate>` flags.
Gates never re-infer an operator pin, except through the hard shell below, so a
misused pin cannot escape it. `change-code` refuses `--flow quick` on a Plan
with `Mx` rows: a pin it cannot honour should say so while the operator is
there.

**The hard shell.** "Is this small?" is exactly the call that gets made wrong
(#263 looked like one keybinding), so the inference at entry is not trusted
either. A quick issue must stay inside all four limits, whatever its provenance:

- At most 2 code files changed.
- At most 100 changed lines, counted as insertions in code files — the number
  the churn report and calibration ledger already use.
- No declared shared surface touched. A repo declares its shared surfaces in
  `.sdlc/shared-surfaces`, one path pattern per line (for
  parley: the keybinding registry, `config.lua`'s option schema,
  `construct/vocabulary/*.cue`, any cross-module seam). This is the criterion
  that would have caught #263 — the *feature* was one chord, but it touched the
  registry, and registry-shaped work grows guard generalizations and
  cross-module extractions whatever its headline size.
- No `Mx` milestones.

Crossing any one limit crosses the shell. Code files exclude docs, the process
trees (`workshop/`, `atlas/`, `*.md`) and tests, so writing a test never pushes a
change out of the quick flow. The limits are single-sourced in the binary: the
check, the help text and the constitution's durable-plan rule all read the same
constants.

**Gates only upgrade.** A gate that finds the shell crossed rewrites the issue to
`{kind: full, provenance: inferred}`, logs the measured reason, and carries on as
the full flow. No gate ever moves an issue from `full` to `quick`. `sdlc close`
measures the diff, so it is where a crossing is usually found; `milestone-close`
finds one when an `Mx` row appears mid-work. Crossing the shell neither refuses
nor warns, so from the agent's side the flow looks the same.

The upgrade is recorded with the close's other writes at finalize. A REWORK
writes nothing (#139), and the next close re-derives the upgrade from the same
window, so "no gate downgrades" holds for the recorded flow.

An upgraded issue gets the full review at close, but not the plan or estimate
it skipped. A plan written after the code has no value, and an estimate made
afterwards is not a prediction, so the issue enters no est/actual calibration.

### 2. What the quick flow drops

- The durable plan and the plan-quality judge.
- The structured estimate: `estimate_hours` and the `## Estimate` block.
  `actual_hours` is still measured at close, because measuring it costs nothing.
  With no estimate to pair it with, though, quick-flow rows leave est/actual
  velocity calibration rather than entering it half-filled.
- `change-code`'s structural gate.

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
- If `## Spec` or `## Revisions` changed after `change-code`, `## Done when` must
  have changed too, or the agent acknowledges it did not need to
  (`--no-done-when-fresh`). Revisions counts because the constitution records a
  mid-stream reframe by appending there. Cheap, and it protects the one oracle
  left. The anchor is `change-code`, not `claim`: a claim can come before the Spec
  exists (#113), while `change-code` is where the contract is fixed. change-code
  records short hashes of both sides in the quick flow record on every run, so
  the check needs no git history and the anchor moves when the contract is
  re-fixed.

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
  ARCH-\* lens is off, with no tripwires. Which principles apply is a field on
  each entry of the single-source registry
  `cmd/sdlc/internal/judge/architecture.md` — `quick-flow: yes|no`, beside the entry's
  `at-plan`/`at-review` lenses — so changing the set means flipping a field there,
  and the recipe never restates the list.
- **Cap the tail.** #263's four rounds were mostly *repeat families* caused by
  fixing the site a finding named. The recipe should require, on the **first**
  finding in a family, the enumeration of that family before any fix. A small
  diff has a small enumeration — this is cheap exactly where it is most often
  skipped.

`sdlc close` selects the recipe from `flow.kind`, so an upgraded issue gets the
full review with no special case.

### 5. Split out: the declarative change

The declarative tier — a config change admitted on mechanical facts, with no
close review at all — is now its own flow, `kind: config`, tracked in #233. This
issue does not depend on it. #233 builds on §1's shared-surface declaration and
extends `#Flow`.

## Done when

- `construct/vocabulary/issue.cue` models `flow?: #Flow` with
  `#Flow: {kind: "full" | "quick", provenance: "inferred" | "operator", spec?:
  string, done?: string}` (the hashes 8 lowercase hex), so a
  mistyped value fails validation. The one-line form round-trips through the
  line-based frontmatter helpers, an absent field reads as `full`, and `sdlc
  issue --help` documents the field.
- `sdlc change-code` infers the kind — `full` if the Plan has `Mx` rows or a
  durable plan exists, otherwise `quick` — and writes `flow:` with `provenance:
  inferred`. `--flow quick|full` pins it with `provenance: operator`, and
  `--flow quick` is refused on a Plan with `Mx` rows.
- On the quick flow, no durable plan, plan-quality judge, structured estimate or
  `change-code` structural gate runs. `sdlc close` is the only review gate.
- A repo can declare its shared surfaces.
- A quick issue that crosses the hard shell — more than 2 code files, more than
  100 changed lines, a declared shared surface, or an `Mx` row — is upgraded to
  `{kind: full, provenance: inferred}` by the gate that finds it, whatever its
  provenance, with the measured reason in the Log, and gets the full review at
  close. No gate downgrades. Tests pin each limit: exactly at it stays quick,
  one past it upgrades.
- An end-to-end test drives the same verb sequence (claim → change-code → close)
  through a small issue and a large one, and each lands on the right flow and
  review recipe without the agent choosing either.
- `sdlc close` refuses a quick-flow issue with an empty `## Done when`, and runs
  the Done-when-freshness check deterministically.
- `small-diff-review.md` exists as its own recipe, and `sdlc close` selects the
  recipe from `flow.kind`.
- Every entry in `judge/architecture.md` declares `quick-flow: yes|no`, and a test
  refuses an entry that doesn't, so a new principle cannot be added without
  deciding. Today DRY, PURE and PURPOSE say yes. A test asserts the rendered
  recipe carries exactly the markers marked yes.
- The recipe requires every finding to enumerate its family across the window.
  Whether repeat families actually drop is judged after adoption, from the
  close-gate ledgers, not at this issue's close.
- Calibration: the ledger tags rows with `kind`, `provenance` and whether the
  issue was upgraded. Est/actual drift excludes quick and upgraded rows, since
  they carry no estimate, and a trailing quick row does not switch the drift
  check off; throughput keeps their hours. The trade-off itself is judged after
  adoption on gate rounds per issue, measured actual hours and escaped defects
  (follow-up fixes citing a quick-flow issue) against comparable full-flow rows.
  If those do not improve, this gets reverted on evidence.

## Estimate

*Produced via `brain/data/life/42shots/velocity/estimate-logic-v3.1.md` against
`baseline-v3.1.md`. Method A only.* Design hours are taken without Step 3's ×0.2
spec discount: the durable plan was written after `claim`, so its design time is
inside the measured window. `impl` is 40% of the v2 table (v3.1), and the design
buffer is +15% because a thorough plan doc exists.

Items in order:

- M1: the flow package; one milestone parser; `#Flow` in cue; the change-code
  step; the flow-conditional surfaces; docs; review.
- M2: the classifiers; shell + surfaces; the close integration; the registry
  field + recipe; the end-to-end test; docs; review.
- M3: ledger + drift; the #263 fixture run; docs; the final close review.

```estimate
model: estimate-logic-v3.1
familiarity: 1.0
item: greenfield-go-module    design=0.8  impl=0.24
item: smaller-go-module       design=0.0  impl=0.12
item: smaller-go-module       design=0.05 impl=0.12
item: smaller-go-module       design=0.15 impl=0.2
item: cross-cutting-refactor  design=0.4  impl=0.16
item: atlas-docs              design=0.1  impl=0.06
item: milestone-review        design=0.0  impl=0.14
item: smaller-go-module       design=0.05 impl=0.12
item: smaller-go-module       design=0.2  impl=0.16
item: greenfield-go-module    design=0.8  impl=0.32
item: smaller-go-module       design=0.2  impl=0.2
item: smaller-go-module       design=0.1  impl=0.2
item: atlas-docs              design=0.1  impl=0.08
item: milestone-review        design=0.0  impl=0.14
item: smaller-go-module       design=0.1  impl=0.16
item: milestone-review        design=0.0  impl=0.14
item: atlas-docs              design=0.05 impl=0.06
item: milestone-review        design=0.0  impl=0.14
design-buffer: 0.15
total: 6.33
```

## Plan

Durable plan: `workshop/plans/000231-a-quick-path-for-small-diffs-scale-the-gate-set-and-the-review-recipe-to-size-plan.md`.

- [x] Decide whether test files count toward the shell's limits. Decided
      2026-09-17: they do not.
- [x] Decide the shell/constitution threshold. Decided 2026-09-17: the shell's
      2 code files, 100 lines; the constitution cites the binary.
- [x] Decide the shared-surface declaration format: `.sdlc/shared-surfaces`, a
      repo-owned line file (see the durable plan).
- [x] M1 — the flow record and change-code: `#Flow` in cue, the `flow` package,
      one milestone parser, inference + `--flow` pin + gate skips, and a reachable
      quick path (start-plan, the constitution and skills made flow-conditional),
      docs.
- [ ] M2 — the shell, the Done-when checks and the small-diff review at close:
      one path classifier, shared surfaces, the upgrade, the recipe, the
      end-to-end test, help tokens, docs.
- [ ] M3 — calibration columns and drift exclusion, and the parley.nvim#263
      fixture run.

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
- 2026-09-17: closed M1 — round 3 after BR-11..BR-14 fixed as rules: go test ./... green except TestFleetPlanHasAuthoritativeCorrectedCoreConceptInventory (pre-existing, #210, red on main); round-2 mutations each red (stale content written -> TestRecordChangeCodeFlowKeepsConcurrentEdit; drift check off -> ...RefusesFlowChangingEdit; empty-hash check relaxed -> TestFlowRecordCorpus); flow corpus asserted by both Go (TestFlowRecordCorpus) and cue (TestValidateInstance_FlowRecordCorpus); BR-13 verb sweep command recorded in the plan (26 hits, 3 qualified); git diff --check clean; review verdict: SHIP

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

### 2026-09-17 — claimed; two shell decisions

Claimed and entered planning. The operator settled two open questions: test
files do not count toward the shell's limits, and the constitution's durable-plan
threshold moves from >3 files to the shell's >2, so the two are one number owned
by the binary.

### 2026-09-17 — durable plan written

Wrote the durable plan after three read-only code surveys: `change-code`, the
close/judge path, and the ledgers/config. Three milestones: record +
change-code, close, calibration. The surveys changed the design in four places,
recorded in the plan's Decisions: the upgrade is recorded at finalize because
`computeClose` writes nothing and a REWORK must leave the issue unwritten (#139);
changed lines are insertions, as churn counts them; the three disagreeing path
classifiers collapse into one per-path set in `churn`, whose test detection widens
to non-Go layouts because parley.nvim is Lua; and the colon-only milestone regex in
`sizing.go` retires onto the one `close.go` uses.

### 2026-09-17 — plan-quality round lost to the sandbox

The first `change-code` run's plan-quality judge never reached the API: the
`claude` subprocess got `403 Connection blocked by network allowlist` from the
session sandbox, and the gate recorded a BLOCKED round with a protocol error and
no findings. No review happened, so the untracked ledger file it created was
removed rather than letting an environment failure spend one of the three
rounds. Re-ran outside the sandbox.

### 2026-09-17 — change-code passed; branch created

Plan-quality passed on round 2, with three advisory findings (PQ-8..PQ-10),
folded in below. The estimate-quality verdict was INFO, and its main finding
stands: `sdlc actual` already measures about 5.4h before any code, because the
window opens at the first `#231:` commit (a pre-claim `issue sync` on 09-16), not
at `claim`. The 6.33h estimate will come in low, ratio about 0.7. It stays as
derived; v3.1 has no "sunk, already measured" primitive. A tooling observation
worth its own issue: pre-claim `issue sync` checkpoints (#206) move the
active-time window earlier than #113 intends. The branch was created in place.

### 2026-09-17 — M1 built; boundary review round 1

M1 is committed across five commits: the milestone parser, the flow package,
`#Flow` in cue, change-code inference, and the reachable-path surfaces. The
boundary review returned FIX-THEN-SHIP with two Important findings, both fixed
as classes. BR-4: the guard now derives every caller of `MilestonesInPlanOrder`
instead of trusting a hand list. BR-5: the flow record is written only after
change-code's gates pass, so a refused run leaves no sticky `full` behind.

Two notes for later:

- M1's docs describe close-time behaviour in the present tense: the small-diff
  review, the upgrade and the Done-when refusals. M2 builds it. That is correct
  only because M2 lands before this branch merges; the branch is not merged.
- Known limitation of the contract hash: it covers everything from `## Revisions`
  to the next `##` heading. A Log entry hand-appended at the end of a file whose
  last section is Revisions therefore moves the `spec` hash, and the freshness
  check refuses falsely. `--no-done-when-fresh` covers it. `sdlc close` inserts
  Log lines into `## Log`, never at end of file, so the gate's own writes don't
  trigger it.

### 2026-09-17 — M2 built

The close side landed in seven commits: one per-path classifier set in `churn`;
the pure shell (`Measure`, `Crossings`, shared surfaces, `Upgrade`); the
small-diff recipe, with its principles chosen by the registry's new
`quick-flow:` field; close's measurement, upgrade, Done-when checks and recipe
selection; and the end-to-end test that drives the same verbs through a small
issue and a large one. Each wire was mutation-checked red.

A measurement cut-over to know about: `isTestPath` now recognises non-Go test
layouts (Lua `*_spec.lua`, Python, JS/TS, shell `*.test.sh`, and
`test/`/`tests/`/`spec/` directories). From #231 on, calibration-ledger rows for
non-Go repos count those as `churn_test`, where earlier rows counted them as
`churn_prod`, so compare #187 cost metrics across the cut-over with that in
mind. ariadne's own `scripts/test/*.test.sh` moves the same way.

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

### 2026-09-17 — the gates infer the flow; the operator can pin it

Reason: the operator decided that the agent should not choose a flow. From its
side the flow is the same; the gates behave differently by flow, and nothing
warns. The record carries `kind` and `provenance` (`inferred | operator`). Having
no milestones is necessary for quick, not sufficient.

Delta:

- `flow:` is now `{kind, provenance}`, kept on one line because the frontmatter
  helpers are line-based.
- §1 rewritten. `change-code` infers the kind from `Mx` rows and the presence of
  a durable plan; `--flow` is now an operator pin. The hard shell gains `Mx` rows
  as a fourth limit and no longer refuses: a gate that finds it crossed upgrades
  the issue to `full` and carries on. Gates only upgrade. This replaces the
  previous revision's refusal at close, its re-route through `change-code`, and
  the `sdlc state` drift line.
- An upgraded issue gets the full review but no retroactive plan or estimate,
  and stays out of est/actual calibration.
- The shell reads as "crossing any one limit", matching the constitution's own
  "or". Single-sourcing the two thresholds is a Plan item.
- §2: dropped the separate `milestone-close` refusal; the `Mx` limit covers it.
- Done when and Plan updated to match, including an end-to-end test that the
  agent's verb sequence is identical on both flows.

### 2026-09-17 — Done-when made checkable at close; plan decisions folded in

Reason: two Done-when bullets described adoption outcomes (repeat families
dropping, the est/actual trade-off paying off) that no gate can observe when this
issue closes. Planning also settled details the Spec left open.

Delta: those two bullets now name what is checkable at close — the recipe's
family-enumeration rule; the ledger columns, the drift exclusion and a trailing
quick row not disabling drift — and state that the outcomes are judged after
adoption. §1 now defines changed lines as insertions in code files, names
`.sdlc/shared-surfaces` as the declaration, and says the upgrade is recorded at
finalize. §3's freshness check now anchors at `change-code` instead of `claim`,
and names its skip flag. `## Plan` became three milestones pointing at the
durable plan.

### 2026-09-17 — the quick principles are a registry field

Reason: the operator chose to set which principles the quick flow checks in the
ARCH registry rather than in a Go list.

Delta: §4 and the matching Done-when bullet now say each registry entry declares
`quick: yes|no`, every entry must declare it, and the recipe carries exactly the
markers marked yes.

### 2026-09-17 — plan-quality round 1

Reason: the plan-quality gate found that the quick flow was unreachable on the
documented path (start-plan, the constitution and the brainstorming skill send
every issue into a durable plan, which infers full), and that a reframe recorded
under `## Revisions` would slip past the freshness check.

Delta: §3's freshness check watches Spec + Revisions, anchored by hashes that
change-code records in the quick flow record (§1 notes the extra fields). M1 now
makes every surface that routes work into a durable plan flow-conditional, and the
constitution work moves there from M3.

### 2026-09-17 — advisory plan findings PQ-9, PQ-10

Reason: plan-quality found that an all-digit hash parses as a YAML int, which
cue then rejects (PQ-9), and that §4 and Done-when still named `quick:` and a
hash-less `#Flow` after the plan renamed and reshaped them (PQ-10).

Delta: §1's example record carries quoted hashes, and says why. §4 and Done-when
say `quick-flow:`, and the `#Flow` bullet includes the optional hashes. Swept
every name Done-when uses against the plan; the only remaining mentions of the
old names are in dated Revisions entries.

