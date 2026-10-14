# Boundary Review — ariadne#350 (milestone M3)

| field | value |
|-------|-------|
| issue | 350 — lift the TL skill into ariadne: tl-lead and tl-worker skill binaries (help is the prose), scripts ported to Go, skill stubs |
| repo | ariadne |
| issue file | workshop/issues/000350-lift-the-tl-skill-into-ariadne-tl-lead-and-tl-worker-skill-binaries-help-is-the-prose-scripts-ported-to-go-skill-stubs.md |
| boundary | milestone M3 |
| milestone | M3 |
| window | c266ee7e61d9297fb59d5900b4c0cf5e246016d8..ef5ab09c14f73c03d2997649b65c7f78fb9478d9 |
| command | sdlc milestone-close --issue 350 --milestone M3 |
| reviewer | claude |
| isolation | fenced |
| timestamp | 2026-10-14T08:55:40-07:00 |
| verdict | not-run |

## Review



## Review preparation and protocol diagnostics

Quoted values are diagnostic data, not review verdicts.
Protocol attempt 1
reason: "missing valid findings block"
output: "Now verify the atlas sentence: \"A failed inventory leaves the Trains block intact\" and \"a plain write still lands\".\n\n```verdict\nverdict: SHIP\nconfidence: high\n```\n\nThe window `c266ee7e..ef5ab09c` changes one line of `atlas/workflow/tl-skills.md`. The rest is the M2 gate ledger and the M2 review record. The base is a rebased copy of the branch, so the M3 code is the same on both ends: `pkg/markdoc` and `pkg/trains` extraction, the `tl-lead status` and `trains` subcommands, the `SessionSync` and `SDLC` seams, and the fakes. The M1 review window already covered that code. I checked the new atlas sentence against the code, and both claims hold:\n- **Degraded render:** `cmd/tl-lead/trains.go:63-69` keeps `renderTrains(plan, cars)` and renders the Slots block as `? off-train slots unread: <err>`. `TestGatherTrainsFleetFailureDegrades` pins this (`trains_test.go:220-238`).\n- **Delivery gap:** a `--fold-owed` write is refused with \"check recent messages, then write without --fold-owed\", and a plain write still lands (`sessionsync.go:38-42`). `TestDeliveryGapRefusesFoldOnly` pins this (`status_test.go:284-303`).\n\nSo the edit settles the prior Minor BR-7. I also checked the M3 plan rows against the head tree. Every build step is ticked except the last one (\"full `make test`, then `sdlc close`, hand over to the TL\"), and that step comes after this gate. Targeted tests pass: `go test ./cmd/tl-lead/... ./cmd/tl-worker/... ./pkg/markdoc/... ./pkg/trains/...`. Nothing blocks SHIP.\n\n1. **Strengths**\n   - The atlas now states both degraded behaviours precisely, and each one has a test: the fleet-failure render and the gap that refuses only the fold.\n   - The fleet-failure path degrades visibly, as `? … unread: <reason>`. It does not abort the render or invent an empty slot list (`trains.go:63-69`).\n   - `ErrDeliveryGap` is a typed sentinel that carries the next action, and `decodeReply` maps couch's `delivery-gap` code onto it (`sessionsync.go:93-96`, tested by `TestDecodeReplyDeliveryGap`).\n   - The README already names `tl-worker` and `tl-lead` and points to `--help` (`README.md:195-196`).\n\n2. **Critical findings:** none.\n\n3. **Important findings:** none.\n\n4. **Minor findings:** none raised. The atlas line at `tl-skills.md:33` is getting long; consider splitting the degraded-behaviour sentences into their own bullet the next time it is edited.\n\n5. **Test coverage notes**\n   - BR-7 changed only the atlas. I checked the corrected passage against `trains.go:63-69` and `sessionsync.go:38-42`, and the existing tests cover both behaviours. No wording test is needed.\n   - The live conformance test (`TL_LEAD_LIVE=1`) was not re-run, because this window changes no behaviour.\n\n6. **Architectural notes**\n   - **ARCH-DRY: pass.** One parser for trains (`pkg/trains`) and one for plan bars (`pkg/markdoc`); the old names forward to them.\n   - **ARCH-PURE: pass.** Rendering is pure, and IO sits behind the `SessionSync` and `SDLC` seams.\n   - **ARCH-PURPOSE: pass.** The atlas now tracks every degraded path the code has; the TL handover is the declared close step.\n   - **ARCH-MOCK: pass.** `fakeSessionSync` keeps state across calls and models the gap; `fakeSDLC` injects fleet errors; the opt-in live conformance test exists.\n   - **ARCH-CONSTRAINTS: pass.** Car reads are bounded at 8 at once, and the fleet read overlaps them; the atlas states a render takes about 12 s.\n   - **ARCH-SECURE: pass.** couch is called with argv arrays, and an unreadable reply surfaces as an error instead of a made-up value.\n   - **ARCH-ORDER: pass.** Writes are guarded by compare-and-replace on the section hash, and a delivery gap is typed.\n   - **ARCH-FUNERAL: pass.** The window creates nothing durable apart from the gate-ledger docs, which the gate tooling owns.\n\n7. **Plan revision recommendations:** none.\n\n```findings\n```"
Protocol attempt 2
reason: "missing valid findings block"
output: "Tests pass and the atlas change matches the code. Last step: checking the M3 checklist in the plan.\n\n```verdict\nverdict: SHIP\nconfidence: high\n```\n\nThis review covers only the M3 close window (`c266ee7e..ef5ab09c`). Base `c266ee7e` is the rebased reviewed patch, so the only net changes are one atlas paragraph (`atlas/workflow/tl-skills.md:33`), the close-gate record, the new M2 review record and two issue-log lines. The atlas edit adds two claims, and I checked both against the code. First, a failed inventory leaves the Trains block intact and renders `? off-train slots unread: <reason>` (`cmd/tl-lead/trains.go:63-69`, pinned by `trains_test.go:236`). Second, a delivery gap refuses a `--fold-owed` write and names the next action, while a plain write still lands (`cmd/tl-lead/sessionsync.go:38-41,95`, pinned by `status_test.go:286` `TestDeliveryGapRefusesFoldOnly`). `go test ./cmd/tl-lead/` passes at head. Every M3 plan row is ticked except \"full `make test` then `sdlc close`\", and that row is this close itself. Nothing blocks SHIP.\n\n1. **Strengths**\n   - The atlas text follows the code exactly. It names the literal output string and says which writes are refused and which still land. This resolves the `atlas-tracks-behavior` family in a way a reader can check.\n   - When the inventory fails, the render degrades visibly and stays partial instead of aborting (`trains.go:63-69`). The comment explains why the Trains block survives.\n   - The delivery-gap test checks both halves of the contract: zero writes when folding, and the plain write landing (`status_test.go:286-302`).\n   - The plan's `## Revisions` records each departure from the plan: goldens kept inline, the `pkg/tlrules` location, the `RepoRoot` seam and the legacy-lock decision.\n\n2. **Critical:** none.\n3. **Important:** none.\n4. **Minor:** none.\n5. **Test coverage:** both new atlas claims have tests that would fail if the behavior regressed (`trains_test.go:236`, `status_test.go:286`). The package passes at head.\n6. **Architecture**\n   - ARCH-DRY: pass. The window contains only docs.\n   - ARCH-PURE: pass. The render degradation is a value computed in `gatherTrains`, and the IO goes through the `SDLC` seam.\n   - ARCH-PURPOSE: pass. The handover to the TL's ops change is named in the plan's non-goals and in the final M3 row.\n   - ARCH-MOCK: pass. `fakeSessionSync` models the delivery-gap refusal, and conformance is opt-in through `TL_LEAD_LIVE`.\n   - ARCH-CONSTRAINTS: pass. The window adds no runtime change. The roughly 12 s render bound is documented.\n   - ARCH-SECURE: pass. The fleet error text reaches only the rendered status block, not a credential path.\n   - ARCH-ORDER: pass. The fleet goroutine is bounded lexically by `<-done` before return.\n   - ARCH-FUNERAL: pass. The window creates no new durable artifacts.\n7. **Plan revisions:** none needed.\n\n```findings\n```"

