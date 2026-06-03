---
id: 000067
status: working
deps: []
github_issue:
created: 2026-06-02
updated: 2026-06-02
estimate_hours: 1.5
---

# sdlc close: per-gate --no-<gate> bypass flags (narrower than --force) + fix stale FORCE=1 messages

## Problem

Closing #66 (a pure `insertLogLine` bugfix, no new architectural surface)
tripped `close`'s atlas-change gate. The only escape was `--force`, which
bypasses **all** of close's 7 guards at once (actual, verified, atlas, verdict,
plan-unchecked, project, re-close). Too blunt — waiving atlas shouldn't also
silently waive the evidence requirements. Want a narrower, *acknowledged*
bypass: `--no-atlas` = "I've consciously determined no atlas change is needed,"
not "skip everything." Operator's call: make it systematic — **every gate gets
its own `--no-<gate>` flag**; `--force` stays as the umbrella (≡ all of them).

Bonus: the refusal messages say "set `FORCE=1`" — a stale Makefile/Python-era
env var that **nothing reads** (no `os.Getenv("FORCE")` anywhere). They should
name the precise `--no-<gate>` flag + `--force`.

## Spec

7 gates in `runClose`, each guarded by `!f.Force`. Per-gate flags:

| gate | flag |
|------|------|
| actual-hours required | `--no-actual` |
| verified-evidence required | `--no-verified` |
| re-close guard (already `done`) | `--no-reclose-guard` |
| atlas/ changed in window | `--no-atlas` |
| every milestone has `Review-Verdict:` | `--no-verdict` |
| `## Plan` has no unchecked items | `--no-plan-check` |
| project detail-block updated | `--no-project` |

- DRY: 7 bool fields on `closeFlags` + `func (f *closeFlags) skip(gate string)
  bool` → `f.Force || <that field>`. Each gate site: `!f.Force` → `!f.skip("…")`.
- Audit: bypass fires + logs a `cwarn` only when the gate would actually have
  refused. `--no-actual`/`--no-verified` get a stronger caution (they weaken
  velocity calibration). Rationale stays in `--verified` (no separate reason flag).
- Rewrite the stale `FORCE=1` messages to name the specific flag + `--force`.
- `milestone-close` pass-through: same 7 flags on `milestoneCloseFlags`, threaded
  into the `closeFlags` it builds, so `sdlc milestone-close --no-atlas` works.
- Docs: `close.md` help (flag family + table + `--force` ≡ all); AGENTS.md §5
  (the convention — flag = explicit acknowledgment, not forgetting; `--force` =
  all; why in `--verified`).

`merge`/`push` gate adoption of the same pattern is a noted follow-up (merge
already has `--no-judge`, the precedent).

## Done when

- Each of close's 7 gates has a `--no-<gate>` flag that bypasses ONLY that gate;
  `--force` still bypasses all.
- A bypass emits an audit `[!]` line (acknowledged, not silent) and only when the
  gate would have fired.
- `sdlc close --no-atlas` closes a no-atlas-change issue while still enforcing
  actual/verified; `--no-actual` alone still requires verified (per-gate isolation).
- Refusal messages name the precise `--no-<gate>` flag (no stale `FORCE=1`).
- `sdlc milestone-close --no-atlas` works (pass-through).
- AGENTS.md §5 + close.md document the convention; tests cover `skip()` + flag
  registration.

## Plan

- [ ] Implement the `--no-<gate>` family on `close` (+ `milestone-close`
  pass-through), audit warns, message rewrites, docs (close.md + AGENTS.md §5),
  and tests (`skip()` logic + flag registration). Verify per the plan's manual
  matrix; dogfood by closing this issue (real atlas change → no `--no-atlas`
  needed for its own close).

## Log

### 2026-06-02
