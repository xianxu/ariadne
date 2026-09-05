# Boundary Review — ariadne#215 (whole-issue close)

| field | value |
|-------|-------|
| issue | 215 — ARCH-ORDER architecture principle |
| repo | ariadne |
| issue file | workshop/issues/000215-arch-order-architecture-principle.md |
| boundary | whole-issue close |
| milestone | — |
| window | 03cd1d27f810fa73b519d1f0defc689b8582e0e6..d70114fd89d561c10aa15ccb1af09873b10d8601 |
| command | sdlc close --issue 215 |
| reviewer | claude |
| timestamp | 2026-09-04T22:17:18-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

The seventh registry entry lands cleanly and every wiring claim in the issue holds up under independent verification: `sdlc arch-principles` renders 7 with a derived header count, exactly the four `{{ARCH_BLOCK}}`-bearing goldens changed, their added lines are byte-identical to the registry entry (I reconstructed the diff rather than reading the commit message), the three non-registry goldens are correctly untouched, all eight `## Revisions` deltas are present in the shipped text, and the one failing test is genuinely pre-existing (`workshop/plans/000200-…` is absent at the base commit too; tracked as ariadne#210). Nothing blocks the boundary. What holds it back from a clean SHIP is the other deliverable: the atlas paragraph restates three of the registry entry's clauses in near-verbatim prose — one sentence is character-identical — in the fleet's canonical route-don't-restate file, with no test pinning the copy. That is the ARCH-DRY shape this very issue is about, and it is a five-minute fix.

## 1. Strengths

- **The single-source claim is true end-to-end, and I checked it rather than took it.** `ArchitectureBlock` derives the count from `len(ArchitectureMarkers())` (`cmd/sdlc/internal/judge/architecture.go:57`), `{{ARCH_STAR}}` derives from the same call (`review.go:44`), and exactly four templates carry `{{ARCH_BLOCK}}` (`prompts/dry.md:4`, `milestone-review.md:3`, `plan-quality.md:81`, `pure.md:4`) — matching exactly the four goldens that changed. `estimate-quality.prompt`, `plan.prompt`, `specs.prompt` carry no registry and correctly did not move. One hand-written list exists (`judge_test.go:363`) and it was updated, last, in registry order — the deliberate tripwire, not derived.
- **The golden re-capture honored `golden_test.go:35` instead of citing it.** Reconstructing the added lines independently: each golden gained the 45-line entry verbatim plus one blank separator, plus one header-count line (`6` → `7`), plus one `{{ARCH_STAR}}` line in `milestone-review.prompt`. `--numstat` 189/5 reconciles exactly to the Log's arithmetic. Nothing rode along.
- **All eight accepted review deltas actually shipped.** The Spec's post-revision block and `architecture.md:146+` are character-identical, and each delta is individually present: the ARCH-SECURE provenance-vs-temporal split, the 2^N sentence scoped to "carried across events", capacity-vs-extent against ARCH-CONSTRAINTS, target-rather-than-sweep, the falsifiable `N/A`, and the oracle clause promoted to lead `at-review`. PQ-1's failure mode ("tick every box and ship the un-revised draft") did not happen.
- **The pre-existing-failure claim is honest.** `git ls-tree 03cd1d2` shows `000200-sdlc-fleet-thread-inventory-plan.md` already absent at the base, and `workshop/issues/000210-fleet-plan-test-hardcoded-path.md` tracks it. Everything outside `cmd/sdlc` is green.
- **The cross-repo evidence survives an independent check.** `pair#182` (relaunch/park-then-resume) and `pair#185` (status-row notices) exist with the described subjects, and pair#185's own close record states the `-race` story first-hand ("quoted from ONE run of a 30%-flaky test; reproduced at 2/10") — so the oracle clause the entry leads with is grounded in a transcript, exactly as claimed.

## 2. Critical findings

None.

## 3. Important findings

**`atlas/workflow/architecture-principles.md:79-89` — the atlas restates registry clause wording near-verbatim (ARCH-DRY).** Three sentences are copies of `cmd/sdlc/internal/judge/architecture.md`, one of them exact:

| atlas | registry |
|---|---|
| `:88` "a green run of such a test is a sample of size one that reports no coverage, so it confirms whichever ordering the author happened to get" | `architecture.md:176` — character-identical |
| `:82-86` "`N/A` must be written as a falsifiable claim ("holds no state between events because X") … a bare `N/A` is exactly what the author who cannot see the ordering will write" | `architecture.md:171-173` — near-verbatim tail |
| `:61-62` "when a component holds state across events arriving from outside it, the legal states and the transitions between them are design, not an emergent property of the code" | `architecture.md:148-150` — near-verbatim |

Nothing pins the copy, so the next wordsmith of the registry strands it silently. This is the pattern the fleet has explicitly decided against three times (`construct/intents/superpowers.md:184`, `:196`, `:204` — "routes to ARCH-PURPOSE … does not restate the principle") and that this same atlas file names at `:24-27` ("asserting wording is what would let the line become a second copy of the principle"), and it is out of line with the sibling paragraphs: ARCH-SECURE (`:50-60`), ARCH-MOCK and ARCH-CONSTRAINTS all paraphrase or compress, and the ARCH-ORDER treatment is 38 lines against ARCH-SECURE's 11. Fix sketch: keep the two boundary paragraphs (ARCH-PURE sibling-not-bullet, ARCH-SECURE provenance-vs-temporal) and the `pair#182`/`#185` origin paragraph — those are atlas-only content and are the best part of the addition — and compress `:79-89` to one map-level sentence naming the two shape choices (at-plan targets rather than sweeps; at-review leads with the oracle clause) with the rationale left to `sdlc arch-principles`.

## 4. Minor findings

- `cmd/sdlc/internal/judge/architecture.md:146-190` — ARCH-ORDER is 531 words, 1.43× the next-largest entry (ARCH-PURPOSE, 371) and 5.8× ARCH-DRY; the registry grew 9,510 → 12,868 bytes (+35%) on one entry, and every registry-bearing prompt gained ~3.3 KB (dry/pure +32%). No length norm exists for an artifact injected into six prompt paths, and the `at-plan` bullet alone is ~230 words asking for six distinct things inside an entry whose own text warns against ceremony (ARCH-CONSTRAINTS lens on the prompt budget, not on runtime).
- `workshop/issues/000215-…:Done-when` — the atlas bullet asks the atlas to document "the count"; it does not, and should not. The count is derived at `architecture.go:57` and the atlas has never carried one (no "six"/"seven" ever appeared in it). The criterion is what's wrong here, not the implementation — see §7.

## 5. Test coverage notes

- The entry's text is pinned byte-exact by the four goldens, so an *unintended* wording change reds the suite. That is the drift tripwire and it demonstrably worked this round.
- `TestArchitectureRegistry_Content` (`judge_test.go:119`) derives its marker set and checks the `## ARCH-ORDER` heading plus all three lens bullets — verified passing, and the "found it in prose only" branch is armed: the entry cites ARCH-PURE/ARCH-SECURE/ARCH-CONSTRAINTS, all headed, and `ARCH-FSM` correctly stayed in the issue. Rendered marker set is exactly the seven headed names.
- Gap, consistent with precedent rather than introduced here: ARCH-ORDER has no clause contract analogous to `constraintsClauseContracts` (`judge_test.go:156`), so an *intended* `-update-golden` re-capture can weaken the at-plan/at-review clauses with no test naming what was lost. ARCH-SECURE set that precedent in #208; the class is now two entries wide. See §6.
- `go test ./...` outside `cmd/sdlc` green; `cmd/sdlc` green except the ariadne#210 hardcoded-path test, confirmed pre-existing against the base commit.

## 6. Architectural notes for upcoming work

- **ARCH-DRY** — flagged, §3; wiring itself is exemplary (one extraction site, `markersIn` shared with the deferred guard). **ARCH-PURE** — pass: the only code change is a test literal; `ArchitectureBlock`/`markersIn` stay pure over an embedded constant. **ARCH-PURPOSE** — pass: shadow-sweep run over all six consumer paths plus the deferred file (disjoint, `ARCH-AUTHORITY` only); the sole remaining hand-maintained restatement is the atlas prose in §3. **ARCH-MOCK** — N/A: no external binary or service is consumed; goldens run in-process. **ARCH-CONSTRAINTS** — see the prompt-budget Minor. **ARCH-SECURE** — N/A: the diff adds no input parsing and no credential; the registry is a compile-time `//go:embed` constant. **ARCH-ORDER** (applied to itself) — N/A as a falsifiable claim, per its own rule: the registry holds no state between events because it is an embedded constant read once at process start, and the tests are deterministic string comparisons with no injectable ordering.
- If a third entry lands without a clause contract, generalize `architectureClauseContract` from the ARCH-CONSTRAINTS-specific `constraintsClauseContracts` into a per-marker `map[string][]contract` — the machinery at `judge_test.go:150-320` (including its mutation tests) is already marker-agnostic; only the variable name and the single caller are not.
- The entry makes a bare `N/A` non-conforming, which is a deliberate divergence from ARCH-CONSTRAINTS' "mark irrelevant categories `N/A`". `plan-quality.md:76-81` treats the block as advisory with no per-marker coverage requirement, so this should not force ceremony at the gate — but it is the thing to watch. If the next few plan-quality rounds start carrying a ritual "ARCH-ORDER: N/A because …" sentence on plainly stateless issues, that is the signal to soften the clause, and `atlas/workflow/gate-state.md`'s round counts are where it will show.

## 7. Plan revision recommendations

One `## Revisions` entry on `workshop/issues/000215-arch-order-architecture-principle.md`, correcting the Done-when atlas bullet:

> ### 2026-09-04 — Done-when correction: the atlas does not carry the count
>
> Reason: close-gate review. The bullet reads "`atlas/workflow/architecture-principles.md` documents the entry, the count, and both boundaries." The atlas has never carried an entry count, and adding one would create exactly the hand-maintained restatement of a derived fact (`len(ArchitectureMarkers())`, `architecture.go:57`) that the same file's "Adding an entry" section says every site but one avoids. Delta: drop "the count" from the criterion; the shipped atlas satisfies the remaining three clauses.

If the §3 Important is taken, append the atlas trim to the same entry so the artifact records why the paragraph is shorter than the draft.

```findings
findings:
  - id: new
    severity: Important
    family: atlas-restates-single-source
    title: |
      atlas paragraph restates ARCH-ORDER clause wording near-verbatim instead of mapping to it
    detail: |
      atlas/workflow/architecture-principles.md:79-89 copies three registry sentences
      (":88" is character-identical to architecture.md:176; the N/A clause and the
      principle opener are near-verbatim). Nothing pins the copy, so a later rewording
      of the registry strands it. Fleet convention is route-don't-restate
      (construct/intents/superpowers.md:184,196,204) and the same atlas file says so at
      :24-27; sibling paragraphs (ARCH-SECURE :50-60, ARCH-MOCK, ARCH-CONSTRAINTS) all
      paraphrase. Keep the two boundary paragraphs and the pair#182/#185 origin — those
      are atlas-only content — and compress :79-89 to one map-level sentence. ARCH-DRY.
  - id: new
    severity: Minor
    family: registry-entry-length-unbudgeted
    title: |
      ARCH-ORDER is 1.4x the next-largest entry and grew the injected registry 35 percent
    detail: |
      531 words vs ARCH-PURPOSE's 371 and ARCH-DRY's 92; architecture.md went 9,510 to
      12,868 bytes, and every registry-bearing prompt gained ~3.3 KB (dry/pure +32%).
      Six prompt paths carry it and no length norm exists for the artifact. Note for the
      next entry, not a defect in this one. ARCH-CONSTRAINTS, on the prompt budget.
  - id: new
    severity: Minor
    family: derived-fact-restated
    title: |
      Done-when asks the atlas to document the entry count, which is derived and was never carried there
    detail: |
      The count comes from len(ArchitectureMarkers()) at architecture.go:57 and no
      "six"/"seven" ever appeared in atlas/workflow/architecture-principles.md. Writing
      one in would be the hand-maintained restatement the file's own "Adding an entry"
      section rules out. The criterion is wrong, not the implementation — drop "the
      count" from the bullet via a Revisions entry.
```
