---
id: '000038'
status: done
created: 2026-05-27
updated: 2026-05-27
estimate_hours: 4
actual_hours: 3.0
---

# substrate simplification — symlink-only, drop vendor/copy, bootstrap-clones-peers

## Problem

#37 landed the `construct/go.mod` split — surgical Go vendoring (substrate-tool source in `construct/vendor/`, app deps separate at root). The split solved the immediate "pair's vendor/ ballooned" complaint but kept the dual-mode (`--symlink` vs `--vendor`) machinery around in setup.sh.

After further design work, the dual-mode + vendor concept turns out to be unnecessary complexity:

1. **Privacy was overstated.** Ariadne's artifacts aren't really sensitive — what gets shipped as substrate (AGENTS.md, manifests, skill docs) is hard to keep private once any consumer ships derived content. Making ariadne public (or accepting that its content effectively leaks through consumers) removes vendor's main justification.

2. **Vendor is just symlink-with-ergonomic-veneer (modulo privacy).** Functionally, vendoring a snapshot is equivalent to "symlink to a frozen branch in upstream + don't pull updates." The differences are operational ergonomics (where the divergent copy lives, how clone DX works) — not semantic.

3. **Bootstrap-clones-peer addresses the clone DX argument for vendor.** If the only reason for vendor was "operator can clone derivative without needing upstream sibling," then `make bootstrap` that auto-clones missing peers achieves the same DX without bundling source.

4. **Per-artifact `copy` for divergence has its own headache.** Mixed copy/symlink within a logical command (cmd/gmail with some files copied, some symlinked) makes state hard to reason about. Per-operator branches in source repos (one branch per operator's customizations, shared across all that operator's brains) is a cleaner divergence mechanism than per-artifact copy.

The simplification: drop vendor mode entirely, drop `copy` action entirely, everything is symlink. Divergence happens via per-operator branches in source repos. Clone DX happens via bootstrap-cascades-and-clones.
