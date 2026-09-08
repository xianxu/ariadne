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

---

## Re-review — 2026-09-07T19:16:10-07:00 (FIX-THEN-SHIP)

| field | value |
|-------|-------|
| issue | 209 — queue datatype for advisory work ordering |
| repo | ariadne |
| issue file | workshop/issues/000209-queue-datatype.md |
| boundary | whole-issue close |
| milestone | — |
| window | 00d7c75c5371bbb3e2f01f5e28780565c274f3a9..66065f6a7cd58a30f943011eca52c31bc2e6451d |
| command | sdlc close --issue 209 |
| reviewer | claude |
| timestamp | 2026-09-07T19:16:10-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

The fix round did real work and I verified it by mutation rather than by reading the commit message: `Intent.Validate` genuinely runs in the verb before `Update` (deleting it turns all six `…ValidationRefusesBeforeTouchingTheTrunk` sub-cases red), `ValidateTag` is pinned at two independent sites, converge now merges instead of replacing (a wholesale-replace mutation turns two tests red), the unparsed-line report exists, the `#207` note landed and is substantive, and the guard decision is recorded. Full suite is green except the pre-existing ariadne#210 failure, and the seed really is on `origin/main` — four entries, six `queue: …` commits, LF, em-dash, tags. What holds SHIP back is that two prior Importants were closed on a claim rather than a check — **BR-21 was dismissed as "stale" in the commit message and is live** (I reproduced tag-shadowing: with a tag named `main`, `rev-parse --verify main` and `cat-file blob main:q.md` return the *tag*, so `ReadDegraded`'s fallback at `trunkfile.go:169` serves the tag's bytes while warning "used local main"), and BR-25's `Kind` cell still flips a peer's project line to an issue line on a re-add that never mentioned `--project`. And the BR-23 enumeration was of *fields*, not of what makes a field safe: `sdlc queue add project:foo "why"` and `sdlc queue add a#1 "  padded  "` both pass `Validate` and publish lines that re-parse to a different record — the first duplicates on re-add and cannot be removed by the ref that was typed, the second is invisible to the listing and unremovable, both exit 0. All of it is cheap; none of it corrupts anything outside `workshop/queue.md`.

## 1. Strengths

- **`line.go:110-123`** — the parse self-check (`if l.String() != s { return unparsed, false }`) makes the round-trip invariant hold by construction. It is the best idea in the diff, and finding #1 below is simply "you built this at the read boundary and not at the write boundary."
- **`queue.go:188-196` + `queue_test.go:186-215`** — BR-24's fix is real and the test is not vacuous. I removed the `Validate` call in a scratch copy and every sub-case failed with `Update ran 1 times`. The fake's `calls` counter is the seam that makes the claim observable.
- **`intent.go:126-152`** — converge-by-merge with a note that enumerates what actually changed; a wholesale-replace mutation reddens both `TestQueueEdit_ConvergePreservesUnmentionedFields` and `…ConvergeMergeSemantics/omitting_the_tag_keeps_it`. The CRLF cell BR-25 named is genuinely fixed (verified: converging onto `\r\n` keeps `\r\n`).
- **`cmd/sdlc/helptext/queue.md:17-30`** — the LINE FORMAT section is exactly what BR-26 asked for, including the em-dash-not-hyphen sentence and the "preserved but not listed" contract, and `runQueueList` now backs it with a real count.
- **`workshop/issues/000207-sync-without-worktree.md`** (tail) — not a pointer but an argument: it names the seam, and it says plainly that #207's "both issue files land" Done-when is itself the bug for a colliding id. That is the note BR-28 wanted.
- **`atlas/workflow/sdlc-binary.md:647-727`** — routes rather than restates, and explains *why* `push <commit>:main` beats a cleanliness check.

## 2. Critical findings

None.

## 3. Important findings

**A — a validated field can still render a line that re-parses to a different record (`line.go:197-208`, `intent.go:73-89`). 2nd in family `unvalidated-format-field`.**
Do not patch these two inputs — state the rule. **RULE: a field validator that enumerates forbidden characters does not make the record round-trip; the write path must assert what the read path already asserts.** Two reproduced instances, both exit 0:
- `Intent{Op: OpAdd, Ref: "project:foo", WhyNow: "why"}` → `Validate` returns nil → publishes `- project:foo — why` → re-reads as `{Ref: "foo", Kind: KindProject}`. `queue remove project:foo` then reports *"project:foo was not in the queue; nothing to remove"* and leaves it; a second `add project:foo` appends a duplicate instead of converging. `ValidateRef` rejects `[`, `]`, whitespace and ` — ` but not the `project:` prefix, which is structure in that position.
- `Intent{Op: OpAdd, Ref: "a#1", WhyNow: "  padded  "}` → `Validate` returns nil → publishes `- a#1 —   padded  `, which `ParseLine`'s self-check rejects: `Entries()` is empty, `UnrecognizedItems()` is 1. The verb creates the very state BR-26's new warning attributes to *"a hand-edit that used a hyphen"*.

Fix that covers both and any future field: in `applyAdd`, after building `line`, refuse unless `ParseLine(line.String())` returns an equal record — the same mechanical guarantee `line.go:110-123` already gives the read side (`ARCH-DRY`, `ARCH-SECURE`).

