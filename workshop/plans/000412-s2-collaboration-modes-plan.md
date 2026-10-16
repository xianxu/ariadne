# S2 M1: collaboration modes in xx-universal-review: implementation plan

> **For agentic workers:** Consult AGENTS.md Section 3 (Subagent Strategy) to determine the appropriate execution approach: use superpowers-subagent-driven-development (if subagents are suitable per AGENTS.md) or superpowers-executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Each section of a reviewed document gets a collaboration mode (discussion, collaboration, chat) that says
where an ask's answer goes, as the operator-decided S2 design specifies.

**Architecture:** The protocol stays prose in the shared skill, `construct/local/universal-review/`. Parley's
standalone prompt runs it with no binary, so the prose is the one source (ARCH-DRY), and nothing in Go restates it.
A new `collaboration-modes.md` holds the modes and the side-doc contract. `round.md` §4 resolves them and routes each
ask. `examples/` holds a before/after document set per mode. A Go test checks the examples against the side-doc
contract, so the reference fixtures can't drift from the prose.

**Tech Stack:** Markdown (skill protocol), Go (`pkg/reviewexport`, its tests, `cmd/tl-lead` help and topic test).

**Source of truth:** ops `teams/devinfra/projects/universal-agent-input-interface/design.side/01-s2-section-modes.md`
at ops db130ac0 (sections 2 to 7), with side docs 02 to 06. Below it's cited as **D§n**. Design owner: parley.nvim:4.

---

## Core concepts

### Pure entities

| Name | Lives in | Status |
|---|---|---|
| `collaboration-modes.md` (the protocol) | `construct/local/universal-review/collaboration-modes.md` | new |
| round §4 (routing) | `construct/local/universal-review/round.md` | modified |
| `examples/<mode>/` (reference rounds) | `construct/local/universal-review/examples/` | new |
| `Member` role `collaboration-modes` | `pkg/reviewexport/export.go` | modified |
| `sideDoc` / `checkExample` (fixture checker) | `pkg/reviewexport/examples_test.go` | new |

- **collaboration-modes.md**: the one authored statement of D§2 to D§5. It covers:
  - the frontmatter (block, flow and bare-scalar forms);
  - heading matching and its four error cases;
  - the three protocols, with the imperative/question rule and its strictness;
  - the instructions-that-act table;
  - how it combines with `process-instruction:`;
  - the side-doc creation contract;
  - the state table, rows 1 to 17.

  `round.md` restates none of it, the way it already treats the grammar.
  - **Relationships:** 1:1 with round.md §4, which points at it. 1:N with examples (one set per mode).
  - **DRY rationale:** D§6 places the protocols in one file. tl-lead's help cites it rather than restating the
    Operator-instructions rules (below).
  - **Future extensions:** S3 adds the asked fold-in, the `discussions:` index, closing and renames to the same file.
    M2's per-type defaults are templates, not this file.
