# doc-review through the #409 policy Implementation Plan

> **For agentic workers:** Consult AGENTS.md Section 3 (Subagent Strategy) to determine the appropriate execution approach: use superpowers-subagent-driven-development (if subagents are suitable per AGENTS.md) or superpowers-executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** One reviewer policy shared by `sdlc` and `doc-review`. doc-review then picks its second-vendor reviewer from the document's authors, and never reviews a document with its author's own family.

**Architecture:** Move #409's pure policy out of `cmd/sdlc/internal/judge` into `pkg/reviewpolicy`, so `cmd/doc-review` can import it. Go's `internal` rule is what blocks that today. `judge` keeps type aliases and re-exported names, so no sdlc caller changes, and keeps everything run-shaped. doc-review gets a thin IO shell: the document's trailers, PATH availability, and a launch-failure fallback.

**Tech Stack:** Go standard library; `git` (already used by doc-review's tests via a real repo).

---

## Decisions (approved by the hub, 2026-10-18)

- **doc-review never self-reviews.** A repo's `same` mode, or the default, counts as `other` for doc-review, because a second vendor is its whole purpose. `explicit` pairs are honored, and an explicit agent argument (`doc-review gemini file.md`) still wins. The help says so.
- **Quota detection for doc-review runs is out of scope.** Its run shape (codex `-o` report capture) differs from judge's, so a failure there keeps today's message. Only a missing CLI or a failed launch moves it to the next candidate.

## Core concepts

### Pure entities

| Name | Lives in | Status |
|------|----------|--------|
| `Agent`, `Family`, `FamilyOf`, `ChangeKind` (+ `KindDesignHeavy`, `KindBounded`), `defaultOrders` | `pkg/reviewpolicy/policy.go` | new (moved from judge `family.go`) |
| `Mode` (+ `Same`, `Other`, `Explicit`), `DefaultMode`, `Policy` (`Validate`, `Reviewer`, orders, models) | `pkg/reviewpolicy/policy.go` | new (moved from judge `agent_resolver.go`) |
| `Candidate`, `Skip`, `Availability`, `Plan` | `pkg/reviewpolicy/policy.go` | new (moved; `judge.ReviewPlan` becomes a thin wrapper) |
| `AuthorFamilies(session Family, trailerValues []string) []Family` | `pkg/reviewpolicy/policy.go` | new (the trailer → family step of `authorFamiliesFromLog`, shared) |
| `Record(used, authors, skips, forced, degraded)` | `pkg/reviewpolicy/record.go` | new (moved from `judge.ReviewerRecord`, which wraps it) |
| judge aliases (`AgentCLI = reviewpolicy.Agent`, `ReviewerPolicy = reviewpolicy.Policy`, `Family`, `Candidate`, …) and re-exported constants | `cmd/sdlc/internal/judge/policyalias.go` | new |
| doc-review's reviewer choice (`chooseReviewers`: never-same mode, explicit arg) | `cmd/doc-review/choose.go` | new |

- **Unchanged in judge** (run-shaped, not shared): `Unavailable`, `LocalAvailability` (with the containment rule), `DispatchChoice`, `UnavailableError`, `NoReviewerError`, `AgentRun.Began`.
- **ARCH-DRY:** doc-review's own `AgentCLI` and its agent constants are deleted; it uses `reviewpolicy.Agent`.

### Integration points

| Name | Lives in | Status | Wraps |
|------|----------|--------|-------|
| `reviewpolicy.Load(root)` | `pkg/reviewpolicy/load.go` | new (moved from `cmd/sdlc/reviewconfig.go`; the sdlc loader calls it) | `workshop/review.json` |
| doc-review authors (`git log --format=%(trailers:key=Agent,valueonly) -- <file>`) | `cmd/doc-review/choose.go` | new | `git` |
| doc-review availability (`exec.LookPath`, a package var for tests) and launch-failure fallback in `runReview` | `cmd/doc-review/review.go` | modified | agent CLIs (the existing `runAgent` seam) |

- ARCH-MOCK: doc-review's existing `runAgent` fake keeps serving the tests; git runs against a temp repo, as in the sdlc tests.
- ARCH-SECURE: trailers are untrusted. The same rule as #409 applies: the session's family is always among the authors, so a forged trailer can only narrow the candidates. `review.json` stays local-gates; this moves its loader only.
- ARCH-CONSTRAINTS: N/A beyond one `git log` per run (a single file's history).
- ARCH-FUNERAL: creates nothing durable beyond the existing report file, whose `reviewer:` field gains the record.
- ARCH-ORDER: holds no state between events; it's one run with an ordered fallback list.

## Plan

- [ ] M1 — Extract `pkg/reviewpolicy` (policy, plan, author families, record, loader) with its tests moved from judge's family/agent_resolver/reviewchoice tests. judge gets aliases and thin wrappers; cmd/sdlc's `reviewconfig.go` and `authorFamiliesFromLog` delegate. Verification: every existing judge and cmd/sdlc test passes unchanged, and `go vet` is clean. An atlas note says where the policy lives.
- [ ] M2 — doc-review: delete its agent list and use `reviewpolicy.Agent`. `chooseReviewers` takes the doc's trailer authors plus the session, applies never-same (same/default count as other; explicit honored; an explicit arg wins), uses the design-heavy order and PATH availability. `runReview` tries candidates in order, announcing a skip on a missing CLI or failed launch, refuses naming each candidate when none is left, and writes the record into the report's `reviewer:` field. Tests: a claude-authored doc is reviewed by codex; a codex-authored doc by claude; with codex missing, gemini runs and the record says degraded; nothing available refuses; an explicit arg wins; `same` in review.json still cross-reviews. Update the help (with the never-self-review sentence) and the atlas.

## Revisions

### 2026-10-18 — plan-quality round 1 (PQ-1 and four Minors)

- **PQ-1: the session agent and trailer parsing also cross the `internal` boundary.** `cmd/sdlc/internal/agentid`'s
  environment and log parsing move to a new `pkg/agentenv`. The environment part is `Env`, `EnvFrom`, `Known`,
  `Main`, `InSession` and the `Claude`/`Codex`/`Gemini` names; the log part is `TrailerKey`, `Commit`, `LogFormat` and
  `ParseLog`. `agentid` keeps them as aliases and wrappers, so no sdlc caller changes; its transcript detection,
  stamping and `Identity` stay put. doc-review reads the session with `agentenv.EnvFrom(os.Getenv).Main()` and the
  document's authors with `git log --format=agentenv.LogFormat -- <file>` and `agentenv.ParseLog`: one parser,
  nothing hand-rolled. **An unknown session** (no agent signals) resolves to claude, as `Env.Main` already does. That
  is conservative for never-self-review: doc-review then excludes anthropic, as it does today by defaulting to codex.
- **The forced path is extracted too (minor):** `reviewpolicy.Resolve(explicit, explicitSet, env, policy, authors,
  kind, available) Choice` holds the `--agent`/`AGENT_CMD` precedence and the `Forced` flag, and `judge.ResolveReview`
  wraps it. doc-review passes its positional agent as `explicit`, and applies never-same by mapping `same` and the
  default to `other` before calling `Resolve`.
- **Fallback rule, one only (minor):** doc-review moves to the next candidate only when the CLI is missing (a
  pre-launch skip) or `runAgent` returns a non-`*exec.ExitError` error, which is a failed launch. A codex crash is a
  run that exits non-zero (an `*exec.ExitError` under `CombinedOutput`), so it keeps today's hint message and doesn't
  switch. The Spec's "the codex-crash hint becomes the automatic fallback" is corrected to that.
- **Derived from the candidate that ran (minor):** `buildArgs`' codex `-o` temp file is made per attempt, and
  `reportPath` and the report's `reviewer:` record are computed from the candidate that produced the report. `--out`
  still overrides the path.
- **Tests, compressed (minor):** `chooseReviewers` is table-tested over (policy mode × document authors × explicit arg ×
  availability), with the property that no candidate shares an author family unless forced. One `runReview` test with
  the `runAgent` fake covers a missing CLI skipped (degraded record), a failed launch moving on, all unavailable
  refusing, and a crash not switching.
- Core concepts additions: `pkg/agentenv` (env, log parsing), new; `reviewpolicy.Resolve`/`Choice`, new.

### 2026-10-18 — M1 boundary review round 1 (BR-1..3, Minors)

**As built, each row grepped against the head** (supersedes the Core concepts table above):

| Entity | Lives in |
|---|---|
| `Agent`, `Family`, `FamilyOf`, `ChangeKind`, `defaultOrders`, `Mode`, `DefaultMode`, `Policy` (`EffectiveMode`, `Reviewer`, `Validate`), `Candidate`, `Skip`, `Availability`, `Plan`, `AuthorFamilies(session Family, commits []agentenv.Commit)` | `pkg/reviewpolicy/policy.go` |
| `Choice` (`First`, `Refusal`), `Resolve`, `ErrNoReviewer`, `NoReviewerError`, `Record(used, choice, skips)` | `pkg/reviewpolicy/choice.go` (there is no `record.go`) |
| `ConfigPath`, `Load` | `pkg/reviewpolicy/load.go` |
| `Env`, `EnvFrom`, `Known`, `Main`, `InSession`, agent names, `TrailerKey`, `Commit`, `LogFormat`, `ParseLog` | `pkg/agentenv/agentenv.go` |
| judge's aliases and wrappers: `AgentCLI` and the agent constants (`dispatch.go`); `ReviewerMode`, `ReviewerPolicy` and the mode constants (`agent_resolver.go`); `Family`, `ChangeKind`, `Candidate`, `Skip`, `Availability`, the family and kind constants, `FamilyOf`, `ReviewPlan` (`family.go`); `ReviewChoice`, `NoReviewerError`, `ResolveReview`, `ReviewerRecord`, `Refusal` (`reviewchoice.go`) | `cmd/sdlc/internal/judge/` (there is no `policyalias.go`) |

- **The move keeps judge's contract (BR-1):** `reviewpolicy.NoReviewerError` unwraps to the package's own
  `ErrNoReviewer`. Every sdlc site reports a refusal through `judge.Refusal(choice)`, which matches both
  `*NoReviewerError` and `judge.ErrReviewDidNotRun` and names `--agent` as the remedy, as it did before the move.
  `TestRefusalKeepsJudgesContract` pins this in judge.
- The shared packages carry their own tests: `pkg/agentenv` (`TestEnvMain`, `TestParseLog`) and `pkg/reviewpolicy`
  (`TestLoad`, `TestAuthorFamilies`, and the moved policy tests).