**B — the spine-guard decision is recorded in prose but pinned by nothing, and the enumeration that pins every other guard now excludes it (`queue.go:85,112,134`, `repoguard.go:8-14`, `repoguard_test.go:17-25`). 2nd in family `unguarded-write-verb`.**
Do not just add three assertions — state the rule. **RULE: a guard is only "wired" when it is in the enumeration the drift test walks; a guard reachable only by remembering to call it is a comment.** `repoguard.go:8-11` says the guard is wired into "exactly the lifecycle verbs … `processmanual.WorkflowVerbs`; the drift test enumerates it" — that sentence is now false, since `queue add/remove/move` call `guardSpineRepo` and are not in `WorkflowVerbs` (correctly so: adding them would distort the #172 friction instrument). Consequence: no test anywhere enters `NewQueueCmd`'s command tree at all — not the guard, not `--project`→`Kind`, not `--tag`, not the `--before`/`--after` mutual exclusion — so a refactor that drops the guard call ships silently, in the base-layer binary, to every downstream repo. Also `newQueueMoveCmd` runs its flag validation *before* the guard, contrary to the guard-first placement `repoguard.go:58-60` states. Fix: correct the "exactly" sentence to name the two sets, and add a brain-repo table test over the three write subcommands (plus one asserting the bare list is *not* guarded, which is the decision's other half).

## 4. Minor findings

- **5th in family `duplicated-helper`.** The item-marker predicate is written twice and the two disagree: `line.go:73` tests `CutPrefix(trimmed, "- ")` on the raw line, `doc.go:75` tests `HasPrefix(TrimSpace(l.raw), "- ")`. An ordinary indented prose list (`  - the top line is next`) is reported as 2 lines that "look like entries but do not parse" — reproduced — which is precisely the reader-training failure `doc.go:68-71` says it avoids. **RULE (already stated at `trunkfile.go:230-235`): two functions differing only in a parameter's value, or in which part of one answer they keep, are one function.** Extract `looksLikeItem(s string) bool` and have `ParseLine` and `UnrecognizedItems` both call it.
- **New family `record-region-boundary`.** `intent.go:153` appends at the end of the *file*, not the end of the *record region*: adding to `"# Queue\n\n- a#1 — first\n\n<!-- keep this at the bottom -->\n"` lands the new entry below the trailing comment, while `move --after a#1` would land it above. Reproduced.
- Appending to a CRLF document produces mixed endings (`- a#1 — one\r\n- b#2 — two\r\n- c#3 — three\n`) — the converge path carries `cr`, the append path does not. Folded into BR-25's note.
- `queue.go:194`'s comment says "BEFORE any git call", but `guardSpineRepo` and `openTrunkStore` each run `git rev-parse --show-toplevel` first (two identical invocations per write). The property that matters — no fetch, no masked cause — does hold; the sentence overstates it.

## 5. Test coverage notes

- The pure core is genuinely IO-free and the fuzz target with a checked-in corpus entry is the right guard for a hand-editable file.
- Three claimed fixes are mutation-verified above. One is not verifiable because it has no test at all: the spine guard (finding B).
- The converge table (`intent_test.go:214-242`) covers `Tag` and "unchanged" but has no `Kind` row and no `cr` row — the two cells BR-25 explicitly enumerated. A `Kind` row is what would have caught the remaining half.
- Still uncovered, unchanged from last round: no test drives `Intent.Apply` through a *real* non-fast-forward rejection. `fakeTrunk.Update` (`queue_test.go:53-69`) invokes `transform` exactly once and fires `peer` before it, so `TestQueueEdit_PeerEditSurvivesTheReplay` passes against an `Update` with no retry. The new `reached` field is incremented at two sites and read at none — the doc comment at `queue_test.go:36-38` credits it with making the claim testable, but the assertion is on `calls`.

## 6. Architectural notes

- **ARCH-DRY — flag.** BR-29's two split pairs are untouched (`pathPresent:219`/`modeOf:426`, `refPresent:178`/`resolve:356`, both halves run per `Update` attempt), and the diff added a third instance in the item predicate.
- **ARCH-PURE — pass.** Validation moved out of the transform into the shell; the merge policy is pure, the CAS retry is at the seam, neither knows the other's rules.
- **ARCH-PURPOSE — flag.** Shadow-sweep of "the line format": code (authoritative) ✓, `--help` (now derives) ✓, atlas (routes) ✓, `construct/datatype/queue.md` (still silent, and it is the artifact the Done-when names) ✗. And finding A is the same axis on findings: BR-23 asked for an enumeration and got one — of fields, not of what makes a field safe.
- **ARCH-MOCK — flag.** `gitx` is exemplary (real bare origin throughout). The verb layer's double still cannot produce the rejection the design exists for.
- **ARCH-CONSTRAINTS — pass.** Bound of 3, one fetch per attempt, `signs()` hoisted. Residue: `queueRefusal`'s extra full fetch, and no timeout (declared in the plan's audit table).
- **ARCH-SECURE — flag** (finding A). Trunk content is treated as untrusted and fuzz-guarded on the read side; the write side has no symmetric parse-back check, so invalid state stays representable.
- **ARCH-ORDER — flag.** The interleaving table is enumerated and each cell tested at the pure layer, but the converge cell still loses `Kind`, and `Kind`'s zero value is `KindIssue`, so "not mentioned" and "explicitly an issue line" are one state. Collapsing that (a `KindUnspecified` zero, or `c.Flags().Changed("project")`) is the fix for BR-25's residue.

## 7. Plan revision recommendations

Extend the "2026-09-07 — M2 as built" entry (it is honest about the red-green ticks; complete the sweep it began):

1. Add a task row for `04804ff` (the real-git e2e) — real work, no plan row, called out last round and still absent.
2. Core-concepts table: `queueCmd` → `NewQueueCmd` (lines 81, 87, 91), and add a row for `gitx.FirstLine`, newly exported in this window.
3. Record that Task 12's Done-when clause "the prose states … the line format" is met in `helptext/queue.md` but **not** in `construct/datatype/queue.md`, and say which one is intended to carry it.
4. Record the `project:` prefix as the project-line marker — a design decision made during implementation that appears in no plan or Spec text, and the one finding A turns on.
5. Record that `queue add/remove/move` are `guardSpineRepo`'d while the bare list is not, since that is a lifecycle-surface change the plan never contemplated.

