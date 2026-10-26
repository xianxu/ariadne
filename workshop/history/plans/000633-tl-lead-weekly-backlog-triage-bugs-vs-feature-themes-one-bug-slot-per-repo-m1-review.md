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

---

## Re-review — 2026-10-25T13:16:04-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 633 — tl-lead: weekly backlog triage (bugs vs feature themes, one bug slot per repo, closing-velocity insight, Needs-you output) |
| repo | ariadne |
| issue file | workshop/issues/000633-tl-lead-weekly-backlog-triage-bugs-vs-feature-themes-one-bug-slot-per-repo.md |
| boundary | milestone M1 |
| milestone | M1 |
| window | c5ba88778116a3881ddea2cde230ff83dbd1fd4c..3528c2ed5e7fb0c5a109c8ff5ec8ef91d9bda1e4 |
| command | sdlc milestone-close --issue 633 --milestone M1 |
| reviewer | claude (anthropic; author anthropic) |
| isolation | fenced |
| timestamp | 2026-10-25T13:16:04-07:00 |
| verdict | SHIP |

## Review

I've read the diff and the plan. Next I'm checking the ARCH-EVOLVE unknown-key test and running the tests.

```verdict
verdict: SHIP
confidence: high
```

All five findings from round 1 are fixed, and I checked each fix against the code rather than the commit messages. `sdlc issue list --json` now has its own `CardRow`/`IssueListJSON` types in `cmd/sdlc/issuelist_json.go`, and a reflection test pins the version 1 keys. The plan has a dated Revisions entry recording the shipped envelope and the decode policy (tolerant on fields, strict on `schema_version`). The README Classifying issues section now covers kind, `set-kind`, `--kind` and `--json`. Rows from malformed details files are kept as unreadable rows. The atlas notes that the enum cannot grow additively. The targeted tests pass (`pkg/vocab`, `cmd/sdlc/internal/issue`, and the `TestIssueList*`, `TestIssueSetKind`, `TestParseCardRefusesUnknownKind`, `TestCardRowOf*` and setter/recovery tests in `cmd/sdlc`). I reverted two of the fixes one at a time in the disposable checkout, saw their tests fail, and restored them; the tree is clean. Nothing blocks SHIP. There is one new Minor: a stale tracker read is reported only on stderr, not in the JSON.

1. **Strengths**
   - `cmd/sdlc/cardsetters.go:186-215`: `set-severity` and `set-kind` share one `newCardEnumSetterCmd`, and `issue.CheckEnum` is generic, so a third enum setter is a one-line addition (ARCH-DRY).
   - `cmd/sdlc/state.go:369-388`: `issueStates(…, history)` reuses the reader that `state` already uses and only lifts the filter on terminal card-only cards. There is no second walk over the tracker.
   - `cmd/sdlc/issuelist_json.go:48-60`: `cardRowOf` is a pure projection, and `TestCardRowOfUnparsedDetails` tests it directly without IO. When I disabled the empty-status mapping, that test failed.
   - `cmd/sdlc/issuelist_json_test.go:80-96`: the table of filter queries asserts that the unreadable row survives `--kind`, `--status done` and `--status open`. When I removed the bypass, the test failed.
   - `cmd/sdlc/cardkind_test.go:12-56`: the test asserts that a done card's `updated` is unchanged after `set-kind`. That is the close-date assumption the weekly trend and the TL's back-fill depend on.

2. **Critical findings:** none.

3. **Important findings:** none.

4. **Minor findings**
   - `cmd/sdlc/issuelist_json.go:76-78`: when the tracker read falls back to the local mirror (`stale`), only `cwarn` on stderr says so. The JSON has no field for it, so tl-lead, which reads stdout through its `SDLC` seam, will count weekly trends from an old snapshot without knowing. Fix: add a `stale` boolean to `IssueListJSON`. That is an additive v1 key, and `TestIssueListJSONFieldSetIsPinned` needs updating in the same change. (ARCH-SECURE)
   - Plan body: the `CardSummary` bullet and the ARCH-SECURE "decodes `CardSummary` strictly" bullet still describe the old shape. A dated Revisions entry supersedes both and the Core concepts table is correct, so this is cosmetic only. Rewriting the bullets in place would spare M2's implementer the cross-reference. I'm not raising it as a new finding.

