---
id: 000281
status: open
deps: []
github_issue:
created: 2026-10-01
updated: 2026-10-01
estimate_hours:
card_mirror: '1bdbf1d52dd3311de9459e0d54dd8e888205348b' # card fields mirrored from issue-cards; edit via sdlc
---

# weave overwrites generated outputs instead of refusing on mismatch

## Problem

`weave compile` refuses with `apply: preserving authored replacement at <path>`
whenever a recorded output's bytes differ both from what weave last wrote
(`construct/generated/weave/ownership.json`) and from what it wants to write
(`plan/ownership.go`, `ApplyManaged`). The refusal has no recovery path:

- `weave compile` has no override flag, and the message names no next action.
- The only ways out are deleting the file, which loses the "protected" edit
  anyway, or recreating the exact recorded bytes by hand.

**Observed in parley.nvim slot2, 2026-10-01.** ariadne#277 added the `claimant`
card field. Compile regenerated the committed
`construct/generated/vocabulary/issue.json` and recorded the new bytes. Later
something put the stale HEAD version back. The likely cause is a git
restore/stash to get the clean tree `weave refresh` requires; it happened twice,
at 17:52 and at 18:10. From then on, `weave compile` and `weave refresh` both
refused on a file nobody had edited by hand. The way out was committing the
regenerated JSON (parley PR #221). Nothing in the error pointed there.

The check came from #239 ("one verb, two ownership classes"), where weave used
to clobber a repo's own root `Makefile`. #239 itself fixed that by no longer
generating the root Makefile. Repo-specific behaviour now goes in overlay sources
weave never writes (`AGENTS.local.md`, `Makefile.local`). So for files weave does
generate, an edit belongs in an overlay, and refusing protects nothing: it only
delays losing the edit. The refusal also leaves `.gitignore` half-migrated
(#265 item 1).

## Spec

- **Generated outputs always converge.** When a recorded output's bytes differ
  from the last write, weave overwrites it and prints a warning. The warning
  names the path and says where the edit belongs: the overlay for that output
  family, or the layer source for vocabulary and datatype exports. Weave never
  refuses on content mismatch for a path it owns in the inventory.
- **Never lose bytes silently.** Before overwriting a mismatched output, copy the
  old bytes to a backup beside the inventory (e.g. `construct/generated/weave/
  displaced/<path>`). Each successful compile keeps only the latest displaced copy
  per path, so the backup's lifecycle is named (ARCH-FUNERAL).
- **The inventory keeps one job: safe retirement.** Weave removes an output it no
  longer generates only when the file still matches the recorded identity.
  Otherwise it leaves the file, and warns once and names it.
- **Unowned occupied destinations** are a different case: a path never in the
  inventory, holding a file weave did not create. Today's refusal stays for
  these, but the message must give the next action (move the file aside, or
  declare it an overlay).
- **Vendored outputs.** Some generated outputs are also committed. parley.nvim
  commits `construct/generated/vocabulary/issue.json` for fresh clones (parley
  #208). For these, a compile that changes a tracked output reports "regenerated
  tracked output <path>: commit it" instead of leaving an unexplained dirty
  tree.

## Done when

- A compile over a recorded output whose bytes were reverted to an older weave
  write (the parley case: HEAD's stale copy) overwrites it, warns, keeps a
  displaced copy, and exits 0 (regression test).
- A compile over a recorded output with an arbitrary hand edit does the same,
  and its warning names the overlay or layer source where the edit belongs.
- Retirement still deletes only matching files; a mismatched retired output is
  left in place with a warning (test).
- An unowned occupied destination still refuses, and the message states the next
  action (test).
- A compile that changes a tracked output prints the commit hint (test).
- `atlas/workflow/weave.md` describes the convergence rule. "preserving
  authored replacement" no longer appears in the code.

## Plan

- [ ]

## Log

### 2026-10-01
- Filed from a parley.nvim slot2 session after weave refused on a reverted vendored vocabulary file. Related, not blocking: #265 (refusal leaves .gitignore half-migrated), #239 (origin of the check), parley.nvim#208 (vendored vocabulary), parley PR #221 (workaround).
