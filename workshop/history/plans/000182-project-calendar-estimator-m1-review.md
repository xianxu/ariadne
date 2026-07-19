# Boundary Review — ariadne#182 (milestone M1)

| field | value |
|-------|-------|
| issue | 182 — project calendar estimator: mechanize the commit reality-check (effort→calendar bridge) |
| repo | ariadne |
| issue file | workshop/issues/000182-project-calendar-estimator.md |
| boundary | milestone M1 |
| milestone | M1 |
| window | 0657d8dfaee80f5d93b703d41856718c7f821c93^..HEAD |
| command | sdlc milestone-close --issue 182 --milestone M1 |
| reviewer | claude |
| timestamp | 2026-07-19T01:06:51-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

```verdict
verdict: FIX-THEN-SHIP
confidence: medium
```

**Summary.** #182 M1 (throughput measurement + blessed baseline) is a clean, well-scoped milestone: a pure `SpanThroughput` + baseline TSV codec in `internal/estimate/throughput.go`, and a thin `sdlc project throughput` shell in `projectthroughput.go` (`--bless`/`--ceiling`/`--brain-dir` + a read-only bare form). The pure/IO separation is exemplary (ARCH-PURE), it reuses the existing ledger parser instead of adding a second one (ARCH-DRY), and edge cases (unparsable dates, untrusted rows, empty span, corrupt baseline) are handled and tested. I found **no correctness bugs** in the M1 surface. What holds this back from a clean SHIP is non-blocking: the README project quick-start doesn't list the new `throughput` subcommand (Docs gate), and one plan step contradicts the shipped code. Note: I could not run the suite — my Bash is broken (EPERM on `~/.claude/session-env`), the known boundary-review-subagent limitation — so this is static analysis at **medium** confidence; **the main agent must run `go build ./... && go test ./cmd/sdlc/... ./pkg/vocab/` and confirm green before crossing.**

**Scope note.** The review window (`0657d8d^..HEAD`) predates the #180/#171 merges, so the diff carries all of #180 and #171. Those milestones already had their own boundary reviews (per the continuations + lessons.md). I reviewed **only the #182 M1 deliverables** (Chunk 1) and treated the rest as prior, separately-reviewed context — not re-litigated here.

---

### 1. Strengths (confirmed-good ground)

