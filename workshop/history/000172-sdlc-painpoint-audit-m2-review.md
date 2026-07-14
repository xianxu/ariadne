# Boundary Review — ariadne#172 (milestone M2)

| field | value |
|-------|-------|
| issue | 172 — sdlc painpoint audit |
| repo | ariadne |
| issue file | workshop/issues/000172-sdlc-painpoint-audit.md |
| boundary | milestone M2 |
| milestone | M2 |
| window | 5d3df75bac48bfccf4dcfc072a4cf7b62fce29f9..HEAD |
| command | sdlc milestone-close --issue 172 --milestone M2 |
| reviewer | claude |
| timestamp | 2026-07-14T13:38:08-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

I have everything I need. Composing the review now — one correctness gap in the firing-order ladder is the headline finding; everything else held up under scrutiny.

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

M2 delivers what the boundary claims: `detectRefusalRetries` and `detectFiringOrder` are pure detectors over the M1 invocation stream, folded into the report (markdown + `--json`), with every deferred M1-review Minor actually landed (per-invocation dedupe threaded into `aggregate`, zero-transcripts error, render/JSON-shape tests, `toolResultText` array test, stated-limits footer). Tests and vet are green (`go test ./cmd/sdlc/...` all ok), the plan's M2 Revisions entry honestly records the deviations, and atlas + helptext are updated. The one substantive defect: the firing-order ladder treats **every** `close`/`merge` invocation as having completed — a gate-**refused** close still raises the ladder to stage 4, so a subsequent `change-code` flags as `change-code-after-close` even though no clean close ever happened. That contradicts the documented semantics ("change-code after a **clean** close/merge", both in the code comment and atlas), and it may inflate the 16-count headline that M4's triage will consume. It's cheap to fix with machinery already in the file, hence FIX-THEN-SHIP rather than REWORK.

## 1. Strengths

