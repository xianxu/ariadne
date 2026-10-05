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

---

## Re-review — 2026-10-05T15:29:46-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 294 — layergraph: export DeclaredSubstrates, including absent substrates |
| repo | ariadne |
| issue file | workshop/issues/000294-layergraph-export-declaredsubstrates-including-absent-substrates.md |
| boundary | whole-issue close |
| milestone | — |
| window | 2a839ee2e84112dfeb38690a3aa0704693c5a537..4f436c2d723cf0397eecea61aafd0736d3d6716e |
| command | sdlc close --issue 294 |
| reviewer | claude |
| timestamp | 2026-10-05T15:29:46-07:00 |
| verdict | SHIP |

## Review

I've finished checking the diff and the prior findings. Here is the verdict.

```verdict
verdict: SHIP
confidence: high
```

All three prior Minor findings are fixed, and I found nothing new. `DeclaredSubstrates` and `Walk` now share one traversal (`declaredGraph`), so the two can't diverge. A test pins that `Walk`'s layers are exactly the present entries of `DeclaredSubstrates`. Every Done-when item except landing on main has a matching test (landing happens after close). `go test ./pkg/layergraph` passes, and `go list -deps` shows the package uses only the standard library.

**1. Strengths**
- **One traversal, two views:** `walk.go` `declaredGraph` returns both the edge map and the declared list. Skipping the root and raising the loud error for a substrate that is present but has no `base.manifest` happen once, inside the shared loop.
- **Source column kept:** `substrateTargets` now reads rows through `ParseRows`, filtered to `substrate` rows, so `Source` survives without a second parser.
- **Bounded reads behind an optional interface:** `DeclarationReader` is optional and `OSFS` implements it. Weave's own file system and the test fakes use the fallback without any changes.
- **Real tests:** they use a real temporary file system through `OSFS`, with no mocks. `plainFS` cleverly shadows `ReadDeclaration` so the fallback path also gets tested.

**2. Critical findings:** none.

**3. Important findings:** none.

**4. Minor findings:** none new. One thing to know, already documented in `readDeclaration`: for an unreadable `construct/deps`, weave (fallback path) still treats it as "no deps", while `datatype` and `vocabulary` (`OSFS`) now fail loudly. That difference is intended and written down, not a defect.

**5. Test coverage notes**
- Every case listed in Done-when has a test, plus three more: oversized declarations on both read paths, a refused symlink, and unreadable / relative-root cases.
- The unreadable case uses a directory where the file should be. That covers the "non-symlink error" that BR-1 asked for.

**6. Architecture**

| Principle | Result | Why |
|---|---|---|
| ARCH-DRY | pass | resolution and traversal are shared, not copied |
| ARCH-PURE | pass | IO stays behind the `FS` seam |
| ARCH-PURPOSE | pass | an absent substrate is reported with `Present: false`, which is what pair#387's reconciler needs |
| ARCH-MOCK | pass | real temporary file system, plus the existing fakes |
| ARCH-CONSTRAINTS | pass | declaration reads are bounded; the fallback documents that its limit applies to parsing, not memory |
| ARCH-SECURE | pass | `OSFS` refuses symlinks and FIFOs and enforces the byte limit; failures are loud |
| ARCH-ORDER | pass | it's a one-shot walk that holds no state between events |
| ARCH-FUNERAL | pass | it's a read-only walk and creates nothing durable |

**7. Plan revisions:** none. The issue's Log already records the behavior change and the round's fixes.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      Logged in the issue Log; tested by TestUnreadableDeclarationIsLoudAndRootsAreAbsolute (a directory at construct/deps errors for both Walk and DeclaredSubstrates).
  - id: BR-2
    disposition: addressed
    note: |
      The readDeclaration comment now says the fallback bounds what is parsed, not what ReadFile loads.
  - id: BR-3
    disposition: addressed
    note: |
      DeclaredSubstrates calls filepath.Abs before physical(), the doc says so, and a relative-root test after t.Chdir checks that the paths are absolute.
```
