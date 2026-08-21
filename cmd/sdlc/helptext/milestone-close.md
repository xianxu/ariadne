Close one milestone of an issue AND auto-dispatch the post-milestone
fresh-context code review (AGENTS.md §3). The canonical closing path
for milestone work — bundles the mechanical close + the mandatory
review into one invocation so neither half is skipped.

WHAT IT DOES

  1. Runs the mechanical milestone close:
     - ticks the `- [ ] Mx — ...` item in ## Plan
     - updates the project file's task row + detail block (if any)
     - appends a verification log entry
     - refuses without --actual / --verified (unless --force)
     - refuses if atlas/ wasn't touched in the window (unless --force);
       auto-satisfied when the window has no code surface (#177)

  2. Auto-dispatches `sdlc judge milestone-review`:
     - Diff window: the PREVIOUS review boundary..the reviewed commit — the prior
       milestone close's commit (the one carrying its Review-Verdict:
       trailer), or the branch start for the first milestone. Basing on
       the prior boundary (not the first `#<issue> <milestone>` commit)
       means inter-milestone `#<issue>`-but-not-`<milestone>` commits
       (side-quests, fixes) land in exactly one window instead of
       slipping the gap between two milestones (#58). Matches close's
       atlas check window exactly.
     - Builds the milestone-review prompt with issue ref + base/head
     - Invokes the configured agent (claude by default)
     - Surfaces findings + classifies clean / info / failure
     - Parses the first line for SHIP | FIX-THEN-SHIP | REWORK

  3. Emits a trailer block to stdout — paste verbatim into the close
     commit message so `sdlc close` (full-issue close) can later verify
     each milestone was reviewed:

         Review-Verdict: SHIP
         Review-Window: abc1234..def5678
         [Review-Reason: --no-judge]   (only when verdict is not-run)

     On FIX-THEN-SHIP it also prints the post-verdict protocol (#174):
     fix the findings before committing, bundle them into the one
     milestone-close commit, do NOT re-run milestone-close.

  4. Appends "; review verdict: <verdict>" to the just-written log line
     in the issue file so a human grep finds it.

If the close succeeds but the judge dispatch fails (agent CLI missing,
no commits matched, etc.), the verb does NOT fail the close — it logs
a warning, records verdict as `not-run` with a reason, and exits
successfully. The close is the durable mutation; the review is a
follow-on. The trailer block is still emitted so the audit chain stays
intact (operator can re-run the judge and amend the trailer later).

FLAGS

  --issue <n>           ariadne workshop issue ID (required, positive)
  --milestone <Mx>      milestone tag (required)
  --actual <hours>      focused dev-hours — MEASURED, not typed. Omit it and
                        close ADOPTS the measured value (active-time-v3; #178 —
                        the info line states the cumulative-window semantics),
                        or run `sdlc actual --issue N` to preview; don't hand-type.
                        A passed value is sanity-checked against the measurement
                        (#87, inherited from close): ≥3× warns, ≥10× refuses.
  --verified '<line>'   one-line behavior evidence
  --force               bypass close's guards (record reason in --verified)
  --dry-run             plan only; skip both close mutation and judge dispatch
  --no-judge            run the close but skip the auto-dispatched judge
  --agent <name>        agent CLI for the judge: claude | codex | gemini.
                        Default: explicit --agent, then AGENT_CMD, then
                        PAIR_AGENT/current known agent signals, then claude.
  --brain-dir <path>    brain root for the calibration ledger (default ../brain);
                        project files are discovered across the fleet (#171)
  --issues-dir <path>   directory holding issue files
  --plans-dir <path>    durable plans + gate sidecars (default workshop/plans);
                        where the cost report finds the plan-gate ledger (#187)

USAGE

  # (--actual 6 below is the MEASURED value from `sdlc actual` / the omit-suggestion,
  #  not a typed estimate)
  sdlc milestone-close --issue 31 --milestone M4 --actual 6 --verified '...'

  # Skip the review (already ran it manually, or this is a no-code milestone):
  sdlc milestone-close --issue 31 --milestone M4 --actual 0.5 \
    --verified 'docs-only milestone, no code to review' --force --no-judge

  # Preview without mutating or dispatching:
  sdlc milestone-close --issue 31 --milestone M4 --actual 4 --verified '...' --dry-run

RELATED

  sdlc close             whole-issue close (auto-dispatches the end-of-issue
                         boundary review; refuses --milestone — #146)
  sdlc judge milestone-review --base SHA --head HEAD
                         manual milestone-review invocation for ad-hoc windows

THE GATE LEDGER (#194)

  Milestone reviews share the issue's one boundary ledger
  (`workshop/plans/NNNNNN-slug-close-gate.md`), with each round stamped by its
  boundary. The round cap and the open-findings set scope PER BOUNDARY, so M2's
  rounds cannot push the whole-issue close past its cap; finding families stay
  visible across milestones, which is the point of one file rather than several.

  A milestone close refuses when the ledger still holds an open blocking finding,
  even on a SHIP verdict. `--no-ledger` waives that one refusal;
  WF_BOUNDARY_ROUND_CAP (default 3) bounds the rounds.

  See `sdlc close --help` for the full contract — it is the same gate.
