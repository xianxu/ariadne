# Workflow Observations Implementation Plan

> **For agentic workers:** Consult AGENTS.md Section 3 (Subagent Strategy) to determine the appropriate execution approach: use superpowers-subagent-driven-development (if subagents are suitable per AGENTS.md) or superpowers-executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** `sdlc issue show N --json` answers, from any checkout and without asking any agent:
- who owns an issue, and where its work is;
- which checkpoints have passed;
- whether it landed.

Every observation names its source and authority, and distinguishes absent from stale and from unknown (unreadable or failed).

**Architecture:**
- **Shape.** A pure assembler (`internal/observe`) turns raw collected inputs into one versioned `Observation`. Each section is `{state, source, error}` plus its data.
- **Collection.** IO collectors in `cmd/sdlc` read the existing authorities only:
  - the tracker card: status, claimant, completion binding;
  - committed evidence on the issue's branch ref, or on main once landed: the details' flow and plan, gate ledgers, `Review-Verdict` trailers, the archive path;
  - worktrees: which hold the issue branch, plus their dirty and ahead/behind activity.
- **No new state store.** Nothing writes. The only side effect is the tracker fetch, which updates the local remote-tracking ref, as `state` and `issue show` already do.

**Tech Stack:** Go (`cmd/sdlc`, new `internal/observe`). Real-git tests on the #277/#278 harness (`newTrackerRepo`, `closeReady`, linked worktrees, `withClaimant`).

**Operator decisions (2026-10-01, issue Log):**
- **Surface:** `issue show --json`, with the text view showing the same sections.
- **Freshness:** fetch, fall back to stale, carrying the tracker commit and `observed_at`.
- **Targeting:** repo-wide via git worktrees, plus `--repo <path>`.

## Non-goals

- No Couch RPC, no agent round-trip, no multi-machine view. Another machine's worktree is reported as `other-machine`, never probed.
- `state` and `fleet` are unchanged. Folding them onto the observation is a later possibility, not this issue.
- No inference of liveness, abandonment or correctness. Activity is reported as activity.
- No new recorded state. If evidence isn't already written somewhere, the answer is `absent`.

## Core concepts

### The contract (`schema_version: 1`)

```
{
  "schema_version": 1, "issue": "000279", "observed_at": "<RFC3339>",
  "tracker":     {"state": "fresh|stale|absent|unknown", "ref": "<commit oid>", "error"?},
  "card":        {"state", "source": "issue-tracker:<path>", "revision": "<blob oid>", "status", "title", "started"?, "updated"?},
  "assignment":  {"state", "authority": "tracker", "claimant"?: {...}, "relation": "this-workspace|other-workspace|unattributed|unknown",
                  "claimant_worktree": "holds-branch|elsewhere|missing|other-machine|unknown"},
  "workspaces":  {"state", "authority": "worktree", "holding": [{"path", "address"?, "branch", "head", "dirty_count", "ahead", "behind", "is_claimant": bool}]},
  "branch":      {"state", "authority": "committed", "ref", "head", "commits_ahead_of_main", "last_commit_at"},
  "checkpoints": {"state", "authority": "committed", "source": "<ref>",
                  "flow"?: {"kind", "provenance"}, "plan": {"total", "ticked"},
                  "reviews": [{"boundary": "plan|close|M1…", "verdict", "commit", "at", "open_blocking"}]},
  "completion":  {"state", "authority": "tracker", "token", "evidence_commit", "reviewed_head", "landed_commit"?},
  "landing":     {"state", "authority": "tracker", "outcome": "landed|not-landed", "landed_commit"?, "archived"?: "<path on main>"}
}
```

- **`state` is read quality only**, in every section: `present`, `absent` (the authority was read and has no such evidence), `stale` (read from a last-fetched tracker), and `unknown` (the read failed; `error` says why). An `unknown` is never collapsed into `absent`.
  - Every *value* the read yields lives in its own field: `assignment.relation`, `assignment.claimant_worktree`, `landing.outcome`, review `verdict`. So, for example, a stale read of a landed issue is `{state: stale, outcome: landed}`.
  - The same rule holds for any section added later. That is the part of `schema_version 1` that is hard to reverse.
- **Authority classes:**
  - `tracker` covers claim, status, completion and landing.
  - `committed` covers branch evidence: checkpoints and review verdicts.
  - `worktree` covers *activity only*: dirty files and ahead/behind are not progress.

  The help and the atlas say this explicitly: a working card proves a claim, not execution.
