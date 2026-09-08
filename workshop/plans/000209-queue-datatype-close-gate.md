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
    - "n": 3
      timestamp: "2026-09-07T17:51:06-07:00"
      agent: claude
      dispose:
        - id: BR-3
          disposition: not-addressed
          note: 'Classifier half landed and is now structurally pinned (hardwiring refExists to true reddens TestTrunkFile_NoRemoteIsNamed), but the byte assertion is still missing: hardwiring readLocal to return nil,nil leaves the whole package green. That gap is what let I-1 ship.'
          round: 3
        - id: BR-11
          disposition: not-addressed
          note: trunkfile.go:150 and ReadDegraded:127-130 unchanged; Update:290 still reports "the trunk moved" for any commitAndPush failure coinciding with a peer push.
          round: 3
        - id: BR-12
          disposition: not-addressed
          note: trunkfile.go:350-352 unchanged; still no sentence saying .gitattributes resolves from the checked-out tree rather than from the trunk being written.
          round: 3
        - id: BR-14
          disposition: addressed
          note: 'Verified twice: collapsing pathPresent to two states reddens TestTrunkFile_UnreadablePathRefusesRatherThanTruncating, and a stub-free end-to-end against a real deleted loose blob gives Read->err, transform not called, trunk intact.'
          round: 3
        - id: BR-15
          disposition: not-addressed
          note: grep -n "combined\|Combined" trunkfile.go still returns :37 and :332, both unchanged.
          round: 3
        - id: BR-16
          disposition: not-addressed
          note: Plan :86, :136, :207 unchanged. Same enumeration also shows gitx.FirstLine, a newly exported symbol, missing from the plan's Core-concepts surface.
          round: 3
        - id: BR-17
          disposition: addressed
          note: One table over true/yes/1/on/false/0/off; reverting --type=bool reddens the yes, 1 and on rows. The negative cases are now covered, which the two old bodies never were.
          round: 3
        - id: BR-18
          disposition: addressed
          note: Fetch half pinned — reverting the attempt==1 hoist reddens TestTrunkFile_OneFetchPerAttempt with "3 fetches for 2 attempts". The signs() hoist has no test; behavior-neutral, and a call-count assertion would only restate the implementation.
          round: 3
      findings:
        - id: BR-19
          severity: Important
          title: readLocal treats a legitimately absent local branch as unreadable, so ReadDegraded errors obscurely in a fresh repo while warning that it used the local branch
          detail: |-
            4th finding in this family — complete the RULE, do not patch the site.
            BR-14 stated half of it (an ambiguous signal must error). The missing half
            is that an unambiguous BENIGN signal must not be swept into that error.
            pathPresent (trunkfile.go:220) maps every non-zero ls-tree exit onto
            "cannot tell", which is right for an unreadable tree and wrong for a ref
            that does not resolve — a state readLocal reaches by design. Reproduced
            against the real type: `git init -b main` with no commits, and a repo whose
            local default branch is `master`, both give
            err="ls-tree main -- <path>: exit status 128 / fatal: Not a valid object
            name main" WITH warn ending "used local main", asserting something that did
            not happen. This is a regression introduced in this window; at 850eabd
            exists() swallowed it and the call returned (nil, warn, nil). It violates
            the Done-when clause "a repo with no origin remote at all reads the local
            ref and says so … rather than erroring obscurely". Re-run enumeration under
            the completed rule readRef's tracking ref cannot legitimately be absent
            (fetch precedes it, ReadDegraded guards with refExists) so its failures are
            genuine; readLocal's branch can, and must be guarded with the same
            refExists signal used one line above at trunkfile.go:132.
          family: git-output-string-matching
          round: 3
        - id: BR-20
          severity: Minor
          title: '`_ = errOut` at trunkfile.go:200 is a no-op left behind by the three-states edit'
          detail: |-
            2nd finding in this family — state the rule rather than deleting the line.
            Rule no statement or parameter may exist solely to reference something;
            errOut is already consumed by the error branch at :198, so the assignment
            does nothing and go vet will not catch it. The enumeration is
            `grep -n '^\s*_ = ' cmd/sdlc/internal/gitx/`, which returns exactly this one
            site — measured prevalence 1, so the sweep is cheap and complete.
          family: dead-parameter
          round: 3
      boundary: M1
      blocked: true
    - "n": 4
      timestamp: "2026-09-07T18:03:54-07:00"
      agent: claude
      dispose:
        - id: BR-3
          disposition: addressed
          note: isMissingPath/isNonFastForward are gone from the tree (grep over cmd/ returns nothing); classification is now exit-code based via gitExitCode/gitAbsentExit and pinned by TestGitExitCode_NonExitFailureIsNotAbsent plus TestTrunkFile_RefPresentPropagatesNonAbsentFailure, which I mutation-verified goes red when the error branch is collapsed.
          round: 4
        - id: BR-11
          disposition: not-addressed
          note: 'Unchanged, and now upgraded from wording to ARCH-SECURE. RULE a message must describe the state the code observed, not the state it hoped for. Enumeration, 3 sites, all reproduced against the real type: (1) offlineError trunkfile.go:192 and (2) ReadDegraded''s warn trunkfile.go:158 both say "unreachable (offline?)" for a reachable remote missing the branch — measured "origin unreachable (offline?): fatal: couldn''t find remote ref refs/heads/main"; (3) trunkfile.go:170 appends "used local main" even when no local main exists, so a repo whose default branch is master yields bytes="" warn="…used local main" err=nil — a fabricated empty value carrying a claim that did not happen. Fix by branching on the observed signal already in hand (fetch''s exit/stderr class; the refPresent result readLocal computes).'
          round: 4
        - id: BR-12
          disposition: not-addressed
          note: Unchanged. trunkfile.go:404-406 still explains what --path buys with no sentence noting the attributes come from the working tree/index, not from the trunk being written.
          round: 4
        - id: BR-15
          disposition: not-addressed
          note: Unchanged. grep -n "combined\|Combined" trunkfile.go still returns :37 ("returning combined output", contradicted by :39) and :386 ("Returns git's combined output", where commitAndPush returns stderr only).
          round: 4
        - id: BR-16
          disposition: not-addressed
          note: Unchanged, and the enumeration is larger than the three lines named. Beyond plan lines 86, 136 and 207, the same rule catches line 7 ("a sibling runEnv"; the code adds one runGitIn and leaves run alone), the Revisions entry's own "replaced isMissingPath's phrase matching with cat-file -e's exit code" (cat-file -e was itself rejected as two-state; the code uses ls-tree), and line 171's checked-off promise of a real signed-commit test with t.Skip, which the delivered test does not do.
          round: 4
        - id: BR-19
          disposition: addressed
          note: 'Verified by mutation, not by the diff: deleting the refPresent(t.branch) guard at trunkfile.go:217-223 turns TestTrunkFile_FreshRepoReadsEmptyNotError red with "ls-tree main -- workshop/queue.md: exit status 128 / fatal: Not a valid object name main". The residual false "used local main" warning is folded into BR-11.'
          round: 4
        - id: BR-20
          disposition: addressed
          note: grep -n '^\s*_ = ' over cmd/sdlc/internal/gitx/ now returns only the test's idiomatic _ = cmd.Run(); the no-op assignment is gone.
          round: 4
      findings:
        - id: BR-21
          severity: Important
          title: readLocal and readRef are the same three-line body with a different ref, and the guard exists on only one of them
          detail: |-
            This is the 3rd finding in family `duplicated-helper`. Earlier rounds fixed
            instances (firstLineOf/firstLine, the two signing tests). Do NOT merge just
            these two — state the rule. RULE two functions differing only in the value
            of a parameter are one function; when they diverge, the divergence is a bug
            in the one that lacks it, not a feature of the one that has it.
            Enumeration by that rule over cmd/sdlc/internal/gitx/: 1 remaining site.
            trunkfile.go:224-235 (readLocal) and :241-252 (readRef) both run
            pathPresent(ref, path) then cat-file blob ref:path, differing only in
            t.branch vs t.trackingRef() and in the error string. readLocal additionally
            guards ref existence via refPresent; readRef does not — which is exactly the
            BR-19 defect class, currently unreachable only because ReadDegraded and
            Update happen to check the tracking ref before calling it. Collapse to
            readAt(ref, path) carrying the guard, with readLocal/readRef as one-line
            callers. Doing so also fixes a second thing for free: readLocal passes the
            UNQUALIFIED branch name to refPresent, pathPresent and cat-file, so a tag
            named `main` shadows refs/heads/main under rev-parse's DWIM order; readAt
            should take a fully-qualified ref and readLocal should pass
            "refs/heads/" + t.branch (ARCH-DRY).
          family: duplicated-helper
          round: 4
        - id: BR-22
          severity: Minor
          title: Update rebuilds the tree entry from scratch, so an executable path on the trunk comes back mode 100644
          detail: |-
            trunkfile.go:413-414 hardcodes `--cacheinfo 100644,<blob>,<path>`. Verified
            against the real type: a 100755 blob on the trunk is 100644 after an
            Update that only changed content. RULE when replacing one field of an
            existing record, carry the fields you were not asked to change. The mode is
            already on the wire — pathPresent (trunkfile.go:272) runs ls-tree and
            discards everything but the name; dropping --name-only there yields the
            existing mode for free. Alternatively state "regular file, mode 100644" as
            the type's contract in the doc comment and the atlas entry. Harmless for
            M1's .md consumers, but ariadne#207 inherits this primitive and no test
            would notice.
          family: unspecified-fields-not-preserved
          round: 4
      boundary: M1
      blocked: false
    - "n": 5
      timestamp: "2026-09-07T18:52:12-07:00"
      agent: claude
      dispose:
        - id: BR-11
          disposition: not-addressed
          note: 'Unchanged at trunkfile.go:192 and :158, and the enumeration is now larger — M2 added two sites. (4) queue.go:212 prints "the queue on the trunk right now (N entries)" from a ReadDegraded whose warning it discards; reproduced offline, it printed "(0 entries)" for a trunk that has four, i.e. a fabricated state presented as observed. (5) Update:317 calls offlineError unconditionally, so a repo with NO origin gets "origin unreachable (offline?)" although ReadDegraded has a dedicated branch for exactly that case. (6) intent.go:96 reports "kept the newer why-now" for an edit that also dropped the tag and flipped the kind (see the new I-3). Same rule: a message may only name state the code observed.'
          round: 5
        - id: BR-12
          disposition: not-addressed
          note: Unchanged. trunkfile.go:383-385 still explains what --path buys with no sentence noting the attributes are resolved from the working tree/index, not from the trunk being written.
          round: 5
        - id: BR-15
          disposition: addressed
          note: 'Verified by grep: `grep -in "combined" trunkfile.go` now returns only :39 ("returned SEPARATELY, not combined"), and commitAndPush:365 reads "Returns git''s STDERR".'
          round: 5
        - id: BR-16
          disposition: addressed
          note: 'Verified by grep over the plan: no CombinedOutput at lines 86/136/207 or anywhere outside the Revisions entry describing the history. Residual plan-vs-code staleness (line 7''s "sibling runEnv" vs the delivered runGitIn; line 171''s t.Skip signing test vs the delivered -S interception; the Revisions entry''s "cat-file -e" vs the delivered ls-tree) is carried forward under the new plan-traceability finding rather than re-raised here.'
          round: 5
        - id: BR-21
          disposition: not-addressed
          note: 'The main defect IS fixed and properly pinned — readLocal/readRef collapsed into readFrom carrying the guard; verified by mutation (deleting the refPresent block turns both ReadFromRefusesUndeterminableRef and FreshRepoReadsEmptyNotError red). The only residue is the half the finding asked for and the diff did not do: trunkfile.go:169 still passes the UNQUALIFIED t.branch. Reproduced against the real type — a repo with a tag named "main" and no tracking ref returns the TAG''s bytes while warning "used local main". One-line fix: pass "refs/heads/" + t.branch.'
          round: 5
        - id: BR-22
          disposition: addressed
          note: 'modeOf + TestTrunkFile_PreservesFileMode. Verified by mutation, not by the diff: replacing the modeOf call with a hardcoded "100644" fails the test with "mode = [100644 ... run.sh], want 100755 preserved". The new-path 100644 case is covered in the same test. (Note M-1: the fix added a SECOND ls-tree rather than widening the one pathPresent already runs, which is what the finding sketched.)'
          round: 5
      findings:
        - id: BR-23
          severity: Important
          title: --tag is the one user-supplied field that lands in the line format with no validator; a newline in it publishes a forged file with exit 0
          detail: 'Reproduced against the real code path. Intent{Op: OpAdd, Ref: "b#2", WhyNow: "why", Tag: "x\n- forged#9 — injected [y"} returns nil and publishes three entries where one was intended, mangling the real entry to "- b#2 — why [x" and forging "- forged#9 — injected [y]". ValidateWhyNow exists to stop exactly this forge (line_test.go:96 demonstrates it) and ValidateRef covers the third field; Tag was simply not in the enumeration. RULE every user-supplied field that lands in the line format is validated by the same rule. Add ValidateTag (newline/control/sep/bracket, plus non-blank-after-trim, since a whitespace-only tag is silently absorbed into why-now) and call it from applyAdd beside the other two.'
          family: unvalidated-format-field
          round: 5
        - id: BR-24
          severity: Important
          title: Validation runs inside the transform, so "rejected before any git call" is false and the offline error masks it
          detail: 'Path is RunE -> openTrunkStore -> gitx.RepoTopLevel (git) -> Update -> signs (git config) -> fetch (network) -> transform -> ValidateRef. Reproduced with an unreachable origin: `queue add "bad ref" x` prints "origin unreachable (offline?)" and a fabricated "(0 entries)" listing; the operator never learns the ref was malformed. The Spec''s Done-when, the Spec''s "Input validation, before any git call" paragraph and the plan''s Task 9 table all promise the opposite. queue_test.go:168 claims to pin this but only asserts the fake''s content is unchanged, so it cannot observe a git call and passes today. Fix: an Intent.Validate() called in runQueueEdit before openTrunkStore, keeping the in-Apply calls as the belt, plus a call-counting seam in the test.'
          family: validation-behind-io
          round: 5
        - id: BR-25
          severity: Important
          title: Converge replaces the whole line, silently flipping kind and dropping the tag, while the note claims only the why-now moved
          detail: 'This is the 2nd finding in family unspecified-fields-not-preserved. Do NOT fix only this site — state the rule. RULE when replacing one field of an existing record, carry every field you were not asked to change, and a note may only claim the change it actually made. BR-22 stated it for git tree modes; the record here is {Ref, WhyNow, Tag, Kind, cr} and intent.go:90-98 rebuilds it from the Intent alone. Reproduced: `add sdlc-fleet "sharper reason"` onto "- project:sdlc-fleet — the whole area is next [sdlc]" yields "- sdlc-fleet — sharper reason" with the note "kept the newer why-now (was ...)" — project marker and tag gone, unmentioned. Second cell: converging onto a CRLF line yields "- a#1 — one prime\n- b#2 — two\r\n", mixing line endings. This is the interleaving table''s concurrent-add cell, so a peer''s --project/--tag vanishes on the path built to make peers survive; the table test asserts only refs plus a note substring and cannot catch it.'
          family: unspecified-fields-not-preserved
          round: 5
        - id: BR-26
          severity: Important
          title: The line format is stated in no user-facing artifact, and a near-miss hand-edit disappears from the listing without a word
          detail: 'The Done-when requires the datatype prose to state "the line format". construct/datatype/queue.md does not, and routes the operational contract to `sdlc queue --help`, which does not carry it either. The grammar — "- <ref> — <why> [tag]", the em-dash separator, the `project:` prefix that is the ONLY marker of a project line, the why-now restrictions — lives only in Go comments. That matters because the design explicitly invites hand-editing (line.go:45-48, the fuzz target, and the comment that an unparsed line is one `queue remove` cannot find). Compounding: runQueueList prints Entries() only, so a hand-edited "- a#1 - why" (ASCII hyphen) is silently absent from the listing while sitting in the file. Two cheap fixes: a "Line format" section in the datatype prose and the helptext, and an stderr line in runQueueList reporting len(lines)-len(entries) unparsed lines.'
          family: format-undocumented
          round: 5
        - id: BR-27
          severity: Important
          title: A new trunk-writing verb shipped into the base-layer binary with no guardSpineRepo decision recorded either way
          detail: repoguard.go:5-14 states the guard is wired into the lifecycle verbs and that reads stay unguarded by construction. `queue add/remove/move` are writes that push to origin/main, and they are unguarded — in a brain repo that is a CAS push at a gcrypt::ssh remote, and in a repo without workshop/issues/ it creates workshop/queue.md on main. This issue's own Spec calls the brain-queue question "an open disagreement, not a settled no", so shipping the unguarded path quietly settles it; ariadne is the base layer, so the verb reaches every downstream repo. migrate.go:460 is the established shape for the other answer — an explicit comment saying why the exemption holds. Either wire guardSpineRepo into the three mutating subcommands or record the exemption the way migrate does.
          family: unguarded-write-verb
          round: 5
        - id: BR-28
          severity: Important
          title: 'Task 13 Step 3 is ticked but no note exists on either #207 file, and three more Chunk-2 rows claim artifacts the tree does not have'
          detail: 'This is the 3rd finding in family plan-traceability. Do NOT fix only this site — state the rule. RULE a "- [x]" is a claim about the tree, not about intent; before a close, sweep every checked row in the closing chunk against the artifact it names. Enumeration over Chunk 2, 4 rows fail: (a) Task 13 Step 3 — grep -i "trunkfile|#209" over both workshop/issues/000207-*.md returns nothing, so #207 still has no record that gitx.TrunkFile exists nor of the retry-semantics defect the Spec promised to flag there (say which of the two #207 files gets it); (b) Task 11 Step 5''s commit does not exist, folded into c93be7c; (c) Task 13 Step 4''s commit does not exist — the seed landed as `queue: add ...` commits on origin/main; (d) 04804ff (the real-git e2e) is real work with no plan row. M1 recorded exactly this divergence in Revisions; M2''s chunk has no such entry.'
          family: plan-traceability
          round: 5
        - id: BR-29
          severity: Minor
          title: pathPresent/modeOf and refPresent/resolve are each one function split in two, and Update runs both halves of both pairs per attempt
          detail: 'This is the 4th finding in family duplicated-helper. Do NOT merge only these — state the rule, which the code already states at trunkfile.go:230-235: two functions differing only in a parameter''s value, or in which field of one result they keep, are one function. Enumeration over cmd/sdlc/internal/gitx/, 2 remaining sites. trunkfile.go:219 (pathPresent) and :426 (modeOf) run the same ls-tree on the same (ref, path), differing only by --name-only — and BR-22''s own fix sketch said to drop --name-only and take the mode from the call already on the wire, so the fix for BR-22 created this instance. trunkfile.go:178 (refPresent) and :356 (resolve) run the same rev-parse --verify, differing only by --quiet and return type; Update calls both on trackingRef each attempt. Collapse each pair to one call returning the richer result.'
          family: duplicated-helper
          round: 5
        - id: BR-30
          severity: Minor
          title: fakeTrunk calls the transform exactly once and fires its peer before it, so two tests are named for properties the double cannot produce
          detail: 'This is the 2nd finding in family vacuous-test-guard. State the rule rather than patching a site: a test may not be named for a property its double cannot produce. queue_test.go:37-51 — Update calls transform once and runs `peer` BEFORE it, so TestQueueEdit_PeerEditSurvivesTheReplay ("the property the whole design exists for") would pass against an Update with no retry at all, and TestQueueEdit_ValidationRefusesBeforeTouchingTheTrunk cannot see a git call. Consequence: no test anywhere drives Intent.Apply through a real non-fast-forward rejection — gitx''s two-publisher test uses a raw append transform and the e2e has no peer. Give fakeTrunk a rejection mode that re-invokes the transform, and add a peer push to queue_e2e_test.go.'
          family: vacuous-test-guard
          round: 5
        - id: BR-31
          severity: Minor
          title: Core-concepts table names queueCmd where the symbol is NewQueueCmd, and gitx.FirstLine has no row
          detail: Same family as above; listed separately only because it is a table edit rather than a checkbox. FirstLine was newly exported in this window (moved out of issueids.go) and is part of the package's surface.
          family: plan-traceability
          round: 5
        - id: BR-32
          severity: Minor
          title: queueRefusal discards ReadDegraded's staleness warning and issues a second full fetch on the error path
          detail: queue.go:212 binds the warning to _ , so the handoff prints "the queue on the trunk right now" with no indication the read was degraded. Folded into the BR-11 enumeration above; noted here for the file:line.
          family: error-misattribution
          round: 5
      blocked: true
    - "n": 6
      timestamp: "2026-09-07T19:16:10-07:00"
      agent: claude
      dispose:
        - id: BR-11
          disposition: not-addressed
          note: trunkfile.go:191-193 and Update:317 are untouched by this round (trunkfile.go is not in 66065f6). queueRefusal now prints the staleness warning, which was BR-32's half, not this one.
          round: 6
        - id: BR-12
          disposition: not-addressed
          note: Untouched. trunkfile.go:383-385 still explains what --path buys with no sentence noting the attributes resolve from the working tree/index, not from the trunk being written.
          round: 6
        - id: BR-21
          disposition: not-addressed
          note: 'The fix commit dismisses this as "stale" on the readFrom merge, but the finding''s second half is live and reproduced. trunkfile.go:169 still passes the UNQUALIFIED t.branch; I confirmed against real git that with both refs/tags/main and refs/heads/main present, `rev-parse --verify main` and `cat-file blob main:q.md` return the TAG. So ReadDegraded''s no-tracking-ref fallback can serve a tag''s bytes while warning "used local main". One-line fix, unchanged: pass "refs/heads/" + t.branch.'
          round: 6
        - id: BR-23
          disposition: addressed
          note: 'ValidateTag exists and is pinned twice. Verified by mutation, not by the diff: short-circuiting ValidateTag to nil reddens TestIntent_Validate_CoversEveryInterpolatedField/tag and two sub-cases of TestQueueEdit_ValidationRefusesBeforeTouchingTheTrunk.'
          round: 6
        - id: BR-24
          disposition: addressed
          note: 'Verified by mutation: deleting the in.Validate() call from runQueueEdit reddens all six sub-cases with "Update ran 1 times". Residue, not re-raised: guardSpineRepo and openTrunkStore each run `git rev-parse --show-toplevel` before validation, so queue.go:194''s "BEFORE any git call" is literally still false (and the two calls are redundant with each other).'
          round: 6
        - id: BR-25
          disposition: not-addressed
          note: 'Tag and cr are now carried (both mutation-verified), but the Kind cell the finding enumerated is not. Reproduced: `add sdlc-fleet "sharper reason"` onto "- project:sdlc-fleet — the whole area is next [sdlc]" yields "- sdlc-fleet — sharper reason [sdlc]" with the note "updated why-now ... and kind" — the peer''s project marker is destroyed by an edit that never mentioned it, now announced rather than silent. Root cause is that KindIssue is Kind''s zero value, so "not mentioned" is unrepresentable; fix with a KindUnspecified zero or c.Flags().Changed("project"). Second residue in the same rule: the APPEND branch (intent.go:153) does not carry the document''s line ending, so adding to a CRLF file yields mixed endings.'
          round: 6
        - id: BR-26
          disposition: not-addressed
          note: The listing half is fully done and pinned (TestQueueList_ReportsUnrecognizedItems), and helptext/queue.md:17-30 now carries the grammar. The datatype-prose half is untouched — construct/datatype/queue.md is not in the fix commit and still says the operational contract "is not restated here", so the Done-when clause "the prose states ... the line format" remains unmet. One short section closes it.
          round: 6
        - id: BR-27
          disposition: addressed
          note: guardSpineRepo is wired into all three write subcommands and the read/write asymmetry is argued in queue.go:22-34. The residue — no test pins it and repoguard.go's enumeration claim is now false — is raised separately as its own finding rather than re-raised here.
          round: 6
        - id: BR-28
          disposition: addressed
          note: 'The #207 note landed on 000207-sync-without-worktree.md, names which file it chose, and carries the retry-semantics amendment. The plan gained an M2 Revisions entry that is unusually honest about the blanket-regex ticking. Residue carried into the plan-revision recommendations: 04804ff still has no task row.'
          round: 6
        - id: BR-29
          disposition: not-addressed
          note: Untouched — trunkfile.go is not in the fix commit. pathPresent:219/modeOf:426 and refPresent:178/resolve:356 both remain, and Update still runs both halves of both pairs per attempt.
          round: 6
        - id: BR-30
          disposition: not-addressed
          note: 'Half addressed. TestQueueEdit_ValidationRefusesBeforeTouchingTheTrunk can now see a git call (it asserts f.calls == 0, mutation-verified). The replay half is unchanged: fakeTrunk.Update still calls transform exactly once and fires peer BEFORE it, so TestQueueEdit_PeerEditSurvivesTheReplay would pass against an Update with no retry, and queue_e2e_test.go still has no peer push. Also, the `reached` field added for this is incremented at two sites and read at none, while queue_test.go:36-38 credits it with making the claim testable.'
          round: 6
        - id: BR-31
          disposition: not-addressed
          note: Untouched. workshop/plans/000209-queue-datatype-plan.md still names queueCmd at lines 81, 87 and 91, and has no row for gitx.FirstLine.
          round: 6
        - id: BR-32
          disposition: addressed
          note: 'queue.go:249-251 now prints "(this snapshot is itself stale: ...)". The redundant second full fetch on the error path remains and is folded into the BR-11 enumeration.'
          round: 6
      findings:
        - id: BR-33
          severity: Important
          title: A field that passes Validate can still render a line that re-parses to a different record, producing entries that duplicate and cannot be removed
          detail: 'This is the 2nd finding in family unvalidated-format-field. Do NOT patch these two inputs — state the rule. RULE a validator that enumerates forbidden characters does not make the record round-trip; the write path must assert what the read path already asserts. Two instances reproduced against the real code, both exit 0. (1) Intent{OpAdd, Ref "project:foo", WhyNow "why"} validates, publishes "- project:foo — why", and re-reads as {Ref "foo", Kind KindProject}; `queue remove project:foo` then reports "project:foo was not in the queue; nothing to remove" and leaves it, and a second identical add appends a duplicate instead of converging. ValidateRef rejects brackets, whitespace and the em-dash separator but not the project: prefix, which is structure in that position. (2) Intent{OpAdd, Ref "a#1", WhyNow "  padded  "} validates and publishes "- a#1 —   padded  ", which ParseLine''s own self-check rejects — Entries() is empty and UnrecognizedItems() is 1, so the verb manufactures the state runQueueList blames on "a hand-edit that used a hyphen". Fix covering both and every future field - in applyAdd, after building the Line, refuse unless ParseLine(line.String()) yields an equal record, the same mechanical guarantee line.go:110-123 gives the read side (ARCH-DRY, ARCH-SECURE).'
          family: unvalidated-format-field
          round: 6
        - id: BR-34
          severity: Important
          title: The spine guard on the queue write verbs is pinned by no test, and repoguard.go's enumeration claim is now false
          detail: 'This is the 2nd finding in family unguarded-write-verb. Do NOT just add three assertions — state the rule. RULE a guard is wired only when it is in the enumeration the drift test walks; a guard reachable solely by remembering to call it is a comment. repoguard.go:8-11 states the guard is wired into "exactly the lifecycle verbs (... processmanual.WorkflowVerbs; the drift test enumerates it)", which is now untrue - queue add/remove/move call guardSpineRepo and are outside WorkflowVerbs, correctly so since adding them there would distort the #172 friction instrument. Consequence - no test constructs NewQueueCmd''s command tree at all, so nothing pins the guard, the --project to Kind wiring, --tag, or the --before/--after mutual exclusion, and a refactor dropping the guard ships silently from the base-layer binary to every downstream repo. Separately, newQueueMoveCmd runs flag validation before the guard, contrary to the guard-first placement repoguard.go:58-60 states. Fix - correct the "exactly" sentence to name both sets, and add a brain-repo table test over the three write subcommands plus one asserting the bare list is NOT guarded.'
          family: unguarded-write-verb
          round: 6
        - id: BR-35
          severity: Minor
          title: The item-marker predicate is written twice with different rules, so indented prose bullets are reported as unparsed entries
          detail: This is the 5th finding in family duplicated-helper. Do NOT fix this instance — the rule is already stated at trunkfile.go:230-235 and just needs extending from functions to predicates - two implementations of one predicate are one function, and where they diverge one of them is a bug. line.go:73 tests CutPrefix(trimmed, "- ") on the raw line; doc.go:75 tests HasPrefix(TrimSpace(l.raw), "- "). Reproduced - a queue file whose prose header contains an ordinary indented list ("  - the top line is next") reports 2 lines that "look like entries but do not parse", which is exactly the reader-training failure doc.go:68-71 says the TrimSpace-free rule avoids. TestQueueList_ReportsUnrecognizedItems uses only unindented prose so it cannot see this. Extract looksLikeItem(s string) bool and call it from both.
          family: duplicated-helper
          round: 6
        - id: BR-36
          severity: Minor
          title: add appends at the end of the FILE, so a new entry lands below a trailing comment block, and move --after disagrees with it
          detail: intent.go:153 appends to d.lines unconditionally. Reproduced - adding b#2 to "# Queue\n\n- a#1 — first\n\n<!-- keep this at the bottom -->\n" renders the new entry BELOW the comment, while `move b#2 --after a#1` would place it above. RULE in a document that interleaves records with free text there is a record region, and an insert belongs at that region's edge, not the file's. The file has no prose today so nothing breaks yet, but Doc exists precisely to support hand-authored prose and the plan's Task 8 names a trailing comment block as a case. TestIntent_Apply_PreservesProse asserts the comment survives but not where the entry lands.
          family: record-region-boundary
          round: 6
      blocked: true
    - "n": 7
      timestamp: "2026-09-07T19:32:13-07:00"
      agent: claude
      dispose:
        - id: BR-11
          disposition: not-addressed
          note: All three sites reproduced unchanged against real git; trunkfile.go:200 and :166-169 untouched this round.
          round: 7
        - id: BR-12
          disposition: not-addressed
          note: trunkfile.go:392-394 untouched — still no sentence on where .gitattributes resolves from.
          round: 7
        - id: BR-21
          disposition: not-addressed
          note: Behaviour now correct (verified both directions), but mutating localRef back to the unqualified branch leaves the whole gitx package green — no test pins it.
          round: 7
        - id: BR-25
          disposition: not-addressed
          note: Fixed at the type layer and unreachable from the CLI — queue.go:90 hardcodes KindIssue, so KindUnspecified is produced at zero production call sites; reproduced against the built binary at exit 0.
          round: 7
        - id: BR-26
          disposition: addressed
          note: 'construct/datatype/queue.md now carries a "The line format" section stating the grammar, the project: prefix and the why-now restrictions; the listing half was already pinned.'
          round: 7
        - id: BR-29
          disposition: not-addressed
          note: trunkfile.go not in this round's commit for these sites — pathPresent/modeOf and refPresent/resolve both remain, and Update still runs all four per attempt.
          round: 7
        - id: BR-30
          disposition: not-addressed
          note: fakeTrunk still calls transform once with peer firing before it; queue_e2e_test.go still has no peer push; `reached` is still written at two sites and read at none.
          round: 7
        - id: BR-31
          disposition: not-addressed
          note: Plan lines 81/87/91 still say queueCmd; no gitx.FirstLine row; no row for the e2e (04804ff).
          round: 7
        - id: BR-33
          disposition: addressed
          note: Mutation-verified — disabling the round-trip check reddens TestIntent_Validate_RejectsRecordsThatDoNotSurviveARoundTrip; the padded-why-now instance is caught by the same mechanism without being enumerated.
          round: 7
        - id: BR-34
          disposition: not-addressed
          note: Claim corrected and presence pinned (mutation-verified), but the guard-first residue is live — newQueueMoveCmd still validates --before/--after before guardSpineRepo, and a source-grep test cannot observe ordering.
          round: 7
        - id: BR-35
          disposition: not-addressed
          note: Reproduced — an indented prose list yields unrecognized=2; line.go:82 and doc.go:86 still disagree on the item-marker predicate.
          round: 7
        - id: BR-36
          disposition: not-addressed
          note: Reproduced — a new entry renders below a trailing comment block; intent.go:190 still appends to the end of the file.
          round: 7
      findings:
        - id: BR-37
          severity: Important
          title: A converged no-op still pushes an empty commit to origin/main whose subject claims the edit it did not make
          detail: |-
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
          family: error-misattribution
          round: 7
        - id: BR-38
          severity: Minor
          title: The "Apply must refuse too (defence in depth)" rows pass for an unrelated reason on move and remove, because Apply validates only OpAdd
          detail: |-
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
          family: vacuous-test-guard
          round: 7
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

