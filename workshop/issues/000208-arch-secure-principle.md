---
id: 000208
status: codecomplete
deps: []
github_issue:
created: 2026-09-02
updated: 2026-09-02
estimate_hours: 1.13
started: 2026-09-02T15:19:59-07:00
actual_hours: 1.72
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

The split is the point. The implementation-level lens — untrusted input,
credentials, blast radius — fires on defects this fleet actually produced, so it
earns a slot in every gate prompt. The architecture-level one describes a
property that needs an authority topology worth reasoning about, and gating it
before that exists buys ceremony rather than design; it is written down where it
can be activated by moving a section, and a test keeps it out of the prompts
until someone does. Both texts are given below verbatim, because the registry is
embedded whole-file and its prose *is* the deliverable.

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

### The marker-enumeration class

Adding a sixth entry is **not** a one-line paste. `ArchitectureMarkers()` is a
genuine single source at runtime, but five sites restate its contents by hand:

| site | on a 6th entry |
| --- | --- |
| `judge_test.go:120` `TestArchitectureMarkers` | hard-fails (`len != len` → `t.Fatalf`) |
| `judge_test.go:332` `TestArchitectureRegistry_Content` | silently stops covering the new marker |
| `judge_test.go:364` `{{ARCH_STAR}}` full-set literal | passes only if the entry is appended AFTER `ARCH-CONSTRAINTS` |
| `cmd/sdlc/archprinciples_test.go:20` | silently stops covering the new marker |
| four `testdata/golden/*.prompt` | byte-drift + "each of the 5 entries" → `-update-golden` |

Patching five lists is the instance. The rule is: **exactly one site pins the
marker set by hand — the tripwire that makes an accidental registry edit
visible — and every other site derives from `ArchitectureMarkers()`.**
`TestArchitectureMarkers` is that one site, because pinning the expected set is
what it is *for*. The other three tests build their expectations from the
function, so entry number seven touches the registry, the tripwire, and the
goldens, and nothing else.

The goldens legitimately carry registry bytes; regenerating them with
`-update-golden` is the blessed path for a registry edit
(`atlas/workflow/architecture-principles.md:16-20`), not a workaround.

### Operating envelope (`ARCH-CONSTRAINTS`)

The registry is embedded verbatim into four gate prompts plus `start-plan`
output. A sixth entry costs roughly **+1.4 KB per dispatch** and one more
mandatory item in the block header's "work through each of the N entries".
That is accepted deliberately: the entry's clauses each trace to a defect this
fleet actually shipped, and the alternative — a lens nobody is prompted with — is
what the issue exists to end. The deferred file costs zero, which is the reason
ARCH-AUTHORITY goes there rather than being trimmed into the registry.

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
  from any prompt. The guard DERIVES both halves: the forbidden set by scanning
  `architecture-deferred.md` with the existing `archMarkerRE`, and the gate-facing
  text by walking `promptFS` plus `ArchitectureBlock` (both lenses),
  `CodeReviewBody` and `ArchitectureRegistry`. Moving a section into
  `architecture.md` must therefore *empty* the forbidden set and leave the guard
  green — activation is the move and nothing else, exactly as the file claims.
- The guard classifies WHY its forbidden set is empty rather than assuming:
  sections but no parsed marker means the headings broke (fail); nothing at all
  means every entry was activated (skip). Renaming or deleting the file is
  already a build error via `//go:embed`. The `retired` state must **skip, never
  fail** — failing it would red the suite on activation and break the bullet
  above, which is the tension an earlier draft of this bullet got wrong.
- Only ONE site hardcodes the marker set (`TestArchitectureMarkers`); the other
  three test sites derive theirs from `ArchitectureMarkers()`. Adding a seventh
  entry touches the registry, that one tripwire, and the goldens — nothing else.
- The four `testdata/golden/*.prompt` files are regenerated and their entry count
  reads 6.
