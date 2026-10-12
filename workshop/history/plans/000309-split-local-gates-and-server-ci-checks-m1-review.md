# Boundary Review — ariadne#309 (milestone M1)

| field | value |
|-------|-------|
| issue | 309 — Split local gates and server CI checks |
| repo | ariadne |
| issue file | workshop/issues/000309-split-local-gates-and-server-ci-checks.md |
| boundary | milestone M1 |
| milestone | M1 |
| window | c1afb049c5305d3eb0deeafcfb8e10db3df23a72..80160793457bf91321cc6724cf9b861771e51c28 |
| command | sdlc milestone-close --issue 309 --milestone M1 |
| reviewer | claude |
| isolation | checkout |
| timestamp | 2026-10-12T12:25:10-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

Reviewing the legacy merge path's head pinning, then running the targeted tests.

Timed out at 600s. I'll find which test hung.

The hang is in a git subprocess inside `revalidateLanding`, not the CI gate. Checking whether the same test hangs at base.

Both pass alone (~80s each), so the earlier timeout came from slow tests running under load, not from this change. Next, a mutation check on both call sites.

```verdict
verdict: FIX-THEN-SHIP
confidence: medium
```

M1 does what the issue asks: `sdlc merge` now refuses red or pending CI. Both merge paths call the gate, and a mutation test shows each site's test fails when that call alone is bypassed. The gate's logic is a pure function (`decideCIGate`), the parser is pure (`parsePRChecks`) and fails closed, and both are tested well. The legacy path now pins the merge with `--match-head-commit`, and the help text and atlas are updated. Nothing blocks SHIP. One Important gap: the legacy path never checks that the checks it read belong to the head it pins. The durable path gets this for free from `selectLandingPR`. The legacy path relies on `--match-head-commit` alone, which leaves a narrow window right after a push where an unchecked head can land. One plan bookkeeping item is also open.

**Inspection:** I ran the stat, name-status, full and targeted diffs. `go test` on `TestDecideCIGate`, `TestParsePRChecks`, `TestPRMergeArgsPinTheHead`, `TestRunCIGate`, `TestRunMerge_RedCIRefusesPreMerge`, `TestRunMerge_GreenCIPinsPushedHead` and `TestLandingRedCIRefusesBeforeMerge` passed. The `processmanual` package passed. Running all `CIGate|Merge|Landing` tests together hit the 420s/600s timeout inside `TestLandingPreservesDirtyRestingRecovery`, stuck in a git subprocess under `revalidateLanding`. That test passes alone at both base (79s) and head (85s), so the timeout is the slow suite under load, not this change. Mutation check in a scratch worktree: forcing `noCI=true` at `landing.go:420` made `TestLandingRedCIRefusesBeforeMerge` fail, and doing the same at `merge.go:563` made `TestRunMerge_RedCIRefusesPreMerge` fail.

### 1. Strengths
- `cigate.go:56-83`: `decideCIGate` is pure, and the order red > unknown > pending is explicit. An unknown bucket fails closed. The test table covers every bucket, each refusal row checks its own reason, and the refusal names every non-green check.
- `cigate.go:132-152`: `parsePRChecks` reads stdout before the exit code, which matches gh's exit 1 and exit 8 behaviour. It treats the "no checks reported" message as an empty list, treats auth failures, empty output and truncated JSON as errors, and rejects a check that has no name or bucket.
- `ghclient.go:170`: `prMergeArgs` is pulled out and tested against the exact argv. The legacy merge is now pinned to a head (`TestRunMerge_GreenCIPinsPushedHead` asserts the pinned SHA equals `origin/feature`).
- `landing.go:418-422`: the gate sits after the publish gate and before `revalidateLanding`/`LandingMerge`, as the plan says, and the head it uses was already matched to `pr.HeadOID` by `selectLandingPR`.
- `stubGH.PRChecks` returns green by default, so existing merge tests keep passing without each one being edited.

### 2. Critical
None.

### 3. Important
- **ARCH-ORDER, `merge.go:552-566`:** the legacy path pins `head` from the local `origin/<branch>` tracking ref. `gh pr checks <N>` then reports checks on whatever head GitHub's PR object shows at that moment, and nothing confirms the two match. In the usual "push, then `sdlc merge` right away" flow, GitHub's PR head can still be the previous commit. The old commit's green checks pass the gate, and if GitHub syncs before `gh pr merge --match-head-commit <new>` runs, the new commit lands without any checks. The plan's D1 ("the checks read must belong to the head that merges") does not hold on this path. **Fix:** in `runCIGate`, or just before it on the legacy path, read the PR's `headRefOid` (`gh pr view N --json headRefOid`, or reuse `LandingPRs`). Refuse if it is not `head`, and treat a mismatch as pending. Add a fake test where the PR head differs from the pinned head.

