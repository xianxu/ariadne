# Boundary Review — ariadne#578 (whole-issue close)

| field | value |
|-------|-------|
| issue | 578 — Train repos: pre-close and boarding use the affected check, so the full suite runs once, in the train |
| repo | ariadne |
| issue file | workshop/issues/000578-train-repos-pre-close-and-boarding-use-the-affected-check-so-the-full-suite-runs.md |
| boundary | whole-issue close |
| milestone | — |
| window | f6b23cb7a8eb866b080df65c97f92334624b17bd..b8a7d874c2e96c494274c37f5f31c34e08c4b295 |
| command | sdlc close --issue 578 |
| reviewer | claude (anthropic; author anthropic) |
| isolation | fenced |
| timestamp | 2026-10-23T17:50:36-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

Still reviewing: found README.md:1008 still says to run `--name suite` before close. Next I'm checking the remaining guidance and running the tests.

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

The mechanism is correct and well tested. A pure `preCloseCheck` picks the check, `resolvePreCloseCheck` reads its config from main, and `local-check run --pre-close` runs whichever check was picked. The two code-conflict refusals and close's VERIFIED example now name that command through one constant. The train test counts the suite runs that physically happen, not the posted statuses. I ran the targeted `cmd/sdlc` tests (pre-close, refusals, train landing) and the whole `cmd/tl-worker` package. All passed. Two things stop a plain SHIP. First, the sweep of pre-close instructions is incomplete: README.md and the atlas suite paragraph still tell agents to run `--name suite` before close. Second, the new `--pre-close` flag is not mentioned in README at all. Both are cheap documentation fixes.

Done-when: current — both clauses are delivered and tested in both states. The trains-on and trains-off cases are covered by `TestLocalCheckCmdPreClose` and `TestResolvePreCloseCheck`. The single full run per landing is covered by `TestLandingTrainLandsEveryCarOnOneSuite`. The help text and AGENTS §5 are covered by the tl-worker tests. The Revisions entry (the refusals name `--pre-close` instead of resolving a check themselves) narrows the Plan, but does not touch Done-when.

**1. Strengths**
- `cmd/sdlc/precloseck.go:13` — the rule for each repository lives in one pure function with four cases, each returning a reason. The table test needs no IO (ARCH-PURE pass).
- `cmd/sdlc/precloseck.go:7` — `preCloseRunCommand` is the single source for the command, and the refusal tests assert on it (`closefixed_test.go:33`, `publishgate_test.go:425`).
- `cmd/sdlc/localcheck_cmd.go:169` — the check is resolved on the pinned `origin/main` and printed with its reason. Boarding mode is excluded both by cobra and by `affectedModeFor`, so ordinary mode still falls back to recording `suite` when the selection is full.
- `cmd/sdlc/landingtrain_test.go:240` — exactly one saved suite run is counted, which makes "only full run" checkable from what actually ran.
- If a config file cannot be read, the resolver falls back to suite, and the test covers both files.

**2. Critical** — none.

**3. Important**
- `README.md:1008` still says "Run `sdlc local-check run --name suite` on the pushed issue head before close." In the same sweep:
  - `atlas/workflow/ci-merge-check.md:282` says "Agents run it before close with `sdlc local-check run --name suite`."
  - "pre-close suite" wording remains at `README.md:654`, `atlas/workflow/sdlc-binary.md:432-434` and `sdlc-binary.md:2814`.
  
  These are pre-close instructions that do not come from `preCloseRunCommand`. A worker reading the README in pair would still run the full suite twice (ARCH-PURPOSE).
- README has no entry for the new `--pre-close` flag. The local-check section (around `README.md:779-830`) and the affected paragraph (`README.md:887`) should describe it.

**4. Minor**
- The Plan task names atlas `sdlc-binary.md`'s local-check section, but the paragraph went into `ci-merge-check.md`. That is acceptable, but `sdlc-binary.md:101` (the `local-check run` row) doesn't mention `--pre-close`.
- In a train repo, `--pre-close` on a head with no change against main fails with "run --name suite", from the affected path. That is an odd next step for a pre-close command, though it rarely happens.
- `merge_main.go:27` and `helpercapability.go:220` still name `--name suite`. Both are recovery paths, not pre-close guidance, so they are acceptable as they are.

**5. Test coverage notes**
- The table test covers all four combinations. The fixture-repo test covers pair's shape, ariadne's shape, a serial repo, trains without affected config, and both unreadable files. The CLI test covers trains on and off. The flag-exclusivity test covers `--name`, `--all` and `--boarding`.
- Reverting the resolver would turn `TestLocalCheckCmdPreClose` red. I reasoned this from the status assertions; I did not run a revert.

