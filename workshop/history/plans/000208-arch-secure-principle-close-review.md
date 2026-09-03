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

---

## Re-review — 2026-09-02T16:02:59-07:00 (FIX-THEN-SHIP)

| field | value |
|-------|-------|
| issue | 208 — ARCH-SECURE: a security lens in the principle registry |
| repo | ariadne |
| issue file | workshop/issues/000208-arch-secure-principle.md |
| boundary | whole-issue close |
| milestone | — |
| window | deb530a3f873b33456c9e4dd148105e549d3edae..34f0f59c7a43f02ef5d9efdf63cfb47d9c173a45 |
| command | sdlc close --issue 208 |
| reviewer | claude |
| timestamp | 2026-09-02T16:02:59-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

The boundary delivers what the Spec committed to, and the round‑1 Important findings are genuinely fixed — I verified each by reverting in a scratch copy rather than reading the commit message. `parseDeferred`/`Verdict()` is a real pure split with a table test that goes red when I disable section counting; `markersIn` is now the single extraction site for both the gated and the deferred sets; `sdlc-binary.md`'s contradicting "flows into every consumer with no other edit" sentence is corrected. What holds it back from SHIP is that the BR‑1 fix reintroduced, through a new sibling test, the exact regression the round‑1 implementation had identified and designed around: `TestDeferredFileIsGuarding` duplicates the `deferredBroken` `t.Fatalf` that already exists, and its *only* unique effect is to fail on `deferredRetired` — the state the classification exists to bless. Activating the (single) deferred entry now reds the suite, falsifying six prose sites that say activation is a pure move. It is cheap to fix. The four remaining Minors from round 1 are unaddressed and re-disposed below.

## 1. Strengths

- **The guard's derivation is real, not decorative.** I probed it three ways: appending `ARCH-AUTHORITY` to `prompts/specs.md` reds `TestDeferredPrinciplesReachNoGate` with the right message (`architecturedeferred_test.go:160`), breaking the deferred heading to `ARCH_AUTHORITY` reds the vacuity branch, and the `promptFS` walk plus the zero-entries `t.Fatal` means a prompt added next year is covered without anyone remembering #208. This is the load-bearing part of the issue and it works.
- **BR‑1's fix is a genuine ARCH‑PURE split, verified by revert.** `parseDeferred(content) → deferredState` takes content, not a path (`architecturedeferred_test.go:41`); replacing `d.Sections++` with a no-op reds two named subtests. The three-state table covers the cases the committed tree can't exhibit.
- **BR‑2's `markersIn` extraction is the right consolidation** (`architecture.go:20-37`). One scan-and-dedupe rule computing both the gated and the not-gated sets is precisely what a disjointness guard needs, and the comment says why (ARCH‑DRY).
- **`TestArchitectureRegistry_Content` gained real content in exchange for its restated list** (`judge_test.go:130-146`) — every declared marker must carry all three lens bullets, which nothing checked before. That is a derived test that asserts *more*, not less.
- **`TestArchitectureMarkers`'s comment explains why it is deliberately hand-written** (`judge_test.go:353-360`), which is what stops a future agent from "fixing" the tripwire into vacuity.

## 2. Critical findings

None.

## 3. Important findings

### I‑1 — `TestDeferredFileIsGuarding` re-breaks the activation contract; it is redundant except in the one state it should allow

`cmd/sdlc/internal/judge/architecturedeferred_test.go:111-117`

**This is the 2nd finding in family `unbacked-existing-behavior`.** Per the escalation protocol I am not asking for this instance to be patched — here is the rule and the enumeration.

Measured behavior (scratch-copy probes against `34f0f59`):

| probe | `TestDeferredPrinciplesReachNoGate` | `TestDeferredFileIsGuarding` |
|---|---|---|
| heading broken (`ARCH_AUTHORITY`) | FAIL (`t.Fatalf`, line 136) | FAIL |
| file emptied / last entry activated | pass (`t.Skip`, line 143) | **FAIL** |