### 4. Minor
- `cigate_test.go:68`: `exitErr` builds the exit command with `string(rune('0'+code))`, which only works for codes 0–9. Use `strconv.Itoa`.
- `gatesig.go:163`: `RefusalPat` does not match the gate's other two refusals: the `CI gate: read <head>'s tree…` error and `CI gate: cannot pin the merge`. Those refusals won't be attributed to the gate.
- Plan: the M1 bullet "dry-run reports what the gate would say" was dropped. That is recorded only in the issue Log, not in the plan's `## Revisions`.
- The legacy path does not fetch, so a stale `origin/<branch>` makes `--match-head-commit` fail with a raw `gh pr merge` error. That fails safe, but the message doesn't say "re-fetch or push, then retry".

### 5. Test coverage notes
- The pure layer, the adapter fixtures (real pass and no-checks output, constructed fail and pending output) and both call sites are covered, and both call sites are confirmed by the mutation check.
- The landing `--no-ci` test only checks that the merge happened, not that the warning was printed. `TestRunCIGate` does check the warning.
- Missing: a test where the PR head differs from the pinned head (see the Important finding).
- The fail and pending fixtures are still constructed. The plan replaces them after the post-landing live check.

### 6. Architectural notes
- **ARCH-DRY:** pass. There is one `runCIGate` and one `decideCIGate` for both paths. The three fake `PRChecks` overrides are near-copies of each other, which is acceptable for test seams.
- **ARCH-PURE:** pass. Decision and parsing are pure. The gh call and the git `ls-tree` are injected (`ghCaller`, `git func`).
- **ARCH-PURPOSE:** pass for M1's slice of Done-when ("refuses red or pending"), apart from the head-binding gap above.
- **ARCH-MOCK:** pass with a note. gh sits behind the `ghCaller` seam, and stateless fakes are enough for a one-shot read. Conformance with real gh is one live check after landing, not a recurring one, so a future change to gh's JSON or exit codes would show up only as fail-closed refusals.
- **ARCH-CONSTRAINTS:** pass. gh has a 2-minute timeout with `WaitDelay`, and the gate makes one call per merge.
- **ARCH-SECURE:** pass. Arguments are passed as an argv array, `landingRepoValid` is checked, and the external JSON is parsed into a typed struct and validated. An unreadable result refuses instead of being read as green.
- **ARCH-ORDER:** flagged (Important above). The durable path confirms the checks belong to the head via `selectLandingPR` plus the pin; the legacy path relies on the pin alone.
- **ARCH-FUNERAL:** pass. The change creates nothing durable: the gate is a read plus a log line.

### 7. Plan revision recommendations
- `## Revisions`: "2026-10-12, M1: dropped 'dry-run reports what the gate would say'. Neither path's dry-run contacts GitHub for PR state, and adding that is outside this issue."
- If the Important fix lands: amend D1 to say both paths confirm the PR's `headRefOid` equals the pinned head before trusting `gh pr checks`.

```findings
findings:
  - id: new
    severity: Important
    family: checks-bound-to-merged-head
    title: |
      Legacy merge path never verifies the checks gh reports belong to the pinned head
    detail: |
      merge.go:552-566 pins head from local origin/<branch> but gh pr checks reads GitHub's current PR head; right after a push the PR object can lag, old-head green checks pass the gate, and --match-head-commit only protects if GitHub has not yet synced. Compare PR headRefOid to head (treat mismatch as pending) and add a fake test.
  - id: new
    severity: Minor
    family: plan-revisions-record-scope-changes
    title: |
      Dropped M1 dry-run bullet is recorded only in the issue Log, not the plan Revisions
    detail: |
      Add a Revisions entry so the plan stops claiming the dry-run behaviour.
  - id: new
    severity: Minor
    family: gate-catalog-covers-all-refusals
    title: |
      no-ci RefusalPat misses the head-tree-read and cannot-pin refusals
    detail: |
      gatesig.go:163 matches "PR N has|reports" and "could not read", but not "CI gate: read <head>'s tree" or "CI gate: cannot pin the merge".
  - id: new
    severity: Minor
    family: test-helper-correctness
    title: |
      exitErr test helper builds the exit code via rune arithmetic, valid only for 0-9
    detail: |
      cigate_test.go:68; use strconv.Itoa.
```

---

## Re-review — 2026-10-12T12:56:15-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 309 — Split local gates and server CI checks |
| repo | ariadne |
| issue file | workshop/issues/000309-split-local-gates-and-server-ci-checks.md |
| boundary | milestone M1 |
| milestone | M1 |
| window | c1afb049c5305d3eb0deeafcfb8e10db3df23a72..1f915d6065bf307057fd2c7c4a88d6bdcfe67b13 |
| command | sdlc milestone-close --issue 309 --milestone M1 |
| reviewer | claude |
| isolation | checkout |
| timestamp | 2026-10-12T12:56:15-07:00 |
| verdict | SHIP |

## Review

Prior-round fixes look in place; now checking whether the broadened catalog pattern holds up and running the tests.

