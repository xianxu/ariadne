---
type: pensive
date: 2026-06-02
topic: sdlc / workflow-harness retrospective from the shared-brain close + nous#41 session
mode: thoughts
description: Battle-tested feedback on the sdlc workflow harness after a long session that exercised the full lifecycle (claim → change-code → milestone-close ×6 → close ×7 → pr → merge). What worked, what cost the most friction, and the highest-leverage fixes — with the specific session moment behind each.
references: [AGENTS.md, workshop/issues/000064-sdlc-judge-reads-base.md]
---

# Pensive: sdlc / workflow-harness retrospective

Captured by the agent at the operator's request after a session that touched almost
every part of the `sdlc` harness: closing down the `shared-brain` project (declare-done
+ fold-in wave-2 + close/archive 11 issues), filing `ariadne#64`, then implementing
`nous#41` (collaborator-lifecycle hardening) end-to-end across four reviewed milestones
and merging via PR. Each item below names the concrete moment that produced it, so it's
falsifiable rather than vibes. Comment inline — I treat battle-tested feedback with
urgency.

---

## What worked

**W1 — Gate-not-state-machine is the right spine.** `sdlc` codifies the *transitions*
where drift recurs and leaves the stages as prose. `change-code` refused until `#41` had
a `## Spec`; `milestone-close` refused without an atlas touch + actuals + a
`Review-Verdict` trailer. When the errors are good they read as *next-action specs*
("no upstream → run `sdlc pr` first"). This is the best property of the whole system —
keep it.

**W2 — Mandatory fresh-eyes review at every boundary earned its keep (not theater).**
Three milestones, three genuine finds:
- M1 review flagged fp→login resolution preferring a stale `verified.yaml` hint → fixed in M3.
- M3 review caught a `reset --hard` data-loss hazard (Critical) → became the dirty-tree guard.
- M4 review caught a factually-wrong doc claim (`AcceptInvitation` no-op) → corrected.
The "always subagent the review, main session carries confirmation bias" rule held up.

**W3 — The `target` datatype was the standout.** `collaborator-state-machine.md` is what
let me *find* the `nous#32` leave-resurrection bug — I reasoned from invariant #2 ("clears
every store") and went looking. A plain issue tracker would not have produced that.
Invariants-defended-from-drift is a real idea, not bureaucracy.

**W4 — Subagent exploration kept main context clean.** "Map the code touch-points →
digest" before planning, and the per-milestone reviewers, meant I planned/reviewed against
tight summaries instead of dumping 1000-line files into my own context.

---

## What can be improved

Numbered for easy reference in comments.

### F1 — Judge *result handling* is brittle, and the only escape hatch is too blunt. (Biggest friction.)
- On `sdlc push`, the plan-completeness judge reads `origin/main` (the base), not HEAD — so
  a *close-and-archive* push structurally can never pass (the issues are still `working` on
  main by definition). Hit twice. → `ariadne#64`.
- On `sdlc merge`, the judge correctly read the full diff and returned **VERDICT: CLEAN**
  ("no action required to ship") — but `sdlc`'s result parser saw the judge's
  *explicitly-non-blocking note* and classified the whole run as `failure`.  🤖[we should schematize the protocol a bit. give me some options]
- Both forced `--no-judge`, which disables **all** judges (specs, lessons, plan). Routing
  around one parser bug by switching off three valuable reviewers is the wrong granularity. 🤖[in another session, I asked AI to add --no-xxx, where xxx is a critieria in a gate.]
- **The judges are good; the *result interpretation* is the weak link.** Fixes: read HEAD
  (or the merge result), parse the leading `VERDICT:` line as authoritative (a CLEAN with a
  non-blocking note must pass), and add **per-judge** skip flags so I never disable all to
  dodge one. Fold the merge false-positive into `ariadne#64` or a sibling.

### F2 — Velocity calibration is collecting fabricated data.
Every close warned me to run `active-time-v3.py`, but the v3 telemetry returns **0 events**
for these sessions (flagged as far back as `nous#34`). So every `actual_hours` this session —
across ~7 closes — was a manual guess passed with `--force`. The whole estimate-calibration
loop the `project` datatype is built around is running on made-up numbers. It has the *form*
of calibration with none of the substance. Either telemetry capture is broken for this
harness, or the actuals signal needs to be something that actually exists here. Until then,
the `--actual` gate is ceremony. 🤖[make a ticket, we need to fix active-time-v3.py. one potentially issue is project shared-brain had been "dormant for a couple of weeks as I took detour fixing many other things and as a result, there might be some parameters to the script, such as how far we look back, that might be not working."]