```findings
dispose:
  - id: BR-11
    disposition: not-addressed
    note: |
      trunkfile.go:191-193 and Update:317 are untouched by this round (trunkfile.go is not in 66065f6). queueRefusal now prints the staleness warning, which was BR-32's half, not this one.
  - id: BR-12
    disposition: not-addressed
    note: |
      Untouched. trunkfile.go:383-385 still explains what --path buys with no sentence noting the attributes resolve from the working tree/index, not from the trunk being written.
  - id: BR-21
    disposition: not-addressed
    note: |
      The fix commit dismisses this as "stale" on the readFrom merge, but the finding's second half is live and reproduced. trunkfile.go:169 still passes the UNQUALIFIED t.branch; I confirmed against real git that with both refs/tags/main and refs/heads/main present, `rev-parse --verify main` and `cat-file blob main:q.md` return the TAG. So ReadDegraded's no-tracking-ref fallback can serve a tag's bytes while warning "used local main". One-line fix, unchanged: pass "refs/heads/" + t.branch.
  - id: BR-23
    disposition: addressed
    note: |
      ValidateTag exists and is pinned twice. Verified by mutation, not by the diff: short-circuiting ValidateTag to nil reddens TestIntent_Validate_CoversEveryInterpolatedField/tag and two sub-cases of TestQueueEdit_ValidationRefusesBeforeTouchingTheTrunk.
  - id: BR-24
    disposition: addressed
    note: |
      Verified by mutation: deleting the in.Validate() call from runQueueEdit reddens all six sub-cases with "Update ran 1 times". Residue, not re-raised: guardSpineRepo and openTrunkStore each run `git rev-parse --show-toplevel` before validation, so queue.go:194's "BEFORE any git call" is literally still false (and the two calls are redundant with each other).
  - id: BR-25
    disposition: not-addressed
    note: |
      Tag and cr are now carried (both mutation-verified), but the Kind cell the finding enumerated is not. Reproduced: `add sdlc-fleet "sharper reason"` onto "- project:sdlc-fleet — the whole area is next [sdlc]" yields "- sdlc-fleet — sharper reason [sdlc]" with the note "updated why-now ... and kind" — the peer's project marker is destroyed by an edit that never mentioned it, now announced rather than silent. Root cause is that KindIssue is Kind's zero value, so "not mentioned" is unrepresentable; fix with a KindUnspecified zero or c.Flags().Changed("project"). Second residue in the same rule: the APPEND branch (intent.go:153) does not carry the document's line ending, so adding to a CRLF file yields mixed endings.
  - id: BR-26
    disposition: not-addressed
    note: |
      The listing half is fully done and pinned (TestQueueList_ReportsUnrecognizedItems), and helptext/queue.md:17-30 now carries the grammar. The datatype-prose half is untouched — construct/datatype/queue.md is not in the fix commit and still says the operational contract "is not restated here", so the Done-when clause "the prose states ... the line format" remains unmet. One short section closes it.
  - id: BR-27
    disposition: addressed
    note: |
      guardSpineRepo is wired into all three write subcommands and the read/write asymmetry is argued in queue.go:22-34. The residue — no test pins it and repoguard.go's enumeration claim is now false — is raised separately as its own finding rather than re-raised here.
  - id: BR-28
    disposition: addressed
    note: |
      The #207 note landed on 000207-sync-without-worktree.md, names which file it chose, and carries the retry-semantics amendment. The plan gained an M2 Revisions entry that is unusually honest about the blanket-regex ticking. Residue carried into the plan-revision recommendations: 04804ff still has no task row.
  - id: BR-29
    disposition: not-addressed
    note: |
      Untouched — trunkfile.go is not in the fix commit. pathPresent:219/modeOf:426 and refPresent:178/resolve:356 both remain, and Update still runs both halves of both pairs per attempt.
  - id: BR-30
    disposition: not-addressed
    note: |
      Half addressed. TestQueueEdit_ValidationRefusesBeforeTouchingTheTrunk can now see a git call (it asserts f.calls == 0, mutation-verified). The replay half is unchanged: fakeTrunk.Update still calls transform exactly once and fires peer BEFORE it, so TestQueueEdit_PeerEditSurvivesTheReplay would pass against an Update with no retry, and queue_e2e_test.go still has no peer push. Also, the `reached` field added for this is incremented at two sites and read at none, while queue_test.go:36-38 credits it with making the claim testable.
  - id: BR-31
    disposition: not-addressed
    note: |
      Untouched. workshop/plans/000209-queue-datatype-plan.md still names queueCmd at lines 81, 87 and 91, and has no row for gitx.FirstLine.
  - id: BR-32
    disposition: addressed
    note: |
      queue.go:249-251 now prints "(this snapshot is itself stale: ...)". The redundant second full fetch on the error path remains and is folded into the BR-11 enumeration.
findings:
  - id: new
    severity: Important
    family: unvalidated-format-field
    title: |
      A field that passes Validate can still render a line that re-parses to a different record, producing entries that duplicate and cannot be removed
    detail: |
      This is the 2nd finding in family unvalidated-format-field. Do NOT patch these two inputs — state the rule. RULE a validator that enumerates forbidden characters does not make the record round-trip; the write path must assert what the read path already asserts. Two instances reproduced against the real code, both exit 0. (1) Intent{OpAdd, Ref "project:foo", WhyNow "why"} validates, publishes "- project:foo — why", and re-reads as {Ref "foo", Kind KindProject}; `queue remove project:foo` then reports "project:foo was not in the queue; nothing to remove" and leaves it, and a second identical add appends a duplicate instead of converging. ValidateRef rejects brackets, whitespace and the em-dash separator but not the project: prefix, which is structure in that position. (2) Intent{OpAdd, Ref "a#1", WhyNow "  padded  "} validates and publishes "- a#1 —   padded  ", which ParseLine's own self-check rejects — Entries() is empty and UnrecognizedItems() is 1, so the verb manufactures the state runQueueList blames on "a hand-edit that used a hyphen". Fix covering both and every future field - in applyAdd, after building the Line, refuse unless ParseLine(line.String()) yields an equal record, the same mechanical guarantee line.go:110-123 gives the read side (ARCH-DRY, ARCH-SECURE).
  - id: new
    severity: Important
    family: unguarded-write-verb
    title: |
      The spine guard on the queue write verbs is pinned by no test, and repoguard.go's enumeration claim is now false
    detail: |
      This is the 2nd finding in family unguarded-write-verb. Do NOT just add three assertions — state the rule. RULE a guard is wired only when it is in the enumeration the drift test walks; a guard reachable solely by remembering to call it is a comment. repoguard.go:8-11 states the guard is wired into "exactly the lifecycle verbs (... processmanual.WorkflowVerbs; the drift test enumerates it)", which is now untrue - queue add/remove/move call guardSpineRepo and are outside WorkflowVerbs, correctly so since adding them there would distort the #172 friction instrument. Consequence - no test constructs NewQueueCmd's command tree at all, so nothing pins the guard, the --project to Kind wiring, --tag, or the --before/--after mutual exclusion, and a refactor dropping the guard ships silently from the base-layer binary to every downstream repo. Separately, newQueueMoveCmd runs flag validation before the guard, contrary to the guard-first placement repoguard.go:58-60 states. Fix - correct the "exactly" sentence to name both sets, and add a brain-repo table test over the three write subcommands plus one asserting the bare list is NOT guarded.
  - id: new
    severity: Minor
    family: duplicated-helper
    title: |
      The item-marker predicate is written twice with different rules, so indented prose bullets are reported as unparsed entries
    detail: |
      This is the 5th finding in family duplicated-helper. Do NOT fix this instance — the rule is already stated at trunkfile.go:230-235 and just needs extending from functions to predicates - two implementations of one predicate are one function, and where they diverge one of them is a bug. line.go:73 tests CutPrefix(trimmed, "- ") on the raw line; doc.go:75 tests HasPrefix(TrimSpace(l.raw), "- "). Reproduced - a queue file whose prose header contains an ordinary indented list ("  - the top line is next") reports 2 lines that "look like entries but do not parse", which is exactly the reader-training failure doc.go:68-71 says the TrimSpace-free rule avoids. TestQueueList_ReportsUnrecognizedItems uses only unindented prose so it cannot see this. Extract looksLikeItem(s string) bool and call it from both.
  - id: new
    severity: Minor
    family: record-region-boundary
    title: |
      add appends at the end of the FILE, so a new entry lands below a trailing comment block, and move --after disagrees with it
    detail: |
      intent.go:153 appends to d.lines unconditionally. Reproduced - adding b#2 to "# Queue\n\n- a#1 — first\n\n<!-- keep this at the bottom -->\n" renders the new entry BELOW the comment, while `move b#2 --after a#1` would place it above. RULE in a document that interleaves records with free text there is a record region, and an insert belongs at that region's edge, not the file's. The file has no prose today so nothing breaks yet, but Doc exists precisely to support hand-authored prose and the plan's Task 8 names a trailing comment block as a case. TestIntent_Apply_PreservesProse asserts the comment survives but not where the entry lands.
