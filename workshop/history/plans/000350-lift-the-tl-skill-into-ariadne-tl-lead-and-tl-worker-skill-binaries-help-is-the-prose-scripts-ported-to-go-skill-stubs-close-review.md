# Boundary Review — ariadne#350 (whole-issue close)

| field | value |
|-------|-------|
| issue | 350 — lift the TL skill into ariadne: tl-lead and tl-worker skill binaries (help is the prose), scripts ported to Go, skill stubs |
| repo | ariadne |
| issue file | workshop/issues/000350-lift-the-tl-skill-into-ariadne-tl-lead-and-tl-worker-skill-binaries-help-is-the-prose-scripts-ported-to-go-skill-stubs.md |
| boundary | whole-issue close |
| milestone | — |
| window | 1c39182532fa3cc78ea4935abff4ab1b87a6f01d..60e462468f2815550b4026cda2fe6e2b5c10db3b |
| command | sdlc close --issue 350 |
| reviewer | claude |
| isolation | fenced |
| timestamp | 2026-10-14T10:30:47-07:00 |
| verdict | REWORK |

## Review

Disposing of the two open findings is done (BR-6 is fixed; BR-7 no longer applies since the off-train render was dropped). Now running the tests and checking the new slots/threads code.

```verdict
verdict: REWORK
confidence: medium
```

The code holds up. All tests pass for `cmd/tl-lead`, `cmd/tl-worker`, `pkg/markdoc`, `pkg/trains`, `pkg/tlrules` and the sdlc `issue`/`project` packages. The new `status slots` and `status thread` work (db844120) is pure where it should be, and it is covered by `TestRenderSlots`, `TestGatherSlots`, `TestStatusSlots`, `TestUpsertThread` and `TestThreadDropAndPlace`. Both open findings are resolved:

- **BR-6:** the comment now matches what the test does.
- **BR-7:** no longer applies, because the scope revision removed the off-train render.

That same scope revision is the problem. The TL decision (cc315116) was recorded only in the issue's Log. The plan's Core concepts table and Revisions still describe `RenderOffTrain`, an off-train block, and `TestGatherTrainsFleetFailureDegrades`, and none of them exist in the code. The plan also never names the slots and threads entities that replaced them. The review contract treats a Core concepts table that contradicts the code as Critical, so this blocks. It is cheap to fix.

One test failure is outside this work: `pkg/workspace` `TestEnvironmentLegacyAndArchive` fails ("nil Git reader"), but this range doesn't touch `pkg/workspace`, so it isn't attributed here.

