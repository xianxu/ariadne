Release the lock on issues this workspace owns (#284). The owner is cleared;
the status never changes.

  sdlc unclaim --issue N
  sdlc unclaim --issue 284,285 --note "split into #290; nothing pending"

Unclaiming is a normal operation: the owner (a slot) lets go so another slot,
machine or operator may take the issue. It never moves the lifecycle — an open
issue stays open, started work stays started.

AN OPEN CLAIM (shaping). Before the owner is cleared, each issue whose local
details differ from main's is published to main (`sdlc issue publish`), while
the claim is still held — so a claim-to-shape cycle never loses its edits. If
that publication refuses (main moved since this checkout's base), the claim is
kept and nothing is released. Several open claims release together in one
tracker commit, all or nothing.

STARTED WORK (working, blocked, codecomplete) is a handoff, one issue at a
time. Run it from the issue's branch with a clean tree (untracked files
included): a handoff carries only committed work. The note, if any, is
committed on the branch so it travels; the branch is pushed to the publication
remote (leased on the copy this checkout last fetched — the owner is its only
writer, but a tip it never saw is not overwritten); the card records the
release with the branch and its tip; this checkout returns to its resting
branch, freeing the branch for another worktree. A plain `sdlc claim` in any
other checkout — slot, machine or operator — takes it over at exactly that tip.
The pushed branch is the issue's own: it goes away when the issue lands (or is
abandoned).

--note adds a dated line to each issue's ## Log before publishing: where the
work stands and what comes next, for whoever takes it.

Every release records who let go. A rerun after a lost response finds the
cards already released by this workspace and only finishes the checkout (a
resting branch fast-forwards to main).

FLAGS

  --issue <n>[,<n>…]    required issue ID(s)
  --note "<line>"       one line for each issue's ## Log
  --issues-dir <path>   override $WF_ISSUES_DIR / workshop/issues
  --dry-run             check and describe; change nothing

RELATED

  sdlc claim            take the lock (a released card is claimable by anyone)
  sdlc issue publish    publish details edits without releasing
  sdlc reclaim          the operator's transfer from an owner who cannot release
