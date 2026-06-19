# workshop/benchmarks/

Repo-local artifacts for the `sdlc bench` multi-agent benchmark harness (#119).
The harness itself is base-layer (`cmd/sdlc`); the task library and results here
are this repo's own.

## Layout

- `tasks/<id>.md` — **frozen tasks** (datatype `benchmark-task`): an immutable
  spec snapshot + `base_sha` + grading rubric, frozen from a live issue. Agents
  branch from `base_sha` and never merge, so a task replays identically forever.
- `runs/<task>-<runid>.md` — **run records**: one per (task × agent-set ×
  execution). Per agent: branch, captured metrics, objective scorecard, and the
  blind judge + operator verdicts. Pins each agent's CLI version.
- `leaderboard.md` — generated; per-`(agent, version)` objective-metric
  distributions + head-to-head win-rates per subjective dimension.

## Flow

```
sdlc bench freeze --issue N            # snapshot a live issue → tasks/<id>.md
sdlc bench run --task <id> --agents claude,codex   # fan out worktrees, capture
sdlc bench grade --run <runid>         # measure (Stage A) + blind judge (Stage B)
sdlc bench review --run <runid>        # fold in operator rankings, de-anonymize
sdlc bench leaderboard                 # aggregate → leaderboard.md
```

See `sdlc bench --help` and the `benchmark-task` datatype
(`construct/datatype/benchmark-task.md`).
