# Boundary Review — ariadne#205 (whole-issue close)

| field | value |
|-------|-------|
| issue | 205 — Make operating constraints explicit |
| repo | ariadne |
| issue file | workshop/issues/000205-make-operating-constraints-explicit.md |
| boundary | whole-issue close |
| milestone | — |
| window | 00fe01d88eb75f15be9a99b19096702b20d24967..c3eb3cc076f461ff6bc4b3a02e38214e94da3920 |
| command | sdlc close --issue 205 |
| reviewer | codex |
| timestamp | 2026-08-29T18:19:56-07:00 |
| verdict | REWORK |

## Review

```verdict
verdict: REWORK
confidence: high
```

The single-source registry design, generated prompts, atlas updates, and pure test helpers are sound. Shipping is blocked because the new principle omits the Spec’s explicit input-scale constraint, and the plan’s Core concepts table claims nonexistent or unmodified entities. The CLI tests also do not prove semantic delivery as required by Done-when.

1. Strengths

- `cmd/sdlc/internal/judge/architecture.md:96` adds one concise registry entry without introducing parallel production configuration.
- `cmd/sdlc/internal/judge/judge_test.go:168-212` cleanly isolates entries and clauses with pure helpers that fail closed on malformed structure.
- All four architecture-aware prompts derive the complete registry through `ArchitectureBlock`; the shadow-sweep found no hand-maintained prompt-body copy.
- Atlas coverage is present in `atlas/workflow/architecture-principles.md`, `atlas/workflow/sdlc-binary.md`, and `atlas/index.md`. No README update is needed because no command syntax, flag, or usage step changed.

2. Critical findings

- `cmd/sdlc/internal/judge/architecture.md:100` — `ARCH-PURPOSE`: the Spec requires “workload/input scale and growth,” but the delivered principle says only “workload and growth.” That can fail to prompt for dataset/request/model size—the material parameter motivating this issue. `cmd/sdlc/internal/judge/judge_test.go:138` repeats the omission, so tests bless the under-delivered contract. Add explicit input scale and pin it in the clause contract before regenerating derived goldens.
- `workshop/plans/000205-make-operating-constraints-explicit-plan.md:19-41` — the Core concepts table is not code-traceable as required. `ArchitectureMarkerSet`, `ArchitecturePrincipleDelivery`, and `ArchitecturePromptGoldens` are conceptual aliases rather than greppable entities, while `ArchitectureMarkers`, `ArchitectureBlock`, and the delivery files are labeled modified despite no diff at their stated production paths. Append a revision naming actual symbols/files and accurately distinguishing the new registry entry, changed golden files, and unchanged consumers whose output changes through embedding.

3. Important findings

- `cmd/sdlc/archprinciples_test.go:12-23` and `cmd/sdlc/startplan_test.go:21-42` — `ARCH-PURPOSE`: Done-when requires tests to fail if required semantics disappear from every derived delivery surface, but these CLI tests assert markers only. In a scratch copy, replacing both `ArchitectureBlock` calls with marker-only text left both tests green. Assert that each CLI output contains `ArchitectureRegistry`, or at minimum the complete `ARCH-CONSTRAINTS` entry, so marker-only delivery fails.

4. Minor findings

None.

5. Test coverage notes

- Focused architecture contract/delivery tests: pass.
- `go vet -p 20 ./...`: pass.
- `git diff --check`: pass.
- Full suite: the unchanged, pre-existing #200 archived-plan test fails; the base commit already lacks its active plan path. One additional failure is reviewer-sandbox-only because `.git/sdlc.lock` cannot be created.
- Full suite excluding those two tests: pass.
- Both `arch-principles` lens smokes: pass.
- Scratch mutation proved the CLI semantic-delivery test gap described above.

6. Architectural notes for upcoming work

- `ARCH-DRY`: pass; registry content has one production source, and prompt goldens are derived evidence.
- `ARCH-PURE`: pass; the production change is declarative and the new validation helpers are deterministic and IO-free.
- `ARCH-PURPOSE`: flag; input scale is omitted, and CLI semantic delivery is not enforced by tests.
- `ARCH-MOCK`: pass/not applicable; no external service or binary dependency was introduced.
- `ARCH-CONSTRAINTS`: pass for this implementation’s own envelope—the plan bounded prompt size, runtime mechanism, delivery coverage, and test parallelism.

7. Plan revision recommendations

Append a dated `## Revisions` entry that:

- Restores explicit input-scale coverage in the registry and contract test.
- Records semantic assertions for both CLI consumers and the marker-only mutant evidence.
- Replaces conceptual Core concepts aliases with actual greppable entities/files and correct statuses.

