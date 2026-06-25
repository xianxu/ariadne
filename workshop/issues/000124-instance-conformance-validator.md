---
id: 000124
status: working
deps: [ariadne#122]
github_issue:
created: 2026-06-24
updated: 2026-06-25
estimate_hours: 4.55
started: 2026-06-25T12:14:37-07:00
---

# Instance-conformance: validate typed-markdown artifacts against their datatype schema (extract then cue vet)

## Problem

#122 makes `construct/vocabulary/issue.cue` the formal type of an issue (frontmatter
enum/fields + lifecycle), and wires `sdlc`'s *verbs* to it. But nothing validates that
a real `workshop/issues/NNNNNN-*.md` **file** conforms to that type. A direct edit —
an LLM "tidying" frontmatter to `status: in-progress` (the enum says `working`), a
bulk migration, a mangled merge, a `nous-resolve` AI-merge — bypasses the verb guards
and lands an ill-formed instance that no tool flags. Symptoms are silent: `claim`
assumes already-claimed, `close` clobbers, `state` mis-categorizes. Because issue files
publish to `main` (the bulletin board), the bad value propagates to peer agents.

The shell around a typed artifact's *shape* is today only **prose** (`AGENTS.md`: "do
NOT hand-edit `status:`") — a soft shell an LLM violates by default, since the Edit
tool knows nothing about the lifecycle. The vocabulary layer makes a **hard** shell
possible for the first time (there's now a formal `#Issue` to vet against); this issue
builds it. Generalizes beyond `issue`: it's the **instance-conformance** closed-loop
for *all* typed markdown (datatype = prose skill + schema). Design captured in
`workshop/pensive/2026-06-24-01-pensive-cue-schema-layer-nouns-verbs.md` ("Typing
markdown — the general frame").

## Spec

A validator: **`artifact → (locator) → list of typed instances → cue vet each`.**

- **The validator is well-formedness only.** Enum membership, required fields, field
  types, required/allowed sections. It does **not** check semantic quality — `Spec ≥ 50
  words` and friends are soul-checks in disguise and belong to the LLM, not here (see
  #122 Log + pensive). The schema defends the skeleton; the LLM owns the soul.
- **The extractor/locator is the genuinely-new component** (`cue vet` is the easy
  half). Per type: where do instances live and how are they normalized into the
  schema's shape? Generic locators for frontmatter (YAML) and sections (`## ` split);
  bespoke locators for fenced blocks / `- [ ]` lines. Open design question: can a
  locator be *declared* per datatype, or do bespoke ones stay imperative glue?
- **No deterministic write-back.** The validator emits clear diagnostics; the fix is the
  LLM (or the human in parley) editing the markdown, then re-validate — same agentic
  loop. Deterministic = the check; LLM = the actuator.
- **One binary, many triggers** (single mechanism): the agent runs it after an Edit (or
  a PostToolUse hook fires it), parley shells out on save (LSP-style inline
  diagnostics), the pre-merge gate runs it before `main`, CI runs it on PR. The type
  lives in *one* `construct/vocabulary/*.cue` all of them read.
- **Reuse, don't rebuild:** the frontmatter parse (`pkg/frontmatter`), the
  `cmd/vocabulary` cue runner + DAG-merge (#122 M2), and the section-extraction the
  `sdlc` structural gates already do imperatively.

## Done when

- A validator (`sdlc issue validate` and/or a `vocabulary check-instance` surface) takes
  a typed markdown file, extracts its instances, and `cue vet`s them against the type;
  a `status: in-progress` issue file is **rejected with a clear diagnostic**, a valid
  one passes.
- It is wired at ≥1 fail-closed boundary — the **pre-merge / `sdlc push` gate** (the
  high-value one: catches it before `main`) — and is invocable on demand by an agent.
- Diagnostics are actionable enough that an LLM, given the output, fixes the file and
  re-validates clean (the agentic-loop write-back).
- Generalizes past `issue`: at least one other datatype (e.g. `pensive`) gets a
  `construct/vocabulary/<type>.cue` and validates through the same path — proving the
  locator/validator isn't issue-specific.
- No semantic-quality checks smuggled into the validator (well-formedness only).

## Estimate

```estimate
model: estimate-logic-v2
familiarity: 1.0
item: smaller-go-module    design=0.2 impl=1.0
item: smaller-go-module    design=0.2 impl=0.7
item: typed-data-prototype design=0.1 impl=0.4
item: milestone-review     design=0.0 impl=1.8
design-buffer: 0.30
total: 4.55
```

Build is 3 milestones, each a review boundary. M1 (smaller-go-module, impl 1.0) is the
meatiest: the `validate-instance` engine + the `cue vet -d` spike + completing `#Issue` to a
closed all-fields type (null-handling fiddly) + the generic section check + a real-corpus
test harness. M2 (smaller-go-module, impl 0.7) wires the fail-closed pre-merge gate across
push+merge+preflight + `sdlc issue validate` + a process-level vocabulary fake. M3
(typed-data-prototype, impl 0.4) is `pensive.cue` + the genericity-proof test.
milestone-review = 3 × 0.6 (one fresh review per boundary). estimate-logic-v2 runs ~2.3× high
vs ship-wall-clock (#127); this is the honest un-compensated v2 number for a clean ledger point.

## Plan

Design captured in `workshop/plans/000124-instance-conformance-validator-plan.md` (locator =
generic+imperative; engine = one `vocabulary validate-instance` binary; **frontmatter on every
changed issue, sections on newly-ADDED issues only** reusing `structural.go`'s policy;
**open** `#Issue`; loud escape hatch). See `## Revisions` for the section + open-schema pivots.

- [x] **M1 — validator engine: generic frontmatter conformance.** Spike DONE (`cue vet -d '#Issue'`); **open** `#Issue` (add `...`; keep `id`/`status`/`estimate_hours?`/`actual_hours?`+done-guard; NO `requiredSections`); `CueRunner.VetInstance`; pure cue-stderr→diagnostic transform (fixture-tested); `vocabulary validate-instance --type <noun> <file>`. Tests: whole real issue corpus passes FRONTMATTER; `status: in-progress` (bad enum) / `statuss:` typo (→ `status` absent → required-field fail) / `done`-missing-actuals each rejected with a clear diagnostic.
- [ ] **M2 — wire the fail-closed gate + section overlay + on-demand surface.** Export `issue.CheckSectionsPresence` from `cmd/sdlc/internal/issue` (Spec / Plan-has-items / Done-when-bullet-or-`related:`) — `structural.go` composes its ≥50-word check on top; NO new pkg (both callers are in cmd/sdlc). `sdlc issue validate [--issue N|<file>|--all]` (full check). `validateChangedIssues` in `preflight.go` BEFORE the judge loop (push + merge): frontmatter on ALL changed issues, sections on newly-ADDED only (git `A`); **loud** `--no-validate`/`--force`; process-level vocabulary fake. Tests: rejects a modified-file bad-`status`; rejects an added file missing `## Plan`; PASSES a window modifying a legacy ticket lacking `## Done when` (grandfathered); `--no-validate` warns loudly.
- [ ] **M3 — generalize to a second datatype.** `construct/vocabulary/pensive.cue` (`#Pensive`, closed; `mode` enum; no required sections); validate a real `workshop/pensive/*.md` through the same frontmatter path. Tests: real pensive passes; `mode: musing` fails — proving the only per-datatype addition is the `.cue`.

## Revisions

### 2026-06-25 — section validation: forward-only, single-sourced from structural.go
**Reason:** the plan-quality judge MEASURED the corpus and disproved the original flat
`requiredSections: [Spec, Plan, Done when]` design — `000052` (status `working`, in-flight) has
no `## Done when`, ~23 history files lack it, and a flat list diverges from `structural.go`'s
Done-when-**or**-`related:` fallback. A flat hard gate would retroactively fail valid legacy
tickets and double-source the section policy.
**Delta:** (1) **Frontmatter** validation stays hard + universal (every changed issue) — it's the
clean invariant and still catches the motivating bad-`status` case on existing tickets. (2)
**Section** validation is enforced at the gate **only on newly-ADDED issue files** (grandfather
legacy/in-flight; validate forward — operator). (3) The section policy is **single-sourced from
`structural.go`** (lifted to a shared `pkg/`), not a new `requiredSections` list in the `.cue`.
(4) Added a **loud escape hatch** (`--no-validate`/`--force`, announced on use —
[[feedback_escape_hatch_loud_claim]]). Done-when's section criterion is preserved, scoped to
added files.

### 2026-06-25 — `#Issue` open, not closed
**Reason:** the plan-quality judge measured the frontmatter field set (not just sections) and
found a *closed* `#Issue` would reject real files carrying `target:`/`references:`/`related:`
(incl. `000122`, this issue's dep) — and named the deeper cost: a closed schema is a field
allowlist that must track organically-growing frontmatter, and false positives at a fail-closed
merge gate train `--no-validate`.
**Delta:** OPEN `#Issue` (add `...`) instead of enumerating + closing. It still catches the
high-value cases — bad `status` *value* (enum) + a typo'd *required* field (`statuss:` →
`status` absent) — and only misses optional-field typos (low-stakes), in exchange for robustness
+ zero false positives at the gate. Schema change shrinks to one `...`. Also (judge): section
policy is an exported `issue.CheckSectionsPresence` in `cmd/sdlc/internal/issue`, NOT a new
`pkg/` (both section callers are in cmd/sdlc).

## Log

### 2026-06-24

- Filed as a follow-up to #122 (deps: ariadne#122 — needs the vocabulary layer + the
  `cmd/vocabulary` cue infra). Motivated by the hand-edited-bad-status walkthrough: the
  formal enum from #122 guards *verbs*, not *files*; this closes the instance-conformance
  loop so a typed artifact actually defends its own shape.

### 2026-06-25

- **Design (start-plan + Explore digest of the vocabulary/validation infra).** Durable plan:
  `workshop/plans/000124-instance-conformance-validator-plan.md`. Resolved the Spec's open
  questions (operator-confirmed the two architectural forks):
  - **Locator = generic + imperative** (frontmatter YAML split + `## ` section split); defer a
    declared-per-datatype locator DSL until a 3rd datatype needs it (Simplicity-First / YAGNI).
  - **Engine = one binary**: `vocabulary validate-instance --type <noun> <file>` (reuses
    `resolveVocab` + `CueRunner` in place); sdlc's gate + `sdlc issue validate` shell out to it
    (matches the Spec's "one binary, many triggers").
  - **Sections via `requiredSections` cue-data + Go presence-check** — the `.cue` stays the
    single complete source for the type without forcing markdown through cue unification.
  - **Complete + keep `#Issue` closed**: discovered `#Issue` currently models only
    {id,status,estimate,actual} but is closed-by-default, so a real file (with
    deps/github_issue/created/updated/started) would fail vet on every "unknown" field. Completing
    it is required for the validator to work — and closedness then catches `statuss:`-style typos
    (true well-formedness). Touches the #122 artifact `issue.cue` (ARCH-PURPOSE: the type must
    serve its now-real consumer).
  - **Load-bearing mechanism:** `cue vet -d '#Issue' <frontmatter.yaml> issue.cue` — M1 spikes
    this first (fallback: synthesize `out: #Issue & {<data>}` and vet). Extends `CueRunner` with
    `VetInstance`.
  - Scope: full M1–M3 build (operator-confirmed), 3 review boundaries.
- **Plan-quality FAILURE addressed (2026-06-25).** The judge caught a real cross-artifact
  inconsistency: the durable plan still had 4 milestones while the issue `## Plan` (which
  `milestone-close` reads) had 3 — reconciled the durable plan down to 3 (folding frontmatter +
  section validation into M1). Also spiked the two flagged unknowns before coding:
  (1) `cue vet -d '#Issue'` works (bad-status rejected, closed-#Issue rejects extra fields);
  (2) the date/timestamp-vs-`string` trap does NOT materialize — cue's YAML loader keeps
  unquoted `created:`/`started:` as strings, and empty `github_issue:`→null + `deps: []` vet
  clean. Folded the judge's refinements into the plan: pure cue-stderr→diagnostic transform
  (unit-tested on fixtures), `.yaml` temp extension, reuse `SectionBody`'s `ok` bool (lift to
  `pkg/frontmatter` is forced — cmd/vocabulary can't import cmd/sdlc/internal).
- **2nd plan-quality FAILURE addressed (2026-06-25) — the important one.** The judge MEASURED
  the corpus and disproved the flat `requiredSections` design (see `## Revisions`): it would
  retroactively fail valid in-flight tickets (`000052`/working has no `## Done when`) and diverge
  from `structural.go`. Operator steer: do BOTH frontmatter + section validation, but
  "validate forward, don't fail old tickets." Resolved: frontmatter hard+universal; sections on
  newly-ADDED issues only, policy single-sourced from `structural.go` (lifted to shared `pkg/`);
  loud `--no-validate` escape. A measure-before-rebuild win — the design was wrong by intuition,
  right by measurement.
