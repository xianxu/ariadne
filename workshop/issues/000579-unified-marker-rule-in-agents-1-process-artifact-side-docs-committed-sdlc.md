---
id: 000579
status: codecomplete
deps: []
github_issue:
created: 2026-10-23
updated: 2026-10-23
estimate_hours:
collaboration-mode: collaboration
card_mirror: '9e7d385aeba81e688ce43b0360a13fad5b7b6518' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-23T16:38:06-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:5
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot5/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: full, provenance: inferred}
actual_hours: 1.70
---

# Unified marker rule in AGENTS §1; process-artifact side docs committed; sdlc refuses links to uncommitted files

## Problem

## Spec

**Side docs in process artifacts, and the unified marker rule (operator, 10-23).** Follows the single collaborative
mode (ariadne#575): in issues, plans and projects too, a `🤖[…]` question gets a side doc (the marker becomes
`→ [discussion](…)`), an instruction is acted on, and `🤖[fold in]` collapses a side doc back into the text.
Discussion: ops `session.side/2026-10-23-07-side-docs-in-process-artifacts.md`.

1. **AGENTS §1** (the shared constitution): replace "a 🤖[H] in any artifact … answer it, and resolve it in place
   (fold the answer into the prose)" with the unified rule above, plus: side docs of process artifacts live next to
   them (`<artifact>.side/…`), are committed with them, and archive with the issue. The operator's private session
   side docs (`session.side/`, git-ignored in ops) stay uncommitted: they're operational context.
2. **A guard in sdlc** where an issue's text reaches main (`issue publish`, `change-code`, `close`): refuse when a link
   in the issue, its plan or a referenced project file points at a file that isn't committed or is git-ignored,
   naming the link and the fix (commit the side doc, or fold it in).
3. Keep the review convention (`.agents/skills/xx-fix/review-convention.md`) consistent with 1.

## Done when

- AGENTS §1 and the review convention state the unified rule and commit-with-artifact (doc check).
- `sdlc issue publish`, `change-code` and `close` refuse a link to an uncommitted or ignored file and pass once it's
  committed or removed (tests).

## Plan

Durable plan: [000579 plan](../plans/000579-unified-marker-rule-in-agents-1-process-artifact-side-docs-committed-sdlc-plan.md).

- [x] The rule in prose: AGENTS.base.md §1, review convention §6, collaboration-modes §4
- [x] `sidedoc.LocalLinks` extracted; `RebaseLinks` over it
- [x] `linkguard` census; the side-links gate in change-code, close and issue publish
- [x] (D1) publish carries the side docs the details link to
- [x] atlas, targeted runs, recorded suite (the suite runs on the pushed head before merge, which requires it)

## Log

### 2026-10-23
- 2026-10-23: closed — BR-3 (Minor, advisory after the SHIP): TestGitLinkStateLiteralPath judges x*.md beside an ignored xy.md as uncommitted; removing :(literal) fails it (x*.md: 2, want 1), restored. Test-only delta at 5ead6cb6. (fixed-then-ship); review verdict: SHIP
- 2026-10-23: closed — Rule: AGENTS.base.md section 1, review-convention section 6, collaboration-modes section 4 (review export regenerated). Guard tests: TestCensus, TestLocalLinks, TestIssueSideLinkRefusals, TestIssueSideLinkRefusalsJudgeProjects (peer project refused; unborn peer warns), TestChangeCodeRefusesUncommittedSideLink (runChangeCode exit 1, --force ack, --no-side-links, pass once committed), TestCloseRefusesIgnoredSideLink (refuses; --no-side-links skips the check; passes once folded), TestIssuePublishRefusesUncommittedLink, TestIssuePublishCarriesLinkedSideDocs (D1). Round-1 BR-1/BR-2 and Minors fixed at 712b54c0. go test -run Publish|Asset|SideLink|ChangeCode|Close|GateCatalog|Archive pass but for sandbox-only host boot identity landing tests; review verdict: SHIP
- 2026-10-23: flow upgraded quick → full — 563 added lines in code files (limit 100); an earlier round of this close already ran the full review
- Claimed and started from ariadne:5 after #565 landed. Survey: publish carries only details and `assets/`; change-code's checkpoint commits issue, plan and gate ledger; nothing commits `.side/` before the archive. So a refusal at publish from a resting branch has no "commit it" fix unless publish carries linked side docs: decision D1 to the TL.
- The link grammar already lives in `sidedoc.RebaseLinks`; the guard reuses it via an extracted `LocalLinks` (ARCH-DRY).
- TL approved the plan and D1 (publish carries linked side docs). change-code inferred the quick flow; the diff will likely upgrade it at close.
- Built: prose (AGENTS.base §1, review convention §6, collaboration-modes §4, export manifest regenerated); `LocalLinks`; `internal/linkguard` census; the guard in change-code (every flow, before the gate loop, since the gate list is skipped on quick), close/milestone-close (`--no-side-links`, catalog rows) and publish (pre-effect for every issue; carry in the one main commit, finish restores or commits the carried paths). Deviations are in the plan's Revisions.
- Found while testing: publish's own asset bundle must count as committed (the asset tests refused otherwise); the guard defers to the asset code there. Mutations (no carry; ignored not refused) each fail their tests.
- Targeted runs: sidedoc, linkguard, processmanual, reviewexport, and `go test ./cmd/sdlc/ -run 'Publish|Close|ChangeCode|GateCatalog|SideLink|Archive|MilestoneClose|Asset'` pass, but for five landing/train tests that fail only on the sandbox's host boot identity (operation not permitted).
- Close review round 1: FIX-THEN-SHIP, BR-1 (close's census ran under its bypass and died on project errors) and BR-2 (no runChangeCode or project-branch tests) Important, five Minors; all fixed this round, see the plan's Revisions.