```

---

## Re-review — 2026-09-07T19:32:13-07:00 (REWORK)

| field | value |
|-------|-------|
| issue | 209 — queue datatype for advisory work ordering |
| repo | ariadne |
| issue file | workshop/issues/000209-queue-datatype.md |
| boundary | whole-issue close |
| milestone | — |
| window | 00d7c75c5371bbb3e2f01f5e28780565c274f3a9..c3feaefa8faa99b88b59754d9749c0cd6382e1bc |
| command | sdlc close --issue 209 |
| reviewer | claude |
| timestamp | 2026-09-07T19:32:13-07:00 |
| verdict | REWORK |

## Review

```verdict
verdict: REWORK
confidence: high
```

Round 6's two commits are real work — the round-trip validator (`Intent.Validate` renders and re-parses rather than banning substrings) is the right answer to BR-33 and I mutation-verified it goes red when disabled; the datatype prose now carries the line format; `repoguard.go`'s enumeration claim is corrected and pinned by a test that reddens when the guard is deleted. What blocks SHIP is BR-25: `KindUnspecified` was introduced as the representation of "not mentioned," a test was written against it, and the *only production caller* — `queue.go:90` — still hardcodes `queue.KindIssue`, so the fix is reachable from no code path a user can run. I reproduced it against the built binary: `sdlc queue add sdlc-fleet "sharper reason"` onto a peer's `- project:sdlc-fleet — … [sdlc]` publishes `- sdlc-fleet — sharper reason [sdlc]`, exit 0. That is the third round this defect has survived, and it now ships with a green test certifying it fixed — which is worse than shipping it open. Alongside it, four of the twelve open findings (BR-11, BR-12, BR-29, BR-35) were not touched at all this round, and one new Important surfaced: a converged no-op still publishes an empty commit to `origin/main` whose subject claims the edit it did not make.

## 1. Strengths

- **`Intent.Validate`'s round-trip check** (`intent.go:92-117`) is the right generalisation of BR-33 — it asks "does this record survive a write and a read" instead of enumerating another forbidden prefix, and `TestIntent_Validate_AllowsOddButConsistentRecords` proves the check is *precise* rather than a blanket ban. Mutation-verified: `if false && (…)` reddens `TestIntent_Validate_RejectsRecordsThatDoNotSurviveARoundTrip`. BR-33's second instance (`WhyNow: "  padded  "`) is also caught by the mechanism without being enumerated, which is what a rule-level fix looks like.
- **`readFrom`** (`trunkfile.go:236-265`) is the collapse BR-21 asked for, and the doc comment states the rule that produced it rather than just the merge.
- **`TestQueueCmd_WriteVerbsAreSpineGuarded`** — mutation-verified: deleting `guardSpineRepo` from `newQueueAddCmd` reddens it with `"add": guardSpineRepo present=false, want true`. The read/write asymmetry is argued, not assumed.
- **`FuzzDocRoundTrip` plus a committed corpus entry** (`testdata/fuzz/…/98a4570c0f30ed9d` = `"-  0 — 0"`) — a real fuzz find preserved as a regression, and the `parsed bool` field that came out of it is the correct fix for a two-states-in-one-field bug.
- **Atlas** (`sdlc-binary.md:647-727`) genuinely explains *why* — the CAS push as the concurrency primitive, the separate-streams rationale, the offline asymmetry — and routes rather than restates against the datatype and helptext.

## 2. Critical findings

None.

## 3. Important findings

**BR-25 residue — `cmd/sdlc/queue.go:90`.** `kind := queue.KindIssue` runs whenever `--project` is absent, so the verb never constructs `KindUnspecified`. `applyAdd`'s guard (`intent.go:178`, `in.Kind != KindUnspecified && in.Kind != old.Kind`) therefore never protects the CLI path. Reproduced end-to-end against a real bare origin: the peer's `project:` marker is destroyed at exit 0. `TestIntent_Apply_ConvergeDoesNotFlipUnmentionedKind` passes because it constructs the zero-value Intent directly. Fix: `kind := queue.KindUnspecified; if c.Flags().Changed("project") { … }`, and add a `queue_test.go` case that drives `newQueueAddCmd`'s RunE (or at minimum asserts the Intent the verb builds).

**New — a converged no-op publishes an empty commit to the trunk that claims the edit.** Reproduced: `sdlc queue remove nonexistent#99` prints `nothing to remove` on stderr and then pushes commit `queue: remove nonexistent#99` with an empty diff to `origin/main`. Same for a re-`add` that changes nothing. `trunkfile.go:342` calls `commitAndPush` unconditionally on whatever the transform returned. The Spec's row reads "Converge. No-op, succeed," and the storage rationale is "the history is the point — `git log workshop/queue.md` records how priority actually moved." Fix in `Update`: `if bytes.Equal(next, old) { return nil }` before `commitAndPush`. This is the 3rd finding in family `error-misattribution` — the rule (a message may only name state the code observed) had been swept over stderr sites only; the enumeration must include the **commit subject**, which is the most durable message this code emits.

