---
id: '000067'
status: done
created: 2026-06-02
updated: 2026-06-02
estimate_hours: 1.5
actual_hours: 2
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
