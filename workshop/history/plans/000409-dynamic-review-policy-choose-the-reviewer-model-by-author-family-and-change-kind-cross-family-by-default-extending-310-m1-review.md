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

---

## Re-review — 2026-10-18T05:55:37-07:00 (FIX-THEN-SHIP)

| field | value |
|-------|-------|
| issue | 409 — Dynamic review policy: choose the reviewer model by author family and change kind (cross-family by default), extending #310 |
| repo | ariadne |
| issue file | workshop/issues/000409-dynamic-review-policy-choose-the-reviewer-model-by-author-family-and-change-kind-cross-family-by-default-extending-310.md |
| boundary | milestone M1 |
| milestone | M1 |
| window | d3ebb55260c6647fbc70b423efed20cd8c734c5b..948d85dd7dce63dd71cc6514dbf37b8b6f5dcd21 |
| command | sdlc milestone-close --issue 409 --milestone M1 |
| reviewer | claude |
| isolation | fenced |
| timestamp | 2026-10-18T05:55:37-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

Checking the prior-round fixes: running the tests and reverting the Reviewer() fallback to see if any test catches it.

```verdict
verdict: FIX-THEN-SHIP
confidence: medium
```

M1's pure core is solid. `family.go` holds `FamilyOf`, `ChangeKind`, `ReviewPlan`, `Unavailable` and `LocalAvailability` with no IO. `Validate` now rejects empty, single-family and duplicate orders and flag-like model values. Model flags reach each CLI in `agentArgs`. Help, atlas and the plan Revisions were all corrected. The judge and `cmd/sdlc` test suites pass at head.

One prior finding fails the revert check, so this is not SHIP yet. With `Model: policy.Models[reviewer]` removed from `cmd/sdlc/milestoneclose.go:874`, every test stays green. That line is the entire BR-4 symptom ("a configured model is silently ignored"), so BR-4 stays `not-addressed`. One test of `boundaryReviewOptionsFor` that asserts `opts.Model` fixes it. I did the same revert test on the `Reviewer()` default-order fallback for BR-1, and it also stays green. That one doesn't block: `Validate` makes the branch unreachable for any policy loaded from the file, and the `Validate` tests do pin the real fix.

1. **Strengths**
   - `ReviewPlan` (`cmd/sdlc/internal/judge/family.go:88`) is pure and checks `availability` before launch. Its table test also checks an invariant on every case: under `other`, no candidate shares an author's family. That is stated separately from the expected lists, not copied from them.
   - `Unavailable` (`family.go:141`) treats unclear evidence as "the review started": an unparseable stream, any codex activity line, or a 400 all count as started. It reads codex's cause only from `ERROR:` lines because codex echoes the prompt to stderr, and a test pins that case.
   - `LocalAvailability` (`family.go:211`) takes `lookPath` as a parameter, so its tests need no IO.
   - It reuses `apiFailureRE`, `lineAt`, `firstLine` and `valueOrDefault` instead of copying them (ARCH-DRY).
   - `Validate` (`agent_resolver.go:124-150`) checks the new config keys when the file is loaded, and refuses flag-like model values before they reach an argv (ARCH-SECURE).

2. **Critical:** none.

3. **Important**
   - BR-4 is still open (see below). The wiring at `cmd/sdlc/milestoneclose.go:874` has no test. Fix: a test that loads a policy with `reviewer_models`, calls `boundaryReviewOptionsFor`, and asserts `opts.Model`. It should also cover the `--agent` forced path.

4. **Minor**
   - The close help (`cmd/sdlc/helptext/close.md:415-421`) and the atlas describe behaviour that only lands in M2: the order depends on the change kind, a full-flow close tries claude first, and gemini is a fallback chosen by availability. Today the boundary review goes through `Reviewer()`, which always uses `KindBounded` and ignores availability. So a gemini main agent on a full-flow close gets codex, not the claude the help promises. This is the 2nd finding in `docs-track-behavior-change`; the rule is in the findings block.
   - `TestReviewerOtherNeverSameFamily` passes on the pre-fix code. None of its orders leaves `ReviewPlan` empty, so the default-order fallback is never run.
   - Unclassified edge cases in `Unavailable`:
     - A claude non-stream plain-text error on stdout.
     - `agent == ""`, which `Dispatch` treats as claude but which goes to the gemini branch here.

     Both are conservative (they halt), so they're fine for now.

5. **Test coverage notes**
   - Each test below fails if its fix is reverted:
     - The `Validate` cases for an empty order, a single-family order, a duplicate agent and a flag-like model.
     - `TestAgentArgsPassTheModel`.
     - The codex and gemini cases in `TestUnavailable`.
   - What does not go red when reverted: the call site at `milestoneclose.go:874` (BR-4) and the fallback inside `Reviewer()`.
   - `Unavailable`, `ReviewPlan` and `LocalAvailability` have no production consumer yet. That is expected: the M2 dispatch loop is their consumer.

