---
id: 000091
status: working
deps: []
github_issue:
created: 2026-06-11
updated: 2026-06-11
estimate_hours: 5
---

# Session continuation: datatype + distill skill

## Problem

A live coding session carries human-meaningful working state that the flattened
repo loses — even when working state was faithfully written to issues / `sdlc state`:

- back-and-forth **deliberation** on a topic that never landed in an artifact;
- **cross-issue cross-pollination** — a session often spans several related issues,
  and a decision in one silently informs another;
- **dead ends** and the reasoning behind decisions made — the "tried X, rejected
  for Y" that the final repo state erases;
- the **"you are here / what's the next concrete step"** sense a flattened tree
  can't convey.

There's no durable, portable way to materialize that so work can resume later
(after a break), by another person, on another machine, or under a different
agent. Native agent session stores (`~/.claude/...`, `.antigravitycli/...`, codex)
are machine-faithful but per-agent, locally-scoped, subject to recycling, and the
*wrong kind* of state for this purpose.

## Spec

Organizing principle:

> **`resume` restores machine state. `continue` restores human understanding.**
> Two different *kinds* of state, not two fidelities. A native session store is
> byte-faithful machine state (what `resume` wants). A *continuation* captures the
> human-meaningful projection of the session — what a person needs to pick the
> work back up. The two are orthogonal.

This issue is the agent-agnostic **core**; the pair-side substrate + UX is `pair#50`.

**1. `continuation` datatype** (`construct/datatype/continuation.md`, base-layer →
propagates to ariadne-styled repos). A continuation is **session-scoped** (may
reference several issues), durable (version-controlled, travels via git), and
follows the issue/plan lifecycle (archived to `workshop/history/`). The prototype
file itself is authored in `construct/datatype/type.md`'s six-section skeleton
(lede · frontmatter-shape table · body skeleton · authoring instructions · search
recipes · rules) — it *describes* the instance shape below, it isn't a fill-in template.

File: `workshop/continuation/<YYYYMMDDTHHMMSS>-slug.md` — **timestamp prefix**
(chronological order, no shared counter, no hand-scan; matches the `xx-introspect`
run-id convention), kebab slug (typeable; often the pair tag). Same-second
collisions are resolved by the differing slug; on an exact filename clash the skill
appends a short suffix. **No sdlc coupling** — the skill writes the file directly
(see §2 below on why the `sdlc continuation new` verb was dropped).

Frontmatter:

```
type: continuation
slug: <typeable>
agent: <the agent whose session this distills — the ORIGINAL; pairs with session_id>
session_id: <native id of that original session — provenance only; NOT a resume handle>
created: <ISO>
supersedes: <prior continuation slug, or empty>
branch: <git branch>
worktree: <local path — a hint, not portable>
issues: [NNNNNN, ...]
```

Body sections:

- **## NEXT ACTION** — front-loaded; the single next concrete step. The "you are
  here" the flattened repo can't give.
- **## State of play** — per-issue status; points at `sdlc state`, doesn't
  duplicate it. Cross-issue links live here ("#71's decision Y constrains #73").
- **## Live deliberations** — open threads not yet resolved into artifacts;
  current leaning + why-not-yet.
- **## Decisions & dead ends** — decisions + reasoning; roads not taken + why
  rejected.
- **## Pointers** — issues to read first (NOT auto-loaded), key files,
  branch/worktree. (AGENTS.md is auto-loaded via `CLAUDE.md` → no instruction to
  read it needed.)

**2. Distill skill** (`xx-continuation`, base-layer `xx-*` skill) — the **only**
producer; **no sdlc verb**. (Why dropped: a continuation is a recovery tool, not
an SDLC stage; it needs no shared counter — the timestamp filename suffices — so
threading it through the `sdlc` binary would be coupling for coupling's sake.) The
skill is source-agnostic:

- input: **a rendered session transcript** (plain text) + the repo's `sdlc state` /
  touched issues;
- it drafts the structured doc → **surfaces for the user's approval → writes**
  (operating principle from `xx-introspect`; no silent writes);
- it allocates the timestamped filename and writes `workshop/continuation/…` directly.

Two production contexts, same skill, same output:

- agent alive: distills from its own warm understanding (a superset of the rendered view);
- agent unavailable but session not exited: a fresh agent distills from the
  rendered transcript (pair supplies it — `pair#50`).

The rendered transcript being a *human projection* (collapsed tool I/O,
agent-rendered) is a **feature** — it drops exactly the verbatim cruft a
continuation shouldn't carry. The skill consumes "a rendered transcript"; it does
**not** parse native stores (wrong kind of state, and would need per-agent parsers).

**3. Commit + push on create.** A continuation is disaster-recovery — the whole
point is to externalize agent state to the repo and *off the host*. So on write
the skill **commits the new file to `main` and pushes immediately**. Precedent: a
record artifact going straight to main is exactly what `sdlc issue new` / `sdlc
claim` already do for issue files, via the on-main sync helper (`cmd/sdlc/claim.go`)
— the PR→merge norm governs *feature code*, not tracker records. A continuation is
the same class: a single, append-only, conflict-light record, and an unpushed
recovery doc defeats its own purpose. The skill does the `git add/commit/push`
itself (decoupled from sdlc). (Archival to `history/` still happens at periodic
cleanup, not here.)

## Done when

- `construct/datatype/continuation.md` defines schema/frontmatter/lifecycle, conforms to `type.md`'s six-section prototype skeleton, and is surfaced via the `construct/datatype` base-manifest symlink (automatic) + documented in `atlas/workflow/data-artifacts.md`.
- A produced continuation lands at `workshop/continuation/<timestamp>-slug.md` with conformant frontmatter/sections, and is auto-committed + pushed to `main` on create; covered by a test.
- `xx-continuation` turns a rendered transcript + repo state into a valid continuation doc via draft → approval → write.
- A continuation spanning ≥2 issues round-trips: produced, then a fresh agent reads it and can state the NEXT ACTION without the original session.
- Atlas updated (datatype entry + the `resume`-vs-`continue` principle).

## Plan

- [ ] Define `continuation` datatype (`construct/datatype/continuation.md`) + index link
- [ ] `xx-continuation` writes the timestamped file directly + commit/push-on-create (no sdlc verb) + test
- [ ] `xx-continuation` distill skill (rendered-transcript input → doc; draft → approve → write)
- [ ] Atlas: datatype entry + `resume`-vs-`continue` principle
- [ ] Verify: round-trip a multi-issue continuation through a fresh agent

## Log

### 2026-06-11