**6. Architectural notes**
- ARCH-DRY: pass. There is one constant and one resolver, and the resolver reuses `readLandingSchedule` and `readAffectedConfig`.
- ARCH-PURE: pass. The decision is pure, and the git reads are injected through `git func`.
- ARCH-PURPOSE: flag. The code and tl-worker/AGENTS consumers were updated, but the README and atlas restatements were not. See Important above.
- ARCH-EVOLVE: pass. No persisted format changed.

**7. Plan revision recommendations**
- Add to `## Revisions`: "atlas: the pre-close paragraph went to `ci-merge-check.md` instead of `sdlc-binary.md`."

```findings
findings:
  - id: new
    severity: Important
    kind: defect
    family: pre-close-guidance-single-source
    title: |
      README and atlas still tell agents to run --name suite before close
    detail: |
      README.md:1008 and atlas/workflow/ci-merge-check.md:282 still name `sdlc local-check run --name suite` as the pre-close run, and the same "pre-close suite" wording remains at README.md:654 and atlas/workflow/sdlc-binary.md:432-434 and 2814. These restate the rule instead of naming `sdlc local-check run --pre-close`, so pair workers following the README still run the full suite twice (ARCH-PURPOSE shadow-sweep). Update all five sites together.
  - id: new
    severity: Important
    kind: defect
    family: readme-new-surface
    title: |
      README update appears missing for the new local-check run --pre-close flag
    detail: |
      The diff adds the --pre-close flag (cmd/sdlc/localcheck_cmd.go:47), which excludes --name, --all and --boarding, but README.md does not mention it in its local-check section (around lines 779-830) or in the affected paragraph (line 887). The sdlc-binary.md local-check run row (line 101) also omits it.
  - id: new
    severity: Minor
    kind: defect
    family: next-action-error-messages
    title: |
      --pre-close on a head with no change against main suggests --name suite
    detail: |
      In a train repo, the affected path refuses an empty change with "run --name suite" (localcheck_affected.go:61), which is an odd next step for a --pre-close caller. It rarely happens.
```

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"d6fd26fd-b9ad-4328-8d36-b6f23418964d","repository":"github.com/xianxu/ariadne","issue_id":"000578","recipe":"small-diff-review","round":1,"attempt":1,"agent":"claude","root":"/Volumes/TestScratch/review-checkouts/d326ff6e41a680dd","reviewed_head":"b8a7d874c2e96c494274c37f5f31c34e08c4b295","slot":"ariadne:2","requested_at":"2026-10-24T00:48:18.968148Z"},"phase":"finished","pid":52229,"started_at":"2026-10-24T00:48:19.085396Z","finished_at":"2026-10-24T00:50:35.988796Z","session_id":"3453467a-8963-403a-a3df-66ee7e7430cf","outcome":"returned"}] -->

---

## Re-review — 2026-10-23T18:01:53-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 578 — Train repos: pre-close and boarding use the affected check, so the full suite runs once, in the train |
| repo | ariadne |
| issue file | workshop/issues/000578-train-repos-pre-close-and-boarding-use-the-affected-check-so-the-full-suite-runs.md |
| boundary | whole-issue close |
| milestone | — |
| window | f6b23cb7a8eb866b080df65c97f92334624b17bd..097344071d5a13e9d7544d1b103b1a55e0e77fd3 |
| command | sdlc close --issue 578 |
| reviewer | claude (anthropic; author anthropic) |
| isolation | fenced |
| timestamp | 2026-10-23T18:01:53-07:00 |
| verdict | SHIP |

## Review

Reading in pieces.

Reviewing the diff; now checking generated AGENTS.md and running the new tests.

```verdict
verdict: SHIP
confidence: high
```

The change does what the issue asks. A single pure function, `preCloseCheck` (`cmd/sdlc/precloseck.go:13`), decides which check a branch runs before close. The new `--pre-close` flag on `local-check run` asks it, and every instruction that used to tell agents to run the full suite before close now names `--pre-close` instead. That covers the tl-worker help, AGENTS.base.md §5, the README, the atlas pages, the embedded local-check help, both code-conflict refusals and close's VERIFIED example. I re-grepped for `--name suite` and "pre-close suite". The hits left in the tree are about other things: the boarding refusal's full-proof alternative, bootstrap, resume/republish, and the legacy-worktree merge path. None of them is a pre-close instruction, so BR-1 and BR-2 are fixed. I ran the targeted tests on the pinned head and they pass: `TestPreCloseCheck`, `TestResolvePreCloseCheck`, `TestPreCloseFlagExcludesAChosenCheck`, `TestLocalCheckCmdPreClose`, `TestFixedThenShipRefusal`, `TestClassifyPublishDelta`, the help tests, `TestLandingTrainLandsEveryCarOnOneSuite`, and the whole `cmd/tl-worker` package. Nothing blocks SHIP.

