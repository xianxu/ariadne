# Boundary Review — ariadne#310 (whole-issue close)

| field | value |
|-------|-------|
| issue | 310 — Choose the boundary reviewer's agent per repo, and model read-only |
| repo | ariadne |
| issue file | workshop/issues/000310-choose-the-boundary-reviewer-s-agent-per-repo-and-model-read-only.md |
| boundary | whole-issue close |
| milestone | — |
| window | 15b365d31f1d17f8217476a7baf62fb741f47549..3c84da0afce2fd327726c6d3521e0a9f71de9195 |
| command | sdlc close --issue 310 |
| reviewer | claude |
| isolation | checkout |
| timestamp | 2026-10-11T09:52:15-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

Still inspecting the diff; checking for stale references and the fence-reason wiring.

Verified the EPERM fix goes red when reverted; now checking the lock's age fallback, then writing the verdict.

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

The issue's main goals are delivered and tested. `workshop/review.json` can set `same`, `other` or `explicit`. The pure `judge.ReviewerPolicy` and `ResolveReviewerCLI` keep `--agent` and `AGENT_CMD` above the policy, and `ResolveAgentCLI` is now just the `same` case. A bad config file stops the review and names the file. An end-to-end test shows a claude-driven boundary reviewed by codex and the reverse, with the reviewer recorded in the ledger round. `Containment` replaces the old `IsSandbox`/`Fenced` booleans, and no judge gets an auto-approve flag any more. Build, vet, `go test ./internal/judge/` and the targeted `cmd/sdlc` tests all pass.

One thing stops a plain SHIP: the help text and atlas say "every judge runs read-only", which overclaims. Claude's allowed tools still include unrestricted `Bash` under `dontAsk`, and codex at checkout level runs `workspace-write`. So a reviewer's write through a shell command still succeeds. In the disposable checkout that is acceptable by design. But for `ContainNone` judges (plan-quality, `sdlc judge`) it lands in the operator's own tree. That is the same kind of mutation the issue's evidence describes, such as leaving the checkout detached.

1. **Strengths**
   - `agent_resolver.go`: the policy is pure, its zero value means `same`, and `Validate` rejects typos and pairs that don't match the mode. The loader at `reviewconfig.go:39` uses `DisallowUnknownFields`, so a misspelled key fails loudly instead of quietly falling back to `same`.
   - `dispatch.go:154-171`: one `Containment` enum replaces two independent booleans that had unclear combinations. `TestBuildArgs_ReadOnlyPerContainment` checks the full agent × level grid and also checks that no auto-approve flag appears.
   - `milestoneclose.go:806`: the policy is read from the source checkout *before* the root moves to the review clone. That is the right place to trust config from.
   - `TestBoundaryReviewOtherPolicySwapsReviewer` covers the Done-when end to end: the swap, the read-only argv and the ledger `Agent` field.
   - The `processAlive` side-quest fix is proven. I reverted it and `TestProcessAliveTreatsEPERMAsAlive` failed (`processAlive(1) = false`); I then restored it.

2. **Critical**: none.

3. **Important**
   - `cmd/sdlc/helptext/close.md:376-384` and `atlas/workflow/sdlc-binary.md:996-1000` say every judge runs read-only. In fact claude's allowed tools are `Read,Grep,Glob,Bash` (`prompts.go:194`), and `dontAsk` lets any Bash command run. The live check only tried the `Write` tool, not a Bash write. For `ContainNone` judges, a Bash write such as `git checkout` still reaches the operator's tree (ARCH-SECURE, ARCH-PURPOSE). Fix: either narrow Bash to read-only command patterns (for example `Bash(git log:*)`, `Bash(git diff:*)`, `Bash(go test:*)`), or reword the docs. The honest version is: "Edit/Write refused; shell writes are contained by the review checkout or fence, not prevented."

4. **Minor**
   - `milestoneclose.go:675`: the codex "cannot nest" warning fires whenever there is no fence. That includes Linux without bwrap, where codex's own Landlock sandbox works, so there the warning is misleading. Fix: gate it on the reason actually being a failed Seatbelt probe.
   - The plan's checkbox claims an atlas update to `openshell-sandbox`, but the range doesn't touch that file. Nothing in it is stale about judges, so the plan line is what needs fixing.
   - `agent_resolver.go:99`: "reviewer_pairs is read only with…" can be misread as "read-only". Better: "reviewer_pairs applies only to boundary_reviewer explicit".
   - `cmd/doc-review/review.go:165`: this reviewer still passes claude `bypassPermissions`. It is outside #310's boundary-review scope; note it for a later sweep.
   - `term.go:3`: the reflowed header comment is now one overlong line.
   - Treating EPERM as alive means a same-host lock whose PID gets reused by another user's process never reads as stale, because `Observe` has no age fallback on the same host. This is rare and was already a PID-reuse risk.

