---
gate: plan-quality
issue: 190
id_prefix: PQ
rounds:
    - "n": 1
      timestamp: "2026-07-29T16:36:13-07:00"
      agent: claude
      findings:
        - id: PQ-1
          severity: Important
          title: ARCH-DRY — the repo-qualified ref grammar already exists twice; the plan adds a third encoding while claiming to consolidate
          detail: |-
            migrate.go:45 refScanRE is `([A-Za-z0-9][A-Za-z0-9_.-]*)?#([0-9]{1,6})\b` — identical to
            the proposed issueref pattern — and its doc defers to parseRef (resolve.go:57), the
            canonical grammar documented as MUST NOT be re-encoded. Since both live in package main
            they can import internal/issueref, so make issueref the single source (refScanRE derived
            from it) or state why the active-time path keeps its own scanner. Also settle two
            divergences: `\d+` vs the pinned `[0-9]{1,6}` bound, and exact-match IsLocal vs
            resolveRepoDir's prefix matching (parley to parley.nvim).
          round: 1
        - id: PQ-2
          severity: Important
          title: Options.RepoName is derived from the process cwd but the commit path parses opts.GitRepo, which the standalone verb takes as a flag
          detail: |-
            Task 3 Step 4 sets RepoName from repoIdentity() (cwd git root), while loadWindowCommits
            reads commits from opts.GitRepo — supplied by --git-repo at activetime.go:203/221. With
            --git-repo pointed at a peer, that peer's own self-qualified refs are dropped as foreign
            and ariadne-qualified refs are admitted as local, reproducing the #190 bug class in the
            diagnostic verb. sdlc actual is unaffected (actual.go:110 passes repoTop), so testing
            will not surface it. Default the qualifier to filepath.Base(opts.GitRepo) when RepoName
            is empty, and state which repo the qualifier names.
          round: 1
        - id: PQ-3
          severity: Important
          title: Task 4's regression command exits 2 before computing — wrong flag names and two required flags missing
          detail: |-
            sdlc active-time requires --dir, --issue and --git-repo (activetime.go:29-41); the
            command supplies none of the first or third. The flag is repeatable --issue
            (activetime.go:224), not --issues with a comma list, and --threshold is --threshold-min
            (:227). Fix the invocation and name the transcript dirs, or state the sdlc actual
            fallback for the archived issue file concretely.
          round: 1
        - id: PQ-4
          severity: Important
          title: Done-when bullet on per-segment attribution-rule output is neither planned nor explicitly dropped
          detail: |-
            The Revisions ledger lists retained and dropped Spec items but omits "sdlc actual output
            states which rule attributed each segment", and no task covers it. If the existing
            "mention fallback without issue commit boundary" warning (compute.go:136-148) already
            satisfies it, dispose it in the Revisions section; otherwise add a task.
          round: 1
        - id: PQ-5
          severity: Minor
          title: Plan restates the diff — full implementation source and five pre-written commit messages
          detail: |-
            The corpus-derived test table and the Step 5 mutation-verify earn their place. The
            complete ref.go body and verbatim commit messages will be rewritten during
            implementation and are stale on arrival; compress to the signature plus the strategy
            line per function.
          round: 1
      blocked: true
---

# Gate ledger — ariadne#190 (plan-quality)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-07-29T16:36:13-07:00 (claude) — BLOCKED

### Raised

- **PQ-1** [Important] ARCH-DRY — the repo-qualified ref grammar already exists twice; the plan adds a third encoding while claiming to consolidate
  migrate.go:45 refScanRE is `([A-Za-z0-9][A-Za-z0-9_.-]*)?#([0-9]{1,6})\b` — identical to
  the proposed issueref pattern — and its doc defers to parseRef (resolve.go:57), the
  canonical grammar documented as MUST NOT be re-encoded. Since both live in package main
  they can import internal/issueref, so make issueref the single source (refScanRE derived
  from it) or state why the active-time path keeps its own scanner. Also settle two
  divergences: `\d+` vs the pinned `[0-9]{1,6}` bound, and exact-match IsLocal vs
  resolveRepoDir's prefix matching (parley to parley.nvim).
- **PQ-2** [Important] Options.RepoName is derived from the process cwd but the commit path parses opts.GitRepo, which the standalone verb takes as a flag
  Task 3 Step 4 sets RepoName from repoIdentity() (cwd git root), while loadWindowCommits
  reads commits from opts.GitRepo — supplied by --git-repo at activetime.go:203/221. With
  --git-repo pointed at a peer, that peer's own self-qualified refs are dropped as foreign
  and ariadne-qualified refs are admitted as local, reproducing the #190 bug class in the
  diagnostic verb. sdlc actual is unaffected (actual.go:110 passes repoTop), so testing
  will not surface it. Default the qualifier to filepath.Base(opts.GitRepo) when RepoName
  is empty, and state which repo the qualifier names.
- **PQ-3** [Important] Task 4's regression command exits 2 before computing — wrong flag names and two required flags missing
  sdlc active-time requires --dir, --issue and --git-repo (activetime.go:29-41); the
  command supplies none of the first or third. The flag is repeatable --issue
  (activetime.go:224), not --issues with a comma list, and --threshold is --threshold-min
  (:227). Fix the invocation and name the transcript dirs, or state the sdlc actual
  fallback for the archived issue file concretely.
- **PQ-4** [Important] Done-when bullet on per-segment attribution-rule output is neither planned nor explicitly dropped
  The Revisions ledger lists retained and dropped Spec items but omits "sdlc actual output
  states which rule attributed each segment", and no task covers it. If the existing
  "mention fallback without issue commit boundary" warning (compute.go:136-148) already
  satisfies it, dispose it in the Revisions section; otherwise add a task.
- **PQ-5** [Minor] Plan restates the diff — full implementation source and five pre-written commit messages
  The corpus-derived test table and the Step 5 mutation-verify earn their place. The
  complete ref.go body and verbatim commit messages will be rewritten during
  implementation and are stale on arrival; compress to the signature plus the strategy
  line per function.

## Open findings

- **PQ-1** [Important] ARCH-DRY — the repo-qualified ref grammar already exists twice; the plan adds a third encoding while claiming to consolidate
- **PQ-2** [Important] Options.RepoName is derived from the process cwd but the commit path parses opts.GitRepo, which the standalone verb takes as a flag
- **PQ-3** [Important] Task 4's regression command exits 2 before computing — wrong flag names and two required flags missing
- **PQ-4** [Important] Done-when bullet on per-segment attribution-rule output is neither planned nor explicitly dropped
- **PQ-5** [Minor] Plan restates the diff — full implementation source and five pre-written commit messages
