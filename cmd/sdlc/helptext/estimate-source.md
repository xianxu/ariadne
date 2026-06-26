Name BOTH halves of estimate derivation in one place (#134) — so an agent never
satisfies the `## Estimate` block grammar while picking per-primitive hours from
memory. The estimate-side counterpart to `sdlc arch-principles`.

WHY THIS EXISTS

  Estimation is split correctly but discovery was implicit: the SHARED METHOD
  (the block grammar + closed vocabulary) is single-sourced in `sdlc`, while the
  REPO-LOCAL CALIBRATION (the actual per-primitive hour ranges, which drift as
  closes accrue) lives in a brain artifact. An agent could satisfy the syntax
  `sdlc change-code` enforces without ever reading the calibrated numbers. This
  command makes the split — and the exact source path — discoverable.

WHAT IT PRINTS

  - Shared method: `sdlc change-code --help` (grammar + vocabulary,
    helptext/estimate.md) + the recognized model versions.
  - Repo-local calibration: the resolved path to the model's calibration doc,
    tagged [ok | stale | MISSING], with a status-specific next-action.

SOURCE RESOLUTION

  $WF_ESTIMATOR_SRC (if set) wins; else <brain-dir>/data/life/42shots/velocity/
  <model>.md (e.g. estimate-logic-v2.md). `stale` means the sibling
  calibration-ledger.tsv is newer than the doc — closes have accrued since the
  last recalibration (tracked in #127), so treat the numbers as provisional.

FAIL-LOUD CONTRACT

  When the calibration source is MISSING/inaccessible, this PULL command exits
  non-zero with a next-action — it never silently lets you fall back to memory.
  (The start-plan / change-code PUSHes only warn-and-continue, so a downstream
  repo with no brain never breaks the gates; this command is the loud one.)

FLAGS

  --model <m>         estimate model whose calibration doc to resolve
                      (default estimate-logic-v2)
  --brain-dir <path>  brain repo path (holds the calibration doc + ledger;
                      default ../brain)

RELATED

  sdlc change-code    enforces the `## Estimate` block; its missing-block error
                      points here.
  sdlc start-plan     pushes a one-line estimate-source pointer at planning time.
