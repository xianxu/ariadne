---
type: continuation
name: continuation
description: "Use when parking or handing off a live coding session — distilling its human-meaningful working state into a durable, portable doc so the work can resume later, on another machine, by another person, or under another agent. Built as the connective narrative over the session's durable artifacts (pensive/issues/targets): it carries the next action, the thread's arc, a model of the user's intention, the open questions to resolve on resume, decisions + dead ends, and lessons learned. Triggers on 'park this session', 'create a continuation', 'hand this off', `/xx-datatype continuation`, or pair's park/continue flow. Distinct from `pensive` (a single thought) and from a native session `resume` (machine state, not human understanding)."
---

# continuation

A continuation captures the **human-meaningful projection** of a coding session — the "you are here, here's the next step, here's what's still open and why" that a flattened repo (even one with faithfully-updated issues) can't convey. It exists so work survives a gap in time (a break, a vacation), a change of hands (someone else picks it up), a change of machine, or a change of agent stack.

Its organizing contrast: a native session **`resume`** restores *machine state* (the agent's own transcript + session id); a **continuation** restores *human understanding*. These are different *kinds* of state, not different fidelities — so a continuation is distilled from the session's **rendered** view (what a human actually read on screen) and deliberately drops verbatim tool I/O. The collapsed, human-readable projection is the right substrate, not a lossy fallback.

A continuation is the **connective narrative over the session's durable artifacts.** The detail lives in the artifacts it points at — the `pensive` notes, the issue files, the `target`s, the code. The continuation's job is to explain *how they connect*, the *history and reasoning* that produced them, where the **user's head is**, and what comes **next**. This is what keeps it short while making it richer: you earn the right to be terse by first **flushing** loose understanding into durable artifacts (see Authoring step 1), then narrating over them rather than restating them.

## Frontmatter shape

| Field | Required | Notes |
|---|---|---|
| `type` | yes | `continuation` |
| `slug` | yes | Short, typeable handle (3–5 words, kebab). Often the pair tag. Used in the filename and to resume (`pair continue <slug>`). |
| `agent` | yes | The agent whose session this distills — the **original** (e.g. `claude`). Pairs with `session_id`. Not necessarily who *produced* the doc (a fresh agent may distill a dead one). |
| `session_id` | no | Native session id of that original session. **Provenance only** — never a resume handle (byte-faithful resume is `resume`'s job). Empty if unknown. |
| `created` | yes | ISO datetime (`YYYY-MM-DDTHH:MM:SS`) the continuation was written. Corresponds to the filename's timestamp prefix (same instant; the filename uses the compact `YYYYMMDDTHHMMSS` form). |
| `supersedes` | no | Slug of the prior continuation this one continues from (the chain). Empty if first. |
| `branch` | no | Git branch the work is on. |
| `worktree` | no | Local worktree path — a hint for resuming on *this* machine; not portable. |
| `issues` | yes | Inline list of issue IDs the session touched: `[000071, 000073]`. A session often spans several; this is what makes a continuation session-scoped rather than issue-scoped. |

## Body skeleton

An instance's body, in order (`*` = always present; `†` = include only when non-empty):

1. `# Continuation: <slug>` — title.
2. **## NEXT ACTION** `*` — front-loaded. The single concrete next step, specific enough that a fresh agent can start without guessing. The most load-bearing section; never empty. **Tie it to the arc/lessons** below — say *why* this is the next step, not just what it is.
3. **## State of play** `*` — per-issue status (done / in-flight / blocked). Point at `sdlc state` and the issue files rather than duplicating them. Cross-issue links live here ("#71's decision Y constrains #73").
4. **## Thread arc & user model** `†` — the session's arc and the model of the user behind it. Where the thread started, how it pivoted across topics, the **underlying connection** among those pivots, and the **latent intention** they reveal → a working model of the user's mental model. Two constraints on that model: it must be (a) **internally self-consistent** and (b) **fit the observed interactions** this session. A few tight paragraphs, not a transcript. (This is the durable **checkpoint** of the user-model discipline that ariadne#103 maintains *live* every turn — keep the two criteria in sync with #103; that issue is the canonical home for the discipline, this section is where a session's instance of it is persisted.)
5. **## Open questions** `†` — internal inconsistencies in that user-model, or ambiguities the session left unresolved. **Lead this section with the resume directive, verbatim:** *"On resume, resolve these open questions with the user before continuing with the NEXT ACTION."* This is how "ask on resume" is delivered — embedded in the doc itself, so any resume path (a fresh `pair continue`, another agent, another machine) honors it without special tooling.
6. **## Artifact map** `*` — the durable artifacts this session produced or leaned on, **with their history, reasoning, and connections** — a narrative, not a bare list. For each load-bearing artifact say *why it exists and what it connects to*: "pensive `<slug>` — written because …, constrains issue #NN, points toward target `<slug>`." Include read-first ordering (issues are NOT auto-loaded), key files, the branch/worktree. For cross-repo work, **pin the peer repo's path** — a bare slug or `repo#NN` is ambiguous when the next step lives in a sibling repo. (AGENTS.md *is* auto-loaded via `CLAUDE.md`, so don't instruct reading it.)
7. **## Live deliberations** `†` — open threads not yet resolved into artifacts: options under consideration, the current leaning, and *why it isn't decided yet*. What's still in the air on the *work* (distinct from Open questions, which is about the *user-model*).
8. **## Decisions & dead ends** `†` — decisions made + the reasoning, and roads not taken + why rejected. The "tried X, dropped it for Y" that the final repo state erases.
9. **## Lessons learned** `†` — transferable lessons from the session: about the codebase, the process, the tooling, or working with this user. Distinct from Decisions & dead ends (which records *this work's* choices); Lessons is the **meta**, the thing worth carrying into the next session regardless of this issue.

Skip a `†` section only when genuinely empty (a session with no dead ends omits that one). **NEXT ACTION** and **State of play** are always present; **Artifact map** is effectively always present too (a session with no artifacts to point at is a sign step 1's flush was skipped).

## Authoring instructions

When the dispatcher applies this prototype:

1. **Flush first — capture key exchanges into durable artifacts.** Before distilling, scan the session for results of key exchanges that aren't yet captured durably, and **flush them to `pensive`** (the low-friction sink for not-yet-structured insight). The continuation is the checkpoint — this is a good time to do it, and the continuation is *built around* these artifacts (steps 4/6/9 narrate over them), so they must exist first.
   - **Automated flush is `pensive`-only.** Do **not** auto-create `target`s — promoting an insight into a target is a human-review-gated act of committing to a goal, not something to mint while parking. `meeting-notes` is not the right fit here either. Issues *already in flight* may be updated as normal work (that's not speculative creation).
   - If the key understanding is already captured in existing artifacts, say so and proceed — don't manufacture pensives for their own sake.
2. **Substrate — what you distill from:**
   - **Self-mode (the common, recommended park):** the current session. Distill from your own working understanding — it is a superset of what's on screen.
   - **Dead-agent mode:** when parking a session you are *not* in (the original agent is rate-limited / down / wedged but the session hasn't exited), the producer supplies the path to a **rendered transcript** (pair's `pair-scrollback-render --plain` for that tag). Read it as your substrate. Do **not** parse native session stores — the rendered, human-projection view is the correct input; collapsed tool I/O is a feature, not a loss.
3. **Gather the fields:** `slug` (default to the pair tag when available), `agent` + `session_id` (the *original* session's — from pair's `config-<tag>-<agent>.json` when present), `issues` (scan the session / `sdlc state` for touched issue IDs), `branch` / `worktree` (current git state), `supersedes` (the prior continuation for this slug, if any).
4. **Draft → confirm:** draft the body, then show the user the **NEXT ACTION**, the **Thread arc & user model** summary, and the **Open questions** (these three are where you're most likely wrong about intent), plus the section outline; get approval (the dispatcher's normal confirm-before-write).
5. **Finalize via the writer — do NOT write the file yourself.** Continuation is the **one datatype that does not** use the dispatcher's default "write the file and leave it uncommitted." Instead, hand the gathered fields + approved body to the **continuation writer** (the producer provides it; in a pair session the park/continue flow invokes it). The writer deterministically renders the frontmatter, allocates the timestamped collision-safe filename, writes the file, and **commits + pushes it** (a path-scoped commit, to origin on the current branch — it reaches `main` when that branch merges) — because a continuation is disaster-recovery and an unpushed one is useless. The writer's only structural guard is that the body contains a `## NEXT ACTION`; the rest of the skeleton is your authoring discipline, not a binary check. The dispatcher must not Write the instance itself, and must not apply the leave-uncommitted default. **If no writer is available** (e.g. a standalone `/xx-datatype continuation` outside a pair session — the writer ships with `pair#50`), do *not* fall back to hand-writing the file: tell the user this datatype is currently finalized only through pair's park/continue flow, and stop.
6. **Default location (explicit):** `workshop/continuation/<YYYYMMDDTHHMMSS>-<slug>.md`. Do **not** run the usual `memory/`/`data/` location discovery — continuation instances live under `workshop/`, and the directory may not exist yet (the writer creates it).

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

# Open questions still owed to the user (resolve these on resume)
rg -A6 "^## Open questions" workshop/continuation/

# Lessons learned across past sessions
rg -A6 "^## Lessons learned" workshop/continuation/
```

## Rules

- **NEXT ACTION is mandatory and concrete.** "Continue the refactor" is not an action; "run `make test-queue`, then fix the `push_front` index bug in `nvim/draft/init.lua`" is.
- **Build around the artifacts, don't restate them.** The continuation narrates the connections among durable artifacts (the Artifact map) and the session's arc/lessons; the detail lives in the artifacts it points at. Flush first (Authoring step 1) so there's something to point at.
- **Human understanding, not machine state.** Distill the rendered/understood session; never dump raw tool I/O or transcribe verbatim. If you want byte-faithful pickup, that's `resume`, not a continuation.
- **The reflective sections are concise.** Thread arc & user model, Live deliberations, Lessons learned — each is a few high-signal lines that *reference* artifacts, not an essay. The flush-first step is what earns the terseness.
- **Open questions get resolved on resume.** When the section is present it leads with the verbatim resume directive, so the next session asks the user before charging into the NEXT ACTION.
- **Append-only chain.** A new continuation for the same work `supersedes:` the prior one; don't edit a shipped continuation to "update" it — write the next link.
- **Auto-committed — unlike every other datatype.** A continuation is committed + pushed on creation (via the writer). It is a recovery artifact; leaving it uncommitted defeats the purpose. This is a deliberate exception to the dispatcher's "never auto-commit data artifacts" default.
- **Resumable, not exhaustive.** Enough for a fresh agent to pick up confidently; not a transcript. If it reads like a log, it's too long.
