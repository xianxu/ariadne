---
gate: plan-quality
issue: 224
id_prefix: PQ
rounds:
    - "n": 1
      timestamp: "2026-09-12T18:31:04-07:00"
      agent: claude
      findings:
        - id: PQ-1
          severity: Important
          title: Plan delivers the entry to pair via `sdlc propagate-base`, but the registry is not in base.manifest; delivery is a rebuild of the shared binary
          detail: architecture.md is go:embed'ed (architecture.go:47) and pair's `sdlc` is a symlink to ariadne/bin/sdlc, so the step is `make sdlc-build` (Makefile.workflow:796), not propagate-base, which would commit into every dependent for no effect. Fix the sixth checkbox and the fourth Done-when.
          family: unbacked-existing-behavior-claim
          round: 1
        - id: PQ-2
          severity: Minor
          title: Registry growth per entry has no stated bound or retirement path in the atlas paragraph
          detail: Each entry adds ~50 lines to four gate prompts; the entry about ends should say where a registry entry ends (architecture-deferred.md is the existing route). One sentence in the atlas map paragraph.
          family: arch-constraints-prompt-budget
          round: 1
      blocked: true
    - "n": 2
      timestamp: "2026-09-12T18:33:09-07:00"
      agent: claude
      dispose:
        - id: PQ-1
          disposition: addressed
          note: Plan step 6 and Done-when 4 now say make sdlc-build; verified embed at architecture.go:47, target at Makefile.workflow:796, no manifest row for architecture.md.
          round: 2
        - id: PQ-2
          disposition: addressed
          note: Atlas step now names the retirement route; cmd/sdlc/internal/judge/architecture-deferred.md exists and is the activation/deactivation seam.
          round: 2
      blocked: false
content_hash: fe36082e6936831587b2e1cd274a73d56690b7949f933c1e75cd02e719e7ba42
---

# Gate ledger — ariadne#224 (plan-quality)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-09-12T18:31:04-07:00 (claude) — BLOCKED

### Raised

- **PQ-1** [Important] `unbacked-existing-behavior-claim` Plan delivers the entry to pair via `sdlc propagate-base`, but the registry is not in base.manifest; delivery is a rebuild of the shared binary
  architecture.md is go:embed'ed (architecture.go:47) and pair's `sdlc` is a symlink to ariadne/bin/sdlc, so the step is `make sdlc-build` (Makefile.workflow:796), not propagate-base, which would commit into every dependent for no effect. Fix the sixth checkbox and the fourth Done-when.
- **PQ-2** [Minor] `arch-constraints-prompt-budget` Registry growth per entry has no stated bound or retirement path in the atlas paragraph
  Each entry adds ~50 lines to four gate prompts; the entry about ends should say where a registry entry ends (architecture-deferred.md is the existing route). One sentence in the atlas map paragraph.

## Round 2 — 2026-09-12T18:33:09-07:00 (claude) — passed

### Disposed

- PQ-1 — addressed — Plan step 6 and Done-when 4 now say make sdlc-build; verified embed at architecture.go:47, target at Makefile.workflow:796, no manifest row for architecture.md.
- PQ-2 — addressed — Atlas step now names the retirement route; cmd/sdlc/internal/judge/architecture-deferred.md exists and is the activation/deactivation seam.

## Open findings

(none — every finding has been disposed)
