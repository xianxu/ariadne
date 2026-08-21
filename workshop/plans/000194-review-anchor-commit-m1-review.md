# Boundary Review — ariadne#194 (milestone M1)

| field | value |
|-------|-------|
| issue | 194 — boundary reviews: anchor to the reviewed commit, and remember across rounds |
| repo | ariadne |
| issue file | workshop/issues/000194-review-anchor-commit.md |
| boundary | milestone M1 |
| milestone | M1 |
| window | 9cb22e79b3ae4df8838ecf57a412e743c56581ed^..6e4662276a018cee0361fd9aa431a7b515ff2f17 |
| command | sdlc milestone-close --issue 194 --milestone M1 |
| reviewer | claude |
| timestamp | 2026-08-20T17:46:16-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

M1 delivers what Spec A and F promised: the review head is resolved once under the repo lock and spent by every downstream consumer (diff, prompt, `Review-Window:` trailer, sidecar, finalize check, dispatch info line) — I grepped `cmd/` and `pkg/` and no literal `"HEAD"` survives on the boundary-review path, so the ARCH-PURPOSE shadow-sweep is genuinely complete rather than partially done. The delta classifier is well-shaped: pure decision logic, a thin git shell, real reuse of `publishGateHasCodeSurface`, and a fourth `diverged` state whose necessity is argued from `git diff A B`'s behavior on unrelated commits. `go build ./... && go test ./cmd/sdlc/... ./pkg/...` passes clean; `gofmt -l` is empty. Nothing here is a correctness bug that ships broken behavior, so nothing blocks. What does need fixing before the boundary is a cluster of contract-accuracy gaps: an atlas paragraph that now states the *old* rule as fact, helptext that promises more safety than the code delivers, and a refusal path that lost its next-action instruction — all three are the kind of thing that misleads the next agent, which in this repo is the primary consumer.

## 1. Strengths

- **`reviewanchor.go:56-92` — the four-state design and its justification.** `anchorDiverged` exists because `git diff A B` between unrelated commits returns paths perfectly happily, so a rebase-away would otherwise classify as doc-only and finalize. That reasoning is stated in the code and pinned by a named fixture (`reviewanchor_test.go:35`). This is the non-obvious correctness argument in the milestone and it was found, not missed.
- **`reviewanchor.go:65` — ARCH-DRY reuse, verified.** `classifyReviewAnchor` *calls* `publishGateHasCodeSurface` rather than restating the docs-vs-code rule, so the close gate and the publish gate cannot drift. The consequence that matters (`cmd/sdlc/helptext/*.md` is `//go:embed`ed, therefore code) is pinned at `reviewanchor_test.go:26-31`.
- **`state.go:369-383` + `TestAbbrevSHA_DoesNotResolveSymbolicRefs` — a real trap caught.** Noticing that `shortSHA` *resolves* its argument, so `shortSHA("HEAD")` would print the ambient repo's HEAD — a commit the review never read — and routing the trailer through the pure `abbrevSHA` so a degraded window stays visibly `..HEAD`. The planned duplicate `abbrevSHA` was dropped in favor of the existing one instead of being written twice.
- **`close_finalize_test.go:475` — ARCH-MOCK-clean interleaving tests.** Both new integration tests block the fake reviewer on a channel and land a *real* commit through the real `git` binary in the hermetic repo, exactly as the plan required. No second git double was introduced.
- **`close.go:1224-1234` — the strict/relaxed split is drawn at the right line.** Only the HEAD check was loosened; the issue-file and project-file checks stay byte-exact, which is what protects `applyClose`'s lost-update hazard on `r.issueText`.

## 2. Critical findings

None.

## 3. Important findings

