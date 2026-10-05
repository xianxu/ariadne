# Boundary Review — ariadne#295 (whole-issue close)

| field | value |
|-------|-------|
| issue | 295 — layergraph: a present substrate without base.manifest is a typed NotLayerError |
| repo | ariadne |
| issue file | workshop/issues/000295-layergraph-a-present-substrate-without-base-manifest-is-a-typed-notlayererror.md |
| boundary | whole-issue close |
| milestone | — |
| window | 6191354435032f7244790b92233983e969e23788..6e8e4d8ca4f328fb6b912fdd650df2092f4c6f35 |
| command | sdlc close --issue 295 |
| reviewer | claude |
| timestamp | 2026-10-05T16:02:13-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

This change does what the issue set out to do, and I found nothing that blocks shipping. `declaredGraph` now returns `&NotLayerError{Path: dep, Owner: cur}` at the #155 site (`pkg/layergraph/walk.go:113`). `Walk` and `DeclaredSubstrates` pass it through without wrapping it, so `errors.As` works in both. `Error()` (`walk.go:135-140`) uses the same format string and the same three arguments as the old `fmt.Errorf`, so the message is byte-identical. The new test checks the message against a literal. All three `## Done when` clauses that apply to this boundary are covered. Typed error from both views, with Path/Owner: `notlayer_test.go:13`. Malformed row and unreadable `construct/deps` are both still errors, and neither is a `NotLayerError`, through both views: `notlayer_test.go:36`. The "landed on main" clause belongs to the merge step, not this review. `go test ./pkg/layergraph/` passes. The atlas entry was updated (`atlas/workflow/weave.md:162`). The Log records why the optional `Layer: false` variant was declined.

1. **Strengths**
   - `Error()` keeps the original format and argument order exactly, and the test builds the expected string independently rather than calling `Error()`, so any drift in the message fails it.
   - The test runs both views (`Walk` and `DeclaredSubstrates`) against both kinds of other failure (malformed row and unreadable `construct/deps`), so each two-sided `## Done when` clause is checked on both sides.
   - Reverting the fix makes the test fail: with the untyped error back, `errors.As` fails in `TestNotLayerError`.
   - The package still imports only the standard library, and no other error changed type.
   - The new comment on `DeclaredSubstrates` ("Errors are Walk's: … a present substrate without construct/base.manifest …") is still accurate.

2. **Critical:** none.
3. **Important:** none.
4. **Minor:** none worth raising.
5. **Test coverage:** this is adequate for a change that only adds a type. The unreadable case makes `construct/deps` a directory, which the bounded read rejects. The test requires `err != nil`, so it isn't passing vacuously.
6. **Architecture:**
   - **ARCH-DRY: pass.** The message now lives in one place, `Error()`. The only other copy is the deliberate literal in the test, and no other code in the tree builds this string.
   - **ARCH-PURE: pass.** It is a plain value type with a pure `Error()`. Nothing new does IO.
   - **ARCH-PURPOSE: pass.** The consumer the issue names, pair#387, needs `errors.As` from `DeclaredSubstrates`, and it now works. Both views are covered, and nothing that is the point of the issue was deferred.
7. **Plan revisions:** none. The plan matches the code.

```findings
```