- **Pure core / thin shell is textbook (ARCH-PURE).** `SpanThroughput`, `ParseBaselineTSV`, `RenderBaselineRow`, `BaselineHeader` are IO-free and unit-tested directly (`throughput_test.go`); all `os.*` calls live in `blessThroughput`/`appendBaselineRow`/`showThroughput` and are exercised against `t.TempDir()` fixtures (`projectthroughput_test.go`).
- **Genuine reuse, no parallel parser (ARCH-DRY).** Consumes `estimate.LedgerRow`/`ParseRows`/`VelocityPath`/`Header` as-is (`projectthroughput.go:85`, `throughput.go:36`); `calibrationLedgerPath` mirrors `appendCalibrationRow`'s `WF_CALIB_LEDGER`-first resolution (`projectthroughput.go:52-56`). No second ledger row type.
- **Inclusive-day math is correct and DST-safe.** `days := int(toT.Sub(fromT).Hours()/24) + 1` (`throughput.go:48`) is exact because `time.Parse("2006-01-02", …)` yields UTC (no DST hour-drift); `d.Before/After` gives closed-interval membership, matching the tests (a#3 on the 28th is in-span).
- **Right fail-loud/fail-quiet asymmetry.** A *corrupt* baseline surfaces as an error (`ParseBaselineTSV` → `throughput.go:113-125`, wired at `projectthroughput.go:143-145`); an *absent* baseline is a quiet hint (`showThroughput:137-142`). A malformed numeric column errors rather than silently reading 0.
- **Fixtures use the real 10-column ledger shape** (`throughput_test.go:11-17`), and untrusted rows are counted-but-included with the count surfaced as a warning (`throughput.go:63-65`, `projectthroughput.go:102-104`) — the honest choice for the rate.

### 2. Critical findings

None.

### 3. Important findings

- **README project quick-start omits `throughput` (Docs gate).** `README.md:15-37` frames itself as "the model-derived CLI surface" and enumerates new/list/show/validate/set-status/status/retro/close/find, but not `throughput` — a new subcommand a reader types. The canonical in-CLI doc (`helptext/project.md:16,39-53`) *is* thorough, so this is a one-line add, not a real gap in coverage. (Reasonable counter-argument the operator may weigh: `throughput` is a measurement utility, arguably peripheral to that section's "create and inspect" narrative — but the section lists nearly every other verb.) Fix: add `sdlc project throughput --bless <FROM>..<TO>` (and the bare form) to the README block.

### 4. Minor findings

- `isoLayout` (`projectthroughput.go:169`, package main) and `isoDate` (`throughput.go:27`, package estimate) are two consts for the same `"2006-01-02"` layout; the literal is already repeated ~15× across package main. Pre-existing pattern, cross-package — not a #182 regression; noting only.
- `showThroughput:137-142` swallows *all* baseline read errors (including permission-denied) as "no blessed baseline yet." Fine for an informational read; a stat-based existence check would report a real IO error more honestly.
- `--ceiling` is silently a no-op in bare (show) mode — only recorded on `--bless`. Harmless; the help text already scopes it to bless.

### 5. Test coverage notes

Coverage is strong for the delivered surface. Small gaps, all low-value:
- `appendBaselineRow`'s "existing file without trailing newline" branch (`projectthroughput.go:128-130`) isn't exercised — both bless paths in `TestProjectThroughput_BlessAppends` write a trailing `\n`.
- `ParseBaselineTSV`'s `len(c) < 6` (too-few-columns) error branch isn't directly tested (the bad-float branch is).
- `TestSpanThroughput_BadSpanBounds` covers a bad *from*-date and `to<from`, but not a bad *to*-date specifically (symmetric code path, so low risk).

### 6. Architectural notes for upcoming work

- **ARCH-DRY / ARCH-PURE / ARCH-PURPOSE: all PASS for M1.** M1 correctly ships only the measured-baseline substrate; the forecast core (M2) and consumers (M3) are genuine separate boundaries, not a deferral of "the point." The baseline file *is* the interface M2/M3 read.
- The un-tagged (lock-free) throughput command is the right seam for M2/M3: it writes only the brain baseline (outside the repo transaction) — so future consumers reading the baseline stay cheap and lock-free like `resolve`. Keep the brain-file writes out of the repo-lock path when M3 wires `set-status →committed`.
- When M2 adds `ComputeForecast`, the throughput baseline's `Ceiling` field is already carried per-row — the forecast's ceiling *warning* should read `cur.Ceiling` from the current baseline row rather than re-deriving a constant.

### 7. Plan revision recommendations

- **`workshop/plans/000182-project-calendar-estimator-plan.md`, Task 1.2 Step 3** says *"Read-only bare form; bless is `markMutatingCommand`."* The shipped code deliberately does **not** tag the command `markMutatingCommand` (`projectthroughput.go:35-42`, with a rationale comment: cobra can't vary the lock per-invocation, the bare `show` must stay lock-free, and the write target is a brain file outside the repo transaction). `repolock_test.go:19-32` confirms the intended behavior — it lists `{"project","close"}` as lock-taking but **not** `{"project","throughput"}`. The code + test agree; the plan text is stale. Add a `## Revisions` entry recording that `throughput` is intentionally lock-free (bless/show duality on one command + brain-resident write), superseding the `markMutatingCommand` instruction — so the plan stops claiming what the code correctly doesn't do. (This is a plan↔code contradiction to reconcile, not a correctness defect — the deviation is the better choice.)
