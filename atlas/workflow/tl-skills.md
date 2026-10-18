# TL skill binaries (tl-worker, tl-lead)

A Couch fleet is run by a TL in the hub slot (AGENTS §15). The rules for both
roles ship as **skill binaries** in the shape of `sdlc --help` and
`couch --skill`: the prose is embedded in the binary, and the skill is a
pointer stub, so the rules can't drift from the tools they describe (#350).

| Role | Binary | Prose | Stub (weave-delivered) |
|------|--------|-------|------------------------|
| worker slot | `tl-worker` (`cmd/tl-worker`) | `cmd/tl-worker/help.md` | `construct/local/tl-worker` → `xx-tl-worker` |
| TL in the hub | `tl-lead` (`cmd/tl-lead`) | `cmd/tl-lead/help.md` | `construct/local/tl-lead` → `xx-tl-lead` |

**Load path.** AGENTS §15 tells every Couch slot to run `tl-worker --help` at
the start of each dispatched item (the hub runs `tl-lead --help`). The
constitution is always loaded, and `couch --prepare` reloads a slot onto it, so
no Couch change is needed. The stubs make the same pointer discoverable by
description.

**What lives where.** The worker prose holds only rules specific to a TL-run
fleet: permissions, `make -k` past a cited known failure, real hashes, the
project file is the TL's, filing and unclaiming, live checks and
`sdlc stage`, and landing on SHIP. Rules for any Couch-fleet agent stay in
AGENTS §15, test-budget rules in §5, and messaging mechanics in
`couch --skill`. The worker prose cites them rather than restating them;
`cmd/tl-worker/main_test.go` checks both, against the rule table in
`cmd/tl-worker/testdata/rules.tsv`.

**The lead prose** is the TL's current procedure: the order of play, stewardship and the operator's two design rules, the status file, grooming trains, dispatch, watching, landing and the `sdlc stage` smoke loop, attention and load, and fleet traps. It holds rules only. Incident history stays in ops learnings, and a new kink lands as a rule change here (`cmd/tl-lead/main_test.go` refuses kink numbers, dated stories, bare `#N` refs and "until X lands" interims). It cites `tl-worker --help` for the worker's rules: `pkg/tlrules` parses the shared rule table, and the lead's test refuses any worker-owned phrase.

