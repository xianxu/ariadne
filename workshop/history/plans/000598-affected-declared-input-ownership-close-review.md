# Boundary Review — ariadne#598 (whole-issue close)

| field | value |
|-------|-------|
| issue | 598 — Affected planning: declared non-Go path ownership and audited reader inputs |
| repo | ariadne |
| issue file | workshop/issues/000598-affected-declared-input-ownership.md |
| boundary | whole-issue close |
| milestone | — |
| window | 65726091f368c5713d1ec162e407be6dec8c79bf..a31433be3309b76b9f1c9e405eb595584ec744fa |
| command | sdlc close --issue 598 |
| reviewer | claude (anthropic; author anthropic) |
| isolation | fenced |
| timestamp | 2026-10-24T17:32:26-07:00 |
| verdict | SHIP |

## Review

Prior rounds' open findings are BR-6 and BR-7. Let me check what changed in the last two commits.

Checking the two open findings (BR-6 doc fix, BR-7 boarding test) against the code; now running the tests.

I finished verifying the two open findings and ran the tests (all green). Verdict: SHIP.

```verdict
verdict: SHIP
confidence: high
```

Both open findings from the earlier rounds are fixed and checked against the code, and I'm raising no new findings. For BR-6, the corrected README and atlas text now matches the four main callers of `readAffectedConfig`. Pre-close and a train car's stand-in fall back to requiring the suite. An explicit `--name affected` run and its resume return the parse error. For BR-7, the malformed-sidecar test now runs in both ordinary and boarding mode, checks the refusal and the cause it names, and passes. The scope commit moving M2 and M3 to ariadne#616 is legitimate: #616 exists on `main` (`b01ce4497`) with the moved criteria. The issue's original Done-when is M1's scope, which is delivered, and the Revisions entry records the change and the re-derived estimate. Tests in `internal/affected`, `internal/gitx` and the affected-related sdlc tests passed at the head.

1. **Strengths**
   - Declarations are read only from the pinned main (`localcheck_affected.go` `loadAffectedMapping`, called with `base`), so a branch can't declare or undeclare its own selection.
   - A cancelled read is reported as an interruption, not as a reason to select everything (`planAffectedSelection`). A test checks this in both modes.
   - `gitx.TreeFilesSelected` is extracted with guards (an empty `dir` or a `-`-prefixed revision is refused) and is shared by `TrunkView.FilesSelected`, so there is one implementation.
   - `splitImpact` loads each endpoint's own mapping and uses `affected.Owned` in its uncertain fallback, so attribution and selection make the same ownership claim.

2. **Critical:** none.
3. **Important:** none.
4. **Minor:** none new. Two `readAffectedConfig` callers aren't in the atlas's "Older binaries" list:
   - republishing a saved known-flake affected pass (`localcheck_cmd.go:377` via `declaredCheck`) refuses;
   - train split attribution (`landingattribution_impact.go:65`) becomes unverified.

   Both are conservative and neither can narrow a selection, so the doc's "None of them can narrow anything" still holds. I put this in the BR-6 note rather than raising a finding.

5. **Test coverage notes:** I ran `TestAffectedDeclaredMalformedSidecarRunsTheSuite` and it passes. The boarding branch checks the refusal text, the sidecar cause, and that neither the suite nor affected ran. I didn't revert-check it because the change is a test: it is the evidence BR-7 asked for, not a behavior fix.

6. **Architecture notes**
   - **ARCH-DRY:** pass. `fullMatch` and `claimPath` replace the inline loops, and `affectedIsDoc` replaces the two duplicate lambdas.
   - **ARCH-PURE:** pass. `Plan`, `Compile` and `Owned` are pure, and IO stays in `loadAffectedMapping` and the `mapping` field of `affectedDeps`.
   - **ARCH-PURPOSE:** pass. M1 is what pair#853 needs, and M2/M3 are separable additions that moved to #616 with the operator's approval.
   - **ARCH-MOCK:** pass. Tests use the same `affectedDeps` seam as production, against temporary real-git repos.
   - **ARCH-CONSTRAINTS:** pass. The benchmark of 2,000 paths against 1,200 rules takes about 30 ms, and each sidecar convention costs one tree listing and one batch read.
   - **ARCH-SECURE:** pass. Sidecars are untrusted input: they're parsed strictly, and a bad one selects the full suite and names the file.
   - **ARCH-ORDER:** pass. Nothing is held between events; each run is a single computation.
   - **ARCH-FUNERAL:** pass. This creates nothing durable; sidecars are files the operator writes.
   - **ARCH-EVOLVE:** pass. The new `affected.json` members are additive, the strict refusal by older binaries is documented with per-caller behavior, and those binaries never narrow a selection.

7. **Plan revision recommendations:** none. The plan's Revisions entry already records the move to #616.

```findings
dispose:
  - id: BR-6
    disposition: addressed
    note: |
      README.md:938 and the atlas Older binaries bullet now match precloseck.go:33 and landingcar.go:164 (suite) and localcheck_cmd.go:219 and localcheck_resume_cmd.go:89 (error). Known-flake republish (localcheck_cmd.go:377) and split attribution (landingattribution_impact.go:65) are unlisted but also never narrow, so the doc's claim holds.
  - id: BR-7
    disposition: addressed
    note: |
      TestAffectedDeclaredMalformedSidecarRunsTheSuite now loops over affectedOrdinary and affectedBoarding. Boarding checks the refusal, the sidecar cause and that nothing ran. It passes at the head.
```

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"b698f88e-c787-400b-be9a-63d669b84736","repository":"github.com/xianxu/ariadne","issue_id":"000598","recipe":"milestone-review","round":3,"attempt":1,"agent":"claude","root":"/Volumes/TestScratch/review-checkouts/452d896038bc3133","reviewed_head":"a31433be3309b76b9f1c9e405eb595584ec744fa","slot":"ariadne:5","requested_at":"2026-10-25T00:29:00.332278Z"},"phase":"finished","pid":38283,"started_at":"2026-10-25T00:29:00.439015Z","finished_at":"2026-10-25T00:32:24.553573Z","session_id":"4ae32bfd-d3b6-4a02-bafa-a3354e2898a7","outcome":"returned"}] -->
