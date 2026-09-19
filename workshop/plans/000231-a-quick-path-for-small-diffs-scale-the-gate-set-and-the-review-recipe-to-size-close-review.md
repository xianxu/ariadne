# Boundary Review — ariadne#231 (whole-issue close)

| field | value |
|-------|-------|
| issue | 231 — A quick path for small diffs: scale the gate set and the review recipe to size |
| repo | ariadne |
| issue file | workshop/issues/000231-a-quick-path-for-small-diffs-scale-the-gate-set-and-the-review-recipe-to-size.md |
| boundary | whole-issue close |
| milestone | — |
| window | adaa53a60caea9815780269c76a852df03496f77..ed3ceb78e73d89dbec443bbda197b72cc58eae61 |
| command | sdlc close --issue 231 |
| reviewer | claude |
| timestamp | 2026-09-18T19:07:07-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

The issue is delivered. `change-code` infers the flow and lets the operator pin it. Close measures the size limits, upgrades a quick issue to full when it crosses one (the upgrade sticks across a REWORK), picks the review recipe, and runs the Done-when checks. The publish check re-measures fixes made after the verdict, and calibration skips quick and upgraded rows. I ran the whole suite in a scratch clone at `ed3ceb7`: it is green except the known #210 failure (`TestFleetPlanHasAuthoritativeCorrectedCoreConceptInventory`). I found no correctness bug. Three small things stand between this and a clean SHIP, all cheap:
- **BR-32's test proves nothing.** I reverted its `-z` fix and it stayed green.
- **Two outputs and one atlas page still state the old rules.** One is the message close prints after a FIX-THEN-SHIP verdict.
- **The friction audit can't see the new publish refusal.** Its pattern for merge/push doesn't match the new message.

**Strengths**
1. **The compiler now enforces BR-16.** `Flow`'s fields are unexported (`flow.go:52-56`) and `Rule` has unexported values (`flow.go:219-225`), so outside the package a flow can only come from `Parse`/`Decide`/`WithContract`/`Upgrade`. The old source-scan guard is gone.
2. **One corpus is read by both the Go parser and the cue model.** Each reject row names the branch that refuses it, and the test requires every branch to have a row (`flow_test.go:205-233`). I disabled the duplicate-key check and the test went red.
3. **change-code and close apply the same size limits.** Both call `flow.Measure`/`Crossings` (`changecode_flow.go:47`, `closeflow.go:125`). The limits are written once and rendered into help and output by `ShellSummary`/`AfterReviewSummary`.
4. **The upgrade survives a REWORK.** Each boundary-ledger round records the recipe it ran (`boundaryledger.go:179-183`), and an unstamped round reads as the full review (`closeflow.go:152-163`), so the safe reading wins.
5. **The quick principles are chosen in the registry.** It splits into sections without losing a byte, every entry must declare the field, and the small-diff golden prompt carries exactly 3 principles. Calibration uses one `calibratable` predicate for both the drift sample and the latest-row pick (`drift.go:78`).

**Critical findings:** none.

**Important findings:** none.

**Minor findings**
- **New #1 — three places still state the old rules** (7th in `doc-claim-contradicts-code`):
  - After a FIX-THEN-SHIP verdict, close prints "Do NOT re-run `sdlc close` — this verdict already sanctions shipping after the fixes" (`close.go:1871-1873`). That holds for every issue except a quick one whose fixes pass 200 lines, which the publish check now refuses (`publishgate.go:176`).
  - The publish gate's own atlas page (`atlas/workflow/pre-merge-checks.md:24-46`) lists its refusal conditions and that protocol without the quick re-measure.
  - Close prints "header upgraded to the #187 column set" (`close.go:981`). The only upgrade that can now fire is #187 → #231.
- **New #2 — the friction audit can't count the new publish refusal** (2nd in `gate-key-undeclared`). The merge/push `no-judge` pattern is `publish gate: \d+ commit\(s\) landed after` (`gatesig.go:150,156`). It doesn't match "publish gate: #N is on the quick flow…".
- **BR-32 is still open.** See its note in the block below.

**Test coverage notes**
- Every size limit is tested exactly at the limit and one past it, at each gate: 100/101 lines, 500/501 design lines, and 200/201 at publish. The 500/501 case at close uses a real durable plan.
- REWORK → shrink → re-close keeps the full review.
- The corpus is asserted by both the Go parser and the cue model.
- My mutation probes:

  | Fix | Result when the fix is reverted |
  |---|---|
  | BR-17 | red, as it should be |
  | BR-27 | red, as it should be |
  | BR-32 | **green** — the test doesn't catch it |

- A fixture with a quoted doc/test name beside a small code file, expected to stay quick, does go red without `-z`.

**Architecture walk**
- **ARCH-DRY: pass.** `windowFileStats` is shared by the churn report, close and the publish check. The code/doc path checks are all built from the same `churn` predicates. `boundary-tail.md` is one include used by both recipes, and there is one milestone parser.
- **ARCH-PURE: pass.** The flow package is pure, and `closeFlowStep`/`quickGrewPastReview` are thin IO around pure decisions.
- **ARCH-PURPOSE: flag (new #1).** Of the surfaces that describe post-verdict fixes and publish refusals, the help pages and `sdlc-binary.md` derive from `AfterReviewSummary`. The runtime FIX-THEN-SHIP message and `pre-merge-checks.md` still restate the old unconditional rule.
- **ARCH-MOCK: pass.** git runs in temp repos, the judge runs through the `stubJudge` seam, and nothing new is external.
- **ARCH-CONSTRAINTS: pass.** The added cost is one numstat per close and one per quick codecomplete issue at publish; the registry is parsed once.
- **ARCH-SECURE: pass.** A malformed flow record resolves to full, an unreadable ledger counts as a crossing, and paths are parsed with `-z`. One note is under Architectural notes.
- **ARCH-ORDER: pass.** `Decide` and `Upgrade` are the only transitions and there is no downgrade. An edit made while change-code's gates run is re-derived, and `flowDrift` refuses if it changed the flow.
- **ARCH-FUNERAL: pass.** The flow record is one line replaced in place. The recipe field lives in the per-issue ledger, which is archived with the issue. Each calibration row gains three short columns; the ledger's own lifecycle is unchanged.

**Architectural notes for upcoming work**
- **The publish check trusts the working-tree record.** Editing `kind: full` into the issue after close is a docs-only change, and it would switch the re-measure off. Reading the record from the codecomplete anchor commit instead would close that.
- **ariadne's own composed `AGENTS.md`/`CLAUDE.md` are stale.** They are gitignored weave output and still say "Mx milestones or a durable plan mean the full flow". Run `make weave` here too.
- **#235 (plan lookup by exact name) is the remaining hole in the design limit.** If the plan file's name doesn't match, the plan isn't counted.

**Plan revision recommendations**
- **Correct the BR-32 Revisions line** ("the test reddens without `-z`", plan:503). Since f0141e0 it doesn't; name the quoted doc/test fixture that does.
- **Tidy the stale plan prose.** The Revisions at plan:446 already record the true names, so this is not a table-vs-code conflict:
  - The Core concepts row `judge.ArchitectureSection` (plan:32) should read `architectureSections`/`archSection`.
  - The close-step prose "`churn.ParseNumstat(git diff --numstat …)`" (plan:89) should read `-z` + `ParseNumstatZ`.
  - The `Flow` shape should show unexported fields with accessors.

```findings
dispose:
  - id: BR-15
    disposition: addressed
    note: |
      change-code.md step 0, issue-lifecycle.md:54 and sdlc-binary.md:36 now say the record is written after the gates; the flowDrift refusal is documented under THE FLOW.
  - id: BR-16
    disposition: addressed
    note: |
      Flow fields are unexported with accessors (flow.go:52-56) and Rule is opaque (flow.go:219-225), so the compiler enforces the single producer; the source-scan guard is gone.
  - id: BR-17
    disposition: addressed
    note: |
      Reason codes plus the every-reason-has-a-row assertion (flow_test.go:205-233). Probed: disabling the seen[k] check turns TestFlowRecordCorpus red on both assertions.
  - id: BR-27
    disposition: addressed
    note: |
      AGENTS.base.md:52 points at the catalog; help pages render the gate flags; the build-closure prose went with shared surfaces. Probed: hand-listing a gate flag in close.md turns TestGateFlagListsAreRenderedFromTheCatalog red.
  - id: BR-31
    disposition: withdrawn
    note: |
      Overtaken by design: f0141e0 removed shared surfaces and their declaration, so no grammar is left to describe.
  - id: BR-32
    disposition: not-addressed
    note: |
      -z is in place (closeflow.go:141), but TestCloseQuotedNameCodeFileLinesCount cannot fail. It uses a code file, and IsCodeFile's CodeProd default counts the quoted path as code either way; once f0141e0 dropped the -z name-list intersection, the test lost its teeth. Probed: restoring `-c core.quotePath=false diff --numstat` with ParseNumstat leaves it green. The discriminating fixture goes red: docs/a"b.md and tests/x"y_spec.lua at 150 lines beside a 5-line cmd/a.go, expecting quick. Plan:503's "the test reddens without -z" is false at HEAD.
  - id: BR-33
    disposition: withdrawn
    note: |
      Overtaken by design: ParseSurfaces and the fuzz target were removed with shared surfaces (f0141e0).
findings:
  - id: new
    severity: Minor
    family: doc-claim-contradicts-code
    title: |
      Two runtime messages and the publish gate's atlas page restate rules this window changed: FIX-THEN-SHIP's unconditional "do not re-run close", and the "#187 column set" header message
    detail: |
      This is the 7th finding in doc-claim-contradicts-code. Instances: close.go:1871-1873 says "Do NOT re-run sdlc close — this verdict already sanctions shipping after the fixes", yet publishgate.go:176 now refuses a quick issue past MaxAddedLinesAfterReview and sends it back to close; atlas/workflow/pre-merge-checks.md:24-46 (the publish gate's page) lists its refusals and that protocol without the quick re-measure; close.go:981 prints "header upgraded to the #187 column set", and the only upgrade left to fire is #187 to #231. Enumeration: 7 surfaces describe what the publish gate refuses or what the verdict sanctions (close.md:222, merge.md:35, push.md:10, sdlc-binary.md:303 and publishgate.go:6 derive; close.go:1871 and pre-merge-checks.md do not). 4 surfaces name the ledger's column set, and close.go:981 is wrong. Prevalence is 3 of 11. Why earlier sweeps missed these: they keyed on the new mechanism's name, while these sentences state the old conclusion. Rule: a runtime string that states a policy renders it from the policy's owner, never a literal. formatFixThenShipProtocol takes the flow and renders flow.AfterReviewSummary() on quick; the header message states no version. When a change adds a policy function, the sweep greps for the old policy's conclusion ("do not re-run", "landed after", "#187"), not the new gate's name (ARCH-PURPOSE shadow sweep).
  - id: new
    severity: Minor
    family: gate-key-undeclared
    title: |
      The publish check's new quick-growth refusal matches no GateCatalog RefusalPat, so the friction instrument never counts it
    detail: |
      This is the 2nd finding in gate-key-undeclared. The merge/push no-judge RefusalPat is `publish gate: \d+ commit\(s\) landed after` (gatesig.go:150,156). "publish gate: #N is on the quick flow, and its diff grew…" (publishgate.go:180) is unattributed by classifyOutputLine, although --no-judge waives it. Enumeration of refusals added in this window that a catalogued flag waives: DoneWhenFresh (catalogued) and the publish quick-growth refusal (not catalogued), so 1 of 2. Rule: every refusal a catalogued flag waives must be attributable by the catalog. The tests that produce a refusal assert that their (command, flag) RefusalPat matches it, the inverse of assertNoGatesigCollision. Fix: widen the two patterns to `publish gate: (\d+ commit\(s\) landed after|#\d+ is on the quick flow)` and add that assertion to TestRunPublishGate_QuickGrewPastReview.
```
