---
id: 000159
status: blocked
deps: [pair#104]
github_issue:
created: 2026-07-01
updated: 2026-07-06
estimate_hours:
started: 2026-07-06T16:45:44-07:00
---

# alt+shift+c continuation improvements

a couple of things. 

1. I think the restart stopped working. I want alt+shift+c to make a continuation, and then automatically restart with same config but a new session (pair or pair-dev), then continue on. this is a bug fix. 

2. if the draft pane at * position is not empty, it's content would be incorporated as part of next steps. automatically. this is a small new feature

## Problem

All the code for this lives in the **`pair` repo** (not ariadne, not parley.nvim). `alt+shift+c` → zellij `config.kdl:269-280` (`bind "Alt C"`) focuses the draft pane and calls `PairConfirmCompact()` (`pair/nvim/init.lua:3327`), which confirms then `send_to_agent(COMPACT_PROMPT)` (`pair/nvim/init.lua:3315`). `COMPACT_PROMPT` delegates BOTH steps to the agent: (1) write a continuation via the `pair-continuation` writer, then (2) *itself run* `pair continue <slug>`. Because the restart is step 2 of an NL prompt — agent judgment, not code — it is **non-deterministic**: if the agent writes the doc but skips `pair continue`, nothing restarts. That is the "restart stopped working" bug (the mechanism was never automatic).

## Spec

Decisions from brainstorm (2026-07-06):

1. **Home = `pair` repo.** Move this work to pair's tracker on resume (all code is there; keeps it beside the colliding #104 and its close review). Recreate as `pair#NNN`, close this ariadne issue as moved.
2. **Restart bug → make it deterministic in the writer.** After `pair-continuation` writes + commits the continuation, the writer itself triggers the park+restart using the slug it just wrote (the `pair continue` path) — removing agent judgment from the restart step. Restarts with the same config (pair vs pair-dev via the `PAIR_DEV` env riding the exec loop) and a new session.
3. **Feature #2 → fold draft into NEXT ACTION.** On compaction, if `draft-<tag>.md` (`pair/nvim/init.lua:471`, `draft_path_for_tag()`) is non-empty, incorporate its content — **minus comments** — into the continuation's `## NEXT ACTION`, automatically. ("`* position`" = the current draft file as a whole.)

## Done when

- alt+shift+c reliably: writes a continuation → **automatically** restarts the tag with same config (pair/pair-dev) + a new session → continues from the doc's NEXT ACTION — no agent step required for the restart.
- When the draft pane has content at compaction time, that content (sans comments) appears in the continuation's NEXT ACTION.

## Plan

Blocked on **pair#104** (single-pair-binary), claimed 2026-07-06, which rewrites the exact `pair restart`/`pair continue` call surface this depends on. Building first would just be rewritten. On unblock: move to pair, then implement in the writer (`pair/cmd/internal/continuationcmd/`, assembly at `continuation.go:73-124`) + the draft read (`draft_path_for_tag()`).

- [ ] (deferred until pair#104 lands) move to pair; deterministic writer-triggered restart; draft→NEXT-ACTION fold

## Log

### 2026-07-06

Claimed + brainstormed. Traced the full flow (all in `pair`): binding `config.kdl:269-280` → `PairConfirmCompact` (`nvim/init.lua:3327`) → `COMPACT_PROMPT` (`nvim/init.lua:3315`, agent-delegated restart) → agent writes continuation via `pair-continuation` (`cmd/internal/continuationcmd/`) + runs `pair continue` → `runCompaction` (`cmd/internal/launcher/compaction.go`) → outer relaunch loop (`createflow.go`). Root cause of #1: restart is agent-judgment, not code. Decided: move to pair, writer-triggers-restart, draft-sans-comments → NEXT ACTION (see Spec). **Blocked on pair#104** (rewrites this call surface) per operator — wait for the single-pair-binary to land, then move + implement.

### 2026-07-01
