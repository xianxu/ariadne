# Boundary Review — ariadne#190 (whole-issue close)

| field | value |
|-------|-------|
| issue | 190 — active-time mention-fallback attributes replay work to the replayed issue |
| repo | ariadne |
| issue file | workshop/issues/000190-active-time-mention-fallback-attributes-replay-work-to-the-replayed-issue.md |
| boundary | whole-issue close |
| milestone | — |
| window | defe6306415900ce49395fa76106af0012b6da12..HEAD |
| command | sdlc close --issue 190 |
| reviewer | claude |
| timestamp | 2026-07-29T17:23:33-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

```verdict
verdict: FIX-THEN-SHIP
confidence: medium
```

The fix is correct and well-targeted: the root cause was correctly re-diagnosed (the filed Spec's "commit boundaries outrank mentions" rule would have fixed nothing, because `Commit.Issues` is poisoned by the same parse and `attributeRun` splits `weight × active` equally across it), all three broken sites now derive from one grammar, `migrate.go`'s two encodings genuinely compose from the same fragment, and the exclusion is made observable rather than silent. I found no correctness defect in the shipped behavior and no surviving plan-gate carry-forward findings (PQ-6/7/8 are all delivered as the round-3 ledger claims). What holds it back from SHIP is bookkeeping and one coverage hole: the issue file has a corrupted duplicated `## Estimate` block, the atlas package tree never learned about the new `internal/issueref` package, the plan's Core-concepts table says `refScanRE`/`spanRefRE` were *deleted* when both still exist, the end-to-end `Compute` test the plan named was not written, the evidence doc records only the drifting-`HEAD` `sdlc actual` rather than the fixed-window `active-time` run Task 5 Step 2 designated as *the* measurement, and `## Log` is empty. **Caveat on confidence:** every `Bash` invocation in this session was refused by the sandbox, so I verified by reading source and tests, not by running `go test`/`go vet`.

## 1. Strengths

- **The Spec supersede is the real win.** `workshop/issues/000190-…md:227-274` — recognizing that the commit path was poisoned too, and that a precedence rule "fixes nothing when both sides are poisoned by the same parse," is what made this a 20-line fix instead of an attribution-policy rewrite. `commit_test.go:99` pins exactly that path.
- **A test that pinned the bug was inverted, not quietly relaxed.** `cmd/sdlc/internal/gitx/window_test.go:47-56` records that the old row `{"prefix#42 suffix", []string{"42"}}` encoded the #190 defect as intended behavior, and says so inline. That is the right way to change an expectation.
- **The fragment-not-regexp export is the detail that makes it 5 → 1.** `issueref/ref.go:40` exports `QualifiedIDPattern` as a `const string` so `migrate.go:62`'s anchored `spanRefRE` composes instead of restating. I verified the group-index shift is inert: both migrate regexes are used only via `MatchString`/`FindAllStringIndex` (`migrate.go:104, 134, 142`), never `Submatch`. Zero migrate test edits is consistent with that.
- **`DiscoverWindowIssues` moved onto the package `run` shim** (`window.go:405`), with the rationale recorded (`RepoTopLevel` bypasses the shim, so an internally-resolved qualifier would have shipped with no guard). That seam correction is what made `TestDiscoverWindowIssuesExcludesForeignRefs` possible at all — ARCH-MOCK done right.
- **The `selfRepo ""` degradation is pinned honestly.** `util_test.go:63-66` and `window_test.go:74` both assert that an unresolvable self-repo makes every qualified ref foreign rather than guessing. That's the conservative direction.

## 2. Critical findings

None.

## 3. Important findings

**I1 — `workshop/issues/000190-…md:154-178`: the issue file is corrupted with a mangled heading and 24 orphaned lines of a superseded `## Estimate` table.**
Line 154 reads `## Revisions` because both its mechanism and its fix were wrong |`` — a heading spliced into the tail of a table cell — followed by lines 155-178, which are a stale duplicate of the estimate table and calibration prose that lines 122-152 already supersede. The real `## Revisions` is at :180. Failure scenario: this file is about to be archived to `workshop/history/` as the durable tracker record, and a reader (or a future `sdlc` section consumer) sees two estimate tables and a heading that is neither. It currently parses by luck — `internal/issue/section.go:15`'s `^## Revisions\s*\n` does not match line 154 because of the trailing junk, so `## Estimate` terminates at :154 and the real Revisions at :180 still resolves. Fix: delete lines 154-178.

**I2 — no end-to-end coverage of `Compute`'s own wiring (`compute.go:84`); the plan named this test and it wasn't written.**
Plan Task 3 Step 1 specifies `TestComputeDoesNotAttributeToForeignIssue` ("PerIssue["127"] == 0; PerIssue["187"] holds the whole segment"). It does not exist — `grep 'func Test' cmd/sdlc/internal/activetime` confirms. The three new activetime tests all call `parseEventMentions` / `loadWindowCommits` directly. Failure scenario: someone changes `compute.go:84` to `newMentionScope("", opts.Issues)` or drops `sc` from the `loadEventsWithFiles` call, and every test in the package still passes — the mention path silently reverts to counting bare refs only (or, with the wrong qualifier, re-admits foreign ones). The existing `Compute` tests can't catch it because their fixtures use bare `#N`, where the qualifier is irrelevant. Fix: add the planned test — a `gitRun` fixture with `#187 M2: pair#127 …`, a transcript event mentioning `pair#127`, `Issues: []string{"187","127"}`, `GitRepo: "/tmp/x/ariadne"`, assert `PerIssue["127"] == 0`.

**I3 — Docs update gate: `internal/issueref/` is missing from the atlas architecture tree.**
`atlas/workflow/sdlc-binary.md:357-372` enumerates every `internal/` package (`repolock/`, `gitx/`, `issue/`, `judge/`, `project/`, plus `activetime/` and `transcripts/` above), and a new one landed without an entry. Related, in the same file: the `gitx` line (:361-364) and the measured-actuals narrative (:584-610) still describe `DiscoverWindowIssues` without the `selfRepo` parameter or the local/foreign rule. Only `ledger-landscape.md` was updated, and its paragraph (:47-55) is accurate — including the `#174-#176` claim, which I traced and confirmed: `-` being a non-word char is what lets `\b` succeed after `#174`, so the leftmost match consumes it and `174-` is never read as a qualifier. Fix: add `internal/issueref/  new (#190): the one qualifier+id scan grammar (Find/IsLocal/LocalNums/CountLocal); parseRef stays the validator` to the tree, and a clause on the `gitx` line.

**I4 — Core-concepts table contradicts the code: `refScanRE` and `spanRefRE` are listed `deleted`, but both still exist.**
`workshop/plans/000190-…-plan.md:85-86` marks both `deleted`; `migrate.go:49` is `var refScanRE = issueref.ScanRE` and `migrate.go:62` is `var spanRefRE = regexp.MustCompile(...)`. The table also contradicts the plan's own Task 4, which says "Replace" and "Recompose," not "delete." I am not calling this Critical because the substance the table was tracking — one encoding of the grammar — *is* delivered; it is the status word that is wrong. Failure scenario: a future ARCH-DRY audit greps for `refScanRE`, finds it, and concludes the consolidation didn't land. Fix: change both rows to `modified`, plus the `## Revisions` entry in §7.

**I5 — `workshop/plans/000190-evidence.md`: the measurement the plan designated as primary isn't recorded.**
Task 5 Step 2 specifies a fixed-window `sdlc active-time --since 2026-07-29T10:00 --until 13:00 --issue 187 --issue 127 …` run — the invocation plan-quality round 1 corrected as PQ-3 precisely so the before/after would be like-for-like. The evidence file records only `sdlc actual --issue 187`, twice, over `f59f49cb → HEAD` — and `HEAD` moved between the runs, which the doc itself concedes ("2.29h → 2.32h → 2.83h → 3.83h … the earlier drift is the moving HEAD"). The `84.5m → 130.6m = +46.1m` segment comparison is sound, because those segment boundaries (`10:25 → 12:36`) are identical in both outputs — but the issue's Plan row claims "measured over a fixed window," and no fixed-window run appears. Fix: either run Step 2's command and paste it, or amend the evidence doc + Plan row to say the identical-boundary segment comparison replaced it and why.

**I6 — `## Log` is empty.** `workshop/issues/000190-…md:276-279` is a bare `### 2026-07-29` heading with nothing under it, while plan Task 6 Step 3 says "write the `## Log` entry" and AGENTS.md §3 requires the review outcome to be logged there. `sdlc close` will insert its own line, but the narrative entry (root cause, the `pair#129` bonus catch, the gate cost) exists only in `## Revisions`.

## 4. Minor findings

- `activetime.go:104` — `formatAttributionWarning` discriminates window-scoped from segment-scoped warnings by `Start.IsZero() && End.IsZero()`. An explicit field on `AttributionWarning` (`Kind`, or `WindowScoped bool`) would survive a future non-segment warning that happens to carry a time. The struct is internal; the change is cheap.
- `migrate.go:49` — `var refScanRE = issueref.ScanRE` is a vestigial alias. Its two call sites (`:104`, `:142`) can name `issueref.ScanRE` directly; the doc comment can move to the call site or drop.
- `foreignRefWarnings` is defined in `compute.go:140` but tested in `util_test.go:39`. Colocate.
- `scripts/close-issue.py:212` still carries `re.findall(r"#(\d+)\b", line)` in `discover_window_issues`, live on the pre-binary fallback (`Makefile.workflow:124`). It only prints a suggested `sdlc active-time` command line, so it can't corrupt a measured number — but the Revisions claim "no encoding of the qualifier+id grammar remains outside `cmd/sdlc/internal/issueref/ref.go`" is true of Go only. Scope the claim or fix the Python.
- `selfQualifier` is computed twice per `Compute` call — `compute.go:84` and again inside `loadWindowCommits` (`commit.go:53`). Same input, same answer; threading one derivation through would read better.
- `actual.go:101` — `filepath.Base(repoTop)` is the *directory* name. In a git worktree at `…/ariadne-190-something`, `ariadne#180` becomes foreign: dropped from the tracked set *and* reported as "another repo's issue." Rare (1 self-qualified ref in 400 subjects) but worth a sentence in the doc.
- `compute.go:160` — a `gh#42` ref renders as "foreign ref ignored — **another repo's issue**." Plan risk #4 correctly accepts gh refs as non-local, but the reason text mislabels a GitHub-inbox ref. One `if r.Qualifier == "gh"` branch fixes the wording.
- `window_test.go:11-16` splits `testfix` and `issueref` into two import groups; fold into one.
- `event_test.go:10, 51, 100, …` still name `mentionScope` values `pat` — a leftover from the `*regexp.Regexp` era.

## 5. Test coverage notes

Coverage of the *grammar* is strong and corpus-derived: `ref_test.go` covers every real repo-name shape in the workspace (`brain-family`, `parley.nvim`, `42shots`, `xianxu.dev`), the `#174-#176` range, the 7-digit rejection, exact-not-prefix `IsLocal`, and the anchored-composition case including the #179 corruption strings. `TestForeignRefWarningsNameTheDroppedRefs` covers the observability requirement including the `selfRepo ""` degradation. `TestSelfQualifierComesFromGitRepo` covers the inverse-direction bug that no test through `sdlc actual` could reach.

The gap is at the **seam**, not the leaves (I2): three of the four tests the plan specified for Task 3 exist; the fourth — the only one that would exercise `Compute`'s construction of `mentionScope` from `opts.GitRepo` — does not. `TestFormatAttributionWarningOmitsZeroTimeRange` is a good addition and correctly also asserts the *unchanged* segment-warning shape, so it can't pass by deleting the range.

I also checked the parity fixtures the plan flagged as a risk: `parity_test.go` carries no qualified refs (grep for `[A-Za-z0-9]#[0-9]` over `cmd/**/*_test.go` returns nothing from it), so the v3 golden is genuinely unaffected — the plan's risk #1 holds.

## 6. Architectural notes

- **ARCH-DRY — pass, with one residue.** Five Go encodings collapsed to one, and the fragment export is what made it 1 rather than 2. Flagged: `scripts/close-issue.py:212` is a sixth encoding still carrying the unbounded pattern (Minor above), and `refScanRE` survives as an alias.
- **ARCH-PURE — pass, with a note.** `issueref` is genuinely pure (no IO, no clock, no git) and its tests run without fakes. `mentionScope` and `foreignRefWarnings` are pure and unit-tested directly. Note: `selfQualifier` (`commit.go:87`) touches the environment (`filepath.Abs` reads cwd, `expandUser` reads `$HOME`) and sits inside what `Compute`'s doc calls the pure region. It's defensible — `--git-repo .` *needs* cwd resolution — but note the asymmetry with `gitx.DiscoverWindowIssues`, which takes `selfRepo` as an explicit parameter for exactly the testability reason. `TestSelfQualifier:143-147` has to weaken its `"."` assertion because of this, which is the tell. Worth one sentence in `selfQualifier`'s doc naming the asymmetry as deliberate.
- **ARCH-PURPOSE — pass.** Shadow-sweep of the consumers: `gitx.DiscoverWindowIssues` ✓ derives; `activetime` commit path ✓ derives; `activetime` mention path ✓ derives; `migrate.refScanRE` ✓ derives; `migrate.spanRefRE` ✓ composes from the fragment; `parseRef` (`resolve.go:57`) — a hand-written parser, but explicitly the declared canonical *validator* with the division of labor documented on both sides, so it's a deliberate boundary, not a deferred consumer; `scripts/close-issue.py` — the one remaining hand-maintained restatement, deprecated path, Minor. The issue's stated purpose (the 46.1m returns to #187) is delivered and measured, not deferred.
- **ARCH-MOCK — pass.** No new external dependency. Both git surfaces stay behind their existing shims, and this diff *improved* the seam by moving `DiscoverWindowIssues` off a direct `exec.Command` onto the package `run` shim. Production and test flow share the same boundary in both packages.

## 7. Plan revision recommendations

Add to `workshop/plans/000190-cross-repo-refs-parsed-as-local-plan.md`'s `## Revisions`:

1. **Core-concepts table: `refScanRE` / `spanRefRE` are `modified`, not `deleted`** (I4). Both identifiers survive in `migrate.go` — `refScanRE` as an alias to `issueref.ScanRE` (:49), `spanRefRE` as a recomposition from `QualifiedIDPattern` (:62). Task 4's own Steps 1-2 say "Replace"/"Recompose," so the table was already inconsistent with the task at plan time. Update rows 85-86 to `modified`.
2. **Task 3 Step 1: `TestComputeDoesNotAttributeToForeignIssue` was not written** (I2) — either record it as delivered once added, or record the deliberate omission and what covers `compute.go:84` instead.
3. **Task 5 Step 2: the fixed-window `sdlc active-time` run is not in the evidence** (I5) — record either that it was run and where the output is, or that the identical-boundary segment comparison from `sdlc actual` was substituted, and why that is equivalent. The issue's Plan row "measured over a fixed window" needs the same correction.
