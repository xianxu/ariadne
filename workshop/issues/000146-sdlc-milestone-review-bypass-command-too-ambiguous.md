---
id: 000146
status: codecomplete
deps: []
created: 2026-06-30
updated: 2026-07-05
started: 2026-07-05T16:28:52-07:00
estimate_hours: 0.53
actual_hours: 1.35
---

# sdlc milestone review bypass command too ambiguous

agent used `sdlc close --milestone Mx`, instead of `sdlc milestone-close`. `sdlc close --milestone Mx` looks too innocent. there should be some text like "force" etc. to indicate it is skipping normal path. 

let's start with checking what does th `sdlc close` verb do, and then systematically update escape patches to have word force in it, e.g. `sdlc force-close --milestone Mx`. I'm not sure about the actual contour of the verbs, let's do some investigation first. 

## Done when

- `sdlc close --milestone Mx` no longer silently performs a no-review milestone
  close: the flag is removed from `close`'s public surface, and invoking it
  refuses with a redirect to `sdlc milestone-close` (reviewed) / `--no-judge`
  (explicit skip).
- The milestone close's actual/verified gate re-run hints suggest `sdlc
  milestone-close …`, never `sdlc close --milestone …` (the misdirection is gone).
- `milestone-close`'s mechanical close (via `computeClose` directly — #139's
  compute→review→finalize, NOT `runClose`) is unchanged.
- Docs updated: `helptext/close.md`, `helptext/milestone-close.md`,
  `milestoneclose.go` doc comment, and the base-layer `AGENTS.base.md` (`[--milestone
  Mx]` dropped from the close example — propagates downstream).
- The composed entry files (`CLAUDE.md`/`AGENTS.md`/`GEMINI.md`, gitignored) are
  regenerated via `make weave` so this repo's own agent-facing docs no longer
  advertise `--milestone` (shadow-sweep the derived consumers, not just the source).

## Spec

**Root cause (investigated — see Log).** `sdlc close --milestone Mx` and
`sdlc milestone-close --milestone Mx` share the same mechanical close body
(`runClose`); the ONLY difference is that `milestone-close` auto-dispatches the
mandatory fresh-context review while `close --milestone` structurally
short-circuits before it (`close.go:785 return runClose`) — no review, no
`Review-Verdict:` trailer, no warning. Two failure modes: (1) the `--milestone`
flag reads like a normal parameter, not a bypass; (2) worse, the actual-gate
error hint (`explainActual`/`explainVerified`) tells an agent who forgot
`--actual` on `milestone-close` to re-run with `sdlc close --milestone …` — the
non-reviewing path (documented in `workshop/pensive/2026-07-01-01-…`).

**Reframe.** Every OTHER escape affordance in the binary is already self-labeling
(`--force`, the eleven `--no-<gate>` flags — all announce they turn something
off). `close --milestone` is the sole unlabeled escape — and it is a **redundant
second spelling** of `sdlc milestone-close --no-judge` (which already exists and
prints "skipping milestone-review per --no-judge"). So the operator's original
"add force to escape names" reduces to fixing this one outlier; the elegant fix
is **removal**, not a rename (ARCH-DRY / Simplicity-First: kill the redundant
footgun, keep the labeled path).

**Chosen design (operator-approved).** Remove `--milestone` from `sdlc close`'s
public surface. Keep the flag registered but hidden so `sdlc close --milestone Mx`
still parses (no ugly cobra "unknown flag") and instead **refuses** with a
redirect:

```
Error: `sdlc close` no longer closes milestones (it would skip the boundary review).
  reviewed:       sdlc milestone-close --issue N --milestone Mx
  explicit skip:  sdlc milestone-close --issue N --milestone Mx --no-judge
