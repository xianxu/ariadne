# `codecomplete` Status + Two-Gate Publish Implementation Plan

> **For agentic workers:** Consult AGENTS.md Section 3 (Subagent Strategy) to determine the appropriate execution approach: use superpowers-subagent-driven-development (if subagents are suitable per AGENTS.md) or superpowers-executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking. This is a **multi-milestone** plan — each `Mx` is its own review boundary closed via `sdlc milestone-close`.

**Goal:** Add an intermediate `codecomplete` issue status so `sdlc close` (local) is the LLM acceptance gate that flips `working → codecomplete`, and `sdlc merge`/`push` (deterministic publish) flip `codecomplete → done` after verifying nothing drifted since close — folding in #142 (remove merge-time `plan`/`specs` LLM judges; the close boundary review owns docs sync incl. README).

**Architecture:** Two gates. **`close`** owns all LLM review (boundary review + the relocated `lessons` ping) and writes `status: codecomplete` — the *only* writer of that value (`set-status` refuses it, like `done`), which makes the commit carrying `codecomplete` a trustworthy anchor. **`merge`/`push`** run no LLM: they verify `HEAD == the codecomplete anchor commit` (the Q5 invariant — nothing landed since close), refuse with a re-run-`close` message otherwise, then flip `codecomplete → done` and archive. `codecomplete` is an **active** status (work isn't finished until merged).

**Tech Stack:** Go (`cmd/sdlc`, `pkg/vocab`), CUE vocabulary (`construct/vocabulary/issue.cue` → `make vocab-embed` → embedded `issue.json`), embedded prompts, golden/conformance tests.

**Architectural principles applied:**
- **ARCH-PURPOSE** — the issue's purpose is a *lifecycle* fix (stop `done` meaning "I think I'm finished"). Delivering it means every consumer of the status model derives `codecomplete` from the CUE source (`pkg/vocab`), not hand-restated enums — and the invariant is *enforced* at merge, not just documented. The shadow-sweep of status consumers is M1's job.
- **ARCH-DRY** — one status model (`issue.cue`) drives set-status help, gates, archiving; the invariant reuses existing issue-file git-history helpers; the two publish paths (merge/push) share the flip+invariant code.
- **Base-layer awareness** — `issue.cue` is a base-layer file; the vocabulary change propagates downstream (`atlas/workflow/base-layer.md`). M1 weighs it.

---

## Core concepts

### Pure entities (the conceptual core)

| Name | Lives in | Status |
|------|----------|--------|
| `issue.cue` codecomplete model | `construct/vocabulary/issue.cue` | modified |

- **`issue.cue` codecomplete model** — adds `codecomplete` to `categories.active`, a `when` line, and lifecycle transitions (`working→codecomplete` [close], `codecomplete→done` [merge], `codecomplete→working` [reopen], `codecomplete→{wontfix,punt}` [abandon/defer]). The `actual_hours!` compiled guard extends from `done` to `{done, codecomplete}`. Regenerated into `pkg/vocab/issue.json` via `make vocab-embed` (which runs `go generate ./pkg/vocab/...`; the `construct/generated/vocabulary/*` materialization is gitignored — NOT committed).
  - **DRY rationale:** single source; `pkg/vocab` predicates (`IsActive`/`IsTerminal`) + set-status help + close/merge gates all derive. No new predicate needed — `codecomplete` is `IsActive`; value-specific sites test the literal `"codecomplete"` (the #122 convention, mirroring literal `"done"`).
  - **Future extensions:** further active sub-states (e.g. `in-review`) slot into `categories.active`.

> **No `anchorSatisfiesInvariant` pure function** (review #6): a named unit-tested `head==anchor` is ceremony. The comparison inlines into `runPublishGate`; the test budget goes to `codecompleteAnchorCommit` (git derivation) + the drift/multi-issue enumeration cases, where the real risk lives.

### Integration points (where pure meets the world)

| Name | Lives in | Status | Wraps |
|------|----------|--------|-------|
| close status write | `cmd/sdlc/close.go:486` | modified | issue-file mutation |
| set-status codecomplete refusal | `cmd/sdlc/setstatus.go` | modified | transition guard |
| `mergedCodecompleteIssues` | `cmd/sdlc/publishgate.go` | new | `git diff` origin/main..HEAD |
| `codecompleteAnchorCommit` | `cmd/sdlc/publishgate.go` | new | `git log` on issue path |
| `runPublishGate` | `cmd/sdlc/publishgate.go` | new | invariant check + flip |
| `touchedIssuesNotDone` codecomplete carve-out | `cmd/sdlc/push.go`, `cmd/sdlc/merge.go` | modified | status classification |
| `lessons` ping relocation | `cmd/sdlc/close.go`, `cmd/sdlc/merge.go`, `cmd/sdlc/push.go` | modified | reminder output |
| README docs gate | `cmd/sdlc/internal/judge/code-review.md` | modified | boundary-review prompt |
| merge/push flip+archive | `cmd/sdlc/merge.go`, `cmd/sdlc/push.go` | modified | git + issue-file |

- **close status write** — `close.go:486` changes `SetField(newFM, "status", "done")` → `"codecomplete"` (+ the log/msg strings). The re-close guard (`close.go:401`) keeps keying on `"done"` (re-closing a *published* issue is the guard case); re-closing a `codecomplete` issue is the *normal rework* path and stays allowed. `actual_hours` is still recorded here (measured at close).
- **set-status codecomplete refusal** — `set-status` must refuse `→ codecomplete` (only `close` writes it, so the anchor is trustworthy), exactly as it refuses `→ done` (`setstatus.go:244`). This is the load-bearing enforcement for Q5.
- **`mergedCodecompleteIssues(issuesDir) []string`** — enumerates the issues this merge/push publishes: issue files changed in `origin/main..HEAD` whose current `status == "codecomplete"`. Mirrors `touchedIssuesNotDone`'s window-scan (ARCH-DRY). Solves review #3 — merge/push have no `--issue`, so they discover the set from the diff.
- **`codecompleteAnchorCommit(issuePath) string`** — the most recent commit touching the issue file that leaves it at `status: codecomplete`. Since only `close` writes `codecomplete`, this is necessarily a close commit. **Derivation (NOT the trailer-grep pattern):** `previousReviewBoundary`/`latestVerdictCommit` grep a *commit message* trailer; this needs the file's *resulting content* — walk `git log --format=%H -- <issuePath>` newest-first and return the first commit where `git show <sha>:<issuePath>` parses to `status: codecomplete`. A genuinely different derivation (a content read, not a message grep) — so it's a *new* helper, not an ARCH-DRY reuse of the trailer helpers.
- **`runPublishGate(...)`** — shared by merge and push (replaces the removed `plan`/`specs` preflight judges). Enumerate `mergedCodecompleteIssues`; compute each one's `codecompleteAnchorCommit`; take the **latest** anchor. The last close on the branch is a *whole-issue* close whose boundary-review window is `branch-point..HEAD` — it covered the entire branch (incl. earlier issues' code), so a **branch-level** check (HEAD unchanged since the last close) is sufficient and avoids a false per-issue "drift" refusal on multi-issue branches (review #3). Refuse unless `rev-list <latest-anchor>..HEAD` is empty — nothing landed after the last close (message: *"commits landed after `sdlc close`; re-run it to re-review"*). Comparison inlined (no `anchorSatisfiesInvariant`, review #6).
  - **Injected into:** merge step 5 (invariant, pre-merge) + the main-side flip; push step 4 likewise.
- **`touchedIssuesNotDone` codecomplete carve-out** — it (`push.go:506`, `merge.go:369`) treats any non-`IsTerminal` status as "NOT done," which would fire the "Continue anyway?" prompt on *every* normal merge now that the pre-merge state is `codecomplete` (active). Carve `codecomplete` out (it's about to be flipped) — warn only for open/working/blocked (review #2).
- **`lessons` ping relocation** — `LessonsReminder` emission moves to `close`; `merge`/`push` stop emitting it. `merge.go:315`/`push.go:135` pass an **explicit** `[]judge.Category{Plan, Specs, Lessons}` that overrides preflight's default, so M2 only *adds* the ping to close; M3 removes it from merge/push when it replaces the preflight call with `runPublishGate` (review #1).
- **merge/push flip+archive** — after the (pre-merge) invariant check + the actual merge, the main-side step flips each merged `codecomplete` issue `→ done`, then archives (archiving keys on `IsTerminal`, so it must run *after* the flip). `merge.go:568` / `push.go` archiving ordering adjusts.

**Test surface:** `issue.cue` laws via `pkg/vocab/conformance_test.go` (codecomplete must be reachable + escapable); `codecompleteAnchorCommit`, `mergedCodecompleteIssues`, and `runPublishGate` (incl. the drift-refusal + multi-issue cases) via temp-repo integration tests (pattern: `milestonewindow_test.go`); close/set-status transitions via existing test files.

---

## M1 — Vocabulary: add `codecomplete` to the model

**Review boundary.** Closes via `sdlc milestone-close --issue 160 --milestone M1`.

**Files:**
- Modify: `construct/vocabulary/issue.cue`
- Regenerate + commit: `pkg/vocab/issue.json` ONLY (via `make vocab-embed` → `go generate ./pkg/vocab/...`). `construct/generated/vocabulary/issue.json` is **gitignored** — do NOT commit it (review #4).
- Test: `pkg/vocab/conformance_test.go`, `pkg/vocab/vocab_test.go`
- Docs: `atlas/workflow/issue-lifecycle.md`, `atlas/workflow/vocabulary.md`

- [ ] **Step 1: Read the model + its laws**

Run: `go test ./pkg/vocab/ -v` — baseline PASS. Note the laws (`reachable`, `escapable`, `documented-value`) that adding a status must satisfy.

- [ ] **Step 2: Write a failing vocab test for codecomplete**

Add to `vocab_test.go`: assert `vocab.Issue().IsActive("codecomplete")` is true, `IsTerminal("codecomplete")` is false, and `When["codecomplete"] != ""`.
Run: `go test ./pkg/vocab/ -run CodeComplete -v` → FAIL.

- [ ] **Step 3: Edit `issue.cue`**

- `categories.active: ["working", "blocked", "codecomplete"]`
- `when.codecomplete: "code complete; passed local acceptance review, awaiting merge"`
- Lifecycle: relocate `working → done` to `working → codecomplete` (event `close`, same guards `["actual-recorded", "verified", "atlas-updated"]`); add:
  - `{from: "codecomplete", to: "done", event: "merge", guards: ["reviewed-head-unchanged"]}`
  - `{from: "codecomplete", to: "working", event: "reopen"}`
  - `{from: "codecomplete", to: "wontfix", event: "abandon"}`
  - `{from: "codecomplete", to: "punt", event: "defer"}`
  - **relocate BOTH close edges** — `sdlc close` writes the close-target status *unconditionally* (`close.go:486`, no transition-table consultation, no `blocked` handling), so a `blocked` issue closed via `close` goes straight to `codecomplete`. The model MUST allow it or it contradicts the code (plan-quality FAILURE). So: change `working → done` **and** `blocked → done` to `working → codecomplete` and `{from: "blocked", to: "codecomplete", event: "close", guards: ["actual-recorded","verified","atlas-updated"]}`. (Forcing unblock-first would be a *separate* behavior change needing a `blocked`-refusal in `close.go` — not in scope.) Verify `reachable`/`escapable` laws still hold (done reachable via codecomplete; codecomplete reachable via working+blocked; escapable via done/working).
- Extend the compiled guard: `if status == "done" || status == "codecomplete" { actual_hours!: (number & >0) | #ActualNotApplicable }`.

- [ ] **Step 4: Regenerate + verify laws**

Run: `make vocab-embed` (runs `go generate ./pkg/vocab/...` → regenerates `pkg/vocab/issue.json`; it also asserts `git diff --exit-code -- pkg/vocab`). Then `git diff pkg/vocab/issue.json` — confirm codecomplete appears in active + when + lifecycle. (Ignore any `construct/generated/vocabulary/*` churn — gitignored.)
Run: `go test ./pkg/vocab/ -v` — PASS (laws hold, new test green). If a law fails, the transition graph is incomplete — fix `issue.cue`, don't edit JSON.

- [ ] **Step 5: Shadow-sweep status consumers (ARCH-PURPOSE)**

`grep -rn '"done"\|IsTerminal\|IsActive' cmd/sdlc pkg` — for each site, confirm it behaves correctly with codecomplete present. Record the sweep in `## Log`, noting at minimum: `state.go` `detectDrift` handles codecomplete gracefully with no change (not terminal → no "should be archived" drift `state.go:316`; not open/`"working"` → not a close-off candidate `state.go:346`); the two sites that DO need changes are deferred to their milestones — `touchedIssuesNotDone` (`push.go:506`/`merge.go:369` → M3, review #2) and archiving order (`merge.go:568` → M3).

- [ ] **Step 6: Update atlas**

`atlas/workflow/issue-lifecycle.md` + `vocabulary.md`: add `codecomplete` to the status list + the lifecycle diagram/prose (working → codecomplete → done). Keep `atlas/index.md` links intact.

- [ ] **Step 7: Commit + milestone-close**

```bash
git add construct/vocabulary/issue.cue pkg/vocab/issue.json atlas/
git commit -m "#160 M1: add codecomplete active status to the issue vocabulary model"
sdlc milestone-close --issue 160 --milestone M1
```
Paste the `Review-Verdict:` trailer into the close commit; fix Critical/Important before crossing.

---

## M2 — `close` → `codecomplete` (+ lessons relocation + README docs gate)

**Review boundary.** Closes via `sdlc milestone-close --issue 160 --milestone M2`.

**Files:**
- Modify: `cmd/sdlc/close.go` (:486 status write, msg; verify re-close guard :401)
- Modify: `cmd/sdlc/setstatus.go` (refuse `→ codecomplete`)
- Modify: `cmd/sdlc/close.go` (ADD the `lessons` ping to close; the removal from merge/push is M3, review #1)
- Modify: `cmd/sdlc/internal/judge/code-review.md` + `judge_test.go:376` + golden testdata (README docs gate — folded #142 Task 4)
- Modify: `cmd/sdlc/helptext/close.md`
- Test: `cmd/sdlc/close_test.go`, `cmd/sdlc/setstatus_test.go`, `cmd/sdlc/internal/judge/*`
- Docs: `atlas/workflow/issue-lifecycle.md` (close semantics)

- [ ] **Step 1 (TDD): close flips to codecomplete.** Add/adjust a `close_test.go` case asserting a full-issue close on a finalizing verdict writes `status: codecomplete` (not `done`) and still records `actual_hours`. Run → FAIL.
- [ ] **Step 2:** Change `close.go:486` to `SetField(newFM, "status", "codecomplete")`; update the `msg` strings (`flipped … → status: codecomplete`). Run → PASS.
- [x] **Step 3 (TDD): set-status refuses codecomplete. → DONE IN M1** (rebalanced): relocating `working → done` made that transition illegal, breaking `TestCheckTransitionGuards_RefusesDone` — the set-status enforcement is the direct counterpart of the model change (ARCH-PURPOSE: enforce, don't just declare), so it shipped with M1. Added Guard 1b (`→ codecomplete` → refuse, route to `sdlc close`), updated Guard 1 (`→ done` now routes through the close→merge publish flow), and updated/added the guard tests (`RefusesDone` now exercises `codecomplete → done`; new `RefusesCodecomplete`; clean codecomplete reopen/abandon/defer edges added to `NormalTransitions`).
- [ ] **Step 4: Re-close guard.** Verify `close.go:401` still keys on `"done"` (re-closing a `codecomplete` issue must be ALLOWED — the rework path). Add a `close_test.go` case: closing an already-`codecomplete` issue succeeds (re-review) and produces a fresh codecomplete commit. Adjust only if the guard wrongly blocks codecomplete.
- [ ] **Step 5 (TDD): lessons at close.** Assert `sdlc close` emits `judge.LessonsReminder` (the no-LLM ping). Run → FAIL → add the reminder emission to the close path → PASS. (Do NOT assert merge/push stop emitting it here — they pass explicit categories including `Lessons`; that removal happens in M3 when the whole preflight call is replaced by `runPublishGate`. Review #1 — asserting it in M2 would fail against M2's file set.)
- [ ] **Step 6: README docs gate (folded #142 Task 4).** Apply #142 plan Task 4 verbatim: `code-review.md:55` cross-ref + `:73` section rename to "Docs update gate (atlas + README)" with the README bullet; update `judge_test.go:376` `"Atlas update gate"` → `"Docs update gate"`; regenerate golden (`go test ./cmd/sdlc/internal/judge -run TestBuildPrompt_Golden -update-golden`). Run judge tests → PASS.
- [ ] **Step 7: Help + atlas.** `close.md`: close now flips to `codecomplete` (not done), owns all LLM review + the lessons ping. `issue-lifecycle.md`: close semantics.
- [ ] **Step 8: Commit + milestone-close.** `#160 M2: close flips working→codecomplete; owns lessons + README docs gate`. `sdlc milestone-close --issue 160 --milestone M2`.

---

## M3 — Publish gate: `merge`/`push` flip `codecomplete → done`

**Review boundary.** Closes via `sdlc milestone-close --issue 160 --milestone M3` (the final milestone; then `sdlc close --issue 160`).

**Files:**
- Create: `cmd/sdlc/publishgate.go` (`mergedCodecompleteIssues`, `codecompleteAnchorCommit`, `runPublishGate` — comparison inlined, no `anchorSatisfiesInvariant`) + `cmd/sdlc/publishgate_test.go`
- Modify: `cmd/sdlc/merge.go` (step 5 → publish gate; flip before archive), `cmd/sdlc/push.go` (step 4 → publish gate; flip before archive)
- Modify: `cmd/sdlc/preflight.go` (remove `plan`/`specs`/`lessons` from the pre-merge path — now empty of LLM judges; delete `runPreflightJudges` merge-usage or repurpose)
- Modify: `cmd/sdlc/helptext/merge.md`, `cmd/sdlc/helptext/push.md`
- Docs: `atlas/workflow/pre-merge-checks.md`, `atlas/workflow/ci-merge-check.md`

- [ ] **Step 1 (TDD, integration): `codecompleteAnchorCommit`.** Temp-repo test (pattern: `milestonewindow_test.go`): a commit that sets the issue to `status: codecomplete` is returned; a later commit (code or issue-file that leaves status unchanged) is NOT the anchor. Implement by walking `git log --format=%H -- <issuePath>` newest-first and returning the first commit where `git show <sha>:<issuePath>` parses to `status: codecomplete` (a content read — NOT the trailer-grep of `previousReviewBoundary`). PASS. (No separate `anchorSatisfiesInvariant` — the `rev-list <anchor>..HEAD` empty-check lives inline in `runPublishGate`, review #6.)
- [ ] **Step 2 (TDD, integration): `mergedCodecompleteIssues`.** Temp-repo: two issues changed in `origin/main..HEAD`, one `codecomplete` one `working` → only the codecomplete one is returned. Implement by scanning the `origin/main..HEAD` issue-file diff + reading each file's current status (mirror `touchedIssuesNotDone`). PASS.
- [ ] **Step 3 (TDD): `runPublishGate` refuses on drift, incl. multi-issue.** Temp-repo cases: (a) single codecomplete issue, HEAD == its anchor → passes; (b) a commit after close → refuses with the re-run-`close` message; (c) **two** codecomplete issues closed in sequence (anchors X then Y=HEAD) → passes (uses the *latest* anchor Y = the last whole-issue close, which reviewed `branch-point..HEAD`), NOT a false per-issue refusal (review #3); (d) mixed-state branch — one `codecomplete` + one still-`working` issue → `mergedCodecompleteIssues` returns only the codecomplete one, and the gate behaves on its anchor alone. Implement `runPublishGate`: enumerate → latest anchor → refuse iff `rev-list <latest>..HEAD` non-empty. PASS.
- [ ] **Step 4: Wire merge.** Replace `merge.go` step 5 (`runPreflightJudgesFn(...)` with plan/specs/lessons) with `runPublishGate(...)` (pre-merge, on the feature branch — before the step-10 merge). After the merge, in the main-side step, flip each `mergedCodecompleteIssues` entry `codecomplete → done` (record nothing new — actuals already set at close) BEFORE `archiveDoneIssuesInDir` (which keys on `IsTerminal`). Adjust the `merge.go:568` flow.
- [ ] **Step 5: Wire push.** Same for `push.go` (direct-to-main): publish gate → flip `codecomplete → done` → archive → push. (Q3: push = "merge without a PR".)
- [ ] **Step 6: `touchedIssuesNotDone` carve-out (review #2).** Update `touchedIssuesNotDone` (`push.go:506`, `merge.go:369`) so `codecomplete` is NOT flagged as "not done" (it's about to be flipped) — warn only for open/working/blocked. Add a test asserting a codecomplete issue does not trip the "Continue anyway?" prompt.
- [ ] **Step 7: Remove the LLM judges + lessons ping.** Delete `plan`/`specs`/`lessons` from the merge/push explicit category slices (`merge.go:315`, `push.go:135`) — the publish gate replaces them; this also completes the M2 lessons relocation (review #1). Keep `sdlc judge plan/specs/lessons` as *ad-hoc* commands (the categories still exist) — only the auto-dispatch at merge/push is removed. Update `merge_e2e_test.go`/`push_test.go` expectations.
- [ ] **Step 8 (TDD): end-to-end path + re-close note.** Cover `working → close(codecomplete) → merge(done)`; and `close → code fixup → merge refuses → re-close → merge succeeds`. **Re-close must yield a fresh issue-file commit** to advance the anchor — the standard `sdlc close --verified '<evidence>'` appends a `## Log` line (`close.go:502-509`), guaranteeing a diff; document that a same-day re-close WITHOUT `--verified` may produce no issue-file change (stale anchor, review #5) — so re-close always with `--verified`. Also confirm a README gap is caught at close (M2 gate), not merge.
- [ ] **Step 9: Help + atlas.** `merge.md`/`push.md`: the two-gate model, `codecomplete → done` flip, the invariant refusal, no LLM at publish. `pre-merge-checks.md`: LLM judges now close-time; `ci-merge-check.md`: still the server-side deterministic gate (unchanged). Note the stale `make check` refs.
- [ ] **Step 10: Full suite + commit + close.** `go test ./...`. Commit `#160 M3: merge/push flip codecomplete→done + invariant; remove pre-merge LLM judges`. `sdlc milestone-close --issue 160 --milestone M3`, then `sdlc close --issue 160 --verified '<evidence>'`.

---

## Verification (whole-issue, before final `sdlc close`)

- [ ] `go test ./...` green; `go build ./cmd/sdlc`.
- [ ] Behavior: `sdlc close --issue N --dry-run` shows the flip target is `codecomplete` + the README docs gate text in the prompt; `set-status codecomplete` is refused; `sdlc merge --dry-run` on a clean close (HEAD==anchor) proceeds, and after an extra commit refuses with the re-run-`close` message; `lessons` prints at close, not merge.
- [ ] Downstream: note in `## Log` that the `issue.cue` change propagates via the base layer; no downstream repo currently has a `codecomplete` issue, so it's additive (no data migration).

## Done-when cross-check (issue #160)

- [x] codecomplete in the model, derived by `pkg/vocab` (M1).
- [x] close flips working→codecomplete carrying guards (M2).
- [x] merge flips codecomplete→done after the deterministic invariant check, no LLM (M3).
- [x] merge refuses on post-close drift with a re-run-close message (M3).
- [x] close boundary review owns README docs sync (M2, folded #142).
- [x] plan/specs auto-dispatch removed from merge/push; help updated (M3).
- [x] Tests: working→codecomplete→done; refuse on drift; README caught at close (M3).
