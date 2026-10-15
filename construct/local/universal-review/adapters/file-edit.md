# Adapter: editing the file

You answer a round (`../round.md`) by changing the document on disk. Which
path you take depends on whether Couch tracks the document. Never edit a
tracked document directly: the operator's lock and the compare-and-swap below
are what keep your write from overwriting their unsaved work. Every `couch`
call is the whole shell command (AGENTS §15).

## Is the document tracked?

Decide in this order, and edit directly only on a confirmed **untracked**:

1. Outside a Couch slot (`COUCH_TREE` unset): **untracked**.
2. Frontmatter `type: session-sync`: **tracked**, whatever the registry says
   (Couch routes these through its own storage without a registration).
3. `couch --session-sync lookup FILE --json` exits 0 with Code `ok`:
   a non-null Registration is **tracked**; an explicit `null` Registration is
   **untracked**.
4. Anything else (an error, a malformed answer, a usage message from an older
   Couch) is **unknown**: write nothing, and tell the operator the lookup
   failed. Unknown ownership is never permission to edit directly.

## Tracked

One round reviews one snapshot: the text whose hash you write against.

1. **Owed events:** `couch --session-sync status --json FILE`. Keep the IDs it
   lists as owed, and the contents of those deliveries (the operator's
   submits): they are the asks this round answers. A `delivery-gap` or
   `not-recording` refusal means owed events are unknown, not zero: follow the
   refusal's advice before writing.
2. **Changes since your last read:** `couch --session-sync diff FILE`. It must
   come before step 3, which marks everything seen and empties it.
3. **The snapshot:** `couch --session-sync read --all FILE` prints the text and
   its hash H. This text is the round's input: the instruction comes from its
   frontmatter, and the markers from its body. The diff only points at what
   moved; an edit made between steps 2 and 3 is in the snapshot even though
   the diff missed it, so read the markers from the snapshot, not the diff.
4. **Compute** the round (`../round.md`) on the snapshot, and write the whole
   new document, frontmatter kept, to a body file made with
   `mktemp "${TMPDIR:-/tmp}/review.XXXXXX"`.
5. **Write:** `couch --session-sync write --all BODYFILE FILE --if-hash H --fold ID[,ID]`,
   folding exactly the IDs kept in step 1, which this write answers. A
   delivery that arrived after step 1 stays owed for the next round. With no
   IDs kept, leave out `--fold`. With nothing to change, write the snapshot
   back as read, so the asks are still folded.

Refusals:

- **Exit 3, the operator holds the lock:** they are still editing. Don't wait
  in a loop or force it. Keep the pending ask, and answer it after their next
  submit, starting again from step 1. The ask stays yours to answer even if
  that round's diff is empty. If `status` shows the lock held for more than 30
  minutes, ping the operator in one line in your pane (for the TL's session
  file, a Needs you item instead).
- **Exit 6, stale:** the document changed since your read. Go back to step 1,
  refresh every input, and recompute on the new snapshot.

## Untracked

Edit the file in place with your harness's file tools. The changes since your
last read are the differences from the copy you last read in this
conversation; on a first read there are none, so read the whole file. There is
no lock, so read again right before writing if the operator may have edited
since.
