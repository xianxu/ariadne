---
gate: boundary-review
issue: 231
id_prefix: BR
rounds:
    - "n": 1
      timestamp: "2026-09-17T20:53:40-07:00"
      agent: sdlc
      findings:
        - id: BR-1
          severity: Minor
          title: Tests are listed case by case in prose; the risky parsers have no one-line adversarial strategy
          detail: |-
            Name each test and replace the case lists with one line per risky function. For example: fuzz flow.FromFrontmatter seeded with malformed, nested and duplicate-key maps (never panics, malformed resolves to full); fuzz ParseSurfaces (a bad pattern becomes a crossing); DoneWhenFresh section extraction over fenced headings.
            (carried from plan-quality PQ-8, deferred to the boundary review)
          family: test-prose-enumeration
          round: 1
        - id: BR-2
          severity: Minor
          title: Unquoted 8-hex hashes in the flow record become YAML ints, and cue checks at push/merge refuse them
          detail: |-
            Verified with local cue: spec 12345678 gives "mismatched types int and string" against the plan's own #Flow. About 2.3 percent per hash, roughly 4.6 percent of quick records. validategate.go runs this on every changed issue at push and merge, and yaml.v3 hides it on the Go side. Rule: a persisted value must parse to the same type under every reader. Fix: Format quotes the hashes, and the round-trip test, FuzzFromFrontmatter and the validate_test flow case are seeded with an all-digit hash.
            (carried from plan-quality PQ-9, deferred to the boundary review)
          family: record-roundtrips-every-reader
          round: 1
        - id: BR-3
          severity: Minor
          title: Spec section 4 and Done-when still say quick yes/no after the plan renamed it quick-flow, and Done-when's Flow omits the hash fields
          detail: |-
            Third finding in this family. The rule for all of them: a revision that renames or reshapes an artifact Done-when names restates that bullet in the same edit. Apply it now by checking every name in Done-when against the plan (both drifts here) before implementing. DoneWhenFresh will enforce this mechanically once it ships.
            (carried from plan-quality PQ-10, deferred to the boundary review)
          family: acceptance-oracle-freshness
          round: 1
      boundary: '*'
      no_cap: true
      blocked: false
    - "n": 2
      timestamp: "2026-09-17T20:53:40-07:00"
      agent: claude
      findings:
        - id: BR-4
          severity: Important
          title: decideChangeCodeFlow's PlanItemsBody routing is unpinned — not in planItemBodySources, no fenced-Mx test
          detail: changecode_flow.go:48 hands a plan body to the exempt helper issue.MilestonesInPlanOrder, whose exemption says callers are covered by planItemBodySources (commitpathspec_guard_test.go:341), which lists only close.go. Reverting to PlanSectionBody in a scratch copy left every Flow/Guard/PlanItem/ChangeCode test green. Add a fenced-Mx case to TestDecideChangeCodeFlow (must stay quick) and sweep all MilestonesInPlanOrder callers (ARCH-PURPOSE).
          family: hand-listed-guard-incomplete
          round: 2
        - id: BR-5
          severity: Important
          title: change-code writes the flow record before its gates, so a refused run leaves a sticky full/inferred
          detail: applyChangeCodeFlow runs at changecode.go:137 before the gate loop; a structural/plan-quality/estimate refusal exits with flow full/inferred written (uncommitted), and Decide keeps a recorded full, so removing the plan or Mx rows and re-running stays full. The plan's ARCH-ORDER table says REWORK writes nothing. Decide before the gates, write after they pass (before syncIssue); test that a refused run leaves the file byte-identical (ARCH-ORDER).
          family: state-persisted-on-refused-transition
          round: 2
        - id: BR-6
          severity: Minor
          title: M1 docs describe close-time flow behaviour (small-diff review, upgrade, Done-when refusal) that M2 hasn't built
          detail: helptext/change-code.md THE FLOW, helptext/issue.md flow field, AGENTS.base.md section 2, atlas issue-lifecycle item 6 and flowInfoLine state it in the present tense; at HEAD close ignores the flow entirely. Acceptable only if M2 lands before any merge; note it in the Log.
          family: doc-claim-contradicts-code
          round: 2
        - id: BR-7
          severity: Minor
          title: flowReason re-derives Decide's rule order instead of Decide returning the rule that fired
          detail: changecode_flow.go:66-81 mirrors flow.go Decide rules; a reorder in Decide silently mislabels the reason (ARCH-DRY).
          family: decision-logic-restated
          round: 2
        - id: BR-8
          severity: Minor
          title: Rewriting a block-form hand-edited flow record orphans its child lines into invalid frontmatter
          detail: SetField replaces only the flow line; a probe produced flow full/inferred followed by indented kind/provenance lines, while the warning says the record is being rewritten (ARCH-SECURE).
          family: untrusted-input-rewrite
          round: 2
        - id: BR-9
          severity: Minor
          title: FuzzFromFrontmatter does not assert that the error path resolves to full
          detail: The plan names "every error path resolves to full" as the fuzz oracle; the body returns on error without checking Kind == Full.
          family: test-oracle-weaker-than-plan
          round: 2
        - id: BR-10
          severity: Minor
          title: issue.cue carries unrelated cue-fmt whitespace realignment of categories, Status and Transition
          family: unrelated-churn
          round: 2
      boundary: M1
      blocked: true
    - "n": 3
      timestamp: "2026-09-17T21:07:39-07:00"
      agent: claude
      dispose:
        - id: BR-1
          disposition: addressed
          note: Plan names every test. FuzzFromFrontmatter plus the fenced-heading case in TestContractHashes cover M1's parsers; FuzzParseSurfaces is M2 scope; keeping case lists is a recorded decision.
          round: 3
        - id: BR-2
          disposition: addressed
          note: Format quotes hashes with %q (flow.go:66-72); scratch revert to %s turns TestFlowRoundTrip red on 12345678; validate_test covers quoted-valid and unquoted-invalid through real cue.
          round: 3
        - id: BR-3
          disposition: addressed
          note: Spec section 4 and Done-when say quick-flow and the Flow bullet carries spec/done; old names survive only in dated Revisions.
          round: 3
        - id: BR-4
          disposition: addressed
          note: Scratch revert of changecode_flow.go:50 or close.go:1719 to PlanSectionBody turns TestPlanItemReadersUsePlanItemsBody red; the fenced-Mx case in TestDecideChangeCodeFlow also goes red.
          round: 3
        - id: BR-5
          disposition: addressed
          note: Moving recordChangeCodeFlow before the gate loop turns TestRunChangeCodeRecordsFlowAfterGates red. The fix opened a stale-snapshot window, raised separately.
          round: 3
        - id: BR-6
          disposition: addressed
          note: The 2026-09-17 "M1 built" Log entry records that the present-tense close docs hold only because M2 lands before merge.
          round: 3
        - id: BR-7
          disposition: addressed
          note: Decide returns the Rule that fired, flowReason is gone, and TestDecide asserts the rule for each case.
          round: 3
        - id: BR-8
          disposition: addressed
          note: SetField drops the indented block continuation (frontmatter.go:73-99), pinned by TestSetFieldReplacesBlockValue.
          round: 3
        - id: BR-9
          disposition: addressed
          note: Scratch revert of FromFrontmatter's error value to Flow{} turns FuzzFromFrontmatter's seed corpus red.
          round: 3
        - id: BR-10
          disposition: addressed
          note: The issue.cue diff is now 17 pure additions, with no realignment.
          round: 3
      findings:
        - id: BR-11
          severity: Important
          title: change-code writes the flow record from issue text read before the gates, clobbering edits made while plan-quality ran
          detail: 'recordChangeCodeFlow (changecode_flow.go:100-108) writes d.content, built at changecode.go:138 from the step-2 read, after the multi-minute gate loop. A scratch probe (write, report, edit the file on disk, record) lost the edit, and syncIssue then commits it and, on main, pushes it. It fires on every issue''s first successful run, because the record is always new. The repo lock does not cover editor or agent edits; close handles the same hazard by validating its snapshot after the review (close.go:1218-1231). Fix: re-read issuePath at record time and apply SetField to the fresh frontmatter, recomputing quick hashes over the fresh body, or refuse as close does. Add the probe as a regression test (ARCH-ORDER, second actor on the same state).'
          family: write-from-stale-snapshot
          round: 3
        - id: BR-12
          severity: Minor
          title: flow.Parse accepts spec "" and done "" that cue Flow rejects, contradicting Parse's own doc comment
          detail: 'This is the 2nd finding in record-roundtrips-every-reader. Rule: the Go codec and the cue model are two readers of one record and must accept exactly the same set of values. Fix the rule, not the instance: a shared testdata corpus of records with expected accept or reject, asserted by both flow_test (Parse) and cmd/vocabulary validate_test (cue). That also catches #233 adding config to one reader only. Also reject empty hashes in Parse. Prevalence: 1 divergence among 3 probed edge shapes.'
          family: record-roundtrips-every-reader
          round: 3
        - id: BR-13
          severity: Minor
          title: helptext/estimate.md:5 still says change-code refuses unless the Estimate block reconciles, with no flow qualification
          detail: 'This is the 2nd finding in doc-claim-contradicts-code. Rule: when a change makes a gate conditional, every sentence asserting that gate''s action is flow-qualified, and the enumeration comes from grepping the gate''s action verbs, not from a hand-copied surface list. PQ-1''s own regex matched estimate.md through estimate_hours, but the list dropped it. Measured: git grep over helptext, atlas, AGENTS.base.md and construct for change-code with refus, requir, demand, parses or asks-for, plus "estimate gate" and "plan-quality gate", gives about 15 hits; 1 is unqualified and false.'
          family: doc-claim-contradicts-code
          round: 3
        - id: BR-14
          severity: Minor
          title: The no-frontmatter branch of decideChangeCodeFlow decides full itself, bypassing Decide and its pin validation
          detail: 'This is the 2nd finding in decision-logic-restated. changecode_flow.go:35-39 returns full with an ad hoc Rule string and empty provenance. An invalid --flow fast, or a --flow quick pin, is silently ignored, and the info line prints "flow: full (: the issue has no frontmatter)". Rule: Decide is the only producer of a (Flow, Rule) pair; callers pass degenerate input (Recorded nil) through it and handle only "nowhere to write". Prevalence: 1 site; line 46 builds Decide''s input, which is fine.'
          family: decision-logic-restated
          round: 3
      boundary: M1
      blocked: true
    - "n": 4
      timestamp: "2026-09-17T21:21:22-07:00"
      agent: claude
      dispose:
        - id: BR-11
          disposition: addressed
          note: Scratch revert (write ran.content, skip the fresh read) turns TestRecordChangeCodeFlowKeepsConcurrentEdit and TestRecordChangeCodeFlowRefusesFlowChangingEdit red.
          round: 4
        - id: BR-12
          disposition: addressed
          note: Reverting Parse's seen[key] hash check to h != "" turns TestFlowRecordCorpus red on spec "" and done ""; the cue corpus test ran (cue on PATH) and passes.
          round: 4
        - id: BR-13
          disposition: addressed
          note: estimate.md:5, claim.md:17 and issue.md:93 are flow-qualified; re-running the recorded sweep finds no unqualified gate-action claim. A different verb class is raised separately.
          round: 4
        - id: BR-14
          disposition: addressed
          note: Restoring the ad hoc no-frontmatter branch turns TestDecideChangeCodeFlow and TestOnlyFlowPackageBuildsFlowValues red. The guard's holes are raised separately.
          round: 4
      findings:
        - id: BR-15
          severity: Minor
          title: change-code help and two atlas pages say the flow is recorded first, but since BR-5 it is written after the gates
          detail: 'This is the 3rd finding in doc-claim-contradicts-code. helptext/change-code.md:4 (step 0 infers and records), atlas/workflow/issue-lifecycle.md:54 (first records, then runs structural checks) and atlas/workflow/sdlc-binary.md:36 (First it infers and records) were written in 24ea2b9; 1bd3b03 moved the write after the gates and past the dry-run return without updating them. The BR-13 sweep keyed on gate-action verbs and could not see an effect-timing change. Rule: every behaviour change in a fix round names its effect verb and subject (records ... flow) and that pair joins the sweep regex in the same round, so the regex derives from the diff''s changed effects, not a fixed verb list. Prevalence: 5 sentences describe writing the flow record; the 3 that state an order are all false. The new flowDrift refusal is also undocumented in change-code.md.'
          family: doc-claim-contradicts-code
          round: 4
        - id: BR-16
          severity: Minor
          title: TestOnlyFlowPackageBuildsFlowValues misses field assignment and untyped Rule strings, so BR-14's own defect passes it
          detail: 'This is the 2nd finding in test-oracle-weaker-than-plan. Probe: rewriting BR-14''s original branch as var fl flow.Flow; fl.Kind = flow.Full, with rule set to a string literal, leaves the guard green; it scans only composite literals and conversions. Rule: a pin claimed for "only package X builds Y" must fail on every syntactic shape of the violation, so prefer compiler enforcement. Fix: make Flow''s fields unexported with accessors and have Decide be the only producer of Rule values (ARCH-ORDER structural enforcement), before M2''s close-time upgrade becomes the first new consumer.'
          family: test-oracle-weaker-than-plan
          round: 4
        - id: BR-17
          severity: Minor
          title: The shared flow corpus dropped Parse's duplicate-key reject branch when it replaced TestParseRejects
          detail: 'This is the 3rd finding in record-roundtrips-every-reader. Deleting the seen[k] duplicate check in a scratch copy leaves every flow test green; Go would then read {kind: full, kind: quick, provenance: operator} as quick/operator while cue rejects it (both reject at HEAD, probed). Rule: the corpus enumerates Parse''s reject branches; give each error return a reason code, tag reject rows with it, and assert every code has at least one row, so no branch can exist unpinned. Prevalence: 8 reject branches, 6 in the corpus; the unparseable-YAML branch is backstopped by the non-map check, so duplicate-key is the one real gap.'
          family: record-roundtrips-every-reader
          round: 4
      boundary: M1
      blocked: false
    - "n": 5
      timestamp: "2026-09-17T21:51:13-07:00"
      agent: claude
      findings:
        - id: BR-18
          severity: Important
          title: 'A full-recipe REWORK is not sticky: a fix that shrinks the net diff sends the next round back to the small-diff recipe, and the issue stays recorded quick'
          detail: 'closeFlowStep/decideCloseFlow (cmd/sdlc/closeflow.go:36,60) re-measure only the current base..HEAD net diff. A REWORK persists its round to the ledger, but gatestate.Round records nothing about which recipe ran. Reproduced in a scratch copy of 60962e2: 3 code files -> full review -> REWORK -> a commit deleting cmd/c.go -> the re-close dispatches the small-diff recipe, and the record stays {kind: quick}. The Spec/plan claim that the re-close "re-derives the upgrade from the same window" holds only when no commit lands between rounds. That is the one ordering TestCloseQuickReworkThenReclose exercises (ARCH-ORDER: one observable interleaving). Effects: the round after a deletion is judged under 3 principles, not 8. Open full-recipe findings go to a judge told to raise nothing outside those 3. M3 will count the issue as a pure quick row. Fix: stamp the recipe on each boundary round. Make "a prior round of this boundary ran the full recipe" a crossing, recorded at finalize like the others. Add a test for REWORK -> shrinking fix -> re-close. If the operator keeps today''s behaviour, record that as a Revision and correct the "same window" claim in the Spec, the plan Decisions and close.md.'
          family: round-decision-not-sticky
          round: 5
        - id: BR-19
          severity: Important
          title: .sdlc/shared-surfaces guards the declaration but not the code that enforces the shell, and ariadne's sdlc is rebuilt from the working tree, so a quick branch can loosen its own shell
          detail: 'This is the 2nd finding in family hand-listed-guard-incomplete. Rule: a guard that claims "the change cannot loosen X" must cover every artifact that defines X, and a test must derive that set rather than trust a list. Here X is the shell. Only the declaration is covered (base union head). Undeclared: flow/limits.go (MaxCodeFiles, MaxChangedLines), flow/shell.go and surfaces.go, churn/classify.go (IsCodeFile, isTestPath) and cmd/sdlc/closeflow.go. The operator''s sdlc shell function runs go build ./cmd/sdlc from the checkout, where the default in-place branch lives. So a one-line quick edit raising MaxChangedLines closes under the binary it just changed. This contradicts close.md:210, .sdlc/shared-surfaces:8 and the plan''s ARCH-SECURE bullet. The header''s own criterion ("what the judges are told ... the boundary-review procedure") also omits boundary-tail.md, which this diff created and which carries the boundary contract, and judge/prompts/. Fix: declare cmd/sdlc/internal/flow/, cmd/sdlc/internal/churn/classify.go and cmd/sdlc/closeflow.go, and decide on the judge prompt files. Extend TestRepoDeclarationParses to glob the shell''s implementation files and require Match for each.'
          family: hand-listed-guard-incomplete
          round: 5
        - id: BR-20
          severity: Minor
          title: The shell's prose says "tests and docs excluded" and "docs are never code", but IsCodeFile counts cmd/**/*.md, which IsDoc itself classifies as docs
          detail: 'This is the 4th finding in family doc-claim-contradicts-code. Rule: prose restating what a classifier or gate does must be rendered from the owning package, or pinned by a test against that classifier''s own exemplars. It must not be hand-written, and a test must not pin the wording alone. Instances: ShellSummary (flow/limits.go:20), rendered into start-plan, change-code help, close help and flowInfoLine; close.md:207-208; the limits.go:11 comment; and TestShellSummaryReadsTheConstants (flow_test.go:152), which pins the inaccurate "tests and docs". TestIsDoc marks cmd/sdlc/helptext/close.md as a doc, while TestIsCodeFile counts cmd/**/*.md as code. The plan''s Core concepts table drift is the same class (see plan revisions). Fix: own the exclusion clause in churn beside IsCodeFile, with a test checking each clause against IsCodeFile exemplars, including the embedded-markdown exception. ShellSummary should use it, and close.md should use QUICK_SHELL instead of restating it.'
          family: doc-claim-contradicts-code
          round: 5
        - id: BR-21
          severity: Minor
          title: A trailing-slash surface pattern containing a glob (lua/parley/*/) passes ParseSurfaces but Match compares it as a literal prefix, so it never matches
          detail: 'surfaces.go validates the pattern with path.Match, but Match uses strings.HasPrefix for the trailing-slash form, so the declared guard silently guards nothing. Verified: "lua/*/" validates without error, and HasPrefix of lua/parley/x.lua is false. Reject glob metacharacters in the directory form at parse time (that becomes a crossing, failing toward full), or match the directory form segment-wise.'
          family: accepted-input-never-effective
          round: 5
        - id: BR-22
          severity: Minor
          title: declarationAt re-implements the ls-tree presence plus blob read that already exists (gitx TrunkFile.entryOf/readFrom, pathTrackedAtHEAD)
          detail: ARCH-DRY. closeflow.go:160 is the third copy of the question "is ref:path present, absent, or unknown" (gitx/trunkfile.go:255-287, synctrunk.go:319). It is also the only copy without --end-of-options. Extract gitx.FileAt(ref, path) and route all three through it.
          family: existing-helper-not-reused
          round: 5
        - id: BR-23
          severity: Minor
          title: checkQuickDoneWhen calls f.skip("done-when"), a key closeFlags.skip has no case for, so it means "--force only" by silently falling through
          detail: closeflow.go:85 against close.go:111-137. The behaviour is intended, but it rests on the switch default, so a reader takes done-when for a real gate key, and a typo'd key elsewhere fails the same way. Write f.Force explicitly, or add a case so TestCloseFlags_Skip covers it.
          family: gate-key-undeclared
          round: 5
      boundary: M2
      blocked: true
    - "n": 6
      timestamp: "2026-09-17T22:12:00-07:00"
      agent: claude
      dispose:
        - id: BR-18
          disposition: addressed
          note: Recipe stamp (boundaryledger.go:182) + EarlierFullReview crossing (closeflow.go:80); dropping the crossing reddens TestCloseQuickReworkShrinkThenReclose, dropping the stamp reddens TestCloseQuickSmallDiffReworkStaysQuick; Spec/plan/close.md corrected.
          round: 6
        - id: BR-19
          disposition: addressed
          note: cmd/sdlc/ declared (covers flow, churn, closeflow, judge prompts, boundary-tail.md); narrowing it reddens TestRepoDeclarationCoversTheShell. The build closure outside cmd/sdlc is raised as a new finding.
          round: 6
        - id: BR-20
          disposition: addressed
          note: churn.CodeFileRule beside IsCodeFile, TestCodeFileRule checks exemplars incl. cmd/**/*.md; ShellSummary renders it; close.md no longer restates it.
          round: 6
        - id: BR-21
          disposition: addressed
          note: surfaces.go refuses a glob in a directory entry; removing the check reddens TestParseSurfacesRejectsGlobInDirectory.
          round: 6
        - id: BR-22
          disposition: not-addressed
          note: entryOf and declarationAt now share EntryAt, but TrunkFile.readFrom (trunkfile.go:283) still issues its own cat-file blob, identical to the new BlobAt; route it through BlobAt.
          round: 6
        - id: BR-23
          disposition: addressed
          note: closeflow.go:92 checks f.Force explicitly; behaviour unchanged and still pinned by TestCloseQuickEmptyDoneWhenRefuses.
          round: 6
      findings:
        - id: BR-24
          severity: Important
          title: The shell's shared-surface check sees only rename destinations, so moving a file out of a surface, or renaming .sdlc/shared-surfaces itself, stays quick
          detail: 'Measure matches Surfaces against gitx.DiffNames (git diff --name-only), which lists a rename''s destination only. Reproduced at sdlc close: git mv pkg/vocab/x.go other/x.go under a pkg/vocab/ declaration, and git mv of the declaration to .sdlc/shared-surfaces.old, both close {kind: quick} with the small-diff recipe. The second defeats the declaration self-match that Surfaces.Match, the file header and close.md promise, and disables the shell for later branches once merged. Enumeration in this window: Measure is the only DiffNames consumer the diff adds; code-file counting by destination is right, surface matching needs both sides (the atlas gate''s atlas/ split has the same class, pre-existing). Fix: match surfaces over both sides of renames/copies (--no-renames name list, or name-status with old and new paths), keep destination-only counting, add rename-out and rename-declaration fixtures.'
          family: diff-paths-drop-rename-source
          round: 6
        - id: BR-25
          severity: Important
          title: The gate-machinery declaration stops at cmd/sdlc/, but sdlc's build closure also includes pkg/frontmatter and go.mod/go.sum, and the covering test only walks cmd/sdlc
          detail: 'This is the 3rd finding in family hand-listed-guard-incomplete. Rule: the set a guard protects must be derived from the protected artifact''s real dependency graph, never from a directory the author picked. For the gate binary that is go list -deps ./cmd/sdlc restricted to this module, plus go.mod and go.sum. pkg/frontmatter (Split, used by internal/issue/frontmatter.go to read the frontmatter holding flow:) is compiled into sdlc and undeclared, so a quick branch editing it closes under the binary it changed, which contradicts the declaration header and close.md. Fix: declare pkg/frontmatter/ (or pkg/), go.mod, go.sum, and make TestRepoDeclarationCoversTheShell derive its set from go list -deps instead of WalkDir over cmd/sdlc.'
          family: hand-listed-guard-incomplete
          round: 6
      boundary: M2
      recipe: milestone-review
      blocked: true
    - "n": 7
      timestamp: "2026-09-17T22:29:42-07:00"
      agent: claude
      dispose:
        - id: BR-22
          disposition: addressed
          note: gitx.EntryAt/BlobAt (with --end-of-options) now serve TrunkFile.entryOf, TrunkFile.readFrom and close's declarationAt; pathTrackedAtHEAD keeps its injected gitRunner seam by design (plan Revisions). Residual EntryAt-then-BlobAt composition is a nit.
          round: 7
        - id: BR-24
          disposition: addressed
          note: Measure matches surfaces over gitx.DiffPathsBothSides (--no-renames -z) and counts code over destinations; reverting to diffFiles reddens TestCloseRenameOutOfASurfaceUpgrades and TestCloseRenamingTheDeclarationUpgrades.
          round: 7
        - id: BR-25
          disposition: addressed
          note: Declaration adds pkg/, go.mod, go.sum; TestRepoDeclarationCoversTheShell derives from go list -deps (Go + embed files) plus go.mod/go.sum; dropping pkg/ or go.mod reddens it. Residuals (Skip on go-list failure, go.work/vendor) raised separately as Minor.
          round: 7
      findings:
        - id: BR-26
          severity: Important
          title: ParseSurfaces accepts a bare directory, a double-star glob and a leading-bang pattern, none of which Match ever honours, so a declaration can guard nothing
          detail: 'Probed at head: "pkg/vocab" parses but Match("pkg/vocab/x.go") is false; "lua/parley/**" matches one level only (lua/parley/a/b.lua false); "!pkg/x.go" parses and never matches. Fail-open on exactly the #263 case, for the next declaration (parley). 2nd in family after BR-21. Rule: every form the parser accepts must take the effect its reader expects, or be refused at parse. Fix: refuse double-star and a leading bang, let a glob-free literal match itself or anything under it, optionally treat a pattern matching no tracked path at base or head as a crossing, and pin the forms with a Match table test (the fuzz oracle only checks no-panic).'
          family: accepted-input-never-effective
          round: 7
        - id: BR-27
          severity: Important
          title: Prose hand-restates the declared build-closure scope and the close gate-flag set, and this window changed both without sweeping them
          detail: '5th in family. Instances in this window: helptext/close.md:210-213 says ariadne declares cmd/sdlc/ as the shell''s own code (BR-25 made it the whole build closure); AGENTS.base.md:52 lists close guards and flags without --no-done-when-fresh (added here) or --no-ledger; atlas/workflow/sdlc-binary.md:891-895 says close has 8 gates, with a stale flag list. Rule: prose never restates a set, count or scope that code owns. It renders from the owner (a help token, e.g. from processmanual.GateCatalog) or points at it without restating, and any fix that changes such a fact sweeps restatements keyed on the old fact in the same round.'
          family: doc-claim-contradicts-code
          round: 7
        - id: BR-28
          severity: Minor
          title: TestRepoDeclarationCoversTheShell skips when go list fails, so the BR-25 guard can silently turn off
          detail: '3rd in family. Rule: a guard test fails when the derivation it rests on fails; replace t.Skipf with t.Fatalf. The plan says the closure is derived and pinned by this test, and a skip is neither.'
          family: test-oracle-weaker-than-plan
          round: 7
        - id: BR-29
          severity: Minor
          title: go.work, go.work.sum and vendor/ change what go build compiles but are not declared shared surfaces
          detail: '4th in family. go list -deps can only see inputs that exist today; a branch that adds go.work with a replace directive, or a vendor/ tree, changes sdlc''s binary without touching go.mod, go.sum or a package directory. Rule: the protected set is every input the go command reads in module mode, including files that do not exist yet. Declare go.work* and vendor/.'
          family: hand-listed-guard-incomplete
          round: 7
        - id: BR-30
          severity: Minor
          title: Measure gets code-file paths from quoted DiffNames output but surface paths from -z output, so non-ASCII doc and test paths count as code
          detail: git diff --name-only (and --numstat) quote non-ASCII paths ("docs/caf\303\251.md"), which IsDoc and isTestPath then fail to recognise, so they count as code and can upgrade a quick issue for nothing. The error is on the safe side (toward full), but the -z lesson cited on DiffPathsBothSides was applied to only one of the two listings Measure pairs. Use -z (or core.quotePath=false) for the code-file list too.
          family: porcelain-output-parsed-as-data
          round: 7
      boundary: M2
      recipe: milestone-review
      blocked: true
    - "n": 8
      timestamp: "2026-09-17T22:50:41-07:00"
      agent: claude
      dispose:
        - id: BR-26
          disposition: addressed
          note: Literal-or-subtree branch and the refusals of a leading bang, double-star and globbed dir entries are pinned by TestSurfaceForms and the extended rejection table; reverting either reddens its test (verified in a scratch export). Residual degenerate forms raised separately.
          round: 8
        - id: BR-27
          disposition: not-addressed
          note: 'Named sites fixed, class not: the new test is page-granular. close.md FLAGS (262-276) and milestone-close.md FLAGS (78-81, written this round) both omit --no-ledger and pass because it appears elsewhere on each page; atlas sdlc-binary.md:893-894 claims every list is pinned complete; sdlc-binary.md:754 still restates "12 gates / 16 sigs" against a 20-row GateCatalog. Render per-command gate-flag lists from GateCatalog via a help token, or make the oracle list-granular.'
          round: 8
        - id: BR-28
          disposition: addressed
          note: surfaces_test.go:125 now t.Fatalf on a go list failure.
          round: 8
        - id: BR-29
          disposition: addressed
          note: go.work, go.work.sum and vendor/ declared and in the test's shell list; removing go.work and vendor/ from the declaration reddens TestRepoDeclarationCoversTheShell.
          round: 8
        - id: BR-30
          disposition: addressed
          note: DiffNames uses -z; reverting it reddens TestCloseNonASCIIPathsClassifyAsThemselves. The numstat half is untested and raised separately.
          round: 8
      findings:
        - id: BR-31
          severity: Important
          title: BR-26 changed the shared-surface grammar, but the declaration header and the atlas still describe the old two-form grammar
          detail: 'This is the 6th finding in family doc-claim-contradicts-code. .sdlc/shared-surfaces:5-6 says every non-directory line is a path.Match glob, so pkg/vocab would mean exactly that path, and it omits the refusals; atlas/workflow/sdlc-binary.md:299 says the same. The code (surfaces.go:18-31,68-84) now treats a glob-free literal as path-or-subtree and refuses the double-star and leading-bang forms. The rule BR-27 stated (a fix that changes a fact sweeps restatements keyed on it in the same round) was violated by the same commit. Rule-level fix: give the grammar one owner, e.g. a flow.SurfaceForms string rendered into sdlc close --help via a token and pinned clause-by-clause against TestSurfaceForms (the churn.CodeFileRule / TestCodeFileRule pattern). Point the declaration header and atlas at it, and sweep in the same round with git grep for path.Match, trailing-slash and for-a-subtree phrasing.'
          family: doc-claim-contradicts-code
          round: 8
        - id: BR-32
          severity: Minor
          title: 'The window numstat is still parsed as quoted output: its BR-30 change has no red test, and core.quotePath=false still quotes some paths'
          detail: 'This is the 2nd finding in family porcelain-output-parsed-as-data. Reverting the quotePath change in windowFileStats (closeflow.go:142) leaves the whole suite green; a probe with cmd/über.go at 150 lines then closes quick. The plan''s "mutation-checked" claim covers only DiffNames. At HEAD, cmd/a"b.go at 150 lines closes quick, because quotePath=false still quotes a double-quote, a backslash and control characters, and the quoted row never joins the -z code-file list. Rule: every git listing a gate reads as data uses -z. Use diff --numstat -z (renames arrive as separate NUL fields) with a -z parse, and add a close test with a git-quoted code-file name over the line limit.'
          family: porcelain-output-parsed-as-data
          round: 8
        - id: BR-33
          severity: Minor
          title: ParseSurfaces still accepts root and non-canonical patterns (/, ./, ., pkg/./vocab, pkg//vocab/) that match nothing
          detail: 'This is the 3rd finding in family accepted-input-never-effective. Probed at head: each parses without error and matches no path. The forms are refused by hand-enumeration while the fuzz oracle still checks only no-panic, so the next form slips too. Rule: accept only canonical, non-empty patterns (path.Clean(p) == p and p is not "."). Make the oracle a property instead of a table: FuzzParseSurfaces asserts that an accepted glob-free pattern matches itself, and a directory entry matches pattern + "x". Also decide whether a glob naming a directory (pkg/v*) covers its subtree the way the literal form now does, or document the asymmetry.'
          family: accepted-input-never-effective
          round: 8
      boundary: M2
      recipe: milestone-review
      blocked: false
    - "n": 9
      timestamp: "2026-09-18T19:07:07-07:00"
      agent: claude
      dispose:
        - id: BR-15
          disposition: addressed
          note: change-code.md step 0, issue-lifecycle.md:54 and sdlc-binary.md:36 now say the record is written after the gates; the flowDrift refusal is documented under THE FLOW.
          round: 9
        - id: BR-16
          disposition: addressed
          note: Flow fields are unexported with accessors (flow.go:52-56) and Rule is opaque (flow.go:219-225), so the compiler enforces the single producer; the source-scan guard is gone.
          round: 9
        - id: BR-17
          disposition: addressed
          note: 'Reason codes plus the every-reason-has-a-row assertion (flow_test.go:205-233). Probed: disabling the seen[k] check turns TestFlowRecordCorpus red on both assertions.'
          round: 9
        - id: BR-27
          disposition: addressed
          note: 'AGENTS.base.md:52 points at the catalog; help pages render the gate flags; the build-closure prose went with shared surfaces. Probed: hand-listing a gate flag in close.md turns TestGateFlagListsAreRenderedFromTheCatalog red.'
          round: 9
        - id: BR-31
          disposition: withdrawn
          note: 'Overtaken by design: f0141e0 removed shared surfaces and their declaration, so no grammar is left to describe.'
          round: 9
        - id: BR-32
          disposition: not-addressed
          note: '-z is in place (closeflow.go:141), but TestCloseQuotedNameCodeFileLinesCount cannot fail. It uses a code file, and IsCodeFile''s CodeProd default counts the quoted path as code either way; once f0141e0 dropped the -z name-list intersection, the test lost its teeth. Probed: restoring `-c core.quotePath=false diff --numstat` with ParseNumstat leaves it green. The discriminating fixture goes red: docs/a"b.md and tests/x"y_spec.lua at 150 lines beside a 5-line cmd/a.go, expecting quick. Plan:503''s "the test reddens without -z" is false at HEAD.'
          round: 9
        - id: BR-33
          disposition: withdrawn
          note: 'Overtaken by design: ParseSurfaces and the fuzz target were removed with shared surfaces (f0141e0).'
          round: 9
      findings:
        - id: BR-34
          severity: Minor
          title: 'Two runtime messages and the publish gate''s atlas page restate rules this window changed: FIX-THEN-SHIP''s unconditional "do not re-run close", and the "#187 column set" header message'
          detail: 'This is the 7th finding in doc-claim-contradicts-code. Instances: close.go:1871-1873 says "Do NOT re-run sdlc close — this verdict already sanctions shipping after the fixes", yet publishgate.go:176 now refuses a quick issue past MaxAddedLinesAfterReview and sends it back to close; atlas/workflow/pre-merge-checks.md:24-46 (the publish gate''s page) lists its refusals and that protocol without the quick re-measure; close.go:981 prints "header upgraded to the #187 column set", and the only upgrade left to fire is #187 to #231. Enumeration: 7 surfaces describe what the publish gate refuses or what the verdict sanctions (close.md:222, merge.md:35, push.md:10, sdlc-binary.md:303 and publishgate.go:6 derive; close.go:1871 and pre-merge-checks.md do not). 4 surfaces name the ledger''s column set, and close.go:981 is wrong. Prevalence is 3 of 11. Why earlier sweeps missed these: they keyed on the new mechanism''s name, while these sentences state the old conclusion. Rule: a runtime string that states a policy renders it from the policy''s owner, never a literal. formatFixThenShipProtocol takes the flow and renders flow.AfterReviewSummary() on quick; the header message states no version. When a change adds a policy function, the sweep greps for the old policy''s conclusion ("do not re-run", "landed after", "#187"), not the new gate''s name (ARCH-PURPOSE shadow sweep).'
          family: doc-claim-contradicts-code
          round: 9
        - id: BR-35
          severity: Minor
          title: The publish check's new quick-growth refusal matches no GateCatalog RefusalPat, so the friction instrument never counts it
          detail: 'This is the 2nd finding in gate-key-undeclared. The merge/push no-judge RefusalPat is `publish gate: \d+ commit\(s\) landed after` (gatesig.go:150,156). "publish gate: #N is on the quick flow, and its diff grew…" (publishgate.go:180) is unattributed by classifyOutputLine, although --no-judge waives it. Enumeration of refusals added in this window that a catalogued flag waives: DoneWhenFresh (catalogued) and the publish quick-growth refusal (not catalogued), so 1 of 2. Rule: every refusal a catalogued flag waives must be attributable by the catalog. The tests that produce a refusal assert that their (command, flag) RefusalPat matches it, the inverse of assertNoGatesigCollision. Fix: widen the two patterns to `publish gate: (\d+ commit\(s\) landed after|#\d+ is on the quick flow)` and add that assertion to TestRunPublishGate_QuickGrewPastReview.'
          family: gate-key-undeclared
          round: 9
      recipe: milestone-review
      blocked: false
---

# Gate ledger — ariadne#231 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-09-17T20:53:40-07:00 (sdlc) — passed

### Raised

- **BR-1** [Minor] `test-prose-enumeration` Tests are listed case by case in prose; the risky parsers have no one-line adversarial strategy
  Name each test and replace the case lists with one line per risky function. For example: fuzz flow.FromFrontmatter seeded with malformed, nested and duplicate-key maps (never panics, malformed resolves to full); fuzz ParseSurfaces (a bad pattern becomes a crossing); DoneWhenFresh section extraction over fenced headings.
  (carried from plan-quality PQ-8, deferred to the boundary review)
- **BR-2** [Minor] `record-roundtrips-every-reader` Unquoted 8-hex hashes in the flow record become YAML ints, and cue checks at push/merge refuse them
  Verified with local cue: spec 12345678 gives "mismatched types int and string" against the plan's own #Flow. About 2.3 percent per hash, roughly 4.6 percent of quick records. validategate.go runs this on every changed issue at push and merge, and yaml.v3 hides it on the Go side. Rule: a persisted value must parse to the same type under every reader. Fix: Format quotes the hashes, and the round-trip test, FuzzFromFrontmatter and the validate_test flow case are seeded with an all-digit hash.
  (carried from plan-quality PQ-9, deferred to the boundary review)
- **BR-3** [Minor] `acceptance-oracle-freshness` Spec section 4 and Done-when still say quick yes/no after the plan renamed it quick-flow, and Done-when's Flow omits the hash fields
  Third finding in this family. The rule for all of them: a revision that renames or reshapes an artifact Done-when names restates that bullet in the same edit. Apply it now by checking every name in Done-when against the plan (both drifts here) before implementing. DoneWhenFresh will enforce this mechanically once it ships.
  (carried from plan-quality PQ-10, deferred to the boundary review)

## Round 2 — 2026-09-17T20:53:40-07:00 (claude) — BLOCKED

### Raised

- **BR-4** [Important] `hand-listed-guard-incomplete` decideChangeCodeFlow's PlanItemsBody routing is unpinned — not in planItemBodySources, no fenced-Mx test
  changecode_flow.go:48 hands a plan body to the exempt helper issue.MilestonesInPlanOrder, whose exemption says callers are covered by planItemBodySources (commitpathspec_guard_test.go:341), which lists only close.go. Reverting to PlanSectionBody in a scratch copy left every Flow/Guard/PlanItem/ChangeCode test green. Add a fenced-Mx case to TestDecideChangeCodeFlow (must stay quick) and sweep all MilestonesInPlanOrder callers (ARCH-PURPOSE).
- **BR-5** [Important] `state-persisted-on-refused-transition` change-code writes the flow record before its gates, so a refused run leaves a sticky full/inferred
  applyChangeCodeFlow runs at changecode.go:137 before the gate loop; a structural/plan-quality/estimate refusal exits with flow full/inferred written (uncommitted), and Decide keeps a recorded full, so removing the plan or Mx rows and re-running stays full. The plan's ARCH-ORDER table says REWORK writes nothing. Decide before the gates, write after they pass (before syncIssue); test that a refused run leaves the file byte-identical (ARCH-ORDER).
- **BR-6** [Minor] `doc-claim-contradicts-code` M1 docs describe close-time flow behaviour (small-diff review, upgrade, Done-when refusal) that M2 hasn't built
  helptext/change-code.md THE FLOW, helptext/issue.md flow field, AGENTS.base.md section 2, atlas issue-lifecycle item 6 and flowInfoLine state it in the present tense; at HEAD close ignores the flow entirely. Acceptable only if M2 lands before any merge; note it in the Log.
- **BR-7** [Minor] `decision-logic-restated` flowReason re-derives Decide's rule order instead of Decide returning the rule that fired
  changecode_flow.go:66-81 mirrors flow.go Decide rules; a reorder in Decide silently mislabels the reason (ARCH-DRY).
- **BR-8** [Minor] `untrusted-input-rewrite` Rewriting a block-form hand-edited flow record orphans its child lines into invalid frontmatter
  SetField replaces only the flow line; a probe produced flow full/inferred followed by indented kind/provenance lines, while the warning says the record is being rewritten (ARCH-SECURE).
- **BR-9** [Minor] `test-oracle-weaker-than-plan` FuzzFromFrontmatter does not assert that the error path resolves to full
  The plan names "every error path resolves to full" as the fuzz oracle; the body returns on error without checking Kind == Full.
- **BR-10** [Minor] `unrelated-churn` issue.cue carries unrelated cue-fmt whitespace realignment of categories, Status and Transition

## Round 3 — 2026-09-17T21:07:39-07:00 (claude) — BLOCKED

### Disposed

- BR-1 — addressed — Plan names every test. FuzzFromFrontmatter plus the fenced-heading case in TestContractHashes cover M1's parsers; FuzzParseSurfaces is M2 scope; keeping case lists is a recorded decision.
- BR-2 — addressed — Format quotes hashes with %q (flow.go:66-72); scratch revert to %s turns TestFlowRoundTrip red on 12345678; validate_test covers quoted-valid and unquoted-invalid through real cue.
- BR-3 — addressed — Spec section 4 and Done-when say quick-flow and the Flow bullet carries spec/done; old names survive only in dated Revisions.
- BR-4 — addressed — Scratch revert of changecode_flow.go:50 or close.go:1719 to PlanSectionBody turns TestPlanItemReadersUsePlanItemsBody red; the fenced-Mx case in TestDecideChangeCodeFlow also goes red.
- BR-5 — addressed — Moving recordChangeCodeFlow before the gate loop turns TestRunChangeCodeRecordsFlowAfterGates red. The fix opened a stale-snapshot window, raised separately.
- BR-6 — addressed — The 2026-09-17 "M1 built" Log entry records that the present-tense close docs hold only because M2 lands before merge.
- BR-7 — addressed — Decide returns the Rule that fired, flowReason is gone, and TestDecide asserts the rule for each case.
- BR-8 — addressed — SetField drops the indented block continuation (frontmatter.go:73-99), pinned by TestSetFieldReplacesBlockValue.
- BR-9 — addressed — Scratch revert of FromFrontmatter's error value to Flow{} turns FuzzFromFrontmatter's seed corpus red.
- BR-10 — addressed — The issue.cue diff is now 17 pure additions, with no realignment.

### Raised

- **BR-11** [Important] `write-from-stale-snapshot` change-code writes the flow record from issue text read before the gates, clobbering edits made while plan-quality ran
  recordChangeCodeFlow (changecode_flow.go:100-108) writes d.content, built at changecode.go:138 from the step-2 read, after the multi-minute gate loop. A scratch probe (write, report, edit the file on disk, record) lost the edit, and syncIssue then commits it and, on main, pushes it. It fires on every issue's first successful run, because the record is always new. The repo lock does not cover editor or agent edits; close handles the same hazard by validating its snapshot after the review (close.go:1218-1231). Fix: re-read issuePath at record time and apply SetField to the fresh frontmatter, recomputing quick hashes over the fresh body, or refuse as close does. Add the probe as a regression test (ARCH-ORDER, second actor on the same state).
- **BR-12** [Minor] `record-roundtrips-every-reader` flow.Parse accepts spec "" and done "" that cue Flow rejects, contradicting Parse's own doc comment
  This is the 2nd finding in record-roundtrips-every-reader. Rule: the Go codec and the cue model are two readers of one record and must accept exactly the same set of values. Fix the rule, not the instance: a shared testdata corpus of records with expected accept or reject, asserted by both flow_test (Parse) and cmd/vocabulary validate_test (cue). That also catches #233 adding config to one reader only. Also reject empty hashes in Parse. Prevalence: 1 divergence among 3 probed edge shapes.
- **BR-13** [Minor] `doc-claim-contradicts-code` helptext/estimate.md:5 still says change-code refuses unless the Estimate block reconciles, with no flow qualification
  This is the 2nd finding in doc-claim-contradicts-code. Rule: when a change makes a gate conditional, every sentence asserting that gate's action is flow-qualified, and the enumeration comes from grepping the gate's action verbs, not from a hand-copied surface list. PQ-1's own regex matched estimate.md through estimate_hours, but the list dropped it. Measured: git grep over helptext, atlas, AGENTS.base.md and construct for change-code with refus, requir, demand, parses or asks-for, plus "estimate gate" and "plan-quality gate", gives about 15 hits; 1 is unqualified and false.
- **BR-14** [Minor] `decision-logic-restated` The no-frontmatter branch of decideChangeCodeFlow decides full itself, bypassing Decide and its pin validation
  This is the 2nd finding in decision-logic-restated. changecode_flow.go:35-39 returns full with an ad hoc Rule string and empty provenance. An invalid --flow fast, or a --flow quick pin, is silently ignored, and the info line prints "flow: full (: the issue has no frontmatter)". Rule: Decide is the only producer of a (Flow, Rule) pair; callers pass degenerate input (Recorded nil) through it and handle only "nowhere to write". Prevalence: 1 site; line 46 builds Decide's input, which is fine.

## Round 4 — 2026-09-17T21:21:22-07:00 (claude) — passed

### Disposed

- BR-11 — addressed — Scratch revert (write ran.content, skip the fresh read) turns TestRecordChangeCodeFlowKeepsConcurrentEdit and TestRecordChangeCodeFlowRefusesFlowChangingEdit red.
- BR-12 — addressed — Reverting Parse's seen[key] hash check to h != "" turns TestFlowRecordCorpus red on spec "" and done ""; the cue corpus test ran (cue on PATH) and passes.
- BR-13 — addressed — estimate.md:5, claim.md:17 and issue.md:93 are flow-qualified; re-running the recorded sweep finds no unqualified gate-action claim. A different verb class is raised separately.
- BR-14 — addressed — Restoring the ad hoc no-frontmatter branch turns TestDecideChangeCodeFlow and TestOnlyFlowPackageBuildsFlowValues red. The guard's holes are raised separately.

### Raised

- **BR-15** [Minor] `doc-claim-contradicts-code` change-code help and two atlas pages say the flow is recorded first, but since BR-5 it is written after the gates
  This is the 3rd finding in doc-claim-contradicts-code. helptext/change-code.md:4 (step 0 infers and records), atlas/workflow/issue-lifecycle.md:54 (first records, then runs structural checks) and atlas/workflow/sdlc-binary.md:36 (First it infers and records) were written in 24ea2b9; 1bd3b03 moved the write after the gates and past the dry-run return without updating them. The BR-13 sweep keyed on gate-action verbs and could not see an effect-timing change. Rule: every behaviour change in a fix round names its effect verb and subject (records ... flow) and that pair joins the sweep regex in the same round, so the regex derives from the diff's changed effects, not a fixed verb list. Prevalence: 5 sentences describe writing the flow record; the 3 that state an order are all false. The new flowDrift refusal is also undocumented in change-code.md.
- **BR-16** [Minor] `test-oracle-weaker-than-plan` TestOnlyFlowPackageBuildsFlowValues misses field assignment and untyped Rule strings, so BR-14's own defect passes it
  This is the 2nd finding in test-oracle-weaker-than-plan. Probe: rewriting BR-14's original branch as var fl flow.Flow; fl.Kind = flow.Full, with rule set to a string literal, leaves the guard green; it scans only composite literals and conversions. Rule: a pin claimed for "only package X builds Y" must fail on every syntactic shape of the violation, so prefer compiler enforcement. Fix: make Flow's fields unexported with accessors and have Decide be the only producer of Rule values (ARCH-ORDER structural enforcement), before M2's close-time upgrade becomes the first new consumer.
- **BR-17** [Minor] `record-roundtrips-every-reader` The shared flow corpus dropped Parse's duplicate-key reject branch when it replaced TestParseRejects
  This is the 3rd finding in record-roundtrips-every-reader. Deleting the seen[k] duplicate check in a scratch copy leaves every flow test green; Go would then read {kind: full, kind: quick, provenance: operator} as quick/operator while cue rejects it (both reject at HEAD, probed). Rule: the corpus enumerates Parse's reject branches; give each error return a reason code, tag reject rows with it, and assert every code has at least one row, so no branch can exist unpinned. Prevalence: 8 reject branches, 6 in the corpus; the unparseable-YAML branch is backstopped by the non-map check, so duplicate-key is the one real gap.

## Round 5 — 2026-09-17T21:51:13-07:00 (claude) — BLOCKED

### Raised

- **BR-18** [Important] `round-decision-not-sticky` A full-recipe REWORK is not sticky: a fix that shrinks the net diff sends the next round back to the small-diff recipe, and the issue stays recorded quick
  closeFlowStep/decideCloseFlow (cmd/sdlc/closeflow.go:36,60) re-measure only the current base..HEAD net diff. A REWORK persists its round to the ledger, but gatestate.Round records nothing about which recipe ran. Reproduced in a scratch copy of 60962e2: 3 code files -> full review -> REWORK -> a commit deleting cmd/c.go -> the re-close dispatches the small-diff recipe, and the record stays {kind: quick}. The Spec/plan claim that the re-close "re-derives the upgrade from the same window" holds only when no commit lands between rounds. That is the one ordering TestCloseQuickReworkThenReclose exercises (ARCH-ORDER: one observable interleaving). Effects: the round after a deletion is judged under 3 principles, not 8. Open full-recipe findings go to a judge told to raise nothing outside those 3. M3 will count the issue as a pure quick row. Fix: stamp the recipe on each boundary round. Make "a prior round of this boundary ran the full recipe" a crossing, recorded at finalize like the others. Add a test for REWORK -> shrinking fix -> re-close. If the operator keeps today's behaviour, record that as a Revision and correct the "same window" claim in the Spec, the plan Decisions and close.md.
- **BR-19** [Important] `hand-listed-guard-incomplete` .sdlc/shared-surfaces guards the declaration but not the code that enforces the shell, and ariadne's sdlc is rebuilt from the working tree, so a quick branch can loosen its own shell
  This is the 2nd finding in family hand-listed-guard-incomplete. Rule: a guard that claims "the change cannot loosen X" must cover every artifact that defines X, and a test must derive that set rather than trust a list. Here X is the shell. Only the declaration is covered (base union head). Undeclared: flow/limits.go (MaxCodeFiles, MaxChangedLines), flow/shell.go and surfaces.go, churn/classify.go (IsCodeFile, isTestPath) and cmd/sdlc/closeflow.go. The operator's sdlc shell function runs go build ./cmd/sdlc from the checkout, where the default in-place branch lives. So a one-line quick edit raising MaxChangedLines closes under the binary it just changed. This contradicts close.md:210, .sdlc/shared-surfaces:8 and the plan's ARCH-SECURE bullet. The header's own criterion ("what the judges are told ... the boundary-review procedure") also omits boundary-tail.md, which this diff created and which carries the boundary contract, and judge/prompts/. Fix: declare cmd/sdlc/internal/flow/, cmd/sdlc/internal/churn/classify.go and cmd/sdlc/closeflow.go, and decide on the judge prompt files. Extend TestRepoDeclarationParses to glob the shell's implementation files and require Match for each.
- **BR-20** [Minor] `doc-claim-contradicts-code` The shell's prose says "tests and docs excluded" and "docs are never code", but IsCodeFile counts cmd/**/*.md, which IsDoc itself classifies as docs
  This is the 4th finding in family doc-claim-contradicts-code. Rule: prose restating what a classifier or gate does must be rendered from the owning package, or pinned by a test against that classifier's own exemplars. It must not be hand-written, and a test must not pin the wording alone. Instances: ShellSummary (flow/limits.go:20), rendered into start-plan, change-code help, close help and flowInfoLine; close.md:207-208; the limits.go:11 comment; and TestShellSummaryReadsTheConstants (flow_test.go:152), which pins the inaccurate "tests and docs". TestIsDoc marks cmd/sdlc/helptext/close.md as a doc, while TestIsCodeFile counts cmd/**/*.md as code. The plan's Core concepts table drift is the same class (see plan revisions). Fix: own the exclusion clause in churn beside IsCodeFile, with a test checking each clause against IsCodeFile exemplars, including the embedded-markdown exception. ShellSummary should use it, and close.md should use QUICK_SHELL instead of restating it.
- **BR-21** [Minor] `accepted-input-never-effective` A trailing-slash surface pattern containing a glob (lua/parley/*/) passes ParseSurfaces but Match compares it as a literal prefix, so it never matches
  surfaces.go validates the pattern with path.Match, but Match uses strings.HasPrefix for the trailing-slash form, so the declared guard silently guards nothing. Verified: "lua/*/" validates without error, and HasPrefix of lua/parley/x.lua is false. Reject glob metacharacters in the directory form at parse time (that becomes a crossing, failing toward full), or match the directory form segment-wise.
- **BR-22** [Minor] `existing-helper-not-reused` declarationAt re-implements the ls-tree presence plus blob read that already exists (gitx TrunkFile.entryOf/readFrom, pathTrackedAtHEAD)
  ARCH-DRY. closeflow.go:160 is the third copy of the question "is ref:path present, absent, or unknown" (gitx/trunkfile.go:255-287, synctrunk.go:319). It is also the only copy without --end-of-options. Extract gitx.FileAt(ref, path) and route all three through it.
- **BR-23** [Minor] `gate-key-undeclared` checkQuickDoneWhen calls f.skip("done-when"), a key closeFlags.skip has no case for, so it means "--force only" by silently falling through
  closeflow.go:85 against close.go:111-137. The behaviour is intended, but it rests on the switch default, so a reader takes done-when for a real gate key, and a typo'd key elsewhere fails the same way. Write f.Force explicitly, or add a case so TestCloseFlags_Skip covers it.

## Round 6 — 2026-09-17T22:12:00-07:00 (claude) — BLOCKED

### Disposed

- BR-18 — addressed — Recipe stamp (boundaryledger.go:182) + EarlierFullReview crossing (closeflow.go:80); dropping the crossing reddens TestCloseQuickReworkShrinkThenReclose, dropping the stamp reddens TestCloseQuickSmallDiffReworkStaysQuick; Spec/plan/close.md corrected.
- BR-19 — addressed — cmd/sdlc/ declared (covers flow, churn, closeflow, judge prompts, boundary-tail.md); narrowing it reddens TestRepoDeclarationCoversTheShell. The build closure outside cmd/sdlc is raised as a new finding.
- BR-20 — addressed — churn.CodeFileRule beside IsCodeFile, TestCodeFileRule checks exemplars incl. cmd/**/*.md; ShellSummary renders it; close.md no longer restates it.
- BR-21 — addressed — surfaces.go refuses a glob in a directory entry; removing the check reddens TestParseSurfacesRejectsGlobInDirectory.
- BR-22 — not-addressed — entryOf and declarationAt now share EntryAt, but TrunkFile.readFrom (trunkfile.go:283) still issues its own cat-file blob, identical to the new BlobAt; route it through BlobAt.
- BR-23 — addressed — closeflow.go:92 checks f.Force explicitly; behaviour unchanged and still pinned by TestCloseQuickEmptyDoneWhenRefuses.

### Raised

- **BR-24** [Important] `diff-paths-drop-rename-source` The shell's shared-surface check sees only rename destinations, so moving a file out of a surface, or renaming .sdlc/shared-surfaces itself, stays quick
  Measure matches Surfaces against gitx.DiffNames (git diff --name-only), which lists a rename's destination only. Reproduced at sdlc close: git mv pkg/vocab/x.go other/x.go under a pkg/vocab/ declaration, and git mv of the declaration to .sdlc/shared-surfaces.old, both close {kind: quick} with the small-diff recipe. The second defeats the declaration self-match that Surfaces.Match, the file header and close.md promise, and disables the shell for later branches once merged. Enumeration in this window: Measure is the only DiffNames consumer the diff adds; code-file counting by destination is right, surface matching needs both sides (the atlas gate's atlas/ split has the same class, pre-existing). Fix: match surfaces over both sides of renames/copies (--no-renames name list, or name-status with old and new paths), keep destination-only counting, add rename-out and rename-declaration fixtures.
- **BR-25** [Important] `hand-listed-guard-incomplete` The gate-machinery declaration stops at cmd/sdlc/, but sdlc's build closure also includes pkg/frontmatter and go.mod/go.sum, and the covering test only walks cmd/sdlc
  This is the 3rd finding in family hand-listed-guard-incomplete. Rule: the set a guard protects must be derived from the protected artifact's real dependency graph, never from a directory the author picked. For the gate binary that is go list -deps ./cmd/sdlc restricted to this module, plus go.mod and go.sum. pkg/frontmatter (Split, used by internal/issue/frontmatter.go to read the frontmatter holding flow:) is compiled into sdlc and undeclared, so a quick branch editing it closes under the binary it changed, which contradicts the declaration header and close.md. Fix: declare pkg/frontmatter/ (or pkg/), go.mod, go.sum, and make TestRepoDeclarationCoversTheShell derive its set from go list -deps instead of WalkDir over cmd/sdlc.

## Round 7 — 2026-09-17T22:29:42-07:00 (claude) — BLOCKED

### Disposed

- BR-22 — addressed — gitx.EntryAt/BlobAt (with --end-of-options) now serve TrunkFile.entryOf, TrunkFile.readFrom and close's declarationAt; pathTrackedAtHEAD keeps its injected gitRunner seam by design (plan Revisions). Residual EntryAt-then-BlobAt composition is a nit.
- BR-24 — addressed — Measure matches surfaces over gitx.DiffPathsBothSides (--no-renames -z) and counts code over destinations; reverting to diffFiles reddens TestCloseRenameOutOfASurfaceUpgrades and TestCloseRenamingTheDeclarationUpgrades.
- BR-25 — addressed — Declaration adds pkg/, go.mod, go.sum; TestRepoDeclarationCoversTheShell derives from go list -deps (Go + embed files) plus go.mod/go.sum; dropping pkg/ or go.mod reddens it. Residuals (Skip on go-list failure, go.work/vendor) raised separately as Minor.

### Raised

- **BR-26** [Important] `accepted-input-never-effective` ParseSurfaces accepts a bare directory, a double-star glob and a leading-bang pattern, none of which Match ever honours, so a declaration can guard nothing
  Probed at head: "pkg/vocab" parses but Match("pkg/vocab/x.go") is false; "lua/parley/**" matches one level only (lua/parley/a/b.lua false); "!pkg/x.go" parses and never matches. Fail-open on exactly the #263 case, for the next declaration (parley). 2nd in family after BR-21. Rule: every form the parser accepts must take the effect its reader expects, or be refused at parse. Fix: refuse double-star and a leading bang, let a glob-free literal match itself or anything under it, optionally treat a pattern matching no tracked path at base or head as a crossing, and pin the forms with a Match table test (the fuzz oracle only checks no-panic).
- **BR-27** [Important] `doc-claim-contradicts-code` Prose hand-restates the declared build-closure scope and the close gate-flag set, and this window changed both without sweeping them
  5th in family. Instances in this window: helptext/close.md:210-213 says ariadne declares cmd/sdlc/ as the shell's own code (BR-25 made it the whole build closure); AGENTS.base.md:52 lists close guards and flags without --no-done-when-fresh (added here) or --no-ledger; atlas/workflow/sdlc-binary.md:891-895 says close has 8 gates, with a stale flag list. Rule: prose never restates a set, count or scope that code owns. It renders from the owner (a help token, e.g. from processmanual.GateCatalog) or points at it without restating, and any fix that changes such a fact sweeps restatements keyed on the old fact in the same round.
- **BR-28** [Minor] `test-oracle-weaker-than-plan` TestRepoDeclarationCoversTheShell skips when go list fails, so the BR-25 guard can silently turn off
  3rd in family. Rule: a guard test fails when the derivation it rests on fails; replace t.Skipf with t.Fatalf. The plan says the closure is derived and pinned by this test, and a skip is neither.
- **BR-29** [Minor] `hand-listed-guard-incomplete` go.work, go.work.sum and vendor/ change what go build compiles but are not declared shared surfaces
  4th in family. go list -deps can only see inputs that exist today; a branch that adds go.work with a replace directive, or a vendor/ tree, changes sdlc's binary without touching go.mod, go.sum or a package directory. Rule: the protected set is every input the go command reads in module mode, including files that do not exist yet. Declare go.work* and vendor/.
- **BR-30** [Minor] `porcelain-output-parsed-as-data` Measure gets code-file paths from quoted DiffNames output but surface paths from -z output, so non-ASCII doc and test paths count as code
  git diff --name-only (and --numstat) quote non-ASCII paths ("docs/caf\303\251.md"), which IsDoc and isTestPath then fail to recognise, so they count as code and can upgrade a quick issue for nothing. The error is on the safe side (toward full), but the -z lesson cited on DiffPathsBothSides was applied to only one of the two listings Measure pairs. Use -z (or core.quotePath=false) for the code-file list too.

## Round 8 — 2026-09-17T22:50:41-07:00 (claude) — passed

### Disposed

- BR-26 — addressed — Literal-or-subtree branch and the refusals of a leading bang, double-star and globbed dir entries are pinned by TestSurfaceForms and the extended rejection table; reverting either reddens its test (verified in a scratch export). Residual degenerate forms raised separately.
- BR-27 — not-addressed — Named sites fixed, class not: the new test is page-granular. close.md FLAGS (262-276) and milestone-close.md FLAGS (78-81, written this round) both omit --no-ledger and pass because it appears elsewhere on each page; atlas sdlc-binary.md:893-894 claims every list is pinned complete; sdlc-binary.md:754 still restates "12 gates / 16 sigs" against a 20-row GateCatalog. Render per-command gate-flag lists from GateCatalog via a help token, or make the oracle list-granular.
- BR-28 — addressed — surfaces_test.go:125 now t.Fatalf on a go list failure.
- BR-29 — addressed — go.work, go.work.sum and vendor/ declared and in the test's shell list; removing go.work and vendor/ from the declaration reddens TestRepoDeclarationCoversTheShell.
- BR-30 — addressed — DiffNames uses -z; reverting it reddens TestCloseNonASCIIPathsClassifyAsThemselves. The numstat half is untested and raised separately.

### Raised

- **BR-31** [Important] `doc-claim-contradicts-code` BR-26 changed the shared-surface grammar, but the declaration header and the atlas still describe the old two-form grammar
  This is the 6th finding in family doc-claim-contradicts-code. .sdlc/shared-surfaces:5-6 says every non-directory line is a path.Match glob, so pkg/vocab would mean exactly that path, and it omits the refusals; atlas/workflow/sdlc-binary.md:299 says the same. The code (surfaces.go:18-31,68-84) now treats a glob-free literal as path-or-subtree and refuses the double-star and leading-bang forms. The rule BR-27 stated (a fix that changes a fact sweeps restatements keyed on it in the same round) was violated by the same commit. Rule-level fix: give the grammar one owner, e.g. a flow.SurfaceForms string rendered into sdlc close --help via a token and pinned clause-by-clause against TestSurfaceForms (the churn.CodeFileRule / TestCodeFileRule pattern). Point the declaration header and atlas at it, and sweep in the same round with git grep for path.Match, trailing-slash and for-a-subtree phrasing.
- **BR-32** [Minor] `porcelain-output-parsed-as-data` The window numstat is still parsed as quoted output: its BR-30 change has no red test, and core.quotePath=false still quotes some paths
  This is the 2nd finding in family porcelain-output-parsed-as-data. Reverting the quotePath change in windowFileStats (closeflow.go:142) leaves the whole suite green; a probe with cmd/über.go at 150 lines then closes quick. The plan's "mutation-checked" claim covers only DiffNames. At HEAD, cmd/a"b.go at 150 lines closes quick, because quotePath=false still quotes a double-quote, a backslash and control characters, and the quoted row never joins the -z code-file list. Rule: every git listing a gate reads as data uses -z. Use diff --numstat -z (renames arrive as separate NUL fields) with a -z parse, and add a close test with a git-quoted code-file name over the line limit.
- **BR-33** [Minor] `accepted-input-never-effective` ParseSurfaces still accepts root and non-canonical patterns (/, ./, ., pkg/./vocab, pkg//vocab/) that match nothing
  This is the 3rd finding in family accepted-input-never-effective. Probed at head: each parses without error and matches no path. The forms are refused by hand-enumeration while the fuzz oracle still checks only no-panic, so the next form slips too. Rule: accept only canonical, non-empty patterns (path.Clean(p) == p and p is not "."). Make the oracle a property instead of a table: FuzzParseSurfaces asserts that an accepted glob-free pattern matches itself, and a directory entry matches pattern + "x". Also decide whether a glob naming a directory (pkg/v*) covers its subtree the way the literal form now does, or document the asymmetry.

## Round 9 — 2026-09-18T19:07:07-07:00 (claude) — passed

### Disposed

- BR-15 — addressed — change-code.md step 0, issue-lifecycle.md:54 and sdlc-binary.md:36 now say the record is written after the gates; the flowDrift refusal is documented under THE FLOW.
- BR-16 — addressed — Flow fields are unexported with accessors (flow.go:52-56) and Rule is opaque (flow.go:219-225), so the compiler enforces the single producer; the source-scan guard is gone.
- BR-17 — addressed — Reason codes plus the every-reason-has-a-row assertion (flow_test.go:205-233). Probed: disabling the seen[k] check turns TestFlowRecordCorpus red on both assertions.
- BR-27 — addressed — AGENTS.base.md:52 points at the catalog; help pages render the gate flags; the build-closure prose went with shared surfaces. Probed: hand-listing a gate flag in close.md turns TestGateFlagListsAreRenderedFromTheCatalog red.
- BR-31 — withdrawn — Overtaken by design: f0141e0 removed shared surfaces and their declaration, so no grammar is left to describe.
- BR-32 — not-addressed — -z is in place (closeflow.go:141), but TestCloseQuotedNameCodeFileLinesCount cannot fail. It uses a code file, and IsCodeFile's CodeProd default counts the quoted path as code either way; once f0141e0 dropped the -z name-list intersection, the test lost its teeth. Probed: restoring `-c core.quotePath=false diff --numstat` with ParseNumstat leaves it green. The discriminating fixture goes red: docs/a"b.md and tests/x"y_spec.lua at 150 lines beside a 5-line cmd/a.go, expecting quick. Plan:503's "the test reddens without -z" is false at HEAD.
- BR-33 — withdrawn — Overtaken by design: ParseSurfaces and the fuzz target were removed with shared surfaces (f0141e0).

### Raised

- **BR-34** [Minor] `doc-claim-contradicts-code` Two runtime messages and the publish gate's atlas page restate rules this window changed: FIX-THEN-SHIP's unconditional "do not re-run close", and the "#187 column set" header message
  This is the 7th finding in doc-claim-contradicts-code. Instances: close.go:1871-1873 says "Do NOT re-run sdlc close — this verdict already sanctions shipping after the fixes", yet publishgate.go:176 now refuses a quick issue past MaxAddedLinesAfterReview and sends it back to close; atlas/workflow/pre-merge-checks.md:24-46 (the publish gate's page) lists its refusals and that protocol without the quick re-measure; close.go:981 prints "header upgraded to the #187 column set", and the only upgrade left to fire is #187 to #231. Enumeration: 7 surfaces describe what the publish gate refuses or what the verdict sanctions (close.md:222, merge.md:35, push.md:10, sdlc-binary.md:303 and publishgate.go:6 derive; close.go:1871 and pre-merge-checks.md do not). 4 surfaces name the ledger's column set, and close.go:981 is wrong. Prevalence is 3 of 11. Why earlier sweeps missed these: they keyed on the new mechanism's name, while these sentences state the old conclusion. Rule: a runtime string that states a policy renders it from the policy's owner, never a literal. formatFixThenShipProtocol takes the flow and renders flow.AfterReviewSummary() on quick; the header message states no version. When a change adds a policy function, the sweep greps for the old policy's conclusion ("do not re-run", "landed after", "#187"), not the new gate's name (ARCH-PURPOSE shadow sweep).
- **BR-35** [Minor] `gate-key-undeclared` The publish check's new quick-growth refusal matches no GateCatalog RefusalPat, so the friction instrument never counts it
  This is the 2nd finding in gate-key-undeclared. The merge/push no-judge RefusalPat is `publish gate: \d+ commit\(s\) landed after` (gatesig.go:150,156). "publish gate: #N is on the quick flow, and its diff grew…" (publishgate.go:180) is unattributed by classifyOutputLine, although --no-judge waives it. Enumeration of refusals added in this window that a catalogued flag waives: DoneWhenFresh (catalogued) and the publish quick-growth refusal (not catalogued), so 1 of 2. Rule: every refusal a catalogued flag waives must be attributable by the catalog. The tests that produce a refusal assert that their (command, flag) RefusalPat matches it, the inverse of assertNoGatesigCollision. Fix: widen the two patterns to `publish gate: (\d+ commit\(s\) landed after|#\d+ is on the quick flow)` and add that assertion to TestRunPublishGate_QuickGrewPastReview.

## Open findings

- **BR-32** [Minor] `porcelain-output-parsed-as-data` The window numstat is still parsed as quoted output: its BR-30 change has no red test, and core.quotePath=false still quotes some paths
- **BR-34** [Minor] `doc-claim-contradicts-code` Two runtime messages and the publish gate's atlas page restate rules this window changed: FIX-THEN-SHIP's unconditional "do not re-run close", and the "#187 column set" header message
- **BR-35** [Minor] `gate-key-undeclared` The publish check's new quick-growth refusal matches no GateCatalog RefusalPat, so the friction instrument never counts it
