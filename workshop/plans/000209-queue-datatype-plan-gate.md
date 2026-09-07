---
gate: plan-quality
issue: 209
id_prefix: PQ
rounds:
    - "n": 1
      timestamp: "2026-09-07T16:46:41-07:00"
      agent: claude
      findings:
        - id: PQ-1
          severity: Critical
          title: M1's plumbing cannot be built on the seam the plan names — import cycle, and no seam carries GIT_INDEX_FILE
          detail: |-
            Task 1's test calls gitx.NewTrunkFile(execGitRunner{}, ...) but execGitRunner is
            package main (cmd/sdlc/runner.go:12) and internal/gitx is imported BY main
            (cmd/sdlc/close.go:39) — an import cycle that will not compile; gitx's own git
            access is the package-level run shim (internal/gitx/window.go:32), a different
            seam. Separately, Task 2 Step 3 needs GIT_INDEX_FILE on read-tree/update-index/
            write-tree, and neither gitRunner.Git/GitInDir (runner.go:34-45) nor gitx.run
            accepts env; the only cmd.Env in the binary is judge/dispatch.go:97, for a
            non-git child. read-tree has --index-output but update-index/write-tree have no
            equivalent, so env is unavoidable. State the package placement and the seam
            extension — this changes an interface used across ~20 files.
          family: unverified-seam-capability
          round: 1
        - id: PQ-2
          severity: Important
          title: No operating envelope for the fetch on every queue read — offline, no-origin, and latency are unstated
          detail: |-
            ARCH-CONSTRAINTS: every sdlc queue invocation including the bare list does a
            network fetch, with no budget, no offline policy, and no behavior named for a
            repo with no origin remote. The repo already settled this at issueids.go:40-49
            and 126-145 — offline is not a refusal, degrade to the stale ref and announce it
            LOUDLY. Adopt that policy (ARCH-DRY) or state why queue diverges, and cover it
            in Task 1 or Task 9.
          family: undeclared-degraded-read-policy
          round: 1
        - id: PQ-3
          severity: Important
          title: The wholesale-hand-edit interleaving row has no task, no test, and no implementation site
          detail: |-
            Done-when requires a test for every row of the Spec's interleaving table, but
            the "Wholesale hand-edit of the file / Remote wins; report the drop" row is
            absent from Task 8's case table and from every other task. It is also
            undesigned: nothing in this architecture ever reads the working-tree copy, so
            there is no point at which the drop is detectable. Either name the task that
            compares the worktree copy against the trunk and warns, or drop the row from the
            Spec so Done-when stops promising it.
          family: donewhen-clause-without-task
          round: 1
        - id: PQ-4
          severity: Important
          title: Doc.Parse/Render is guarded by one handcrafted doc, not a mechanical round-trip guard
          detail: |-
            Doc parses a human-editable file arriving from origin/main — input this process
            did not produce (ARCH-SECURE). The adversarial class is arbitrary bytes: no
            trailing newline, CRLF, a near-miss line, unicode vs ASCII dashes, a bare "- " at
            EOF. Add FuzzDocRoundTrip asserting Render(Parse(b)) == b, seeded with those
            malformed forms; that is what makes Task 6 Step 3's "a malformed line is kept
            as-is, never discarded" true across the whole input space rather than for four
            chosen inputs.
          family: parser-lacks-mechanical-guard
          round: 1
        - id: PQ-5
          severity: Minor
          title: Task 11's git diff --stat check on construct/generated/ is a no-op — that path is gitignored
          detail: |-
            .gitignore:30 ignores /construct/generated/ (confirmed via git check-ignore), so
            the diff prints nothing whether or not a hand-edit occurred and the step passes
            vacuously. Per atlas/workflow/data-artifacts.md:20 the generated SKILL.md is
            per-repo and never committed; the file a hand-edit would wrongly touch is
            cmd/datatype/SKILL.md.tmpl. Keep the grep for queue; point the diff check at
            cmd/datatype/ and construct/local/datatype/ instead.
          family: vacuous-verification-step
          round: 1
        - id: PQ-6
          severity: Minor
          title: why-now is free text written straight into the line format with no validation
          detail: |-
            A newline in the why-now argument silently becomes a second queue entry and
            breaks Doc's round-trip invariant; an embedded " — " or "[...]" re-parses as a
            different why/tag split. Task 8 already rejects a malformed ref before any git
            call — say the same for WhyNow (reject control characters and newlines, or state
            the escaping).
          family: unvalidated-input-breaks-format
          round: 1
      blocked: true
    - "n": 2
      timestamp: "2026-09-07T16:50:36-07:00"
      agent: claude
      dispose:
        - id: PQ-1
          disposition: addressed
          note: TrunkFile sited in gitx on its own run/runEnv shims; gitx imports only issueref, so no cycle; gitRunner left untouched.
          round: 2
        - id: PQ-2
          disposition: addressed
          note: Adopts issueids.go's offline policy in shape — read degrades loudly, write refuses; Task 5 covers no-origin.
          round: 2
        - id: PQ-3
          disposition: addressed
          note: Row deleted from the Spec with the undesignability recorded; Done-when no longer promises it.
          round: 2
        - id: PQ-4
          disposition: addressed
          note: Task 8 Step 5 adds FuzzDocRoundTrip over arbitrary bytes with the malformed seeds.
          round: 2
        - id: PQ-5
          disposition: addressed
          note: Verified .gitignore:30 via check-ignore; check repointed at cmd/datatype/ and construct/local/datatype/.
          round: 2
        - id: PQ-6
          disposition: addressed
          note: why-now validated before any git call; rejection rather than escaping, with the reason stated.
          round: 2
      blocked: false
    - "n": 3
      timestamp: "2026-09-07T16:55:01-07:00"
      agent: claude
      dispose:
        - id: PQ-1
          disposition: addressed
          note: Cycle resolved by keeping TrunkFile in leaf package gitx with its own run/runEnv shims; verified main imports gitx at actual.go:26 and execGitRunner is package main at runner.go:34.
          round: 3
        - id: PQ-2
          disposition: addressed
          note: Read degrades to the stale tracking ref, write refuses, no-origin says so; mirrors issueids.go:126-145 rather than inventing a second policy.
          round: 3
        - id: PQ-3
          disposition: addressed
          note: The undesignable hand-edit row is deleted and the reasoning recorded in the Spec instead of a fabricated worktree-comparison task.
          round: 3
        - id: PQ-4
          disposition: addressed
          note: Task 8 Step 5 adds FuzzDocRoundTrip with an adversarial seed corpus and a named fuzztime run.
          round: 3
        - id: PQ-5
          disposition: addressed
          note: Repointed at cmd/datatype/SKILL.md.tmpl; confirmed /construct/generated/ is .gitignore:30 and .dynamic-skill is the only tracked file under construct/local/datatype/.
          round: 3
        - id: PQ-6
          disposition: addressed
          note: why-now newline/control-char/em-dash/bracket rejection is now a Done-when clause plus three Task 9 table rows, validated before any git call.
          round: 3
      findings:
        - id: PQ-7
          severity: Critical
          title: gitx.run carries argv only — the plan's TrunkFile still needs Dir and stderr from it, and Task 1/2/4 would run git against the real ariadne repo
          detail: 'This is the 3rd instance in family unverified-seam-capability; PQ-1 fixed the Env instance. Do not patch Task 1''s sample — state the rule in the plan (enumerate every exec.Cmd parameter the plumbing needs against the shim''s signature at file:line before naming it the seam) and sweep the whole enumeration this round. window.go:32-34 defines run as exec.Command(name, args...).Output(): no Dir, no Env, stderr dropped into ExitError.Stderr. The plan''s NewTrunkFile(repo, "origin", "main") never names -C or a Dir-carrying shim, and Task 1''s test does not chdir (contrast window_test.go:24,117 which use testfix.Chdir), so Read fetches from and Update pushes <commit>:main to the REAL ariadne origin, overwriting real workshop/queue.md on real main with fixture bytes. Separately, the "surface the last rejection" (Task 4), "warning text names the risk" (Task 5) and "print the trunk state and the intent" (Task 11) clauses need git''s stderr; the cited ARCH-DRY precedent issueids.go:126-145 gets it free via execGitRunner.Git''s CombinedOutput (runner.go:36-38), which .Output() does not provide. Measured prevalence: 3 of 6 enumerable exec.Cmd parameters (Env fixed, Dir open, stderr open). If the chosen fix is -C, also state that the GIT_INDEX_FILE path must be absolute, since it resolves against the process cwd rather than -C.'
          family: unverified-seam-capability
          round: 3
      blocked: true
