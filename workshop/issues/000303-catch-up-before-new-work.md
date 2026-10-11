---
id: 000303
status: working
deps: []
github_issue:
created: 2026-10-08
updated: 2026-10-11
estimate_hours:
card_mirror: 'db1303a6d82bbc2843328cce4cb1ee0382c920da' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-11T08:51:25-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: MacBook Pro
    workspace: ariadne:4
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: full, provenance: inferred}
---

# Catch the slot and its dependencies up to main before new work

## Problem

A slot (`:1+`) starts new work from whatever its checkout and its substrate dependencies were when last touched. After a merge, the slot's resting branch stays behind `main` ("workspace retained on unchanged main-slotN"). The slot's `sdlc` (built from the checkout) and its dependency repos (read live through `construct/deps`) are stale too. Observed after #284 landed: `issue new` on the slot still printed #283-era guidance, because the slot's resting branch, and the binary built from it, hadn't caught up. Today the operator remembers `weave refresh` per slot; nothing enforces it.

Operator idea (2026-10-08): the very start of new work should check that the slot's repo and its dependencies are clean and can fast-forward to their `origin/main`, and catch them up.

## Spec

Decided (2026-10-11):
- **Where:** `start-plan`, on the resting branch, before the issue branch is created. `claim` stays a cheap lock: it keeps its #284 fast-forward of the slot's own resting branch and does no dependency sweep (N fetches plus a compile is not cheap).
- **Probe:** for the slot root and each repo on `substrateChain` (construct/deps, transitive), fetch `origin main` (bounded) and ask whether `HEAD` lacks commits of `origin/main`. A repo with no `origin`, or a fetch that fails, is reported as unchecked: a warning, never "up to date".
- **Catch-up = `weave refresh` (ARCH-DRY):** if anything is behind, start-plan runs the installed `weave refresh` in the slot. Refresh already owns the transaction: it fetches the slot and its substrates, refuses before any branch moves unless every checkout is clean and fast-forwardable, then advances and compiles. sdlc never fast-forwards a dependency itself.
- **Refuse vs warn when refresh fails:** #269's `internal/freshness` decides, so start-plan never refuses what close's integrate gate accepts. A behind repo that is `Stale` (it lacks code that landed on main before the last noon) refuses, and the error carries refresh's output, which names what is dirty or diverged, plus the next action. A repo behind only by today's or docs-only landings warns, and planning continues.
- **The binary:** the running `sdlc` was built before the refresh. After a successful catch-up, start-plan stops before it branches or starts the card, and asks for a re-run so the refreshed build plans the issue. The re-run is idempotent and finds everything current.
- **Boundary with pair#434 (`couch --prepare`, landed):** prepare is the TL's outside-in step before dispatch: it checks the slot is idle, runs the same `weave refresh`, then reloads the agent's context. #303 is the agent's inside-out guard at the start of work, for slots nobody prepared. Both delegate the catch-up to `weave refresh`; only prepare reloads context, only start-plan applies the freshness tolerance. After a prepare, start-plan's probe finds nothing behind and adds one fetch per repo.
- ARCH-FUNERAL: creates nothing durable. The fetch updates remote-tracking refs, which git already owns.

## Done when

- Starting new work in a `:1+` slot whose resting branch or dependencies are behind `origin/main` catches them up (fast-forward) when clean, and otherwise stops with a next action naming what is dirty or diverged.
- A fixture with a stale slot and a stale dependency shows both caught up before the issue branch is created.

## Plan

- [x] Design: where the check lives (claim / start-plan), refuse vs warn for dependencies, reuse of `weave refresh`.
- [x] `catchup.go`: probe slot + substrate chain against origin/main; run `weave refresh` through a seam when behind; freshness decides refuse vs warn; stop for a re-run after a catch-up. Wire it into `startPlanBranch` on the resting branch.
- [x] Tests: a stale slot and a stale dependency are both caught up before the issue branch is created (fake weave); a failed refresh with code-stale input refuses and names refresh's output, and nothing changed; a failed refresh on a docs-only lag warns and plans; a current fleet doesn't run weave; the seam is stubbed in TestMain so no test runs a real weave.
- [x] Atlas + start-plan help text.

## Log

### 2026-10-11
- 2026-10-11: closed — make test: 12 sdlc shards green (1077 tests); processgroup fails only in the sandbox (/bin/ps denied) and passes unsandboxed. Round 1 fixes: targeted catch-up/start-plan/propagate suites green. catchup_test.go: stale slot + stale dependency both caught up before the issue branch (fake weave), re-run runs no weave; stale-code refresh failure refuses naming dep + missing code; post-cutoff and docs-only lags warn and plan; a failed refresh that advanced a checkout stops for re-run; absent/no-origin/fetch-failed checkouts warn unchecked; real weave invoked from PATH in root. Mutation: disabling the start-plan wiring fails the fixture tests.; review verdict: SHIP
- 2026-10-11: flow upgraded quick → full — 199 added lines in code files (limit 100); an earlier round of this close already ran the full review
- Claimed at ariadne:4 (TL dispatch). `weave refresh` already is the catch-up transaction; pair#434's prepare already delegates to it. Design reuses both rather than a parallel fast-forward.

### 2026-10-08
- Implemented `cmd/sdlc/catchup.go` and wired it into `startPlanBranch` on the resting branch. Tests in `catchup_test.go` cover: TestStartPlanCatchesUpSlotAndDependency (the Done-when fixture: both caught up before the branch; the re-run with nothing behind runs no weave), TestStartPlanRefusesAStaleCheckoutRefreshCannotCatchUp, and TestStartPlanToleratesALagFreshnessAccepts (code after the cutoff; docs before it). Mutation check: with the wiring disabled, all three fail. TestMain stubs `weaveRefresh` so no test runs the installed weave.
- Edge noted: shaping under a claim leaves the details dirty on the rest, so `weave refresh` refuses if main moved since the claim. That refusal is only blocking when freshness calls the lag Stale; publish the details (`sdlc issue publish`) and re-run.
- Revision to Done when (2026-10-11): "otherwise stops" is narrowed by the freshness tolerance. A checkout that cannot be caught up stops only when it lacks code main had by the last noon, matching close's integrate gate; a later or docs-only lag warns.
- Close review round 1 (FIX-THEN-SHIP): BR-1 — a failed refresh may already have advanced checkouts. Now HEADs are snapshotted before weave; any move stops for a re-run (TestStartPlanStopsWhenAFailedRefreshAdvancedACheckout). BR-2 — an absent dependency, a repo with no origin and a failed fetch now warn as unchecked; `substrateWalk` returns absent targets alongside `substrateChain`'s present ones (TestCatchUpWarnsUncheckedCheckouts). Minors fixed in round: explicit fetch refspec matching weave's; the production `runInstalledWeave` is tested from PATH, including a missing weave (TestRunInstalledWeaveRunsRefreshInRoot).
