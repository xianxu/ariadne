---
type: continuation
slug: quick-flow-trial
agent: claude
session_id: a99ec60f-d0a5-4065-9547-56320127da3f
created: 2026-09-18T09:35:30
branch: 000231-a-quick-path-for-small-diffs-scale-the-gate-set-and-the-review-recipe-to-size
worktree: /Users/xianxu/workspace/ariadne
issues: [000231, 000233, 000234, 000235]
---

# Continuation: quick-flow-trial

## NEXT ACTION

Gather the trial evidence before touching code. In each fleet repo the operator
worked in since 2026-09-18 (pair, parley.nvim, tools, …), classify every issue
that closed with a `flow:` line into four groups:

- **stayed quick**;
- **quick, then upgraded at close** — note which crossing, from its Log line
  `flow upgraded quick → full — …`;
- **full**, via a durable plan or `Mx` rows;
- **operator-pinned**.

Start with `git -C /Users/xianxu/workspace/<repo> log --since=2026-09-18 -p -- workshop/issues | grep '^+flow:'`,
then read the flow columns (indices 20–22) in
`/Users/xianxu/workspace/brain/data/life/42shots/velocity/calibration-ledger.tsv`.
Then ask the operator for their read of how it felt, and only then turn findings
into fixes on the #231 branch. Why this order: the operator wants to learn
whether the quick path is *useful* and tweak it from real use. The evidence
lives in other repos' issues, and the Open questions below are exactly what that
data should settle.

## State of play

- **#231** (working; branch `000231-a-quick-path-for-small-diffs-scale-the-gate-set-and-the-review-recipe-to-size`):
  - **M1 is closed**, SHIP after 3 boundary rounds.
  - **M2 is closed**, FIX-THEN-SHIP at the round cap after 4 rounds; the fixes
    are bundled in close commit `3381378`.
  - **M3's work is done**: ledger flow columns, the drift exclusion, docs, and
    the parley.nvim#263 fixture replay.
  - **Deliberately NOT run yet:** the final `sdlc close --issue 231` (the
    whole-issue review, meant to cover trial fixes), then `sdlc pr` → `sdlc merge`.
- **ariadne's primary checkout stays on the #231 branch, clean.** The operator's
  `sdlc` shell function rebuilds `bin/sdlc` from `~/workspace/ariadne`'s working
  tree on every call, so every repo runs this branch, uncommitted edits
  included.
- **#233** (open): the `sdlc config` flow for declarative changes, split out of
  #231 and building on it. Decide after the trial.
- **#234** (open, depends on #231): split actuals into build (claim → first
  boundary-review dispatch) and converge (→ finalize); calibrate k =
  converge/build; backfill the 526-row ledger. Its columns go after #231's.
- **#235** (open): durable-plan lookup is exact-name, found in pair#283. It takes
  the full flow in ariadne, by the operator's decision.
- **Tests:** `go test ./...` is green except the pre-existing #210 failure
  (`TestFleetPlanHasAuthoritativeCorrectedCoreConceptInventory`).
- **Untracked here:** `workshop/issues/000234-*.md` and `000235-*.md` are
  already on trunk and arrive with `main`.

## Thread arc & user model

