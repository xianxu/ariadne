# Boundary Review — ariadne#201 (whole-issue close)

| field | value |
|-------|-------|
| issue | 201 — Boundary-review artifacts persist the harness transcript, not just the review |
| repo | ariadne |
| issue file | workshop/issues/000201-boundary-review-artifacts-capture-the-harness-trust-dialog-preamble.md |
| boundary | whole-issue close |
| milestone | — |
| window | 4a2ebbd6708dc5cc0c5843521303da0717e4bfba..d548b669c9eb088a69c22d920c87d79234ff9c4b |
| command | sdlc close --issue 201 |
| reviewer | codex |
| timestamp | 2026-08-25T12:56:00-07:00 |
| verdict | REWORK |

## Review

```verdict
verdict: REWORK
confidence: high
```

The implementation correctly separates process streams and keeps diagnostics out of review parsing and persistence. The full suite passes. The boundary is blocked by the plan’s incorrect PURE classification of a side-effecting function, as the Core concepts contract explicitly makes that Critical. The live-agent conformance test also does not actually verify that the requested semantic response appears on stdout.

1. Strengths

- [`Run`](/Users/xianxu/workspace/ariadne/cmd/sdlc/internal/judge/dispatch.go:94) captures stdout and stderr independently while preserving PID and exit behavior.
- [`classifyRunResult`](/Users/xianxu/workspace/ariadne/cmd/sdlc/internal/judge/dispatch.go:216) is the one process-result transition used by synchronous and heartbeat paths, satisfying ARCH-DRY.
- The boundary regression at [closereview_test.go:299](/Users/xianxu/workspace/ariadne/cmd/sdlc/closereview_test.go:299) verifies verdict parsing, structured findings, terminal routing, `reviewResult.Output`, and the persisted sidecar together.
- Launch failures, non-zero exits, nil-diagnostic-sink behavior, heartbeat routing, and a real OS subprocess are covered.
- Atlas and command help accurately document the artifact split. No README update is required because no new command, flag, keybinding, or usage step was introduced.

2. Critical findings

- [Plan:31](/Users/xianxu/workspace/ariadne/workshop/plans/000201-boundary-review-artifacts-capture-the-harness-trust-dialog-preamble-plan.md:31), [dispatch.go:216](/Users/xianxu/workspace/ariadne/cmd/sdlc/internal/judge/dispatch.go:216) — **ARCH-PURE: `classifyRunResult` is incorrectly declared PURE.** It writes to an injected `io.Writer` and reads process/environment metadata on launch errors. Its tests exercise it through fake process and writer seams. Promote it to INTEGRATION in the Core concepts table and revise the ARCH-PURE narrative, or split pure classification from diagnostic/environment IO. Add a dated `## Revisions` entry because the current plan contradicts the delivered code.

3. Important findings

- [live_stream_conformance_test.go:45](/Users/xianxu/workspace/ariadne/cmd/sdlc/internal/judge/live_stream_conformance_test.go:45) — **ARCH-MOCK / ARCH-PURPOSE: the live conformance check accepts any non-empty stdout.** A CLI could emit a banner on stdout and put `STREAM_OK`—the semantic response—on stderr; the test would pass while production discards the review. Assert that trimmed stdout equals `STREAM_OK` and that stderr does not contain `STREAM_OK`. This pins the external contract the issue depends on.

4. Minor findings

None.

5. Test coverage notes

Passed:

- `go test ./cmd/sdlc/internal/judge -count=1`
- `go test ./cmd/sdlc -count=1`
- `go test ./... -count=1`
- `git diff --check 4a2ebbd..d548b669`

The opt-in live Claude/Codex/Gemini conformance test was not run, consistent with the plan’s explicit opt-in requirement. The fake boundary regression would fail if diagnostic stderr were rejoined into semantic output or persisted into the review sidecar.

6. Architectural notes for upcoming work

- **ARCH-DRY: pass.** Both dispatch modes converge on one completion helper.
- **ARCH-PURE: flag.** The code is a reasonably thin IO boundary, but the plan misclassifies that boundary as pure.
- **ARCH-PURPOSE: flag.** The production behavior fulfills the main artifact-separation purpose, but its promised real-adapter drift check does not prove semantic output uses stdout.
- **ARCH-MOCK: flag.** Fakes and the real `sh` seam are good; the external-agent conformance assertion must validate the actual channel contract.

7. Plan revision recommendations

Append a dated `## Revisions` entry recording:

