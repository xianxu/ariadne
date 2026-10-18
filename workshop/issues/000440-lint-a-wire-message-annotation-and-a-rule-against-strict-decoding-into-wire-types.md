---
id: 000440
status: working
deps: []
github_issue:
created: 2026-10-18
updated: 2026-10-18
estimate_hours:
collaboration-mode: collaboration
card_mirror: '44f4523b6ce37631fa2c9657f8c1ca30df374a9c' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-18T05:43:21-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:1
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: quick, provenance: inferred, spec: "6499ed27", done: "0f940315"}
---

# lint: a wire-message annotation and a rule against strict decoding into wire types

## Problem

Strict JSON decoding of cross-version peer or wire messages (`DisallowUnknownFields`) breaks mixed-version fleets (pair#594), but the same call is right for local files. A syntactic rule can't tell them apart, so #407 deferred it.

## Spec

Mark named Go types decoded from independently evolving JSON producers with a
type-doc comment `// wire: <why the producer can evolve separately>`. The lint
catalog builds one syntax-only index of those declarations across the module.
It flags a direct `encoding/json.Decoder` chain when a decoder created by
`json.NewDecoder` calls `DisallowUnknownFields` and then `Decode` into a marked
type. Report the finding at `Decode`, where the destination is known. Do not
flag local-file and version-pinned records merely because they use strict
decoding.

The matcher follows imported aliases, package files, local typed variables,
composite literals, and pointer targets. It stays conservative on control flow:
if a decoder or target is rebound, or its provenance is ambiguous, it emits no
finding. The annotation is an explicit opt-in; the atlas explains when new
types need it. `ARCH-DRY` keeps one index for all files; `ARCH-PURE` keeps the
index and matcher free of filesystem reads and type-checking.

Annotate these existing tolerant JSON inputs: `judge.streamEvent` (Claude
stream-json), `main.goTestJSONEvent` (`go test -json`), `usage.record` (Codex and
Claude transcript lines), and `usage.codexUsage` (Codex usage payload). Do not
annotate tracker receipts, fleet policy, or issue observations: those are
version-pinned persisted contracts where strict decode is intentional.

## Done when

- An annotation convention marks wire message types (for example a `// wire:` comment), and a pkg/golint rule flags strict decoding into a marked type, with flag/no-flag tests citing pair#594.
- The four existing tolerant input types named in Spec carry the annotation;
  atlas states when a new type needs one and why strict local-file decoding is
  still valid.
- Tests cover cross-file and imported marked types, alias imports, pointer
  targets, direct strict/non-strict decoder chains, ambiguous or rebound
  decoders, and a counted N/2N source-index workload.

## Plan

- [x] Approve the durable plan and pass `sdlc change-code`.
- [x] Add the red lint fixtures, implement the pure index and decoder matcher,
  and annotate Ariadne's current tolerant input types.
- [ ] Update atlas, run focused tests and `make -k`, and record the pushed-head
  suite for close.

Then close and land on SHIP.

## Log

### 2026-10-18

- Claimed and started in ariadne:1. The hub approved a syntax-only source
  index and direct decoder-chain rule, and asked that Ariadne's existing wire
  types be annotated in this issue; Pair adopts the rule in pair#665.
- `sdlc change-code` admitted the quick flow, with the boundary review set to
  upgrade if the implementation leaves its size shell. Red tests first showed
  the missing same-file and cross-file findings, then caught false positives
  from decoder reassignment, conditional flags, shadowed targets, mixed
  packages, and shadowed `encoding/json`. The matcher now keeps only proven
  direct chains; unsupported helper/control-flow cases remain unflagged.
- Focused `go test ./pkg/golint -count=1` passes after the four annotations;
  the repository scan checked 1,172 tracked Go files against six rules in
  331 ms in a measured run. `make -k tools` built all six tools, and
  `git diff --check` passed. The index's counted N=8 and 2N=16 fixture
  confirms one file and one type-declaration visit per fixture file.

## Revisions

- 2026-10-18, close checklist: the last Plan checkbox now names evidence
  attainable before `sdlc close`; close and merge remain mandatory workflow
  actions immediately after that evidence, not a self-referential checkbox.