content_hash: 17f2c457676bedd3e7dfe99b2d9bd14966c5957394b7f6b9546b85acacd52413
---

# Gate ledger — ariadne#209 (plan-quality)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-09-07T16:46:41-07:00 (claude) — BLOCKED

### Raised

- **PQ-1** [Critical] `unverified-seam-capability` M1's plumbing cannot be built on the seam the plan names — import cycle, and no seam carries GIT_INDEX_FILE
  Task 1's test calls gitx.NewTrunkFile(execGitRunner{}, ...) but execGitRunner is
  package main (cmd/sdlc/runner.go:12) and internal/gitx is imported BY main
  (cmd/sdlc/close.go:39) — an import cycle that will not compile; gitx's own git
  access is the package-level run shim (internal/gitx/window.go:32), a different
  seam. Separately, Task 2 Step 3 needs GIT_INDEX_FILE on read-tree/update-index/
  write-tree, and neither gitRunner.Git/GitInDir (runner.go:34-45) nor gitx.run
  accepts env; the only cmd.Env in the binary is judge/dispatch.go:97, for a
  non-git child. read-tree has --index-output but update-index/write-tree have no
  equivalent, so env is unavoidable. State the package placement and the seam
  extension — this changes an interface used across ~20 files.
- **PQ-2** [Important] `undeclared-degraded-read-policy` No operating envelope for the fetch on every queue read — offline, no-origin, and latency are unstated
  ARCH-CONSTRAINTS: every sdlc queue invocation including the bare list does a
  network fetch, with no budget, no offline policy, and no behavior named for a
  repo with no origin remote. The repo already settled this at issueids.go:40-49
  and 126-145 — offline is not a refusal, degrade to the stale ref and announce it
  LOUDLY. Adopt that policy (ARCH-DRY) or state why queue diverges, and cover it
  in Task 1 or Task 9.