**I1 — `atlas/workflow/sdlc-binary.md:88` now states the old rule as current fact (docs gate).**
The paragraph reads: *"…reacquire before finalization and refuse to write if HEAD, the issue file, or any prepared project file edit changed while the lock was released."* That is precisely the behavior this diff replaced. `git diff --name-only <base>..HEAD -- atlas/` shows only `ledger-landscape.md` changed in this window, so the file documenting the close/milestone-close lock protocol was left asserting the superseded contract — worse than a missing update, because a reader trusts it. *Fix:* amend to "…refuse to write if the issue file or any prepared project-file edit changed, or if the commits landing during the review carry code surface (#194); a doc-only delta finalizes."

**I2 — `cmd/sdlc/helptext/close.md:55` over-claims what is safe to commit mid-review.**
*"So committing lessons/atlas/plan bookkeeping during a review is safe."* Two of the most common close-time bookkeeping targets are not safe: an edit to the **issue file** (a `## Log` line — the single most likely mid-review write) and to a **project file** both still refuse, via the deliberately-strict checks at `close.go:1231-1246`. `TestCloseCommands_IssueChangedDuringBoundaryReview_DoesNotFinalize` proves it. This is embedded, shipped, agent-facing text in the base-layer repo, so the inaccuracy propagates downstream. *Fix:* narrow the sentence — e.g. "…committing `workshop/lessons.md`, `atlas/`, and `workshop/plans/` bookkeeping during a review is safe. The **issue file** and any **project file** are still snapshotted byte-exact — edit those after the close, not during."

**I3 — the non-anchor refusal paths lost their next-action instruction (`close.go:1149`).**
Plan Task 1.2 Step 7 dropped `cwarn(… "re-run \`%s\` so the review covers the current repo state")` on the grounds that `formatAnchorRefusal` carries a precise instruction. But that `cwarn` was **unconditional**, while `formatAnchorRefusal` is reached only for `anchorCodeDelta`/`anchorDiverged`. The issue-file, project-file, and git-error branches now surface as a bare `… close NOT finalized: workshop/issues/000069-x.md changed` with no next step. AGENTS.md §5 makes "errors are next-action specs" an explicit property of this gate. *Fix:* re-add the `cwarn` guarded on the anchor branches not having supplied one, or (simpler) have the three non-anchor branches append `— re-run \`<verb>\``.

**I4 — the milestone's stated "actual defect fix" has no test (`milestoneclose.go:607`).**
Plan Step 3 calls passing `p.Head` (not `"HEAD"`) into `collectDiff`/`PromptInput` *"the actual defect fix"* — it closes the lock-release window where the reviewed diff could name a different commit than the snapshot pinned. Nothing asserts it. `TestRunCloseWithReview_IssueClose_Dispatches` was updated to check the **trailer**, not the dispatched diff; the two interleaving tests commit after `collectDiff` has already run, so they don't exercise the race either. A regression to `collectDiff(..., "HEAD", ...)` would pass the entire suite. *Fix:* a direct test of `boundaryReviewDispatchOptions` with `Head` set to an older commit, asserting the returned diff excludes the newer commit's content.

