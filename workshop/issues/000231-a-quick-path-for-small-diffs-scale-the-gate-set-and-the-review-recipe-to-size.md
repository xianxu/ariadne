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

**Inference at `change-code`.** The shell below, as far as it can be measured
before any code exists — two facts about artifacts the agent already produces:

- `Mx` milestone rows in `## Plan` → `full`.
- A design past its limit → `full`. The design is `## Spec`, `## Plan` and the
  durable plan in `workshop/plans/`, counted in lines. A plan runs longer than
  the code it describes, since it cites code and justifies it, so it gets its
  own limit rather than a multiple of the code limit.

Neither → `quick`. Having no milestones is necessary for quick, not sufficient;
the design's length is the second signal. A plan's mere existence is not: a
short one stays quick. `full` runs the plan-quality and estimate gates as
today; `quick` skips both. change-code and close measure through the same
`flow.Measure` and `Crossings`, so they cannot disagree about a limit both see.

**Operator pin.** The operator can set the flow ("use sdlc quick path", or
`full` for a small change that should get the whole process). The agent passes
`sdlc change-code --flow quick|full`, which writes `provenance: operator`. The
flag attests that the operator asked, in the spirit of the `--no-<gate>` flags.
Gates never re-infer an operator pin, except through the hard shell below, so a
misused pin cannot escape it. `change-code` refuses `--flow quick` on an issue
already outside the shell (`Mx` rows, or a design past its limit): a pin it
cannot honour should say so while the operator is there.

