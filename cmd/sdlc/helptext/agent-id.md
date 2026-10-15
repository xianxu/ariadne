Print this session's commit trailer: `Agent: <agent>/<model>/<effort>` (#367).

Every agent-authored commit carries it, next to the human-facing
`Co-Authored-By:` line, and sdlc stamps it on the commits it makes. The values
are the names the agent's own transcript records (`claude/claude-opus-5-5/high`,
`codex/gpt-6-astra/medium`), so commits join `sdlc usage` and `sdlc actual`
without a mapping. An unknown part is `-`; with no agent session at all (the
operator in a plain shell) the value is `human`. History from before #387 says
`sdlc` for the same thing, and readers treat it as `human`.

Where the values come from: the agent from the session's environment (the same
signals the judge uses: PAIR_AGENT, CODEX_THREAD_ID, CLAUDECODE); the model and
effort from the newest records of the session's own transcript (Claude Code:
~/.claude/projects/*/$CLAUDE_CODE_SESSION_ID.jsonl; Codex: the rollout file named
by $CODEX_THREAD_ID), with CLAUDE_EFFORT as Claude's effort fallback.
Only the transcript's last 1 MiB is read, so a model whose newest record was
pushed out by larger later records reads as `-`. PAIR_AGENT names which agent
a session is, but alone (a pair slot's plain shell) it is not a session.
SDLC_AGENT_ID=<agent>/<model>/<effort> names it outright.

USAGE

  sdlc agent-id          # Agent: claude/claude-opus-5-5/high
  sdlc agent-id --json   # {"agent":…,"model":…,"effort":…,"trailer":…}

Copy the printed line into the commit: `git commit -m '…' --trailer 'Agent: …'`.
`sdlc close` lists the window's commits that lack it.