5. **Test coverage notes**
   - The ARCH-EVOLVE gate in M1 step 1 rests on the existing `TestOlderCardReaderPreservesUnknownFieldsThroughSetters` (`cmd/sdlc/internal/issue/cardunknown_test.go:36`). It exists and passes, and the Log records the reason no new test was written. I accept that.
   - The key-pinning test compares JSON tag names only. It does not pin `omitempty` or the meaning of values. Values come from `IssueState` through `cardRowOf`, so a change to what `state` means by `Status` could still reach the format. That is acceptable for v1, but keep `cardRowOf` as the only coupling point.
   - Coverage of the legacy repository (`tracker: false`) and of the terminal card-only path is adequate.

6. **Architectural notes for upcoming work (M2)**
   - **ARCH-DRY: pass.** The enum setter and enum check are consolidated.
   - **ARCH-PURE: pass.** `cardRowOf` and `issueListFilter.matches` are pure; `runIssueListJSON` is thin glue.
   - **ARCH-PURPOSE: pass for M1.** The kind field, the reader of every card and the docs are delivered; stats and the duty belong to M2, as planned.
   - **ARCH-MOCK: pass.** The tests run against the tracker fixture repository. M2's `fakeSDLC.cards` must decode the same raw JSON that sdlc produces.
   - **ARCH-CONSTRAINTS: pass.** It is a single batch read per repository with no new fan-out.
   - **ARCH-SECURE: pass, with the Minor above.** Unreadable input stays visible as rows; the stale tracker read does not reach the JSON.
   - **ARCH-ORDER: pass.** It holds no state between events, because every call recomputes from the tracker records.
   - **ARCH-FUNERAL: pass.** It creates nothing durable apart from the `kind` key, which lives as long as its card.
   - **ARCH-EVOLVE: pass.** `kind` is an additive card key and survives setter rewrites; the v1 keys are pinned; the closed-enum caveat is documented in `atlas/workflow/vocabulary.md`.
   - For M2: the tl-lead decoder should refuse any `schema_version` other than 1 by name, ignore unknown keys, and treat `status: "unreadable"` rows as a separate count rather than a status.

7. **Plan revision recommendations:** none required. Optionally, rewrite the stale `CardSummary` and ARCH-SECURE bullets in the plan body to match the round 1 revision.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      Core concepts table names CardRow/IssueListJSON in issuelist_json.go; Revisions "M1 boundary review round 1" records the shipped envelope and tolerant-fields/strict-version decode policy.
  - id: BR-2
    disposition: addressed
    note: |
      README.md Classifying issues now documents kind, set-kind, --kind and --json (lines 236-257 at head).
  - id: BR-3
    disposition: addressed
    note: |
      CardRow is its own type (issuelist_json.go:29-42); TestIssueListJSONFieldSetIsPinned pins v1 document and row keys.
  - id: BR-4
    disposition: addressed
    note: |
      cardRowOf maps empty status to unreadable; TestCardRowOfUnparsedDetails and the filter-query table go red when the fix or the bypass is reverted.
  - id: BR-5
    disposition: addressed
    note: |
      atlas/workflow/vocabulary.md and the plan's future-extensions line state the enum is closed in the reader and cannot grow additively.
findings:
  - id: new
    severity: Minor
    kind: defect
    family: machine-output-carries-degradation
    title: |
      issue list --json reports a stale tracker read only on stderr, not in the document
    detail: |
      issuelist_json.go:76-78 calls cwarn when issueStates returns stale, but IssueListJSON has no field for it, so tl-lead reading stdout through its SDLC seam computes weekly trends from an old mirror without knowing. Add an additive stale boolean to v1 and update the pinned key test.
```

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"90b09e28-eab2-4236-a840-ad557f7839f1","repository":"github.com/xianxu/ariadne","issue_id":"000633","recipe":"milestone-review","round":2,"attempt":1,"agent":"claude","root":"/Volumes/TestScratch/review-checkouts/086de6687c72f0f4","reviewed_head":"3528c2ed5e7fb0c5a109c8ff5ec8ef91d9bda1e4","slot":"ariadne:2","requested_at":"2026-10-25T20:14:16.046995Z"},"phase":"finished","pid":66088,"started_at":"2026-10-25T20:14:16.157434Z","finished_at":"2026-10-25T20:16:03.213304Z","session_id":"e26dbf58-eaa6-4207-a419-615f4779ce0b","outcome":"returned"}] -->