### F3 — `milestone-close` over-couples to the project file.
Every `nous#41` milestone-close tried to write a per-milestone detail block into the
*closed* `shared-brain` project — only because `#41` appears there as a follow-on row —
forcing `--force` four times. The logic treats "any issue referenced in a project file" as
"a project milestone needing calibration prose." It should distinguish `mvp_scope`
milestones from issues merely *mentioned* as follow-ons (e.g. key off `mvp_scope` membership
or the issue's `target:` instead of a bare text reference). 🤖[I see, #41 was later added to project, not originally planned. but as we test, I decided we need to harden that poriton of system enough. you asked me if we should treat #41 part of project, and I said yes I think. if project is closed, is the remedy to either reopen the project; or you push back to not add #41 to a closed project?]

### F4 — Two independent fresh-context reviews per milestone, and the second is slow.
I ran my own `superpowers-requesting-code-review` subagent *and* `milestone-close`
auto-dispatched `sdlc judge milestone-review` — two separate agents reviewing the same diff.
That's the deliberate "form first, judge second" design, but the milestone-review judges took
3–10 min each (M3's worst), serialized the workflow, and mostly *confirmed* my review rather
than adding new findings. For a 4-milestone issue that's a lot of redundant latency. Worth
asking: should the boundary review and the judge be the *same* pass? Or should the judge only
deep-dive when the boundary review flags something / on a sampling basis? 🤖[I think those two review should be merged togehter. the truth is superpower is borrowed from external and milestone-close is more home grown. make a ticket to see how to merge those two. here, the sdlc milestone-close intend to provide a form, essenence I think in ariadne design, should be the adapted superpowers-requesting-code-review]

### F5 — §5 mandates process-level fakes, but `gh` has none — so `nous#41 #11` shipped with zero automated coverage.
`gh` shells out to the CLI with no injectable seam, so the re-invite hard-error fix is only
"dogfood-verified." Meanwhile `gpg-agent` won't run in the sandbox (every integration test
needed `--dangerouslyDisableSandbox`), and `LeaveBrain` refuses on `file://` so its full flow
can't be e2e'd without GitHub. The test story fractured into three tiers (pure /
gpg-unsandboxed / GitHub-dogfood) with no tooling support for the layering. A documented `gh`
fake would directly close a real coverage gap and make the §5 mandate enforceable instead of
aspirational.

### F6 — The plan-mode file is ephemeral and disconnected from `workshop/plans/`.
`EnterPlanMode` wrote my plan to `~/.claude/plans/…` (not version-controlled). The M2 judge
even recommended adding a `## Revisions` entry there — to a file that won't survive. The
durable plan record had to live in the issue's Spec/Log instead. For multi-milestone work the
constitution wants `workshop/plans/NNNNNN-…`; the harness's plan file isn't that, and nothing
bridges them. Either teach plan-mode to land in `workshop/plans/`, or have `change-code`
ingest the plan-mode file into one.

### F7 — `milestone-close` mutates the issue file but doesn't commit, and the judge runs async.
Result: uncommitted issue-file edits straddling milestone boundaries, and hand-managing which
checkbox tick belonged to which close commit. The tool-edits-but-doesn't-commit pattern plus
the background judge created a coordination tax on every milestone. Consider: milestone-close
stages+commits its own issue-file mutation (with the trailer), or emits the trailer
synchronously so there's no async gap to manage.

---

## Highest-leverage fixes (ranked)

1. **Fix judge result handling** (F1: `ariadne#64` base-read + the CLEAN-with-note
   false-positive) and add **per-judge** skip flags. Unblocks the normal close/merge flow;
   stops training me to reach for `--no-judge`.
2. **Make `active-time-v3` actually capture, or change the actuals signal** (F2). Otherwise
   stop pretending to calibrate — the fabricated data is worse than no gate.
3. **Decouple `milestone-close` project-sync from follow-on mentions** (F3).
4. **Ship a `gh` process-level fake** (F5) so GitHub-layer code stops being dogfood-only.

F1 and F2 cost the most this session. `ariadne#64` already exists for the F1 base-read half.

---

## Meta-note on my own process

Not tooling, but honest: I burned real time deliberating over the `nous#41` leave test
(F5's GitHub-coupling is a genuine constraint, but I circled it longer than I should have),
and I ran milestone-close judges in the background which created idle waits I then had to
manage. If the tooling fixes above land, several of these self-inflicted detours disappear
too.
