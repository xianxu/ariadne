---
id: 000399
status: working
deps: [parley.nvim#332, pair#573]
github_issue:
created: 2026-10-15
updated: 2026-10-15
estimate_hours:
card_mirror: 'e1514d8edef259613b78c11d7f6ec74e85ed46bc' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-15T16:24:58-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:2
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: quick, provenance: inferred, spec: "9ac9bf24", done: "43b2372e"}
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

Quick flow: prose plus one small Go validator; no Mx.

Layout, `construct/local/universal-review/` (weave lowers it to `xx-universal-review` in `.claude/skills` and `.agents/skills`):
- `SKILL.md`: the trigger is a `submitted: PATH` (or `review request: PATH`) delivery for a markdown document. The steps: read `round.md`, choose the adapter, and follow it. It is short and restates no protocol.
- `round.md`: the harness-neutral round. It covers:
  - the inputs: the document, the frontmatter `process-instruction: [mode][mode] text` (`free-form` when there is no mode), and the changes since your last read;
  - mode resolution: an exact name, else a unique prefix; an ambiguous or unknown name is refused, naming the modes; several modes give every body but only the first mode's directives;
  - what the mode flags mean as directives: scope, deletions, frontier;
  - marker readiness and the reply rules. It refers to the grammar and restates none of it.
- `modes/*.md`: the six stock modes, byte-for-byte from parley's `lua/parley/skills/review/modes/`. They become canonical here; parley regenerates its copy (#319).
- `adapters/file-edit.md`: editing the file itself.
  - A **registered** document (`couch --session-sync lookup`) goes `diff`, then `read --all` for the hash, then `write --all --if-hash H --fold <owed IDs>`. Exit 3 (the operator holds the lock) queues the write until the next submit, and a lock older than 30 minutes is pinged. Exit 6 (stale) re-reads. The CAS is never bypassed.
  - An **unregistered** document is edited in place; changes since the last read come from your own last read, else the whole file.
  - The `propose_edits` tool adapter stays parley-owned (#319's `local-adapter.md`).
- `review-convention.md`: a symlink to `../fix/review-convention.md`, the one authored grammar. A relative link then works both lexically, through `.claude/skills/xx-universal-review/`, and physically.
- `export.json`: the versioned manifest that parley's `sync-review-protocol.py` snapshots.
  - Header: `{"schema":"ariadne/universal-review-export","version":1,"members":[…]}`.
  - Each member has `role` (round | grammar | adapter | mode), `path` relative to the skill directory, `bytes` and `sha256`.
  - The grammar member also has `source: construct/local/fix/review-convention.md`.
  - Mode members are dense records: `name`, `order`, `scope`, `deletions`, `frontier`.
  - Members are sorted by role and then order, with no timestamps, so the file is deterministic.

Tasks:
- [x] Author `SKILL.md`, `round.md` and `adapters/file-edit.md` via the superpowers-writing-skills workflow. Copy the six modes and add the grammar symlink.
- [x] Add a `pkg/reviewexport` validator:
  - it parses mode frontmatter as parley does (flat keys; name equals basename; valid enum values);
  - it builds the manifest;
  - `TestExportCurrent` fails on a stale or missing member, and `UPDATE_REVIEW_EXPORT=1` regenerates the manifest;
  - a determinism test builds it twice and checks the two are identical.
- [x] A discovery test: an ariadne-derived fixture whose weave lowering yields `xx-universal-review` in both harness directories, with `review-convention.md` resolving through the harness path.
- [x] Point `tl-lead --help` and `tl-worker --help` at the skill for review requests, with phrase guards. The `xx-couch` skill gets the same one-line pointer.
- [x] Atlas: a short `atlas/workflow/universal-review.md`, linked from the indexes.
- [ ] Live check, coordinated by the hub: one operator Alt+Return on a registered document is answered by the skill under the lock and the CAS.

## Log

2026-10-15 — filed from parley.nvim:1 while planning the operator-requested S2a;
unclaimed for hub dispatch. See ops teams/devinfra/projects/universal-agent-input-interface/design.md
S2 and project.md S2a. This is the shared-skill half; no Parley code belongs here.
- 2026-10-15 (ariadne:2): claimed. TL decision (ops:0): Alt+Return sends ONLY 'submitted: PATH'; the skill carries the whole protocol (frontmatter process-instruction, couch --session-sync diff for changes since last read, review modes, marker grammar). parley.nvim:1 confirms #319 is submitted-only: it deletes the verbose Couch builder/diff baseline and keeps only the inline in-editor fallback outside Couch, so the export must serve both: the skill (in Couch) and the inline fallback (standalone).
- 2026-10-15: TL approved the plan at 6c384550 (delivery 18480ed1). The TL also asked to keep round.md open for S2's section-mode step, which is now its §4 (reserved).
- 2026-10-15: parley.nvim:1's consumer review (delivery 97b175aa) is folded in:
  - **Direct edits:** a file is edited directly only on a confirmed-untracked lookup. A `type: session-sync` file always takes the lock and CAS path. An unknown lookup result writes nothing.
  - **Snapshot:** the round reviews the `read --all` snapshot whose hash is H. It folds only the owed IDs captured in `status` before the diff; anything delivered later stays owed. With nothing owed, `--fold` is left out. On a lock, the pending ask carries over to the next submit, even if that submit's diff is empty.
  - **Composition:** round.md names no Couch verb, so it works outside Couch. Parley's local prompt is round, grammar, the selected modes and its own `propose_edits` adapter; it excludes SKILL.md and the file-edit adapter.
  - **Manifest pins:**
    - sizes and hashes cover each file's resolved contents, including the grammar symlink;
    - the order is total: role, then mode order with unordered modes last, then name, then path;
    - an omitted mode flag takes parley's default.
- 2026-10-15: built. These tests are green:
  - `pkg/reviewexport`: ParseMode, its defaults and refusals; the build order; missing-member refusals; determinism; TestExportCurrent.
  - `cmd/weave` TestUniversalReviewSkillDiscoveredInDerivedRepo: both harness directories, with the grammar bytes equal to the canonical file.
  - The tl-lead and tl-worker phrase guards.
- 2026-10-15: dry-run check. A fresh subagent, given only the skill, answered `submitted: note.md` (untracked, `[copy] keep it short`):
  - it resolved `[copy]` to copy-editing by unique prefix;
  - it left the typo above the frontier untouched;
  - it turned the ready grammar ask and a misspelling below the frontier into strike proposals;
  - it left the pending `{…}` marker and the marker inside a fence.
  It also flagged ambiguities, so round.md now states:
  - how an explicit ask works under propose-strike: the proposal replaces the human marker and is the answer;
  - that replacement forms defer to the grammar's §3;
  - the frontier's edges: fenced markers don't count, the referred text is open, and with no human marker nothing is settled;
  - "edit region" in place of the overloaded word "scope";
  - "pending" as any agent turn, whoever wrote it;
  - where the operator's free text applies.
- 2026-10-15: parley.nvim:1 checked the pin at 8bead12b (delivery c732ec76): all 9 member hashes match, and the six modes are byte-identical to parley's. The export shape is accepted. Two corrections to the file-edit adapter followed:
  - `type: session-sync` is tested first, before `COUCH_TREE`. A legacy session-sync file stays tracked outside a slot, and when its storage is unreachable the adapter refuses instead of editing directly.
  - The write folds only the kept IDs whose asks it answers. A kept delivery that wasn't answered stays owed.
