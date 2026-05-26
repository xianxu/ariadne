---
id: 000031
status: working
deps: []
github_issue:
created: 2026-05-25
updated: 2026-05-25
estimate_hours:
actual_hours:
---

# sdlc checkpoint binary

## Problem

Two related problems, captured in `docs/vision/2026-05-25-01-pensive-sdlc-checkpoint-binary.md`:

1. **Progressive-disclosure leak.** Markdown-index skills lose context-budget control when agents glob `**/*.md`. The disclosure boundary is conventional, not enforced.
2. **Workflow drift defense lives in Make.** Ariadne's checkpoint guards (close-issue, fetch, worktree, push, pull-request, merge, issue-sync, check-*) live as Makefile targets with weak arg parsing, env-export shims, no embedded help, and no shared discoverability surface.

The lift: a single Go binary `sdlc` that collects existing checkpoint guards as subcommands, ships its own guidance via `sdlc --help` / `sdlc <verb> --help`, exposes workflow state via `sdlc state`, and names the anti-collusion judge pattern as a first-class primitive (`sdlc judge ...`).

## Spec

### Binary and build

- Language: **Go**. Matches `cmd/nous/` in ../nous and ariadne's existing `make build` convention.
- Layout: `cmd/sdlc/main.go` + `cmd/sdlc/internal/<verb>/` per subcommand.
- Per-subcommand `help.md` embedded via `//go:embed` and served by `sdlc <verb> --help` (rich prose + flags + examples, not just usage line).
- Build target: `make sdlc-build`, mirroring `make nous-build` in ../nous:

  ```makefile
  sdlc-build:
  	@mkdir -p bin cmd/sdlc/bin
  	@echo "building cmd/sdlc/bin/sdlc"
  	@go build -o cmd/sdlc/bin/sdlc ./cmd/sdlc
  	@ln -sf ../cmd/sdlc/bin/sdlc bin/sdlc
  ```
