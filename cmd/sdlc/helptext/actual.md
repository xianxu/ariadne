Compute an issue's focused dev-hours via the in-binary active-time-v3 engine over
the correct transcript directories — so `--actual` for a close is *measured*, not
hand-typed (#68).

WHY THIS EXISTS

  The v3 procedure was prose: "run this 6-line python command over the issue's
  commit window." Nobody ran it — the easy `--git-repo . --issue N` form omits
  `--dir`, and v3's events come ONLY from transcript `.jsonl` files, so it
  returned a silent 0 that got passed off as a guess. This verb runs the right
  invocation for you.

WHAT IT DOES

  1. Pins the code and tracker histories. In tracker repositories the window
     comprises ownership intervals: claim to release, with every reclaim opening
     a fresh interval. Relocation retains the original session root within that
     continuous engagement. The output prints the actual right edge.
  2. Admits sessions from those owning worktrees, irrespective of the checkout
     running this command. A spawned session needs an admitted parent and a
     dispatch within the engagement. Unbound brain sessions are excluded; the
     provenance reports their count and active minutes. Missing or stale required
     evidence refuses automatic adoption. Repositories without a tracker retain
     the explicitly labeled repository-plus-brain heuristic.
  3. Runs the native `activetime` engine (`--commit-weight 1.0 --threshold-min 15
     --include-assistant`, in-process — no python3) and prints measured
     hours. If attribution is suspect (for example one long run
     dominates the issue total, or the engine had to use mention fallback because
     no issue commit boundary existed), warning lines are printed before the
     close suggestion.
  4. Splits the measurement at the issue's boundary reviews (#234), read from
     its boundary gate ledger: **build** runs to each boundary's first review
     dispatch; **converge** runs from there to the boundary's last #N commit at
     or before its final dispatch. The verification after that commit (the
     final test run and review) is the uncounted **tail**, at every boundary.
     The actual is build + converge. With `--milestone Mx` it is Mx's phases
     only, the increment `milestone-close` adopts.
     Fresh cuts use the exact code pin's executable issue Plan. An absent,
     never-finalized named boundary contributes no cut; the replacement owns
     work since the last finalized boundary (or issue start). Finalized history,
     still-executable open boundaries and whole-issue rounds remain. This
     accounting fallback does not mark a review SHIP or clear findings/caps.
     A requested unfinalized milestone must be executable in that pinned Plan,
     even with no open review rounds. Missing or ambiguous required Plan evidence
     refuses with recovery guidance.

OUTPUT

  - measured            → "measured actual for #N: <h>h", then the build /
                            converge / uncounted-tail line and which evidence cut it
                            (dispatch stamps, finalize times, or no-review). The effective
                            selected phase interval and code pin are shown;
                            provenance distinguishes admitted sessions from
                            actual contributing sources. Empty selected phases
                            and retained executable open boundaries stay visible.
  - telemetry unavailable → window has commits but no transcript events (work
                            lived in cwds/worktrees not under brain/repo, or aged
                            out / in an unsupported harness transcript shape).
                            Ask the TL/hub to restore the owning slot's transcripts,
                            then re-run sdlc actual --issue N.
  - legacy suite attempt → exact queue endpoints were never recorded. The
                            refusal names every legacy attempt and the matching
                            close or milestone-close --no-actual command only
                            when all blockers are legacy; include the reason
                            with verification evidence.
  - malformed suite data → restore the named recoverable evidence and re-run.
  - unverified coverage → "N min in a retained train wait" describes overlapping
                            whole waiter invocations with verified car membership.
                            It makes no new subtraction: substantive work and
                            polls keep their existing accounting. Pre-boarding
                            and released gaps are included; this is not proven
                            idle time or time aboard a train.
  - no window           → commit the issue's work with #N in its subject, then
                            re-run sdlc actual --issue N. Do not invent hours.

A migrated working issue with no tracker owner refuses with
`no owner since cutover: run sdlc claim --issue N from the slot doing the work`.
Reconcile does not create ownership. Work between cutover and claim is unowned
and not counted; successful hybrid measurements and retained replay expose that
gap beside the legacy prefix and owned suffix. Claim does not retroactively
credit it. Existing activity continuity may cross claim, but only its portion
inside ownership can count. For a suite queued before first ownership,
hybrid measurement clips queue-idle exclusion to the owned portion, preserving
its raw endpoints. Waits crossing release/reclaim still refuse.

Completed issues replay the measurement inputs retained with their close evidence:
code/tracker pins, source byte prefixes and digests, ownership windows, and phase
cuts. Later transcript appends and bookkeeping cannot move that measurement.
Saved cuts are replayed exactly: retiring a Plan row later does not rewrite an
already-adopted sample. The fresh-cut fallback is not a historical repair.
Legacy closes without retained inputs reconstruct their historical completion
binding and report that limitation.

Retained train waits (#547) record ticket-owned start, verified car/conductor
binding and owner-return end facts in the existing 30-day landing audit. Fresh
measurements select the issue's immutable head, owner engagement and pinned source
windows, union descriptive overlap, and retain optional coverage for replay.
Incomplete/expired facts or diagnostic capacity limits report unavailable coverage
without changing measured hours or calibration validity. Optional coverage is
bounded to 16 intervals and 8 KiB within the existing snapshot limit; if even a
compact limitation note cannot fit, replay preserves accounting without a coverage
claim. Old snapshots remain unchanged. Historical records lacking the new identity
or owner-return endpoint cannot yield an exact wait duration or corrected hours.

Close and milestone-close measure their current candidate head and tracker
authority, including fixed-then-ship re-closes. They retain a fresh snapshot with
the new reviewed generation; completed preview keeps its prior immutable pins.

To repair a completed sample, use sdlc issue correct-actual --issue N --reason TEXT
(--dry-run previews it). It computes hours, preserves the close generation, and
records provenance. --exclude removes a proven invalid calibration sample when
reliable numeric evidence is unavailable. Interrupted corrections resume through
sdlc issue recovery reconcile --issue N without measuring again.

Ambiguous legacy calibration rows need an explicit --target-row <full SHA256>.
The refusal lists the exact ledger path, completion and eligible row IDs with
metadata and retry commands. Ask the TL/operator to identify the row by provenance;
never choose by ordering. Selection cannot admit a foreign or superseded row.
A pending or completed same-intent retry cannot change its original target.
If retained measurement cannot support a numeric repair, the error names an exact
same-target --exclude command for a proven invalid sample. It never supplies hours.

For the full per-segment breakdown, run `sdlc active-time` (the standalone engine).

FLAGS

  --issue <n>         issue ID to measure (required)
  --brain-dir <path>  brain capture root (bound descendants only; default ../brain)
  --milestone <Mx>    measure one milestone's increment (#234)

RELATED

  sdlc close          consumes this — its missing-`--actual` explainer runs the
                      same engine and prints the suggestion inline.
