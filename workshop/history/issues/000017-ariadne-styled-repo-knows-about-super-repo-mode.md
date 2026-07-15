---
id: 000017
status: done
deps: []
created: 2026-04-30
updated: 2026-06-03
---

# ariadne styled repo knows about super-repo mode

make ariadne styled repo understand super-repo mode, e.g. peer repos with .parley marker file defined. to facilitate various cross repo reference and search. 



## Done when

- An ariadne repo recognizes peer/super-repo siblings and supports cross-repo
  reference + blocking deps.

## Spec

Realized — the marker just evolved. #17 imagined a `.parley` marker file; the
model that shipped uses **`construct/`** to identify an ariadne-styled peer and
**`.brain/config.md`** for the special "brain" peer (cross-cutting `project`/
`roadmap` state). Documented in AGENTS.md's "Peer Repo" section (§1); cross-repo
blocking deps go through `sdlc issue new --deps repo#N`. The `.parley` marker was
never built (zero references repo-wide). Used daily — e.g. `../nous` is referenced
routinely (nous#14 / nous#41 surfaced in the #68 actuals work).

## Plan

- [x] Realized via the peer/brain model: AGENTS.md "Peer Repo" section
  (`construct/` = ariadne-styled peer; `.brain/config.md` = brain) + `sdlc issue
  new --deps repo#N` cross-repo deps. No #17-specific implementation — built
  across the construct/brain work; `.parley` marker superseded.

## Log

### 2026-06-03 — done (marker evolved `.parley` → `construct/`/`.brain`)
- 2026-06-03: closed — peer/super-repo awareness realized: AGENTS.md Peer Repo section (construct/ = ariadne-styled peer, .brain/config.md = brain) + sdlc issue new --deps repo#N cross-repo deps; used daily (../nous referenced routinely). .parley marker superseded (zero refs). Validation close: no #17-specific impl (--no-actual), doc-only (--no-judge/--no-atlas).; review verdict: not-run
- Operator: "we refer to `../nous` all the time" — peer-repo reference is routine
  practice. The capability exists and is in the constitution (AGENTS.md Peer Repo
  section + `--deps`); the original `.parley`-marker mechanism was superseded by
  `construct/`/`.brain/config.md`. Closing against intent (peer/super-repo
  awareness), not the deprecated marker. Validation close — no implementation to
  attribute to #17 specifically.

### 2026-04-30

