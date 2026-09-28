#!/bin/bash
# legacy-equivalence.sh — differential check that an sdlc build runs the legacy
# (pre-#252) issue workflow exactly as the frozen pre-#252 binary does, in a
# repository without an issue tracker (#252). Runs two scripted scenarios with
# each binary on identical fresh repositories and compares every step's exit
# code and normalized output, and the published main's tree and issue files.
#
# usage: legacy-equivalence.sh <old-sdlc> <new-sdlc>
#   build the old one from the frozen revision, e.g.
#     git archive pre-252-freeze | tar -x -C "$d" && (cd "$d" && go build -o old-sdlc ./cmd/sdlc)
# exit 0 when every step and the end state are identical.
set -uo pipefail
OLD=$1; NEW=$2; WORK=$(mktemp -d "${TMPDIR:-/tmp}/legacy-equivalence.XXXXXX")
norm() { sed -E 's/[0-9]{4}-[0-9]{2}-[0-9]{2}(T[0-9:]+[-+][0-9:]+)?/DATE/g; s/[0-9a-f]{7,40}/SHA/g; s#(/private)?'"$WORK"'/(old|new)[12]#W#g; s/[0-9]+\.[0-9]+h/Nh/g'; }
seed() {
  export GIT_CONFIG_NOSYSTEM=1
  git init -q --bare -b main origin.git
  git init -q -b main repo && cd repo
  git config user.email t@t; git config user.name t; git config commit.gpgsign false
  mkdir -p workshop/issues workshop/history/issues workshop/plans cmd
  printf -- '---\nid: 000001\nstatus: open\ncreated: 2026-09-01\n---\n\n# One\n\n## Problem\n\nFirst.\n\n## Log\n' > workshop/issues/000001-one.md
  printf -- "$1" > workshop/issues/000002-two.md
  printf -- '---\nid: 000003\nstatus: done\nactual_hours: 1\n---\n\n# Gone\n\n## Problem\nx\n' > workshop/history/issues/000003-gone.md
  git add -A; git commit -qm seed; git remote add origin ../origin.git; git push -q -u origin main
}
report() {
  git fetch -q origin
  echo "=== origin/main tree"; git ls-tree -r --name-only origin/main | sort
  echo "=== origin/main issue files"
  for p in $(git ls-tree -r --name-only origin/main -- workshop | sort); do echo "--- $p"; git show "origin/main:$p"; done
  echo "=== issue-tracker on origin: $(git ls-remote origin refs/heads/issue-tracker)"
}
scenario1() { # new → claim → start-plan → change-code → close → sync → list → set-status → show
  B=$1; mkdir -p "$2" && cd "$2" || exit 2
  seed '---\nid: 000002\nstatus: working\ncreated: 2026-09-01\n---\n\n# Two\n\nThe report.\n\n## Plan\n\n- [ ] x\n'
  step() { local n=$1; shift; "$B" "$@" > "../out-$n.txt" 2>&1; echo "$n exit=$?"; }
  step 01-new issue new "legacy cycle" --slug cycle
  step 02-claim claim --issue 4
  step 03-startplan start-plan --issue 4
  f=workshop/issues/000004-cycle.md
  perl -0pi -e 's/## Problem\n/## Problem\n\nA gap.\n\n## Spec\n\nA thing.\n\n## Done when\n\n- it works\n\n## Plan\n\n- [x] do it\n\n## Scratch\n/' "$f"
  step 04-changecode change-code --issue 4 --worktree=no --no-judge --no-estimate --no-estimate-recon --flow quick
  echo "branch=$(git branch --show-current)"
  printf 'package cycle\n' > cmd/cycle.go; git add cmd/cycle.go; git commit -qm "#4: implement"
  step 05-close close --issue 4 --verified "legacy cycle" --actual 1 --no-atlas --no-judge
  git add -A; git commit -qm "#4: close" >/dev/null 2>&1
  step 06-syncpush issue sync --issue 4 --push
  step 07-list issue list
  step 08-setstatus issue set-status blocked --issue 1
  step 09-show issue show 1
  report
}
scenario2() { # set-status, state, lint-ids, milestone-close, actual, close, publish, push landing
  B=$1; mkdir -p "$2" && cd "$2" || exit 2
  seed '---\nid: 000002\nstatus: working\ncreated: 2026-09-01\n---\n\n# Two\n\nThe report.\n\n## Spec\n\nS.\n\n## Done when\n\n- d\n\n## Plan\n\n- [x] M1 — first half\n- [ ] M2 — second half\n\n## Log\n'
  step() { local n=$1; shift; "$B" "$@" > "../out-$n.txt" 2>&1; echo "$n exit=$?"; }
  step 11-setstatus-legal issue set-status blocked --issue 2
  step 12-setstatus-back issue set-status working --issue 2
  step 13-state state
  step 14-lintids issue lint-ids
  git switch -q -c 000002-two
  printf 'package two\n' > cmd/two.go; git add cmd/two.go; git commit -qm "#2 M1: first half"
  step 15-milestoneclose milestone-close --issue 2 --milestone M1 --verified m1 --actual 0.5 --no-atlas --no-judge
  git add -A; git commit -qm "#2 M1: close" >/dev/null 2>&1
  step 16-actual actual --issue 2
  printf 'package two // more\n' > cmd/two.go; git commit -qam "#2 M2: second half"
  step 17-close close --issue 2 --verified m2 --actual 1 --no-atlas --no-judge
  git add -A; git commit -qm "#2: close" >/dev/null 2>&1
  printf -- '# note\n' > workshop/plans/000002-two-note.md; git add workshop/plans/000002-two-note.md; git commit -qm "#2: doc note"
  step 18-publish issue publish --commit "$(git rev-parse HEAD)"
  git switch -q main; git pull -q --no-rebase --no-edit >/dev/null 2>&1
  git merge -q --no-ff --no-edit 000002-two >/dev/null 2>&1; echo "merge-exit=$?"
  step 19-push push --yes --no-validate
  report
}
status=0
for s in 1 2; do
  (scenario$s "$OLD" "$WORK/old$s") > "$WORK/end-old$s.txt" 2>&1
  (scenario$s "$NEW" "$WORK/new$s") > "$WORK/end-new$s.txt" 2>&1
  if cmp -s <(norm < "$WORK/end-old$s.txt") <(norm < "$WORK/end-new$s.txt"); then echo "scenario $s end state: same"; else echo "scenario $s end state: DIFFERENT"; diff <(norm < "$WORK/end-old$s.txt") <(norm < "$WORK/end-new$s.txt") | head -20; status=1; fi
  for f in "$WORK/old$s"/out-*.txt; do n=$(basename "$f")
    if cmp -s <(norm < "$f") <(norm < "$WORK/new$s/$n"); then echo "  same  $n"; else echo "  DIFF  $n"; diff <(norm < "$f") <(norm < "$WORK/new$s/$n") | head -10; status=1; fi
  done
done
echo "work dir: $WORK"
exit $status