- **Strict JSON,** following fleet's convention: unknown and duplicate keys are rejected on decode, invariants are validated on marshal (e.g. `unknown` ⇒ `error`), collections are non-null, and a golden fixture is kept.

### Pure entities

| Name | Lives in | Status |
|------|----------|--------|
| `Observation` + section types, `State` enum | `cmd/sdlc/internal/observe/types.go` | new |
| `Assemble(Inputs) Observation` | `cmd/sdlc/internal/observe/assemble.go` | new |
| `Observation.Validate` / strict `UnmarshalJSON` | `cmd/sdlc/internal/observe/json.go` | new |
| `RenderText(Observation)` | `cmd/sdlc/internal/observe/text.go` | new |

- **Inputs** holds raw collected facts, each with its own read error:
  - the card bytes and blob OID, the tracker ref OID, and stale/error flags;
  - this workspace's identity;
  - the worktree list with per-tree facts;
  - the branch ref and head, and its ahead-count and last commit time;
  - the details bytes at that ref;
  - ledger bytes by boundary and review trailers;
  - the main archive existence.

  `Assemble` derives every judgment (relation, `claimant_worktree`, plan counts, review list, landing). It reuses `issue.CardClaimant` / `MatchClaimant`, `issue.CardCompletion`, `flow.FromFrontmatter`, `issue.CountPlanItems` and `gatestate.Decide`. It is tested table-wise with no IO.
- **DRY rationale:** each judgment is one call into the existing authority's parser. Observations derive; they never restate.

### Integration points

| Name | Lives in | Status | Wraps |
|------|----------|--------|-------|
| `collectObservation(ctx, repoDir, id)` | `cmd/sdlc/observe.go` | new | `loadIssueRecords(PreferFresh)`, tracker snapshot, `gitx.ParseWorktrees`, `git show/log/rev-list`, `claimantIdentity` |
| `issue show --json`, `--repo` | `cmd/sdlc/issue.go` | modified | — |

- **Branch selection:**
  - The details stem (`NNNNNN-slug`) names the issue branch. Use the local `refs/heads/<stem>` if it exists, else the publication remote's `refs/remotes/<remote>/<stem>`. That reaches a parked or no-agent slot, because worktrees share refs.
  - If neither exists and the card is done, read checkpoints from main's archived details and plans.
  - Otherwise checkpoints are `absent`.
