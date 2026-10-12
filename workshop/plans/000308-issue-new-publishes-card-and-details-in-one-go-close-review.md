# Boundary Review — ariadne#308 (whole-issue close)

| field | value |
|-------|-------|
| issue | 308 — issue new publishes card and details in one go |
| repo | ariadne |
| issue file | workshop/issues/000308-issue-new-publishes-card-and-details-in-one-go.md |
| boundary | whole-issue close |
| milestone | — |
| window | ab362ac4c35f137125f25e36a45f7b994d7493ad..f4ebe569cfd632f11100b6b50d85724e3d9115c1 |
| command | sdlc close --issue 308 |
| reviewer | claude |
| isolation | checkout |
| timestamp | 2026-10-12T10:23:44-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

Prior finding BR-6 is confirmed addressed. Now checking the M2 peer-publish code against the old peerwrite behavior (brain repos, tests).

Found one likely regression (brain legacy projects no longer ticked). Running the M2 and tracker tests to confirm the suite is green.

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

Issue #308 does what it set out to do. `issue new` now files the card and details on origin through the `creation.main` stage and writes nothing into the checkout. A peer repository's project tick goes to that peer's origin through `TrunkFile.Update`, built fresh from the fetched bytes each time. The M1 review findings were fixed, and BR-6 is fixed too. The targeted tests pass, both in `cmd/sdlc` (PeerProject, NoProject, Recovery, IssueNew, workspace and environment close, MoveDetail) and in `internal/tracker`, `internal/issue` and `internal/processmanual`. One behaviour change was not declared, and it should be settled before shipping. Close used to write the tick into a brain repository's old `data/project` project files and leave nous to commit it. It now skips those files, because they count as peer projects, and then refuses them as a brain. The project is never ticked, and the warning tells the operator that "the TL or owner applies the tick there". The fix is cheap: keep the local write for brain projects, or declare that the behaviour is retired.

1. **Strengths**
   - `cmd/sdlc/internal/tracker/detailpublish.go:14`: `DetailPublication` is one shared builder for the main-branch commit. Creation and transfer refuse an existing destination in exactly the same way (ARCH-DRY).
   - `cmd/sdlc/projecttick.go:27`: `projectCloseEdit` is a pure function. Close uses it for the local preview, and the peer publish runs it again on origin's bytes, so the two cannot drift apart. The detail-block refusal is now a typed error rather than a `die` buried inside the loop.
   - `cmd/sdlc/peerproject.go:63`: the transform re-checks `tl:` owners on the fetched bytes and refuses if it cannot read them. `peerProjectPath` follows symlinks and refuses a path that escapes the repository.
   - `cmd/sdlc/internal/tracker/store.go:131`: `NeedsSourceCheckout` alone now decides which checkout must finish a receipt, and recovery adds no exception of its own for a particular operation.
   - Tests run against real git repositories with bare origins. Several tests that used to expect a local write in the environment slot now check the peer's origin and confirm the checkout was left alone.

2. **Critical findings:** none.

3. **Important findings**
   - **Brain projects in the old `data/project` folder are no longer ticked at close** (`cmd/sdlc/close.go:851`, `cmd/sdlc/close.go:926-928`, `cmd/sdlc/peerproject.go:81`).
     - `walkProjectRoots` (`internal/project/discover.go:166`) still finds a brain repository's `data/project/*.md` files, and `ActiveOnly` keeps the active ones.
     - The brain's root is not `repoTop`, so `computeClose` marks the edit `peer: true`. `applyClose` then skips the local write, and `publishPeerProject` refuses because `IsBrainRepo` is true.
     - Before this change, the base `applyClose` wrote every edit locally. `planPeerWrites` only stopped the auto-commit for a brain, with the next action "leave it — nous sweeps it".
     - The old atlas paragraph saying so was deleted without a replacement, and the close `Long` text still says close "updates the matching brain-side project file".
     - The test `TestClose_PeerProjectTickRefusesBrainAndNoOrigin` puts its fixture in `workshop/projects`, which a brain is never scanned for, so this path has no test.
     - **Fix sketch:** pick one of two options.
       - Keep the old behaviour: treat a brain match as a local write, not a peer (`peer := … && !gitx.IsBrainRepo(m.RepoDir)`), because writing into the brain checkout is how brain is meant to be updated.
       - Or declare the change: record the retirement in the plan's Revisions, the atlas page and the close help, and make the warning say why a brain is skipped.
     - Either way, add a test that puts the project under a brain's `data/project`.
     - The rule behind this: when you replace a path, list every case the old path handled and say what happens to each one.

