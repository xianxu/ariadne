---
gate: boundary-review
issue: 242
id_prefix: BR
rounds:
    - "n": 1
      timestamp: "2026-09-22T22:56:14-07:00"
      agent: codex
      findings:
        - id: BR-1
          severity: Critical
          title: Malformed all-zero HEAD values are accepted as a valid unborn workspace
          detail: '`Classify` treats any all-zero string as unborn (`pkg/workspace/identity.go:63`) without requiring the Git OID grammar. A malformed value such as `HEAD 0` therefore resolves a primary successfully with `head: null`, contradicting the plan''s fail-closed malformed-observation contract. Validate HEAD as exactly a 40/64-character zero OID for unborn state, or reject it; add a regression test that fails before the fix.'
          family: malformed-git-evidence-fails-open
          round: 1
        - id: BR-2
          severity: Important
          title: README update is missing for the new workspace CLI surface
          detail: '`cmd/sdlc/main.go:112` adds `sdlc workspace [address] --json`, but README.md is unchanged and has no usage or contract entry. Add the command and its basic invocation/JSON purpose to README.md.'
          family: user-facing-surface-docs
          round: 1
      recipe: milestone-review
      blocked: true
    - "n": 2
      timestamp: "2026-09-22T23:07:42-07:00"
      agent: codex
      dispose:
        - id: BR-1
          disposition: addressed
          note: '`Classify` now validates every non-bare HEAD with strict 40/64-character lowercase OID grammar before interpreting all-zero values; regression coverage is present in `identity_test.go`.'
          round: 2
        - id: BR-2
          disposition: addressed
          note: README.md now documents `sdlc workspace`, JSON usage, state integration, and links the workspace contract.
          round: 2
      recipe: milestone-review
      blocked: false
---

# Gate ledger — ariadne#242 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-09-22T22:56:14-07:00 (codex) — BLOCKED

### Raised

- **BR-1** [Critical] `malformed-git-evidence-fails-open` Malformed all-zero HEAD values are accepted as a valid unborn workspace
  `Classify` treats any all-zero string as unborn (`pkg/workspace/identity.go:63`) without requiring the Git OID grammar. A malformed value such as `HEAD 0` therefore resolves a primary successfully with `head: null`, contradicting the plan's fail-closed malformed-observation contract. Validate HEAD as exactly a 40/64-character zero OID for unborn state, or reject it; add a regression test that fails before the fix.
- **BR-2** [Important] `user-facing-surface-docs` README update is missing for the new workspace CLI surface
  `cmd/sdlc/main.go:112` adds `sdlc workspace [address] --json`, but README.md is unchanged and has no usage or contract entry. Add the command and its basic invocation/JSON purpose to README.md.

## Round 2 — 2026-09-22T23:07:42-07:00 (codex) — passed

### Disposed

- BR-1 — addressed — `Classify` now validates every non-bare HEAD with strict 40/64-character lowercase OID grammar before interpreting all-zero values; regression coverage is present in `identity_test.go`.
- BR-2 — addressed — README.md now documents `sdlc workspace`, JSON usage, state integration, and links the workspace contract.

## Open findings

(none — every finding has been disposed)
