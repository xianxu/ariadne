# Boundary Review — ariadne#194 (whole-issue close)

| field | value |
|-------|-------|
| issue | 194 — boundary reviews: anchor to the reviewed commit, and remember across rounds |
| repo | ariadne |
| issue file | workshop/issues/000194-review-anchor-commit.md |
| boundary | whole-issue close |
| milestone | — |
| window | 33343455e21e25f056f41ab066b5face9e70f745..20449bd73ddc3ac471ce6595bb5794227c8e029f |
| command | sdlc close --issue 194 |
| reviewer | claude |
| timestamp | 2026-08-20T23:16:29-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

I have everything I need. Here is the review.

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

The issue delivers all nine Done-when items, and — unusually — it produced its own acceptance evidence: M3 took four rounds and the family escalation sharpened one rule three times (r5 "the fix shipped untested" → r6 "revert the fix to verify" → r7 "reverting structurally cannot catch a guarded assertion"), with round 7 explicitly invalidating round 6's rule for its own case. A memoryless reviewer could not have argued that. I verified the tree independently at `20449bd`: `go build ./...`, `go vet ./...`, `go test ./...`, `gofmt -l cmd pkg`, `construct/vocabulary/vet_test.sh`, `go run ./cmd/vocabulary check`, and `sdlc process-manual` are all clean, and every Core-concepts row resolves to a real symbol at the stated file. What keeps this off SHIP is a scoping gap that this gate demonstrates on itself: the whole-issue close is the **last** read before publish, and it was handed *"This is the FIRST round of this gate at this boundary — there are no prior findings to dispose of"* while the very ledger it writes lists **15 open findings, three of them Important**. Nothing after this gate will ever ask about them. I disposed the ten I could verify from the code (which I could only do because I read the ledger file myself, not because the gate showed it to me) — but the mechanism, not this round's diligence, has to be what closes that loop.

## 1. Strengths

- **`prompt.go:37` `RenderPriorFindingsScoped(scoped, full)` — the seam, verified live.** Families come from the full ledger, disposal from the boundary-scoped view. My own prompt carried ten families spanning M2 and M3 rounds, which is exactly the cross-milestone recurrence BR-20 said was structurally invisible.
- **`reviewanchor.go` is the shape the rest of `cmd/sdlc` should move toward.** Gather → classify → format, git confined to `gatherReviewAnchorDelta`, `classifyReviewAnchor` calling `publishGateHasCodeSurface` rather than restating the docs-vs-code rule (`reviewanchor.go:65`), and a fourth `anchorDiverged` state argued from `git diff A B`'s behavior on unrelated commits. The whole decision layer unit-tests with zero repo.
- **`planreview.go:80-100` `gateLedgerKind` — the ARCH-DRY extraction was actually done, not mirrored.** Task 2.1 Step 4 made it conditional; the two gates differed only by the triple, so the bodies were shared. One copy of "a corrupt sidecar is an error, never a silent reset."
- **Two tests whose *names* are the documentation.** `TestFamilyCounts_TrueSynonymsAreNotMerged_AcceptedResidualRisk` (`family_test.go:78`) records the limit D3 accepted instead of pretending normalization solves synonyms; `TestBoundaryWindowBase_WholeIssueStaysAtMergeBase` (`boundaryledger_test.go:486`) is a rejected-alternative test whose failure message explains *why* M4 lost.
- **`ledger.go:74-89` — `NoCap`'s negative spelling plus an honest KNOWN LIMITATION.** Pre-#194 rounds default to counting, so no historical ledger changes meaning; and the comment states plainly that an interrupted review is byte-indistinguishable from a non-compliant one and is deliberately *not* excluded, naming the cost this issue itself paid. That is a comment that survives being checked.

## 2. Critical findings

None.

## 3. Important findings

**I1 — the whole-issue close, the last gate, is shown zero prior findings while 15 are open in the same ledger.** *(New family `boundary-scope-strands-findings`; BR-20 is the same underlying rule filed under a different slug — see below.)*

Evidence, not inference — this round's own prompt opens with *"This is the FIRST round of this gate at this boundary — there are no prior findings to dispose of"*, while `workshop/plans/000194-review-anchor-commit-close-gate.md`'s `## Open findings` lists BR-10, BR-14, BR-15, BR-16, BR-17, BR-18, BR-19, BR-26, BR-27, BR-28, BR-29, BR-30, BR-33, BR-35, BR-36 — three at **Important**.

The two surfaces disagree by construction. `Render` (`render.go:120`) computes `## Open findings` from `OpenFindings(l)` on the **full** ledger; `boundaryPriorFindings` (`boundaryledger.go:81`) hands the reviewer `FilterBoundary(l, p.Milestone)`, and at a whole-issue close `Milestone == ""` drops every M1/M2/M3 round. So the durable artifact says 15 open and the final gate says 0, and there is no later gate — `sdlc merge`/`push` run no LLM judge.

Four of those fifteen are in fact **fixed and merely undisposed** (BR-10, BR-14, BR-15, BR-36 — I verified each against the source and am disposing them in the fence below), which is the second half of the defect: a finding fixed *after* its boundary's last round has no path to disposal at all, because the only path is another round at that same boundary and boundaries close once. BR-16 is a genuine half-miss that nobody will ever be asked about again: `gate-state.md` was correctly updated, but `atlas/index.md:14` still reads *"the plan-gate → boundary-review carry-forward (#187; **#183 is the second intended consumer**)"* — #194 *is* that consumer and it landed.

