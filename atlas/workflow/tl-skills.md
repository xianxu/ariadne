# TL skill binaries (tl-worker, tl-lead)

## Workingdoc settlement (#507)

`tl-lead doc settle SIDE-DOC [--to PATH]` preserves the live editing surface and
publishes one tracked copy, defaulting to `workshop/workingdoc/<date>-<slug>.md`.
Its managed source pointer pins the target and full commit; re-settlement uses
that same path. `doc status` compares a shared pure projection with committed
blobs, distinguishing drift, destination conflicts and unknown evidence.

`doc_model.go`, `doc_project.go` and `doc_markers.go` own metadata, content and
marker validation. Literal masks have one shared owner, `sidedoc.CodeLiteralMask`, used by marker
validation, link rebasing and side-link discovery. They come from Goldmark's actual Markdown code nodes;
unmatched backticks across paragraphs or containers cannot hide active questions.
The counted linear work covers our input, AST walk, masks and scanner. Goldmark's
code-span parser can rescan pathological input; the existing 1 MiB document cap
bounds accepted input, rather than a claim that dependency internals are linear.
The phase model returns the next permitted publication/stamp effect, and both
external calls require that permission before executing; failed or unconfirmed
outcomes stop further effects. Remaining active questions, proposals and malformed markers
refuse; the operator's editor owns acceptance. `doc_settle.go` coordinates a
repository lock, source/target guards, immutable publication and Couch's source
CAS stamp. `doc_repository.go` validates repository identity and batches Git
blob reads. Products reuse the SDLC trunk publisher; verified ops primary-main
publication is `doc_operating.go`, with a private Git index and scoped child
processes. The real filesystem, bare-remotes and stateful Couch/repository fakes
exercise interruptions and late edits through the production seams.

Publishing and stamping have separate outcomes. A confirmed copy survives a
locked/stale source; an uncertain response is never absence. Reconcile the named
source/target and retry the same command. Independent destination edits refuse.
The copy is durable knowledge removed through ordinary repository edits; temporary
payload/index resources end at command return. The constant-size common-directory
lock inode remains so concurrent callers cannot lock different inodes.

The [workingdoc datatype](../../construct/datatype/workingdoc.md) supplies the
authoring contract; `tl-lead --help` supplies the operator procedure.

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

