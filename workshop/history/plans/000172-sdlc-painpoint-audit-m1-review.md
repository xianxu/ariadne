# Boundary Review — ariadne#172 (milestone M1)

| field | value |
|-------|-------|
| issue | 172 — sdlc painpoint audit |
| repo | ariadne |
| issue file | workshop/issues/000172-sdlc-painpoint-audit.md |
| boundary | milestone M1 |
| milestone | M1 |
| window | 5c005147dbe18556a97dddf2634a5bcc9e48d525^..HEAD |
| command | sdlc milestone-close --issue 172 --milestone M1 |
| reviewer | claude |
| timestamp | 2026-07-14T12:27:00-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

All verification is done — tests pass, the catalog matches every real emit site, and I reproduced the headline live (no-judge 17, no-verified 0, brain 19/pair 15/ariadne 3). The live run also confirmed the one substantive defect I suspected: the top-ranked gate's observability label is wrong. Writing the review.

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

M1 delivers what it claims, and the claims check out under independent verification: I validated all 16 catalog signatures against the actual emit sites in `close.go`/`milestoneclose.go`/`changecode.go`/`merge.go`/`push.go`/`validategate.go` (every AckPat and RefusalPat matches a real runtime string, and the `:219` warmup twin is correctly distinguished from the `:1093` real refusal), ran the full test suite (green, including the drift guard against live cobra registrations), confirmed helptext carries none of the signature strings (so the deliberate don't-skip-`--help` choice is safe), and re-ran `--friction-report` over the real corpus — it reproduces the Log's headline exactly (no-judge 17 dominant, no-verified 0, bypasses concentrated in brain 19/pair 15 vs ariadne 3). Nothing blocks the boundary. What keeps this from SHIP: the report's observability column — the design's honesty pillar — mislabels its own top-ranked gate (live output shows `no-judge | 17 | 12 | flag-omitted ⚠️` when the 17 bypasses are mostly fully-observable close/mclose ACKs), and the durable plan was not reconciled with three execution deviations (all M1 step checkboxes still unticked, no Revisions entry).

**1. Strengths**

- **The catalog is ground truth, verified.** Every one of the 16 `GateCatalog` rows (`gatesig.go`) matches the real source emit site — e.g. no-actual ACK ↔ `close.go:367`, its refusal tail `only when measurement is not applicable` ↔ `close.go:1093` while rejecting the byte-adjacent warmup at `:219`; no-verified's `genuinely no behavior` ↔ `close.go:1194` vs the warmup's `genuinely nothing`. For a measurement instrument this is the property that matters most, and it holds.
- **Anti-contamination works and is pinned by load-bearing tests.** `TestAggregateAntiContamination` and the real-fixture `TestClassifyOutputLine` rejection cases (warmup, source `cwarn(`, cat-n reads, format verbs, wrong-verb ACK) cover exactly the bug class this instrument could ship. The live run confirms the effect: raw-grep no-judge ~64 → anchored 17.
- **The drift guard has real teeth** (`gates_test.go:20`): it introspects each spine command's *live* registered `--no-*` flags via cobra and diffs against `GateFlagsFor` — a new gate cannot land without the audit noticing, without touching production registration code.
- **ARCH-PURE shape is clean:** `classifyOutputLine`/`aggregate`/`repoLabel`/`sdlcInvocations` are pure over bytes with real-fixture tests; `enumerateClaudeTranscripts`+`RunFrictionReport` are the single thin IO seam (injectable root, tested with a temp tree).
- **Honest residual disclosure** in the Log (~2/19 no-judge ACKs unclassified; lossy slug→repo) — the right posture for a measurement tool.

**2. Critical findings** — none.

**3. Important findings**

- **Observability label collapses per-flag, mislabeling the headline gate** — `friction.go` `aggregate()`: `obs[ev.Gate] = ev.Observability` is last-write-wins keyed by flag alone, but `no-judge` spans five commands with three different observabilities (close/mclose cinfo = `full`, change-code G2 = `force-only`, merge/push refusals = `flag-omitted`). Live report output: `| no-judge | 17 | 12 | flag-omitted ⚠️ |` — the 17 bypasses are predominantly fully-observable close/mclose ACKs, and change-code's force-only caveat (the one the footnote exists for) disappears entirely. This undermines the design's stated pillar ("state observability limits honestly", ARCH-PURPOSE) exactly where it matters — the T2 triage of the top gate will read a wrong caveat. Fix sketch: key `GateStat` by (command, flag) as the plan's master table and the drift guard already do — or keep the per-flag row but render observability as the set of per-command observabilities (`full/force-only/flag-omitted`), and split no-judge's counts per command.
- **Durable plan not reconciled with execution (plan-vs-code drift; #97 lesson class).** All M1 task step checkboxes in `workshop/plans/000172-sdlc-painpoint-audit-plan.md` are still `- [ ]` although the issue's Plan ticks M1, and there is no M1 `## Revisions` entry for three real deviations: (a) Task 1's `cmd/sdlc/gates.go` `AllGates()` + "point registrations at AllGates()" was **not** built — the delivered drift guard introspects registered flags instead (arguably a better design: production untouched, same invariant — but the plan still claims the other shape); (b) Task 3's "extend `parseEvents` to attach + RETAIN the raw output on every KindSDLCPrompt" was replaced by the sibling `sdlcInvocations` scanner, leaving `parseEvents` untouched; (c) Edit/Write→`KindFileEdit` deferred to M2 (disclosed in the issue Log, not in the plan). Per the review contract, plan-table rows claiming what code doesn't deliver need a Revisions entry so the plan stops claiming it; I grade this Important rather than Critical because the deviations are functionally sound and (c) is disclosed in the Log.

**4. Minor findings**

- One `no-validate` refusal emits **two** matching lines (`validategate.go:82` cwarn + the die-wrapped error from `:86`, both matching `instance-conformance gate: \d+ nonconforming`) → refusal double-count; will skew M2's refusal→retry resolution rates if not deduped per invocation.
- `sdlcVerbRE` (`session.go:387`) misses dev-style invocations (`go run ./cmd/sdlc <verb>`, `bin/sdlc <verb>`) — deflates counts precisely in the repo that dogfoods sdlc, i.e. the "ariadne 3" side of the concentration headline. Empirically small (I probed the ariadne corpus; the go-run hits are almost all quoted plan text), but it belongs in the stated-limitations footnote.
- Silent zero: missing/unreadable `~/.claude/projects` → "0 transcripts scanned", exit 0 — the repo's own #68 lesson (misinvocation must not look like a real empty answer). Cheap: error when zero transcripts enumerate.
- Compound commands (`sdlc close … && sdlc merge …`) anchor only the first verb; the second verb's gate lines are rejected by verb-anchoring (conservative under-count).
- `--json` without `--friction-report` is silently ignored; `IsHelp` is parsed but unused (keep for M2 or drop).
- No test covers `toolResultText`'s array-of-`{text}` legacy form, nor `renderFrictionReport`/JSON output shape.

**5. Test coverage notes**

`go test ./cmd/sdlc/... ./cmd/sdlc/internal/processmanual/` passes. Coverage of the classifier is strong and uses real captured fixtures including the rejection cases — the exact trap the plan's review history warned about (fabricated fixtures). Gaps: the two known-unclassified no-judge ACK shapes have no fixture yet (capture them when M2 tightens the tail); the aggregation/rendering layer (observability labeling, JSON) is untested — which is exactly where the one real defect above lives, so add a test asserting a mixed-observability flag renders correctly when fixing it.

**6. Architectural notes for upcoming work**

- **ARCH-DRY: pass, with one watch item.** `GateCatalog` genuinely single-sources the classifier, drift guard, `SpineVerbs`, and report rows; the codex format lives once in `atlas/workflow/introspect.md` with the fork-skip-40-not-119 trap called out for M3's Go reader. Watch item: `sdlcInvocations` (friction.go) duplicates `parseEvents`' scan/tool_use_id-linkage loop shape (shared `rec` + `sdlcVerbRE` mitigate this; the code comment acknowledges it) — the repo's "N parallel walkers drift silently" lesson applies; consider extracting the shared scan-and-link core before M3 adds a codex sibling.
- **ARCH-PURE: pass.** Pure classification/aggregation core, one injectable IO seam, real-fixture tests, no mocks.
- **ARCH-PURPOSE: pass for M1's scope, flagged on the honesty sub-goal.** The clean command-anchored measure replaces the contaminated grep and reproduces the headline over the real corpus (verified live); refusal→retry, firing-order, and codex are genuine milestone sequencing, not deferred-purpose games. The one under-delivery is the observability mislabel (Important #1) — the "state limits honestly" clause is the part of the purpose the current render gets wrong.
- **Docs gates:** atlas updated in-window (`sdlc-binary.md` friction-audit section; index already links it) — pass. README documents no per-flag sdlc surface at all, so no README obligation for `--friction-report` — pass.
- For M2: dedupe gate events per invocation (the no-validate double-line), and decide whether firing-order's segment-context attribution should also cover the compound-command case.

**7. Plan revision recommendations**

Append one dated `## Revisions` entry to `workshop/plans/000172-sdlc-painpoint-audit-plan.md` and tick the M1 Task 1–4 step boxes:

> `2026-07-14 — M1 execution deviations (recorded at boundary review):` (1) No `gates.go`/`AllGates()`; flag registrations remain literal — the drift-guard invariant is enforced the other direction, by `gates_test.go` introspecting each spine command's registered `--no-*` flags against `GateFlagsFor` (production untouched, same lockstep guarantee). (2) `parseEvents` was not extended to retain raw output; a sibling pure scanner `sdlcInvocations` (friction.go) does the anchored scan + `tool_use_id` linkage instead, sharing `rec` + `sdlcVerbRE`. (3) Edit/Write/MultiEdit→`KindFileEdit` deferred to M2, where the skill-late detector consumes it. Also record the boundary-review fix: `GateStat` observability was per-flag last-wins and mislabeled no-judge; now keyed/rendered per (command, flag) [or per-command set].