```findings
findings:
  - id: new
    severity: Critical
    family: operating-envelope-semantic-completeness
    title: |
      The registry omits the Spec's explicit input-scale constraint
    detail: |
      cmd/sdlc/internal/judge/architecture.md:100 reduces workload/input scale and growth to workload and growth, and cmd/sdlc/internal/judge/judge_test.go:138 pins that omission. Add explicit input scale and a regression assertion before regenerating derived outputs (ARCH-PURPOSE).
  - id: new
    severity: Critical
    family: core-concept-inventory-accuracy
    title: |
      The Core concepts table names conceptual or unmodified entities as modified code
    detail: |
      workshop/plans/000205-make-operating-constraints-explicit-plan.md:19-41 includes nongreppable aliases and claims modifications at production paths absent from the diff. Append a plan revision using actual symbols/files and statuses consistent with the committed range.
  - id: new
    severity: Important
    family: derived-consumer-semantic-coverage
    title: |
      CLI tests permit marker-only delivery with the principle body missing
    detail: |
      cmd/sdlc/archprinciples_test.go:12-23 and cmd/sdlc/startplan_test.go:21-42 stayed green in a scratch mutation after ArchitectureBlock was replaced by marker-only text. Assert full registry or complete entry delivery at both seams to satisfy Done-when (ARCH-PURPOSE).
```

---

## Re-review — 2026-08-29T18:34:59-07:00 (REWORK)

| field | value |
|-------|-------|
| issue | 205 — Make operating constraints explicit |
| repo | ariadne |
| issue file | workshop/issues/000205-make-operating-constraints-explicit.md |
| boundary | whole-issue close |
| milestone | — |
| window | 00fe01d88eb75f15be9a99b19096702b20d24967..256359af438fbb9bacb35c89082e4b92db895c8d |
| command | sdlc close --issue 205 |
| reviewer | codex |
| timestamp | 2026-08-29T18:34:59-07:00 |
| verdict | REWORK |

## Review

```verdict
verdict: REWORK
confidence: high
```

The registry and delivery implementation satisfy the issue’s functional purpose, and BR-1/BR-3 are mutation-proven. Shipping remains blocked because BR-2 has no regression that fails without its plan correction, which the claimed-fix contract explicitly requires. A new repeat semantic-coverage gap also allows every required predicate to be negated while tests remain green.

1. Strengths

- `cmd/sdlc/internal/judge/architecture.md:96-116` adds the complete principle in the single-source registry, including explicit input scale.
- All four judge prompts and both CLI seams derive the full registry; marker-only mutations fail both CLI tests.
- `cmd/sdlc/internal/judge/judge_test.go:168-212` cleanly isolates entries and clauses with IO-free helpers.
- Atlas coverage is complete. No README update is required because command syntax, flags, and usage steps did not change.

2. Critical findings

- BR-2 — `workshop/plans/000205-make-operating-constraints-explicit-plan.md:387-411`: the revised inventory is accurate by inspection, but no test fails when it is removed or regressed. The review contract says a claimed fix without such a test is `not-addressed`. Add a repository-contract test or validator that pins the authoritative inventory’s greppable names, paths, kinds, and changed/unchanged statuses.

3. Important findings

- `cmd/sdlc/internal/judge/judge_test.go:132-163,230-239` — ARCH-PURPOSE: the validator uses positive substring membership, so prefixing any of its 18 required phrases with a case-preserving “Do not” leaves the contract green. The existing negation mutant only fails because it changes `Classify` to lowercase `classify`. An actual mutation changing the registry to “do not confirm material uncertainty with the operator” passed `TestArchitectureRegistry_ConstraintsContract`.

  **This is the 2nd finding in family `operating-envelope-semantic-completeness`.** Do not fix only operator confirmation: state the rule covering all affirmative predicates and sweep all 18, such as table-driven case-preserving negation mutants backed by validation that rejects negative contexts.

4. Minor findings

None.

5. Test coverage notes

- Focused architecture and delivery suite: pass.
- BR-1 mutation removing `input scale`: expected test failure.
- BR-3 marker-only CLI mutations: expected test failures at both seams.
- New negation mutation: unexpectedly passed.
- `go vet -p 20 ./...`: pass.
- Pinned-range `git diff --check`: pass.
- Both CLI lens smokes render five complete entries.
- Full suite has one failure: `TestFleetPlanHasAuthoritativeCorrectedCoreConceptInventory`. The same stale test and absent active plan path exist at the base commit, confirming it is pre-existing.

