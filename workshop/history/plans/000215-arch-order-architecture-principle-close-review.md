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

---

## Re-review — 2026-09-04T22:28:48-07:00 (FIX-THEN-SHIP)

| field | value |
|-------|-------|
| issue | 215 — ARCH-ORDER architecture principle |
| repo | ariadne |
| issue file | workshop/issues/000215-arch-order-architecture-principle.md |
| boundary | whole-issue close |
| milestone | — |
| window | 03cd1d27f810fa73b519d1f0defc689b8582e0e6..d4136d1a451b41397464609c7bca8d3e108e44ee |
| command | sdlc close --issue 215 |
| reviewer | claude |
| timestamp | 2026-09-04T22:28:48-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

The seventh registry entry is correctly wired and I verified every mechanical claim independently rather than reading the Log back: `sdlc arch-principles` renders 7 with a derived header count, the ARCH-ORDER entry is byte-identical inside all four `{{ARCH_BLOCK}}` goldens, `--numstat` reconciles exactly (189/5 = 46×4 entry lines + 4 header counts + 1 `{{ARCH_STAR}}` line, with the 5 deletions their counterparts), the shadow sweep finds exactly one hand-written marker list (the documented `judge_test.go:363` tripwire) and it was updated last in registry order, and the only failing test is genuinely pre-existing (`workshop/plans/000200-…` is absent, tracked as ariadne#210). BR-1's fix is real and I measured it: a 12-word-window scan of the atlas paragraph against the registry entry now finds **0** shared spans, down from 26 overlapping windows (≈6 distinct spans) pre-fix — and no sibling paragraph in the atlas restates its entry either (ARCH-SECURE/PURPOSE 0 shared 8-grams, MOCK/CONSTRAINTS 1 each). Nothing blocks the boundary. Two Minor residues remain, both the same shape: the *instances* the prior round named were fixed, but each finding's class-level rule was left unwritten — one sibling site of BR-3 still stands, and BR-1's rule has no durable home outside a `## Revisions` section headed for `workshop/history/`.

## 1. Strengths

- **The single-source property holds end-to-end, and it is enforced rather than documented.** `ArchitectureBlock` derives the header count from `len(ArchitectureMarkers())` (`cmd/sdlc/internal/judge/architecture.go:57`), `{{ARCH_STAR}}` derives from the same call (`review.go:44`), and `TestArchitectureRegistry_Content` (`judge_test.go:119`) derives its per-entry contract from `ArchitectureMarkers()` — so ARCH-ORDER got the heading + three-lens check for free, with no test edited to admit it. This prompt you are reading carries ARCH-ORDER, which is production confirmation, not a golden.
- **BR-1 was swept as a class, not patched at the cited span.** The finding's line cite had drifted and the fix says so; the implementor re-derived the overlap by measurement instead of trusting the cite, then rewrote the whole paragraph to map-level. I reproduced both numbers. The surviving 11-word overlap (`at a single-shot parse of input the component did not produce`) is the unavoidable descriptor of the *neighbour* entry in a boundary sentence — that is map content, not restatement.
- **The golden re-capture respected `golden_test.go:35` instead of citing it.** Exactly four goldens moved; `estimate-quality.prompt`, `plan.prompt`, `specs.prompt` carry no registry and correctly did not. Only `milestone-review.prompt` gained a second changed line, and it is the `{{ARCH_STAR}}` list — the one template using that token.
- **The pre-existing-failure claim is honest.** `TestFleetPlanHasAuthoritativeCorrectedCoreConceptInventory` fails on a hardcoded path to an archived plan; `workshop/issues/000210-fleet-plan-test-hardcoded-path.md` tracks it and the file is absent at base too. Everything else in `./cmd/sdlc/...` is green.
- **`TestArchitectureMarkers` was updated the right way** — appended last, in registry order, with the "do not fix this by deriving it" comment intact (`judge_test.go:352-361`). The temptation to derive the one deliberate tripwire was resisted.

## 2. Critical findings

None.

## 3. Important findings

None.

## 4. Minor findings

- `workshop/issues/000215-arch-order-architecture-principle.md:243` — the Plan step still lists "the count" as an atlas deliverable and is ticked `[x]`, though Done-when now says the opposite. **2nd finding in family `derived-fact-restated`** (details below).
- `atlas/workflow/architecture-principles.md:143` — "Adding an entry" enumerates three touch points and omits the atlas paragraph, so BR-1's rule is nowhere a future entry-adder will read it. **2nd finding in family `atlas-restates-single-source`** (details below).

## 5. Test coverage notes

Appropriate for the boundary. The three properties that could regress are each pinned by a test that would go red without the change: the marker set and its order (`TestArchitectureMarkers`), the per-entry lens contract (`TestArchitectureRegistry_Content`, derived — it covers ARCH-ORDER without being edited), and the four prompt bodies (goldens). The one prose deliverable — the atlas paragraph — is deliberately unpinned, which is correct: `TestFixTheClassLine_RoutesToArchPrinciples` is the precedent that asserting *routing* and never *wording* is what stops a doc becoming a second copy. That is why finding 2 asks for the rule in the checklist rather than a test.

## 6. Architectural notes for upcoming work

Working the lenses explicitly against this diff:

- **ARCH-DRY** — flagged (finding 2, class-level). The instance-level duplication is gone and measured gone. The residual is that the rule preventing recurrence lives only in an issue bound for `workshop/history/`, which AGENTS.md marks "don't read."
- **ARCH-PURE** — pass. One test-literal edit; no logic touched. The registry is a compile-time embedded string; every consumer reads it through one pure extraction (`markersIn` → `ArchitectureMarkers`).
- **ARCH-PURPOSE** — pass on the shadow sweep, flagged on the finding-class axis (findings 1 and 2). Enumerated every consumer: four `{{ARCH_BLOCK}}` templates, two `ArchitectureBlock` callers (`startplan.go`, `archprinciples.go`), one `{{ARCH_STAR}}` template, one hand-written list. All seven derive except the deliberate tripwire. No consumer left as a hand-maintained restatement. Both remaining findings are the same pattern the entry itself warns about: the named site fixed, the enumerable sibling left.
- **ARCH-MOCK** — N/A. No external binary or service enters the diff; the goldens are the pinned-artifact seam for prompt generation and production/test flows share it.
- **ARCH-CONSTRAINTS** — noted only, per BR-2's own scoping. ARCH-ORDER is 531 words against ARCH-PURPOSE's 371, adding ~3.3 KB to each of six registry-bearing prompt paths. The registry has no length norm while its cost multiplies across consumers; that is a registry-level question, not a defect here.
- **ARCH-SECURE** — N/A: the diff touches no untrusted input and no credential. Markdown compiled into the binary at build time, read by tests as string constants.
- **ARCH-ORDER** — N/A, written as the entry demands rather than as a bare marker: the registry holds no state between events because it is an embedded compile-time constant, and every test in the window is a deterministic string assertion with no scheduler, clock, or arrival order in play. Worth noting the one order-sensitive thing that *is* pinned: `TestArchitectureMarkers` asserts registry order positionally, and `judge_test.go:395` records that the `{{ARCH_STAR}}` literal used to be order-sensitive and passed only while entries were appended after ARCH-CONSTRAINTS — exactly the "legal vs. representable" bug the new entry describes, already caught once by #208.

For the next entry: the two Minors below plus BR-2 converge on one gap — the registry has an enforced mechanism but an unwritten *authoring* contract (length norm, atlas-paragraph shape, what counts as derived). That is one small issue, and filing it would also give BR-2's stated "worth its own issue" an artifact.

## 7. Plan revision recommendations

One `## Revisions` entry, covering finding 1:

> **BR-3 sibling site.** The Done-when bullet was corrected but Plan step 5 (`:243`) still enumerates "the count" as an atlas deliverable and is ticked `[x]`, asserting delivery of something Done-when now explicitly forbids. Corrected to `the entry, the two boundaries, and the shaping choices`. Rule, stated once so it covers the class: **a fact derived at runtime from the registry — the entry count, the marker list, `{{ARCH_STAR}}` — must never appear as a hand-maintained requirement in any artifact (Done-when, Plan step, atlas, README, helptext), only as a verification criterion about a derived output.** Enumeration swept for this issue: `grep -n "the count\|seven\|\b7\b"` over the issue file and the atlas returns one demand site (`:243`, now fixed); the remaining hits are verification assertions, which the rule permits.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      Verified by measurement, not by the commit message: 12-word-window scan of the atlas paragraph against the registry entry finds 0 shared spans post-fix vs 26 overlapping windows pre-fix; no sibling atlas paragraph restates its entry either.
  - id: BR-2
    disposition: addressed
    note: |
      Recorded in Revisions with the measurement and the reason not to trim; the finding itself scoped this as a note for the next entry. No follow-up issue was filed for the stated "worth its own issue".
  - id: BR-3
    disposition: addressed
    note: |
      Done-when bullet corrected as the finding asked. A sibling site in the same family remains at Plan step 5 and is raised below rather than re-raised here.
findings:
  - id: new
    severity: Minor
    family: derived-fact-restated
    title: |
      Plan step 5 still lists "the count" as an atlas deliverable and is ticked, contradicting the corrected Done-when
    detail: |
      This is the 2nd finding in family `derived-fact-restated`. BR-3 named the Done-when bullet and that bullet was fixed; the enumerable sibling at workshop/issues/000215-arch-order-architecture-principle.md:243 was not — "Update atlas...: the entry, the count, the ARCH-PURE boundary..." is still there and marked [x], asserting delivery of the hand-maintained restatement Done-when now forbids. Do not fix only this site: state the rule — a fact derived at runtime from the registry (entry count, marker list, ARCH_STAR) may appear as a verification criterion about a derived output, never as a hand-maintained requirement in any artifact — and sweep the enumeration. Measured prevalence for this issue: grep over the issue file and atlas returns exactly one demand site (:243); all other "seven"/"7" hits are verification assertions the rule permits. ARCH-PURPOSE.
  - id: new
    severity: Minor
    family: atlas-restates-single-source
    title: |
      The map-don't-restate rule BR-1 established has no durable home; the atlas "Adding an entry" checklist still omits the atlas paragraph entirely
    detail: |
      This is the 2nd finding in family `atlas-restates-single-source`. BR-1's instances are genuinely gone (0 shared 12-word spans, and no sibling paragraph restates), so the sweep is clean — what is missing is the rule. atlas/workflow/architecture-principles.md:143 says "a new entry touches: architecture.md, that one list, and the goldens", omitting the map-level paragraph in that same file that every entry since ARCH-MOCK has in fact required. The rule is currently recorded only in this issue's ## Revisions, which archives to workshop/history/ that AGENTS.md marks "don't read", so the next entry-adder reads a three-item checklist and repeats BR-1. Fix the class: add the fourth touch point to :143 stating the shape — this file gets a MAP-level paragraph (boundaries against confusable neighbours, deliberate shaping choices, in-fleet origin) and never a restatement of the entry's clauses, the same route-don't-restate discipline gatefindings.go already follows at :24-30. Measured prevalence: 1 restatement in 1 entry-add under the unwritten rule; 0 remaining after this round's sweep. ARCH-DRY, ARCH-PURPOSE.
```