**BR-34 residue — `cmd/sdlc/queue.go:131-134`.** `newQueueMoveCmd` returns the `--before`/`--after` error *before* `guardSpineRepo`, contradicting `repoguard.go:64-66` ("Called first in each lifecycle verb's RunE — before flag validation, so the refusal … is what the agent sees"). Also, the delivered test is a source-text grep, so it structurally cannot observe ordering — it would pass on this. Move the guard above the flag check.

**BR-21 residue — pinned by no test.** The behaviour is now correct; I verified both directions against real git. But reverting `localRef()` to `return t.branch` leaves the entire `gitx` package green. Add this (it fails on the mutation and passes on HEAD):

```go
func TestTrunkFile_LocalFallbackUsesTheBranchNotATagNamedForIt(t *testing.T) {
	repo := testfix.Repo(t, testfix.InitialCommit())
	// commit "from branch", then commit "from TAG", tag it `main`, reset back
	// … then ReadDegraded must return "from branch\n", not the tag's bytes.
}
```

## 4. Minor findings

- **BR-11** — all three sites reproduced unchanged. A reachable origin with no `main` yields `origin unreachable (offline?): fatal: couldn't find remote ref refs/heads/main`; a repo whose local branch is `master` yields `bytes="" err=nil warn="…used local main"` — a fabricated empty value carrying a claim that did not happen (`ARCH-SECURE`).
- **BR-12** — `trunkfile.go:392-394` still has no sentence noting `.gitattributes` resolves from the working tree/index, not from the trunk being written.
- **BR-29** — `pathPresent:228`/`modeOf:435` (same `ls-tree`, differ by `--name-only`) and `refPresent:187`/`resolve:365` (same `rev-parse --verify`, differ by `--quiet`) untouched; `Update` still runs both halves of both pairs per attempt.
- **BR-30** — `fakeTrunk.Update` still calls `transform` once with `peer` firing before it, so `TestQueueEdit_PeerEditSurvivesTheReplay` would pass against an `Update` with no retry; `queue_e2e_test.go` still has no peer push. The `reached` field is incremented at two sites and read at none, while the fake's doc comment (`queue_test.go:22`) credits it with making the claim testable.
- **BR-31** — plan still names `queueCmd` (symbol is `NewQueueCmd`) at lines 81/87/91; no row for `gitx.FirstLine`.
- **BR-35** — reproduced: `Parse("# Queue\n\n  - the top line is next\n  - then this one\n\n- a#1 — real\n")` gives `entries=1 unrecognized=2`. `line.go:82` uses `CutPrefix(trimmed, "- ")`; `doc.go:86` uses `HasPrefix(TrimSpace(raw), "- ")`.
- **BR-36** — reproduced: adding `b#2` to a doc ending `<!-- keep this at the bottom -->\n` renders the entry *below* the comment.
- **New** — `TestIntent_Validate_CoversEveryInterpolatedField`'s "Apply must refuse too (defence in depth)" assertion holds only for `OpAdd`; `applyMove`/`applyRemove` never call `Validate`. A *valid* move on an empty doc errors identically, so the assertion cannot distinguish a validated intent from an unvalidated one.

## 5. Test coverage notes

