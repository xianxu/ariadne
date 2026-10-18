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

## Revisions

### 2026-10-18 — plan-quality round 1 (PQ-1..5 and minors), and the hub's decisions

The hub approved the design: one resolver for every review site, a switch only on a confirmed could-not-run, the
`Review-Reviewer:` trailer, and the order openai, google, anthropic as the overridable bounded order. doc-review stays
as is (#441). Changing the fleet default is **the operator's call**, so the default is one constant, kept at `same`
until they answer. Changes to the design above:

- **The #310 policy is the floor (PQ-3); no new `cross` mode.** `same` and `explicit` keep their meaning, and the
  dynamic layer never overrides them. `other` *becomes* the dynamic, family-aware mode: today it means "claude ↔
  codex"; now it means "a family absent from the authors, in the order for this change kind". That generalizes
  without breaking anything, since a claude-authored change under `other` still goes to codex first.
  `Validate` accepts the same three modes plus optional `order` (per kind) and `models`. The default is one constant,
  `defaultBoundaryReviewer = "same"`. The operator's answer is a one-line change to `"other"`, and it won't land
  without that answer.
- **Change kind is the spec's (PQ-2):** `design-heavy` (a new mechanism, state or invariant) or `bounded`. It comes from
  what sdlc already decides: plan-quality and planning review always review a design, so they're design-heavy. A
  boundary review takes the issue's flow, so a full flow (Mx milestones, or a design over the shell's limit) is
  design-heavy and a quick flow is bounded. Fix reviews inherit their close's kind. **The default rule uses kind:**
  design-heavy prefers anthropic, then openai, then google; bounded prefers openai, then google, then anthropic. In
  both, the author's families are removed first. The spec's evidence motivates this: Claude's reviews caught Codex
  designs over-engineering, and Codex reviews caught bugs in Claude code. Each kind's order can be overridden in
  review.json.
- **"Could not run" is evidence that no review began (PQ-1).** `Unavailable(run AgentRun, stderr string, launchErr
  error) (reason string, ok bool)` returns true only if:
  - (a) the process failed to launch (a missing binary or an exec error), or
  - (b) it exited non-zero **and its parsed stream holds no assistant or tool event**, **and** its output or stderr
    matches the unreachable, quota or auth patterns.
  A run whose stream can't be parsed counts as started (conservative). `RunFailure` is unchanged and still decides
  `VerdictNotRun` for everything else. The loop calls `Unavailable` first, and switches only on its yes.
  `Dispatch`'s swallowed non-zero exit means quota and auth text arrive in the output, which is why `Unavailable`
  reads the run rather than an `err`.
- **Pre-launch skips (PQ-4):** a candidate known not to work here is skipped before launch, with the reason recorded:
  codex under `ContainCheckout` on darwin (its Seatbelt sandbox fails where the fence probe did, as
  `milestoneclose.go:699` already warns), or a CLI absent from PATH. `ReviewPlan` takes an `available(agent)
  (reason, ok)` predicate, and the shell supplies it.
- **Author families are untrusted (minor):** a forged `Agent:` trailer could make the window look like another
  family's work. Author families are therefore the union of the window's trailers **and** the closing session's own
  `agentid.Detect` family. A forged trailer can add a family (narrowing the candidates) but can't remove the closer's
  family, so it can't steer the review back to the closer's own family.
- **Non-goals (PQ-5):**
  - Budget or quota *accounting*: only reactive detection of a quota error is built, and there's no "over budget"
    prediction, because reviewusage records usage but holds no budget.
  - doc-review's hard-coded codex (#441).
  - Semantic kind detection from the diff: kind comes from the flow decision.
  - Detecting a vendor that quietly degrades its model.
- **Tests (minor):** one strategy per risky function: table tests for `ReviewPlan` (mode × author set × kind ×
  availability) and `Unavailable` (captured outputs per agent: launch failure, unreachable, quota, auth, started then
  failed); one fake-runner test of the loop's three outcomes (skip then run, all unavailable refuses, started never
  switches).