## Round 3 — 2026-09-07T17:51:06-07:00 (claude) — BLOCKED

### Disposed

- BR-3 — not-addressed — Classifier half landed and is now structurally pinned (hardwiring refExists to true reddens TestTrunkFile_NoRemoteIsNamed), but the byte assertion is still missing: hardwiring readLocal to return nil,nil leaves the whole package green. That gap is what let I-1 ship.
- BR-11 — not-addressed — trunkfile.go:150 and ReadDegraded:127-130 unchanged; Update:290 still reports "the trunk moved" for any commitAndPush failure coinciding with a peer push.
- BR-12 — not-addressed — trunkfile.go:350-352 unchanged; still no sentence saying .gitattributes resolves from the checked-out tree rather than from the trunk being written.
- BR-14 — addressed — Verified twice: collapsing pathPresent to two states reddens TestTrunkFile_UnreadablePathRefusesRatherThanTruncating, and a stub-free end-to-end against a real deleted loose blob gives Read->err, transform not called, trunk intact.
- BR-15 — not-addressed — grep -n "combined\|Combined" trunkfile.go still returns :37 and :332, both unchanged.
- BR-16 — not-addressed — Plan :86, :136, :207 unchanged. Same enumeration also shows gitx.FirstLine, a newly exported symbol, missing from the plan's Core-concepts surface.
- BR-17 — addressed — One table over true/yes/1/on/false/0/off; reverting --type=bool reddens the yes, 1 and on rows. The negative cases are now covered, which the two old bodies never were.
- BR-18 — addressed — Fetch half pinned — reverting the attempt==1 hoist reddens TestTrunkFile_OneFetchPerAttempt with "3 fetches for 2 attempts". The signs() hoist has no test; behavior-neutral, and a call-count assertion would only restate the implementation.

