# Boundary Review — ariadne#367 (whole-issue close)

| field | value |
|-------|-------|
| issue | 367 — Commits record the coding agent, model and effort in one standard trailer (today: 6+ inconsistent Co-Authored-By spellings, many commits none) |
| repo | ariadne |
| issue file | workshop/issues/000367-commits-record-the-coding-agent-model-and-effort-in-one-standard-trailer-today-6-inconsistent-co-authored-by-spellings-many-commits-none.md |
| boundary | whole-issue close |
| milestone | — |
| window | 18915f6c17d9b4aeb5e618d3f736705ec355d2b3..a1ca04c516becee90f7d9f9f38104baee5d6cdea |
| command | sdlc close --issue 367 |
| reviewer | claude |
| isolation | fenced |
| timestamp | 2026-10-14T22:47:12-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

The build passes and every commit in the window carries the trailer. Next I'm checking for duplicated logic in judge/sessions, other places that create commits, and test hermeticity.

I've finished reading the diff and checking the dependency graph. Now I'm running the targeted tests.

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

The issue's purpose is delivered. The trailer format is defined in one place. All 16 `commit`/`commit-tree` sites that sdlc runs, plus the landing merge, are stamped. A sweep test refuses any new unstamped site, `sdlc agent-id` prints the trailer, and `sdlc close` warns about window commits without one. The convention is in AGENTS.base.md §12 and the atlas. The build passes, and the targeted tests pass in agentid, cmd/sdlc (the sweep, verb, close-report, close-evidence and issue-new tests), tracker and gitx. `git log --first-parent --format='%(trailers:key=Agent)'` over this window shows `claude/claude-opus-5-5/medium` on all six commits. I also checked the record shapes against real transcripts on this machine: Codex `turn_context.payload.{model,effort}` and Claude `message.model` plus top-level `effort` both match. One thing blocks a clean SHIP. The plan said gitx would receive the identity as a parameter. Instead, gitx calls `agentid.StampCurrent`, which reads the environment and transcript files itself. That makes the low-level git package depend on `judge`, `reviewbox`, `reviewprepare` and `estimate`. It also explains why agentid can't reuse the `transcripts` package's root helpers: importing it would create a cycle. Everything else is minor, mostly plan text that no longer matches the code.

**1. Strengths**
- `cmd/sdlc/agentstamp_test.go:145`: the sweep test enforces the "every consumer derives" rule. Each exemption is named, with a reason (ARCH-PURPOSE pass).
- `cmd/sdlc/internal/agentid/agentid.go`: rendering, parsing, stamping, log parsing and `MissingAgent` are all pure, and tests cover them without IO. That includes the round trip, rejection of malformed values, joining an existing trailer block, and leaving an already-stamped message alone.
- `cmd/sdlc/internal/agentid/detect.go:640` and `:685`: the `Detector{Home, Getenv}` seam keeps detection testable. The read is bounded, with a test pinning the bound (`TestTailIsBounded`), and a missing or garbled transcript degrades visibly to `-` rather than a guess.
- `testfix.PinAgent` in TestMain keeps commit messages in tests independent of whoever runs the suite. `TestTestfixPinsTheOverride` pins the duplicated constant in testfix to agentid's.
- `closeagent_test.go`: the close-report test uses a real repo containing a stamped commit, a bare commit, a commit merged in from main and a merge commit. It checks both inclusion and exclusion.

**2. Critical findings**
None.

**3. Important findings**
- **The gitx package now detects the agent itself, against the plan and the layering (ARCH-PURE)**
  - Where: `cmd/sdlc/internal/gitx/commitpublication.go:217` and `cmd/sdlc/internal/gitx/refbootstrap.go:178`.
  - The plan says: "gitx gets the identity as a parameter (no env reads inside gitx) so its tests stay hermetic".
  - What the code does: gitx calls `agentid.StampCurrent`, which reads the process environment and `~/.claude` / `~/.codex` files.
  - Effect on dependencies: `go list -deps ./cmd/sdlc/internal/gitx` now includes `judge`, `reviewbox`, `reviewprepare`, `estimate` and `agentid`. At base it was only `processgroup` and `workspace`. As a result, judge can never import gitx, and agentid can't reuse `transcripts`.
  - Fix sketch, either of:
    - Give `TrunkFile` a `StampMessage func(string) string` (or an `agentid.Identity`) that the caller in trackerenv sets.
    - Split agentid so gitx imports only a pure core with no `judge` dependency, with `Detect`/`Current` in a separate package.