**I5 — the #172 no-collision constraint is hand-restated instead of derived (`reviewanchor_test.go:56`, ARCH-DRY).**
`TestFormatAnchorDocsOnly_SharesNoRefusalVocabulary` checks a hand-written list of five forbidden words. The repo already has the derived form: `assertNoGatesigCollision` (`close_atlasskip_test.go:15`) matches a rendered line against **every** `processmanual.GateCatalog` Ack/Refusal pattern, ANSI-stripped, and is used by the two sibling info lines (#177 `atlasAutoSatisfyLine`, #178 `formatAdoptLine`). `formatAnchorDocsOnly` is a new unconditional `cinfo` on the close path — the same class — and skips it. I checked manually and there is no live collision today, so this is a missing guard rather than a bug; but a hand list cannot track a future catalog row. *Fix:* add `assertNoGatesigCollision(t, "\x1b[1;36m==>\x1b[0m "+formatAnchorDocsOnly(d))` alongside the existing word check.

**I6 — the plan's Core-concepts "Pure entities" table misclassifies two rows (plan revision needed).**
`workshop/plans/000194-review-anchor-commit-plan.md:182-183` lists `closeReviewSnapshot` and `resolveReviewWindow` under **Pure entities**. Neither is: `closeReviewSnapshot.validate()` does `os.ReadFile` plus git via `gatherReviewAnchorDelta`, and `resolveReviewWindow` now shells `gitx.Capture("rev-parse","HEAD")` directly (this diff made it *more* impure) on top of `boundaryWindowBase`. `TestResolveReviewWindow_HeadIsConcreteSHA` requires a real temp git repo (`windowRepo`) to run at all. The review checklist directs Critical for a table/code contradiction; I am reporting Important instead and flagging the judgement explicitly so you can override: the *code* satisfies ARCH-PURE — the decision logic (`classifyReviewAnchor`, both formatters) really is pure and really does unit-test with zero IO — and both rows were mislabeled before this diff touched them. It is the plan artifact that is wrong, not the architecture. *Fix:* move both rows to the "Integration points" table.

## 4. Minor findings

- `reviewanchor.go:78` + `close.go:1224`: when `resolveReviewWindow` degrades `head` to the literal `"HEAD"`, the anchor check becomes **silently inert** — `gatherReviewAnchorDelta("HEAD")` resolves the symbolic ref, so `HEAD..<current>` is always empty and every case classifies as `anchorDocsOnly`, printing the false-reassuring `anchored to HEAD; 0 doc-only commit(s) arrived since`. Narrow (needs `rev-parse` to fail at capture and succeed at validate), but a `0 commit(s)` docs-only pass is by construction impossible otherwise — cheap to guard: treat a non-SHA `reviewed` as unanchored and `cwarn` rather than `cinfo`.
- `reviewanchor.go:130`: `"%d code commit(s) landed after…"` counts **all** commits in the range, not only the code-bearing ones — a mixed doc+code delta reports "3 code commit(s)" when one touches code. The `code surface:` line below it is accurate; the count label over-claims.
- `close.go:477`: the atlas gate still uses literal `"HEAD"` while the review uses the pinned SHA, yet `resolveReviewWindow`'s comment claims the two "provably cover the same commits." True only because both run under one lock — i.e. incidental, which is exactly the property #194 argues against elsewhere. Passing the pinned head there would make it structural.
- `workshop/plans/…-plan.md:276` and issue `## Spec` A: "67 of the 70 archived sidecar files record `<base>..HEAD`". Measured: 66 files / 84 rows carry `..HEAD` (70 archived, 86 window rows — those two are right). This is the plan-gate's PQ-7 residual, disposed as cosmetic in round 3; it's the last inaccurate number left and is a one-token fix.
- Plan Task 1.1 skips **Step 6** (numbering jumps 5 → 7).
- All 38 checkboxes in the durable plan are still `- [ ]` after M1 landed. The issue's `## Plan` M1 row is correctly ticked, so the close gate is satisfied, but the plan itself no longer says which steps are done.
- `Review-Window:` renders base via `shortSHA` (git's minimal-unique length, typically 7) and head via `abbrevSHA` (fixed 8) — cosmetically asymmetric.

## 5. Test coverage notes

Coverage of the pure layer is good: `TestClassifyReviewAnchor` tables all four outcomes plus the embedded-helptext-is-code case, and both formatters are pinned. The interleaving tests use real git and real commits. Three gaps, in priority order:

1. **The dispatched-diff pinning (I4)** — the milestone's headline fix, untestable-by-nothing today.
2. **`anchorDiverged` against real git** — asserted only at the classifier level with a hand-built struct. The branch that actually decides it is `gitx.RunGit("merge-base","--is-ancestor",…)` returning non-zero (`reviewanchor.go:87`), and that line is never executed by any test. It guards the scariest failure mode (finalize after a history rewrite). A `git reset --hard HEAD~1 && git commit` inside the existing channel-blocked harness would cover it in ~15 lines.
3. **The sidecar's concrete SHA** is a named Done-when item with no assertion — every `Head:` in `reviewsidecar_test.go` is fixture *input* (legitimately unchanged per the plan's trap (b)), so the `| window |` row's real value is unverified. It flows from the same `p.Head`, so this is low-risk; folding an assertion into the I4 test would close it for free.

Note also that the error branches of `gatherReviewAnchorDelta` (`log …: %w`, `diff …: %w`) are unreachable in tests — acceptable, since they fail closed.

## 6. Architectural notes for upcoming work

- **ARCH-DRY: pass**, with I5 as the one flag. The `publishGateHasCodeSurface` reuse, the `closeVerb` reuse for the re-run verb, and dropping the planned duplicate `abbrevSHA` are all the right calls. `formatAnchorDocsOnly` vs `formatPublishGateDocsOnly` reads as duplication but is deliberate and correct — gatesig attribution *requires* the two gates' lines be distinguishable.
- **ARCH-PURE: pass** on the code (flag is against the plan table, I6). `reviewanchor.go` is the shape the rest of `cmd/sdlc` should move toward: gather → classify → format, with git confined to `gather`.
- **ARCH-PURPOSE: pass** on the shadow-sweep (no hand-maintained restatement of the reviewed head remains), flagged on the docs (I1, I2) — the surfaces that *describe* the new contract to its consumers are where the purpose is currently under-delivered.
- **ARCH-MOCK: pass.** No new double; the hermetic real-git repo and the `judge.Run` override are reused as the plan required. Worth noting for M2/M3: `judge.Run` is a *stateless* override, which is adequate while a boundary review is a single call, but M2 makes the reviewer stateful across rounds (ledger in → dispositions out). At that point a stateless canned-output override stops modeling the dependency, and the fake should carry round state.
- **For M2 specifically:** a doc-only commit that lands during the *whole-issue* close review escapes every review window permanently — the next window base is the close commit, which sits above it. That is the same tradeoff #174 already accepted post-close, so it isn't a regression, but M2's ledger gives you a place to *record* it if you want the property to be visible rather than inferred.
- **For friction measurement:** the anchor refusal is a new close-refusal signal with no `--no-<gate>` flag, so it has no `GateCatalog` row and will be unattributed in #172 accounting. That's arguably correct (it isn't a bypassable gate), but it's a deliberate choice worth a line in `## Log` rather than something discovered later.

## 7. Plan revision recommendations

Add a `## Revisions` entry to `workshop/plans/000194-review-anchor-commit-plan.md`:

> **### 2026-08-20 (M1 close review) — Core-concepts classification corrected**
> **Reason.** The M1 boundary review flagged that the Core-concepts table lists `closeReviewSnapshot` and `resolveReviewWindow` under **Pure entities**, but neither is pure: `closeReviewSnapshot.validate()` does `os.ReadFile` plus git via `gatherReviewAnchorDelta`, and M1 added a direct `gitx.Capture("rev-parse","HEAD")` to `resolveReviewWindow` on top of its existing `boundaryWindowBase` call — `TestResolveReviewWindow_HeadIsConcreteSHA` needs a real git repo to run.
> **Delta.** Both rows move to **Integration points**, with `gatherReviewAnchorDelta` named as the injected seam for the first and `boundaryWindowBase` for the second. The genuinely pure entities of M1 — `reviewAnchorDelta`, `anchorOutcome`, `classifyReviewAnchor`, `formatAnchorDocsOnly`, `formatAnchorRefusal` — stay, and are confirmed unit-tested with no IO.
> Also corrected: the sidecar count (66 files / 84 window rows carry `..HEAD`, of 70 archived files / 86 rows — closing the PQ-7 residual), and Task 1.1's missing Step 6 renumbered.

Separately, tick M1's Task 1.1 and 1.2 step boxes so the plan records what landed.