### Raised

- **BR-19** [Important] `git-output-string-matching` readLocal treats a legitimately absent local branch as unreadable, so ReadDegraded errors obscurely in a fresh repo while warning that it used the local branch
  4th finding in this family — complete the RULE, do not patch the site.
  BR-14 stated half of it (an ambiguous signal must error). The missing half
  is that an unambiguous BENIGN signal must not be swept into that error.
  pathPresent (trunkfile.go:220) maps every non-zero ls-tree exit onto
  "cannot tell", which is right for an unreadable tree and wrong for a ref
  that does not resolve — a state readLocal reaches by design. Reproduced
  against the real type: `git init -b main` with no commits, and a repo whose
  local default branch is `master`, both give
  err="ls-tree main -- <path>: exit status 128 / fatal: Not a valid object
  name main" WITH warn ending "used local main", asserting something that did
  not happen. This is a regression introduced in this window; at 850eabd
  exists() swallowed it and the call returned (nil, warn, nil). It violates
  the Done-when clause "a repo with no origin remote at all reads the local
  ref and says so … rather than erroring obscurely". Re-run enumeration under
  the completed rule readRef's tracking ref cannot legitimately be absent
  (fetch precedes it, ReadDegraded guards with refExists) so its failures are
  genuine; readLocal's branch can, and must be guarded with the same
  refExists signal used one line above at trunkfile.go:132.