**Hub takeover (ariadne#528).** TAKEOVER in `cmd/tl-lead/help.md` orders recovery
after a forced switch or fresh hub session. The session's Threads stamp bounds
any transcript gap; live slots, landing/queue state and project files establish
the current picture before a status-writer refresh and one-line operator notice.
`cmd/tl-lead/takeover_help_test.go` guards both rendered entry points and keeps
each instruction inside its ordered step.

**External dependency incidents (ariadne#540).** TALKING TO THE OPERATOR in
`cmd/tl-lead/help.md` owns immediate escalation: the incident goes first in
Needs you with one desktop notification before diagnosis, a known operator
remedy includes its exact command, and verified recovery gets a second notice
and affected-slot checks. This is an explicit exception to the ask-only
Needs-you rule. Worker REPORTS owns reporting the incident to the hub while
preserving operation recovery state. Both help contracts are tested across
the shared rendered instruction surface. Diagnostic observations come from
[`sdlc net check`](sdlc-binary.md#network-incidents); no automated outage
watcher or remote-write retry is introduced.

**Worker verification (ariadne#508).** TESTS in `cmd/tl-worker/help.md` owns
verification across every agent kind in tests and live checks, recorded real
capture fixtures replayed in the normal suite with their live bug fixes, and
reuse of the shared disposable-Couch harness (pair#744) instead of the live
fleet or private fixtures. Rendered-section and shared rule-table tests pin
the worker surface; the lead continues citing it. Pair#744's landed harness is
`cmd/internal/testcouch`, exercised with `make test-couch-fleet` in Pair.

**Project shaping (ariadne#518).** SHAPING in `cmd/tl-lead/help.md` precedes
issue-level execution: six operator-reviewed steps cover users/moments, flows,
the clickable mock, visual design, components and slices. It assigns the driver
of each step, versioned mock/sketch conventions, `couch mock serve`, shared
fixtures and the development-only screen-by-screen gradient with a production
mock refusal. The optional project summaries are defined by
[`construct/datatype/project.md`](../../construct/datatype/project.md#optional-shaping-sections);
the TL keeps detailed design and mock artifacts in the project folder.
`cmd/tl-lead/shaping_help_test.go` checks the rendered procedure at both entry
points, including step-local ownership, order, the feedback-round/alternative
version distinction, and shipping boundary.

**Operator design review (ariadne#505).** STEWARDSHIP in `cmd/tl-lead/help.md`
routes failure-sensitive workflow designs to the operator before code. It names
the recognition shapes and examples, the options/state/recovery/failure contents
of the side doc, and the boundary where ordinary TL approval still applies.
Rendered-help tests pin that boundary and reject the former taste/smoke-only
project summary.

**Couch lifecycle notices (ariadne#433).** Couch's system notices for slot start, park, relaunch, recovery, needs-relaunch, and restart summary update the TL's `## Slots` view without a reply. The restart summary replaces routine per-slot polling after a Couch restart; the TL peeks only uncertain or stalled slots. A peer message cannot become a system notice by claiming to be one. The authoritative procedure is `cmd/tl-lead/help.md` under WATCHING.

**Planning rules.** Both help surfaces carry the third-design-delta re-plan
tripwire and the owner-lifetime lock/lease question before an absence proof.
The shared body is embedded from `pkg/tlrules/planning.md` and composed into
both command helps by `tlrules.ComposeHelp`. Existing help tests phrase-guard
both outputs (ariadne#404).
For state that a user or automation can act on, that shared body also requires
an action and failure-mode table: internal cases map to a few visible states,
each with one recovery and stated possible loss; the tradeoff goes to the TL,
then to the operator when user-facing (ariadne#410).
For hot paths, the same shared rule asks for expected operations per growth
driver and a counted test at N and 2N rather than timing alone (ariadne#401).

**Approval and dispatch boundary.** Worker help owns small-fix self-approval,
issue-backed flake handling, Git-output hash evidence and direct same-repo
filings. Lead help owns the TL side of cross-repo proposals, work-kind dispatch
(Claude for design, Codex for fixed-design execution), landing-line ownership
of admission order and operator-agreed session retros. Both outputs retain the
shared planning tripwire; help tests guard new rules and retired central approval
wording (ariadne#406).

**`tl-lead` subcommands** replace the ops Python scripts (`ss.py`, `render-trains.py`, `opinstr.py`; `tl-status.py` was already dead and is retired, not ported):

- `tl-lead status submit|section|thread|slots|seen|owed|check|clobbered` does the TL's status-file work over `couch --session-sync`, behind the `SessionSync` seam (`cmd/tl-lead/sessionsync.go`). Couch owns the guards: the operator's lock (exit 3, passed through) and compare-and-replace by section hash (exit 6 on stale; `tl-lead` re-reads, recomputes and retries up to 4 times). `--fold-owed` folds only owed events an earlier `submit`/`owed` listed: the seen-set `.<file>.ss-seen`, bounded by the owed set. Tests run against `fakeSessionSync`, a stateful model of couch's refusals. The opt-in `TestSessionSyncConformance` (`TL_LEAD_LIVE=1`) checks the model against the real couch. A delivery gap (owed events unknown) refuses a `--fold-owed` write with the next action, while a plain write still lands. `thread` places one `###` subsection in `## Threads`: it replaces the subsection with the same heading, or goes in before `### Miscellaneous`.
- **Rendered sections.** `tl-lead status slots FILE` renders `## Slots` behind the `SDLC` seam: hub-owned numbered slots from `sdlc fleet inventory --json` by default (`--all` shows the machine view; no `.hub` keeps legacy scope), each busy slot's phase from `sdlc issue show N --json` and its bar from the checkout's issue file, then the free slots. It takes about 18 s, bounded by the inventory read. Rendered sections are rows in `renderedSections` (`cmd/tl-lead/status.go`), which is where a later `status quota` from `sdlc usage` (ariadne#365) registers. `tl-lead trains PROJECT` prints a project's trains block, for its `### Project` thread. It resolves each car's repository with `sdlc workspace <repo> --json`, then runs `sdlc issue show` per car, at most 8 at once, about 10 s in all. A car whose read failed renders `?`. The status-file layout (Operator instructions / Needs you / Quota / Threads / Slots) follows the ops draft at e7ea8e5 (ariadne#377).
- **Answer migration (ariadne#455).** `tl-lead status migrate-answers FILE [--apply]` moves the older `qa/<operator>/` answers into the status file's side folder once. `planMigration` (`cmd/tl-lead/migrate.go`) is pure over a snapshot. It classifies each entry (docs and attachments move, dot paths stay, a `.md.lock`, `.md.clobbered` or symlink is a conflict), gives answers the side-doc frontmatter and backlink, and rebases every relative link with `sidedoc.RebaseLinks` (`pkg/sidedoc/rebase.go`). The executor (`migrate_apply.go`) creates each file exclusively, keeps its mtime, moves couch's registration (`SessionSync.Rename`), then removes the old file, so a rerun resumes from any stop. The session file's links and index change in one `SessionSync.Apply`, whose hash is `fileHash` of a Peek; the formula is pinned by live conformance. Bare mentions are listed, not rewritten.
- **Teams (ariadne#456).** A team is an ops repo whose `:0` slot is the TL. A slot whose checkout has `.hub` at its root runs `tl-lead`; every other slot runs `tl-worker`. `.hub`, Couch's file, is the one ownership list: claim lines only, a repo or `repo:N`. tl-lead's copy of the grammar (`cmd/tl-lead/team.go`, `parseHub`) is pinned by pair's own test vectors. `ownerOf` mirrors Couch's `Resolve`: a hub owns its `:0`, and an exact claim beats a repo claim. A tie or an unclaimed slot is reported as such, because the command-line order that settles it is invisible to tl-lead; so are Couch's runtime moves. Four verbs:
  - `team new PATH --claim C` creates or adopts an ops repo with `.hub` and a `team.md` that points at it. It can be rerun, refuses a file that differs, and never commits or pushes (`team_cmd.go`).
  - `team list [REPO]` reads the sibling repos that have `.hub`.
  - `team issues` lists open issues nobody owns in the repos `.hub` claims, through `sdlc state --json` (`SDLC.RepoIssues`, `team_issues.go`).
  - `fleet resume [--apply] [--all]` (`fleet.go`) joins `couch --actors`, sdlc claims and `couch --resume-candidates` (the `CouchResume` seam). The pure `planResume` classes each slot this hub owns alone as running, idle, resume, ambiguous or unknown. `--apply` is operator-started, never automatic at startup. It resumes the eligible class serially, awaiting each terminal JSON receipt through Couch's existing poller before proceeding. `resume_result.go` classifies matching terminal receipts with consistent normal exits; missing/mismatched receipts, signal death and timeouts remain uncertain. Its six-minute outer bound covers Couch's five-minute budget plus admission/teardown. The first unsuccessful result stops with nonzero status, retained diagnostics, earlier successes and the exact untouched list. The stopped/untouched result and inspection guidance precede uncertainty's optional actor observation (10 s plus at most 1 s pipe cleanup); it never proves success or triggers a retry. Inspect the receipt/slot before further manual action. Unclaimed/foreign rows are counted by default; `--all` exposes them without widening mutation authority. A rerun skips live slots.
The steward is the first hub on Couch's command line.
- **Keep the conversation, soak holds (ariadne#482).** `tl-lead --help` DISPATCH gives a just-landed slot the next same-area car without `couch --prepare`, which would start a fresh conversation. LANDING lets the TL hold a slot to soak a risky change. The hold lives in the rendered `## Slots` section: `status slots FILE --hold repo:N='until X'` / `--release repo:N`. `parseHolds` and `applyHolds` (`cmd/tl-lead/slots.go`) read it back from the section being replaced, at write time, so every render carries it, and a held slot is never on the free line. A scoped render that would hide an existing hold refuses unless explicitly released; `--all` retains the full view.
- **Freshness (ariadne#414).** `writeSection`, the one write path, strips a leading `_updated YYYY-MM-DD HH:MM_` stamp before computing and puts a fresh one on what it writes (`cmd/tl-lead/stamp.go`); Operator instructions is never stamped. `tl-lead status check FILE` reads the stamps against `sectionBudgets` (`cmd/tl-lead/freshness.go`): Threads and Slots 2 h plus the landing rule (stale once a hub-claimed SDLC repository's newest first-parent merge on local `origin/main`, brains excluded, is newer; direct filing, publish, archive and tick commits do not count, and stale output identifies the repository, merge hash and time), Quota 12 h, Needs you only when unstamped. `submit`, `owed` and every write end with the same lines. The stamp is the only state: it lives in the section it describes. `## Slots` also reads `couch --actors --json` (the `CouchActors` seam, `cmd/tl-lead/actors.go`) for each slot's agent and idleness; a failed read renders `?` with one diagnostic. `scope.go` shares `.hub` ownership across Slots, resume planning and freshness; exact-slot claims include their repository for freshness, nested claimed paths are followed, and no-hub freshness retains sibling discovery. Invalid scope refuses/marks freshness unknown rather than widening the view.
- **Operator instructions.** `tl-lead status instructions list FILE` prints the open bullets of `## Operator instructions` (chat items by `pkg/sidedoc.ChatItems`: `N.` lines at the shallowest indent for a `@@chat@@`-tagged section, also `-` and `*` for a legacy-declared or undeclared one, chosen from a `Peek` of the file, #412, #576) by position (the operator reuses numbers) and a `HASH:GRAMMAR` token, the section's hash and the item grammar the positions were counted in; `link FILE POS TARGET --hash TOKEN` appends ` → [answer](TARGET)` to that bullet's last non-blank line. The parser and linker are pure (`cmd/tl-lead/instructions.go`). `link` is the one write that never retries: a section whose hash moved since `list` refuses with exit 6, since a position from the older read could name another bullet; it refuses the same way when a `@@chat@@` tag or legacy declaration outside the hashed section changed the grammar (#576). The section is a chat section of xx-universal-review (ariadne#412): the status file puts `@@chat@@` above it (#576; older files' heading comment or map still reads as chat), the answers are dated answer docs in the gitignored `session.side/`, and the rest of the file is ordinary collaborative editing. `link` validates its target (ariadne#455, `cmd/tl-lead/answers.go`): a dated side doc of the status file, already written, whose `parent:` names it back, so every answer shows in parley's attention finder. A bullet the operator deletes leaves its answer doc in place (ariadne#424: side docs have no closed state); `tl-lead --help` says where a later follow-up goes.
- **The discussions: index (ariadne#424).** The status file's frontmatter `discussions:` lists the side docs waiting for the operator's review, by the one-way rule of xx-universal-review: deleting a link drops its entry, removing an entry means reviewed. `instructions link` writes the answer doc's entry first and the link second (a split leaves an entry without a link, which a rerun completes and a reconcile otherwise drops). `submit` reconciles on each submit, and `status index FILE [--add DOC]` reconciles or lists a side doc linked elsewhere. The write is couch's frontmatter scope (ariadne#426, pair#643): `read --frontmatter` and `write --frontmatter --if-hash` (`ReadFrontmatter`/`WriteFrontmatter` on the seam). Both mark only the frontmatter seen, so the index never consumes or prints the operator's unread section edits. The body's links are read by `Peek` (the file on disk, baseline untouched), which also decides a no-op without any couch read. `pkg/sidedoc.FrontBlock` splits a document the way couch projects it, and the rule itself is `pkg/sidedoc.ReconcileIndex`, shared with the review skill's checker. The CAS covers only the frontmatter, so a body edit after the Peek can only delay a drop to the next reconcile. The fake models the frontmatter-only baseline; `TestSessionSyncConformance` (against a scratch file) pins that a frontmatter read or write leaves a section edit in the diff and keeps every body byte.
- **One parser.** The trains table is parsed by `pkg/trains`, and the bar is counted by `pkg/markdoc`'s plan scanners. Both moved out of `cmd/sdlc/internal` (#350), and the old names forward to them, so the TL's bar counts what the close gate counts.

**Distribution.** `make tools` builds both binaries. `construct/dev-aliases.sh`
exposes every ariadne-owned `cmd/X`, downstream environments included.

### Resume receipt conformance

When closing a change to resume and whenever upgrading the supported Couch
binary, run `TL_LEAD_RESUME_CONFORMANCE=1 go test ./cmd/tl-lead -run '^TestCouchResumeReceiptConformance$' -count=1 -v`.
The real CLI uses an isolated scratch singleton selection and a private stateful
broker. The check covers queued/running polling, standalone terminal JSON
receipts, exit mapping and our classifier without contacting live slots or
starting a GUI. Default test runs skip this opt-in check.

**Root hub session (ariadne#504).** `session.md` and `session.side/` are
local, gitignored operator/TL conversation. Decisions enter tracked issues,
project Logs or charters immediately. Existing explicit FILE forms remain;
`withDefaultStatusFile` (`cmd/tl-lead/status_path.go`) supplies the hub Git
root's session.md for omitted FILE. `status move [OLD] [--apply]` discovers
one legacy file or accepts OLD. It plans pure transformations in
`status_move.go`, executes copy/switch/verify/cleanup in `status_move_apply.go`,
and uses typed batched `RegistrationSnapshot` list replies plus Couch `Rename` to preserve
identity, unread baselines and owed events. Source sessions require prior
registration so Couch imports their legacy metadata; tl-lead leaves those
companions in place. Dry runs write nothing. A quiet window with editors
closed is required: observed locks refuse, but there is no editor lease.
Copies are exclusive and identical copies are reusable after interruption;
changed copies refuse before switching. Confirmed root switching authorizes source-only cleanup; live destination edits and locks are allowed. Retry confirms the prior root rename and never overwrites live copies. The old tree must remain quiet between retries. Distinct registration conflicts refuse without disposal.
The old session stays in use before switching; afterwards the new session
works while cleanup can be retried. No journal is created.

`pkg/sidedoc.RebaseIndex` accepts old/new parent coordinates and rewrites only
validated discussions path scalars, preserving unknown metadata bytes. The
older answer migration delegates stationary index rebasing to it. Parley's
attention finder derives `.side/` from the registered parent path, so root
session.side/ works without a peer change; nested side-doc discovery retains
its existing one-level-per-open behavior and migrated mtimes.

Migration metadata validation accepts only plain parent scalars matching the
syntax emitted by tl-lead; every broader parent form refuses before effects.
Session routing fields are validated only for session-sync documents, so
ordinary documents retain business owner metadata. Missing ignore lines are
appended without rewriting existing policy; `verifyMoveIgnores` runs real Git
checks over every planned and existing retry destination before copies. A later
negation causes refusal with sources intact; only ignore additions may remain.

Session migration refuses `.gitignore` files anywhere in the old or new side tree before effects. It rechecks effective Git protection before switching registrations and before each source deletion.

Migration Couch calls have a five-second budget with bounded output draining.
Read failures report unreachable Couch before local effects; uncertain rename
outcomes preserve sources. Registration observations use fixed-count list
snapshots, not per-document lookup/status. Git policy uses complete batched
membership checks before copy/switch and one direct check per deletion. Counted
8/16-document tests bound process growth; real scratch conformance interrupts
after switching, edits the live root and retries cleanup.

## Drain and hub automation

The worker's DRAIN section in `cmd/tl-worker/help.md` owns graceful/fast background
work, durable checkpoints, deferral of new work messages and continuation after a
lift. Graceful waits only for work already running; queued work remains queued
and is logged as outstanding. There is no formal drain-ack command or parsed completion reply. Couch owns
authenticated notice text and receipts; custom notes are data. Restart-lost flags
never grant recovery authority.

DRAIN THE FLEET in `cmd/tl-lead/help.md` owns the steward's separate owner-scoped
queue controls and restart-safety judgment. Paused queued work stays outstanding;
a flag or quiet screen is not proof of a durable checkpoint or completed work.
Recovery waits for the operator's explicit instruction and uses the existing
owned-slot resume procedure. Where Couch provides hub automation controls, its
first genuine operator message can enable that hub unless told to keep it off;
automation does not grant recovery authority.

`cmd/tl-worker/drain_test.go` and `cmd/tl-lead/drain_test.go` guard both
rendered help entry points. Source-backed canonical notice vectors and the
recorded stale real-CLI DRAIN excerpt live in `cmd/tl-worker/testdata/drain/`;
these are replay data, not a live drain or a private Couch fleet harness.