`internal/queue` is genuinely PURE — its tests run with no git, no fs, no clock (verified by running them). `gitx` runs 25 tests against a real bare origin (`ARCH-MOCK` satisfied for the primitive). The gap is at the *seam*: no test anywhere drives `Intent.Apply` through a real non-fast-forward rejection. `gitx`'s two-publisher test uses a raw append transform; the verb's replay test uses a fake that calls the transform once; the e2e has no peer. So the single property the design exists for is asserted only where it cannot be observed. Second gap: nothing constructs the cobra command tree and exercises a RunE, which is why the `--project`→`Kind` wiring defect (BR-25) is invisible to the suite. `go test ./cmd/sdlc/...` is green except the pre-existing `TestFleetPlanHasAuthoritativeCorrectedCoreConceptInventory` (ariadne#210).

## 6. Architectural notes

- **ARCH-DRY — flag.** `duplicated-helper` is at 5 findings; BR-29 and BR-35 are both untouched this round. The rule is *already written in the code* at `trunkfile.go:239-244`; it has not been swept.
- **ARCH-PURE — pass.** Clean split: merge policy in the pure core, CAS in `TrunkFile`, `queue.go` is argument parsing and rendering only.
- **ARCH-PURPOSE — flag.** `KindUnspecified` is a single-source change whose one consumer does not derive from it (BR-25) — a deferred consumer, not a finished one. And a fix answered at the type layer while the caller that motivated it was left alone is the instance, not the class.
- **ARCH-MOCK — pass with caveat.** Real bare origin throughout `gitx`; `fakeTrunk` is stateful, not a call mock. Caveat: it does not model the retry interleaving its real counterpart guarantees (BR-30) — the same class of gap that produced BR-24's false test.
- **ARCH-CONSTRAINTS — pass.** One fetch per attempt (mutation-pinned), `signs()` hoisted, bounded at 3. Residues: `queueRefusal`'s second full fetch, and the empty-commit write-amplification on shared `main`.
- **ARCH-SECURE — mostly pass.** Fuzz-guarded parse of untrusted trunk bytes, round-trip write validation, 0700 `MkdirTemp` index with no unclaimed-name window, `--end-of-options` on `ls-tree`, structural argv. Flag: BR-11 site 3 substitutes a fabricated empty value with `err=nil` and a claim downstream reads as evidence.
- **ARCH-ORDER — flag.** The interleaving table is written down and `Update` observes git's state rather than parsing its prose. But BR-30 is exactly this entry's highest-leverage at-review clause: tests that can only observe one interleaving, with no seam to inject another.

## 7. Plan revision recommendations

A `## Revisions` entry covering:

1. **Core-concepts table corrections (BR-31):** `queueCmd` → `NewQueueCmd`; add a row for `gitx.FirstLine` (`cmd/sdlc/internal/gitx/trunkfile.go`, new — moved out of `issueids.go`, now package surface); add a row for `cmd/sdlc/queue_e2e_test.go` (commit `04804ff`), which is real delivered work with no task row.
2. **The `Kind` wiring is a verb-layer contract, not a type-layer one:** record that `KindUnspecified` only means "not mentioned" if the cobra layer forwards flag-presence, and name `c.Flags().Changed("project")` as the seam — otherwise the next reader repeats BR-25.
3. **`Update`'s no-op contract:** the plan's Task 4 sequence has no "skip the push when the transform returned the base unchanged" step, so the implementation followed the plan correctly and still publishes empty commits. Add it to the plumbing sequence.

```findings
dispose:
  - id: BR-11
    disposition: not-addressed
    note: |
      All three sites reproduced unchanged against real git; trunkfile.go:200 and :166-169 untouched this round.
  - id: BR-12
    disposition: not-addressed
    note: |
      trunkfile.go:392-394 untouched — still no sentence on where .gitattributes resolves from.
  - id: BR-21
    disposition: not-addressed
    note: |
      Behaviour now correct (verified both directions), but mutating localRef back to the unqualified branch leaves the whole gitx package green — no test pins it.
  - id: BR-25
    disposition: not-addressed
    note: |
      Fixed at the type layer and unreachable from the CLI — queue.go:90 hardcodes KindIssue, so KindUnspecified is produced at zero production call sites; reproduced against the built binary at exit 0.
  - id: BR-26
    disposition: addressed
    note: |
      construct/datatype/queue.md now carries a "The line format" section stating the grammar, the project: prefix and the why-now restrictions; the listing half was already pinned.
  - id: BR-29
    disposition: not-addressed
    note: |
      trunkfile.go not in this round's commit for these sites — pathPresent/modeOf and refPresent/resolve both remain, and Update still runs all four per attempt.
  - id: BR-30
    disposition: not-addressed
    note: |
      fakeTrunk still calls transform once with peer firing before it; queue_e2e_test.go still has no peer push; `reached` is still written at two sites and read at none.
  - id: BR-31
    disposition: not-addressed
    note: |
      Plan lines 81/87/91 still say queueCmd; no gitx.FirstLine row; no row for the e2e (04804ff).
  - id: BR-33
    disposition: addressed
    note: |
      Mutation-verified — disabling the round-trip check reddens TestIntent_Validate_RejectsRecordsThatDoNotSurviveARoundTrip; the padded-why-now instance is caught by the same mechanism without being enumerated.
  - id: BR-34
    disposition: not-addressed
    note: |
      Claim corrected and presence pinned (mutation-verified), but the guard-first residue is live — newQueueMoveCmd still validates --before/--after before guardSpineRepo, and a source-grep test cannot observe ordering.
  - id: BR-35
    disposition: not-addressed
    note: |
      Reproduced — an indented prose list yields unrecognized=2; line.go:82 and doc.go:86 still disagree on the item-marker predicate.
  - id: BR-36
    disposition: not-addressed
    note: |
      Reproduced — a new entry renders below a trailing comment block; intent.go:190 still appends to the end of the file.
findings:
  - id: new
    severity: Important
    family: error-misattribution
    title: |
      A converged no-op still pushes an empty commit to origin/main whose subject claims the edit it did not make
    detail: |
      This is the 3rd finding in family `error-misattribution`. Do NOT fix this
      instance — state the rule and sweep it. RULE a message may only name state
      the code observed, and a COMMIT SUBJECT is the most durable message this
      code emits. Earlier rounds swept the rule over stderr only (offlineError,
      ReadDegraded's warn, queueRefusal's snapshot, Applied.Note), so the commit
      subject was never in the enumeration. Reproduced against a real bare
      origin: `sdlc queue remove nonexistent#99` prints "nothing to remove" and
      then publishes commit "queue: remove nonexistent#99" with an empty diff;
      a re-add that changes nothing does the same. trunkfile.go:342 calls
      commitAndPush unconditionally on whatever the transform returned. The
      Spec's interleaving row says "Converge. No-op, succeed", and the storage
      rationale is that `git log` records how priority actually moved. Fix at
      the primitive so every consumer inherits it: in Update, return nil without
      committing when the transform's output equals the base bytes (ARCH-DRY,
      ARCH-CONSTRAINTS — it also stops write-amplifying shared main).
  - id: new
    severity: Minor
    family: vacuous-test-guard
    title: |
      The "Apply must refuse too (defence in depth)" rows pass for an unrelated reason on move and remove, because Apply validates only OpAdd
    detail: |
      This is the 3rd finding in family `vacuous-test-guard`. Do NOT patch the
      row — state the rule. RULE a test may not be named for a property its
      fixture satisfies for an unrelated reason; the assertion must be able to
      distinguish the property present from the property absent.
      intent_test.go:256's "move anchor" case applies to Parse(nil), so it gets
      ErrSubjectMissing — and a perfectly VALID move on the same empty doc
      errors identically, so the assertion cannot tell a validated intent from
      an unvalidated one. Verified: applyMove (intent.go:216) and applyRemove
      (:200) never call Validate; only applyAdd (:149) does. Either extend
      Validate's call to all three arms and assert the returned error is the
      VALIDATION error, or drop the "defence in depth" claim for the arms that
      do not have it.