- **BR-20** [Minor] `dead-parameter` `_ = errOut` at trunkfile.go:200 is a no-op left behind by the three-states edit
  2nd finding in this family — state the rule rather than deleting the line.
  Rule no statement or parameter may exist solely to reference something;
  errOut is already consumed by the error branch at :198, so the assignment
  does nothing and go vet will not catch it. The enumeration is
  `grep -n '^\s*_ = ' cmd/sdlc/internal/gitx/`, which returns exactly this one
  site — measured prevalence 1, so the sweep is cheap and complete.

## Round 4 — 2026-09-07T18:03:54-07:00 (claude) — passed

### Disposed

- BR-3 — addressed — isMissingPath/isNonFastForward are gone from the tree (grep over cmd/ returns nothing); classification is now exit-code based via gitExitCode/gitAbsentExit and pinned by TestGitExitCode_NonExitFailureIsNotAbsent plus TestTrunkFile_RefPresentPropagatesNonAbsentFailure, which I mutation-verified goes red when the error branch is collapsed.
- BR-11 — not-addressed — Unchanged, and now upgraded from wording to ARCH-SECURE. RULE a message must describe the state the code observed, not the state it hoped for. Enumeration, 3 sites, all reproduced against the real type: (1) offlineError trunkfile.go:192 and (2) ReadDegraded's warn trunkfile.go:158 both say "unreachable (offline?)" for a reachable remote missing the branch — measured "origin unreachable (offline?): fatal: couldn't find remote ref refs/heads/main"; (3) trunkfile.go:170 appends "used local main" even when no local main exists, so a repo whose default branch is master yields bytes="" warn="…used local main" err=nil — a fabricated empty value carrying a claim that did not happen. Fix by branching on the observed signal already in hand (fetch's exit/stderr class; the refPresent result readLocal computes).
- BR-12 — not-addressed — Unchanged. trunkfile.go:404-406 still explains what --path buys with no sentence noting the attributes come from the working tree/index, not from the trunk being written.
- BR-15 — not-addressed — Unchanged. grep -n "combined\|Combined" trunkfile.go still returns :37 ("returning combined output", contradicted by :39) and :386 ("Returns git's combined output", where commitAndPush returns stderr only).
- BR-16 — not-addressed — Unchanged, and the enumeration is larger than the three lines named. Beyond plan lines 86, 136 and 207, the same rule catches line 7 ("a sibling runEnv"; the code adds one runGitIn and leaves run alone), the Revisions entry's own "replaced isMissingPath's phrase matching with cat-file -e's exit code" (cat-file -e was itself rejected as two-state; the code uses ls-tree), and line 171's checked-off promise of a real signed-commit test with t.Skip, which the delivered test does not do.
- BR-19 — addressed — Verified by mutation, not by the diff: deleting the refPresent(t.branch) guard at trunkfile.go:217-223 turns TestTrunkFile_FreshRepoReadsEmptyNotError red with "ls-tree main -- workshop/queue.md: exit status 128 / fatal: Not a valid object name main". The residual false "used local main" warning is folded into BR-11.
- BR-20 — addressed — grep -n '^\s*_ = ' over cmd/sdlc/internal/gitx/ now returns only the test's idiomatic _ = cmd.Run(); the no-op assignment is gone.

