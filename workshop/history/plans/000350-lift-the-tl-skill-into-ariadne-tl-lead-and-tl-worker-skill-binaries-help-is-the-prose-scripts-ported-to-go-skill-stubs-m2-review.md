# Boundary Review — ariadne#350 (milestone M2)

| field | value |
|-------|-------|
| issue | 350 — lift the TL skill into ariadne: tl-lead and tl-worker skill binaries (help is the prose), scripts ported to Go, skill stubs |
| repo | ariadne |
| issue file | workshop/issues/000350-lift-the-tl-skill-into-ariadne-tl-lead-and-tl-worker-skill-binaries-help-is-the-prose-scripts-ported-to-go-skill-stubs.md |
| boundary | milestone M2 |
| milestone | M2 |
| window | 1d88a7dea4628a527ed136e9929745ea88c212d1..9cb63797c7e4567798c31622edb61fe68a432110 |
| command | sdlc milestone-close --issue 350 --milestone M2 |
| reviewer | claude |
| isolation | fenced |
| timestamp | 2026-10-14T08:53:42-07:00 |
| verdict | SHIP |

## Review

Targeted tests pass. Last check: which test file holds the session-sync fake, since the plan names `fake_test.go`.

```verdict
verdict: SHIP
confidence: high
```

Only one code line changed in the pinned window `1d88a7de..9cb63797`: the doc comment of `TestSessionSyncConformance` (`cmd/tl-lead/conformance_test.go:16-17`). This answers M1's Minor finding BR-6. The rest of the window is the M1 gate ledger, the M1 re-review record and a one-line `## Log` entry. The base is a rebased copy of the branch, so the M2 code itself (the commits `6039a71a` and `3da466a3`) is identical on both ends. That code was covered by the M1 review window `1c391825..392c1079`. I still checked M2's plan items against the head tree. All M2 rows are delivered:
- `cmd/tl-lead/{main.go,help.md,main_test.go}`
- the stub `construct/local/tl-lead/SKILL.md`
- `tl-lead-build` in the Makefile's `tools`
- `tl-lead` in `atlas/workflow/tl-skills.md` (lines 11, 28, 30-33)

`go vet ./cmd/tl-lead/` is clean, and `go test` passes for `./cmd/tl-lead/...`, `./cmd/tl-worker/...`, `./pkg/tlrules/...`, `./pkg/markdoc/...` and `./pkg/trains/...`. Nothing blocks SHIP.

1. **Strengths**
   - The rewritten comment now matches the test exactly. Lines 53-68 fold `00000000-not-owed` and then check that section A is unchanged, which is what "a refusal, which changes nothing" claims.
   - The comment's summary line (line 14) already lists "a fold of an ID not owed refuses", so the header and the closing sentence no longer contradict each other.
   - The plan's 2026-10-14 Revisions entry matches the files: `fake_test.go` holds `TestFakeSessionSyncRefusals`, `sessionsync_fake_test.go` holds the `fakeSessionSync` model, and `fakeSDLC` is in `trains_test.go`.

2. **Critical findings:** none.

3. **Important findings:** none.

4. **Minor findings**
   - `conformance_test.go:15` runs long ("…exit 3. Opt-in (TL_LEAD_LIVE=1, couch on PATH): run it") compared with the wrapping around it. This is cosmetic, it predates the window, and I'm not raising it.

5. **Test coverage notes**
   - BR-6 changed only a comment, so no regression test applies. I checked the corrected text against the code it describes (lines 53-68).
   - The live conformance test still skips unless `TL_LEAD_LIVE=1` is set. I didn't re-run it live here because this window doesn't change behavior.

6. **Architectural notes** (none of them is touched by this window's one-line change; each was checked against the head tree)
   - **ARCH-DRY:** pass. The scanners live in `pkg/markdoc` and `pkg/trains`, and the old packages forward to them.
   - **ARCH-PURE:** pass. Rendering is pure, and IO goes through the `SessionSync` and `SDLC` seams.
   - **ARCH-PURPOSE:** pass. All of M2's prose, stub, Makefile and atlas rows are delivered.
   - **ARCH-MOCK:** pass. `fakeSessionSync` keeps state across calls, and the live conformance test now covers the fold refusals the fake models.
   - **ARCH-CONSTRAINTS:** pass. Car reads run at most 8 at once, and the slow fleet read overlaps them.
   - **ARCH-SECURE:** pass. couch is called with argv arrays.
   - **ARCH-ORDER:** pass. Writes are guarded by the section hash (`ErrStale`), and a delivery gap is a typed `ErrDeliveryGap`.
   - **ARCH-FUNERAL:** pass. The window creates no new artifacts, and the conformance test works in `t.TempDir()`.
   - **Upcoming:** `atlas/workflow/tl-skills.md:33` says how a failed car read renders (`?`). It could also mention the degraded off-train line added for BR-5 when M3's atlas pass happens.

7. **Plan revision recommendations:** none. The plan matches the code. The M2 item "Update the atlas, run the targeted tests, then `sdlc milestone-close --milestone M2`" is the step this gate completes.

```findings
findings:
  - id: new
    severity: Minor
    kind: defect
    family: atlas-tracks-behavior
    title: |
      tl-skills atlas does not mention the degraded off-train render when the fleet read fails
    detail: |
      atlas/workflow/tl-skills.md:33 says a failed car read renders as ?, but not that a failed sdlc fleet inventory now renders "? off-train slots unread: <reason>" (TestGatherTrainsFleetFailureDegrades); add one clause at M3's atlas pass.
```