5. **Test coverage**
   - The policy, loader, argv grid, the end-to-end swap and a bad config are all covered.
   - Missing: a test that a claude judge cannot write through Bash. Today it can, which is the Important finding above.
   - The live conformance test lost `IsSandbox: true`, so codex and gemini now run in their read-only defaults there. That fits a prompt that uses no tools.

6. **Architecture**
   - **ARCH-DRY: pass.** `ResolveAgentCLI` delegates to the policy resolver.
   - **ARCH-PURE: pass.** The policy and argv mapping are pure, and the file read is a thin loader.
   - **ARCH-PURPOSE: flag**, as in the Important finding. "Read-only" is enforced for the edit tools but not for shell writes.
   - **ARCH-MOCK: pass.** The tests go through the `judge.Run` seam and the `reviewbox.Probe` fake.
   - **ARCH-CONSTRAINTS: pass.** Nothing new runs on a hot path.
   - **ARCH-SECURE: flag** (same finding). Config input is parsed into typed values with strict decoding, which is good.
   - **ARCH-ORDER: pass.** The two booleans collapsed into an enum, and nothing new holds state between events.
   - **ARCH-FUNERAL: pass.** The only new artifact is one optional config file the operator owns.

7. **Plan revisions**
   - Add a `## Revisions` entry: "atlas: only sdlc-binary.md updated; openshell-sandbox.md needed no change (its auto-approve text is about the interactive sandbox aliases, not judges)."
   - Add a `## Revisions` entry: "Read-only means Edit/Write refused for every CLI; Bash/shell writes are bounded by the review checkout or fence, not refused."

```findings
findings:
  - id: new
    severity: Important
    family: read-only-claim-exceeds-enforcement
    title: |
      Help/atlas say every judge runs read-only, but claude's unrestricted Bash (and codex workspace-write) still let shell writes through
    detail: |
      AllowedTools is Read,Grep,Glob,Bash under dontAsk, so any Bash command runs. For ContainNone judges (plan-quality, sdlc judge) that write reaches the operator tree. Either narrow Bash to read-only patterns or reword close.md:376-384 and sdlc-binary.md:996-1000.
  - id: new
    severity: Minor
    family: warning-scope-too-broad
    title: |
      The codex cannot-nest warning fires on any unfenced review, including Linux without bwrap where codex's sandbox works
  - id: new
    severity: Minor
    family: plan-claims-undelivered-item
    title: |
      Plan checkbox claims an atlas update to openshell-sandbox; the range does not touch that file
  - id: new
    severity: Minor
    family: error-message-clarity
    title: |
      Validate's "reviewer_pairs is read only with" wording can be misread as read-only
  - id: new
    severity: Minor
    family: read-only-claim-exceeds-enforcement
    title: |
      cmd/doc-review still passes claude bypassPermissions to its reviewer (outside #310's scope)
```

---

## Re-review — 2026-10-11T09:55:58-07:00 (FIX-THEN-SHIP)

| field | value |
|-------|-------|
| issue | 310 — Choose the boundary reviewer's agent per repo, and model read-only |
| repo | ariadne |
| issue file | workshop/issues/000310-choose-the-boundary-reviewer-s-agent-per-repo-and-model-read-only.md |
| boundary | whole-issue close |
| milestone | — |
| window | 15b365d31f1d17f8217476a7baf62fb741f47549..4ea45451f02a6104f95f63ca7048c7c2d34c75b0 |
| command | sdlc close --issue 310 |
| reviewer | claude |
| isolation | checkout |
| timestamp | 2026-10-11T09:55:58-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

Checking the fix round now: four of the five prior findings look addressed. The remaining question is whether the new "judges always run read-only" flag text repeats BR-1. Next I'm running the tests.

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

The issue's purpose is delivered and tested, and four of the five prior findings are fixed. BR-1 is only partly fixed:

- **What's done.** A repo can choose its reviewer in `workshop/review.json` (`same`, `other` or `explicit`). An invalid file halts the review and names the file. A claude-driven close is reviewed by codex, and the reverse, with the reviewer recorded in the ledger round. `Containment` maps each CLI to its own no-auto-approve mode.
- **What's checked.** Build and vet are clean. These all pass: `go test ./cmd/sdlc/internal/judge/`, `./cmd/doc-review/`, and the targeted `cmd/sdlc` tests (`TestLoadReviewerPolicy`, `TestBoundaryReview*`, `TestProcessAlive*`).
- **What blocks SHIP.** BR-1 was fixed at the two places the finding named, and those two now say plainly that Bash writes still go through. But the same range added four more user-facing strings that say "judges always run read-only". They sit on exactly the judges where that is least true: `sdlc judge` and `change-code` (plan-quality) run in the operator's tree with no fence. The new `close.md` text itself admits a claude judge's Bash can still write there.

