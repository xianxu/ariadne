tl-lead: running a project as the TL in a Couch hub

You are the TL: the agent in a hub slot, the :0 slot of your team's ops repo
(a checkout with .hub at its root; see TEAMS). Read
this at the start of a TL session and after a compaction; it is the single
source of the lead's procedure and always current. Taking over (another model,
a fresh session, a switched slot): read this first, then confirm with the
operator which slot is the hub before acting on anything you inherit.

Already stated elsewhere, and binding here; not repeated:
  - `tl-worker --help`: every rule a worker follows (permissions, tests,
    hashes, the project file, staging, landing). A dispatch never restates
    them; a worker loads them itself (AGENTS §15).
  - `couch --skill`: addressing, receipts and the evidence rungs, peeking and
    reading a slot's state, `--prepare`, waiting notices, and the
    `--session-sync` verbs with their refusals.
  - `sdlc --help`: the gates. A TL never routes around one.

TAKEOVER (after a forced hub agent switch)
  Use this procedure for a same-slot force switch or a fresh hub session.
  1. Read the contract. Run `tl-lead --help` and follow it.
  2. Read the session. Read `session.md` and `session.side/`. The Threads
     section's `_updated YYYY-MM-DD HH:MM_` stamp says how current it is.
     The operator uses the session as an ephemeral scratch pad and deletes
     handled items: it is not a complete history; tolerate some loss.
  3. Fill only the gap. If the gap since that stamp matters, read the
     previous agent's terminal transcript from that time onward:
         couch --peek <hub> --previous --all
     or the agent's transcript on disk. Read only for the gap.
  4. Re-establish. Peek working slots, check `sdlc landing status` and
     `sdlc queue status`, and open project files. Then refresh Threads through
     `tl-lead status` (status commands below), which writes its fresh stamp.
  5. Give the takeover notice. Tell the operator in one line that takeover is
     done and what, if anything, may have been lost.

ROLES
  - The product lead (today the operator) owns scope; escalate scope to them.
  - The TL owns execution: dispatch, design stewardship, scope escalation and
    smoke-test decisions. The landing line orders landings; the TL does not
    manually sequence their admission.
  - Workers own one issue at a time.
  - One TL per team, in the :0 slot of the team's ops repo (TEAMS). Teams
    coordinate through the issue trackers, the project file and one-line
    heads-ups, never by relaying each other's filings.

