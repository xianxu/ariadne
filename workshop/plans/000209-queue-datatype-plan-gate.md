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

## Open findings

- **PQ-1** [Critical] `unverified-seam-capability` M1's plumbing cannot be built on the seam the plan names — import cycle, and no seam carries GIT_INDEX_FILE
- **PQ-2** [Important] `undeclared-degraded-read-policy` No operating envelope for the fetch on every queue read — offline, no-origin, and latency are unstated
- **PQ-3** [Important] `donewhen-clause-without-task` The wholesale-hand-edit interleaving row has no task, no test, and no implementation site
- **PQ-4** [Important] `parser-lacks-mechanical-guard` Doc.Parse/Render is guarded by one handcrafted doc, not a mechanical round-trip guard
- **PQ-5** [Minor] `vacuous-verification-step` Task 11's git diff --stat check on construct/generated/ is a no-op — that path is gitignored
- **PQ-6** [Minor] `unvalidated-input-breaks-format` why-now is free text written straight into the line format with no validation