4. **Minor findings**
   - `close.go:851`: whether a peer publishes at all is decided from the **local** copy (`newPT != pt`). A stale peer checkout can cause two problems:
     - The close can die on a missing detail block that origin already has.
     - It can skip a publish when the local copy already shows the row ticked but origin does not.

     Consider running the peer edit only against origin, and treating the local copy as a preview only.
   - `peerproject.go:56`: the trunk branch is hard-coded to `"main"`, while the code picks the publication remote using the resting branch. It works for today's fleet, but a comment would help.

5. **Test coverage notes**
   - M2's main cases have tests: the tick lands on origin with the checkout untouched, a symlinked project, an origin that moved, `tl:` gained on origin, no origin, `--no-project`, and the review snapshot leaving peers out.
   - The brain legacy-folder case has no test.
   - BR-6 was a dead clause, so removing it changes nothing at runtime, and no failing-first test is needed.
   - The targeted suites listed above all pass.

6. **Architectural notes**
   - **ARCH-DRY: pass.** `DetailPublication`, `projectCloseEdit` and `confirmedMain` are each shared.
   - **ARCH-PURE: pass.** The edit and the decision about who claims are pure. The IO is limited to `publishPeerProject` and `CreationOp`.
   - **ARCH-PURPOSE: flag, at the finding above.** The old peer-write logic was replaced without accounting for one of its cases (brain).
   - **ARCH-MOCK: pass.** Tests use real git and bare origins through the production `TrunkFile` interface.
   - **ARCH-CONSTRAINTS: pass.** Retries are bounded by `maxUpdateAttempts` and `MaxPublicationAttempts`, and a peer failure only warns.
   - **ARCH-SECURE: pass.** There is a symlink-escape check, the `tl:` read refuses on error, and a brain is never written.
   - **ARCH-ORDER: pass.** The two creation stages go through the receipt stepper. Old receipts are recognised by their shape and released, and the publish is idempotent when retried.
   - **ARCH-FUNERAL: pass.** Nothing new is stored durably. The leftover uncommitted peer edits of the old approach are gone.

7. **Plan revision recommendations**
   - Add a Revisions entry that records the brain decision (keep the local write, or retire it), because the plan's `publishPeerProject` section says "refuse a brain" without noting that this ends the brain-side tick.

```findings
dispose:
  - id: BR-6
    disposition: addressed
    note: |
      c929e244 removes the creation clause at issuerecovery.go:131; no other creation-specific checkout exception remains (store.go NeedsSourceCheckout has no creation case). Dead-clause removal, no behaviour test needed.
findings:
  - id: new
    severity: Important
    family: replaced-path-case-dropped
    title: |
      Close no longer ticks a brain repo's legacy data/project project: it is classed as a peer, skipped locally, then refused as a brain
    detail: |
      close.go:851 marks every match outside repoTop as a peer, applyClose skips its local write (close.go:926) and publishPeerProject refuses IsBrainRepo (peerproject.go:81). The base wrote the brain file and left it for nous to sweep. The atlas line saying so was deleted, the close Long text still promises a brain-side update, and no test covers data/project. Either exclude brains from peer (keep the local write) or record the retirement in Revisions, atlas and help, and add a data/project fixture test.
  - id: new
    severity: Minor
    family: replaced-path-case-dropped
    title: |
      Peer publish is gated on the stale local copy (newPT != pt) rather than on origin's bytes
    detail: |
      A stale peer checkout can die on a missing detail block that origin has, or skip a publish that origin still needs.
```

---

## Re-review — 2026-10-12T10:38:34-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 308 — issue new publishes card and details in one go |
| repo | ariadne |
| issue file | workshop/issues/000308-issue-new-publishes-card-and-details-in-one-go.md |
| boundary | whole-issue close |
| milestone | — |
| window | ab362ac4c35f137125f25e36a45f7b994d7493ad..ab8a8dfbbf6287ac56354d19ace0ae65958d29d6 |
| command | sdlc close --issue 308 |
| reviewer | claude |
| isolation | checkout |
| timestamp | 2026-10-12T10:38:34-07:00 |
| verdict | SHIP |

## Review

Focus: BR-7 fix (last commit) and BR-8.

Checking that the BR-7 fix holds up: I'll run its test, then revert the fix to see whether the test goes red.

BR-7's test goes red when the fix is reverted, so that one is settled. Next I'm checking BR-8 the same way.

The broad run timed out, so I'm rerunning just the peer-project tests against the reverted gate.

No test fails with BR-8's gate change reverted, so BR-8 has no regression test. Last step: checking how the plan records these fixes.

```verdict
verdict: SHIP
confidence: medium
```