**Do not fix this by dropping the filter.** D1's cap argument still holds and I checked it: this ledger has 8 counted rounds (rounds 1 and 2 carry no `no_cap`, being host-sleep interruptions), so `Decide` on an unfiltered view would report `CapReached` and silently demote every Important on round one of the gate this work exists to strengthen.

**The rule** — the same one BR-20 forced onto the prompt's family counts, applied one field over: *scope per boundary only what the round cap needs; every other read of the ledger wants the full issue.* BR-20's slug (`family-plumbing-incomplete`) named its symptom rather than that rule, which is why the second instance did not escalate; re-slugging it would be the honest bookkeeping.

*Fix sketch (cheap, one line + one test):* in `boundaryPriorFindings`, pass the full ledger as `scoped` when `p.Milestone == ""` — the whole-issue close **is** the boundary that covers everything, so its dispose-set should be issue-wide. Keep `Decide(FilterBoundary(...))` unchanged so blocking semantics don't shift in the same commit; whether a demoted-at-M2 Important should block the final close is a separate design call worth stating in `gate-state.md` beside D1. Pin it with a test that seeds an `M1`-boundary open finding and asserts it appears in the `Milestone: ""` prompt — the mirror of `TestBoundaryPriorFindings_FamiliesSpanMilestones`.

## 4. Minor findings

- **`test-pins-the-invariant`, 4th instance — do NOT add a test; record the rule's third exception.** I mutation-verified M1-I3's `windowHead` pin (`close.go:474-478`): deleting it leaves `go test ./cmd/sdlc/` green (52.9s). That is *correct*, not a gap — both `rev-parse` calls run under one lock hold, so no interleaving can distinguish them; the fix converts an incidental identity into a structural one and is unfalsifiable by construction. `lessons.md:941` states the rule as "a fix is complete only when a test FAILS WITHOUT IT", which here would force a fake test or mark good work incomplete. This is a base-layer file that propagates downstream, so the exception belongs in it: *when a fix removes a possible divergence rather than an actual one, the honest record is "structural — no behavioral difference to pin" in the Log, not a manufactured test.* That is the third limit this issue has found in its own rule (r6 revert-to-verify → r7 guarded assertions → this).
- Three stale line numbers survive BR-32's correction: the plan's `Round.Boundary → ledger.go:57` (actual `:79`), `closeReviewSnapshot → close.go:1182` (actual `:1244`), `resolveReviewWindow → milestoneclose.go:243` (actual `:270`). Every symbol resolves at the stated *file*, so this is navigational drift only — and it is precisely what #198's deterministic path/symbol check is filed to catch. Don't patch by hand.
- `printBoundaryReviewDryRun` (`milestoneclose.go:566`) sets neither `PlansDir` nor `PriorFindings`, so `--dry-run` shows a prompt a real run would never send.
- `isResolvedSHA` (`reviewanchor.go:33`) accepts any ≥7-char all-hex string, so a branch literally named `deadbeef` would read as a resolved anchor. Not worth code, but worth knowing.

## 5. Test coverage notes

Coverage of the pure layer is genuinely good and the fixtures discriminate: `TestFilterBoundary_KeepsEachBoundaryUnderTheRoundCap` asserts the *unfiltered* precondition before the filtered claim, so removing the scoping fails rather than passing vacuously; `TestConvergenceLine_LaterRoundsAreNotPriorFamilies` puts the debut family in the *middle* round; `FuzzRenderParseRoundTrip` now fuzzes `family` with a migrated crasher seed. Remaining gaps:

1. **No test enters at the `Milestone: ""` / cross-boundary prompt seam (I1).** `TestBoundaryPriorFindings_FamiliesSpanMilestones` covers families crossing boundaries; nothing covers *open findings* not crossing, which is the behavior I1 is about, and nothing would fail if that behavior changed either way.
2. `TestBoundaryReview_EmitsConvergenceLine` (`boundaryledger_test.go:474`) still asserts `Contains("round 2")` / `Contains("1 repeat family")` / `Contains("Not converging")` — the shape pin BR-36 asked for landed in `TestConvergenceLine` (`family_test.go:105`, exact string) but not in the test that exercises the *emission* path. Adequate in combination; worth knowing the stderr path is substring-only.
3. `pkg/vocab/finding_test.go:85` still does not assert `family:` (BR-29 item two, disposed `not-addressed` below). The goldens cover it indirectly, but a judge never told to emit `family` defeats the milestone silently, and the model↔prompt drift guard is where that invariant belongs.
4. `ConvergenceLine`'s stderr emission is never exercised through a full `close` / `milestone-close` run against M2's stateful reviewer fake.

## 6. Architectural notes for upcoming work

