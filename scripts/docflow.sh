#!/usr/bin/env bash
# docflow — branch-scoped prose review with per-round git journaling.
#
# Wraps an `xx-fix` co-authoring review (operator + agent trading 🤖 markers in a
# markdown doc) in a git branch and journals every round, so the full back-and-
# forth — and the agent's rationale — survives as durable, attributable history
# instead of dying with the chat transcript.
#
# Git is the only state store; this script is guards + orchestration over git,
# not a state machine. The companion is the `xx-fix` skill (the prose mechanic);
# this script owns the git side effects (branch / commit / merge).
#
# VERBS
#   start <file>...        create review/<slug> from the current branch (or add
#                          files to the current review branch); track untracked
#                          drafts as round 0.
#   round --side h|a ...    journal one side of a round (two commits/round →
#                          attributable log). Called by xx-fix, before & after
#                          the agent edits. Flags: -m/--summary, --body, files.
#   status                 current branch, base, rounds, in-scope files + 🤖 count.
#   finish [--force]       guard (no 🤖 left) → --no-ff merge to base → delete the
#                          review branch. --force merges as-is (== "abandon").
#
# HISTORY MODEL
#   No squash. `finish` does a --no-ff merge so `git log --first-parent <base>`
#   shows one clean merge line per reviewed batch, while a plain `git log` keeps
#   every round (forensics). Deleting the review branch loses nothing — the round
#   commits stay reachable as the merge commit's second parent.
#
# ENV
#   DOCFLOW_AGENT_AUTHOR   author/Co-Authored-By identity for agent-side commits
#                          (default: "Claude <noreply@anthropic.com>").
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=/dev/null
source "$SCRIPT_DIR/lib.sh"

AGENT_AUTHOR="${DOCFLOW_AGENT_AUTHOR:-Claude <noreply@anthropic.com>}"

# ── Output helpers ────────────────────────────────────────────────────────────
ok()   { printf "  ${GREEN}✓ %s${RESET}\n" "$*" >&2; }
info() { printf "  ${CYAN}➜ %s${RESET}\n" "$*" >&2; }
warn() { printf "  ${YELLOW}⚠ %s${RESET}\n" "$*" >&2; }
die()  { printf "${RED}error: %s${RESET}\n" "$*" >&2; exit 1; }

# ── Pure helpers (string→string / file→int; unit-tested directly) ─────────────

# slugify <name> → lowercased, dash-joined alnum stem of a filename's basename.
slugify() {
    local s="${1##*/}"      # basename
    s="${s%.*}"             # strip extension
    s="$(printf '%s' "$s" | tr '[:upper:]' '[:lower:]' | tr -c '[:alnum:]' '-')"
    s="$(printf '%s' "$s" | sed -E 's/-+/-/g; s/^-//; s/-$//')"
    printf '%s' "$s"
}

# review_branch_name <slug> → review/<slug>
review_branch_name() { printf 'review/%s' "$1"; }

# round_subject <slug> <side> <n> <summary> → commit subject (greppable convention)
round_subject() { printf 'review(%s): %s r%s — %s' "$1" "$2" "$3" "$4"; }

# marker_count <file> → number of 🤖/㊷ markers outside fenced code blocks.
# The awk pass strips ``` fences (ASCII-only logic, robust across awks); grep -o
# counts the emoji bytes. `|| true` keeps a zero-match grep from tripping set -e.
marker_count() {
    [[ -f "$1" ]] || { printf '0'; return 0; }
    local n
    n=$(awk '/^```/{f=!f; next} !f{print}' "$1" \
        | grep -o -e '🤖' -e '㊷' | wc -l | tr -d ' ' || true)
    printf '%s' "${n:-0}"
}

# ── Git context ───────────────────────────────────────────────────────────────
current_branch() { git branch --show-current; }

# docflow_meta_dir <review-branch> → per-review state dir under the git dir.
# State lives here as plain files, NOT in .git/config: the sandbox denies config
# (and hooks) writes — they can execute code — but plain files under .git/ are
# writable, so this keeps docflow fully sandbox-compatible (#84). `--git-dir` is
# worktree-local, which matches docflow's start→round→finish-in-one-place model.
docflow_meta_dir() { printf '%s/docflow/%s' "$(git rev-parse --git-dir)" "${1#review/}"; }

# review_base <review-branch> → the branch it was forked from (recorded at start).
review_base() { cat "$(docflow_meta_dir "$1")/base" 2>/dev/null || echo main; }

