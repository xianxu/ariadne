# spine off-workflow guards — Implementation Plan

> **For agentic workers:** single-pass atomic change (one review boundary — no Mx tags). TDD; checkbox steps.

**Goal:** the sdlc lifecycle refuses to run where the workflow state says it shouldn't: (1) in a **brain/non-SDLC repo** (the #172 audit's off-workflow concentration — brain's own merged constitution *invites* sdlc, so the binary must own the gate, #69 pattern); (2) **start-plan/change-code on a `done` issue** (terminal per `issue.cue` — only re-close was guarded until now).

**Inventory result (prerequisite, recorded in the issue Log):** brain holds 1 open + 9 archived issues; the archived #1–#11 are June-era engineering work predating the capture-repo charter — nothing to migrate (already done+archived). The audit's "brain #116/#128" anomalies were **cwd-label noise** (sessions started in brain that `cd`'d to pair/parley — the transcript slug stays brain). The one open issue (brain#12 bootstrap-mac) stays as an inert file; the guard blocks lifecycle verbs, not files.

**Architecture — one gate family, three pieces:**
1. `guardSpineRepo(stderr)` — shared preflight at the top of each lifecycle verb's RunE (claim, start-plan, change-code, milestone-close, close, merge, push — exactly `processmanual.WorkflowVerbs()`, the ARCH-DRY tie): `.brain/config.md` at repo top → **die** with the charter + positive path (captures via datatype/plain git; engineering in a peer work repo); else no `workshop/issues/` dir → **die** "not an SDLC repo". Escape hatch: `WF_SPINE_GUARD=off` env (not a per-verb flag — 7 new flags is surface; the env is the documented emergency) which **cwarn-ACKs** so the #172 instrument can measure bypasses. Reads (estimate-source, actual, state, process-manual, issue list/show) stay unguarded by construction — sdlc legitimately READS brain.
2. **done-issue guard** in start-plan + change-code: issue `status: done` → die "terminal — open a new issue referencing #N (deliberate reopen: `sdlc set-status`)". Close-side already has the reclose guard; this closes the front door.
3. **brain constitution override** (in the brain repo): `AGENTS.local.md` gains the charter section so brain's merged entry files lead with "no sdlc lifecycle here + what to do instead"; re-run `weave compile` in brain (prose faces). Committed in brain locally; push left to the operator (gcrypt remote).

**Instrument safety:** the two die messages + the env-bypass cwarn must not collide with `GateCatalog` patterns (shared `assertNoGatesigCollision`).

## Tasks

### Task 1: guardSpineRepo + done-issue guard (TDD)
**Files:** new `cmd/sdlc/repoguard.go` + `repoguard_test.go`; wire into `claim.go`, `startplan.go`, `changecode.go`, `milestoneclose.go`, `close.go`, `merge.go`, `push.go`.
- [x] **Step 1: Failing tests** — (a) command-tree drift test: for EVERY verb in `processmanual.WorkflowVerbs()`, `buildRoot().Execute()` in a hermetic repo carrying `.brain/config.md` → expectDie with the charter message (a new catalog lifecycle verb automatically demands the guard); (b) no-`workshop/issues` repo → die "not an SDLC repo"; (c) `WF_SPINE_GUARD=off` → proceeds past the guard with the cwarn ACK; (d) normal SDLC repo → guard silent; (e) done-issue: `start-plan`/`change-code` on `status: done` → die with the new-issue/reopen next-action; non-done statuses pass; (f) gatesig no-collision for all three new lines.
- [x] **Step 2–4:** implement; `go test ./cmd/sdlc/...` PASS.
- [x] **Step 5: Commit** — `#176: spine repo guard + done-issue guard (one gate family)`.

### Task 2: docs + brain constitution + verify + close
- [x] Helptext: a GUARDS note on the affected verbs' help (close.md already documents gates — add the family) + `AGENTS.base.md` §0/§5 one-liner + atlas sdlc-binary.md gate section.
- [x] Brain: append the charter section to `~/workspace/brain/AGENTS.local.md`, run `weave compile` there (prose faces), commit locally (operator pushes — gcrypt).
- [x] Live verify: real brain repo → each guard arm observed (`sdlc claim` in brain dies with charter; `WF_SPINE_GUARD=off sdlc state`-class reads unaffected); hermetic non-SDLC dir → dies; done-issue arm against an archived done issue in a hermetic repo.
- [x] Close: `sdlc close --issue 176 --verified '<evidence>'`.

## ARCH notes
- **ARCH-DRY:** the guard's verb set derives from `processmanual.WorkflowVerbs()` in the drift test — the same catalog the friction instrument uses; one guard func, 7 one-line call sites.
- **ARCH-PURE:** guard decision = pure predicate over (brainMarker, issuesDirExists, env); IO confined to two `os.Stat`s.
- **ARCH-PURPOSE / Root Cause:** fixes the constitution-invites-sdlc root (brain's own AGENTS.local.md) AND enforces in the binary; the charter stops living only in one agent's memory.
- **Simplicity:** env escape, not 7 new flags; reads exempt by wiring, not by classification.

## Revisions

### 2026-07-14 — executed (both tasks complete)

Deltas: (1) `guardSpineRepo`'s issues-dir stat anchors at the REPO TOP when
WF_ISSUES_DIR is unset (cwd-relative stat false-positived from subdirectories —
caught by the existing close command-tree test running with cwd inside
cmd/sdlc; env override still honored cwd-relative, matching the verbs). (2) The
done-issue guard wires into start-plan's RunE (locate-and-check; skips
silently when the issue file can't be found — the verb's own error is better)
and into runChangeCode right after path resolution. (3) Brain-side: the charter
section landed in AGENTS.local.md and weave-compiled into brain's AGENTS.md +
CLAUDE.md (ariadne-built weave binary — `go run` from brain cwd picks up
brain's go.mod); brain's autosave daemon committed it. Live-verified all three
arms in the REAL brain: claim → charter refusal; estimate-source (read) →
unaffected; WF_SPINE_GUARD=off → cwarn ACK + proceed.

### 2026-07-14 — close-review fixes (FIX-THEN-SHIP)

Important #1 (ARCH-PURPOSE): the "instrument can measure the env bypass" claim
was a hand-maintained restatement — the #172 classifier derives exclusively
from GateCatalog, which has no row for this family. Claim corrected at all
three sites (repoguard.go, atlas, this plan): the ACK is
**transcript-greppable; instrument wiring is deliberate follow-up** (needs an
env-gate GateSig variant + a drift-guard exemption — noted for #170's orbit).
Important #2: the non-done-pass test now exists
(TestGuardIssueNotDone_WorkingIssuePasses) — guard-layer for both consumers +
command-tree through start-plan; change-code can't be driven past the guard in
tests (its downstream gates exit via bare exitWithCode). Minors: notSDLCRepoMsg
names the actual failing dir (incl. WF_ISSUES_DIR overrides); start-plan's
done-guard resolution now repo-top anchored like guardSpineRepo;
"three new lines" → four. Also noted per review §7: the constitution edit
landed in the §1 Peer-Repo brain bullet (not §0/§5), and Done-when #3 ("brain
drops out of the friction report") is verifiable only at the next re-measure.
