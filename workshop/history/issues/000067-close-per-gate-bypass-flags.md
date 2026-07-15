---
id: 000067
status: done
deps: []
github_issue:
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

- [x] Implement the `--no-<gate>` family on `close` (+ `milestone-close`
  pass-through), audit warns, message rewrites, docs (close.md + AGENTS.md §5),
  and tests (`skip()` logic + flag registration). Verify per the plan's manual
  matrix; dogfood by closing this issue (real atlas change → no `--no-atlas`
  needed for its own close).

## Log

### 2026-06-02
- 2026-06-02: closed — go test ./cmd/sdlc/... + vet green; TestCloseFlags_Skip proves per-gate isolation; e2e via --dry-run: --no-atlas waives atlas but plan-check still fires; (3) both → reaches dry-run; fresh-eyes review SHIP (all 7 gate rewrites preserve semantics, audit only on would-refuse path)

- Implemented in `close.go`: 7 `No*` bool fields on `closeFlags` + 7 `--no-*`
  flags; `skip(gate string) bool` (= `Force || <field>`) as the single arbiter;
  rewrote each of the 7 gate sites to `if <fail> { if !f.skip("…") {refuse}; cwarn(audit) }`
  so a bypass only fires + logs when the gate would actually have refused.
- `--no-actual`/`--no-verified` get a stronger caution line (they weaken velocity
  calibration); the others log a plain `[!] --no-X (or --force): skipping …`.
- **Message modernization** (the stale-`FORCE=1` part of the issue): `FORCE=1`
  was a Makefile/Python-era var that *nothing reads* — rewrote the explainers
  (`explainActual`/`explainVerified`/`explainNoAtlas`/`explainMissingVerdicts`),
  the re-close + plan + project dies, and the warmup contract to name the precise
  `--no-<gate>` flag, and modernized their `make close-issue … ACTUAL=…` re-run
  hints to current `sdlc close --issue … --actual … --verified …` syntax.
- `milestone-close` pass-through: mirrored the 7 `--no-*` flags on
  `milestoneCloseFlags`, threaded into the delegated `closeFlags`.
- Docs: `close.md` (gate→flag table + FLAGS), AGENTS.md §5 (the convention —
  flag = explicit acknowledgment, not forgetting; `--force` = all), atlas
  `sdlc-binary.md` (per-gate bypass + `skip()` arbiter).
- Tests: `TestCloseFlags_Skip` (force ⇒ all; each `--no-X` gates only its own;
  clean set skips none; unknown gate not skipped), `TestCloseCmd_Registered`,
  `TestMilestoneCloseCmd_RegistersBypasses`. `go test ./cmd/sdlc/...` + `go vet`
  green.
- **End-to-end behavioral check** (via `--dry-run`, gates run before the
  write-skip): (1) no bypass → atlas gate fires naming `--no-atlas`; (2)
  `--no-atlas` → atlas waived but plan-check STILL fires (per-gate isolation
  proven); (3) `--no-atlas --no-plan-check` → both waived (audit lines), reaches
  `DRY=1 — no files written`.
- gofmt: `close.go`/`milestoneclose.go` flagged by `gofmt -l`, but the diff is
  entirely pre-existing newer-gofmt doc-comment reflow on untouched lines (same
  drift noted in #63); my added code is clean — didn't reformat unrelated lines.
- Behavior change for the agent: prefer the precise `--no-atlas` (reason in
  `--verified`) over `--force` for pure bugfixes going forward.
- Follow-up (noted, out of scope): extend `--no-<gate>` to `merge`/`push` gates.
