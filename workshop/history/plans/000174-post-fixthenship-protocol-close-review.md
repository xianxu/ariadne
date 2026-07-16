# Boundary Review — ariadne#174 (whole-issue close)

| field | value |
|-------|-------|
| issue | 174 — close: specify the post-FIX-THEN-SHIP protocol (stop the re-close loop and the bookkeeping publish-gate trip) |
| repo | ariadne |
| issue file | workshop/issues/000174-post-fixthenship-protocol.md |
| boundary | whole-issue close |
| milestone | — |
| window | ecc705b8a9db00883ff90cff158734dbf91ce69f..HEAD |
| command | sdlc close --issue 174 |
| reviewer | claude |
| timestamp | 2026-07-14T20:06:26-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

```verdict
verdict: FIX-THEN-SHIP
confidence: medium
```

The boundary delivers all three legs of #174 faithfully to the plan: a verdict-conditional FIX-THEN-SHIP protocol block wired into `finalizeBoundaryReview` (close.go:1028-1033), a docs-only tolerance in the publish gate reusing the #177 `hasCodePath` classifier with fail-closed git-error posture (publishgate.go:114-132), and an append-only reclose-refusal extension that preserves the gatesig-pinned span verbatim (verified against `gatesig.go:84` and the frozen codex fixture at `codex_test.go:208`). Wiring, formatters, helpers (`gitx.DiffNames`, `judge.VerdictFixThenShip`, all test seams) all exist where claimed; the docs sweep covers every live consumer I found. Nothing blocks the boundary. One caveat on confidence: **the Bash tool is broken in this session** (harness `session-env` mkdir EPERM, not sandbox-overridable), so I could not execute `go test` — my review is by reading; imports, helpers, and constants all check out, but I have no green run of my own. The one Important finding is a coverage seam the new gate policy opens: embedded helptext counts as "doc-only."

## 1. Strengths

- **Pinned-span discipline is exemplary.** Both refusal messages extend append-only; I verified the reclose head span still matches `gatesig.go:84`'s `RefusalPat` and the publish refusal still matches lines 120/126. The new pass line (`publishgate.go:172-175`) deliberately avoids "landed after" and is empirically collision-tested via `assertNoGatesigCollision` — the #172 instrument stays clean.
- **Content-based drift semantics are better than commit-counting.** `DiffNames(anchor, HEAD)` compares *trees*, so a post-close code commit that's later reverted nets to docs-only and passes — correct, since HEAD's code equals the reviewed code. A nice emergent property of the endpoint-diff choice.
- **Multi-issue anchor soundness holds.** `minAhead` tracking picks the newest anchor (publishgate.go:110-111); an inter-anchor code commit sits inside the newest close's branch-point→HEAD review window, so `newestAnchor..HEAD` docs-only is sound — and the new multi-issue subtest (publishgate_test.go) pins exactly this interaction, which a naive test suite would have missed.
- **Verb threading matches the REWORK arm** (ARCH-DRY pass): `formatFixThenShipProtocol(verb)` uses the same `closeVerb` threading as close.go:1041, and the test pins the milestone variant. The Log records this as a plan-quality-judge finding acted on mid-stream — good loop hygiene.
- **Both positive and negative wiring tests** (`TestClose_FixThenShip_EmitsProtocol` / `TestClose_Ship_NoProtocolBlock`) — the SHIP negative is the test most implementations skip, and it's the one that pins verdict-conditionality.

## 2. Critical findings

None.

## 3. Important findings

