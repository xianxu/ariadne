---
collaboration-mode: discussion
---

# Weekly backlog triage (tl-lead) Implementation Plan

> **For agentic workers:** Consult AGENTS.md Section 3 (Subagent Strategy) to determine the appropriate execution approach: use superpowers-subagent-driven-development (if subagents are suitable per AGENTS.md) or superpowers-executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** the TL triages each repository's open issues weekly (bug or feature; ready, waiting, stale or duplicate), sizes one bug slot per repository from measured velocity, proposes feature themes as projects, and reports to the operator in Needs you, backed by a read-only `tl-lead triage stats`.

**Architecture:** the bug/feature mark becomes a card field (`kind`), owned by sdlc like `severity` (#348), so triage writes it once and every later reading derives from it. sdlc gains one machine-readable reader of every card, terminal ones included (`sdlc issue list --json`), because no reader gives closed issues today and tl-lead cannot import sdlc's internals. tl-lead reads it per repository through its existing `SDLC` seam, and a pure core buckets the cards into weeks. The duty itself is prose in `tl-lead --help`.

**Tech stack:** Go, cobra, CUE vocabulary (`construct/vocabulary/issue.cue`), `pkg/vocab`.

---

## Facts this design rests on (surveyed 2026-10-25)

- **No bug/feature field.** Cards carry `severity` (14 of 629 set), `flake_tests` (15) and `area`; nothing says bug or feature (`issue.cue:100-125`).
- **No reader of closed issues.**
  - `sdlc state --json` skips terminal card-only issues (`state.go:372-374`) and has no `created`.
  - `sdlc issue list` reads details files and prints text only.
  - `cmd/sdlc/internal/*` is not importable by tl-lead.
- **Closed cards keep their claimant** (for example `sdlc issue show 617 --json`: done, with `claimant.workspace: ariadne:3`). So closes per slot per week can be measured from cards.
- **`updated` is the close date of a terminal card**, at day resolution: `set-status` and the close and merge flips bump it. Weekly buckets need nothing finer.
- **tl-lead's pattern to copy is `team issues`:**
  - The pure core `teamIssues(out, s SDLC, …)` walks `claimedRepos(h)`, and a failed repository prints `repo: ? (err)` without stopping the others.
  - Its test uses `fakeSDLC` (`trains_test.go:16-25`).
  - Help tests flatten text and check phrases and their order (`takeover_help_test.go`).

## Core concepts

### Pure entities

| Name | Lives in | Status |
|------|----------|--------|
| `kind` card field (`bug`, `feature`) | `construct/vocabulary/issue.cue` | new |
| `CardSummary` JSON row | `cmd/sdlc/issue_list.go` (new file beside `issue.go`) | new |
| `weekStats` (per-repo week buckets) | `cmd/tl-lead/triage_stats.go` | new |
| `slotVelocity` | `cmd/tl-lead/triage_stats.go` | new |

- **`kind`** is a shared enum like `severity`, with setter `sdlc issue set-kind`. Unset means unclassified; triage's first pass sets it.
  - The stats report the unclassified count instead of guessing. A title heuristic (such as "flake:" or "fix") would mislabel issues and give the trend a false precision.
  - DRY: one mark, set once, and every consumer reads it.
  - Future extensions: more kinds, such as `chore`, are a one-word enum edit.
- **`CardSummary`** is one card: `{id, title, status, created, updated, kind, severity, area[], claimant_workspace}`.
  - The output is `{schema_version: 1, repo, cards[], unreadable[]}`.
  - A card that fails to parse goes into `unreadable` with the error, so it is never silently dropped (ARCH-SECURE).
- **`weekStats(cards, now, weeks)`** gives, for each ISO week (Monday start, in the operator's local zone), the count of bugs opened (`created`), bugs closed (terminal with `updated` in the week) and open bugs at the week's end.
  - It also gives features opened and closed, and the unclassified count.
  - It is pure: it takes `now` and has no IO, so it is unit-tested directly.
- **`slotVelocity(cards, now, weeks)`** counts issues closed per claimant workspace per week over the window (4 weeks by default). It returns the median over the weeks in which a slot closed anything.
  - The suggested bug quota is the median for the repository's slots, rounded down, and at least 1.
  - A slot with no closes in the window is not counted, so an idle slot does not drag the median down.

### Integration points

| Name | Lives in | Status | Wraps |
|------|----------|--------|-------|
| `sdlc issue list --json` | `cmd/sdlc/issue_list.go` (new; `issue.go`'s `runIssueList` dispatches to it) | new | the issue-tracker cards (`origin/issue-tracker`) |
| `sdlc issue set-kind` | `cmd/sdlc/cardsetters.go` | new | card CAS, like `set-severity` |
| `SDLC.RepoCards(root)` | `cmd/tl-lead/sdlc.go` | new | `sdlc issue list --json` per repository |
| `tl-lead triage stats` | `cmd/tl-lead/triage_stats.go` | new | stdout |

- **`sdlc issue list --json`** lists every card in a repository with an issue tracker, terminal ones included, read through the tracker reader that `state` and `show` already use.
  - In a legacy repository it reads the details files as today.
  - The text output is unchanged. `--status`, `--area` and `--severity` still filter, and `--kind` is added.
  - ARCH-MOCK: the existing tracker test fixtures (a temp repository with an `issue-tracker` branch) back its tests, so no new fake is needed.
- **`SDLC.RepoCards`** is injected into the triage core. The tests extend `fakeSDLC` with `cards map[root]CardList` and `cardsErr`, a stateful fake keyed by repository.
- **Repositories covered:** `claimedRepos(h)`, the same scope `team issues` uses. `--repo R` narrows the run to one.

## Operating envelope (ARCH-CONSTRAINTS)

- **Workload:** an operator-initiated batch read, run weekly. There is no interactive path.
- **Scale:**
  - Today about 630 cards per repository and 2 to 4 repositories.
  - Each repository costs one sdlc call, within the seam's existing 60 s timeout.
  - The core is O(cards × weeks), with weeks at 12 by default and 52 at most (a flag).
- **When a call exceeds its timeout or fails:** that repository prints `repo: ? (err)` and the others still report. Nothing retries.

## Other principles

- **ARCH-ORDER:** holds no state between events. Every run recomputes from the cards, and nothing is cached or written.
- **ARCH-FUNERAL:** creates nothing durable apart from the `kind` key on cards, which lives as long as its card. `stats` writes nothing.
- **ARCH-EVOLVE:** `kind` is an additive card field, and `CardSummary` is versioned (`schema_version: 1`), so new fields are additive.
  - Risk: an older sdlc (pair's pinned binary) reading or rewriting a card that carries `kind`.
  - M1's first step proves with a test that the current card reader keeps an unknown frontmatter key through a setter rewrite.
  - If it does not, the plan stops and comes back to the TL with that finding, because the pin-bump ordering would then matter.
- **ARCH-SECURE:**
  - The cards are repo-written input. They are parsed into typed rows at the sdlc boundary, and the malformed ones are reported.
  - tl-lead decodes `CardSummary` strictly by schema version, and refuses an unknown `schema_version` naming it.
  - No credentials are involved.
- **ARCH-PURPOSE:** the duty is the point, and `stats` serves it. Help documents the full weekly procedure (marks, capacity, themes, Needs-you shape) in the same milestone that ships stats, so neither half is a follow-up.

## The duty (help.md text, M2)

A new section, `WEEKLY BACKLOG TRIAGE`, sits beside GROOMING THE BACKLOG INTO TRAINS. It is prompted by the weekly scheduled task (couch scheduling, pair#737). It says:

1. **Run `tl-lead triage stats`.** Read each repository's trend: bugs opened against bugs closed, and the open-bug line.
2. **Mark each repository's open issues:**
   - **Kind:** `sdlc issue set-kind --issue N bug|feature`.
   - **Readiness:** ready (relevant, deps met, small enough), waiting (on what), stale (propose closing, with the reason) or duplicate (of which).
   - **Where the readiness marks live:** in the triage report, not on the card. They are judgements for this week's list, which the operator approves, so they are not durable fields. A stale or duplicate issue closes only after the operator approves, through `sdlc abandon`.
3. **Capacity:**
   - One bug slot per repository by default.
   - Its weekly quota is the suggested number from stats, the median closes per slot per week. Pick that many ready bugs, by severity and then age; the rest wait.
4. **Features:** group the ready and waiting features into themes. Propose each theme to the operator as a candidate project (the project datatype), not as a dispatch.
5. **Needs you, before anything proceeds,** in this shape:
   - Per repository, one line of trend, then the week's bug list and the close proposals.
   - Then the feature themes proposed as projects.
   - Then, when the open-bug line is rising, the trade-off: shift capacity from features to fixes, with a recommendation. The operator decides.
6. **The 21:00 ready-queue check (ops#11)** counts against this approved ready list.

## Plan

- [ ] M1: the `kind` card field and the card reader (sdlc)
  - [x] Test first: an unknown frontmatter key survives a `set-severity` rewrite (the ARCH-EVOLVE gate above).
  - [x] Add `kind` to `issue.cue` (enum and card field, setter `sdlc issue set-kind`). Regenerate whatever `pkg/vocab` derives, and add the `vocab` tests.
  - [x] `sdlc issue set-kind`: follow the `set-severity` path in `cardsetters.go`, with tests for set, clear, an unknown value refused, and a rerun that makes no write.
  - [x] `sdlc issue list --json` and `--kind`:
    - Tests over a tracker fixture with open, working and done cards (terminal ones included) and an unreadable card.
    - A legacy-repository case.
    - The text output stays unchanged.
  - [x] Atlas: `sdlc-binary.md`, the issue-card fields and the `issue list --json` contract.
- [ ] M2: `tl-lead triage stats` and the duty
  - [ ] Pure core tests first:
    - `weekStats` over a fake card list spanning week boundaries, including a bug opened and closed in the same week, a terminal card closed before the window, and unclassified cards.
    - `slotVelocity`, including an idle slot and the quota's floor of 1.
  - [ ] `SDLC.RepoCards` and the `fakeSDLC` extension; decode tests against raw JSON, including an unknown schema version.
  - [ ] `tl-lead triage stats [--repo R] [--weeks N]`: for each repository, a table of the weeks, the unclassified count and the suggested bug quota; a failed repository prints `?` and the rest continue.
  - [ ] `help.md`: the WEEKLY BACKLOG TRIAGE section. A help test checks the steps are in order and that the marks, the one-bug-slot default, velocity sizing, themes as projects and the Needs-you shape are all present.
  - [ ] Atlas: the tl-lead entry.

**Proofs:**
- Targeted tests per package, one run at a time.
- `tl-lead triage stats` run against the live ariadne tracker, with the output pasted into the Log as a smoke check.
- The pre-close suite.

**Non-goals:**
- Automating the marks. The TL judges them and the operator approves.
- Creating the project files. Proposing themes is the duty; making a project follows the operator's yes, with the project datatype.
- Back-filling `kind` on the 629 existing cards. The first triage pass marks the open ones, and closed cards stay unclassified, which the stats report as such.

## Revisions

### 2026-10-25 — TL approval

- `kind` is a card field (`bug` or `feature`), and the stats use no heuristics.
- Back-fill, which replaces the non-goal above: after M1 lands, the TL marks the closed cards from the last 4 weeks once (the bug quota's window), so the first triage has a real quota and trend. Older cards stay unclassified. This work is the TL's, not this issue's.
- The ARCH-EVOLVE unknown-key test stays M1's first step. If a setter drops unknown keys, the work stops and goes back to the TL.

### 2026-10-25 — plan-quality round 1 (3 Minors, folded in)

- **`updated` must stay the close date.** `updated` is written only by the close and status paths (`close.go:811`, `trackercompletion.go:151`, `completeop.go:118`, `externalmerge.go:417`), never by the generic setter. The TL's back-fill depends on that, so the `set-kind` tests assert that `updated` is unchanged on a done card.
- **`issue list --json` walks the full record set.** `runIssueList` reads through `listIssueStates`, the same path as `state`, which drops terminal card-only records (`state.go:371-374`). So `--json` walks the tracker's record set directly, and a test with a done card-only issue proves it is listed. `issue_list.go` is a new file.
- **Legacy repositories have no close history.** Their closed issues live in `workshop/history/`, so `--json` there lists the open details files only, and `triage stats` reports such a repository as "no tracker: open counts only".
- **`weekStats` test strategy.** Generated card sets that span DST changes and ISO week-year boundaries. Property: for each week, bugs opened minus bugs closed equals the change in open bugs, given no reopens.
- **Open at a week's end** means `created <= end`, and either the card is non-terminal or `updated > end`. A reopened card makes the figure approximate, and help says so.