- **ARCH-DRY — pass, one flag.** `gateLedgerKind`, `roundCapFromEnvVar`, `NormalizeFamily → issue.Slugify`, `abbrevSHA` reused rather than duplicated, `classifyReviewAnchor → publishGateHasCodeSurface`. The flag is the one M2's review predicted and that materialized: the plan gate's persist tail stamps `Blocked` **and** `Forced` (`changecode.go:539-540`), the boundary tail stamps only `Blocked` (`boundaryledger.go:180`) — that is BR-17, still open, and it is exactly "the two persist tails diverge again at the one line nobody notices." Before a third gate adopts this ledger, extract `applyGateRound(kind, ledger, report, cap) (Ledger, Decision)`.
- **ARCH-PURE — pass.** `family.go`, `prompt.go`, `FilterBoundary`, `CountedRounds`, `Decide`, `classifyReviewAnchor` and both anchor formatters are deterministic and IO-free; the `gatestate` suite needs exactly one `os.ReadFile` (the `tools#1` fixture) and no mocks. IO stays in `boundaryledger.go`/`planreview.go`/`gatherReviewAnchorDelta`. `RenderPriorFindingsScoped` widened a pure signature rather than pushing a boundary parameter into `Decide` — the same caller-side-transform discipline `FilterBoundary` set.
- **ARCH-PURPOSE — flag (I1), plus one shadow worth naming.** The `family` shadow-sweep: prompt vocabulary ✓, escalation ✓, convergence line ✓, plan-gate seed ✓, write-path normalizers ✓ — but the two **durable** consumers do not derive. `render.go:110-116` prints id/severity/title/detail with no family (BR-28, open), and the convergence line is stderr-only. So once the session scrolls away, no persisted artifact renders the families the gate is tracking; the data survives only as unrendered YAML frontmatter. Combined with I1, the durable record is the surface this issue served least.
- **ARCH-MOCK — pass.** No external dependency added. Git stays behind the hermetic real-repo fixture (real `git`, real commits, channel-blocked reviewer for the interleaving tests); the reviewer stays behind `judge.Run`, now with M2's genuinely stateful `fakeReviewer` that reads its own prior-findings block back out of the prompt — the gap M1's review flagged for exactly this milestone. The `tools#1` four-round history is a copied `testdata/` fixture rather than a sibling-checkout read, so the acceptance test is hermetic. No live conformance check exists for the reviewer CLI's fence compliance, but the persisted `ProtocolError` round *is* the modeled response to drift, which is the right posture for a dependency whose contract is a prompt.
- **For #197/#198.** #197 (derive the window from the ledger) now has a second consumer argument in `milestoneHasVerdictCommit` — worth noting there that the ledger's `Boundary` field is the join key it would use. #198 should include the plan's *line numbers*, not just paths and symbols; three drifted while the five BR-32 corrected stayed right.

## 7. Plan revision recommendations

The plan matches the code — every Core-concepts row resolves to a real symbol at the stated file, and the M3 `## Revisions` entry records the scoped/full split, the dropped `DispositionCounts` reuse, and both adopted rules. Two additions:

- **New `## Revisions` entry — "D1 scoped more than the cap required (close review, I1)":** D1's stated justification is the round cap, and that justification still holds (measured: 8 counted rounds vs a cap of 3). But D1 scoped the *open-findings set* too, and only the cap needed it — so the whole-issue close, the last gate before publish, is shown none of the issue's open findings while the ledger's own `## Open findings` lists fifteen. Record the seam (`boundaryPriorFindings` passes the full ledger as `scoped` when `Milestone == ""`) and the open design call (does a demoted-at-M2 Important block the final close?), so the next gate to adopt this ledger inherits the rule instead of re-deriving it. Mirror the decision into `atlas/workflow/gate-state.md` beside the existing D1 paragraph.
- **Extend the M3 entry with the mutation verifications** it adopts as a rule but does not itself record: reverting `Family: f.Family` fails `TestSeedFromPlanGate_CarriesFamilyAcrossGates`; reverting `r.N >= round` fails `TestConvergenceLine_LaterRoundsAreNotPriorFamilies`; and — the new datum — reverting M1-I3's `windowHead` pin leaves the suite green *correctly*, which is the third exception the rule needs.

