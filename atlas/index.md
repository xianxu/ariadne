# Ariadne Atlas

## Overview
Central directory for atlas entries — practical pointers for future developers and agents to understand the sketch of functionalities, history, and intention. Details live in the code.

## 1. Workflow System (base layer)
- [Workflow Index](workflow/index.md) — issue-based development loop
- [Issue Lifecycle](workflow/issue-lifecycle.md) — GitHub → local → archive flow; closing checklist (actual_hours, side-quest log, auto calibration-ledger append)
- [Artifact Hierarchy](workflow/artifact-hierarchy.md) — where workshop/ files live + lifecycle
- [Ledger Landscape](workflow/ledger-landscape.md) — where state and evidence live across all surfaces (issue file, git history + trailers, transcripts, memory, atlas, project file); design principles for picking the right ledger
- [Pre-merge Checks](workflow/pre-merge-checks.md) — constitution enforcement
- [Directory Conventions](workflow/directory-conventions.md) — standard repo layout
- [sdlc Binary](workflow/sdlc-binary.md) — unified checkpoint-guard binary (`cmd/sdlc/`) guarding the workflow from claim → ship (incl. the `issue` group, `active-time` the in-binary v3 attribution engine #110, and the read-only `resolve`/`open` ref resolver #144, and `migrate` — cross-repo artifact move with ref rewrite #179) replacing the Make-target surface; embedded `--help` per subcommand; fresh-context judges for anti-collusion
- [Gate State](workflow/gate-state.md) — how an SDLC gate remembers what it asked for: the `finding` CUE noun, the schema'd ` ```findings ` handoff, binary-assigned stable ids, the pure block/pass decision, the content-hash pass-through, and the plan-gate → boundary-review carry-forward, which #194 delivered as a SEED into the boundary review's own ledger (#187, #194)
- [Architecture Principles](workflow/architecture-principles.md) — single-source `ARCH-*` registry pushed through planning/review gates and `sdlc arch-principles`; includes stateful external doubles and explicit runtime operating envelopes.
- [Sandbox](workflow/sandbox.md) — Claude Code sandbox vs OpenShell container sandbox, zellij multiplexer usage
- [OpenShell Sandbox](workflow/openshell-sandbox.md) — the containerized dev sandbox in the workflow: setup, what's inside, git transport (HTTPS-not-SSH, #152), base-layer provisioning
- [Data Artifacts](workflow/data-artifacts.md) — typed markdown documents (xx-datatype skill, prototypes, capture flow)
- [Introspection](workflow/introspect.md) — postmortem mining of past agent transcripts (Claude + codex) into auto-loading taste-rule skills (xx-introspect + introspect-&lt;activity&gt;)
- [sdlc process-manual](workflow/process-manual.md) — the process manual: `sdlc process-manual` regenerates one linked markdown doc unrolling every always-on injection source (sdlc prompts, help text, skills, lessons, AGENTS chain, memories); deterministic, navigate-to-source, `🤖[]`-editable (#153: M1 static catalog + M2 judge-prompts-as-embedded-markdown). `--session <jsonl|current>` reconstructs which injection points actually *fired* in a session, in order, matched to the catalog (#157)
- [docflow](workflow/docflow.md) — branch-scoped prose review with per-round git journaling (`review/<slug>` branch, `--no-ff` merge); companion to the `xx-fix` skill (#79)
- [Session Retro](workflow/session-retro.md) — exported skill for evidence-backed development-process findings from current or supplied session evidence
- [Tart VMs](workflow/base-layer.md#what-gets-installed) — `make tart` family for macOS VM testing (Apple Silicon only). Vendored under `.tart/`; details in the base-layer entry.
- [Colima VMs](workflow/colima-vm.md) — `make colima` family for clean **Linux** VM testing (the tart counterpart). Profile-per-repo, live workspace mount, VNC GUI path. Vendored under `.colima/`.

## 2. Base Layer Infrastructure
- [Base Layer](workflow/base-layer.md) — how to adopt ariadne's base layer, path conventions, runtime artifacts
- [Setup & Replication](workflow/setup-and-replication.md) — `construct/setup.sh` mechanism, ancestor discovery, bootstrap cascade
- [weave compiler](workflow/weave.md) — `cmd/weave`, the intent compiler **replacing `setup.sh`** (in progress, #95): layer DAG → composed `AGENTS.md` + served skills; pure core, `construct/deps`-only edges, hybrid intents, symlink-only
- [Harness integration](workflow/harness-integration.md) — how weave serves many agent harnesses (Claude/Codex/Gemini) from ONE checkout: the `Compile(C,T)`/Union model, the per-harness face map, the assumption ledger + its guard (`make harness-check`), onboard/triage runbooks (#107)
- [pkg/layergraph + pkg/frontmatter](../pkg/) — module-level shared libraries (#115 M1): the transitive `construct/deps` layer-graph walk (`pkg/layergraph`, imported by weave, the `datatype` binary, and (via `MergeByName`, #122) `cmd/vocabulary` — the "single mechanism" so DAG-aware tools never diverge on topology) + a flat-YAML `description:` parser (`pkg/frontmatter`). The DAG-merged datatype prototypes + per-repo gitignored dynamic-skill materialization (#115) are documented in [Data Artifacts](workflow/data-artifacts.md) + [weave atlas → Dynamic skills](workflow/weave.md).
- [Vocabulary Layer](workflow/vocabulary.md) — formal CUE models of nouns + lifecycles (`construct/vocabulary/`); the single source consumers derive from. `issue.cue` is the first (#122, M1–M4 landed: model + `cmd/vocabulary` + weave wiring + `pkg/vocab` Go binding that sdlc consumers derive from + enforced lifecycle gate at `set-status`). #124 adds **instance-conformance** (`vocabulary validate-instance` / `sdlc issue validate` + a fail-closed push/merge gate) + a second noun (`pensive.cue`). Propagates like datatype.
- [Data Dependencies](workflow/data-deps.md) — content peers (looser git submodule): sibling clone + relative symlink via `construct/deps` `data` rows (#60); add / remove / bootstrap
- [base.manifest](../construct/base.manifest) — canonical list of portable paths
- [setup.sh](../construct/setup.sh) — bootstrapper for consuming repos

## 3. Brain Convention
The brain-manifest convention is defined in `AGENTS.md` §1 (lives here so it travels with the substrate). Detailed implementation pointers (nous#3, #4, #6, charon#21) and the lifecycle doc live downstream with the implementation — see [nous/atlas/nous/brain-manifest.md](../../nous/atlas/nous/brain-manifest.md).

## 4. Vision & Strategy
- [Founding Context](../docs/vision/founding-context.md) — core thesis, product progression, bootstrapping strategy, 3-month focus
- [Pitch Deck](../docs/plans/pitch-deck.md) — seed-stage funding pitch deck outline

## 5. Product Specs
*(To be created as the product takes shape)*