**Couch lifecycle notices (ariadne#433).** Couch's system notices for slot start, park, relaunch, exit, and recovery update the TL's `## Slots` view without a reply. Its restart summary replaces routine per-slot polling after a Couch restart; the TL peeks only uncertain or stalled slots. A peer message cannot become a system notice by claiming to be one. The authoritative procedure is `cmd/tl-lead/help.md` under WATCHING.

**Planning rules.** Both help surfaces carry the third-design-delta re-plan
tripwire and the owner-lifetime lock/lease question before an absence proof.
The shared body is embedded from `pkg/tlrules/planning.md` and composed into
both command helps by `tlrules.ComposeHelp`. Existing help tests phrase-guard
both outputs (ariadne#404).

**Approval and dispatch boundary.** Worker help owns small-fix self-approval,
issue-backed flake handling, Git-output hash evidence and direct same-repo
filings. Lead help owns the TL side of cross-repo proposals, work-kind dispatch
(Claude for design, Codex for fixed-design execution), landing-line ownership
of admission order and operator-agreed session retros. Both outputs retain the
shared planning tripwire; help tests guard new rules and retired central approval
wording (ariadne#406).

**`tl-lead` subcommands** replace the ops Python scripts (`ss.py`, `render-trains.py`, `opinstr.py`; `tl-status.py` was already dead and is retired, not ported):

- `tl-lead status submit|section|thread|slots|seen|owed|check|clobbered` does the TL's status-file work over `couch --session-sync`, behind the `SessionSync` seam (`cmd/tl-lead/sessionsync.go`). Couch owns the guards: the operator's lock (exit 3, passed through) and compare-and-replace by section hash (exit 6 on stale; `tl-lead` re-reads, recomputes and retries up to 4 times). `--fold-owed` folds only owed events an earlier `submit`/`owed` listed: the seen-set `.<file>.ss-seen`, bounded by the owed set. Tests run against `fakeSessionSync`, a stateful model of couch's refusals. The opt-in `TestSessionSyncConformance` (`TL_LEAD_LIVE=1`) checks the model against the real couch. A delivery gap (owed events unknown) refuses a `--fold-owed` write with the next action, while a plain write still lands. `thread` places one `###` subsection in `## Threads`: it replaces the subsection with the same heading, or goes in before `### Miscellaneous`.
- **Rendered sections.** `tl-lead status slots FILE` renders `## Slots` behind the `SDLC` seam: every numbered slot from `sdlc fleet inventory --json`, each busy slot's phase from `sdlc issue show N --json` and its bar from the checkout's issue file, then the free slots. It takes about 18 s, bounded by the inventory read. Rendered sections are rows in `renderedSections` (`cmd/tl-lead/status.go`), which is where a later `status quota` from `sdlc usage` (ariadne#365) registers. `tl-lead trains PROJECT` prints a project's trains block, for its `### Project` thread. It resolves each car's repository with `sdlc workspace <repo> --json`, then runs `sdlc issue show` per car, at most 8 at once, about 10 s in all. A car whose read failed renders `?`. The status-file layout (Operator instructions / Needs you / Quota / Threads / Slots) follows the ops draft at e7ea8e5 (ariadne#377).
- **Freshness (ariadne#414).** `writeSection`, the one write path, strips a leading `_updated YYYY-MM-DD HH:MM_` stamp before computing and puts a fresh one on what it writes (`cmd/tl-lead/stamp.go`); Operator instructions is never stamped. `tl-lead status check FILE` reads the stamps against `sectionBudgets` (`cmd/tl-lead/freshness.go`): Threads and Slots 2 h plus the landing rule (stale once any sibling SDLC repository's newest first-parent merge on local `origin/main`, brains excluded, is newer; direct filing, publish, archive and tick commits do not count, and stale output identifies the repository, merge hash and time), Quota 12 h, Needs you only when unstamped. `submit`, `owed` and every write end with the same lines. The stamp is the only state: it lives in the section it describes. `## Slots` also reads `couch --actors --json` (the `CouchActors` seam, `cmd/tl-lead/actors.go`) for each slot's agent and idleness; a failed read renders `?`.
- **Operator instructions.** `tl-lead status instructions list FILE` prints the open bullets of `## Operator instructions` (chat mode's items: `-`, `*` and `N.` lines at the shallowest indent, #412) by position (the operator reuses numbers) and the section's hash; `link FILE POS TARGET --hash H` appends ` → [answer](TARGET)` to that bullet's last non-blank line. The parser and linker are pure (`cmd/tl-lead/instructions.go`). `link` is the one write that never retries: a section whose hash moved since `list` refuses with exit 6, since a position from the older read could name another bullet. The section is a chat section of xx-universal-review (ariadne#412): the status file declares `collaboration-mode: {chat: ["Operator instructions"]}`, the answers are dated answer docs in the gitignored `teams/<team>/tl-status-<operator>.side/` (older bullets keep their `qa/<operator>/` links), and the rest of the file is discussion. A bullet the operator deletes leaves its answer doc in place (ariadne#424: side docs have no closed state); `tl-lead --help` says where a later follow-up goes.
- **The discussions: index (ariadne#424).** The status file's frontmatter `discussions:` lists the side docs waiting for the operator's review, by the one-way rule of xx-universal-review: deleting a link drops its entry, removing an entry means reviewed. `instructions link` writes the answer doc's entry first and the link second (a split leaves an entry without a link, which a rerun completes and a reconcile otherwise drops). `submit` reconciles on each submit, and `status index FILE [--add DOC]` reconciles or lists a side doc linked elsewhere. The write is couch's whole-file `write --all --if-hash` (`ReadAll`/`WriteAll` on the seam, modeled in the fake and in `TestSessionSyncConformance`), and the rule itself is `pkg/sidedoc.ReconcileIndex`, shared with the review skill's checker. Couch's whole-file read and write mark every edit seen, so the index never consumes the operator's unread edits: whether anything changes is decided by `Peek` (the file on disk, baseline untouched), and a needed write first prints the pending diff (`unreadChanges`). The fake models couch's seen-baseline, and the conformance test pins that `read --all` empties the diff.
- **One parser.** The trains table is parsed by `pkg/trains`, and the bar is counted by `pkg/markdoc`'s plan scanners. Both moved out of `cmd/sdlc/internal` (#350), and the old names forward to them, so the TL's bar counts what the close gate counts.

**Distribution.** `make tools` builds both binaries. `construct/dev-aliases.sh`
exposes every ariadne-owned `cmd/X`, downstream environments included.
