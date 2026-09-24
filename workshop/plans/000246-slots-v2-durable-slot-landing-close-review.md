# Boundary Review — ariadne#246 (whole-issue close)

| field | value |
|-------|-------|
| issue | 246 — Slots v2: land while retaining the workspace |
| repo | ariadne |
| issue file | workshop/issues/000246-slots-v2-durable-slot-landing.md |
| boundary | whole-issue close |
| milestone | — |
| window | 9977b163e4367cfbfa7ba249ccd6467578e13485..96918897e2fe20e59e758f5bf40997b4ca87b571 |
| command | sdlc close --issue 246 |
| reviewer | codex |
| timestamp | 2026-09-23T17:30:36-07:00 |
| verdict | REWORK |

## Review

```verdict
verdict: REWORK
confidence: high
```

The range delivers the durable landing architecture with strong GitHub evidence, archive proof, retry handling, and stateful integration fixtures. It is not ready to ship because cleanup incorrectly rejects a dirty resting branch after successfully switching to it, leaving the integrated workspace stranded and making retry impossible without manual cleanup.

1. Strengths

- Structured PR identity and full SHA validation in `cmd/sdlc/ghlanding.go`.
- Archive publication uses scoped PR-owned ancestry and conditional remote updates.
- Stateful bare-Git fixtures cover merge, squash, rebase, uncertainty, occupancy, and interruption paths.
- README, atlas, and embedded help were updated.
- Focused landing/archive tests passed: `go test ./cmd/sdlc -run 'TestLanding|TestArchive' -count=1 -timeout=15m`.
- `git diff --check` passed.

2. Critical findings

- `cmd/sdlc/landing.go:246`, `:274`, `:290` — `revalidateLanding` always calls `landingClean`. After switching from the issue branch to the unchanged resting branch, this rejects any preserved tracked dirt on the resting branch. The switch may already have succeeded, so integration is complete but cleanup returns an error; subsequent `sdlc merge --branch ...` fails at the initial cleanliness check as well. Separate pre-integration cleanliness from post-switch identity/occupancy validation, and add a regression fixture with non-colliding tracked dirt on the resting branch. `ARCH-ORDER`, `ARCH-PURPOSE`.

3. Important findings

None.

4. Minor findings

- `workshop/issues/000246-slots-v2-durable-slot-landing.md:57` still cites undefined `ARCH-STATE`; the plan corrected this to `ARCH-ORDER`, but the issue specification remains inconsistent.

5. Test coverage notes

The focused suite passed, but it does not test a dirty current resting branch. Existing dirty-neighbor coverage does not exercise the failing post-switch state.

6. Architectural notes

- ARCH-DRY: pass; shared archive and lifecycle helpers are used.
- ARCH-PURE: pass; phase decisions and archive planning are separated from IO.
- ARCH-PURPOSE: flag; dirty resting slots are not fully preserved through cleanup.
- ARCH-MOCK: pass; stateful GitHub/bare-Git fakes exist.
- ARCH-CONSTRAINTS: pass; bounded PR, commit, tree, and subprocess work is explicit.
- ARCH-SECURE: pass; repository identity, SHAs, paths, and external evidence are validated.
- ARCH-ORDER: flag; post-switch validation applies pre-switch cleanliness rules to a new state.
- ARCH-FUNERAL: pass; completed issue refs and branch configuration have explicit removal paths.

7. Plan revision recommendations

- Add a `## Revisions` entry documenting the distinction between pre-integration tracked-dirt refusal and post-switch preservation of non-colliding resting-branch dirt.
- Add the corresponding regression test to the Task 3 acceptance evidence.

```findings
findings:
  - id: new
    severity: Critical
    family: resting-checkout-dirt-preservation
    title: |
      Post-switch cleanup rejects preserved tracked dirt on the resting branch
    detail: |
      cmd/sdlc/landing.go:246, :274, and :290 route post-switch validation through landingClean, which rejects tracked dirt. A non-colliding dirty resting branch can therefore be switched to successfully but cannot complete deletion or retry, stranding the already-integrated workspace. Separate pre-integration cleanliness from post-switch identity validation and add a regression test. ARCH-ORDER, ARCH-PURPOSE.
  - id: new
    severity: Minor
    family: architecture-marker-integrity
    title: |
      Issue specification retains undefined ARCH-STATE marker
    detail: |
      workshop/issues/000246-slots-v2-durable-slot-landing.md:57 cites ARCH-STATE, while the registry and corrected plan use ARCH-ORDER. Update the issue specification for consistent architectural traceability.
```