```verdict
verdict: SHIP
confidence: high
```

All three Important/Minor fixes from round 1 hold up, and the one that changes behavior (BR-1) was confirmed by reverting it. `runCIGate` now reads GitHub's PR head before and after `gh pr checks` and refuses unless both match the pinned head (`cmd/sdlc/cigate.go:216-230`). Both merge paths share this code. With the comparison disabled, `TestRunCIGate` (heads `[old old]` passed) and `TestRunMerge_LaggingPRHeadRefuses` (`died=false calls=1`) both failed. After restoring it, both suites pass: `go test ./cmd/sdlc/` passed in 508s and `processmanual` passed. BR-2 and BR-4 are fixed. One gap remains, and it does not block the ship. BR-3 broadened the catalog pattern to `CI gate: `, but no classifier test covers any `no-ci` refusal. Reverting `gatesig.go` to the earlier, narrower pattern leaves `processmanual` green. Under the #194 rule, BR-3 therefore stays not-addressed (Minor).

1. **Strengths**
   - `cigate.go:211-230`: the check reads the head before and after the checks. This also catches GitHub's head moving *during* the read (the `{"h","new"}` and `{"old","h"}` cases), not just a lag that is already there.
   - `unread` gives every query failure the same refusal wording. "Unread is not green" (fail closed) is applied the same way to the head read and the checks read.
   - The `e2eGH.PRHeadOID` fake (`merge_e2e_test.go:120`) defaults to the real pushed `origin/feature` and can play back a scripted lagging sequence. Production and tests use the same `ghCaller` seam, so ARCH-MOCK passes.
   - Success lines were renamed to `CI checks:`. That makes the claim "every line with the `CI gate: ` prefix is a refusal" true by construction; `grep` confirms the only non-test sites are `cigate.go` and `merge.go:557`.

2. **Critical:** none.

3. **Important:** none.

4. **Minor**
   - BR-3 has no regression test. Add `TestClassifyOutputLine` cases for:
     - each `ciVerdict.refusal` reason, `unread`, the tree-read error and `cannot pin` (expect refusal, `no-ci`, `ObsFlagOmitted`);
     - the `CI checks:` success line and the `--no-ci` ACK as negatives.

     The `RefusalPat` doc comment at `gatesig.go:56` still says patterns are "grammar+digit-anchored". The new prefix pattern is neither, so it relies on verb anchoring plus `isContamination`, and the comment should say so.
   - The `ciHeadLag` refusal falls into the default advice, "wait for the checks … stays pending here; only then is `--no-ci`". For a lagging head, better advice is "wait for GitHub to sync the push, then re-run". This is cosmetic.

5. **Test coverage**
   - The BR-1 bracket is pinned in both the unit test and the legacy end-to-end test, and the revert proves both.
   - The durable path's `landingFakeGH.PRHeadOID` always returns the matching head, so a lag on that path is covered only through the shared `runCIGate` unit test. That is acceptable.
   - An inherent limit, outside this round's scope: if GitHub has registered only some workflows for a new head, the green ones can still pass.

6. **Architecture**
   - **ARCH-DRY: pass.** There is one shared `runCIGate` and one `unread` helper.
   - **ARCH-PURE: pass.** `decideCIGate` and `parsePRChecks` are pure; the IO goes through the injected `gh` and `git` calls.
   - **ARCH-PURPOSE: pass.** Both merge paths are gated. BR-3 was fixed for the whole class (a prefix rule), not just the reported instance, though the test is missing.
   - **ARCH-MOCK: pass.** The fakes are stateful and sit behind the `ghCaller` seam.
   - **ARCH-CONSTRAINTS: pass.** Each check makes two extra `gh pr view` calls, which is bounded.
   - **ARCH-SECURE: pass.** The OID is checked with `landingOIDValid`, and the repo and number are validated before use.
   - **ARCH-ORDER: pass.** Comparing the head before and after is the correct evidence that the checks were read at a single head, and `--match-head-commit` covers anything that happens after the read.
   - **ARCH-FUNERAL: pass.** The change creates nothing durable.

7. **Plan revisions:** none needed. The BR-1 and BR-2 entries in the plan match the code.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      cigate.go:216-230 reads the head before and after the checks; reverting the comparison turns TestRunCIGate and TestRunMerge_LaggingPRHeadRefuses red.
  - id: BR-2
    disposition: addressed
    note: |
      The plan's Revisions section now records the dropped dry-run bullet (2026-10-12, M1 implementation).
  - id: BR-3
    disposition: not-addressed
    note: |
      The prefix pattern covers the class, but no classifier test pins any no-ci refusal: reverting gatesig.go to the 80160793 pattern leaves processmanual tests green. Add TestClassifyOutputLine cases for every refusal string, plus negatives for the CI checks: success line and the ACK.
  - id: BR-4
    disposition: addressed
    note: |
      cigate_test.go:72 now uses strconv.Itoa.
```
