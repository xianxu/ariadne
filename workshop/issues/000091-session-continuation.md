---
id: 000091
status: working
deps: []
github_issue:
created: 2026-06-11
updated: 2026-06-11
estimate_hours: 2
---

# Session continuation: datatype prototype

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

**2. Production reuses the `xx-datatype` dispatcher — no new skill (ARCH-DRY).**
`xx-datatype` already owns conversation-distillation, location discovery,
prototype-as-spec application, and confirm-before-write; a custom `xx-continuation`
skill would re-implement it. The continuation prototype's *Authoring instructions*
direct the dispatcher to:

- distill the body from the **current session** (self-mode — the common, recommended
  park); or, for the **dead-agent** case, from a **rendered transcript** at a path the
  producer supplies (`pair#50` drives that trigger explicitly);
- gather `slug` / `agent` / `session_id` / `issues` / `branch` / `worktree`;
- **finalize via the deterministic writer (§3)** — never by hand-writing the file.

The dispatcher (LLM) does the irreducible *judgment*: what the NEXT ACTION is, what's
still in deliberation, which decisions/dead-ends matter. The rendered transcript being
a *human projection* (collapsed tool I/O) is a **feature** — it drops the verbatim
cruft a continuation shouldn't carry; the dispatcher never parses native stores. (No
`sdlc continuation new` verb — a continuation needs no shared counter, so it stays out
of the SDLC binary.)

**3. Deterministic writer — the robustness boundary (ARCH-PURE).** The mechanics that
must not depend on the LLM remembering are enforced in code, implemented by the
producer (`pair#50`, e.g. `cmd/pair-continuation`):

- **pure core (unit-tested, no IO):** render complete, conformant frontmatter from the
  gathered fields; allocate the `<YYYYMMDDTHHMMSS>-slug.md` name, collision-safe against
  the existing directory; assemble frontmatter + approved body.
- **thin IO seam (injected fakes in tests):** clock (timestamp), filesystem (write), git
  (add / commit / push to `main`).

This guarantees the disaster-recovery invariants — every continuation has well-formed
frontmatter, a unique timestamped name, and is **committed + pushed the instant it's
written** (an unpushed recovery doc is useless). Precedent for straight-to-main: `sdlc
issue new` / `claim` already push issue files directly (`cmd/sdlc/claim.go`); the
PR→merge norm governs *feature code*, not record artifacts. **Layering:** `#91` defines
the format + invariants (agent-agnostic, base-layer); `pair#50` implements the writer
that enforces them (downstream producer implements the base-layer contract — no backward
dependency). (Archival to `history/` still happens at periodic cleanup, not here.)

## Done when

- `construct/datatype/continuation.md` defines the frontmatter shape, body skeleton, and authoring instructions; conforms to `type.md`'s six-section prototype skeleton; surfaced via the `construct/datatype` base-manifest symlink (automatic) + documented in `atlas/workflow/data-artifacts.md`.
- The prototype's authoring instructions specify the substrate (current session, or a supplied rendered transcript), the field set, and that finalization goes through the writer (not hand-written).
- The disaster-recovery invariants — conformant frontmatter, unique `<timestamp>-slug.md` name, commit+push-on-create — are enforced by the writer (built + tested in `pair#50`), not left to the dispatcher.
- Format validated by a **hand-authored sample fixture** (≥2 issues; exempt from the never-hand-write rule since the writer is `pair#50`): a fresh agent reads it and states the NEXT ACTION without the original session — confirming the format is resumable. Writer-enforced invariants (frontmatter, naming, commit+push) are validated in `pair#50`.
- Atlas updated (datatype entry + the `resume`-vs-`continue` principle).

## Plan

- [x] Author `construct/datatype/continuation.md` (type.md prototype skeleton: lede · frontmatter shape · body skeleton · authoring instructions · search recipes · rules); body skeleton = NEXT ACTION · State of play · Live deliberations · Decisions & dead ends · Pointers
- [x] Authoring instructions: substrate (self-mode vs supplied rendered transcript); field gathering; **explicit default location** `workshop/continuation/<YYYYMMDDTHHMMSS>-slug.md` (the dispatcher's `memory/`/`data/` tree-scan won't find it — Finding 3); **delegate the entire write to the writer binary** — continuation is the one datatype that does NOT use the dispatcher's Write-and-leave-uncommitted default (`xx-datatype` SKILL.md:178 "Never auto-commit"); state that the dispatcher must not write the file itself and the writer owns the commit+push (Finding 2)
- [x] Atlas: `data-artifacts.md` entry (+ built-in-types row) + `resume`-vs-`continue` principle
- [x] Validate via a **hand-authored sample fixture** (≥2 issues, exempt from never-hand-write since the writer is `pair#50`): a fresh agent reads it and states the NEXT ACTION → format is resumable (Finding 1)
> **Writer note:** the deterministic writer that enforces the invariants (frontmatter, naming, commit+push) is built + tested in `pair#50` — tracked there, not a `#91` deliverable.

## Log

### 2026-06-11 — session summary

Designed + landed the continuation datatype core. Key decisions (see Spec): a
continuation restores *human understanding* (vs `resume`'s machine state),
distilled from the **rendered** session; production **reuses the `xx-datatype`
dispatcher** (ARCH-DRY — no new skill); the deterministic writer enforcing the
invariants (timestamped filename, frontmatter, commit+push) is deferred to
`pair#50` (ARCH-PURE, correct layering — base defines the contract, downstream
producer enforces it). Continuation is the **first auto-committing datatype** —
an explicit override of the dispatcher's "never auto-commit" default
(`xx-datatype` SKILL.md:178), spelled out in the prototype.

Shipped: `construct/datatype/continuation.md` + `atlas/workflow/data-artifacts.md`.
Format resumability validated by a fresh-agent round-trip on a hand-authored
fixture — given only the doc, the agent named the correct NEXT ACTION + reading
order; its caveat (pin peer-repo paths for cross-repo work) was folded into the
prototype Pointers. `change-code` plan-quality judge: INFO/pass after fixing 3
findings (dogfood-via-fixture, the auto-commit override, explicit location hint).

Next: `pair#50` — `pair-scrollback-render --plain`, the `cmd/pair-continuation`
writer (TDD: pure core + thin clock/fs/git seam), `pair continue`, Alt+x park-nudge.
