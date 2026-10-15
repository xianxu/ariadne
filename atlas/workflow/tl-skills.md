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

**`tl-lead` subcommands** replace the ops Python scripts (`ss.py`, `render-trains.py`, `opinstr.py`; `tl-status.py` was already dead and is retired, not ported):

- `tl-lead status submit|section|thread|slots|seen|owed|clobbered` does the TL's status-file work over `couch --session-sync`, behind the `SessionSync` seam (`cmd/tl-lead/sessionsync.go`). Couch owns the guards: the operator's lock (exit 3, passed through) and compare-and-replace by section hash (exit 6 on stale; `tl-lead` re-reads, recomputes and retries up to 4 times). `--fold-owed` folds only owed events an earlier `submit`/`owed` listed: the seen-set `.<file>.ss-seen`, bounded by the owed set. Tests run against `fakeSessionSync`, a stateful model of couch's refusals. The opt-in `TestSessionSyncConformance` (`TL_LEAD_LIVE=1`) checks the model against the real couch. A delivery gap (owed events unknown) refuses a `--fold-owed` write with the next action, while a plain write still lands. `thread` places one `###` subsection in `## Threads`: it replaces the subsection with the same heading, or goes in before `### Miscellaneous`.
- **Rendered sections.** `tl-lead status slots FILE` renders `## Slots` behind the `SDLC` seam: every numbered slot from `sdlc fleet inventory --json`, each busy slot's phase from `sdlc issue show N --json` and its bar from the checkout's issue file, then the free slots. It takes about 18 s, bounded by the inventory read. Rendered sections are rows in `renderedSections` (`cmd/tl-lead/status.go`), which is where a later `status quota` from `sdlc usage` (ariadne#365) registers. `tl-lead trains PROJECT` prints a project's trains block, for its `### Project` thread. It resolves each car's repository with `sdlc workspace <repo> --json`, then runs `sdlc issue show` per car, at most 8 at once, about 10 s in all. A car whose read failed renders `?`. The status-file layout (Operator instructions / Needs you / Quota / Threads / Slots) follows the ops draft at e7ea8e5 (ariadne#377).
- **Operator instructions.** `tl-lead status instructions list FILE` prints the open bullets of `## Operator instructions` by position (the operator reuses numbers) and the section's hash; `link FILE POS TARGET --hash H` appends ` → [answer](TARGET)` to that bullet's last non-blank line. The parser and linker are pure (`cmd/tl-lead/instructions.go`). `link` is the one write that never retries: a section whose hash moved since `list` refuses with exit 6, since a position from the older read could name another bullet. The answers live in gitignored QA files under `teams/<team>/qa/<operator>/` in ops. A bullet the operator deletes closes its QA file, and no follow-up goes into that file; `tl-lead --help` says where follow-ups go (ariadne#388).
- **One parser.** The trains table is parsed by `pkg/trains`, and the bar is counted by `pkg/markdoc`'s plan scanners. Both moved out of `cmd/sdlc/internal` (#350), and the old names forward to them, so the TL's bar counts what the close gate counts.

**Distribution.** `make tools` builds both binaries. `construct/dev-aliases.sh`
exposes every ariadne-owned `cmd/X`, downstream environments included.
