---
type: continuation
slug: project-vocabulary-lift
agent: claude
session_id: 5e983583-ca62-4c6d-aba6-dae46dac27e4
created: 2026-07-15T23:21:21
branch: main
issues: [000171, 000174, 000175, 000179, 000180, 000181]
---

# Continuation: project-vocabulary-lift

## NEXT ACTION

Run `sdlc start-plan --issue 180` and design the project vocabulary model.
#180 is already claimed and fully brainstormed — its `## Spec` is the
converged design input (lifecycle funnel, two-phase estimation, kanban
baseline/derived/logged split, `workshop/projects/` residency). The first
design decision, before any cue authoring: whether to structure the whole
project-management lift (this issue + #171's gate half + the verbs) **as the
first dogfood project file** — the Spec proposes it, and it changes whether
#180 is one multi-boundary issue or the anchor of several. Then the durable
plan per the Plan row: cue shape (cross-repo discovery is the novel part),
transition-guard mechanics, Phase-A estimate vocabulary, which gate owns
conformance, verb set, ordering vs #171.

## State of play

- **#180 (project vocabulary model)** — working, claimed, brainstorm
  converged into the Spec (operator refinements folded 2026-07-15). Design
  not started. This is the resume target.
- **#171 (brain-vs-repos tension)** — working, claimed. Direction settled
  after a three-stage arc (see Revisions): project files live in coding
  repos under `workshop/projects/` (plural confirmed), parley navigates,
  NO meta repo, brain unchanged. Its implementation (project-gate lift in
  `sdlc close`, moving the 5 brain project files) should CONSUME #180's
  model — model first or together.
- **Done + merged this session:** #175 (verdict-gate single-pass recovery,
  PR 96), #174 (post-FIX-THEN-SHIP protocol + doc-tolerant publish gate,
  PR 97 — the protocol has since fired live on three closes), #179
  (`sdlc migrate`, PR 98), #181 (history subfolders, PR 99 + fleet
  migration: all 9 ariadne-shaped repos incl. metis migrated and pushed).
- `sdlc state` for drift; #52, #119 remain in-flight from earlier work.

## Thread arc & user model

The session opened as friction-audit cleanup (#175, #174 from #172's
measurements) and pivoted when the operator asked about #171. That
brainstorm ran three stages across two days — meta repo decided → two-axis
reframe (scope × commit-rhythm; brain vindicated as a correct per-user
auto-save store) → meta DROPPED for peer-repo `repo#id` addressing + lifting
project management into the sdlc spine. #179 and #181 were built mid-arc as
supporting infrastructure the moment they crystallized.

The latent intention connecting every pivot: the operator is upgrading the
SDLC from single-issue tooling to **portfolio/project-level tooling**, by
first principles rather than convention — they rejected the Jira-shaped hub
because containers ossify org structure, insisted breadth is a
view/resolution concern not a storage concern, and want project management
to get the same lift issues got (formal model, binary-owned gates,
calibrated estimation, derived views). Decision style observed: chooses
crisply when given weighed options, revises freely and without ego
(meta → drop), expects decision arcs recorded honestly (Revisions, not
overwrites), and consistently prefers mechanisms over mandates.

## Open questions

On resume, resolve these open questions with the user before continuing
with the NEXT ACTION.

- **Dogfood structure:** should #180's design phase begin by creating the
  first project file (the project-management lift itself, PRD'd by its own
  emerging machinery), or land the cue/vocab model first and dogfood after?
  This decides whether #180 stays single-issue or becomes the anchor of a
  multi-issue project.
- **Ordering vs #171:** model-first (do #180 fully, then #171's gate lift
  consumes it) or interleaved (one plan covering both)? The issues say
  "soft ordering: model first or together" — the operator hasn't picked.
- **Project archive-on-done:** `workshop/history/projects/` (the #181
  layout has the slot; `vocab.ArchiveSubdirs` widens in one edit) vs
  projects staying in place as records (the datatype prose's current
  claim). Parked in #180's Spec, needs an operator call at design time.

