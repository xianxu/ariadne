# Boundary Review — ariadne#148 (whole-issue close)

| field | value |
|-------|-------|
| issue | 148 — sdlc merge: guard against a reused branch name silently skipping unmerged commits |
| repo | ariadne |
| issue file | workshop/issues/000148-sdlc-merge-guard-against-a-reused-branch-name-silently-skipping-unmerged-commits.md |
| boundary | whole-issue close |
| milestone | — |
| window | 85928d8a1ea07276fb44d9c7fd0d6cf06eaf9a6a..HEAD |
| command | sdlc close --issue 148 |
| reviewer | claude |
| timestamp | 2026-07-05T21:00:34-07:00 |
| verdict | unknown |

## Review

The full `cmd/sdlc` package test run just completed with **exit code 0** — the whole package is green, confirming the evidence behind my review. My verdict stands: **FIX-THEN-SHIP** (high confidence). The shipped guard is correct and safe; the two non-blocking Important findings are the unpinned `git fetch` in `_FinishesCleanup` (I1) and the stale `merge.md` help text (I2), both cheap to address.

---

## Re-review — 2026-07-05T21:09:28-07:00 (FIX-THEN-SHIP)

| field | value |
|-------|-------|
| issue | 148 — sdlc merge: guard against a reused branch name silently skipping unmerged commits |
| repo | ariadne |
| issue file | workshop/issues/000148-sdlc-merge-guard-against-a-reused-branch-name-silently-skipping-unmerged-commits.md |
| boundary | whole-issue close |
| milestone | — |
| window | 8d0bb6f06792328ea7c6405975f6b9a65237f531..HEAD |
| command | sdlc close --issue 148 |
| reviewer | claude |
| timestamp | 2026-07-05T21:09:28-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

I have everything I need. Let me confirm the pre-guard context: `remoteRef` is verified-to-exist and HEAD-synced before the guard (lines 307–322), so the count's head is valid and fails safe if not.

Let me write up the review.

```verdict
verdict: FIX-THEN-SHIP
confidence: medium
```

The #148 guard itself is **ship-quality**: correct, pure-core/thin-IO, well-tested, and fully documented. `decideMergeAction` stays a pure decision (extended with `unmergedCount`), the count IO is isolated in a fakeable `countUnmerged` seam, the caller fetches `origin/main` before counting (staleness handled), fails safe on any fetch/count error, and refuses **before** any switch/pull/archive/delete. Unit tests pin the decision table and the count seam; the two e2e tests exercise real git for both arms, and the `_FinishesCleanup` test now genuinely pins the guard's fetch. I ran `go vet` (clean) and the merge tests (all PASS) rather than trusting the Log. What keeps this from a clean SHIP is **not** the #148 code — it's that the review window and branch are polluted by out-of-scope `seed`/`work` junk commits, which (a) push the window base *past* the implementation commit so this mandatory boundary review's configured window never sees the guard, and (b) carry a whole in-tree `origin.git/` bare repo + scratch files toward main. These look like test-sandbox scaffolding; the main agent has the session context to confirm and clean before recording the close.

### 1. Strengths
- **ARCH-PURE, textbook.** `decideMergeAction(openPR, mergedExists, unmergedCount)` (`cmd/sdlc/merge.go:238`) is a pure switch; the git IO lives in `countUnmerged(r gitRunner, …)` (`merge.go:224`) behind the injected runner. The "pure decision + fakeable seam + thin caller" split is exactly right, and the caller (`merge.go:457-467`) is the minimal glue.
- **Fail-safe is real, not decorative.** Both the fetch error (`merge.go:460`) and the count error (`merge.go:464`) `die` rather than defaulting to 0 — so an unverifiable branch is never cleaned up. `countUnmerged` also errors on non-numeric output (`merge.go:229`), and `TestCountUnmerged` pins that path.
- **Refusal ordering is correct.** `actionResumeBlocked` dies at `merge.go:489` *before* the `if merged {` block (`merge.go:496`) and the archive (`merge.go:536`), so nothing is switched/deleted — and the e2e test asserts tree-untouched at four points (still on feature, branch present, issue not archived, not in history).
- **The I1 fetch-pin is genuinely load-bearing.** `merge_e2e_test.go:317-323` forces `refs/remotes/origin/main` stale after the push, so the test now passes *only* if the guard's fetch refreshes it. That closes the "test passes even with a broken fetch" gap the earlier review flagged. Verified the reasoning holds: stale tracking ref → count 1 → refuse → test fails.
- **Docs are complete for the surface.** root.md PUBLISH note, merge.md REFUSES-IF entry, and `atlas/workflow/sdlc-binary.md` all describe the guard. No new subcommand/flag, so README needs nothing — the help-text update is the right home.

### 2. Critical findings
None in the #148 code.

### 3. Important findings

