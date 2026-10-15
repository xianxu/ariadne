---
id: 000399
status: open
deps: [parley.nvim#332, pair#573]
github_issue:
created: 2026-10-15
updated: 2026-10-15
estimate_hours:
card_mirror: '4d1322dd4e320147e1b1fea6afa626e9b3148e4b' # card fields mirrored from issue-cards; edit via sdlc
---

# Universal review skill and canonical round export

## Problem

Universal-agent-input-interface S2a (parley.nvim#319) should send a short review
request rather than repeating the review protocol every round. Agents need one
shared skill, and standalone Parley still needs an inline fallback derived from
the same protocol. S1 is parley.nvim#332 plus pair#573.

## Spec

- Add the exported skill construct/local/universal-review/SKILL.md (discovered
  as xx-universal-review through the existing skill manifest), available to all
  slots. tl-lead and tl-worker reference it for document review requests.
- Own the generic review-round protocol in round.md and the existing six stock
  mode files in modes/. Preserve unique-prefix/multiple-mode behavior and the
  first mode's scope/deletions/frontier semantics. Do not add section modes or
  side-doc lifecycle here; those are later project slices.
- Reference the canonical marker grammar in ../fix/review-convention.md; never
  create another authored grammar. Define file editing vs the Parley tool
  adapter separately. For registered agentic documents use Couch read/diff/write
  CAS, respect the operator lock and fold owed deliveries; do not bypass S1.
- Publish a small versioned export manifest naming round.md, modes/*.md and
  the canonical marker grammar as the inputs a Parley release snapshots. The
  skill, generated inline fallback and local review adapter must derive from
  these inputs; no independently maintained Parley mode/protocol definitions.
- Land before Parley #319 implementation. Coordinate the final export shape
  with its owner; the Parley issue owns generator/loader/message changes.

## Done when

- The skill is discoverable in an ariadne-derived fixture for the supported
  harness directories, and TL skills point to it.
- A representative review request can be answered using the skill, preserving
  mode and marker behavior; registered files use the existing lock/CAS path.
- The exported round/modes/grammar input manifest is deterministic and supplies
  Parley's standalone inline fallback without authored duplicate semantics.
- Validation covers the actual export and existing marker/skill discovery
  contracts, plus a live operator check coordinated by the hub.

## Plan

- [ ] Confirm the shared export contract with parley.nvim#319, then author the
  skill via the writing-skills workflow and land it before the consumer.

## Log

2026-10-15 — filed from parley.nvim:1 while planning the operator-requested S2a;
unclaimed for hub dispatch. See ops teams/devinfra/projects/universal-agent-input-interface/design.md
S2 and project.md S2a. This is the shared-skill half; no Parley code belongs here.
