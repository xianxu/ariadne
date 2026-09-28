---
id: '000086'
status: done
created: 2026-06-05
updated: 2026-06-05
estimate_hours: 0.5
actual_hours: 0.11
---

# fix stale --actual prose that primes hand-typing instead of measuring (AGENTS.md/issues-SKILL/milestone-close help) — #74 follow-up

## Problem

`#68`/`#74` made `sdlc close`'s `--actual` **computed-and-suggested** (active-time-v3):
omit `--actual` and close runs the engine and prints `→ close with: --actual <h>`; the
authoritative description is "**measured, not hand-typed**" (`atlas/workflow/sdlc-binary.md:139`,
`helptext/actual.md:2`, `helptext/close.md:24,105`).

But `#74` only cleaned up `close.md`/warmup/`explainActual`. The **most agent-facing surfaces
still show `--actual h` as a value you supply, with no pointer to `sdlc actual`** — so an agent
reads "pass a number" and types one from memory. This actually bit us: during nous#42 the agent
hand-passed `--actual 13.5` (a fabricated sum of per-milestone *estimates*) to every close; the
measured value was `0.30h`. A wrong 45× value sailed straight through and was recorded as
`actual_hours`, polluting velocity calibration — precisely the "earned, not guessed" failure the
gate exists to prevent.

Root cause traced to stale prose, not the tool. The tool is correct.

### Stale references (audited 2026-06-05)

🔴 **Primes hand-typing, agent-facing — fix:**
- `AGENTS.md:49` — `sdlc close … --actual h …`; never mentions `sdlc actual`. **Base layer**
  (symlinked, `base.manifest:43`) so it propagates the wrong habit to every downstream repo.
- `construct/local/issues/SKILL.md:17` — `sdlc close --issue N --actual <h> --verified`.
- `cmd/sdlc/helptext/milestone-close.md` (`:12,51,63,66,70`) — examples `--actual 6/0.5/4`;
  unlike `close.md` it never says the value is computed/suggested. (Tool is fine —
  `milestone-close` is a thin wrapper over the `close` path and shares `explainActual`; this is
  **doc-only**.)

🟡 **Minor — point at close syntax with bare `--actual <hours>`:**
- `cmd/sdlc/helptext/close.md:7-8` — top examples show literal `--actual 7 / 2.5` with no
  "(measured)" annotation (lines `:24,105` already explain it).
- `cmd/sdlc/helptext/set-status.md:24` and `cmd/sdlc/setstatus.go:207` — the "next, close with"
  guidance string shows `--actual <hours>`.

🟢 **Correct exemplars to mirror (no change):** `atlas/workflow/sdlc-binary.md:139-140`,
`helptext/actual.md`, `helptext/close.md:24,82,105-107`, `cmd/sdlc/actual.go`/`close.go:680-682`.

⚪ **Out of scope:** `Makefile.workflow:104` (env passthrough mechanism); `workshop/history/*`,
`docs/vision/*` (archived). Separate concern noted in
`workshop/pensive/2026-06-02-…:74`: active-time-v3.py may itself have stale look-back params
(brain dormancy) — i.e. the *engine* may undercount; that's not this issue (this is doc prose).
