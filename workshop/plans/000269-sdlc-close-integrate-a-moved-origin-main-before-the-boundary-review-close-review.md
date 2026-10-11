# Boundary Review — ariadne#269 (whole-issue close)

| field | value |
|-------|-------|
| issue | 269 — sdlc close: integrate a moved origin/main before the boundary review |
| repo | ariadne |
| issue file | workshop/issues/000269-sdlc-close-integrate-a-moved-origin-main-before-the-boundary-review.md |
| boundary | whole-issue close |
| milestone | — |
| window | 73b717f853f34208840eada67b7614653894ba32..92b6bf38fe85138a0b0a28547e3e607092d4c52e |
| command | sdlc close --issue 269 |
| reviewer | claude |
| timestamp | 2026-10-11T01:06:27-07:00 |
| verdict | SHIP |

## Review

Reviewing the #269 integrate gate diff; now checking the git runner's error wrapping and the tests.

```verdict
verdict: SHIP
confidence: high
```

**Verdict: SHIP.** The gate does what the Spec, the Done-when and the durable plan say. Before any review cost, `computeClose` refuses a branch missing main's code from before the most recent noon. The remedy is a rebase for an unpushed branch and a merge for a pushed one. Docs-only staleness passes with an info line. A failed fetch prints "could not check … NOT verified as up to date" and never the up-to-date line. `--no-integrate` works on both `close` and `milestone-close`, `--force` also skips the gate, and the flag is in the gatesig catalog and the help text. I ran the targeted tests at HEAD `92b6bf38` and they pass: `internal/freshness`, `processmanual`, and in `cmd/sdlc` the `TestIntegrateGate*`, `TestCloseRefusesAStaleBranchBeforeTheReview`, the skip tests and the flag-registration tests. Only Minors remain, and none blocks the boundary.

**Strengths**
1. **Pure core and thin shell (ARCH-PURE).** `freshness.Cutoff` is pure. `Check` takes a narrow `Git` seam plus the caller's `isCode` predicate. `integrategate.go` only resolves main, loads the zone and renders the result. `formatIntegrateRefusal` is pure.
2. **Real git instead of mocks (ARCH-MOCK).** Every probe test uses a real repo with a bare remote and a peer clone. Each remedy is run from the refused state and then passes (`integrategate_test.go`, the rebase and merge tests). The tracker test covers both the pinned-SHA probe and an unreachable origin.
3. **Three outcomes are honest.** `Check` treats only exit code 1 of `merge-base --is-ancestor` as "not contained"; any other failure is Inconclusive (`freshness.go:120-125`). A landing that was later reverted counts as Fresh, not as DocsOnly with zero files.
4. **The cutoff is tested at the edges.** `TestCutoff` covers DST start and end, exactly noon, after midnight, a year boundary and a zone comparison. The first-parent test proves an old commit merged after the cutoff is not yet required.
5. **The code predicate is shared (ARCH-DRY).** The gate reuses `publishGateHasCodeSurface`, so it agrees with the publish gate on what counts as code; `TestIntegrateGate_EmbeddedMarkdownIsCode` pins that embedded help text counts as code.

**Critical findings:** none.

**Important findings:** none.

**Minor findings**
- `integrategate.go:76`: outside tracker repos, `git fetch` runs through `exec.Command` with no context or timeout. The `ctx` passed in is ignored there, so a hung SSH remote stalls `close` with no limit (ARCH-CONSTRAINTS). Fix: pass `ctx` to the fetch, or put a deadline on it.
- `integrategate.go:84`: `integrateBranchPushed` treats any `@{upstream}` as "pushed". An unpushed issue branch created with `-b x origin/main` tracks main itself, so it gets the merge remedy instead of the rebase one. Fix: ignore an upstream equal to main's ref.

**Test coverage notes:** the tests cover every outcome:
- fresh
- stale, unpushed and pushed
- docs-only
- embedded markdown counted as code
- fetch failure
- both bypass flags
- a bad timezone file
- `computeClose` wiring
- a tracker repo
- `milestone-close` forwarding the flag

