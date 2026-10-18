# Boundary Review — ariadne#409 (milestone M1)

| field | value |
|-------|-------|
| issue | 409 — Dynamic review policy: choose the reviewer model by author family and change kind (cross-family by default), extending #310 |
| repo | ariadne |
| issue file | workshop/issues/000409-dynamic-review-policy-choose-the-reviewer-model-by-author-family-and-change-kind-cross-family-by-default-extending-310.md |
| boundary | milestone M1 |
| milestone | M1 |
| window | d3ebb55260c6647fbc70b423efed20cd8c734c5b..7fe6c39b917990a37c04efa94aa6b96e954646c9 |
| command | sdlc milestone-close --issue 409 --milestone M1 |
| reviewer | claude |
| isolation | fenced |
| timestamp | 2026-10-18T05:35:52-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

Now the plan.

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

M1 delivers what it set out to: a pure, well-tested core. It adds `Family`/`FamilyOf`, the two change kinds (design-heavy, bounded), the family-aware `other` mode, one default constant (`DefaultReviewerMode`, still `same`), and the pure `ReviewPlan`, `Unavailable` and `LocalAvailability`. `go vet` and `go test ./internal/judge/` are clean, and so are `TestLoadReviewerPolicy` and `TestResolveReviewer*` in `cmd/sdlc`. Five Important findings stand between this and SHIP, and none needs a redesign:

1. A repo's `reviewer_order` can quietly put the change back with the author's own family, which the issue forbids.
2. `Unavailable` can never fire for codex or gemini, because their output is not parsed as a stream. So the quota fallback cannot work for the agent that is the default reviewer of Claude's code.
3. The close help text is now wrong: M1 changed who reviews gemini's work.
4. The config accepts `reviewer_models`, but no code reads it yet, and no milestone is assigned to wire it.
5. The plan's Core concepts table no longer matches the code.

### 1. Strengths
- `Reviewer()` now takes its answer from `ReviewPlan` (`agent_resolver.go:85-88`), so the static and dynamic paths share one source. Claude ↔ codex under `other` still behaves as #310 promised, and `TestResolveReviewerCLI_Policy` pins it.
- `Unavailable` (`family.go:128-149`) is conservative in the right way. Output that is not a stream, any assistant or tool event (`Began`), or any verdict means "it started", so the loop never switches. `TestUnavailable` covers the "quota after the review began" case.
- `TestReviewPlan` has an independent invariant check: under `other`, no candidate may share an author's family. It tests the rule itself, not just a list of expected outputs.
- `LocalAvailability` takes `lookPath` and `goos` as inputs, so it can be tested without touching the system.
- Choosing codex, claude, gemini as the bounded order (instead of the approved codex, gemini, claude) to keep #310's floor is a sound call, and the Log records it.

### 2. Critical findings
None.

### 3. Important findings
- **I-1: `reviewer_order` can silently send review back to the author's family** (`agent_resolver.go:85-91`, `Validate` at `:120-129`).
  - **What happens:** with `{"boundary_reviewer":"other","reviewer_order":{"bounded":["claude"]}}` and a claude main agent, `ReviewPlan` returns an empty plan. `Reviewer` then falls through to `return main`, so claude reviews claude with no notice. An empty list (`"bounded": []`) does the same. This breaks the Spec's rule that a review never falls back silently to the author's family. It is reachable today, because close and milestone-close already call `Reviewer`. Plain `Validate` accepts the config; only main-agent detection needs to resolve to claude.
  - **Fix:** in `Validate`, reject an order that is empty, has duplicates, or covers fewer than two families. Separately, `Reviewer` should not return `main` under `other`; until M2's refusal path exists, return the first entry of the order that is not the main agent's family, or an error. (ARCH-PURPOSE)
- **I-2: `Unavailable` is inert for codex and gemini** (`dispatch.go:629-633`, `family.go:131-134`).
  - **Cause:** only claude's output goes through `ReadStream`; for codex and gemini, `Stream` is always false. So `Unavailable` returns false for every codex or gemini failure except a failed launch. Codex is the default bounded reviewer for Claude-authored code, so a spent codex quota will always halt the review instead of falling back. That is the main case the issue's fallback exists for.
  - **Gap against the plan:** M1 promised "tests for the quota and auth texts each CLI prints". `TestUnavailable` only uses claude stream-json fixtures, and the plan never states this consequence.
  - **Fix:** in M2, either pass codex `exec --json` and parse its events into `Began`, or define a narrow, tested pre-review signature for each CLI. At minimum, add a plan Revision that states the limit and a test that pins it. (ARCH-PURPOSE, ARCH-ORDER)
