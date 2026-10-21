# Boundary Review — ariadne#526 (whole-issue close)

| field | value |
|-------|-------|
| issue | 526 — landing: drop one car from a composed train before merge (no per-car hold today) |
| repo | ariadne |
| issue file | workshop/issues/000526-landing-drop-one-car-from-a-composed-train-before-merge-no-per-car-hold-today.md |
| boundary | whole-issue close |
| milestone | — |
| window | 510aab6ff3a55bfdf5472e45eedf5512f6d04420..8fb25eb263f95e8c8169e4db3bc2560f4be1c1f3 |
| command | sdlc close --issue 526 |
| reviewer | codex (openai; author openai) |
| isolation | fenced |
| timestamp | 2026-10-21T13:56:57-07:00 |
| verdict | REWORK |

## Review

```verdict
verdict: REWORK
confidence: high
```

The normal three-car drop flow, cleanup barrier, ownership checks, and documentation are well covered. One reproduced ordering bug blocks shipping: an accepted drop can lose precedence to a concurrent membership update, preventing survivor re-formation.

```findings
findings:
  - id: new
    severity: Critical
    kind: defect
    family: drop-preempts-completion
    title: |
      Accepted drops must fence non-merge completion updates
    detail: |
      cmd/sdlc/landingtrain.go:176 checks pending drops before computing and publishing a completion, but UpdateTrain at cmd/sdlc/internal/landingline/traincontrol_io.go:178 commits membership changes without checking accepted drops under line.lock. A drop accepted between those operations can target a car removed by a bisect completion; the next ApplyTrainDrops then fails with “drop does not pin a unique riding car,” and the surviving conductor exits instead of automatically re-forming. A temporary Go-overlay regression reproduced this failure. Fence completion updates atomically against pending drops and route the pending result through applyTrainDrops using the pre-completion train; cover membership-changing and terminal completions. ARCH-ORDER, ARCH-PURPOSE.
```

1. **Strengths**

   - Drop acceptance and merge admission share `line.lock`; the durable merge-start marker survives reintegration.
   - Cleanup requires positive PR closure and leased-ref deletion evidence, including lost replies and retargeted remotes.
   - Real Git fixtures verify survivor ancestry, dropped-commit exclusion, and both dropped-conductor and dropped-middle cases.
   - README, help, and atlas explain authority, recovery, compatibility, and the merge boundary.

2. **Critical findings**

   - **Accepted drop versus completion race:** [traincontrol_io.go:178](/Volumes/TestScratch/review-checkouts/7df8449c136c8f51/cmd/sdlc/internal/landingline/traincontrol_io.go:178), called from [landingtrain.go:199](/Volumes/TestScratch/review-checkouts/7df8449c136c8f51/cmd/sdlc/landingtrain.go:199). The merge path handles `ErrTrainDropPending`, but ordinary completion updates lack the equivalent fence. Preserve the original train until the locked update confirms that no accepted drop takes precedence.

3. **Important findings**

   None.

4. **Minor findings**

   None.

5. **Test coverage notes**

   Passed:

   - Complete `landingline` package tests.
   - Targeted command drop and authorization tests, including real three-car re-formation.
   - Race-enabled control, transition, and publication tests.
   - Seven Python deadline/process controls.
   - `git diff --check`.

   The temporary overlay probe reproduced: bisect completion computed → drop accepted → membership update committed → drop application fails. Existing tests omit this ordering. Repository files remained unchanged; the full repository suite was not rerun.

6. **Architectural notes**

   - **ARCH-DRY — pass:** reuses locks, codec, abort effects, and normal admission.
   - **ARCH-PURE — pass:** decisions and transitions remain free of IO; integration seams are separate.
   - **ARCH-PURPOSE — flag:** the reproduced race interrupts promised automatic survivor re-formation.
   - **ARCH-MOCK — pass:** stateful GH doubles and portable real Git fixtures exercise production seams.
   - **ARCH-CONSTRAINTS — pass:** bounded records, scoped polling, and counted batch-work checks.
   - **ARCH-SECURE — pass:** caller/target separation, strict decoding, and destination revalidation.
   - **ARCH-ORDER — flag:** pending-drop precedence is enforced at merge admission but not every completion commit.
   - **ARCH-FUNERAL — pass:** cleanup obligations persist; records and exclusions have bounded replacement/retirement paths.
   - **ARCH-EVOLVE — pass:** strict known fields and nested unknown-field preservation are implemented and tested.

7. **Plan revision recommendations**

   Add a `## Revisions` entry requiring every pre-merge completion commit to serialize against accepted drops, including membership reductions and terminal refusals. Require deterministic tests for both lock orderings and preservation of the pre-completion train when drop wins.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"82b3bd5e-cfe2-4e36-9dcd-4aa6bb778cff","repository":"github.com/xianxu/ariadne","issue_id":"000526","recipe":"milestone-review","round":1,"attempt":1,"agent":"codex","root":"/Volumes/TestScratch/review-checkouts/7df8449c136c8f51","reviewed_head":"8fb25eb263f95e8c8169e4db3bc2560f4be1c1f3","slot":"ariadne:1","requested_at":"2026-10-21T20:54:36.203878Z"},"phase":"finished","pid":28366,"started_at":"2026-10-21T20:54:36.304738Z","finished_at":"2026-10-21T20:56:57.182586Z","session_id":"01a1603e-952a-7120-aafc-4ee8b5a583f5","outcome":"returned"}] -->