Done-when: current. The Spec and Revisions changed only how the refusals say the rule: they name the command, which resolves at run time, instead of naming affected or suite. Both Done-when clauses still describe what this diff builds:
- **Train repo:** `TestLocalCheckCmdPreClose` with trains on records only affected. The train test counts the saved suite attempts and requires exactly one. `sdlc close` has no suite-record gate; I grepped `close*.go` to confirm.
- **Repos without trains:** they stay on the suite, shown by the `trains=false` case and the resolver's "ariadne" fixture.

1. **Strengths**
   - `cmd/sdlc/precloseck.go:13-38` keeps the pure decision separate from a thin resolver that reads git through an injected `git` function. Any config file that can't be read falls back to the suite and says why, the same way the car gate does.
   - `cmd/sdlc/localcheck_cmd.go:169-173` maps `--pre-close` onto the existing `Names` path, so nothing about how checks run is duplicated. `--boarding` can't be combined with it, so `affectedModeFor` correctly stays in ordinary mode (a full selection still records `suite`).
   - `cmd/sdlc/landingtrain_test.go:240-255` counts the suite attempts that actually ran, not the posted statuses. That is the right measure for "the only full run".
   - The refusals use the shared `preCloseRunCommand` constant rather than repeating the rule in prose (ARCH-DRY).
   - The resolver test covers both pair's layout and ariadne's (no `landing.json`), plus each kind of unreadable config.

2. **Critical findings:** none.

3. **Important findings:** none.

4. **Minor findings**
   - The `--full-suite` flag on `close --fixed-then-ship` can now carry affected evidence. The flag's help text says so (`close.go:207`), but the name is now slightly misleading. This needs no action this round.

5. **Test coverage notes**
   - "Close passes on an affected record alone" holds because close has no suite gate, not because a new test checks it. Adding a negative test would be cheap insurance if a close suite gate is ever added.
   - `TestLocalCheckCmdPreClose` and the train test need the host preflight, as the Log notes. They passed in this checkout.

6. **Architectural notes**
   - ARCH-DRY: pass. One resolver, and one constant for the command string.
   - ARCH-PURE: pass. `preCloseCheck` is pure and tested from a table. The IO is a small resolver with an injected `git`.
   - ARCH-PURPOSE: pass. I enumerated every place that restated the pre-close run and each now names `--pre-close`. Close's VERIFIED example and both refusals use the constant.
   - ARCH-EVOLVE: pass. No persisted format changes; the code only reads existing configs.
   - Whether trains are on comes from `landing.json` alone. The machine plan only sets the offset (`landingconfig.go:101-124`), so `readLandingSchedule` is the right source.

7. **Plan revision recommendations:** none. The existing Revisions entries already record where the atlas paragraph went and the BR-1/BR-2 sweep.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      All five sites now name --pre-close: README.md:654 and 1013, ci-merge-check.md:281-282, sdlc-binary.md:432-434 and 2814. The remaining --name suite hits are the boarding alternative, bootstrap, resume, republish and the legacy merge path, none of them pre-close.
  - id: BR-2
    disposition: addressed
    note: |
      README.md:900-903 documents --pre-close (resolution, unreadable fallback, exclusivity) and the sdlc-binary.md:101 local-check run row names it; embedded help documents it too.
  - id: BR-3
    disposition: withdrawn
    note: |
      With no change against main there is nothing affected to vouch for, so the suite is a correct next step for a --pre-close caller too (rationale in the issue Log).
```

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"26224410-8a10-415f-86a9-1672785289fa","repository":"github.com/xianxu/ariadne","issue_id":"000578","recipe":"small-diff-review","round":2,"attempt":1,"agent":"claude","root":"/Volumes/TestScratch/review-checkouts/cf269f8cc08d579b","reviewed_head":"097344071d5a13e9d7544d1b103b1a55e0e77fd3","slot":"ariadne:2","requested_at":"2026-10-24T00:59:06.87524Z"},"phase":"finished","pid":1762,"started_at":"2026-10-24T00:59:07.120784Z","finished_at":"2026-10-24T01:01:51.332463Z","session_id":"a9908857-b5ea-421e-a683-918c04994652","outcome":"returned"}] -->
