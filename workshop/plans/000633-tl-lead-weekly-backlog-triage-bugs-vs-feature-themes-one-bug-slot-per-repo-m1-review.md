# Boundary Review — ariadne#633 (milestone M1)

| field | value |
|-------|-------|
| issue | 633 — tl-lead: weekly backlog triage (bugs vs feature themes, one bug slot per repo, closing-velocity insight, Needs-you output) |
| repo | ariadne |
| issue file | workshop/issues/000633-tl-lead-weekly-backlog-triage-bugs-vs-feature-themes-one-bug-slot-per-repo.md |
| boundary | milestone M1 |
| milestone | M1 |
| window | c5ba88778116a3881ddea2cde230ff83dbd1fd4c..b6b5b2baf45d32f907c43eb94d44474a97a2c891 |
| command | sdlc milestone-close --issue 633 --milestone M1 |
| reviewer | claude (anthropic; author anthropic) |
| isolation | fenced |
| timestamp | 2026-10-25T13:03:03-07:00 |
| verdict | REWORK |

## Review

Checking the existing unknown-key test and the CUE schema's openness, then running the tests.

```verdict
verdict: REWORK
confidence: high
```

The code for M1 is good, but the plan does not describe what was built. `kind` is now a shared-enum card field. `set-kind` and `set-severity` share one setter. `issue list --json` includes terminal card-only cards. All the targeted tests pass: `pkg/vocab`, `internal/issue`, `internal/recovery`, and the `cmd/sdlc` set-kind, list-json, filter, help and catalog tests. The ARCH-EVOLVE gate also holds: the existing `TestOlderCardReaderPreservesUnknownFieldsThroughSetters` (`cmd/sdlc/internal/issue/cardunknown_test.go:36`) shows that a setter rewrite keeps unknown keys, and the CUE frontmatter is open (`...`).

What blocks SHIP is the plan's `CardSummary` row and its contract. The plan promises a `CardSummary` row shaped `{id,title,status,created,updated,kind,severity,area[],claimant_workspace}`, inside a `{schema_version, repo, cards[], unreadable[]}` document, decoded strictly by tl-lead. The code ships something different: `{schema_version, tracker, issues[]}`, built from `state --json`'s `IssueState` rows, with unreadable cards kept as rows. The Log goes further and says tl-lead will decode "tolerantly". There is no `## Revisions` entry, and M2's consumer will be built against whichever of these it reads. Two Important gaps should be fixed at the same time:
- The README was not updated for the new surface.
- The schema-v1 promise is not enforced: the rows are shared with `state --json`, so a change there changes this format without a version bump.

**1. Strengths**
- `newCardEnumSetterCmd` (`cmd/sdlc/cardsetters.go:141`) and `issue.CheckEnum` (`classify.go:216`) remove the duplication between severity and kind instead of copying it (passes ARCH-DRY).
- `issueStates(…, history)` (`state.go:370`) reuses the one existing record walk. It does not add a second tracker reader, and the text listing stays unchanged, which the test checks.
- `TestIssueSetKind` checks the property M2 depends on: setting a kind leaves a done card's `updated` (its close date) alone. It also covers a rerun with no write, refusals that change nothing, and clear.
- Unreadable cards stay in the output as rows with their reason (`issue_list.go:374`), so nothing is dropped silently (passes ARCH-SECURE). The new flag and setter are wired into the recovery catalog with tests that pin them.

**2. Critical**
- **The Core concepts table contradicts the code.** `CardSummary` does not exist. The output is `{schema_version, tracker, issues[IssueState]}`, not `{schema_version, repo, cards[], unreadable[]}`. The owner shows as `owner.workspace`, not `claimant_workspace`. ARCH-SECURE says tl-lead decodes "strictly", but the Log says "tolerantly".
  - Fix: add a `## Revisions` entry that names the shipped shape and the decode policy M2 will follow, and update the Core concepts table to match.

**3. Important**
- **README update appears missing for `set-kind`, `--kind` and `--json`.** The "Classifying issues" section (`README.md:234-255`) documents severity, `set-severity` and `--severity`, but none of the new surface. Fix: add one or two sentences about kind, `set-kind`, `--kind` and `issue list --json`.
- **ARCH-EVOLVE: nothing enforces the schema_version promise.** `IssueListJSON` (`issue_list.go:341-348`) says a field is "never renamed or re-meant without a schema_version bump", but its rows are `IssueState`, which also belongs to `state --json`. A future edit to `state --json` would change this format under v1 without any signal.
  - Fix: either define a separate row type for this format, or add a golden test that pins the v1 row's field set and fails when `IssueState` changes.

**4. Minor**
- Rows with an empty status (unparseable details files) go through the filter, unlike `"unreadable"` rows, so `--kind` or `--status` can drop them silently (`issue_list.go:374`).
- The JSON path returns errors, but the text path calls `die()` (`issue.go:316-321`). Error handling is inconsistent within one command.
- `issue_list.go` and `issuelist.go` sit side by side with names that are easy to confuse.
- The plan says more kinds are "a one-word enum edit", but `ParseCard` refuses any kind outside the enum. A card marked with a new kind would be unreadable to an older pinned sdlc, so the enum cannot grow additively. This is the same as severity today; note it before anyone adds `chore`.