```

---

## Re-review — 2026-09-07T19:51:21-07:00 (FIX-THEN-SHIP)

| field | value |
|-------|-------|
| issue | 209 — queue datatype for advisory work ordering |
| repo | ariadne |
| issue file | workshop/issues/000209-queue-datatype.md |
| boundary | whole-issue close |
| milestone | — |
| window | 00d7c75c5371bbb3e2f01f5e28780565c274f3a9..7770416999cfc8e47343a24c29a0f5975aa0447d |
| command | sdlc close --issue 209 |
| reviewer | claude |
| timestamp | 2026-09-07T19:51:21-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

The shipping code is correct and, on the paths that matter, genuinely well tested: `gitx.TrunkFile` runs against a real bare origin with the peer's push injected at the transform seam, so the retry contract, the attempt bound, the non-retryable fast-fail and the fetch budget are all pinned by tests that fail when mutated (I verified four of this round's claims by reverting them in a scratch copy — BR-21, BR-25, BR-37 and the mode/signing guards all go red). The datatype, helptext and atlas are complete, and the seed really was written through the verb (`origin/main` carries four entries under `queue: add …` / `queue: move …` commits). What blocks a clean SHIP is not the code but its pinning: three of this round's fixes are held by a test that reads `queue.go` as *text* or by no test at all — I deleted the `guardSpineRepo` call from `sdlc queue add`, left a comment naming it, and the whole suite stayed green, while a 20-line behavioural test built from the seam this repo already has (`hermeticRepo` + `writeBrainMarker` + `expectDie`, used one file over in `repoguard_test.go`) went red immediately. That is the same defect class the last four rounds have been closing, and it is cheap to close for real. Everything else outstanding is Minor doc/plan hygiene.

**1. Strengths**

- `cmd/sdlc/internal/gitx/trunkfile_test.go:174,207,322,578` — the interleaving is injected *inside the transform*, so the peer's push lands at a chosen point against a real origin. This is the ARCH-MOCK/ARCH-ORDER shape done right: a function-call mock could not produce the non-fast-forward rejection, and the tests observe more than one ordering.
- `cmd/sdlc/internal/queue/line.go:129` — the parse is accepted only if it renders back byte-identically. Making the round-trip true *by construction* rather than by five separate trim arguments is the elegant answer, and `FuzzDocRoundTrip` (with a persisted corpus entry) is the mechanical backstop the Spec asked for.
- `cmd/sdlc/internal/queue/intent.go:98-122` — validation by render-and-re-parse instead of another banned substring, plus `TestIntent_Validate_AllowsOddButConsistentRecords` proving the check is *precise* rather than a blanket prefix ban. That test is the one that stops the rule from over-firing.
- `cmd/sdlc/internal/gitx/trunkfile.go:346` — the no-op skip landed at the **primitive**, so `ariadne#207` inherits it; mutation-verified (removing it moves the trunk and `TestTrunkFile_UnchangedContentPushesNothing` fails).
- `atlas/workflow/sdlc-binary.md:647` — "the noun is in `construct/datatype/queue.md`, the verb contract is in `sdlc queue --help`; neither restates the other." Routing instead of restating is the right instinct (see Minor 3 for where the diff doesn't follow it).
- `workshop/plans/000209-queue-datatype-plan.md:368` — the Revisions entry admitting the M2 TDD checkboxes were ticked by a blanket regex without the red-green order being followed. Recording that rather than quietly re-ticking is worth more than the ticks were.

**2. Critical findings**

None.

**3. Important findings**

- **`cmd/sdlc/queue_test.go:299,326,351,386` — this round's fixes are pinned by source-text greps or by nothing.** *4th finding in family `vacuous-test-guard`.* Do not patch these three sites — the rule is: **a test must be able to distinguish the property present from the property absent; reading the implementation's own source text cannot, because a comment satisfies it, and no test at all cannot.** Enumeration over this round's fixes, each mutation-verified against the real tree:
  1. `subcommandGuardSource` (`queue_test.go:326`) reads `queue.go` and greps. Replacing `guardSpineRepo(c.ErrOrStderr())` in `newQueueAddCmd` with `// guardSpineRepo runs first here` leaves `go test ./cmd/sdlc/` **green**; a behavioural probe over the three write subcommands failed on the same mutation. Its justifying comment ("invoking the verb would exit the process via `die()`") is false in this codebase — `die` is a package-level var and `expectDie` (`die_test.go:33`) exists for exactly this, which is how `repoguard_test.go:39` tests the sibling guard.
  2. `doc.go:76` `looksLikeItem` — reintroducing the divergence (`HasPrefix(TrimSpace(raw), "- ")` in `UnrecognizedItems`) leaves every package green.
  3. `intent.go:199-205` insert-after-last-entry — collapsing it back to `at := len(out.lines)` leaves every package green.
  Fix as one: a `queue_guard_test.go` using `hermeticRepo`/`writeBrainMarker`/`expectDie` over `{add, remove, move}` plus a not-guarded case for the bare list (and, on the same command tree, `--project` → `KindProject` and the `--before/--after` exclusion, which today are only greps), and two six-line table cases in `internal/queue` for the indented-prose count and the trailing-comment insert position.

- **BR-34 remains open on its core claim.** The `repoguard.go:8-17` enumeration sentence is now correct and `newQueueMoveCmd` puts the guard first — both real. But "the guard is pinned by no test" is still true under mutation (above), so a refactor dropping it still ships silently from the base-layer binary to every downstream repo.

**4. Minor findings**

- BR-11, BR-12, BR-29, BR-30, BR-31, BR-38 all verified unchanged this round — see the dispositions below for the file:line evidence.
- The line format and the why-now validation rules are stated four times (`line.go:159-178`, `construct/datatype/queue.md:56-61`, `cmd/sdlc/helptext/queue.md` "LINE FORMAT", `queue.go:184-187`) with nothing tying them to the validators; `fleet_readme_test.go:10` is the in-repo precedent for the cheap doc-conformance test that would.
- `queue_test.go:308` reports `present=%v` from `!guarded[name]`, i.e. from the expectation rather than the observation. Correct only by coincidence of the branch condition.

**5. Test coverage notes**

`go test ./cmd/sdlc/...` is green except the pre-existing `TestFleetPlanHasAuthoritativeCorrectedCoreConceptInventory` (ariadne#210, unrelated). The real gap is not breadth but *pinning*: no test anywhere drives `Intent.Apply` through a genuine non-fast-forward rejection — `gitx`'s two-publisher tests use a raw append transform, `fakeTrunk.Update` calls the transform exactly once with `peer` firing *before* it, and `queue_e2e_test.go` has no peer push. `TestQueueEdit_PeerEditSurvivesTheReplay` therefore still passes against an `Update` with no retry at all, despite being named for "the property the whole design exists for" (BR-30).

**6. Architectural notes**

- **ARCH-DRY — flag** (BR-29): `pathPresent`/`modeOf` and `refPresent`/`resolve` are still one function each, split in two; `Update` runs all four per attempt. Pass elsewhere — `firstLine`→`gitx.FirstLine`, the `readFrom` collapse and `looksLikeItem` are all correct consolidations.
- **ARCH-PURE — pass.** `internal/queue` has no git, no fs, no clock; `queue.go` is argument parsing and rendering only; the CAS lives behind `TrunkFile`.
- **ARCH-PURPOSE — pass on the deliverable** (datatype + verb + a seed genuinely written through the verb onto `origin/main`), **flag on the method**: this round again fixed the sites the findings named without writing the enumeration that would sweep the class — see the Important finding.
- **ARCH-MOCK — pass at `gitx`, flag at the verb.** Real bare origin throughout `trunkfile_test.go`; but `fakeTrunk` is a double that cannot reject, and the e2e (its only conformance path) has no peer, so the divergence is unobservable.
- **ARCH-CONSTRAINTS — pass with residue.** `signs()` hoisted out of the loop, one fetch per attempt (pinned), bound of 3. Residue: two redundant git calls per attempt (BR-29) and no context/timeout on fetch/push — the latter is a declared, reasoned residue in the plan.
- **ARCH-SECURE — pass.** Typed validation at the boundary before any git call, argv (never shell) interpolation, temp index inside a 0700 `MkdirTemp` dir with no unclaimed-name window, fuzz over the hand-editable input. Undocumented nuance: `hash-object --path` resolves `.gitattributes` from the working tree, not the trunk (BR-12).
- **ARCH-ORDER — pass at the primitive, flag at the verb.** The interleaving table is enumerated in the Spec and each cell has a case; ordering is injectable at the transform seam. At the verb layer only one interleaving is observable (BR-30).
- **Docs gate:** atlas is updated for both the primitive and the verb/noun, and `atlas/index.md` links it. README is deliberately *not* a verb catalog in this repo (`migrate` #179 and `issue sync` #206 are likewise absent), so no README finding.

**7. Plan revision recommendations**

- Core-concepts table: rename `queueCmd` → `NewQueueCmd` at lines 81, 87 and 91; add a row for `gitx.FirstLine` (newly exported into the package surface in this window) and for `cmd/sdlc/queue_e2e_test.go` (BR-31).
- Add a `## Revisions` entry for the design changes forced by close-review rounds 3–7 — `KindUnspecified` as the zero value, converge-by-merge, `Update`'s no-op skip, insert-after-last-entry, one item-marker predicate. "M2 as built" predates all five, so the plan currently describes a design the code has moved past.

```findings
dispose:
  - id: BR-11
    disposition: not-addressed
    note: |
      trunkfile.go:200-202 and :166-170 unchanged; Update:326 still calls offlineError for a repo with no origin at all. Round 7 stated this exact rule and swept commit subjects, but not the family's oldest instance.
  - id: BR-12
    disposition: not-addressed
    note: |
      grep for "attributes" in trunkfile.go returns exactly one line (:399); still no sentence on where .gitattributes resolves from.
  - id: BR-21
    disposition: addressed
    note: |
      readFrom carries the guard (:245-265), localRef is qualified (:130); mutation-verified — reverting localRef to t.branch turns TestTrunkFile_LocalFallbackPrefersTheBranchNotATag red.
  - id: BR-25
    disposition: addressed
    note: |
      Merge-not-replace pinned (dropping the KindUnspecified check reddens two tests); CRLF half pinned by TestIntent_Apply_AppendMatchesDocumentLineEnding. The CLI half is pinned only by a source grep — see the new finding.
  - id: BR-29
    disposition: not-addressed
    note: |
      pathPresent:228/modeOf:442 and refPresent:187/resolve:372 both remain; Update still runs all four per attempt.
  - id: BR-30
    disposition: not-addressed
    note: |
      Only the dead `reached` field was removed. fakeTrunk still calls the transform once with peer firing before it, and queue_e2e_test.go still has no peer push, so no test drives Intent.Apply through a real rejection.
  - id: BR-31
    disposition: not-addressed
    note: |
      Plan lines 81, 87 and 91 still say queueCmd; still no gitx.FirstLine row.
  - id: BR-34
    disposition: not-addressed
    note: |
      Enumeration sentence and guard-first ordering are fixed and real; the guard itself is still unpinned — deleting the call and leaving a comment naming it keeps the suite green.
  - id: BR-35
    disposition: addressed
    note: |
      One predicate now (doc.go:76), called from both sites; behaviour verified fixed. Unpinned by any test — folded into the new finding rather than re-raised.
  - id: BR-36
    disposition: addressed
    note: |
      intent.go:199-205 inserts after the last entry; behaviour verified fixed. Unpinned by any test — folded into the new finding.
  - id: BR-37
    disposition: addressed
    note: |
      trunkfile.go:346 skips the write at the primitive so #207 inherits it; mutation-verified — removing it moves the trunk and TestTrunkFile_UnchangedContentPushesNothing fails.
  - id: BR-38
    disposition: not-addressed
    note: |
      applyMove:232 and applyRemove:216 still never call Validate; intent_test.go:263's "Apply must refuse too" still passes on the move row via ErrSubjectMissing.
findings:
  - id: new
    severity: Important
    family: vacuous-test-guard
    title: |
      This round's fixes are pinned by tests that read source text, or by no test at all
    detail: |
      4th in family. Do NOT patch the three sites — state the rule: a test must be able to
      distinguish the property present from absent, and a grep over the implementation's own
      source cannot (a comment satisfies it). Enumeration, each mutation-verified against the
      real tree: (1) queue_test.go:326 subcommandGuardSource — replacing the guardSpineRepo
      call in newQueueAddCmd with a comment naming it leaves go test ./cmd/sdlc/ green, while
      a behavioural probe over add/remove/move failed on the same mutation; the justifying
      comment ("invoking the verb would exit via die()") is false, since expectDie
      (die_test.go:33) exists and repoguard_test.go:39 uses it for the sibling guard.
      (2) doc.go:76 looksLikeItem — reintroducing the divergence leaves every package green.
      (3) intent.go:199-205 — collapsing the insert back to end-of-file leaves every package
      green. One fix covers all three: a behavioural queue-guard test built from
      hermeticRepo/writeBrainMarker/expectDie (which also pins --project -> KindProject and
      the --before/--after exclusion, today only grepped), plus two table cases in
      internal/queue for the indented-prose count and the insert position.
  - id: new
    severity: Minor
    family: format-undocumented
    title: |
      The line format and its validation rules are stated four times with nothing tying them to the validators
    detail: |
      2nd in family. The rule, not the instance: the format has one authoritative statement and
      every other surface routes to it or is checked against it. Today line.go:159-178 defines
      it, and construct/datatype/queue.md:56-61, cmd/sdlc/helptext/queue.md ("LINE FORMAT") and
      queue.go:184-187 each restate the em-dash separator and the newline/control/bracket bans
      independently — so a change to ValidateWhyNow silently falsifies three documents. The
      atlas entry already declares the routing discipline ("neither restates the other"), and
      fleet_readme_test.go:10 is the in-repo precedent for the doc-conformance test that would
      enforce it.
```
