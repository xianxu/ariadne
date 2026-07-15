# Boundary Review — ariadne#178 (whole-issue close)

| field | value |
|-------|-------|
| issue | 178 — close: adopt the measured actual when --actual is omitted (kill the compute-then-ask loop) |
| repo | ariadne |
| issue file | workshop/issues/000178-close-adopts-measured-actual.md |
| boundary | whole-issue close |
| milestone | — |
| window | e4e06cbe00c83de01f87c727252b55c4d51d32af..HEAD |
| command | sdlc close --issue 178 |
| reviewer | claude |
| timestamp | 2026-07-14T17:02:02-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

Review complete — all checks done (code, tests, gatesig instrument, milestone routing, docs sweep, test suite run). Final report:

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

The change delivers its purpose cleanly: the omit-path now measures once through a stubbed seam and adopts on `actualMeasured`, the refusal survives byte-identical for unmeasurable statuses (the gatesig instrument's refusal anchor at `close.go:1165` is unchanged, and a new test proves the adopt line classifies to neither ACK nor refusal), and the deviation check correctly skips a value the engine itself produced. `go test ./cmd/sdlc/` passes (26.9s). What keeps this from SHIP: the milestone-mode auto-adopt records a **cumulative** number into **per-milestone** project detail blocks — contradicting the repo's own documented rule (`workshop/lessons.md:513`: "per-milestone actuals are INCREMENTS") — and the wording sweep missed one shadow doc that still teaches the old two-step. Both are cheap; neither blocks the gate.

**1. Strengths**

- The refusal-arm reuse is the right call: `adoptOmittedActual` returns the measurement so `explainActual` renders the diagnosis without a second engine run (`close.go:1131-1143`) — the naive port would have measured twice on every refusal.
- `resolveOmittedActual` pins the ledger value to `%.2f` (`close.go:1102-1107`) so frontmatter, ledger, and info line can never disagree — and the test pins both `1.234→"1.23"` and `7.0→"7.00"`.
- `TestAdoptLineNoGatesigCollision` (`close_adopt_test.go:100-117`) tests against the *rendered* line including `cinfo`'s exact ANSI prefix (verified: matches `term.go:29,34` byte-for-byte) across both modes and every `GateCatalog` pattern — this is exactly the #172-instrument regression the diff could have shipped.
- The fall-through comment at `close.go:379-381` is accurate: I traced the control flow and the `--no-actual` warn is genuinely the only way to reach that block now.
- Plan `## Revisions` honestly records the deltas (signature change, the `milestone-close.md` straggler, the omit+no-verified live-verify trick) instead of pretending the plan was executed verbatim.

**2. Critical findings**

None.

**3. Important findings**

- **Milestone-mode adoption writes cumulative hours into per-milestone records, contradicting the documented increments rule** — `close.go:1120-1122` + `close.go:567-568`. Each milestone gets its own project detail block (`AnchorFor` includes the milestone, `close.go:563`), so at M2+ the auto-adopted cumulative value makes successive blocks double-count when summed (M1: 2h, M2: 5h → reads as 7h of a 5h issue). `workshop/lessons.md:513-517` explicitly instructs increments (`cumulative − Σ(prior milestones)`) and even proposes the fix (window prev-boundary..HEAD). The old flow suggested the same wrong number but left the agent a decision point to apply the lesson; the new flow records it before the agent can. The calibration ledger is safe (`shouldLogCalibration` excludes milestones, `close.go:696`), which is why this isn't Critical. Fix sketch — pick one: (a) keep the refusal in milestone mode only (`mode == "milestone"` → don't adopt); (b) adopt the windowed increment for milestones; or (c) explicitly supersede the lesson (update `lessons.md:513` + the project-file convention to "cumulative snapshot per block") in this same range. Option (a) is the smallest honest change. (ARCH-PURPOSE: disclosed-in-a-log-line is not the same as resolved.)
- **Wording sweep missed a shadow doc**: `construct/local/issues/SKILL.md:17` (the source of the `xx-issues` skill agents load) still says "omit it (close computes + suggests it)" — it teaches the deleted two-step. This is precisely the shadow-sweep class `lessons.md` #122 warns about (a hand-maintained restatement of the contract left describing the old behavior). One-line fix: "omit it (close measures + adopts it)". (ARCH-PURPOSE shadow-sweep.)

**4. Minor findings**

- `close.go:134` — the `--actual` flag help string still reads "(sdlc computes it; see `sdlc actual`)"; could say "measured and adopted when omitted".
- `close.go:1117` — `"#" + strings.Join(res.Peers, ", #")` hand-rolls what `prefixHash` (`actual.go:234`) already does; reuse it (ARCH-DRY, trivial).
- `actual.go:214` — `sdlc actual`'s "→ close with: --actual %.2f" is now mildly stale advice: copying it back triggers a needless #87 deviation round-trip vs simply omitting the flag at close.
- `close_adopt_test.go:107` hardcodes `cinfo`'s ANSI prefix; rendering through `cinfo` into a buffer would keep the test honest if the prefix ever changes.

**5. Test coverage notes**

The new tests pin the decision table (all five statuses), the format (both modes, cumulative note), the wiring (single engine call, `f.Actual` adoption, no side effects on refusal), and the gatesig non-collision — the right pure-function coverage per ARCH-PURE, and stubbing via the seam means no real transcripts needed. Two gaps, both explainable: (1) no in-process test drives `computeClose` omit-path through to `actual_hours:` in frontmatter — `exitWithCode`/`die` call `os.Exit` (the `lessons.md:519` debug aside documents this constraint), and the adopted value enters the identical pre-existing pass-path at `close.go:508-509`, which existing tests cover; (2) the refusal *exit path* (b) and N/A sentinel (c) from the plan's test list lean on pre-existing tests (`close_test.go:398`) rather than new ones. Acceptable, but the plan should say so (see §7).

**6. Architectural notes**

- **ARCH-DRY: pass** (one `prefixHash` nit). One measurement feeds adopt, warnings, and the refusal explainer; `resolveActualRoots` reused; no new verb.
- **ARCH-PURE: pass.** Decision (`resolveOmittedActual`) and rendering (`formatAdoptLine`) are pure and tested without IO; the engine sits behind one package-var seam matching the file's existing pattern; `adoptOmittedActual` is thin wiring.
- **ARCH-PURPOSE: pass with one flag.** The diff deletes the loop at its root rather than teaching agents around it, and the doc sweep covered warmup, both helptexts, AGENTS.base.md, and atlas — but the shadow-sweep (per this repo's own #122 lesson) turns up `construct/local/issues/SKILL.md:17` still restating the old contract, and the milestone-mode finding above is the one place the diff states a problem instead of resolving it.
- For upcoming work: `checkActualDeviation` still re-runs the full engine on the explicit pass-path while the omit-path now has a measurement in hand — if explicit `--actual` after a preview stays common, threading one measurement through both would drop a second multi-second engine run.

**7. Plan revision recommendations**

Add one `## Revisions` entry to `workshop/plans/000178-close-adopts-measured-actual-plan.md`: Task 1 tests (a)–(c) as-written claim frontmatter-landing, refusal-exit, and N/A-sentinel pins; what shipped pins those at the helper layer (`resolveOmittedActual` / `adoptOmittedActual`) and leans on pre-existing `computeClose` tests for frontmatter and sentinel, because the exit paths call `os.Exit` in-process. The substance is covered; the plan's test list shouldn't overclaim what `close_adopt_test.go` itself does. If the milestone finding is resolved via option (a) or (b), that's a second Revisions entry (task (f)'s "pre-existing semantics, now stated" would no longer describe the code).
