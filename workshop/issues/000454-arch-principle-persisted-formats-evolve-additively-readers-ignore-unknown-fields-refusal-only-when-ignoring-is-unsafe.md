---
id: 000454
status: codecomplete
deps: []
github_issue:
created: 2026-10-18
updated: 2026-10-18
estimate_hours:
collaboration-mode: collaboration
card_mirror: 'e61f703fdbe9184fa98227fedd1beb3fed24394f' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-18T21:03:58-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:3
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: quick, provenance: inferred, spec: "40ed78ad", done: "d2d585d6"}
actual_hours: 0.29
---

# ARCH principle: persisted formats evolve additively; readers ignore unknown fields; refusal only when ignoring is unsafe

## Problem

The ARCH registry has no principle for how persisted formats evolve. ARCH-SECURE covers reading input written by an *older* version, but nothing says what an older reader should do with data from a *newer* writer. Designs default to refusing unknown records (pair#684's first draft, precedent #198). The operator (2026-10-18): "typical robust strategy would be compatibility, that old binary will just ignore new attributes ... both simpler and also more robust ... if this is not your default, I wonder if we should add [it] to architecture".

## Spec

Add a principle to the registry (name to settle, e.g. ARCH-EVOLVE): persisted formats (records, ledgers, manifests, config) evolve additively. A reader tolerates fields it doesn't know **and preserves them**: it keeps working, and any read-modify-write writes the unknown members back unchanged (a tolerant writer that drops them silently erases newer data; found by pair:1 designing pair#684: several CLI paths rewrite records in whatever build a slot runs). Types, required fields and duplicate keys stay strict; a writer never changes the meaning of an existing field. A version bump, with older readers refusing, is reserved for the cases where ignoring a new field would be unsafe (it changes how existing fields must be interpreted, or it guards a destructive action an old reader would otherwise take), and the design must say why. at-plan prompt: for each persisted format the design changes, say whether the change is additive, and if a reader must refuse, why ignoring is unsafe. Delivered through sdlc arch-principles, start-plan and the plan-quality judge like the others.

## Done when

- The principle is in the registry, printed by sdlc arch-principles and start-plan.
- The plan-quality judge cites it on a design that adds a field with a refusal but no stated reason (a test fixture).

## Plan

- [x] Add red-first contract and delivery guards for ARCH-EVOLVE, including
  arch-principles, start-plan and the actual plan-quality prompt; keep the
  deliberate marker tripwire and affected golden snapshots current.
- [x] Add the registry entry with quick-flow inclusion, additive evolution,
  unknown-member capture and preservation through rewrites, strict validation
  of known fields/required members/duplicate keys, and a justified unsafe-ignore
  exception. Reuse every existing registry consumer (ARCH-DRY).
- [x] Commit self-contained plan fixtures and an opt-in live judge conformance
  test through the existing dispatcher in a private checkout; verify a refusal
  without an unsafe-ignore reason is cited under the new principle, with
  preserving-additive and justified-unsafe controls. Normal suites stay offline.
- [x] Update README and atlas and run targeted checks. Dispatch the recorded
  suite before close; review may overlap its queue, but landing requires its pass.


## Log

### 2026-10-18
- 2026-10-18: closed — BR-1 red reproduced both stale-count failures (44.816s); all rendered numeric count assertions now derive from registry/shared renderer. Corrected command/start-plan/flow tests pass (44.209s); full golint 1.277s, judge 12.919s, processmanual 0.171s pass. BR-2 maps updated and full manual regenerated. Narrow hardcoded-count lint red/green added. Prior actual three-case Codex conformance passed (52.904s): unjustified refusal gets ARCH-EVOLVE; both controls get zero findings. Corrected-head full suite dispatched, must pass before landing. (done-when: contract moved since change-code; judged by the close review); review verdict: SHIP

- Claimed and started in ariadne:3. Hub dispatch specifies preserving unknown
  fields through rewrites, with strict known-field validation and duplicate-key
  checks; the existing issue Spec already records that operator requirement.
- Proposed marker ARCH-EVOLVE and quick-flow=yes to the hub before code. An
  alternative that expands ARCH-SECURE would mix provenance validation with
  compatibility across versions; a separate registry entry keeps their lenses
  distinct. Duplicating requirements in each consumer would violate ARCH-DRY.
- Scope is the registry and its tested delivery, not migration of existing
  persisted formats. No new runtime state, locks, file formats, credentials or
  hot path are introduced. The live test uses disposable private fixture state
  and the existing agent dispatcher; regular suites use deterministic fixtures
  and prompt construction. New prompt bytes scale linearly with one entry.

### 2026-10-18 — implementation and observed judge conformance

- Hub approved ARCH-EVOLVE, quick-flow=yes and the three-fixture live test.
  All runtime consumers already derive from the embedded registry; no new
  delivery mechanism was introduced. Updated the deliberate marker/quick-set
  tests, counts, and only affected architecture blocks/lists in six goldens.
- Red: registry, arch-principles and start-plan guards failed because ARCH-EVOLVE
  was absent. The first attempt also saw a concurrent fixture write in the
  repository mutation guard; a stable-tree rerun reproduced only expected
  missing-principle failures before implementation. No tests and writers overlap
  in subsequent verification.
- Green: complete judge package passed in 16.860s with the model-calling test
  skipped by default. Focused real arch-principles command-tree and start-plan
  delivery checks passed in 7.232s.
- Observed live conformance: `SDLC_LIVE_ARCH_EVOLVE=1 go test
  ./cmd/sdlc/internal/judge -run '^TestLiveArchitectureEvolve$' -count=1 -v
  -timeout 20m` passed in 52.904s using Codex through the existing dispatcher in
  private committed fixture checkouts. The optional-field-refusal fixture raised
  one Important finding: "ARCH-EVOLVE: Optional display metadata does not justify
  incompatible version admission". Its rationale identifies retained label/URL
  semantics and the missing unsafe-ignore reason. The additive-round-trip and
  unsafe-action-guard controls each produced zero findings. These are actual
  parsed model responses, not canned judge output.
- README and atlas describe the policy, quick-flow inclusion and opt-in command.
  The policy preserves unknown nested members and strict known-field validation;
  this item changes no existing storage codec or version.

### 2026-10-18 — recorded full-suite result

- Final runtime head passed recorded attempt `97fd6c380961433cbc2a260db6b0b191`: 12 shards over 1,658 cmd/sdlc tests and 70 other packages, every exit 0 and zero failures, wall 776.4s after a 2,661s machine-queue wait. Saved suite log SHA-256 prefix `d965264a71cf`; recorded as `sdlc-local/suite`. No retries or gate waivers.

## Revisions

### 2026-10-18 — verification bookkeeping

Plan checkboxes track completed implementation deliverables and targeted checks;
the recorded full suite is dispatched before the close review and may queue in
parallel under the worker contract. It must pass on the final code before landing;
no test or review gate is waived.

### 2026-10-18 — round-one registry consumer sweep

The review found stale numeric header assertions in two flow tests (BR-1) and
outdated architecture maps/manual (BR-2). These are the enumerable consumers of
registry expansion: rendered-header assertions, the independent membership and
quick-policy tripwires, prompt snapshots, and architecture map/generated manual.
The remaining numeric Go assertions now derive from the shared registry renderer
or marker count, while both intentional policy tripwires remain independent.
The architecture maps describe these exceptions accurately; the full process
manual was regenerated from current source. Added the narrow mechanical catalog
guard required by the worker contract, with red cases for both eight and nine.

The queued first suite attempt `e97df7bb0c7b87da36334f52f3a47f61` was cancelled
before tests started so the final corrected code gets its own recorded run.
Own red reproduction of both flow failures took 44.816s; corrected flow tests
passed in 38.347s. The lint red cases fail before its matcher is added. These
are test/documentation corrections within the approved policy scope; no registry
semantics or runtime delivery mechanism changed.

- Final targeted verification after the sweep passed: full golint 1.277s, full
  judge package 12.919s, processmanual 0.171s, and the combined real command/
  start-plan/flow checks 44.209s. The model-calling test remains opt-in; its
  earlier three-case observed conformance is unchanged by these test/doc fixes.
- Regeneration used `go run ./cmd/sdlc process-manual --full --out
  atlas/process-manual.md`; a deterministic trailing-whitespace normalization
  removed an empty-window placeholder's trailing space before diff checking.
