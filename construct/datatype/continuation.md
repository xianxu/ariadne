---
type: continuation
name: continuation
description: "Use when parking or handing off a live coding session — distilling its human-meaningful working state (next action, open deliberations, decisions + dead ends, cross-issue links) into a durable, portable doc so the work can resume later, on another machine, by another person, or under another agent. Triggers on 'park this session', 'create a continuation', 'hand this off', `/xx-datatype continuation`, or pair's park/continue flow. Distinct from `pensive` (a single thought) and from a native session `resume` (machine state, not human understanding)."
---

# continuation

A continuation captures the **human-meaningful projection** of a coding session — the "you are here, here's the next step, here's what's still open and why" that a flattened repo (even one with faithfully-updated issues) can't convey. It exists so work survives a gap in time (a break, a vacation), a change of hands (someone else picks it up), a change of machine, or a change of agent stack.

Its organizing contrast: a native session **`resume`** restores *machine state* (the agent's own transcript + session id); a **continuation** restores *human understanding*. These are different *kinds* of state, not different fidelities — so a continuation is distilled from the session's **rendered** view (what a human actually read on screen) and deliberately drops verbatim tool I/O. The collapsed, human-readable projection is the right substrate, not a lossy fallback.

## Frontmatter shape

| Field | Required | Notes |
|---|---|---|
| `type` | yes | `continuation` |
| `slug` | yes | Short, typeable handle (3–5 words, kebab). Often the pair tag. Used in the filename and to resume (`pair continue <slug>`). |
| `agent` | yes | The agent whose session this distills — the **original** (e.g. `claude`). Pairs with `session_id`. Not necessarily who *produced* the doc (a fresh agent may distill a dead one). |
| `session_id` | no | Native session id of that original session. **Provenance only** — never a resume handle (byte-faithful resume is `resume`'s job). Empty if unknown. |
| `created` | yes | ISO datetime (`YYYY-MM-DDTHH:MM:SS`) the continuation was written. Matches the filename's timestamp prefix. |
| `supersedes` | no | Slug of the prior continuation this one continues from (the chain). Empty if first. |
| `branch` | no | Git branch the work is on. |
| `worktree` | no | Local worktree path — a hint for resuming on *this* machine; not portable. |
| `issues` | yes | Inline list of issue IDs the session touched: `[000071, 000073]`. A session often spans several; this is what makes a continuation session-scoped rather than issue-scoped. |

## Body skeleton

An instance's body, in order:

1. `# Continuation: <slug>` — title.
2. **## NEXT ACTION** — front-loaded. The single concrete next step, specific enough that a fresh agent can start without guessing. The most load-bearing section; never empty.
3. **## State of play** — per-issue status (done / in-flight / blocked). Point at `sdlc state` and the issue files rather than duplicating them. Cross-issue links live here ("#71's decision Y constrains #73").
4. **## Live deliberations** — open threads not yet resolved into artifacts: options under consideration, the current leaning, and *why it isn't decided yet*. What's still in the air.
5. **## Decisions & dead ends** — decisions made + the reasoning, and roads not taken + why rejected. The "tried X, dropped it for Y" that the final repo state erases.
6. **## Pointers** — issues to read first (they are NOT auto-loaded), key files, the branch/worktree. For cross-repo work, **pin the peer repo's path** — a bare slug or `repo#NN` is ambiguous when the next step lives in a sibling repo. (AGENTS.md *is* auto-loaded via `CLAUDE.md`, so don't instruct reading it.)

Skip a section only when genuinely empty (a session with no dead ends may omit that one) — but **NEXT ACTION** and **State of play** are always present.

## Authoring instructions

When the dispatcher applies this prototype:

1. **Substrate — what you distill from:**
   - **Self-mode (the common, recommended park):** the current session. Distill from your own working understanding — it is a superset of what's on screen.
   - **Dead-agent mode:** when parking a session you are *not* in (the original agent is rate-limited / down / wedged but the session hasn't exited), the producer supplies the path to a **rendered transcript** (pair's `pair-scrollback-render --plain` for that tag). Read it as your substrate. Do **not** parse native session stores — the rendered, human-projection view is the correct input; collapsed tool I/O is a feature, not a loss.
2. **Gather the fields:** `slug` (default to the pair tag when available), `agent` + `session_id` (the *original* session's — from pair's `config-<tag>-<agent>.json` when present), `issues` (scan the session / `sdlc state` for touched issue IDs), `branch` / `worktree` (current git state), `supersedes` (the prior continuation for this slug, if any).
3. **Draft → confirm:** draft the body, show the user the NEXT ACTION plus the section outline, get approval (the dispatcher's normal confirm-before-write).
4. **Finalize via the writer — do NOT write the file yourself.** Continuation is the **one datatype that does not** use the dispatcher's default "write the file and leave it uncommitted." Instead, hand the gathered fields + approved body to the **continuation writer** (the producer provides it; in a pair session the park/continue flow invokes it). The writer deterministically renders the frontmatter, allocates the timestamped collision-safe filename, writes the file, and **commits + pushes to `main`** — because a continuation is disaster-recovery and an unpushed one is useless. The dispatcher must not Write the instance itself, and must not apply the leave-uncommitted default.
5. **Default location (explicit):** `workshop/continuation/<YYYYMMDDTHHMMSS>-<slug>.md`. Do **not** run the usual `memory/`/`data/` location discovery — continuation instances live under `workshop/`, and the directory may not exist yet (the writer creates it).

## Search recipes

```sh
# All continuations
rg -l "^type: continuation" workshop/continuation/

# The latest continuation for a slug (the resume target)
ls workshop/continuation/*-<slug>.md | sort | tail -1

# Continuations touching a given issue
rg -l "^type: continuation" workshop/continuation/ | xargs rg -l "^issues:.*000071"

# Continuations distilled from a given agent's session
rg -l "^type: continuation" workshop/continuation/ | xargs rg -l "^agent: claude"

# The next action across every continuation on disk
rg -l "^type: continuation" workshop/continuation/ | xargs rg -A3 "^## NEXT ACTION"
```

## Rules

- **NEXT ACTION is mandatory and concrete.** "Continue the refactor" is not an action; "run `make test-queue`, then fix the `push_front` index bug in `nvim/draft/init.lua`" is.
- **Human understanding, not machine state.** Distill the rendered/understood session; never dump raw tool I/O or transcribe verbatim. If you want byte-faithful pickup, that's `resume`, not a continuation.
- **Append-only chain.** A new continuation for the same work `supersedes:` the prior one; don't edit a shipped continuation to "update" it — write the next link.
- **Auto-committed — unlike every other datatype.** A continuation is committed + pushed on creation (via the writer). It is a recovery artifact; leaving it uncommitted defeats the purpose. This is a deliberate exception to the dispatcher's "never auto-commit data artifacts" default.
- **Resumable, not exhaustive.** Enough for a fresh agent to pick up confidently; not a transcript. If it reads like a log, it's too long.
