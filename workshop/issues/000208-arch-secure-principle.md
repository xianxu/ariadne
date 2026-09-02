---
id: 000208
status: working
deps: []
github_issue:
created: 2026-09-02
updated: 2026-09-02
estimate_hours:
started: 2026-09-02T15:19:59-07:00
---

# ARCH-SECURE: a security lens in the principle registry

## Problem

The ARCH-* registry has five entries — DRY, PURE, PURPOSE, MOCK, CONSTRAINTS —
and **zero security-ish words across all of them**: no threat, secret,
credential, untrusted, injection, privilege. Every other quality axis assumes
indifferent inputs; nothing in the registry asks what happens when an input is
chosen by someone who wants the system to fail.

The fleet is dev tooling that runs on a host behind firewalls, which removes the
remote attacker and makes the gap easy to live with. But the threat model that
*survives* a firewall is credentials and local state, and that is exactly what
recent work kept hitting. From parley.nvim#205 alone, three security-shaped
defects were caught only incidentally, each filed under a non-security marker:

- a spec run overwrote the operator's **live** cliproxy config with a test API
  key; the running proxy reloaded it and began rejecting their own bearer
  (landed as a test-hygiene lesson)
- opening the agent picker **minted a 0600 `management.key`** as a side effect of
  gathering host/port (landed as ARCH-PURPOSE)
- a persisted `catalog.json`, readable across sessions and versions, crashed the
  picker on a malformed row (landed as a crash bug)

None was found by a security lens, because there is none. Each was found by a
reviewer noticing something adjacent.

There is a second, higher-level concern — authority gradients and blast radius —
which is deliberately NOT in scope here; see Spec.

## Spec

Add **one** gated principle, and write down a second **without** gating it.

### ARCH-SECURE (gated) — implementation level

Goes in `cmd/sdlc/internal/judge/architecture.md`, so it is embedded into
start-plan, plan-quality and boundary-review like every other entry.

```markdown
## ARCH-SECURE — Name what is trusted

- **principle:** Every component has inputs it did not produce and secrets it
  must not leak. Trust is a property of provenance, not of location: an input
  crossing a process, session, or version boundary is untrusted even when this
  same program wrote it, and a credential is a credential whether it reaches
  disk, a log line, a subprocess argument, or a test fixture. Prefer making
  invalid state unrepresentable — parse input into a typed value at the boundary
  — over checking for bad values downstream, and prefer structural separation
  (parameterised queries, argv arrays, escaped contexts) over sanitising strings.
- **at-plan:** Name the boundaries this design crosses and what it assumes on
  each side. For every persisted artifact or external response it reads, say what
  happens when that input is truncated, hand-edited, or written by an older
  version — and where it is turned into a typed value. For every credential it
  touches, say where it is created, who can read it, and whether anything creates
  it as a side effect of an unrelated action. Flag a design that widens blast
  radius for convenience: a test that can reach real user or system state, a
  component that mints a secret on a path whose purpose is something else, an
  artifact trusted because this program produced it. Mark `N/A` when the work
  touches neither untrusted input nor secrets; do not fill a ceremonial checklist.
- **at-review:** Flag input from outside the process treated as well-formed
  (persisted caches, external API bodies, files under a shared directory);
  credentials reaching a log, an error message, an argument list, or a fixture;
  tests able to write real user or system state; and secrets created as a side
  effect of an unrelated call. Where the diff parses untrusted data, check the
  failure path degrades visibly rather than crashing or substituting a fabricated
  value that downstream code will read as evidence.
```

Every clause is drawn from an incident rather than from a generic checklist, so
the lens catches this fleet's failures rather than web-app ones.

### ARCH-AUTHORITY (NOT gated) — architecture level

Goes in a **new** `cmd/sdlc/internal/judge/architecture-deferred.md`, which is
NOT embedded.