- **`cmd/sdlc/helptext/*.md` is embedded binary surface but classifies as "doc-only"** — `publishgate.go:124` + `close.go:1313-1324`. `hasCodePath` treats any `.md` as docs, but helptext is compiled into the binary via `helptext.Get` and shapes shipped agent-facing behavior; content-pinning tests (`estimate_helptext_test.go`) can even go red on a helptext edit. So a post-close helptext commit now publishes with **no review coverage and no CI-equivalent check at the gate**, while the pass line claims "reviewed-HEAD-unchanged holds for code." This exceeds what the measured friction (6/6 = lessons/plan-ticks/atlas) required. Fix sketch: at the publish-gate call site only (don't touch the atlas gate's use), tighten the predicate — e.g. `hasCodePath(paths) || anyUnder(paths, "cmd/")` — so embedded docs keep the refusal while workshop/atlas/root-md bookkeeping passes. If instead this is a deliberate accepted residual (the #177-shared-definition ARCH-DRY argument is legitimate), record that decision in the plan's `## Revisions` and the `pre-merge-checks.md` paragraph rather than leaving it implicit.

## 4. Minor findings

- `formatFixThenShipProtocol`'s escape-hatch parenthetical "(doc-only deltas pass on their own)" and "the anchor advances" are issue-close/publish-gate semantics; when rendered with the `sdlc milestone-close` verb they're inapplicable (milestone closes write no codecomplete anchor). Slight imprecision in an agent-facing instruction.
- The fail-closed `derr` branch (publishgate.go:121-123) is untested — plan Task 2's test-surface note promised "git-error fail-closed unchanged" coverage; only `revCount`'s posture is exercised. Hard to isolate (PATH-empty trips earlier calls first); acceptable to leave, worth a note.
- `commitDocs` ignores `os.WriteFile`'s error — but it exactly mirrors the pre-existing `commitCode` idiom (publishgate_test.go:70-75), so style-consistent; not actionable.
- The `## Done when` re-measure bullet (friction-report deltas) is inherently lagging and remains open — the plan already declares tests the leading proof; fine, just don't claim it satisfied at close.

## 5. Test coverage notes

Coverage is strong for the shipped surface: formatter contract tests with gatesig-collision assertions on both new lines; wiring tests for FIX-THEN-SHIP (positive), SHIP (negative), and reclose refusal via `expectDie`; publish-gate subtests for docs-only pass, mixed refuse, and the multi-issue anchor case; the pre-existing code-drift and re-close-after-drift subtests still pin the refusal path. The Log claims a mutation check (inverted `hasCodePath` reddens the docs-only subtests) — plausible but unverifiable by reading. Gap: the new `DiffNames` fail-closed branch (above). **I could not run the suite** — the session's Bash harness is broken (EPERM on its own session-env dir); the main agent should re-run `go test ./cmd/sdlc/... && go test ./cmd/sdlc/internal/processmanual/` before committing this close, per the plan's own verification step.

## 6. Architectural notes

- **ARCH-DRY: pass.** Leg C reuses `hasCodePath` + `gitx.DiffNames` instead of a second docs classifier; formatters are single-source and colocated with the other `format*` helpers. The doc-tolerance prose is restated across 5 helptext/atlas files, but that's the established pattern for human docs (not derivable).
- **ARCH-PURE: pass.** Both formatters are pure and tested without IO; the gate branch stays a thin IO seam; wiring tests use real temp git repos, not mocks.
- **ARCH-PURPOSE: pass, with the Important finding as the one shadow.** The purpose — close the FIX-THEN-SHIP ambiguity and the bookkeeping trip without new bypass pressure — is delivered end-to-end: protocol stated at the moment of ambiguity, gate tolerates the measured friction shape, reclose refusal names the recovery, and every doc consumer I swept (`helptext` close/milestone-close/merge/push, `pre-merge-checks.md`, `issue-lifecycle.md`; `vocabulary.md`/`ci-merge-check.md`/`sdlc-binary.md` verified non-stale) is consistent. The helptext classification seam is the one place the delivered predicate is *broader* than the purpose required.
- Forward note: if verdict-specific protocols grow (the plan's own `formatVerdictProtocol(v)` dispatch idea), the REWORK arm's inline `cwarn` pair at close.go:1040-1041 should fold into the same dispatch.

## 7. Plan revision recommendations

Add one `## Revisions` entry to `workshop/plans/000174-post-fixthenship-protocol-plan.md` covering three deltas the code already shipped past the plan text:
1. `formatFixThenShipProtocol` is verb-parameterized, not "Pure, no args" (Core concepts prose bullet + the Task 1 test sketch are stale; the issue Log records the change but the plan doesn't).
2. Task 4 named `atlas/workflow/sdlc-binary.md` as a modify target; the sweep landed in `pre-merge-checks.md` + `issue-lifecycle.md` instead (verified correct — sdlc-binary.md carries no stale claim).
3. Whatever disposition the Important finding gets (tightened predicate at the publish-gate call site, or an explicitly recorded accepted-residual for embedded helptext).