**The hard shell.** "Is this small?" is exactly the call that gets made wrong
(#263 looked like one keybinding), so the inference at entry is not trusted
either. A quick issue must stay inside all three limits, whatever its
provenance:

- At most 100 added lines in code files, counted as insertions — the number the
  churn report and calibration ledger already use. Deleted lines do not count,
  and neither does how many files the lines spread across.
- A design of at most 500 lines: `## Spec`, `## Plan` and the durable plan.
- No `Mx` milestones.

The shell measures size, not reach. How many files a change spreads across and
where in the tree it lands both shape its blast radius, not its bug profile, and
under 100 lines the tests (plus a smoke test where needed) and the one close
review are the guard. In ariadne, where sdlc runs from the branch under review,
a quick change to sdlc is closed by the binary it just changed; that review sees
the diff.

Crossing any limit crosses the shell. Code files exclude tests, docs and the
process trees (`workshop/`, `atlas/`) — except markdown under `cmd/`, which ships
in the binary and counts — so writing a test never pushes a change out of the
quick flow. (`churn.CodeFileRule` states it where the classifier lives.) The limits are single-sourced in the binary: the
check, the help text and the constitution's durable-plan rule all read the same
constants.

**Gates only upgrade.** A gate that finds the shell crossed rewrites the issue to
`{kind: full, provenance: inferred}`, logs the measured reason, and carries on as
the full flow. No gate ever moves an issue from `full` to `quick`. `sdlc close`
measures the diff, so it is where a crossing is usually found; `milestone-close`
finds one when an `Mx` row appears mid-work. Crossing the shell neither refuses
nor warns, so from the agent's side the flow looks the same.

The upgrade is recorded with the close's other writes at finalize. A REWORK
writes nothing to the issue (#139), but every boundary round is stamped with the
recipe it ran in the boundary ledger, and an earlier full-review round is itself a
crossing. So a quick issue that once needed the full review stays full, even if
the fix shrinks the diff back inside the shell.

**Fixes after the verdict.** Close measures the head its review saw. The fixes
a SHIP or FIX-THEN-SHIP verdict leads to ride into the close commit, which is
also the publish anchor, so close never measures them. The publish check
(`merge`/`push`) therefore re-measures a quick issue's final diff over close's
own window. Up to twice the line limit (200 added code lines) it publishes: the
fixes answer the review's own findings, and the margin keeps a re-review rare.
Past it, the check refuses and sends the issue back to `sdlc close`, which finds
the shell crossed, upgrades the issue and runs the full review. The publish
check stays LLM-free; the review runs in close.

An upgraded issue gets the full review at close, but not the plan or estimate
it skipped. A plan written after the code has no value, and an estimate made
afterwards is not a prediction, so the issue enters no est/actual calibration.

### 2. What the quick flow drops

- The plan-quality judge, and the requirement to plan. A plan is optional: a
  very small task writes none (`## Plan` may stay empty), and a short one stays
  inside the design limit.
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
issue does not depend on it. #233 extends `#Flow`.

## Done when

- `construct/vocabulary/issue.cue` models `flow?: #Flow` with
  `#Flow: {kind: "full" | "quick", provenance: "inferred" | "operator", spec?:
  string, done?: string}` (the hashes 8 lowercase hex), so a
  mistyped value fails validation. The one-line form round-trips through the
  line-based frontmatter helpers, an absent field reads as `full`, and `sdlc
  issue --help` documents the field.
- `sdlc change-code` infers the kind — `full` if the Plan has `Mx` rows or the
  design (`## Spec` + `## Plan` + the durable plan) runs past its limit,
  otherwise `quick` — and writes `flow:` with `provenance: inferred`. `--flow
  quick|full` pins it with `provenance: operator`, and `--flow quick` is refused
  on an issue already outside the shell.
- On the quick flow, no plan-quality judge, structured estimate or
  `change-code` structural gate runs, and no plan is required. `sdlc close` is
  the only review gate. The constitution, the brainstorming skill and
  start-plan say a plan is optional inside the shell, and a blank issue's
  `## Plan` seed is not a plan item, so an empty Plan passes close.
- A quick issue that crosses the hard shell — more than 100 added lines in code
  files, a design past 500 lines, or an `Mx` row — is upgraded to `{kind: full,
  provenance: inferred}` by the gate that finds it, whatever its provenance,
  with the measured reason in the Log, and gets the full review at close. No
  gate downgrades. Tests pin both size limits at change-code and at close:
  exactly at a limit stays quick (however many files carry the lines), and one
  past it upgrades.
- The publish check re-measures a quick issue's final diff, fixes after the
  verdict included, over close's own window. Past twice the line limit it
  refuses and names `sdlc close`; a full issue is not re-measured. Tests pin
  200 and 201 added code lines, test lines not counting, and the docs-only pass
  path.
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
- [x] M2 — the shell, the Done-when checks and the small-diff review at close:
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
- 2026-09-17: closed M2 — round 4 after BR-26..BR-30 fixed as rules: go test ./... green except TestFleetPlanHasAuthoritativeCorrectedCoreConceptInventory (pre-existing, #210, red on main); TestSurfaceForms pins every accepted surface form, refused forms tested; TestGateFlagListsInHelpAreComplete pins help gate-flag lists to GateCatalog (completed merge/push/milestone-close); build-closure guard now fails on go list error and covers go.work*/vendor; TestCloseNonASCIIPathsClassifyAsThemselves red with DiffNames unquoted; git diff --check clean; review verdict: FIX-THEN-SHIP
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

### 2026-09-17 — M3: ledger columns, and the #263 fixture run

The calibration ledger gained `flow_kind`, `flow_provenance` and `flow_upgraded`
(indices 20–22). Drift skips quick and upgraded rows, and a trailing quick row no
longer switches the drift check off.

The fixture ran with a real judge through `smalldiffreplay_test.go`, a manual-tagged
harness that builds close's own dispatch with the small-diff category. It replayed
parley.nvim#263 in two windows, from clones checked out at each reviewed head.

- **`bbe05eef..ac60a055`** (what #263's first close review read) → REWORK:
  - "the one place the `<C-g>` spelling leads" is false; `outline` and
    `chat_drill_in` lead too. This is **BR-2** exactly.
  - The insert-mode press of the chord leaves insert mode (Critical). This is
    **BR-1's class**: a two-mode contract exercised in one mode.
  - The repeated command preamble, #263's `duplicated-command-preamble` family.
  - A streaming race the original first round did not raise (Critical).
  - Test oracles that observe the mechanism rather than the behaviour.
- **`bbe05eef..62c7f292`** (after #263's round-2 fix) → REWORK:
  - `get_paste_line`'s doc comment now sits above `exchange_index_at`. This is
    **BR-9** exactly.
  - The insert-mode Critical again, a restated-scan duplication, and the missing
    README.

Architecture lenses: markers cited were ARCH-DRY 5, PURE 2 and PURPOSE 1 in the
first run, and DRY 4, PURE 2 and PURPOSE 4 in the second. No round went to the
five principles marked `quick-flow: no`. The recipe's aim holds: it surfaces the
local-correctness, doc-truth and duplication classes #263's review found, in
its first round. Two limits: a single run per window is an LLM sample, not a
benchmark; and #263 would not actually have been quick (it touched a
keybinding registry), so this tests the recipe's aim rather than the admission.

### 2026-09-18 — ariadne keeps its full build-closure declaration

Reading pair#283 (7 code files, 30 added lines; inferred full because it has a
durable plan) surfaced #235: plan lookup is by exact name. That raised whether
#235 could be a quick-flow example in ariadne. It cannot while sdlc's whole build
closure is a declared shared surface. The operator decided to keep the
declaration rather than narrow it to the shell's defining files, as not worth
the extra rule and risk. So sdlc work in ariadne always takes the full flow, and
the quick flow's trial runs in other repos.

### 2026-09-18 — first trial reading: pair#283, and what to watch

pair#283 was the first real task under the branch: 7 code files, 30 added lines,
a 427-line durable plan, and inferred full because of that plan. By the shell it
is not quick: the file limit catches a change that is tiny but spread along a
pipeline (vt → endpoint → frame → renderer) and renumbers a state enum, which is
ARCH-ORDER territory, and the quick recipe does not check that. The review it got
was right. The plan is where the cost went unearned (plan-quality passed in round
1 with two Minors).

Observation for the trial: **quick-then-upgrade already yields "no plan, full
review"**. With no durable plan, #283 would have been inferred quick at
change-code and upgraded at close by its file count. The constitution's
plan threshold says to write a plan outside the shell, so spread-but-tiny
changes currently pay for one. Two questions to settle on trial data, not
yet: should spread-but-tiny work skip the durable plan? And should a
substantial in-issue `## Plan` count toward full, via the plan-item count
`issue.ComputeSizingFromContent` already computes?

### 2026-09-18 — trial reading, and the shell narrowed to lines

The trial evidence, from every fleet repo with work since the branch went live
(9/17, about 20:40):

- **No issue has taken the quick flow yet** in pair, parley.nvim or tools. None
  stayed quick, none was upgraded at close, none was pinned.
- **pair#283** closed full because of its durable plan. Its first change-code
  inferred **quick**, because the plan's file name missed #235's exact-name
  lookup. The agent noticed, renamed the plan and re-ran, and got full. So #235
  is an admission bug, not a nit: a plan-bearing issue can silently skip
  plan-quality and get the small-diff review.
- **tools#70** is full through its Mx rows and its plan. It is still open.
- **pair#262 and pair#280 are not trial data.** At 20:30 on 9/17 ariadne's tree
  failed to compile mid-edit. The pair session built sdlc from ariadne's
  committed HEAD (main's sdlc) in its scratchpad and used that binary for the
  rest of the session. Every repo's `sdlc` builds from ariadne's working tree, so
  trial fixes are now made in a worktree and land on the branch only when
  green.
- **pair#283 is a case the quick flow could have sped up:** 7 code files, 30
  added lines, and a 427-line plan that plan-quality passed in round 1. Its
  review was right, but the plan was not earned.
- The operator's first pick for a quick trial was pair#282, which turned out
  more involved. tools#76, a deletion, is now in progress.

The operator's decisions on this reading, recorded under Revisions:

- Spread-but-tiny work is simple work. The shell drops the file-count limit and
  keeps 100 added lines in code files, plus the Mx rule.
- Shared surfaces are gone, ariadne's build-closure declaration included. That
  reverses the entry above ("ariadne keeps its full build-closure
  declaration"): a surface measures blast radius, not bug profile, and under the
  line limit the tests and the one close review are the guard.
- A sizeable in-issue `## Plan` stays out of inference.

Keep watching the trial for quick issues, and for an issue in the quick flow
that a file-count or location rule would have caught.

### 2026-09-18 — the design limit, and why the other repos saw none of it

The operator's follow-up on the same reading: a plan runs longer than the code it
describes, because it cites code and justifies it. So the design side of the
shell is now a length — `## Spec`, `## Plan` and the durable plan together, at
most 500 lines — rather than "a durable plan exists". That reverses the entry
above on the in-issue `## Plan`: it now counts. For very small tasks the
constitution now says a plan is optional, down to none at all.

Measured before choosing the number, over the calibration ledger's 190 issues
with churn and an issue file: 29 added at most 100 code lines. Their durable
plans ran 414 lines at the 90th percentile, and only one passed 500
(ariadne#205, 522 lines for 22 lines of code, 3 review rounds). So the limit
rarely binds; it is a backstop for a design-heavy small diff, and pair#283's
427-line plan would now have stayed quick. Only about 15% of past issues were
under 100 code lines at all, which sets the trial's expected volume.

Why no trial issue went quick: pair, tools and parley.nvim read a composed
`AGENTS.md` that only `make weave` refreshes, and theirs still said
"Non-trivial task (>3 files or >100 lines) → durable plan". Their skills link
to ariadne's working tree, so the brainstorming change reached them; the
constitution did not. tools#76, a deletion, was writing a 20 KB plan under that
rule. `make weave` in each repo is part of landing this revision.

A blank issue's `## Plan` seed (`- [ ]`, no text) was already not a plan item,
so an empty Plan never tripped close's unchecked-plan gate. It is now pinned by
`TestScaffoldPlanSeedIsNotAnItem`, since the new guidance relies on it.

### 2026-09-18 — trial: tools#76, a "deletion" that was not small

tools#76 (delete #66's orphaned source-provenance chain) closed full: inferred
from its durable plan under the old rule, estimate 1.92h, actual 1.32h, 3 gate
rounds. Before it started it looked like the best quick candidate — a deletion,
so almost no added lines. It was not: keeping the one live piece
(`projectLanguageText`) moved it into a new `language_text.go`, and the window
added 129 code lines (461 deleted). Under the new shell it would have entered
quick (a 264-line plan is inside the design limit) and been upgraded at close
on its added lines. That is the shell doing its job: the entry-time judgment
("it's a deletion") was wrong, and only the close-time measurement knew. Still
no issue has stayed quick.

### 2026-09-18 — trial: pair#279, the first issue to stay quick

pair#279 (alt+d dead in the switcher) closed quick/inferred, not upgraded:
0.85h actual, one close round, FIX-THEN-SHIP with six Minors. What the flow did:

- The same verbs as ever — claim, start-plan, change-code (a dry run first),
  close, pr, merge — with no bypass flag and no `--flow` pin. change-code printed
  the quick line and ran none of its gates. No durable plan; the in-issue Plan
  was five plain checkboxes.
- The issue had been misfiled: it asked for detaching the selected row, while
  #170's contract (and #282's key help) is detach-all-and-leave. The agent traced
  it, asked the operator, and restated Spec and Done-when with a Revisions entry
  before change-code, so the contract hashes anchor the corrected contract.
- The root cause was elsewhere than the issue guessed: kitty keyboard flags are
  per-screen, and the presenter pushed once (`f32bb4cf`), so on the alternate
  screen alt+d arrived as legacy `ESC d`.
- The small-diff review cited only ARCH-DRY, PURE and PURPOSE, and its findings
  enumerated their families. Four were fixed; one (re-using vt's kitty model in
  a test) was declined with a reason (an independent oracle); one family's
  pre-existing remainder, beyond this window, was filed as pair#289.
- Size: 91 added code lines at the reviewed head, across 4 code files. The old
  2-file limit would have upgraded it; the line-only shell kept it quick.

**Gap found: FIX-THEN-SHIP fixes are not measured.** The shell measures the
review window's head; the fixes land after the verdict, in the close commit.
Here they took the diff from 91 to 96 added lines — still inside, but unmeasured.
A larger fix could carry a quick issue past the shell with only the small-diff
review behind it.

### 2026-09-18 — trial: pair#289, the second quick issue

pair#289 (retire hostty's test-only control surface), filed from pair#279's
close review, closed quick/inferred: 0.48h, one close round, SHIP with four
Minors. change-code inferred quick; close measured "45 added lines in code
files, a design of 44 lines" and ran the small-diff review. BR-1 enumerated
"all instances in window". All four Minors were fixed in the round, one of
which surfaced another orphan (`joinArgs`, fake-only) and moved it.

It sharpens the gap above: the fixes after a **SHIP** verdict are unmeasured
too, not only FIX-THEN-SHIP's. They took the diff from 45 to 50 added lines and
from 4 to 6 code files. Any post-verdict change bundled into the close commit
skips the shell.

Friction outside the flow: `sdlc merge` refuses on dirty tracked files, and
pair (and tools) carry uncommitted base-layer drift — `Makefile` changed from
its committed symlink into a regular file, plus `bootstrap.sh`,
`merge-check.yml` and `scripts/` from ariadne#213. The session backed them up,
stashed, merged, popped, and verified them by checksum. That was safe, but every
merge in those repos pays for it until the drift is committed or discarded.

### 2026-09-18 — the publish check re-measures a quick diff

Built what the pair#279 and pair#289 readings asked for. The operator's first
idea was for sdlc to switch the review prompt, and it does that already for
every round; the gap is that fixes after the verdict get no round. Their
threshold, twice the line limit, keeps the extra review rare. So the publish
check re-measures and, past it, routes through close rather than running the
review itself: close owns the upgrade record, the Log reason, the review-ledger
round and the calibration row, and the publish check stays LLM-free. Neither
trial issue came near it (96 and 50 lines against 200).

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

### 2026-09-17 — M2 review: code files, shared surfaces, and ariadne's own flow

Reason: M2's boundary review found that §1's code-file sentence contradicted the
classifier, and that ariadne's own shell could be loosened from a branch.

Delta: §1 now says markdown under `cmd/` counts as code. A shared surface is
matched on both sides of a rename. In ariadne, sdlc's whole build closure is a
declared shared surface (`cmd/sdlc/`, `pkg/`, `go.mod`, `go.sum`), because sdlc
runs from the branch under review. Consequence to know: code changes to sdlc in
ariadne always take the full flow. The quick flow in ariadne covers everything
else, and in every other repo the shell is as specified.

### 2026-09-18 — the shell is 100 added lines and Mx rows; shared surfaces removed

Reason: the operator's decisions after the first trial reading (Log,
2026-09-18). Spread-but-tiny changes such as pair#283 are simple work. A shared
surface measures blast radius, not bug profile, and for a change under 100 lines
the operator relies on tests (plus a smoke test where needed) and the one close
review.

Delta:

- §1's shell is two limits, 100 added lines in code files and no `Mx` rows. The
  2-code-file limit is gone, and so is the shared-surface criterion with its
  `.sdlc/shared-surfaces` declaration, ariadne's build-closure declaration
  included. That supersedes the previous revision's ariadne consequence: sdlc
  code changes in ariadne can now take the quick flow, and are then closed by
  the binary they changed. §1 says why the shell measures size, not reach.
- §5 no longer says #233 builds on the declaration.
- Done when: dropped "A repo can declare its shared surfaces". The upgrade
  bullet lists the two limits, and its tests pin the line limit at 100 and 101,
  across many files.
- An in-issue `## Plan` stays out of inference; unchanged.

### 2026-09-18 — the design limit replaces "a durable plan exists"

Reason: the operator's follow-up to the trial reading (Log, 2026-09-18). A plan
runs longer than the code it describes, so a small change with a short plan
should stay quick, and a very small task needs no plan at all. `## Spec` counts
too, because brainstorming lands the design there; counting only the Plan would
move design text out of the measure.

Delta:

- §1's inference: `Mx` rows or a design past its limit → full, otherwise quick.
  A durable plan's existence is no longer a signal. change-code measures the
  shell through the same `flow.Measure` and `Crossings` close uses, and a quick
  pin is refused on an issue already outside the shell.
- §1's shell gains a third limit: a design of at most 500 lines (`## Spec` +
  `## Plan` + the durable plan). Close checks it on the design as it stands, so
  a plan that grew during the work upgrades the issue.
- §2: the quick flow drops the requirement to plan, not the plan itself.
- Done when: the inference, quick-flow and upgrade bullets name the design
  limit; the tests pin it at 500 and 501 at change-code and at close.
- This supersedes the previous revision's last line: an in-issue `## Plan` now
  counts, as part of the design.

### 2026-09-18 — fixes after the verdict are re-measured at publish

Reason: pair#279 and pair#289 showed that fixes made after a SHIP or
FIX-THEN-SHIP verdict ride into the close commit unmeasured. The operator set
the tolerance at twice the line limit and agreed to route a crossing through
close.

Delta: §1 gains "Fixes after the verdict": the publish check re-measures a
quick issue's final diff over close's window and, past twice the line limit,
refuses and sends it back to close. Done when gains the matching bullet.