- AGENTS.md's "Core Design Principles" narrative is **marker-agnostic by design**
  (#128 deliberately removed the per-marker enumeration from the constitution),
  so `TestArchitecture_NarrativeRoutesToArchPrinciples` needs no change — asserted
  by leaving it untouched and green, not by adding the restatement #128 deleted.
- `atlas/index.md`'s Architecture Principles hook does not enumerate a coverage
  list that goes stale on the next entry.

## Estimate

```estimate
model: estimate-logic-v3.1
familiarity: 1.0
design-buffer: 0.15
item: atlas-docs             design=0.05 impl=0.08
item: smaller-go-module      design=0.05 impl=0.20
item: cross-cutting-refactor design=0.05 impl=0.16
item: atlas-docs             design=0.05 impl=0.06
item: milestone-review       design=0.00 impl=0.20
item: milestone-review       design=0.00 impl=0.20
total: 1.13
```

In Plan order:

1. `atlas-docs` — the two prose artifacts: ARCH-SECURE into `architecture.md`
   and the new `architecture-deferred.md`. Both texts are written verbatim in
   the Spec, so this is transcription, not authoring.
2. `smaller-go-module` — the deferred-marker guard. The only genuine design
   here: derive the forbidden set from `architecture-deferred.md` via
   `archMarkerRE`, derive the gate-facing text by walking `promptFS` plus the
   three render helpers, and assert the deferred file parses ≥1 marker so a
   rename can't make it vacuously pass. Lives in package `judge` for access to
   the unexported embed.
3. `cross-cutting-refactor` — collapse four marker-restating test sites onto
   `ArchitectureMarkers()`, keep `TestArchitectureMarkers` as the one tripwire,
   regenerate four goldens with `-update-golden`.
4. `atlas-docs` — the workflow page entry, the activation note, and
   de-enumerating `atlas/index.md`'s hook.
5. `milestone-review` — one close boundary (plain checkboxes, one review).

Design is `×0.2` spec-quality discounted: the Spec carries both registry texts
verbatim, names all five marker sites with file:line, and settles the guard's
derivation strategy — the remaining design is reading. Buffer `+15%` on that
basis (v3.1 step 4). Familiarity `1.0`: Go, this package, embed + regex, all
routine here.

**Review is priced as TWO rows at the scaled ceiling**, not one. The first cut
put a single `milestone-review` at the ceiling and argued from ariadne#206 —
which budgeted 0.15 for one boundary, spent six rework rounds, and closed 1.31
against 4.43 actual. The estimate-quality judge pointed out the obvious flaw:
moving 0.15 → 0.20 recovers 33% of a 340% miss, and the primitive is per-*chunk*,
so the honest lever for "I expect N rounds" is N rows. Predicting multiple rounds
in prose and then pricing one is the same error in a different place. Two rounds
is the read here: the same shape as #206 over a smaller class (one registry, one
guard, five marker sites, four goldens), and plan-quality already took two.

**Convention this block assumes:** the actual will include pre-implementation
design time. The claim landed at 15:19 but the window opens earlier, so issue
authoring and both plan-quality rounds are inside it. No `issue-spec` row is
budgeted against that, because the design was authored by another session before
this one picked the issue up — so at close this will read under, for a reason
that is not implementation sizing. (ariadne#205, which added ARCH-CONSTRAINTS,
assumed the opposite: it carried `item: issue-spec design=1.0` and closed at 0.44
because its window did *not* capture the spec.) Recorded so the ledger row is
read correctly rather than treated as a sizing signal.

*Produced via `brain/data/life/42shots/velocity/estimate-logic-v3.1.md` against
`baseline-v3.1.md`. Method A only.*

## Plan

- [x] Add ARCH-SECURE to `architecture.md`; extend the marker list in
      `archprinciples_test.go`; assert the rendered block counts 6 entries and
      carries the marker under both lenses.