**4. Minor findings**
- **ARCH-DRY:** `Detector.Detect` (`detect.go:657-663`) rebuilds `judge.AgentDefaultEnv` field by field, duplicating `judge.CurrentAgentDefaultEnv`.
  - Fix: add a `judge.AgentDefaultEnvFrom(getenv)` and have both callers use it.
  - The `~/.claude/projects` and `~/.codex/sessions` roots also duplicate `transcripts/defaults.go`. That duplication is forced by the gitx cycle above.
- **Plan text no longer matches the code:**
  - The plan says the tail read is 256 KiB; the code reads 1 MiB (`detect.go:638`).
  - The plan lists the `gh pr merge` body as stamped; it is not (the issue log explains it is GitHub's message).
  - Fixtures are inline constants, not `testdata/`.
  - The plan's task boxes are unticked; the issue's are ticked.
- **Mislabel risk in plain shells:** if `PAIR_AGENT` is set (say, in a pair-slot shell where the operator types sdlc commands by hand), `Detect` labels the commit `claude/-/-`. The reason is that `MainAgent` falls back to `claude` instead of `sdlc`. Unverified whether slot shells actually export `PAIR_AGENT`.
- **Truncation:** if the last assistant record is pushed out of the tail by more than 1 MiB of later records (large tool results), the model degrades to `-`. That is visible, but worth noting in the help text.
- **Stamp false positive:** `Stamp` treats a body paragraph starting `Agent: …` in the last paragraph as already stamped. Negligible.

**5. Test coverage notes**
- Plan Task 4 promised trailer assertions on the `claim` sync, the `changecode` checkpoint and the landing merge message. Only close evidence, `issue new` and tracker cards were extended. The sweep test covers stamping at the source-line level, so this is minor.
- The close report is tested through `reportMissingAgent` directly. Its wiring in `computeClose` (`close.go:663`) isn't exercised.
- Test packages that reach agentid but don't pin it (fleet, freshness, project, activetime, transcripts) don't assert commit messages, so there is no flake risk today.

**6. Architectural notes**

| Marker | Result | Note |
|---|---|---|
| ARCH-DRY | flag (Minor) | See the Minor findings. |
| ARCH-PURE | flag (Important) | gitx detects the agent itself. The pure core is otherwise well separated. |
| ARCH-PURPOSE | pass | Every commit site is stamped or exempt with a reason, and the sweep test enforces it. The close check is a warning, which the Spec allows ("flags"). |
| ARCH-MOCK | pass | Temp-home transcript fixtures behind `Detector`; real git in temp repos. |
| ARCH-CONSTRAINTS | pass | One bounded read per process, memoised with `sync.OnceValue`. |
| ARCH-SECURE | pass | Malformed JSON lines are skipped and missing values become `-`. The session id comes from the same user's environment; no credentials are involved. |
| ARCH-ORDER | pass | The only state is a per-process memo of a read-only lookup. |
| ARCH-FUNERAL | pass | The only durable output is one trailer line per commit. |

If #365 consumes `Parse`, it benefits from agentid's pure core being split from detection. Doing that split would also fix the Important finding.

**7. Plan revision recommendations**
- `## Revisions`: gitx stamps through `agentid.StampCurrent`. Either keep this and record that gitx now depends on judge, or restore "identity passed in" once the Important finding is fixed.
- `## Revisions`: the tail bound is 1 MiB (`tailBytes = 1 << 20`), not 256 KiB.
- `## Revisions`: the `gh pr merge` commit and abandon's merge-conclusion commit are exempt (merges carry no work); the commit-site list should say so.
- `## Revisions`: detection builds `AgentDefaultEnv` from the injected `Getenv` rather than calling `judge.CurrentAgentDefaultEnv()`. Fixtures are inline in `detect_test.go`, not under `testdata/`.

```findings
findings:
  - id: new
    severity: Important
    kind: defect
    family: io-seam-injected-not-reached-for
    title: |
      gitx calls agentid.StampCurrent (env + transcript reads) instead of taking the identity as a parameter, so gitx now depends on judge/reviewbox/estimate
    detail: |
      commitpublication.go:217 and refbootstrap.go:178 call agentid.StampCurrent, against the plan's "gitx gets the identity as a parameter (no env reads inside gitx)". go list -deps for gitx went from processgroup+workspace to also judge, reviewbox, reviewprepare, estimate, agentid, which also stops agentid reusing transcripts' roots (cycle). Fix: inject a StampMessage func or Identity on TrunkFile from trackerenv, or split agentid's pure core from detection so gitx imports no judge (ARCH-PURE).
  - id: new
    severity: Minor
    kind: defect
    family: shared-helper-not-reused
    title: |
      Detector.Detect rebuilds judge.AgentDefaultEnv and hardcodes transcript roots already defined in judge and transcripts
    detail: |
      detect.go:657-663 duplicates judge.CurrentAgentDefaultEnv field by field; add judge.AgentDefaultEnvFrom(getenv) and use it from both. The ~/.claude/projects and ~/.codex/sessions roots duplicate transcripts/defaults.go (ARCH-DRY).
  - id: new
    severity: Minor
    kind: defect
    family: plan-claims-drift-from-code
    title: |
      Plan text drifts from the code: 256 KiB vs 1 MiB tail, gh pr merge body listed as stamped, testdata fixtures, gitx parameter, unticked tasks
    detail: |
      Add a Revisions section to workshop/plans/000367-agent-trailer-plan.md recording what the code actually does.
  - id: new
    severity: Minor
    kind: defect
    family: plan-promised-test-missing
    title: |
      Plan Task 4 promised trailer assertions for claim sync, changecode checkpoint and the landing merge message; none were added
    detail: |
      The sweep test covers stamping at the source-line level, but there is no behavioural assertion for these three sites, and none for the close.go:663 wiring of reportMissingAgent.
```

---

## Re-review — 2026-10-14T23:39:58-07:00 (FIX-THEN-SHIP)

| field | value |
|-------|-------|
| issue | 367 — Commits record the coding agent, model and effort in one standard trailer (today: 6+ inconsistent Co-Authored-By spellings, many commits none) |
| repo | ariadne |
| issue file | workshop/issues/000367-commits-record-the-coding-agent-model-and-effort-in-one-standard-trailer-today-6-inconsistent-co-authored-by-spellings-many-commits-none.md |
| boundary | whole-issue close |
| milestone | — |
| window | 18915f6c17d9b4aeb5e618d3f736705ec355d2b3..a6da38026389acf2358c0132581d61e9cbf6bd63 |
| command | sdlc close --issue 367 |
| reviewer | claude |
| isolation | fenced |
| timestamp | 2026-10-14T23:39:58-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

Now the fix commit itself.

Prior-round fixes check out on reading. Next I'm running the targeted tests, then removing individual stamps to see whether the new assertions fail without them.

Now revert checks for the claim sync and changecode stamps.

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

This is ready to ship. The round-1 blocker (BR-1) is fixed: `agentid` now uses only the standard library and owns the agent environment, and `judge.AgentDefaultEnv` is just another name for `agentid.Env`. `go list -deps ./cmd/sdlc/internal/gitx` now adds only `agentid`; judge, reviewbox, reviewprepare and estimate are gone. The plan's new Revisions section records why gitx calls `StampCurrent` instead of taking the identity as a parameter. The build is clean and the targeted tests pass. Every commit in this issue's window has exactly one `Agent:` value, so the Done-when holds. One Minor item is left: the claim-sync stamp has no behavioural test. When I removed it, every test still passed, because the trunk publication path adds the stamp again on the published commit. That does not block the boundary.

**1. Strengths**
- **Done-when holds on the issue's own window.** `git log --format='%(trailers:key=Agent)'` shows `claude/claude-opus-5-5/medium` on all 7 commits.
- **The `agentid` package is a real leaf** (`cmd/sdlc/internal/agentid/env.go`).
  - `EnvFrom(getenv)` is now the one parser of the agent environment, used by both the judge and the detector.
  - `ClaudeProjects`/`CodexSessions` are the only production statements of the transcript roots. Grep finds no other production literals; only tests restate them.
- **Untrusted transcript input is handled carefully** (`transcript.go`). The read is capped at 1 MiB. A truncated first line or an unknown record is skipped, and a miss gives empty parts rather than a guess. `TestTailIsBounded` checks both the cap and that the newest record lands inside it.
- **The sweep test makes the purpose enforceable.** `TestEveryCommitSiteStamps` fails on any new commit site that skips the stamp, and each exemption carries a reason (`cmd/sdlc/agentstamp_test.go:17-27`).
- **The new assertions catch real regressions.** With the change-code stamp removed, `TestChangeCodeRefreshesCardFieldsAndCheckpointsLocally` fails as it should (`checkpoint Agent trailer ""`). The close test now checks that the report is actually printed, which covers the `close.go:666` wiring.

**2. Critical findings:** none.

**3. Important findings:** none.

**4. Minor findings**
- **BR-4 is still partly open** (disposed `not-addressed` below). The commit stamp in `syncInPlace` (`claim.go:597`) has no behavioural test. I removed it and ran `-run 'Sync|Claim|IssueNew|Publish'`: nothing related failed. `TestIssueSync_PublishMatrix` checks the commit on `origin/main`, and `TrunkFile.commitTree` stamps that commit again. The local commit made with `--no-push` is never checked.
  - Fix: add the same `%(trailers:key=Agent,valueonly)` check to a `--no-push` sync test, or to the local HEAD before publication.
  - The source-level sweep test still guards this site, so this stays Minor.
- **README:** the new `agent-id` verb appears in the help text, AGENTS.base.md and the atlas, but not in README.md. README has no per-verb catalog to add it to, so I am not raising this.

**5. Test coverage notes**
- Passing: the `agentid`, `judge`, `transcripts` and `processmanual` packages, plus the targeted `cmd/sdlc` tests (sweep, checkpoint, issue sync, landing integration, tracker close).
- Two tests fail in this review checkout whether or not anything is reverted: `TestLegacyRepoLockConcurrentIssueNewSerializesAllocation` and `TestPushPublishSequence_CodecompleteFlippedThenArchived`. Both stop with "no issue tracker is reachable", meaning the checkout has no issue-tracker branch. This is an environment problem, not part of this diff.
- Revert checks:
  - Change-code stamp: the test fails without it, as it should.
  - Claim-sync stamp: no test fails without it (see BR-4).
- The checkout was restored afterwards; `git status` is clean.

**6. Architectural notes**
- **ARCH-DRY: pass.** The agent environment and transcript roots each have one production source; `knownAgent` and `MainAgent` delegate to `agentid`.
- **ARCH-PURE: pass, with a recorded exception.** gitx calls `StampCurrent`, which reads the environment and a file once per process via `sync.OnceValue`. The plan's Revisions section justifies this: with 16 constructors, any one of them could forget an injected identity. Test hermeticity comes from `PinAgent` in each `TestMain` that asserts commit messages.
- **ARCH-PURPOSE: pass.**
  - Commit sites are enforced by the sweep test, and the close report covers commits agents make by hand.
  - The report is a warning, not a gate. This is deliberate, recorded in the plan, and promoting it later is one catalog row.
- **ARCH-MOCK: pass.** Tests use real git in temp repos, transcripts come from inline fixtures in a temp home, and the environment comes through an injected `Getenv`.
- **ARCH-CONSTRAINTS: pass.** One seek plus a bounded read of at most 1 MiB, once per process.
- **ARCH-SECURE: pass.** Only the session's own transcript is read. A malformed `SDLC_AGENT_ID` is ignored rather than trusted (covered by a test case). No credentials are involved.
- **ARCH-ORDER: pass.** The only state is a per-process memo of a read-only lookup.
- **ARCH-FUNERAL: pass.** The only durable output is one trailer line per commit.
- **For #365:** `Parse` and `ParseLog` are the stable surface to group usage by agent.

**7. Plan revision recommendations**
- One correction to the Revisions bullet "Behavioural trailer assertions now cover … issue sync". The issue-sync assertion checks the published commit, which `gitx.commitTree` stamps. It does not cover `syncInPlace`'s own stamp. Either reword the bullet or add the local-commit assertion (BR-4).

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      agentid is now standard-library only, and go list -deps for gitx adds only agentid (judge, reviewbox, reviewprepare and estimate are gone). The parameter-vs-StampCurrent decision is recorded with its reasons in the plan Revisions, and tests are hermetic through testfix.PinAgent in the gitx, tracker and cmd/sdlc TestMains.
  - id: BR-2
    disposition: addressed
    note: |
      agentid.EnvFrom is the single agent-env parser (judge.CurrentAgentDefaultEnv delegates to it, env.go:24). ClaudeProjects and CodexSessions are used by transcripts/defaults.go, processmanual.go, processmanual/session.go and processmanual/memory.go, and grep finds no remaining production literals.
  - id: BR-3
    disposition: addressed
    note: |
      The plan Revisions now record the 1 MiB tail, the gh merge exemption, the inline fixtures, gitx using StampCurrent, and PAIR_AGENT-only shells reading as sdlc. All tasks are ticked, and each claim matches detect.go, env.go and agentstamp_test.go.
  - id: BR-4
    disposition: not-addressed
    note: |
      The change-code checkpoint (verified red when its stamp is reverted), landing merge and close-report wiring now have assertions. The claim-sync stamp at claim.go:597 does not: with it removed, -run 'Sync|Claim|IssueNew|Publish' showed no related failures, because the issue-sync test checks origin/main, which gitx.commitTree stamps again.
```