- `make build` (the existing cmd/*/main.go scanner in `Makefile.workflow`) will also pick `sdlc` up automatically — `sdlc-build` is the explicit dev-flow target.
- `make sdlc-bootstrap` — first-time install on a fresh machine. Idempotent. Mirrors `make nous-bootstrap` in ../nous: verifies Go toolchain present, runs `sdlc-build`, places `sdlc` on PATH (e.g., symlink `bin/sdlc` into `~/bin/` or `go install ./cmd/sdlc`). One-shot for developers cloning ariadne to get `sdlc` working from any directory.

### Skill surface

Thin SKILL.md at `construct/local/sdlc/SKILL.md` pointing at `sdlc --help`. No prose duplication. Skill becomes greppable signpost; binary is source of truth. Convention beats invention — agents try `--help` reflexively, no new verb to teach.

**SKILL.md generation.** `sdlc --index` emits the SKILL.md content (frontmatter + body) from the same `//go:embed`-backed prose that powers `--help`. The on-disk SKILL.md becomes regenerable, not hand-maintained:

    sdlc --index > construct/local/sdlc/SKILL.md

One source of truth in the binary; two renderings (one for CLI `--help`, one for the skill loader). No drift possible between what the binary does and what the skill says.

**Workflow stages (prose, not code).** The SKILL.md describes the SDLC workflow in prose — not as a state machine. We don't formalize the journey, we name the stages so agents know which checkpoint they're at and which tools live at each stage:

1. **Ideation** — `workshop/parley/`, `docs/vision/` (pensives)
2. **Brainstorming** — `superpowers-brainstorming`
3. **Planning** — `superpowers-writing-plans` → `workshop/plans/`
4. **Build** — `superpowers-executing-plans` with milestones in `workshop/issues/`
5. **Milestone review** — `sdlc judge dry|pure|plan|specs|lessons` (auto-dispatched from `sdlc milestone-close`)
6. **Test** — currently manual; web-app work will borrow `playwright` and/or gstack's browser integration
7. **Close / ship** — `sdlc close` → `sdlc push` (direct-on-main) or `sdlc pr` → `sdlc merge` (branch)
8. **Postmortem / introspect** — `xx-introspect`, `workshop/lessons.md`

Same shape as gstack's Think → Plan → Build → Review → Test → Ship → Reflect, with ariadne-specific tooling at each stage. The `sdlc` binary owns the *checkpoints between stages*; the stages themselves remain prose and human-driven.

### Lift table — Make targets → `sdlc` subcommands

**Flag convention.** `--issue N` always refers to the ariadne workshop issue ID (6-digit, in `workshop/issues/` or `workshop/history/`). `--github-issue N` refers to a GitHub issue number. Convention applies across all subcommands; the bare `--issue` flag never means a GitHub issue.

| Today | Lift to | Defends what | Lift cost |
|---|---|---|---|
| `make fetch <N>` | `sdlc fetch --github-issue <N>` | Issue file shape, next-ID assignment, frontmatter | Low — port shell logic |
| `make worktree [name]` | `sdlc start [--issue N \| --name X]` | Single-untracked-issue auto-detect, pre-commit before branch | Low — drops MAKECMDGOALS hack |
| `make pull-request` | `sdlc pr` | Branch ≠ main, links touched issues into PR body | Low — wraps `gh pr create` |
| `make merge` | `sdlc merge [--yes]` | Upstream synced, ahead=0, undone-issue scan, irreversible-action confirm | Medium — longest shell script |
| `make push` | `sdlc push` | Clean tree, pre-merge checks pass, archive done issues. Direct-on-main; `sdlc merge` is the branch counterpart. | Medium |
| `make close-issue ISSUE=N ACTUAL=h VERIFIED='...'` | `sdlc close --issue N --actual h --verified '...'` | Evidence required, milestone plan ticked, atlas touched in window | Medium — port `scripts/close-issue.py` |
| `make issue-sync` | `sdlc lock [--issue N]` | Issue file landed on main as parallelization lock | Low |
| `make check-dry/pure/plan/specs/lessons` | `sdlc judge dry\|pure\|plan\|specs\|lessons` | LLM-judge: principle adherence with fresh-context anti-collusion. Both standalone and auto-dispatched (see below). | Medium — wraps subagent dispatch |

**Judge auto-dispatch.** Judges run consistently because they're invoked *from* checkpoint verbs, not relied upon as a separate human step. `sdlc milestone-close` dispatches `sdlc judge milestone-review`; `sdlc push` and `sdlc merge` run `sdlc judge plan|specs|lessons` as pre-flight. Manual invocation (`sdlc judge dry`) remains available, but the default mode is implicit via the checkpoint that needs the verification. This is the answer to "checks aren't run consistently today" — embed them in the verbs that already get invoked.

### New subcommands (no Make equivalent today)

| Subcommand | Purpose |
|---|---|
| `sdlc state [--json]` | Single source of truth for "where am I": current branch, worktrees, working issues, plan progress. Compaction recovery primitive. Also runs **drift detection**: cross-checks issue status vs. project file task ticks vs. recent commits vs. atlas mentions, flags inconsistencies. The structural mitigation is **funnel-mutations-through-binary** — `sdlc set-status`, `sdlc close`, `sdlc milestone-close` own all status writes, so bypassing them is the only way drift can enter, and `sdlc state` then surfaces it. |
| `sdlc set-status --issue N <status>` | Enforces status transitions: `working` requires `estimate_hours`, `done` delegates to `sdlc close`, refuses `done → open` without log entry. |
| `sdlc milestone-close --issue N --milestone Mx` | Today folded into `close-issue MILESTONE=Mx`; promote to explicit verb with §3 post-milestone code review dispatch as follow-on. |
| `sdlc judge milestone-review --base <sha> --head <sha>` | Codifies AGENTS.md §3 mandatory post-milestone review as a checkpoint, not a prose rule. | 

### Out of scope

- `make refresh` — substrate adoption, belongs to `construct`, not `sdlc`.
- `/construct ...` slash commands — own coherent surface; separate binary candidate later.
- `make build` — generic Go build, stays as Make target.
- Velocity / `active-time-v3.py` — orthogonal binary candidate (`velocity`?), not folded in.
- xx-issues, xx-datatype, xx-fix, xx-introspect, xx-voice-* — prose convention skills, stay markdown. xx-issues SKILL.md becomes a thin pointer to `sdlc fetch --help` for file-shape contract.

### Migration posture

Make targets stay in place during the lift — each Make target gets rewritten to shell out to `sdlc <verb>`. This means existing muscle-memory (`make close-issue ISSUE=...`) keeps working, and downstream repos that vendor `Makefile.workflow` don't break. Once all subcommands ship and bake, Make targets can deprecate.

## Plan

Build order chosen to maximize drift-defense per unit work and front-load the load-bearing patterns (embedded `--help`, fresh-context judge).

- [x] **M1 — scaffold + close.** `cmd/sdlc/` skeleton, subcommand dispatch with `//go:embed`-backed `--help`, `make sdlc-build` + `make sdlc-bootstrap`. Port `scripts/close-issue.py` to `sdlc close`. Top-level `sdlc --help` + `sdlc close --help` carry the skill narrative + checkpoint contract. Keep `make close-issue` as a thin wrapper invoking `sdlc close`.
- [x] **M2 — state.** `sdlc state` + `sdlc state --json`. Reads git + `workshop/issues/` + worktree list. No Make equivalent to migrate.
- [x] **M3 — judge.** `sdlc judge <category>` wraps the subagent-dispatch pattern from `scripts/pre-merge-checks.sh` / `scripts/parallel-checks.sh`. Make `check-*` targets become wrappers. Establishes the anti-collusion primitive explicitly. (Auto-dispatch wiring into `push`/`merge`/`milestone-close` lands in M5/M6 when those verbs ship.)
- [x] **M4 — issue lifecycle.** `sdlc fetch`, `sdlc start`, `sdlc lock`, `sdlc set-status`. Port shell logic from Makefile.workflow.
- [x] **M5 — push/pr/merge verbs.** `sdlc push`, `sdlc pr`, `sdlc merge`. The longest Make scripts; lift mechanically. Wire auto-dispatch of `sdlc judge plan|specs|lessons` as pre-flight in `push` and `merge`.
- [x] **M6 — milestone-close + milestone-review.** Promote milestone close from a flag on `sdlc close` to its own verb; wire up the post-milestone code review dispatch.
- [ ] **M7 — SKILL.md generation + atlas.** Implement `sdlc --index` to emit the SKILL.md content. Regenerate `construct/local/sdlc/SKILL.md` from it (no hand-edits going forward — binary owns the prose). Atlas update: `atlas/workflow/sdlc-binary.md` covering the binary surface (replacing scattered references to individual Make targets).
- [ ] **M8 — deprecate Make wrappers.** Once usage has shifted, simplify or remove the wrapper targets.

Post-milestone code review (per AGENTS.md §3) is **mandatory at each milestone close** — the binary is partly building its own review primitive, so we eat our own dog food early.

## Log







- 2026-05-25: closed M6 — M6 implementation: thin wrapper composing runClose + judge milestone-review dispatch; 71 tests pass; --no-judge dogfood close since claude unreachable in sandbox
- 2026-05-25: closed M5 — M5 review subagent stalled (watchdog 600s); main session self-review verified archive-gating, gh-close differential push vs merge, prompter stdin matches shell — 68 tests pass; full M5 surface smoke-tested
- 2026-05-25: closed M4 — code review of a6f0ece..dc35d6a: Ship (no Critical), 2 Important (I1 regex, I4 file-exists) + 1 Minor (unused arg) addressed; 111 tests pass; live sdlc state shows M1-M4 verbs registered
- 2026-05-25: closed M3 — code review of f17f753..4d59d80: no Critical, 4 of 5 Important addressed in 41ce6f5; 67 tests pass; --dry-run smoke verified prompt + argv for all agents
- 2026-05-25: closed M2 — M2 review of fbc55f5..d7789e0: C1+C2 fixed, I1-I5 addressed (gitx.Capture seam, rune-safe truncate, plan regexes in internal/issue); 49 tests pass; live drift surfacing still works
- 2026-05-25: closed M1 — code review of 9e8625e..fa4010c: 1 Critical (C1 ordering) + 1 Critical-deferred (C2 wrapper) + 5 Important addressed in 90342de; 40 tests pass; bin/sdlc smoke-verified
### 2026-05-25 — session summary

Issue created from discussion captured in `docs/vision/2026-05-25-01-pensive-sdlc-checkpoint-binary.md`. Two threads converged: (1) progressive-disclosure leak in markdown-index skills, fixed by binary-backed disclosure via `<bin> --help` (gmail/charon precedent, but using standard `--help` flag instead of a custom `instructions` verb — convention beats invention); (2) checkpoint guards over FSM modeling — defend known commit moments, don't formalize the journey. The `sdlc` binary unifies ariadne's existing Make-target checkpoint guards under one verb namespace with embedded `--help`, structured state inspection, and anti-collusion judge subcommands. Go chosen to match ../nous's `cmd/nous` precedent. Awaiting user verification of scope before commencing.