- [x] Add `architecture-deferred.md` with ARCH-AUTHORITY.
- [x] Guard "documented but not gated" as a checked property: a test in package
      `judge` that DERIVES the set of gate-facing text — every `prompts/*.md`
      walked out of `promptFS` and rendered, plus `ArchitectureBlock` under both
      lenses, `CodeReviewBody`, and `ArchitectureRegistry` itself — and asserts
      none contains `ARCH-AUTHORITY`. Enumerating gate outputs by hand is the
      restatement that goes stale; a prompt added later must be covered without
      anyone remembering this issue.
- [x] Collapse the marker-restating sites: `TestArchitectureRegistry_Content`,
      the `{{ARCH_STAR}}` golden expectation, and `archprinciples_test.go` all
      derive from `ArchitectureMarkers()`; `TestArchitectureMarkers` stays the
      single hand-written tripwire. Regenerate the four goldens.
- [x] Atlas: `atlas/workflow/architecture-principles.md` gains ARCH-SECURE and a
      note on the deferred file + how to activate an entry (move the section);
      `atlas/index.md`'s hook stops enumerating a coverage list.

## Revisions

### 2026-09-02 — one review boundary, not three

The Plan arrived tagged `M1`/`M2`/`M3`. In this workflow an `Mx` tag is a
**review boundary**, not a task label (AGENTS.md §3): each one commits to its own
`sdlc milestone-close` with a full fresh-context LLM review and a
`Review-Verdict:` trailer. This is ~3 files and well under 100 lines of
production code — a paste into a registry, one new markdown file, one test, one
atlas edit — so three boundaries would buy three reviews of a change that has one
coherent story, plus the redundant milestone-close/issue-close double-log §3
warns about.

Retagged as plain checkboxes: one close, one review. Fresh evidence for the
sizing: ariadne#206 just spent **six** boundary review rounds on one close, and
the cost tracked the number of members in the class under review, not the number
of milestones. Splitting a small change into more boundaries multiplies the
reviews without adding classes to check.

The step list also gained a row: the deferred-marker guard was folded inside M2's
one line, and it is the only part of this issue that is genuinely design rather
than transcription — it deserves its own row and its own reasoning.

### 2026-09-02 — "a simple paste" was wrong

Plan-quality round 1 refuted the framing this issue was picked up under. Adding a
sixth entry looked like one edit because `ArchitectureMarkers()` is a real
single source *at runtime* — the block header count, `{{ARCH_STAR}}`,
`start-plan`, and all four gate prompts genuinely derive. What does not derive is
the **test** layer: five sites restate the marker set by hand, one of them a hard
`t.Fatalf` on length, one of them order-sensitive in a way nothing states.

The response is the rule rather than five patches: one hand-written tripwire,
everything else derived. That is the same shape ariadne#206 converged on after
six rounds, arrived at here in one — which is the gate working as intended.

Two smaller corrections in the same round: the deferred-marker guard must derive
its forbidden set from `architecture-deferred.md` (hardcoding `ARCH-AUTHORITY`
would make the file's own "moving a section is the whole activation step" false
— activation would red the guard), and the Done-when's claim about the AGENTS.md
drift guard was backwards: that guard is marker-agnostic *by design* since #128,
and "confirm it sees the new marker" invited re-adding the restatement #128
deliberately removed.

## Log

### 2026-09-02
- 2026-09-02: closed — Round-2 BR-9 fixed and bundled per #174. go test ./cmd/... ./pkg/... green except the pre-existing unrelated fleet_plan test (#210). TestDeferredFileIsGuarding deleted: it asserted the committed file is always `guard`, which would red the suite when the last deferred entry is activated — contradicting the activation contract — while its other branch duplicated the deferredBroken fatal. The verdict->action mapping is now single-sourced on deferredVerdict and the six describing sites point at it. Verified by probe against the real tree: moving ARCH-AUTHORITY into architecture.md leaves the deferred guard GREEN and fails exactly TestArchitectureMarkers (the deliberate hand-written tripwire) and TestBuildPrompt_Golden (re-capture) — the two the atlas documents as what activation touches. Breaking a deferred heading still goes red on the vacuity branch; leaking the marker into prompts/dry.md still goes red naming that prompt. Orphaned godoc comment folded; atlas page gained markersIn and the verdict table.; review verdict: FIX-THEN-SHIP

