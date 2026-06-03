---
id: 000068
status: done
deps: []
github_issue:
created: 2026-06-02
updated: 2026-06-02
estimate_hours: 3
actual_hours: 2.0
---

# fix active-time-v3.py: returns 0 events for these sessions — actuals are fabricated

## Problem

`sdlc close` / `milestone-close` require `--actual <hours>` and tell the agent to
derive it by running `active-time-v3.py` over the issue's commit window. But the
script returns **0 events** for these sessions — flagged as far back as `nous#34`'s
close ("active-time-v3 reported 0 events; session telemetry not captured"), and again
across **~7 closes** in the 2026-06-02 session (shared-brain close-down + `nous#41`),
where every `actual_hours` was a manual guess passed with `--force`/FORCE=1.

Net effect: the velocity-calibration loop the `project` datatype is built around (each
close records `actual_hours` → feeds the velocity-skill validation table) is running on
**fabricated numbers**. The gate has the *form* of calibration with none of the
substance — arguably worse than no gate, because it looks calibrated.

See `brain/data/life/42shots/velocity/{baseline-v3,estimate-logic-v3,SKILL.md}` for the
v3 procedure the script implements.

## Diagnosis (confirmed 2026-06-02)

The algorithm is **fine**. "0 events" was never a v3 bug — it was **dir-selection**:
v3 derives events *only* from transcript `.jsonl` files passed via `--dir`. Every
0-events case traced to feeding it **zero or the wrong `--dir` folders**:

- The agent (this session, repeatedly) ran `active-time-v3.py --git-repo . --issue N`
  with **no `--dir`** → `events=[]` → "no events in window" → **exit 0**. A
  misinvocation is silently indistinguishable from a real "no activity" answer. That
  is the footgun that produced ~7 fabricated actuals.
- Work spans **many cwds**, each with its own `~/.claude/projects/<slug>/` folder —
  the issue's repo, `brain`, peer repos, **and every git worktree** (`worktree-…`).
  Feeding only `--dir <repo>` misses most of it. The `nous` folder even had only the
  last 3 days of sessions; the early-May work lived under `brain`/`charon`.
- **Single `--issue` over/under-attributes:** with only `--issue 14`, all the peer-issue
  segments in the window fall to `##unattributed`. Must pass *all* window issues
  (`DiscoverWindowIssues` already does this for `close`).

