Enter the implementation phase for an issue. Composes the gates
between planning (which happens on `main`) and code-changing work:

  0. Flow                — infers the issue's flow (#231; see THE FLOW
                           below). On the quick flow, gates 1–3 do not run.
                           The flow is RECORDED in the frontmatter only
                           after the gates pass, just before the sync
                           commit, so a refused run leaves the issue as it
                           was.
  1. Structural sanity   — does the issue have a filled-in Spec, a
                           non-empty Plan, and Done-when criteria?
  2. Plan-quality judge  — fresh-context LLM review (skip with
                           --no-judge): is the plan executable as-written?
                           STATEFUL since #187 — see THE PLAN GATE below.
                           It runs BEFORE the estimate gates (#187 B1),
                           because the estimate is a function of the plan.
  3. Estimate gates      — a positive `estimate_hours:` (#113) that
                           RECONCILES with an itemized `## Estimate`
                           block (#117): a fenced ```estimate block of
                           v2-lineage primitives whose design/impl hours sum to
                           estimate_hours (no unitemized estimate). Derive
                           it AFTER the plan clears plan-quality — nothing
                           above this point mentions the estimate.
                           --no-estimate / --no-estimate-recon bypass the
                           two halves; the block grammar + vocabulary live
                           in helptext/estimate.md. Then estimate-quality
                           (#117: was the derivation actually applied, or
                           back-fitted to a predetermined total?).
  4. Branching strategy  — defaults to in-place (a branch in the
                           current checkout, carrying the working tree
                           forward). `--worktree=yes` for an isolated
                           worktree; `--worktree=ask` to be prompted
                           (with a sizing hint) or, headless, get the
                           agent sentinel.

THE FLOW (#231)

  Every issue runs one of two flows, and you run the same verbs either way —
  the gates decide what applies. change-code infers the flow from what the
  design produced:

    Mx milestone rows in `## Plan`, or a durable plan
    (workshop/plans/<issue>-plan.md)                        → full
    neither                                                  → quick

  and records it on the issue as one line,
  `flow: {kind: quick, provenance: inferred, spec: "…", done: "…"}`.

  full   today's gates: structural, plan-quality, the estimate gates, and the
         full boundary review at close.
  quick  none of change-code's gates, no durable plan, no estimate — and one
         review, at close, with the small-diff recipe. The hard shell is

           {{QUICK_SHELL}}

         `sdlc close` measures the real diff, and a quick issue outside the
         shell is upgraded to full and gets the full review, whoever chose
         quick. The `spec`/`done` hashes let close refuse a contract that moved
         without its `## Done when`.

  The operator can pin the flow with `--flow quick|full` (provenance:
  operator). A pin to quick is refused on a Plan with Mx rows. Gates never
  downgrade full to quick; only a pin does.

  The record is re-derived from the issue as it is when written, so an edit
  made while the gates ran survives. If that edit changed the flow itself (say,
  Mx rows appeared), the gates ran for the wrong flow: change-code refuses and
  asks for a re-run.

THE PLAN GATE (stateful since #187)

  The plan-quality gate REMEMBERS. Its findings persist to
  `workshop/plans/NNNNNN-slug-plan-gate.md` with binary-assigned stable
  ids, and each re-run must dispose of every prior finding
  (addressed / not-addressed / withdrawn) before raising new ones.

  A finding names one instance; what disposes of it is the class it belongs
  to — see ARCH-PURPOSE (`sdlc arch-principles`), which the gate's own
  refusal routes to.

  Only *undisposed* Critical or Important findings block. New Minor
  findings are recorded and carried to the close review — they cost no
  round-trip. Past WF_PLAN_ROUND_CAP rounds (default 3) only Critical
  blocks; the rest are recorded and reach the boundary review, which is
  instructed to read the ledger.

  If the issue + plan are byte-identical to what a previous round
  ACCEPTED, the gate passes through without re-dispatching the judge —
  so fixing an estimate refusal costs milliseconds, not a fresh review.
  The estimate itself is excluded from that comparison, since #187 B1
  removed it from this gate's remit.

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
  --no-estimate-recon skip the `## Estimate` reconciliation gate (#117).
  --flow <kind>       pin the flow: quick | full — the operator's decision
                      (provenance: operator). Default: inferred (#231).
  --dry-run           print would-be operations; do nothing.
  --agent <cli>       agent for the plan-quality judge.
                      Default: explicit --agent, then AGENT_CMD, then
                      PAIR_AGENT/current known agent signals, then claude.

ENVIRONMENT

  WF_BOUNDARY_ROUND_CAP
                      the same knob for the boundary review's gate ledger
                      (`sdlc close` / `milestone-close`, #194); default 3
  WF_PLAN_ROUND_CAP   rounds after which only Critical findings block the
                      plan gate (default 3). Bounds the tail of a review
                      that keeps descending severity levels.

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
