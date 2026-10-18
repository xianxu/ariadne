Read completed local suite attempts from the repositories in this workspace's
sibling fleet. This command does not run tests, post statuses, or contact
GitHub. It works from a numbered slot by resolving the Git-verified primary
repository and then inspecting each sibling's Git common directory.

`sdlc test-stats` defaults to the last seven days. Use `--since 12h` or
`--since 30d` for a different window, `--repo pair` for one repository, and
`--json` for machine-readable numbers and null missing measurements.

Wall and queue p50/p90 use nearest-rank percentiles over attempts with those
measurements. Slow package and shard rows rank by p90; package groups are kept
distinct from individual packages. A confirmed flake is a named failure
followed by a passing suite on the same exact head. Failures during a diagnostic
main-baseline suite are listed separately as suspected. The confirmed flake
rate denominator includes only complete per-attempt records.

Older per-head logs are partial: they cannot recover queue time or overwritten
retries, and some formats do not report total wall time. The command labels
missing metrics as n/a and reports unreadable or malformed records as coverage
warnings rather than treating them as successful runs.
The reader caps each repository at 10,000 records and 64 MiB of record JSON;
if a cap is reached, it reports the omitted coverage.

`sdlc test-stats snapshot` publishes a versioned, immutable JSON measurement at
`<brain>/memory/work/devinfra/test-economy/weekly/YYYY-MM-DD.json`. It defaults
to the latest completed Monday 00:00 UTC boundary; `--end 2026-10-19T00:00:00Z`
selects a fixed seven-day interval, and `--brain-root` overrides the sibling
brain capture repository. The date is the exclusive UTC end. A rerun returns an
existing validated schema-v2 artifact without recollecting or changing its
`observed_at`; an incompatible or corrupt file is refused. The writer keeps its
byte-wise immutable check for concurrent publication. The brain owner commits and
backs up the artifact under its capture charter; source attempts remain local
for 30 days. The ops TL links the path from its test-economy project.

The snapshot's physical suite statistics include attempts completed in the
seven-day window. Its headline ratio uses issues landed in that window and all
exactly attributable attempts from first enqueue through landing, including
attempts before the week. These populations can differ. The report separates
outside-week attribution, in-week attempts outside the landing cohort and
pre-enqueue attempts; absent or contradictory evidence makes the ratio null.
The ratio also stays null for an issue whose creation day is not strictly after
the retained source-history floor: the 30-day attempt/event cutoff, the
first retained landing event and the first retained structured suite attempt.
An older issue could have earlier admission or
suite records that were swept, even when its latest ticket looks complete.
Each repository's headline puts the ratio beside its covered and observed
landed-issue counts, unknown count and three most common coverage reasons.
An incomplete source makes the observed count a lower bound and the covered
count zero; the snapshot never presents a partial numerator as a full ratio.
Named failed tests include an explicit unknown class. Slow packages have their
own top-five list and coverage count, distinct from shard durations.

Snapshot collection fetches main and tracker evidence. Git commands have a
30-second/8-MiB bound, landing logs a 64-MiB bound and attempts retain the
10,000-record limit. Unavailable or over-limit sources are reported as unknown
coverage. Snapshot publication always covers the discovered SDLC fleet so a
partial `--repo` run cannot claim the immutable weekly path.

Schema v2 adds `resources`: metadata-only byte totals for Pair's declared
backup-source families, the ten largest Couch ledgers, per-repo suite wall
durations, Go build-cache size, AgentScratch free/total bytes, and unprivileged
disk write/wear gauges. Pair's `sources.json` supplies the family definitions;
`--resource-inventory`, `--resource-pair-data`, `--resource-scratch`, and
`--resource-disk` select the source paths when defaults do not fit. The section's
`observed_at` is the actual UTC collection time, distinct from the suite window.
Point-resource deltas and growth alerts require complete observations 6–8 days
apart; otherwise they are unknown with the actual interval. Suite-duration
deltas use the fixed weekly windows. Missing tools, source paths, prior weeks,
or readable gauges remain named unknowns or first-week baselines, never zeros.
No scheduler is installed for this command; the TL runs the completed-week
snapshot after each Monday 00:00 UTC boundary.