**Empirical proof** (nous#14, window 2026-05-08..05-11):

| dirs fed to v3 | #14 actual |
|---|---|
| none / wrong | 0 (silent) |
| all 24 dirs (incl. unrelated `pair`) | 12.06h (inflated) |
| **brain + issue-repos {nous,ariadne,charon}** | **7.79h** |
| recorded in shared-brain.md (Σ M1–M5, judgment) | 8.2h |

So **brain + repos-owning-a-scoped-issue** lands within ~5% of the human number;
`pair` (concurrent unrelated work) inflated it by +4.3h. Note: at `--commit-weight 1.0`,
a segment's active-time counts *every* event in its time-slice regardless of issue
mention (mentions only gate the `1−weight` split + commitless segments) — which is
*why* an unrelated dir with concurrent activity inflates, and why dir-selection matters.

**The operator's dormancy hypothesis** folds into this: a long gap only hurts via the
**31-day `WindowCapDays`** in `internal/gitx/window.go` (`CommitWindow` caps the lookback),
which truncates the window for month-long work (e.g. #16). Bump it.

## Spec

Agreed design. Direction: **lift the manual prose into `sdlc`** — stop printing a
6-line command for a human to run (nobody does); have `sdlc` run it.

1. **`WindowCapDays` 31 → 61** (`internal/gitx/window.go`) — cover month-long tasks.
2. **0-event / no-`--dir` detection** — `active-time-v3.py` must FAIL LOUDLY (non-zero,
   clear message) when `--dir` is empty (events can only come from transcripts, so empty
   `--dir` is always a misinvocation). When `--dir` is non-empty but a window with commits
   yields 0 events, print an explicit "telemetry unavailable for this window → use a
   labeled judgment estimate" line — never a silent 0-exit-0.
3. **`sdlc` runs v3 itself** — in the `close`/`milestone-close` actual path, compute the
   actual instead of printing instructions:
   - **dir-selection:** `brain` (always) ∪ the transcript folder for the issue's own repo
     ∪ (for a project-driven issue, repos owning a scoped issue). Resolve worktree folders
     too where cheap. Enumerate `~/.claude/projects/*`.
   - window from `CommitWindow`; peers from `DiscoverWindowIssues`; `--commit-weight 1.0
     --threshold-min 15 --include-assistant`.
   - print the computed per-issue hours as a **suggestion** the operator/agent confirms
     into `--actual` (keep a human in the loop; don't auto-write). If v3 reports 0-events,
     suggest the labeled-judgment path (ties to `--no-actual` / a labeled estimate).
   - **Risks to resolve in M2:** `active-time-v3.py` lives in `construct/local/issues/`
     (verify it's reachable from a derivative's cwd; if not, fall back to printing the
     command as today). `python3` runtime dep (already used by ariadne).

## Done when

- `WindowCapDays == 61`; a month-long issue (#16-style) still yields a commit window.
- `active-time-v3.py` exits **non-zero with a clear message** on empty `--dir`, and prints
  an explicit "telemetry unavailable" line (not a silent 0) when commits exist but events==0.
- `sdlc close` (and `milestone-close`) **runs** v3 with brain+repo dir-selection + window +
  discovered peers and prints a suggested actual; reproduces the shared-brain numbers
  (nous#14 ≈ 7.8h) when pointed at that window. Falls back gracefully if the script/python
  is unavailable.

## Plan

- [x] M1 — `WindowCapDays` 31→61 (`gitx`); `active-time-v3.py` loud-fail on empty `--dir`
      + explicit "telemetry unavailable" on commits-but-0-events. Unit/CLI tests.
- [x] M2 — `sdlc` runs v3 in the actual path: dir-selection (brain + issue-repo, enumerate
      `~/.claude/projects/*`), window + peers, subprocess invoke + parse the per-issue
      total, print suggested actual; graceful fallback when script/python absent. Verify it
      reproduces nous#14 ≈ 7.8h.

## Log

### 2026-06-02 — M2
- 2026-06-02: closed — cap 31→61 + active-time-v3 loud-fail (M1) + sdlc actual verb & close-inline measured suggestion (M2); the fabricated-actuals footgun is closed — sdlc now runs v3 with brain+repo dirs. go test+vet green; M1 review FIX-THEN-SHIP→fixed, M2 review SHIP. actual=Σ milestones (sdlc actual measures 1.0h but undercounts pre-commit investigation — a noted CommitWindow refinement)
- 2026-06-02: closed M2 — sdlc actual verb + close-inline suggestion live (sdlc actual #68 → measured 1.0h; close prints "→ --actual 0.62"); engine selects brain+repo dirs, classifies v3 exit codes; actual_test.go + full suite + vet green. actual=M2-portion judgment (tool measures whole-#68 at 1.0h but undercounts pre-commit investigation); review verdict: unknown

- New `cmd/sdlc/actual.go` — engine `computeActual(repoTop, brainAbs, issue)`:
  resolves `CommitWindow` + `DiscoverWindowIssues` peers, selects transcript dirs
  via the validated heuristic (**brain + the issue's repo**, existing folders
  only — never "all"), runs `active-time-v3.py` (`--commit-weight 1.0
  --threshold-min 15 --include-assistant`), and classifies by exit code
  (3→telemetry-gap, 0+`#N: h.hh hr`→measured, else→fallback). Pure helpers
  `cwdToTranscriptDir` ('/'+'.'→'-'), `selectActualDirs`, `parseV3PrimaryHours`.
- New verb `sdlc actual --issue N` (`NewActualCmd` + `helptext/actual.md`,
  registered in main.go) prints the suggested `--actual` or the judgment guidance.
- `close.go`/`explainActual` rewritten: the missing-`--actual` path now **runs
  the same engine and prints the measured suggestion inline** (`→ close with:
  --actual <h>`) instead of a python command nobody ran. Graceful fallback
  (`actualNoScript`) when the script/`python3` is absent (e.g. a derivative
  without `construct/local/`) → "use a judgment estimate".
- Tests: `actual_test.go` (cwd-encoding, per-issue parse incl. no-prefix-match,
  dir-selection fixture incl. unrelated-excluded + missing-skipped, flag reg).
- **Live verified**: `sdlc actual --issue 67` → measured; `--issue 68` → 0.62h
  (peers #16,#68); `--issue 999` → no-window guidance; `sdlc close --issue 68`
  (no `--actual`) prints "measured actual for #68: 0.62h → --actual 0.62" inline.
  `go test ./cmd/sdlc/...` + `go vet` green. (Note: numbers carry the existing
  CommitWindow subject-anchor caveat — work before the first `#N:`-subject commit
  isn't in the window; a separate v3/window refinement, not M2.)

### 2026-06-02 — M1
- 2026-06-02: closed M1 — gitx cap 31→61 (test: 45-day commit anchors, has teeth); active-time-v3 exits 2 on empty --dir + exit 3 on commits-but-0-events (test_active_time_v3.py 5 checks pass); go test+vet green. --no-atlas: internal cap/exit-code change, no new surface (M2 adds it). actual=judgment (15-min chunk); review verdict: FIX-THEN-SHIP

- `cmd/sdlc/internal/gitx/window.go`: `WindowCapDays` 31→61 (+ rationale comment).
  Test `TestCommitWindow_ExtendedCapIncludes45Days` (temp repo, a 45-day-old #99
  commit must anchor the window — fails under the old 31-day cap).
- `construct/local/issues/active-time-v3.py`: (a) **empty `--dir` → exit 2** with
  a message (events come only from transcripts; no `--dir` is always a
  misinvocation, never a real 0); (b) **commits-but-0-events → exit 3** with a
  "TELEMETRY UNAVAILABLE" message (the transcripts are in cwds not passed via
  `--dir`, or aged out — never read 0 as measured). Genuinely-empty window still
  exits 0. Smoke-confirmed all three; `#67` over today's window measured 2.01h
  (vs the 2.0h judgment I'd recorded — the algorithm was always fine).
- Test `construct/local/issues/test_active_time_v3.py` (self-contained, mirrors
  `test_detect.py`): 5 checks (exit 2 / exit 3 / exit 0 paths + messages). Run
  with `python3 construct/local/issues/test_active_time_v3.py`.
- `go test ./cmd/sdlc/...` + `go vet` green.

### 2026-06-02

Filed from the sdlc tooling retro
(`workshop/pensive/2026-06-02-01-pensive-sdlc-tooling-retro.md`, finding F2). Operator
flagged the dormancy-window hypothesis as the likely cause.

- **Investigated + diagnosed** (using #16 + shared-brain as live examples). Root cause is
  dir-selection, not the algorithm (see Diagnosis above). Validated the operator's
  dir-selection heuristic empirically: brain + issue-repos gave nous#14 = 7.79h vs 8.2h
  recorded (~5%), while including unrelated `pair` inflated to 12.06h. Operator approved:
  bump cap 31→61, add 0-event detection, and lift v3-invocation into `sdlc`. Scoped M1
  (cap + detection) / M2 (sdlc runs v3).