6. Architectural notes for upcoming work

- ARCH-DRY: pass; one registry source feeds all consumers.
- ARCH-PURE: pass; validation logic is deterministic and IO-free.
- ARCH-PURPOSE: flag; the semantic guard does not enforce affirmative meaning across its enumerated contract.
- ARCH-MOCK: pass/not applicable; no external dependency was introduced.
- ARCH-CONSTRAINTS: pass for the implementation’s declared prompt-size, runtime, delivery, and test-parallelism envelope.

7. Plan revision recommendations

Append a dated revision that:

- Replaces the claim that the current negation mutant proves reversed semantics cannot stay green with an all-18-predicate mutation strategy.
- Records the regression mechanism added to pin the corrected Core concepts inventory.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      Removing explicit input scale from the registry makes TestArchitectureRegistry_ConstraintsContract fail.
  - id: BR-2
    disposition: not-addressed
    note: |
      The revised inventory is accurate, but no regression fails without it as required by the claimed-fix contract.
  - id: BR-3
    disposition: addressed
    note: |
      Marker-only replacements at both CLI delivery calls make their seam tests fail.
findings:
  - id: new
    severity: Important
    family: operating-envelope-semantic-completeness
    title: |
      The semantic contract accepts case-preserving negation of every required predicate
    detail: |
      cmd/sdlc/internal/judge/judge_test.go:132-163 validates 18 phrases by substring, while lines 230-239 mutate only one phrase and accidentally lowercase it. This is the 2nd finding in family operating-envelope-semantic-completeness: define the affirmative-semantics rule and sweep all 18 predicates rather than fixing only operator confirmation (ARCH-PURPOSE).
```

---

## Re-review — 2026-08-29T18:46:59-07:00 (REWORK)

| field | value |
|-------|-------|
| issue | 205 — Make operating constraints explicit |
| repo | ariadne |
| issue file | workshop/issues/000205-make-operating-constraints-explicit.md |
| boundary | whole-issue close |
| milestone | — |
| window | 00fe01d88eb75f15be9a99b19096702b20d24967..a22d7e5773d31c614051f881ed3b9cfe494f193f |
| command | sdlc close --issue 205 |
| reviewer | codex |
| timestamp | 2026-08-29T18:46:59-07:00 |
| verdict | REWORK |

## Review

```verdict
verdict: REWORK
confidence: high
```

The registry addition, consumer fan-out, atlas updates, and corrected Core concepts inventory are sound. BR-4 still blocks closure: the semantic guard rejects only immediately adjacent negators. An adversarial sweep inserting “Do not ever” before every required predicate passed validation for all 18, confirming that the implementation fixed one syntax rather than the affirmative-semantics class.

1. Strengths

- `ARCH-CONSTRAINTS` covers the Spec’s workload classification, scale/resource dimensions, basis, exceeded behavior, and review checks in `cmd/sdlc/internal/judge/architecture.md:95`.
- All delivery surfaces derive from `ArchitectureBlock` or `ArchitectureMarkers`; no parallel production copy was introduced.
- Both CLI seam tests now require the complete registry, preventing marker-only delivery.
- BR-2’s revised inventory accurately distinguishes changed artifacts from unchanged derived consumers, with an archive-safe regression test.
- Atlas coverage is complete. No README update is needed because no command, flag, configuration key, or invocation changed.

2. Critical findings

None.

3. Important findings

- **BR-4 remains not addressed — `operating-envelope-semantic-completeness` (`cmd/sdlc/internal/judge/judge_test.go:217-280`).** `predicateIsNegated` recognizes a negator only when it is the immediate suffix before the required phrase. Every mutation shaped as `Do not ever <required predicate>` passed `constraintsContractViolations`. **This is the 3rd finding in family `operating-envelope-semantic-completeness`.** Do not add another nearby negation token. State and enforce the class-level rule—for example, require each semantic predicate within a canonical affirmative clause whose surrounding text is constrained—and sweep all 18 predicates with separated/modifying negation variants.

4. Minor findings

None.

5. Test coverage notes

- Focused architecture, delivery, inventory, and mutation suite: pass.
- `go vet -p 20 ./...`: pass.
- Pinned-range `git diff --check`: pass.
- Both `arch-principles` lens smokes report and render all five entries.
- Adversarial separated-negation sweep: failed as expected, exposing BR-4 across all 18 predicates.
- Full `go test -p 20 ./... -count=1`: one failure in `TestFleetPlanHasAuthoritativeCorrectedCoreConceptInventory`; the referenced #200 plan is already absent at the pinned base, confirming this is pre-existing.

6. Architectural notes for upcoming work

- `ARCH-DRY`: pass—one registry source feeds every production consumer.
- `ARCH-PURE`: pass—new validation and extraction logic is deterministic and IO-free; repository IO remains in the integration test helper.
- `ARCH-PURPOSE`: flag—BR-4 fixes an immediate-negation instance but does not enforce the committed affirmative-semantics class.
- `ARCH-MOCK`: pass/not applicable—no external dependency was introduced.
- `ARCH-CONSTRAINTS`: pass for the implementation’s own declared runtime and test-worker envelope.

7. Plan revision recommendations

Append a dated `## Revisions` entry recording that the “immediately negated context” implementation did not satisfy BR-4’s affirmative-semantics rule. Replace it with the class-level invariant and enumerate the separated/modifying negation mutations applied across all 18 predicates.

