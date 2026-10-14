Report token usage from local Claude and Codex transcripts, without contacting
or launching an agent. These are reported response counters; they are not an
account quota or invoice. Cache reads/writes are included in input, and reasoning
is included in output. Unsupported dimensions are null in JSON, not assumed zero.

The first rows are a fleet glance per agent: tokens in the last hour, wall-clock
burn over --burn-window (default 3h, range 1h–3h), and a calibrated account
projection when evidence supports one. These rows always cover the fleet's
recent window, independently of the detail window or --issue.

DETAIL
  --since 24h               default detail window
  --since 2026-10-14T00:00:00Z
  --issue 365               whole issue unless --since was explicitly supplied
  --issue 365 --brain-dir /path/to/brain
  --json                    schema_version 1, integer counters, explicit coverage

Issue usage follows the owning slot's claim→release intervals and explicitly
bound descendants. A reclaim starts a fresh interval. A completed issue ends
at its exact completion evidence commit. Worker, reviewer (rounds and attempts),
and total tokens appear separately; reviewer bindings take precedence so spend
is counted once. Active time is reported with its own provenance and window,
which deliberately excludes the verification tail that token usage includes.
The local tracker snapshot is read without fetching; refresh it separately when
needed. Repositories without a tracker retain a labelled cwd heuristic.

Reviewer sessions are tied to recorded per-attempt dispatch evidence, including
the disposable review root or exact captured session ID. Historical reviews
without those fields stay explicit coverage gaps. Unmapped transcripts still
appear in unknown-slot; ambiguity does not become guessed ownership.

CALIBRATION
  sdlc usage calibrate --agent codex --remaining 29276 --unit credits
  sdlc usage calibrate --agent claude --remaining 70 --unit percent

Record two observations of the same agent and unit. The meter's decrease is
calibrated against measured local tokens between those observations, across all
slots. One observation, a refill, zero activity, stale readings (7 days), or
incomplete spend evidence produces a status and null projection. A projected
time is an empirical local rate, not a provider limit. Percent must be 0–100;
credits must be finite and nonnegative.

Only calibrate writes: $XDG_STATE_HOME/sdlc/usage-meter.json, or
~/.local/state/sdlc/usage-meter.json. The bounded file retains two observations
per agent/unit, uses a context-aware lock and private atomic writes. Plain
usage has no cache and performs no tracker fetch or provider interaction.

The scan is streaming and bounded (60s, 16384 sources, 4MiB lines, 250000
response identities). Coverage names malformed/unsupported records, missing
baselines, resets, provisional responses, truncation, and excluded attribution.
A bound is a visible incomplete report; cancellation exits unsuccessfully.
