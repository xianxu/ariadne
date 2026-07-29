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
    - "n": 2
      timestamp: "2026-07-29T16:45:03-07:00"
      agent: claude
      dispose:
        - id: PQ-1
          disposition: addressed
          note: Task 4 retires refScanRE; issueref.ScanRE adopts the pinned [0-9]{1,6} and IsLocal stays exact — both divergences settled explicitly.
          round: 2
        - id: PQ-2
          disposition: addressed
          note: selfQualifier now derives from opts.GitRepo (commit.go:42), and Options.RepoName plus the CLI plumbing are gone.
          round: 2
        - id: PQ-3
          disposition: addressed
          note: Invocation verified against activetime.go:29-41/:224/:227 plus --include-assistant, with a like-for-like BEFORE run.
          round: 2
        - id: PQ-4
          disposition: addressed
          note: Disposed as already satisfied via AttributionWarning.Reason (compute.go:68, activetime.go:100-105), and Task 3 Step 4 adds the foreign-refs-ignored warning.
          round: 2
        - id: PQ-5
          disposition: addressed
          note: Compressed to signatures plus a strategy line; commit messages dropped; corpus table and mutation-verify retained.
          round: 2
      findings:
        - id: PQ-6
          severity: Minor
          title: ARCH-DRY — spanRefRE (migrate.go:55) is a fifth encoding of the same qualifier+id grammar, so "4 → 1" is not the real count
          detail: |-
            `^([A-Za-z0-9][A-Za-z0-9_.-]*)?#[0-9]{1,6}( M[0-9]+[a-z]?)?$` restates the grammar in
            anchored form. After Task 4 the count is 5 → 2. Compose both from one exported
            qualifier+id fragment, or state why the anchored variant stays separate.
          round: 2
        - id: PQ-7
          severity: Minor
          title: The self-qualified-is-local case is untestable in gitx because RepoTopLevel bypasses the run shim (window.go:524)
          detail: |-
            selfRepoName() calls RepoTopLevel, which shells out via exec.Command directly, so the new
            test can only assert the foreign case. A "" or wrong basename would silently drop
            ariadne#180 — the plan's own must-not-regress row — with no guard. Pass selfRepo as a
            parameter, or route RepoTopLevel through run.
          round: 2
        - id: PQ-8
          severity: Minor
          title: The issue's Plan rows are the superseded ones, so Task 6 Step 3's "tick every Plan row" would record dropped work as done
          detail: |-
            The issue file still lists the precedence rule and the terminal-status guard, both
            explicitly dropped by the Revisions ledger. Rewrite the issue's Plan to mirror Tasks 1-6
            before the close, so the plan-unchecked gate is satisfied honestly.
          round: 2
      blocked: false
content_hash: 0383d183aef77374384439b2c0861f8f39f3ff983fa2a25c7bb22d83464f0715
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

## Round 2 — 2026-07-29T16:45:03-07:00 (claude) — passed

### Disposed

- PQ-1 — addressed — Task 4 retires refScanRE; issueref.ScanRE adopts the pinned [0-9]{1,6} and IsLocal stays exact — both divergences settled explicitly.
- PQ-2 — addressed — selfQualifier now derives from opts.GitRepo (commit.go:42), and Options.RepoName plus the CLI plumbing are gone.
- PQ-3 — addressed — Invocation verified against activetime.go:29-41/:224/:227 plus --include-assistant, with a like-for-like BEFORE run.
- PQ-4 — addressed — Disposed as already satisfied via AttributionWarning.Reason (compute.go:68, activetime.go:100-105), and Task 3 Step 4 adds the foreign-refs-ignored warning.
- PQ-5 — addressed — Compressed to signatures plus a strategy line; commit messages dropped; corpus table and mutation-verify retained.

### Raised

- **PQ-6** [Minor] ARCH-DRY — spanRefRE (migrate.go:55) is a fifth encoding of the same qualifier+id grammar, so "4 → 1" is not the real count
  `^([A-Za-z0-9][A-Za-z0-9_.-]*)?#[0-9]{1,6}( M[0-9]+[a-z]?)?$` restates the grammar in
  anchored form. After Task 4 the count is 5 → 2. Compose both from one exported
  qualifier+id fragment, or state why the anchored variant stays separate.
- **PQ-7** [Minor] The self-qualified-is-local case is untestable in gitx because RepoTopLevel bypasses the run shim (window.go:524)
  selfRepoName() calls RepoTopLevel, which shells out via exec.Command directly, so the new
  test can only assert the foreign case. A "" or wrong basename would silently drop
  ariadne#180 — the plan's own must-not-regress row — with no guard. Pass selfRepo as a
  parameter, or route RepoTopLevel through run.
- **PQ-8** [Minor] The issue's Plan rows are the superseded ones, so Task 6 Step 3's "tick every Plan row" would record dropped work as done
  The issue file still lists the precedence rule and the terminal-status guard, both
  explicitly dropped by the Revisions ledger. Rewrite the issue's Plan to mirror Tasks 1-6
  before the close, so the plan-unchecked gate is satisfied honestly.

## Open findings

- **PQ-6** [Minor] ARCH-DRY — spanRefRE (migrate.go:55) is a fifth encoding of the same qualifier+id grammar, so "4 → 1" is not the real count
- **PQ-7** [Minor] The self-qualified-is-local case is untestable in gitx because RepoTopLevel bypasses the run shim (window.go:524)
- **PQ-8** [Minor] The issue's Plan rows are the superseded ones, so Task 6 Step 3's "tick every Plan row" would record dropped work as done
