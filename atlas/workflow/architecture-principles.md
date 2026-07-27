# Architecture Principles

The canonical `ARCH-*` registry lives in
`cmd/sdlc/internal/judge/architecture.md`. It is embedded into plan-quality and
boundary-review prompts, and `go run ./cmd/sdlc arch-principles` renders it for
non-gate design work.

Key consumers:

- `cmd/sdlc/internal/judge/architecture.go` extracts markers and renders the
  registry block.
- `cmd/sdlc/internal/judge/judge_test.go` pins marker extraction and prompt
  embedding.
- `cmd/sdlc/internal/judge/testdata/golden/*.prompt` pins the generated prompt
  bodies that carry the registry.

`ARCH-MOCK` codifies the external dependency rule: every relied-on external
binary/service should sit behind a seam with a stateful fake for integration and
end-to-end tests, plus live conformance checks where practical to keep the fake
honest against the real dependency. For owned libraries, services, and binaries,
the storage/backend layer should also boot from portable file folders and/or
database configuration without depending on production configuration or
production databases.
