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
