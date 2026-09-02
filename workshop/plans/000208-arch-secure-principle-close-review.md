# Boundary Review — ariadne#208 (whole-issue close)

| field | value |
|-------|-------|
| issue | 208 — ARCH-SECURE: a security lens in the principle registry |
| repo | ariadne |
| issue file | workshop/issues/000208-arch-secure-principle.md |
| boundary | whole-issue close |
| milestone | — |
| window | deb530a3f873b33456c9e4dd148105e549d3edae..b0f2fc72e3cbc8d7027880727c187b5a53b0aa1a |
| command | sdlc close --issue 208 |
| reviewer | claude |
| timestamp | 2026-09-02T15:50:36-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

The boundary delivers what the Spec committed to: `ARCH-SECURE` lands in `architecture.md` byte-identical to the Spec text, the non-gated `architecture-deferred.md` carries `ARCH-AUTHORITY` verbatim, `TestDeferredPrinciplesReachNoGate` derives *both* halves (forbidden set via `archMarkerRE`, gate-facing text via a `promptFS` walk + the three render helpers), the four marker-restating test sites collapse onto `ArchitectureMarkers()` with `TestArchitectureMarkers` left as the single hand-written tripwire, and the four goldens are re-captured with the header reading `6 entries`. I verified the runtime end of the Done-when directly: `./bin/sdlc arch-principles` prints "each of the 6 entries" and zero occurrences of `ARCH-AUTHORITY`; `go test ./cmd/sdlc/internal/judge/` is green and `gofmt`/`go vet` are clean. Nothing here is Critical. What holds it back from SHIP is that the guard's *own* anti-vacuity machinery — the section-vs-marker classification that the `## Log` says "three probes now pin" — has no committed test and neither of its branches is reachable in the current tree, plus two derive-don't-restate residuals of exactly the class this issue exists to close.

## 1. Strengths

- **The activation property is genuinely derived, not asserted.** `deferredMarkers` (`cmd/sdlc/internal/judge/architecturedeferred_test.go:31`) scans the deferred file rather than hardcoding `ARCH-AUTHORITY`, so moving the section into `architecture.md` empties the forbidden set and the guard stays green with no edit. That is what makes the deferred file's "moving a section is the whole activation step" claim true rather than aspirational, and it is a materially better answer than PQ-2 asked for.
- **`gateFacingTexts` walks `promptFS` instead of listing prompt names** (`architecturedeferred_test.go:108`), and renders through `BuildPrompt` — the production substitution path — so a marker leaking via *any* token, in a prompt added years from now, is caught. The `len(entries) == 0` bail (`:117`) closes the "walked to nothing" hole. This is the load-bearing design decision in the diff and it's right.
- **`TestArchitectureRegistry_Content` traded a restated list for a real assertion** (`judge_test.go:129-145`): every declared marker must have a `## <marker>` heading and all three lens bullets. Nothing checked entry *shape* before; the derived version covers strictly more than the literal it replaced.
- **The `{{ARCH_STAR}}` literal was correctly diagnosed as order-sensitive** and replaced with `strings.Join(ArchitectureMarkers(), ", ")` (`judge_test.go:395`), with the order-pinning left where it belongs — in the one tripwire.
- **The tripwire's comment says *why* it is hand-written** (`judge_test.go:353-360`), which is the thing that stops a future agent from "fixing" it into a tautology.

## 2. Critical findings

None.

## 3. Important findings

**I1 — `architecturedeferred_test.go:71-82`: the anti-vacuity classification is untestable and untested; the `## Log`'s "three probes now pin it" is not backed by a committed test.** (ARCH-PURE)

With `ARCH-AUTHORITY` present, `deferred` is length 1, so the `switch` is skipped entirely and `sections` is used only inside a message string — both branches are dead in the committed tree. The `## Log` states "Three probes now pin it — activate → green, break a heading → red, leak `ARCH-AUTHORITY` into a prompt → red", but the file contains one test and two helpers; those probes were manual and left no artifact. The reason they *can't* be committed is structural: `deferredMarkers` fuses the embed read, the parse, and `t.Fatalf` into one helper, so classification can only be exercised by editing the real `architecture-deferred.md`.

