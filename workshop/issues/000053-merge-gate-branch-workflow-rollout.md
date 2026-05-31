---
id: 000053
status: working
deps: []
created: 2026-05-31
updated: 2026-05-31
estimate_hours: 1
---

# Rollout conductor — merge-gate + branch-in-place workflow

**Conductor only.** Sequences three existing tickets + a one-time cut-over so the
program doesn't get lost across sessions. Substantive work + actuals live in the
sub-tickets; this file is the portfolio view and the ordering.

Sub-tickets: **ariadne #52** (generic CI merge-check mechanism), **ariadne #51**
(in-place-branch replaces direct-on-main), **you-decide #4** (the review-gate).

## The four end-results (operator's words)

1. Generic merge-gate mechanism for ariadne + derivatives. → #52
2. Enable the gate for you-decide; gate logic = **two different AI stacks agree on
   fact validation** (`review: passed` AND `reviewed-by ≠ generated-by`). → #4
3. you-decide's current work was done **on main directly**, so it can't use the gate
   — manually close this first (bootstrap) series of commits, ungated. → cut-over
4. Make **branch-in-place** the default (replacing work-on-main) in the ariadne
   workflow. → #51

## Dependency insight

The gate only engages on PRs → you must be on branches → so #51 (branch default) is the
hinge that makes the gate live. And you stop working on main only *after* pushing the
accumulated main work. Hence a single cut-over: **bootstrap era ends with a final manual
push; the branch regime begins.** #51's own implementation is the last bootstrap-era work
(it can't dogfood itself before it exists).

## Sequence

**Phase A — close the bootstrap era (#3):**
- [x] A1. Cross-stack (Codex) review of the unpushed mechanism + gate (#52, you-decide #4 M3). **Done 2026-05-31.** 3 correctness bugs fixed in you-decide (`e9f7f41`: silent-pass on bad range, loose frontmatter parser, escaped shebang). 3 enforcement findings (PR-mutable gate code; `--no-verify`; empty-dir/required-checks) confirm what we already chose — the gate is *advisory* until server-side trusted enforcement; folded into #52 M2 (Phase C). Conclusion: mechanism is sound for advisory use; the "real teeth" are Phase C.
- [x] A2. Resolved you-decide `bootstrap.sh` (committed `5a781f1` — refresh synced it to the current base layer).
- [x] A3. **Pushed 2026-05-31** — you-decide `4488ddb..5a781f1` (+ tag `bootstrap-close`), ariadne `9d27d44..6518c58`. Direct-on-main era closed. (Note: Phase B / #51 is the *one* remaining bootstrap-era exception — it can't be built on a branch that lacks the tooling it adds.)

**Phase B — branch-in-place default (#4 of goals / ariadne #51):** ✅ DONE 2026-05-31 (built + dogfooded live)
- [x] #51 M1–M3: `change-code` default → in-place (`--worktree=yes`/`=ask` overrides); `merge.go` in-place vs worktree topology split (server-side PR merge, then in-place switches back to main); `sdlc push` kept (operator decision); docs (AGENTS.md, helptext, atlas). Tests + vet green.
- [x] **Live end-to-end dogfood (ariadne #54):** `claim` → `change-code --worktree=no` (in-place branch, working tree carried forward) → push.md edit + `TestPushEmbedded` → `pr` (PR #4) → `merge` (PR #4 merged server-side, switched back to main, archived #54, deleted branch). Merge cleanup plumbing now exercised against real git+gh. **Caught a real bug:** the *deployed* downstream `you-decide/bin/sdlc` was a month stale (pre-#51) and died on the in-place merge with `find main worktree: …`. Root cause = downstream prebuilt binaries don't auto-rebuild on base-layer tool changes. Fixed (`make sdlc-build` in you-decide) + documented in `atlas/workflow/sdlc-binary.md` (downstream staleness gotcha).

**Phase D — make you-decide's gate mean "two stacks agree" (#2):** ✅ DONE 2026-05-31 (you-decide #4 M2)
- [x] Cross-stack publish gate: `scripts/cross-stack-gate.sh` + `merge-checks.d/20-cross-stack-gate.sh` assert a *passed* substrate file has `reviewed-by ≠ generated-by` (exact inequality; fail-closed on missing/dup stack fields). Factored shared `lib-substrate.sh` (review-gate refactored onto it, behavior-preserving). Fresh-eyes review (SHIP); `scripts/tests/test-gates.sh` 20/20. `audit-review.sh` same-stack report stays as the dashboard (you-decide #4 M4, deferred).

**Phase C — finish generic mechanism (#1 / #52 M2), optional/when-needed:**
- [ ] `make remote-init` (gh repo create + opt-in required-check). you-decide can stay advisory + merge-on-green.

**Phase E — gate live by construction; resume work on branches:** ✅ DONE 2026-05-31 (you-decide PR #4)
- [x] First real branch→PR (you-decide #4 M2) validated CI end-to-end. **Found + fixed a real #52 mechanism bug:** the CI merge-check failed (exit 127) because `scripts/run-merge-checks.sh` is a sibling symlink into `../ariadne`, absent in the isolated GitHub Actions checkout. Fix: the seeded workflow now runs `BOOTSTRAP_CLONE_ONLY=1 ./bootstrap.sh` to clone the base-layer peer chain as siblings before the checks (operator's call — the single-script mechanism). CI green; both gates run live. **Pending: backport the workflow step to the ariadne #52 seed template** (`ariadne/.github/workflows/merge-check.yml`) so future derivatives seed correctly — see Phase C / #52.

Order of kinds: **A → B → D → (C optional) → E.** Operator chose framing (i): push-now, build-after.

## Current position

→ **Phases A, B, D, E DONE** (2026-05-31). B dogfooded the in-place flow (#54) + caught a stale-downstream-binary bug. D built the cross-stack gate (you-decide #4 M2: `reviewed-by ≠ generated-by`, the literal "two stacks agree"). E validated CI end-to-end on you-decide PR #4 and caught + fixed a #52 mechanism bug (symlinked runner absent in CI → now `bootstrap.sh` clones peers in CI). you-decide PR #4 is green and mergeable. **Remaining:** (1) **#52 seed backport** — apply the CI bootstrap step to `ariadne/.github/workflows/merge-check.yml` so future derivatives seed correctly (existing seeded derivatives using CI need the patch too — seed is write-once); (2) **Phase C** (#52 M2 `make remote-init` + opt-in required-check) — optional, turns the advisory gate into a hard one; (3) the spun-off **ariadne #55** (sdlc single-owner binary + PATH + freshness). All new work is branch→PR (in-place flow proven). Cosmetic: `sdlc pr` double-prefixes a cross-repo `github_issue` in the Fixes line (`Fixes #xianxu/you-decide#3`) — minor, worth a side-quest.
