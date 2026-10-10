---
gate: boundary-review
issue: 270
id_prefix: BR
rounds:
    - "n": 1
      timestamp: "2026-10-09T20:26:25-07:00"
      agent: claude
      recipe: milestone-review
      blocked: false
      protocol_error: no valid findings block
    - "n": 2
      timestamp: "2026-10-09T20:30:42-07:00"
      agent: claude
      findings:
        - id: BR-1
          severity: Important
          title: sdlc-binary.md still says window-start is the earlier of parent and claim, and names deleted DiscoverWindowIssues
          detail: atlas/workflow/sdlc-binary.md around line 1005 says "the window-start is the earlier of" the two anchors, which the new paragraph contradicts, and lines 715 and 980 still name DiscoverWindowIssues, which this diff deletes. Rewrite the sentence to "claim when present, else parent" and name activetime.WindowIssues.
          family: atlas-stale-after-surface-change
          round: 2
        - id: BR-2
          severity: Minor
          title: gitx/window.go import block keeps a blank line before the closing paren
          family: gofmt-clean
          round: 2
        - id: BR-3
          severity: Minor
          title: standalone sdlc active-time builds Options without a Scope, so it can disagree with sdlc actual
          family: single-source-consumer-sweep
          round: 2
        - id: BR-4
          severity: Minor
          title: scope is off when HEAD equals the merge base, so other issues' tracker commits bound segments before the branch's first commit
          family: boundary-scope-consistency
          round: 2
        - id: BR-5
          severity: Minor
          title: loadWindowCommits walks all history twice per measurement; keeping --since as a committer-date prefilter would bound it
          family: unbounded-history-walk
          round: 2
        - id: BR-6
          severity: Minor
          title: no test covers the scope wiring in computeActual (MergeBaseWithMain into Options.Scope)
          family: glue-wiring-untested
          round: 2
      recipe: milestone-review
      blocked: true
    - "n": 3
      timestamp: "2026-10-09T20:48:38-07:00"
      agent: claude
      dispose:
        - id: BR-1
          disposition: addressed
          note: sdlc-binary.md:715/980/1005 now say claim-else-parent and name activetime.WindowIssues; grep finds no stale DiscoverWindowIssues/"earlier of" in atlas.
          round: 3
        - id: BR-2
          disposition: addressed
          note: gitx/window.go import block fixed; gofmt -l cmd/sdlc/ is clean.
          round: 3
        - id: BR-3
          disposition: not-addressed
          note: --branch-point wiring (activetime.go:222-227, incl. the needs---issue error) is a behavior change with no test exercising the flag; add a small cobra-level test.
          round: 3
        - id: BR-4
          disposition: addressed
          note: actualScope falls back to BranchPoint() on the issue's own undiverged branch; TestActualScope asserts it and goes red without the branch.
          round: 3
        - id: BR-5
          disposition: withdrawn
          note: 'Decline is correct: git --since walk-termination uses committer dates and would reintroduce the rebase sensitivity #270 removes; ~50ms cost is acceptable.'
          round: 3
        - id: BR-6
          disposition: addressed
          note: computeActual now derives scope via actualScope(issueNum), pinned by cmd/sdlc/actualscope_test.go.
          round: 3
      recipe: milestone-review
      blocked: false
---

# Gate ledger — ariadne#270 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-09T20:26:25-07:00 (claude) — passed

**Protocol error:** no valid findings block — this round contributed no findings.

## Round 2 — 2026-10-09T20:30:42-07:00 (claude) — BLOCKED

### Raised

- **BR-1** [Important] `atlas-stale-after-surface-change` sdlc-binary.md still says window-start is the earlier of parent and claim, and names deleted DiscoverWindowIssues
  atlas/workflow/sdlc-binary.md around line 1005 says "the window-start is the earlier of" the two anchors, which the new paragraph contradicts, and lines 715 and 980 still name DiscoverWindowIssues, which this diff deletes. Rewrite the sentence to "claim when present, else parent" and name activetime.WindowIssues.
- **BR-2** [Minor] `gofmt-clean` gitx/window.go import block keeps a blank line before the closing paren
- **BR-3** [Minor] `single-source-consumer-sweep` standalone sdlc active-time builds Options without a Scope, so it can disagree with sdlc actual
- **BR-4** [Minor] `boundary-scope-consistency` scope is off when HEAD equals the merge base, so other issues' tracker commits bound segments before the branch's first commit
- **BR-5** [Minor] `unbounded-history-walk` loadWindowCommits walks all history twice per measurement; keeping --since as a committer-date prefilter would bound it
- **BR-6** [Minor] `glue-wiring-untested` no test covers the scope wiring in computeActual (MergeBaseWithMain into Options.Scope)

## Round 3 — 2026-10-09T20:48:38-07:00 (claude) — passed

### Disposed

- BR-1 — addressed — sdlc-binary.md:715/980/1005 now say claim-else-parent and name activetime.WindowIssues; grep finds no stale DiscoverWindowIssues/"earlier of" in atlas.
- BR-2 — addressed — gitx/window.go import block fixed; gofmt -l cmd/sdlc/ is clean.
- BR-3 — not-addressed — --branch-point wiring (activetime.go:222-227, incl. the needs---issue error) is a behavior change with no test exercising the flag; add a small cobra-level test.
- BR-4 — addressed — actualScope falls back to BranchPoint() on the issue's own undiverged branch; TestActualScope asserts it and goes red without the branch.
- BR-5 — withdrawn — Decline is correct: git --since walk-termination uses committer dates and would reintroduce the rebase sensitivity #270 removes; ~50ms cost is acceptable.
- BR-6 — addressed — computeActual now derives scope via actualScope(issueNum), pinned by cmd/sdlc/actualscope_test.go.

## Open findings

- **BR-3** [Minor] `single-source-consumer-sweep` standalone sdlc active-time builds Options without a Scope, so it can disagree with sdlc actual