```findings
dispose:
  - id: BR-10
    disposition: addressed
    note: |
      Both halves fixed and verified in source — ApplyChecked now rejects per-disposition (ledger.go:248-263) and changecode.go:528-533 uses ApplyChecked's own `applied` round instead of hand-rebuilding one.
  - id: BR-14
    disposition: addressed
    note: |
      TestGateLedgerRefusal_BlocksAPassingVerdictAndIsBypassable covers the passing-verdict refusal AND --no-ledger through executeSDLCTestCommand; TestBlockOnLedgerFailure_FailsClosed covers the unusable-ledger branch.
  - id: BR-15
    disposition: addressed
    note: |
      ledger.go:74-89 now claims exactly TWO kinds and states the interrupted-review case as a KNOWN LIMITATION rather than as motivation; the issue Log's "Not decided here" is replaced by "Cap accounting: DECIDED". All three artifacts agree.
  - id: BR-16
    disposition: not-addressed
    note: |
      gate-state.md is fixed (the CountedRounds / no_cap paragraphs are there). The second half is not — atlas/index.md:14 still reads "#183 is the second intended consumer" after #194 delivered it, and that blurb still enumerates only #187's surface.
  - id: BR-17
    disposition: not-addressed
    note: |
      Verified by grep — "Forced" appears nowhere in cmd/sdlc/boundaryledger.go, while changecode.go:540 stamps it. A --force or --no-ledger bypass at the boundary still leaves no durable record.
  - id: BR-26
    disposition: not-addressed
    note: |
      All four sub-items unchanged, and all four misfired on THIS round's prompt — it rendered family totals as "3 new findings", named only escalation-copy-precision of ten families in play with only its ordinal, and swept six count-1 families into "a rule that has already been patched at least once".
  - id: BR-28
    disposition: not-addressed
    note: |
      render.go:110-116 still prints id/severity/title/detail with no family, in either projection. With the convergence line being stderr-only, no durable artifact renders the families the gate tracks.
  - id: BR-29
    disposition: not-addressed
    note: |
      Item one stays covered by FuzzRenderParseRoundTrip. Item two is not — grep for "family" in pkg/vocab/finding_test.go returns nothing, so the model-to-prompt drift guard still does not pin the key.
  - id: BR-33
    disposition: not-addressed
    note: |
      Ran vet_test.sh myself (ok). It has -d '#Project' instance cases at :45 and :48 and still no -d '#Finding' equivalent, so the "closed schema, an unmodeled key fails instance validation" rationale remains unenforced.
  - id: BR-36
    disposition: addressed
    note: |
      helptext/close.md:72-74 now shows a shape the formatter can emit (the ", N disposed" segment is present in both examples) and family.go:148 dropped the markdown; TestConvergenceLine pins both exact strings.
findings:
  - id: new
    severity: Important
    family: boundary-scope-strands-findings
    title: |
      The whole-issue close sees zero prior findings while the same ledger holds 15 open, three at Important — and no gate follows it
    detail: |
      boundaryledger.go:81 scopes the prompt to FilterBoundary(l, ""), which drops every
      M1/M2/M3 round, while render.go:120 builds the durable "## Open findings" from the
      FULL ledger. This round's own prompt therefore read "FIRST round ... no prior
      findings to dispose of" against a ledger listing BR-10/14/15/16/17/18/19/26/27/28/
      29/30/33/35/36. Four of those are already fixed and merely undisposed, because a
      finding fixed after its boundary's last round has no path to disposal at all. Do
      NOT drop the filter — measured, this ledger has 8 counted rounds against a cap of
      3, so an unfiltered Decide would demote every Important on round one. The RULE, and
      it is BR-20's rule one field over (BR-20's slug named its symptom, which is why the
      second instance did not escalate): scope per boundary only what the round cap
      needs; every other read wants the full issue. Cheapest correct fix — pass the full
      ledger as `scoped` when Milestone == "", since the whole-issue close IS the
      boundary that covers everything, and pin it with the mirror of
      TestBoundaryPriorFindings_FamiliesSpanMilestones.
  - id: new
    severity: Minor
    family: test-pins-the-invariant
    title: |
      The revert-to-verify rule needs a third exception — a fix that removes a possible divergence has no failing test by construction
    detail: |
      Mutation-verified: deleting close.go:474-478 (M1 I3's windowHead pin) leaves
      `go test ./cmd/sdlc/` fully green. That is correct rather than a gap — both
      rev-parse calls run under one lock hold, so no interleaving can distinguish them;
      the fix converts an incidental identity into a structural one. lessons.md:941 as
      written ("a fix is complete only when a test FAILS WITHOUT IT") would force a fake
      test or mark good work incomplete. Do NOT add a test for windowHead. This is the
      4th instance of the family and the third limit this issue has found in its own
      rule (r6 revert-to-verify, r7 guarded assertions, now this), so record the
      exception in lessons.md — which is a base-layer file that propagates downstream:
      when the fix removes a POSSIBLE divergence rather than an actual one, the honest
      record is "structural, no behavioral difference to pin" in the Log.
```

---

## Re-review — 2026-08-21T10:41:27-07:00 (FIX-THEN-SHIP)

| field | value |
|-------|-------|
| issue | 194 — boundary reviews: anchor to the reviewed commit, and remember across rounds |
| repo | ariadne |
| issue file | workshop/issues/000194-review-anchor-commit.md |
| boundary | whole-issue close |
| milestone | — |
| window | 33343455e21e25f056f41ab066b5face9e70f745..cf18f34ca235cd04cba75fb596159083baa34fd8 |
| command | sdlc close --issue 194 |
| reviewer | claude |
| timestamp | 2026-08-21T10:41:27-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

The issue delivers all ten Done-when items, and this round I could verify the *verifications* rather than trust them: I reverted BR-37's `openScopeFor` (two tests go red), BR-19's `dispositionsOfBoundaryAllFindings` (two assertions go red), and — the strongest evidence available — ran the real `000194-*-close-gate.md` through the shipped `DecideScoped`, which now returns `Block=true` naming BR-16 and BR-37, against `CapReached=false` at 1 counted close-boundary round. Round 9's prompt read *"FIRST round … no prior findings"* against that same file; mine carried 14 open findings and 11 families. That is the design hole closed and demonstrated on itself. `go build`, `go vet`, `gofmt -l cmd pkg`, `go test ./...`, `vet_test.sh`, `vocabulary check`, and `process-manual` are all clean at `cf18f34`, and the tree is clean. What keeps this off SHIP is that **two of the ten fixes in `cf18f34` do not do what the commit says**: `BR-17`'s `Forced` stamp is inert — `p.ForcedRationale` is assigned by none of the seven `boundaryReviewParams` construction sites, so deleting the line leaves the whole `cmd/sdlc` suite green — and BR-27's fix is correct but pinned by nothing. Both are exactly what this issue's own adopted rule (`lessons.md:942`, revert-to-verify) exists to catch, and both shipped in the commit that closed the findings that produced that rule. The same shape repeats one file over: `helptext/` still asserts the open-findings scoping BR-37 replaced, and `lessons.md:934`'s own grep recipe names `cmd/*/helptext/` as the place to look.

## 1. Strengths

