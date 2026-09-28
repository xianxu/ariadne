---
id: '000159'
status: wontfix
started: 2026-07-06T16:45:44-07:00
created: 2026-07-01
updated: 2026-07-06
---

# alt+shift+c continuation improvements

## Problem

All the code for this lives in the **`pair` repo** (not ariadne, not parley.nvim). `alt+shift+c` → zellij `config.kdl:269-280` (`bind "Alt C"`) focuses the draft pane and calls `PairConfirmCompact()` (`pair/nvim/init.lua:3327`), which confirms then `send_to_agent(COMPACT_PROMPT)` (`pair/nvim/init.lua:3315`). `COMPACT_PROMPT` delegates BOTH steps to the agent: (1) write a continuation via the `pair-continuation` writer, then (2) *itself run* `pair continue <slug>`. Because the restart is step 2 of an NL prompt — agent judgment, not code — it is **non-deterministic**: if the agent writes the doc but skips `pair continue`, nothing restarts. That is the "restart stopped working" bug (the mechanism was never automatic).
