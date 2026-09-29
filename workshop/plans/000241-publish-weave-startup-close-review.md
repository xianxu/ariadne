# Boundary Review — ariadne#241 (whole-issue close)

| field | value |
|-------|-------|
| issue | 241 — Publish weave and cut over startup |
| repo | ariadne |
| issue file | workshop/issues/000241-publish-weave-startup.md |
| boundary | whole-issue close |
| milestone | — |
| window | bfe260bff8962d64044abcfc5a15d16155fa7930..54e9cc5ec34688b5b2bfab7d878662321ef2614d |
| command | sdlc close --issue 241 |
| reviewer | claude |
| timestamp | 2026-09-28T19:18:32-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

The close window does what the issue's Spec and Plan ask. It adds an MIT `LICENSE`, and every release archive now carries `weave` plus `LICENSE`, which unit and release tests both pin. The formula template now declares the license with the stanza order `brew audit` accepts. The #250 source-build fallback is gone from `merge-check.yml`, and `portable-ci.test.sh` now asserts that it is absent. The sweep of README and atlas text leaves only historical `#241` mentions. I re-ran `go test ./cmd/weave/internal/release/` and `scripts/test/portable-ci.test.sh`, and both pass. I did not re-run the heavy `release-weave.test.sh`; the Log records its PASS and the CI build-of-record run. Three of the four prior findings are fixed. The fourth (BR-4) is still open. It is Minor, and its deferral target #265 does not record it yet. Nothing blocks shipping.

1. **Strengths**
   - `cmd/weave/internal/release/main.go:78-83` reads `LICENSE` before `staging.New`, so a missing license fails before any staging. `TestMissingLicenseFailsBeforeStaging` pins this, including that no stages or output are left behind.
   - `writeArchive` writes its entries in a table-driven loop. `TestArchiveCarriesBinaryAndLicense` checks each entry's name, mode and content, and that no extra entry follows.
   - `release-weave.test.sh:60-61` checks that the archive's `LICENSE` is byte-identical to the repo's, and replaces the old single-member assertion.
   - The fallback removal is enforced both ways: `portable-ci.test.sh:45` asserts `'brew tap' not in workflow and 'ariadne-source' not in workflow`, and the published-tap event row is `install → --prefix → compile`.
   - README publishing steps 1–5 match the Log's evidence: build of record, `shasum -c`, `gh release create --verify-tag`, and verification with an empty trust store.

2. **Critical:** none.

3. **Important:** none.

4. **Minor**
   - BR-4 remains open (disposed below). The Log says the item "belongs to #265 and should be added there when it's claimed", but #265 on origin/main never mentions seed or merge-check retirement. The deferral exists only in this issue's prose.
   - `atlas/workflow/weave.md:40` is a reflowed line that exceeds the paragraph's wrap width. Cosmetic only.

5. **Test coverage:** The new behaviours each have a test that fails without them:
   - Two archive members: the unit test and the release test both fail if `LICENSE` is dropped.
   - License read before staging: the missing-LICENSE test fails if the read moves after `staging.New`.
   - Fallback removed: the portable-ci absence assertion fails if the `elif` branch returns.
   - `license "MIT"` in the formula: the release test's formula assertion fails without it.

6. **Architecture**
   - **ARCH-DRY: pass.** The archive entries come from one table, with no copied write blocks.
   - **ARCH-PURE: pass.** The change is small, and archive writing is tested directly without mocks.
   - **ARCH-PURPOSE: pass.** Every Done-when item has Log evidence: the release, the clean tap install, Linux CI via parley.nvim, and fleet migration. The docs sweep is complete.
   - **ARCH-MOCK: pass.** The `brew` and `go` fakes in portable-ci sit behind PATH, which is the same seam CI uses. The live conformance check was the parley.nvim CI run.
   - **ARCH-CONSTRAINTS: N/A.** This is release packaging with no runtime-envelope change.
   - **ARCH-SECURE: pass.** `LICENSE` is a repo-owned input and a missing file fails loudly. No credentials are touched.
   - **ARCH-ORDER: pass.** The only ordering that matters is the provenance invariant: tag, then no rebase, then merge keeping the tagged SHA reachable. The issue states it explicitly, and `gh release create --verify-tag` enforces it.
   - **ARCH-FUNERAL: flag (Minor, BR-4).** Consumers' seed-once copies of `merge-check.yml` still carry the dormant fallback, and nothing retires it.

7. **Plan revisions:** none needed; the plan matches the code. To close BR-4 cheaply, add a line to #265's Spec about retiring seed-once residue, rather than relying on the Log's "add it when claimed".

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      writeArchive now emits weave (0755) and LICENSE (0644); TestArchiveCarriesBinaryAndLicense and release-weave.test.sh:60-61 fail without it; Log confirms the keg has LICENSE.
  - id: BR-2
    disposition: addressed
    note: |
      README.md now ends with "## License / MIT — see LICENSE" and states that archives carry LICENSE (MIT).
  - id: BR-3
    disposition: addressed
    note: |
      The template now puts license after version (packaging/homebrew/Formula/weave.rb:5-6), so future releases publish verbatim; the one-time tap swap is logged.
  - id: BR-4
    disposition: not-addressed
    note: |
      Deferred to #265 in Log prose only; #265 on origin/main has no seed/merge-check retirement item. Record it in #265 so the deferral is durable.
```
