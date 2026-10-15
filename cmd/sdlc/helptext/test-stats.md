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