- **Boundary-review window excludes the implementation (base landed on a junk commit).** The configured window is `8d0bb6f..HEAD`, but `8d0bb6f` is a `"work"` junk commit that sits *after* the real implementation `05a9e14` ("#148: guard sdlc merge…"). So the window this mandatory close-review is pointed at contains only the fix-findings commit + docs — the guard code, `decideMergeAction`, `countUnmerged`, and their unit tests are *not* in it. I reviewed the full issue range (`85928d8..HEAD`) manually and it's sound, but the base computation is wrong: the review that gates the close doesn't cover the code it's gating. *Fix:* recompute the boundary base to the actual branch point (`85928d8`, matching the prior close-review.md's own `window` field) after resolving the junk commits below, so the recorded verdict reflects a review of the implementation.

- **The branch carries non-#148 pollution headed for main.** `git ls-tree HEAD` shows an entire in-tree `origin.git/` bare repo (18 files: HEAD, config, hooks/*, …) plus an `f` file, introduced by the `seed` (`1727757`) and `work` (`8d0bb6f`) commits — which also dragged in unrelated docs (`…000159-alt-shift-c-continuation-improvements.md`, a pensive note, a 1021-line `…agentic-process-documentation-strategy.md`). None of this is issue #148. It looks like sandbox/e2e scaffolding (and `main` itself is at `8d0bb6f`, so it's already there, not newly introduced by this close). *Fix:* the main agent should confirm whether this is a known test sandbox; if this is the real ariadne branch, these commits must be dropped/`.gitignore`d before the close, since committing a bare git repo into the tree will confuse tooling and bloat history. Flagging because a fresh reviewer cannot tell scaffolding from an accident, and the checklist asks about production-readiness of what crosses the boundary.

### 4. Minor findings
- `countUnmerged` hardcodes the base to `"origin/main"` at the call site (`merge.go:462`) and the message hardcodes "not in main" (`merge.go:490`). Correct today because `sdlc merge` always targets main, but if a non-main base ever becomes possible this is the spot that silently assumes otherwise. Fine to leave; note for future.
- `git fetch origin main` now makes the *resume-cleanup* path require network, where a fully-local resume previously did not. This is the intended fail-safe (per Plan), but it's a behavior change worth a one-line mention in the atlas/help if offline resume was ever relied on. Non-blocking.

### 5. Test coverage notes
Coverage is strong and pins real logic, not mocks reasserting the implementation:
- `TestDecideMergeAction` gained the two decisive rows — `{"", true, 16, actionResumeBlocked}` and `{"", true, 0, actionResume}` — plus "open PR wins even with unmerged commits" guarding the switch ordering.
- `TestCountUnmerged` covers numeric parse, `0`, git-error propagation, and non-numeric → error (all four fail-safe branches).
- `TestRunMerge_ResumeMergedPR_UnmergedCommits_Refuses` is the exact regression for parley.nvim#116 (real git graph disagrees with a merged-PR stub → abort + tree untouched). This is the kind of bug the diff could ship, and it's covered.
The one gap: no test exercises the **fetch-error → die** caller branch (`merge.go:460`) at the e2e level — only the count-error path is implicitly reachable. Low value to add given the pure seam is tested; noting only.

### 6. Architectural notes for upcoming work
- **ARCH-DRY — pass (with a noted, defensible divergence).** The pre-existing ahead-of-upstream check at `merge.go:314` also runs `rev-list --count` but via `gitx.Capture` (unfakeable). The Log consciously keeps it separate because `countUnmerged` is the fakeable home for the *new* guard's count. That's the right call — they serve different purposes and only one needs injection; consolidating would drag the unfakeable global into the pure seam. No action.
- **ARCH-PURE — pass.** Business logic (the decision + the parse) is pure and unit-tested without IO; the only IO (fetch + rev-list) is the injected `mergeRunner`. Nothing to extract.
- **ARCH-PURPOSE — pass.** Shadow-sweep of the "reused branch name" purpose: the guard refuses with the actionable rename→`sdlc pr`→`sdlc merge` message, leaves the tree untouched, AND both doc consumers derive the behavior (root.md PUBLISH, merge.md REFUSES-IF, atlas). This delivers the issue's actual purpose (refuse loudly + guide recovery), not a cheap subset. The Done-when bullet requiring *automated* both-branch coverage is met (not left as a manual step).

### 7. Plan revision recommendations
None for the #148 plan itself — the Core-concepts entities (`actionResumeBlocked`, `decideMergeAction` extended, `countUnmerged` new, fake-runner test, e2e both-arms, root.md note) all exist at their stated paths and match the code. The plan and code are in agreement.

The only "revision" worth recording is procedural, not in the plan: the junk `seed`/`work` commits and the resulting mis-set boundary base (findings §3) should be resolved before the close verdict is trusted — that's a branch-hygiene action for the main agent, not a plan edit.