The session opened with lookups (couch's pid, the couch-slots project, "which
task is the fast path"). It then turned into the operator reshaping #231 over
several passes:

- skip the estimate; one close gate; only ARCH-DRY, PURE and PURPOSE;
  operator-instructable; split the config flow out;
- then: no new verb — the gates infer the flow, and the agent's sequence of
  verbs is identical on both flows; `flow: {kind, provenance}` in frontmatter
  and cue;
- then: build it to smoke-test readiness, leave ariadne on the branch, trial it
  in other repos.

Midway the operator pushed on the process's economics:

- review findings, Minors included, are fixed in the round, because a follow-up
  issue costs more in process than the fix;
- an estimate should not price review rounds; a calibration factor should. That
  became #234.

**User model:** the operator is optimizing the throughput of the agentic SDLC
itself. They cut ceremony where its expected value is low (small diffs), but keep
hard, mechanical backstops (the shell) rather than trusting judgment. They prefer
measurement over prediction ("fix actuals; recompute history"), simple rules over
clever ones ("keep it — not worth additional rule/risk"), and process the agent
doesn't have to think about (gates decide). They will judge the quick path on
real use, not on argument.

## Open questions

On resume, resolve these open questions with the user before continuing with the NEXT ACTION.

- **Spread-but-tiny changes and the durable plan.** pair#283 was 7 code files
  and 30 added lines. The quick-then-upgrade path already yields "no plan, full
  review" for such changes, but the constitution still asks for a plan outside
  the shell. Should they skip it?
- **An in-issue `## Plan` of real size.** Should it count toward full (a
  plan-item count via `issue.ComputeSizingFromContent`), or stay invisible to
  inference?
- **The limits.** Are 2 code files and 100 added lines right (tests and docs
  excluded)? Let the trial data say.
- **Other repos' shared surfaces.** Should they declare `.sdlc/shared-surfaces`?
  parley has none, so its keybinding registry and config schema are
  unprotected, which was #263's case.

## Artifact map

- **Read first:** `/Users/xianxu/workspace/ariadne/workshop/issues/000231-a-quick-path-for-small-diffs-scale-the-gate-set-and-the-review-recipe-to-size.md`.
  - Spec §1–5 are the design as revised with the operator; Done-when is the
    acceptance.
  - The Log records each milestone, the #263 fixture results, the first trial
    reading (pair#283), and the build-closure decision.
  - Revisions is the dated record of every operator decision.
- **Then the durable plan**, `workshop/plans/000231-…-plan.md`: Core concepts;
  Decisions with ARCH-*; Revisions, one per review round, BR-4..BR-33, each
  naming the rule a class of findings became.
- **Review evidence:** `workshop/plans/000231-…-{plan-gate,close-gate,m1-review,m2-review}.md`.
- **Code, all under `cmd/sdlc/`:**
  - `internal/flow/`: the record codec, `Decide`, limits and `ShellSummary`, the
    contract hashes and Done-when checks, `Measure`/`Crossings`, surfaces and
    `SurfaceForms`.
  - `changecode_flow.go`: inference; the record is written after the gates.
  - `closeflow.go`: the shell at close, the upgrade, the recipe; stickiness
    comes from the ledger's `Round.Recipe`.
  - The judge: `internal/judge/prompts/small-diff-review.md`, `boundary-tail.md`,
    and the `quick-flow:` field in `architecture.md`.
  - `internal/processmanual` `GateCatalog.Gate`, rendered through `{{GATE_FLAGS}}`.
  - The ledger's flow columns in `internal/estimate`.
  - `.sdlc/shared-surfaces`: ariadne declares sdlc's whole build closure.
- **Tool:** `cmd/sdlc/smalldiffreplay_test.go` (manual tag) replays the recipe on
  another repo's shipped window. The parley.nvim#263 replay hit BR-2, BR-1's
  class and BR-9. Replay from clones checked out at the reviewed head, because
  archived issue files break manifest resolution.
- **Related issues:** #233, #234 and #235 (ariadne `workshop/issues/`), and peer
  `/Users/xianxu/workspace/pair/workshop/issues/000283-*` (the first trial task).
- **brain:** `data/life/42shots/velocity/SKILL.md` carries a flow-column filter
  note, uncommitted, awaiting brain's auto-commit.

## Decisions & dead ends

- **No `sdlc quick` verb.** The flow is frontmatter state that gates infer and the
  operator pins (`change-code --flow`). A verb would duplicate claim/change-code
  and couldn't size a diff that doesn't exist yet.
- **The shell is enforced at close and upgrades; it never refuses.** Dropped:
  refuse-and-reroute, and an `sdlc state` warning. Both made the flow visible to
  the agent.
- **The upgrade is written at finalize, since REWORK writes nothing (#139).**
  Stickiness across rounds comes from the boundary ledger's per-round recipe.
- **Done-when freshness compares contract hashes** recorded at change-code, not
  git history.
- **ariadne keeps its build-closure declaration**, so sdlc work in ariadne is
  always full. Rejected: narrowing it (#235 would then have qualified), and
  gates-from-main (an issue of its own if it's ever wanted).

## Lessons learned

- **Boundary review converges only on rule-level fixes.** Its "not converging:
  fix rules, not instances" was right each time. Hand-listed sets, restated
  prose, porcelain git output and accepted-but-ineffective inputs each recurred
  until they were derived, rendered, or pinned by a test.
- **Mutation-check every fix.** Two "green" tests were proving nothing: a `-run`
  regex that missed the test's name, and a stamp in the other branch.
- **Verify judge claims before acting on them.** `12e45678` reads as a string
  to both YAML readers, not a float.
- **In ariadne, never move HEAD during a running `milestone-close` review.** Run
  change-code, close and replays outside the sandbox: the judge needs the
  network.
