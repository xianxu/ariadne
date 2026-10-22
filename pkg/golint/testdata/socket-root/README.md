# Pair socket-root source capture

These are verbatim Git blobs of
`pair/cmd/internal/testcouch/agentfixture/resources_test.go`:

- `pair-unsafe.go.txt`: commit `39625e2c9`, before pair#792 BR-3 corrected the inherited root.
- `pair-safe.go.txt`: commit `e6535a0c2`, corrected short-root fixture with unrelated long-TMPDIR construction.

Source provenance: local Pair Git object database, read with `git show COMMIT:PATH`.
The reported real error in ariadne#549 is:
`isolated message socket exceeds 103 bytes; choose a shorter COUCH_ISOLATED_ROOT`.
That sentence is issue evidence, not a claimed verbatim terminal capture. Both
source captures are replayed by TestSocketFixtureRootRecordedPairSource through
the real catalog in the normal suite; the unsafe finding must locate the
isolation-root setting, not the hashed namespace argument.
