# Adapter: editing the file

You answer a round (`../round.md`) by changing the document on disk. Which
path you take depends on whether Couch tracks the document. Never edit a
tracked document directly: the operator's lock and the compare-and-swap below
are what keep your write from overwriting their unsaved work. Every `couch`
call is the whole shell command (AGENTS §15).

**Temporary files.** Each temporary file below states when you delete it. All
of them live under `$TMPDIR`, so one you could not delete (a refused tool
call, a lost conversation) is left in scratch space, not in the repository.

## Is the document tracked?

Decide in this order, and edit directly only on a confirmed **untracked**:

1. Frontmatter `type: session-sync`: **tracked**, whatever the registry says
   and wherever you are (Couch routes these through its own storage without a
   registration). If Couch can't be reached (outside a Couch slot, or a
   failing call), write nothing and tell the operator; never fall back to a
   direct edit.
2. Outside a Couch slot (`COUCH_TREE` unset): **untracked**.
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
2. **Changes since your last read:** `couch --session-sync diff FILE`. Couch
   moves its baseline on your write, so the diff never holds your own edits
   (`couch --skill`: "anything outside what you last wrote or read with
   `--all`"). It must come before step 3, which marks everything seen and
   empties it.
3. **The snapshot:** `couch --session-sync read --all FILE` prints the text and
   its hash H. This text is the round's input: the instruction comes from its
   frontmatter, and the markers from its body. The diff only points at what
   moved; an edit made between steps 2 and 3 is in the snapshot even though
   the diff missed it, so read the markers from the snapshot, not the diff.
4. **Compute** the round (`../round.md`) on the snapshot, and write the whole
   new document, frontmatter kept, to a body file made with
   `mktemp "${TMPDIR:-/tmp}/review.XXXXXX"`.
5. **Write:** `couch --session-sync write --all BODYFILE FILE --if-hash H --fold ID[,ID]`,
   folding only those IDs kept in step 1 whose asks this write answers. A
   kept delivery you didn't answer, and one that arrived after step 1, stay
   owed for the next round. With no IDs to fold, leave out `--fold`. When the
   answer changes nothing, write the snapshot back as read, so its asks are
   still folded. Then delete BODYFILE, whether the write succeeded or was
   refused.

Refusals:

- **Exit 3, the operator holds the lock:** they are still editing. Don't wait
  in a loop or force it. Keep the pending ask, and answer it after their next
  submit, starting again from step 1. The ask stays yours to answer even if
  that round's diff is empty. If `status` shows the lock held for more than 30
  minutes, ping the operator in one line in your pane (for the TL's session
  file, a Needs you item instead).
- **Exit 6, stale:** the document changed since your read. Go back to step 1,
  refresh every input, and recompute on the new snapshot.

## Side docs and a parent

A round in discussion or chat mode may create side docs and may fold into a
parent document (`../collaboration-modes.md` §4 orders the writes). Write a new
side doc as a new file, directly: no one holds a lock on a file that doesn't
exist yet. Every existing document the round changes, the submitted one and a
parent alike, takes its own path above or below by its own tracked or untracked
lookup. A tracked parent gets its own snapshot (`read --all`, its hash) and its
own `write --all --if-hash`, with no `--fold`: its owed events belong to its
own rounds. On exit 6 (stale), read the parent again and recompute the fold.
On exit 3 (the operator holds the parent's lock), end the round before the
submitted document is written, so its ask stays for the next submit.

## Untracked

Edit the file in place with your harness's file tools. There is no lock, so
read the file again right before writing.

**Changes since your last read** exclude your own edits. After each write,
re-read the file and save that text, the file as you left it, to a copy made
with `mktemp "${TMPDIR:-/tmp}/review-base.XXXXXX"`; remember the copy's path,
and delete the copy it replaces. Delete the last copy when you are done with
the document: the operator says so, or you hand the slot to other work.
Next round, the changes are `diff -u COPY FILE`: the operator's edits since
you left it, never your own, and never measured against the operator's
previous send. Without a saved copy (a first read, or a new conversation),
read the whole file instead.
