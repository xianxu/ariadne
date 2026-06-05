---
id: 000086
status: working
deps: []
github_issue:
created: 2026-06-05
updated: 2026-06-05
estimate_hours: 0.5
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

## Spec

Align every 🔴 and 🟡 surface to the established 🟢 wording: **`--actual` is measured, not
hand-typed — omit it to let close compute+suggest it, or run `sdlc actual --issue N` first.**
Don't restate the full method (that's `helptext/actual.md`'s job, per #74); just stop priming a
typed number and point at the command.

- `AGENTS.md:49`: drop the bare `--actual h` framing; show the close without `--actual` and add
  a clause that the hours are computed (`sdlc actual` / omit to auto-compute), never typed from
  memory. Keep the "refuses without verification + actuals + atlas" contract.
- `construct/local/issues/SKILL.md:17`: same one-line correction.
- `helptext/milestone-close.md`: add the "computed/suggested, see `sdlc actual` / omit to
  compute" note that `close.md` already has; leave examples but annotate one as measured.
- `helptext/close.md:7-8`, `helptext/set-status.md:24`, `setstatus.go:207`: annotate the example
  hours as measured / cross-reference `sdlc actual`.

No behavior change to the binary's logic — prose + help strings only. (The code backstop that
would catch a *fabricated* passed value is the sibling issue **#87**, cross-referenced.)

## Done when

- All 🔴 + 🟡 surfaces above no longer present `--actual` as a hand-supplied number; each points
  at measurement (`sdlc actual` / omit-to-compute), mirroring the 🟢 exemplars.
- `go build ./cmd/sdlc` clean; `sdlc close --help` and `sdlc milestone-close --help` render the
  corrected prose; a grep sweep shows no agent-facing `--actual <h|hours>` example lacking a
  measurement pointer.
- AGENTS.md change is weighed for downstream propagation (base layer) and noted in `## Log`.

## Plan

- [ ] Edit `AGENTS.md:49` (base layer) — measured-not-typed wording.
- [ ] Edit `construct/local/issues/SKILL.md:17`.
- [ ] Edit `cmd/sdlc/helptext/milestone-close.md` — add the computed/suggested note.
- [ ] Edit `cmd/sdlc/helptext/close.md:7-8`, `helptext/set-status.md:24`, `cmd/sdlc/setstatus.go:207` — annotate/point at `sdlc actual`.
- [ ] `go build ./cmd/sdlc`; eyeball `--help` for close + milestone-close; grep-sweep for remaining bare examples.
- [ ] Update atlas only if surface/terminology changed (likely `--no-atlas`: pure prose; sdlc-binary.md already states the truth).

## Log

### 2026-06-05

Filed from the nous#42 retro: agent hand-passed a fabricated `--actual 13.5` (measured 0.30h)
because AGENTS.md §5 and the issues SKILL show `--actual h` with no pointer to `sdlc actual`.
Sibling **#87** adds the code-level deviation backstop. Audit of all `--actual` refs captured in
Problem above.
