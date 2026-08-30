# Architecture Principles

The canonical `ARCH-*` registry lives in
`cmd/sdlc/internal/judge/architecture.md`. It is embedded into plan-quality and
boundary-review prompts, `sdlc start-plan` pushes it into planning sessions, and
`go run ./cmd/sdlc arch-principles` renders it for non-gate design work.

Key consumers:

- `cmd/sdlc/internal/judge/architecture.go` extracts markers and renders the
  registry block.
- `cmd/sdlc/startplan.go` and `cmd/sdlc/archprinciples.go` are the planning-time
  push and standalone pull paths; both reuse `ArchitectureBlock`.
- `cmd/sdlc/internal/judge/judge_test.go` pins marker extraction and prompt
  embedding.
- `cmd/sdlc/internal/judge/testdata/golden/*.prompt` pins the generated prompt
  bodies that carry the registry. A deliberate registry edit re-captures these
  (`-update-golden`); the ⛔ in `golden_test.go` forbids re-capturing to paper
  over *drift*, which is a different case.
- `cmd/sdlc/gatefindings.go` routes every fixer-facing findings refusal to
  `ARCH-PURPOSE` (#203) — the judges get the registry inlined because a marker
  alone would dangle in a fresh context, but these lines are read by the main
  thread, which already holds the block from `sdlc start-plan`. Guarded by
  `TestFixTheClassLine_RoutesToArchPrinciples`, which asserts the ROUTING and
  never the wording: asserting wording is what would let the line become a
  second copy of the principle.

`ARCH-PURPOSE` covers two axes of the same rule: deliver the issue's purpose
rather than its easy subset, and answer a review finding with the *class* it
names rather than the one site — a finding's named site being the easy subset
again. That is why #203 extended it instead of coining a fifth marker.

`ARCH-MOCK` codifies the external dependency rule: every relied-on external
binary/service should sit behind a seam with a stateful fake for integration and
end-to-end tests, plus live conformance checks where practical to keep the fake
honest against the real dependency. For owned libraries, services, and binaries,
the storage/backend layer should also boot from portable file folders and/or
database configuration without depending on production configuration or
production databases.

`ARCH-CONSTRAINTS` makes runtime behavior an explicit design input. Planning
classifies the workload/interaction path and records each material parameter's
budget or range, basis, and bounded behavior when exceeded; review checks the
implementation and representative measurements against that operating envelope.
Its domain prompts cover latency, workload growth, CPU/memory/IO, concurrency,
environment/co-tenancy, and overload without imposing universal numeric defaults.