- **PQ-3** [Important] `donewhen-clause-without-task` The wholesale-hand-edit interleaving row has no task, no test, and no implementation site
  Done-when requires a test for every row of the Spec's interleaving table, but
  the "Wholesale hand-edit of the file / Remote wins; report the drop" row is
  absent from Task 8's case table and from every other task. It is also
  undesigned: nothing in this architecture ever reads the working-tree copy, so
  there is no point at which the drop is detectable. Either name the task that
  compares the worktree copy against the trunk and warns, or drop the row from the
  Spec so Done-when stops promising it.
- **PQ-4** [Important] `parser-lacks-mechanical-guard` Doc.Parse/Render is guarded by one handcrafted doc, not a mechanical round-trip guard
  Doc parses a human-editable file arriving from origin/main — input this process
  did not produce (ARCH-SECURE). The adversarial class is arbitrary bytes: no
  trailing newline, CRLF, a near-miss line, unicode vs ASCII dashes, a bare "- " at
  EOF. Add FuzzDocRoundTrip asserting Render(Parse(b)) == b, seeded with those
  malformed forms; that is what makes Task 6 Step 3's "a malformed line is kept
  as-is, never discarded" true across the whole input space rather than for four
  chosen inputs.
- **PQ-5** [Minor] `vacuous-verification-step` Task 11's git diff --stat check on construct/generated/ is a no-op — that path is gitignored
  .gitignore:30 ignores /construct/generated/ (confirmed via git check-ignore), so
  the diff prints nothing whether or not a hand-edit occurred and the step passes
  vacuously. Per atlas/workflow/data-artifacts.md:20 the generated SKILL.md is
  per-repo and never committed; the file a hand-edit would wrongly touch is
  cmd/datatype/SKILL.md.tmpl. Keep the grep for queue; point the diff check at
  cmd/datatype/ and construct/local/datatype/ instead.
