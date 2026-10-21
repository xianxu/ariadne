---
id: 000517
status: open
created: 2026-10-20
updated: 2026-10-20
estimate_hours:
github_issue:
---

# Prepare generated inputs before affected graph discovery

## Problem

Pair's fresh affected graph checkout has no generated runtime bundle. Native go list -test ./... fails at cmd/internal/runtimebundle/embed.go:8:41 with pattern assets/runtime/files: no matching files found. Running the existing runtimebundle generator in that same checkout makes the graph load (408 entries measured on 2026-10-20). Every affected run currently falls back to a full suite, including docs-only changes. Pair's command runs after graph discovery, too late to establish the prerequisite.
