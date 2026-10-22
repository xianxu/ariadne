# Boundary Review — ariadne#548 (whole-issue close)

| field | value |
|-------|-------|
| issue | 548 — Actual queue evidence must belong to the measured issue head |
| repo | ariadne |
| issue file | workshop/issues/000548-actual-queue-evidence-must-belong-to-the-measured-issue-head.md |
| boundary | whole-issue close |
| milestone | — |
| window | 9b6896e4123f0a8d2790b6ec3b6f3a54c0d6a792..e51f0ec669c17280ae8d9a0ffb782a8f722b0e40 |
| command | sdlc close --issue 548 |
| reviewer | codex (openai; author openai) |
| isolation | fenced |
| timestamp | 2026-10-22T04:10:16-07:00 |
| verdict | REWORK |

## Review

```verdict
verdict: REWORK
confidence: high
```

The attribution behavior passed review and targeted tests. The sole blocker is a documentation-only contradiction in the plan’s core-concepts tables, classified Critical by this review’s explicit table-consistency rule. No runtime defect was found.

1. **Strengths**

   - Membership filtering precedes final queue ownership selection in `cmd/sdlc/actualinputs.go:400–423`.
   - The real Git/tracker fixture verifies foreign exclusion, exact own waits, ambiguous ownership refusal, and unknown-head diagnostics (`cmd/sdlc/actualqueue_issue_test.go:83`).
   - Immutable reads disable replacement objects; graph tests verify deduplication and resource limits.
   - Atlas documentation describes the new attribution flow. No new command, flag, or configuration surface requires a README update.

2. **Critical findings**

```findings
findings:
  - id: new
    severity: Critical
    family: core-concept-traceability
    title: |
      Correct the filter location in both plan tables
    detail: |
      workshop/plans/000548-suite-head-attribution-plan.md:13 and :83 place filterIssueQueueAttempts in actualqueue_head.go, but its definition is in cmd/sdlc/actualqueue_filter.go:33. This contradicts the required core-concepts location check. Update both rows and add a Revisions entry recording the file split; the correction is prose-only and needs no wording-presence test.
```

3. **Important findings**

   None.

4. **Minor findings**

   None.

5. **Test coverage notes**

   Both targeted runs passed, including the broader actual-measurement compatibility selection (`cmd/sdlc`: 98.881s). The full `pkg/golint` package passed separately; diff whitespace checks passed. The checkout remained unchanged. I did not independently rerun the owner’s Echo previews or the recorded full suite.

6. **Architectural notes**

   - **ARCH-DRY — pass:** shared issue-reference grammar and existing ownership selector reused.
   - **ARCH-PURE — pass:** parsers, reducer, and filter remain IO-free.
   - **ARCH-PURPOSE — pass behavior:** foreign evidence is excluded while uncertain relevant evidence stays strict; plan traceability correction remains above.
   - **ARCH-MOCK — pass:** injected graph fake and private real-Git fixtures exercise the reader boundary.
   - **ARCH-CONSTRAINTS — pass:** node, byte, batch, and depth ceilings have executable controls.
   - **ARCH-SECURE — pass:** immutable reads reject malformed framing and bypass replacement refs.
   - **ARCH-ORDER — pass:** classification precedes attribution; read failures cannot produce partial adopted membership.
   - **ARCH-FUNERAL — pass:** caches are command-local; no new durable artifact family.
   - **ARCH-EVOLVE — pass:** persisted formats remain unchanged, and historical replay retains its existing path.

7. **Plan revision recommendation**

   Add: `### Filter module location` — record that the pure filtering helpers live in `cmd/sdlc/actualqueue_filter.go`, and correct both table references.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"b4cf1618-e638-4b66-960a-0000cc511468","repository":"github.com/xianxu/ariadne","issue_id":"000548","recipe":"milestone-review","round":1,"attempt":1,"agent":"codex","root":"/Volumes/TestScratch/review-checkouts/e619f74c7203281b","reviewed_head":"e51f0ec669c17280ae8d9a0ffb782a8f722b0e40","slot":"ariadne:1","requested_at":"2026-10-22T11:06:46.356778Z"},"phase":"finished","pid":76099,"started_at":"2026-10-22T11:06:46.591701Z","finished_at":"2026-10-22T11:10:16.559271Z","session_id":"01a1634a-c59f-7371-9197-fa8d47abcf44","outcome":"returned"}] -->

