---
id: 000240
status: open
deps: []
github_issue:
created: 2026-09-20
updated: 2026-09-20
estimate_hours:
---

# Issue updates for non-working issues strand on the feature branch — one thread per repo is the wrong model

## Problem

Working #239 on a feature branch, three unrelated issues (#236, #237, #238) had
in-flight edits in the working tree. There was no correct place to put them.

**The asymmetry, measured this session:**

| Verb | From a feature branch | Result |
|---|---|---|
| `sdlc issue new` | **publishes to the trunk** ("Issue changes are on the trunk") | ✅ filing works |
| `sdlc issue sync --issue N` | commits **locally, to the current branch** ("not pushed") | ❌ update strands |

So *filing* an issue from anywhere lands on main, which is right. *Updating* one
lands on whatever branch you happen to be standing on. For the issue you are
working, that is fine — it merges with the work. For any **other** issue it is
wrong twice over:

1. The update is invisible to peer agents until an unrelated feature branch
   merges — the exact thing `claim`/`sync` exist to prevent (AGENTS.md §2:
   "Issue files are workflow state… They need to land on `main` quickly so peer
   agents see status changes without waiting for the feature branch to merge").
2. It couples unrelated content to that branch's fate. If #239 is abandoned or
   rebased, #236–#238's edits go with it.

**The workaround is worse than the gap.** With nowhere to put them, the files sit
dirty across the whole session, and every `git add -A` / `git add <dir>` sweeps
them into an unrelated commit. That happened **twice** in this session; one of
them was `.claude/settings.ariadne.json`, a **fleet-propagated** egress
allowlist, which would have reached every derivative inside a #239 commit,
undeclared. Both were caught and backed out, but only by inspection. See
`workshop/lessons.md` ("Stage by path, never by directory…").

There is also no verb at all for a **non-issue** edit that belongs on main — the
settings file above had to be committed onto the #239 branch with a `side-quest:`
message explaining why, which is a workaround, not a home.

**Secondary friction:** `sdlc issue sync` takes exactly one `--issue N` ("the
commit message names it"), so N dirty issues need N invocations. Reasonable given
the one-commit-per-issue contract, but it compounds the above.

## Spec

### The diagnosis: one thread per repo is the wrong model

The current model assumes a repo has **one** line of work at a time — a branch,
a claimed issue, a working tree. Everything else you might legitimately be doing
has to borrow that thread. But a repo routinely carries at least two independent
conversations:

- a **coding thread** — the claimed issue, its branch, its diff;
- a **product/discussion thread** — filing, refining and triaging issues,
  writing specs, recording decisions, which is trunk-shaped work that should
  never touch a feature branch.

These have different destinations (branch vs trunk), different cadences, and
different collision semantics. Today they share one slot, so the product thread
inherits the coding thread's branch by accident.

### Proposed direction: threads with slots

Give a repo **multiple named thread slots**, each with its own destination and
its own working state, so an update routes by *what it is* rather than by *where
the checkout happens to be standing*. At minimum:

- `coding` — the claimed issue's branch; code + that issue's own files.
- `product` — trunk-targeted; issue bodies, specs, roadmaps, targets for issues
  you are **not** implementing.

Open design questions, not decided here:

- **Mechanism.** A second worktree per slot is the obvious realization and needs
  no new git concepts — but `sdlc issue sync`'s existing commit-tree + CAS push
  to `refs/heads/main` (the `claim` path, #207) already lands on the trunk with
  **no checkout at all**, and may be the cheaper answer: route trunk-shaped
  writes through it regardless of current branch.
- **Routing rule.** Explicit (`--slot product`) or inferred (an issue file whose
  id is not the claimed issue ⇒ product slot)? Inference is friendlier and is
  probably right for the common case, with an explicit override.
- **Scope of the product slot.** Issue bodies clearly. Also `workshop/parley/`,
  `pensive/`, `targets/`, `projects/`? And what about a non-issue base-layer edit
  like `.claude/settings.ariadne.json` — is that a third slot, or does it simply
  have no business landing outside a claimed issue?
- **Collision semantics.** Two slots writing the same issue file; the #222
  lost-update problem one level up.

### Smallest useful first step

Even before slots exist, the asymmetry above is fixable on its own: make
`sdlc issue sync --issue N` route to the **trunk** when N is not the issue
claimed on the current branch, using the same CAS path `claim` already has. That
removes the stranding and the sweep hazard without any new model.

## Done when

- `sdlc issue sync --issue N` for an issue that is NOT the current branch's
  claimed issue lands on the trunk, not the feature branch — verified from a
  feature branch with a dirty unrelated issue file.
- Syncing the issue that IS claimed on the current branch keeps today's
  behaviour (it belongs with the work).
- There is a stated, documented home for a non-issue edit that belongs on main,
  or an explicit decision that no such edit is legitimate mid-issue.
- The thread/slot model is either designed (a durable plan) or explicitly
  deferred with the smallest-useful-step above shipped.
- `atlas/workflow/` records whichever model lands.

## Plan

- [ ] design not yet done — brainstorm the slot model, then decide whether to
      ship the smallest-useful-step first or go straight to slots.

## Log

### 2026-09-20

Filed from the #239 session. Evidence above is all measured in that session, not
hypothesized: the `issue new` vs `issue sync` destination split, and two live
instances of the sweep hazard the gap creates.

Related: **#222** (same-issue concurrent trunk edits are last-writer-wins) is the
collision half of the same area; **#207** built the no-checkout CAS trunk push
this issue would reuse.
