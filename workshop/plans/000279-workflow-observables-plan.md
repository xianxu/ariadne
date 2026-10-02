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
  "landing":     {"state": "landed|not-landed|unknown", "authority": "tracker", "archived"?: "<path on main>"}
}
```

- **`state` vocabulary** (per section): `present`, `absent` (the authority was read and has no such evidence), `stale` (read from a last-fetched tracker), and `unknown` (the read failed; `error` says why). An `unknown` is never collapsed into `absent`.
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
- **Reviews:**
  - `Review-Verdict` / `Review-Window` trailers come from `git log <main>..<ref>` (or the landed range) on commits whose subject is the issue's (`#N: close`, `#N Mx: milestone close`).
  - Each boundary's `open_blocking` comes from `gatestate.Decide` over its ledger at that ref.
- **`claimant_worktree`:**
  - `holds-branch` when the recorded worktree is a local worktree on the branch;
  - `elsewhere` when it exists on another branch;
  - `missing` when the path is gone;
  - `other-machine` when the fingerprint differs, in which case it is never probed;
  - `unknown` on a read error.
- **`--repo <path>`** resolves the repository containing the path and runs the same collector there.

### Lifecycle and ordering

- It writes nothing and creates nothing durable. The tracker fetch's remote-tracking update is the one existing side effect, and it is documented.
- One query reads each authority once. The sources can come from slightly different instants (tracker at fetch, worktrees at scan). That is why every answer carries `observed_at` and the tracker ref: callers compare evidence, not wall-clock assumptions.

## M1 — The contract and its tracker sections

- [ ] `internal/observe`: types, `State`, strict JSON (validate, reject unknown fields), and a golden fixture. Test strategy: a table over every section's state × error invariants, a round-trip, and unknown-key rejection.
- [ ] `Assemble` for tracker, card, assignment (relation, `claimant_worktree`), completion and landing. Table-tested for: unattributed, mine, foreign, other machine, missing worktree, stale tracker, unknown tracker and a landed card.
- [ ] Collector for those sections, plus `issue show --json`. The text view appends an "observations" block, leaving the existing output first and unchanged.
- [ ] Real-git tests:
  - an issue claimed in a parked linked worktree (no agent) with uncommitted files, queried from the primary;
  - a stale tracker (remote unreachable) gives `stale` with an error, and never `absent`;
  - the query leaves `git status` and every local branch ref unchanged.
- [ ] M1 milestone-close.

## M2 — Checkpoints, workspaces/activity, other repositories, docs

- [ ] `Assemble` and the collectors for workspaces (activity), branch, and checkpoints (flow, plan, reviews with verdicts and open blocking). Tests:
  - a closed-then-landed issue shows `close` SHIP and its evidence, then `landed` with the archive path;
  - a milestone issue lists M1's verdict;
  - a missing branch gives `absent`, and an unreadable ledger gives `unknown`.
- [ ] `--repo <path>`: a test queries a second tracker repository from the first.
- [ ] Contract test cases: a conflicting worktree (the claimant worktree is on another branch while a third worktree holds the issue branch), a deleted claimant worktree, and another machine.
- [ ] Docs:
  - `issue show` help gets the contract, the authority classes, the freshness semantics and "activity is not progress".
  - Atlas: an `issue-tracker.md` Observations section and the verb-table row.
  - `TestEveryFlagAppearsInItsHelp`, and the process manual.
- [ ] Close.

## Revisions
