Report native usage from local Claude, Codex and Qoder transcripts, without contacting
or launching an agent. These are reported response tokens or native Qoder credits; they are not an
account quota or invoice. Cache reads/writes are included in input, and reasoning
is included in output. Unsupported dimensions are null in JSON, not assumed zero.

The first rows are a fleet glance per agent: native tokens or credits in the last hour, wall-clock
burn over --burn-window (default 3h, range 1h–3h), and a calibrated account
projection when evidence supports one. These rows always cover the fleet's
recent window, independently of the detail window or --issue.

DETAIL
  --since 24h               default detail window
  --since 2026-10-14T00:00:00Z
  --issue 365               whole issue unless --since was explicitly supplied
  --issue 365 --brain-dir /path/to/brain
  --json                    schema_version 2, nullable tokens, exact decimal credits, explicit coverage

Issue usage follows the owning slot's claim→release intervals and explicitly
bound descendants. A reclaim starts a fresh interval. A completed issue ends
at its exact completion evidence commit. Worker, reviewer (rounds and attempts),
and total native usage appear separately; reviewer bindings take precedence so spend
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
  sdlc usage calibrate --agent qoder --remaining 500 --unit credits

Record two observations of the same agent and unit. The meter's decrease is
calibrated against measured local tokens (Claude/Codex) or credits (Qoder)
between those observations, across all
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
baselines, resets, provisional responses, truncation, excluded attribution, and
unresolved reviewer spend. A launched review without usable transcript evidence
blocks complete coverage and the affected agent's calibrated projection.
Codex usage before a valid executing-session identity is skipped with incomplete
spend coverage; later headers cannot make that prefix safe retroactively. A
missing working directory affects attribution, separately from spend coverage.
A bound is a visible incomplete report; cancellation exits unsuccessfully.

Qoder reads ~/.qoder/projects/*/*.jsonl. Workspace metadata supplies slot
attribution, and the reported model (including auto) is preserved. Terminal
billable usage.credits is native spend; original_credits is not added. Streaming
snapshots and copied responses are deduplicated. Qoder placeholder token fields
are unsupported and stay null, including measured_dimensions. A measured zero
credit response is distinct from missing charge evidence.

Mixed reports expose null full totals for any unit unknown in an included
response. measured_dimensions and measured_credits retain only the known
subtotals and are null when that unit has no measured response. Text labels
these known subtotals; credits are never converted into tokens. Hourly, group,
issue worker/reviewer/total and unbound evidence preserve the same distinction.