**5. Test coverage notes**
- The tracker fixture covers open, working, done card-only (with kind and owner) and unreadable cards. It also covers `--kind` and the legacy case.
- Nothing covers `--status done --json`, the main query M2 will run.
- Nothing pins the v1 JSON field set (see Important 2).

**6. Architectural notes for M2**
- ARCH-DRY: pass.
- ARCH-PURE: pass. The filter is pure and the new JSON path is a thin layer around it.
- ARCH-PURPOSE: pass for M1.
- ARCH-MOCK: pass. The tests use the existing tracker fixture.
- ARCH-CONSTRAINTS: pass. One record walk, and `withOwnership` does no history lookups.
- ARCH-SECURE: pass on the sdlc side. M2 must settle strict versus tolerant decoding.
- ARCH-ORDER: not applicable, because nothing keeps state between runs.
- ARCH-FUNERAL: pass. The only new durable thing is the `kind` key, which lives as long as its card.
- ARCH-EVOLVE: flagged (Important 2 and the Minor about growing the enum).
- M2 should read owners from `owner.workspace` and treat `tracker:false` as "open counts only".

**7. Plan revision recommendations**
- A `## Revisions` entry that:
  - replaces the `CardSummary` row and the `{repo, cards, unreadable}` envelope with the shipped `{schema_version:1, tracker, issues[IssueState + created, kind]}`;
  - says unreadable cards are rows with `status: "unreadable"`;
  - says the owner is `owner.workspace`;
  - states whether tl-lead decodes strictly or tolerantly.

```findings
findings:
  - id: new
    severity: Critical
    kind: defect
    family: plan-table-matches-code
    title: |
      Core concepts CardSummary row and the issue list --json envelope contradict the shipped shape
    detail: |
      The plan promises CardSummary {id,title,status,created,updated,kind,severity,area[],claimant_workspace} inside {schema_version, repo, cards[], unreadable[]}, and says tl-lead decodes it strictly (ARCH-SECURE). cmd/sdlc/issue_list.go ships {schema_version, tracker, issues[IssueState]}, with unreadable cards inline and the owner under owner.workspace, and the Log says tl-lead will decode it tolerantly. There is no Revisions entry, so M2 has two conflicting contracts. Fix: add a Revisions entry and update the table to the shipped shape and the chosen decode policy.
  - id: new
    severity: Important
    kind: defect
    family: readme-tracks-user-surface
    title: |
      README update appears missing for issue set-kind, list --kind and list --json
    detail: |
      README.md Classifying issues (lines 234-255) documents severity, set-severity and --severity, but the diff adds set-kind, --kind and --json and README.md is not touched in the range. Add a sentence or two in that section.
  - id: new
    severity: Important
    kind: defect
    family: versioned-format-pinned-by-test
    title: |
      issue list --json schema_version 1 is not enforced, because its rows are state --json's IssueState
    detail: |
      IssueListJSON (issue_list.go:341-348) promises no rename or re-meaning without a schema_version bump, but its rows are IssueState, which state --json also owns. Any edit there silently changes v1 (ARCH-EVOLVE). Either give the format its own row type, or add a golden test pinning the v1 row field set.
  - id: new
    severity: Minor
    kind: defect
    family: list-filter-keeps-unreadable
    title: |
      Rows with an empty status (malformed details) are filtered out, unlike unreadable rows
    detail: |
      issue_list.go:374 bypasses the filter only for status "unreadable". The DetailErr rows with status "" in state.go go through the filter, so --kind or --status drops them silently.
  - id: new
    severity: Minor
    kind: defect
    family: closed-enum-growth
    title: |
      Adding a kind value later makes cards unreadable to older pinned sdlc
    detail: |
      ParseCard refuses a kind outside the enum (TestParseCardRefusesUnknownKind), so the plan's claim that more kinds are a one-word enum edit is not additive for older readers. The same is already true of severity; note it before growing the enum.
```

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"dd20a9d5-7e73-49f5-81e3-ba2d4c293212","repository":"github.com/xianxu/ariadne","issue_id":"000633","recipe":"milestone-review","round":1,"attempt":1,"agent":"claude","root":"/Volumes/TestScratch/review-checkouts/fd154526bd2e02b9","reviewed_head":"b6b5b2baf45d32f907c43eb94d44474a97a2c891","slot":"ariadne:2","requested_at":"2026-10-25T19:56:50.027325Z"},"phase":"finished","pid":64450,"started_at":"2026-10-25T19:56:50.201033Z","finished_at":"2026-10-25T20:03:02.363275Z","session_id":"b3cb3463-adca-4277-af2f-1d021646e1e0","outcome":"returned"}] -->
