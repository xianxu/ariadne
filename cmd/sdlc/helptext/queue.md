`sdlc queue` — the advisory work queue: of the loose issues that could be done in
any order, which one is next, and why.

The file is ONE file per repo at `workshop/queue.md`, and it lives on the trunk.
Every operation reads and writes `origin/main` directly, so the queue is the same
from every checkout — a feature branch, a detached HEAD, a worktree, another
machine. Nothing holds a competing copy, so nothing merges.

  sdlc queue                                  list, from origin/main
  sdlc queue add <ref> "<why-now>"            append an entry
  sdlc queue add <ref> "<why>" --tag <t>      ... grouped under a project tag
  sdlc queue add <ref> "<why>" --project      ... as a project line, not an issue
  sdlc queue remove <ref>                     drop the entry
  sdlc queue move <ref> --before <ref>        reorder
  sdlc queue move <ref> --after <ref>

LINE FORMAT

  - <ref> — <why-now> [tag]
  - project:<ref> — <why-now> [tag]

  The separator is an EM-DASH surrounded by spaces (` — `), not a hyphen. The
  trailing `[tag]` is optional and must close the line. `<why-now>` may not
  contain a newline, a control character, ` — `, or brackets — those are
  rejected rather than escaped, because the format's value is that a human reads
  it as a list.

  A line that does not parse is PRESERVED in the file, never discarded — but it
  is not listed as an entry, and `sdlc queue` says how many such lines it found.
  A hand-edit that used a hyphen is the usual cause.

ADVISORY, NEVER BINDING

  The queue indicates a section of work that will likely happen. It is not a
  commitment, and it never overrides `deps:` — which is the authoritative record
  of hard blocking. If a line is really a block, it belongs in `deps`, or there
  are two truths about blocking that will disagree.

  It also carries no status. Labels tolerate staleness; state does not. A stale
  why-now is still better than an order you cannot evaluate, but a stale status
  goes confidently wrong.

TWO KINDS OF ENTRY

  An ISSUE line is a next action — something that can be started.
  A PROJECT line (`--project`) is a declared intent to work in an area. It is
  coarser, and it should be replaced by an issue line once it becomes the actual
  next thing. A project line that never becomes actionable is permanently
  present, carries no ordering information, and trains the reader to stop reading
  the file.

CONCURRENCY

  `origin/main` is the permanent base. Each operation is fetch → apply → push,
  and a rejected push retries the same three steps (bounded at 3) — re-applying
  YOUR INTENT to the base a peer just moved, so both edits survive. That is why
  the operations are `add`/`remove`/`move` rather than "here is the new file":
  an intent replays, content cannot.

  Where an intent cannot be replayed, the command REFUSES and prints the current
  trunk queue alongside the edit it could not apply, so you or an agent can
  re-derive it. The cases:

    remove of a ref a peer already removed   converges — same end state
    add of a ref a peer already added        converges — newer why-now wins
    move whose anchor or subject is gone     REFUSES — no defensible position
    the trunk moved 3 times running          REFUSES — surfaces git's rejection

  Offline: a LIST degrades to the last-fetched trunk state and says so loudly. An
  edit refuses, because a compare-and-swap push has nothing to compare against.

READING IT

  Read the queue through this command, not by opening `workshop/queue.md`. The
  file is tracked, so your branch has a copy — and that copy is whatever your
  branch last saw, which is the staleness this command exists to avoid.

RELATED
  `deps:`                      hard blocking; authoritative
  project `status` / roadmap   where PROJECT-level prioritization lives
  a project's `## Breakdown`   authoritative for ordering WITHIN that project

  The queue is for loose ends. Overlapping it with a project's breakdown gives
  you two orderings that will disagree.
