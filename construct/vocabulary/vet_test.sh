#!/bin/sh
# M1 gate (ariadne#122): the issue vocabulary model validates, a broken model is
# rejected, and the EXPORT carries `categories` + `lifecycle`. The export check is
# load-bearing: CUE `#`-definitions don't `cue export`, so this guards the concrete
# data sdlc actually consumes (caught in plan review).
set -e
dir="$(dirname "$0")"

cue vet "$dir/issue.cue" || { echo "FAIL: valid model did not vet"; exit 1; }

if cue vet "$dir/testdata/issue_invalid.cue" 2>/dev/null; then
  echo "FAIL: invalid model passed vet"; exit 1
fi

json="$(cue export "$dir/issue.cue" --out json)"
echo "$json" | grep -q '"categories"' || { echo "FAIL: categories not in export"; exit 1; }
echo "$json" | grep -q '"lifecycle"'  || { echo "FAIL: lifecycle not in export";  exit 1; }

echo ok
