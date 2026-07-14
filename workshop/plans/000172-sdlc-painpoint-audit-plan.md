# sdlc Painpoint Audit — Implementation Plan

> **For agentic workers:** Consult AGENTS.md Section 3 (Subagent Strategy) to determine the appropriate execution approach: use superpowers-subagent-driven-development (if subagents are suitable per AGENTS.md) or superpowers-executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build `sdlc process-manual --friction-report` — a whole-corpus, cross-repo, both-agent instrument that measures cleanly (no grep contamination) where the `sdlc` spine creates friction: per-gate **bypass** rates, gate **refusal→retry** loops, and workflow **firing-order** anomalies — across the **12 bypass gates** on the five spine commands `close`/`milestone-close`/`change-code`/`merge`/`push` (`claim`'s `--no-start` is a workflow toggle, not a gate — excluded). Then triage the top-bypassed gates (T2) and read for un-gated recurring-error moments (T3).

**Architecture:** The bypass/refusal signal is already captured — sdlc's stderr is folded into the Bash tool's captured stdout, which is the `tool_result` block content linked to its call by `tool_use_id`; `extractStdout` already reads the right stream. The problem is **discrimination, not capture**: this repo *develops* sdlc, so `close.go` source and cat-n log reads spray every gate string into tool output. So the design (1) **anchors to `Bash(sdlc <verb>)` invocations only** (drops all source/edit/log-read noise — the dominant contamination), (2) classifies each output line against a **per-gate signature catalog** (three distinct ACK grammars + per-gate refusal shapes, matched on the runtime `\x1b[0m ` reset and rejecting `%s`/`cwarn(`/`NN\t` contamination markers), and (3) states honestly what is **unobservable** (change-code's silent no-ACK bypasses; refusals that never name their flag). Reuses `process-manual`'s pure parser; adds a whole-corpus discovery walk, three pure detectors, and a net-new Go codex parser derived from the one atlas format spec.

**Tech Stack:** Go (stdlib + cobra), `go test` — table-driven over **real captured** transcript fixtures (copied verbatim from `~/.claude/projects` / `~/.codex/sessions`, including contamination cases that must be REJECTED), no mocks.

---

## Core Concepts

### The signal catalog (ground truth — the load-bearing artifact)

Built from source + verified against the real corpus (2,356 Claude files + 592 codex rollouts). This table IS the classifier's data; the code is a table-driven matcher over it.

**Two discriminators — do NOT apply the reset trick to refusals (verified fix, v3.1).** cwarn/cinfo/cok ACKs always render with the runtime reset `\x1b[0m ` immediately before the message (`term.go:20-40`, ANSI is unconditional), so **ACK** detection requires that reset — contamination (`close.go` source, `%s`/`%d`, `cwarn(stderr,"`, `fmt.Sprintf("`, leading `"`, `NN\t` cat-n) lacks it. But **refusals** are built as plain `\n`-joined strings (`explainActual`/`explainNoAtlas` etc.) or `die()` (reset at line-END), so they carry **no** `\x1b[0m ` before the gate string — applying the reset rule to them filters refusals to zero. Refusals are instead discriminated by a **grammar-anchored `refusalRE`**: real numbers (`\d+`, not the `N`/`…` placeholders that appear when this very catalog or doc/atlas prose is read back as a transcript), colon-delimited counts (`: \d+ nonconforming`), no leading quote, and the exact per-gate tail (not a shared prefix — the C1 warmup lesson). A rejection test feeds the plan's own catalog text + a `grep -oh "Pass --no-actual…"` meta line → both `none`.

**Three ACK grammars — one regex will NOT do:**
- **G1 close/milestone-close:** `--no-<gate> (or --force): <verb…>` (paren + colon).
- **G2 change-code:** `<gate> gate[s] bypassed (--force: <reason>)` — flag NOT `--no-`-prefixed; **only emitted on `--force`** — used alone the gate skips **silently** (no ACK).
- **G3 merge/push:** `--no-<gate>: <verb…>` — colon, **no `(or --force)`**; `no-validate` adds `⚠️ …SKIPPING`.

| command | flag | ACK grammar / key pattern (post-`[0m`) | refusal pattern (or NONE) | runtime-bypass (claude / codex) | notes |
|---|---|---|---|---|---|
| close·mclose | no-actual | G1 `--no-actual (or --force): closing with actual_hours: N/A` | `ACTUAL=<hours> required for … close (§5 step 3)` (explainActual head, `close.go:1079`); tail `…only **when measurement is not applicable**` | 27 / 21 | **warmup twin** `…only if there's genuinely nothing` (`close.go:219`, on SUCCESS, 341×) — MUST reject |
| close·mclose | no-verified | G1 `--no-verified (or --force): closing with NO verification evidence` | `VERIFIED=… required for … close (§5 step 1)` (`close.go:1180`) | 0 / 0 | design works (Spec) |
| close | no-reclose-guard | G1 `--no-reclose-guard (or --force): re-closing …#N (already done)` | `…#N is already status: done — pass --no-reclose-guard (or --force) to re-close` (`close.go:405`) | 0 / **30** | codex re-closes far more |
| close·mclose | no-atlas | G1 `--no-atlas (or --force): skipping atlas/ change check` | `no atlas/ changes in <sha>..HEAD (§5 step 5)` … `or pass --no-atlas (or --force) with` (`close.go:1233`) | 20 / 19 | |
| close | no-verdict | G1 `--no-verdict (or --force): skipping Review-Verdict check for N milestone(s)` | `milestones … lack Review-Verdict trailer …` … `Or pass --no-verdict (or --force); record` (`close.go:1362`) | 14 / 3 | |
| close·mclose | no-plan-check | G1 `--no-plan-check (or --force): closing … with N unchecked ## Plan item(s)` | `## Plan has N unchecked item(s)` … **comma form** `(pass --no-plan-check, or --force, to close anyway)` (`close.go:478`) | 0 / 7 | |
| close·mclose | no-project | G1 `--no-project (or --force): skipping detail-block update for <a id=` | `no detail block … (§5 step 4)` … **comma form** `(--no-project, or --force, …)` (`close.go:571`) | 8 / 0 | |
| close | no-judge | **cinfo** `==> skipping issue boundary review per --no-judge (or --force)` (`close.go:818`) | **NONE** (auto-dispatch; REWORK verdict ≠ refusal) | 25 / 6 | |
| milestone-close | no-judge | **cinfo** `==> skipping milestone-review per --no-judge (or --force)` (`milestoneclose.go:161`) | **NONE** | 40 / — | |
| change-code | no-structural | **G2 (silent alone)** `structural gates bypassed (--force: …)` (`changecode.go:130`) | `structural-sanity gates failed:` … `re-run with --force <reason>` (`:127`) | 0 (alone: unobservable) | |
| change-code | no-estimate | **G2 (silent alone)** `estimate gate bypassed (--force: …)` (`:148`) | `estimate gate failed:` … **slash** `--no-estimate / --force <reason>` (`:145`) | 0 | |
| change-code | no-estimate-recon | **G2 (silent alone)** `estimate-reconciliation gate bypassed (--force: …)` (`:163`) | slash `--no-estimate-recon / --force` (`:160`) | 1 | |
| change-code | no-judge (plan-quality / estimate-quality) | **G2** `plan-quality gate bypassed (--force: …)` (`:174`); `estimate-quality gate bypassed (--force: …)` (`:180`) | `plan-quality: findings reported — … OR re-run with --force`; `estimate-quality: … --no-judge / --force` (`:387/:456`) | pq 52 / eq 41 | heavy `--force` use |
| merge | no-validate | **G3** `⚠️ --no-validate: SKIPPING the instance-conformance gate (#124)` (`merge.go:328`) | `instance-conformance gate: N nonconforming …, or --no-validate to bypass` (`validategate.go:82` + `merge.go:325`) | 3 (shared) / 19 | ACK byte-identical to push — disambiguate by verb |
| merge | no-judge | **G3** `--no-judge: skipping the pre-merge publish gate (#160 …)` (`merge.go:345`) | publish-gate `publish gate: N commit(s) landed after sdlc close` (`publishgate.go:116`) — **never names the flag** | 1 / 2 | refusal not flag-linkable |
| push | no-validate | **G3** identical `⚠️ --no-validate: SKIPPING …` (`push.go:129`) | bare `instance-conformance gate: N nonconforming …` (`push.go:126`, no hint) | (shared) / — | |
| push | no-judge | **G3** `--no-judge: skipping the pre-push publish gate (#160 …)` (`push.go:144`) | bare publish-gate err (`push.go:140`) | 1 / 6 | |

**Runtime ranking (the corrected headline):** judge (close 25 + mclose 40 = **65**) dominates; then change-code plan-quality (52) / estimate-quality (41) via `--force`; no-actual (27), no-atlas (20), no-verdict (14), no-project (8), no-validate (**3**, not the 87 a naive grep suggests — that was source), no-verified (**0** — the design works), no-reclose-guard (0 claude / 30 codex).

**Observability limits (stated, not hidden — ARCH-PURPOSE honesty):**
- **Silent bypasses:** change-code's four gates used *alone* emit no ACK — a bypass is observable only when `--force` was used. The report must label these gates "bypass observable only via --force" rather than imply a complete count.
- **Flag-omitting refusals:** merge/push `--no-judge` and push `--no-validate` refusals never name the flag, so refusal→retry for them is attributed by verb+gate-context, best-effort, and flagged as such.

### Pure entities

| Name | Lives in | Status |
|------|----------|--------|
| `GateSig` (command, flag, grammar, ackRE, refusalRE, warmupTrap, silentAlone, refusalNamesFlag) | `cmd/sdlc/internal/processmanual/gatesig.go` | new |
| `gateCatalog []GateSig` (the table above, single-sourced) | `cmd/sdlc/internal/processmanual/gatesig.go` | new |
| `SdlcInvocation` (verb, args, issueID, output, time, transcript, agent, repo, isHelp) | `cmd/sdlc/internal/processmanual/friction.go` | new |
| `GateEvent` (Bypass\|Refusal, gate, command, viaForce, observability, invocation) | `cmd/sdlc/internal/processmanual/friction.go` | new |
| `classifyOutputLine` (line → GateEvent\|none, via `gateCatalog` + runtime/contam discriminator) | `cmd/sdlc/internal/processmanual/friction.go` | new |
| `RefusalRetry` | `cmd/sdlc/internal/processmanual/friction.go` | new |
| `FiringOrderAnomaly` + `workflowOrder` (per-issue, iteration-aware) | `cmd/sdlc/internal/processmanual/friction.go` | new |
| `FrictionReport` + `renderFrictionReport` | `cmd/sdlc/internal/processmanual/friction.go` | new |
| `AllGates()` drift guard (catalog vs registered `--no-*` across all cmd files) | `cmd/sdlc/gates_test.go` | new |
| `classifyToolUse` (+ `KindFileEdit` for Edit/Write/MultiEdit) | `cmd/sdlc/internal/processmanual/session.go` | modified |
| `parseEvents` (attach linked tool_result output to every `KindSDLCPrompt`, not just close/mclose) | `cmd/sdlc/internal/processmanual/session.go` | modified |
| `codexMetaKind` + `parseCodexEvents` | `cmd/sdlc/internal/processmanual/codex.go` | new (M3) |

- **`GateSig` / `gateCatalog`** — the catalog table encoded as data; `classifyOutputLine` is a pure matcher over it. **DRY:** one source for the classifier, the drift guard, and the report's per-gate rows. A new gate = one table row (+ its flag registration).
- **`classifyOutputLine`** — strips ANSI, then: for a **bypass ACK** requires the runtime `[0m ` reset (G1/G3) or the change-code `gate bypassed (--force:` shape (G2), matched against each `GateSig`'s `ackRE`; for a **refusal** matches the grammar-anchored `refusalRE` (digits, colon-delimited, exact tail — NOT reset-gated). Both REJECT contamination markers (`%s`/`%d`, `cwrite(`/`cwarn(`, `fmt.Sprintf(`, leading `"`, `NN\t`) and the `warmupTrap`. The G3 `no-validate` ACK regex must tolerate the literal `⚠️ ` emoji (`\x{26a0}\x{fe0f}`) + **two** spaces before `--no-validate` (`merge.go:328`/`push.go:129`). Returns the `GateEvent` with `observability` (`full` | `force-only` | `flag-omitted`).
- **`SdlcInvocation`** — a `Bash(sdlc <verb> …)` call + its `tool_use_id`-linked result output. `issueID` parsed from `--issue N`/`#N` in args (keys firing-order per issue). `isHelp` (a `--help` invocation) excluded — its output legitimately lists every flag.
- **`workflowOrder` / `FiringOrderAnomaly`** — ordered stages from AGENTS.md's flow `claim(0) ≺ start-plan(1) ≺ change-code(2) ≺ milestone-close(3) ≺ close(4) ≺ merge(5)`, **keyed per issueID**, **iteration-aware**: legal loops that must NOT flag — `milestone-close→change-code` (next milestone), `start-plan` re-runs (AGENTS.md "re-run per design"), `close→change-code`/`close→start-plan` after a REWORK/reopen (`codecomplete→working`, `issue.cue:129/132`). Only a verb below the issue's max-reached stage with **no** intervening legal-loop trigger flags. **`merge`/`push` carry NO `--issue`** (they derive touched issues from the git diff, invisible to the transcript — `merge.go:105-110`, `push.go`), so stage-5 `merge` is **attributed from segment context** (the nearest preceding `--issue`-bearing invocation in the same segment) or, if none, recorded as `unattributed` and kept OUT of any per-issue ladder — never bucketed under a global `""` key that would cross-contaminate. `skill-late` = a plan/TDD `Skill` load after a `KindFileEdit` in the same segment+issue.

### Integration points

| Name | Lives in | Status | Wraps |
|------|----------|--------|-------|
| `enumerateAllTranscripts()` (all Claude slugs + all codex rollouts; worktree-normalized, temp-excluded repo labels) | `cmd/sdlc/internal/processmanual/friction.go` | new | `~/.claude/projects/*` + `~/.codex/sessions/**` |
| `--friction-report` + `--json` flag + dispatch | `cmd/sdlc/processmanual.go` | modified | cobra CLI |

- **`enumerateAllTranscripts`** — globs every `~/.claude/projects/<slug>/*.jsonl` + `~/.codex/sessions/**/rollout-*.jsonl`. Repo label from `<slug>`/`session_meta.cwd`, **normalizing ariadne worktrees** (`…-worktree-ariadne-…` → `ariadne`, review M2) and **excluding temp dirs** (`-private-tmp-`, `-private-var-folders-`). The only new IO seam; pure detectors receive parsed invocations.
- **`--friction-report`** — new dispatch branch in `processmanual.go`; mutually exclusive with `--session`; whole-corpus, so not repo-bound.

**Test surface.** All entities/detectors pure — table tests over **real captured** fixtures INCLUDING the rejection cases (the warmup line, `close.go` source lines, cat-n log reads, `%s` format strings → all classify to `none`). `enumerateAllTranscripts`/`classifyToolUse` are the thin IO/parse seam (real temp files). **Load-bearing tests:** (a) anti-contamination — a transcript with a `Grep(close.go)` + a cat-n log read + one real `sdlc close --no-atlas` → exactly ONE bypass; (b) warmup rejection — a successful `sdlc close` whose output contains the `:219` warmup → ZERO no-actual refusals; (c) cross-language codex golden — Go parser's keep/skip/bypass/refusal decisions match Python `agent_codex.py`/`normalize.py` on a shared fixture (spec-derived golden, no live `python3`).

---

## Chunk 1: M1 — the signature catalog + anchored bypass measurement (Claude, whole-corpus)

### Task 1: `GateSig` catalog + `AllGates()` drift guard (all commands)
**Files:** Create `cmd/sdlc/internal/processmanual/gatesig.go`; `cmd/sdlc/gates.go` (`AllGates()`); Tests `gatesig_test.go`, `gates_test.go`. Modify flag registrations in `close.go`/`milestoneclose.go`/`changecode.go`/`merge.go`/`push.go` to source names from `AllGates()`.
- [x] **Step 1: Failing test** — `AllGates()` == the **12 bypass-gate** flags across the **five spine commands** (close, milestone-close, change-code, merge, push). The drift guard asserts `AllGates()` equals every `--no-*`/gated-`--force` flag registered by those five command files, **explicitly excluding a known-non-gate allowlist** — currently `{no-start}` (`claim.go:62`, a workflow toggle on a sixth command, NOT a bypass gate: no ACK, no refusal). Without the allowlist the guard fails on day one (there are 13 registered `--no-*` flags; one is `no-start`). `gateCatalog` has one `GateSig` per row of the table above.
- [x] **Step 2: Run** → FAIL → **Step 3: Implement** `gatesig.go` (the table as data) + `gates.go`; point registrations at `AllGates()`.
- [x] **Step 4: Run** `go test ./cmd/sdlc/...` → PASS → **Step 5: Commit** — `#172 M1: gate signature catalog + cross-command drift guard`.

### Task 2: `classifyOutputLine` over REAL fixtures (3 grammars + rejections)
**Files:** Create `friction.go`; Test `friction_test.go` with real captured lines.
- [x] **Step 1: Failing tests** — one positive per grammar (G1 `[!]…[0m --no-verdict (or --force): skipping…`→Bypass; cinfo `==>…[0m skipping…per --no-judge (or --force)`→Bypass; G2 `[!]…[0m plan-quality gate bypassed (--force: x)`→Bypass{force-only}; G3 `[!]…[0m ⚠️ --no-validate: SKIPPING…`→Bypass); one refusal per shape incl comma/slash forms; and the REJECTIONS: the `:219` warmup line→none, a `cwarn(stderr, "--no-plan-check (or --force):` source line→none, a `17\t==> …` cat-n line→none, a `%s#%s is already status`→none.
- [x] **Step 2: Run** → FAIL → **Step 3: Implement** `classifyOutputLine` (ANSI strip, require `[0m ` runtime reset or the G2 shape, match per-`GateSig` `ackRE`/`refusalRE`, reject contamination markers + `warmupTrap`).
- [x] **Step 4: Run** → PASS → **Step 5: Commit** — `#172 M1: classifyOutputLine (3 ACK grammars, warmup+source rejection)`.

### Task 3: `SdlcInvocation` from anchored calls + Edit/Write capture + output linkage
**Files:** Modify `session.go` (`classifyToolUse` +`KindFileEdit`; `parseEvents` attach linked output to all `KindSDLCPrompt`); `friction.go` builder; Tests.
- [x] **Step 1: Failing tests** — `Bash(sdlc close --no-atlas --issue 173)` + its linked `tool_result` → `SdlcInvocation{verb:close, issueID:"173", output:<ack>}`; `Edit`/`Write`/`MultiEdit` → `KindFileEdit`; `sdlc close --help` → `isHelp`.
- [x] **Step 2–4:** implement — note `parseEvents` currently parses the verdict from close/mclose output then **discards the raw text** (`session.go:216`); extend it to (a) link results to EVERY `KindSDLCPrompt` and (b) **retain the raw output** on the event for `classifyOutputLine`. Per-session golden unchanged; PASS.
- [x] **Step 5: Commit** — `#172 M1: SdlcInvocation (anchored, issue-keyed) + Edit/Write capture + output linkage`.

### Task 4: whole-corpus walk + `detectGateEvents` + aggregate + render (Claude)
**Files:** `friction.go`, `processmanual.go`; Tests + temp-dir corpus test.
- [x] **Step 1: Failing tests** — anti-contamination test (a); warmup-rejection test (b); `enumerateAllTranscripts` over temp `<slugA>`/`<slugB>` incl a `-worktree-ariadne-` slug (→ labeled `ariadne`) and a `-private-tmp-` slug (→ excluded); `aggregate` per-gate + per-repo; `detectGateEvents` sets `observability`.
- [x] **Step 2–3:** implement Claude glob, `detectGateEvents`, `aggregate`, `renderFrictionReport` (markdown + `--json`), `--friction-report` + `--session` conflict guard.
- [x] **Step 4: Run** → PASS; smoke over the real corpus; the clean ranking should reproduce judge≈65-dominant, no-verified 0. Record clean per-gate + per-repo numbers in the Log; note the peer-vs-ariadne split.
- [x] **Step 5: Commit** — `#172 M1: whole-corpus bypass measure, anti-contamination + repo labeling (claude)`.

**M1 close:** `sdlc milestone-close --issue 172 --milestone M1`.

---

## Chunk 2: M2 — refusal→retry + firing-order detectors (Claude)

### Task 5: `detectRefusalRetries` (per-gate refusal sigs, flag-omitted best-effort)
- [ ] **Step 1: Failing tests** — `[close Refusal{no-atlas}, …, close Bypass{no-atlas}]` in one transcript → `RefusalRetry{Resolved:true}`; a `no-verdict` comma-form refusal never retried → `Resolved:false`; a merge `no-judge` publish-gate refusal (flag not named) paired by verb+context → `observability:flag-omitted`; the `:219` warmup must NOT produce a refusal.
- [ ] **Step 2–4:** implement (pair a `Refusal` with the next same-verb+issue invocation; carry `observability`); PASS.
- [ ] **Step 5: Commit** — `#172 M2: refusal→retry pairing (per-gate sigs)`.

### Task 6: `detectFiringOrder` (per-issue, iteration-aware)
- [ ] **Step 1: Failing tests** — `close` before any `change-code` for issue N → anomaly; **`milestone-close`→`change-code` (same issue) → none**; **`start-plan` re-run after `change-code` → none**; **`close`→`change-code` after a REWORK → none**; cross-issue interleave (`claim A; close A; claim B; start-plan B`) → none (keyed per issue); **`merge` (no `--issue`) after `close #N` in the same segment → attributed to N; a `change-code #N` after N was merged → anomaly**; a `merge` with no preceding `--issue` context → `unattributed`, not flagged; `Skill(writing-plans)` after a `KindFileEdit` → `skill-late`.
- [ ] **Step 2–4:** implement `workflowOrder` (AGENTS.md-sourced table, cited) + per-issue max-stage tracking with the legal-loop triggers; PASS.
- [ ] **Step 5: Commit** — `#172 M2: firing-order (per-issue, iteration-aware)`.

### Task 7: fold both into the report
- [ ] Extend `FrictionReport`/render (+ JSON) with refusal→retry (resolution rate, observability caveats) + firing-order; update golden; `go test ./cmd/sdlc/...` PASS; smoke + Log. Commit — `#172 M2: report renders refusal→retry + firing-order`.

**M2 close:** `sdlc milestone-close --issue 172 --milestone M2`.

---

## Chunk 3: M3 — codex coverage (net-new Go parser from the atlas spec)

### Task 8: `codexMetaKind` + `parseCodexEvents`
**Files:** Create `codex.go`; Tests + real `testdata/codex-*.jsonl`. (Note: `transcripts/codex.go:69 codexCWDFromBytes` is cwd-only + unexported — share only the "iterate to first `session_meta`" loop shape; extract `forked_from_id`/`parent_thread_id`/`agent_nickname` net-new, review M1.)
- [ ] **Step 1: Failing tests** from the atlas spec — `codexMetaKind`→fork-replay/sub-agent/root; `parseCodexEvents` maps a `function_call` running `sdlc <verb>` → `SdlcInvocation`, derives Bypass/Refusal from the plain-string `function_call_output.output` via the SAME `classifyOutputLine`; FIRST `session_meta`; skip fork-replay, keep sub-agent.
- [ ] **Step 2–4:** implement; PASS.
- [ ] **Step 5: Commit** — `#172 M3: Go codex parser (atlas-spec-derived)`.

### Task 9: wire codex into the walk + cross-language golden test
- [ ] **Step 1: Failing tests** — walk over 1 Claude + 1 codex transcript each with a `--no-judge` bypass → both counted, agent-tagged (codex adds e.g. no-reclose-guard 30); cross-language golden (Go decisions == Python `agent_codex.py`/`normalize.py` on a shared fixture, spec-derived snapshot).
- [ ] **Step 2–4:** codex glob in `enumerateAllTranscripts` + dispatch; PASS; smoke over both-agent corpus; record per-agent split in Log.
- [ ] **Step 5: Commit** — `#172 M3: codex coverage in friction-report (both agents)`.

**M3 close:** `sdlc milestone-close --issue 172 --milestone M3`.

---

## Chunk 4: M4 — T2 triage + T3 coverage-gap read (analysis)

### Task 10: T2 — triage the top-bypassed gates
- [ ] Run `--friction-report`. Triage each top gate (judge 65, change-code pq/eq via --force, no-actual, no-atlas, no-verdict) as **workflow gap** (mis-designed/too costly → fix/relax) vs **legit escape hatch**, using refusal→retry + firing-order as evidence; note the `observability` caveats where a count is `force-only`. Confirm `no-verified`=0 left alone. Record verdicts + actions in `## Findings`; file follow-ups.

### Task 11: T3 — coverage-gap read
- [ ] From firing-order anomalies + refusal→retry loops + segment context, identify un-gated workflow moments that recur as errors. Record gaps + candidate gates/relaxations. Commit — `#172 M4: T2 triage + T3 coverage-gap findings`.

**Issue close:** `sdlc close --issue 172`.

---

## ARCH notes
- **ARCH-DRY:** `gateCatalog` single-sources the classifier, drift guard, and report rows; codex format from the one atlas spec (cross-language golden); one `classifyOutputLine` for claude + codex; reuses `process-manual`'s parser.
- **ARCH-PURE:** all entities/detectors pure (real-fixture tests, incl. rejection cases, no mocks); `enumerateAllTranscripts` the single IO seam.
- **ARCH-PURPOSE:** measures the whole spine (all ~13 gates, both agents, cross-repo) AND does the T2/T3 analysis; states observability limits honestly rather than reporting a falsely-complete count.
- **ARCH-SIMPLICITY:** firing-order is a narrow per-issue observed-sequence-vs-workflow detector, not a hook/timing model.
- **Root Cause:** anchoring + the signature catalog attack the real problem (discrimination); the two prior plan-review rounds forced this precision (v1 mis-located the signal, v2 mis-shaped it).

## Revisions

### 2026-07-14 — v3: built on an exhaustively-enumerated signature catalog

Two fresh-eyes plan-review rounds + a ground-truth enumeration reshaped this:
- **v1** assumed the signal was in a discarded `.stderr` field → empirically false (it's in the captured tool-result output; `extractStdout` already reads it). Problem is discrimination, not capture.
- **v2** affirmed the anchoring architecture but hand-picked fixtures that didn't match reality — the review caught that the canonical "refusal" fixture was actually the `printSemanticWarmup` success line (C1), refusal shapes aren't uniform (I1), the 8-gate catalog missed spine gates on other commands (I2), and firing-order lacked per-issue keying + legal loops (I3).
- **v3** is built on a source+corpus-verified **signature catalog** (three ACK grammars, per-gate refusal shapes, warmup traps, runtime-vs-contamination counts). Corrections: refusal keys on the exact per-gate tail (not the warmup); **all spine gates** across close/mclose/change-code/merge/push (operator-approved scope); the corrected runtime ranking is **judge 65-dominant** (not `--no-validate` 87, which was source contamination — runtime 3); **observability limits stated** (change-code silent no-ACK bypasses observable only via --force; flag-omitting merge/push refusals); firing-order **per-issue + iteration-aware** (mclose→change-code, start-plan re-runs, rework/reopen all legal); codex parser extracts fork fields net-new (can't reuse the cwd-only unexported `codexCWDFromBytes`); repo labels normalize ariadne worktrees + exclude temp dirs.

### 2026-07-14 — v3.1: third review verdict = BUILDABLE; precision fixes folded

The third fresh-eyes verification confirmed the architecture is sound and empirically reproduced the headline (judge ≈ 64 measured). All three Important precision fixes are now in the plan text:
1. **Refusal discriminator is ACK-only** — runtime refusals carry no `\x1b[0m ` reset (plain `\n`-joined / `die()` reset-at-end), so refusals discriminate by grammar-anchored `refusalRE` (`\d+` not `N`/`…`, colon-delimited counts) — else refusal→retry filters to zero. A rejection test feeds this catalog's own text + grep meta.
2. **`no-start` excluded** — 13 registered `--no-*`, but `no-start` (`claim.go:62`) is a non-gate on a sixth command; the drift guard carries a `{no-start}` allowlist. Catalog = 12 spine gates.
3. **merge/push have no `--issue`** — stage-5 attributed from segment context or recorded `unattributed`, never bucketed globally.

Minors folded: the G3 `no-validate` ACK regex tolerates the `⚠️ ` emoji + double-space; `parseEvents` must **retain** the raw output (today it parses the verdict then discards it); **push `no-validate` refusal IS flag-named** (the shared `validategate.go:82` cwarn fires for both merge and push — so push no-validate is observability `full`, not `flag-omitted`); the change-code `no-judge` ACK (`gate bypassed (--force:)`) is a `--force` override of a *failed* judge, distinct from the silent `--no-judge` skip (label both, don't conflate). Per-gate runtime counts are approximate (±a few) but the ranking + sum hold.

**Status: approved-ready for implementation** (`sdlc change-code` → estimate → M1 TDD). No Critical ever surfaced across the three rounds; the architecture (anchor to `Bash(sdlc <verb>)` + table-driven signature classifier + honest observability limits) was affirmed by rounds 2 and 3.

### 2026-07-14 — M1 execution deviations (recorded at boundary review)

M1 shipped (verdict FIX-THEN-SHIP; both Important fixed before crossing). Three
deviations from the as-planned M1 task shapes, all functionally sound — treat this
entry as authoritative over the Chunk-1 task text:

1. **No `cmd/sdlc/gates.go` / `AllGates()`; registrations stay literal.** Task 1
   planned to source the flag names from `AllGates()`. Delivered instead as a
   drift-guard enforced the *other* direction: `cmd/sdlc/gates_test.go` introspects
   each spine command's live-registered `--no-*` flags (via cobra) and diffs against
   `GateFlagsFor` — production registration code untouched, same lockstep guarantee,
   and it catches a real registration without a catalog edit.
2. **`parseEvents` not extended; a sibling scanner does the anchored scan.** Task 3
   planned to extend `parseEvents` to retain raw output on every `KindSDLCPrompt`.
   Delivered as `sdlcInvocations` (friction.go) — a pure sibling that scans + links
   by `tool_use_id`, sharing `rec` + `sdlcVerbRE`, and reads the tool_result
   **content-block** (more complete than `toolUseResult.stdout`: no-judge ACK 105× vs
   49×). Watch item (review §6): extract the shared scan-and-link core before M3 adds
   a codex sibling, so the two walkers can't drift.
3. **Edit/Write/MultiEdit → `KindFileEdit` deferred to M2**, where the firing-order
   skill-late detector actually consumes it (disclosed in the issue Log).

**Boundary-review fix (Important #1):** `GateStat` observability was per-flag
last-write-wins and mislabeled the headline gate (no-judge shown `flag-omitted` when
its 17 bypasses are full-observable close/mclose). Now keyed per **(command, flag)**
and derived from the gate's *intrinsic* caveat (`gateObs`), so the label is correct
regardless of which event type was seen — close/mclose no-judge `full`, change-code
`force-only`, merge/push no-judge `flag-omitted`. Test: `TestObservabilityPerCommand`.

**Deferred to M2 (review Minors):** dedupe gate events per invocation (the
`no-validate` refusal double-line: `validategate.go:82` cwarn + the die-wrapped
error); `sdlcVerbRE` misses `go run ./cmd/sdlc <verb>` dev invocations (footnote);
error (not silent-0) when zero transcripts enumerate; compound-command second-verb
anchoring; a `renderFrictionReport`/JSON shape test + `toolResultText` array-form test.
