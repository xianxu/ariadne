# Boundary Review — ariadne#294 (whole-issue close)

| field | value |
|-------|-------|
| issue | 294 — layergraph: export DeclaredSubstrates, including absent substrates |
| repo | ariadne |
| issue file | workshop/issues/000294-layergraph-export-declaredsubstrates-including-absent-substrates.md |
| boundary | whole-issue close |
| milestone | — |
| window | 2a839ee2e84112dfeb38690a3aa0704693c5a537..866aac6358ce176d27004a1bd4488e777212b328 |
| command | sdlc close --issue 294 |
| reviewer | claude |
| timestamp | 2026-10-05T15:24:50-07:00 |
| verdict | SHIP |

## Review

Tests pass and the package is stdlib-only. Finishing the review write-up now.

```verdict
verdict: SHIP
confidence: high
```

This change delivers what the Spec and every Done-when item ask for, except landing on `main`, which comes after close. `declaredGraph` (`pkg/layergraph/walk.go:66`) is now the only BFS. `Walk` reads its edge map through `discoverEdges`, and `DeclaredSubstrates` reads its declared list. The two share self-exclusion, the loud error for a present substrate with no manifest, and `substrateTargets`' resolution. Because of that they can't diverge, and `TestWalkIsThePresentSubsetOfDeclaredSubstrates` pins it.

I checked it myself:
- `go test ./pkg/layergraph/` passes.
- `go vet` is clean.
- `go list -deps` shows only the package itself outside stdlib.
- The only real consumers of `Walk` are weave (its own FS) and datatype and vocabulary (`OSFS`).

Nothing blocks shipping. The minor findings are below.

1. **Strengths**
   - One traversal, two views (`walk.go:51-62`). It's ARCH-DRY by construction rather than two copies that happen to agree.
   - `substrateTargets` now uses `ParseRows` (`walk.go:140`) instead of `ParseDeps`. Grammar and errors are unchanged: `ParseDeps` is a projection of `ParseRows` (`deps.go:48`). It also carries `Source` with no second parser.
   - `readDeclaration` (`walk.go:170`) bounds reads without changing the `FS` interface. The optional `DeclarationReader` lets weave's FS and the test fakes keep working.
   - In `plainFS`, a method with a different signature shadows the promoted `ReadDeclaration`. That lets the bounded fallback be tested for real against a non-reader FS.
   - The tests use real temp directories through `OSFS` and check behaviour (discovery order, owner, source, present), not the implementation.

2. **Critical:** none.

3. **Important:** none.

4. **Minor**
   - **The `DeclarationReader` path turns more errors loud (`walk.go:171-176`).** Only `ErrNotExist` still means "no declaration". With `OSFS`, any other error now aborts the walk, where before the declaration was silently treated as absent. That covers a permission error, `ENOTDIR` when `construct` is a file, and a symlinked `construct/deps`. This changes `Walk`'s behaviour for datatype and vocabulary too. It's arguably right (ARCH-SECURE: degrade visibly), but it isn't in the Spec or the Log; it should be recorded as an intended change.
   - **The plain-FS fallback reads the whole file before checking the limit (`walk.go:178-184`).** It bounds what gets parsed, not what gets read into memory. That's fine for fakes and weave's FS, but the comment says "bounded".
   - **`DeclaredSubstrates` doesn't state that `root` must be absolute**, though `Walk` does. A relative root gives a relative `Owner`, which breaks the doc's "absolute" promise.
   - **The test file differs from the plan.** The plan names `walk_test.go`; the tests are in the new `declared_test.go`.

5. **Test coverage notes**
   - Every case in Done-when has a test: transitive chain, absent substrate, source column, malformed row, unresolvable parent, dedup with the first owner kept, the Walk-subset property, and an oversized declaration. There's also a test that the root is never its own substrate.
   - Missing: a test that a non-`ErrNotExist` read error through `OSFS` is reported, not swallowed. Only the symlink case is tested.

6. **Architectural notes**
   - ARCH-DRY: pass.
   - ARCH-PURE: pass. The IO stays behind the `FS` seam; `physical` is a real-disk concern that was already there.
   - ARCH-PURPOSE: pass. The consumer (pair#387) gets owner, source and presence without re-implementing resolution.
   - ARCH-MOCK: pass. Filesystem only, using the existing seam.
   - ARCH-CONSTRAINTS: pass. Reads are bounded, the BFS is linear, and dedup uses `seen`.
   - ARCH-SECURE: pass. Declaration reads go through the no-follow, regular-file-only reader. Note the intentional loudening in Minor #1.
   - ARCH-ORDER: pass. It's a single synchronous walk that keeps no state between events.
   - ARCH-FUNERAL: pass. It creates nothing durable.
   - Pair will consume `DeclaredSubstrate` as a public API. Adding fields later is fine; renaming `Present` or `Owner` would break pair's pin.

7. **Plan revision recommendations**
   - A `## Revisions` (or Log) line saying that under `OSFS`, unreadable, symlinked or non-regular `construct/deps` files now raise errors for every `Walk` consumer, where they used to be silently skipped.

```findings
findings:
  - id: new
    severity: Minor
    family: undocumented-behavior-change
    title: |
      DeclarationReader path makes non-ENOENT construct/deps read errors loud for all Walk consumers
    detail: |
      Under OSFS, permission, ENOTDIR, and symlinked construct/deps now abort Walk in datatype and vocabulary, where they were silently treated as absent. Arguably correct, but not in the Spec or Log; record it and add a test for one non-symlink error.
  - id: new
    severity: Minor
    family: bounded-read-claim
    title: |
      Plain-FS fallback reads the whole file before applying DeclarationLimit
    detail: |
      readDeclaration's fallback bounds what is parsed, not what is read into memory. Fine for fakes and weave's FS; the comment should say so.
  - id: new
    severity: Minor
    family: api-precondition-doc
    title: |
      DeclaredSubstrates does not say root must be absolute, unlike Walk
    detail: |
      A relative root gives a relative Owner, contradicting the field doc's absolute-path promise.
```
