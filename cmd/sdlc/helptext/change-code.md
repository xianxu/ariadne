Enter the implementation phase for an issue. Composes the gates
between planning (which happens on `main`) and code-changing work:

  1. Structural sanity   — does the issue have a filled-in Spec, a
                           non-empty Plan, and Done-when criteria?
  2. Estimate gate       — a positive `estimate_hours:` (#113) that
                           RECONCILES with an itemized `## Estimate`
                           block (#117): a fenced ```estimate block of
                           v2 primitives whose design/impl hours sum to
                           estimate_hours (no unitemized estimate). Set
                           it at start-plan. --no-estimate /
                           --no-estimate-recon bypass the two halves; the
                           block grammar + vocabulary live in
                           helptext/estimate.md.
  3. Quality judges      — fresh-context LLM review (skip with
                           --no-judge): plan-quality (is the plan
                           executable as-written — vague items, missing
                           test surface, undefined acceptance criteria?)
                           and estimate-quality (#117: was the
                           ## Estimate derivation actually applied, or
                           back-fitted to a predetermined total?).
  4. Branching strategy  — defaults to in-place (a branch in the
                           current checkout, carrying the working tree
                           forward). `--worktree=yes` for an isolated
                           worktree; `--worktree=ask` to be prompted
                           (with a sizing hint) or, headless, get the
                           agent sentinel.

WHEN TO USE

  After `sdlc claim --issue N` and after you've written the Plan
  section (and any separate workshop/plans/NNNN-*-plan.md). This is
  the verb that says "I'm done planning, let's change code." It
  refuses to start if the plan isn't ready — the gates catch the
  most common forms of "code first, plan after" drift.

  The pre-#39 verb `sdlc start` (which bundled claim + worktree
  creation) was split into `sdlc claim` + `sdlc change-code` because
  in practice the planning step takes hours-to-days, and the
  worktree decision wants to wait until you can size the work.

FLAGS

  --issue <n>         workshop issue ID; derives branch name from
                      issues/NNNNNN-*.md
  --name <X>          explicit branch name (overrides --issue)
  --worktree=<v>      yes (worktree) | no (in-place) | ask (prompt).
                      Empty (default) = in-place, silently. `ask` prompts
                      via tty or, when stdin isn't a tty, emits
                      ASK_BRANCHING_STRATEGY + exits 2 for the agent
                      harness to handle.
  --force <reason>    bypass gate refusals; the rationale is logged
                      to stderr and recorded in the audit trail.
  --no-judge          skip the plan-quality LLM judge.
  --no-structural     skip the deterministic structural checks.
  --no-estimate       skip the estimate_hours gate (#113).
  --dry-run           print would-be operations; do nothing.
  --agent <cli>       agent for the plan-quality judge.
                      Default: explicit --agent, then AGENT_CMD, then
                      PAIR_AGENT/current known agent signals, then claude.

EXIT CODES

  0   branch created successfully
  1   gate refused (structural / estimate / plan-quality) without --force
  2   branching decision deferred — see AGENT PROTOCOL below

AGENT PROTOCOL

  When `--worktree=ask` AND stdin is not a tty, change-code emits
  the sizing hint on stderr, prints `ASK_BRANCHING_STRATEGY` on
  stdout, and exits 2. (An unset flag now defaults to in-place silently,
  so this fires only on an explicit `ask`.) The xx-sdlc skill recognizes
  this contract:

    on exit 2 + ASK_<TOPIC> stdout line:
      issue the corresponding AskUserQuestion
      re-invoke sdlc change-code with the answer flag (here:
        --worktree=yes or --worktree=no)

  Future asks follow the same `ASK_<TOPIC>` pattern; xx-sdlc grows
  one mapping per topic.

EXAMPLES

  sdlc change-code --issue 39
    # default: structural checks + plan-quality judge + in-place branch
  sdlc change-code --issue 39 --worktree=yes
    # create an isolated worktree instead
  sdlc change-code --issue 39 --worktree=ask
    # be prompted (tty) / get the agent sentinel (headless)
  sdlc change-code --issue 39 --no-judge
    # structural only; skip LLM judge (faster, for trivial changes)
  sdlc change-code --issue 39 --force "quick docs typo fix"
    # bypass all refusals; rationale recorded

RELATED

  sdlc claim      claim the issue (commit + push the issue file)
  sdlc judge      manually invoke any judge category, including
                  plan-quality
  sdlc close      close an issue or milestone (the matching exit
                  gate at the other end of implementation)