- **`DecideScoped` (`decide.go:59`) splits the two questions instead of picking one.** The naive fix — drop the filter — is refused with a measured argument, and I confirmed the measurement: unfiltered, this ledger's 9 counted rounds against a cap of 3 would demote every Important on round one. Cap per boundary, open findings per issue. `TestWholeIssueClose_RoundCapStaysBoundaryScoped` pins the half that the obvious fix would have broken.
- **BR-19 was a genuine correctness bug, found and fixed at the right layer.** `FilterBoundary` (`ledger.go:125-145`) now carries a `BoundaryAll` finding's *disposal* across boundaries as a synthesized `NoCap` round, so the seeded finding stops re-opening forever. Mutation-verified: reverting it re-opens `BR-1` at `"M2"` and `""`.
- **`ConvergenceLine`'s exact shape is pinned twice** (`family_test.go:105,115`) against two full strings, and the helptext examples were corrected to a shape the formatter can actually emit. That is BR-36's rule — prose about the gate's own output must be pinned to it — applied rather than restated.
- **The commit message records a mis-fix and why it was wrong.** BR-26 was first "fixed" by raising the dead `counts[fam] >= 1` guard to 2, the acceptance test caught it, and the reasoning ("removing the no-op guard is the fix; lowering the bar was the wrong reading") is in the record. A wrong turn documented is worth more than a clean-looking diff.
- **`reviewanchor.go` remains the shape the rest of `cmd/sdlc` should move toward** — gather → classify → format, git confined to `gatherReviewAnchorDelta`, `classifyReviewAnchor` *calling* `publishGateHasCodeSurface` rather than restating the docs-vs-code rule. The whole decision layer unit-tests with zero repo.

## 2. Critical findings

None.

## 3. Important findings

**I1 — `cf18f34` claims two fixes it did not make.** *(5th finding in family `test-pins-the-invariant`.)*

Mutation-measured, not inferred:

| fix | mutation | suite |
|---|---|---|
| BR-37 `openScopeFor` | always filter | **FAIL** (2 tests) |
| BR-19 disposal carry | drop the branch | **FAIL** (2 assertions) |
| BR-27 `display` counter | `display := round` | **PASS — nothing caught it** |
| BR-17 `Forced` stamp | delete the line | **PASS — nothing caught it** |

BR-27 is at least *correct*; it is only unpinned. BR-17 is worse — it is **inert**. `boundaryledger.go:190` assigns `l.Rounds[last].Forced = p.ForcedRationale`, and `ForcedRationale` (`milestoneclose.go:555`) is set at **zero** of the seven `boundaryReviewParams{…}` construction sites (`close.go:1010,1023,1052,1060`; `milestoneclose.go:193,226,234`). So `Round.Forced` is still `""` on every boundary round and a `--no-ledger`/`--force` bypass still leaves no durable record — BR-17's stated defect, unchanged. Two further layers: the justifying comment at `boundaryledger.go:188` names the "close-time `gate_forced` metric", but `churnreport.go:111` reads only `readPlanGateLedger` — the boundary ledger is never read by that metric at all; and even wired, the stamp is unconditional where the plan gate gates on the decision (`changecode.go:540`, `forcedRationale(f.Force, d.Block)`), so it would over-report.

Per the escalation I am not asking for these two instances to be patched. **The rule is already written** — `lessons.md:942`: *"A fix for a gate finding is complete only when a test FAILS WITHOUT IT — verified by reverting the fix … not by inspection."* It was adopted at round 7 and violated at round 10 by the commit closing the findings that produced it. So restating it is not the fix; **enforcing it is**, and the enforceable version is consumer-side:

> A reviewer must not dispose a finding `addressed` on the strength of the code reading correctly. Revert the fix and confirm a named test goes red; where no test can distinguish the two worlds, say which recorded exception applies.

That belongs in `cmd/sdlc/internal/judge/code-review.md`, beside the dispose-first contract — one paragraph, and the binary already owns that prompt. The evidence for the asymmetry: the producer-side habit has now failed 6 times on this issue (BR-21, BR-22, BR-29 ×2, BR-27, BR-17); the consumer-side check has caught it 3 for 3 (rounds 6, 7, 10). Measured prevalence of `test-pins-the-invariant` on this issue: **6 instances**. When you do fix BR-17, do it as the `applyGateRound(kind, ledger, report, cap) (Ledger, Decision)` extraction M2's review recommended — this persist tail has now diverged three times at the one line nobody notices (`Blocked` at BR-3, `Forced` at BR-17, `Forced`-but-unwired now), which is the ARCH-DRY root cause.

**I2 — BR-37's behavior change left the helptext asserting the contract it replaced, and atlas has no entry for the new rule.** *(New family `doc-asserts-replaced-mechanism`; the underlying rule has 3 prior instances under other ids — M1 I1, M2 I3, BR-16.)*

Two embedded, agent-facing, base-layer sites now state a false contract:

- `cmd/sdlc/helptext/milestone-close.md:99` — *"The round cap **and the open-findings set** scope PER BOUNDARY"*. After `f2da4c4` the open-findings set does **not** scope per boundary at the whole-issue close; that is the entire point of the change.
- `cmd/sdlc/helptext/close.md:46` — *"the next review **at the same boundary** is shown them"*. The whole-issue close is now shown every boundary's findings.