Fix sketch: split the pure part out —
```go
func parseDeferred(text string) (markers []string, sections int)
func classifyDeferred(markers []string, sections int) (skip bool, fatal string)
```
— and table-test `classifyDeferred` over the three states (markers+sections → check; sections, no markers → fatal; neither → skip), plus `parseDeferred` over a fixture with a broken `### `/`##`-without-space heading. That makes the discriminator the Log describes a checked property rather than a described one, and it is the test that fails without the fix.

**I2 — `architecturedeferred_test.go:37-45` duplicates `ArchitectureMarkers()`'s extraction loop (`architecture.go:19-29`).** (ARCH-DRY)

The dedupe-preserving-order scan over `archMarkerRE.FindAllStringSubmatch` is copy-pasted with only the input string changed. This is not cosmetic: the guard's correctness depends on the forbidden set being extracted *the same way* the gated set is. If `ArchitectureMarkers` ever normalizes differently (trailing punctuation, case, a marker-name change), the two diverge and the guard goes silently weaker — a false green, which is the exact failure mode the rest of this issue is defending against.

Fix sketch: extract `func markersIn(text string) []string` in `architecture.go`; `ArchitectureMarkers()` becomes `markersIn(ArchitectureRegistry)` and `deferredMarkers` calls `markersIn(string(b))`.

**I3 — `atlas/workflow/sdlc-binary.md:880-895` was not updated and now contradicts the atlas page this diff *did* write.** (ARCH-PURPOSE, Docs update gate)

That paragraph is the per-entry narrative — it has a line for `#71 ARCH-SHIM`, `#126 ARCH-PURPOSE`, `#205 ARCH-CONSTRAINTS` — and has no line for `#208 ARCH-SECURE`, no mention of `architecture-deferred.md` or the documented-but-not-gated concept (genuinely new architectural surface), and it still asserts at line 893 that "Adding an `ARCH-*` entry **flows into every consumer with no other edit**." The new `atlas/workflow/architecture-principles.md` "Adding an entry" section says the opposite and more accurate thing: registry + the one tripwire + the goldens. A reader who follows `sdlc-binary.md` edits only `architecture.md` and gets a red suite.

Fix sketch: one sentence for `#208` in the same style as `#205`, a clause pointing at the deferred file, and correct the "no other edit" claim to route to `architecture-principles.md#adding-an-entry`.

## 4. Minor findings

- `architecturedeferred_test.go:76-81` — the comment claims the only real disarm vector is a broken heading, but an **emptied/truncated** file (bad merge, botched sed) yields `sections == 0, markers == 0` → silent `t.Skip`, a green suite covering nothing. Deletion/rename is a build error as claimed; truncation is not. Cheapest tightening: require the file's title line before allowing the skip path. Relatedly, the issue's Done-when still reads "That guard also asserts `architecture-deferred.md` parses **at least one** marker" — the implementation deliberately replaced that with a classified skip, and the `## Log` explains why, but the Done-when was never corrected (see §7).
- `judge_test.go:133` — `strings.Index(ArchitectureRegistry, "## "+m)` is prefix-fragile: `"## ARCH-PURE"` matches inside `"## ARCH-PURPOSE"`. Correct today only because `ARCH-PURE` precedes `ARCH-PURPOSE` in the file; a reorder would silently lens-check the wrong entry. Match `"## "+m+" "` or `"\n## "+m+" —"`.
- `cmd/sdlc/startplan_test.go:36` — a sixth hand-written marker mention (`"ARCH-DRY", "ARCH-CONSTRAINTS"`) that the Spec's five-site enumeration missed. Benign, because line 41's `strings.Contains(out, judge.ArchitectureRegistry)` already covers new markers derivedly — but it's a residual member of the swept class and the two literals now buy nothing.
- `atlas/workflow/architecture-principles.md` / issue `## Estimate` — the declared envelope is "roughly **+1.4 KB** per dispatch"; measured is **+1,991 bytes** (`architecture.md` 7,519 → 9,510), ~39% over, across four prompts plus `start-plan`. The tradeoff is still clearly worth it; the number should be replaced with the measured one so the ledger records a fact rather than a pre-implementation estimate.
- **Pre-existing, out of window:** `cmd/sdlc/fleet_plan_test.go:14` fails — `workshop/plans/000200-sdlc-fleet-thread-inventory-plan.md` was archived to `workshop/history/plans/` in `dfeba9c`, which is an ancestor of the base. `go test ./cmd/sdlc/...` is therefore RED at this boundary through no fault of this diff. Not this issue's to fix, but the close evidence should be scoped to the packages this diff touches rather than claiming a green suite — and the test itself should resolve the plan through `workshop/plans/` *or* `workshop/history/plans/`.