# inscope_files <review-branch> → the docs recorded at start, one per line. The
# source of truth for "what's under review": round stages exactly these, never a
# blanket `git add -u` that would sweep unrelated tracked WIP (lessons.md).
inscope_files() { cat "$(docflow_meta_dir "$1")/files" 2>/dev/null || true; }

# round_count <base> <slug> <side> → committed rounds for that side on this branch.
round_count() {
    git log --format='%s' "$1..HEAD" | grep -c "review($2): $3 r" || true
}

# ── Verbs ─────────────────────────────────────────────────────────────────────
cmd_start() {
    [[ $# -ge 1 ]] || die "usage: docflow start <file>..."
    is_git_repo || die "not a git repository"
    local cur slug rb base
    cur="$(current_branch)"
    if [[ "$cur" == review/* ]]; then
        rb="$cur"; slug="${rb#review/}"; base="$(review_base "$rb")"
        info "adding to existing review branch $rb (base $base)"
    else
        slug="$(slugify "$1")"
        [[ -n "$slug" ]] || die "could not derive a slug from '$1'"
        rb="$(review_branch_name "$slug")"
        base="$cur"
        [[ -n "$base" ]] || die "detached HEAD — check out a branch before starting a review"
        git show-ref --verify --quiet "refs/heads/$rb" \
            && die "branch $rb already exists — check it out, or start with a different first file"
        git checkout -q -b "$rb"
        ok "created review branch $rb (base $base)"
    fi
    local md; md="$(docflow_meta_dir "$rb")"
    mkdir -p "$md"
    [[ -f "$md/base" ]] || printf '%s\n' "$base" > "$md/base"
    local f
    for f in "$@"; do
        [[ -f "$f" ]] || die "no such file: $f"
        # Record as in-scope (deduped) so `round` stages exactly the review docs.
        grep -qxF "$f" "$md/files" 2>/dev/null || printf '%s\n' "$f" >> "$md/files"
        if git ls-files --error-unmatch "$f" &>/dev/null; then
            info "$f already tracked — journaled on first round"
        else
            git add -- "$f"
            git commit -q -m "$(round_subject "$slug" track 0 "${f##*/}")"
            ok "tracked $f (round 0)"
        fi
    done
}

cmd_round() {
    is_git_repo || die "not a git repository"
    local side="" summary="" body=""; local -a files=()
    while [[ $# -gt 0 ]]; do
        case "$1" in
            --side)        side="$2"; shift 2;;
            -m|--summary)  summary="$2"; shift 2;;
            --body)        body="$2"; shift 2;;
            --)            shift; files+=("$@"); break;;
            -*)            die "unknown flag: $1";;
            *)             files+=("$1"); shift;;
        esac
    done
    [[ "$side" == human || "$side" == agent ]] \
        || die "round requires --side human|agent"
    local cur slug base
    cur="$(current_branch)"
    [[ "$cur" == review/* ]] || die "not on a review branch (run 'docflow start' first)"
    slug="${cur#review/}"; base="$(review_base "$cur")"

    if [[ ${#files[@]} -gt 0 ]]; then
        git add -- "${files[@]}"
    else
        # Stage exactly the recorded in-scope docs — never `git add -u`, which
        # sweeps unrelated tracked WIP into the review. Read into an array so
        # paths with spaces survive.
        local f; local -a inscope=()
        while IFS= read -r f; do [[ -n "$f" ]] && inscope+=("$f"); done < <(inscope_files "$cur")
        if [[ ${#inscope[@]} -gt 0 ]]; then
            git add -- "${inscope[@]}"
        else
            warn "no in-scope files recorded for $cur — staging all tracked changes (legacy branch?)"
            git add -u
        fi
    fi
    if git diff --cached --quiet; then
        info "no changes to journal for $side round — skipping"
        return 0
    fi

    local n; n=$(( $(round_count "$base" "$slug" "$side") + 1 ))
    [[ -n "$summary" ]] || summary="$side round $n"
    local -a args=(commit -q -m "$(round_subject "$slug" "$side" "$n" "$summary")")
    [[ -n "$body" ]] && args+=(-m "$body")
    if [[ "$side" == agent ]]; then
        # Attribute the agent's round to the agent identity (the load-bearing
        # detail that keeps `git log` readable: human rounds = operator, agent
        # rounds = agent), plus a Co-Authored-By trailer per repo convention.
        args+=(-m "Co-Authored-By: $AGENT_AUTHOR")
        git "${args[@]}" --author="$AGENT_AUTHOR"
    else
        git "${args[@]}"
    fi
    ok "$side r$n — $summary"
}

cmd_status() {
    is_git_repo || die "not a git repository"
    local cur; cur="$(current_branch)"
    if [[ "$cur" != review/* ]]; then
        printf "${BOLD}not on a review branch${RESET} (current: %s)\n" "$cur" >&2
        local brs; brs=$(git for-each-ref --format='%(refname:short)' refs/heads/review/ 2>/dev/null || true)
        [[ -n "$brs" ]] && { printf "open review branches:\n" >&2; printf '  %s\n' $brs >&2; }
        return 0
    fi
    local slug base; slug="${cur#review/}"; base="$(review_base "$cur")"
    printf "${BOLD}review:${RESET} %s   ${BOLD}base:${RESET} %s\n" "$cur" "$base" >&2
    printf "${BOLD}rounds:${RESET} %s human / %s agent\n" \
        "$(round_count "$base" "$slug" human)" "$(round_count "$base" "$slug" agent)" >&2
    printf "${BOLD}in-scope files:${RESET}\n" >&2
    local f m total=0
    while IFS= read -r f; do
        [[ -z "$f" ]] && continue
        if [[ -f "$f" ]]; then
            m="$(marker_count "$f")"; total=$((total + m))
            printf "  %s  (🤖 %s)\n" "$f" "$m" >&2
        else
            printf "  %s  (deleted)\n" "$f" >&2
        fi
    done < <(git diff --name-only "$base..HEAD")
    printf "${BOLD}outstanding 🤖:${RESET} %s\n" "$total" >&2
}

cmd_finish() {
    is_git_repo || die "not a git repository"
    local force=0
    while [[ $# -gt 0 ]]; do
        case "$1" in
            --force) force=1; shift;;
            *)       die "unexpected arg: $1";;
        esac
    done
    local cur slug base; cur="$(current_branch)"
    [[ "$cur" == review/* ]] || die "not on a review branch"
    slug="${cur#review/}"; base="$(review_base "$cur")"

    if ! git diff --quiet || ! git diff --cached --quiet; then
        die "uncommitted changes — journal a final round first (docflow round --side ...)"
    fi

    local f m total=0 offenders=""
    while IFS= read -r f; do
        [[ -z "$f" || ! -f "$f" ]] && continue
        m="$(marker_count "$f")"
        [[ "$m" -gt 0 ]] && { total=$((total + m)); offenders+="  $f (🤖 $m)\n"; }
    done < <(git diff --name-only "$base..HEAD")

    if [[ "$total" -gt 0 ]]; then
        if [[ "$force" -eq 1 ]]; then
            printf "${YELLOW}merging with %s outstanding 🤖 (drafts with published:false won't render):${RESET}\n" "$total" >&2
            printf '%b' "$offenders" >&2
        else
            printf "${RED}refusing to finish: %s outstanding 🤖 marker(s):${RESET}\n" "$total" >&2
            printf '%b' "$offenders" >&2
            die "resolve them (re-run /xx-fix) or 'docflow finish --force' to merge as-is"
        fi
    fi

    # Count actual round commits (human/agent), not the round-0 track commit.
    local rounds; rounds=$(git log --format='%s' "$base..HEAD" | grep -cE "review\($slug\): (human|agent) r" || true)
    git checkout -q "$base" || die "could not check out base '$base' (checked out in another worktree?)"
    git merge --no-ff -q "$cur" -m "review($slug): merge — $rounds round-commit(s)"
    # `git branch -d` tries to prune the branch's .git/config section; the sandbox
    # denies that write (harmless — the branch still deletes, exit 0). Hush the
    # misleading "could not write config file" warning it prints to stderr.
    git branch -d "$cur" >/dev/null 2>&1
    rm -rf "$(docflow_meta_dir "$cur")"
    ok "merged $cur → $base (--no-ff), deleted the review branch"
    info "clean view: git log --first-parent $base    full: git log $base"
}

usage() {
    sed -n '2,40p' "${BASH_SOURCE[0]}" | sed 's/^# \{0,1\}//' >&2
}

main() {
    [[ $# -ge 1 ]] || { usage; exit 1; }
    local verb="$1"; shift
    case "$verb" in
        start)            cmd_start "$@";;
        round)            cmd_round "$@";;
        status)           cmd_status "$@";;
        finish)           cmd_finish "$@";;
        -h|--help|help)   usage;;
        *)                die "unknown verb: $verb (start|round|status|finish)";;
    esac
}

# Only dispatch when executed; sourcing (e.g. the test) exposes helpers without running.
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi
