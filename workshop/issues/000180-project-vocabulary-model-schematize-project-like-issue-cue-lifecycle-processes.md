---
id: 000180
status: working
deps: []
github_issue:
created: 2026-07-15
updated: 2026-07-15
estimate_hours: 8.1
started: 2026-07-15T15:12:34-07:00
---

# project vocabulary model: schematize project like issue (cue + lifecycle + processes)

## Problem

`project` is a datatype in prose only (`construct/datatype/project.md`:
frontmatter table, `status: active|paused|done|dropped`, done_when, MVP
scope, single-operator discipline) — it is NOT a vocabulary model.
`construct/vocabulary/` holds issue.cue / pensive.cue / verdict.cue; issue's
status enum + lifecycle are formally modeled and ENFORCED (sdlc reads them
via `pkg/vocab`: set-status transition guards, the compiled done-guard,
discovery dirs feeding resolve). Project has none of that:

- `sdlc close`'s project gate parses project files by convention (task tick
  + detail-block upsert against an unchecked shape);
- no gate validates a project instance's conformance (issues get the
  instance-conformance gate at merge; projects get nothing);
- the status enum exists only as a markdown table nothing reads — a
  hand-maintained restatement, not an enforced source (the exact
  ARCH-PURPOSE gap the issue vocabulary work closed for issues).

This matters more now: #171 lifts project management into the sdlc spine
(project files move into coding repos; the close gate resolves them across
peers; project verbs join sdlc). A lift onto an unschematized noun bakes the
prose-only shape into code.

## Spec

*(Brainstorm converged 2026-07-15, operator + agent. The organizing insight:
the project lifecycle is the issue lifecycle one level up — ideation→
brainstorm, PRD→Spec, project estimation→## Estimate, breakdown→Plan,
execution→implementation, retro→Log, close→measured close. The artifact
answer follows fractally: one file growing sections through gated stages,
binary-owned gates, calibrated estimation, derived views.)*

**Taxonomy** (the three nouns): *issue* = a concrete thing in a single repo;
*pensive* = unstructured musing/realization; *project* = a structured,
TIME-BOUND push for a major change in a product, across dependencies/repos.
A project is not merely a container of issues — it carries a deadline.

**Lifecycle enum (start with this set):**
`ideation → defined → committed → executing → done | dropped` (+ `paused`).
The transitions are the gates:
- **define** (ideation→defined): a PRD exists in the project file. PRD
  authoring is itself tracked as a normal issue; ideation lives in parley
  (linked via `sources:`). Candidate: a fresh-eyes PRD review (plan-quality
  judge's sibling).
- **commit** (defined→committed): the Phase-A estimate exists + the reality
  check passes (fits the roadmap month's capacity, given calibrated velocity
  and competing priorities). Commit SETS THE BASELINE: `deadline:` (the
  time-bound attribute) + planned finish + parallelism intent.
- **breakdown** (committed→executing): PRD converted to issues across repos
  (`deps:` qualified refs). Candidate judge: does the issue set COVER the
  PRD, and does it include the infra/maintenance work PRDs ignore?
- **close** (executing→done|dropped): mandatory PROJECT RETRO entry + the
  fog-factor calibration row (below).

**Time-bound attribute:** frontmatter gains `deadline:` (and the committed
baseline: planned finish, set at commit). Distinguishes project from a mere
issue container; feeds on-track computation.

**Two-phase estimation, one calibration process, different vocabularies:**
- *Phase A (at commit, from the PRD):* a project-level estimate with its OWN
  vocabulary and axes — PRD-level primitives + an explicit uncertainty
  multiplier — because the issue primitive table needs specificity that
  doesn't exist yet. Recorded in the project file's `## Estimate`.
- *Phase B (at breakdown):* the existing per-issue machinery
  (estimate-logic-v3.1), unchanged.
- *Calibration bridge (the fog factor):* at project close, roll up the
  `mvp_scope` issues' measured actuals from the calibration ledger and
  record Phase-A-estimate vs Σ-actuals — a project-level ledger row. Over
  projects this calibrates the PRD-stage multiplier the same way closes
  calibrated issue estimation.

**Kanban: baseline STORED, progression DERIVED, re-forecasts LOGGED** (three
aspects, kept separate — operator decision):
- *Baseline* (stored at commit/breakdown in `## Breakdown`): the
  non-derivable intent — deadline, planned finish, thread assignments,
  sequencing decisions and why. Never overwritten.
- *Progression* (derived, never stored): `sdlc project status` computes the
  board from live issue frontmatter across repos via the resolve machinery —
  dependency frontier (what's unblocked), Σ remaining estimates vs deadline,
  parallel threads as independent dep-subgraphs. A hand-maintained board is
  the drifting-copy lesson waiting to bite.
- *Re-forecasts* (logged): retro entries append "where we are + new
  forecast" to `## Log`; the baseline stays intact so slippage is visible.
- Parallelism ceiling is OPERATOR ATTENTION (~2 concurrent sessions per the
  #117 measurements), not agent count — the constant the estimate unit-note
  said lives "one level up"; this is that level.

**Retro mechanism, not mandate:** `sdlc project retro` verb computes the
on-track summary and prompts the Log entry; the issue-close project gate
(which already touches the project file) nudges when the last retro entry on
an executing project is stale (>1 week); project close REQUIRES a retro
entry (gate, like issue close requires --verified).

**Schema/machinery candidates (as filed, still current):**
- `construct/vocabulary/project.cue` — fields incl. `deadline:`, the
  lifecycle above, transition rules, and a `discovery:` block: home =
  **`workshop/projects/`** (operator, 2026-07-15 — plural, matching
  issues/plans/targets; project files live under workshop/ like every SDLC
  artifact, one dir per repo; cross-repo resolution globs it across peers
  per #171). Settled at design (operator, 2026-07-15): done projects
  ARCHIVE to `workshop/history/projects/` per #181's subfolder layout —
  `workshop/projects/` stays the live portfolio; the datatype prose's
  "file becomes a record" claim is satisfied by the history copy.
- `pkg/vocab` accessor (`vocab.Project()`) mirroring `Issue()`; no consumer
  hardcodes the enum.
- Project verbs on the spine (new/list/show/set-status/status/retro; tick
  semantics at issue close), transition guards, instance conformance (the
  gate class issues get at merge).
- Prose doc derives, not duplicates: `construct/datatype/project.md` cites
  the cue as schema authority (procedure refers, registry defines), drift
  test binding the two.

**Dogfood:** REVERSED at plan approval (operator, 2026-07-16; supersedes the
2026-07-15 deferral): `workshop/projects/project-management-primitive.md`
exists from ideation, hand-authored to the emerging model shape, and is the
live test subject as machinery comes online — M2's conformance gate validates
it, M3's verbs transition it, M4's board/retro/close drive it. #180 itself
stays a single multi-boundary issue (that part of the deferral stands); the
project file tracks the wider lift (#180 + #171, with #182 explicitly out).

Out of scope (own tickets later): `product` and `roadmap` deserve the same
lift; project first — it is the one the sdlc spine touches.

Related: #171 (residency/navigation/close-gate half — consumes this model;
soft ordering: model first or together).

## Estimate

Item→milestone mapping: M1 = typed-data-prototype (project.cue) + two
smaller-go-module (vocab helper extraction; Project() binding) +
cross-cutting-refactor (kind-keyed ArchiveSubdir, 9 non-test + 2 test call
sites). M2 =
greenfield-go-module (Doc/Task parser) + two smaller-go-module (tick reimpl;
validate-gate noun table). M3 = two smaller-go-module (verb skeleton +
helptext; new/list/show/validate) + greenfield-go-module (guard registry +
set-status). M4 = greenfield-go-module ×2 (computeBoard/status; project
close) + smaller-go-module (retro + nudge) + pensive (Phase-A method doc).
M5 = atlas-docs (prose demotion + drift test + atlas). milestone-review
impl=1.0 aggregates the five boundary reviews at 0.2 each.

```estimate
model: estimate-logic-v3.1
familiarity: 1.0
item: typed-data-prototype    design=0.2 impl=0.16
item: smaller-go-module       design=0.1 impl=0.16
item: smaller-go-module       design=0.1 impl=0.2
item: cross-cutting-refactor  design=0.1 impl=0.16
item: greenfield-go-module    design=0.3 impl=0.28
item: smaller-go-module       design=0.1 impl=0.16
item: smaller-go-module       design=0.1 impl=0.16
item: smaller-go-module       design=0.2 impl=0.2
item: smaller-go-module       design=0.2 impl=0.24
item: greenfield-go-module    design=0.3 impl=0.28
item: greenfield-go-module    design=0.4 impl=0.4
item: smaller-go-module       design=0.2 impl=0.2
item: greenfield-go-module    design=0.4 impl=0.4
item: pensive                 design=0.4 impl=0.08
item: atlas-docs              design=0.2 impl=0.2
item: milestone-review        design=0.0 impl=1.0
design-buffer: 0.15
total: 8.1
```

*Produced via `brain/data/life/42shots/velocity/estimate-logic-v3.1.md`
against `baseline-v3.1.md`. Method A only. Design column includes the
already-spent brainstorm/plan sessions (actuals measure from claim); impl
values are v3.1's 40%-scaled v2 table; +15% design buffer (thorough plan
doc).*

## Done when

- `construct/vocabulary/project.cue` models the project noun (fields,
  status enum, lifecycle) and `pkg/vocab` exposes it; no consumer hardcodes
  the enum.
- `sdlc close`'s project-file update parses through typed records rather than
  substring convention (lessons.md #167); #171 owns enabling close-time model
  validation after migrating the grandfathered brain project records.
- A project instance failing conformance is caught by a gate (which gate —
  merge instance-conformance vs close — is a design decision).
- `construct/datatype/project.md` cites the cue as schema authority; a
  drift test binds prose table ↔ model.
- xx-vocabulary skill's claim ("the system's nouns are formally modeled in
  construct/vocabulary/*.cue") becomes true for project.
- The lifecycle funnel (`ideation → defined → committed → executing → done |
  dropped`, + `paused`) and the `deadline:` attribute are in the model, with
  transition gates at define / commit / breakdown / close (close requires a
  retro entry).
- Two-phase estimation is designed: a Phase-A (PRD-level) vocabulary + a
  project-level ledger row at close (Phase-A estimate vs Σ mvp_scope
  actuals — the fog factor).
- Kanban split holds in the tooling: baseline stored, progression derived
  (`sdlc project status` over live cross-repo issue state), re-forecasts
  logged — no hand-maintained board anywhere.

## Plan

- [x] brainstorm: taxonomy, lifecycle funnel, two-phase estimation,
      kanban baseline/derived/logged split, retro mechanism (Spec)
- [ ] design at start-plan: cue shape (esp. cross-repo discovery),
      transition guard mechanics, Phase-A estimate vocabulary, which gate
      owns conformance, verb set — single multi-boundary issue (dogfood
      deferred; #171 consumes the finished model) →
      `workshop/plans/000180-project-vocabulary-model-plan.md`
- [x] M1 — model + binding: project.cue (funnel, baseline guard, discovery,
      scaffold, laws) + vet block, pkg/vocab lifecycle-helper extraction,
      `vocab.Project()`, kind-keyed `ArchiveSubdir` (+projects)
- [x] M2 — typed parsing + conformance: `internal/project.Doc`/`Task`,
      tick mutators re-implemented over Doc (same contract), validate gate
      generalized to a noun table (project instances at push/merge)
- [x] M3 — verbs: `sdlc project` family (new/list/show/set-status/validate),
      model-derived helptext, named-guard registry (unknown guard = refusal)
- [x] M4 — derived board + calibrated close: `project status` (computeBoard:
      frontier, Σ remaining, dep-subgraph threads), `project retro` +
      stale-retro nudge in the issue-close gate, `project close` (retro gate,
      fog-factor ledger row, archive to history/projects), Phase-A method doc
      v1 in brain; live-fixture dogfood pass
- [ ] M5 — docs + drift: datatype prose demoted to cite the cue, prose↔model
      drift test (bite-proofed both ways), atlas + xx-vocabulary claim

## Log

### 2026-07-16 — M5.1 datatype prose demotion

`construct/datatype/project.md` now treats `project.cue` as schema authority,
documents the modeled funnel and commit-time deadline baseline, teaches the four
gated scaffold sections, and points authoring/search/archive behavior at
`workshop/projects/` and `workshop/history/projects/`. The task-detail,
single-operator, scope-event, and MVP disciplines remain intact (ARCH-PURPOSE).

### 2026-07-16 — M5.2 prose/model drift binding

`TestProjectProseCitesModel` derives every status and scaffold section from
`vocab.Project()`, requires the schema-authority citation, and forbids both
retired `active` byte forms. The bite check failed against the stashed old
prototype on all intended predicates, then passed after restoration (ARCH-DRY,
ARCH-PURPOSE).

### 2026-07-16 — M5.3 atlas and served-skill sweep

The vocabulary atlas now maps the datatype-prose drift binding and #171's
legacy-record/close-validation handoff. Both served `xx-vocabulary` skill copies
already enumerate `project`, so no regeneration was needed. The live shadow-doc
sweep corrected `construct/datatype/product.md`'s canonical project-home example
to `workshop/projects/<slug>.md`; historical issue/plan migration references
remain as provenance (ARCH-DRY, ARCH-PURPOSE).

Verification: `go test ./... -count=1`; vocabulary vet and generated-output
freshness; live `project validate` and `project status` against
`project-management-primitive`; datatype served-output assertions; live-doc
shadow sweep; and `git diff --check` pass.

### 2026-07-16 — M4.3 project close
- 2026-07-16: closed M4 — Fog ledger insertion is pure, fence-aware, section-scoped, and safe at EOF with or without existing rows; full suite, expanded race suite, vocabulary vet, and diff check pass.; review verdict: FIX-THEN-SHIP
- 2026-07-16: closed M4 — Post-fourth-REWORK: finite-positive validation rejects quoted/unquoted NaN/Inf/zero and close seam non-finites; stale-retro status and board name use typed metadata; close/drop/blocked meaning derives from lifecycle events; today captured once; M5 core-concept rows marked planned; full and expanded race suites plus vocabulary vet and diff check pass.; review verdict: SHIP

Added the dedicated close/drop boundary: lifecycle edges remain model-owned;
retro and fog-ledger bypasses are explicit cataloged gates; Phase-A fog rolls
up measured issue actuals; terminal records archive through
`vocab.ArchiveSubdir`. The process-manual scanner now recognizes the full
two-word `project close` catalog key in both Claude and Codex transcripts
(ARCH-DRY, ARCH-PURPOSE). Hermetic close/gate/scanner tests, `go test
./cmd/sdlc/... -count=1`, and `git diff --check` pass.

### 2026-07-16 — M4.4 Phase-A method

Captured the project estimation method at
`brain/data/life/42shots/velocity/estimate-logic-project-v1.md`: workstream
midpoints, explicit default/median fog, Phase-A recording, and the close-time
calibration ledger. `sdlc project` help now links it and lists the complete M4
command surface. Extending `estimate-source` to dispatch both issue and project
methods remains a natural follow-up, intentionally outside #180.

### 2026-07-16 — M4.5 live-run correction

The process-level fixture exposed a semantic mismatch: an unblocked `working`
issue disappeared from frontier because the implementation used `IsOpen`
(not-yet-started) where the Spec says “what's unblocked.” Root-cause tracing
confirmed the lookup and dependency graph were correct. Added the failing
active-issue regression and changed frontier membership to non-terminal after
the blocked checks (ARCH-PURPOSE).

### 2026-07-16 — M4.5 process-level dogfood

Built the feature binary and ran it from symlinked cwd
`/tmp/ariadne-180-live-link` over a scratch Git repo: `project new`; substantive
PRD; ideation→defined; committed transition first refused without
deadline/planned_finish, then passed with reality evidence; two dependency-linked
fixture issues seeded; executing transition passed with coverage evidence;
status rendered `frontier: #1`, `blocked: #2`, one thread, and 5h remaining;
retro appended; both issues/ticks closed; final status rendered 2/2 and 0h;
`project close --no-ledger` emitted its cataloged acknowledgement and archived
the record to `workshop/history/projects/alpha.md` with Phase-A 6h, actuals 4h,
fog 0.67. This exercises command IO, `$PWD`/symlink repo discovery, cross-file
lookup, mutation, and archive behavior beyond unit seams.

### 2026-07-16 — M4.5 full-suite integration fixes

The first `go test ./... -count=1` exposed two broader drift guards omitted by
the focused M4.3 run: nested `project close` did not yet call the #176 brain/
non-SDLC spine guard, and fixed process-manual/repo-lock census tests still
pinned the pre-M4 command set. Added guard-first execution (manual slug check
keeps it ahead of Cobra required-flag validation), corrected the multiword test
argv, and widened the lock and 14-gate pins. Targeted regressions pass; the CUE
vocabulary harness passes via `bash construct/vocabulary/vet_test.sh` (the file
is intentionally/non-regressively non-executable).

Re-run evidence: `go test ./... -count=1` passes across every package;
`bash construct/vocabulary/vet_test.sh` passes; `git diff --check` passes.

### 2026-07-16 — M4 boundary review: REWORK

The gate-owned fresh review found three blocking contract gaps: close selected
guard names instead of failing closed over the full modeled list; unavailable
MVP actuals could create a false partial fog row; and the plan mislabeled the
human Phase-A method as PURE. Important findings were non-atomic project/ledger
writes, missing README runnable examples, and incomplete documentation of the
real-instance dogfood. Remediation: close-owned complete guard registry;
availability-aware actual parsing and complete-only calibration; staged writes
with compensation and forced-failure coverage; Phase-A reclassified as an
Integration/process document; README command/bypass examples; live status +
retro dry-run against the real project. Added the general prevention rule to
`workshop/lessons.md` (ARCH-DRY, ARCH-PURPOSE).

Real-instance evidence after remediation: `project status --slug
project-management-primitive` rendered 0/3, 8.1h remaining, frontier
`ariadne#180, ariadne#171`, blocked `ariadne#182`, and two dependency threads;
`project retro --dry-run` rendered the same board in a dated retro without
mutating the still-honest ideation record.

Remediation verification: `go test ./... -count=1`, the review's focused race
suite (`go test -race ./cmd/sdlc ./cmd/sdlc/internal/project
./cmd/sdlc/internal/processmanual -run 'Project|Board|Retro|Gate|TwoWord'
-count=1`), `bash construct/vocabulary/vet_test.sh`, and `git diff --check` all
pass.

### 2026-07-16 — M4 re-review: REWORK (malformed estimate)

The second fresh review confirmed the guard/calibration/transaction/docs fixes
and found one remaining derived-board bug: `lookupIssueMeta` discarded
`ParseFloat` errors for `estimate_hours`, turning malformed input into a
plausible zero-hour forecast. Added a real sibling-repo filesystem regression
that proves `estimate_hours: invalid` becomes an explicit board warning and
cannot enter remaining-hours/frontier calculations; parsing now fails with the
ref and bad value. Also narrowed the transaction comment to distinguish
compensated commit-step failures from post-commit backup-cleanup errors
(ARCH-PURPOSE).

### 2026-07-16 — M4 third review: REWORK (YAML + Phase-A semantics)

The third review found that M4 consumers still interpreted model-valid YAML
through flat line strings (breaking quoted status and block-form scope/deps),
and that Phase-A absence/invalidity collapsed into an implicit ledger bypass.
It also requested compensation coverage after ledger replacement and the
planned last-retro age. Remediation adds a shared typed YAML metadata decoder,
a shared absent/valid/invalid Phase-A parser used by both guard and close,
explicit no-ledger acknowledgement for legacy absence, invalid/zero refusal,
an injected post-ledger archive-rename failure proving both originals restore,
and age rendering from injected `today`. Pure and real-IO regressions cover
quoted/flow/block forms; live project status remains correct (ARCH-DRY,
ARCH-PURPOSE).

Third-review remediation verification: `go test ./... -count=1`; focused race
suite including Phase/Metadata; `bash construct/vocabulary/vet_test.sh`; live
`project status`; and `git diff --check` all pass.

### 2026-07-16 — M4 fourth review: REWORK (non-finite + residual consumers)

The fourth review found that NaN/Inf could evade a `<= 0` actual check and enter
the fog ledger; board name and stale-retro nudge still used raw scalar reads;
close dates were sampled repeatedly; close/blocked semantics retained enum
literals; and two Core concepts rows claimed planned M5 files as delivered.
Remediation rejects non-finite/non-positive metadata (quoted and YAML numeric),
defends the injected close seam, moves the remaining consumers to typed
metadata, captures one transaction date, derives close/drop and blocked meaning
from model lifecycle events, and labels M5 rows `planned M5`. Focused regressions
cover every edge (ARCH-DRY, ARCH-PURPOSE).

Fourth-review remediation verification: full `go test ./... -count=1`, expanded
focused race suite including vocab event/number tests, vocabulary vet, and diff
check all pass.

### 2026-07-16 — M4 later review: REWORK (duplicate logical scope refs)

A later gate-owned fresh review found that exact or alias-equivalent
`mvp_scope` entries could count one issue twice and permanently inflate the
Phase-A fog calibration. Close now canonicalizes every reference to a resolved
repository-plus-issue-ID identity and refuses duplicates before looking up
actuals or preparing either durable mutation. Regressions cover
`[ariadne#1, ariadne#1]` and `[ariadne#1, '#1']` (ARCH-PURPOSE).
Verification: `go test ./... -count=1`; the expanded focused race suite;
`bash construct/vocabulary/vet_test.sh`; and `git diff --check` all pass.

### 2026-07-16 — M4 follow-up review: REWORK (unavailable peer compatibility)

The next review found that filesystem-required duplicate normalization changed
an absent peer from an incomplete actual into an unconditional error, preventing
the documented `--no-ledger` degraded close. Canonicalization is now
best-effort: resolvable aliases share a durable path identity; absent peers use
a normalized repository token; unparseable refs remain lookup failures. The
regression proves ordinary close refuses incomplete actuals while
`--no-ledger` archives with `actuals: incomplete` and `fog: n/a`
(ARCH-PURPOSE).
Verification: `go test ./... -count=1`; expanded focused race suite;
vocabulary vet; and `git diff --check` pass.

### 2026-07-16 — M4 further review: REWORK (alias-sensitive threads)

The next review found that the derived dependency graph keyed tasks and deps by
authored ref strings, so `ariadne#3` and `#3` could produce false parallel
threads. Issue lookup now supplies the shared canonical repository-plus-ID
identity; board maps use it while display preserves the first authored ref.
Duplicate task refs contribute one effort value/component. Regressions cover
current-repo aliases, peer-prefix aliases, and exact duplicates (ARCH-DRY,
ARCH-PURPOSE).
Verification: full `go test ./... -count=1`; expanded focused race suite;
vocabulary vet; and `git diff --check` pass.

### 2026-07-16 — M4 structural review: REWORK (fog-ledger EOF/heading)

The next review found a valid EOF-without-newline panic in fog-ledger insertion
and unanchored heading selection that could target prose/fenced examples. The
pure ledger transformer now uses the shared fence-aware Markdown line scanner,
inserts into the first contiguous table of the real level-two section, and
reconstructs lines without assumed delimiter bytes. Regressions cover divider
and existing-row EOF plus inline/fenced fake headings (ARCH-DRY, ARCH-PURE,
ARCH-PURPOSE).
Verification: full `go test ./... -count=1`; expanded focused race suite;
vocabulary vet; and `git diff --check` pass.

### 2026-07-16 — M4 boundary finalized

The final gate verdict is `FIX-THEN-SHIP`: all code and architecture checks
passed; the sanctioned follow-up updated `atlas/workflow/ledger-landscape.md`
from the brain-era, hand-maintained portfolio description to per-repository live
projects, derived status, and terminal archive residency. This final record
supersedes the earlier provisional “closed M4 / SHIP” Log line, which was
written before later gate re-runs uncovered and resolved additional findings.

### 2026-07-15

Filed from the #171 thread (operator): "is project a datatype? we should
lift it to be properly schematized just like issue and think about
processes around it." Current state verified: datatype prose exists,
vocabulary model does not; sdlc's project gate parses by convention.

### 2026-07-15 — brainstorm converged (operator + agent)

Operator refinements folded into the Spec: (1) two-phase estimation follows
the SAME calibration process with different vocabularies/axes per phase;
(2) kanban has two distinct aspects kept separate — the committed baseline
("parallel threads → 2 weeks") vs live progression + re-forecast — mapped to
stored-baseline / derived-progression / logged-re-forecasts; (3) lifecycle
enum adopted as proposed, PLUS projects are time-bound, not issue
containers → `deadline:` attribute set at commit; (4) project close runs a
mandatory project retro. Organizing insight recorded: the project lifecycle
is the issue lifecycle one level up, so the artifacts follow fractally (one
file growing gated sections; derived views; calibrated loops). Dogfood: the
project-management lift itself becomes the first project file.

### 2026-07-15 — design start: three parked questions settled (operator)

At `sdlc start-plan`, the operator resolved the continuation's open
questions: (1) **dogfood deferred** — model fully first; #180 stays a
single multi-boundary issue, the first real project file comes after the
machinery ships (rejected: creating the file mid-flight or hand-authoring
it up front); (2) **ordering: model first** — #171's gate lift consumes
the finished model as its own plan (rejected: one interleaved plan);
(3) **archive-on-done: `workshop/history/projects/`** per #181's layout
(`vocab.ArchiveSubdirs` widens; live portfolio = directory membership),
rejected: stay-in-place records. Spec updated in place where these were
parked.

### 2026-07-16 — durable plan authored + reviewed
- 2026-07-16: closed M3 — go test ./... passes; construct/vocabulary/vet_test.sh passes; hostile YAML scalar scaffold round-trips and passes process-level vocabulary validation; shared ResolvePath rejects traversal for new/show/validate/set-status before IO; README, atlas, plan Core concepts/revision, and lessons updated; live dogfood list/show/validate succeeds; review verdict: FIX-THEN-SHIP
- 2026-07-16: M3 built — `sdlc project new/list/show/validate/set-status`; scaffolds and lifecycle facts derive from #Project, command IO wraps pure render/summary/guard cores, unknown model guards refuse, and done stays close-owned. `go test ./...`, vocabulary vet harness, and live list/show/validate against the dogfood project pass (ARCH-DRY, ARCH-PURE, ARCH-PURPOSE).
- 2026-07-16: closed M2 — go test ./... and construct/vocabulary/vet_test.sh pass; make vocab-embed clean; live project-management-primitive conforms to #Project; scratch sdlc push refuses workshop/projects/bad.md specifically because status shipped is outside the project enum; milestone-increment deviation regression pinned; review verdict: FIX-THEN-SHIP
- 2026-07-16: M2 review fixes — issue discovery now falls back to vocab.Issue().Discovery().Home (ARCH-DRY); fenced headings no longer corrupt section spans (ARCH-PURE); Doc.Tasks is Breakdown-only while close ticking uses an explicit whole-document compatibility seam (ARCH-PURPOSE); #171 owns close-time validation after legacy migration. Targeted regressions and full-suite evidence recorded in the M2 boundary commit.
- 2026-07-16: closed M1 — go test ./... + vet_test.sh green (bare); invalid fixture fails vet on the enum conflict specifically; live pm-primitive instance vets clean against #Project; embeds byte-identical under vocabulary binary; ArchiveSubdir guard test widened, 11 call sites migrated; review verdict: FIX-THEN-SHIP

Plan landed at `workshop/plans/000180-project-vocabulary-model-plan.md`
(5 milestones, M1–M5 as review boundaries). Three fresh-eyes chunk reviews
dispatched per the writing-plans skill: chunk 1 approved; chunks 2+3 found
real defects, all folded — a would-be-vacuous invalid-model fixture (fixed:
self-contained copy so vet fails on the enum conflict, verified failure
mode), a `paused→done` model/verb contradiction (fixed: close requires
`executing` exactly, resume-first pointer — ARCH-PURPOSE, verb never
bypasses the model), a dir-override regression in the noun-table validate
gate (fixed: issue row keeps `f.IssuesDir`/`WF_ISSUES_DIR`), a silently
dropped Spec computation (fixed: `Threads` dep-subgraphs added to the
board), and a drift-test regex that matched nothing (fixed: exact-byte
assertions + stash-based bite-proof). Chunk-2 reviewer ran the plan's CUE
through `cue vet`/`cue export`/`validate-instance` end-to-end — model
valid, baseline guard bites, JSON carries what ProjectModel expects.
Estimate set: 8.1h (v3.1 Method A, itemized in `## Estimate` at change-code:
design 3.3h ×1.15, impl 4.28h incl. five 0.2h boundary reviews). *Produced via
`brain/data/life/42shots/velocity/estimate-logic-v3.1.md` against
`baseline-v3.1.md`. Method A only.*

### 2026-07-16 — M1 built (model + binding)

Four commits on the feature branch: project.cue (invalid fixture verified to
fail on the enum conflict, not a missing ref; live pm-primitive instance
vets clean against #Project — the plan-Revisions requirement), lifecycle
helper extraction (behavior-preserving, pins unchanged), vocab.Project()
(embed byte-identical under the vocabulary binary; conformance test pins the
deliberate paused→done absence), kind-keyed ArchiveSubdir with all 11 call
sites migrated and the source-scan guard widened to projects. Full suite +
vet_test.sh green (run bare, per lessons.md).

### 2026-07-16 — dogfood reversal + calendar-estimator spun off

At plan approval the operator (1) identified the effort→calendar gap ("we
need a higher level time estimator") — filed as #182, deliberately out of
#180 scope, upgrading the `reality-check` guard from evidence-flag to
computed check later; (2) reversed the dogfood deferral: created
`workshop/projects/project-management-primitive.md` at ideation as the
guinea pig (see the plan's Revisions). The plan's M2 live-check and M4
dogfood pass now run against a real instance instead of scratch fixtures.

### 2026-07-15 — residency dir: workshop/projects/

Operator: project files live in `workshop/projects/` (per coding repo;
plural confirmed) — the workshop/ family, alongside issues/plans/targets,
not the brain-era `data/project/` path. Folded into the cue discovery
candidate; archive-on-done → `workshop/history/projects/` per #181's
subfolder layout, decided at design.