So `TestDeferredFileIsGuarding` adds nothing on `deferredBroken` — that state was already fatal inside the guard — and its sole marginal effect is to red the suite on `deferredRetired`, which the classification was built to distinguish *as legitimate*. The `## Log` documents the implementor discovering this exact failure in round 1 ("I moved ARCH-AUTHORITY into `architecture.md` and the guard went red, falsifying the very claim the deferred file makes"); the BR‑1 fix reintroduced it.

**The rule (this is what to fix, not the one test):** *the classifier owns the pass/fail/skip action for each of its states, at one site; no second consumer may independently assign an outcome to a state, and no prose may restate the mapping.* Concretely: give `deferredVerdict` the action (`broken → fatal`, `retired → skip`, `guard → enforce`) in one helper both tests call, then delete or narrow `TestDeferredFileIsGuarding` to `Verdict() != deferredBroken`.

The class this rule covers — six sites asserting "activation is a pure move", none deriving from the code, now all false at suite level:

1. `cmd/sdlc/internal/judge/architecture-deferred.md:6` — "that is the whole activation step" (ships to the future operator).
2. `architecturedeferred_test.go:30` — "moving the section empties it and the guard stays green with no other edit".
3. `architecturedeferred_test.go:141-142` — "activation stayed a pure MOVE, which is exactly what the deferred file promises".
4. `atlas/workflow/architecture-principles.md:67` and `:84` — "Activation is a MOVE and nothing else"; "(no sections, no markers — skips)".
5. `atlas/workflow/sdlc-binary.md:909` — "the move empties the set and the guard stays green with no edit to it".
6. The issue's Done‑when bullet 3 — which is in direct tension with Done‑when bullet 4 ("asserts … **at least one** marker"). With exactly one deferred entry, both cannot hold. That tension, not this test, is the root: the issue asked for two incompatible things and each round has satisfied a different one.

Note also that `atlas/workflow/architecture-principles.md` was written in `b0f2fc7` and not touched by the fix commit, so it now describes a guard that no longer behaves as documented and omits `TestDeferredFileIsGuarding` and `markersIn` entirely (ARCH‑PURPOSE at‑review: the doc consumer did not derive).

Secondary, same finding: `architecturedeferred_test.go:20-29` is an orphaned doc comment for `deferredMarkers`, a function the fix deleted. It is fused to `deferredState`'s comment with no blank line, so godoc renders "deferredMarkers returns…" as the doc for `deferredState`.

## 4. Minor findings

- `architecturedeferred_test.go:158` matches deferred markers with `strings.Contains(g.text, m)` — substring, not delimiter-anchored; a deferred marker that is a prefix of a gated one yields a confusing false RED that the exact-compare disjointness loop above it would not explain. Same rule as BR‑5 (see disposition).
- `gateFacingTexts` walks `prompts/` non-recursively and passes every entry name to `BuildPrompt`, which `panic`s on a missing template — a future `prompts/<subdir>/` would panic rather than fail.
- The guard lives in package `judge`, so `cmd/sdlc`'s own gate-facing prose (`startplan.go`'s surrounding text, help output) is structurally outside its reach. Matches the Done‑when's declared scope; noted only so the boundary is a choice rather than an assumption.

## 5. Test coverage notes

- `go test ./cmd/sdlc/internal/judge/` — green. `gofmt -l` and `go vet ./cmd/sdlc/...` — clean.
- `go test ./...` — one failure, `TestFleetPlanHasAuthoritativeCorrectedCoreConceptInventory` (`cmd/sdlc/fleet_plan_test.go:14`), pre-existing and out of window (BR‑8, still red).
- Done‑when verified directly: `go run ./cmd/sdlc arch-principles` prints "each of the **6** entries" and zero `ARCH-AUTHORITY`; all four goldens read `6 entries`; `milestone-review.prompt:112` carries the derived `{{ARCH_STAR}}` full set ending in `ARCH-SECURE`.
- Coverage gap remaining: nothing pins that `deferredRetired` is a *green* terminal state end-to-end — which is the finding above, not an additional one.

## 6. Architectural notes for upcoming work

