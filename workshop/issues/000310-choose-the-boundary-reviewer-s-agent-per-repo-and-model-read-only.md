---
id: 000310
status: working
deps: []
github_issue:
created: 2026-10-09
updated: 2026-10-11
estimate_hours:
card_mirror: '48ea532cecf1488eff351a5f46e2d7a2dbd6bd35' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-11T08:51:05-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: MacBook Pro
    workspace: ariadne:2
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: quick, provenance: inferred, spec: "9403771f", done: "5037d7ec"}
---

# Choose the boundary reviewer's agent per repo, and model read-only

## Problem

The boundary reviewer's agent is chosen by `ResolveAgentCLI` (`cmd/sdlc/internal/judge/agent_resolver.go:34-51`). The order is `--agent`, then `AGENT_CMD`, then `PAIR_AGENT`, then codex/claude env detection, then a claude default. That is effectively "same as the main agent".

#129 (`workshop/history/issues/000129-default-sdlc-judges-to-current-agent.md:24-33`) proposed `same`, `other` (claude reviewed by codex and vice versa) and `explicit` pairs. Only `same` was built. There is no per-repo config.

"Read-only" isn't modeled either (`dispatch.go` `BuildArgs` ~130-155):
- claude gets `--allowedTools` but also `--permission-mode bypassPermissions`, with Bash in the list;
- codex runs `exec --full-auto`;
- gemini runs `--yolo`.

Reviewers have mutated the checkout, e.g. left it detached, so close refused (evidence `a1-15e9a:1453`).

## Spec

- Per-repo config for the reviewer agent: `same` (default; the most robust, since if the main agent works, its CLI works), `other`, or `explicit` pairs. Later possibly per review category.
- A per-agent mapping to its real read-only mode, passed by the dispatcher. No `bypassPermissions`/`--full-auto`/`--yolo` for reviewers.
- Companion to #204, the isolated checkout. Isolation is the agent-agnostic way to make read-only *true*. The operator notes a fresh worktree may be heavy (deps clone plus `weave compile`), so measure the setup cost. Fallback: a detached worktree reusing the slot's built deps.

## Done when

- A repo can configure `other`. A claude-driven close is then reviewed by codex and vice versa, recorded in the ledger.
- A reviewer that attempts a write fails on read-only, not via prose.
- `sdlc close --help` documents the config.

## Plan

Design (2026-10-11, ariadne:2):

- **Config.** `workshop/review.json` (repo-local, like `issue-tracker.json`; not base-layer):
  `{"boundary_reviewer": "same" | "other" | "explicit", "reviewer_pairs": {"claude": "codex", …}}`.
  Absent file = `same`. A malformed file or unknown mode/agent makes the boundary review
  *not run* with a named error (the close halts), never a silent `same`.
- **Resolution** (pure, `judge.ReviewerPolicy`): `--agent`, then `AGENT_CMD` stay operator
  overrides above the policy. Otherwise the *main* agent is detected as today
  (`PAIR_AGENT`, codex env, claude env, claude) and the policy maps it: `same` → itself;
  `other` → claude↔codex, gemini→claude; `explicit` → `reviewer_pairs[main]`, else itself.
  `ResolveAgentCLI` becomes the `same` case of it (ARCH-DRY). Scope: the boundary review
  (close + milestone-close); plan-quality/`sdlc judge` keep `same`.
- **Ledger.** The round already stamps the resolved reviewer (`Agent`); the dispatch line
  names the policy (`via codex (reviewer: other of claude)`). No new ledger field.
