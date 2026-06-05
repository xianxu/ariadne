#!/usr/bin/env bash
# docflow.test.sh — unit tests for the pure helpers + a real-git e2e of the flow.
# No mocks: the e2e drives an actual throwaway repo under one guarded temp root
# and asserts the resulting log shape, author attribution, and the ship guard.
#
# Run:  scripts/docflow.test.sh   (exit 0 = all pass; cleanly SKIPs the file/e2e
# tiers if no temp dir is available, e.g. a restricted sandbox — never falls
# through to the host repo, the failure mode #79's own bug shipped.)

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
DOCFLOW="$SCRIPT_DIR/docflow.sh"

# Source for unit tests (the BASH_SOURCE guard keeps main() from dispatching),
# then relax the strict opts docflow.sh set so assertions continue past non-zero.
# shellcheck source=/dev/null
source "$DOCFLOW"
set +e +u +o pipefail

fails=0
pass() { printf '  \033[1;32mok\033[0m   %s\n' "$1"; }
fail() { printf '  \033[1;31mFAIL\033[0m %s\n' "$1"; fails=$((fails + 1)); }
skip() { printf '  \033[1;33m--\033[0m   SKIP %s\n' "$1"; }
eq()   { if [[ "$2" == "$3" ]]; then pass "$1"; else fail "$1 (expected '$2', got '$3')"; fi; }

# One guarded temp root for every file-touching test. If it can't be made, the
# file/e2e tiers SKIP (not FAIL) — and crucially nothing ever writes to the host.
work="$(mktemp -d 2>/dev/null || true)"
have_tmp=1
[[ -z "$work" || ! -d "$work" ]] && have_tmp=0
[[ "$have_tmp" -eq 1 ]] && trap 'cd / 2>/dev/null; rm -rf "$work"' EXIT

echo "── unit: pure helpers ──"
eq "slugify basename+ext+case+space" "my-post"      "$(slugify 'src/data/My Post.md')"
eq "slugify underscores+digits"      "the-value-2"  "$(slugify 'The_Value 2.md')"
eq "slugify trims dashes"            "x"            "$(slugify '__x__.md')"
eq "review_branch_name"              "review/foo"   "$(review_branch_name foo)"
eq "round_subject"                   "review(foo): agent r3 — did stuff" \
                                     "$(round_subject foo agent 3 'did stuff')"
eq "marker_count missing file"       "0"            "$(marker_count /no/such/file)"

if [[ "$have_tmp" -eq 1 ]]; then
    mtmp="$work/marker.md"
    printf '%s\n' 'hello 🤖{a}' 'world' '```' 'code 🤖{ignored}' '```' 'tail 🤖[b]' > "$mtmp"
    eq "marker_count skips fenced markers" "2" "$(marker_count "$mtmp")"
    printf '%s\n' 'clean prose' > "$mtmp"
    eq "marker_count clean file"           "0" "$(marker_count "$mtmp")"
else
    skip "marker_count file tests (no temp dir)"
fi

echo "── e2e: real git flow ──"
if [[ "$have_tmp" -eq 0 ]]; then
    skip "e2e (no temp dir; run unsandboxed/CI for full coverage)"
