---
gate: boundary-review
issue: 224
id_prefix: BR
rounds:
    - "n": 1
      timestamp: "2026-09-12T18:43:55-07:00"
      agent: claude
      findings:
        - id: BR-1
          severity: Important
          title: Close review dispatched from stale /tmp/claude-501/sdlc (7 entries); delivery Done-when for boundary-review prompts not evidenced
          detail: This review's prompt enumerates seven markers without ARCH-FUNERAL, matching /tmp/claude-501/sdlc (mtime 18:26) rather than bin/sdlc (18:39, eight entries). Remove the temp copy, re-run close via ~/.local/bin/sdlc, confirm the prompt shows eight entries. Same rule as lessons.md:1046 (#171).
          family: stale-binary-evidence
          round: 1
        - id: BR-2
          severity: Minor
          title: principle clause says every ARCH-CONSTRAINTS constraint is felt at the time, but that entry lists "scale and growth"
          detail: architecture.md:200 — drop "every"; the atlas paragraph states the load/residue split without the universal claim.
          family: neighbour-boundary-overclaim
          round: 1
      blocked: true
    - "n": 2
      timestamp: "2026-09-12T18:48:48-07:00"
      agent: claude
      dispose:
        - id: BR-1
          disposition: addressed
          note: $TMPDIR/sdlc is gone, only ~/.local/bin/sdlc (symlink to bin/sdlc, mtime 18:44) is on PATH, it renders 8 entries, and this round's dispatched prompt enumerates ARCH-FUNERAL.
          round: 2
        - id: BR-2
          disposition: addressed
          note: architecture.md:199-206 now names what ARCH-CONSTRAINTS budgets instead of claiming "every constraint"; goldens re-captured 1:1 and the judge package passes.
          round: 2
      blocked: false
---

# Gate ledger — ariadne#224 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-09-12T18:43:55-07:00 (claude) — BLOCKED

### Raised

- **BR-1** [Important] `stale-binary-evidence` Close review dispatched from stale /tmp/claude-501/sdlc (7 entries); delivery Done-when for boundary-review prompts not evidenced
  This review's prompt enumerates seven markers without ARCH-FUNERAL, matching /tmp/claude-501/sdlc (mtime 18:26) rather than bin/sdlc (18:39, eight entries). Remove the temp copy, re-run close via ~/.local/bin/sdlc, confirm the prompt shows eight entries. Same rule as lessons.md:1046 (#171).
- **BR-2** [Minor] `neighbour-boundary-overclaim` principle clause says every ARCH-CONSTRAINTS constraint is felt at the time, but that entry lists "scale and growth"
  architecture.md:200 — drop "every"; the atlas paragraph states the load/residue split without the universal claim.

## Round 2 — 2026-09-12T18:48:48-07:00 (claude) — passed

### Disposed

- BR-1 — addressed — $TMPDIR/sdlc is gone, only ~/.local/bin/sdlc (symlink to bin/sdlc, mtime 18:44) is on PATH, it renders 8 entries, and this round's dispatched prompt enumerates ARCH-FUNERAL.
- BR-2 — addressed — architecture.md:199-206 now names what ARCH-CONSTRAINTS budgets instead of claiming "every constraint"; goldens re-captured 1:1 and the judge package passes.

## Open findings

(none — every finding has been disposed)