```

`milestone-close` is untouched: it calls `runClose` (with the milestone set)
in-process, NOT via the cobra `close` command, so its mechanics keep working.
Fix both re-run hints to pick the verb by mode via a small `closeVerb(milestone)`
helper (extracted from the existing inline selection at `close.go:868-871`, ARCH-DRY).

**Out of scope / non-goals.** No new `force-*` verb (the labeled `milestone-close
--no-judge` already covers "skip the review"). No change to the `--no-*`/`--force`
conventions — they already signal correctly. Auto-redirecting (silently running
the reviewed close) is rejected: the operator wanted a SIGNAL, and a refusal
teaches the correct verb where a silent redirect hides the correction.

## Estimate

```estimate
model: estimate-logic-v3.1
familiarity: 1.0
item: smaller-go-module   design=0.15 impl=0.2
item: atlas-docs          design=0.05 impl=0.1
design-buffer: 0.15
total: 0.53
```

*Produced via `brain/data/life/42shots/velocity/estimate-logic-v3.1.md` against `baseline-v3.1.md`. Method A only.*
Decomposition: (1) `close.go` — extract `closeVerb`, repoint the two re-run hints,
hide+refuse the `--milestone` flag, + unit tests (one focused Go file); (2) docs —
`helptext/close.md` + `helptext/milestone-close.md` + `milestoneclose.go` doc comment
+ base-layer `AGENTS.base.md` + atlas. Design trimmed (plan pre-resolves the fork +
call sites); `impl=` at v3.1's 40%; familiarity 1.0 (warm — just shipped #145 in this
same `close.go`/vocab neighborhood).

## Plan

Durable design: [`workshop/plans/000146-remove-close-milestone-bypass-plan.md`](../plans/000146-remove-close-milestone-bypass-plan.md).
Single-pass (no `Mx` — one `sdlc close`).

- [x] Extract `closeVerb(milestone)` helper; reuse in `reviewThenFinalize` (DRY).
- [x] Repoint `explainActual` + `explainVerified` re-run hints to `milestone-close` (kill the misdirection) — via a pure `rerunCmd` builder (ARCH-PURE).
- [x] Hide `--milestone` on `close` + refuse with a redirect to `milestone-close` (the mechanical `runClose(Milestone=…)` path stays for milestone-close). Repurposed the obsolete guard test.
- [x] Docs: `helptext/close.md`, `helptext/milestone-close.md`, `milestoneclose.go` comment, `close.go` comment, base-layer `AGENTS.base.md`; then `make weave` recomposed the gitignored `CLAUDE.md`/`AGENTS.md`/`GEMINI.md`.
- [x] Build + `go test ./...` (25 pkgs green) + manual e2e (refusal fires, `--help` drops flag, hint points at milestone-close, entry files flag-free); atlas; close.

## Log

### 2026-06-30

Filed: agent used `sdlc close --milestone Mx` instead of `sdlc milestone-close`.

### 2026-07-05
- 2026-07-05: closed — Third re-review: comprehensive runClose→computeClose comment sweep (milestoneclose.go 3/46/90/369-371, close.go 65/153/778, atlas 520). No production runClose reference is inaccurate now (grep-verified). No behavior change since anchor 94ea6e9 — comment/doc accuracy + the Makefile milestone-close routing (verified DRY-run) + atlas call-graph. go build/vet/test ./cmd/sdlc green.; review verdict: SHIP
- 2026-07-05: closed — Re-review the Makefile.workflow executable-consumer fix + atlas call-graph corrections (found by the prior re-close). make close-issue MILESTONE=Mx now routes to milestone-close (verified DRY-run); go test ./cmd/sdlc green. No product Go change since anchor 94ea6e9 — this delta is the Makefile target + atlas/docs.; review verdict: FIX-THEN-SHIP
- 2026-07-05: closed — Re-close to re-review the post-close delta (lessons.md shadow-sweep-at-section-granularity addition; the FIX-THEN-SHIP fixes are within anchor 94ea6e9). No product-code change since the reviewed close. go build/vet/test ./... — 25 pkgs green. close --milestone removed + refuses with redirect; hints point at milestone-close; docs+entry files swept.; review verdict: FIX-THEN-SHIP
- 2026-07-05: closed — go build/vet/test ./... — 25 pkgs, 0 failures. Removed the unlabeled close --milestone no-review bypass: runCloseWithReview refuses a milestone with a returnable-error redirect to `sdlc milestone-close` (reviewed) / `--no-judge` (labeled skip); flag hidden. E2E: `sdlc close --issue 999 --milestone M1` refuses (before issue-existence); `close --help` FLAGS drops --milestone; no `close --milestone` in helptext/AGENTS.base.md/entry files (make weave recomposed). Hints (explainActual/explainVerified) point at milestone-close via pure rerunCmd+closeVerb. milestone-close mechanics unchanged (in-process runClose; repurposed guard test asserts the refusal). Prior close-review verdict was spuriously "unknown" (empty review body — transient dispatch failure), now retried.; review verdict: FIX-THEN-SHIP

Investigated the verb/escape-hatch contour (Explore digest + direct read).
Confirmed the bug is real and precise (`close.go:781-787` short-circuit skips the
review that `milestone-close` dispatches). Enumerated every escape affordance:
all are self-labeling flags (`--force`, `--no-<gate>` ×11 across close/
milestone-close/change-code/merge/push) EXCEPT `close --milestone`. Found the
misdirecting re-run hint (`explainActual` `close.go:942`, `explainVerified`
`close.go:1048`) that points agents at the no-review path — the exact trap the
`2026-07-01-01` pensive recorded ("I fell into this"). Key: `sdlc milestone-close
--no-judge` already exists as the labeled skip, so `close --milestone` is a
redundant footgun. Brainstormed the fork with operator → **remove the flag**
(over refuse-unless-`--no-judge` / rename-to-`force-close`). Spec written above.

change-code plan-quality judge: **FAILURE** → two findings folded into the plan
before re-run: (1) [Important, ARCH-PURPOSE] the `[--milestone Mx]` close example
also lives in the composed entry files `CLAUDE.md`/`AGENTS.md`/`GEMINI.md` (weave
outputs of `AGENTS.base.md`) — must `make weave` to regenerate them (verified
gitignored, `.gitignore:22-24`, so no commit, but the recompose is required else
this repo's own docs stay stale); (2) [Info] the refusal must be a **returnable
error**, not `die()` (`die`→`os.Exit(1)` would kill the test binary before
`cmd.Execute()` returns; `runCloseWithReview` already returns `error` under
`SilenceErrors`, the `main.go:47-49` soft-error path). Both adopted.

change-code plan-quality re-run: **FAILURE** again (2 more valid catches, folded
in): (1) [Important] `TestRunCloseWithReview_MilestoneClose_DoesNotDispatch`
(`closereview_test.go:253`) calls `runCloseWithReview` DIRECTLY with `Milestone:"M1"`
and asserts success — it guards the exact short-circuit being removed, so it must be
**repurposed** into a refusal assertion (my "must stay green" claim was wrong; the
in-process `runClose` path via milestone-close is what's untouched, not this direct
caller). (2) [Info] `helptext/close.md:96` has a hand-authored `--milestone <Mx>`
FLAGS entry — must be deleted or `sdlc close --help` keeps listing the flag. Plan
updated: Task 3 Step 3b repurposes the test; Task 4 Step 1 names close.md:96.

change-code round 3: plan-quality **INFO (pass)**; estimate-quality advisory-only
(0.53 slightly low — a known v3.1 property, no change). One INFO folded in
(close.md:18 "both forms remain valid" clause + broadened verify grep). Re-crossed
with `--no-judge` (judges adjudicated).

**Implemented + verified.** `go build/vet/test ./...` — 25 pkgs, 0 failures. E2E:
(1) `sdlc close --issue 999 --milestone M1 …` refuses with the milestone-close
redirect (fires before issue-existence, as designed); (2) `close --help` FLAGS
block no longer lists `--milestone`; (3) no `close --milestone` bypass anywhere in
`helptext/`, `AGENTS.base.md`, or the recomposed `CLAUDE.md`/`AGENTS.md`/`GEMINI.md`.
Tests: `TestCloseVerb`, `TestRerunCmd` (hints → milestone-close), `TestClose_
MilestoneRefusesWithRedirect` (command-level), `TestRunCloseWithReview_Milestone
Refuses` (repurposed guard). `make weave` recomposed the entry files (settings.json
write needed sandbox off — deny-listed path, unrelated to this change).