And `atlas/` has no row for it: `grep -rn "DecideScoped\|openScopeFor\|last gate before publish" atlas/` returns nothing, so a new exported `gatestate` API and the rule it encodes ("scope per boundary only what the round cap needs; every other read wants the full issue") exist only in a Go comment. `gate-state.md` documents the cap and the demotion but not the scoping split it now depends on.

This is the miss the diff's own `lessons.md:930-936` entry predicts, down to the grep target: *"`grep -rn "<the old mechanism's name>" atlas/ cmd/*/helptext/` — searching for what you REMOVED."* The rule is written, has an explicit recipe, and the very next behavior change violated it. Same meta-shape as I1: a rule adopted about this issue's own process, landed in `lessons.md`, with no gate behind it. Fix the two helptext sentences, and add the D1/BR-37 scoping split to `gate-state.md` beside the existing cap paragraphs (it is the decision the next gate to adopt this ledger will inherit).

## 4. Minor findings

- **The plan contradicts itself about `DispositionCounts`.** *(3rd finding in family `plan-artifact-lags-code`.)* `plan.md:244` still says `ConvergenceLine` *"Reuses the existing `DispositionCounts` (`ledger.go:202`)"*; `plan.md:584` — the plan's own `## Revisions` — says it *"did **not** reuse `DispositionCounts`"*. The code uses `len(r.Dispositions)`, and `DispositionCounts` is at `ledger.go:342`, not `:202`. Do not patch the line: BR-32 already stated the rule (the durable plan has no gate) and filed **#198**. What this adds is a **scope refinement for #198**: its check must cover *prose claims about reuse*, not only the path/symbol table rows — a row-only check would pass this file while it asserts a reuse the same file retracts. Also still drifted: `ledger.go:57` (is `Dispositions`, `Round.Boundary` is `:79`), `close.go:1182`, `milestoneclose.go:243`.
- `Decision.Reason` now mixes scopes at the whole-issue close — *"no open blocking findings after 1 round(s)"* counts close-boundary rounds while the findings span the issue.
- `ConvergenceLine`'s `display` (`family.go:158`) excludes the current round when it is itself `NoCap`, so a never-dispatched round re-prints the previous round's number.
- `printBoundaryReviewDryRun` still sets neither `PlansDir` nor `PriorFindings`, so `--dry-run` shows a prompt no real run would send.
- `reviewThenFinalize`'s dispatch path (`close.go:1023`, `milestoneclose.go:193`) is unreachable in production — every caller short-circuits before it — but it carries no `PriorFindings`, so it is a latent re-entry point for BR-1's defect.

## 5. Test coverage notes

Mutation coverage on this round's four code fixes is **2 of 4** (I1's table). Beyond that:

1. Nothing exercises `ConvergenceLine` with a `NoCap` round present — the only case BR-27 was about. `grep NoCap family_test.go boundaryledger_test.go` returns nothing.
2. Nothing asserts `Round.Forced` on a boundary ledger, which is why an inert assignment reads as a fix.
3. `TestBoundaryReview_EmitsConvergenceLine` (`boundaryledger_test.go:475`) still asserts substrings on the emission path; the exact shape is pinned only in `family_test.go`. Adequate in combination.
4. `seedFromPlanGate` is still only ever exercised on an empty ledger, so BR-13's `AssignIDsAt` fix stays correct-by-construction and unpinned.
5. Genuinely good: `TestFilterBoundary_KeepsEachBoundaryUnderTheRoundCap` asserts the *unfiltered* precondition before the filtered claim, so the scoping cannot pass vacuously; `FuzzRenderParseRoundTrip` fuzzes `family` with a migrated crasher seed; `TestFamilyCounts_TrueSynonymsAreNotMerged_AcceptedResidualRisk` documents a limit in its name rather than pretending it is solved.

## 6. Architectural notes for upcoming work

- **ARCH-DRY — flag (I1).** Real consolidations: `gateLedgerKind`/`readGateLedger`/`writeGateLedger`, `roundCapFromEnvVar`, `NormalizeFamily`→`issue.Slugify`, `abbrevSHA`, `classifyReviewAnchor`→`publishGateHasCodeSurface`. The flag is the persist tail: `changecode.go:539-540` and `boundaryledger.go:189-190` are the same six-step sequence written twice, and the only step that has ever diverged is the one nobody reads. Third divergence. Extract `applyGateRound` before a third gate adopts this ledger.
- **ARCH-PURE — pass.** `family.go`, `prompt.go`, `decide.go`, `FilterBoundary`, `CountedRounds`, `openScopeFor`, `classifyReviewAnchor` and both anchor formatters are deterministic and IO-free; the `gatestate` suite needs exactly one `os.ReadFile` (the `tools#1` fixture) and no mocks. `DecideScoped` widened a pure signature rather than pushing a boundary parameter into `Decide` — the caller-side-transform discipline `FilterBoundary` set.
- **ARCH-PURPOSE — flag (I1, I2).** The `family` shadow-sweep is now complete on every consumer except the CUE `#Finding` source, which enforces nothing (BR-33, open) so the Go struct and the `family: <slug>` literal remain hand restatements. New this round: `Round.Forced` is a field declared, stamped, and never populated — a consumer that does not derive from anything; and the two helptext sites are restatements of a model the code no longer implements.
- **ARCH-MOCK — pass.** No external dependency added. Git stays behind the hermetic real-repo fixture (real `git`, real commits, channel-blocked reviewer for the interleaving tests); the reviewer stays behind `judge.Run` with M2's stateful `fakeReviewer`; the `tools#1` history is a copied `testdata/` fixture, so the acceptance test is hermetic. No live conformance check that a real reviewer CLI emits `family:`, but the persisted `ProtocolError` round is the modeled response to that drift — the right posture for a dependency whose contract is a prompt.
- **The generalizable point behind I1 and I2.** This issue adopted two rules about its own process, wrote both to `workshop/lessons.md`, and both were violated by the next commit. `lessons.md` is read at session start (AGENTS.md §4) and has no gate — the same diagnosis BR-32 reached for the durable plan, which became #198. Rules this issue discovers about *gate behavior* should land where the gate reads them (`code-review.md`, the close gate), with `lessons.md` as the narrative, not the enforcement.

