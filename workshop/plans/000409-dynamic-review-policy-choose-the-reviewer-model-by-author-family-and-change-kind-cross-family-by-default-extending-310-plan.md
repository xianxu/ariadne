# Dynamic review policy Implementation Plan

> **For agentic workers:** Consult AGENTS.md Section 3 (Subagent Strategy) to determine the appropriate execution approach: use superpowers-subagent-driven-development (if subagents are suitable per AGENTS.md) or superpowers-executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Every review sdlc dispatches (boundary, fix, plan-quality, estimate-quality, planning) chooses its reviewer agent and model by the author's family and the change kind, and is cross-family by default. A reviewer that can't run degrades visibly, never silently to the author's family.

**Architecture:** A pure `judge.ReviewPlan` resolves (author families, change kind, policy, availability) to an ordered list of candidate reviewers. A thin dispatch loop tries them in order, skipping a candidate only for a confirmed "could not run" failure, and records what happened. One resolver serves every review site (ARCH-DRY: today plan-quality, estimate-quality, planning review and `sdlc judge` bypass the #310 policy).

**Tech Stack:** Go; the existing `workshop/review.json` (#310) and `judge` dispatch.

---

## What exists (measured on main, 2026-10-18)

- `workshop/review.json` → `loadReviewerPolicy` (`cmd/sdlc/reviewconfig.go`) → `judge.ReviewerPolicy{same|other|explicit}` (`internal/judge/agent_resolver.go`). Its default is **`same`**, and this repo has no file, so today the author's own agent reviews.
- Only `close` and `milestone-close` consult the policy (`boundaryReviewOptionsFor`). Plan-quality (`changecode.go:725`), estimate-quality (`:965`), planning review and `sdlc judge` call `ResolveAgentCLI`, which is always `same`.
- No family concept: `agentid.Known` = claude, codex, gemini. No reviewer model is ever passed, so each CLI uses its own default.
- No fallback. A review that can't run ends `VerdictNotRun` and the close halts. Failures are classified as `ErrAPIUnreachable` (a regex) or `ErrReviewDidNotRun`; a missing binary gives a PATH error. Quota and auth errors aren't detected.
- Who reviewed is recorded in the ledger round's `agent` and the sidecar's `reviewer` row. The trailers don't name the reviewer.
- `workshop/review.json` is a **local-gates path** (`local-gates-guard.yml`), so changing its contents needs the operator's label. Code under `cmd/sdlc` is deliberately not on that list.

## Decisions to confirm

1. **The default changes from `same` to `cross`** in code, for every repository (base layer). The issue title asks for it. Effects: closes in a claude-authored repo are reviewed by codex (or gemini), which spends that vendor's quota, and the operator's machine must have the other CLIs installed and signed in. claude, codex and gemini are all on PATH here. A repo that wants the old behaviour writes `"boundary_reviewer": "same"`; that edit is local-gates, so it needs the operator's label.
2. **Family = vendor:** claude → anthropic, codex → openai, gemini → google; human and unknown → none. qoder resolves through its underlying model when the trailer names one, and is otherwise unknown.
3. **No same-family fallback without an explicit choice.** When no other-family reviewer can run, the review refuses, visibly. The operator or agent can then retry later, or rerun with `--agent <same-family>`, which is recorded as `degraded: same-family by choice`.

## Core concepts

### Pure entities

| Name | Lives in | Status |
|------|----------|--------|
| `Family`, `FamilyOf(agent, model)` | `cmd/sdlc/internal/judge/family.go` | new |
| `ChangeKind` (plan, estimate, small-diff, milestone, fix, docs-only) | `cmd/sdlc/internal/judge/family.go` | new |
| `ReviewerPolicy` gains mode `cross` (the new default), per-kind overrides and per-agent `models` | `cmd/sdlc/internal/judge/agent_resolver.go` | modified |
| `ReviewPlan(authors []Family, kind, policy) []Candidate` | `cmd/sdlc/internal/judge/reviewplan.go` | new |
| `Unavailable(err) (reason string, ok bool)`: a confirmed could-not-run classification | `cmd/sdlc/internal/judge/apifailure.go` | modified |

- **ReviewPlan:** for `cross`, the candidates are every known agent whose family is absent from the window's authors, in a stable preference order (policy `order`, else openai, google, anthropic), each with its configured model. `same` returns the author's agent; `explicit` keeps #310's pairs. An empty authors set (human or untrusted trailers) returns all agents in order, noted as `author family unknown`. ARCH-PURE: no IO; table tests cover every mode × author set × kind.
- **Author families:** derived from the window's `Agent:` trailers (`agentid.ParseLog`), with the session's own `agentid.Detect` as a fallback for an unstamped window. Every family present counts, so a mixed claude+codex window is reviewed by google.

### Integration points

| Name | Lives in | Status | Wraps |
|------|----------|--------|-------|
| `dispatchReview` loop over candidates | `cmd/sdlc/internal/judge/dispatch.go` | modified | agent CLIs |
| every review site resolves through `ResolveReview` | `close`/`milestone-close`/`changecode.go`/`planningreview.go`/`judge.go` | modified | — |
| ledger round + sidecar + `Review-Reviewer:` trailer | `gatestate`, `reviewsidecar.go`, `milestoneclose.go` | modified | files, git |

- **Dispatch loop:** try candidate 1. If the failure is a confirmed could-not-run (binary missing or launch failed, `ErrAPIUnreachable`, or a quota/auth message matched by new patterns in `Unavailable`), print `[!] reviewer codex (openai) unavailable: <reason>; trying gemini (google)` and try the next. Any other failure (no verdict, a timeout mid-run, a review that ran) is **not** grounds to switch: it ends as today, `VerdictNotRun`. That keeps a review that might have partly run from quietly becoming a second opinion (ARCH-ORDER). When every candidate is unavailable, the review refuses and names each candidate with its reason; the close isn't finalized.
- **The `--agent` and `AGENT_CMD` precedence stays.** A forced agent is recorded as `forced`, and if it's the author's family, `degraded: same-family by choice`.
- **Recording (visible):** the ledger round gains `family`, `model`, `author_families` and `degraded` (the skipped candidates with reasons). The sidecar shows the same rows. The close commit gains a `Review-Reviewer: codex/gpt-6-astra (openai; author anthropic)` trailer, with `; degraded: …` when anything was skipped or forced. So `git log` shows cross-family coverage per close.
- **Model:** a per-agent `models` map in review.json, defaulting to none (the CLI's default), passed as `--model`/`-m` in `agentArgs`.

### State design table (tl-worker)

Actions: run close or review; retry; force `--agent X`; edit review.json (operator, local-gates).

| Internal state / failure mode | Visible state | Recovery action | Possible loss |
|---|---|---|---|
| first cross-family candidate runs | normal verdict; `Review-Reviewer: codex/… (openai; author anthropic)` | — | none |
| first candidate could not run (missing, unreachable, quota/auth), a later cross-family one runs | verdict, plus `[!] … unavailable … trying …`; the trailer carries `degraded: codex unavailable: <reason>` | none needed | none (the first never started a review) |
| no cross-family candidate can run | review refuses; the close isn't finalized; every candidate is listed with its reason | retry later, or `--agent <same>` | time; nothing written |
| `--agent` forced to the author's family | verdict; the trailer and ledger say `degraded: same-family by choice` | — | cross-family coverage (recorded) |
| candidate started but gave no or garbled verdict, or timed out mid-run | as today: one retry on the same agent, then `VerdictNotRun`, and the close halts; no switch to another family | retry | time |
| author family unknown (human, unstamped) | the ledger notes `author family unknown`; the first candidate in order runs | — | none |
| invalid review.json | as today: the review halts naming the file | operator fixes the file (local-gates) | none |

**Covered:** every could-not-run cause the code can confirm before a review starts. **Not covered:** a vendor that runs but quietly degrades its model (unobservable), and a quota exhausted mid-review (that's the "started" row and halts as today, by design).

## Operating envelope (ARCH-CONSTRAINTS)

Reviews are interactive batch work of minutes. The fallback adds at most one failed launch or API probe per skipped candidate: seconds, bounded by the number of known agents (3). N/A beyond that.
ARCH-FUNERAL: creates nothing new on disk. It adds fields to existing ledger rows and sidecars, both bounded by review rounds, and one trailer line per close.
ARCH-SECURE: the `Agent:` trailers are untrusted input (anyone can write a trailer). A forged trailer can only steer the reviewer toward another family, never away from review. Families are parsed against the `Known` list; unknown values mean "unknown", never a guess.
ARCH-MOCK: the dispatch loop is tested with the existing fake runners. A fake per agent returns launch-failure, unreachable or quota text, or a verdict.

## Plan

- [ ] M1 — `Family`/`FamilyOf`, `ChangeKind`, policy mode `cross` (the default) with per-kind overrides and `models`, and the pure `ReviewPlan`, with table tests (modes × author sets, including mixed and unknown, × kinds). `Unavailable` classification, with tests for the quota and auth texts each CLI prints (from their docs or a captured sample) and for launch failures.
- [ ] M2 — Dispatch loop with visible fallback and refusal. Every review site (close, milestone-close, fix review, plan-quality, estimate-quality, planning review, `sdlc judge`) goes through `ResolveReview`. Ledger/sidecar fields and the `Review-Reviewer:` trailer. Fake-runner tests: first unavailable then second runs; all unavailable refuses; started-but-garbled doesn't switch; forced same-family recorded as degraded. Help (close.md REVIEWER), atlas (#310 paragraph), and doc-review's hard-coded codex noted for a follow-up if it should derive from the policy.
