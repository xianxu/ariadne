# Ariadne Atlas

## Overview
Central directory for atlas entries — practical pointers for future developers and agents to understand the sketch of functionalities, history, and intention. Details live in the code.

## 1. Workflow System (base layer)
- [Workflow Index](workflow/index.md) — issue-based development loop
- [Issue Lifecycle](workflow/issue-lifecycle.md) — GitHub → local → archive flow; closing checklist (actual_hours, side-quest log, validation-log entry)
- [Artifact Hierarchy](workflow/artifact-hierarchy.md) — where workshop/ files live + lifecycle
- [Ledger Landscape](workflow/ledger-landscape.md) — where state and evidence live across all surfaces (issue file, git history + trailers, transcripts, memory, atlas, project file); design principles for picking the right ledger
- [Pre-merge Checks](workflow/pre-merge-checks.md) — constitution enforcement
- [Directory Conventions](workflow/directory-conventions.md) — standard repo layout
- [sdlc Binary](workflow/sdlc-binary.md) — unified checkpoint-guard binary (`cmd/sdlc/`), 11 verbs (incl. the `issue` group) replacing the Make-target surface; embedded `--help` per subcommand; fresh-context judges for anti-collusion
- [Sandbox](workflow/sandbox.md) — Claude Code sandbox vs OpenShell container sandbox, zellij multiplexer usage
- [Data Artifacts](workflow/data-artifacts.md) — typed markdown documents (xx-datatype skill, prototypes, capture flow)
- [Introspection](workflow/introspect.md) — postmortem mining of past Claude transcripts into auto-loading taste-rule skills (xx-introspect + introspect-&lt;activity&gt;)
- [docflow](workflow/docflow.md) — branch-scoped prose review with per-round git journaling (`review/<slug>` branch, `--no-ff` merge); companion to the `xx-fix` skill (#79)
- [Tart VMs](workflow/base-layer.md#what-gets-installed) — `make tart` family for macOS VM testing (Apple Silicon only). Vendored under `.tart/`; details in the base-layer entry.

## 2. Base Layer Infrastructure
- [Base Layer](workflow/base-layer.md) — how to adopt ariadne's base layer, path conventions, runtime artifacts
- [Setup & Replication](workflow/setup-and-replication.md) — `construct/setup.sh` mechanism, ancestor discovery, bootstrap cascade
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