## 5. Test coverage notes

Coverage of the *shipped surface* is good and mostly derived: registry containment in all four architecture-aware prompts (`TestArchitectureRegistry_EmbeddedInPrompts`) covers "`ARCH-SECURE` reaches plan-quality and boundary-review" without naming it; the goldens pin the bytes and the `6 entries` header; `TestArchitectureMarkers` pins the set and its order. `TestArchitecture_NarrativeRoutesToArchPrinciples` was correctly left untouched and is green, satisfying that Done-when bullet the way PQ-3 asked.

The gap is entirely in the *guard's own* failure modes (I1): a leak into a prompt is caught by construction, but "the guard stopped guarding" is not exercised at all. That's the bug class this test exists to prevent, and it is the one class the test suite would ship silently.

## 6. Architectural notes for upcoming work

- The two-file split (`architecture.md` gated / `architecture-deferred.md` not) is the right call over a `gates:` field for one entry, and the Spec's reasoning holds. Watch the threshold: at three or four deferred entries, "a second file the build must not embed" starts to want entry-level parsing anyway, and at that point the marker-extraction helper from I2 is the natural place for it.
- `gateFacingTexts` is quietly the most reusable thing in this diff — "every string this package can put in front of a judge, derived." Future guards (no raw secrets in prompts, no absolute host paths, prompt size ceilings) should consume it rather than re-enumerate. Worth promoting out of `architecturedeferred_test.go` into a shared test helper the first time a second caller appears.
- The measurement in the `## Log` (4,775 ARCH citations, 71% in generated ledgers, 142 in code comments) is the real success metric for this entry. `ARCH-SECURE` will be judged by whether it moves code-comment citations, not by whether it appears in review sidecars — worth re-running that count in ~30 days rather than assuming the entry landed.

## 7. Plan revision recommendations

One `## Revisions` entry on `workshop/issues/000208-arch-secure-principle.md`, because the Done-when still claims behavior the code deliberately does not implement:

> ### 2026-09-02 — the anti-vacuity assertion became a classified skip
>
> Done-when read "That guard also asserts `architecture-deferred.md` parses **at least one** marker, so renaming or deleting the file cannot make it vacuously pass" (PQ-2's second half). Implementation showed an unconditional assert makes activating the LAST deferred entry fail, falsifying the deferred file's own "activation is just a move" claim. Shipped instead: markers and `## ` sections are counted independently — sections without markers fails (headings stopped parsing), neither fails nor passes but skips (everything activated). Rename/delete remains impossible to sneak past because `//go:embed` on a missing path is a build error; a *truncated* file is the residual hole. Done-when amended to state the classification rather than the assert.

Optionally fold the measured registry delta (+1,991 B, vs the estimated ~1.4 KB) into the same entry so the `### Operating envelope` section carries a measured basis at close.

```findings
findings:
  - id: new
    severity: Important
    family: unexercised-guard-branch
    title: |
      Deferred-guard vacuity classification has no test and neither branch is reachable in the committed tree
    detail: |
      architecturedeferred_test.go:71-82 — with ARCH-AUTHORITY present the switch is
      never entered, so both the fatal (sections>0, no markers) and skip (neither)
      branches are dead, and `sections` is used only inside a message string. The
      issue's Log claims "Three probes now pin it — activate/break a heading/leak",
      but the file ships one test and two helpers; those probes were manual and left
      no artifact. deferredMarkers fuses the embed read, the parse and t.Fatalf, so
      classification cannot be exercised without editing the real file (ARCH-PURE).
      Split out pure parseDeferred(text) and classifyDeferred(markers, sections) and
      table-test the three states plus a broken-heading fixture.
  - id: new
    severity: Important
    family: derive-not-restate
    title: |
      Marker extraction is copy-pasted between ArchitectureMarkers and the guard's deferredMarkers
    detail: |
      architecturedeferred_test.go:37-45 duplicates architecture.go:19-29 —
      the same order-preserving dedupe over archMarkerRE.FindAllStringSubmatch with
      only the input string changed (ARCH-DRY). The guard's correctness depends on the
      forbidden set being extracted identically to the gated set; divergence yields a
      silent false green, the exact failure this issue defends against. Extract
      markersIn(text string) []string in architecture.go and have both call it.
  - id: new
    severity: Important
    family: doc-enumeration-drift
    title: |
      atlas/workflow/sdlc-binary.md not updated and now contradicts the atlas page this diff wrote
    detail: |
      sdlc-binary.md:880-895 carries the per-entry narrative (lines for #71, #126,
      #205) with no line for #208 and no mention of architecture-deferred.md or the
      documented-but-not-gated concept — new architectural surface. Worse, line 893
      still asserts "Adding an ARCH-* entry flows into every consumer with no other
      edit", which the new architecture-principles.md "Adding an entry" section
      directly contradicts (registry + the one tripwire + the goldens). A reader
      following sdlc-binary.md edits only architecture.md and gets a red suite.
  - id: new
    severity: Minor
    family: unbacked-existing-behavior
    title: |
      Done-when claims an unconditional at-least-one-marker assert; the code ships a classified skip, and a truncated file skips silently
    detail: |
      The implementation deliberately replaced PQ-2's unconditional assert with a
      sections-vs-markers classification (the Log explains why: an unconditional
      assert reds on activating the last entry). The Done-when bullet was never
      amended, so the issue still claims behavior the code does not have. Separately,
      architecturedeferred_test.go:76-81 comments that broken headings are the only
      real disarm vector — an emptied or truncated file also yields a silent t.Skip.
      Require the file's title line before allowing the skip path.
  - id: new
    severity: Minor
    family: prefix-match-fragility
    title: |
      Entry-section lookup matches ARCH-PURE inside ARCH-PURPOSE
    detail: |
      judge_test.go:133 — strings.Index(ArchitectureRegistry, "## "+m) is correct only
      because ARCH-PURE precedes ARCH-PURPOSE in the file. A reorder would silently
      lens-check the wrong entry. Match "## "+m+" " (or the em-dash form) instead.
  - id: new
    severity: Minor
    family: marker-enumeration-sites
    title: |
      startplan_test.go holds a sixth hand-written marker list the class sweep missed
    detail: |
      cmd/sdlc/startplan_test.go:36 spot-checks "ARCH-DRY", "ARCH-CONSTRAINTS" by
      hand. Benign — line 41's Contains(out, judge.ArchitectureRegistry) already
      covers new markers derivedly, which is what satisfies the "start-plan delivers
      it" Done-when — but it is a residual member of the swept class and the two
      literals now assert nothing the registry check doesn't.
  - id: new
    severity: Minor
    family: operating-envelope-unstated
    title: |
      Declared prompt-growth budget is 1.4 KB; measured is 1,991 bytes
    detail: |
      architecture.md grew 7,519 -> 9,510 bytes, ~39% over the Spec's stated
      "roughly +1.4 KB per dispatch", across four prompts plus start-plan
      (ARCH-CONSTRAINTS at-review: implementation vs declared envelope). The tradeoff
      remains clearly worth it; replace the estimate with the measured value in the
      issue's Operating envelope section and atlas so the ledger records a fact.
  - id: new
    severity: Minor
    family: stale-path-in-test
    title: |
      Pre-existing red test blocks any "full suite green" close evidence (out of window)
    detail: |
      cmd/sdlc/fleet_plan_test.go:14 fails — it reads
      workshop/plans/000200-sdlc-fleet-thread-inventory-plan.md, archived to
      workshop/history/plans/ in dfeba9c, an ancestor of this window's base. Not
      caused by this diff. Scope the close's verification evidence to the packages
      touched (judge + cmd/sdlc archprinciples/startplan are green), and have that
      test resolve the plan through workshop/plans/ or workshop/history/plans/.
```
