---
id: 000238
status: open
deps: []
github_issue:
created: 2026-09-18
updated: 2026-09-18
estimate_hours:
---

# atlas: split by purpose — map, journeys, workflow — with agent instructions

## Problem

Atlas was meant as key pointers that help an agent understand what a repo is
and how it's laid out, not a comprehensive or user-facing manual. Over time it
has accumulated three purposes in one folder:

1. **The map:** what's where, terminology, why things are shaped this way.
2. **User-facing behavior:** what the user sees and does.
3. **How the code works inside:** internals written out as prose.

Surveys (2026-09-18):
- **parley:** 60 pages, about 43 on features the user sees but with no journeys,
  and implementation words leak into them ("generation" appears in 20 files,
  "grant" in 15).
- **pair:** about 5,100 lines, mostly internals (`couch.md` alone is 1,882
  lines), with no page written from the user's side.

AGENTS.md §8 already says "Map, don't over-specify; details live in code +
issues". Prose that restates code drifts. Both surveys found real
contradictions:
- **parley:** `atlas/chat/lifecycle.md:4` says "save edits with `:write`", while
  `packaging/tutorials/basics.md:35` says it autosaves.
- **pair:** atlas says the resume picker "offers up to three options"; the README
  lists four.

`prd` (ariadne#237) needs a baseline of current user-facing behavior to check a
PRD against, and no atlas has one.

## Spec

**Three purposes, three places:**

| folder | purpose | reader | kept current by |
|---|---|---|---|
| `atlas/` (root + feature folders) | **Map:** what's where, terminology, why it's shaped this way. Short pages of pointers. | every agent, at orientation | close gate (as today) |
| `atlas/journeys/` *(new)* | **User journeys:** one page per central journey, with steps in the user's words, an interruption table of *current* behavior, and links down to map pages | `prd` and its check (ariadne#237) | issues with a PRD update their journey page at close (ariadne#237) |
| `atlas/workflow/` *(exists)* | **Process:** sdlc and the workflow (a symlink to ariadne's in consumers) | agents following sdlc | ariadne |

**No `mechanism/` folder.** A folder for internals would make over-specification
look like accepted practice.

**Sorting rule for existing content**, applied per section:

| kind of content | where it goes |
|---|---|
| User-visible behavior | `journeys/` |
| An invariant | `workshop/targets/` |
| A pointer or a design reason | stays in the map, cut to a few lines |
| Prose that only restates the code | deleted (code is the truth; issues and plans hold the history) |

**Journey page shape as a datatype.** Add `journey` to
`construct/datatype/` so `datatype show journey` gives agents the shape and the
authoring instructions:
- frontmatter: `type: journey`, name, created/updated
- a lede: who, and the goal
- `## Steps`, in the user's words
- `## Interruptions`: a table of step × interruption → what the user sees.
  Rows that no source settles are marked *undecided*, not guessed.
- `## Links`: pointers to the map pages

The standard interruption list lives in the datatype, not in each page:
- **The user:** edits nearby, waits, undoes, stops, regenerates, switches away,
  quits and reopens.
- **The system:** fails or never returns, crashes mid-effect, is slow, is busy
  with another task.

Journeys use user vocabulary only. An implementation term in a journey page is
a finding.

**`atlas/index.md` sections by purpose:** Map, Journeys, Workflow. Every file is
still linked, as today.

**Agent instructions.** Rewrite AGENTS.md §8 (base layer, composed to
consumers by weave):
- what goes where
- the sorting rule
- journeys in user words, never implementation terms
- "map, don't over-specify" restated as a rule with the sorting rule as its
  mechanism

The close gate's atlas check doesn't change here. Making issues with a PRD update
their journey page is ariadne#237's job.

**Migration.** One issue per repo that has its own atlas content, each depending
on this one:
- parley.nvim, pair, nous, you-decide, kbench, tools, metis, kaggle, xianxu.dev
- ariadne's own two pages are migrated here, as the reference
- 42shots and astro have only the workflow symlink: nothing to migrate
- brain repos are excluded by the brain charter (no sdlc issues); their atlases
  can be moved later as a plain edit

## Done when

- AGENTS.md §8 (base) describes the three purposes, the sorting rule and the
  vocabulary rule. `weave` composes it into consumers.
- `construct/datatype/journey.md` exists, `datatype show journey` prints it, and
  it includes the standard interruption list.
- ariadne's own `atlas/index.md` is organized by purpose and its pages conform.
- The migration issues exist in each listed repo with `deps: [ariadne#238]`, and
  ariadne#237 depends on this issue.

## Plan

- [ ] Write the `journey` datatype (shape + authoring instructions + the
  interruption list).
- [ ] Rewrite AGENTS.md §8 (base); run `weave` to compose it into consumers.
- [ ] Migrate ariadne's own atlas as the reference example.
- [ ] Check the per-repo migration issues are filed (filed 2026-09-18, see Log).

## Log

### 2026-09-18

Filed from a brain advisor session. Order agreed with the operator: this issue
first, then each repo's migration, then `prd` (ariadne#237). Surveys of parley
and pair are summarized in the Problem. Discussion:
`brain/workshop/pensive/2026-09-18-01-pensive-product-lens-user-model.md`.

Migration issues filed, each with `deps: [ariadne#238]`:
- parley.nvim#269 and pair#290, with details from the survey
- nous#53, you-decide#14, kbench#28, tools#78, metis#69, kaggle#10 and
  xianxu.dev#3, with a page inventory and a first step to identify their
  journeys
