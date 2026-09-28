---
id: '000074'
status: done
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
