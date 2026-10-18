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
brain capture repository. The date is the exclusive UTC end. An identical rerun
does nothing; changed bytes at that path are refused. The brain owner commits and
backs up the artifact under its capture charter; source attempts remain local
for 30 days. The ops TL links the path from its test-economy project.

The snapshot's physical suite statistics include attempts completed in the
seven-day window. Its headline ratio uses issues landed in that window and all
exactly attributable attempts from first enqueue through landing, including
attempts before the week. These populations can differ. The report separates
outside-week attribution, in-week attempts outside the landing cohort and
pre-enqueue attempts; absent or contradictory evidence makes the ratio null.
Named failed tests include an explicit unknown class. Slow packages have their
own top-five list and coverage count, distinct from shard durations.

Snapshot collection fetches main and tracker evidence. Git commands have a
30-second/8-MiB bound, landing logs a 64-MiB bound and attempts retain the
10,000-record limit. Unavailable or over-limit sources are reported as unknown
coverage. Snapshot publication always covers the discovered SDLC fleet so a
partial `--repo` run cannot claim the immutable weekly path.
