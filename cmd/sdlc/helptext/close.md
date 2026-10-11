Close an issue or a milestone — perform AGENTS.md §5's mechanical closing
steps. Edits files in place; for an issue whose details mirror a tracker card
(#252) the whole-issue close then commits its own evidence and publishes the
card (see TRACKER-ERA ISSUES); otherwise the agent commits, usually bundling
close with other work.

{{OWNERSHIP_GATE}}
The check runs before the review, so a refused close dispatches no judge.

`sdlc close` is the LOCAL ACCEPTANCE GATE (#160): it runs the fresh-context
boundary review (all LLM review — code quality, requirements traceability, docs
sync incl. README, architecture) and flips a full issue to `codecomplete`, NOT
`done`. The deterministic publish gate (`sdlc merge`/`push`) later flips
`codecomplete → done` after verifying nothing drifted since close. close is the
SOLE writer of `codecomplete` (set-status refuses it), which makes the commit
carrying it a trustworthy anchor for that reviewed-HEAD-unchanged invariant.

POST-VERDICT PROTOCOL (#174): on FIX-THEN-SHIP, close prints it — fix the
findings NOW (before committing), bundle fixes + the issue-file mutations +
bookkeeping (lessons, plan ticks) into ONE commit so the publish anchor is
HEAD, and do NOT re-run close. On the quick flow, one exception:
{{QUICK_AFTER_REVIEW}}. Fixes that must land after the close commit:
`sdlc close --fixed-then-ship` reviews only that delta and advances the anchor
(see FIXED-THEN-SHIP; no bypass flag needed at codecomplete). Doc-only
post-close commits pass the publish gate on their own. "Fix the findings" means the CLASS each one names, not only the
site it points at — ARCH-PURPOSE (`sdlc arch-principles`).

TRACKER-ERA ISSUES (#252)

  status, actual_hours and updated live on the issue's card, not in the details.
  After a finalizing verdict the whole-issue close commits its evidence itself —
  the details' Log line, the gate ledgers, review sidecars and this repository's
  project records, with the Review-Verdict/Review-Window trailers and a
  `Close-Actual:` trailer — as one narrow commit on the issue branch (unrelated
  staged work is left alone). It then publishes codecomplete, the actual hours
  and a completion binding (reviewed head, evidence commit) on the card. Merge,
  push and slot landings select closes by that binding and check the reviewed
  state against the evidence commit. close reports success only after the card
  is confirmed; an interruption is finished by `sdlc issue recovery reconcile
  --issue N`. A second narrow commit then mirrors the codecomplete card into
  the details (#275). The evidence commit cannot carry that mirror, because
  the card names the evidence commit. An uncommitted details edit keeps its
  body and gets the same mirror.

  On FIX-THEN-SHIP close publishes nothing yet: commit the fixes on the branch,
  then run `sdlc issue recovery reconcile --issue N`, which lands the evidence
  commit after them and publishes the card bound to it. To have the fixes
  reviewed instead, `sdlc close --issue N --fixed-then-ship` reviews just them
  and supersedes the pending close.

FIXED-THEN-SHIP (#183)

  `sdlc close --issue N --fixed-then-ship --verified '<evidence>'` is the narrow
  re-close, run after a finalized close review (SHIP or FIX-THEN-SHIP) when code
  landed after it: the review's fixes, a post-close change, or a resolution of a
  conflict with main (the publish gate names this path when it refuses).

    - Window: the interdiff since the head that close review read, replayed onto
      today's main (the same replay as a later milestone's window, #304). Main's
      own changes never enter it; a conflict shows only its resolution.
    - Review: the fix-review recipe. It disposes of each prior finding, raises
      only defects the interdiff introduces, judges every conflict resolution,
      and returns REWORK for a scope change (`## Spec`, `## Revisions` or
      `## Done when` changed beyond recording the fix): run a full close then.
      The window's diff excludes the issues directory, so close compares those
      sections itself (the issue as the close review read it, against now) and
      hands each changed section to the reviewer, before and after.
    - The atlas gate and the quick-flow shell are measured on the branch patch,
      which the earlier close already held to them.
    - Finalizing is an ordinary close: a new evidence commit, the card rebound to
      it, so the publish gate's anchor advances.

  It refuses, before any review:
    - when no close review has finalized (nothing to fix against), naming the
      full close;
    - on the quick flow, when the branch patch has left the shell (the full
      review is due);
    - when the window resolves a conflict with main in a code file and no
      `--full-suite '<evidence>'` records a full test run. Only the review
      narrows; the suite still runs (operator decision, #320). The evidence
      goes into the close's ## Log line.
    - with the review skipped (its bypass, or --force), because it IS a review.

  A milestone has no such flag: fixes after a milestone close fall in the next
  boundary's interdiff.

MODES

  Issue close:      sdlc close --issue 15 --actual 7 --verified '<evidence>'
  Milestone close:  sdlc milestone-close --issue 15 --milestone M4 --actual 2.5 --verified '<evidence>'

  (Milestone closing lives on `sdlc milestone-close` — `close` no longer takes
   `--milestone` (#146). See `sdlc milestone-close --help`.)
  (The --actual values above are MEASURED — from `sdlc actual` or, when the
   flag is omitted, adopted directly from the measurement (#178) — not typed
   estimates. See the --actual flag note.)

  THE BOUNDARY REVIEW (#69). A standalone full-issue close auto-dispatches the
  one binary-owned fresh-context review on the whole-issue window — the branch
  patch, diff(merge-base(main, HEAD), HEAD), which main merged in never enters
  (#304); close prints it as "branch patch vs main@<base>: <n> issue commit(s)" (the same
  reviewer `milestone-close` runs per-milestone). For a no-milestone issue this
  is the single review the boundary gets; for a multi-milestone issue it's the
  end-of-issue integration review on top of the per-milestone ones. The agent
  does NOT separately run `superpowers-requesting-code-review` (AGENTS.md §3).
  Skip with `--no-judge` (records a not-run trailer). Milestone slices are
  reviewed by `sdlc milestone-close` (per-milestone); to skip THAT review
  explicitly, use `sdlc milestone-close --no-judge` — `close` no longer has a
  `--milestone` path to skip it silently (#146).

  THE GATE LEDGER (#194). The boundary review REMEMBERS across rounds, the way
  `change-code`'s plan-quality gate has since #187. Each review's findings are
  written to `workshop/plans/NNNNNN-slug-close-gate.md` with binary-assigned
  stable ids (`BR-1`, `BR-2`, …), and the next review is shown them and must
  dispose of each — `addressed`, `not-addressed`, or `withdrawn` — before
  raising anything new. That is what stops a re-run renumbering the same
  finding forever. The implementer's answer is `sdlc finding dispute` (#202):
  a rationale plus a counter-measurement, which the next review must withdraw
  the finding over or re-raise it against.

  A MILESTONE review is shown its own boundary's findings. The WHOLE-ISSUE
  close is shown ALL of them, from every milestone: it is the last gate before
  publish, and a finding left undisposed when its milestone closed has no other
  path to disposal. (The round cap still scopes per boundary — three milestones'
  rounds must not arrive at the close already past it.)

  Two consequences worth knowing before you hit them:

    - A close can REFUSE DESPITE A PASSING VERDICT. Finalizing needs the
      verdict AND the ledger to clear; a SHIP carrying an undisposed
      Important means the reviewer contradicted itself. Fix the findings, or
      have the next review dispose them explicitly — and fix the CLASS each
      names, not only the site (ARCH-PURPOSE, `sdlc arch-principles`).
      `--no-ledger` waives just this refusal.
    - The ledger is NOT the prose sidecar. `-close-review.md` holds boundary
      metadata plus the reviewer's semantic final response for a human;
      `-close-gate.md` holds addressable findings for the next round. Harness
      diagnostics and progress stay on terminal stderr. Neither replaces the
      other.

  FINDING FAMILIES (#194). Each finding carries a `family:` slug naming the
  underlying RULE it is an instance of, not its symptom. When a family already
  has a finding from an earlier round, the next reviewer is told so and its
  recommendation changes: instead of "fix this instance" it is asked to state
  the rule that covers all of them. That is the difference between patching
  four shapes of one missing rule across four rounds and writing the rule on
  the second.

  That is also what the field asks of YOU (#203). A finding names one instance;
  the deliverable is the class — and `family:` names the class for you, so the
  ledger is a worklist, not a label. Each gate refusal says so and routes to the
  principle that governs it: **ARCH-PURPOSE** (`sdlc arch-principles`). A family
  that repeats across rounds is this ledger reporting that the enumeration was
  never written.

  Every round also emits a convergence line:

      round 4 — 2 new findings, 0 repeat families, 6 disposed. Converging.
      round 2 — 3 new findings, 2 repeat families, 1 disposed. Not converging:
        fix rules, not instances.

  Capping rounds on finding COUNT is arbitrary; a family that stops repeating
  is an honest signal that the work is converging.

  Findings the plan gate deferred (Minor, or demoted past its round cap) are
  seeded into this ledger on its first round, so they arrive as `BR-*` ids you
  can dispose of rather than as a file the reviewer is told to go read.

  Past WF_BOUNDARY_ROUND_CAP rounds (default 3) only Critical still blocks. A
  demotion here is louder than at the plan gate on purpose: there is no later
  gate to pick up what it drops, so each demoted finding is named.

  ANCHORED TO THE COMMIT IT READ (#194). The review records the concrete SHA it
  reviewed — in the `Review-Window:` trailer and the sidecar — and finalizes
  against that commit, not against "HEAD has not moved". A commit landing while
  the review runs is classified, not fatal:

      doc-only delta  the close finalizes; the reviewed code surface is intact,
                      and the same classifier `sdlc merge` applies at publish
                      time decides it (#174)
      code delta      refused, naming every commit the review did not cover
      rewritten       refused as diverged (HEAD no longer descends from the
                      reviewed commit), since the delta is not describable

  So committing `workshop/lessons.md`, `atlas/` and `workshop/plans/` bookkeeping
  during a review is safe. Landing code is not, and the refusal now tells you which
  commits to re-review. Every mutable prose artifact named to the reviewer is
  separately snapshotted: the **issue file**, any **project file** the close
  edits (a TL-owned one is skipped and not snapshotted, #326), and the
  canonical durable **plan** (including whether that plan exists). Any edit,
  creation, deletion, or replacement refuses finalization — the review READ that
  prose — so write your `## Log` line AFTER the close, not during it.
  Repository/worktree/branch identity and the boundary ledger also participate;
  the first boundary round protects the plan ledger it seeds from. Validation
  happens after reacquiring the lock and before persisting ANY verdict, sidecar
  or ledger round. Stale inputs, cancellation, failed dispatch and failed lock
  reacquisition cannot be waived by --force.

  External reviews run without holding the repository lock. The review's
  timeout scales with its window (30m up to 500 added lines, +15m per further
  1,000, at most 2h; #300); WF_REVIEW_TIMEOUT overrides it with a Go duration
  from 1s through 2h. Interruption stops
  without a review result; reviewer shutdown and pipe draining have a five-second
  grace bound.

  The reviewer runs in a disposable shared clone of the pinned head (#204),
  removed when it exits; a crashed owner's clone is swept by the next review.
  An OS write fence (sandbox-exec / bwrap), verified live before each review,
  keeps it out of every worktree of the repository and its git dir; where the
  fence cannot nest (inside an agent sandbox) the round runs clone-only with a
  warning. A no-verdict retry starts from a reset clone. The round and sidecar record
  `isolation: fenced|checkout`. WF_REVIEW_CHECKOUT_ROOT relocates the clones.

  (All lifecycle verbs — claim/start-plan/change-code/milestone-close/close/
   merge/push — also carry the repo guard (#176): they refuse in a brain
   (capture) repo and in repos without workshop/issues/. WF_SPINE_GUARD=off is
   the logged emergency hatch.)

WHAT THE GUARD DEFENDS

  --actual <hours>     focused dev-hours. Omit it: close measures and ADOPTS
                       the value (#178; loud info line with window + attribution).
                       An explicit value is deviation-checked (#87). Run
                       `sdlc actual --issue N` to preview. Required by hand only
                       when measurement fails.
  --verified '<line>'  one-line evidence the work meets done-when (behavior,
                       not artifacts: "tests pass" beats "code written").
                       Required.

  Plus structural checks:
    - integrate (#269, close and milestone-close; runs first, before any
      review): the branch must contain every commit that landed on main
      (first parent, freshly fetched) by the most recent noon, in the zone of
      .sdlc/freshness.json ({"version":1,"timezone":"<IANA>"}; absent → the
      local zone). Missing code refuses with the remedy: `git merge <main>`
      for a pushed branch, `git rebase <main>` for an unpushed one; then
      re-run the suite and the close. Missing docs only → an info line. A
      failed fetch warns "could not check", never "up to date". Integrating
      never widens the review window (#304).
    - atlas/ must have changed in the issue's commit window (§5 step 5) —
      auto-satisfied when the window has NO code surface (docs/workshop-only
      deltas have nothing to map, #177)
    - issue's `## Plan` has no unchecked items (issue close only)
    - each milestone listed in ## Plan must carry a `Review-Verdict:`
      trailer on its close commit (issue close only; AGENTS.md §3).
      A Plan with only plain `- [ ]` checkboxes (no `Mx` rows) has no
      such requirement — atomic single-pass work closes in ONE `sdlc
      close`, no milestone-close, one `closed —` log line. Reserve `Mx`
      tags for ≥2 genuinely separate review boundaries (AGENTS.md §3).
      TRAILING Mx rows that never got a milestone-close (over-split
      single-pass work) are accepted with an info line — the issue-close
      boundary review covers them (#175) — unless --no-judge skips that
      review; a MIDSTREAM miss (a later Mx closed with review) still
      refuses.
    - milestone-close ticks the `- [ ] M4 — ...` row; refuses if absent
    - every project record across the fleet referencing <repo>#<id>
      (each peer's workshop/projects/*.md; the deprecated brain legacy
      home still scans with a migration warning, #171) gets its task row
      ticked + detail block updated; a PEER repo's edit is committed
      there only when git state is unambiguous (main, clean), else
      reported with the exact finish command
    - a TL-OWNED project (frontmatter `tl:`, #326) is skipped entirely: no
      tick, no detail block (so no --no-project refusal), no peer-write, no
      review snapshot. An info line names it and its TL; report the landing
      to the TL, who owns the file

  BYPASSING A GATE (#67) — each gate has its own --no-<gate> flag, so you
  can waive exactly the one that doesn't apply (and acknowledge it) instead
  of reaching for the blanket --force:

{{GATE_FLAGS}}

  Each bypass logs an audit "[!] --no-X: skipping ..." line (it's an
  explicit acknowledgment, not a silent skip) and the rationale belongs in
  --verified. --force ≡ all of them at once (emergencies). Prefer the
  precise flag: e.g. a pure bugfix with no new architectural surface closes
  with `--no-atlas` (reason in --verified), NOT --force.

WHAT IT DOES

  - Ticks the milestone box in the issue's ## Plan (milestone mode)
  - Flips status: codecomplete (#160 — NOT done; the deterministic publish gate
    `sdlc merge`/`push` flips codecomplete → done), sets actual_hours (number or
    N/A) and updated (issue mode) — in the issue file, or, for a tracker-era
    issue, on its card after the evidence commit (#252)
  - Appends a log line to ## Log: "YYYY-MM-DD: closed — <verified>"
  - Emits the no-LLM `lessons` reminder on a whole-issue close (#160 Q4 — moved
    here from the publish gate so it fires while findings are fresh)
  - Ticks the project task row + upserts **actual:** and **closed:** in the
    detail block (skipping TL-owned projects, those with `tl:`, #326)
  - Prints the COST REPORT (#187) — see below
  - Does NOT git-commit (except a tracker-era issue's evidence commit, #252),
    does NOT move the file to workshop/history/

THE QUICK FLOW (#231)

  An issue whose frontmatter records `flow: {kind: quick, …}` (written by
  `sdlc change-code`, see its THE FLOW) closes with ONE review, and close first
  measures the review window against the hard shell:

    {{QUICK_SHELL}}

  Added lines come from the window's diff, over code files as the shell above
  classifies them. The design is measured as it stands at close, so a plan that
  grew during the work crosses too. Neither how many files a change spreads
  across nor where in the tree it lands counts, only its size: for a diff this
  small, the tests and the one close review are the guard. (In ariadne, a quick
  change to sdlc itself is closed by the binary it just changed; that review
  sees the diff.)

  Inside the shell, the boundary review runs the small-diff recipe: the full
  procedure, aimed at the bug classes small diffs ship, over only the ARCH-*
  principles the registry marks `quick-flow: yes`. Outside it — whoever chose
  quick — the issue is upgraded to `flow: {kind: full, provenance: inferred}`,
  the measured reason goes to ## Log, and the full review runs. Crossing the
  shell never refuses. The upgrade is written with the close's other edits at
  finalize, so a REWORK leaves the issue unwritten — but each boundary round
  records its recipe in the boundary ledger, and a round that already ran the
  full review keeps the issue on it, even if the fix shrinks the diff.

  Before that one review, two checks (issue close only):
    - `## Done when` has a bullet — it is the review's only oracle.
    - If `## Spec` or `## Revisions` changed since change-code but `## Done when`
      did not (F3, #183), the close review judges whether Done-when still covers
      the work. Its prompt says so, a stale Done-when blocks the verdict, and
      the ## Log line records "done-when: contract moved since change-code;
      judged by the close review" with the verdict. A deterministic check cannot tell a
      reframe from a review-fix Revisions entry; the reviewer can. With the
      review skipped there is no reviewer, so close refuses: restate the
      acceptance criteria, or pass the Done-when freshness bypass (BYPASSING A
      GATE) with the reason in --verified.

  Fixes made after the verdict ride into the close commit, which close does
  not measure, so the publish check (`sdlc merge`/`push`) re-measures over
  close's own window: {{QUICK_AFTER_REVIEW}}. That close finds the shell
  crossed, upgrades the issue and runs the full review.

  An issue with no `flow:` line (every issue before #231) closes exactly as before.

THE COST REPORT (#187)

  Two lines, over the same window the boundary review saw:

    [ok] churn: prod 554 / test 300 / atlas 20 / workshop 778 (final 1652, rework 2.4×)
    [ok] plan gate: 6 round(s), 1 forced; findings 4 addressed / 1 withdrawn / 0 disputed / 2 still open

  rework is insertions-across-commits over insertions-in-the-final-diff: 1.0
  means every line written survived; 2.4× means 2.4 lines written per line kept.
  It is the number a final diff cannot show — a file rewritten five times looks
  identical to one written once.

  Both lines are UNCONDITIONAL: unlike the calibration row, they print on a
  milestone close, under --no-actual, and with no sibling brain/. Every value
  degrades to zero with a warning rather than failing the close; an ABSENT
  plan-gate sidecar is not a failure (it is every issue closed before #187) and
  reads as honest zeroes.

  The same values append to the calibration ledger as ten columns
  (churn_prod … gate_open, indices 10–19). Columns are APPENDED, never
  reordered — ParseRows indexes positionally and legacy rows must keep parsing.

REVIEWER (#310)

  The boundary reviewer (close and milestone-close) is chosen per repo by
  workshop/review.json, a repo-local file the base layer never carries:

    {"boundary_reviewer": "other"}
    {"boundary_reviewer": "explicit", "reviewer_pairs": {"claude": "codex"}}

  same (the default; no file) — the main agent reviews itself
  other    — claude and codex review each other; gemini is reviewed by claude
  explicit — reviewer_pairs[main agent]; an unpaired main agent reviews itself

  The main agent is detected from PAIR_AGENT, then codex's env, then claude.
  --agent and AGENT_CMD override the file. An unknown mode, agent or key halts
  the review with the file named. The ledger round records the reviewer.

  No judge runs in an auto-approve mode. claude gets `--permission-mode
  dontAsk`: only Read,Grep,Glob,Bash run, so Edit/Write are refused, but Bash
  itself is unrestricted. codex gets `--sandbox workspace-write` in the review
  checkout (writes only the checkout), `danger-full-access` under the OS fence,
  `read-only` outside a checkout. gemini runs without --yolo (headless, it
  refuses tools needing approval). What stops a shell write to the source
  checkout is the OS fence (#204), whatever the agent; without it, only
  codex's own sandbox does. Plan-quality and `sdlc judge` run in the operator's
  tree with no fence, so a claude judge's Bash can still write there.
  Codex's sandbox is its own Seatbelt on macOS, which cannot nest where the
  fence probe failed (sdlc inside Claude Code's sandbox): its commands are then
  refused and close warns. Run a codex reviewer from an unsandboxed sdlc.

WARMUP

  On the first 2 invocations per shell session, prints the close-issue
  contract to stderr. After that, silent. Reset by starting a new shell.

FLAGS

  --issue <n>           ariadne workshop issue ID (numeric, zero-pad
                        applied internally; required)
  --actual <hours>      focused dev-hours (required unless --no-actual/--force)
  --verified '<line>'   one-line behavior evidence (required unless --no-verified/--force)
  --force               bypass ALL gates (≡ every --no-* flag); reason in --verified
  --no-<gate>           one flag per gate, each waiving exactly that gate — the
                        table under BYPASSING A GATE; docs-only windows
                        auto-satisfy the atlas gate (#177)
  --agent <cli>         agent CLI for the boundary review (claude | codex | gemini)
                        Default: explicit --agent, then AGENT_CMD, then the
                        repo's reviewer policy (see REVIEWER) applied to the
                        main agent: PAIR_AGENT/current agent signals, or claude.
  --dry-run             print what would change, write nothing
  --fixed-then-ship     review only the changes since the last finalized close
                        review, with the fix-review recipe (see FIXED-THEN-SHIP)
  --full-suite '<line>' with --fixed-then-ship: evidence the full test suite
                        passed; required when the window resolves a code conflict
  --brain-dir <path>    brain root for the calibration ledger (default ../brain);
                        project files are discovered across the fleet (#171)
  --issues-dir <path>   issues directory (default workshop/issues)
  --plans-dir <path>    durable plans + gate sidecars (default workshop/plans);
                        where the cost report finds the plan-gate ledger (#187)

DEEP-DIVE REFERENCES

  AGENTS.md §5                       closing checklist
  brain/data/life/42shots/velocity/  v3 attribution method
    baseline-v3.md
  construct/datatype/project.md      project-file shape & detail blocks

If --actual is missing, close runs active-time-v3 itself (brain + repo
transcript dirs, this issue's window + auto-discovered peers) and ADOPTS the
measured value in the same invocation (#178) — a loud info line shows the hours,
window, and peer attribution. When the telemetry isn't available it refuses and
points you to a labeled judgment estimate. Attribution warnings print alongside
the adoption; inspect them, because they identify dominant long runs or mention
fallback where commit boundary evidence was weak. The measured value is BUILD
+ CONVERGE (#234): it is split at each boundary review's first dispatch, and the
verification after a finalized boundary's last #N commit (the final test run and
review) is not counted, at every boundary. `sdlc actual --issue N` shows the
split. If
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
guessed --actual would otherwise sail through. (milestone-close inherits it,
comparing against the milestone's own measured increment, #234.)

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