### Raised

- **BR-21** [Important] `duplicated-helper` readLocal and readRef are the same three-line body with a different ref, and the guard exists on only one of them
  This is the 3rd finding in family `duplicated-helper`. Earlier rounds fixed
  instances (firstLineOf/firstLine, the two signing tests). Do NOT merge just
  these two — state the rule. RULE two functions differing only in the value
  of a parameter are one function; when they diverge, the divergence is a bug
  in the one that lacks it, not a feature of the one that has it.
  Enumeration by that rule over cmd/sdlc/internal/gitx/: 1 remaining site.
  trunkfile.go:224-235 (readLocal) and :241-252 (readRef) both run
  pathPresent(ref, path) then cat-file blob ref:path, differing only in
  t.branch vs t.trackingRef() and in the error string. readLocal additionally
  guards ref existence via refPresent; readRef does not — which is exactly the
  BR-19 defect class, currently unreachable only because ReadDegraded and
  Update happen to check the tracking ref before calling it. Collapse to
  readAt(ref, path) carrying the guard, with readLocal/readRef as one-line
  callers. Doing so also fixes a second thing for free: readLocal passes the
  UNQUALIFIED branch name to refPresent, pathPresent and cat-file, so a tag
  named `main` shadows refs/heads/main under rev-parse's DWIM order; readAt
  should take a fully-qualified ref and readLocal should pass
  "refs/heads/" + t.branch (ARCH-DRY).
