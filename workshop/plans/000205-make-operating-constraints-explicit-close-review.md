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
