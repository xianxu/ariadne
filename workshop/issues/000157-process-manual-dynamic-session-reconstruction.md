---
id: 000157
status: working
deps: [ariadne#153]
github_issue:
created: 2026-07-01
updated: 2026-07-01
estimate_hours: 2.6
started: 2026-07-01T16:06:44-07:00
---

# process-manual dynamic session reconstruction

Split out of ariadne#153 (was its M3). #153 delivers the **static catalog** (M1) +
**judge-prompts-as-markdown** (M2); this issue is the **dynamic** counterpart: given a
session transcript, reconstruct which catalogued injection points actually fired.

## Problem

The M1 process manual (`sdlc process-manual`) is the *static* catalog — what CAN be
injected. It can't say which injection points actually **fired** in a given session, in
what order. The parley's second output: mine a session transcript → the ordered stream
of fired injections, matched to the M1 catalog, so a human can see the process as it
actually ran (`workshop/parley/2026-07-01.10-47-05.962_agentic-process-documentation-strategy.md`).

## Spec

**Feasibility is grounded** (M3 exploration digest, this issue's Log). Bottom line:
the deterministic core is real and **Go-native**; the ambitious "anomaly-first /
injected-but-ignored" framing is **deferred** as fuzzy.

### Feasible core (build this)

`sdlc process-manual --session <jsonl>` (or `--session current`) parses a Claude session
JSONL (`~/.claude/projects/<slug>/*.jsonl`) → an **ordered, timestamped list of fired
injection events**, each matched to its M1 `Kind`:

- `Bash` `sdlc <verb>` calls → `KindSDLCPrompt` / `KindHelpText` (recover the review
  verdict for `close`/`milestone-close` from the Bash **stdout** — the forked review's
  output is streamed back into the parent result).
- `Skill` tool_use (`.input.skill`) → `KindSkill`.
- `Read`/cat of `workshop/lessons.md` → `KindLessons`.

Reuse the **60-min-gap / `away_summary` segmentation** (port the constant + tool-input
extraction table from `introspect`'s `segment_text.py`; **don't shell to Python** — no
structured output to reuse, and M3 must match against M1's in-process Go `InjectionSource`
catalog → **ARCH-DRY** = one process). Parse is ~15 lines of `encoding/json`.

### Two hard limits (document as stated gaps, don't fight them)

1. **`agents-chain` (AGENTS/CLAUDE.md) + `memory` are session-start *system-prompt*
   injections that never appear in the transcript** — verified 0/68 local sessions
   contain the injection wrappers. M3 can only assert they were *available* (from the M1
   catalog), never that they fired or were ignored.
2. **Forked review *prompts* aren't in the main JSONL** — only their *output* (via Bash
   stdout) + the durable sidecar. The prompt template lives in `judge/prompts/*.md` (M2).

### Deferred (NOT this issue)

Anomaly / "injected-but-ignored" detection. `agents-chain`/`memory` firing is
undetectable; "offered-but-never-fired" (skill_listing − Skill tool_use) is a weak signal;
"followed the guidance?" is an LLM-judge problem, not a parse. A later issue can add a weak
set-difference anomaly and/or an LLM-judge semantic pass — a separate, fuzzier problem.

### Sidecar correlation (for close/milestone-close review linking)

Deterministic path from issue#+milestone (`reviewsidecar.go:37`):
`workshop/plans/NNNNNN-<slug>-{close-review|m<x>-review}.md`, archived to
`workshop/history/` on issue close (search both). Re-runs append a `## Re-review — <ts>`
section to the same file; the Bash call's timestamp/`window` disambiguates which section.

### Annotation surface

Same as M1: markdown report with live links + `🤖[]` slots (composes with `xx-fix`).

## Done when

- `sdlc process-manual --session <jsonl|current>` emits, in timestamp order, each fired
  injection event mapped to its M1 `Kind` (sdlc verbs + verdicts, skills, lessons reads),
  segmented on the 60-min/away_summary boundary.
- The two hard limits (agents-chain/memory invisibility; forked prompts) are documented as
  stated gaps in the output + skill.
- Go-native (no Python shell-out); matches against the M1 `InjectionSource` catalog.

## Plan

- [ ] Coarse — detailed plan at `start-plan` via `superpowers-writing-plans`. Shape:
      Go JSONL parser (tolerate unknown record `type`s) → ordered event stream →
      segmentation → match to M1 catalog Kinds → linked markdown report. Reuse M1's
      `InjectionSource` + `renderManual`.

## Estimate

Single-pass atomic work (no `Mx` tags → one `sdlc close`, one review boundary),
shaped like #153's M1 (a pure Go core + wiring + atlas + review — the closest
calibrated sibling, same author). Design is largely spent (grounding digest +
reviewed plan), reflected in the design hours. *Produced via
`brain/data/life/42shots/velocity/estimate-logic-v3.1.md` against
`baseline-v3.1.md`. Method A only.*

