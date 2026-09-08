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

## Open findings

- **BR-1** [Important] `vacuous-test-guard` Temp-index assertions in TestTrunkFile_TransformErrorLeavesNoIndexAndNoCommit are unreachable
- **BR-2** [Important] `git-output-string-matching` isNonFastForward matches bare "rejected", retrying non-retryable push refusals
- **BR-3** [Important] `git-output-string-matching` isMissingPath carries a phrase git no longer emits, and the no-remote fallback silently depends on it
- **BR-4** [Important] `git-config-bool` signs() compares the raw config string to "true", so commit.gpgsign = 1|yes|on yields unsigned commits
- **BR-5** [Important] `untested-first-write` No test for Update on a path absent from the trunk — the exact path M2's seed takes
- **BR-6** [Minor] `dead-parameter` tempIndexPath(dir string) never uses dir
- **BR-7** [Minor] `temp-file-race` os.CreateTemp then os.Remove leaves a symlink-plant window on the index path
- **BR-8** [Minor] `duplicated-helper` firstLineOf duplicates firstLine in cmd/sdlc/issueids.go:112
- **BR-9** [Minor] `stale-comment` Test comment claims runGitIn uses CombinedOutput, which it deliberately does not
- **BR-10** [Minor] `pure-logic-untested` isNonFastForward, isMissingPath and firstLineOf are pure but only exercised through integration tests
- **BR-11** [Minor] `error-misattribution` offlineError labels every fetch failure "unreachable (offline?)"
- **BR-12** [Minor] `gitattributes-source` hash-object --path resolves .gitattributes from the working tree, not from the trunk being written
- **BR-13** [Minor] `plan-traceability` Plan Tasks 1-6b left unchecked and the plan has no Revisions section despite mid-stream edits