- **BR-22** [Minor] `unspecified-fields-not-preserved` Update rebuilds the tree entry from scratch, so an executable path on the trunk comes back mode 100644
  trunkfile.go:413-414 hardcodes `--cacheinfo 100644,<blob>,<path>`. Verified
  against the real type: a 100755 blob on the trunk is 100644 after an
  Update that only changed content. RULE when replacing one field of an
  existing record, carry the fields you were not asked to change. The mode is
  already on the wire — pathPresent (trunkfile.go:272) runs ls-tree and
  discards everything but the name; dropping --name-only there yields the
  existing mode for free. Alternatively state "regular file, mode 100644" as
  the type's contract in the doc comment and the atlas entry. Harmless for
  M1's .md consumers, but ariadne#207 inherits this primitive and no test
  would notice.

## Round 5 — 2026-09-07T18:52:12-07:00 (claude) — BLOCKED

### Disposed

- BR-11 — not-addressed — Unchanged at trunkfile.go:192 and :158, and the enumeration is now larger — M2 added two sites. (4) queue.go:212 prints "the queue on the trunk right now (N entries)" from a ReadDegraded whose warning it discards; reproduced offline, it printed "(0 entries)" for a trunk that has four, i.e. a fabricated state presented as observed. (5) Update:317 calls offlineError unconditionally, so a repo with NO origin gets "origin unreachable (offline?)" although ReadDegraded has a dedicated branch for exactly that case. (6) intent.go:96 reports "kept the newer why-now" for an edit that also dropped the tag and flipped the kind (see the new I-3). Same rule: a message may only name state the code observed.
- BR-12 — not-addressed — Unchanged. trunkfile.go:383-385 still explains what --path buys with no sentence noting the attributes are resolved from the working tree/index, not from the trunk being written.
- BR-15 — addressed — Verified by grep: `grep -in "combined" trunkfile.go` now returns only :39 ("returned SEPARATELY, not combined"), and commitAndPush:365 reads "Returns git's STDERR".
- BR-16 — addressed — Verified by grep over the plan: no CombinedOutput at lines 86/136/207 or anywhere outside the Revisions entry describing the history. Residual plan-vs-code staleness (line 7's "sibling runEnv" vs the delivered runGitIn; line 171's t.Skip signing test vs the delivered -S interception; the Revisions entry's "cat-file -e" vs the delivered ls-tree) is carried forward under the new plan-traceability finding rather than re-raised here.
- BR-21 — not-addressed — The main defect IS fixed and properly pinned — readLocal/readRef collapsed into readFrom carrying the guard; verified by mutation (deleting the refPresent block turns both ReadFromRefusesUndeterminableRef and FreshRepoReadsEmptyNotError red). The only residue is the half the finding asked for and the diff did not do: trunkfile.go:169 still passes the UNQUALIFIED t.branch. Reproduced against the real type — a repo with a tag named "main" and no tracking ref returns the TAG's bytes while warning "used local main". One-line fix: pass "refs/heads/" + t.branch.
- BR-22 — addressed — modeOf + TestTrunkFile_PreservesFileMode. Verified by mutation, not by the diff: replacing the modeOf call with a hardcoded "100644" fails the test with "mode = [100644 ... run.sh], want 100755 preserved". The new-path 100644 case is covered in the same test. (Note M-1: the fix added a SECOND ls-tree rather than widening the one pathPresent already runs, which is what the finding sketched.)

### Raised

- **BR-23** [Important] `unvalidated-format-field` --tag is the one user-supplied field that lands in the line format with no validator; a newline in it publishes a forged file with exit 0
  Reproduced against the real code path. Intent{Op: OpAdd, Ref: "b#2", WhyNow: "why", Tag: "x\n- forged#9 — injected [y"} returns nil and publishes three entries where one was intended, mangling the real entry to "- b#2 — why [x" and forging "- forged#9 — injected [y]". ValidateWhyNow exists to stop exactly this forge (line_test.go:96 demonstrates it) and ValidateRef covers the third field; Tag was simply not in the enumeration. RULE every user-supplied field that lands in the line format is validated by the same rule. Add ValidateTag (newline/control/sep/bracket, plus non-blank-after-trim, since a whitespace-only tag is silently absorbed into why-now) and call it from applyAdd beside the other two.
- **BR-24** [Important] `validation-behind-io` Validation runs inside the transform, so "rejected before any git call" is false and the offline error masks it
  Path is RunE -> openTrunkStore -> gitx.RepoTopLevel (git) -> Update -> signs (git config) -> fetch (network) -> transform -> ValidateRef. Reproduced with an unreachable origin: `queue add "bad ref" x` prints "origin unreachable (offline?)" and a fabricated "(0 entries)" listing; the operator never learns the ref was malformed. The Spec's Done-when, the Spec's "Input validation, before any git call" paragraph and the plan's Task 9 table all promise the opposite. queue_test.go:168 claims to pin this but only asserts the fake's content is unchanged, so it cannot observe a git call and passes today. Fix: an Intent.Validate() called in runQueueEdit before openTrunkStore, keeping the in-Apply calls as the belt, plus a call-counting seam in the test.
- **BR-25** [Important] `unspecified-fields-not-preserved` Converge replaces the whole line, silently flipping kind and dropping the tag, while the note claims only the why-now moved
  This is the 2nd finding in family unspecified-fields-not-preserved. Do NOT fix only this site — state the rule. RULE when replacing one field of an existing record, carry every field you were not asked to change, and a note may only claim the change it actually made. BR-22 stated it for git tree modes; the record here is {Ref, WhyNow, Tag, Kind, cr} and intent.go:90-98 rebuilds it from the Intent alone. Reproduced: `add sdlc-fleet "sharper reason"` onto "- project:sdlc-fleet — the whole area is next [sdlc]" yields "- sdlc-fleet — sharper reason" with the note "kept the newer why-now (was ...)" — project marker and tag gone, unmentioned. Second cell: converging onto a CRLF line yields "- a#1 — one prime\n- b#2 — two\r\n", mixing line endings. This is the interleaving table's concurrent-add cell, so a peer's --project/--tag vanishes on the path built to make peers survive; the table test asserts only refs plus a note substring and cannot catch it.
- **BR-26** [Important] `format-undocumented` The line format is stated in no user-facing artifact, and a near-miss hand-edit disappears from the listing without a word
  The Done-when requires the datatype prose to state "the line format". construct/datatype/queue.md does not, and routes the operational contract to `sdlc queue --help`, which does not carry it either. The grammar — "- <ref> — <why> [tag]", the em-dash separator, the `project:` prefix that is the ONLY marker of a project line, the why-now restrictions — lives only in Go comments. That matters because the design explicitly invites hand-editing (line.go:45-48, the fuzz target, and the comment that an unparsed line is one `queue remove` cannot find). Compounding: runQueueList prints Entries() only, so a hand-edited "- a#1 - why" (ASCII hyphen) is silently absent from the listing while sitting in the file. Two cheap fixes: a "Line format" section in the datatype prose and the helptext, and an stderr line in runQueueList reporting len(lines)-len(entries) unparsed lines.
- **BR-27** [Important] `unguarded-write-verb` A new trunk-writing verb shipped into the base-layer binary with no guardSpineRepo decision recorded either way
  repoguard.go:5-14 states the guard is wired into the lifecycle verbs and that reads stay unguarded by construction. `queue add/remove/move` are writes that push to origin/main, and they are unguarded — in a brain repo that is a CAS push at a gcrypt::ssh remote, and in a repo without workshop/issues/ it creates workshop/queue.md on main. This issue's own Spec calls the brain-queue question "an open disagreement, not a settled no", so shipping the unguarded path quietly settles it; ariadne is the base layer, so the verb reaches every downstream repo. migrate.go:460 is the established shape for the other answer — an explicit comment saying why the exemption holds. Either wire guardSpineRepo into the three mutating subcommands or record the exemption the way migrate does.
- **BR-28** [Important] `plan-traceability` Task 13 Step 3 is ticked but no note exists on either #207 file, and three more Chunk-2 rows claim artifacts the tree does not have
  This is the 3rd finding in family plan-traceability. Do NOT fix only this site — state the rule. RULE a "- [x]" is a claim about the tree, not about intent; before a close, sweep every checked row in the closing chunk against the artifact it names. Enumeration over Chunk 2, 4 rows fail: (a) Task 13 Step 3 — grep -i "trunkfile|#209" over both workshop/issues/000207-*.md returns nothing, so #207 still has no record that gitx.TrunkFile exists nor of the retry-semantics defect the Spec promised to flag there (say which of the two #207 files gets it); (b) Task 11 Step 5's commit does not exist, folded into c93be7c; (c) Task 13 Step 4's commit does not exist — the seed landed as `queue: add ...` commits on origin/main; (d) 04804ff (the real-git e2e) is real work with no plan row. M1 recorded exactly this divergence in Revisions; M2's chunk has no such entry.
- **BR-29** [Minor] `duplicated-helper` pathPresent/modeOf and refPresent/resolve are each one function split in two, and Update runs both halves of both pairs per attempt
  This is the 4th finding in family duplicated-helper. Do NOT merge only these — state the rule, which the code already states at trunkfile.go:230-235: two functions differing only in a parameter's value, or in which field of one result they keep, are one function. Enumeration over cmd/sdlc/internal/gitx/, 2 remaining sites. trunkfile.go:219 (pathPresent) and :426 (modeOf) run the same ls-tree on the same (ref, path), differing only by --name-only — and BR-22's own fix sketch said to drop --name-only and take the mode from the call already on the wire, so the fix for BR-22 created this instance. trunkfile.go:178 (refPresent) and :356 (resolve) run the same rev-parse --verify, differing only by --quiet and return type; Update calls both on trackingRef each attempt. Collapse each pair to one call returning the richer result.
- **BR-30** [Minor] `vacuous-test-guard` fakeTrunk calls the transform exactly once and fires its peer before it, so two tests are named for properties the double cannot produce
  This is the 2nd finding in family vacuous-test-guard. State the rule rather than patching a site: a test may not be named for a property its double cannot produce. queue_test.go:37-51 — Update calls transform once and runs `peer` BEFORE it, so TestQueueEdit_PeerEditSurvivesTheReplay ("the property the whole design exists for") would pass against an Update with no retry at all, and TestQueueEdit_ValidationRefusesBeforeTouchingTheTrunk cannot see a git call. Consequence: no test anywhere drives Intent.Apply through a real non-fast-forward rejection — gitx's two-publisher test uses a raw append transform and the e2e has no peer. Give fakeTrunk a rejection mode that re-invokes the transform, and add a peer push to queue_e2e_test.go.
- **BR-31** [Minor] `plan-traceability` Core-concepts table names queueCmd where the symbol is NewQueueCmd, and gitx.FirstLine has no row
  Same family as above; listed separately only because it is a table edit rather than a checkbox. FirstLine was newly exported in this window (moved out of issueids.go) and is part of the package's surface.
- **BR-32** [Minor] `error-misattribution` queueRefusal discards ReadDegraded's staleness warning and issues a second full fetch on the error path
  queue.go:212 binds the warning to _ , so the handoff prints "the queue on the trunk right now" with no indication the read was degraded. Folded into the BR-11 enumeration above; noted here for the file:line.

## Round 6 — 2026-09-07T19:16:10-07:00 (claude) — BLOCKED

### Disposed

- BR-11 — not-addressed — trunkfile.go:191-193 and Update:317 are untouched by this round (trunkfile.go is not in 66065f6). queueRefusal now prints the staleness warning, which was BR-32's half, not this one.
- BR-12 — not-addressed — Untouched. trunkfile.go:383-385 still explains what --path buys with no sentence noting the attributes resolve from the working tree/index, not from the trunk being written.
- BR-21 — not-addressed — The fix commit dismisses this as "stale" on the readFrom merge, but the finding's second half is live and reproduced. trunkfile.go:169 still passes the UNQUALIFIED t.branch; I confirmed against real git that with both refs/tags/main and refs/heads/main present, `rev-parse --verify main` and `cat-file blob main:q.md` return the TAG. So ReadDegraded's no-tracking-ref fallback can serve a tag's bytes while warning "used local main". One-line fix, unchanged: pass "refs/heads/" + t.branch.
- BR-23 — addressed — ValidateTag exists and is pinned twice. Verified by mutation, not by the diff: short-circuiting ValidateTag to nil reddens TestIntent_Validate_CoversEveryInterpolatedField/tag and two sub-cases of TestQueueEdit_ValidationRefusesBeforeTouchingTheTrunk.
- BR-24 — addressed — Verified by mutation: deleting the in.Validate() call from runQueueEdit reddens all six sub-cases with "Update ran 1 times". Residue, not re-raised: guardSpineRepo and openTrunkStore each run `git rev-parse --show-toplevel` before validation, so queue.go:194's "BEFORE any git call" is literally still false (and the two calls are redundant with each other).
- BR-25 — not-addressed — Tag and cr are now carried (both mutation-verified), but the Kind cell the finding enumerated is not. Reproduced: `add sdlc-fleet "sharper reason"` onto "- project:sdlc-fleet — the whole area is next [sdlc]" yields "- sdlc-fleet — sharper reason [sdlc]" with the note "updated why-now ... and kind" — the peer's project marker is destroyed by an edit that never mentioned it, now announced rather than silent. Root cause is that KindIssue is Kind's zero value, so "not mentioned" is unrepresentable; fix with a KindUnspecified zero or c.Flags().Changed("project"). Second residue in the same rule: the APPEND branch (intent.go:153) does not carry the document's line ending, so adding to a CRLF file yields mixed endings.
- BR-26 — not-addressed — The listing half is fully done and pinned (TestQueueList_ReportsUnrecognizedItems), and helptext/queue.md:17-30 now carries the grammar. The datatype-prose half is untouched — construct/datatype/queue.md is not in the fix commit and still says the operational contract "is not restated here", so the Done-when clause "the prose states ... the line format" remains unmet. One short section closes it.
- BR-27 — addressed — guardSpineRepo is wired into all three write subcommands and the read/write asymmetry is argued in queue.go:22-34. The residue — no test pins it and repoguard.go's enumeration claim is now false — is raised separately as its own finding rather than re-raised here.
- BR-28 — addressed — The #207 note landed on 000207-sync-without-worktree.md, names which file it chose, and carries the retry-semantics amendment. The plan gained an M2 Revisions entry that is unusually honest about the blanket-regex ticking. Residue carried into the plan-revision recommendations: 04804ff still has no task row.
- BR-29 — not-addressed — Untouched — trunkfile.go is not in the fix commit. pathPresent:219/modeOf:426 and refPresent:178/resolve:356 both remain, and Update still runs both halves of both pairs per attempt.
- BR-30 — not-addressed — Half addressed. TestQueueEdit_ValidationRefusesBeforeTouchingTheTrunk can now see a git call (it asserts f.calls == 0, mutation-verified). The replay half is unchanged: fakeTrunk.Update still calls transform exactly once and fires peer BEFORE it, so TestQueueEdit_PeerEditSurvivesTheReplay would pass against an Update with no retry, and queue_e2e_test.go still has no peer push. Also, the `reached` field added for this is incremented at two sites and read at none, while queue_test.go:36-38 credits it with making the claim testable.
- BR-31 — not-addressed — Untouched. workshop/plans/000209-queue-datatype-plan.md still names queueCmd at lines 81, 87 and 91, and has no row for gitx.FirstLine.
- BR-32 — addressed — queue.go:249-251 now prints "(this snapshot is itself stale: ...)". The redundant second full fetch on the error path remains and is folded into the BR-11 enumeration.

### Raised

- **BR-33** [Important] `unvalidated-format-field` A field that passes Validate can still render a line that re-parses to a different record, producing entries that duplicate and cannot be removed
  This is the 2nd finding in family unvalidated-format-field. Do NOT patch these two inputs — state the rule. RULE a validator that enumerates forbidden characters does not make the record round-trip; the write path must assert what the read path already asserts. Two instances reproduced against the real code, both exit 0. (1) Intent{OpAdd, Ref "project:foo", WhyNow "why"} validates, publishes "- project:foo — why", and re-reads as {Ref "foo", Kind KindProject}; `queue remove project:foo` then reports "project:foo was not in the queue; nothing to remove" and leaves it, and a second identical add appends a duplicate instead of converging. ValidateRef rejects brackets, whitespace and the em-dash separator but not the project: prefix, which is structure in that position. (2) Intent{OpAdd, Ref "a#1", WhyNow "  padded  "} validates and publishes "- a#1 —   padded  ", which ParseLine's own self-check rejects — Entries() is empty and UnrecognizedItems() is 1, so the verb manufactures the state runQueueList blames on "a hand-edit that used a hyphen". Fix covering both and every future field - in applyAdd, after building the Line, refuse unless ParseLine(line.String()) yields an equal record, the same mechanical guarantee line.go:110-123 gives the read side (ARCH-DRY, ARCH-SECURE).
- **BR-34** [Important] `unguarded-write-verb` The spine guard on the queue write verbs is pinned by no test, and repoguard.go's enumeration claim is now false
  This is the 2nd finding in family unguarded-write-verb. Do NOT just add three assertions — state the rule. RULE a guard is wired only when it is in the enumeration the drift test walks; a guard reachable solely by remembering to call it is a comment. repoguard.go:8-11 states the guard is wired into "exactly the lifecycle verbs (... processmanual.WorkflowVerbs; the drift test enumerates it)", which is now untrue - queue add/remove/move call guardSpineRepo and are outside WorkflowVerbs, correctly so since adding them there would distort the #172 friction instrument. Consequence - no test constructs NewQueueCmd's command tree at all, so nothing pins the guard, the --project to Kind wiring, --tag, or the --before/--after mutual exclusion, and a refactor dropping the guard ships silently from the base-layer binary to every downstream repo. Separately, newQueueMoveCmd runs flag validation before the guard, contrary to the guard-first placement repoguard.go:58-60 states. Fix - correct the "exactly" sentence to name both sets, and add a brain-repo table test over the three write subcommands plus one asserting the bare list is NOT guarded.
- **BR-35** [Minor] `duplicated-helper` The item-marker predicate is written twice with different rules, so indented prose bullets are reported as unparsed entries
  This is the 5th finding in family duplicated-helper. Do NOT fix this instance — the rule is already stated at trunkfile.go:230-235 and just needs extending from functions to predicates - two implementations of one predicate are one function, and where they diverge one of them is a bug. line.go:73 tests CutPrefix(trimmed, "- ") on the raw line; doc.go:75 tests HasPrefix(TrimSpace(l.raw), "- "). Reproduced - a queue file whose prose header contains an ordinary indented list ("  - the top line is next") reports 2 lines that "look like entries but do not parse", which is exactly the reader-training failure doc.go:68-71 says the TrimSpace-free rule avoids. TestQueueList_ReportsUnrecognizedItems uses only unindented prose so it cannot see this. Extract looksLikeItem(s string) bool and call it from both.
- **BR-36** [Minor] `record-region-boundary` add appends at the end of the FILE, so a new entry lands below a trailing comment block, and move --after disagrees with it
  intent.go:153 appends to d.lines unconditionally. Reproduced - adding b#2 to "# Queue\n\n- a#1 — first\n\n<!-- keep this at the bottom -->\n" renders the new entry BELOW the comment, while `move b#2 --after a#1` would place it above. RULE in a document that interleaves records with free text there is a record region, and an insert belongs at that region's edge, not the file's. The file has no prose today so nothing breaks yet, but Doc exists precisely to support hand-authored prose and the plan's Task 8 names a trailing comment block as a case. TestIntent_Apply_PreservesProse asserts the comment survives but not where the entry lands.

## Round 7 — 2026-09-07T19:32:13-07:00 (claude) — BLOCKED

### Disposed

- BR-11 — not-addressed — All three sites reproduced unchanged against real git; trunkfile.go:200 and :166-169 untouched this round.
- BR-12 — not-addressed — trunkfile.go:392-394 untouched — still no sentence on where .gitattributes resolves from.
- BR-21 — not-addressed — Behaviour now correct (verified both directions), but mutating localRef back to the unqualified branch leaves the whole gitx package green — no test pins it.
- BR-25 — not-addressed — Fixed at the type layer and unreachable from the CLI — queue.go:90 hardcodes KindIssue, so KindUnspecified is produced at zero production call sites; reproduced against the built binary at exit 0.
- BR-26 — addressed — construct/datatype/queue.md now carries a "The line format" section stating the grammar, the project: prefix and the why-now restrictions; the listing half was already pinned.
- BR-29 — not-addressed — trunkfile.go not in this round's commit for these sites — pathPresent/modeOf and refPresent/resolve both remain, and Update still runs all four per attempt.
- BR-30 — not-addressed — fakeTrunk still calls transform once with peer firing before it; queue_e2e_test.go still has no peer push; `reached` is still written at two sites and read at none.
- BR-31 — not-addressed — Plan lines 81/87/91 still say queueCmd; no gitx.FirstLine row; no row for the e2e (04804ff).
- BR-33 — addressed — Mutation-verified — disabling the round-trip check reddens TestIntent_Validate_RejectsRecordsThatDoNotSurviveARoundTrip; the padded-why-now instance is caught by the same mechanism without being enumerated.
- BR-34 — not-addressed — Claim corrected and presence pinned (mutation-verified), but the guard-first residue is live — newQueueMoveCmd still validates --before/--after before guardSpineRepo, and a source-grep test cannot observe ordering.
- BR-35 — not-addressed — Reproduced — an indented prose list yields unrecognized=2; line.go:82 and doc.go:86 still disagree on the item-marker predicate.
- BR-36 — not-addressed — Reproduced — a new entry renders below a trailing comment block; intent.go:190 still appends to the end of the file.

### Raised

- **BR-37** [Important] `error-misattribution` A converged no-op still pushes an empty commit to origin/main whose subject claims the edit it did not make
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
- **BR-38** [Minor] `vacuous-test-guard` The "Apply must refuse too (defence in depth)" rows pass for an unrelated reason on move and remove, because Apply validates only OpAdd
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

## Open findings

- **BR-11** [Minor] `error-misattribution` offlineError labels every fetch failure "unreachable (offline?)"
- **BR-12** [Minor] `gitattributes-source` hash-object --path resolves .gitattributes from the working tree, not from the trunk being written
- **BR-21** [Important] `duplicated-helper` readLocal and readRef are the same three-line body with a different ref, and the guard exists on only one of them
- **BR-25** [Important] `unspecified-fields-not-preserved` Converge replaces the whole line, silently flipping kind and dropping the tag, while the note claims only the why-now moved
- **BR-29** [Minor] `duplicated-helper` pathPresent/modeOf and refPresent/resolve are each one function split in two, and Update runs both halves of both pairs per attempt
- **BR-30** [Minor] `vacuous-test-guard` fakeTrunk calls the transform exactly once and fires its peer before it, so two tests are named for properties the double cannot produce
- **BR-31** [Minor] `plan-traceability` Core-concepts table names queueCmd where the symbol is NewQueueCmd, and gitx.FirstLine has no row
- **BR-34** [Important] `unguarded-write-verb` The spine guard on the queue write verbs is pinned by no test, and repoguard.go's enumeration claim is now false
- **BR-35** [Minor] `duplicated-helper` The item-marker predicate is written twice with different rules, so indented prose bullets are reported as unparsed entries
- **BR-36** [Minor] `record-region-boundary` add appends at the end of the FILE, so a new entry lands below a trailing comment block, and move --after disagrees with it
- **BR-37** [Important] `error-misattribution` A converged no-op still pushes an empty commit to origin/main whose subject claims the edit it did not make
- **BR-38** [Minor] `vacuous-test-guard` The "Apply must refuse too (defence in depth)" rows pass for an unrelated reason on move and remove, because Apply validates only OpAdd