- **I-3: close help is stale after a behaviour change made in M1** (`cmd/sdlc/helptext/close.md:417`).
  - It still says "gemini is reviewed by claude", but the code now picks codex (see the changed `judge_test.go:1336`).
  - The new `reviewer_order` and `reviewer_models` keys are missing from the REVIEWER section.
  - The atlas paragraph at `sdlc-binary.md:1191-1192` still describes `other` as "claude and codex review each other" and nothing more.
  - The fix is a few lines; it should not wait for M2.
- **I-4: `reviewer_models` is loaded and checked, but nothing uses it** (`reviewconfig.go:24`, `agentArgs` in `dispatch.go:219-258`).
  - An operator can set a model and get the CLI's default with no warning.
  - The plan's Integration section says the model is "passed as `--model`/`-m` in `agentArgs`", but neither M1 nor M2 lists that work.
  - **Fix:** add it to M2's checklist explicitly, or warn when `Models` is set until the wiring lands. (ARCH-PURPOSE)
- **I-5: the plan's Core concepts table doesn't match the code.** The table rule calls any table/code contradiction Critical; I rated it Important because the fix is a plan Revision, not code. The mismatches:
  - The table puts `ReviewPlan` in `reviewplan.go` and `Unavailable` in `apifailure.go`; both live in `family.go`.
  - It lists `FamilyOf(agent, model)`; the code has `FamilyOf(agent)`. Decision 2's "qoder resolves through its underlying model" was dropped, and the Revisions don't say so.
  - It lists `ReviewPlan(authors, kind, policy)`; the code is `ReviewPlan(policy, main, authors, kind, available)`.
  - `LocalAvailability` and `AgentRun.Began` are not in the table at all.

### 4. Minor findings
- `trimReason` (`family.go:158`) has no caller.
- `Validate` accepts an order that names the same agent twice, which produces duplicate candidates.
- The `reviewer_models` check allows a value starting with `-`. Even passed as its own argument, a CLI could read it as a flag, so reject it now before M2 wires it in (ARCH-SECURE; review.json is protected by the local-gates label, so the risk is low).
- `Reviewer()` always uses `KindBounded`. That's fine for M1, but M2 must pass the real kind at full-flow boundaries.
- The test row "the default mode is the constant" builds its expected value with a map lookup, which is hard to read. If `DefaultReviewerMode` were `explicit`, the lookup would yield `""`.

### 5. Test coverage notes
The tests check real logic and do no IO. Missing:
- A case where `reviewer_order` lists only the author's family (would catch I-1).
- Captured quota and auth output from codex and gemini (I-2).
- A test that `ReadStream` sets `Began` on a `user` event without any assistant event.
- M2 still needs to test the dispatch loop with fake runners.

