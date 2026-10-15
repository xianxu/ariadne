---
type: type
name: project
description: Use when starting a coordination plane for cross-issue work — a TL sequencing and driving several issues, possibly across products, toward a falsifiable done_when. Triggers on "let's start a project for X", "set up a project file for the launch push", "create a project to track this", "/xx-datatype project". Distinct from product (durable charter) and roadmap (month-level aggregate).
---

# project

A project is the **coordination plane** for work that spans issues: a place to discuss cross-issue problems, sequence the work into trains, and drive execution. A TL runs it (ariadne#329). It cuts across issues, products and repos.

The question a project answers is "what is the right next move across these issues, and who makes it?" It does not answer "will this ship on time?". A project carries **no deadline**. The forecast is derived from the trains and the calibration ledger and is never typed in. When a team needs a date, it sets one against the forecast, and the date never drives status.

Distinct from sibling datatypes:

- `product` — the durable charter ("what is being built"). Static. A project advances one or more products.
- `roadmap` — month-level aggregate ("what we want true by month T"). A project lives inside a roadmap month.

## Layers

| Layer | What | Changes | Where |
|---|---|---|---|
| Charter | problem, cross-issue requirements, `done_when`, scope | rarely, by decision | `## Charter` and frontmatter |
| Plan | trains, cross-train edges, allocation per train | at each grooming | `## Trains`, owned by the TL |
| Decisions | cross-issue decisions only: dated, who decided, one line of why | append-only | `## Decisions`, end of file |
| Live | who is doing what now, what needs the operator | constantly | the TL's status file, outside the datatype |
| Forecast | how much focused work remains on the critical path | derived on read | `sdlc project forecast`, never typed |

**Where a decision goes.**
- **Issue-scoped** decisions (where an issue stops, how it is built, what it measures) go in that issue's `## Revisions`. The worker writes the entry when the TL hands the decision down.
- **Cross-issue** decisions (anything that moves work between issues or constrains several) go in the project: requirements in the Charter, the rest in `## Decisions`.
- **Process** rules (how the TL and workers operate) go in the TL skill, not the project.

## Frontmatter shape

| Field | Required | Notes |
|---|---|---|
| `type` | yes | `project` |
| `name` | yes | Slug form, lowercase-hyphenated. Matches the filename without `.md`. |
| `goal` | yes | One sentence. Why this project exists. |
| `done_when` | yes | The MVP boundary as a falsifiable criterion. *"What would make me say this project is finished?"* |
| `status` | yes | See `construct/vocabulary/project.cue` (the schema authority): `defined` → `executing` → `closing` → `closed`, or `dropped`. |
| `landing_tick` | optional | `true`/`false`: whether `sdlc merge` ticks this project when a cited issue lands (ariadne#373). Unset: yes without `tl:`, no for a TL-owned project. |
| `tl` | optional | The TL driving the project: a slot handle (`xian-xu@ops:0`), or a list of them where each owns trains. A TL-owned project is written only by its TLs. `sdlc close` and `milestone-close` skip it entirely (no tick, no detail block, no review snapshot) and print an info line, so workers report landings to the TL instead (ariadne#326). |
| `operator` | optional | The human the project serves. Default = self. |
| `mvp_scope` | optional | List of issue refs (`[<repo>#<id>, ...]`) in the MVP. The `executing` → `closing` guard requires every one to have landed. |
| `explicitly_out` | optional | List of issue refs deliberately excluded from the MVP. The conversation about what's *out* is more useful than the in-list. Record it here. |
| `created` | yes | ISO date. |
| `updated` | yes | ISO date of the last edit. |
| `sources` | optional | Lineage: files, parley chat IDs and URLs the agent read when authoring. |

`deadline` and `planned_finish` are **banned** in a live record. The schema rejects them, so `sdlc project validate` flags a record still shaped as a shipping plan. Archived records are frozen history and are never validated.

## Lifecycle

The schema authority is `construct/vocabulary/project.cue`; this prose explains
how to author within the model rather than maintaining a second enum.

- `defined` — Charter written (problem, cross-issue requirements, done_when, scope); no train dispatched yet
- `executing` — the first train is dispatched; the TL grooms Trains and appends Decisions
- `closing` — every mvp_scope issue has landed; the done_when audit is the remaining work
- `closed` — done_when audit passed; retro recorded and the file archived
- `dropped` — no longer worth pursuing

Guards:
- `defined` → `executing` needs a trains table with at least one train (`trains-present`).
- `executing` → `closing` needs every `mvp_scope` issue in a terminal status (`mvp-landed`).
- `closing` → `executing` (reopen) is unguarded: new MVP work was filed.
- `closing` → `closed` and a drop from `executing` or `closing` need a dated retro (`retro-recorded`).

Ordinary transitions go through `sdlc project set-status`. Closing and the
post-execution drop go through `sdlc project close`, which owns the retro gate
and the archive move. A close takes `--verified '<evidence>'` (recorded in the
Log) and re-checks that every `mvp_scope` issue landed (`--no-mvp` waives it); `--dry-run` reports the
gates without writing.

## Ownership

- **The TLs own the project file; workers never edit it by hand.** A worker reports its landing to the TL; the row tick is the landing tick below (or the TL's, while that is off).
- A worker's close neither ticks a TL-owned project nor treats an edit to it as staling the worker's review (ariadne#326).
- A project without `tl:` is still ticked by the close gate, as before.
- **At landing** (ariadne#373), `sdlc merge` ticks the landed issue's rows and appends `(PR #N)`, by one commit on the project repository's origin main, never in a checkout. A project without `tl:` is ticked by default. A TL-owned project is ticked only with `landing_tick: true` in its frontmatter; otherwise merge prints the `sdlc project tick` command for the TL. `landing_tick: false` opts any project out.

## Body skeleton

An instance of `project` has, in order:

1. `# <name>` — title matching the slug.
2. **Lede paragraph** — one short paragraph. **Call out the headline omission explicitly** (what's NOT in the MVP).
3. `## Charter` — the problem, the cross-issue requirements (numbered, so decisions can cite them) and the scope narrative.
4. `## Trains` — the trains table and the cross-train edge list (grammar below).
5. `## Breakdown` — the ordered checkbox list of issue rows. A row may carry a short pointer. The detail lives in the issue.
6. `## Log` — append-only retros, forecasts and project-level events. Retro headings use `### <ISO date> — retro`.
7. `## Decisions` — append-only, at the end of the file: `- <ISO date> — <decision> (<who decided>): <one line of why>`.

### Trains

A train is a chain of issues in one code area, run in order by one slot. The TL re-grooms the table at every filing and batch edge.

```markdown
| Train | Cars, in order | Slot |
|---|---|---|
| Main integration | ariadne#183 → #269 → #303 | ariadne:4 |
| Merge and CI | pair#324 ⇒ ariadne#196 → #309 | ariadne:2 (from #196, 10-12) |

- ariadne#183 and ariadne#320 ⇒ ariadne#269/#303 (requirement 6 applies only after 5)
- ariadne#183 ~ ariadne#234: whichever lands second accounts for the other
```

Grammar (parsed by `sdlc project forecast`, `find` and the `trains-present` guard):
- The table header names `Train`, `Cars…` and `Slot` columns.
- In a cars cell, a bare `#N` inherits the repo of the car before it; the first takes the project's repo. Parenthesized text is an annotation and is ignored. Every separator (`→`, `⇒`, `/`, `,`) means "next car": a train runs serially.
- An edge is a bullet holding `⇒`: each ref before it is a prerequisite of each ref after it. Prose after a colon is the reason. `~` bullets are overlap notes for the TL, which the forecast ignores.

### Task line format

Keep one line per task short — **title + ref only**, plus an optional short pointer after landing (`(PR #190)`).

```markdown
- [ ] provider interface skeleton [charon#13 M1]
- [x] OpenAI provider impl [charon#13 M2] (PR #41)
- [ ] write release notes
```

#### Checkbox states

- `[ ]` — open / not started
- `[x]` — done
- `[.]` — blocked (reason in the issue)
- `[-]` — cancelled / removed from scope mid-project

#### Reference syntax

- `[<repo>#<id>]` — issue in any peer repo with a `workshop/issues/` tracker (`charon`, `ariadne`, `metis`, etc.). Brain repos hold no issue trackers (capture/measurement only — the spine refuses there, #176).
- `[<repo>#<id> M<N>]` — milestone-level granularity within an issue.
- Plain text — for items that don't fit any issue tracker (e.g., `write release notes`, `email investors`).

Scope events for a single issue (broaden, demote, punt) go in that issue's `## Revisions`, not in the project. The project's row keeps pointing at the issue.

## Forecast

`sdlc project forecast` (also printed by `project show` and `project status`) reads each car's card (status and `estimate_hours`) and the calibration ledger:
- Each unlanded car is expected to take estimate × Σ(build+converge)/Σestimate, computed over the ledger's split rows. That is the estimate ÷ the calibrated build rate × (1 + converge k) of ariadne#234.
- The range comes from the 25th–75th percentile of the per-issue multipliers.
- Trains run in parallel, and ⇒ edges hold a car until its prerequisite lands.
- The output is the critical path as a range of focused hours, plus the total remaining. It never gives a date.
- An unestimated car takes the median ledger estimate, and the statement names it.

## Authoring instructions

When the dispatcher applies this prototype:

1. **Distill before asking.** Pull the project name, goal and candidate issues from the recent conversation. Pre-fill what's clear; ask only for what's missing.

2. **Required fields the dispatcher must resolve before writing:**
   - `name` (slug) — usually obvious from "project for X". Confirm if X has spaces.
   - `goal` — one sentence. If you can't extract it, ask.
   - `done_when` — the MVP boundary. **Force this conversation upfront**, even if the user resists. Push for a falsifiable criterion ("daily reliance on charon for ≥1 week"), not a vague goal ("charon is good").

3. **Force the MVP scoping conversation.** Ask explicitly: *"What's NOT in this project? What's the smallest version that delivers external value?"* Resolve `mvp_scope` and `explicitly_out` as `[<repo>#<id>, ...]` lists.

4. **Name the TL** (`tl:`) when a TL will drive it. Otherwise leave it out, and the close gate keeps ticking rows.

5. **Write the Charter, then the Trains.** Group issues into trains by code area, one slot each, and write the ⇒ edges between trains. Breakdown rows follow the trains' order.

6. **Residency:** `workshop/projects/<slug>.md` in the project's
   **center-of-gravity repo** — the top product repo by default. This is a soft
   rule, not an address: every ref inside is a qualified `repo#id`, so the record
   reads identically from any vantage, and tooling discovers it fleet-wide
   (`sdlc project find --issue <ref>` / `sdlc resolve --kind project <ref>`
   navigate to it and name the issue's train and car). **Never in brain** —
   brain is capture/measurement and holds no SDLC process artifacts (#171).
   **Moving a record** when the center of gravity shifts: `sdlc migrate`
   relocates it and rewrites refs for the new vantage. Terminal records archive
   to `workshop/history/projects/`.

7. **Updates preserve everything else.** Edit in place and never rewrite the file. Decisions and Log only grow.

8. **Confirm before writing** a new project file: show the destination path, lede line, mvp_scope, explicitly_out and the trains. One round of confirmation.

## Search recipes

```sh
# All projects
rg -l "^type: project"

# Executing projects
rg -l "^type: project" | xargs rg -l "^status: executing"

# Projects involving a specific repo (look for refs in the body)
rg -l "^type: project" | xargs rg -l "\[charon#"

# All open tasks across all active projects
rg "^- \[ \] " workshop/projects/

# Which project, train and car carry an issue
sdlc project find --issue ariadne#329

# Cross-issue decisions across projects
rg -A40 "^## Decisions" workshop/projects/
```

## Rules

- One project per file. Slug, filename, and `name:` field must agree.
- A project always has a falsifiable `done_when`. "Vague goal" is a planning failure; force the conversation at authoring time.
- `mvp_scope` and `explicitly_out` together form the MVP commitment. The `out` list is the load-bearing one.
- No deadline fields. Read the forecast; never type one.
- Cross-issue decisions go in `## Decisions`; issue-scoped ones go in the issue's `## Revisions`.
- A TL-owned project is written only by its TLs.
- A project doesn't replace an issue tracker. Issues describe units of work that exist regardless of timing; a project is the plane coordinating what is currently in flight. The same issue can appear in multiple projects over its lifetime.