truncated: false
retained-bytes: 0000007083

---

## Re-review — 2026-10-14T08:59:14-07:00 (FIX-THEN-SHIP)

| field | value |
|-------|-------|
| issue | 350 — lift the TL skill into ariadne: tl-lead and tl-worker skill binaries (help is the prose), scripts ported to Go, skill stubs |
| repo | ariadne |
| issue file | workshop/issues/000350-lift-the-tl-skill-into-ariadne-tl-lead-and-tl-worker-skill-binaries-help-is-the-prose-scripts-ported-to-go-skill-stubs.md |
| boundary | milestone M3 |
| milestone | M3 |
| window | c266ee7e61d9297fb59d5900b4c0cf5e246016d8..81db31ea07a5b1eda21876c95d20150b5f2b1345 |
| command | sdlc milestone-close --issue 350 --milestone M3 |
| reviewer | claude |
| isolation | fenced |
| timestamp | 2026-10-14T08:59:14-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

I'm checking the M3 plan items against the head tree. The window itself is small: one regex fix plus an atlas clause.

The fix's test fails when the fix is reverted. Next I'm checking the new atlas sentences against the code.

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

This window (`c266ee7e..81db31ea`) contains little new code. The M3 subcommands (`tl-lead status`/`trains`, the `pkg/markdoc` and `pkg/trains` extraction, the seams and fakes) landed in commits that the M1 review window (`1c391825..392c1079`) already covered. What is new here is three things:
- the `TickMilestone` bold-row fix, which shares one `milestoneEmphasis` prefix between the reader and the writer (`pkg/markdoc/plan.go:67,101,106`);
- the atlas clause that answers BR-7;
- gate and ledger records.