- `classifyRunResult` is an INTEGRATION boundary because it writes diagnostics and reads launch-error context; update the Core concepts and ARCH-PURE text accordingly.
- The live conformance acceptance criterion changes from “stdout is non-empty” to “stdout contains exactly the requested semantic sentinel and stderr does not contain it.”

```findings
findings:
  - id: new
    severity: Critical
    family: core-concept-classification
    title: |
      classifyRunResult is declared PURE despite performing IO
    detail: |
      The plan lists classifyRunResult under Pure entities, but the implementation writes to an injected io.Writer and reads launch-error environment context. Promote it to INTEGRATION or split pure classification from the IO shell, and append a plan revision.
  - id: new
    severity: Important
    family: external-contract-conformance
    title: |
      The live agent stream check does not verify the semantic channel
    detail: |
      The test requests STREAM_OK but accepts any non-empty stdout. Assert trimmed stdout equals STREAM_OK and stderr does not contain it so Claude, Codex, or Gemini stream drift cannot pass unnoticed.
```

---

## Re-review — 2026-08-25T13:11:37-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 201 — Boundary-review artifacts persist the harness transcript, not just the review |
| repo | ariadne |
| issue file | workshop/issues/000201-boundary-review-artifacts-capture-the-harness-trust-dialog-preamble.md |
| boundary | whole-issue close |
| milestone | — |
| window | 4a2ebbd6708dc5cc0c5843521303da0717e4bfba..9b0ab3ff70f4b57ab6b0fcfcbcec627ab6ba286d |
| command | sdlc close --issue 201 |
| reviewer | codex |
| timestamp | 2026-08-25T13:11:37-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

The stream separation fulfills the issue’s purpose: semantic stdout alone reaches parsing, display, and durable sidecars, while diagnostics remain on stderr. Both prior findings are addressed and no new blocking findings were found.

1. Strengths

- `ProcessOutput` preserves stdout/stderr independently through the replaceable subprocess seam ([dispatch.go:76](/Users/xianxu/workspace/ariadne/cmd/sdlc/internal/judge/dispatch.go:76)).
- Both synchronous and heartbeat paths converge on `classifyRunResult`, preserving exit behavior without duplicated routing logic ([dispatch.go:175](/Users/xianxu/workspace/ariadne/cmd/sdlc/internal/judge/dispatch.go:175)).
- The boundary regression verifies verdict/finding parsing, terminal routing, `reviewResult.Output`, and sidecar persistence together ([closereview_test.go:299](/Users/xianxu/workspace/ariadne/cmd/sdlc/closereview_test.go:299)).
- The plan now correctly classifies `classifyRunResult` as INTEGRATION and records the correction in `## Revisions`.
- Atlas and help text accurately describe the review-sidecar/gate-ledger split. No public command, flag, or configuration surface requires a README update.

2. Critical findings

None.

3. Important findings

None.

4. Minor findings

None.

5. Test coverage notes

Fresh targeted verification passed:

- `go test ./cmd/sdlc/internal/judge -run 'TestDispatch|TestRun_RealSubprocess|TestLiveAgentStreamConformance' -count=1`
- `go test ./cmd/sdlc -run 'TestDispatchBoundaryReview.*Semantic|TestClose.*Sidecar' -count=1`
- `git diff --check 4a2ebbd..9b0ab3f`

The opt-in live-agent test skipped normally as designed. Its exact `STREAM_OK` assertion would fail under the former non-empty-output condition.

A full-suite attempt was not usable as evidence because the review sandbox forbids `.git/sdlc.lock`; an isolated clone then lacked ignored base-layer fixtures used by unrelated repository tests.

6. Architectural notes for upcoming work

- **ARCH-DRY: pass.** One completion helper owns diagnostic routing and exit classification.
- **ARCH-PURE: pass.** `ProcessOutput` is a pure value; subprocess capture and diagnostic writing are explicitly integration boundaries.
- **ARCH-PURPOSE: pass.** All review consumers—verdict parsing, findings parsing, terminal output, and sidecar persistence—derive from semantic stdout.
- **ARCH-MOCK: pass.** Production and fake flows share the `Run` seam, with a real subprocess test and opt-in live CLI conformance check.

7. Plan revision recommendations

None.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      The Core concepts table and ARCH-PURE narrative now classify classifyRunResult as an INTEGRATION boundary, with a dated revision recording the correction.
  - id: BR-2
    disposition: addressed
    note: |
      The live conformance test now requires trimmed stdout to equal STREAM_OK exactly; the plan documents why stderr may legitimately echo the prompt sentinel.
```
