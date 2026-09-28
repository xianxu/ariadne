# Ariadne Workshop Extensions

This is the ariadne repo — where the base layer itself is developed. If you're
changing base-layer files here (anything listed in `construct/base.manifest`),
it propagates to downstream repos, so weigh downstream impact. See
`atlas/workflow/base-layer.md` for how the base layer, manifest, and the
`.ariadne.`/`.local.` settings merge work.

Run the sdlc Go suite with `make test` (sharded across processes, about 4
minutes); a serial `go test ./cmd/sdlc/...` takes about 25 and needs
`-timeout 60m`. See `atlas/workflow/sdlc-binary.md` → "Running the test suite".
