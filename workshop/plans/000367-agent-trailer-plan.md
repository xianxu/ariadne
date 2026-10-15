# Agent commit trailer Implementation Plan

> **For agentic workers:** Consult AGENTS.md Section 3 (Subagent Strategy) to determine the appropriate execution approach: use superpowers-subagent-driven-development (if subagents are suitable per AGENTS.md) or superpowers-executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Every agent-authored and sdlc-made commit carries one machine-readable `Agent: <agent>/<model>/<effort>` trailer, and `sdlc close` reports window commits that lack it.

**Architecture:** A new pure package `cmd/sdlc/internal/agentid` owns the trailer value (`Identity`, `String`, `Parse`, `Stamp`) and reads the session's model and effort from its own transcript tail (the thin IO seam). sdlc's commit sites stamp their messages through it. A `sdlc agent-id` verb prints the trailer for agents to put on their own commits. The close verb lists window commits without one.

**Tech Stack:** Go, git trailers (`git interpret-trailers` semantics, `%(trailers:key=Agent)`), cobra.

---

## Format (coordinated with ariadne#365)

`Agent: <agent>/<model>/<effort>`, one per commit, values as the transcripts record them so commits join to `sdlc usage` without a mapping table:

| Part | Source | Examples |
|---|---|---|
| agent | `judge.MainAgent(env)` (existing; `PAIR_AGENT`, `CODEX_*`, `CLAUDECODE`) | `claude`, `codex` |
| model | Claude: `"model"` on the session transcript's last assistant record; Codex: `turn_context.payload.model` | `claude-opus-5-5`, `gpt-6-astra` |
| effort | Claude: `"effort"` on the same record, else `CLAUDE_EFFORT`; Codex: `turn_context…settings.reasoning_effort` | `medium`, `high` |

An unknown part is `-` (e.g. `claude/claude-opus-5-5/-`). With no agent environment at all (an operator running sdlc in a plain shell), sdlc's commits say `Agent: sdlc`. `Co-Authored-By` stays as the human-facing line; `Agent:` is the machine record.

## Core concepts

### Pure entities

| Name | Lives in | Status |
|------|----------|--------|
| `Identity` (+ `String`, `Parse`) | `cmd/sdlc/internal/agentid/agentid.go` | new |
| `Stamp(msg, id)` | `cmd/sdlc/internal/agentid/agentid.go` | new |
| `claudeFromTail`, `codexFromTail` | `cmd/sdlc/internal/agentid/transcript.go` | new |
| `MissingAgent(commits)` | `cmd/sdlc/internal/agentid/agentid.go` | new |

