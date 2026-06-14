---
id: 000096
status: open
deps: []
github_issue:
created: 2026-06-14
updated: 2026-06-14
estimate_hours:
---

# Peer-repo work legibility: make cross-repo base-layer edits visible at the gate

> **Stage:** brainstorm parked mid-design. Captured so we can resume after the
> ariadne base-layer migration (#95 weave). One open question remains (see end of
> ## Spec) — resume there.

## Problem

We co-develop the **base layer (ariadne)** alongside example apps, with multiple
work threads spanning peer repos at once (e.g. brain: kaggle + a workflow thread;
ariadne: weave/colima; parley; pair). The pain:

- **ariadne is a shared mutable singleton.** The `sdlc` shell function hardcodes
  the path — it rebuilds-and-runs `/Users/xianxu/workspace/ariadne/bin/sdlc` from
  *that absolute path* on every call, regardless of which repo you're in. And
  `brain/construct/{adapted,datatype,local,config.json,dev-aliases.sh}` are **live
  symlinks** into `../../ariadne/construct/`. So an in-place edit to ariadne is
  globally live across every peer — and changes the behavior of the workflow tool
  every peer runs — the moment it's saved.
- **Worktrees solve leaf work, not base-layer work.** An app is a leaf — fork it,
  nothing points back, worktrees just work. The base layer is the root everyone
  points at, by *absolute path* (the `sdlc` function) and *relative symlink*
  (`../../ariadne/`). A worktree is invisible to both, so we branch ariadne
  **in-place** — and concurrent threads on one working tree tangle.
- **Concrete failure mode.** A session in brain edits ariadne's files *through the
  symlink* while on a brain branch. brain's git sees nothing (it tracks the
  *link*, not the target). The change lands dirty in **ariadne** with no issue, no
  branch, no claim → unattributed → "unrelated changes show up," and both agent
  and human get confused about who changed what.

## Spec

### Two concerns, two responses

| Concern | What happens | Right response | Status |
|---|---|---|---|
| **Moving ground** | You *read* ariadne's tree live; it shifts under you (even by a legit other thread, even if you never edit it) | **inform** | already built |
| **Write collision** | You *edit* ariadne in-place while another thread does too; in-place edits tangle | **guard** (block / require-ack) | to build |

### What already exists — do NOT rebuild

`sdlc start-plan` already does a **dependency-path contention heads-up** (#82/#83,
`cmd/sdlc/startplan.go`). It walks `construct/deps` transitively (`substrateChain`)
to find the upstream repos this one reads live via symlinks, and for *each*
upstream `gatherBaseContention` reports:

- **branch** — on `main`? → covers **rung 3.1 (branched)**
- **dirty *code*-file count**, with `workshop/`/tracker files **excluded** via
  `assessDirty().Blocking` → the **3.2a vs 3.2b split is already built**
  (dirty-only-in-workshop is deliberately *not* contention)
- other **`status:working`** issues in that repo's tracker → the published-claim
  signal (`sdlc claim` pushes `status:working` to `origin/main`)

It renders one line per upstream (green `clear to plan` / yellow `planning against
a moving base`) and is **non-blocking**. For brain (`construct/deps:
substrate ../ariadne`) this already prints ariadne's contention at plan time.

So three of the four "rungs" already have detection. `change-code` has **zero**
peer-awareness (grep-confirmed). `sdlc state` is read-only, this-repo-only, has
`--json`, no `--all-peers`.

### Decisions made

1. **Posture = guard at the gate** (enforcement). Chosen over *observe-only*
   (passive glance) and over the *relocatable-singleton* route (parameterize the
   `sdlc` function + symlinks so base-layer work runs against its own ariadne
   worktree — rejected: too complex, and manual pair/parley verification wants
   stable paths anyway).

2. **Legibility-by-construction is the core principle.** *Always route
   peer/shared-upstream edits through sdlc — even a one-liner; edits to your own
   working repo are governed by size.* The real axis is **shared-vs-private, not
   size**:
   - **Peer / shared-upstream edit → always sdlc.** The value isn't plan rigor —
     it's *attribution*: you're mutating a tree other threads read live, so you owe
     them a visible marker.
   - **Own repo → size governs.** Small/obvious → skip fine (only thread there,
     attributed by locality); large → sdlc for planning/review value.
   - *Why it works:* the contention signals (branch, `status:working`) are
     **produced by** going through sdlc. Skip sdlc → emit no signal → invisible →
     collision. Routing through sdlc upgrades a *transient* dirty-file blip (which
     vanishes on commit and carries no intent) into a **durable** "thread A owns
     ariadne" marker.