- **PQ-6** [Minor] `unvalidated-input-breaks-format` why-now is free text written straight into the line format with no validation
  A newline in the why-now argument silently becomes a second queue entry and
  breaks Doc's round-trip invariant; an embedded " — " or "[...]" re-parses as a
  different why/tag split. Task 8 already rejects a malformed ref before any git
  call — say the same for WhyNow (reject control characters and newlines, or state
  the escaping).

## Round 2 — 2026-09-07T16:50:36-07:00 (claude) — passed

### Disposed

- PQ-1 — addressed — TrunkFile sited in gitx on its own run/runEnv shims; gitx imports only issueref, so no cycle; gitRunner left untouched.
- PQ-2 — addressed — Adopts issueids.go's offline policy in shape — read degrades loudly, write refuses; Task 5 covers no-origin.
- PQ-3 — addressed — Row deleted from the Spec with the undesignability recorded; Done-when no longer promises it.
- PQ-4 — addressed — Task 8 Step 5 adds FuzzDocRoundTrip over arbitrary bytes with the malformed seeds.
- PQ-5 — addressed — Verified .gitignore:30 via check-ignore; check repointed at cmd/datatype/ and construct/local/datatype/.
- PQ-6 — addressed — why-now validated before any git call; rejection rather than escaping, with the reason stated.

## Round 3 — 2026-09-07T16:55:01-07:00 (claude) — BLOCKED

### Disposed

- PQ-1 — addressed — Cycle resolved by keeping TrunkFile in leaf package gitx with its own run/runEnv shims; verified main imports gitx at actual.go:26 and execGitRunner is package main at runner.go:34.
- PQ-2 — addressed — Read degrades to the stale tracking ref, write refuses, no-origin says so; mirrors issueids.go:126-145 rather than inventing a second policy.
- PQ-3 — addressed — The undesignable hand-edit row is deleted and the reasoning recorded in the Spec instead of a fabricated worktree-comparison task.
- PQ-4 — addressed — Task 8 Step 5 adds FuzzDocRoundTrip with an adversarial seed corpus and a named fuzztime run.
- PQ-5 — addressed — Repointed at cmd/datatype/SKILL.md.tmpl; confirmed /construct/generated/ is .gitignore:30 and .dynamic-skill is the only tracked file under construct/local/datatype/.
- PQ-6 — addressed — why-now newline/control-char/em-dash/bracket rejection is now a Done-when clause plus three Task 9 table rows, validated before any git call.

### Raised

- **PQ-7** [Critical] `unverified-seam-capability` gitx.run carries argv only — the plan's TrunkFile still needs Dir and stderr from it, and Task 1/2/4 would run git against the real ariadne repo
  This is the 3rd instance in family unverified-seam-capability; PQ-1 fixed the Env instance. Do not patch Task 1's sample — state the rule in the plan (enumerate every exec.Cmd parameter the plumbing needs against the shim's signature at file:line before naming it the seam) and sweep the whole enumeration this round. window.go:32-34 defines run as exec.Command(name, args...).Output(): no Dir, no Env, stderr dropped into ExitError.Stderr. The plan's NewTrunkFile(repo, "origin", "main") never names -C or a Dir-carrying shim, and Task 1's test does not chdir (contrast window_test.go:24,117 which use testfix.Chdir), so Read fetches from and Update pushes <commit>:main to the REAL ariadne origin, overwriting real workshop/queue.md on real main with fixture bytes. Separately, the "surface the last rejection" (Task 4), "warning text names the risk" (Task 5) and "print the trunk state and the intent" (Task 11) clauses need git's stderr; the cited ARCH-DRY precedent issueids.go:126-145 gets it free via execGitRunner.Git's CombinedOutput (runner.go:36-38), which .Output() does not provide. Measured prevalence: 3 of 6 enumerable exec.Cmd parameters (Env fixed, Dir open, stderr open). If the chosen fix is -C, also state that the GIT_INDEX_FILE path must be absolute, since it resolves against the process cwd rather than -C.

## Open findings

- **PQ-7** [Critical] `unverified-seam-capability` gitx.run carries argv only — the plan's TrunkFile still needs Dir and stderr from it, and Task 1/2/4 would run git against the real ariadne repo