```findings
dispose:
  - id: BR-2
    disposition: addressed
    note: |
      The authoritative revision uses actual symbols and artifacts, distinguishes unchanged derived consumers, and its archive-safe mutant test fails when any recorded row is removed.
  - id: BR-4
    disposition: not-addressed
    note: |
      This is the 3rd finding in family operating-envelope-semantic-completeness: all 18 predicates still validate when a modifier separates the negator from the required phrase, such as “Do not ever”; enforce canonical affirmative context rather than extending the immediate-prefix list.
```

---

## Re-review — 2026-08-29T18:59:40-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 205 — Make operating constraints explicit |
| repo | ariadne |
| issue file | workshop/issues/000205-make-operating-constraints-explicit.md |
| boundary | whole-issue close |
| milestone | — |
| window | 00fe01d88eb75f15be9a99b19096702b20d24967..4a753c6de28bf8158a43a4b44c8e1fdb5b652358 |
| command | sdlc close --issue 205 |
| reviewer | codex |
| timestamp | 2026-08-29T18:59:40-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

The pinned range fulfills issue #205: `ARCH-CONSTRAINTS` is defined once, reaches every enumerated consumer, and is guarded by semantic and delivery tests. BR-4 is addressed with verified red/green mutant evidence. No blocking findings remain.

## Strengths

- The registry entry covers the complete operating-envelope contract, including scale, basis, exceeded behavior, and review checks ([architecture.md](/Users/xianxu/workspace/ariadne/cmd/sdlc/internal/judge/architecture.md:96)).
- Exact normalized affirmative clauses replace incomplete negator enumeration ([judge_test.go](/Users/xianxu/workspace/ariadne/cmd/sdlc/internal/judge/judge_test.go:127)).
- Both CLI delivery seams assert the complete registry, not marker-only output ([archprinciples_test.go](/Users/xianxu/workspace/ariadne/cmd/sdlc/archprinciples_test.go:14), [startplan_test.go](/Users/xianxu/workspace/ariadne/cmd/sdlc/startplan_test.go:22)).
- Atlas coverage documents the new principle and its delivery graph. No README change is required because no command, flag, configuration, or usage surface changed.

## Critical findings

None.

## Important findings

None.

## Minor findings

None.

## Test coverage notes

- Focused architecture, consumer, inventory, and mutant suites: pass.
- BR-4 red-side verification: restoring the prior prefix detector caused all 18 separated-negation cases to fail.
- `go vet -p 20 ./...`: pass.
- Full suite excluding the independently verified pre-existing `TestFleetPlanHasAuthoritativeCorrectedCoreConceptInventory` failure: pass.
- The unfiltered suite’s sole failure exists at the pinned base; this range does not modify that test.

## Architectural notes

- `ARCH-DRY`: Pass — one registry source; derived consumers and goldens contain no independent production definition.
- `ARCH-PURE`: Pass — semantic validation and rendering remain pure; CLI IO seams stay thin.
- `ARCH-PURPOSE`: Pass — the shadow-sweep covers both CLI paths, four architecture-aware prompts, and boundary marker expansion.
- `ARCH-MOCK`: Pass — no external dependency or new IO behavior was introduced.
- `ARCH-CONSTRAINTS`: Pass — the declarative change stays within its stated static-rendering and 20-worker test envelope.

## Plan revision recommendations

None; the final revision matches the code and generated artifacts.

```findings
dispose:
  - id: BR-4
    disposition: addressed
    note: |
      Exact normalized affirmative clauses now guard all 18 predicates; reverting to the prior prefix detector makes every separated-negation regression fail.
```