- **Done-when's weekly tally (minor):** `Review-Reviewer:` is a git trailer in a fixed form, `<agent>/<model>
  (<reviewer family>; author <families>)[; degraded: …]`. A test parses it back with `git log
  --format=%(trailers:key=Review-Reviewer)` and a tally helper, `sdlc review-tally` style, counts by author and reviewer
  family. Help documents the git one-liner.

### 2026-10-18 — M1 boundary review round 1 (BR-1..5, two Minors)

Core concepts, corrected to the code:
- `Family`, `FamilyOf(agent)` (by agent only; qoder and anything unknown map to "", never guessed through a model),
  `ChangeKind`, `defaultOrders`, `Candidate`, `Skip`, `Availability`, `ReviewPlan(policy, main, authors, kind,
  available)`, `Unavailable(agent, run, stderr, nonZeroExit, launchErr)` and `LocalAvailability(containment, goos,
  lookPath)` all live in `cmd/sdlc/internal/judge/family.go`. `AgentRun.Began` is in `stream.go`.
  `DispatchOptions.Model` and the per-CLI model flags are in `dispatch.go`. There is no `reviewplan.go`.
- **The bounded default order is codex, claude, gemini.** This keeps #310's claude ↔ codex pairing under `other`, as
  the floor requires. With gemini second, a codex-authored bounded change would go to gemini. Gemini is the fallback
  in both kinds. The hub had approved openai, google, anthropic, so this deviation is reported to them.
- **Unavailability evidence per CLI (BR-2):**
  - claude: a stream with no assistant or tool event.
  - codex and gemini: non-stream, a non-zero exit, and nothing at all on stdout, where their review text would be.
    For codex, its stderr must also show no activity block (`codex`/`exec`/`thinking` lines), and the cause is read
    only from its `ERROR:` lines, because codex echoes the prompt to stderr.
  The patterns gain HTTP 403, 5xx and the CLI refusing (`IneligibleTierError`, `no longer supported`). A 400 (bad
  model) is a misconfiguration and halts. The samples were measured on this machine: codex with a bad model, and
  gemini, which refuses entirely here (`IneligibleTierError: This client is no longer supported…`). So on this
  machine gemini can't be a fallback; that's a fact for the operator.
- **No silent same-family review (BR-1):** `Validate` requires every `reviewer_order` to name agents of at least two
  families, with no repeats. `Reviewer()` under `other` falls back to the default orders, and never returns the main
  agent's family. Tested by `TestReviewerOtherNeverSameFamily`.
- **`reviewer_models` now reaches the CLI (BR-4):** `DispatchOptions.Model` is passed as claude `--model`, codex
  `-m` or gemini `-m`. The boundary review sets it from the policy for the resolved reviewer, and M2's resolver sets
  it at every site. A model value starting with `-` is refused (ARCH-SECURE).
- The help (close.md REVIEWER) and the atlas describe the new `other`, `reviewer_order` and `reviewer_models` (BR-3).

### 2026-10-18 — M2 boundary review round 1 (REWORK: BR-9..13 and Minors)

**As built, each row checked by grep against the head** (BR-10; this table supersedes the Core concepts and
integration tables above):

| Entity | Lives in |
|---|---|
| `Family`, `FamilyOf`, `ChangeKind`, `defaultOrders`, `Candidate`, `Skip`, `Availability`, `ReviewPlan`, `Unavailable`, `LocalAvailability` | `cmd/sdlc/internal/judge/family.go` |
| `ResolveReview`, `ReviewChoice` (`First() (Candidate, bool)`, `Refusal()`), `DispatchChoice`, `UnavailableError` (short `Reason`, operator-only `Detail`), `NoReviewerError`, `ReviewerRecord` | `cmd/sdlc/internal/judge/reviewchoice.go` |
| `ReviewerPolicy.Orders`/`Models`, `DefaultReviewerMode` | `cmd/sdlc/internal/judge/agent_resolver.go` |
| `AgentRun.Began` | `cmd/sdlc/internal/judge/stream.go` |
| `DispatchOptions.Model`, per-CLI model flags, `UnavailableError` from `classifyRunResult`, the attempt-2 demotion | `cmd/sdlc/internal/judge/dispatch.go` |
| `resolveReviewChoice`, `reviewAuthorFamilies`, `branchWindowBase`, `boundaryChangeKind`, `reviewAvailability` (test seam) | `cmd/sdlc/reviewchoice.go` |
| `planningReviewChoice` | `cmd/sdlc/changecode.go` |
| `judgeChangeKind` | `cmd/sdlc/judge.go` |
| ledger `Round.Reviewer` (one string, the trailer's text; not separate family/model/author/degraded fields) | `cmd/sdlc/internal/gatestate/ledger.go` |
| `Review-Reviewer:` trailer | `cmd/sdlc/milestoneclose.go` (`reviewTrailers`) |

Retired: `ResolveAgentCLI` and `ResolveReviewerCLI` are deleted, and their tests now run through `ResolveReview`. The
parse and tally helpers are test-only (the tally is the git one-liner in `close --help`).

Decisions in this round:
- **A reviewer that began never hands over (BR-9, Critical):** an `UnavailableError` from attempt 2 means attempt 1
  ran, so `dispatchAttempts` demotes it to `ErrReviewDidNotRun`. `TestUnavailableRetryNeverHandsOver` fails without
  this; checked by revert.
- **An empty choice is never dispatched (BR-11):** `First()` reports emptiness, and every site refuses with
  `Refusal()`: the boundary review as not-run, plan-quality and estimate-quality as an error, and `sdlc judge` by
  dying.
- **Every site is pinned (BR-12):** `TestPlanQualityChoosesThroughThePolicy`, `TestEstimateQualityChoosesThroughThePolicy`
  and `TestSdlcJudgeChoosesThroughThePolicy` each run in an `other` repo with a claude session and expect codex.
- **Docs (BR-13):** the atlas no longer names `ResolveAgentCLI` as live, and says the ledger round's `reviewer`
  carries the record. The rule followed: every retired symbol is grepped across atlas/ and helptext/ in the same
  commit.
- **Fix reviews are bounded, a deliberate deviation from "inherit their close's kind":** a fix review judges only
  the fix's small interdiff, whatever the close's flow was, so `boundaryChangeKind` maps FixReview to bounded.
- **A launch-failure reason leads with the cause** (`could not launch: <exec error>`). The owner bin and PATH stay in
  the error's `Detail` for the operator, never in the trailer or ledger.
