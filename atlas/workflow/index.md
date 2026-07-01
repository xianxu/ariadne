# Workflow System

The ariadne workflow is an issue-based development loop designed for AI-assisted software engineering. It's codified in `AGENTS.md` (the "constitution") and guarded by the `sdlc` checkpoint binary. `Makefile.workflow` remains a compatibility wrapper for downstream repos that have not built `sdlc` yet.

## Entries

- [Issue Lifecycle](issue-lifecycle.md) — how work flows from GitHub issue to completion
- [Artifact Hierarchy](artifact-hierarchy.md) — where things live and when they move
- [Pre-merge Checks](pre-merge-checks.md) — constitution enforcement via agent-driven review
- [CI Merge-check](ci-merge-check.md) — pluggable server-side publish gate (seed shim + symlinked runner + scaffolded `merge-checks.d/`); deterministic checks, complements the LLM judges
- [Issue Sync](issue-sync.md) — syncing issue state to main from any branch with `sdlc claim`
- [Directory Conventions](directory-conventions.md) — the `workshop/` structure and why
- [Sandbox](sandbox.md) — Claude Code sandbox vs OpenShell container sandbox, zellij multiplexer
- [OpenShell Sandbox](openshell-sandbox.md) — containerized dev environment setup, daily use, base layer integration
- [Colima VM](colima-vm.md) — `make colima` Linux VM testing (the tart counterpart); shared `vm-log.sh` step/log helper
- [Base Layer](base-layer.md) — how to adopt ariadne's base layer, modes, path conventions, runtime artifacts
- [Setup & Replication](setup-and-replication.md) — `construct/setup.sh` mechanism, symlink-only model, bootstrap cascade
- [weave](weave.md) — the layer-composition compiler that replaced `setup.sh`: pure intent pipeline, per-harness face lowering, prune / golden / verify-complete
- [Harness Integration](harness-integration.md) — the Compile(C,T)/Union model, per-harness faces (claude→`CLAUDE.md`+`.claude/skills`; codex/gemini→`AGENTS.md`/`GEMINI.md`+`.agents/skills`), and the per-harness assumption ledger
- [Data Dependencies](data-deps.md) — content peers (looser git submodule): sibling clone + relative symlink via `construct/deps` `data` rows (#60); how to add / remove / bootstrap
- [Construct: Adaptation is Ariadne-Only](construct-adaptation.md) — why only ariadne runs `/construct adapt`; how derivatives inherit via `construct/adapted/`
- [Ledger Landscape](ledger-landscape.md) — where state and evidence live across all surfaces; principles for picking the right ledger
- [sdlc Binary](sdlc-binary.md) — checkpoint-guard binary; 10 verbs (incl. the `issue` group); embedded help; fresh-context judges
- [Data Artifacts](data-artifacts.md) — typed markdown documents via the xx-datatype skill
- [Introspection](introspect.md) — postmortem mining of Claude transcripts into auto-loading taste-rule skills
- [sdlc retro](retro.md) — the process manual: `sdlc retro` unrolls every always-on injection source (sdlc prompts, help text, skills, lessons, AGENTS chain, memories) into one linked markdown doc (#153, M1)
- [docflow](docflow.md) — branch-scoped prose review with per-round git journaling; companion to the `xx-fix` skill (`--no-ff` merge keeps the back-and-forth + rationale, `--first-parent` stays clean)
