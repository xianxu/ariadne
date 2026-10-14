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

**Distribution.** `make tools` builds both binaries. `construct/dev-aliases.sh`
exposes every ariadne-owned `cmd/X`, downstream environments included.
