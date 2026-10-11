# Boundary Review — ariadne#189 (whole-issue close)

| field | value |
|-------|-------|
| issue | 189 — sdlc judge plan-quality renders an empty issue — no --issue path populates IssueContent/PlanContent |
| repo | ariadne |
| issue file | workshop/issues/000189-sdlc-judge-plan-quality-renders-an-empty-issue-no-issue-path-populates-issuecontent-plancontent.md |
| boundary | whole-issue close |
| milestone | — |
| window | 8b9052040674db7b431ad3db3844837e0c561366..f4fd6f8e55fd687d6d6dced68fcd2e5b11d38e9f |
| command | sdlc close --issue 189 |
| reviewer | claude |
| timestamp | 2026-10-10T17:40:39-07:00 |
| verdict | SHIP |

## Review

Diff inspected; now checking tree-wide plan-path builders and the tests.

```verdict
verdict: SHIP
confidence: high
```

The diff fixes the bug the issue describes. `sdlc judge plan-quality --issue N --dry-run` now renders the issue, the durable plan and the ledger's prior findings. It builds the prompt through the same `planQualityPromptInput` that change-code's gate uses, so the two prompts cannot drift apart. A live run refuses before dispatching and points at `sdlc change-code`, so the no-ledger question is settled in code and tested. Durable-plan paths are now built in one place, `planArtifactPath`. I found no hand-built path left anywhere in the Go tree: `resolve.go:146` is a suffix classifier and `startplan.go:242` is guidance text. The new tests and the `TestJudgeAgentDefault_*` tests pass when run with `-count=1`. The main test would fail without the fix: before it, plan-quality went through the `collectDiff` path, so `IssueContent` was empty and the `SPEC-SENTINEL` assertion fails. Nothing blocks shipping. There are two Minor notes.

1. **Strengths**
   - `cmd/sdlc/changecode.go:598` `planQualityPromptInput` is the only place the plan-quality `PromptInput` is built. Both verbs call it (`changecode.go:643`, `judge.go:300`), which meets the ARCH-DRY Done-when.
   - `judge.go:277-283`: both refusals (no `--issue`, live run) happen before any file is read or any agent is dispatched. `TestJudgePlanQuality_LiveRunRefuses` checks that the stub agent is called zero times.
   - `judge_issue_test.go`: both states of the plan clause are tested (plan plus ledger, and neither). Each test asserts the judge's prompt is byte-equal to change-code's dry-run prompt and that the ledger bytes are unchanged, which pins the read-only claim.
   - The sweep of plan-path builders covered all of them: `changecode.go:150`, `changecode.go:284`, `readOptionalPlanFile` (which also serves `closeflow.go:79` and `changecode_flow.go:107`), and `reviewwindow.go:155`.
   - Changing `cli_signal_unix_test.go` to use `judge dry` is the right call: the test needs a category that still dispatches live, and that's what it exercises.

2. **Critical:** none.

3. **Important:** none.

4. **Minor**
   - **The "exactly" claim has an exception.** `atlas/workflow/sdlc-binary.md:1104` says the dry run "renders exactly the prompt the gate sends", and `helptext/judge.md:23` says the same. In a repo that uses the issue tracker, change-code's gate is given `gateContent`, which has the card mirror refreshed. The judge renders the file as it is on disk. If the mirror is stale, the two prompts differ in the card fields. The issue's Decision section records this, but the user-facing atlas and helptext don't. Add "(card mirror as on disk)" or similar to both.
   - **ARCH-DRY, small.** `planQualityJudgePrompt` (`judge.go:285-299`) repeats change-code's read steps: read the issue file, `captureReviewArtifact` the plan, `readPlanGateLedger`. The paths and the prompt input are shared, but these read steps are not. Reasonable for now. If a third consumer appears, extract a `loadPlanQualityInputs(plansDir, name, issuePath, issueNum)`.

5. **Test coverage**
   - Both plan states, the live-run refusal and the missing-`--issue` refusal are covered.
   - The byte-equality tests feed `runPlanQualityJudge` issue bytes that the test reads itself. They don't go through `runChangeCode`'s read path, so the card-mirror difference above is never exercised. That is acceptable given the Decision section.
   - The Done-when asks for "a golden prompt". The tests use sentinels plus byte-equality with change-code's prompt rather than a golden file. That is at least as strong: an empty render cannot pass.

6. **Architecture**
   - **ARCH-DRY: pass.** One path helper, one issue-ref helper and one prompt-input builder; only the minor read-step overlap above remains.
   - **ARCH-PURE: pass.** The new helpers are pure. `planQualityJudgePrompt` is a thin IO layer around them.
   - **ARCH-PURPOSE: pass.** The purpose of the issue (a real render, shared resolution, the no-ledger question settled) is delivered in full. The estimate-quality exclusion is justified: `judge.IsValid` rejects it, so there is no manual command to fix.

7. **Plan revisions:** none needed. The issue's Plan and Decision match the code.

```findings
findings:
  - id: new
    severity: Minor
    family: doc-claim-overstates-equivalence
    title: |
      atlas and judge helptext say the dry run renders "exactly" change-code's prompt, but card-mirror refresh can differ
    detail: |
      atlas/workflow/sdlc-binary.md:1104 and cmd/sdlc/helptext/judge.md:23 claim an exact match. In a tracker-era repo, change-code's gate gets the refreshed card mirror (gateContent), while the judge reads the file as it is on disk. Qualify the claim in both places.
  - id: new
    severity: Minor
    family: shared-input-loader
    title: |
      planQualityJudgePrompt repeats change-code's issue/plan/ledger read steps
    detail: |
      judge.go:285-299 repeats the ReadFile + captureReviewArtifact + readPlanGateLedger steps that runChangeCode and runPlanQualityJudge perform. The paths and prompt input are shared; the read steps could become one loader if a third consumer appears (ARCH-DRY).
```