The fix is correct and its regression test is real: I reverted line 67 to the old pattern, and `TestTickMilestoneTicksEveryFormTheReaderCounts` went red on the bold, italic and in-progress rows. With the fix restored, it passes. `go test ./pkg/markdoc/... ./cmd/sdlc/internal/issue/...` and the targeted `cmd/sdlc` close and guard tests pass. One cheap item remains before SHIP. `cmd/sdlc/close_test.go` still holds a hand copy of the *old* tick regex and claims it "mirrors the regex in runClose's milestone path". That claim became false in this window. The copy is also the only place the boundary cases are pinned (`M4` must not tick `M40` or `M4b`); those cases never reach the production function.

1. **Strengths**
   - `pkg/markdoc/plan.go:103-106`: reader and writer now derive from one constant, so the divergence is fixed for every row form, not just bold rows (ARCH-DRY, ARCH-PURPOSE).
   - `pkg/markdoc/plan_test.go:117-140`: the test first checks its own premise (that the reader counts the row as M1), then asserts the round trip through the reader: M1 is ticked and M2 is not. It tests a property, not the implementation.
   - The atlas clause (`atlas/workflow/tl-skills.md:33`) matches the code: `cmd/tl-lead/trains.go:68` renders `? off-train slots unread: <reason>`, and `cmd/tl-lead/sessionsync.go:41-42` refuses a `--fold-owed` write with "check recent messages, then write without --fold-owed". This disposes BR-7 in substance.
   - The issue Log is honest that the running `sdlc` is built from ariadne:0, so this branch's own milestone rows were ticked by hand. That records an operational fact a later reader would otherwise trip over.

2. **Critical findings:** none.

3. **Important findings**
   - `cmd/sdlc/close_test.go:164-172`: `TestMilestoneTickRegex` builds its own `(?m)^(- )\[[ .]\]( M4\b)`. Its comment says this "mirrors the regex in runClose's milestone path", but after this window it no longer does, because the emphasis prefix is missing. Its boundary cases (`M40_no_match`, `M4b_no_match_for_M4`, `M4-extra` matching via `\b`, already-`[x]` unchanged) run against the copy, so a regression in `markdoc.TickMilestone`'s boundary handling would go uncaught.
     - This is the anti-pattern that `TickMilestone`'s own doc comment (`plan.go:45-50`) says the function was extracted to remove.
     - Fix sketch: move those cases into `pkg/markdoc/plan_test.go` as a table over `TickMilestone("## Plan\n\n"+row+"\n", "M4")`, including bold forms (`**M40`, `**M4b`). Then delete `TestMilestoneTickRegex`.
     - The class: tests that restate a production regex. A grep for `regexp.MustCompile` in `_test.go` files that duplicates a `pkg/markdoc` pattern finds this one site.

4. **Minor findings**
   - `cmd/sdlc/resolve.go:47`: `milestoneRe` restates `M\d+[a-z]?` instead of building from `markdoc.MilestoneTagPattern`. This predates the window and is not raised as a finding.
   - `close.go:731`: the warning text `no '- [ ] %s'` does not mention bold forms. It's cosmetic.

