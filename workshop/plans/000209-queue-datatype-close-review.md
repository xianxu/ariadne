# Boundary Review — ariadne#209 (whole-issue close)

| field | value |
|-------|-------|
| issue | 209 — queue datatype for advisory work ordering |
| repo | ariadne |
| issue file | workshop/issues/000209-queue-datatype.md |
| boundary | whole-issue close |
| milestone | — |
| window | 00d7c75c5371bbb3e2f01f5e28780565c274f3a9..9c547fdc0ffd2771bf0cae0df276b5b8e6b165e2 |
| command | sdlc close --issue 209 |
| reviewer | claude |
| timestamp | 2026-09-07T18:52:12-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

M2 lands the design it promised: `Line`/`Doc`/`Intent` are genuinely pure (no git, no IO, fuzz-guarded round-trip with a checked-in corpus entry), `Intent.Apply` is handed to `TrunkFile.Update` as the transform so the intent-not-content property is structural rather than asserted, the verb is a thin shell, the seed really landed on `origin/main` through the verb (four entries, verified), and the atlas gained two substantial, non-restating sections. `go test ./cmd/sdlc/...` is green except the pre-existing ariadne#210 failure. What blocks SHIP is a cluster of cheap, empirically-reproduced gaps at the verb's input boundary: `--tag` is the one user-supplied field that lands in the line format with no validator, and a newline in it publishes a forged three-entry file to the trunk with exit 0 (reproduced); validation runs *inside* the transform, so the Spec's and plan's "rejected before any git call" is false and, offline, the validation error is replaced by "origin unreachable" (reproduced); the converge path silently flips a project line to an issue line and drops its tag while reporting only that the why-now changed (reproduced) — the second instance of the rule BR-22 already stated; the line format is documented in no user-facing artifact despite being an explicit Done-when clause; a new trunk-*writing* verb shipped into the base-layer binary with no `guardSpineRepo` decision recorded either way; and Task 13 Step 3 is ticked but no note exists on either #207 file. Of the six open prior findings, three are addressed (two mutation-verified), three remain.

## 1. Strengths

- **`trunkfile.go:70-98`** — the three-state rule is stated once at the top of the file and every query obeys it. `gitExitCode` returning `-1` for a non-exit failure is the right shape: a failed type assertion leaving a zero value is exactly how "could not tell" becomes "absent".
- **`line.go:110-123`** — the parse self-check (`if l.String() != s { return unparsed, false }`) makes the round-trip invariant hold *by construction* rather than by five separate arguments about five separate trims. This is the best decision in the diff, and the comment explains why the obvious fix was rejected.
- **`trunkfile_test.go:722-752` / `:756-771`** — both BR-22 and BR-21 are pinned by tests that go red without the fix. I verified by mutation: hardcoding `mode := "100644"` fails `TestTrunkFile_PreservesFileMode`; deleting the `refPresent` block from `readFrom` fails both `ReadFromRefusesUndeterminableRef` and `FreshRepoReadsEmptyNotError`.
- **`queue_e2e_test.go:22`** — the assembled path (verb → pure core → real git → bare origin) is exercised end to end, including that `workshop/queue.md` never appears in the working tree. That test did not exist before this round and it is the one that would catch a wiring regression.
- **`atlas/workflow/sdlc-binary.md`** — routes rather than restates (noun → `construct/datatype/queue.md`, verb contract → `--help`), and the `TrunkFile` section explains *why* `push <commit>:main` is stronger than a cleanliness check instead of just listing the plumbing.

## 2. Critical findings

None.

## 3. Important findings

**I-1 — `--tag` is the one format-bearing field with no validator; a newline in it publishes a forged file (`intent.go:83-101`, `queue.go:85`).**
Reproduced against the real code path: `Intent{Op: OpAdd, Ref: "b#2", WhyNow: "why", Tag: "x\n- forged#9 — injected [y"}` returns `nil` and publishes

```
- a#1 — first
- b#2 — why [x
- forged#9 — injected [y]
```

Three entries where one was intended, the real entry's why-now mangled, a forged ref, exit 0. `ValidateWhyNow` exists precisely to stop this (`line_test.go:96` demonstrates the identical forge), and `ValidateRef` covers the third field — `Tag` was simply not in the enumeration. Fix: add `ValidateTag` (same newline/control/`sep`/bracket rules, plus non-blank-after-trim since a whitespace-only tag is silently absorbed into why-now) and call it from `applyAdd` beside the other two.