- **Dedupe fixed at root, not at the symptom** — `invocationGateEvents` (friction.go:274) collapses the no-validate double-line once, and *both* consumers (`aggregate` and `detectRefusalRetries`) read the deduped stream; the test uses the real two-line captured shape (friction_test.go `TestInvocationGateEventsDedupe`), and the Log records the bypass headline verified unchanged over the real corpus.
- **The inversion-semantics deviation is exactly right.** Task 6's original test list was genuinely self-contradictory under absence semantics (the plan's own cross-issue example contains a close with no prior change-code); resolving to observed-inversion-only, documenting it in `## Revisions`, and marking the plan row with a pointer is the correct precision-over-recall call, executed transparently.
- **ARCH-DRY on verdict recovery** — `isReworkVerdict` (friction.go:551) reuses `judge.ParseVerdict` (which handles the fenced ```` ```verdict ```` block authoritatively) plus `ParseVerdictTrailer` for trailer-only re-closes, rather than growing a third verdict regex.
- **Honesty counters throughout** — `UnattributedPublish` kept out of every per-issue ladder instead of a cross-contaminating `""` bucket (friction.go:434), the ≤20 anomaly listing says "…and N more (full list in --json)", and the footer states the go-run/compound undercounts. This is the plan's ARCH-PURPOSE honesty pillar actually enforced.
- **`buildFrictionReport` (friction.go:661)** is a clean single composition seam for M3's codex walk — invs + marks in, report out, IO stays in `RunFrictionReport`.

## 2. Critical findings

None.

## 3. Important findings

- **friction.go:474–485 — a refused/failed close or merge still raises the firing-order ladder, producing false `change-code-after-close` anomalies.** The ladder raises `maxStage` for *any* `close`/`milestone-close`/`merge`/`push` invocation. But a close refused by a gate (no-actual, plan-check, atlas…) or a merge refused by the publish gate demonstrably did **not** cross the boundary — and "close refused → do more work → change-code" is a plausible real recovery sequence, so some of the reported 16 anomalies may be this false-positive class. The documented contract (code comment friction.go:403, atlas/workflow/sdlc-binary.md "change-code after a **clean** close/merge") promises clean-close semantics the code doesn't implement. **Fix sketch:** compute `invocationGateEvents(inv)` in the ladder loop (it's already in this file and cheap) and skip the `maxStage` raise when the invocation carries a `GateRefusal` with no accompanying `GateBypass` (the compound `close || close --no-atlas` case must still count as completed — the existing `TestFrictionReportRenderAndJSON` sequence is a legit anomaly precisely because the retry bypassed through). Note the residual: non-gate failures (dirty tree, no claim) have no refusal signature; the transcript's `tool_result.is_error` field — currently not captured on `rec`/`SdlcInvocation` (session.go:131) — would cover those, and M3 touches the scanner anyway. Add the regression test: refused-only close → `change-code` → **no** anomaly. Re-run the corpus and update the Log headline if the 16 moves. (ARCH-PURPOSE: the detector's stated purpose is precision over recall; this is the one place the diff under-delivers it.)

## 4. Minor findings

- friction.go:327–330 + :705 — `invocationGateEvents` is computed twice per invocation (once in `aggregate`, once in `detectRefusalRetries`, and a fix for the Important above would add a third). Compute once in `buildFrictionReport` and pass down (ARCH-DRY-lite; cost is trivial at ~1.2k invocations, so consolidation is about one-source-of-classification, not speed).
- friction.go:349–357 — refusal→retry pairing is unbounded in time (a "retry" days later in the same transcript pairs) and ignores `--milestone`, so an M1 mclose refusal can pair with a clean M2 mclose. Both err toward over-counting "resolved"; worth a one-line footnote or a `gapBoundary` bound.
- Refuse→refuse→satisfy chains yield two records with the first `Resolved:false` — per-record accurate and matches the doc comment, but the aggregate "resolved" rate is mildly understated; fine as-is, just know it when reading M4 numbers.
- source.go `KindFileEdit` widens the injection-source `Kind` enum with a non-injection value; the comment covers it, but if a third non-injection mark kind appears, split the type.
- Plan Core Concepts table names the stage table `workflowOrder`; code says `workflowStage`. Trivial staleness.

## 5. Test coverage notes

Coverage is strong and real-fixture-driven: dedupe uses the actual two-line no-validate shape, the ladder table covers all three legal loops + cross-issue interleave + `--help` exclusion, merge attribution covers both attributed and unattributed arms, skill-late covers `.md` exclusion and correct-order, and the render/JSON test pins the report shape end-to-end. Gaps: (a) the refused-close ladder case above — the *kind of bug this diff shipped* and currently untestable-because-wrong; (b) the skill-late segment-gap reset (friction.go:518) has no test — only the issue-change reset path is implied; (c) a close with empty/unlinked output silently reads as "no REWORK observed" — behavior is intentional and honestly labeled, but a test documenting it would pin the choice.

## 6. Architectural notes for upcoming work

- The M1 watch item stands and has grown: `scanTranscript` and `parseEvents` are now two parallel scan-and-link walkers over `rec`, and `scanTranscript` just gained marks capture. Extract the shared scan/link core before M3 adds the codex sibling, or the three will drift.
- When M3 touches the scanner, capture `tool_result.is_error` onto `SdlcInvocation` — it hardens both detectors (failed invocations shouldn't raise ladders or count as satisfying retries) for one field.
- M3's codex walk should state whether codex rollouts can yield `ActivityMark`s (file edits / skill loads); if not, the skill-late arm is Claude-only and the report footer should say so.
- ARCH-DRY: pass (dedupe shared, judge parser reused, `classifyToolUse` as the one match table; `workflowStage` hand-encodes AGENTS.md prose, which has no machine source — acceptable and cited). ARCH-PURE: pass (all detectors pure over parsed inputs; IO confined to `RunFrictionReport`/enumerate; tests need no mocks). ARCH-PURPOSE: pass at this boundary — both promised detectors shipped with limits stated, and the codex deferral is the planned M3 milestone, not a dodged purpose — with the one flagged exception folded into the Important above (the precision purpose vs the refused-close gap).

## 7. Plan revision recommendations

The plan matches the code via the M2 Revisions entry — nothing stops claiming what isn't delivered. Two small follow-ons when the Important lands: append a line to the M2 Revisions entry noting the clean-close semantics tightening (and the corrected anomaly count if the 16 moves), and fix the `workflowOrder`→`workflowStage` name in the Core Concepts table while in there.