- **ARCH‑DRY** — pass on the production change (`markersIn` is now the single extraction site) with one residual: `TestDeferredFileIsGuarding` duplicates the `deferredBroken` assertion (I‑1).
- **ARCH‑PURE** — pass. `parseDeferred` is content-in/value-out; `deferredFileState` is the thin IO seam; the table test runs with no IO.
- **ARCH‑PURPOSE** — flag (I‑1). The shadow-sweep finds the runtime consumers all deriving, but the *doc* consumers of the activation contract are six hand-maintained restatements, one of which the fix commit left stale. The marker-set sweep also stopped one member short (BR‑6, `startplan_test.go:36`).
- **ARCH‑MOCK** — N/A. No external binary or service crosses this diff; the embedded files are compile-time inputs.
- **ARCH‑CONSTRAINTS** — flag (BR‑7). Declared `+1.4 KB per dispatch`; measured `architecture.md` 7,519 → 9,510 = **1,991 bytes**, ~39% over, across four prompts plus `start-plan`. The tradeoff still reads clearly worth it; the number should be the measured one.
- **ARCH‑SECURE** — pass, and worth noting the recursion: applying the new lens to its own diff, all parsed input is compile-time-embedded (trusted provenance), `parseDeferred` handles empty/truncated content without panicking, the failure path degrades *visibly* with an actionable message rather than substituting a fabricated value, and no credential surface is touched.
- For the next entry: the atlas now promises "registry + one tripwire + goldens". Until I‑1 is resolved, activating a *deferred* entry additionally requires editing a test — write that down or make it untrue.

## 7. Plan revision recommendations

The issue needs a `## Revisions` entry reconciling the two incompatible Done‑when bullets:

> **### 2026‑09‑02 — Done‑when bullets 3 and 4 cannot both hold**
> Bullet 3 requires activation to leave the guard green; bullet 4 requires an unconditional ≥1‑marker assertion. With one deferred entry, activating it empties the file and bullet 4 reds. Round 1 satisfied bullet 3 (section-count classification, retired → skip); the BR‑1 fix satisfied bullet 4 (`TestDeferredFileIsGuarding`) and re-broke bullet 3. Resolution: bullet 3 wins — the anti-vacuity guarantee is delivered by `deferredBroken` (a parse break is the only real disarm vector; deleting the file is a build error via `//go:embed`). Amend bullet 4 to "the guard fails when the file has sections but parses no markers", and state that the state→outcome mapping lives once, on `deferredVerdict`.

Also amend the `### Operating envelope` section: replace "roughly **+1.4 KB per dispatch**" with the measured **1,991 bytes** (BR‑7), and mirror it into `atlas/workflow/architecture-principles.md`.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      Verified by revert — disabling section counting reds two named subtests; breaking the committed heading reds both the classification and the guard.
  - id: BR-2
    disposition: addressed
    note: |
      markersIn is now the single extraction site; ArchitectureMarkers() and the deferred scan both call it.
  - id: BR-3
    disposition: addressed
    note: |
      sdlc-binary.md now separates runtime consumers (derive) from the test layer, and documents the deferred file.
  - id: BR-4
    disposition: addressed
    note: |
      TestDeferredFileIsGuarding closes the truncated-file skip vector and makes the Done-when claim true — at the cost of the incompatible sibling bullet, raised as a new finding rather than re-raised here.
  - id: BR-5
    disposition: not-addressed
    note: |
      judge_test.go:133 still uses strings.Index(ArchitectureRegistry, "## "+m); the class now has a second member at architecturedeferred_test.go:158 (strings.Contains for marker leak) — the rule is that every marker match must be delimiter-anchored.
  - id: BR-6
    disposition: not-addressed
    note: |
      cmd/sdlc/startplan_test.go:36 and :47 still hand-write ARCH-DRY / ARCH-CONSTRAINTS / ARCH-PURE; the file is untouched in this window.
  - id: BR-7
    disposition: not-addressed
    note: |
      Re-measured at head — architecture.md 7,519 to 9,510 bytes = 1,991; the issue still states "roughly +1.4 KB per dispatch".
  - id: BR-8
    disposition: not-addressed
    note: |
      go test ./... still fails at cmd/sdlc/fleet_plan_test.go:14 on the archived plan path; pre-existing and out of window.