6. **Architecture notes**
   - ARCH-DRY: pass. `Reviewer()` derives from `ReviewPlan`, and the regex and line helpers are reused.
   - ARCH-PURE: pass. `family.go` is pure and `lookPath` is injected.
   - ARCH-PURPOSE: pass for M1's scope. Under the plan, M2 must make every review site use one resolver and pass `Model` from the `Candidate`, not from `policy.Models[reviewer]` read again at each site.
   - ARCH-MOCK: pass for now; nothing new calls an external service directly. M2's loop needs the fake-runner tests the plan promises: skip then run, all unavailable, and a started run never switching.
   - ARCH-CONSTRAINTS: N/A. At most 3 candidates.
   - ARCH-SECURE: pass. Models are validated and argv arrays are used. Author trailers come into play in M2, and the plan's rule to always include the closer's own family should be tested there.
   - ARCH-ORDER: pass. `Began` marks the point a review started, and uncertainty halts rather than switching.
   - ARCH-FUNERAL: pass. Nothing durable is created in M1.

7. **Plan revision recommendations**
   - Add a note that until M2, a boundary review under `other` uses the bounded order and ignores availability, or mark the related help and atlas text as M2.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      Validate rejects empty, single-family and duplicate orders (agent_resolver.go:124-142), pinned by TestValidateOrdersAndModels; with Validate, Reviewer() cannot reach same-family. The default-order fallback in Reviewer() is unpinned: reverting it leaves the suite green.
  - id: BR-2
    disposition: addressed
    note: |
      Unavailable now classifies codex (ERROR lines, activity blocks) and gemini (empty stdout, non-zero exit); TestUnavailable has codex quota, bad-model, echoed-prompt and began cases and the gemini IneligibleTierError case; the plan Revision records the per-CLI evidence.
  - id: BR-3
    disposition: addressed
    note: |
      close.md REVIEWER now describes the family-based other mode, reviewer_order and reviewer_models; atlas sdlc-binary.md:1191 no longer says claude and codex only.
  - id: BR-4
    disposition: not-addressed
    note: |
      agentArgs passes the flags (pinned by TestAgentArgsPassTheModel), but deleting Model from milestoneclose.go:874 leaves go test ./cmd/sdlc green, so a configured model can again be silently ignored. Add a boundaryReviewOptionsFor test that asserts opts.Model.
  - id: BR-5
    disposition: addressed
    note: |
      The M1 round-1 Revisions entry corrects the paths (family.go, stream.go, dispatch.go) and the signatures of FamilyOf, ReviewPlan and Unavailable, and adds LocalAvailability and AgentRun.Began; checked against the code.
  - id: BR-6
    disposition: addressed
    note: |
      trimReason is gone; grep finds no definition or caller.
  - id: BR-7
    disposition: addressed
    note: |
      Validate rejects a model starting with '-' and an agent listed twice; both are pinned in TestValidateOrdersAndModels.
findings:
  - id: new
    severity: Minor
    kind: defect
    family: docs-track-behavior-change
    title: |
      Close help and atlas describe M2 behaviour (kind order, availability fallback) that the boundary review does not do yet
    detail: |
      This is the 2nd finding in docs-track-behavior-change. Rule: help and atlas text must describe what the reviewed head does, and anything that lands later is marked with its milestone. Here close.md:415-421 says a full-flow close tries claude first and that gemini is the fallback, but boundaryReviewOptionsFor resolves through Reviewer(), which always uses KindBounded and ignores availability, so a gemini main agent on a full-flow close gets codex. Mark these as (M2) or land the kind with M2 and sweep both docs together.