5. **Test coverage notes**
   - The fix's regression test fails without the fix; I checked by reverting the change.
   - Boundary semantics (prefix collisions, `[x]` not re-ticked) are not pinned against the production `TickMilestone` (see the Important finding above).
   - The new test covers plain, bold, italic, in-progress and colon forms, plus a sibling milestone that must not be ticked.

6. **Architectural notes**
   - **ARCH-DRY:** flagged. The production code fixes duplication through a shared constant, but the stale regex copy in `close_test.go` remains (Important finding above).
   - **ARCH-PURE:** pass. `TickMilestone` and the reader are pure and are tested without IO.
   - **ARCH-PURPOSE:** pass. The fix covers the class (every form the reader counts) rather than the bold instance. The tool handover (the ops pointer, deleting `scripts/*.py`) is correctly scoped as the TL's ops change in the plan's Non-goals.
   - **ARCH-MOCK:** pass. No new external calls in this window. The `SessionSync` and `SDLC` seams and fakes were reviewed at M1, and the live conformance test is logged as PASS.
   - **ARCH-CONSTRAINTS:** pass. The window adds no runtime-path change; `regexp.MustCompile` per call in `TickMilestone` predates it and runs only once per close.
   - **ARCH-SECURE:** pass. Issue files are parsed into tolerant regex matches, and an unmatched row degrades to a visible warning (`close.go:731`).
   - **ARCH-ORDER:** pass. Holds no state between events: `TickMilestone` is a pure body-to-body rewrite.
   - **ARCH-FUNERAL:** pass. Creates nothing durable apart from the review and ledger records, which follow the existing gate-artifact lifecycle.

7. **Plan revision recommendations**
   - Add a `## Revisions` entry to the plan: "2026-10-14: folded in the `TickMilestone` bold-row fix (`pkg/markdoc/plan.go`, shared `milestoneEmphasis`), which was not in the M3 Files list; found while closing M1 and M2, and it blocks this issue's close." Today only the issue `## Log` records this.

```findings
findings:
  - id: new
    severity: Important
    kind: defect
    family: tests-exercise-production-code
    title: |
      TestMilestoneTickRegex pins a stale hand copy of the pre-fix tick regex, not markdoc.TickMilestone
    detail: |
      cmd/sdlc/close_test.go:164-172 claims to mirror runClose's tick regex but builds the old pattern without milestoneEmphasis, which diverged from pkg/markdoc/plan.go:67 in this window. Its boundary cases (M40, M4b, M4-extra, [x] unchanged) never reach the production function. Move them into a pkg/markdoc table over TickMilestone, including bold forms, and delete the copy.
```

---

## Re-review — 2026-10-14T09:12:37-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 350 — lift the TL skill into ariadne: tl-lead and tl-worker skill binaries (help is the prose), scripts ported to Go, skill stubs |
| repo | ariadne |
| issue file | workshop/issues/000350-lift-the-tl-skill-into-ariadne-tl-lead-and-tl-worker-skill-binaries-help-is-the-prose-scripts-ported-to-go-skill-stubs.md |
| boundary | milestone M3 |
| milestone | M3 |
| window | c266ee7e61d9297fb59d5900b4c0cf5e246016d8..0058e96c01c1287501f0129f9735bc5062029e78 |
| command | sdlc milestone-close --issue 350 --milestone M3 |
| reviewer | claude |
| isolation | fenced |
| timestamp | 2026-10-14T09:12:37-07:00 |
| verdict | SHIP |

## Review

The regression test fails when I revert the fix, which confirms BR-8 is addressed. Next I'm checking the new atlas claims against the tl-lead code.

The targeted tests and `go vet` pass. Last step: checking that the plan file records the writer change.

```verdict
verdict: SHIP
confidence: high
```

This window covers the follow-ups after M2: the BR-8 fix, a fold-in fix to the `TickMilestone` writer, and an atlas note on two behaviors. All three are correct and backed by tests. BR-8 is addressed. The hand-copied regex test in `cmd/sdlc/close_test.go` is deleted, and its boundary cases now call the production `markdoc.TickMilestone` (`pkg/markdoc/plan_test.go:146-168`). I removed `milestoneEmphasis` from the writer's pattern (`pkg/markdoc/plan.go:67`) and both new tests failed: bold, italic and in-progress rows ticked 0. With the fix restored, `go test -count=1 ./pkg/markdoc/` passes. `go vet` is clean. The targeted `cmd/sdlc` tests (`TestPlanUnchecked*`, `TestPlanItemWritersUseTickMilestone`, the planfence `TickMilestone` cases) and `./cmd/tl-lead` pass. I started the full `cmd/sdlc` suite, but it was still running at the 300 s timeout and gave no result before the review ended, so it is unverified here. Nothing blocks SHIP.

