---
id: 000074
status: done
deps: []
github_issue:
created: 2026-06-03
updated: 2026-06-03
estimate_hours: 0.25
actual_hours: 0.25
---

# remove manual actual-computation prose from close — sdlc computes it now

## Problem

#68 M2 lifted actual-computation into the binary (`sdlc actual` / close runs v3
and prints the measured suggestion). But two spots in `close.go` still teach the
operator to compute it **by hand** — stale now, and the wrong default:

- `printSemanticWarmup` (the close-issue contract): "ACTUAL = … derived via the
  v3 procedure. Run `active-time-v3.py` over the issue's commit window with
  `--commit-weight 1.0`; read the per-issue total. See baseline-v3.md. Pass
  --no-actual only if you genuinely cannot run the script."
- `explainActual`: "Method: v3 commit-anchored segment-local attribution. See
  baseline-v3.md."

`sdlc close` is the path forward — it computes the number. The prose should point
at that, not at a python command nobody should be running by hand.

## Spec

Remove the manual-invocation how-to from both spots; replace with "sdlc computes
it (`sdlc actual --issue N`, or close suggests it inline)". Keep the *definition*
of ACTUAL (focused dev-hours, not wall-clock) — that's still load-bearing for the
judgment-fallback case — and a light pointer to the methodology doc as a
reference, not as instructions. No behavior change; prose only.

## Done when

- Neither `printSemanticWarmup` nor `explainActual` instructs running
  `active-time-v3.py` / `--commit-weight` / "read the per-issue total" by hand.
- Both point at `sdlc actual` / the inline suggestion as the way to get the number.
- The definition of ACTUAL (focused dev-hours) is retained; build + suite green;
  `sdlc close --help`-adjacent warmup renders the new prose.

## Plan

- [x] Rewrite the ACTUAL block in `printSemanticWarmup` + drop the "Method:"/
  manual lines in `explainActual`; point both at `sdlc actual`. Verify the
  warmup + missing-`--actual` paths render cleanly.

## Log

### 2026-06-03
- 2026-06-03: closed — manual active-time-v3 how-to removed from warmup, explainActual, --actual flag help, and close.md; all point at sdlc actual / inline suggestion; build+test+helptext green; grep sweep confirms no manual-compute prose remains. --no-atlas: pure prose/help-string, no new surface. actual=judgment (sdlc actual measures 0.05h, undercounts pre-#74:-commit scoping)

- Removed manual-compute prose from **four** spots (a sweep caught two beyond the
  original two): `printSemanticWarmup` ACTUAL block + `explainActual` "Method:"
  lines (both now "sdlc computes it — `sdlc actual --issue N`"; kept the ACTUAL
  *definition* + a light baseline-v3.md method pointer); the `--actual` flag help
  ("focused dev-hours (sdlc computes it; see `sdlc actual`)"); and `close.md`
  helptext (the "derived from active-time-v3" line + the "explainer prints a
  tailored active-time-v3 command line" paragraph → "close runs active-time-v3
  itself … prints the measured suggestion inline").
- Verified: `go build` + `go test ./cmd/sdlc/...` + helptext test green; the
  warmup + missing-`--actual` paths render the new prose; grep sweep confirms no
  `active-time-v3 command` / `--commit-weight` / `per-issue total` / `v3
  procedure` manual instructions remain in close.go or close.md.
- No separate fresh-eyes subagent — pure prose/help-string change, verified by
  render + exhaustive grep (plan-quality judge already vetted the plan).
