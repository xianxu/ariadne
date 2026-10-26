# Boundary Review — ariadne#633 (whole-issue close)

| field | value |
|-------|-------|
| issue | 633 — tl-lead: weekly backlog triage (bugs vs feature themes, one bug slot per repo, closing-velocity insight, Needs-you output) |
| repo | ariadne |
| issue file | workshop/issues/000633-tl-lead-weekly-backlog-triage-bugs-vs-feature-themes-one-bug-slot-per-repo.md |
| boundary | whole-issue close |
| milestone | — |
| window | c5ba88778116a3881ddea2cde230ff83dbd1fd4c..70cfcfb2be1445122ecd7214e5151dd1ff97ce95 |
| command | sdlc close --issue 633 |
| reviewer | claude (anthropic; author anthropic) |
| isolation | fenced |
| timestamp | 2026-10-25T16:18:45-07:00 |
| verdict | SHIP |

## Review

Triage tests pass. Next I'm checking that the unrelated tl-lead failures come from the sandbox and not from this diff.

```verdict
verdict: SHIP
confidence: high
```

The issue's two Done-when items are delivered. `tl-lead --help` now has a WEEKLY BACKLOG TRIAGE section in order, covering the bug/feature and ready/waiting/stale/duplicate marks, one bug slot per repo, quota sizing from measured velocity, feature themes proposed as projects, and the Needs-you report shape. `tl-lead triage stats` is a read-only helper with a pure core, and its tests run over generated and fake card histories. The one open prior finding, BR-6, is fixed. `IssueListJSON` has an additive `stale` key, set from the same `issueStates` flag that triggers the stderr warning (`cmd/sdlc/issuelist_json.go:79`). The pinned-keys test lists it (`issuelist_json_test.go:126`), so removing the field turns that test red. tl-lead decodes the key and prints a stale line, and both steps are tested (`triage_stats_test.go:108,133,146`). Nothing blocks SHIP.

On the test run: the targeted sdlc, pkg/vocab, internal/issue and internal/recovery tests are green. So are all the triage, team-issues and help tests in tl-lead. About 80 other tl-lead tests fail in this checkout, and the ones I sampled fail with "git init refused … outside the reviewer's directories (sdlc #339)". That is the review sandbox's confinement, not this diff. None of the failing tests touch code changed in this range.

1. **Strengths**
   - `newCardEnumSetterCmd` and `issue.CheckEnum` (`cmd/sdlc/cardsetters.go:194`, `internal/issue/classify.go:28`) merge `set-severity` and `set-kind` into one path instead of copying it (ARCH-DRY).
   - The week arithmetic works on civil days (`day`, `monday()`), so DST never moves a week boundary. `TestWeekStatsReconcilesOpenBugs` checks the opened − closed = Δopen property across DST changes and the ISO week-year boundary.
   - `CardRow` is a separate type from `IssueState`, and its v1 keys are pinned by a test. tl-lead decodes it tolerantly of extra fields and strictly on `schema_version`, and both sides are tested (ARCH-EVOLVE).
   - `thisHub` is pulled out of `team issues` and reused. `triageStats` follows the `teamIssues` pattern: one failed repo prints `?` and the rest still report.
   - Unreadable cards stay as rows and are counted. They are never silently dropped, and the stats never guess a kind.

2. **Critical:** none.

3. **Important:** none.

4. **Minor:** none new. One observation, not raised as a finding: no sdlc-side test drives a real failed fetch through to `stale: true` in the JSON. The wiring is one line and the key is pinned.

5. **Test coverage**
   - Pure-core tests run without IO: `weekStats`, `slotVelocity`, `bugQuota`, `unclassified`, `decodeRepoCards`.
   - `TestTriageStats` covers error, no-tracker, stale, unreadable and `--repo` through `fakeSDLC`.
   - The sdlc tests use a real tracker fixture with open, working, done card-only and unreadable cards, plus the `--kind`, `--status done` and `--status open` filters.
   - `TestIssueSetKind` asserts that `updated` is unchanged on a done card, which the close-date logic depends on.

6. **Architecture notes**
   - ARCH-DRY: pass.
   - ARCH-PURE: pass. The IO is limited to `execSDLC.RepoCards` and `thisHub`.
   - ARCH-PURPOSE: pass. The duty text and the stats ship together.
   - ARCH-MOCK: pass. `fakeSDLC` keys its state by repo, and the sdlc tests use real git tracker fixtures. Nothing automatically checks tl-lead's decoder against sdlc's actual encoder: each side is pinned separately (raw JSON on one side, key list on the other). The live smoke run in the Log is the only end-to-end evidence. A shared golden fixture would close that gap later.
   - ARCH-CONSTRAINTS: pass. `--weeks` is bounded to 1–52, and each repo costs one call under the existing timeout.
   - ARCH-SECURE: pass. Malformed input becomes visible unreadable counts, and no credentials are involved.
   - ARCH-ORDER: pass. Nothing holds state between runs; every run recomputes from the cards.
   - ARCH-FUNERAL: pass. The only durable addition is the `kind` key, which lives as long as its card.
   - ARCH-EVOLVE: pass. `stale` is additive within v1. The closed-enum limit on `kind` is documented in `vocabulary.md`.

7. **Plan revisions:** none needed. The round-2 revision already records the `stale` key. The Core concepts table still describes the old `CardSummary` envelope, but the BR-1 revision explicitly replaces it, and that finding is already disposed.

```findings
dispose:
  - id: BR-6
    disposition: addressed
    note: |
      IssueListJSON.Stale is set from issueStates (issuelist_json.go:79), pinned by TestIssueListJSONFieldSetIsPinned (red without the key), and decoded and rendered by tl-lead (TestDecodeRepoCards, TestTriageStats stale line).
```

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"41533f21-ac0b-47f9-91e3-455d00e68fe7","repository":"github.com/xianxu/ariadne","issue_id":"000633","recipe":"milestone-review","round":3,"attempt":1,"agent":"claude","root":"/Volumes/TestScratch/review-checkouts/46235bb67e067b90","reviewed_head":"70cfcfb2be1445122ecd7214e5151dd1ff97ce95","slot":"ariadne:2","requested_at":"2026-10-25T23:16:57.362788Z"},"phase":"finished","pid":25409,"started_at":"2026-10-25T23:16:57.520618Z","finished_at":"2026-10-25T23:18:43.891751Z","session_id":"0da0fa5d-44c8-44ea-84de-7d548eaaf3f5","outcome":"returned"}] -->