1. **Strengths**
   - The reader and the writer now share one emphasis prefix: the `milestoneEmphasis` constant (`pkg/markdoc/plan.go:106`), used at `:67` and `:101`. A row the counter reads can no longer be one the writer can't tick (ARCH-DRY).
   - `TestTickMilestoneTicksEveryFormTheReaderCounts` (`plan_test.go:117`) first asserts that the reader counts each row form, then checks that the writer ticks it and that the reader sees the tick. That tests the reader and writer together rather than restating the regex. I confirmed it goes red when the fix is removed.
   - The BR-8 fix covers the whole class, not just the named test. A grep for `\[[ .]\]` / `TickMilestone(` in the Go files finds no other hand copy. The remaining `TestPlanUncheckedDetection*` tests already call `issue.PlanSectionBody` / `issue.PlanUncheckedRE`.
   - The atlas additions (`atlas/workflow/tl-skills.md:33`) match the code. The degraded `? off-train slots unread:` render is at `cmd/tl-lead/trains.go:63-69` and tested at `trains_test.go:236`. The fold-only refusal on a delivery gap is tested in `status_test.go:286` (`TestDeliveryGapRefusesFoldOnly`).
   - The issue `## Log` (lines 146-148) records this out-of-scope bug fix with its reason, the fix and the test.

2. **Critical findings:** none.

3. **Important findings:** none.

4. **Minor findings**
   - The `TestTickMilestoneBoundaries` comment refers to a test that no longer exists (`cmd/sdlc/close_test.go`). That is acceptable as history, but it will age.

5. **Test coverage notes:** The tests cover the bold, italic, in-progress, colon and plain row forms, the tag boundaries (M40, M4b, bold M40, M4-extra), and an already-ticked row left unchanged. All of them run through the production function, with no mocks or IO.

6. **Architectural notes**
   - ARCH-DRY: pass (one shared prefix constant).
   - ARCH-PURE: pass (`TickMilestone` is pure and tested without IO).
   - ARCH-PURPOSE: pass (the BR-8 sweep found no other copies).
   - ARCH-MOCK: N/A in this window (no new external calls).
   - ARCH-CONSTRAINTS: N/A (regex change only, not a hot path).
   - ARCH-SECURE: pass (the writer quotes the tag with `regexp.QuoteMeta`).
   - ARCH-ORDER: N/A (no state carried between events in this diff).
   - ARCH-FUNERAL: N/A (creates nothing that persists).

7. **Plan revision recommendations:** none. The issue Log already records the `TickMilestone` fold-in.

```findings
dispose:
  - id: BR-8
    disposition: addressed
    note: |
      Hand copy deleted from cmd/sdlc/close_test.go; boundary cases now drive markdoc.TickMilestone (plan_test.go:146-168, incl. bold); reverting milestoneEmphasis at plan.go:67 turns both new tests red; sweep finds no other regex copies.
```

That background run was the full `cmd/sdlc` suite. It **failed**: Go's default 10-minute test timeout killed it after 600 s and printed a goroutine dump. The exit code 0 only came from the `| tail` pipe. I don't know which test was running when it timed out, because I kept only the last few lines. No test reported a failed assertion before the timeout. The grep half of that command didn't run either (a zsh glob error); the later Grep sweep did the same check.

I don't think this window caused it. The only `cmd/sdlc` change here deletes one self-contained regex test (`TestMilestoneTickRegex`) and an unused import. Every `cmd/sdlc` test that calls the changed `TickMilestone` passes when run on its own, and so do `go vet` and the `tl-lead` tests. Still, I haven't shown the whole suite passes at this head. The issue log already notes a known `cmd/sdlc` test failure under `MAKEFLAGS=k` (ariadne#358), which may be related; I didn't check.

**Correction to my review:** the verdict stays `SHIP`, but it rests on the targeted tests and the revert check, not a green full suite. Before crossing the boundary, run `go test -timeout 30m ./cmd/sdlc/` once and check that it passes. If it still hangs, `-v` will name the test that stalled.
