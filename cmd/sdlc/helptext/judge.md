Run an LLM-as-judge check against the current diff. Fresh-context
subagent invocation — the anti-collusion property: the judge sees only
the constructed prompt, never the doer's session state. Diff-oriented
categories receive the patch inline; milestone-review receives a compact,
pinned repository-inspection manifest instead. The doer's incentive to declare
success doesn't propagate.

CATEGORIES

  dry              Look for DRY (Don't Repeat Yourself) violations.
  pure             Look for PURE-principle violations (side effects in
                   logic that should be pure).
  plan             Review changed issue files for plan completeness,
                   ticked items, log entries, status correctness.
  specs            Compare diff against atlas/ + README.md; update
                   stale documentation (the only category with write
                   permission).
  lessons          Emit a reminder to capture patterns in
                   workshop/lessons.md. No agent invocation.
  milestone-review Boundary code review per AGENTS.md §3. Resolves --base and
                   optional --head to immutable commits, then supplies exact
                   read-only Git inspection commands instead of patch bytes.

USAGE

  sdlc judge dry                                     review main..HEAD for DRY
  sdlc judge specs                                   review and update atlas/
  sdlc judge milestone-review --base SHA --head SHA  bounded fresh-eyes review
  sdlc judge dry --dry-run                           print prompt + would-be command line
  sdlc judge dry --agent codex                       use codex CLI instead of claude

FLAGS

  --base <ref>          override diff base (default: from gitx.DiffBase,
                        which honors COMPARE-SHA, falls back to origin/main
                        or merge-base main HEAD)
  --head <ref>          override diff head (default: working tree)
  --agent <name>        CLI to invoke: claude | codex | gemini.
                        Default: explicit --agent, then AGENT_CMD, then
                        PAIR_AGENT/current known agent signals, then claude.
  --tools <list>        comma-separated tool allowlist for claude. Default:
                        Read,Grep,Glob,Bash — read-only for ALL categories
                        (#62 M2: a judge is a reviewer, not a doer; it reports
                        findings and the main agent applies fixes). `specs`
                        used to get Edit,Write to auto-fix docs, which could
                        leave a passing gate's tree dirty and strand a merge —
                        that's gone. Pass --tools explicitly to widen.
  --issues-dir <path>   directory holding issue files. Default: $WF_ISSUES_DIR
                        or "workshop/issues".
  --history-dir <path>  directory holding archived issues. Default:
                        $WF_HISTORY_DIR or "workshop/history".
  --plans-dir <path>    directory holding optional durable plans for a manual
                        milestone-review. Default: $WF_PLANS_DIR or
                        "workshop/plans". Used only when --issue is present.
  --issue <n>           optional issue identity for milestone-review. When set,
                        the issue path must resolve and an optional canonical
                        plan is named; when omitted, the prompt visibly says
                        <unspecified> and performs an ad-hoc range review.
  --milestone <Mx>      optional milestone label appended to --issue in the
                        manual review prompt.
  --dry-run             print the prompt + would-be command line; do not
                        invoke the agent. Useful for verifying behavior
                        in restricted environments.
  --sandbox             pass auto-approve flags to codex/gemini (no-op for
                        claude). Set automatically when /.dockerenv exists.

OUTPUT CLASSIFICATION

Each invocation's output is classified as:

  clean       — agent confirmed no violations ("No DRY violations found.",
                "in sync", etc.)
  info        — informational reminder (e.g., the `lessons` REMINDER)
  failure     — agent reported findings (or no output) — content
                demands attention

The exit code is 0 for clean/info, 1 for failure. Callers (push/merge
auto-dispatch in M5/M6) can chain on the exit code.

ANTI-COLLUSION

Every invocation spawns a fresh subprocess. The agent has no inherited
session state from the doer that produced the diff. This is the
structural anti-collusion property — same model is fine, what matters
is that the judge's context is uncontaminated.

DEFERRED TO LATER MILESTONES

  - Parallel multi-category dispatch (auto-run dry+pure+plan+specs in
    one call) — M5/M6 will wire this into `sdlc push` and `sdlc merge`
    pre-flight.
  - --json structured output — current output is the agent's prose;
    parsing it into machine-readable findings is a later concern.
  - Interactive change acceptance for `specs` (the shell version prompts
    "Accept changes? [Y/n]" before committing the agent's edits) — M5
    will gate this behind `sdlc push`'s confirmation flow.
