---
gate: boundary-review
issue: 162
id_prefix: BR
rounds:
    - "n": 1
      timestamp: "2026-08-26T22:24:07-07:00"
      agent: codex
      findings:
        - id: BR-1
          severity: Critical
          title: Absolute tracker-directory overrides do not exclude tracker files
          detail: cmd/sdlc/internal/judge/reviewwindow.go:115 inserts IssuesDir and HistoryDir directly into Git pathspecs. Git matches tracked paths relative to the repository, so absolute in-repository values from --issues-dir, --history-dir, or their environment variables fail to exclude those files. Running the rendered shape against this pinned range included workshop/issues/000162-milestone-close-window-base.md, while the relative-path recipe excluded it. Normalize in-repository exclusion directories relative to RepoRoot before rendering, and pin both directory fields with an executable real-Git regression. This violates ARCH-PURPOSE because the manifest does not preserve the promised exclusions.
          family: repo-relative-pathspecs
          round: 1
        - id: BR-2
          severity: Important
          title: Live-Git conformance does not execute the rendered review recipes
          detail: cmd/sdlc/reviewwindow_test.go:138-167 checks only ref resolution and mode fields. It never executes the rendered committed or working-tree commands, nor creates the promised staged, unstaged, untracked, and custom-directory states. Consequently the checked Task 2 Step 4 and ARCH-MOCK conformance claim are not delivered. Execute the structured argv against a temporary repository and assert range contents and exclusions.
          family: external-command-conformance
          round: 1
      blocked: true
    - "n": 2
      timestamp: "2026-08-26T22:38:08-07:00"
      agent: codex
      dispose:
        - id: BR-1
          disposition: addressed
          note: Both tracker directories are normalized relative to the repository, absolute or escaping renderer inputs are rejected, and removing normalization makes the real-Git regression fail.
          round: 2
        - id: BR-2
          disposition: not-addressed
          note: The test executes all recipes, but only names and full assert included paths; empty stat and targeted recipes still pass.
          round: 2
      findings:
        - id: BR-3
          severity: Important
          title: README omits the new manual milestone-review flag
          detail: cmd/sdlc/judge.go:64 adds the user-facing --plans-dir flag, but README.md is unchanged. Add the manual milestone-review invocation and its plan-directory behavior, or a direct README entry pointing readers to that documented surface.
          family: readme-cli-surface
          round: 2
      blocked: true
    - "n": 3
      timestamp: "2026-08-26T22:47:41-07:00"
      agent: codex
      dispose:
        - id: BR-2
          disposition: addressed
          note: All four rendered recipes execute against real committed and working-tree states; an empty-range stat mutant makes the conformance test fail.
          round: 3
        - id: BR-3
          disposition: not-addressed
          note: README.md documents the flag, but no automated test reads or asserts that documentation, so removing the claimed fix leaves the suite green.
          round: 3
      blocked: true
    - "n": 4
      timestamp: "2026-08-26T23:01:57-07:00"
      agent: codex
      dispose:
        - id: BR-1
          disposition: addressed
          note: Removing tracker-directory normalization makes the live-Git absolute-override regression test fail.
          round: 4
        - id: BR-2
          disposition: addressed
          note: Mutating stat and targeted recipes to empty ranges makes their positive-semantic assertions fail.
          round: 4
        - id: BR-3
          disposition: addressed
          note: Removing the README manual boundary-review section makes the scoped repository contract test fail.
          round: 4
      findings:
        - id: BR-4
          severity: Critical
          title: Durable plan changes do not invalidate an in-flight boundary review
          detail: The manifest names the optional plan as reviewer input, but closeReviewSnapshot captures only issue and project text. A blocked-review integration reproduction changed the plan and close still finalized; snapshot and revalidate every mutable reviewer input, including plan presence and contents.
          family: unlocked-review-snapshot-completeness
          round: 4
      blocked: true
    - "n": 5
      timestamp: "2026-08-26T23:18:31-07:00"
      agent: codex
      dispose:
        - id: BR-4
          disposition: addressed
          note: Both close paths snapshot canonical-plan presence and contents; disabling plan capture makes all eight mutation cases fail.
          round: 5
      blocked: false
---

# Gate ledger — ariadne#162 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-08-26T22:24:07-07:00 (codex) — BLOCKED

### Raised

- **BR-1** [Critical] `repo-relative-pathspecs` Absolute tracker-directory overrides do not exclude tracker files
  cmd/sdlc/internal/judge/reviewwindow.go:115 inserts IssuesDir and HistoryDir directly into Git pathspecs. Git matches tracked paths relative to the repository, so absolute in-repository values from --issues-dir, --history-dir, or their environment variables fail to exclude those files. Running the rendered shape against this pinned range included workshop/issues/000162-milestone-close-window-base.md, while the relative-path recipe excluded it. Normalize in-repository exclusion directories relative to RepoRoot before rendering, and pin both directory fields with an executable real-Git regression. This violates ARCH-PURPOSE because the manifest does not preserve the promised exclusions.
- **BR-2** [Important] `external-command-conformance` Live-Git conformance does not execute the rendered review recipes
  cmd/sdlc/reviewwindow_test.go:138-167 checks only ref resolution and mode fields. It never executes the rendered committed or working-tree commands, nor creates the promised staged, unstaged, untracked, and custom-directory states. Consequently the checked Task 2 Step 4 and ARCH-MOCK conformance claim are not delivered. Execute the structured argv against a temporary repository and assert range contents and exclusions.

## Round 2 — 2026-08-26T22:38:08-07:00 (codex) — BLOCKED

### Disposed

- BR-1 — addressed — Both tracker directories are normalized relative to the repository, absolute or escaping renderer inputs are rejected, and removing normalization makes the real-Git regression fail.
- BR-2 — not-addressed — The test executes all recipes, but only names and full assert included paths; empty stat and targeted recipes still pass.

### Raised

- **BR-3** [Important] `readme-cli-surface` README omits the new manual milestone-review flag
  cmd/sdlc/judge.go:64 adds the user-facing --plans-dir flag, but README.md is unchanged. Add the manual milestone-review invocation and its plan-directory behavior, or a direct README entry pointing readers to that documented surface.

## Round 3 — 2026-08-26T22:47:41-07:00 (codex) — BLOCKED

### Disposed

- BR-2 — addressed — All four rendered recipes execute against real committed and working-tree states; an empty-range stat mutant makes the conformance test fail.
- BR-3 — not-addressed — README.md documents the flag, but no automated test reads or asserts that documentation, so removing the claimed fix leaves the suite green.

## Round 4 — 2026-08-26T23:01:57-07:00 (codex) — BLOCKED

### Disposed

- BR-1 — addressed — Removing tracker-directory normalization makes the live-Git absolute-override regression test fail.
- BR-2 — addressed — Mutating stat and targeted recipes to empty ranges makes their positive-semantic assertions fail.
- BR-3 — addressed — Removing the README manual boundary-review section makes the scoped repository contract test fail.

### Raised

- **BR-4** [Critical] `unlocked-review-snapshot-completeness` Durable plan changes do not invalidate an in-flight boundary review
  The manifest names the optional plan as reviewer input, but closeReviewSnapshot captures only issue and project text. A blocked-review integration reproduction changed the plan and close still finalized; snapshot and revalidate every mutable reviewer input, including plan presence and contents.

## Round 5 — 2026-08-26T23:18:31-07:00 (codex) — passed

### Disposed

- BR-4 — addressed — Both close paths snapshot canonical-plan presence and contents; disabling plan capture makes all eight mutation cases fail.

## Open findings

(none — every finding has been disposed)
