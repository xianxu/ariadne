# Boundary Review — ariadne#241 (milestone M2)

| field | value |
|-------|-------|
| issue | 241 — Publish weave and cut over startup |
| repo | ariadne |
| issue file | workshop/issues/000241-publish-weave-startup.md |
| boundary | milestone M2 |
| milestone | M2 |
| window | 8640c9dbe1d52257601124d2d138cdc832eba14a..a1e9d0f6f5931d56e2196956932eff0621108078 |
| command | sdlc milestone-close --issue 241 --milestone M2 |
| reviewer | claude |
| timestamp | 2026-09-28T19:16:56-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

M2 does what the plan says. The tag `weave-v0.1.0` points at `4de31b2d`, which is the reviewed commit that bundles the LICENSE. The release and the tap are live, and the Log has evidence for each step: the build-of-record run, the checksum re-check, the empty-trust-store install, test and audit, and parley.nvim's Linux CI installing from the tap. The #250 stopgap was removed only after that verification passed, as the plan required, and the test that covered it was updated in the same change. The docs sweep is complete: a grep for `brew tap`, `ariadne-source`, `pending publication` and `until…publish` outside `workshop/` finds nothing stale. I re-ran `go test ./cmd/weave/internal/release/` and `scripts/test/portable-ci.test.sh` and both pass. Nothing blocks SHIP. The only findings are two Minor provenance/residue notes.

1. **Strengths**
   - `cmd/weave/internal/release/main.go:78-83`: LICENSE is read before `staging.New`, so a missing LICENSE fails with nothing left behind. `TestMissingLicenseFailsBeforeStaging` checks both the `ErrNotExist` error and that no output or stages remain. The fixture writes `weave.rb`, so the error really does come from the LICENSE read.
   - `TestArchiveCarriesBinaryAndLicense` checks every archive member's order, mode and content, and that there is no extra member. `release-weave.test.sh:60-61` checks that the bundled notice is byte-identical to the repo's `LICENSE`. Together they cover the class of bug that motivated this change.
   - `portable-ci.test.sh:45` replaces the old `::warning::`/`#241` assertion with a negative one (`'brew tap' not in workflow and 'ariadne-source' not in workflow`). If someone puts the stopgap back, the test fails.
   - The trust question was settled by reading Homebrew's source and then confirming it with an empty trust store, rather than by adding `brew trust` just in case. README and atlas state which forms need trust (a Brewfile entry, a bare `brew upgrade`) and which don't (the full-name install).
   - `family: license-notice-travels-with-artifact` is covered at every level: the archive members, the formula's `license "MIT"`, the tap repo's LICENSE, and the new README `## License` section.

2. **Critical findings:** none.

3. **Important findings:** none.

4. **Minor findings**
   - The template at the tag still has the old stanza order (`license` before `version`). As a result, the tap's `Formula/weave.rb` (`1d3ecf3`) differs from the build-of-record artifact by a hand swap of two lines. This is logged, and the template fix (`9cf6ef7d`) means future releases follow README step 4 exactly, so no action is needed beyond the existing Log entry.
   - Consumers' seed-once `merge-check.yml` copies keep the dormant `elif ! brew tap` fallback indefinitely (ARCH-FUNERAL: residue with no removal path). It is harmless now that the tap resolves, but nothing retires it.

5. **Test coverage notes:** The archive shape is covered at two levels: the Go unit test and the real cross-build in `release-weave.test.sh`. The CI change is covered by portable-ci for both the published-tap row and the ariadne-source row. The live behaviour (tap install, Linux CI) can only be checked by manual evidence, and the Log records it with run IDs.

6. **Architectural notes**
   - **ARCH-DRY:** pass. The table-driven archive entries replace the duplicated header/write code.
   - **ARCH-PURE:** pass. `writeArchive` takes the license as bytes, and the IO stays in `prepare`.
   - **ARCH-PURPOSE:** pass. Every Done-when item has evidence. Deferred work is filed as #265 and #266 and is outside this issue's purpose.
   - **ARCH-MOCK:** pass. The fake `brew` in portable-ci is simplified along the same seam, and the live check was the real parley.nvim CI run.
   - **ARCH-CONSTRAINTS:** N/A. This is release tooling, not a hot path.
   - **ARCH-SECURE:** pass. Checksums were re-verified from the artifact and the release was built from exactly those files. No credentials appear in the diff.
   - **ARCH-ORDER:** pass. The ordering that matters (tag, then build, then verify, then remove the fallback) was followed. The no-rebase provenance invariant is stated in the Plan.
   - **ARCH-FUNERAL:** flagged Minor only, for the leftover consumer fallback above.

7. **Plan revision recommendations:** none. The Plan's M2 bullets match what was delivered. The stanza-order fix to the tap is already in the Log.

```findings
findings:
  - id: new
    severity: Minor
    family: release-artifact-is-published-verbatim
    title: |
      Tap formula differs from the tag's build-of-record artifact by a hand stanza swap
    detail: |
      The template at weave-v0.1.0 (4de31b2d) has license before version, so brew audit --strict needed a manual edit in the tap (1d3ecf3). The fix in 9cf6ef7d means future releases match README step 4 exactly, and the Log records this one exception.
  - id: new
    severity: Minor
    family: seed-once-residue-has-no-retirement
    title: |
      Consumers' seed-once merge-check.yml keep the dormant #250 tap fallback with no removal path
    detail: |
      This is harmless now that the tap resolves, but seed-once copies never receive the upstream deletion (ARCH-FUNERAL). Consider a one-off sweep when a consumer is next touched.
```
