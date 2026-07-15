# Boundary Review — ariadne#172 (whole-issue close)

| field | value |
|-------|-------|
| issue | 172 — sdlc painpoint audit |
| repo | ariadne |
| issue file | workshop/issues/000172-sdlc-painpoint-audit.md |
| boundary | whole-issue close |
| milestone | — |
| window | 164d2610921c1eb2c6c74f124f1be30e5670eac9..HEAD |
| command | sdlc close --issue 172 |
| reviewer | claude |
| timestamp | 2026-07-14T14:32:37-07:00 |
| verdict | SHIP |

## Review

The live run reproduces the issue's documented headline exactly (codex 43 / claude 37 bypasses, exactly 40 fork-replays skipped, 11 + 2 firing-order anomalies, 158 refusals → 149 resolved → only 8 via bypass). All verification done — composing the review.

```verdict
verdict: SHIP
confidence: high
```

This boundary closes #172 whole, and the issue delivers its full stated purpose with the claims verified independently: the T1 instrument (`process-manual --friction-report`) is built across all three milestones (Claude + codex, all 12 spine gates, refusal→retry + firing-order), `go test ./cmd/sdlc/...` and `go vet` are green, the Python golden consumer (`test_normalize.py`) passes, and a live run over the real corpus reproduces every headline number in the Findings — codex 43 / claude 37, exactly the spec's 40 fork-replays skipped, 11 change-code-after-close + 2 skill-late, 158 refusals with 149 resolved and only 8 via bypass. T2's per-gate triage table and T3's coverage-gap read are recorded in `## Findings` with the three follow-ups actually filed (#174, #175, #176 exist in `workshop/issues/`). Every prior milestone's FIX-THEN-SHIP Importants were fixed in-window (walk-level integration test, Python golden consumer, refused-close ladder semantics, per-command observability), the plan's checkboxes are all ticked with honest Revisions entries covering each deviation, and the atlas + helptext document the new surface. Nothing blocks; the findings below are minor notes for future work.

**1. Strengths**

- **The instrument's claims are reproducible, not narrated.** Running `--friction-report --json` live matched the Log's M4 numbers to the digit — for a measurement tool whose entire point is replacing a contaminated grep, this is the property that matters, and it holds (`cmd/sdlc/internal/processmanual/friction.go:1004`).
- **Anti-contamination is pinned by real-fixture tests on both agents**: `TestAggregateAntiContamination` and the codex golden's `root.jsonl` third call both plant cat-n/source noise inside real command output and assert it classifies to none — the exact bug class this diff could ship.
- **The M4 instrument correction is precision-over-recall done right**: the bare-`#N` fallback removal (`friction.go:44` `issueArgRE`, `--issue`-only now) was caught from a live false anomaly, fixed at root, its rationale documented in the regex comment, and the honest cost (unattributed publishes 52→60) stated rather than hidden.
- **Cross-language DRY is now enforced on both sides** (ARCH-DRY): `testdata/codex-golden/` is consumed by the Go reader (`TestCodexGoldenDecisions`) *and* Python introspect (`test_normalize.py::test_codex_golden_shared_fixture`, with a downstream-safe skip) — the shared keep/skip judgment can no longer drift silently.
- **The drift guard has teeth**: `cmd/sdlc/gates_test.go:20` diffs the catalog against each spine command's live-registered cobra flags, so a new `--no-<gate>` can't land without the audit noticing.

**2. Critical findings** — none.

**3. Important findings** — none. (Each milestone's Importants were verified fixed in-window: `TestRunFrictionReportTwoAgentWalk` covers the two-corpus seam including the one-corpus-missing contract; the refused/Failed-close ladder skip exists at `friction.go:523` with its regression tests; observability is keyed per (command, flag) with `TestObservabilityPerCommand`.)

**4. Minor findings**

- `detectFiringOrder` (`friction.go:497`): the change-code anomaly check runs *before* the Failed/refused skip, so a change-code that itself failed after a clean close still flags. Defensible (the attempt is the signal), but it's asymmetric with the close-side semantics — worth a comment or a deliberate test.
- `transcripts_scanned` asymmetry (Claude counts enumerated files including unreadable; codex counts only processed-and-included) — noted in the M3 review, not folded; harmless but the header number is slightly soft.
- `--json` without `--friction-report` is silently ignored (`processmanual.go`); a one-line refusal would match the repo's misinvocation-must-not-look-real posture.
- `detectRefusalRetries`: a merge/push refusal with empty IssueID pairs with *any* later same-verb invocation in the transcript — documented as flag-omitted best-effort in the footnote, fine, just know it when reading M4 numbers.
- `repoLabelFromPath` on a cwd of exactly `/Users/x/workspace` would label the repo "workspace" — an edge no real rollout hits.

**5. Test coverage notes**

Coverage is strong and real-fixture-driven end-to-end: classifier positives per grammar + the load-bearing rejection cases, per-invocation dedupe on the real no-validate double-line, the full firing-order ladder table (legal loops, refused/Failed skips, `--help` exclusion, cross-issue interleave), merge attribution both arms, skill-late three arms, the two-agent walk integration test, the cross-language golden with a Python consumer, and render/JSON shape. The only untested spot I found is the change-code-after-close-while-itself-failed asymmetry above — cheap to pin if the semantics are intended.

**6. Architectural notes for upcoming work**

- **ARCH-DRY: pass.** `GateCatalog` single-sources classifier, drift guard, `SpineVerbs`, and report rows; `forEachRec` is the shared Claude scan core (the M1/M2 "parallel walkers drift" watch item was actually resolved, not just noted); `judge.ParseVerdict` reused for REWORK recovery; the codex format lives once in the atlas spec with the golden enforcing both consumers. Residual: scanner buffer setup appears three times (`forEachRec`, `codexMeta`, `parseCodexInvocations`) — a shared constant would do.
- **ARCH-PURE: pass.** All detectors (`classifyOutputLine`, `aggregate`, `detectRefusalRetries`, `detectFiringOrder`, `codexMeta`, `parseCodexInvocations`) are pure over bytes/parsed inputs, tested without mocks; `RunFrictionReport` is the single thin IO seam with injectable roots.
- **ARCH-PURPOSE: pass.** The issue's Done-when has three parts and all three are delivered — instrument (both agents, per the Spec's explicit both-agents demand), T2 triage with per-gate verdicts and evidence, T3 coverage-gap read with follow-ups filed as real issues. Shadow-sweep of the single-source claims: both spec consumers (Go + Python) derive and are golden-pinned; the catalog's consumers all derive; no hand-maintained restatement remains. The observability-honesty sub-goal — the pillar earlier reviews flagged — is now correct per (command, flag) and stated in the report footer.
- For #170 (the umbrella this feeds): the follow-ups' triage evidence is machine-readable (`--json` carries per-agent `RefusalRetry` records), so #174–#176 can cite exact counts rather than re-running analysis.

**7. Plan revision recommendations** — none. The plan's Revisions entries (M1 through M4) faithfully cover every deviation I could check against the code, including the `parseCodexEvents`→`parseCodexInvocations` mapping and the `workflowOrder`→`workflowStage` rename; the checkbox state matches what shipped.
