---
id: 000078
status: open
deps: []
github_issue:
created: 2026-06-03
updated: 2026-06-03
estimate_hours:
---

# sdlc merge clean-tree guard refuses on unrelated untracked files

## Problem

`sdlc merge`'s clean-tree guard (`worktreeDirty`, `merge.go:112-126`) calls
`git status --porcelain` and refuses the merge if the output is **non-empty** —
which includes **untracked** files (`??` lines), not just modified/staged tracked
changes. So unrelated, pre-existing **untracked** WIP in the working tree —
local-only work belonging to a *different* effort — blocks an otherwise-ready
merge. There is no bypass flag (`merge.go:103-106` exposes only `--yes`,
`--no-judge`, `--dry-run`, `--issues-dir`, `--history-dir`).

Hit live shipping #58: two untracked dirs (`.claude/skills/xx-pair-doctor`,
`construct/local/pair-doctor/`) — a separate in-progress skill, nothing to do
with #58 — made `sdlc merge` refuse with "Uncommitted changes found." The
workaround was to `git stash -u` them around the merge and pop after; worse, the
stash itself needed the sandbox disabled because git couldn't unlink under
`.claude/skills` (a sandbox write-deny path). A multi-step, sandbox-fighting
detour for files that have nothing to do with the merge.

**Why untracked-only is materially safer than tracked-dirty** for this guard's
purpose: the guard (and its #62 M1 re-assert at step 9b, `merge.go:295`) exists
to avoid merging server-side and then **stranding** the operator on the
post-merge `git switch main` / `git pull`. Modified **tracked** files genuinely
risk that (switch/pull refuse to clobber local edits). **Untracked** files
carry across a branch switch untouched, *unless* `main` introduces a tracked
file at the exact same path — in which case `git switch`/`pull` refuse on their
own (self-protecting), and at worst this is the pre-existing strand the resume
path (#62 M3, `actionResume`) already recovers idempotently. So untracked files
are a near-zero strand risk, yet treated identically to the real risk.

## Spec

Loosen the clean-tree guard so **unrelated untracked files do not block a
merge**, while still refusing on **modified/staged tracked** changes (the real
strand risk). Implementer to weigh (don't over-specify):

- **Option A — ignore untracked in the check.** `git status --porcelain
  --untracked-files=no` in `worktreeDirty`. Simplest; relies on `git switch`'s
  own collision-refusal for the rare same-path case. Applies to BOTH the step-2
  check and the step-9b re-assert (one helper, both call sites).
- **Option B — warn, don't refuse, on untracked-only.** If the *only* dirt is
  untracked, print a `cwarn` listing them and continue; refuse (as today) when
  any tracked modification is present. More explicit/visible than A.
- **Option C — opt-in flag.** Keep today's strict default, add
  `--allow-untracked` (per the #67 `--no-<gate>` acknowledgment convention) so
  skipping is explicit and logged. Safest but adds operator burden for the
  common case.

Leaning B (visible + safe-by-default), but the implementer decides. Whatever the
choice, the step-2 guard and the step-9b re-assert must stay consistent (both go
through the one `worktreeDirty`-style helper — don't let them diverge).

Note the sandbox-unlink interaction is a *symptom* of needing the stash
workaround, not a thing to fix here — once the guard tolerates untracked files,
the stash (and its sandbox fight) disappears.

## Done when

- `sdlc merge` succeeds with unrelated untracked files present (the #58 shape),
  without stashing.
- `sdlc merge` still refuses on modified/staged **tracked** changes (unchanged).
- The step-2 check and the step-9b re-assert behave identically (shared helper).
- A unit test (per `merge_test.go`'s fake-runner pattern) pins both: untracked-
  only → proceeds; tracked-modified → refuses. `go test ./cmd/sdlc/...` green.
- `merge.go` doc comment + `helptext/merge.md` "REFUSES IF" wording updated to
  reflect that untracked-only no longer blocks.

## Plan

- [ ]

## Log

### 2026-06-03
Filed from #58's ship. The clean-tree guard refused on two unrelated untracked
WIP dirs, forcing a `git stash -u` (which itself needed the sandbox disabled to
unlink under `.claude/skills`). Untracked files survive `git switch main` fine,
so they don't carry the strand risk the guard protects against — the guard
over-refuses. Base-layer change to `cmd/sdlc/merge.go` (`base.manifest`), so it
propagates downstream — weigh impact. See [[000058]] for the live incident.
