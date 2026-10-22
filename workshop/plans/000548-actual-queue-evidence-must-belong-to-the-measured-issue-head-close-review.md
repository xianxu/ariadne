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