- **Identity** — `{Agent, Model, Effort}`; `String()` renders `agent/model/effort` with `-` for empty parts, `sdlc` when Agent is empty; `Parse` is its inverse (exported for #365 and `sdlc actual`).
  - **DRY rationale:** one renderer and one parser; every commit site and every reader share it (ARCH-DRY).
- **Stamp** — appends `Agent: <id>` to a commit message's trailer block (joins an existing trailer block such as `Review-Verdict:`/`Tracker-Operation:`, else adds a blank line first); idempotent: an existing `Agent:` trailer is left alone.
- **claudeFromTail / codexFromTail** — parse the last records of a transcript (bytes in, `Identity` parts out); tolerate a truncated first line and unknown records, returning empty parts rather than guessing (ARCH-SECURE: the transcript is input this process did not write).
- **MissingAgent** — from `(sha, subject, trailers)` rows, the non-merge commits without an `Agent:` trailer.

### Integration points

| Name | Lives in | Status | Wraps |
|------|----------|--------|-------|
| `Current()` | `cmd/sdlc/internal/agentid/detect.go` | new | env + transcript file read |
| commit sites (list below) | `cmd/sdlc/*.go`, `cmd/sdlc/internal/gitx/*.go` | modified | `git commit` / `git commit-tree` |
| `sdlc agent-id` | `cmd/sdlc/agentid.go` | new | stdout |
| close window report | `cmd/sdlc/close.go` | modified | `git log` over the review window |

- **Current()** — resolves once per process (`sync.Once`): agent from `judge.CurrentAgentDefaultEnv()`; the session file from `CLAUDE_CODE_SESSION_ID` (`~/.claude/projects/*/<sid>.jsonl`) or `CODEX_THREAD_ID` (`~/.codex/sessions/**/rollout-*-<id>.jsonl`, the `CodexHarness` root); reads at most the last 256 KiB. Tests inject home dir and env through a `Detector{Home, Getenv}` value; production uses the real ones.
- **Commit sites stamped** (every non-test site from the grep of `"commit"`/`"commit-tree"`): issuemovedetail.go:401, issuerestore.go:159, issuemigrate.go:620, propagatebase.go:347, changecode.go:311, handoff.go:92, closetracker.go:174 (`commitIndexOnto`, covers close/milestone evidence and `commitMirror`), claim.go:596, abandon.go:536/655, migrate.go:445, push.go:119/283, republish.go:426, trackerenv.go:188, gitx/refbootstrap.go:176, gitx/commitpublication.go:215 (tracker cards, `issue new`/`publish`), the landing merge message (landingintegrate.go:185) and the `gh pr merge` body (ghclient.go:139). Left alone: gitx/recoveryref.go (a recovery ref, not history) and gitx/rebasedpatch.go (a throwaway patch-id commit).
  - Stamping happens where each message is built, through `agentid.Stamp(msg, agentid.Current())`. gitx gets the identity as a parameter (no env reads inside gitx) so its tests stay hermetic.
  - **Sweep test (ARCH-PURPOSE):** `TestEveryCommitSiteStamps` scans non-test `.go` files under `cmd/sdlc` for `"commit"`/`"commit-tree"` git argument literals and fails on any file not in the stamped list or the two named exemptions, so a new commit site can't skip the trailer silently.
- **Close report** — after the review, `sdlc close` lists window commits (`BASE..HEAD`, `--no-merges`) without `Agent:` as a warning line with short shas and subjects. A report, not a gate: older Codex slots and hand commits would otherwise block closes before agents adopt the convention. Promoting it to a `--no-agent-trailer` gate later is one catalog row.

## Operating envelope and lifecycle

- **ARCH-CONSTRAINTS:** runs once per sdlc process on a commit path (interactive CLI). Transcripts reach 100+ MB, so the read is bounded to the file's last 256 KiB (one seek + read, well under 10 ms); a miss there yields `-` parts, never a full scan. Memory, concurrency, network: N/A, one small read.
- **ARCH-SECURE:** reads only the session's own transcript named by its env id, never another session's; parses JSON per line and treats any malformed line as absent. Writes no credentials; the trailer holds agent/model/effort names only.
- **ARCH-ORDER:** holds no state between events because `Current()` is a per-process memo of a read-only lookup.
- **ARCH-FUNERAL:** creates nothing durable except one trailer line per commit, which lives as long as the commit.
- **ARCH-MOCK:** the external surfaces are the transcript files (fixtures of real Claude/Codex record shapes under testdata) and git (tests use real git in temp repos via testfix, as the rest of cmd/sdlc does).

## Tasks

- [ ] **Task 1: agentid pure core.** TDD `Identity.String/Parse` (round trip, `-` parts, `sdlc`), `Stamp` (no trailers, existing trailer block, already stamped, message without trailing newline), `MissingAgent`. Files: `cmd/sdlc/internal/agentid/agentid.go`, `agentid_test.go`.
- [ ] **Task 2: transcript readers + Current().** Fixtures: a Claude tail (assistant record with `model`/`effort`, truncated first line) and a Codex tail (`turn_context` with `model`/`reasoning_effort`). TDD `claudeFromTail`, `codexFromTail`, then `Detector.Current` with temp homes (Claude session found by id, Codex rollout found by thread id, missing file → `-` parts, no agent env → `sdlc`, `CLAUDE_EFFORT` fallback). Files: `transcript.go`, `detect.go`, tests, `testdata/`.
- [ ] **Task 3: `sdlc agent-id`.** Prints `Agent: …`; `--json` prints the three parts. Test via `buildRoot()` with a fake env. File: `cmd/sdlc/agentid.go`, test.
- [ ] **Task 4: stamp the commit sites** listed above, then the sweep test. Extend one existing test per helper family to assert the trailer: `commitIndexOnto` (close evidence), tracker card commit (`issue new`), `claim` sync, `changecode` checkpoint, landing merge message.
- [ ] **Task 5: close report.** `sdlc close` warns with the list from `MissingAgent` over the review window; test with a window holding one stamped and one bare commit.
- [ ] **Task 6: docs.** AGENTS.base.md §12: keep `Co-Authored-By`, add the `Agent:` trailer and how to get it (`sdlc agent-id`, then `git commit … --trailer 'Agent: …'`); regenerate AGENTS.md/CLAUDE.md with weave; atlas: `atlas/workflow/sdlc-binary.md` gains the trailer and verb.
- [ ] **Task 7: verify.** Targeted tests per task; full `make test` under the gate before close; `git log --format='%(trailers:key=Agent)'` over this issue's own window shows one value per commit (this session's commits included, since I stamp mine from Task 3 on).

## Out of scope

- Grouping `sdlc actual` or `sdlc usage` by the trailer (#365 owns usage; `Parse` is exported for it).
- Rewriting history or back-filling old commits.
- A git commit-msg hook (local-gates territory; the close report covers detection).