`TestCheck` covers the probe logic directly. The log records a full-suite run with one flaky test that it attributes to load (the #325 class).

**Architecture**
- **ARCH-DRY:** pass. The predicate is shared, and tracker repos share the review window's cached fetch through `reviewMainRef`.
- **ARCH-PURE:** pass.
- **ARCH-PURPOSE:** pass. The gate covers both boundaries, and the `freshness` package is ready for #303 to call.
- **ARCH-MOCK:** pass. Tests use real git with no function-call mocks.
- **ARCH-CONSTRAINTS:** a Minor flag for the unbounded fetch above. Otherwise the cost is seconds, ahead of a review that takes minutes.
- **ARCH-SECURE:** pass. A malformed or unknown zone in `.sdlc/freshness.json` refuses visibly instead of falling back.
- **ARCH-ORDER:** pass. The gate holds no state between events; a main that moves after the cutoff cannot move the requirement backwards.
- **ARCH-FUNERAL:** pass. Nothing durable is created besides one static config file.

**Architectural notes for upcoming work:** #303 should call `freshness.Check` with the same predicate and the same way of resolving main. It's worth moving `integrateMainRef` somewhere both callers can reach, so claim/start-plan don't copy the fetch logic.

**Plan revision recommendations:** none. The existing Revisions entry already matches the code.

```findings
findings:
  - id: new
    severity: Minor
    family: unbounded-external-call
    title: |
      Legacy-repo integrate fetch runs without a context or timeout
    detail: |
      integrateMainRef's legacy path calls gitx.RunGit("fetch", ...) through exec.Command with no deadline and ignores the ctx it receives, so an unresponsive remote hangs close indefinitely (ARCH-CONSTRAINTS). Pass ctx to the fetch or give it a deadline.
  - id: new
    severity: Minor
    family: push-state-heuristic
    title: |
      integrateBranchPushed treats any @{upstream} as pushed
    detail: |
      An unpushed issue branch created with -b x origin/main tracks main as its upstream, so the gate names git merge instead of git rebase. Ignore an upstream equal to main's ref.
```

---

## Re-review — 2026-10-11T01:15:24-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 269 — sdlc close: integrate a moved origin/main before the boundary review |
| repo | ariadne |
| issue file | workshop/issues/000269-sdlc-close-integrate-a-moved-origin-main-before-the-boundary-review.md |
| boundary | whole-issue close |
| milestone | — |
| window | e4104a7602052bfbf5412a4659e41bcbb242b5f0..4687acd352ca33175804525cf91d8e913809d9da |
| command | sdlc close --issue 269 |
| reviewer | claude |
| timestamp | 2026-10-11T01:15:24-07:00 |
| verdict | SHIP |

## Review

Still on the fix review. The diff is read and I'm waiting for the targeted integrate-gate tests to finish.

```verdict
verdict: SHIP
confidence: high
```

Both prior Minors are fixed at the class level, and each fix has a regression test that targets it. In legacy repositories the fetch now runs under a context with a deadline built from the caller's `ctx`. The push-state check now ignores an upstream that is main itself. I ran the targeted tests at HEAD `4687acd3` with `go test ./cmd/sdlc/ -run TestIntegrateGate -count=1` and they pass (`ok … 211.6s`). The merge of origin/main (`4687acd3`) brings in only main's own files, and none of them overlap this window. No conflict resolution needed review, and the interdiff makes no scope change.

**Strengths**
1. `integrategate.go:82-86`: the timeout derives from the caller's `ctx`, so a cancelled close also stops the fetch. A hang becomes the honest "could not check" line instead of a stall. `RunGitContext` reuses the existing `runGitInContext`, which kills git's whole process group, rather than adding another way to run git (ARCH-DRY).
2. `TestIntegrateGate_FetchTimeoutIsInconclusive` uses a real `uploadpack` script that hangs, with no function mock. It would hang or exceed its 20s bound if the deadline were missing.
3. `integrategate.go:96`: the check compares the upstream's abbreviated name (`origin/main`) with `mainName`, which is the same form `TrunkRef` returns. The local-main case (`main` == `main`) is handled the same way.
4. `TestIntegrateGate_UpstreamOfMainIsNotPushed` sets the upstream to `origin/main` and asserts the rebase remedy. Before the fix, that branch would have been told to merge.

**Critical findings:** none.
**Important findings:** none.
**Minor findings:** none new. In a tracker repository with no trunk, `mainName` is a commit SHA, so the upstream comparison can never match. That edge already falls back to the remote-branch check, so it does no harm.

**Test coverage notes:** each fix has a test that targets it. The timeout test tolerates a slow machine (a 500ms deadline against a 20s bound).

**Architecture (interdiff only)**
- **ARCH-DRY:** pass. It reuses `runGitInContext` and `firstLine`.
- **ARCH-PURE:** pass. The changes stay inside the thin IO shell.
- **ARCH-PURPOSE:** pass.
  - `unbounded-external-call`: the only external call this window adds, the legacy fetch, is now bounded. The tracker path's fetch goes through `openTracker(ctx)`, which already carries `ctx`.
  - `push-state-heuristic`: the heuristic had two signals, and the upstream signal is the one the finding targets. The other signal, a remote branch with the same name, already compares against the branch itself.

**Plan revision recommendations:** none.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      Fetch now runs via gitx.RunGitContext under context.WithTimeout(ctx, integrateFetchTimeout); TestIntegrateGate_FetchTimeoutIsInconclusive pins it with a hanging uploadpack.
  - id: BR-2
    disposition: addressed
    note: |
      Upstream is compared by abbrev-ref against mainName and ignored when equal; TestIntegrateGate_UpstreamOfMainIsNotPushed asserts the rebase remedy.
```
