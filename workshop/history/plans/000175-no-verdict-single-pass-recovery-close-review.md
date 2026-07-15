# Boundary Review — ariadne#175 (whole-issue close)

| field | value |
|-------|-------|
| issue | 175 — no-verdict gate: accept the issue-close review for trailing unclosed milestones (single-pass Mx recovery) |
| repo | ariadne |
| issue file | workshop/issues/000175-no-verdict-single-pass-recovery.md |
| boundary | whole-issue close |
| milestone | — |
| window | 41e007669669dbf426ec6c75fea675ef6c74ecc9..HEAD |
| command | sdlc close --issue 175 |
| reviewer | claude |
| timestamp | 2026-07-14T19:18:08-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: medium
```

The boundary delivers all three Spec candidates: the verdict gate now partitions missing-verdict milestones into midstream (still refused) vs trailing (accepted, since the imminent issue-close review's branch-point→HEAD window genuinely covers their work), the refusal texts cite §3 with the fold-to-plain-checkboxes recovery, and the plan-quality judge gains the over-split failure-mode bullet as the forward fix. The gate ordering is correct — `computeClose` runs before both the `--no-judge` finalize path and the review dispatch, so the trailing+`--no-judge` refusal fires before anything is written, and acceptance only finalizes after `reviewThenFinalize`. The `--no-judge` void case, the gatesig pins (#172), and the info-line collision guard (#177 precedent) are all handled. Confidence is medium rather than high for one reason only: the Bash tool is broken in this review session (harness-level EPERM creating its session-env dir, even with the sandbox override), so I could not execute `go test ./cmd/sdlc/...` myself — my static trace of all three integration tests and the gate wiring found no failure vector, and the issue Log reports the suite green with a mutation check, but that execution is unverified by me. The main agent should re-run the suite before recording the close.

**1. Strengths**

- The switch in `computeClose` (`cmd/sdlc/close.go:483-497`) is a clean thin dispatch over the pure partition, and the arm ordering (`--no-verdict` ack → midstream refusal → trailing+`--no-judge` refusal → acceptance) preserves every pre-existing behavior while adding the new ones. ARCH-PURE: pass.
- `verdictNextActionLines` + `verdictBypassClosingLine` (`close.go:1515-1535`) extract exactly what the two refusal formatters share, and the const's comment names the gatesig co-maintenance constraint. ARCH-DRY: pass — I verified `gatesig.go:88-90`'s `AckPat`/`RefusalPat` still match the emitted strings verbatim.
- The integration tests pin *where* the refusal fires via judge call-count assertions (`closereview_test.go`), and the midstream test asserts M2 is *not* named in the refusal — that's the kind of negative assertion that keeps the message honest.
- The acceptance coverage claim is real, not asserted: I cross-checked `atlas/workflow/sdlc-binary.md:668-671` — the whole-issue close review windows `merge-base(main, HEAD)→HEAD`, which contains every trailing milestone's commits, including the reopened-issue shape.
- `formatTrailingVerdictAccepted` is guarded by `assertNoGatesigCollision` (ANSI-stripped, same as the live classifier), closing the exact hole a new info line on this path could open.

**2. Critical findings** — none.

**3. Important findings** — none. (ARCH-PURPOSE shadow-sweep: all consumers of the new rule derive or are updated — gate code, both refusal formatters, helptext/close.md, atlas close-gates prose, plan-quality prompt + regenerated golden + prompt-content pin. README has no close-gate surface, so no README gap.)

**4. Minor findings**

- Plan-order-vs-temporal-order corner: `partitionMissingVerdicts` classifies by plan position, so a plan revision that moved a reviewed milestone *below* an unreviewed one (ordered `[M2, M1]`, M1 reviewed, M2 never closed) classifies M2 as midstream and refuses, even though temporally no boundary was crossed unreviewed. Conservative false refusal with an existing recovery path; fine as designed, worth remembering if it ever surfaces.
- `formatTrailingNeedsJudge` deliberately ends with the no-verdict gatesig signature, so #172 friction attribution will count a refusal whose actual remedy is dropping `--no-judge` under the no-verdict gate. The plan chose this ("one no-verdict signature"); noting it so a future friction re-measure doesn't misread the bucket.
- `rewriteIssuePlan` (`closereview_test.go:431`) runs bare `exec.Command("git", ...)` relying on `closeRepo`'s chdir, while `closeRepo` itself sets `cmd.Dir` explicitly — works, but the explicit form is more robust if a fixture ever stops chdir-ing.
- The acceptance line says "window branch-point→HEAD" — accurate for the whole-issue close per the atlas, but a reader of the mixed/reopened shape might expect prev-boundary→HEAD; the over-coverage is in the safe direction, so this is cosmetic.

**5. Test coverage notes**

The pure partition has a 6-case table test including the reopened-issue shape; the three gate integration tests cover accept-trailing / refuse-midstream / refuse-trailing-with-no-judge, each pinning dispatch counts (lessons.md #63 discipline). The Log records a mutation check (everything-midstream mutant reddens both the table test and the acceptance integration test). One untested composite shape: midstream+trailing mixed *at the gate level* — the pure test covers the partition of it and the gate arm is the same code path as pure-midstream, so I don't require it. The acceptance-path-halts-on-REWORK behavior rides on pre-existing `reviewThenFinalize` coverage, unchanged by this diff.

**6. Architectural notes for upcoming work**

- If a future issue widens acceptance to midstream misses covered by a later milestone's window (the plan's named extension point), the temporal-vs-plan-order proxy in the minor finding above becomes load-bearing — the classification would need commit timestamps, not plan position.
- The no-verdict gate now has three distinct outcomes (ack / refuse / info-accept) sharing one gatesig signature family; if a fourth arm ever lands, consider whether gatesig needs a per-arm pattern rather than stretching the shared closing line further.

**7. Plan revision recommendations** — none. The Core concepts table matches the code (all five pure entities and both integration points exist at the stated paths with the stated status), and the two shared helpers not in the table were explicitly sanctioned by the plan's ARCH-DRY implementer note.