**I-2 — validation runs inside the transform, so "rejected before any git call" is false and the offline error masks it (`queue.go:161-172`, `intent.go:84-89`).**
Path: `RunE` → `openTrunkStore()` → `gitx.RepoTopLevel()` (git) → `Update` → `signs()` (git config) → `fetch()` (network) → *then* `transform` → `ValidateRef`. Reproduced with an unreachable origin: `sdlc queue add "bad ref" x` prints `origin unreachable (offline?): exit status 128` and `the queue on the trunk right now (0 entries):` — the operator never learns the ref was malformed. The Spec's Done-when ("rejected before any git call"), the Spec's "Input validation, before any git call" paragraph, and the plan's Task 9 table all promise otherwise. The test that claims to pin this (`queue_test.go:168`, `…ValidationRefusesBeforeTouchingTheTrunk`) only asserts the fake's content is unchanged — it cannot observe a git call, and passes today. Fix: an `Intent.Validate()` called in `runQueueEdit` before `openTrunkStore`/`Update`; keep the in-`Apply` calls as the belt, and give the test a call-counting seam so it fails if any git runs first.

**I-3 — converge replaces the whole line, silently flipping kind and dropping the tag, while the note claims only the why-now moved (`intent.go:90-98`). This is the 2nd finding in family `unspecified-fields-not-preserved`.**
Do not fix just this site — state the rule. **RULE: when replacing one field of an existing record, carry every field you were not asked to change; and a note may only claim the change it actually made.** BR-22 stated this for git tree modes; the enumeration over the *record* here is `{Ref, WhyNow, Tag, Kind, cr}`. Reproduced: `add sdlc-fleet "sharper reason"` onto `- project:sdlc-fleet — the whole area is next [sdlc]` yields `- sdlc-fleet — sharper reason` with the note `kept the newer why-now (was …)` — the `project:` marker and `[sdlc]` are gone and unmentioned. `cr` is the second cell: converging onto a CRLF line yields `"- a#1 — one prime\n- b#2 — two\r\n"`, mixing line endings in the published file. This is the interleaving table's designed-for concurrent-`add` cell, so a peer's `--project`/`--tag` vanishes on the path built to make peers survive. `TestIntent_Apply_InterleavingTable`'s converge case asserts only refs plus a note substring, so no test can catch it.

**I-4 — the line format is documented in no user-facing artifact, and a near-miss hand-edit vanishes from the listing without a word (`construct/datatype/queue.md`, `cmd/sdlc/helptext/queue.md`, `queue.go:145-153`).**
The Done-when requires the datatype prose to state "the line format"; it does not, and routes the operational contract to `sdlc queue --help`, which does not carry it either. The grammar — `- <ref> — <why> [tag]`, the em-dash separator, the `project:` prefix that is the *only* way a project line is marked, the why-now restrictions — exists only in Go source comments. That matters more than usual here because the design explicitly invites hand-editing (`line.go:45-48`, the fuzz target, "an unparsed line is one `sdlc queue remove` cannot find"). Compounding it: `runQueueList` prints `Entries()` only, so a hand-edited `- a#1 - why` (ASCII hyphen) is silently absent from the listing while sitting in the file. Two fixes, both cheap: add a "Line format" section to the datatype prose (and the `--help`), and have `runQueueList` report `len(lines) - len(entries)` unparsed lines on stderr.

**I-5 — a new trunk-*writing* verb shipped into the base-layer binary with no `guardSpineRepo` decision recorded (`queue.go:36-53`, `repoguard.go:58-61`).**
`repoguard.go:5-14` says the guard is wired into the lifecycle verbs and that "reads … stay unguarded by construction". `queue add/remove/move` are writes that push to `origin/main`, and they are unguarded. In a brain repo that means a CAS push at a `gcrypt::ssh` remote; in a repo with no `workshop/issues/` it creates `workshop/queue.md` on main. This issue's own Spec calls the brain-queue question "an open disagreement, not a settled no" — shipping the unguarded path quietly settles it. `migrate.go:460` shows the established shape for the other answer: an explicit comment saying why it is deliberately not guarded. Either wire `guardSpineRepo` into the three mutating subcommands or record the exemption the way `migrate` does.