## Artifact map

Read in this order on resume:

- `workshop/issues/000180-project-vocabulary-model-…​.md` — THE input. Its
  Spec is the brainstorm output (taxonomy, lifecycle funnel
  `ideation → defined → committed → executing → done|dropped` (+paused),
  `deadline:` attribute, two-phase estimation with the fog-factor
  calibration bridge, kanban three-way split, retro mechanism,
  `workshop/projects/` residency); its Log records why each choice.
- `workshop/issues/000171-the-tension-between-brain-and-other-repos.md` —
  the companion issue: rewritten Done-when (the settled shape) + the full
  Revisions arc (meta decided → reframed → dropped). Constrains #180's
  discovery design: project lookup must resolve ACROSS peer repos.
- `construct/datatype/project.md` — the prose datatype #180 demotes to
  citing the cue as schema authority (drift test planned).
- `construct/vocabulary/issue.cue` + `pkg/vocab/vocab.go` — the pattern to
  mirror (`vocab.Project()` alongside `Issue()`); `ArchiveSubdirs` is the
  named extension point for `history/projects/`.
- `workshop/lessons.md` (tail) — two entries from this session: piped
  test runs gate on the pipe's last exit code; `os.Getwd` returns the
  logical `$PWD` while git returns resolved paths (EvalSymlinks both).
- Shipped tools the design leans on: `sdlc migrate` (cross-repo artifact
  move with ref rewrite — makes the residency default cheap to revise) and
  the subfoldered archive (all 9 fleet repos migrated; brain's 5 project
  files in `brain/data/project/` are the migration #171 will execute).

Branch: main (all session work merged; no feature branch open).

## Live deliberations

- **Cue encoding of cross-repo discovery** — issue.cue's `discovery:` block
  assumes one repo; project discovery is fleet-wide. Encode "glob
  `workshop/projects/*.md` across peers" in the model, or keep the cue
  location-only and let resolution own the cross-repo walk? Leaning: cue
  declares the per-repo home, resolution owns the walk (matches how
  `resolveRepoDir` already works).
- **Phase-A estimate vocabulary** — needs its own primitives + uncertainty
  multiplier, same calibration process as v3.1 but different axes
  (operator's explicit refinement). Undesigned.
- **ArchiveSubdirs widening** (#181 close review note): third return value
  vs kind-keyed map when `history/projects/` lands — kind-keyed scales
  better if more kinds appear.

## Decisions & dead ends

- **Meta repo: decided (7/14) then DROPPED (7/15)** — hub model reproduces
  Jira's container pathology; peer-repo addressing dissolves the container
  question. Full arc in #171 Revisions.
- **Calibration data to ~/.claude: proposed and retracted** — brain is the
  designed per-user versioned store; velocity ledger/transcripts stay.
- **Nous rhythm carve-out in brain: rejected** ("too messy" — operator).
- **Top-product residency as a hard rule: softened** to center-of-gravity
  default once `repo#id` addressing made moves cheap (metis example showed
  work-site vs proof-site divergence makes any hard rule wobble).
- **`sdlc reverify` (#174 candidate B): skipped** — without teeth it's
  `--no-judge` with nicer bookkeeping.
- **SDLC process artifacts in brain: banned outright** (operator amendment,
  including the endpoint-less roadmap I'd proposed parking there).

## Lessons learned

- The #174 protocol works as designed — three FIX-THEN-SHIP closes since
  (fix findings pre-commit, bundle into the close commit, anchor == HEAD,
  publish gate passes). Follow it on #180's closes.
- Fresh-eyes plan reviews caught real bugs every time this session (wrong
  status assertion, silent span corruption, traversal write, missed
  write-site). Keep dispatching them at every plan.
- The "project lifecycle = issue lifecycle one level up" frame is
  load-bearing for #180 — when a design question stalls, ask what the
  issue-level machinery does and lift it.
- Live dogfood after green tests keeps finding what hermetic suites can't
  (the $PWD symlink guard misfire) — budget a real-fixture pass into every
  IO-adjacent milestone.