- **Reviews: the source is the review artifacts, not commit trailers.** Squash or rebase merges and `branch -D` can leave the evidence commits unreachable, while the artifacts are archived with the issue (#143). The evidence location is therefore:
  - **Before landing:** the issue branch ref's `workshop/plans/`.
  - **After landing:** main's `workshop/history/plans/`, where `archivePlanArtifacts` moved them.

  At that location:
  - **verdict:** each boundary's prose sidecar (`<stem>-close-review.md`, `<stem>-mN-review.md`; latest round) gives `verdict` from its `| verdict |` row, plus its window and timestamp.
  - **plan-quality ledger** (`<stem>-plan-gate.md`): gives the `plan` boundary's `open_blocking`.
  - **boundary ledger** (`<stem>-close-gate.md`): this one is issue-wide. Each boundary's `open_blocking` is `gatestate.DecideScoped` over `FilterBoundary(ledger, boundary)`. The ledger is parsed from bytes read at the ref (`git show`), never from the filesystem. The decision uses the binary's *default* round cap, not `WF_BOUNDARY_ROUND_CAP`, so the observation doesn't depend on the caller's environment. Only `OpenBlocking` is reported; it doesn't depend on the cap, while `Block`/`CapReached` do and are omitted.
  - The `Review-Verdict` trailer on `completion.evidence_commit` is consulted only as a cross-check when that commit is reachable. A disagreement is reported as `error` on the review entry.

  **Absent vs unknown:**
  - **`unknown`:** the card shows a close happened (codecomplete or done, or a milestone ticked) but its artifact can't be found or read. That is evidence that should exist.
  - **`absent`:** an issue that never reached that boundary.
- **`claimant_worktree`:**
  - `holds-branch` when the recorded worktree is a local worktree on the branch;
  - `elsewhere` when it exists on another branch;
  - `missing` when the path is gone;
  - `other-machine` when the fingerprint differs, in which case it is never probed;
  - `unknown` on a read error.
- **`--repo <path>`** resolves the repository containing the path and runs the same collector there.

### Operating envelope (ARCH-CONSTRAINTS)

- **Tracker fetch:** a single fetch through `loadIssueRecords(PreferFresh)`. It is bounded by the existing transaction timeout and process-group termination. On timeout or failure the answer is `stale`/`unknown` with the error, never a hang.
- **Git fan-out is bounded by the issue, not the fleet:**
  - one `git worktree list`;
  - per-worktree facts (status, ahead/behind) only for worktrees holding the issue branch (normally 0 or 1) and the claimant worktree;
  - one `git show` per artifact, from a fixed set of at most 2 + milestones;
  - one log read.
- **Budget:** a typical query is well under 2 s locally, excluding the network fetch. A test asserts the subprocess count stays O(holding worktrees + artifacts), by counting runner calls.

### Lifecycle and ordering

- It writes nothing and creates nothing durable. The tracker fetch's remote-tracking update is the one existing side effect, and it is documented.
- One query reads each authority once. The sources can come from slightly different instants (tracker at fetch, worktrees at scan). That is why every answer carries `observed_at` and the tracker ref: callers compare evidence, not wall-clock assumptions.

## M1 — The contract and its tracker sections

- [ ] `internal/observe`: types, `State`, strict JSON (validate, reject unknown fields), and a golden fixture. Test strategy: a table over every section's state × error invariants, a round-trip, and unknown-key rejection.
- [ ] `Assemble` for tracker, card, assignment (relation, `claimant_worktree`), completion and landing. Test strategy: a fuzz test feeds `Assemble` malformed card/details/ledger bytes and asserts it never panics, and that every section validates (unknown ⇒ error, read quality separate from outcome). One table covers the relation × `claimant_worktree` × tracker-freshness space.
- [ ] Collector for those sections, plus `issue show --json`. The text view appends an "observations" block, leaving the existing output first and unchanged.
- [ ] Real-git tests:
  - an issue claimed in a parked linked worktree (no agent) with uncommitted files, queried from the primary;
  - a stale tracker (remote unreachable) gives `stale` with an error, and never `absent`;
  - the query leaves `git status` and every local branch ref unchanged.
- [ ] M1 milestone-close.

## M2 — Checkpoints, workspaces/activity, other repositories, docs

- [ ] `Assemble` and the collectors for workspaces (activity), branch, and checkpoints (flow, plan, reviews with verdicts and open blocking). Test strategy: real-git issues at each lifecycle point, queried from another checkout:
  - working (no reviews);
  - an M1 milestone closed (M1 verdict, scoped open_blocking);
  - closed and landed by **squash with the branch deleted** (verdicts read from main's archive, `outcome: landed`);
  - codecomplete with its sidecar removed (`unknown`, not `absent`).
- [ ] `--repo <path>`: a test queries a second tracker repository from the first.
- [ ] Contract test cases: a conflicting worktree (the claimant worktree is on another branch while a third worktree holds the issue branch), a deleted claimant worktree, and another machine.
- [ ] Docs:
  - `issue show` help gets the contract, the authority classes, the freshness semantics and "activity is not progress".
  - Atlas: an `issue-tracker.md` Observations section and the verb-table row.
  - `TestEveryFlagAppearsInItsHelp`, and the process manual.
- [ ] Close.

## Revisions

- 2026-10-01 — plan-quality round 1 (PQ-1 to PQ-3 Important, two Minor).
  - **PQ-1:** reviews read from the archived-with-the-issue artifacts
    (branch ref, then main's history/plans), not commit trailers, which
    squash merges and branch deletion can strand. A close evidenced but
    unfound is `unknown`. Added a squash-landing test.
  - **PQ-2:** two ledgers. The plan-quality ledger gives the plan boundary;
    the issue-wide boundary ledger is scoped per boundary with
    `FilterBoundary` + `DecideScoped`, parsed from ref bytes, using the
    default cap, and reporting only `OpenBlocking`.
  - **PQ-3:** `state` is read quality only in every section. Values (landing
    `outcome`, `relation`, `verdict`) get their own fields.
  - **Minors:** fuzz plus table test strategy; an operating envelope (bounded
    fetch, per-issue git fan-out, budget assertion).

