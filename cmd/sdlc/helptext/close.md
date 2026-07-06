Close an issue or a milestone — perform AGENTS.md §5's mechanical closing
steps. Edits files in place; does NOT commit (the agent commits, usually
bundling close with other work).

`sdlc close` is the LOCAL ACCEPTANCE GATE (#160): it runs the fresh-context
boundary review (all LLM review — code quality, requirements traceability, docs
sync incl. README, architecture) and flips a full issue to `codecomplete`, NOT
`done`. The deterministic publish gate (`sdlc merge`/`push`) later flips
`codecomplete → done` after verifying nothing drifted since close. close is the
SOLE writer of `codecomplete` (set-status refuses it), which makes the commit
carrying it a trustworthy anchor for that reviewed-HEAD-unchanged invariant.

MODES

  Issue close:      sdlc close --issue 15 --actual 7 --verified '<evidence>'
  Milestone close:  sdlc milestone-close --issue 15 --milestone M4 --actual 2.5 --verified '<evidence>'

  (Milestone closing lives on `sdlc milestone-close` — `close` no longer takes
   `--milestone` (#146). See `sdlc milestone-close --help`.)
  (The --actual values above are MEASURED — from `sdlc actual` or the omit-then-
   suggest path below — not typed estimates. See the --actual flag note.)

  THE BOUNDARY REVIEW (#69). A standalone full-issue close auto-dispatches the
  one binary-owned fresh-context review on the whole-issue window (the same
  reviewer `milestone-close` runs per-milestone). For a no-milestone issue this
  is the single review the boundary gets; for a multi-milestone issue it's the
  end-of-issue integration review on top of the per-milestone ones. The agent
  does NOT separately run `superpowers-requesting-code-review` (AGENTS.md §3).
  Skip with `--no-judge` (records a not-run trailer). Milestone slices are
  reviewed by `sdlc milestone-close` (per-milestone); to skip THAT review
  explicitly, use `sdlc milestone-close --no-judge` — `close` no longer has a
  `--milestone` path to skip it silently (#146).

WHAT THE GUARD DEFENDS

  --actual <hours>     focused dev-hours (sdlc computes it — close suggests a
                       number, or run `sdlc actual --issue N`). Required unless
                       measurement is explicitly not applicable.
  --verified '<line>'  one-line evidence the work meets done-when (behavior,
                       not artifacts: "tests pass" beats "code written").
                       Required.

  Plus structural checks:
    - atlas/ must have changed in the issue's commit window (§5 step 5)
    - issue's `## Plan` has no unchecked items (issue close only)
    - each milestone listed in ## Plan must carry a `Review-Verdict:`
      trailer on its close commit (issue close only; AGENTS.md §3).
      A Plan with only plain `- [ ]` checkboxes (no `Mx` rows) has no
      such requirement — atomic single-pass work closes in ONE `sdlc
      close`, no milestone-close, one `closed —` log line. Reserve `Mx`
      tags for ≥2 genuinely separate review boundaries (AGENTS.md §3).
    - milestone-close ticks the `- [ ] M4 — ...` row; refuses if absent
    - project file (if any, under <brain>/data/project/*.md referencing
      <repo>#<id>) gets its task row ticked + detail block updated

  BYPASSING A GATE (#67) — each gate has its own --no-<gate> flag, so you
  can waive exactly the one that doesn't apply (and acknowledge it) instead
  of reaching for the blanket --force:

    gate                          flag
    actual-hours required         --no-actual
    verified-evidence required    --no-verified
    already-done refusal          --no-reclose-guard
    atlas/ changed in window      --no-atlas
    milestone Review-Verdict      --no-verdict
    ## Plan has no unchecked       --no-plan-check
    project detail-block updated  --no-project
    issue boundary review (#69)   --no-judge

  Each bypass logs an audit "[!] --no-X: skipping ..." line (it's an
  explicit acknowledgment, not a silent skip) and the rationale belongs in
  --verified. --force ≡ all of them at once (emergencies). Prefer the
  precise flag: e.g. a pure bugfix with no new architectural surface closes
  with `--no-atlas` (reason in --verified), NOT --force.

WHAT IT DOES

  - Ticks the milestone box in the issue's ## Plan (milestone mode)
  - Flips status: codecomplete (#160 — NOT done; the deterministic publish gate
    `sdlc merge`/`push` flips codecomplete → done), sets actual_hours (number or
    N/A) and updated (issue mode)
  - Appends a log line to ## Log: "YYYY-MM-DD: closed — <verified>"
  - Emits the no-LLM `lessons` reminder on a whole-issue close (#160 Q4 — moved
    here from the publish gate so it fires while findings are fresh)
  - Ticks the project task row + upserts **actual:** and **closed:** in the
    detail block
  - Does NOT git-commit, does NOT move the file to workshop/history/

WARMUP

  On the first 2 invocations per shell session, prints the close-issue
  contract to stderr. After that, silent. Reset by starting a new shell.

FLAGS

  --issue <n>           ariadne workshop issue ID (numeric, zero-pad
                        applied internally; required)
  --actual <hours>      focused dev-hours (required unless --no-actual/--force)
  --verified '<line>'   one-line behavior evidence (required unless --no-verified/--force)
  --force               bypass ALL gates (≡ every --no-* flag); reason in --verified
  --no-actual           record actual_hours: N/A; skip velocity calibration
  --no-verified         bypass the VERIFIED-evidence requirement
  --no-reclose-guard    re-close an already-done issue (skip the refusal)
  --no-atlas            skip the atlas/ change check (no new architectural surface)
  --no-verdict          skip the milestone Review-Verdict trailer check
  --no-plan-check       close despite unchecked ## Plan items
  --no-project          skip the project detail-block update requirement
  --no-judge            skip the issue boundary review on full-issue close (#69)
  --agent <cli>         agent CLI for the boundary review (claude | codex | gemini)
                        Default: explicit --agent, then AGENT_CMD, then
                        PAIR_AGENT/current known agent signals, then claude.
  --dry-run             print what would change, write nothing
  --brain-dir <path>    project-file lookup root (default ../brain)
  --issues-dir <path>   issues directory (default workshop/issues)

DEEP-DIVE REFERENCES

  AGENTS.md §5                       closing checklist
  brain/data/life/42shots/velocity/  v3 attribution method
    baseline-v3.md
  construct/datatype/project.md      project-file shape & detail blocks

If --actual is missing, close runs active-time-v3 itself (brain + repo
transcript dirs, this issue's window + auto-discovered peers) and prints the
measured suggestion inline — `→ close with: --actual <h>` — or, when the
telemetry isn't available, points you to a labeled judgment estimate. The
suggestion may include attribution warnings; inspect those before recording the
actual, because they identify dominant long runs or mention fallback where commit
boundary evidence was weak. If
--verified is missing, the explainer shows a worked example of a
behavior-grounded VERIFIED string. Read the explainer; the contract is
load-bearing.

Use --no-actual only when focused dev-hours are genuinely not applicable or
cannot be measured without fabricating a number. On a full issue close, this
writes `actual_hours: N/A` to keep the issue schema-valid and records that the
close is excluded from velocity calibration.

If --actual IS passed, close still measures (active-time-v3) and sanity-checks
the value against it (#87): a moderate deviation (≥3×) warns; a wild one (≥10×,
e.g. a typed estimate where the measurement is minutes) is REFUSED — re-run
`sdlc actual` and pass the measured value, or `--force '<why>'` if the
measurement is genuinely wrong. Small absolute gaps (< 0.5h) never trip it.
This is the backstop for the "hand-type a plausible number" failure that a
guessed --actual would otherwise sail through. (milestone-close inherits it.)

CALIBRATION LEDGER (#117 — closing the estimate↔actual loop)

  On a whole-issue close (not a milestone) with a measured numeric --actual, close
  appends one estimate↔actual row to the calibration ledger (default
  brain/data/life/42shots/velocity/calibration-ledger.tsv; override with
  $WF_CALIB_LEDGER) and flags >2× same-direction drift over the last N unique
  window-trusted rows for the latest recognized model revision. Each row records
  whether the actual came from a
  `started:`-windowed measurement (#116) — pre-#116 rows are window-trusted=no and
  excluded from drift stats (a truncated actual isn't a clean data point). Pass
  --mode supervised|delegated to tag the supervision style. If no ledger dir
  exists (a downstream repo with no sibling brain/), close skips it with a warning
  — a missing ledger never breaks the close. `actual_hours: N/A` closes are also
  excluded from the ledger and drift stats.