findings:
  - id: new
    severity: Important
    family: unbacked-existing-behavior
    title: |
      TestDeferredFileIsGuarding duplicates the deferredBroken fatal and its only unique effect is to red the suite on deferredRetired, re-breaking the activation contract
    detail: |
      Probed against 34f0f59 in a scratch copy: breaking the deferred heading fails BOTH
      TestDeferredPrinciplesReachNoGate (t.Fatalf, architecturedeferred_test.go:136) and
      TestDeferredFileIsGuarding; emptying the file (equivalently, activating the only
      deferred entry) fails ONLY TestDeferredFileIsGuarding. So the new test adds nothing
      on the state that matters and vetoes the state the section-count classification was
      built to bless. The Log records the implementor finding and removing exactly this
      failure in round 1; the BR-1 fix reintroduced it.
      This is the 2nd finding in family unbacked-existing-behavior, so the deliverable is
      the rule, not this test. RULE — the classifier owns the pass/fail/skip action for
      each of its states at ONE site; no second consumer assigns an outcome to a state,
      and no prose restates the mapping. Enumeration of the six sites currently asserting
      the now-false "activation is a pure move": architecture-deferred.md:6;
      architecturedeferred_test.go:30; architecturedeferred_test.go:141-142;
      atlas/workflow/architecture-principles.md:67 and :84; atlas/workflow/sdlc-binary.md:909;
      the issue's Done-when bullet 3 (which contradicts its own bullet 4 — that tension is
      the root). Fix: move the action onto deferredVerdict, narrow or delete
      TestDeferredFileIsGuarding, and rewrite the six sites to point at the one mapping.
      Same finding, secondary: architecturedeferred_test.go:20-29 is an orphaned doc comment
      for the deleted deferredMarkers, fused without a blank line to deferredState's comment,
      so godoc renders the wrong doc for that type. atlas/workflow/architecture-principles.md
      was not touched by the fix commit and now omits both markersIn and TestDeferredFileIsGuarding.
```

---

## Re-review — 2026-09-02T16:14:23-07:00 (FIX-THEN-SHIP)

| field | value |
|-------|-------|
| issue | 208 — ARCH-SECURE: a security lens in the principle registry |
| repo | ariadne |
| issue file | workshop/issues/000208-arch-secure-principle.md |
| boundary | whole-issue close |
| milestone | — |
| window | deb530a3f873b33456c9e4dd148105e549d3edae..5903b5d14ff2828f1cf586833a8481b8aa3be3e2 |
| command | sdlc close --issue 208 |
| reviewer | claude |
| timestamp | 2026-09-02T16:14:23-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

BR-9 is genuinely fixed, and I verified it by probe rather than by reading the Log: in a scratch copy at `5903b5d` I moved the `## ARCH-AUTHORITY` section into `architecture.md` and the deferred guard **skipped green**, with exactly two failures — `TestArchitectureMarkers` (the deliberate tripwire) and `TestBuildPrompt_Golden` (re-capture) — which is precisely what the atlas documents activation as touching. (A third failure, `TestArchitecture_NarrativeRoutesToArchPrinciples`, turned out to be a scratch-copy artifact: `AGENTS.md` is generated and untracked, so `git archive` doesn't carry it; with it copied in, the count is two.) I separately probed that breaking the deferred heading reds one test with an actionable message, and that leaking `ARCH-AUTHORITY` into `prompts/estimate-quality.md` reds the guard by name — so the guard is reachable and doing work, not decorative. Both registry texts are byte-identical to the Spec's verbatim blocks (diffed programmatically), the four goldens read `6 entries`, `{{ARCH_STAR}}` derives, and every Done-when bullet traces. What holds it back from SHIP is only cheap paperwork: four Minors carried unaddressed across two rounds plus one new one, including a factual number in the issue's own Operating envelope and a doc sentence in the round-2 fix that contradicts itself.

## 1. Strengths

