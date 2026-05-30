# Workflow System

The ariadne workflow is an issue-based development loop designed for AI-assisted software engineering. It's codified in `AGENTS.md` (the "constitution") and guarded by the `sdlc` checkpoint binary. `Makefile.workflow` remains a compatibility wrapper for downstream repos that have not built `sdlc` yet.

## Entries

- [Issue Lifecycle](issue-lifecycle.md) — how work flows from GitHub issue to completion
- [Artifact Hierarchy](artifact-hierarchy.md) — where things live and when they move
- [Pre-merge Checks](pre-merge-checks.md) — constitution enforcement via agent-driven review
- [Issue Sync](issue-sync.md) — syncing issue state to main from any branch with `sdlc claim`
- [Directory Conventions](directory-conventions.md) — the `workshop/` structure and why
- [Sandbox](sandbox.md) — Claude Code sandbox vs OpenShell container sandbox, zellij multiplexer
- [OpenShell Sandbox](openshell-sandbox.md) — containerized dev environment setup, daily use, base layer integration
- [Base Layer](base-layer.md) — how to adopt ariadne's base layer, modes, path conventions, runtime artifacts
- [Setup & Replication](setup-and-replication.md) — `construct/setup.sh` mechanism, symlink-only model, bootstrap cascade, data dependencies (content peers)
- [Construct: Adaptation is Ariadne-Only](construct-adaptation.md) — why only ariadne runs `/construct adapt`; how derivatives inherit via `construct/adapted/`
- [Ledger Landscape](ledger-landscape.md) — where state and evidence live across all surfaces; principles for picking the right ledger
- [sdlc Binary](sdlc-binary.md) — checkpoint-guard binary; 11 verbs; embedded help; fresh-context judges
- [Data Artifacts](data-artifacts.md) — typed markdown documents via the xx-datatype skill
- [Introspection](introspect.md) — postmortem mining of Claude transcripts into auto-loading taste-rule skills