This round was meant to close BR-7, and it does. Close now treats a match as a peer only if it is outside the current repo **and** not in the legacy brain home (`cmd/sdlc/close.go:856`, `&& !m.Legacy`). So a brain's `data/project` file gets its tick written in place, uncommitted, for nous to commit, the same as before #308. `applyClose` prints an info line saying so (`close.go:938`), and the atlas (`atlas/workflow/sdlc-binary.md:530`) and `helptext/close.md:266` now describe this exception. I checked that the test actually guards the fix: I removed `&& !m.Legacy` and `TestClose_BrainLegacyProjectIsWrittenInPlace` failed with the old "NOT updated on its origin: a brain capture repository is never written" warning; restoring it made the test pass. BR-8 is only half done. Peers now always publish (`|| peer`), but no test fails when that change is reverted. And a stale local peer copy can still kill the close before publishing, because `projectCloseEdit`'s missing-detail-block error goes to `die`. BR-8 is Minor, so it does not block. One caveat: I ran only the targeted tests. A broader `-run 'Peer|Close|Workspace|Fleet'` run hit the 600s timeout and produced no result.

1. **Strengths**
   - BR-7 is fixed where peer classification happens (`close.go:856`), not with a workaround downstream. The review snapshot (`close.go:1593`) and the evidence pathspec (`closetracker.go:345`) behave as they did before #308 for brain-legacy files: the base also put every project edit in the snapshot.
   - The new fixture test is realistic: a real brain marker, a bare origin, and checks on the in-place tick, the brain's HEAD, the origin's content and the absence of the peer warning (`peerproject_test.go:361-399`).
   - `publishPeerProject` keeps its brain refusal as defence in depth (`peerproject.go:81`), and `TestPublishPeerProjectRefusesABrain` still covers it directly.

2. **Critical findings:** none.

3. **Important findings:** none new.

4. **Minor findings**
   - BR-8 is still open, see the dispose block. The publish-anyway change has no test. A stale local peer copy that lacks the detail block still dies at `close.go:848` (the `eerr` → `die` path) even when origin has the block.
   - The new `cinfo` at `close.go:938` assumes that any non-peer edit outside `repoTop` is a brain-legacy file. That is true today, because `peer` is defined as the complement. If a third category is ever added, this line should test `m.Legacy` explicitly (store it on `projectEdit`).

5. **Test coverage notes**
   - BR-7: the regression test exists, passes, and goes red without the fix (verified by reverting it).
   - BR-8: with `|| peer` reverted, every test in `TestClose_PeerProject*`, `TestNoProject*`, `TestCloseReviewSnapshot*`, `TestClose_CurrentRepo*`, `TestPublishPeer*` and `TestClose_BrainLegacy*` still passes. A fixture where the local peer copy is already ticked but origin is not would pin it.

6. **Architectural notes**
   - ARCH-DRY: pass. The peer transform reuses `projectCloseEdit`.
   - ARCH-PURE: pass. The tick logic is pure (`projecttick.go`); git IO stays in `peerproject.go`.
   - ARCH-PURPOSE: pass. The BR-7 class ("a match outside repoTop is not always a peer") has one site, the classification line, and it is fixed there.
   - ARCH-MOCK: pass. Real bare-origin git fixtures.
   - ARCH-CONSTRAINTS: pass. Always publishing adds one fetch per peer project per close: bounded, and only at close time.
   - ARCH-SECURE: pass. Symlinks are resolved and escapes refused (`peerproject.go:30-43`), and a brain is never published into.
   - ARCH-ORDER: flag, Minor (it is BR-8). The publish decision now rests on origin's bytes, but the die-or-continue decision is still made on the local copy before origin is read.
   - ARCH-FUNERAL: pass. The only artifact the fix adds is the in-place brain tick, an edit to an existing tracked file that nous commits.

7. **Plan revision recommendations**
   - Add a `## Revisions` entry: "BR-7: a brain's legacy `data/project` match (`m.Legacy`) is not a peer. Its tick is written in place, uncommitted, for nous, as before #308. The 'refuse a brain' rule in the peer-publish design (plan line 68) applies only to non-legacy peer matches. Peers always publish, and whether anything changes is decided on origin's bytes."

```findings
dispose:
  - id: BR-7
    disposition: addressed
    note: |
      close.go:856 excludes m.Legacy from peer; the in-place write is kept. TestClose_BrainLegacyProjectIsWrittenInPlace goes red when the fix is reverted (verified). Atlas and close help updated.
  - id: BR-8
    disposition: not-addressed
    note: |
      The || peer gate is in, but no test fails when it is reverted (verified). A stale local copy missing the detail block still dies at close.go:848 before origin is read.
```