**I-6 — Task 13 Step 3 is ticked but not delivered. This is the 3rd finding in family `plan-traceability`.**
Do not fix just this site — state the rule. **RULE: a `- [x]` is a claim about the tree, not about intent; before a close, sweep every checked row in the closing chunk against the artifact it names.** Enumeration over Chunk 2, 4 rows fail it: (a) Task 13 Step 3 — `grep -i 'trunkfile\|#209' workshop/issues/000207-*.md` returns nothing in either #207 file (note the tracker carries two, `000207-publish-issue-files-without-a-main-worktree.md` and `000207-sync-without-worktree.md`; say which one gets the note), so `ariadne#207` still has no record that its primitive exists nor of the retry-semantics defect the Spec promised to flag there; (b) Task 11 Step 5's commit does not exist — folded into `c93be7c`; (c) Task 13 Step 4's commit does not exist — the seed landed as `queue: add …` commits on `origin/main`; (d) `04804ff` ("end-to-end over real git") is real, valuable work with no plan row at all. M1's Revisions entry recorded exactly this kind of divergence; M2's chunk has no such entry.

## 4. Minor findings

- **M-1 — 4th finding in family `duplicated-helper`.** Do not merge just these — state the rule. **RULE (already stated at `trunkfile.go:230-235`): two functions differing only in a parameter's value or in which field of one result they keep are one function.** Enumeration over `gitx`, 2 remaining sites: `pathPresent:219` and `modeOf:426` are the same `ls-tree` on the same `(ref, path)` differing only by `--name-only` — `Update` runs both per attempt, and BR-22's own fix sketch said to drop `--name-only` and take the mode from the call already on the wire; `refPresent:178` and `resolve:356` are the same `rev-parse --verify` differing only by `--quiet` and return type — `Update` runs both on `trackingRef` per attempt. Collapse each pair to one call returning the richer result.
- **M-2 — 2nd finding in family `vacuous-test-guard`.** **RULE: a test may not be named for a property its double cannot produce.** `fakeTrunk.Update` (`queue_test.go:37-51`) calls `transform` exactly once and fires `peer` *before* it, so `TestQueueEdit_PeerEditSurvivesTheReplay` ("the property the whole design exists for") passes against an `Update` with no retry at all; likewise `…ValidationRefusesBeforeTouchingTheTrunk` (I-2). Give `fakeTrunk` a rejection mode that re-invokes the transform, or name the tests for what they check.
- **M-3** — plan Core-concepts table names `queueCmd`; the symbol is `NewQueueCmd`. `gitx.FirstLine`, newly exported in this window, is on no table row.
- **M-4** — `queueRefusal` (`queue.go:212`) discards `ReadDegraded`'s warning, and issues a second full fetch on the error path.

## 5. Test coverage notes

- Pure core: genuinely IO-free, and the fuzz target with a checked-in corpus entry is the right guard for a hand-editable file. `TestIntent_Apply_UnknownOpFails` and the input-non-mutation assertion inside the table loop are both real.
- The uncovered composition is the verb's own retry: no test drives `Intent.Apply` through a *real* non-fast-forward rejection. `gitx`'s two-publisher test uses a raw append transform; the e2e test has no peer; the fake cannot reject. Adding a peer push to `queue_e2e_test.go` would close it and would also be the first test of the Done-when clause "a concurrent `add` replays and **both** lines land" as the *queue* rather than as `TrunkFile`.
- Every finding above except I-4/I-5/I-6 is a bug no current test can see. I-1 and I-3 each need one assertion; I-2 needs a call-counting seam.

## 6. Architectural notes