```estimate
model: estimate-logic-v3.1
familiarity: 1.0
item: greenfield-go-module   design=0.5 impl=1.1
item: smaller-go-module      design=0.2 impl=0.4
item: atlas-docs             design=0.0 impl=0.1
item: milestone-review       design=0.0 impl=0.2
design-buffer: 0.15
total: 2.6
```

Derivation: `session.go` (pure core — tolerant `parseEvents` + `classifyToolUse`
+ `segmentEvents` + `renderSessionReport`, all fixture-tested) = greenfield
module, tracking #153 M1's greenfield row (design=0.8 impl=1.2) but smaller
(single file, design mostly done); `locateSessionJSONL` + `--session` cobra
wiring = smaller-go-module; atlas relink = atlas-docs; the single close-review =
milestone-review. Design-buffer 0.15 (thorough reviewed plan).

## Log

### 2026-07-01 — grounding digest (M3 feasibility exploration)

Verified against 68 local sessions + `introspect` scripts + live sidecars.

- **JSONL schema:** flat event log, one JSON/line, ordered; each record has
  `type`/`timestamp`/`uuid`/`sessionId`/`cwd`. `assistant` records carry `tool_use` blocks
  in `.message.content[]` (`{id,name,input}`); matching `tool_result` is a following `user`
  record linked by `tool_use_id`, and Bash/Read also populate top-level `.toolUseResult`
  (`.stdout` / `.file`). Parser must tolerate unknown `type`s (newer sessions add
  bookkeeping types).
- **Load-bearing jq (proof of feasibility):**
  ```
  jq -rc 'select(.type=="assistant") | .timestamp as $ts | .message.content[]?
    | select(.type=="tool_use")
    | select((.name=="Skill")
        or (.name=="Bash" and (.input.command|test("(^|[^a-zA-Z])sdlc ")))
        or (.name=="Read" and (.input.file_path|test("lessons\\.md|SKILL\\.md|AGENTS\\.md|CLAUDE\\.md|MEMORY\\.md"))))
    | [$ts, .name,
       (if .name=="Skill" then .input.skill
        elif .name=="Bash" then ((.input.command|capture("sdlc (?<v>[a-z-]+)") // {v:"?"}).v)
        else (.input.file_path|split("/")|.[-2:]|join("/")) end)] | @tsv' F.jsonl
  ```
- **Forked visibility:** the review **verdict + full findings body IS in the main JSONL**
  (streamed into the `sdlc close`/`milestone-close` Bash stdout); the injected **prompt** is
  not (it went to the forked `claude -p`). Corrects the parley's assumption.
- **introspect reuse:** ~30% as *design* (file-discovery loop, segmentation, in-order walk +
  per-tool input table in `segment_text.py:190-278`), ~0% as code (Python vs Go). `normalize.py`
  emits aggregates (counts/sets) — loses the order + command text M3 needs; `segment_text.py`
  renders truncated text, not structured records.
- **Session-start injections invisible:** 0/68 sessions contain the CLAUDE.md/AGENTS/memory
  injection wrappers — the harness doesn't serialize the system prompt. `# Constitution`/`MEMORY.md`
  appear only when a file was explicitly Read/cat'd mid-session. → the two hard limits above.
- **Key files:** `cmd/sdlc/internal/processmanual/source.go` (M1 catalog, 6 Kinds),
  `reviewsidecar.go:37` (sidecar path), `introspect/scripts/{segment_text.py,normalize.py}`.

Untracked note: `sdlc issue new` could not auto-sync this to main (main not checked out on
the 000153 branch) — needs to land on main separately, NOT bundled into #153's PR.
