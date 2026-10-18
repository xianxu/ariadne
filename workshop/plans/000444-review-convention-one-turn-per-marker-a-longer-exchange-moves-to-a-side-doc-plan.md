---
collaboration-mode: discussion
---

# One turn per review marker implementation plan

> **For agentic workers:** Follow AGENTS.md §3 for execution and use the SDLC boundary review at close.

**Goal:** Keep each new inline review marker to one commentary turn, and put any continuing exchange in a linked side doc without making old marker chains unreadable.

**Architecture:** The canonical convention defines the writer rule and retains the reader grammar. The xx-fix and xx-universal-review instructions route clarification and continued discussion through the existing S3 side-doc path; they do not introduce a new marker syntax or change Pair/Couch apply. Before/after fixtures pin the visible result and reject the old append behavior.

**Tech stack:** Markdown skills and conventions; Go fixture checks in `pkg/reviewexport`.

## Core concepts

| Pure entity | Lives in | Status |
|---|---|---|
| Marker turn policy | `construct/local/fix/review-convention.md` | modified |
| Review route | `construct/local/universal-review/round.md`, `collaboration-modes.md` | modified |
| Legacy chain reader rule | `construct/local/fix/review-convention.md`, `construct/local/fix/SKILL.md` | modified |

The marker grammar accepts zero or more commentary blocks for compatibility, while the writer emits at most one. An ask needing more than one turn becomes a side doc linked from the marker's location. A ready legacy chain is read by its last human turn and resolved normally; an agent-ended legacy chain remains pending until the operator responds, so the agent never fabricates consent.

| Integration point | Lives in | Status | Wraps |
|---|---|---|---|
| Hosted and standalone review instructions | `construct/local/fix/SKILL.md`, `construct/local/universal-review/{round,collaboration-modes}.md` | modified | Agent decisions and Pair/Couch apply records |
| Scenario verifier | `pkg/reviewexport/examples_test.go`, `construct/local/universal-review/examples/` | modified | Before/after skill examples |
| Parley compatibility probe | sibling parley.nvim marker parser tests | existing | Legacy-chain reader |

## Design choice

Use an **emit-only** rule. Rewriting the parser to reject chains would break old documents and parley's reader. Adding a new runtime prohibition to Pair/Couch would make the edit engine own a conversation policy that belongs to the agent's route decision; the issue's test contract can be checked at the skill and fixture boundary. Preserve legacy examples only as explicitly labeled reader cases.

The normal route is: carry out a clear instruction and remove its marker; answer a discussion question via the existing side doc; for uncertainty or disagreement, put the question and context in a new side doc and replace the ready marker with its link. If the question is inside a side doc, a follow-up doc uses the same flat folder and names that side doc as parent. In collaboration mode, the same side-doc route is the deliberate exception to in-place answers when the exchange cannot finish in one response. Keep the index entry and link in the same apply batch on tracked documents. Under `presentation: propose`, #437's ready-marker-to-link exception makes that structural replacement direct; ordinary prose edits still follow presentation.

## Steps

1. [x] **Red:** Add a scenario for `🤖[add the numbers]` with two plausible meanings and one for `🤖[ok]` where a side doc offered two options. Expected outputs replace each marker with a side-doc link, include a matching `discussions:` entry, and contain no newly appended `[]{}` chain. Add a legacy `🤖[H]{R}[H]` case whose final human ask is handled without appending. Run the focused `pkg/reviewexport` test and confirm it fails on the old expectation/instruction.
2. [x] **Green:** Revise the canonical convention and the four agent-facing texts. State the writer/reader distinction in a compact rule, replace every append instruction and example, and retain only labeled legacy-chain reading examples. For standalone xx-fix, use the same side-doc layout and link/index contract; for hosted Pair and tracked Couch, issue one atomic apply with link and index records.
3. [x] **Refactor/verify:** Run `go test ./pkg/reviewexport -count=1`, the relevant skill/test package checks, and a live scratch apply probe for a marker-to-link replacement. In parley.nvim, run its existing marker parser test for a legacy multi-turn chain without changing parley. Inspect the diff against main for accidental grammar or side-doc behavior changes. Run `make -k` where needed and the pushed-head suite through `sdlc local-check run --name suite` before close.

## Acceptance

- No agent instruction tells a writer to append `{…}` to a human marker or add a later block to any marker; every continued exchange has a side-doc route.
- Before/after fixtures make a second-turn append fail, and a legacy chain with a final human turn still produces an answer; agent-ended chains stay pending.
- Parley continues to parse old chains, and the #437 edit path can apply the new link/index batch under `presentation: propose`.

## Revisions

- **2026-10-18, close review round 1:** The fixture guard compared after-files only to the same before-path, falsely classifying an unchanged legacy chain after a move as newly written. A move already has an exact `projectMove` projection check, so apply the turn guard only to ordinary rounds. Verify legacy preservation in the renamed parent, both moved side docs, and the displaced orphan; also mutate the projected parent with an appended turn to prove the move check refuses it. Add a concise README pointer for the user-facing convention.
