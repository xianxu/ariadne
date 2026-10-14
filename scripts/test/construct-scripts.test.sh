#!/usr/bin/env bash
# construct-scripts.test.sh — run every construct/scripts/test/*.test.sh (#362).
# The one runner for those tests: the CI merge check (20-ci-workflow-tests) runs
# it as a scripts/test/ script, and `make test` runs it through the Go test beside
# it (scripts/test/scripts_test.go), so the exclusion list below has one home.
# Each test runs under the bash on PATH (macOS /bin/bash 3.2 locally) in its own
# process; a failing or missing test fails the run, naming it.
#
# Excluded by name, each with its reason (none today; a test that needs a live
# environment or is too slow belongs here, never silently skipped):
EXCLUDED=""
#
# Run: bash scripts/test/construct-scripts.test.sh   (exit 0 = all pass)
set -uo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd -P)" || exit 1
tests=("$ROOT"/construct/scripts/test/*.test.sh)
[ -e "${tests[0]}" ] || { echo "construct-scripts: no construct/scripts/test/*.test.sh found under $ROOT" >&2; exit 1; }
pass=0 fail=0 skipped=0 failed=""
for t in ${tests[@]+"${tests[@]}"}; do
	name="$(basename "$t")"
	case " $EXCLUDED " in
		*" $name "*) skipped=$((skipped + 1)); echo "skip $name (excluded above)"; continue ;;
	esac
	if out="$(bash "$t" 2>&1)"; then
		pass=$((pass + 1)); echo "ok   $name"
	else
		fail=$((fail + 1)); failed="$failed $name"
		echo "FAIL $name"; printf '%s\n' "$out" | tail -20 | sed 's/^/    /'
	fi
done
printf 'construct-scripts: %d passed, %d failed, %d excluded\n' "$pass" "$fail" "$skipped"
[ "$fail" -eq 0 ] || { echo "construct-scripts: failed:$failed" >&2; exit 1; }