- **ARCH-DRY — flag** (M-1): two `ls-tree` pairs and two `rev-parse` pairs; BR-22's fix added a *second* `ls-tree` rather than widening the one already running.
- **ARCH-PURE — flag** (I-2): the pure validators exist and are well-tested, but the only path that reaches them is inside an IO transform, so pure logic is gated behind a network call and its error is masked by IO failure. `firstLine` → `gitx.FirstLine` is a good consolidation.
- **ARCH-PURPOSE — flag** (I-4, I-6): shadow-sweep of the format's consumers — code (authoritative), datatype prose (silent), `--help` (silent), atlas (routes to the two silent ones). The single source is Go comments a queue reader will not open. I-3 is the same axis on findings: BR-22 stated the carry-your-fields rule for one record; the sibling record in the new code repeats it.
- **ARCH-MOCK — pass.** Real bare origin throughout `gitx`, real git in the e2e, and `fakeTrunk` is stateful rather than a call mock. The gap is capability, not shape (M-2).
- **ARCH-CONSTRAINTS — pass.** Bound of 3 enforced and tested; `signs()` hoisted out of the loop; `TestTrunkFile_OneFetchPerAttempt` pins the fetch budget. No context/timeout, named as residue in the plan's audit table.
- **ARCH-SECURE — mostly pass, one flag** (I-1). Trunk content is treated as untrusted and fuzz-guarded; the temp-index directory has no unclaimed-name window; no credentials touched. `--tag` is the one boundary input not parsed into a validated value.
- **ARCH-ORDER — flag** (M-2, and I-3's `Kind`/`cr` cells). The interleaving table is enumerated and each cell tested at the pure layer, and the real rejection is reproducible in `gitx`. The verb layer can observe only one interleaving.

## 7. Plan revision recommendations

Add a `## Revisions` entry — "2026-09-07 — M2 as built" — recording:
1. Commit granularity diverged again: Tasks 7+8 landed as `51a59ea`; Task 11's commit folded into `c93be7c`; Task 13's seed landed as `queue: add …` commits on `origin/main` rather than a branch commit; `04804ff` (the real-git e2e) was unplanned and should be added as a task row.
2. Task 13 Step 3 unticked until the note actually lands on #207 (naming which of the two #207 files).
3. Core-concepts table: `queueCmd` → `NewQueueCmd`; add a row for `gitx.FirstLine`.
4. Task 12's Done-when clause "the line format" recorded as unmet, with where it will be stated.
5. The `project:` prefix as the project-line marker — a design decision made during implementation that appears in no plan or Spec text.

```findings
dispose:
  - id: BR-11
    disposition: not-addressed
    note: |
      Unchanged at trunkfile.go:192 and :158, and the enumeration is now larger — M2 added two sites. (4) queue.go:212 prints "the queue on the trunk right now (N entries)" from a ReadDegraded whose warning it discards; reproduced offline, it printed "(0 entries)" for a trunk that has four, i.e. a fabricated state presented as observed. (5) Update:317 calls offlineError unconditionally, so a repo with NO origin gets "origin unreachable (offline?)" although ReadDegraded has a dedicated branch for exactly that case. (6) intent.go:96 reports "kept the newer why-now" for an edit that also dropped the tag and flipped the kind (see the new I-3). Same rule: a message may only name state the code observed.
  - id: BR-12
    disposition: not-addressed
    note: |
      Unchanged. trunkfile.go:383-385 still explains what --path buys with no sentence noting the attributes are resolved from the working tree/index, not from the trunk being written.
  - id: BR-15
    disposition: addressed
    note: |
      Verified by grep: `grep -in "combined" trunkfile.go` now returns only :39 ("returned SEPARATELY, not combined"), and commitAndPush:365 reads "Returns git's STDERR".
  - id: BR-16
    disposition: addressed
    note: |
      Verified by grep over the plan: no CombinedOutput at lines 86/136/207 or anywhere outside the Revisions entry describing the history. Residual plan-vs-code staleness (line 7's "sibling runEnv" vs the delivered runGitIn; line 171's t.Skip signing test vs the delivered -S interception; the Revisions entry's "cat-file -e" vs the delivered ls-tree) is carried forward under the new plan-traceability finding rather than re-raised here.
  - id: BR-21
    disposition: not-addressed
    note: |
      The main defect IS fixed and properly pinned — readLocal/readRef collapsed into readFrom carrying the guard; verified by mutation (deleting the refPresent block turns both ReadFromRefusesUndeterminableRef and FreshRepoReadsEmptyNotError red). The only residue is the half the finding asked for and the diff did not do: trunkfile.go:169 still passes the UNQUALIFIED t.branch. Reproduced against the real type — a repo with a tag named "main" and no tracking ref returns the TAG's bytes while warning "used local main". One-line fix: pass "refs/heads/" + t.branch.
  - id: BR-22
    disposition: addressed
    note: |
      modeOf + TestTrunkFile_PreservesFileMode. Verified by mutation, not by the diff: replacing the modeOf call with a hardcoded "100644" fails the test with "mode = [100644 ... run.sh], want 100755 preserved". The new-path 100644 case is covered in the same test. (Note M-1: the fix added a SECOND ls-tree rather than widening the one pathPresent already runs, which is what the finding sketched.)
findings:
  - id: new
    severity: Important
    family: unvalidated-format-field
    title: |
      --tag is the one user-supplied field that lands in the line format with no validator; a newline in it publishes a forged file with exit 0
    detail: |
      Reproduced against the real code path. Intent{Op: OpAdd, Ref: "b#2", WhyNow: "why", Tag: "x\n- forged#9 — injected [y"} returns nil and publishes three entries where one was intended, mangling the real entry to "- b#2 — why [x" and forging "- forged#9 — injected [y]". ValidateWhyNow exists to stop exactly this forge (line_test.go:96 demonstrates it) and ValidateRef covers the third field; Tag was simply not in the enumeration. RULE every user-supplied field that lands in the line format is validated by the same rule. Add ValidateTag (newline/control/sep/bracket, plus non-blank-after-trim, since a whitespace-only tag is silently absorbed into why-now) and call it from applyAdd beside the other two.
  - id: new
    severity: Important
    family: validation-behind-io
    title: |
      Validation runs inside the transform, so "rejected before any git call" is false and the offline error masks it
    detail: |
      Path is RunE -> openTrunkStore -> gitx.RepoTopLevel (git) -> Update -> signs (git config) -> fetch (network) -> transform -> ValidateRef. Reproduced with an unreachable origin: `queue add "bad ref" x` prints "origin unreachable (offline?)" and a fabricated "(0 entries)" listing; the operator never learns the ref was malformed. The Spec's Done-when, the Spec's "Input validation, before any git call" paragraph and the plan's Task 9 table all promise the opposite. queue_test.go:168 claims to pin this but only asserts the fake's content is unchanged, so it cannot observe a git call and passes today. Fix: an Intent.Validate() called in runQueueEdit before openTrunkStore, keeping the in-Apply calls as the belt, plus a call-counting seam in the test.
  - id: new
    severity: Important
    family: unspecified-fields-not-preserved
    title: |
      Converge replaces the whole line, silently flipping kind and dropping the tag, while the note claims only the why-now moved
    detail: |
      This is the 2nd finding in family unspecified-fields-not-preserved. Do NOT fix only this site — state the rule. RULE when replacing one field of an existing record, carry every field you were not asked to change, and a note may only claim the change it actually made. BR-22 stated it for git tree modes; the record here is {Ref, WhyNow, Tag, Kind, cr} and intent.go:90-98 rebuilds it from the Intent alone. Reproduced: `add sdlc-fleet "sharper reason"` onto "- project:sdlc-fleet — the whole area is next [sdlc]" yields "- sdlc-fleet — sharper reason" with the note "kept the newer why-now (was ...)" — project marker and tag gone, unmentioned. Second cell: converging onto a CRLF line yields "- a#1 — one prime\n- b#2 — two\r\n", mixing line endings. This is the interleaving table's concurrent-add cell, so a peer's --project/--tag vanishes on the path built to make peers survive; the table test asserts only refs plus a note substring and cannot catch it.
  - id: new
    severity: Important
    family: format-undocumented
    title: |
      The line format is stated in no user-facing artifact, and a near-miss hand-edit disappears from the listing without a word
    detail: |
      The Done-when requires the datatype prose to state "the line format". construct/datatype/queue.md does not, and routes the operational contract to `sdlc queue --help`, which does not carry it either. The grammar — "- <ref> — <why> [tag]", the em-dash separator, the `project:` prefix that is the ONLY marker of a project line, the why-now restrictions — lives only in Go comments. That matters because the design explicitly invites hand-editing (line.go:45-48, the fuzz target, and the comment that an unparsed line is one `queue remove` cannot find). Compounding: runQueueList prints Entries() only, so a hand-edited "- a#1 - why" (ASCII hyphen) is silently absent from the listing while sitting in the file. Two cheap fixes: a "Line format" section in the datatype prose and the helptext, and an stderr line in runQueueList reporting len(lines)-len(entries) unparsed lines.
  - id: new
    severity: Important
    family: unguarded-write-verb
    title: |
      A new trunk-writing verb shipped into the base-layer binary with no guardSpineRepo decision recorded either way
    detail: |
      repoguard.go:5-14 states the guard is wired into the lifecycle verbs and that reads stay unguarded by construction. `queue add/remove/move` are writes that push to origin/main, and they are unguarded — in a brain repo that is a CAS push at a gcrypt::ssh remote, and in a repo without workshop/issues/ it creates workshop/queue.md on main. This issue's own Spec calls the brain-queue question "an open disagreement, not a settled no", so shipping the unguarded path quietly settles it; ariadne is the base layer, so the verb reaches every downstream repo. migrate.go:460 is the established shape for the other answer — an explicit comment saying why the exemption holds. Either wire guardSpineRepo into the three mutating subcommands or record the exemption the way migrate does.
  - id: new
    severity: Important
    family: plan-traceability
    title: |
      Task 13 Step 3 is ticked but no note exists on either #207 file, and three more Chunk-2 rows claim artifacts the tree does not have
    detail: |
      This is the 3rd finding in family plan-traceability. Do NOT fix only this site — state the rule. RULE a "- [x]" is a claim about the tree, not about intent; before a close, sweep every checked row in the closing chunk against the artifact it names. Enumeration over Chunk 2, 4 rows fail: (a) Task 13 Step 3 — grep -i "trunkfile|#209" over both workshop/issues/000207-*.md returns nothing, so #207 still has no record that gitx.TrunkFile exists nor of the retry-semantics defect the Spec promised to flag there (say which of the two #207 files gets it); (b) Task 11 Step 5's commit does not exist, folded into c93be7c; (c) Task 13 Step 4's commit does not exist — the seed landed as `queue: add ...` commits on origin/main; (d) 04804ff (the real-git e2e) is real work with no plan row. M1 recorded exactly this divergence in Revisions; M2's chunk has no such entry.
  - id: new
    severity: Minor
    family: duplicated-helper
    title: |
      pathPresent/modeOf and refPresent/resolve are each one function split in two, and Update runs both halves of both pairs per attempt
    detail: |
      This is the 4th finding in family duplicated-helper. Do NOT merge only these — state the rule, which the code already states at trunkfile.go:230-235: two functions differing only in a parameter's value, or in which field of one result they keep, are one function. Enumeration over cmd/sdlc/internal/gitx/, 2 remaining sites. trunkfile.go:219 (pathPresent) and :426 (modeOf) run the same ls-tree on the same (ref, path), differing only by --name-only — and BR-22's own fix sketch said to drop --name-only and take the mode from the call already on the wire, so the fix for BR-22 created this instance. trunkfile.go:178 (refPresent) and :356 (resolve) run the same rev-parse --verify, differing only by --quiet and return type; Update calls both on trackingRef each attempt. Collapse each pair to one call returning the richer result.
  - id: new
    severity: Minor
    family: vacuous-test-guard
    title: |
      fakeTrunk calls the transform exactly once and fires its peer before it, so two tests are named for properties the double cannot produce
    detail: |
      This is the 2nd finding in family vacuous-test-guard. State the rule rather than patching a site: a test may not be named for a property its double cannot produce. queue_test.go:37-51 — Update calls transform once and runs `peer` BEFORE it, so TestQueueEdit_PeerEditSurvivesTheReplay ("the property the whole design exists for") would pass against an Update with no retry at all, and TestQueueEdit_ValidationRefusesBeforeTouchingTheTrunk cannot see a git call. Consequence: no test anywhere drives Intent.Apply through a real non-fast-forward rejection — gitx's two-publisher test uses a raw append transform and the e2e has no peer. Give fakeTrunk a rejection mode that re-invokes the transform, and add a peer push to queue_e2e_test.go.
  - id: new
    severity: Minor
    family: plan-traceability
    title: |
      Core-concepts table names queueCmd where the symbol is NewQueueCmd, and gitx.FirstLine has no row
    detail: |
      Same family as above; listed separately only because it is a table edit rather than a checkbox. FirstLine was newly exported in this window (moved out of issueids.go) and is part of the package's surface.
  - id: new
    severity: Minor
    family: error-misattribution
    title: |
      queueRefusal discards ReadDegraded's staleness warning and issues a second full fetch on the error path
    detail: |
      queue.go:212 binds the warning to _ , so the handoff prints "the queue on the trunk right now" with no indication the read was degraded. Folded into the BR-11 enumeration above; noted here for the file:line.
```