3. **This collapses the design.** The artifact-bearing rungs (3.1, 3.2) become
   *reliable by construction* — no invisible work to detect. Rungs **3.3**
   (clean main + live pair session) and **3.4** (intent-before-artifact, inferred
   from files-read) are demoted to an **optional early-warning add-on** for the
   pre-artifact window (two sessions *chatting* about ariadne before either edits —
   the rule can't cover that, there's no sdlc-able action yet). Bolt-on-later, not
   core.

### The honest catch

`change-code` **cannot catch the diagnosed failure** — *skipping* sdlc — because a
guard inside `change-code` only fires when you *call* it. The chosen guard protects
thread B (who runs the gate) from colliding with thread A's *already-legible* work;
it does nothing to stop thread A's invisible one-liner. So "don't edit peers raw"
must be enforced where `change-code` isn't:

- **(a) a norm** — one line in `AGENTS.md`: "never edit a peer's files in place;
  route through its sdlc." Cheap, soft, relies on the agent honoring it.
- **(b) a pre-edit hook** — fires when an Edit/Write resolves into a sibling repo's
  tree and you haven't claimed there. Deterministic, general (not base-layer
  specific), catches the skip at the moment of violation.

Recommendation: start with **(a) + a fast legible path**; add **(b)** only if skips
actually recur (fix-forward).

### OPEN QUESTION — resume here

**What's the lightest gesture for a one-liner peer change that still emits a
*durable* signal?**
- a **throwaway branch** in ariadne (cheapest — shows as rung 3.1, no issue, no
  ceremony), or
- a **one-touch express claim** (`sdlc claim` auto-creating a tiny issue, so it
  also shows in `Others`)?

This friction knob decides whether "always route peer changes through sdlc" holds
under speed pressure — *speed is exactly why edits got skipped.*

## Done when

- A peer/shared-upstream edit cannot silently happen without emitting a durable,
  peer-visible signal (branch or claim) — enforced by norm+fast-path (and
  optionally a pre-edit hook).
- `change-code` in a shared-upstream repo **blocks / requires-ack** when that repo
  is already hot (branched / dirty-code / another `status:working` issue) — the
  self-contention root case.
- `sdlc state --all-peers` gives a glanceable per-peer contention summary, reusing
  `substrateChain` / `gatherBaseContention`.
- `AGENTS.md` carries the shared-vs-private legibility rule.

## Plan

> Provisional — not yet plan-reviewed. Candidate steps, not committed milestones.

- [ ] Resolve the OPEN QUESTION (fast-legible-path gesture)
- [ ] Encode the shared-vs-private legibility rule in `AGENTS.md` (the norm)
- [ ] Port `gatherBaseContention`/`substrateChain` into `change-code`; make it
      block/require-ack on write-collision (self-contention when run in the edited
      repo)
- [ ] Add `sdlc state --all-peers` (reuse the existing pure functions)
- [ ] (defer) pre-edit hook for raw peer edits — only if skips recur
- [ ] (defer) early-warning add-on: live pair-session (3.3) / intent inference (3.4)

## Log

### 2026-06-14

- Brainstorm held in a brain session (started 2026-06-13). Parked to do the ariadne
  base-layer migration (#95 weave). ariadne is currently on branch `000095-weave`,
  so it's already hot — a live instance of exactly this problem.
- Grounding facts established this session:
  - `sdlc` is a shell function that `go build`s and runs from hardcoded
    `/Users/xianxu/workspace/ariadne/bin/sdlc` — the singleton.
  - `brain/construct/{adapted,datatype,local,config.json,dev-aliases.sh}` are live
    symlinks into `../../ariadne/construct/`; `brain/construct/deps` is
    `substrate ../ariadne`.
  - `cmd/sdlc/startplan.go` already implements the dependency-path contention
    heads-up (`substrateChain`, `gatherBaseContention`, `baseContention`,
    `baseContentionSummary`, `assessDirty`) — non-blocking, walks `construct/deps`.
  - `cmd/sdlc/changecode.go` has no peer-awareness; `sdlc state` is read-only /
    this-repo-only / `--json` / no `--all-peers`.
- Relation: builds on #82/#83 (the contention heads-up). Coordinate with #95
  (weave) — weave is rewriting the base-layer composition mechanism, so the
  symlink/deps substrate this issue reasons about may move.
