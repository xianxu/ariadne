# Boundary Review — ariadne#160 (milestone M2)

| field | value |
|-------|-------|
| issue | 160 — add an issue status: codecomplete |
| repo | ariadne |
| issue file | workshop/issues/000160-add-an-issue-status-codecomplete.md |
| boundary | milestone M2 |
| milestone | M2 |
| window | 21e90db580c3b69225be5e6e2d2736ebdc565a8f..HEAD |
| command | sdlc milestone-close --issue 160 --milestone M2 |
| reviewer | claude |
| timestamp | 2026-07-02T17:16:11-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

M2 delivers its stated scope cleanly: `sdlc close` now flips a whole-issue close `working → codecomplete` (not `done`), the `lessons` ping relocates to close gated to the whole-issue boundary, and the boundary-review prompt's docs gate is broadened to own README sync. I verified the flip is structurally scoped to whole-issue close (never milestone-close), the lessons emission is correctly gated on both emit paths, the "Atlas → Docs update gate" rename moved *all three* coupled sites (section header, cross-ref, test assertion) with the golden regenerated in sync, and the M2-relevant tests pass in isolation (the 600s `TestSetStatusAlias_BothPathsMutate` timeout was repo-lock contention from my own parallel runs, not this diff). One Important gap blocks a clean SHIP only in the "cheap to fix before the gate" sense: the load-bearing `f.Milestone == ""` lessons guard has no test proving a *milestone*-close stays silent.

**1. Strengths**
- **Correct scoping of the flip** — `close.go:490` sits inside the `else { // issue close }` arm, so milestone-close never touches `status`. Verified, not assumed.
- **Lessons gating is sound on both paths.** `reviewThenFinalize` (shared by milestone + whole-issue) guards with `if f.Milestone == ""` (`close.go:880`); `finishBoundaryReview` (`close.go:907`) emits unconditionally but is genuinely milestone-less by construction — it's only reachable past the `if f.Milestone != "" { return runClose(...) }` early-return at `close.go:785`. The "milestone-less by construction" comment checks out.
- **The docs-gate rename avoided the #142 review-issue-#2 trap** — both the `:73` section header *and* the `:55` cross-ref moved together, `judge_test.go` assertion updated, golden regenerated. Grep confirms zero stale "Atlas update gate" refs remain in `cmd/`, `atlas/`, `construct/` (only the reference plan docs mention it, correctly).
- **Negative paths updated correctly** — REWORK / dispatch-error / unknown verdict tests now assert *no* flip to `codecomplete`, and the lessons test covers SHIP-emits vs REWORK-omits. `LessonsReminder` is reused from the shared `judge` constant (ARCH-DRY holds).
- **Atlas already covers the surface** (from M1: `issue-lifecycle.md` two-gate model + `codecomplete` row, `vocabulary.md` split), so the Docs gate is satisfied for M2 without a new atlas edit; README has no status-flip surface to update (grep confirms).

**2. Critical findings** — none.

**3. Important findings**
- **Missing test: milestone-close must NOT emit the lessons reminder** (`close.go:880`). The `if f.Milestone == ""` guard is the *only* thing keeping the Q4 ping off the milestone boundary, and `reviewThenFinalize` is shared with milestone-close (`milestoneclose.go:168`). But `TestRunMilestoneClose_SHIP_Finalizes` (`close_finalize_test.go:181`) passes `io.Discard` for stdout — delete the guard and every test still stays green while every `sdlc milestone-close` starts spuriously pinging. *Failure scenario:* a refactor drops the `f.Milestone == ""` condition → milestone-close emits the reminder at each boundary, undetected. *Fix (cheap):* capture stdout in `TestRunMilestoneClose_SHIP_Finalizes` into a `strings.Builder` and assert `!strings.Contains(stdout.String(), judge.LessonsReminder)`. (ARCH-PURE lens: the emission decision *is* the logic; it's untested on the milestone branch.)

**4. Minor findings**
- Stale "flipped → done" wording in comments now that the message reads "flipped … → status: codecomplete": `close.go:323`, `close.go:627`, and the test comment/error-string at `close_finalize_test.go:62,83`. The assertions themselves are sound (they match on the `"flipped"` substring, not `"→ done"`), so this is prose-only drift.
- `emitLessonsReminder` (`close.go:915`) prints a leading blank line + reminder; `preflight.go:102` prints the reminder with no leading blank line — the same ping renders slightly differently across gates. Cosmetic; self-resolves when M3 removes the merge/push emission.
- Intermediate double-ping: in the M2 state the lessons reminder fires at **both** close and merge/push (removal from merge/push is M3, per plan review #1). Intentional and documented, but a reader running `sdlc merge` on this branch before M3 lands would see it twice. Noting for awareness, not a defect.

**5. Test coverage notes**
- Finalize paths are well covered: SHIP→`codecomplete`, and REWORK/dispatch-error/unknown→no-flip, plus lessons emit-on-SHIP / omit-on-REWORK for the whole-issue path. The gap is the milestone branch (Important #1).
- Plan M2 Step 4 called for a "re-close an already-`codecomplete` issue succeeds" case; it wasn't added. Low risk — the re-close guard (`close.go:401`) keys on the literal `"done"`, so `codecomplete` passes through trivially, and `TestRunCloseWithReview_RerunAfterREWORK` exercises a nearby rerun. Worth adding when M3 wires the drift/re-close flow (M3 Step 8), not a boundary blocker.
- `TestFrontmatterChainForIssueClose` re-runs the `SetField` sequence rather than driving `computeClose`, so it pins field *ordering* (a real concern) but not close's logic; the end-to-end `codecomplete` write is covered by the `runCloseWithReview` integration tests. Acceptable.

**6. Architectural notes for upcoming work (M3)**
- **ARCH-DRY**: pass. Status value is a literal `"codecomplete"` mirroring the `"done"` #122 carve-out; `LessonsReminder` is a single shared constant; the golden/​code-review.md duplication is the intended golden-snapshot pattern (test-enforced to stay in sync).
- **ARCH-PURE**: pass. The flip is a pure `issue.SetField` frontmatter mutation, unit-tested without IO; the LLM dispatch is a process-level fake (`stubJudge`) injected via `judge.Run`, per the constitution's external-service-fake rule. Only the milestone-branch emission decision lacks a test (Important #1).
- **ARCH-PURPOSE**: pass for M2's slice, with a caveat to carry into M3. The purpose — `codecomplete ⟹ close reviewed HEAD` — is only *enforced* once M3's `runPublishGate` lands; today `codecomplete` is written but the invariant isn't yet checked. That's fine (all milestones ship atomically on this unreleased branch — no downstream repo gets M2-without-M3), but M3 is where the purpose is actually fulfilled: it must (a) add the reviewed-HEAD-unchanged refusal, (b) remove the `plan`/`specs` LLM judges *and* the lessons ping from merge/push (completes Q4, kills the double-ping), and (c) confirm the `"done"`-keyed re-close guard still allows re-closing a `codecomplete` issue.

**7. Plan revision recommendations** — none. The plan still matches the code; the set-status refusal was legitimately pulled into M1 and the plan already records that (M2 Step 3 marked `[x] → DONE IN M1`). No "## Revisions" entry needed.