---

## Re-review — 2026-10-22T04:15:09-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 548 — Actual queue evidence must belong to the measured issue head |
| repo | ariadne |
| issue file | workshop/issues/000548-actual-queue-evidence-must-belong-to-the-measured-issue-head.md |
| boundary | whole-issue close |
| milestone | — |
| window | 9b6896e4123f0a8d2790b6ec3b6f3a54c0d6a792..f4d418291df2bf170ea38006a4a55052526ce60b |
| command | sdlc close --issue 548 |
| reviewer | codex (openai; author openai) |
| isolation | fenced |
| timestamp | 2026-10-22T04:15:09-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

The pinned implementation satisfies the issue’s attribution contract: immutable tested-head membership precedes slot ownership, proven foreign attempts are excluded, and unresolved relevant evidence still refuses. BR-1 is addressed. No new blocking findings emerged.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      Inspected the pinned correction diff: both plan rows at lines 13 and 83 now name actualqueue_filter.go, matching the definition at cmd/sdlc/actualqueue_filter.go:33. The Revisions entry at lines 98–100 records the split. This is a prose-only correction.
```

1. **Strengths**

   - The shared attachment path applies membership filtering before ownership selection (`cmd/sdlc/actualinputs.go:400–423`), covering close and correction.
   - Immutable reads disable replacement refs at the Git boundary and validate complete batch framing (`cmd/sdlc/internal/gitx/commits.go:27`).
   - The real Git/tracker regression distinguishes foreign, own-crossing, own-exact, and unavailable-head evidence (`cmd/sdlc/actualqueue_issue_test.go:20`).
   - Graph tests exercise shared-node deduplication, discovery depth, resource refusal, and first-parent train membership.

2. **Critical findings:** None.

3. **Important findings:** None.

4. **Minor findings:** None.

5. **Test coverage notes**

   Passed independently:

   - Attribution, graph, queue-filter, real tracker, and commit-object tests.
   - Broader owned-actual, correction, replay, queue-selection, and legacy-cutover tests (`cmd/sdlc`: 111.325s).
   - `go test ./pkg/golint -count=1`.
   - Pinned-range whitespace check.

   The checkout remained unchanged. I did not independently repeat the owner-run Echo previews or recorded full suite; those remain release-verification obligations.

6. **Architectural notes**

   - **ARCH-DRY — pass:** shared issue-reference grammar and existing ownership selector are reused.
   - **ARCH-PURE — pass:** parsing, membership reduction, and filtering remain IO-free; graph reads use an injected seam.
   - **ARCH-PURPOSE — pass:** shared close/correction consumers enforce attribution; retained replay preserves historical evidence.
   - **ARCH-MOCK — pass:** stateful graph fixtures exercise the production walker; private real-Git tests validate the adapter.
   - **ARCH-CONSTRAINTS — pass:** object, byte, batch, and depth limits reject incomplete proof.
   - **ARCH-SECURE — pass:** canonical identities, raw-object validation, and replacement isolation protect provenance.
   - **ARCH-ORDER — pass:** explicit classification states preserve unknown evidence; no durable event-driven state is introduced.
   - **ARCH-FUNERAL — pass:** new caches are command-local; no durable artifact family is added.
   - **ARCH-EVOLVE — pass:** persisted formats remain unchanged.

   The atlas documents the new attribution flow. No new command, flag, or configuration surface requires a README update.

7. **Plan revision recommendations:** None.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"18a5a9c3-7bb1-459a-abd9-959bb45873e6","repository":"github.com/xianxu/ariadne","issue_id":"000548","recipe":"milestone-review","round":2,"attempt":1,"agent":"codex","root":"/Volumes/TestScratch/review-checkouts/69db6e07476ba64a","reviewed_head":"f4d418291df2bf170ea38006a4a55052526ce60b","slot":"ariadne:1","requested_at":"2026-10-22T11:11:58.060062Z"},"phase":"finished","pid":62054,"started_at":"2026-10-22T11:11:58.208682Z","finished_at":"2026-10-22T11:15:08.52869Z","session_id":"01a1634f-86c1-71b2-9f43-8ac4c94cf4c5","outcome":"returned"}] -->
