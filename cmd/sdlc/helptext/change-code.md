Enter the implementation phase for an issue. Composes three gates
between planning (which happens on `main`) and code-changing work:

  1. Structural sanity   — does the issue have a filled-in Spec, a
                           non-empty Plan, Done-when criteria, and a
                           positive `estimate_hours:`?
  2. Plan-quality judge  — fresh-context LLM review: is this plan
                           executable as-written, or does it have
                           vague items / missing test surface /
                           undefined acceptance criteria?
  3. Branching strategy  — worktree (isolation) or in-place branch
                           (carry working tree forward)? Decided
                           with a sizing hint derived from the plan.

WHEN TO USE

  After `sdlc claim --issue N` and after you've written the Plan
  section (and any separate workshop/plans/NNNN-*-plan.md). This is
  the verb that says "I'm done planning, let's change code." It
  refuses to start if the plan isn't ready — the gates catch the
  most common forms of "code first, plan after" drift.

  Use the previously-existing `sdlc start` to do both claim and
  change-code in one shot ONLY if you genuinely know the plan
  already (rare). The two verbs were split (#39) because in
  practice the planning step takes hours-to-days, and the worktree
  decision wants to wait until you can size the work.

FLAGS

  --issue <n>         workshop issue ID; derives branch name from
                      issues/NNNNNN-*.md
  --name <X>          explicit branch name (overrides --issue)
  --worktree=<v>      yes (worktree) | no (in-place). Empty (default)
                      asks the operator via tty prompt or, when stdin
                      isn't a tty, emits ASK_BRANCHING_STRATEGY +
                      exits 2 for the agent harness to handle.
  --force <reason>    bypass gate refusals; the rationale is logged
                      to stderr and recorded in the audit trail.
  --no-judge          skip the plan-quality LLM judge.
  --no-structural     skip the deterministic structural checks.
  --dry-run           print would-be operations; do nothing.
  --agent <cli>       agent for the plan-quality judge.
                      Default $AGENT_CMD or claude.

EXIT CODES

  0   branch created successfully
  1   gate refused (structural / plan-quality) without --force
  2   branching decision deferred — see AGENT PROTOCOL below

AGENT PROTOCOL

  When --worktree is unset AND stdin is not a tty, change-code emits
  the sizing hint on stderr, prints `ASK_BRANCHING_STRATEGY` on
  stdout, and exits 2. The xx-sdlc skill recognizes this contract:

    on exit 2 + ASK_<TOPIC> stdout line:
      issue the corresponding AskUserQuestion
      re-invoke sdlc change-code with the answer flag (here:
        --worktree=yes or --worktree=no)

  Future asks follow the same `ASK_<TOPIC>` pattern; xx-sdlc grows
  one mapping per topic.

EXAMPLES

  sdlc change-code --issue 39
    # default: structural checks + plan-quality judge + ask
  sdlc change-code --issue 39 --worktree=yes
    # skip the ask; create a worktree
  sdlc change-code --issue 39 --worktree=no
    # skip the ask; branch in place
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