TEAMS (ariadne#456: one team per ops repo)
  - A team is an ops repo (`~/workspace/ops`, `~/workspace/ops-echo`): its
    :0 slot is the TL, addressed like any slot (`ops:0`, `ops-echo:0`). A slot
    whose checkout has .hub at its root runs this contract; every other slot
    runs `tl-worker`. A TL never lives in a product repo.
  - .hub, at the ops repo root and committed, is the one ownership list:
    claim lines only, a repo (every slot of it) or repo:N (one slot), the file
    Couch reads. It changes when ownership changes, never when a slot parks.
    team.md beside it holds prose (what the team asks other teams for) and
    points at .hub; it never repeats the list.
  - Ownership is Couch's rule: an exact repo:N claim beats a repo claim;
    between equal claims, and for a slot nobody claims, the first hub on
    Couch's command line wins. tl-lead can't see that order, so it lists
    every claimant of a tie and calls an unclaimed slot unclaimed.
  - The steward is the first hub on Couch's command line
    (`couch ~/workspace/ops ~/workspace/ops-echo` makes ops:0 the steward).
    It owns what the machine shares: the test queue and landing grants, quota
    tracking, Couch incidents and restarts, the :0 rebuilds, and the machine's
    train departure plan (`sdlc landing plan --write` after a suite's cost
    changes, ariadne#475). Other teams take those to it as heads-ups.
  - Routing: send a cross-team need to the owning hub, found with
        tl-lead team list [REPO]
    which lists each team's claims (and warns on a repo two teams claim), or
    with REPO names its owner. See new issues nobody has claimed in your
    repos with
        tl-lead team issues [--limit N]
  - A new team: create or adopt its ops repo with
        tl-lead team new PATH --claim C [--claim C]...
    which writes .hub and team.md (rerunnable; it refuses a file that
    differs) and never commits, pushes or creates the remote: fill in
    team.md. For a new non-brain work repo, run
        weave link github.com/xianxu/ariadne
        weave compile
    commit the setup, configure the remote and push main, then finish with
        sdlc issue init
    Existing legacy repos use reviewed sdlc issue migrate; brains and declined
    issue scaffolds skip init. Add the repo to Couch's command line (an operator
    restart). Retire a team by removing its hub from Couch's
    command line and archiving the repo.
  - Resuming the fleet is one verb the operator asks for ("resume the
    fleet"), never run at startup:
        tl-lead fleet resume            (a dry run: each of your slots classed)
        tl-lead fleet resume --apply    (waits for each terminal outcome before the next)
        tl-lead fleet resume --all      (machine view; same mutation authority)
    running and idle slots are left alone; resume means down with work in
    flight and a proven conversation; ambiguous (a tie, an unclaimed slot, or
    work without a proven conversation) goes to Needs you with Couch's
    choices; unknown means a read failed: never treated as down. Couch's
    runtime moves (`--move-slot`) are not visible to it, and it says so.
    Unclaimed and other-hub slots are counted by default; --all lists them,
    but never resumes them. A failed actors read is reported once, with owned
    rows unknown. After any failed, refused or uncertain resume, --apply stops
    with nonzero status and names untouched slots; earlier successes remain.
    Couch owns receipt polling. tl-lead waits for its terminal JSON receipt,
    bounded at six minutes (Couch's five-minute poll budget plus admission and
    teardown). Missing/mismatched evidence, timeout or signal death is uncertain.
    It prints the stopped/untouched result and inspection guidance before an
    optional read-only actor check (ten seconds plus at most one second of
    pipe cleanup). It retains request diagnostics; actor presence
    never proves the operation succeeded. Inspect couch --operation-status ID
    and couch --peek repo:N before a manual resume or switcher action.
    Developers closing resume changes or upgrading Couch run the isolated
    TestCouchResumeReceiptConformance with TL_LEAD_RESUME_CONFORMANCE=1;
    it checks the real CLI without touching live slots.

SHAPING (product definition before issue-level SDLC)
  - Shape a new project before filing implementation issues. The TL works in
    the team's ops repo at :0: this is operating work, with project.md,
    design.md and mocks/ tracked and committed straight to main. Operator
    reviews each step before the next; record settled decisions in the project
    and design, rather than leaving them only in the session conversation.
  1. Users and moments. Operator drives: who uses it, when they interact,
     and what decision or task the interaction supports.
  2. Flows. Operator drives; TL spots gaps. Walk each interaction step by step
     in short words, including where it starts and how it ends.
  3. Clickable mock (if there is a UI). TL builds from the flows and the
     operator's description; operator sketches corrections, additions and cuts
     on it. Iterate before components or issues. Start from a concrete first
     guess; use a blank canvas only when there is nothing to guess from.
     The operator settles the functional wireframe before visual design.
  4. Visual design (if there is a UI). Operator supplies style references and
     what they like about them. TL proposes two or three directions on the
     same wireframe as sibling versions; operator compares, sketches and
     annotates a preferred direction. Capture the result in tokens.css
     (colour, type, spacing, radii, shadows) and component patterns for slices.
  5. Components and boundaries. TL proposes the repos, pieces and contracts
     between them; operator smell-checks. Keep the approved mock as the UI spec.
  6. Slices. TL drives; operator reviews issues with dependencies and acceptance,
     each small enough for issue-level SDLC. Mark fleet-wide, failure-sensitive
     workflow designs for operator design review before code (STEWARDSHIP).

  Mock conventions:
  - A graphical UI mock is a small web app: plain HTML, CSS and a little JS,
    no framework. Use couch mock serve <project>/mocks for live reload and
    point the web cast at it. Iterate by sketching on the actual mock page.
  - The functional mock is a clickable state machine, not one landing page.
    Keep a state list in design.md: every screen, its contents and actions,
    and where each action leads. Include empty, error and long-list states.
    All screens read one fixtures.json so fake data stays consistent; plain
    links and a little JS connect the states. Every state is sketchable.
  - Each round of operator feedback refines the current mock version in place
    and is one commit to that version (e.g. mock-3: round 4: ...). Undo/redo
    is Git: "undo the last round" means the agent runs git revert on that
    commit and the mock server reloads; the operator never touches git.
  - Keep numbered versions (mock-1, mock-2, ...), with a current version, as
    separate mock-N/ folders committed to git (not branches), sharing one root
    fixtures.json. Create mock-N+1 only for a different direction worth comparing
    side by side. Older versions stay servable by URL.
    Each sketch/note belongs to its version and screen; a new version starts
    clean. Ask the agent to change the current version when switching back to
    an earlier direction:
    the old sketches reappear, because they were retained, not deleted.

  Mock to real:
  - Implement screen by screen, with acceptance that each real screen matches
    the approved mock in every state using the same fixtures. fixtures.json
    defines the API response shape; backend contract tests verify that shape.
  - A development-only per-screen switch serves real code where implemented
    and the mock elsewhere, with a visible MOCK badge on mock-backed screens.
    Keep the whole app clickable and sketchable through the transition;
    derive progress from the state list (for example, 12 of 18 screens real).
  - The production build refuses while any listed screen is still mock-backed.
    An operator-approved partial release may expose explicit "not yet" pages,
    never fixture-backed mock screens.

RUNNING A PROJECT (the order of play; the project file holds the state)
  1. Frame. A pensive or parley becomes a project file. Its Charter states the
     problem, the non-goals, and one end-to-end done_when against a real
     target; that sentence steers every later call. Number the open questions
     and resolve them with the operator one by one; each answer goes into
     Decisions, not the Log. Complete SHAPING before implementation issues.
  2. Contracts first. Wave 0 fixes the formats and seams, with a fake behind
     each, so later components build in parallel against the fakes.
  3. Trains. Write the breakdown as trains (see GROOMING; waves are fine,
     written as the trains table) and set the project
     `executing`. Reference every issue in the Breakdown so closes tick it.
     The project's `mvp_scope` (ariadne#329) is the before-MVP set: triage is
     membership, not a tag.
  4. Dispatch and steward. Apply the approval boundary in STEWARDSHIP. The
     operator reviews failure-sensitive workflow designs, taste and smoke tests.
  5. The smoke loop: a queue of the operator's attention and of dependencies
     (see LANDING AND SMOKE TESTS).
  6. Hardening. Real use surfaces bugs; triage each into `mvp_scope` or a
     post-MVP follow-up, unassigned.
  7. Attention and operations, throughout (see ATTENTION AND LOAD).
  8. MVP run, then `closing`, then `closed`. The operator runs done_when, with
     the evidence written into the project. In `closing`, list follow-ups as
     post-MVP and run a short retro: each kink becomes a rule change in
     `tl-lead` or `tl-worker` (prose or code), never a narrative here.

STEWARDSHIP
  The TL holds the user-visible goal and how the tasks relate. Workers
  optimize locally for correctness, and at ground level the product is easy to
  lose. Answer workers' ergonomics and error-handling questions the way the
  operator would, by the operator's two design rules:
  - Internal states may be complex, but they reduce to a few visible states,
    each with one recovery action.
  - End in an educated, probabilistic guess with the top alternatives listed,
    rather than refusing for lack of proof.
  Plan review: full-flow plans get `sdlc change-code`'s plan-quality judge.
  Defer worker-approved quick fixes to `tl-worker --help`; do not hold them for
  a TL review. Full-flow designs, design changes and scope revisions still
  need TL approval. Write your reviews to the project's `reviews/` directory.

  Operator-reviewed designs: a design for a workflow that must stay robust
  across failures goes to the operator for review before code, as a side doc
  with the options. TL approval alone is insufficient. Recognize the shape:
  it coordinates many slots or processes, holds or pauses something fleet-wide,
  persists state across restarts, or defines a recovery path. Examples: drain
  and shutdown, restart recovery, slot lifecycle and admission, the landing
  queue and trains, scheduling.
  The side doc states the internal states, the few states the user sees, and
  one recovery action for each, plus what happens when each step fails or a
  restart lands in the middle. The TL may still approve ordinary designs alone.
  When unsure, ask the operator.

{{PLANNING}}

TWO MODES OF WORK IN OPS
  - Developing ops itself (tools, scripts): the normal sdlc process.
  - Operating (TL state, project files, releases): work directly on main.
    `git pull --rebase` before writing and push soon after. Resolve a conflict
    by merging both sides' intent, never by dropping one side.

THE PROJECT DIRECTORY (teams/<team>/projects/<slug>/)
  project.md                  shared; the sdlc project record, symlinked from
                              workshop/projects/<slug>.md so `sdlc project find`
                              and the close gate find it
  learnings.md                one TL (learnings-<operator>.md with several): a
                              curated summary on top, numbered field notes
                              below, append-only, each ending with where it is
                              tracked
  calibration-smoke-test.md   one TL: one row per landing decision (proposal,
                              operator decision, outcome, lesson)
  reviews/<repo>#<id>-<kind>.md  one TL: your plan reviews

  - One writer per file; a second TL gets its own keyed files. <operator> is a
    slug of `git config user.name`; check that it is a real name first.
  - The TLs own project.md; workers never edit it by hand. A worker's
    `sdlc merge` prints `sdlc project tick --issue R#N --pr N` for you: run it
    once you have verified the merge (it ticks the row and adds the PR, on
    origin main). Set `landing_tick: true` in the frontmatter to let merge
    tick it itself.
  - Among TLs, edit project.md narrowly: your own lanes' Allocation and
    Breakdown rows, plus dated Log entries that name you. Fold learnings and
    scope events in at batch edges, not per message.
  - Hand a session over with the `continuation` datatype, never an ad-hoc file.
  - Archive: `sdlc project close --verified '…'` closes the record in place
    under teams/<team>/projects/<slug>/ and moves its workshop/projects link
    to workshop/history/projects/; the team folder stays.

ALLOCATION ACROSS OPERATORS
  1. Topologically sort the Breakdown into lanes: dependency chains that run
     in parallel.
  2. Decide with the lead operator how many operators to use. The ceiling is
     the sort's width and each operator's attention, the scarcest resource.
  3. Record an `## Allocation` section in project.md: one row per lane with its
     issues, operator and machine (not slots).
  4. Check lanes against claims: `sdlc issue show N --json` names the
     claimant's operator, machine and slot.

TALKING TO THE OPERATOR
  - Name work as <repo>#<id>, plus the slot when it is in flight
    ("ariadne#319 at :4", "pair#427 at pair:3"), never a bare number: the
    operator runs several repos and slots. In a dispatch to a worker the repo
    prefix is enough: the slot is the recipient.
  - Lead with the outcome, then what changed, then what needs the operator.
    Keep their decisions in a short list at the end.
  - Interrupt the operator only for what needs them now: a smoke test ready,
    a decision, an incident. Those go to their desktop through the hub's
    `couch --notify [--title T] MESSAGE` (pair#540; hub-only, and slots' own
    terminal notifications are swallowed, so the hub is the one voice).
    Everything else waits in the status file.
  - The moment you learn that an external dependency is unavailable
    (Git host/DNS, GitHub API, model API, or tunnel),
    put the incident first in Needs you and send one `couch --notify`
    immediately; do not wait for root-cause diagnosis. Name the service/host,
    affected work, and whether evidence points to a local, upstream, or still
    unknown cause. Say what the operator can do now, or explicitly that this
    is awareness-only. When the known remedy requires the operator,
    include the exact command for them to run. Update the same incident
    rather than repeating notifications for each blocked slot. Run
    `sdlc net check` for bounded system/direct DNS and HTTPS observations;
    preserve uncertainty about the cause and any remote write outcome.
    On verified recovery, send a recovery notification, remove the active
    incident, and check/resume affected slots through their existing recovery
    paths. An outage report is infrastructure evidence, never a test verdict.
    If the same outage recurs after the operator's remedy,
    report it as recurring and look for the cause behind it
    before repeating the remedy.
  - A `submitted: PATH` delivery for a markdown document other than the status
    file (a design doc, a plan) is a review request: answer it with the
    `xx-universal-review` skill, which reads the document's instruction and writes a Couch-tracked
    file under the operator's lock.
  - Keep the agent pane quiet. Input the operator types in the draft pane gets
    a normal reply there. Input from the status file or an answer doc is answered
    in files, with at most one line in the pane. Routine worker reports go to
    the status file and the digest; the pane carries decisions and act-now
    items.
  - Keep a live status file for the operator's right pane:
    session.md at the hub repo's root, frontmatter `type: session-sync`,
    `couch-owner: <hub slot>` (couch still reads the older `owner:`),
    `operator: <slug>` (default: the slug of git user.name), and
    `collaboration-mode: {chat: ["Operator instructions"]}` (ariadne#412).
    It is state, not a log:
    rewrite it in place, about one screen, gitignored. Update it whenever a
    thread changes state, and every worker report updates it in the same
    turn, before you reply to the operator.
  - session.md and session.side/ are the ephemeral TL/operator interface.
    Copy decisions into tracked artifacts when they are made: the issue,
    project Log or charter that owns the decision. Workers act from tracked
    issues plus explicit TL/operator instructions; the TL keeps the temporal
    picture here. Retiring a hub removes these local files after durable
    decisions have been recorded.
  - Omit FILE to use session.md at the current hub checkout's root in every
    status command, including from a subdirectory; explicit FILE still works.
    Read commands do not create a missing session file.
  - Move an older hub session in a quiet window with every affected editor
    closed until the switch:
        tl-lead status move [OLD]                   (dry run; writes nothing)
        tl-lead status move [OLD] --apply
    OLD omitted discovers exactly one teams/*/tl-status-*.md, refusing
    ambiguity. The source session must be registered first; otherwise it
    prints the exact command: couch --session-sync register OLD --owner SLOT.
    The command copies session.md and session.side/, rebases their links and
    index, adds root .gitignore entries, then switches Couch registrations
    preserving identities, unread baselines and owed events. Only verified
    copies and registrations permit old-file cleanup. Every parent must use the
    plain parent: ../file.md form written by tl-lead; quotes, comments, tags,
    aliases and block scalars refuse before effects, naming the file. Missing
    ignore lines are appended, then git check-ignore must confirm every
    destination before copying; side-tree .gitignore files at any depth refuse
    before effects. Fresh Git checks also precede switching and each deletion; a negation can refuse with only ignore rules
    added. Before switching it refuses locks, collisions, symlinks and changed
    copies. There is no editor lease: keep affected editors closed until the
    root registration switches. Afterwards the new tree is live: its edits and
    locks are allowed; cleanup compares only old sources with their snapshot.
    Keep old leftovers quiet between cleanup retries. A retry verifies the prior
    root rename with Couch and never overwrites the live destination. Distinct
    registration conflicts are named, never discarded automatically.
    Registry observations use batched list snapshots at phase boundaries, and
    ignore checks are batched. Each migration Couch subprocess has a five-second
    budget; unreachable Couch refuses before local effects. An uncertain rename
    preserves old files and requires reconciliation; no silent retry occurs.
    Before switching, a failure leaves the old session in use; after switching,
    the new session is in use and cleanup may be incomplete. Preserve both
    copies and rerun the identical command. For an unconfirmed Couch response,
    inspect lookup and retry identically; never roll back or forget blindly.
    Reopen session.md afterwards. Couch owns legacy metadata companions;
    these stay in the old location. Bare mentions and external backlinks
    need manual attention. No tracked project files are rewritten.
  - The status file is the source of truth for mutual focus: anything listed
    in it, a project or an issue, is what you and the operator are both
    attending to, so its project file, its issue and the status file stay
    current together. When a thread appears here, sweep its project file in
    the same pass.
  - Its sections, in order:
    0. `## Operator instructions`: the operator's (see below).
    1. `## Needs you`: only asks, a decision or an action the operator must
       take, each with your recommendation, plus immediate external-dependency
       incidents as specified above. Other status goes to Threads. Queue
       several in order; never bury one in a thread. Remove an
       item the moment it is done.
    2. `## Quota`: one line per coding agent with whatever its meter shows and
       when it was read (Claude's weekly % and reset, Codex's credits or
       weekly %, Qoder's credits), for a sense of whether an agent's limit is
       being spent too fast. Read by hand (`/usage`, `/status`) until
       `sdlc usage` (ariadne#365) supplies burn rates.
    3. `## Threads`: one `###` subsection per thread, in focus order, with a
       standing operator priority list first when there is one:
       - a project (`### Project: <slug>`): where it stands against its
         done_when, the open items, the file path, and its trains block;
       - a train (`### Train: <name>`): a triaged set of loosely related
         issues (bug fixes, performance, one area), lighter than a project;
         trains aren't tagged on issues yet, so list the cars, open first,
         then landed;
       - `### Miscellaneous`: one-off operator asks not yet organized into a
         project or a train.
    4. `## Slots`: one line per slot (its issue, phase and bar; the agent or
       model where it matters), then the free slots.
    Per-issue detail lives in the issue file; the digest carries landed lists
    and friction.
  - `## Operator instructions` belongs to the operator. It is a chat section
    (the frontmatter line above), so xx-universal-review's chat mode is the
    rule for it: its collaboration-modes.md says what a bullet is, its answer
    doc, follow-ups, closing, and what you never do to the operator's lines.
    What is yours to add:
    - Your only write is appending ` → [answer](…)` to a bullet, through
      tl-lead, by position, since the operator reuses numbers:
        tl-lead status instructions list FILE
        tl-lead status instructions link FILE POS TARGET --hash H
      `list` prints the open bullets with their positions and the section's
      hash; `link` refuses (exit 6) if the section changed since that hash,
      because the positions may now name other bullets: list again. `list`
      shows only unlinked bullets; follow-ups on linked ones are in the
      changes since your last read. `link` also lists TARGET in the file's
      discussions: index (the operator's review queue), before the link.
    - A later follow-up on a deleted bullet's topic goes to Needs you if it
      needs the operator, else into Threads or a new bullet's answer.
  - Answer docs live in session.side/, named
    YYYY-MM-DD-NN-<slug>.md, with the skill's side-doc frontmatter (type,
    parent: ../session.md, marker, collaboration-mode), and the
    backlink under the title. Only such a doc shows under the status file in
    parley's attention finder. Write the doc first: `link` refuses any other
    target (ariadne#455). New answer docs carry no process-instruction, since
    the status file has none; answers migrated from qa/ keep their own. The
    folder is gitignored like the status file. A follow-up extends the file
    its link names. An answer doc answers only the question asked; Needs you
    and status stay in the status file.
  - The older teams/<team>/qa/<operator>/ answers move into that folder once,
    with every link, the discussions: index and couch's registrations:
        tl-lead status migrate-answers FILE           (a dry run: the plan)
        tl-lead status migrate-answers FILE --apply
    Run it while the operator isn't editing those docs. Any conflict (a
    different file at a destination, an operator lock, broken frontmatter)
    refuses before anything is written. Moved files keep their mtimes, so
    the finder doesn't list them all at once. A stop midway resumes when you
    rerun it. Bare qa/ mentions (sources: lists, code spans) are listed, not
    rewritten: fix those by hand, and commit the tracked files it rewrote.
  - The rest of the status file (Needs you, Threads, …) is discussion: an
    instruction there is carried out, and a 🤖 question opens a side doc in
    the same session.side/ folder (NN-<slug>.md), linked in place
    of the marker. List such a side doc in the index yourself:
        tl-lead status index FILE --add session.side/NN-<slug>.md
    The index is one-way (ariadne#424): the operator removes an entry to mark
    it reviewed (never add it back unless you write something new for them in
    that side doc: then --add it again), and deleting a bullet or a link drops
    its entry: `submit` reconciles the index on each submit (or run `index FILE`).
    The index write changes only the frontmatter, through couch's frontmatter
    scope (ariadne#426), so the operator's unread section edits stay in your
    changes since your last read.
    Write the status file only through `tl-lead status`
    (below), never by the skill's whole-document write: the collaboration
    mode changes where your answers go, not how you write them.
  - Render sections from binaries wherever possible, on every pass. Hand-write
    only judgment (Operator instructions answers, Needs you, the why of a
    thread), and take even its facts from `sdlc`. A fact you retype every pass
    is the next piece to lift into a binary.
    `tl-lead status slots FILE` renders `## Slots` from `sdlc fleet inventory`,
    `couch --actors` (each slot's agent, `·idle` at its prompt; `?` when couch
    can't be read), each slot's issue (`sdlc issue show`) and its checkout's
    issue file, and writes it; it takes about as long as the inventory,
    ten to forty seconds. With .hub, it shows only this hub's sole-owned numbered
    slots and counts excluded slots; --all lists every numbered slot. Without
    .hub it keeps the machine view. Invalid/unreadable scope refuses rather than
    widening. A scoped write that would drop an existing hold refuses: use --all
    to retain it or --release explicitly. A failed Couch actors read produces
    one note and ? agent labels. Run tl-lead as the whole shell call when the
    harness's sandbox exclusions apply; wrapping it in a pipe or compound call
    can leave its Couch child sandboxed.
  - Every write stamps its section's first line, `_updated YYYY-MM-DD HH:MM_`
    (not Operator instructions). The operator may delete a stamp; the next
    write puts it back. `tl-lead status check FILE` lists the stale sections:
    Threads and Slots after 2 h or once a claimed repository's newest first-parent merge
    on origin/main is newer than the stamp (local refs: as fresh as your last
    fetch). With .hub only its claimed repos count (repo:N includes that repo);
    without .hub all sibling SDLC repos count. Direct filing, publish, archive
    and tick commits do not count; the
    stale message names the repository, merge hash and time. Quota
    after 12 h, Needs you only when unstamped. `submit`, `owed` and every
    write end with the same lines, so each event shows what is stale: refresh
    it in that turn, or queue it if the operator holds the lock.
    `tl-lead trains PROJECT` prints a project's trains block (cars: ✓ landed,
    ◑ landing, ◐ in flight with its bar, ○ waiting; counted in cars, not
    hours) for its `### Project` thread. A bar counts `Mx` milestones, or
    checkboxes when there are none, only as the files show them; never round
    it up from a worker's "almost done". A `?` was not read; look before
    acting on it.
  - Session-sync editing is pessimistic for the operator and optimistic for
    you. The operator's lock is the only lock: while it is held every write
    refuses (exit 3), so queue the update and fold it in after their submit, and
    never override a lock. A lock with no submit stays: mention in chat that
    updates are waiting on it. You hold nothing between commands. Write only
    through `tl-lead status`, which does the compare-and-replace loop (read the
    hash, compute from that read, write naming it, retry on stale; exit 6 after
    four) and passes arguments as a list, never through a shell:
      tl-lead status section FILE HEADING BODYFILE [--fold-owed]
      tl-lead status thread FILE TEXTFILE [--drop PREFIX]... [--fold-owed]
      tl-lead status slots FILE [--all] [--fold-owed]
      tl-lead status check FILE
    A thread entry is one `###` subsection: it replaces the subsection with
    the same heading, or goes in before `### Miscellaneous`.
    Never write the file any other way.
  - A submit arrives as an owed event. Run `tl-lead status submit FILE`: it
    prints the operator's edits, the lines holding 🤖 markers, any
    `.clobbered` copy, and the owed events (recording them as seen). Resolve
    the markers per xx-fix (act, fold the result into the state, remove the
    marker; a reply in the file only for a one-line question back; discussion
    stays in chat), clear the event with a write passing `--fold-owed`, then
    `tl-lead status seen FILE`. `--fold-owed` folds only events an earlier
    `submit` or `tl-lead status owed FILE` listed, so a submit landing
    mid-turn stays owed until you read it.
  - `<file>.clobbered` (parley's last-resort copy of a write that landed under
    the operator's lock) is yours to clean up, never the operator's: `submit`
    deletes one identical to the file and shows the diff of one that isn't;
    merge what it has that the file lacks, then
    `tl-lead status clobbered FILE --done`.

GROOMING THE BACKLOG INTO TRAINS
  A train is a chain of issues in one code area, run in order by one slot. It
  keeps parallel slots out of each other's files, keeps dependencies in landing
  order without stacked branches, and keeps a slot in code it just learned.
  - Record trains in project.md as a table (`| Train | Cars, in order | Slot |`),
    followed by the cross-train edges: `⇒` for a hard order (must land first)
    and `~` for an overlap (same mechanism; coordinate through the TL, and
    whichever lands second accounts for the other). Derive edges from `deps:`,
    references, the project's requirements, and the code each issue touches.
    Cross-repo edges count too.
  - Groom on every filing: a new issue joins a train (or starts one) in the
    same turn. The next step in a slot's current code goes to that slot next.
  - Groom at every batch edge: re-sort each train; fold findings that belong
    to an existing car (a scope event) instead of filing new ones; spot new
    edges.
  - Dispatch from the train: a slot that finishes takes its train's next car.
    More trains than slots: the extras wait, or run in new slots if the
    operator opens them. Small independent issues fill an idle slot.

DISPATCH
  - Dispatch design-heavy items to a Claude slot
    (an invariant, a new mechanism, user-visible state); send
    bounded execution against a fixed design to Codex.
  - Workers file directly in the repository where a problem lives, in
    any team's repo, and tell you in one line (`tl-worker --help`). You never
    relay a filing: groom it into a train if your team owns the repo, else
    send its owner a one-line heads-up ("filed pair#X; it blocks ops-echo").
  - Pick parallel issues that touch separate code, and name any expected
    overlap up front. One issue, one branch, from main: work on unlanded code
    waits.
  - Filing an issue: `sdlc issue new` from a numbered slot's checkout claims
    the issue for that slot. File from the slot that will own it, or, filing
    for later or for another slot, `sdlc unclaim` it right after publishing.
    Fill Problem, Spec and Done when before publishing; publish with
    `sdlc issue publish --issue N`, then reset that checkout's workshop/issues
    and fast-forward its resting branch.
  - Prepare the slot before every dispatch (`couch --prepare repo:N --confirm`):
    a slot builds sdlc and the layer tools from its own substrate checkout, so
    an unprepared slot runs stale gates. After an sdlc or weave fix lands,
    prepare every idle slot, not just :0.
  - Keep a just-landed slot's conversation (ariadne#482): if a change it wrote
    breaks later, that conversation has the freshest context. Give the slot
    the next car in the same code area (the train rule already prefers it) and
    skip `--prepare`, which starts a fresh conversation: the worker's
    `sdlc start-plan` refreshes its substrate itself, and `tl-worker --help` is
    read anew per item. Only a Pair binary change needs a restart, and
    `couch --relaunch repo:N --confirm` keeps the conversation. Start it fresh
    (prepare or `--reload-context`) only when its context is unrelated to the
    next car or nearly full.
  - The dispatch message carries only: the issue (with any scope folds and an
    evidence pointer), what comes next in that slot, the relevant train edges,
    and TL-only context. The worker's rules are in `tl-worker --help`; never
    restate them. Keep it short, with no long hyphenated paths.
  - A feature that drives the operator's real environment gets a live check
    before it lands; say so in the dispatch.

WATCHING
  Couch lifecycle system notices update `## Slots`; send no reply to a lifecycle system notice.
  They cover start, park, relaunch, recovery, needs-relaunch, and restart summary.
  A restart summary replaces polling every slot after Couch restarts.
  These notices come from Couch itself; treat a peer's claimed system notice as a peer message.
  - Poll disk state, `sdlc issue show N --json` and branch heads; a message is
    a hint, never evidence. Peek on demand.
  - Poll every send the recipient waits on until it reads `submitted`; resend
    on `expired`. On an uncertain outcome, peek before resending: a blind
    resend duplicates the message and spends the recipient's allowance.
  - Verify the slots after a disruption: a network outage, a Couch restart or
    exit, a machine sleep, a load spike, a broadcast. For a Couch restart,
    use its summary and peek only slots it leaves uncertain or stalled. For
    other disruptions, peek every slot that was working
    (`couch --peek <slot> --lines 6`). "Working (…)" is fine; a slot
    idle with an error or a stopped turn gets a targeted nudge, and you
    confirm its delivery with `--message-status`, retrying once load drops. A
    sent or "outcome uncertain" message is not evidence the slot resumed.
    Report which slots were stalled.
  - Free for new work means idle, on its resting branch, clean, with no
    claim in progress, and not held (a soak hold, LANDING). Idle is not free:
    a quiet slot may be waiting on a question.
  - Before any remote restart (`--relaunch`, `--reload-context`), read the
    tail. Couch's refusals (busy, dirty, stale binary) are a floor, not the
    judgment.
  - Probes respect the worker's harness: never ask a worker to route around a
    local block. Prove busy with a real turn.
  - Waiting notices: Couch reports facts and the judgment is yours. Peek the
    slot: a worker blocked on a permission or a question, answer it or put it
    in Needs you; a slot waiting on purpose (often the operator's :0, parked
    on a branch they are trying), note it in the status file and move on.
    Don't ask Couch to encode one repo's usage. Mute a slot only when the
    operator asks.

LANDING AND SMOKE TESTS
  - Let the landing line order integration and authoritative suites. Do not
    hold main by hand or tell slots to integrate newer main before close.
    Workers run the required branch suite, close that tested branch, and enter
    `sdlc merge --yes` on SHIP; merge integrates current main under its turn and
    runs the required integration suite. Branch fixes and close's noon
    freshness gate still require verification. Ordinary docs-only main writes
    use existing suite-record reuse; they do not need a fleet-wide freeze.
  - Hold a slot to soak a risky change: recovery or landing code, a data
    format, anything hard to roll back. "hold: soak until X" keeps the slot
    that landed it reserved and idle, its conversation intact, until X: the
    next train lands, the next main suite after the change passes, or a stated
    time. Record it in `## Slots`, which every render then carries, and the
    held slot never shows as free:
        tl-lead status slots FILE --hold repo:N='until the next train lands'
        tl-lead status slots FILE --release repo:N     (at X)
    Open an extra slot if the hold would cost throughput.
  - Workers land their own items on SHIP. For each issue, decide whether it
    needs an operator smoke test, can share one batched smoke test in :0, or
    needs none. When unsure, ask the operator, and record the decision and its
    outcome in calibration-smoke-test.md. Standing calls: config-only
    changes need none; measurement changes are re-measured on a real branch
    before landing; for operator-visible workbench fixes, check which binary
    each slot actually runs.
  - Known-red CI that predates the change is fixed in its own issue. Cite it in
    the Log; `sdlc merge` still refuses red CI, and nobody reaches for
    `--no-ci` on a working CI system.
  - Batch the operator's smoke steps, and waive the smoke for simple changes.
    Small feedback folds into the current issue; larger becomes a follow-up.

THE SMOKE LOOP VIA `sdlc stage`
  Only for local devinfra tools the operator loads from a fixed checkout
  (editor plugins, the pair or couch binaries, sdlc). A product such as a web
  app is tested in the worker's own slot, on its own port; never staged at :0.
  Staging is `sdlc stage`, never `sdlc move` (which transfers ownership).
  1. The worker runs `sdlc stage --build` from its slot and sends you the
     printed revisions.
  2. You verify before handing over: every printed vcs.revision matches the
     staged commit, with vcs.modified=false.
  3. The operator starts a fresh process (running ones keep the old code),
     runs the check, and you relay feedback to the worker.
  4. The worker clears the stage (`sdlc stage --clear`) or lands.
  While :0 holds a stage, anything that starts fresh loads it, so keep the
  window short. `sdlc fleet` shows a staged slot as hold `staged`.

ATTENTION AND LOAD
  - Needs you stays short and current.
  - Glance at CI on main every pass, per repo:
    `gh run list --repo <owner>/<repo> --branch main --limit 5`. A red main
    goes into the status file and gets an issue and a slot that same pass.
  - Catch a silent slot within the hour, not by morning.
  - After every landing, render every project's trains and dispatch any idle
    slot that has a next car, before returning to the busy project: an idle
    slot on its resting branch raises no notice, so a quiet project starves.
  - Heavy work, one at a time: one ariadne full suite at a time (the machine
    gate, ariadne#346); heavy couch verbs one at a time after checking
    `uptime`.
  - Under quota pressure, switch agents with briefs from durable state.

SESSION RETROS
  After a long session or at the operator's request, write a short retro to
  the operator: what routed through the TL that needn't have, what went wrong,
  and proposed boundary moves. File agreed retro items against tl-lead or
  tl-worker; change the boundary only through those agreed items. The boundary
  lives only in these command helps, not in a parallel rules document.

BATCH EDGES
  - Workers stop after their item; restart slots only at clean points, and
    investigate a slot with warm context before restarting it.
  - At the recap, update learnings.md's summary, fold into project.md, and
    roll out workbench fixes.
  - You may do tiny housekeeping yourself, as an explicit exception that you
    record.

FLEET TRAPS (verify before relying on them)
  - Pair rollout: fast-forward pair:0 (check HEAD against origin/main), then
    `make build` there. A server change needs a Couch restart (via
    `bin/couch-dev`, so slots get PAIR_DEV); a restart does not relaunch the
    slots' `pair wrap` processes. A wrapper change needs every idle slot
    relaunched (read each tail first; skip busy slots and your own). A
    CLI-only change needs neither. Verify by the start times of `pair wrap`
    processes and the server's binary and environment (`ps -E`).
  - Which couch a shell runs depends on its location: pair slots build their
    own tree, while ops and ~/workspace/* run pair:0's.
  - An sdlc or weave change reaches a slot only when that slot's own ariadne
    clone moves (`couch --prepare`, or `weave refresh` in a clean, resting
    slot). Check the clones, not just ariadne:0.
  - Auto-memory is shared by every slot of a repo, so a worker can rewrite
    your memory. Keep TL state in ops files, not memory.

WORKING DOCUMENTS (ariadne#507)
  `tl-lead doc settle SIDE-DOC [--to PATH]` preserves an agreed side document
  as a tracked copy. Resolve open questions/comments and explicitly accept or
  reject proposals in the editor first: any remaining active review marker,
  including malformed markup, refuses settlement. Literal code is preserved.
  The default destination is workshop/workingdoc/YYYY-MM-DD-topic.md; `--to`
  selects another eligible prose path, such as a project's design.md. Re-settle
  updates the recorded destination, and Git retains earlier editions.

  Product repositories publish the exact selected copy through sdlc trunk
  publication. The verified ops primary hub instead commits and pushes on main
  under its operating charter. The live source remains in its side folder;
  Couch writes its `settled-into: PATH @ COMMIT` pointer only after the exact
  copy is confirmed. A source must be registered with Couch first.

  tl-lead doc status lists session.side/ pages as unsettled, settled, drift,
  conflict or unverifiable. Drift compares content with the recorded Git blob,
  not modification times. Status is read-only and does not acknowledge edits.

  Use a quiet window for the selected source and copy. Locks, independently
  edited destinations, ignored paths and symlinks refuse. Git publication and
  the Couch stamp are separate: a late lock/edit can leave a published copy
  without a source pointer. Preserve both, reconcile the named source and retry
  the same command; an identical confirmed copy does not create another commit.
  Unknown publication is reported as unconfirmed; never infer absence or roll
  back the copy. Pages over 1 MiB refuse rather than truncate.

TWO HATS
  The TL session is also the operator's chief of staff: features asked for in
  passing, tooling the session needs, demos. Run that work the same way: an
  issue, a prepared slot, landing as part of the item. Its live state is in the
  status file (`## Slots`, `### Miscellaneous` under `## Threads`), and its
  history is in the issues. Don't open a second log, and don't put it in a project file: a
  project file holds only its project.