## 7. Plan revision recommendations

- **New `## Revisions` entry — "helptext and atlas not swept for BR-37 (close review round 10)"**: record that `f2da4c4` changed the open-findings scope without updating `helptext/milestone-close.md:99`, `helptext/close.md:46`, or `atlas/workflow/gate-state.md`, and that `lessons.md:934`'s grep recipe names both directories. State the scoping split as a decision the next adopter inherits.
- **Correct `plan.md:244`** to match `plan.md:584` (`ConvergenceLine` counts `len(r.Dispositions)`; it does not reuse `DispositionCounts`, which is at `ledger.go:342`).
- **Finish the Core-concepts table (BR-18).** Nine entities the milestones actually built are still absent: `readGateLedger`/`writeGateLedger`, `seedFromPlanGate`, `persistBoundaryRound`, `boundaryPriorFindings`, `blockOnLedgerFailure`, `roundCapFromEnvVar`, `AssignIDsAt`, `renderFamilyVocabulary`/`renderFamilyEscalation`.
- **Extend the M3 entry with this round's mutation results**, per the rule the same entry adopts: BR-37 and BR-19 fail loudly when reverted; BR-27 and BR-17 do not, and BR-17 is inert rather than merely unpinned.

```findings
dispose:
  - id: BR-16
    disposition: not-addressed
    note: |
      gate-state.md:78-80 now documents CountedRounds and no_cap. The second half is not fixed — atlas/index.md:14 still reads "#183 is the second intended consumer" after #194 delivered that consumer, and the blurb still enumerates only #187's surface.
  - id: BR-17
    disposition: not-addressed
    note: |
      The fix is INERT, not merely unpinned. boundaryledger.go:190 assigns p.ForcedRationale, which zero of the seven boundaryReviewParams construction sites ever set; mutation-verified — deleting the line leaves the whole cmd/sdlc suite green. Round.Forced is still "" on every boundary round. Its comment at :188 also names the gate_forced metric, but churnreport.go:111 reads only the PLAN gate ledger.
  - id: BR-18
    disposition: not-addressed
    note: |
      The table gained gateLedgerKind, CountedRounds, Round.NoCap, DecideScoped and openScopeFor, and Revisions gained the shared-gatestate entry. Nine named entities are still absent — readGateLedger/writeGateLedger, seedFromPlanGate, persistBoundaryRound, boundaryPriorFindings, blockOnLedgerFailure, roundCapFromEnvVar, AssignIDsAt, renderFamilyVocabulary/renderFamilyEscalation.
  - id: BR-19
    disposition: addressed
    note: |
      FilterBoundary carries a BoundaryAll finding's disposal across boundaries via dispositionsOfBoundaryAllFindings; mutation-verified — dropping the branch re-opens BR-1 at both "M2" and "".
  - id: BR-26
    disposition: not-addressed
    note: |
      Two of four fixed — the dead counts[fam] >= 1 guard is gone and the blockquote now loops every family instead of repeats[0]. Two remain and both misfired on THIS round's prompt — family.go:85 rendered "test-pins-the-invariant  4 new findings" for a running total, and family.go:108 asserted "Earlier rounds fixed instances" for doc-claim-exceeds-enforcement, whose only finding (BR-33) has never been fixed.
  - id: BR-27
    disposition: addressed
    note: |
      family.go:158 computes the display position from non-NoCap rounds, so a seed round no longer makes the first real review read as "round 2". Correct but UNPINNED — no test exercises ConvergenceLine with a NoCap round, and reverting to `display := round` leaves the suite green; folded into the new test-pins-the-invariant finding.
  - id: BR-28
    disposition: addressed
    note: |
      render.go:112 and :127 both render familyTag, so the family shows in the per-round Raised list and the Open-findings projection; pinned by TestFamily_SurvivesRoundTripAndIsNamedInTheFence.
  - id: BR-29
    disposition: addressed
    note: |
      Both items — the round-trip is covered by FuzzRenderParseRoundTrip plus TestFamily_SurvivesRoundTripAndIsNamedInTheFence, and pkg/vocab/finding_test.go:130 now pins "family: <slug>" in the emitted fence instruction.
  - id: BR-30
    disposition: not-addressed
    note: |
      One of three — the convergence cinfo now sits with its own comment and the demotion comment is adjacent to its loop again. pkg/vocab/finding.go:78 still says the block instruction is "for the plan-quality prompt", and TestBoundaryWindowBase_WholeIssueStaysAtMergeBase still lives in boundaryledger_test.go rather than beside its siblings in milestonewindow_test.go.
  - id: BR-33
    disposition: not-addressed
    note: |
      Ran vet_test.sh myself (ok). It vets finding.cue at :58 and has -d '#Project' instance cases at :45 and :48, but still no -d '#Finding' equivalent, so the "closed schema, an unmodeled key fails instance validation" rationale remains unenforced.
  - id: BR-35
    disposition: addressed
    note: |
      family.go:110 now says "record the family, with its measured prevalence, in the finding's own detail" — a sink the model defines. Verified by grep that no `Limits` reference survives outside quotations of the old text in the review sidecars.
  - id: BR-37
    disposition: addressed
    note: |
      DecideScoped plus openScopeFor; mutation-verified (reverting openScopeFor fails TestWholeIssueClose_SeesOpenFindingsFromEveryMilestone and TestWholeIssueClose_RoundCapStaysBoundaryScoped), and confirmed against the REAL ledger — DecideScoped now returns Block=true naming BR-16 and BR-37 with CapReached=false at 1 counted close-boundary round.
  - id: BR-38
    disposition: not-addressed
    note: |
      lessons.md:942-989 carries the revert-to-verify rule and the guarded-assertion sharpening, but no third exception — grep for "structural", "possible divergence" and "no behavioral difference" over lessons.md returns nothing.
findings:
  - id: new
    severity: Important
    family: test-pins-the-invariant
    title: |
      Two of the ten fixes in cf18f34 do not do what the commit says — BR-17's is inert, BR-27's is unpinned
    detail: |
      Mutation-measured this round: reverting openScopeFor (BR-37) and the BoundaryAll
      disposal carry (BR-19) both fail loudly; deleting `l.Rounds[last].Forced =
      p.ForcedRationale` (BR-17) and reverting `display` to the raw round number (BR-27)
      both leave the cmd/sdlc suite fully green. BR-17 is worse than unpinned — it is
      INERT: ForcedRationale is set at zero of the seven boundaryReviewParams
      construction sites (close.go:1010,1023,1052,1060; milestoneclose.go:193,226,234),
      so Round.Forced is still "" and a --no-ledger bypass still leaves no durable
      record. Its comment also names the gate_forced metric, which reads only the plan
      gate ledger (churnreport.go:111). Do NOT just wire the field. The rule IS already
      written — lessons.md:942, adopted at round 7 — and was violated at round 10 by the
      commit closing the findings that produced it, so restating it is not the fix.
      Enforce it consumer-side instead, in cmd/sdlc/internal/judge/code-review.md beside
      the dispose-first contract - a reviewer must not dispose `addressed` on the
      strength of the code reading correctly; revert the fix and confirm a named test
      goes red, or say which recorded exception applies. Measured asymmetry - the
      producer-side habit has failed 6 times on this issue (BR-21, BR-22, BR-29 twice,
      BR-27, BR-17); the reviewer-side check has caught it 3 for 3 (rounds 6, 7, 10).
      Measured prevalence of this family - 6 instances. When BR-17 is fixed, do it as the
      applyGateRound extraction M2's review recommended - this persist tail has now
      diverged three times at the same line (ARCH-DRY).
  - id: new
    severity: Important
    family: doc-asserts-replaced-mechanism
    title: |
      The BR-37 behavior change left two helptext sites asserting the contract it replaced, and atlas has no entry for the new rule
    detail: |
      helptext/milestone-close.md:99 states "The round cap AND the open-findings set scope
      PER BOUNDARY" and helptext/close.md:46 states "the next review at the same boundary
      is shown them". Both are now false at the whole-issue close, which is the entire
      point of f2da4c4 — and both are //go:embed'ed, agent-facing, base-layer text that
      propagates downstream. Separately grep -rn "DecideScoped|openScopeFor|last gate
      before publish" atlas/ returns nothing - a new exported gatestate API and the rule
      it encodes ("scope per boundary only what the round cap needs; every other read
      wants the full issue") exist only in a Go comment, and gate-state.md documents the
      cap and the demotion without the scoping split they now depend on. This is the miss
      the diff's OWN lessons.md:930-936 entry predicts, down to the grep target it names
      (`cmd/*/helptext/`). New slug, but the underlying rule already has three instances
      under other ids - M1 I1, M2 I3, and BR-16, whose second half is still open. Rule -
      when a mechanism moves, grep for what you REMOVED across atlas/ AND
      cmd/*/helptext/ in the same commit; and, because a lessons.md entry has no gate,
      record the scoping decision in gate-state.md where the next gate to adopt this
      ledger will read it rather than re-derive it.
  - id: new
    severity: Minor
    family: plan-artifact-lags-code
    title: |
      The plan asserts a DispositionCounts reuse that its own Revisions section retracts
    detail: |
      plan.md:244 says ConvergenceLine "Reuses the existing DispositionCounts
      (ledger.go:202)"; plan.md:584 says Task 3.4 did NOT reuse it. The code uses
      len(r.Dispositions), and DispositionCounts is at ledger.go:342. Three table line
      numbers also remain drifted - ledger.go:57 is Dispositions (Round.Boundary is :79),
      close.go:1182, milestoneclose.go:243. Do NOT patch the line - BR-32 already stated
      the rule (the durable plan is the only major SDLC artifact with no automated check)
      and filed 198. What this instance adds is a SCOPE REFINEMENT for that issue - its
      check must cover prose CLAIMS about reuse and mechanism, not only the
      path-and-symbol table rows, because a row-only check passes this file while the
      file contradicts itself. Prevalence of this family on the issue - 3.
```
