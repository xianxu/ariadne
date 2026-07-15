# Boundary Review — ariadne#160 (milestone M3)

| field | value |
|-------|-------|
| issue | 160 — add an issue status: codecomplete |
| repo | ariadne |
| issue file | workshop/issues/000160-add-an-issue-status-codecomplete.md |
| boundary | milestone M3 |
| milestone | M3 |
| window | 21461a80e6bb515d81ed0d8d047f03d4a90ecde2..HEAD |
| command | sdlc milestone-close --issue 160 --milestone M3 |
| reviewer | claude |
| timestamp | 2026-07-02T17:53:31-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

All tests pass (`go test ./cmd/sdlc/ ./pkg/vocab/` → exit 0, confirmed by two independent runs). I have everything I need.

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

**Summary.** M3 delivers the two-gate model cleanly: the pre-merge LLM judges are removed, `runPublishGate` enforces the reviewed-HEAD-unchanged invariant deterministically, and `codecomplete → done` flips on merge/push before archiving. The invariant logic is sound and rests on a genuinely enforced property (set-status refuses `codecomplete`, so the anchor commit is trustworthy — verified at `setstatus.go:256`). Build is clean, no dangling `judge`/`preflight` references, full suite green. Nothing blocks SHIP. The one thing worth fixing before crossing is a missing regression test the plan explicitly called for (the `touchedIssuesNotDone` carve-out); the rest are Minor/plan-alignment notes.

### 1. Strengths
- **The seam rename preserves the #62 defense.** `runPreflightJudgesFn → runPublishGateFn` keeps the "step-5 hook that dirties the tree" test intact (`merge_e2e_test.go` `TestRunMerge_DirtyAfterJudge_RefusesPreMerge`) even though the judges are gone — the 9b re-check is still exercised. Good instinct to keep that window covered.
- **Anchor trust is enforced, not just asserted.** `codecompleteAnchorCommit`'s doc (publishgate.go:23-39) correctly grounds the invariant in the sole-writer property, and `setstatus.go` Guard 1b (`:256`) is the actual enforcement. The "Residual (by design)" disclosure is exactly the right level of honesty.
- **Multi-issue latest-anchor reasoning is correct** (publishgate.go:104-121): because only whole-issue `close` writes `codecomplete`, the newest anchor is necessarily a whole-branch review, so the branch-level check is valid — no false per-issue drift refusal.
- **Unit coverage of the pure-ish core is strong** (publishgate_test.go): code-drift-doesn't-move-anchor, re-close-advances-anchor, clean/drift/multi/no-op all pinned.
- **Docs rewrite is high quality** — `pre-merge-checks.md` two-gate model, `ci-merge-check.md` three-tier note, and merge/push help all consistent.

### 2. Critical findings
None.

### 3. Important findings

**(a) The `touchedIssuesNotDone` codecomplete carve-out is untested** — `push.go:536`, test at `push_test.go:301`.
Plan M3 Step 6 explicitly required: *"Add a test asserting a codecomplete issue does not trip the 'Continue anyway?' prompt."* `TestTouchedIssuesNotDone` still only seeds `working`/`done`/`open` — no `codecomplete` case. The carve-out (`&& st != "codecomplete"`) is exactly the kind of one-token condition a later refactor drops silently, re-introducing "every merge/push trips the not-done warn." Fix: add `mkIssue("000004-cc.md", "codecomplete")` to that test's touched set and assert it's absent from `notDone`.

**(b) Push-side flip+archive has no integration test.** Merge got `TestRunMerge_CodecompleteFlippedToDoneAndArchived` (merge_e2e_test.go); the structurally-identical push path (step 6.5 flip → step 7 archive/commit, push.go:177-210) has none. The flip-then-archive ordering and the "flip is captured by the archive commit" guarantee are load-bearing and untested for push. Fix: mirror the merge e2e test for `runPush` (a codecomplete issue ends up `done` in history/).

### 4. Minor findings
- **Flip scope diverges from the plan (dir-wide vs window-scoped).** `publishCodecompleteIssues(issuesDir)` (publishgate.go:131) globs and flips *every* codecomplete issue in the dir, but the plan (Step 4 / Core-concepts) specified flipping the `mergedCodecompleteIssues` (window-scoped) set — the same set the gate verified. Benign on a healthy main (no stray codecomplete issues), and consistent with `archiveDoneIssues`'s existing dir-wide behavior, but it means an out-of-window codecomplete issue would be flipped+archived without the invariant being checked (and is a latent concurrent-merge hazard). Either scope the flip to the verified set or record the intentional dir-wide choice (see §7).
- **`revCount` returns 0 on git error** (publishgate.go:159-162) → the gate would silently *pass* if `rev-list anchor..HEAD` ever errored. Unreachable in practice (the anchor always comes from HEAD's own log, so the range is always valid), but a `-1`/error sentinel would be more defensive than treating "git failed" as "no drift."
- **DRY: two near-identical window scans.** `mergedCodecompleteIssues` (publishgate.go:65) and `touchedIssuesNotDone` (push.go:515) both do `git diff --name-only baseRef..HEAD -- issuesDir/*.md` → parse → status-filter; the comment even says "Mirrors touchedIssuesNotDone (ARCH-DRY)" — but mirrors, doesn't reuse. A shared `changedIssueFiles(baseRef, issuesDir) []struct{Path,Status}` would consolidate both (and `publishCodecompleteIssues`/`archiveDoneIssues` share the glob+parse shape too). See §6.

### 5. Test coverage notes
- Pure/invariant logic: well covered (publishgate_test.go, real temp-git integration — correct choice over mocks per AGENTS.md §5).
- Gaps: the carve-out (§3a) and push wiring (§3b). Both are behaviors the plan named and neither is pinned. The `--no-judge`-skips-gate-but-still-flips path (merge.go:317 / push.go:135 vs the ungated step 6.5/10.5) is also untested — worth a case confirming `--no-judge` still flips.

### 6. Architectural notes for upcoming work
- **ARCH-DRY — flag (Minor).** The four issue-file scanners (`mergedCodecompleteIssues`, `touchedIssuesNotDone`, `publishCodecompleteIssues`, `archiveDoneIssues`) are converging on one shape; extract a shared changed/globbed-issues-with-status helper before a fifth appears.
- **ARCH-PURE — pass.** New code is thin git-glue and is exercised against a real process-level git repo, not function mocks; the one bit of pure logic (head==anchor) is correctly inlined per the plan's explicit "no ceremony" decision.
- **ARCH-PURPOSE — pass.** All M3 Done-when items are delivered (flip, invariant, judge removal, help/atlas). The only scope note is an *over*-reach (dir-wide flip), not under-delivery.

### 7. Plan revision recommendations
- Add a `## Revisions` entry to `workshop/plans/000160-codecomplete-status-plan.md` reconciling the Core-concepts table with the code: (1) the flip lives in a **new** `publishCodecompleteIssues` (not listed in the Integration-points table), and `runPublishGate`'s "Wraps: invariant check + flip" is inaccurate — `runPublishGate` does only the check; (2) the flip is **dir-wide** (glob), not the plan's stated window-scoped `mergedCodecompleteIssues`-driven flip — state whether that's intentional (consistency with `archiveDoneIssues`) and note the healthy-main assumption it relies on.