```

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"f2f8d12a-9a73-4926-b20d-9022c59bf6bf","repository":"github.com/xianxu/ariadne","issue_id":"000409","recipe":"milestone-review","round":2,"attempt":1,"agent":"claude","root":"/Users/xianxu/Library/Caches/sdlc/review-checkouts/2d32b7370fc25891","reviewed_head":"948d85dd7dce63dd71cc6514dbf37b8b6f5dcd21","slot":"ariadne:2","requested_at":"2026-10-18T12:40:21.851335Z"},"phase":"finished","pid":69034,"started_at":"2026-10-18T12:40:21.997522Z","finished_at":"2026-10-18T12:55:36.714862Z","session_id":"28e965e1-250f-4b52-8f6c-0784b1160796","outcome":"returned"}] -->

---

## Re-review — 2026-10-18T06:00:31-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 409 — Dynamic review policy: choose the reviewer model by author family and change kind (cross-family by default), extending #310 |
| repo | ariadne |
| issue file | workshop/issues/000409-dynamic-review-policy-choose-the-reviewer-model-by-author-family-and-change-kind-cross-family-by-default-extending-310.md |
| boundary | milestone M1 |
| milestone | M1 |
| window | d3ebb55260c6647fbc70b423efed20cd8c734c5b..7753ca91157a2705e04fad8ec4e694835bd36089 |
| command | sdlc milestone-close --issue 409 --milestone M1 |
| reviewer | claude |
| isolation | fenced |
| timestamp | 2026-10-18T06:00:31-07:00 |
| verdict | SHIP |

## Review

Tests pass. Next I'll revert the model wiring to check that the regression test fails without it.

```verdict
verdict: SHIP
confidence: high
```

M1 is ready to ship. The two findings still open from round 2 are both fixed, and I checked each one against the code rather than the commit messages.

- **BR-4 (configured model ignored):** a `reviewer_models` entry now reaches the reviewer's command line through `DispatchOptions.Model`. I removed the fix in two places and ran the tests each time:
  - Without the `Model:` line in `milestoneclose.go:874`, `TestBoundaryReviewPassesTheConfiguredModel` fails (`reviewer argv lacks "--model claude-opus-5-5"`).
  - Without the `--model` argument in `agentArgs`, `TestAgentArgsPassTheModel` fails.
  - With the fix in place, `go vet ./...`, `go test ./internal/judge/` and the targeted `cmd/sdlc` tests all pass.
- **BR-8 (docs describing M2 behaviour):** `close.md` and the atlas now say that until M2 lands, a boundary review uses the bounded order and its first agent only. That matches the code: `Reviewer()` always uses `KindBounded` and ignores availability.

I found nothing new that blocks the boundary.

**1. Strengths**
- `Reviewer()` gets its answer from `ReviewPlan`, so there is one source for who reviews. A second pass over the default orders means `other` can never return the main agent's own family (`agent_resolver.go:152-156`). `TestReviewerOtherNeverSameFamily` tests that rule directly, not specific outputs.
- The model is wired from the policy into the dispatched command line, and it is tested end-to-end through the fake reviewer, including the negative case where a reviewer with no entry gets no model flag (`reviewconfig_test.go:817-856`).
- Each CLI's flag is placed before the prompt, and the test asserts the prompt stays the last argument (`family_test.go:681-696`).
- `Unavailable` treats a review as begun whenever there is any doubt, so the fallback cannot quietly swap a review that partly ran for a second opinion. The codex case reads only `ERROR:` lines, because codex echoes the prompt to stderr.

**2. Critical findings:** none.

**3. Important findings:** none.

**4. Minor findings:** none raised. One note that is not a finding: a forced `--agent` or `AGENT_CMD` reviewer also gets `policy.Models[reviewer]`. That behaviour is reasonable but not documented; M2's resolver should state it on purpose.

**5. Test coverage notes:** the removal checks above confirm the model wiring is tested at both the CLI-argument layer and the boundary-review layer. M2 still owes the fake-runner tests of the dispatch loop: skip then run, all candidates unavailable, started but never switched.

**6. Architecture**
| Principle | Result | Why |
|---|---|---|
| ARCH-DRY | pass | `Reviewer()` reuses `ReviewPlan` |
| ARCH-PURE | pass | `family.go` does no IO; `lookPath` and `goos` are passed in |
| ARCH-PURPOSE | pass | the remaining review sites (`changecode.go:725` and `:965`, `judge.go:148`) still call `ResolveAgentCLI`, but the plan gives them to M2 explicitly |
| ARCH-MOCK | pass | tests run against the existing fake reviewer |
| ARCH-CONSTRAINTS | N/A | M1 has no loop; at most 3 candidates |
| ARCH-SECURE | pass | model names that start with `-` or contain spaces are refused |
| ARCH-ORDER | pass for M1 | no state is kept between events yet; M2's loop is where this applies |
| ARCH-FUNERAL | pass | nothing new is written to disk |

**7. Plan revisions:** none needed; the second Revisions entry matches the code.

```findings
dispose:
  - id: BR-4
    disposition: addressed
    note: |
      DispatchOptions.Model is set from policy.Models at milestoneclose.go:874 and agentArgs emits --model or -m; removing either one makes TestBoundaryReviewPassesTheConfiguredModel or TestAgentArgsPassTheModel fail.
  - id: BR-8
    disposition: addressed
    note: |
      close.md:72-73 and the atlas paragraph now mark the kind order and fallback as M2, and say a boundary review uses the bounded order's first agent, which matches Reviewer() always using KindBounded.
```

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"aa384400-3112-4490-bac5-df118069e10f","repository":"github.com/xianxu/ariadne","issue_id":"000409","recipe":"milestone-review","round":3,"attempt":1,"agent":"claude","root":"/Users/xianxu/Library/Caches/sdlc/review-checkouts/98ff97a217b8180e","reviewed_head":"7753ca91157a2705e04fad8ec4e694835bd36089","slot":"ariadne:2","requested_at":"2026-10-18T12:57:38.404728Z"},"phase":"finished","pid":35806,"started_at":"2026-10-18T12:57:38.527913Z","finished_at":"2026-10-18T13:00:12.116973Z","session_id":"c99ef9a6-6af1-47e0-9ed1-c994bf6656aa","outcome":"returned"}] -->