- **Read-only.** `DispatchOptions.Fenced bool` + `IsSandbox` auto-approve become one
  `Containment` (none / checkout / fenced) mapped per agent; no auto-approve flag for any
  judge:
  - claude: `--permission-mode dontAsk` (only `--allowedTools` run; Edit/Write are refused),
    never `bypassPermissions`; fenced keeps its own sandbox off (can't nest).
  - codex: none → `--sandbox read-only`; checkout → `--sandbox workspace-write` (cwd is the
    disposable clone); fenced → `danger-full-access` under the OS fence. Never `--full-auto`.
  - gemini: never `--yolo` (headless refuses tools needing approval); fenced → `--sandbox=false`.
  - `--sandbox` on `sdlc judge`/`change-code` is deprecated (hidden, no effect).
- ARCH-FUNERAL: creates nothing durable — one optional repo config file, owned by the operator.

- [x] `judge.ReviewerPolicy` + `ResolveReviewerCLI`, unit tests (same/other/explicit/overrides)
- [x] `workshop/review.json` loader + validation, tests (absent, other, explicit, malformed)
- [x] `Containment` argv mapping replaces Fenced/IsSandbox; argv tests per agent × level
- [x] Wire boundary review: policy from the source root; integration test: claude main + `other` → codex dispatched, ledger round `Agent: codex`, and vice versa
- [x] Help: close.md / milestone-close.md document the config; atlas (sdlc-binary; openshell-sandbox describes the main agents, not judges, so it is unchanged)
- [x] Live check: claude `dontAsk` reviewer asked to Write is refused

## Log

### 2026-10-11
- Implemented on ariadne:2 (quick flow). Dropped the plan's "dispatch line names the policy": the ledger round's `agent` and the `via <agent>` line already say who reviewed; threading the policy label through `boundaryReviewOptionsFor` bought nothing.
- Removed `isSandbox()` (its only consumer was the auto-approve gate); `--sandbox` on `judge`/`change-code` is deprecated, hidden by cobra, no effect.
- Live checks (this slot, inside Claude Code's sandbox): `claude -p --allowedTools Read,Grep,Glob,Bash --permission-mode dontAsk` asked to Write → "Permission to use Write has been denied because Claude Code is running in don't ask mode", no file. `codex exec --sandbox read-only` asked to `echo hi > probe.txt` → `sandbox-exec: sandbox_apply: Operation not permitted`, no file: fails closed, but it also means a codex reviewer can run *no* command where the #204 fence probe fails (the same nesting). Pre-existing (codex exec already defaulted to a read-only seatbelt); now warned at dispatch and documented under REVIEWER in `sdlc close --help`. `other` with a claude main is only useful from an unsandboxed sdlc until the fleet runs sdlc outside the seatbelt.
- side-quest: `make test` failed twice on the hermeticity guard ("leaked .git/sdlc.lock") with no lock left behind. Cause: `processAlive` read `kill(pid,0)` EPERM as dead. Inside the Seatbelt sandbox, a peer slot's live sdlc is not signalable, so the guard (and `repolock.Observe`'s dead-holder reclaim, in production) saw a live holder as dead. Fixed to ESRCH-only death, as `weave/internal/staging` already does; `TestProcessAliveTreatsEPERMAsAlive` (pid 1) fails without it. The `processgroup` TestCancellationKillsDescendants failure is the sandbox denying `/bin/ps` (environmental, not this diff).
- Close review round 1: FIX-THEN-SHIP. BR-1 (Important): the help and atlas claimed every judge runs read-only, but claude's Bash stays unrestricted under `dontAsk`. Reworded `sdlc close --help` REVIEWER, the atlas and the `Containment` doc: Edit/Write are refused, and only the #204 fence stops shell writes. Narrowing Bash to patterns would break reviewers that run tests in the checkout. Minors: the codex nesting warning is now darwin-only (Seatbelt); the Validate wording was changed; the plan row no longer claims an openshell-sandbox atlas edit (that page describes the main agents); `cmd/doc-review`'s claude reviewer moved from bypassPermissions to dontAsk (its allowlist was moot before), pinned in `TestBuildArgs_ReadOnly`.

### 2026-10-09
Filed from the robustness evidence review (ariadne-robustness-1). Revives #129's unbuilt modes; pairs with #204.
