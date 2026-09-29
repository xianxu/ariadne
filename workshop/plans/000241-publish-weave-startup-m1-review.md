# Boundary Review — ariadne#241 (milestone M1)

| field | value |
|-------|-------|
| issue | 241 — Publish weave and cut over startup |
| repo | ariadne |
| issue file | workshop/issues/000241-publish-weave-startup.md |
| boundary | milestone M1 |
| milestone | M1 |
| window | fa5396e6bc7437ad408aaf594bce35be867f21fc..86fc857dca316a0d4b9b591390dae3f4cb9eacb3 |
| command | sdlc milestone-close --issue 241 --milestone M1 |
| reviewer | claude |
| timestamp | 2026-09-28T19:04:38-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

M1 does what the Plan asks for. It adds a root MIT `LICENSE` (Xian Xu, 2026), puts `license "MIT"` in the formula template, and changes `release-weave.test.sh` so the license line is now required where it used to be forbidden. The test's Ruby stand-in for the Formula DSL needed a `license` method to match, and that was added. I ran `TMPDIR=… bash scripts/test/release-weave.test.sh <out> weave-v0.1.0` myself and it ends in `PASS release: four real archives, targets/CGO/version/layout/checksums, formula composition, failures`, so the Log's evidence is confirmed. One thing about scope: the pinned window (`fa5396e6..86fc857d`) also holds a large amount of already-merged main work from other issues (#253, #255–#263 and more). This boundary's own changes are only in commit `86fc857d` plus the plan commits, so I reviewed those against the issue. Nothing blocks M1. The tagged commit is sound for M2's build of record.

1. **Strengths**
   - `scripts/test/release-weave.test.sh:54`: the assertion now runs against the *generated* release `weave.rb` rather than the template. That proves the release tooling carries the license through to the artifact. It is not just a check on the source file.
   - `scripts/test/release-weave.test.sh:88`: adding `license` to the stand-in DSL keeps the formula composition test honest. The Log records the red run (`undefined method 'license'`), which shows the test actually executes the formula.
   - `LICENSE` is the verbatim standard MIT text. The operator's choice is recorded in the Plan ("Operator decisions (2026-09-28)"), which meets the Spec's rule to "obtain the operator's license choice… no license is inferred".
   - The regression test genuinely depends on the fix: removing `license "MIT"` from the template makes the line-54 assertion fail.

2. **Critical:** none.

3. **Important:** none.

4. **Minor**
   - The release archives ship only `weave`, and the test asserts `len(entries)==1` (`scripts/test/release-weave.test.sh:60`). So the MIT notice is not bundled with the binary copies. That is legally fine for the author's own distribution. Still, it is common practice to ship `LICENSE` inside the tarball (and optionally `doc.install "LICENSE"` in the formula) so redistributed copies carry the notice.
   - `README.md` does not mention the license. A one-line "License: MIT" belongs in the README update M2 already plans.
   - The window base `fa5396e6` is stale relative to this milestone and pulls in about 240 files of unrelated merged work. That is worth a look in how the boundary base gets computed after merges from main, but it doesn't affect this review.

5. **Test coverage**
   - The template → generated formula → Ruby composition path is covered end to end by the release test.
   - Nothing checks that `LICENSE` exists or is MIT. That's acceptable, since it's a static legal file and not behavior.

6. **Architecture (checked one by one)**
   - **ARCH-DRY: pass.** The license string lives in one template, and the test asserts on the generated output.
   - **ARCH-PURE: N/A.** The change adds no logic.
   - **ARCH-PURPOSE: pass.** M1's purpose is to license the release, and the generated formula carries the license. Shipping the notice inside the archives is noted above as Minor, not treated as the core purpose.
   - **ARCH-MOCK: pass.** The Ruby DSL stand-in is an existing test seam and was extended the same way as the rest of it.
   - **ARCH-CONSTRAINTS: N/A.**
   - **ARCH-SECURE: N/A.** No untrusted input or credentials are involved. M2's tap-trust work is where this principle will matter.
   - **ARCH-ORDER: N/A for M1.** This milestone holds no state between events. M2 has a real ordering constraint: tag, then build of record, then release, then tap, then clean-store verification, then fallback removal. The Plan already says the fallback comes out only after verification passes, and that the tagged SHA must never be rebased. Keep to both.
   - **ARCH-FUNERAL: N/A for M1.** It creates nothing durable beyond source files. In M2, the #250 fallback's removal is correctly scheduled.

7. **Plan revisions:** none. The plan matches the code.

```findings
findings:
  - id: new
    severity: Minor
    family: license-notice-travels-with-artifact
    title: |
      Release archives contain only the weave binary; the MIT notice is not bundled with distributed copies
    detail: |
      release-weave.test.sh:60 asserts a single tar entry. Consider shipping LICENSE in each archive (and doc.install in the formula) so redistributed binaries carry the notice.
  - id: new
    severity: Minor
    family: readme-reflects-public-surface
    title: |
      README does not state the project license
    detail: |
      Fold a one-line License: MIT into the M2 README sweep alongside the install instructions.
```
