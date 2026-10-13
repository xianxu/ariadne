---
id: 000342
status: codecomplete
deps: []
github_issue:
created: 2026-10-12
updated: 2026-10-13
estimate_hours:
card_mirror: '3800d5a0acb530d861e42644e069fe04ca22928b' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-13T15:47:17-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: MacBook Pro
    workspace: ariadne:3
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: full, provenance: inferred}
actual_hours: 0.28
---

# agents: make temp dirs from $TMPDIR; never cd into a possibly-empty path (macOS mktemp ignores TMPDIR)

## Problem

On macOS, `mktemp` / `mktemp -d` / `mktemp -t x` without a path template ignore `TMPDIR` and create under the per-user `/var/folders/.../T`, which the Claude Code sandbox denies (measured by pair:7, 2026-10-12; the same mkdtemp failure appears in transcripts since 2026-09-11 across ariadne, tools, pair, brain, parley.nvim and slots). A probe script that then `cd "$d"` with an empty `$d` lands in HOME: on 10-12 that wrote `~/Makefile` and `~/inc.mk` (ariadne#299's probe), the second home write that day (cf. #339).

## Spec

Fix it where agents and scripts learn and run:
1. **AGENTS rule** (in the shared base, so every repo's agents get it): make temp dirs with `mktemp -d "${TMPDIR:-/tmp}/name.XXXXXX"`, and never `cd` (or write) into a path built from a variable that may be empty: use `"${d:?}"` or `|| exit 1`.
2. **Sweep this repo's scripts** for template-less `mktemp` and fix them the same way; a test (or lint check in an existing guard) that refuses a template-less `mktemp` in tracked shell scripts.
3. No settings or machine change: allowing `/var/folders/.../T` in the sandbox would put agent churn back on the journaled Data volume (pair#496).

## Done when

- The AGENTS rule is in the base and reaches peers through weave.
- No tracked script in ariadne uses a template-less `mktemp`, and the check fails on one.

## Plan

- [x] AGENTS rule in `AGENTS.base.md` §10 (the shared base weave composes into every
  repo's AGENTS/CLAUDE files): temp dirs via `mktemp -d "${TMPDIR:-/tmp}/name.XXXXXX"`;
  never `cd`/write into a path from a possibly-empty variable (`"${d:?}"`, `|| exit 1`).
- [x] Pure scanner `pkg/shelllint` (`MktempFindings(src) []Finding`), unit-tested on
  the failing forms: no template, `-t prefix`, a literal `/tmp/...XXX` template, a
  Makefile `$$(mktemp -d)`, and `cd` into a `$(mktemp ...)` substitution (bash's
  `cd ""` succeeds and stays put). Comments and quoted text are not code.
- [x] Repo test over tracked shell scripts (`.sh`, Makefiles/`.mk`, shell shebangs;
  `workshop/` excluded) that fails on any finding. One exemption,
  `scripts/parallel-checks.sh` (a local-gates path; #351 fixes it with the operator's
  verification); the test fails if an exemption no longer has a finding.
- [x] Sweep: `.colima/test/colima.test.sh`, `.openshell/overlay/setup.sh` (3),
  `Makefile.workflow` drift check (2), `scripts/docflow.test.sh`,
  `construct/scripts/test/clone-data-deps.test.sh` (cd into mktemp), and the
  session-retro skill's `/tmp` template.
- [x] Atlas note; full suite.

## Log

### 2026-10-12

### 2026-10-13 — survey (ariadne:3)
- 2026-10-13: closed — Delta since 4e2d53b9 is the two advisory minors: pkg/shelllint command positions now include a brace group, prefix commands (sudo, time, env, nohup, nice) and env assignments; the cd rule covers pushd. New table rows (brace group, sudo, time, VAR=x, pushd) failed before and pass now; a ${mktemp} expansion stays clean; repo scan TestTrackedScriptsMakeTempDirsFromTMPDIR still clean. go test ./pkg/shelllint green. Full suite green at 15:54. (fixed-then-ship); review verdict: SHIP
- 2026-10-13: closed — Full make test green 15:54 (12 shards, 1120 cmd/sdlc tests, 47 other packages incl. new pkg/shelllint, 0 failed). pkg/shelllint: MktempFindings table tests (failing forms + look-alikes) and TestTrackedScriptsMakeTempDirsFromTMPDIR over tracked shell scripts/Makefiles; before the sweep it reported exactly the 8 surveyed sites, after it none; an appended probe=$(mktemp -d) fails it; locally fixing the exempt scripts/parallel-checks.sh (reverted) fails it with remove the exemption. Swept scripts run: colima.test.sh pass (original fails here with mkdtemp failed on /var/folders: Operation not permitted), docflow.test.sh ALL PASS, clone-data-deps.test.sh 9/9, make weave-drift-check OK, setup.sh bash -n. AGENTS.base.md §10 rule composed into AGENTS.md/CLAUDE.md by weave compile. parallel-checks.sh (local-gates path) filed as #351, unclaimed, hub told.; review verdict: SHIP
- 2026-10-13: flow upgraded quick → full — 159 added lines in code files (limit 100)

Template-less `mktemp` in tracked executables: `.colima/test/colima.test.sh:7`,
`.openshell/overlay/setup.sh:68,72,73`, `Makefile.workflow:209` (x2),
`scripts/docflow.test.sh:27`, `scripts/parallel-checks.sh:144`. The last is protected by
`local-gates-guard`, so it is filed as #351 (unclaimed; the operator verifies local-gates
changes) and exempted by name here. `construct/scripts/test/clone-data-deps.test.sh:13`
has a template but `cd`s into the substitution: a failed mktemp leaves ROOT = cwd.
`cmd/sdlc/internal/judge/confine_unix_test.go` holds a deliberate template-less
reproduction of #339 inside a Go string; Go files are not scanned. The check is a Go test
because `make test` runs only Go and CI's merge-check runs the local gates, which this
issue must not touch.

### 2026-10-13 — implemented (ariadne:3)

`pkg/shelllint.MktempFindings` (pure, table-tested on every failing form and the
look-alikes: comments, words in strings, `$mktemp`, identifiers) and
`TestTrackedScriptsMakeTempDirsFromTMPDIR` over tracked `.sh`/`.bash`/`.zsh`/`.mk`,
Makefiles and shell shebangs. Before the sweep it reported exactly the eight surveyed
sites and nothing else. Proofs: an appended `probe=$(mktemp -d)` fails it; fixing the
exempted `scripts/parallel-checks.sh` locally (reverted, never committed) fails it with
"remove the exemption". Swept scripts run: `colima.test.sh` passes (the original fails in
this sandbox with `mkdtemp failed on /var/folders/…: Operation not permitted`, the issue's
error), `docflow.test.sh` ALL PASS, `clone-data-deps.test.sh` 9/9, `make
weave-drift-check` OK, `setup.sh` (OpenShell container) `bash -n` only. The rule is in
`AGENTS.base.md` §10 and `weave compile` composes it into AGENTS.md/CLAUDE.md (the compile
then hit the sandbox's denial of `.claude/settings.json`, a protected path, unrelated).
Atlas: sandbox.md *Temp dirs*.

Close review (SHIP, round 1) left two advisory Minors, both fixed: `mktemp` after a
brace group, a prefix command (`sudo`, `time`, `env`, `nohup`, `nice`) or an env
assignment (`VAR=x mktemp`) is now a command position, and the cd rule covers `pushd`.
New table rows for each fail without the change; a `${mktemp}` expansion stays clean,
and the repository scan stays clean.

Close review round 2 (SHIP; "not converging: fix rules, not instances") found more
missed command positions (if/elif/while/until, `sudo -u x`, `time -p`, `env -i`), the
same family as round 1. Fixed as a rule: `commandWords` now treats every standalone,
unquoted `mktemp` in code as a call (`$(…)` inside double quotes is code; literal
quoted text is not), excepting identifiers, `$mktemp`/`${mktemp}` expansions and the
lookups that only name it (`command -v`, `which`, `type`, `hash`). The enumerated
predecessor list and its `VAR=x` regex are gone; the regex had also misread
`m=$(which mktemp)` as a prefix assignment. New rows for each opener and option form
fail on the old rule; lookup and single-quoted rows stay clean; repo scan clean.