- **The activation contract now actually holds, and I measured it rather than trusting the Log.** `architecturedeferred_test.go:148` skips on `deferredRetired`; the probe confirms activating the only deferred entry leaves the guard green. This is the property two prior rounds each broke in a different direction, and it is now the one the code delivers.
- **The `deferredBroken` failure path degrades visibly with an actionable message** (`architecturedeferred_test.go:143`): it names the section count, echoes `archMarkerRE`, and says "fix the heading, don't delete the check." That is the ARCH-SECURE at-review clause applied to the guard's own parse failure.
- **`markersIn` is the single extraction site** (`architecture.go:20-37`), and it is what makes the disjointness guard sound — the gated and the not-gated sets cannot diverge by rule. Verified: `grep` finds exactly two hand-written multi-marker literals left in the tree (`judge_test.go:363`, the deliberate tripwire, and `startplan_test.go:36`, which is BR-6).
- **`gateFacingTexts` walks `promptFS` rather than listing prompts** (`architecturedeferred_test.go:196`), covering `estimate-quality`/`plan`/`specs` for free, with a zero-entries `t.Fatal` so an empty walk can't pass vacuously. The `//go:embed prompts/*.md` pattern is non-recursive, so the "subdir panics `BuildPrompt`" worry raised last round can't actually arise.
- **`architecture-deferred.md` is embedded only from a `_test.go` file** — verified by grep across `cmd/`, so it has no production reachability at all, not merely no textual leak.

## 2. Critical findings

None.

## 3. Important findings

None. BR-9 is disposed `addressed` on probe evidence above.

## 4. Minor findings

- `atlas/workflow/architecture-principles.md:85-94` — the page says the mapping "is documented there and nowhere else, so this page describes it rather than restating the rule", then restates the full rule in a three-row table; `architecturedeferred_test.go:55-56` likewise claims "documented nowhere else." Both sentences are false as written. See the new finding below for the rule.
- `judge_test.go:124-129` and `archprinciples_test.go:23` — the marker portion of these derived loops cannot fail: the markers came from scanning `ArchitectureRegistry`, and the loop then asserts they're contained in it (or in output an adjacent, stronger assertion already pins whole).
- Carried, unchanged: BR-5 (`judge_test.go:133` `"## "+m` prefix match), BR-6 (`startplan_test.go:36`), BR-7 (declared 1.4 KB vs measured 1,991 — re-confirmed: `architecture.md` 7,519 → 9,510), BR-8 (`fleet_plan_test.go:14` still red on the archived plan path).

## 5. Test coverage notes

- `go test ./cmd/sdlc/internal/judge/` green; `go vet ./cmd/sdlc/...` and `gofmt -l cmd/sdlc/` clean.
- `go test ./...` — one failure, `TestFleetPlanHasAuthoritativeCorrectedCoreConceptInventory` (BR-8), pre-existing and out of window. Scope the close's `--verified` evidence to the touched packages.
- Three behavioral probes done in scratch copies (activate / break heading / leak marker into a prompt) — all three land where the design says they should.
- Coverage boundary worth naming: the verdict→action mapping in the guard's `switch` (`architecturedeferred_test.go:140-149`) is itself untested — `TestDeferredVerdict` pins the classification, not the action. That is inherent (a test asserting a test's own skip/fatal), and with the second consumer deleted there is exactly one site, which is what BR-9's rule asked for.

## 6. Architectural notes

- **ARCH-DRY** — pass. `markersIn` consolidated; no duplicated extraction remains. The one residual is documentary (the atlas table), below.
- **ARCH-PURE** — pass. `parseDeferred(content) → deferredState` is content-in/value-out; `deferredFileState` is the thin IO seam; the table test runs with no IO.
- **ARCH-PURPOSE** — pass on the code. Shadow-sweep: every runtime consumer derives, and the test layer now derives except the one deliberate tripwire; the only unswept member is `startplan_test.go` (BR-6, already open). Flag on docs only — the atlas restated the mapping the fix commit had just single-sourced.
- **ARCH-MOCK** — N/A. Both parsed files are compile-time `//go:embed` inputs; no external binary or service crosses this diff.
- **ARCH-CONSTRAINTS** — flag (BR-7, carried). The implementation is within budget in kind but the declared number is 39% low; ARCH-SECURE's own at-review clause names "unsupported performance claims", so the ledger should carry the measured 1,991 bytes.
- **ARCH-SECURE** — pass, applying the new lens to its own diff: all parsed input has compile-time provenance, `parseDeferred("")` returns `retired` without panicking, the truncation path fails visibly rather than substituting a fabricated marker set, and no credential surface is touched.

