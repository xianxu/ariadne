# #377: tl-lead catches up with the ops TL draft — plan

**Goal:** `tl-lead --help` carries every rule the ops draft (`ops/skills-draft/xx-tl/SKILL.md`, at ops e7ea8e5) has that #350 didn't port, and `opinstr.py` becomes `tl-lead status instructions`, so ops can point at `tl-lead --help` and delete its draft and its Python.

**Source read:** the draft at ops e7ea8e5 (13:08, the last commit touching `skills-draft/`), compared section by section with `cmd/tl-lead/help.md` at main 18915f6c.

## The gap (the enumeration, ARCH-PURPOSE)

Each row is one rule; the column says where it lands. A rule the draft has that tl-lead already states is not listed.

| # | Rule in the draft | tl-lead today | Lands in |
|---|---|---|---|
| G1 | Section order: `## Operator instructions` is section 0, before Needs you | four sections, starting at Needs you | help.md, the sections list |
| G2 | Operator instructions belong to the operator: numbered questions and instructions; the TL never edits, deletes or marks them; it only appends `→ [answer](qa/…)` to every bullet, an instruction's file recording what was done and where; the operator `dd`s a bullet when satisfied; an unlinked bullet is open | missing | help.md |
| G3 | QA files: `teams/<team>/qa/<operator>/YYYY-MM-DD-NN-<slug>.md`, gitignored; frontmatter `process-instruction: "[free-form] answer the operator's 🤖 comments in place"`; the question as the title; a plain relative backlink to the status file, no `#anchor` (parley's Alt+o falls back to gf, which reads `#` as part of the filename); an answer file answers only the question asked | missing | help.md |
| G4 | Needs you holds only asks (a decision or an action, with the TL's recommendation); no status there, status goes to Threads; remove an item the moment it's done | "decisions and actions … queue several" (no "only asks", no removal rule) | help.md, the sections list |
| G5 | Glance at CI on main every pass (`gh run list --repo <owner>/<repo> --branch main --limit 5` per repo); a red main goes into the status file and gets an issue and a slot that pass | missing | help.md, ATTENTION AND LOAD |
| G6 | Render sections from binaries wherever possible; hand-write only judgment (Operator instructions, Needs you, the why of a thread), taking even its facts from `sdlc`; a fact retyped every pass is the next piece to lift | "Render from sources, never by hand" covers Slots and trains only | help.md, generalize that paragraph |
| G7 | Quiet pane: input typed in the draft pane gets a normal reply in the agent pane; input from the status file or a QA file is answered in files, at most one line in the pane; routine worker reports go to the status file and digest; the pane carries decisions and "act now" items | missing | help.md, TALKING TO THE OPERATOR |
| G8 | A lock with no submit is never overridden; mention in chat that updates are waiting on it | "never override a lock" (no mention rule) | help.md, the session-sync paragraph |
| G9 | Parse Operator instructions by position, multi-line bullets, answered by `→ [answer](` or `→ done:` | `opinstr.py`, no Go port | `tl-lead status instructions list|link` (below) |

Rows in the draft deliberately **not** ported, with the reason:
- "Until parley.nvim#315 lands: delete a `holder: agent` lock yourself": #315 landed (parley.nvim PR #228), so the workaround is obsolete.
- `ss.py` / `tl-status.py` / `render-trains.py` usage lines: already replaced by `tl-lead status` and `tl-lead trains` (#350).
- Stories and dates ("operator, 10-14", "parley.nvim's merge-check was red for 10 hours"): rules only, per the issue's Spec.

**The two contradictions:**
- **Frontmatter key.** `couch --skill` today says: "frontmatter `type: session-sync` and `couch-owner: <your slot>` (older files say `owner:`, still read when `couch-owner:` is absent)". So couch's canonical key is `couch-owner:`, and the draft's `owner:` is the legacy spelling. tl-lead keeps `couch-owner:` and adds that `owner:` is still read; ops's status file may switch at leisure. (This reverses the issue's phrasing, which assumed `owner:` was what couch reads; the evidence is couch's own skill text.)
- **Section order:** fixed by G1.

**Two rules named in the dispatch but not in the draft:** "slot verification after disruption" and "issue filing from a resting checkout". They are not in `SKILL.md` at e7ea8e5 or in ops's other markdown; I asked the hub for their source text and will port them as rows G10/G11 when it arrives (Revisions entry).

## `tl-lead status instructions` (G9)

**Pure core** (`cmd/tl-lead/instructions.go`, ARCH-PURE):
- `instructionBullets(section string) []Bullet` — a bullet starts at a line `N. ` (N digits) and runs to the next such line or the end of the section; trailing blank lines are dropped; lines before the first bullet are not a bullet. `Bullet{Pos, Lines, Answered}`; `Pos` is 1-based by position (numbers repeat). Answered = the bullet's text matches `→\s*(\[answer\]\(|done:)`.
- `linkBullet(section string, pos int, target string) (string, error)` — appends ` → [answer](TARGET)` to bullet `pos`'s last non-blank line; refuses a missing position; an answered bullet is a no-op reported as such.

It works on the section body (`couch --session-sync read --section`), not the whole file: the Python's `## ` boundary scan becomes the section read couch already does (ARCH-DRY).

**Verbs** (thin shell over the existing seam, `status.go`):
- `tl-lead status instructions list FILE` — reads the section through `SessionSync.ReadSection`, prints the open bullets (`POS<TAB>first line`, continuation lines tab-indented, as `opinstr.py` does), then a last line `hash: H`.
- `tl-lead status instructions link FILE POS TARGET --hash H` — reads the section; if its hash isn't `H`, refuses ("Operator instructions changed since list; list again"), since positions from an older read may name a different bullet; otherwise writes `linkBullet`'s result through the guarded write path. An operator lock refuses as every write does (exit 3). This is the one write that does **not** retry on stale: a recompute against a newer read would silently re-target the position (ARCH-ORDER: the event most likely to be mishandled is the operator deleting a bullet between list and link; the hash pins the read the position came from).

**Tests** (`instructions_test.go`, table-driven golden cases, plus the fake for the verbs):
- multi-line bullets; blank lines inside a bullet (kept) and trailing ones (dropped); repeated numbers (`1. 1. 2.` → positions 1, 2, 3); image lines (`![](assets/x.png)` continuation lines stay in the bullet, and the link goes after the last non-blank line even when it's an image line, as opinstr does); `→ done:` and `→ [answer](…)` both mark answered; preamble lines; an empty section.
- `link`: appends to the right bullet; already answered → no-op, no write; missing position → error; hash mismatch → refused with no write; lock → exit 3 (the stateful `fakeSessionSync`, which already models hash and lock).
- The live conformance test (`TL_LEAD_LIVE=1`) gains nothing: these verbs use only `ReadSection`/`WriteSection`, which it already covers.

## Envelope and residue

- **ARCH-CONSTRAINTS:** the section is operator-typed, tens of lines; one couch read and at most one write per call (milliseconds). N/A beyond that.
- **ARCH-SECURE:** the section is untrusted operator input; it is only scanned as lines and re-emitted, never executed; `TARGET` is TL-supplied and written verbatim into the file (the TL's own link).
- **ARCH-FUNERAL:** creates nothing durable beyond the existing temporary body file, which `writeBody` removes. QA files are the TL's and gitignored; their end is the operator's `dd` of the bullet plus ops's cleanup, outside this binary.
- **ARCH-MOCK:** no new external surface; the verbs go through the existing `SessionSync` seam and its stateful fake.

## Plan

Single pass, one close (no milestones).

- [ ] help.md: G1–G8 into the sections named above, rules only; the frontmatter key line; the instructions verbs in the session-sync paragraph.
- [ ] `instructions.go` + `instructions_test.go`: the pure parser and linker with the golden cases.
- [ ] `status instructions list|link` verbs + their fake-backed tests.
- [ ] A help test that every new verb is named in help.md (the existing verb-listing test, extended if it doesn't already derive from the command tree).
- [ ] Atlas `atlas/workflow/tl-skills.md`: the new verbs and the section order.
- [ ] G10/G11 once the hub sends their source.
- [ ] Fresh comparison: the draft at e7ea8e5 against the new help.md, row by row, recorded in `## Log`.

**Verification:** `go test ./cmd/tl-lead/...`; `tl-lead status instructions list` against a copy of ops's real status file via the fake (no write to the real one); the full `make test` before close.