### 6. Architectural notes
- **ARCH-DRY: pass.** `Reviewer` uses `ReviewPlan`, `unavailableRE` builds on `apiFailureRE`, and `firstLine`/`lineAt` are reused.
- **ARCH-PURE: pass.** All of `family.go` is pure, and availability is passed in as a function.
- **ARCH-PURPOSE: flag** (I-1, I-2, I-4).
- **ARCH-MOCK: pass for M1.** No new external calls; `lookPath` is passed in. The dispatch-loop fakes are M2's job.
- **ARCH-CONSTRAINTS: pass.** The candidate list is bounded by the three known agents.
- **ARCH-SECURE: pass, with a minor note.** review.json is local-gates. Author-trailer handling (unioning in the closer's `Detect` family) is M2 and must happen there.
- **ARCH-ORDER: pass** for the conservative "began" rule, which never turns an unconfirmed run into "unavailable". But `Began` only exists for claude (I-2).
- **ARCH-FUNERAL: pass.** M1 creates nothing on disk.

### 7. Plan revision recommendations
- Fix the Core concepts table:
  - File paths: everything is in `family.go`.
  - Signatures: `FamilyOf(agent)`, `ReviewPlan(policy, main, authors, kind, available)`, `Unavailable(run, stderr []byte, nonZeroExit, launchErr)`.
  - Add `LocalAvailability` and `AgentRun.Began`.
  - Say whether qoder's model-based family is dropped or deferred.
- State that `Unavailable` only applies to claude's stream until codex and gemini output is parsed, and say which milestone fixes that.
- Add "pass `Candidate.Model` through `agentArgs`" to M2's checklist.
- Record a `reviewer_order` validation rule (no empty lists, no duplicates, at least two families) as the guard against silent same-family review.

```findings
findings:
  - id: new
    severity: Important
    kind: defect
    family: no-silent-same-family-review
    title: |
      reviewer_order lacking another family makes Reviewer() silently return the author's own agent
    detail: |
      Mode other with reviewer_order {"bounded":["claude"]} (or an empty list) and a claude main agent makes ReviewPlan return an empty plan, and Reviewer() at agent_resolver.go:85-91 falls through to `return main`, so claude reviews claude with no notice. Validate (agent_resolver.go:120-129) accepts the config. Reject empty, duplicate or single-family orders in Validate, and never return main under other. ARCH-PURPOSE.
  - id: new
    severity: Important
    kind: defect
    family: unavailable-evidence-per-cli
    title: |
      Unavailable can never classify a codex or gemini quota or auth failure; the fallback is inert for the default cross-family reviewer
    detail: |
      dispatch.go:629-633 runs ReadStream only for claude, so codex and gemini runs have Stream=false and Unavailable (family.go:131-134) returns false except on a launch error. M1 promised tests for the quota and auth texts each CLI prints; TestUnavailable has only claude stream-json fixtures. Parse codex exec --json (and gemini's output) into Began, or state and test the limit in a plan Revision. ARCH-PURPOSE, ARCH-ORDER.
  - id: new
    severity: Important
    kind: defect
    family: docs-track-behavior-change
    title: |
      close help still says gemini is reviewed by claude; M1 changed it to codex, and the new review.json keys are missing
    detail: |
      cmd/sdlc/helptext/close.md:417 describes the old `other` mapping. The changed judge_test.go:1336 now expects codex. The REVIEWER section omits reviewer_order and reviewer_models, and atlas sdlc-binary.md:1191 still describes other as claude and codex only. A cheap fix, and it should not wait for M2.
  - id: new
    severity: Important
    kind: defect
    family: config-key-without-consumer
    title: |
      reviewer_models is accepted and validated but has no consumer, and no milestone owns the agentArgs wiring
    detail: |
      reviewconfig.go loads ReviewerModels into Policy.Models and ReviewPlan copies it into Candidate.Model, but agentArgs (dispatch.go:219-258) never passes a model, so a configured model is silently ignored. The plan's Integration section says it is passed as --model/-m, but neither M1 nor M2 lists that work. Add it to M2, or warn while it is unwired.
  - id: new
    severity: Important
    kind: defect
    family: plan-table-matches-code
    title: |
      Plan Core concepts table contradicts the code: file paths, FamilyOf and ReviewPlan signatures, missing entities
    detail: |
      The table puts ReviewPlan in reviewplan.go and Unavailable in apifailure.go (both are in family.go), lists FamilyOf(agent, model) with qoder resolved through its model (the code is FamilyOf(agent) with no qoder handling), lists ReviewPlan(authors, kind, policy) (the code adds main and available), and omits LocalAvailability and AgentRun.Began. Add a Revisions entry.
  - id: new
    severity: Minor
    kind: defect
    family: dead-code
    title: |
      trimReason in family.go has no caller
  - id: new
    severity: Minor
    kind: defect
    family: parse-config-to-typed-value
    title: |
      reviewer_models allows a leading '-', and reviewer_order allows duplicate agents
    detail: |
      Reject a flag-like model value before M2 passes it to the CLI (ARCH-SECURE), and reject an agent listed twice in one order.
```

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"c4bc4de4-13e2-4ce2-b0f5-a06f4142d8e3","repository":"github.com/xianxu/ariadne","issue_id":"000409","recipe":"milestone-review","round":1,"attempt":1,"agent":"claude","root":"/Users/xianxu/Library/Caches/sdlc/review-checkouts/76308b093647e048","reviewed_head":"7fe6c39b917990a37c04efa94aa6b96e954646c9","slot":"ariadne:2","requested_at":"2026-10-18T12:32:24.753047Z"},"phase":"finished","pid":18367,"started_at":"2026-10-18T12:32:24.914333Z","finished_at":"2026-10-18T12:35:51.678033Z","session_id":"0b2d6a50-30f4-45a0-a498-d2d2c3bafad4","outcome":"returned"}] -->
