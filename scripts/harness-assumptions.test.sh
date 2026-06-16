#!/usr/bin/env bash
# harness-assumptions.test.sh — guards ariadne's per-harness skill-discovery
# integration assumptions (issue #107, target base-layer-mechanics).
#
# weave's Option B lowering lowers skills to per-harness DIRS — .claude/skills
# (Claude) + .agents/skills (Codex + Gemini) — and trusts each harness to discover
# its own dir natively, in weave's SKILL.md format, FOLLOWING the symlinks weave
# lowers. The .agents/skills SCAN-PATH is a per-tool convention (Gemini ships it
# "Experimental Preview"; Codex had a ~/.agents/skills regression), so it is the
# UNSTABLE, load-bearing part. This suite pins it: it builds a fixture mirroring
# weave's lowering and asserts, per INSTALLED harness, the contract weave relies on.
# It SKIPs an absent CLI and EXITs non-zero on any FAIL.
#
#   Run:     make harness-check     (or: bash scripts/harness-assumptions.test.sh)
#   Self:    bash scripts/harness-assumptions.test.sh --self-test   (proves it can FAIL)
#   Trigger: before a propagate/re-weave (M4), and whenever a harness CLI updates.
#            CI typically lacks the CLIs → it SKIPs there; the value is the dev gate.
#   Triage a break: atlas/workflow/harness-integration.md.

set -u
PASS=0; FAIL=0; SKIP=0
green()  { printf '\033[32m%s\033[0m\n' "$1"; }
red()    { printf '\033[31m%s\033[0m\n' "$1"; }
yellow() { printf '\033[33m%s\033[0m\n' "$1"; }
pass() { PASS=$((PASS+1)); green  "  PASS: $1"; }
fail() { FAIL=$((FAIL+1)); red    "  FAIL: $1"; }
skip() { SKIP=$((SKIP+1)); yellow "  SKIP: $1"; }

# assert_contains <text> <needle> <msg> ; assert_absent <text> <needle> <msg>
assert_contains() { case "$1" in *"$2"*) pass "$3";; *) fail "$3 (missing: $2)";; esac; }
assert_absent()   { case "$1" in *"$2"*) fail "$3 (unexpected: $2)";; *) pass "$3";; esac; }

# --self-test: prove the guard has teeth (a wrong assertion FAILs + exits non-zero).
if [ "${1:-}" = "--self-test" ]; then
  echo "== self-test (expect 1 FAIL, exit 1) =="
  assert_contains "haystack" "NOT-PRESENT" "deliberately-wrong assertion must FAIL"
  [ "$FAIL" -eq 1 ] && green "self-test OK: a broken assertion FAILs + exits 1" \
                    || { red "self-test BROKEN: FAIL counter did not fire"; exit 2; }
  exit 1
fi

# ---- fixture: mirror weave's lowering (a real + a RELATIVE-symlinked skill dir) ----
FIX=$(mktemp -d "${TMPDIR:-/tmp}/harness-assume.XXXXXX")
trap 'rm -rf "$FIX"' EXIT
mkskill() { # <dir> <name> <description>
  mkdir -p "$1"
  printf -- '---\nname: %s\ndescription: %s\n---\n# %s\nbody.\n' "$2" "$3" "$2" > "$1/SKILL.md"
}
mkskill "$FIX/.agents/skills/probe-real" probe-real "probe REAL dir under .agents/skills"
mkskill "$FIX/srclayer/probe-link"       probe-link "probe SYMLINKED dir under .agents/skills"
ln -s ../../srclayer/probe-link "$FIX/.agents/skills/probe-link"   # the weave lowering form
mkskill "$FIX/.claude/skills/probe-claudeonly" probe-claudeonly "ONLY under .claude/skills"
printf -- '# CLAUDE entry\nCLAUDE-MARKER\n' > "$FIX/CLAUDE.md"
printf -- '# AGENTS entry\nAGENTS-MARKER\n' > "$FIX/AGENTS.md"
echo "fixture: $FIX"; echo

# ---- Gemini: gemini skills list --all (deterministic, local) ----
echo "== gemini =="
if command -v gemini >/dev/null 2>&1; then
  OUT=$(cd "$FIX" && gemini skills list --all 2>&1 || true)
  assert_contains "$OUT" probe-real       "gemini discovers a REAL .agents/skills skill (format parity)"
  assert_contains "$OUT" probe-link       "gemini FOLLOWS a symlinked .agents/skills skill (weave lowering)"
  assert_absent   "$OUT" probe-claudeonly "gemini IGNORES .claude/skills"
else
  skip "gemini not installed"
fi
echo

# ---- Codex: codex debug prompt-input (renders the model-visible prompt; no API call) ----
echo "== codex =="
if command -v codex >/dev/null 2>&1; then
  OUT=$(cd "$FIX" && codex debug prompt-input 2>&1 || true)
  assert_contains "$OUT" probe-real       "codex discovers a REAL .agents/skills skill (format parity)"
  assert_contains "$OUT" probe-link       "codex FOLLOWS a symlinked .agents/skills skill (weave lowering)"
  assert_absent   "$OUT" probe-claudeonly "codex IGNORES .claude/skills"
else
  skip "codex not installed"
fi
echo

# ---- Claude: doc-asserted (no non-interactive skills-list / render-prompt hook) ----
echo "== claude =="
if command -v claude >/dev/null 2>&1; then
  yellow "  DOC-ASSERTED (no deterministic CLI hook to runtime-probe):"
  yellow "    - Claude reads CLAUDE.md (NOT AGENTS.md unless @-imported) + .claude/skills"
  yellow "    - ref: https://code.claude.com/docs/en/memory  (re-verify on major claude updates)"
  yellow "  OPEN: does Claude ALSO read .agents/skills? — affects whether weave still needs"
  yellow "        the .claude/skills face. Manual check; record in the atlas page."
else
  skip "claude not installed"
fi
echo

echo "------------------------------------------------------------"
echo "PASS=$PASS  FAIL=$FAIL  SKIP=$SKIP"
if [ "$FAIL" -gt 0 ]; then red "HARNESS ASSUMPTIONS BROKEN — triage: atlas/workflow/harness-integration.md"; exit 1; fi
green "harness assumptions hold (or skipped)"; exit 0
