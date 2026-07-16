# Boundary Review — ariadne#157 (whole-issue close)

| field | value |
|-------|-------|
| issue | 157 — process-manual dynamic session reconstruction |
| repo | ariadne |
| issue file | workshop/issues/000157-process-manual-dynamic-session-reconstruction.md |
| boundary | whole-issue close |
| milestone | — |
| window | c8dc3067d24585ef9818acab268109538025e0f1..HEAD |
| command | sdlc close --issue 157 |
| reviewer | claude |
| timestamp | 2026-07-01T17:30:24-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

Confirmed empirically against real transcripts:
- **Trailer-only stdout (#138, a real re-close)** → `ParseVerdict` returns `"unknown"` → the close event renders with **no verdict**, even though `Review-Verdict: FIX-THEN-SHIP` is right there in the stdout.
- **Body+trailer stdout (#133)** → `"SHIP"` → works.

I have everything I need. Here is my review.

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

The `--session` dynamic reconstruction is a clean, well-architected addition: a genuinely pure core (`parseEvents`/`classifyToolUse`/`segmentEvents`/`renderSessionReport`) tested against fixture JSONL with zero IO, a thin IO seam (`locateSessionJSONL` + one `os.ReadFile`), and disciplined DRY reuse of the M1 catalog, `claudeProjectSlug`, `judge.ParseVerdict`, and the introspect gap constant. Tests pass (`processmanual` 0.331s; the `--session` cobra test 0.02s; the full `cmd/sdlc` run timed out at the 600s default under lock contention, not on a #157 assertion), the atlas gate is satisfied, and the precision-first verb matcher (command-boundary anchor + real-verb validation) is a correct root-cause fix backed by regression tests. What keeps this from a clean SHIP is one real, empirically-confirmed under-delivery of the Done-when's "verdicts": verdict recovery silently drops the verdict on **re-closes** whose stdout is trailer-only — a recurring workflow case, and the fix is cheap.

### 1. Strengths
- **Real ARCH-PURE separation** (`session.go:148-428`): the four core functions are pure over bytes/slices and unit-tested with no mocks; only `locateSessionJSONL` + `SessionReport`'s `os.ReadFile` touch IO. Textbook pure-core/thin-shell.
- **Precision-over-recall verb matcher** (`session.go:369-408`): the command-boundary regex + `validVerbs` gate is the correct root-cause fix for the smoke-run false positives (`--include=*.go`, commit-message prose), with the regression cases pinned in `session_test.go:129-131`. This is exactly the "fix it right" the constitution asks for.
- **ARCH-DRY verb source** (`session.go:57-69`): deriving `validVerbs` from the catalog's help-text titles makes "classifies" ⟺ "links" one source of truth — verified both `close.md` and `milestone-close.md` exist, so verdict recovery isn't dead code.
- **Hard limits rendered into the output** (`session.go:279-285`), not silently omitted — the right ARCH-PURPOSE call for the *documented* gaps.
- **Segmentation design** (`session.go:239-277`): gap detection over `allTimes` (all activity, not just fired events) correctly avoids false splits from non-injection work between two fired events.

### 2. Critical findings
None.

### 3. Important findings

**I1 — Verdict recovery drops the verdict on trailer-only (re-close) stdout.** `session.go:219` uses only `judge.ParseVerdict`, which parses the reviewer-*body* `VERDICT:` line and returns `unknown` for a `Review-Verdict:` git-trailer. But real closes are not always body+trailer: a re-close (a previously-blocked FIX-THEN-SHIP/REWORK issue re-closed after fixes, or a verdict-reuse close) streams **trailer-only** stdout. Confirmed against real transcripts: issue #138's close stdout is literally `Review-Verdict: FIX-THEN-SHIP\nReview-Window: …`, and `ParseVerdict` on it returns `"unknown"` (probed directly) — so `--session` renders that close event with no verdict, despite the verdict being structurally present. Across local sessions the split was 4 recoverable / 1 trailer-only among verdict-bearing closes, i.e. ~20% missed. This under-delivers the Done-when's "sdlc verbs + **verdicts**" for a normal recurring case; the plan's assumption ("a real close stdout streams the full reviewer body *then* the trailer, so it resolves correctly", `session.go` header / plan line 47) is empirically incomplete.
  - *Fix sketch (cheap, DRY):* add a `judge.ParseVerdictTrailer(output) Verdict` helper in `classify.go` next to `ParseVerdict` (scan for `^Review-Verdict:\s*(\S+)`, validate via `vocab.Verdict().IsEmitted`), and in `parseEvents` fall back to it when `ParseVerdict` returns `VerdictUnknown`. Add a trailer-only fixture to `TestParseEvents_*` asserting the recovered verdict.

### 4. Minor findings
- **Env-var-prefixed invocation is a false negative:** `VAR=1 sdlc close` isn't matched by `sdlcVerbRE` (the char before `sdlc ` is a space, not a boundary separator). Rare; precision-over-recall makes this an acceptable trade, but worth a one-line comment noting the known miss.
- **Oversized line is fatal, not tolerant:** `session.go:164` caps the scanner at 16 MB; a single JSONL line exceeding it makes `parseEvents` return a hard error (`bufio.Scanner: token too long`) rather than skipping the line — a slight contradiction with the stated "tolerant" goal. 16 MB is generous, so low risk.
- **`FiredEvent.Tool` is semi-dead** (`session.go:122`): set on parse, asserted in tests, but never rendered by `renderSessionReport`. Either surface it in the output or drop it.
- **Timestamp monotonicity assumed** in gap detection (`session.go:245-248`): interleaved sidechain/sub-agent records could perturb `allTimes` ordering and produce a spurious segment split. Acceptable for a human-readable report.

### 5. Test coverage notes
Coverage is strong for the pure core, the classify table (including the three smoke-run regression cases), segmentation, rendering (incl. the unresolved-verb-renders-unlinked case), `locateSessionJSONL` env-vs-mtime resolution, and the cobra wiring. The one gap directly corresponds to the shipped defect: **the verdict-recovery fixture (`session_test.go:214`) only exercises body+trailer stdout; there is no trailer-only case** — which is exactly why I1 shipped. Add a trailer-only fixture when fixing I1.

### 6. Architectural notes for upcoming work
- ARCH-DRY: **pass.** Verb source, `claudeProjectSlug`, `Collect`, `judge.ParseVerdict`, gap constant all reused in-process — the stated reason to stay Go-native holds up.
- ARCH-PURE: **pass.** Clean pure-core / IO-shell split; no mocks needed to test the core.
- ARCH-PURPOSE: **flag (I1).** The ordered fired stream + stated hard limits are delivered, but "verdicts" is part of the committed purpose and is dropped in the trailer-only case despite being recoverable — that's the "easy subset," not a documented limit. Fixing I1 closes the gap.
- Downstream/base-layer: `internal/processmanual/session.go` ships to every downstream ariadne repo. The `--session` surface (path-or-`current`) is stable and additive; the `SessionReport(opts, sessionArg, linkPrefix)` signature mirrors `Manual` — good consumer ergonomics.

### 7. Plan revision recommendations
- **`## Revisions` entry (FiredEvent field drift):** the Core concepts table (plan line 27) still lists `FiredEvent{… Link string (resolved from the M1 catalog)}`, but the code dropped `Link` (links resolve at render time in `renderSessionReport`). The deviation is documented in the issue `## Log` but not in the plan. Add a revision noting `Link` was removed from `FiredEvent` and links resolve at render time.
- **`## Revisions` entry (verdict-recovery assumption):** if I1 is fixed, update the plan's Task-1 note / `parseEvents` description — the claim that close stdout "streams the full reviewer body then the trailer, so it resolves correctly" should be amended to "…except trailer-only re-closes, recovered via the `Review-Verdict:` trailer fallback."
