# Single Repo After All

**Date:** 2026-04-12
**Context:** Follow-up to [Phases of Ariadne](2026-04-12-phases-of-ariadne.md). Resolving the state synchronization problem that seemed to threaten the monorepo model.

---

## The Problem

The "repo as world state" principle from the Phases doc assumes everything lives in one filesystem. But when multiple agents/engineers work in parallel, a tension emerges: issue state (progress, logs, traces) changes frequently during execution and needs to be visible to collaborators. Git's commit-push-pull ceremony adds friction proportional to update frequency. This seemed to force a choice: abandon the monorepo, use git submodules/subtrees, or adopt an external system.

## Approaches Considered and Rejected

### Git Submodule / Subtree

Separate `workshop/` into its own repo, mounted inside the monorepo.

- **Submodule:** Separate history and push cadence, but submodules are brittle. Parent pins a specific commit. Everyone needs `git submodule update`. Constant "why is my state stale" bugs.
- **Subtree:** Histories merge back into the parent, so it doesn't actually solve the commit noise problem. Splitting/pushing changes is awkward.

Neither addresses the core issue — they're git plumbing solutions to a problem that isn't about git plumbing.

### External Issue Tracker (Entomb Model)

Use an external system (Linear, GitHub Issues) for coordination, entomb issues into the repo before working on them. Acts as a two-phase commit: claim externally, execute internally.

- Reintroduces a system outside the repo, violating "repo as world state"
- Unnecessary if the repo can handle its own locking

### Two Repos, No Submodule

`ariadne` (code) and `ariadne-workshop` (state), cloned side by side. Push workshop freely, push code deliberately.

- Loses the single-filesystem elegance
- Agents need to know about two locations
- Cross-references between state and code become fragile

## The Resolution: Issue-Level Locking + Worktrees

### The Key Insight

The impedance mismatch was never about git being too slow for high-frequency state. It was about **assuming concurrent writes where issue-level locking prevents them.**

### The Layer Cake Paradox

The Phases doc's roadmap system (projects, quarters, dependencies — all YAML in-repo) works perfectly in git. Code works perfectly in git. So why wouldn't the middle layer (issues, execution state) work?

Analysis of the three layers:
- **Top (roadmap/vision):** slow-moving, human-decided → git works
- **Bottom (code):** changes per-feature, reviewed at PR time → git works
- **Middle (issues/execution state):** changes frequently during active work → git friction?

The middle layer seemed different because of **change frequency × audience size**. But this conflates two things:

1. **Issue state** (assigned, working, blocked, done) — changes at boundary events, infrequent
2. **Execution traces** (agent logs, intermediate thoughts, progress) — high-frequency but only relevant to the owning agent

Once you separate these, issue state is as slow-moving as roadmap items (just finer-grained), and execution traces don't need cross-agent visibility — they're local working state.

### The Model

**Issue-level ownership partitions the write space.** Agent A's traces live under their issue. Agent B's traces live under theirs. They never touch the same files. No concurrent writes, no conflicts.

**Git worktrees provide code isolation.** Each agent works on a feature branch in a worktree. Code changes don't interfere across agents. Merges to main are infrequent boundary events.

### The Synchronization Protocol

```
### 0. Synchronization
- You MUST update issue status to working, commit, and push to origin,
  before commencing work. This is the locking mechanism for parallelized work.
- If push fails, pull and check — if another agent claimed the issue, move on.
```

The sequence:

1. `git checkout main && git pull` — get latest state
2. Update issue status to "working" — claim the lock
3. `git commit && git push origin main` — publish the lock
4. `git worktree add ... feature-branch` — work in isolation
5. Do the work (code + traces, all local to worktree)
6. Merge worktree back to main — release lock with results

**Why this is a true compare-and-swap:** If two agents race on step 3, one will fail to push (non-fast-forward). They pull, see the issue is already claimed, and move on to another issue. Git's push semantics provide atomicity for free.

### Why Traces Don't Cause Problems

Execution traces (agent logs, intermediate state) can be git-tracked without issue because:

- They're partitioned by issue ownership — no two agents write to the same trace files
- They're high-frequency **local** state, not high-frequency **shared** state
- Versioning is a bonus (full history for free), not a coordination burden
- They only reach main when the worktree merges — one atomic event per issue completion

### Scaling Properties

- **Phase 0 (solo):** Trivially works. One person, one machine, no coordination needed.
- **Phase 1 (3 engineers):** Issue claiming is a low-frequency event. Conflict surface is tiny. The protocol handles races gracefully.
- **Phase 2 (10 workers):** Each worker with AI swarms still claims issues atomically. The worktree count grows (dozens of in-flight features), but git handles this fine. The only pressure point is the claim step on main, which remains serializable.

## Conclusion

The monorepo model holds across all three layers and all planned phases. The apparent impedance mismatch was a misdiagnosis — it assumed the middle layer needed real-time shared state, when what it actually needed was **ownership-based write partitioning** with **infrequent boundary synchronization**. Git provides both naturally: issue files for ownership, push/pull for synchronization, worktrees for isolation.

No submodules. No external systems. No sync layers. Just git, used correctly.
