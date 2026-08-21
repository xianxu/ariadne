---
id: 000200
status: open
deps: []
github_issue:
created: 2026-08-21
updated: 2026-08-21
estimate_hours:
---

# sdlc: fleet thread inventory

## Problem

`sdlc state` answers "where am I" for **one** repo. Nothing answers "what work is
open across the fleet" — which is the operator-facing failure that `pair#145`
(couch) exists to fix: threads are forgotten, not mis-ranked, and the forgetting
happens across repos over multiple days.

A cold-revival experiment on 2026-08-20 produced the constraint that makes this
non-trivial. `kbench#24` read `status: working, estimate_hours: 4.98` for a
month while 256 commits of the real work happened elsewhere; git said
`0 ahead, last touched 2026-07-23` and was correct. So an inventory built on
issue frontmatter would hand its caller a **confident lie**, in exactly the case
that matters most.

Design context:
`brain/workshop/pensive/2026-08-20-01-pensive-couch-agent-switcher.md`.

## Spec

**Enumerate on measured facts, never self-declared status.** Branch, last commit
time, ahead/behind, dirty, worktree path are evidence. `status:` is a
self-report — carry it, label it as such, and never let it override measurement.
Same measured-vs-typed discipline the actuals gate already enforces, one level
up.

Scope of a row: repo, thread ref (`repo#id`), branch, worktree path, measured
recency, measured divergence, dirty count, and the self-declared status
alongside them. Plus a one-line human descriptor per thread — the issue title
serves — because couch's fuzzy resolution ("the pair refactor session") needs
something to match against.

**Two staleness signals are distinct** and both belong in the output: git
staleness says the thread has gone cold; mailbox depth (couch's, not sdlc's)
says someone is waiting on it. This issue owns the first only.

**Per-repo concurrency policy** is recorded here as fleet metadata —
`in-place-serial` for repos where the checkout is the installation (pair,
ariadne, parley), `worktree-parallel` otherwise, with a third case for
workspaces carrying heavy local state where worktrees are expensive for
unrelated reasons. It is a stable property of a repo, so it is read
deterministically rather than inferred per spawn.

**Reuse the existing fleet walk** (`project.DiscoverByIssueRef`'s traversal,
#171) rather than adding a second way to enumerate peers.

**JSON first.** couch is the primary consumer, so the machine shape is the
contract and the human rendering derives from it. Pairs with `#199` — this
inventory is a natural candidate for the exposed query set.

**Known gap, deliberately not solved here:** work with no tracker entry is
invisible to an issue-keyed inventory. The rogii-v2 phase that lost a deadline
ran off-spine with no issue file at all — 11 days, 301 commits. Naming it so a
clean result is not mistaken for full coverage; addressing it needs a separate
decision about how untracked work gets named.

## Done when

- One command enumerates open threads across the fleet with measured git facts,
  in JSON, from any working directory.
- A thread whose issue says `working` but whose branch has not moved in weeks is
  reported as cold, with both facts visible and measurement winning.
- Per-repo concurrency policy is readable from the inventory so couch can enforce
  the in-place refusal without inferring it.
- The fleet walk is the existing one, not a second implementation.
- Output carries a human descriptor per thread sufficient for fuzzy resolution.

## Plan

- [ ] Fleet walk + per-repo measured facts, JSON shape.
- [ ] Self-declared vs measured fields distinguished in the schema; drift check
      surfaces disagreement rather than hiding it.
- [ ] Per-repo concurrency policy as recorded fleet metadata.
- [ ] Human rendering derived from the JSON.

## Log

### 2026-08-21

Filed as the inventory half of the couch split — `sdlc` owns what work exists,
couch owns the runtime that brings actors up. Directly motivated by the
2026-08-20 cold-revival experiment, which showed issue frontmatter is the wrong
substrate for enumeration.