- **round.md §4**: replaces the "reserved" text with three steps:
  - resolve `collaboration-mode:` (or fail the round on D§2's two fatal cases);
  - give each ask its section's mode;
  - route it per `collaboration-modes.md`.

  §5's marker rules stay as they are; they become collaboration mode's protocol and the shared marker states. Two
  edits come with §4:
  - §5's opening line: "act on it" becomes "act on it as its section's collaboration mode says (§4)".
  - §6 (the reply): it adds the dangling names and the orphans to what the reply reports.
- **examples/**: one directory per mode, each holding `before.md`, `after.md` and, where the round creates them,
  `after.side/` (named as the parent's `<name>.side/`). Every state-table row is covered by an example or by a named
  check (the test list below). The examples are reference rounds a reviewer or a live check compares output to; they
  aren't exported to parley (agreed with parley.nvim:4, Q1).
- **examples_test.go**: a test-only checker. It isn't production code: the agent and parley read the prose. It
  parses each example's `after.md` and side docs and asserts the contract (below). It lives beside `TestExportCurrent`
  because that package already owns "the committed skill directory is well-formed".

### Integration points

| Name | Lives in | Status | Wraps |
|---|---|---|---|
| the file-edit adapter | `construct/local/universal-review/adapters/file-edit.md` | modified (one paragraph) | `couch --session-sync` |
| `tl-lead status instructions link` | `cmd/tl-lead/instructions.go` | unchanged | `couch --session-sync` |

- **The adapter** gains a single paragraph. A round may create side docs: new, untracked files, which are written
  directly before the parent's one write. Only the parent goes through the lock and the compare-and-swap (D§4, "one
  write per round"). Its stateful fake is Couch's own session-sync, which the existing tests already cover, so nothing
  new is needed.
- **tl-lead**: the session file keeps its own write path (D§6). `instructions link` already accepts a
  `tl-status-<op>.side/…` target: its validation refuses only whitespace and parentheses. One test case pins that.

## Design decisions (ARCH)

- **ARCH-DRY / ARCH-PURPOSE: no Go router.** A `sdlc review sections FILE` helper could compute heading modes
  deterministically. I rejected it because parley's standalone prompt has no binary to call, and a Go copy of the
  rules would be a second source that drifts. The purpose is the protocol, and every consumer derives from it:
  - the agent in a slot (via SKILL.md and round.md);
  - parley (via export.json);
  - the TL (tl-lead help cites the skill).
- **ARCH-SECURE: a marker is untrusted text that lands in YAML.** The side doc's `marker:` is the operator's text,
  one line. The contract says to write it as a YAML double-quoted scalar, escaping `\` and `"`. A `<br>` stays
  literal, and a newline never appears, because markers are single-line. The slug is derived as `[a-z0-9-]`, at most
  40 characters, and falls back to `question` when empty. The checker asserts both on the examples.
- **ARCH-ORDER: a round that creates side docs.** The states per ask are: no side doc; written but unlinked (the
  parent's write was refused); linked. The events:
  - **The parent's write is refused** (exit 3, the lock; exit 6, stale). The side docs stay. The next round finds
    one by its `parent:` plus `marker:` and links it instead of creating a duplicate. This makes the round idempotent
    under retry.
  - **The marker is deleted or rewritten before the retry.** The orphan is kept and named in the reply.
  - **A duplicate submit:** the same as a retry.
  - **Two rounds at once** can't happen: one agent answers a document, and the compare-and-swap rejects a stale
    parent write.

  The event most likely to be mishandled is a retry after exit 6 that re-numbers the side docs. The contract
  therefore says to reuse the side docs by `(parent, marker)` before allocating a number. Example
  `discussion/retry` covers it (row 17b).
- **ARCH-FUNERAL:** side docs and answer docs are never pruned by the agent (D§4). Each comes from one operator ask.
  The operator ends one by deleting it; S3 adds closing. The examples are static fixtures. The adapter's
  temporary-file rules are unchanged.
- **ARCH-CONSTRAINTS:** this is not applicable in practice: a round is one agent turn over one document, and side
  docs add one file write per question.
- **ARCH-MOCK:** this adds no new external dependency. Couch is reached only through the adapter's existing seam.
- **ARCH-PURE:** the checker is a pure function over an `fs.FS`, like `Build`.

## The test list: the state table (D§5) and the contract edges

Each row is covered by an `examples/` case, which the live check runs and the Go checker validates for structure, by
a Go assertion, or by both. A row marked *prose* is a rule the examples can't show without a real Couch (the lock).
It is checked by a grep assertion that the rule's sentence appears in `collaboration-modes.md`.

| Row | Situation | Covered by |
|---|---|---|
| 1 | edits only, no ask | `examples/collaboration/` has a section with direct edits and no marker; after.md is unchanged there |
| 2 | discussion imperative | `examples/discussion/`: 🤖[tighten this paragraph] becomes an edit; under propose-strike it becomes a 🤖~old~{new} proposal |
| 3 | discussion question | `examples/discussion/`: the marker becomes `→ [discussion](before.side/01-….md)`. Checker: the link resolves, the frontmatter is complete, the name is `NN-slug` |
| 4 | small-unclear imperative | `examples/discussion/`: 🤖[add the numbers]{the latency numbers or the cost numbers?} |
| 5 | unsure | `examples/discussion/`: "the retry story feels off" opens a side doc |
| 6 | a follow-up in a side doc | `examples/discussion/`: a side doc whose own question opens `02-…` in the same folder, with `parent:` naming `01-….md`. Checker: no `.side/` inside `.side/` |
| 7 | "ok" in a side doc | `examples/discussion/`: the parent's text holds the proposal, the link reads "(folded 2026-10-16)", and the side doc ends with the applied note |
| 8 | an ambiguous "ok" | `examples/discussion/`: 🤖[ok]{which: …?} in a side doc that proposed two things |
| 9 | "file this" | prose (filing is a real effect); the example shows the side doc's closing record line, with no issue filed |
| 10 | an asked fold-in | `examples/discussion/`: a 🤖[fold in] by an existing link becomes an edit to the parent, with the link marked "(folded 2026-10-16)" (O1, decided). The index and closing stay S3 |
| 11 | collaboration question | `examples/collaboration/`: the answer is folded into the prose, the marker is gone, and no side doc exists. Checker: no `.side/` link in collaboration sections |
| 12 | a new chat item | `examples/chat/`: item 5 gains ` → [answer](before.side/2026-10-16-01-s4-cost.md)`. Checker: the dated name, numbered per day |
| 13 | a chat instruction | `examples/chat/`: the answer doc records what was done and where |
| 14 | a chat follow-up | `examples/chat/`: an edited, linked item; its answer doc gains a follow-up section; the link is unchanged |
| 15 | a deleted chat item | prose: the doc is never reopened. The example keeps an answer doc whose item is gone and leaves it untouched |
| 16 | a dangling heading name | `examples/chat/` declares a misspelt name; the expected reply line is in `examples/chat/reply.txt` |
| 17 | the lock held | prose: the adapter is unchanged (exit 3 defers) |
| 17b | a retry after a refused write | `examples/discussion/retry/`: an unlinked side doc exists in before; after links it with no new number. Checker: the `(parent, marker)` pair is unique |

The contract edges (D§2, D§4), each an example or a checker case:

| Edge | Covered by |
|---|---|
| the block, flow and bare-scalar forms | one example each declares one form (chat: block; collaboration: bare scalar; discussion: flow) |
| absent key → discussion | `examples/discussion/retry/` has no `collaboration-mode:` |
| case, whitespace and a trailing `(…)` ignored; subsections inherit; fenced headings don't count; duplicate names both take the mode; text above the first heading takes the default | `examples/chat/` has a `### Operator instructions (open)` and a fenced `## Notes` |
| a name under two modes → the round fails, unchanged | prose plus `examples/errors/conflict.md` with its expected reply (no after.md: unchanged) |
| an unknown mode → the round fails, listing the three | `examples/errors/unknown-mode.md` with its expected reply |
| side-doc frontmatter: `type`, `parent` (no `#`), `marker` (double-quoted), `collaboration-mode: discussion`, `process-instruction` copied only when the parent has one | checker, on every side doc in every example |
| names: `NN-slug` widens (`099`, then `100`); dated `YYYY-MM-DD-NN-slug`; numbers sort before dates | checker: a pure `nextSideDocName` over a fake listing, unit-tested, used only by the checker |
| propose-strike never strikes a marker→link swap | `examples/discussion/` declares `[free-form]`; its link replaced the marker with no strike |
| chat items are always answered, whatever the frontier | `examples/chat/` declares `[copy-editing]` (frontier on) and has items above the topmost marker |

## Tasks

- [ ] **1. collaboration-modes.md.** Write it from D§1 to D§5, in the skill's own voice (harness-neutral, naming no
  Couch verb; the adapter owns IO). Its sections: the frontmatter, matching sections, the three modes, instructions
  that act, combining with process-instruction, side docs, the state table, older documents. Commit.
- [ ] **2. round.md §4 and the touch points.** Rewrite §4 (resolve, then route). Update §5's opening and §6's reply
  list. Add a pointer in SKILL.md (step 1 reads collaboration-modes.md too). Add the side-doc paragraph to the
  adapter. Commit.
- [ ] **3. examples/.** Write `discussion/` (with `retry/`), `collaboration/`, `chat/` and `errors/`, by hand, to
  the test list. Commit.
- [ ] **4. The checker (TDD).** Write `examples_test.go` failing first: point it at a deliberately broken copy in
  `t.TempDir()` (a bad name, a `#anchor` parent, a nested `.side/`, a duplicate `(parent, marker)`, an unquoted
  marker), then at the real examples. Unit-test `nextSideDocName` (`098`→`099`→`100`, dated per day, sort order).
  Commit.
- [ ] **5. Export.** Add the `collaboration-modes` member to `Build` (rank between grammar and adapter), extend
  `TestBuildRefusesMissingMembers` and the fixture, and regenerate `export.json`. Keep Version 1 (Q1).
  `TestExportCurrent` asserts that exactly one `collaboration-modes` member exists. Commit.
- [ ] **6. tl-lead.** Change `help.md` in three places:
  - the session-file frontmatter gains `collaboration-mode: {chat: ["Operator instructions"]}`;
  - the Operator-instructions paragraph cites xx-universal-review's chat mode for the rules, keeping only the TL's
    write path (`instructions list|link`);
  - QA files become answer docs in `tl-status-<operator>.side/YYYY-MM-DD-NN-<slug>.md`, per the skill. The body
    (Needs you, Threads) is discussion, so a 🤖 question there opens a side doc in the same folder.

  Then update `main_test.go`'s G2/G3 topics to the new phrases, and add the `instructions link` case for a `.side/`
  target. Commit.
- [ ] **7. The atlas.** Update `atlas/workflow/universal-review.md` (collaboration-modes.md, examples/, §4 no longer
  reserved) and `atlas/workflow/tl-skills.md` if it names QA paths. Commit.
- [ ] **8. The live check (the done-when "on a real document").** In this slot, run the round myself on copies of
  each `before.md` under `$TMPDIR` (untracked files, so a direct edit), and diff the results against `after.md`.
  Differences in wording are fine; structure must match: which asks edited, branched or got answer docs, plus the
  links and the frontmatter. Then run one tracked round on a real scratch document through Couch, to exercise the
  adapter's side-doc-then-CAS order. Record the commands and results in `## Log`.
- [ ] **9. Close.** Run the full `make test` in the background as `sdlc local-check run --name suite`, then `sdlc
  close`.

## Open points (for the TL and parley.nvim:4)

- **Q1 to Q4, answered by parley.nvim:4 (message 971f4119, 2026-10-16).** The defaults hold:
  - **Q1:** keep Version 1 and add the role `collaboration-modes`, with `examples/` not exported. Parley's #319
    sync refuses an unknown role and any other version alike, so a bump would spare it nothing. Parley adds the role
    on its side (exactly one, composed after the round and the grammar) once #319 lands.
  - **Q2:** ariadne's part is the export member and a test that the export lists it.
  - **Q3:** `process-instruction:` is copied into a side doc only when the parent has one. The session file's
    answer docs carry none.
  - **Q4:** an item's extent is its line, sub-items and continuation lines, minus the trailing answer link. A
    follow-up is any diff hunk overlapping that extent, deletions included. An edit touching only the link isn't a
    follow-up, and a deleted link makes the item a new ask again. A refused write leaves the item unlinked, and the
    next round relinks the orphan answer doc by `(parent, marker)`.

  These go into `collaboration-modes.md` (task 1) and the examples: `chat/` gains a sub-item edit (row 14) and a
  deleted-link item (a new ask), and `chat/retry/` covers a relinked orphan answer doc (row 17b for chat).
- **O1, row 10 before S3: decided by the TL (message 19edb7f1, 2026-10-16).** Under discussion, a 🤖[fold in] beside
  a link in the parent is an imperative, so M1 carries it out as an edit to the parent from the discussion. It marks
  the link "(folded <date>)", as row 7 does. S3 later adds the index and closing around it. Row 10's test-list entry
  becomes `examples/discussion/`: a 🤖[fold in] beside an existing link, with the folded text in after.md and the
  link marked.
  parley.nvim:4 agreed (e2492347; ops 7d0b425d): one fold-in, two triggers. If the discussion holds no settled
  proposal, the agent asks with a one-line clarifying turn, as for an ambiguous ok (row 8).
- **The flow: pinned to full by the TL.** The change alters how every agent answers every document, so it gets the
  plan-quality review and a real estimate.
- **The live session file.** Adding `collaboration-mode:` to `teams/devinfra/tl-status-xian-xu.md` is ops' file and
  the TL's write. The help states it, and the TL applies it once M1 lands.
