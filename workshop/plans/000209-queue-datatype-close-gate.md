---
gate: boundary-review
issue: 209
id_prefix: BR
rounds:
    - "n": 1
      timestamp: "2026-09-07T17:17:20-07:00"
      agent: claude
      findings:
        - id: BR-1
          severity: Important
          title: Temp-index assertions in TestTrunkFile_TransformErrorLeavesNoIndexAndNoCommit are unreachable
          detail: |-
            trunkfile_test.go:258-265 nests the os.Stat removal check and the
            filepath.IsAbs check under `if seen != ""`, but Update calls the transform
            before commitAndPush, and GIT_INDEX_FILE is only set at trunkfile.go:271
            inside commitAndPush — so the spy never sees it and the block never runs.
            The Done-when clause "the temp index is removed on every exit path,
            including the error paths, and is never $GIT_DIR/index" has no coverage;
            neither does the blob temp file. Assert on a path that reaches
            commitAndPush and t.Fatal when seen is empty so it cannot vacate silently.
          family: vacuous-test-guard
          round: 1
        - id: BR-2
          severity: Important
          title: isNonFastForward matches bare "rejected", retrying non-retryable push refusals
          detail: |-
            trunkfile.go:260. Reproduced: a pre-receive hook declining a push emits
            "! [remote rejected] main -> main (pre-receive hook declined)", which
            substring-matches "rejected". A policy, permission or protected-branch
            refusal is then retried 3 times and reported as "the trunk moved under 3
            attempts" — the wrong next action. "non-fast-forward" and "fetch first"
            already cover the real CAS case. Add a pure table test over the three real
            git outputs.
          family: git-output-string-matching
          round: 1
        - id: BR-3
          severity: Important
          title: isMissingPath carries a phrase git no longer emits, and the no-remote fallback silently depends on it
          detail: |-
            trunkfile.go:185 matches "Not a valid object name"; git 2.50.1 says
            "invalid object name". ReadDegraded's local-branch fallback
            (trunkfile.go:128-134) fires only because that match FAILS, so the
            documented "a repo with no origin reads the local ref" behavior rests on a
            negative match against unpinned prose. If a git version emits the matched
            wording, ReadDegraded returns empty bytes with a warning claiming it read
            the stale ref — a fabricated value (ARCH-SECURE). TestTrunkFile_NoRemoteIsNamed
            asserts only the warning, never the bytes. Sweep the class: pin all three
            classifiers (isMissingPath, isNonFastForward, offlineError) against real git
            output, and prefer rev-parse --verify over prose for ref existence.
          family: git-output-string-matching
          round: 1
        - id: BR-4
          severity: Important
          title: signs() compares the raw config string to "true", so commit.gpgsign = 1|yes|on yields unsigned commits
          detail: |-
            trunkfile.go:319-320. Verified: `git config --get commit.gpgsign` returns
            "1"/"yes" verbatim while `--type=bool` normalizes all of git's truthy forms
            to "true". The function's own comment calls silently-unsigned commits a
            regression, and it is reachable today; the test only exercises the literal
            "true". Use --type=bool and add a "1" row. Compounds with the isNonFastForward
            finding: on a signature-requiring trunk the refused push is misreported as a
            CAS collision.
          family: git-config-bool
          round: 1
        - id: BR-5
          severity: Important
          title: No test for Update on a path absent from the trunk — the exact path M2's seed takes
          detail: |-
            Every Update test seeds queue.md via trunkFixture first, but
            workshop/queue.md does not exist on origin/main, so Task 13's seed is a
            create. I replayed the plumbing by hand against a throwaway bare origin and
            it works, so this is a coverage gap rather than a bug — but it is the only
            path with no regression guard, and it is the next one to run in anger. One
            test on a nested, absent path closes it.
          family: untested-first-write
          round: 1
        - id: BR-6
          severity: Minor
          title: tempIndexPath(dir string) never uses dir
          detail: |-
            trunkfile.go:194 — the parameter is vestigial and the doc comment implies it
            matters. Drop it or use it.
          family: dead-parameter
          round: 1
        - id: BR-7
          severity: Minor
          title: os.CreateTemp then os.Remove leaves a symlink-plant window on the index path
          detail: |-
            trunkfile.go:195-201 and 324-325. os.MkdirTemp plus filepath.Join(dir,
            "index") removes the window (ARCH-SECURE; low exposure on a single-user
            machine, but free to fix).
          family: temp-file-race
          round: 1
        - id: BR-8
          severity: Minor
          title: firstLineOf duplicates firstLine in cmd/sdlc/issueids.go:112
          detail: |-
            trunkfile.go:146. main already imports gitx, so exporting one and deleting
            the other is free (ARCH-DRY).
          family: duplicated-helper
          round: 1
        - id: BR-9
          severity: Minor
          title: Test comment claims runGitIn uses CombinedOutput, which it deliberately does not
          detail: |-
            trunkfile_test.go:202-203. The separation is the correct design and is
            explained at trunkfile.go:34-52; the comment contradicts it in the one place
            a reader checks the rationale.
          family: stale-comment
          round: 1
        - id: BR-10
          severity: Minor
          title: isNonFastForward, isMissingPath and firstLineOf are pure but only exercised through integration tests
          detail: |-
            Direct table tests over these three would have surfaced both
            git-output-string-matching findings, and run in microseconds instead of
            seconds (ARCH-PURE).
          family: pure-logic-untested
          round: 1
        - id: BR-11
          severity: Minor
          title: offlineError labels every fetch failure "unreachable (offline?)"
          detail: |-
            trunkfile.go:104-109 and 140-142. A branch that does not exist on origin, or
            an auth failure, reads to the operator as a network outage.
          family: error-misattribution
          round: 1
        - id: BR-12
          severity: Minor
          title: hash-object --path resolves .gitattributes from the working tree, not from the trunk being written
          detail: |-
            trunkfile.go:281-286. Correct in practice, but "no working tree involved" is
            the type's headline claim and this is the one place it is not strictly true.
            One sentence in the doc comment.
          family: gitattributes-source
          round: 1
        - id: BR-13
          severity: Minor
          title: Plan Tasks 1-6b left unchecked and the plan has no Revisions section despite mid-stream edits
          detail: |-
            workshop/plans/000209-queue-datatype-plan.md — Tasks 6b and 13b were appended
            in this window with no revision note, the seam table still specifies a
            CombinedOutput two-value runGitIn that the code deliberately does not
            implement, and TrunkFile's entry omits the exported ReadDegraded. Commit
            granularity also diverges (Tasks 2+4 merged, 3+5 merged).
          family: plan-traceability
          round: 1
      boundary: M1
      blocked: true
    - "n": 2
      timestamp: "2026-09-07T17:36:16-07:00"
      agent: claude
      dispose:
        - id: BR-1
          disposition: addressed
          note: 'Verified by reverting: moving cleanupIdx() off the defer turns the test red. Blob temp-file cleanup remains unpinned (correct code, no test) — noted in coverage, not re-raised.'
          round: 2
        - id: BR-2
          disposition: addressed
          note: 'Verified by restoring the prose match: TestTrunkFile_NonRetryablePushFailsFast goes red with "transform called 3 times, want 1".'
          round: 2
        - id: BR-3
          disposition: not-addressed
          note: 'Prose matching is gone, but the byte assertion the finding asked for is still absent: I hardwired refExists to true and the entire suite stayed green, so the fabricated-empty-value path remains unpinned. TestTrunkFile_NoRemoteIsNamed:509 still discards the bytes; testfix.InitialCommit writes README="x\n", so this is a one-line assert. The replacement signal''s own conflation is raised separately as C-1.'
          round: 2
        - id: BR-4
          disposition: addressed
          note: 'Verified by reverting to `config --get`: TestTrunkFile_SigningHonorsNonCanonicalBool goes red.'
          round: 2
        - id: BR-5
          disposition: addressed
          note: TestTrunkFile_UpdateCreatesAbsentPath covers the nested absent path and also asserts the sibling file survives read-tree.
          round: 2
        - id: BR-6
          disposition: addressed
          note: tempIndexPath() now takes no parameter.
          round: 2
        - id: BR-7
          disposition: addressed
          note: MkdirTemp 0700 + filepath.Join, pinned by TestTempIndexPath_NoUnclaimedNameWindow including the mode check.
          round: 2
        - id: BR-8
          disposition: addressed
          note: gitx.FirstLine exported with a table test; issueids.go's firstLine is a delegating alias.
          round: 2
        - id: BR-9
          disposition: addressed
          note: trunkfile_test.go:203-205 corrected. Two other sites carrying the same stale claim are raised as a family repeat, not as this finding.
          round: 2
        - id: BR-10
          disposition: addressed
          note: 'Better than asked: isNonFastForward and isMissingPath were eliminated rather than table-tested, and FirstLine got TestFirstLine.'
          round: 2
        - id: BR-11
          disposition: not-addressed
          note: 'trunkfile.go:150 unchanged. Extending the finding to its class: Update:288 also reports "the trunk moved under 3 attempts" for any commitAndPush failure (read-tree/hash-object/write-tree/commit-tree/temp-file) that coincides with a peer push. Rule: an error message may only name a cause the code observed.'
          round: 2
        - id: BR-12
          disposition: not-addressed
          note: trunkfile.go:320-322 unchanged; no sentence saying .gitattributes is resolved from the checked-out tree rather than from the trunk being written.
          round: 2
        - id: BR-13
          disposition: addressed
          note: Revisions section added, Tasks 1-6b ticked, ReadDegraded added to the TrunkFile entry, seam Resolution corrected, commit-granularity divergence recorded. Three unnamed CombinedOutput restatements survive and are raised as a family repeat.
          round: 2
      findings:
        - id: BR-14
          severity: Critical
          title: exists() reads "object unavailable" as "path absent", so Update silently wipes the trunk file and reports success
          detail: |-
            3rd finding in this family — fix the RULE, not the instance. trunkfile.go:202
            uses `cat-file -e <ref>:<path>`, whose non-zero exit means both "not in the
            tree" and "object cannot be produced". Reproduced on git 2.50.1: with the
            loose blob removed, `cat-file -e main:queue.md` exits 1 while `ls-tree
            --name-only` exits 0 and prints the path; on a --filter=blob:none clone with
            an unreachable promisor it exits 128 with the path still in the tree. Driven
            end-to-end through the real type with origin reachable, Read returned
            ("", nil) and Update published a commit replacing "important\ncontent\n"
            with "- mine\n", returning nil. The doc comment at trunkfile.go:198-201
            claims this exact failure is prevented. Rule for the whole family
            a classifier may only assert an outcome its signal distinguishes; the
            ambiguous case must surface as an error, never as a value. Enumeration
            exists (conflates, fix with ls-tree or an exit-code split), refExists (ok),
            Update's post-push check (ok but scoped to all of commitAndPush),
            offlineError (asserts an unobserved cause, BR-11), signs (ok).
          family: git-output-string-matching
          round: 2
        - id: BR-15
          severity: Minor
          title: Two more "combined output" claims survive in trunkfile.go, one contradicting itself four lines later
          detail: |-
            2nd finding in this family — state the rule rather than patching the site.
            Rule when a decision is reversed, sweep every restatement of it; the
            enumeration is a grep for the old term. `grep -n "combined\|Combined"` over
            trunkfile.go returns :37 ("returning combined output", contradicted by :39)
            and :302 ("Returns git's combined output", where commitAndPush actually
            returns stderr only). Round 1 fixed the test-comment site; these two are the
            same claim in the file a reader consults for the rationale.
          family: stale-comment
          round: 2
        - id: BR-16
          severity: Minor
          title: The plan still instructs CombinedOutput at three sites after the Revisions entry corrected one
          detail: |-
            2nd finding in this family — same rule as above, same grep. Plan lines 86
            (TrunkFile's `Seam:` bullet), 136 (Task 1 Step 4) and 207 (Task 4 Step 5)
            still specify CombinedOutput; only line 64's Resolution paragraph was
            corrected. As written the plan tells a future implementor to build the shim
            whose CRLF warning landed inside a parsed blob hash.
          family: plan-traceability
          round: 2
        - id: BR-17
          severity: Minor
          title: TestTrunkFile_SigningHonorsNonCanonicalBool and TestTrunkFile_SigningRepoGetsSignedCommit are near-identical 25-line bodies
          detail: |-
            2nd finding in this family — the rule is one definition per behavior,
            including in tests. trunkfile_test.go:321-350 and :430-466 differ only in the
            config value; both stub runGitIn, strip -S, and assert sawDashS. One table
            over {"true","yes","1","0","false"} collapses them and would also cover the
            negative case, which nothing tests today.
          family: duplicated-helper
          round: 2
        - id: BR-18
          severity: Minor
          title: The contended Update path now costs 6 fetches instead of 3
          detail: |-
            trunkfile.go:279 fetches to decide retryability, discards the result, and the
            next iteration's Read (:255) immediately fetches again. Reuse the post-push
            fetch for the retry's base read (ARCH-CONSTRAINTS repeated expensive work).
            Also hoist signs() out of commitAndPush so it is not re-shelled per attempt.
          family: repeated-external-call
          round: 2
      boundary: M1
      blocked: true
---

# Gate ledger — ariadne#209 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-09-07T17:17:20-07:00 (claude) — BLOCKED

### Raised

- **BR-1** [Important] `vacuous-test-guard` Temp-index assertions in TestTrunkFile_TransformErrorLeavesNoIndexAndNoCommit are unreachable
  trunkfile_test.go:258-265 nests the os.Stat removal check and the
  filepath.IsAbs check under `if seen != ""`, but Update calls the transform
  before commitAndPush, and GIT_INDEX_FILE is only set at trunkfile.go:271
  inside commitAndPush — so the spy never sees it and the block never runs.
  The Done-when clause "the temp index is removed on every exit path,
  including the error paths, and is never $GIT_DIR/index" has no coverage;
  neither does the blob temp file. Assert on a path that reaches
  commitAndPush and t.Fatal when seen is empty so it cannot vacate silently.
- **BR-2** [Important] `git-output-string-matching` isNonFastForward matches bare "rejected", retrying non-retryable push refusals
  trunkfile.go:260. Reproduced: a pre-receive hook declining a push emits
  "! [remote rejected] main -> main (pre-receive hook declined)", which
  substring-matches "rejected". A policy, permission or protected-branch
  refusal is then retried 3 times and reported as "the trunk moved under 3
  attempts" — the wrong next action. "non-fast-forward" and "fetch first"
  already cover the real CAS case. Add a pure table test over the three real
  git outputs.
- **BR-3** [Important] `git-output-string-matching` isMissingPath carries a phrase git no longer emits, and the no-remote fallback silently depends on it
  trunkfile.go:185 matches "Not a valid object name"; git 2.50.1 says
  "invalid object name". ReadDegraded's local-branch fallback
  (trunkfile.go:128-134) fires only because that match FAILS, so the
  documented "a repo with no origin reads the local ref" behavior rests on a
  negative match against unpinned prose. If a git version emits the matched
  wording, ReadDegraded returns empty bytes with a warning claiming it read
  the stale ref — a fabricated value (ARCH-SECURE). TestTrunkFile_NoRemoteIsNamed
  asserts only the warning, never the bytes. Sweep the class: pin all three
  classifiers (isMissingPath, isNonFastForward, offlineError) against real git
  output, and prefer rev-parse --verify over prose for ref existence.
- **BR-4** [Important] `git-config-bool` signs() compares the raw config string to "true", so commit.gpgsign = 1|yes|on yields unsigned commits
  trunkfile.go:319-320. Verified: `git config --get commit.gpgsign` returns
  "1"/"yes" verbatim while `--type=bool` normalizes all of git's truthy forms
  to "true". The function's own comment calls silently-unsigned commits a
  regression, and it is reachable today; the test only exercises the literal
  "true". Use --type=bool and add a "1" row. Compounds with the isNonFastForward
  finding: on a signature-requiring trunk the refused push is misreported as a
  CAS collision.
- **BR-5** [Important] `untested-first-write` No test for Update on a path absent from the trunk — the exact path M2's seed takes
  Every Update test seeds queue.md via trunkFixture first, but
  workshop/queue.md does not exist on origin/main, so Task 13's seed is a
  create. I replayed the plumbing by hand against a throwaway bare origin and
  it works, so this is a coverage gap rather than a bug — but it is the only
  path with no regression guard, and it is the next one to run in anger. One
  test on a nested, absent path closes it.
- **BR-6** [Minor] `dead-parameter` tempIndexPath(dir string) never uses dir
  trunkfile.go:194 — the parameter is vestigial and the doc comment implies it
  matters. Drop it or use it.
- **BR-7** [Minor] `temp-file-race` os.CreateTemp then os.Remove leaves a symlink-plant window on the index path
  trunkfile.go:195-201 and 324-325. os.MkdirTemp plus filepath.Join(dir,
  "index") removes the window (ARCH-SECURE; low exposure on a single-user
  machine, but free to fix).
- **BR-8** [Minor] `duplicated-helper` firstLineOf duplicates firstLine in cmd/sdlc/issueids.go:112
  trunkfile.go:146. main already imports gitx, so exporting one and deleting
  the other is free (ARCH-DRY).
- **BR-9** [Minor] `stale-comment` Test comment claims runGitIn uses CombinedOutput, which it deliberately does not
  trunkfile_test.go:202-203. The separation is the correct design and is
  explained at trunkfile.go:34-52; the comment contradicts it in the one place
  a reader checks the rationale.
- **BR-10** [Minor] `pure-logic-untested` isNonFastForward, isMissingPath and firstLineOf are pure but only exercised through integration tests
  Direct table tests over these three would have surfaced both
  git-output-string-matching findings, and run in microseconds instead of
  seconds (ARCH-PURE).
- **BR-11** [Minor] `error-misattribution` offlineError labels every fetch failure "unreachable (offline?)"
  trunkfile.go:104-109 and 140-142. A branch that does not exist on origin, or
  an auth failure, reads to the operator as a network outage.
- **BR-12** [Minor] `gitattributes-source` hash-object --path resolves .gitattributes from the working tree, not from the trunk being written
  trunkfile.go:281-286. Correct in practice, but "no working tree involved" is
  the type's headline claim and this is the one place it is not strictly true.
  One sentence in the doc comment.
- **BR-13** [Minor] `plan-traceability` Plan Tasks 1-6b left unchecked and the plan has no Revisions section despite mid-stream edits
  workshop/plans/000209-queue-datatype-plan.md — Tasks 6b and 13b were appended
  in this window with no revision note, the seam table still specifies a
  CombinedOutput two-value runGitIn that the code deliberately does not
  implement, and TrunkFile's entry omits the exported ReadDegraded. Commit
  granularity also diverges (Tasks 2+4 merged, 3+5 merged).

## Round 2 — 2026-09-07T17:36:16-07:00 (claude) — BLOCKED

### Disposed

- BR-1 — addressed — Verified by reverting: moving cleanupIdx() off the defer turns the test red. Blob temp-file cleanup remains unpinned (correct code, no test) — noted in coverage, not re-raised.
- BR-2 — addressed — Verified by restoring the prose match: TestTrunkFile_NonRetryablePushFailsFast goes red with "transform called 3 times, want 1".
- BR-3 — not-addressed — Prose matching is gone, but the byte assertion the finding asked for is still absent: I hardwired refExists to true and the entire suite stayed green, so the fabricated-empty-value path remains unpinned. TestTrunkFile_NoRemoteIsNamed:509 still discards the bytes; testfix.InitialCommit writes README="x\n", so this is a one-line assert. The replacement signal's own conflation is raised separately as C-1.
- BR-4 — addressed — Verified by reverting to `config --get`: TestTrunkFile_SigningHonorsNonCanonicalBool goes red.
- BR-5 — addressed — TestTrunkFile_UpdateCreatesAbsentPath covers the nested absent path and also asserts the sibling file survives read-tree.
- BR-6 — addressed — tempIndexPath() now takes no parameter.
- BR-7 — addressed — MkdirTemp 0700 + filepath.Join, pinned by TestTempIndexPath_NoUnclaimedNameWindow including the mode check.
- BR-8 — addressed — gitx.FirstLine exported with a table test; issueids.go's firstLine is a delegating alias.
- BR-9 — addressed — trunkfile_test.go:203-205 corrected. Two other sites carrying the same stale claim are raised as a family repeat, not as this finding.
- BR-10 — addressed — Better than asked: isNonFastForward and isMissingPath were eliminated rather than table-tested, and FirstLine got TestFirstLine.
- BR-11 — not-addressed — trunkfile.go:150 unchanged. Extending the finding to its class: Update:288 also reports "the trunk moved under 3 attempts" for any commitAndPush failure (read-tree/hash-object/write-tree/commit-tree/temp-file) that coincides with a peer push. Rule: an error message may only name a cause the code observed.
- BR-12 — not-addressed — trunkfile.go:320-322 unchanged; no sentence saying .gitattributes is resolved from the checked-out tree rather than from the trunk being written.
- BR-13 — addressed — Revisions section added, Tasks 1-6b ticked, ReadDegraded added to the TrunkFile entry, seam Resolution corrected, commit-granularity divergence recorded. Three unnamed CombinedOutput restatements survive and are raised as a family repeat.

### Raised

- **BR-14** [Critical] `git-output-string-matching` exists() reads "object unavailable" as "path absent", so Update silently wipes the trunk file and reports success
  3rd finding in this family — fix the RULE, not the instance. trunkfile.go:202
  uses `cat-file -e <ref>:<path>`, whose non-zero exit means both "not in the
  tree" and "object cannot be produced". Reproduced on git 2.50.1: with the
  loose blob removed, `cat-file -e main:queue.md` exits 1 while `ls-tree
  --name-only` exits 0 and prints the path; on a --filter=blob:none clone with
  an unreachable promisor it exits 128 with the path still in the tree. Driven
  end-to-end through the real type with origin reachable, Read returned
  ("", nil) and Update published a commit replacing "important\ncontent\n"
  with "- mine\n", returning nil. The doc comment at trunkfile.go:198-201
  claims this exact failure is prevented. Rule for the whole family
  a classifier may only assert an outcome its signal distinguishes; the
  ambiguous case must surface as an error, never as a value. Enumeration
  exists (conflates, fix with ls-tree or an exit-code split), refExists (ok),
  Update's post-push check (ok but scoped to all of commitAndPush),
  offlineError (asserts an unobserved cause, BR-11), signs (ok).
- **BR-15** [Minor] `stale-comment` Two more "combined output" claims survive in trunkfile.go, one contradicting itself four lines later
  2nd finding in this family — state the rule rather than patching the site.
  Rule when a decision is reversed, sweep every restatement of it; the
  enumeration is a grep for the old term. `grep -n "combined\|Combined"` over
  trunkfile.go returns :37 ("returning combined output", contradicted by :39)
  and :302 ("Returns git's combined output", where commitAndPush actually
  returns stderr only). Round 1 fixed the test-comment site; these two are the
  same claim in the file a reader consults for the rationale.
- **BR-16** [Minor] `plan-traceability` The plan still instructs CombinedOutput at three sites after the Revisions entry corrected one
  2nd finding in this family — same rule as above, same grep. Plan lines 86
  (TrunkFile's `Seam:` bullet), 136 (Task 1 Step 4) and 207 (Task 4 Step 5)
  still specify CombinedOutput; only line 64's Resolution paragraph was
  corrected. As written the plan tells a future implementor to build the shim
  whose CRLF warning landed inside a parsed blob hash.
- **BR-17** [Minor] `duplicated-helper` TestTrunkFile_SigningHonorsNonCanonicalBool and TestTrunkFile_SigningRepoGetsSignedCommit are near-identical 25-line bodies
  2nd finding in this family — the rule is one definition per behavior,
  including in tests. trunkfile_test.go:321-350 and :430-466 differ only in the
  config value; both stub runGitIn, strip -S, and assert sawDashS. One table
  over {"true","yes","1","0","false"} collapses them and would also cover the
  negative case, which nothing tests today.
- **BR-18** [Minor] `repeated-external-call` The contended Update path now costs 6 fetches instead of 3
  trunkfile.go:279 fetches to decide retryability, discards the result, and the
  next iteration's Read (:255) immediately fetches again. Reuse the post-push
  fetch for the retry's base read (ARCH-CONSTRAINTS repeated expensive work).
  Also hoist signs() out of commitAndPush so it is not re-shelled per attempt.

## Open findings

- **BR-3** [Important] `git-output-string-matching` isMissingPath carries a phrase git no longer emits, and the no-remote fallback silently depends on it
- **BR-11** [Minor] `error-misattribution` offlineError labels every fetch failure "unreachable (offline?)"
- **BR-12** [Minor] `gitattributes-source` hash-object --path resolves .gitattributes from the working tree, not from the trunk being written
- **BR-14** [Critical] `git-output-string-matching` exists() reads "object unavailable" as "path absent", so Update silently wipes the trunk file and reports success
- **BR-15** [Minor] `stale-comment` Two more "combined output" claims survive in trunkfile.go, one contradicting itself four lines later
- **BR-16** [Minor] `plan-traceability` The plan still instructs CombinedOutput at three sites after the Revisions entry corrected one
- **BR-17** [Minor] `duplicated-helper` TestTrunkFile_SigningHonorsNonCanonicalBool and TestTrunkFile_SigningRepoGetsSignedCommit are near-identical 25-line bodies
- **BR-18** [Minor] `repeated-external-call` The contended Update path now costs 6 fetches instead of 3