```markdown
# Deferred architecture principles

Written down, deliberately NOT embedded into the gate prompts. These describe
system-level properties that need an authority topology worth reasoning about;
gating them before that exists produces ceremony, not design. Move a section into
architecture.md to activate it — that is the whole activation step.

## ARCH-AUTHORITY — Keep authority proportional to the instruction's source

- **principle:** A component acting on instructions it did not originate must not
  wield more authority than whoever supplied them. Where an authority gradient
  exists — a process running with the operator's rights, acting on content chosen
  by someone else — the risk is not bad input but a confused deputy: the
  component is asked to do something legitimate-looking on behalf of a party who
  should not be able to ask. Scope authority to a deliberate act rather than
  making it ambient, and design segmentation so that compromising one part yields
  a bounded amount of the whole.
- **at-plan:** Identify each authority the system holds (filesystem, network,
  credentials, subprocess execution) and, for each, who can cause it to be
  exercised. Flag a design that lets attacker-chosen content reach a path holding
  authority the content's author should not have. State what a compromise of each
  component would reach.
- **at-review:** Flag a privileged operation triggerable by unprivileged input,
  ambient authority that no deliberate act gated, and a component whose failure
  would grant more than its own scope.

**Why deferred (2026-09-02):** the architecture-level security concern in this
fleet is already identified and tracked — parley.nvim#129, "capability-based tool
permission model", which names the confused-deputy risk explicitly and settles on
"knowledge is free; power requires a human act." A registry entry would restate a
decision that already has an owner and a plan. Activate this when a second
instance of the pattern appears somewhere #129 does not cover.
```

### Why a separate file, not a status field

`architecture.md` is embedded **whole-file** (`//go:embed`) and
`ArchitectureMarkers()` regex-scans that same text, so *anything* carrying an
`ARCH-` marker in that file is counted and injected — the block header even says
"work through each of the **N** entries", with N derived from the scan. A
deferred entry inside that file would therefore be gated, which is the opposite
of the intent. A `gates:` field would require entry-level parsing and filtering
in `judge/` — a real feature for one deferred entry. A second, non-embedded file
costs nothing and makes activation a cut-and-paste.

### Naming

One token per marker, matching DRY/PURE/PURPOSE/MOCK/CONSTRAINTS. An earlier
draft used `ARCH-SECURE-CODE` / `ARCH-SECURE-ARCH`; those share the `ARCH-SECURE`
prefix, so any grep by family matches both — rejected for that reason.

## Done when

- `sdlc arch-principles` lists ARCH-SECURE under both lenses, and the block
  header's entry count reads 6.
- `sdlc start-plan` delivers it; the plan-quality and boundary-review prompts
  carry it (they embed the same registry, so this follows — assert it).
- `architecture-deferred.md` exists with ARCH-AUTHORITY and is **not** reachable
  from any prompt: a test asserts no gate output contains `ARCH-AUTHORITY`.
- `cmd/sdlc/archprinciples_test.go` marker list includes ARCH-SECURE.
- AGENTS.md's "Core Design Principles" narrative does not drift from the registry
  (the existing drift guard covers this — confirm it sees the new marker).

## Plan

- [ ] M1 — add ARCH-SECURE to `architecture.md`; extend the marker list in
      `archprinciples_test.go`; assert the rendered block counts 6 entries and
      carries the marker under both lenses.
- [ ] M2 — add `architecture-deferred.md` with ARCH-AUTHORITY, plus a test
      asserting the deferred marker reaches NO gate output (the guard that makes
      "documented but not gated" a checked property rather than a convention).
- [ ] M3 — atlas: `atlas/workflow/architecture-principles.md` gains ARCH-SECURE
      and a note on the deferred file and how to activate an entry.

## Log

### 2026-09-02

- Opened from a parley.nvim session reviewing #205's aftermath. The measurement
  that prompted it: across 5 repos over 14 days, 4,775 ARCH citations landed in
  committed artifacts, but **71% were in generated gate ledgers and review
  sidecars** (the machinery citing itself), 25% in plans/issues, and only **142
  in code comments** — so the registry is currently more review vocabulary than
  design input. Adding a lens that fires on a class the reviews were catching
  only incidentally is the cheapest way to raise that ratio.
- Cross-ref: parley.nvim#129 owns the architecture-side (capability model,
  confused deputy). ARCH-AUTHORITY is deliberately deferred to avoid restating a
  decision that already has an owner.