else
    repo="$work/repo"; mkdir -p "$repo"; log="$work/repo.log"
    cd "$repo" || { fail "cd into temp repo"; exit 1; }
    # Belt-and-suspenders: refuse to run destructive git ops unless cwd is the
    # temp repo — the exact guard the original fall-through bug lacked.
    case "$PWD" in "$repo"*) ;; *) fail "e2e refused: cwd '$PWD' not under temp root"; exit 1;; esac

    git -c init.defaultBranch=main init -q
    git config user.name "Operator"; git config user.email "op@example.com"
    printf 'seed\n' > README.md; git add README.md; git commit -q -m init
    printf 'unrelated\n' > other.md; git add other.md; git commit -q -m "seed other.md (tracked, unrelated WIP)"

    run() { local d="$1"; shift; [[ "$1" == -- ]] && shift; if "$@" >>"$log" 2>&1; then pass "$d"; else fail "$d (cmd failed; see $log)"; fi; }
    post() { printf -- '---\npublished: false\n---\n\n# Draft\n\n%s\n' "$1" > post.md; }

    post 'Hello world.'
    run "start tracks untracked draft" -- "$DOCFLOW" start post.md
    eq  "on review branch"   "review/post" "$(git branch --show-current)"
    eq  "base recorded in .git/docflow (#84)" "main" "$(cat .git/docflow/post/base)"
    eq  "start records in-scope file" "post.md" "$(cat .git/docflow/post/files)"
    eq  "no state leaked into .git/config (#84)" "" "$(git config --get branch.review/post.docflowBase 2>/dev/null || true)"

    printf 'unrelated EDIT\n' > other.md   # #81: unrelated tracked WIP — must NOT be swept into a round
    post 'Hello world. 🤖[tighten this]'
    run "round human r1" -- "$DOCFLOW" round --side human -m "incoming markers"
    case "$(git show --name-only --format= HEAD)" in
        *other.md*) fail "round swept unrelated other.md into the commit (#81)";;
        *)          pass "round did not sweep unrelated other.md (#81)";;
    esac
    eq  "unrelated file left dirty after round" " M other.md" "$(git status --porcelain -- other.md)"
    git checkout -- other.md                # clean up so the later ship sees a clean tree
    post 'Hello, world.'
    run "round agent r1" -- "$DOCFLOW" round --side agent -m "tightened" --body "Removed filler; Oxford comma."

    eq  "human round authored by operator" "op@example.com" \
        "$(git log --format='%ae' --grep='human r1' -1)"
    eq  "agent round authored by agent"    "noreply@anthropic.com" \
        "$(git log --format='%ae' --grep='agent r1' -1)"
    case "$(git log --format='%b' --grep='agent r1' -1)" in
        *"Oxford comma"*) pass "agent rationale kept in commit body";;
        *)                fail "agent rationale kept in commit body";;
    esac

    post 'Hello, world. 🤖[one more pass?]'
    run "round human r2 (re-open)" -- "$DOCFLOW" round --side human -m "re-open"
    if "$DOCFLOW" ship >>"$log" 2>&1; then fail "ship blocks on outstanding marker"; else pass "ship blocks on outstanding marker"; fi
    eq  "still on review branch after blocked ship" "review/post" "$(git branch --show-current)"

    post 'Hello, world.'
    run "round agent r2 (resolve)" -- "$DOCFLOW" round --side agent -m "resolved"
    run "ship merges clean"        -- "$DOCFLOW" ship
    eq  "back on base after ship" "main" "$(git branch --show-current)"
    if git show-ref --verify --quiet refs/heads/review/post; then fail "review branch deleted"; else pass "review branch deleted"; fi
    if [[ -d .git/docflow/post ]]; then fail "meta dir removed after ship (#84)"; else pass "meta dir removed after ship (#84)"; fi
    eq  "first-parent shows one merge line" "1" \
        "$(git log --first-parent --format='%s' main | grep -c 'review(post): merge')"
    eq  "merge message counts 4 round commits (excl. track)" "1" \
        "$(git log --first-parent --format='%s' main | grep -c 'review(post): merge — 4 round-commit')"
    eq  "full log retains all 4 round commits" "4" \
        "$(git log --format='%s' main | grep -cE 'review\(post\): (human|agent) r')"

    post2() { printf -- '---\npublished: false\n---\n\n# Two\n\n%s\n' "$1" > two.md; }
    post2 'draft 🤖[unresolved]'
    run "start second doc"            -- "$DOCFLOW" start two.md
    run "round human r1 (two)"        -- "$DOCFLOW" round --side human -m "draft with marker"
    run "ship --force merges as-is" -- "$DOCFLOW" ship --force
    eq  "back on base after ship --force" "main" "$(git branch --show-current)"
    eq  "forced merge present on first-parent" "1" \
        "$(git log --first-parent --format='%s' main | grep -c 'review(two): merge')"

    # #81: an in-scope path containing a space survives round staging (array, not word-split).
    printf 'note\n' > "my note.md"
    run "start space-named doc"     -- "$DOCFLOW" start "my note.md"
    run "re-start same doc (dedup)" -- "$DOCFLOW" start "my note.md"
    eq  "in-scope recorded once (dedup)" "my note.md" "$(cat .git/docflow/my-note/files)"
    printf 'note 🤖[m]\n' > "my note.md"
    run "round on space-named doc"  -- "$DOCFLOW" round --side human -m "space path"
    case "$(git show --name-only --format= HEAD)" in
        *"my note.md"*) pass "round staged space-named in-scope file (#81)";;
        *)              fail "round did NOT stage space-named in-scope file (word-split?)";;
    esac
    # Deprecated alias: `finish` still merges (via ship) and warns it was renamed.
    alias_out="$("$DOCFLOW" finish --force 2>&1)"; alias_rc=$?
    eq  "finish alias exits 0 (merges via ship)" "0" "$alias_rc"
    eq  "back on base after finish alias"        "main" "$(git branch --show-current)"
    case "$alias_out" in
        *"'finish' is now 'ship'"*) pass "finish alias warns it is deprecated";;
        *)                          fail "finish alias warns it is deprecated (got: $alias_out)";;
    esac
    cd /
fi

echo
if [[ "$fails" -eq 0 ]]; then
    if [[ "$have_tmp" -eq 1 ]]; then printf '\033[1;32mALL PASS\033[0m\n'
    else printf '\033[1;33mPASS (pure-helper tier only; file/e2e SKIPPED — no temp dir)\033[0m\n'; fi
    exit 0
else
    printf '\033[1;31m%s FAILED\033[0m\n' "$fails"; exit 1
fi