1. **Strengths**
   - `agent_resolver.go`: `ResolveAgentCLI` is now just the `same` case of `ResolveReviewerCLI`, so the resolution order lives in one place (ARCH-DRY). The policy is pure, and its zero value is `same`.
   - `reviewconfig.go:39`: `DisallowUnknownFields` plus `Validate` means a typo in the file fails loudly instead of quietly reviewing with `same`.
   - `milestoneclose.go:806`: the policy is read from the source checkout before the root moves to the review clone, so config is trusted from the operator's tree.
   - `TestBuildArgs_ReadOnlyPerContainment` checks every agent × containment level and asserts that no auto-approve flag appears.
   - `TestBoundaryReviewOtherPolicySwapsReviewer` covers the Done-when end to end: the swap, the argv, the darwin-only warning and the ledger `Agent` field.

2. **Critical**: none.

3. **Important**
   - BR-1 is re-raised below (`not-addressed`): the class of overclaiming strings wasn't swept.
   - The sites still overclaiming: `cmd/sdlc/judge.go:68-69` and `cmd/sdlc/changecode.go:104-105` (flag usage and `MarkDeprecated` message), `cmd/sdlc/helptext/judge.md:70`, `cmd/sdlc/helptext/change-code.md:185`, and loosely `cmd/sdlc/helptext/milestone-close.md:44` ("read-only, inside a disposable clone").
   - **This is the 3rd finding in family `read-only-claim-exceeds-enforcement`**, so the rule matters more than another instance fix. The rule: no user-facing text may call a judge "read-only" without the enforced scope, which is "Edit/Write refused, no auto-approve; shell writes bounded only by the review checkout / #204 fence".
   - Fix: grep the help text, flag text and deprecation messages for `read-only` claims about judges. Replace each with the scoped wording, or with "no auto-approve mode", which is true everywhere.

4. **Minor**
   - `term.go:3`: the reflowed header comment is now one overlong line.
   - The darwin-gated codex warning still fires for any unfenced darwin review, whatever `iso.Reason` says. That is acceptable, since on darwin the only realistic reason is a nested Seatbelt.

5. **Test coverage**
   - BR-2: `TestBoundaryReviewOtherPolicySwapsReviewer` pins the warning to `runtime.GOOS == "darwin"`, so on Linux it would fail if the gate were dropped.
   - BR-5: `TestBuildArgs_ReadOnly` asserts `--permission-mode dontAsk`, so it fails without the fix.
   - Still missing: a test that a claude judge cannot write through Bash. It can, by design, so the fix belongs in the docs, not in a new test.

6. **Architecture**
   - **ARCH-DRY: pass.**
   - **ARCH-PURE: pass.** The policy and argv mapping are pure; the loader is a thin IO shell.
   - **ARCH-PURPOSE: flag.** The BR-1 class sweep is incomplete, as above.
   - **ARCH-MOCK: pass.** Tests go through the `judge.Run` seam and the `reviewbox.Probe` fake.
   - **ARCH-CONSTRAINTS: pass.** Nothing new runs on a hot path.
   - **ARCH-SECURE: flag.** Same finding: the trust claim is wider than what's enforced. The config parsing itself is strict and typed.
   - **ARCH-ORDER: pass.** Two booleans collapsed into one enum, and nothing new holds state between events.
   - **ARCH-FUNERAL: pass.** The only new artifact is one optional config file the operator owns.

7. **Plan revisions**
   - Add a `## Revisions` line: "read-only" for judges means no auto-approve mode and Edit/Write refused; shell writes are bounded by the review checkout or fence, not refused, and all help and flag text says so.

```findings
dispose:
  - id: BR-1
    disposition: not-addressed
    note: |
      close.md REVIEWER and atlas sdlc-binary.md were reworded honestly, but the same range added "judges always run read-only" to the --sandbox flag usage and MarkDeprecated text (judge.go:68-69, changecode.go:104-105) and helptext judge.md:70 / change-code.md:185, which are exactly the unfenced ContainNone judges whose claude Bash can write the operator tree; milestone-close.md:44 also says "read-only". 3rd instance of read-only-claim-exceeds-enforcement. Rule: no user-facing text calls a judge read-only without the enforced scope (Edit/Write refused, no auto-approve; shell writes bounded only by checkout/fence). Sweep every such string.
  - id: BR-2
    disposition: addressed
    note: |
      milestoneclose.go:676 gates the warning on runtime.GOOS == "darwin"; TestBoundaryReviewOtherPolicySwapsReviewer asserts warned == (codex and darwin).
  - id: BR-3
    disposition: addressed
    note: |
      The plan checkbox now says openshell-sandbox describes the main agents and is unchanged; no atlas claim remains for that file.
  - id: BR-4
    disposition: addressed
    note: |
      agent_resolver.go:99 now reads "reviewer_pairs applies only to boundary_reviewer \"explicit\"".
  - id: BR-5
    disposition: addressed
    note: |
      doc-review/review.go:165 now uses dontAsk; TestBuildArgs_ReadOnly asserts --permission-mode dontAsk and fails without the fix.
```
