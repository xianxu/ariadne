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

# verdict (ariadne#147): atomic noun — no lifecycle; guard the concrete `categories`
# the binding consumes reaches the export (the `#Emitted`/`#Token` defs don't).
cue vet "$dir/verdict.cue" || { echo "FAIL: valid verdict model did not vet"; exit 1; }
vjson="$(cue export "$dir/verdict.cue" --out json)"
echo "$vjson" | grep -q '"categories"' || { echo "FAIL: verdict categories not in export"; exit 1; }

# project (ariadne#180): lifecycle noun like issue — model vets, a broken model
# (edge targeting a status outside every category) is rejected, and the export
# carries the concrete blocks the binding consumes (categories/lifecycle/discovery).
cue vet "$dir/project.cue" || { echo "FAIL: valid project model did not vet"; exit 1; }

if cue vet "$dir/testdata/project_invalid.cue" 2>/dev/null; then
  echo "FAIL: invalid project model passed vet"; exit 1
fi

pjson="$(cue export "$dir/project.cue" --out json)"
echo "$pjson" | grep -q '"categories"' || { echo "FAIL: project categories not in export"; exit 1; }
echo "$pjson" | grep -q '"lifecycle"'  || { echo "FAIL: project lifecycle not in export";  exit 1; }
echo "$pjson" | grep -q '"discovery"'  || { echo "FAIL: project discovery not in export";  exit 1; }

# project baseline guard (ariadne#171 M1): a `done` record predating baseline
# discipline (no deadline/planned_finish) must VALIDATE — a properly-run project
# still carries a baseline (it passed through executing), but a record archived
# from the pre-baseline era honestly has none, and migration must not fabricate
# dates. The negative control guards against over-relaxing: a LIVE `executing`
# record with no baseline must STILL be rejected.
cue vet "$dir/project.cue" "$dir/testdata/project_done_no_baseline.json" -d '#Project' \
  || { echo "FAIL: done record without baseline should validate (#171 M1)"; exit 1; }

if cue vet "$dir/project.cue" "$dir/testdata/project_executing_no_baseline.json" -d '#Project' 2>/dev/null; then
  echo "FAIL: executing record without baseline must still be rejected"; exit 1
fi

echo ok
