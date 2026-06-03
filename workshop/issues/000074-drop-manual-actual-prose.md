---
id: 000074
status: working
deps: []
github_issue:
created: 2026-06-03
updated: 2026-06-03
estimate_hours: 0.25
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

- [ ] Rewrite the ACTUAL block in `printSemanticWarmup` + drop the "Method:"/
  manual lines in `explainActual`; point both at `sdlc actual`. Verify the
  warmup + missing-`--actual` paths render cleanly.

## Log

### 2026-06-03
