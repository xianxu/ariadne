---
id: 000367
status: working
deps: []
github_issue:
created: 2026-10-14
updated: 2026-10-14
estimate_hours:
card_mirror: '8b088f0b6decee523ca05cc6e74ac83dcd190f4a' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-14T20:40:03-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:4
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: full, provenance: inferred}
---

# Commits record the coding agent, model and effort in one standard trailer (today: 6+ inconsistent Co-Authored-By spellings, many commits none)

## Problem

The operator wants token velocity for planning: which agent/model did how much work (2026-10-14). Commits are the natural record, but over the last 3 days of origin/main: pair has 506 `Claude Opus 5.5`, 290 `GPT-6`, 14 `Codex`, 10 `GPT-6 Codex`, 9 `OpenAI Codex` trailers and 1,718 commits with none (merges, tracker and sdlc-generated commits, some agent commits); ariadne has 238 Claude, 47 `GPT-6`, 9 `GPT-6-Sol`, 8 `GPT-6.1-sol`, 7 `GPT-6.1`, 750 none. Model names are free text and effort is never recorded, so commits can't be joined to `sdlc usage` (ariadne#365) or the bench.

## Spec

One machine-readable trailer on every agent-authored commit, e.g. `Agent: codex/gpt-6.1-sol/medium` (agent/model/effort), alongside the human-facing Co-Authored-By. sdlc-generated commits (publish, tracker, close, merge) stamp it from the invoking session's agent where known, `Agent: sdlc` otherwise. Agents learn it from AGENTS.md/the commit convention; a commit-msg check (or close-gate check over the window) flags agent commits without it. `sdlc usage` (#365) and `sdlc actual` can then group by it.

## Done when

- The convention is in AGENTS.md's commit section and the sdlc commit helpers emit it; a check reports agent commits missing it in an issue's window.
- `git log --format='%(trailers:key=Agent)'` over a new issue's window shows one consistent value per commit.

## Plan

Durable plan: `workshop/plans/000367-agent-trailer-plan.md` (format coordinated with #365; pure `agentid` package, stamped commit sites + sweep test, `sdlc agent-id`, close report, AGENTS.base.md §12).

- [x] agentid pure core (Identity, Stamp, MissingAgent)
- [x] transcript readers + Current()
- [x] sdlc agent-id verb
- [x] stamp commit sites + sweep test
- [x] close report of window commits without Agent:
- [x] AGENTS.base.md §12 + atlas
- [x] verify: full suite, window trailers (recorded sdlc-local/suite pass on 9852afd2)

## Log

### 2026-10-14
- 2026-10-14: closed — recorded suite sdlc-local/suite pass on a6da3802 (exit 0); round-1 fixes: agentid is a stdlib-only leaf owning the agent env (gitx deps add only agentid), transcript roots single-sourced; tests: agentid unit (render/parse, Stamp, Claude/Codex tails, bounded read, detection incl. PAIR_AGENT-only shell), TestEveryCommitSiteStamps sweep, Agent trailer asserted on close evidence, tracker cards, issue new, change-code checkpoint, issue sync and the landing integration merge, close report wiring asserted in the close e2e; every commit in this window carries one Agent value; review verdict: FIX-THEN-SHIP
- 2026-10-14: flow upgraded quick → full — 509 added lines in code files (limit 100); an earlier round of this close already ran the full review
- Format sent to ariadne:3 (#365) and the TL: `Agent: agent/model/effort` with the transcripts' raw names. Env carries no model (only CLAUDE_EFFORT), so model and effort come from the session transcript's tail: Claude assistant records carry `message.model` + `effort`; Codex `turn_context.payload` carries `model` + `effort`.
- Tests run inside agent sessions, so TestMain in cmd/sdlc, gitx and tracker pins `SDLC_AGENT_ID` (`testfix.PinAgent`); otherwise commit messages in tests would depend on who runs the suite.
- Stamped 16 `commit`/`commit-tree` sites plus the landing integration merge; exempt (named in the sweep test): recovery ref, rebased-patch temp commit, abandon's merge conclusion. The GitHub PR merge commit is GitHub's message; merges carry no work and the close report skips them (first-parent, no merges).
- Close report is a warning, not a gate (proposed to the TL; one GateCatalog row promotes it).
- Boundary review round 1 (FIX-THEN-SHIP): BR-1 gitx reached into judge through agentid. Fixed by making agentid a stdlib-only leaf that owns the agent env (judge aliases it); gitx's deps add only agentid. Minors fixed in the same round: transcript roots single-sourced, plan Revisions, behavioural trailer tests for the checkpoint/sync/landing merge/close wiring, PAIR_AGENT-only shells read as `sdlc`, help text notes the 1 MiB tail. Left as is: `Stamp` treats a last paragraph starting `Agent: ` as already stamped (that paragraph is a trailer block by git's own rule).
- Boundary review round 2 (FIX-THEN-SHIP, no blocking): BR-1..3 addressed; BR-4 (no behavioural test for the `syncInPlace` local commit) fixed by asserting the on-main arm's local HEAD in `TestIssueSync_PublishMatrix`, checked red with the stamp removed.