**Close review round 2: BR-9.** My BR-1 fix contradicted the activation
contract. `TestDeferredFileIsGuarding` asserted the committed file is always in
the `guard` state — so activating the *last* deferred entry would red the suite,
which is precisely the "activation is a pure move" promise the whole design
exists to keep. Its other branch duplicated the `broken` fatal already in the
guard. Fixing one property by breaking another, again.

Deleted, and the verdict → action mapping now lives in exactly one place
(`deferredVerdict`'s doc), with the six sites that described it pointing there
instead of restating it. `retired` skips and never fails, and the reason is
written where someone would otherwise "fix" it.

Re-probed after the change: activating ARCH-AUTHORITY leaves the deferred guard
green and fails exactly two tests — `TestArchitectureMarkers` (the deliberate
tripwire) and `TestBuildPrompt_Golden` (re-capture) — which are the two the atlas
already documents as what activation touches. That is the contract holding.

Also: the doc comment orphaned by deleting `deferredMarkers` had fused onto
`deferredState`, so godoc rendered the wrong doc for the type; folded into the
real one. `atlas/workflow/architecture-principles.md` gained `markersIn` and the
verdict table.


**Close review round 1: FIX-THEN-SHIP**, three Important findings, all fixed
before the close commit.

- BR-1 — the vacuity classification had no test, and neither empty branch was
  reachable in the committed tree. I had *probed* those states by mutating the
  file and reverting, which verified the behavior but shipped no coverage: a
  probe is evidence for a reviewer, not a regression. Split into a pure
  `parseDeferred(content) → deferredState` with a three-way `Verdict()`, table
  tested over synthetic content (marker present / heading stopped parsing /
  section with no marker / all activated / empty file / marker in prose only),
  plus `TestDeferredFileIsGuarding` pinning that the committed file is in the
  *enforcing* state — otherwise the whole guard could sit skipped and green.
- BR-2 — I had copy-pasted the scan-and-dedupe loop into `deferredMarkers`.
  Two copies of the extraction rule computing the gated and not-gated sets is the
  one duplication a disjointness guard cannot afford. Extracted `markersIn(text)`;
  `ArchitectureMarkers()` is now `markersIn(ArchitectureRegistry)`.
- BR-3 — `atlas/workflow/sdlc-binary.md` said "Adding an `ARCH-*` entry flows
  into every consumer with no other edit", which is precisely the claim this
  issue disproved, and it now contradicted the page this diff wrote. Corrected to
  distinguish the runtime consumers (which do derive) from the test layer (which
  did not, and now derives except for the one deliberate tripwire).


**Implementation.** The guard's activation property needed a second pass. The
first cut asserted `architecture-deferred.md` parses ≥1 marker (PQ-2's
anti-vacuity half) — and that made activating the LAST deferred entry fail,
because "the set is empty" is indistinguishable from "the file broke". Probing
it caught this: I moved ARCH-AUTHORITY into `architecture.md` and the guard went
red, falsifying the very claim the deferred file makes about activation being
just a move.

The discriminator is section count, read independently of markers: sections but
no markers = the entries are still there and stopped parsing (fail); no sections
and no markers = everything activated (skip). Three probes now pin it — activate
→ green, break a heading → red, leak `ARCH-AUTHORITY` into a prompt → red.

Marker sites collapsed as planned: `TestArchitectureMarkers` keeps the one
hand-written list, and its comment says why so nobody "fixes" it later.
`TestArchitectureRegistry_Content` gained something real in exchange for its
restated list — it now walks each declared marker's section and asserts all
three lens bullets are present, which nothing checked before. Goldens re-captured
via the blessed `-update-golden`; the diff is the new entry, the count 5 → 6, and
the derived `{{ARCH_STAR}}`.


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