1. **Strengths**
   - `renderedSections` (`cmd/tl-lead/status.go:284`) gives a later `status quota` (ariadne#365) a one-row place to register, and every rendered section writes through the guarded `writeSection` path.
   - `TestStatusSlots` (`cmd/tl-lead/trains_test.go:236`) checks that a failed fleet read produces no write and that the error says why.
   - `boundedEach` (`cmd/tl-lead/trains.go:20`) is one shared bounded fan-out (at most 8 at once) for both car reads and slot reads.
   - `issueBar` lists the directory and matches by name instead of globbing, so glob characters in a checkout path can't cause trouble (ARCH-SECURE).
   - 60e46246 makes the two shell tests derive the tool list from `cmd/`, which fixes the whole class of stale tool lists rather than just the one in the earlier finding.

2. **Critical**
   - **The plan's Core concepts table and Revisions contradict the code after the scope revision.** Plan line 28 lists `RenderOffTrain` in `cmd/tl-lead/render.go`; it doesn't exist. Lines 7, 38, 138, 153 and 164 describe an off-train block, `TestRenderOffTrain` and `TestGatherTrainsFleetFailureDegrades`; none of these exist. There are no rows for `slotLines`/`renderSlots`/`gatherSlots` (`cmd/tl-lead/slots.go`) or `upsertThread` (`cmd/tl-lead/threads.go`).
     - **This is the 2nd finding in family `plan-core-concepts-drift`.** Rule: a commit that adds, removes or renames an entity, or that records a scope decision, updates the plan's Core concepts table and adds a `## Revisions` entry in the same commit, not only the issue's Log. Fix: sweep every Core concepts row and every Revisions claim against the tree.

3. **Important:** none.

4. **Minor**
   - **Two comments describe code that isn't there** (family `comment-tracks-code`, 2nd finding):
     - `cmd/tl-lead/trains.go:66` names `carReadConcurrency`; the constant is `readConcurrency`.
     - `cmd/tl-lead/slots.go:18` says Phase is `"?" with the read error`, but `gatherSlots` (`slots.go:86`) throws the error away.
     - Rule: a comment that names an identifier, or data a value carries, must match the code. Sweep with `grep` for identifiers that don't resolve.
   - `threadBlocks` (`cmd/tl-lead/threads.go:15`) splits on any line starting with `### `, without checking for code fences. A fenced `### ` line inside a thread would split it in two. `pkg/markdoc` already has a fence scanner it could reuse (ARCH-DRY).

5. **Test coverage**
   - The slots and threads paths are tested through the stateful fakes.
   - Slot cases covered: free, detached, other branch, slot :0 excluded, failed read rendering `?`.
   - Missing: an upsert where a subsection's body contains a fenced `### ` line.

6. **Architecture, marker by marker**
   - **ARCH-DRY:** pass, apart from the Minor note above.
   - **ARCH-PURE:** pass. `slotLines`, `renderSlots` and `upsertThread` are pure; the `gather*` functions are thin.
   - **ARCH-PURPOSE:** pass. The status file follows the new layout.
   - **ARCH-MOCK:** pass. Production and tests share the SDLC and SessionSync seams.
   - **ARCH-CONSTRAINTS:** pass. Concurrency is bounded at 8.
   - **ARCH-SECURE:** pass.
   - **ARCH-ORDER:** pass. Compare-and-replace is couch's job, and retries are bounded.
   - **ARCH-FUNERAL:** pass. The seen-set file is bounded by the owed set.

7. **Plan revision recommendations**
   - Add a `## Revisions` entry for the 2026-10-14 scope revision: `status trains` and the off-train render removed; `status slots`, `status thread` and the `renderedSections` registration point added.
   - Replace the `RenderOffTrain` Core concepts row with rows for `slots.go` (`slotLines`/`renderSlots`/`gatherSlots`) and `threads.go` (`upsertThread`).
   - Mark the off-train text at lines 7, 38, 138, 153 and 164 as superseded.

```findings
dispose:
  - id: BR-6
    disposition: addressed
    note: |
      cmd/tl-lead/conformance_test.go:12-17 now says the only fold names an ID that isn't owed, matching lines 53-68.
  - id: BR-7
    disposition: withdrawn
    note: |
      Overtaken by the scope revision (cc315116, db844120). The off-train render was removed; no "off-train" text remains in cmd/tl-lead or atlas.
findings:
  - id: new
    severity: Critical
    kind: defect
    family: plan-core-concepts-drift
    title: |
      Plan Core concepts and Revisions still claim RenderOffTrain and the off-train degraded render; slots and threads entities are missing
    detail: |
      2nd in family plan-core-concepts-drift. Plan line 28 lists RenderOffTrain in cmd/tl-lead/render.go, and lines 7, 38, 138, 153 and 164 describe the off-train block, TestRenderOffTrain and TestGatherTrainsFleetFailureDegrades. None of these exist after db844120. There are no rows for slotLines/renderSlots/gatherSlots (slots.go) or upsertThread (threads.go), and the scope decision was logged only in the issue. Rule: any commit that adds or removes an entity, or records a scope decision, updates Core concepts and Revisions in the same commit. Sweep every row and revision claim against the tree.
  - id: new
    severity: Minor
    kind: defect
    family: comment-tracks-code
    title: |
      Comments name a nonexistent carReadConcurrency and claim the slot Phase carries the read error
    detail: |
      2nd in family comment-tracks-code. trains.go:66 names carReadConcurrency (the constant is readConcurrency). slots.go:18 says Phase is "?" with the read error, but gatherSlots (slots.go:86) discards the error. Rule: a comment naming an identifier or carried data must match the code. Sweep with grep for identifiers that don't resolve.
  - id: new
    severity: Minor
    kind: defect
    family: reuse-fence-aware-scanner
    title: |
      threadBlocks splits on any line starting with "### ", including inside a code fence
    detail: |
      cmd/tl-lead/threads.go:15 isn't fence-aware. A pasted thread body with a fenced line starting "### " would be split into a spurious subsection. pkg/markdoc already has a fence scanner it could reuse (ARCH-DRY).
```

---

## Re-review — 2026-10-14T10:38:14-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 350 — lift the TL skill into ariadne: tl-lead and tl-worker skill binaries (help is the prose), scripts ported to Go, skill stubs |
| repo | ariadne |
| issue file | workshop/issues/000350-lift-the-tl-skill-into-ariadne-tl-lead-and-tl-worker-skill-binaries-help-is-the-prose-scripts-ported-to-go-skill-stubs.md |
| boundary | whole-issue close |
| milestone | — |
| window | 1c39182532fa3cc78ea4935abff4ab1b87a6f01d..f1fde220fd4ea232791be65302e340a1728e4f59 |
| command | sdlc close --issue 350 |
| reviewer | claude |
| isolation | fenced |
| timestamp | 2026-10-14T10:38:14-07:00 |
| verdict | SHIP |

## Review

Focus on the latest commit (fixes for BR-9..11).

Prior findings look addressed; now checking that the fence-aware fix has a test that fails without it.

```verdict
verdict: SHIP
confidence: high
```

This round fixes all three open findings, and I checked each against the tree and tests. The plan's Core concepts table and entity notes now describe the code that exists. The off-train and bullet-thread entities are marked `deleted`, and a 2026-10-14 Revisions entry supersedes the old claims by name. Both comments that pointed at the wrong code are fixed. Thread splitting now skips fenced lines, and its regression test fails when the fix is removed. I grepped every backticked identifier in the plan against the Go tree. All of them resolve except entities marked deleted or claims explicitly superseded, plus two cosmetic labels covered in the one Minor finding below. Nothing blocks the close.

1. **Strengths**
   - `cmd/tl-lead/threads.go:23` reuses `markdoc.FenceSpans` instead of adding a second fence scanner (ARCH-DRY). The new test row in `threads_test.go:18` pins the behavior.
   - The plan's 2026-10-14 Revisions entry names each claim it supersedes (`TestRenderOffTrain`, `TestGatherTrainsFleetFailureDegrades`, the fleet overlap) instead of quietly editing history.
   - `workshop/lessons.md` now states the general rule for the `plan-core-concepts-drift` family rather than only fixing this instance, which is how ARCH-PURPOSE says a finding should be answered.
   - `atlas/workflow/tl-skills.md:32-33` matches the restructured surface (`status slots`, `renderedSections`, `###` threads).

2. **Critical findings:** none.

3. **Important findings:** none.

4. **Minor findings**
   - **3rd finding in family `plan-core-concepts-drift`.** Earlier rounds fixed instances, and this round wrote the rule (lessons.md: grep every cited identifier against the tree before a close), but the sweep was not fully applied:
     - Core concepts rows 23–24 name `workerHelp` and `leadHelp`, but both embed variables are `rootHelp` (`cmd/tl-lead/main.go:19`, `cmd/tl-worker/main.go:18`).
     - The entity note at plan line 41 still says `PlanBar(planBody)`, while the code and the table say `planBar`.
     - Recommendation: don't just patch these two. Turn the lesson's grep into the actual close step, so the check runs instead of relying on memory. These are labels next to correct paths, so this does not block.

5. **Test coverage notes**
   - These pass: `go test ./cmd/tl-lead/... ./cmd/tl-worker/... ./pkg/markdoc/... ./pkg/trains/... ./pkg/tlrules/...`.
   - To check BR-11, I replaced the `!fenced[i]` guard with a no-op. `TestUpsertThread` then failed on the new "a fenced ### line stays inside its subsection" row, so the fix is what makes it pass.
   - `pkg/workspace` `TestEnvironmentLegacyAndArchive` fails in this checkout ("nil Git reader"). This range does not touch `pkg/workspace`, so it is outside this review, but it should be looked at before the full `make test` the plan's last M3 row calls for.
   - For BR-10's class, I swept comments in `cmd/tl-lead`, `cmd/tl-worker`, `pkg/trains` and `pkg/tlrules` for camelCase identifiers that don't resolve, and found none.

6. **Architectural notes**
   - ARCH-DRY: pass. The fence scanner is shared, and `pkg/markdoc` and `pkg/trains` are each the single source.
   - ARCH-PURE: pass. `threadBlocks`, `upsertThread`, `slotLines` and `renderSlots` are pure. IO sits behind the `SDLC` and `SessionSync` seams.
   - ARCH-PURPOSE: pass. Each prior finding was answered as a class, with a lesson and a sweep, not just at the site it named.
   - ARCH-MOCK: pass. `fakeSessionSync` and `fakeSDLC` share the production seam, and the live conformance test is opt-in.
   - ARCH-CONSTRAINTS: pass. The plan records the bounded concurrency (8) and the measured timings.
   - ARCH-SECURE: pass. A failed observation renders `?`, never a made-up status, and arguments go to subprocesses as argv arrays.
   - ARCH-ORDER: pass. The plan lists the session-sync event cases, and they are driven through the fake.
   - ARCH-FUNERAL: pass. The seen-set is pruned to the owed set, and `.clobbered` has explicit removal paths.

7. **Plan revision recommendations**
   - Correct Core concepts rows 23–24 to `rootHelp` (embedded `help.md`), or relabel them as descriptions rather than identifiers. Also change entity note line 41 from `PlanBar` to `planBar`.

```findings
dispose:
  - id: BR-9
    disposition: addressed
    note: |
      Core concepts lists slotLines/renderSlots/SlotLine, threadBlocks/upsertThread, lineDiff, renderedSections; RenderOffTrain/ThreadEntries marked deleted; Revisions 2026-10-14 supersedes TestRenderOffTrain and TestGatherTrainsFleetFailureDegrades by name; grep of plan identifiers resolves except deleted/superseded ones.
  - id: BR-10
    disposition: addressed
    note: |
      trains.go:66 now names readConcurrency; slots.go:18 says "?" when the observation failed, matching gatherSlots; comment-identifier sweep over tl-lead/tl-worker/pkg found no unresolved names.
  - id: BR-11
    disposition: addressed
    note: |
      threads.go:23 uses markdoc.FenceSpans; neutralizing the guard turns TestUpsertThread's fenced-### row red.
findings:
  - id: new
    severity: Minor
    kind: defect
    family: plan-core-concepts-drift
    title: |
      Core concepts rows name workerHelp/leadHelp and the note names PlanBar; the code has rootHelp and planBar
    detail: |
      3rd in family. Plan rows 23-24 cite workerHelp/leadHelp but both embeds are rootHelp (cmd/tl-lead/main.go:19, cmd/tl-worker/main.go:18); entity note line 41 says PlanBar while code and table say planBar. The lessons.md rule (grep each cited identifier before close) covers this; mechanize that grep as a close step rather than patching instances.
```
