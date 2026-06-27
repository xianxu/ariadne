---
id: 000132
status: working
deps: []
github_issue:
created: 2026-06-26
updated: 2026-06-27
estimate_hours: 6.24
started: 2026-06-27T11:24:51-07:00
---

# sdlc repo transaction lock

## Problem

`sdlc` mutating commands can race when multiple agents or tool calls run in the same repo. In pair#72, four concurrent `sdlc issue new` invocations contended on `.git/index.lock`, produced partial issue-sync commits, and raced on `origin/main` pushes.

Git does not provide a general "wait for index lock" semantic for `git add` / `git commit`, and waiting on `.git/index.lock` would still be too narrow: an SDLC transaction includes issue ID allocation, file writes, git add/commit, and push. The lock needs to cover the whole SDLC transaction, not just Git's internal index mutation.

## Spec

Add an SDLC-owned repo transaction lock under `.git`, for example `.git/sdlc.lock`, and take it around local mutating `sdlc` verbs.

Initial covered verbs:

- `sdlc issue new`
- `sdlc issue set-status`
- `sdlc claim`
- `sdlc start-plan` if it mutates issue state
- `sdlc change-code` if it mutates branch/issue state
- `sdlc close`
- `sdlc milestone-close`
- `sdlc push`
- `sdlc merge`
- any other verb that writes tracked files, stages/commits, changes branches, or pushes

Lock behavior:

- Acquire with an atomic operation (`mkdir .git/sdlc.lock` or equivalent), not by polling Git's `.git/index.lock`.
- Store useful metadata inside the lock: pid, command, cwd, hostname, start time.
- Wait by default with a bounded timeout and status messages, e.g. `waiting for sdlc repo lock held by pid 12345: sdlc issue new ...`.
- Detect stale locks where possible: missing pid on the same host, or age beyond a conservative threshold with a clear error/recovery message.
- Release on normal exit and best-effort on interruption.
- Keep read-only commands lock-free.
- Still handle remote push races separately. The local lock serializes one repo checkout, but cannot serialize another checkout or machine.

The lock should prevent invalid interleavings like two `issue new` calls allocating from the same scan window or one close racing another issue-sync commit.

## Done when

- [ ] Mutating `sdlc` verbs acquire a repo-scoped SDLC transaction lock before reading/writing issue state or invoking git mutations.
- [ ] Concurrent `sdlc issue new` invocations in the same checkout serialize instead of failing on `.git/index.lock` or allocating conflicting state.
- [ ] Lock wait/status messages identify the holder command and pid when available.
- [ ] Stale-lock behavior is tested or documented with a safe recovery path.
- [ ] Read-only verbs such as `sdlc issue list/show`, `sdlc state`, `sdlc actual`, and `sdlc judge --dry-run` do not take the lock.
- [ ] Remote push/ref races still produce precise retry guidance.

## Estimate

```estimate
model: estimate-logic-v2
familiarity: 1.0
item: greenfield-go-module design=0.6 impl=1.2
item: smaller-go-module design=0.4 impl=1.4
item: cross-cutting-refactor design=0.7 impl=1.1
item: atlas-docs design=0.1 impl=0.2
design-buffer: 0.30
total: 6.24
```

## Plan

- [ ] Identify every `sdlc` verb that mutates repo state.
- [ ] Implement a small internal repo-lock helper with metadata, wait timeout, and stale detection.
- [ ] Wrap mutating verbs at the transaction boundary, not just individual git commands.
- [ ] Add concurrency tests for at least `issue new` and one close/status mutation path.
- [ ] Update help text to explain the lock and the retry behavior.

## Log

### 2026-06-26

Created from pair#81 / pair#72 retro. The triggering incident was agent-caused parallel `sdlc issue new`, but the robust fix belongs in `sdlc`: serialize local mutating transactions with an SDLC-owned lock in `.git`.

### 2026-06-27

Claimed and entered planning. Durable plan saved at `workshop/plans/000132-sdlc-repo-lock-plan.md`. Design shape: one shared repo-lock helper plus root-level Cobra wrapping so mutating leaves share a transaction boundary (`ARCH-DRY`), pure lock metadata/stale decisions stay separated from filesystem/process IO (`ARCH-PURE`), and every mutating verb named in the issue is covered while read-only verbs stay lock-free (`ARCH-PURPOSE`).

Plan-quality gate returned FAILURE before implementation. Updated the plan to make the lock granularity explicit: review-bearing verbs hold the coarse transaction lock through synchronous judges because judges can dirty tracked files; the default wait timeout is sized for long review/ship transactions, and messages must tell quick commands what is happening. Also added the in-process re-entrancy invariant/test so nested Cobra dispatch cannot self-deadlock.
