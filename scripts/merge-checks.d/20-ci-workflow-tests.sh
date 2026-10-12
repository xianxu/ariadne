#!/usr/bin/env bash
# 20-ci-workflow-tests — run every scripts/test/*.test.sh (#309), so a script
# test cannot exist without a runner. Ariadne-local: these tests live here.
#
# Excluded by name, each with its runner:
#   release-weave.test.sh     needs a release dir and tag; weave-release.yml runs it
#   slot-dependencies.test.sh product acceptance that clones and builds the
#                             parley.nvim sibling; run by hand (see its header)
#
# Contract: <check> <base_sha> <head_sha>, exit 0 = pass, findings to stderr.
# The range is unused: the tests are cheap and cover their subjects whole.
set -euo pipefail
cd "$(git rev-parse --show-toplevel)"
for t in scripts/test/*.test.sh; do
    case "$(basename "$t")" in
        release-weave.test.sh|slot-dependencies.test.sh) continue ;;
    esac
    bash "$t" >&2
done