## 7. Plan revision recommendations

None outstanding. The round-2 fix already rewrote Done-when bullet 4 to resolve the bullet-3/bullet-4 tension the prior round flagged, and logged the change under `## Log`. One optional tidy: that Done-when edit was an overwrite rather than a `## Revisions` append (AGENTS.md §1), so a two-line Revisions entry dated 2026-09-02 recording "bullet 4 replaced: unconditional ≥1-marker assert → classified `retired` skip; reason: the unconditional form reds the suite on activation" would keep the artifact's amendment history intact. Fold BR-7's measured number into the `### Operating envelope` section in the same edit.

```findings
dispose:
  - id: BR-9
    disposition: addressed
    note: |
      Verified by probe at 5903b5d, not by the Log — activating ARCH-AUTHORITY leaves TestDeferredPrinciplesReachNoGate SKIPped green and fails exactly TestArchitectureMarkers + TestBuildPrompt_Golden; breaking the heading reds one test with an actionable message; the orphaned deferredMarkers doc comment is folded and the atlas gained markersIn and the verdict table.
  - id: BR-5
    disposition: not-addressed
    note: |
      judge_test.go:133 still uses strings.Index(ArchitectureRegistry, "## "+m); architecturedeferred_test.go:165 still uses strings.Contains for the leak check. Neither is delimiter-anchored.
  - id: BR-6
    disposition: not-addressed
    note: |
      cmd/sdlc/startplan_test.go:36 still hand-writes ARCH-DRY / ARCH-CONSTRAINTS and :47 ARCH-PURE; the file is untouched across the whole window.
  - id: BR-7
    disposition: not-addressed
    note: |
      Re-measured at head — architecture.md 7,519 to 9,510 bytes = 1,991; the issue's Operating envelope still reads "roughly +1.4 KB per dispatch".
  - id: BR-8
    disposition: not-addressed
    note: |
      go test ./... still fails only at cmd/sdlc/fleet_plan_test.go:14 on the archived plan path; pre-existing, out of window, and the reason close evidence must be package-scoped.
findings:
  - id: new
    severity: Minor
    family: doc-enumeration-drift
    title: |
      The fix that single-sourced the verdict mapping restates it in the atlas, and both sites claim it is documented nowhere else
    detail: |
      atlas/workflow/architecture-principles.md:85-88 states the mapping "is documented there
      and nowhere else, so this page describes it rather than restating the rule" — and lines
      90-94 then restate the full rule as a three-row table. architecturedeferred_test.go:55-56
      makes the same false claim from the Go side ("enforced in exactly one place ... and
      documented nowhere else"). This is the 2nd finding in family doc-enumeration-drift (BR-3
      was atlas/workflow/sdlc-binary.md contradicting the page this diff wrote), so per the
      escalation protocol the deliverable is the rule, not this table. RULE — a doc may name a
      code-owned mapping and point at the owning symbol, or it may reproduce the mapping and
      say so, but it may not claim single-sourcing while reproducing. Prevalence in this issue:
      2 of 2 atlas pages touched. Cheapest resolution: keep the table (it is genuinely useful
      map content) and delete the two "nowhere else" claims, replacing them with "the owning
      symbol is deferredVerdict; this table mirrors it."
  - id: new
    severity: Minor
    family: tautological-derivation
    title: |
      Two derived marker assertions are checked against the source they were derived from, so they cannot fail
    detail: |
      judge_test.go:124-129 builds `want` from ArchitectureMarkers() and asserts
      strings.Contains(ArchitectureRegistry, w) — but markersIn extracted those exact
      substrings from ArchitectureRegistry, so that branch is unreachable by construction.
      archprinciples_test.go:23 has the same shape, and is additionally subsumed by the
      Contains(out, judge.ArchitectureRegistry) assertion three lines below it. The non-marker
      entries in both loops ("at-plan", "at-review", "principle:") are real and should stay.
      RULE — derive an expectation from the source, then assert it against a CONSUMER; an
      expectation asserted against its own source is a no-op. Adjacent to BR-6's observation
      that the startplan_test literals "now assert nothing the registry check doesn't".
```
