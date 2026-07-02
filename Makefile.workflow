# AI issue-based workflow — include from your project Makefile:
#   include Makefile.workflow

# Include openshell targets if available
-include .openshell/Makefile

# Include tart targets if available (macOS VM testing — Apple Silicon)
-include .tart/Makefile

# Include colima targets if available (clean Linux VM testing — Apple Silicon)
-include .colima/Makefile
# Override WF_ISSUES_DIR / WF_HISTORY_DIR before the include if your
# issues and history live somewhere other than issues/ and history/.

WF_ISSUES_DIR ?= issues
WF_HISTORY_DIR ?= history
export WF_ISSUES_DIR WF_HISTORY_DIR

# BRAIN_DIR points at the brain repo for cross-cutting state (project files,
# velocity baselines). close-issue.py reads it to update parent project tasks.
# Must default *here* — without ?=, the close-issue: export below would emit
# an empty string when BRAIN_DIR is unset, which silently overrides the
# Python default in scripts/close-issue.py and suppresses project updates.
BRAIN_DIR ?= ../brain

.PHONY: help-workflow worktree fetch push pull-request merge check pre-merge weave weave-drift-check issue-sync

help-workflow:
	@printf '%s\n' \
	"AI Workflow (issue-based):" \
	"" \
	"  Work on main:" \
	"    make fetch 42       Fetch GitHub issue, create $(WF_ISSUES_DIR)/NNNN-slug.md" \
	"    make push           Auto-commit, push, close done issues, archive to $(WF_HISTORY_DIR)/" \
	"" \
	"  Work on a larger issue:" \
	"    make worktree       Auto-detect issue file, commit, create worktree" \
	"    make worktree NAME  Create a worktree with explicit name" \
	"    make pull-request   Push branch, open PR referencing GitHub issues" \
	"    make merge          Merge PR, archive done issues, clean up worktree" \
	"" \
	"  Pre-merge checks (agent-driven, run first in push/merge):" \
	"    make check          Run all checks with interactive selection" \
	"    make check-dry      Check DRY principle" \
	"    make check-pure     Check PURE principle" \
	"    make check-plan     Check issue plan completeness" \
	"    make check-specs    Check atlas/README sync" \
	"    make check-lessons  Check for lessons to capture" \
	"    PRE_MERGE_CHECKS=yynnyn make pre-merge   Preset selection" \
	"" \
	"  Sync issues:" \
	"    make issue-sync     Sync $(WF_ISSUES_DIR)/ changes to main and push" \
	"" \
	"  Close (mechanical §5 checklist):" \
	"    make close-issue ISSUE=N [MILESTONE=Mx] ACTUAL=h VERIFIED='...'" \
	"                        Tick checkboxes, set status/actual_hours, update project file" \
	"" \
	"  Setup:" \
	"    make weave          Re-run $(UPSTREAM_NAME) setup (link + merge settings)" \
	""

# ── Issue sync ────────────────────────────────────────────────────────────────
# Sync issue file changes to main and push, even when on a feature branch.
# Delegates to bin/sdlc claim (renamed from `sdlc lock` in #39) when the
# binary is built; falls back to the shell script otherwise.
issue-sync:
	@if [ -x bin/sdlc ]; then \
	    bin/sdlc claim; \
	else \
	    scripts/issue-sync.sh; \
	fi

# ── Close (issue or milestone) ────────────────────────────────────────────────
# Mechanical part of AGENTS.md §5: tick checkboxes, flip status, write
# actual_hours, update the project file's task row + detail block.
# Does NOT commit — the agent commits, usually bundling other content.
#
# Usage:
#   make close-issue ISSUE=15 MILESTONE=M4 ACTUAL=2.5 VERIFIED="ran ./test, saw X"
#   make close-issue ISSUE=15 ACTUAL=7 VERIFIED="end-to-end run, captured in Log"
# Required for issue close: ACTUAL + VERIFIED.
# Flags:
#   FORCE=1   skip "already done" / "Plan unchecked" / "atlas untouched" guards
#   DRY=1     print what would change, write nothing
#   BRAIN_DIR=../brain   override project-file lookup root
.PHONY: close-issue
close-issue: export ISSUE       := $(ISSUE)
close-issue: export MILESTONE   := $(MILESTONE)
close-issue: export ACTUAL      := $(ACTUAL)
close-issue: export VERIFIED    := $(VERIFIED)
close-issue: export FORCE       := $(FORCE)
close-issue: export DRY         := $(DRY)
close-issue: export BRAIN_DIR   := $(BRAIN_DIR)
# Delegates to bin/sdlc close when the Go binary is built; falls back to
# the Python script otherwise. Both implementations match byte-for-byte
# on stderr (Go is a faithful port of close-issue.py). The fallback path
# keeps downstream repos that haven't run `make sdlc-build` yet working.
# After M8 deprecates the Python script, the fallback branch goes away.
#
# Bash ${VAR:+--flag "$$VAR"} expands to nothing when VAR is unset/empty,
# else to --flag "value" — preserves spaces in VERIFIED across the call.
close-issue:
	@if [ -x bin/sdlc ]; then \
	    bin/sdlc close \
	      $${ISSUE:+--issue "$$ISSUE"} \
	      $${MILESTONE:+--milestone "$$MILESTONE"} \
	      $${ACTUAL:+--actual "$$ACTUAL"} \
	      $${VERIFIED:+--verified "$$VERIFIED"} \
	      $${FORCE:+--force} \
	      $${DRY:+--dry-run} \
	      $${BRAIN_DIR:+--brain-dir "$$BRAIN_DIR"}; \
	else \
	    scripts/close-issue.py; \
	fi

# ── Weave / bootstrap ─────────────────────────────────────────────────────────
# Two verbs, distinct concerns (per ariadne#38):
#
#   make weave           Pure substrate-state sync. Verifies peers are present,
#                        invokes construct/setup.sh to update symlinks. Does
#                        NOT clone peers, does NOT build tools. Errors if a
#                        peer is missing — operator should `make bootstrap`
#                        for first-time setup.
#
#   make bootstrap-peers Cascade peer cloning. Reads construct/go.mod for
#                        replace ../<name> directives; clones missing peers
#                        (URL derived from origin convention; operator can
#                        override). Recursively bootstraps each peer.
#
#   make data-deps       Clone + symlink data dependencies (content peers, not
#                        substrate). Reads construct/data-deps; clones each repo
#                        as a sibling and mounts it via a relative symlink.
#                        Language-agnostic; no-op when the manifest is absent.
#
#   make bootstrap       Composition: bootstrap-peers + weave + tools.
#                        Defined as prereqs-only (no recipe) so derivatives
#                        with their own bootstrap target (e.g. nous's GPG
#                        setup) can extend additively without recipe
#                        collision.
#
# First-time bootstrap of a fresh-clone derivative whose upstreams aren't
# yet checked out beside it: run `./bootstrap.sh` (a real committed file, not
# a symlink — see #42). It reads the real construct/go.mod, clones the upstream
# peer(s) as siblings, then hands off to `make bootstrap`. Without it you hit
# the chicken-and-egg where every make target is unreachable (Makefile itself
# is a dangling symlink into the not-yet-cloned upstream).
#
# Equivalent manual path if `./bootstrap.sh` is absent: clone the upstream as
# a sibling yourself (or run `../<upstream>/construct/setup.sh`), then
# `make bootstrap`. Once substrate has propagated, `make bootstrap` is the
# canonical post-clone command.

# weave now builds + invokes the weave binary (cmd/weave), the intent-compiler
# that replaced construct/setup.sh (#95). weave-build resolves weave's owner by
# LOCATION (construct/dev-aliases.sh --list) and builds the binary in-owner at
# $$owner/bin/weave — the same build-in-owner pattern sdlc-build uses, so a
# derivative needs no go.mod replace. This target then resolves the SAME owner
# and runs the OWNER's binary ($$owner/bin/weave) — NOT a local bin/weave, which
# build-in-owner deliberately never produces in a consumer (#95 M5). When THIS
# repo is the owner ($$owner is this repo's own dir), it runs its own bin/weave —
# unchanged. The recipe runs the bare Union `weave compile`, which compiles THIS
# repo's (the cwd's) layer composition for EVERY harness face (claude + codex +
# gemini): the generic symlinks, the prose-only entry-file compose, the
# settings.json merge, and the per-harness skill-dir lowerings (`.claude/skills`
# + `.agents/skills` — each harness discovers its dir natively, no AGENTS.md menu;
# see plan.Target). compile operates on the caller's cwd, so the owner's binary
# composing the consumer's repo is correct. Under `make bootstrap`,
# bootstrap-peers (clones ancestors) precedes weave, so weave's owner is present
# by the time this runs.
# PATH WIRING (#115 M3): the dynamic-skill marker now calls the `datatype` binary
# by NAME (not `go run`), so datatype must be (a) BUILT and (b) on PATH when weave
# execs the marker. So this target also depends on datatype-build, and exports the
# datatype owner's bin/ onto PATH before running weave compile — weave execs the
# marker via exec.Command, which inherits this PATH. This is shared Makefile.workflow:
# in a derivative, `make weave` resolves datatype's owner = ariadne, builds + PATH-
# exposes ariadne's bin/datatype, then weave (cwd=derivative) execs ariadne's marker
# which writes the DERIVATIVE's construct/generated (leaf-rooted output).
weave: weave-build datatype-build vocabulary-build ensure-cue
	@owner="$$(construct/dev-aliases.sh --list 2>/dev/null | awk -F'\t' '$$1=="weave"{print $$2}')"; \
	dtowner="$$(construct/dev-aliases.sh --list 2>/dev/null | awk -F'\t' '$$1=="datatype"{print $$2}')"; \
	vcowner="$$(construct/dev-aliases.sh --list 2>/dev/null | awk -F'\t' '$$1=="vocabulary"{print $$2}')"; \
	if [ -n "$$owner" ] && [ -x "$$owner/bin/weave" ]; then \
		PATH="$$dtowner/bin:$$vcowner/bin:$$PATH" "$$owner/bin/weave" compile; \
	else \
		echo "Error: weave binary not built (weave-build did not produce $$owner/bin/weave)."; \
		echo "  First-time bootstrap of a fresh derivative: run \`./bootstrap.sh\`,"; \
		echo "  or clone the upstream ariadne beside this repo and \`make bootstrap\`."; \
		exit 1; \
	fi

# weave-drift-check — the dynamic-skill GENERATE-IDEMPOTENCY guard (#115 M3, plan
# decision D2). The old #111 guard `git diff --exit-code`'d the COMMITTED
# construct/local/datatype/SKILL.md — but that body is now the GITIGNORED, per-repo
# construct/generated/<dir>/SKILL.md, regenerated EVERY compile. A gitignored,
# every-compile-regenerated output cannot go stale (git can't even see it), so the
# staleness guard's job evaporated. What still bites is DETERMINISM: the render must
# be byte-stable so two compiles produce identical bytes (else `git status` churns
# nondeterministically). This runs the datatype binary twice into two temp dirs and
# diffs the bytes — failing if the render is non-deterministic. (The renderer's
# byte-stable unit test in cmd/datatype is the in-process counterpart.) Run in CI
# after `make weave`; depends on datatype-build so the binary is on disk.
weave-drift-check: datatype-build
	@echo "==> weave drift check (dynamic-skill render must be deterministic)"
	@owner="$$(construct/dev-aliases.sh --list 2>/dev/null | awk -F'\t' '$$1=="datatype"{print $$2}')"; \
	bin="$$owner/bin/datatype"; \
	if [ ! -x "$$bin" ]; then echo "Error: datatype binary not built at $$bin" >&2; exit 1; fi; \
	d1="$$(mktemp -d)"; d2="$$(mktemp -d)"; \
	trap 'rm -rf "$$d1" "$$d2"' EXIT; \
	"$$bin" --output "$$d1" && "$$bin" --output "$$d2"; \
	if ! diff -q "$$d1/SKILL.md" "$$d2/SKILL.md" >/dev/null; then \
		echo "Error: dynamic-skill render is NON-DETERMINISTIC (two runs differ)." >&2; \
		echo "  The materialized construct/generated/<dir>/SKILL.md would churn git status." >&2; \
		exit 1; \
	fi
	@echo "    OK — dynamic-skill render is byte-stable across runs."

bootstrap-peers:
	@if [ -x construct/scripts/bootstrap-peers.sh ]; then \
		bash construct/scripts/bootstrap-peers.sh; \
	fi

# Clone + symlink DATA DEPENDENCIES (content peers, not substrate). Reads
# construct/data-deps; clones each declared repo as a sibling and mounts it via
# a relative symlink. Language-agnostic; no-op when the manifest is absent.
data-deps:
	@if [ -x construct/scripts/clone-data-deps.sh ]; then \
		bash construct/scripts/clone-data-deps.sh; \
	fi

# ensure-tool — canned recipe (#161): guarantee one dependency is on PATH before
# the step that needs it runs. Idempotent: no-op when the tool is present (won't
# fight asdf/gvm/pipx/manual installs). Auto-installs via Homebrew on macOS;
# elsewhere fails fast with guidance, before any costly downstream cascade. The
# three near-identical ensure-* targets (go/cue/uv) collapsed to one parametrized
# recipe once the rule-of-three tripped (ARCH-DRY).
# Usage: $(call ensure-tool,<tool>,<brew-formula>,<why-needed>,<install-noun>,<url>)
#   tool          command probed with `command -v` (e.g. go, cue, uv)
#   brew-formula  Homebrew formula name (usually == tool)
#   why-needed    reason clause for the fail-fast error (NO literal comma — commas
#                 delimit $(call) args; the trailing comma is appended for you)
#   install-noun  what the manual-install hint says to install (e.g. Go 1.26+, CUE)
#   url           canonical install URL for the manual-install hint
define ensure-tool
	@if command -v $(1) >/dev/null 2>&1; then \
	    :; \
	elif command -v brew >/dev/null 2>&1; then \
	    echo "==> $(1) not found — installing via Homebrew (brew install $(2))"; \
	    brew install $(2); \
	else \
	    echo "Error: $(3)," >&2; \
	    echo "  but '$(1)' is not on PATH and Homebrew isn't available to install it." >&2; \
	    echo "  Install $(4) from $(5) and re-run." >&2; \
	    exit 1; \
	fi
endef

# ensure-go / ensure-cue guard hard build-deps of the base layer's OWN build:
# ariadne ships cmd/sdlc and builds it in `tools` (go, #61), and weave compiles
# construct/vocabulary/*.cue via the cue CLI at weave-compile time (#122).
# Pre-sdlc, ariadne assumed only shell + python at the base and provisioned no
# toolchain. nous keeps its own richer toolchain (GPG/gh/…) separately. ensure-uv
# is the one that reaches PAST the base layer's own build — see its note.
.PHONY: ensure-go
ensure-go:
	$(call ensure-tool,go,go,ariadne ships cmd/sdlc and needs the Go toolchain to build it,Go 1.26+,https://go.dev/dl/)

.PHONY: ensure-cue
ensure-cue:
	$(call ensure-tool,cue,cue,the vocabulary layer (construct/vocabulary/*.cue) needs the CUE CLI,CUE,https://cuelang.org/docs/install/)

# ensure-uv — uv backs the Python data plane, and unlike go/cue it is NOT an
# ariadne build-dep: it's a downstream-CONSUMER runtime dep. metis#1 M3 ships
# pure-Python step-types run hermetically via `uv run --project <root> python -m
# metis.steps.<type>`; kbench + future competition workspaces inherit that
# contract. It's provisioned here at the base anyway (operator's call, #161)
# because uv is fast becoming the universal Python toolchain every derivative
# with a Python surface will want — so like go/cue, provision once and every
# consumer inherits it. uv installs its own managed Python, so it needs no system
# python3. (The base-vs-push-down layer-placement rationale lives in #161's Log.)
.PHONY: ensure-uv
ensure-uv:
	$(call ensure-tool,uv,uv,the Python data plane (metis/kbench step-types) runs via uv,uv,https://docs.astral.sh/uv/getting-started/installation/)

# Prereq-only definition — no recipe. Derivatives can `bootstrap: <my-prereq>`
# additively without colliding. Make composes the prereq list; if any
# derivative defines its own recipe for `bootstrap` (e.g. nous's existing
# GPG/install setup), that recipe is what runs after all prereqs. The ensure-*
# targets are listed first so toolchains are provisioned before the cascade in
# serial make; under `make -j` ordering isn't positional, but the go-build
# targets (sdlc-build, build) depend on ensure-go, so the actual compiles still
# wait for it (#61).
bootstrap: ensure-go ensure-cue ensure-uv bootstrap-peers weave tools sdlc-install data-deps

# ── Pre-merge checks ─────────────────────────────────────────────────────────
check: pre-merge

c:
	@scripts/parallel-checks.sh --audit

pre-merge:
	@scripts/parallel-checks.sh

check-%:
	@if [ -x bin/sdlc ]; then \
	    bin/sdlc judge $*; \
	else \
	    scripts/pre-merge-checks.sh $*; \
	fi

# Worktree management targets
# Capture extra argument after worktree (e.g. make worktree feature-x)
ifeq (worktree,$(firstword $(MAKECMDGOALS)))
  WT_NAME := $(wordlist 2,$(words $(MAKECMDGOALS)),$(MAKECMDGOALS))
  ifneq ($(WT_NAME),)
    $(eval $(WT_NAME):;@:)
  endif
endif

# Capture issue number after fetch (e.g. make fetch 42)
ifeq (fetch,$(firstword $(MAKECMDGOALS)))
  FETCH_NUM := $(wordlist 2,$(words $(MAKECMDGOALS)),$(MAKECMDGOALS))
  ifneq ($(FETCH_NUM),)
    $(eval $(FETCH_NUM):;@:)
  endif
endif

# Create a new git worktree in the parent directory.
# Usage: make worktree <name>    — explicit name
#        make worktree            — auto-detect from single untracked issue file
# Delegates to bin/sdlc change-code (with --worktree=yes + the gate-skipping
# flags that preserve this target's pre-#39 quick-and-dirty semantics).
# Falls back to the inline shell logic when the binary isn't built.
#
# For the gated path (structural + plan-quality checks), use
# `sdlc change-code` directly instead of this target.
worktree:
	@if [ -x bin/sdlc ]; then \
	    if [ -n "$(WT_NAME)" ]; then \
	        bin/sdlc change-code --worktree=yes --no-judge --no-structural --name "$(WT_NAME)"; \
	    else \
	        bin/sdlc change-code --worktree=yes --no-judge --no-structural; \
	    fi; \
	else \
	    name="$(WT_NAME)"; \
	    if [ -z "$$name" ]; then \
	        issues=$$(git ls-files --others --exclude-standard -- '$(WF_ISSUES_DIR)/' 2>/dev/null | grep -E '/[0-9]{6}-.*\.md$$'); \
	        count=$$(echo "$$issues" | grep -c . 2>/dev/null || echo 0); \
	        if [ "$$count" -eq 1 ]; then \
	            name=$$(basename "$$issues" .md); \
	            echo "Auto-detected issue: $$name"; \
	        else \
	            echo "Usage: make worktree <name>"; \
	            if [ "$$count" -gt 1 ]; then \
	                echo "Multiple untracked issue files found:"; \
	                echo "$$issues" | sed 's/^/  /'; \
	            fi; \
	            exit 1; \
	        fi; \
	    fi; \
	    if [ -n "$$issues" ] && [ -f "$$issues" ]; then \
	        echo "==> Committing $$issues before creating worktree..."; \
	        git add "$$issues" && \
	        git commit -m "committing issue file before creating worktree" && \
	        git push || echo "  Warning: push failed, continuing with worktree creation"; \
	    fi; \
	    repo_dir=$$(basename "$$(pwd)"); \
	    mkdir -p "../worktree/$$repo_dir"; \
	    git worktree add -b "$$name" "../worktree/$$repo_dir/$$name" HEAD; \
	    echo "Worktree created at ../worktree/$$repo_dir/$$name on branch $$name"; \
	    printf '%s' "../worktree/$$repo_dir/$$name" > .goto; \
	    echo "Run: g (to cd into worktree)"; \
	fi

# Fetch a GitHub issue and create a local issue file in issues/.
# Usage: make fetch <number>
# Delegates to bin/sdlc fetch when the binary is built; falls back to
# the inline shell logic otherwise.
fetch:
	@if [ -z "$(FETCH_NUM)" ]; then \
		echo "Usage: make fetch <number>"; \
		exit 1; \
	fi
# Makefile-level conditional, not a shell `exit 0` — excludes the fallback when
# the binary exists, so make can't fall through to the next recipe line (#36 M2).
ifneq ($(wildcard bin/sdlc),)
	@bin/sdlc fetch --github-issue "$(FETCH_NUM)"
else
	@set -o pipefail; \
	repo=$$(git remote get-url origin | sed 's|.*github.com[:/]\(.*\)\.git|\1|;s|.*github.com[:/]\(.*\)$$|\1|'); \
	gh_title=$$(gh issue view "$(FETCH_NUM)" --repo "$$repo" --json title --jq '.title') || exit 1; \
	gh_body=$$(gh issue view "$(FETCH_NUM)" --repo "$$repo" --json body --jq '.body // ""'); \
	slug=$$(echo "$$gh_title" | tr '[:upper:]' '[:lower:]' | sed 's/[^a-z0-9]/-/g; s/--*/-/g; s/^-//; s/-$$//'); \
	mkdir -p $(WF_ISSUES_DIR); \
	max_id=$$(ls $(WF_ISSUES_DIR)/ $(WF_HISTORY_DIR)/ 2>/dev/null | grep -oE '^[0-9]{6}-' | sed 's/-//' | sort -n | tail -1); \
	next_id=$$(printf '%06d' $$(( $${max_id:-0} + 1 )) ); \
	issue_file="$(WF_ISSUES_DIR)/$${next_id}-$${slug}.md"; \
	today=$$(date +%Y-%m-%d); \
	printf '%s\n' \
		"---" \
		"id: $$next_id" \
		"status: open" \
		"deps: []" \
		"github_issue: $(FETCH_NUM)" \
		"created: $$today" \
		"updated: $$today" \
		"---" \
		"" \
		"# $$gh_title" \
		"" \
		"$$gh_body" \
		"" \
		"## Done when" \
		"" \
		"-" \
		"" \
		"## Plan" \
		"" \
		"- [ ]" \
		"" \
		"## Log" \
		"" \
		"### $$today" \
		"" \
		> "$$issue_file"; \
	echo "Created $$issue_file (GitHub #$(FETCH_NUM))"
endif

# Push to remote, close GitHub issues for done issues, move done issues to history/.
# Works from main — the direct-on-main workflow counterpart to merge.
# Usage: make push
# Delegates to bin/sdlc push when the binary is built; falls back to
# the inline shell logic otherwise (M5 of #31).
push:
# When bin/sdlc is built, the binary IS the whole target — a Makefile-level
# conditional (not a shell `exit 0`) so the legacy fallback below is *excluded*,
# not merely skipped. The old shape ran `bin/sdlc push; exit 0` in the first
# recipe line, but make then ran the next recipe line (the fallback) anyway,
# falling through to the interactive pre-merge-checks.sh (#36 M2).
ifneq ($(wildcard bin/sdlc),)
	@bin/sdlc push $(if $(YES),--yes) $(if $(NO_JUDGE),--no-judge)
else
	@branch=$$(git branch --show-current); \
	if [ "$$branch" != "main" ]; then \
		echo "Error: make push must be run from main (current branch: $$branch)"; \
		exit 1; \
	fi
	@untracked=$$(git ls-files --others --exclude-standard); \
	if [ -n "$$untracked" ]; then \
		echo "  [x] Untracked files found — add or .gitignore them first"; \
		echo "$$untracked" | sed 's/^/       /'; \
		exit 1; \
	fi; \
	dirty=$$(git status --porcelain); \
	if [ -n "$$dirty" ]; then \
		echo "==> Auto-committing tracked changes..."; \
		commit_msg=""; \
		for f in $(WF_ISSUES_DIR)/[0-9][0-9][0-9][0-9][0-9][0-9]-*.md; do \
			[ -f "$$f" ] || continue; \
			if ! git diff --quiet -- "$$f" 2>/dev/null || ! git diff --cached --quiet -- "$$f" 2>/dev/null; then \
				topic=$$(grep -m1 '^# ' "$$f" | sed 's/^# *//'); \
				if [ -n "$$topic" ]; then \
					if [ -n "$$commit_msg" ]; then \
						commit_msg="$$commit_msg"$$'\n'"$$topic"; \
					else \
						commit_msg="$$topic"; \
					fi; \
				fi; \
			fi; \
		done; \
		if [ -z "$$commit_msg" ]; then \
			commit_msg="auto-commit before push"; \
		fi; \
		git commit -a -m "$$commit_msg" || exit 1; \
	fi
	@$(MAKE) pre-merge
	@$(call check_undone_issues,origin/main,$(WF_ISSUES_DIR)) \
	git push || exit 1; \
	repo=$$(git remote get-url origin | sed 's|.*github.com[:/]\(.*\)\.git|\1|;s|.*github.com[:/]\(.*\)$$|\1|'); \
	moved=0; \
	if [ -d $(WF_ISSUES_DIR) ]; then \
		for f in $(WF_ISSUES_DIR)/[0-9][0-9][0-9][0-9][0-9][0-9]-*.md; do \
			[ -f "$$f" ] || continue; \
			status=$$(grep -m1 '^status:' "$$f" | sed 's/^status:[[:space:]]*//'); \
			if [ "$$status" = "done" ] || [ "$$status" = "wontfix" ] || [ "$$status" = "punt" ]; then \
				if [ "$$status" = "done" ]; then \
					gh_num=$$(grep -m1 '^github_issue:' "$$f" | sed 's/^github_issue:[[:space:]]*//'); \
					if [ -n "$$gh_num" ] && [ "$$gh_num" != "" ]; then \
						echo "==> Closing GitHub issue #$$gh_num..."; \
						gh issue close "$$gh_num" --repo "$$repo" --comment "Fixed on main." || true; \
					fi; \
				fi; \
				mkdir -p $(WF_HISTORY_DIR); \
				echo "==> Archiving $$f to $(WF_HISTORY_DIR)/..."; \
				mv "$$f" "$(WF_HISTORY_DIR)/$$(basename $$f)"; \
				moved=1; \
			fi; \
		done; \
	fi; \
	if [ "$$moved" -eq 1 ]; then \
		echo "==> Committing archived history..."; \
		git add $(WF_ISSUES_DIR)/ $(WF_HISTORY_DIR)/ && \
		git commit -m "archive completed issues to history" && \
		git push; \
	fi; \
	echo "Done."
endif

# Create a GitHub pull request from the current worktree branch to main.
# Scans issues/ files touched since branch point for github_issue frontmatter.
# Must be run from inside a worktree (not from main).
# Delegates to bin/sdlc pr when the binary is built; falls back to
# the inline shell logic otherwise (M5 of #31).
pull-request:
# Makefile-level conditional (see push/fetch) — excludes the fallback when the
# binary exists, so make can't fall through to the next recipe line (#36 M2).
ifneq ($(wildcard bin/sdlc),)
	@bin/sdlc pr
else
	@branch=$$(git branch --show-current); \
	if [ -z "$$branch" ] || [ "$$branch" = "main" ]; then \
		echo "Error: run this from a worktree branch, not main"; \
		exit 1; \
	fi; \
	git push -u origin "$$branch" || exit 1; \
	repo=$$(git remote get-url origin | sed 's|.*github.com[:/]\(.*\)\.git|\1|;s|.*github.com[:/]\(.*\)$$|\1|'); \
	base=$$(git merge-base main HEAD 2>/dev/null || echo main); \
	touched_issues=$$(git diff --name-only "$$base"..HEAD -- '$(WF_ISSUES_DIR)/*.md' 2>/dev/null); \
	gh_nums=""; \
	for f in $$touched_issues; do \
		[ -f "$$f" ] || continue; \
		num=$$(grep -m1 '^github_issue:' "$$f" | sed 's/^github_issue:[[:space:]]*//'); \
		if [ -n "$$num" ] && [ "$$num" != "" ]; then \
			gh_nums="$$gh_nums $$num"; \
		fi; \
	done; \
	fixes=""; \
	if [ -n "$$gh_nums" ]; then \
		fixes=$$(echo $$gh_nums | tr ' ' '\n' | sort -u | sed 's/^/#/' | paste -sd ', ' -); \
		fixes="Fixes $$fixes"; \
	fi; \
	commits=$$(git log main..HEAD --pretty=format:'- %s' 2>/dev/null); \
	if [ -n "$$fixes" ]; then \
		echo "Including in PR body: $$fixes"; \
		body="$$commits"; \
		if [ -n "$$body" ]; then \
			body="$$body"$$'\n\n'"$$fixes"; \
		else \
			body="$$fixes"; \
		fi; \
		gh pr create --repo "$$repo" --base main --head "$$branch" --fill-first --body "$$body"; \
	else \
		gh pr create --repo "$$repo" --base main --head "$$branch" --fill; \
	fi
endif

# Merge the current worktree branch into main (if a PR exists),
# move done issues to history/, clean up the worktree.
# Must be run from inside a worktree (not from main).
# Delegates to bin/sdlc merge when the binary is built; falls back to
# the inline shell logic otherwise (M5 of #31).
merge:
# Makefile-level conditional (see push/fetch) — excludes the fallback when the
# binary exists, so make can't fall through to the next recipe line (#36 M2).
ifneq ($(wildcard bin/sdlc),)
	@bin/sdlc merge $(if $(YES),--yes) $(if $(NO_JUDGE),--no-judge)
else
	@branch=$$(git branch --show-current); \
	if [ -z "$$branch" ] || [ "$$branch" = "main" ]; then \
		echo "Error: run this from a worktree branch, not main"; \
		exit 1; \
	fi; \
	echo "==> Branch: $$branch"; \
	uncommitted=$$(git status --porcelain); \
	if [ -n "$$uncommitted" ]; then \
		echo "  [x] Uncommitted changes found — cannot merge"; \
		git status --short; \
		echo "Commit or stash them before merging."; \
		exit 1; \
	fi; \
	echo "  [ok] No uncommitted changes"; \
	upstream=$$(git rev-parse --abbrev-ref --symbolic-full-name @{u} 2>/dev/null || true); \
	if [ -z "$$upstream" ]; then \
		echo "  [x] No upstream configured for $$branch"; \
		echo "Push the branch first (e.g. make pull-request or git push -u origin $$branch)."; \
		exit 1; \
	fi; \
	ahead=$$(git rev-list --count "$$upstream..HEAD" 2>/dev/null || echo 0); \
	if [ "$$ahead" -gt 0 ]; then \
		echo "  [x] Unpushed local commits detected: $$ahead commit(s) ahead of $$upstream"; \
		echo "Push your branch before merging."; \
		exit 1; \
	fi; \
	echo "  [ok] No unpushed local commits (HEAD synced with $$upstream)"
	@$(MAKE) pre-merge \
	wt_path=$$(git rev-parse --show-toplevel); \
	main_path=$$(git worktree list | grep '\[main\]' | awk '{print $$1}'); \
	if [ -z "$$main_path" ]; then \
		echo "  [x] Could not find main worktree — is main checked out?"; \
		exit 1; \
	fi; \
	repo=$$(git remote get-url origin | sed 's|.*github.com[:/]\(.*\)\.git|\1|;s|.*github.com[:/]\(.*\)$$|\1|'); \
	unmerged=$$(git log "main..HEAD" --oneline 2>/dev/null); \
	if [ -n "$$unmerged" ]; then \
		echo "  [ok] Unmerged local commits found:"; \
		echo "$$unmerged" | sed 's/^/       /'; \
	else \
		echo "  [ok] No unmerged local commits (branch is clean)"; \
	fi; \
	$(call check_undone_issues,main,$(WF_ISSUES_DIR)) \
	printf "Final confirmation: proceed with irreversible merge/cleanup actions? [y/N] "; \
	read final_answer; \
	if [ "$$final_answer" != "y" ] && [ "$$final_answer" != "Y" ]; then \
		echo "Aborted."; \
		exit 1; \
	fi; \
	pr_number=$$(gh pr list --repo "$$repo" --head "$$branch" --json number --jq '.[0].number' 2>/dev/null); \
	if [ -n "$$pr_number" ]; then \
		echo "  [ok] Open PR found: #$$pr_number"; \
		echo "==> Merging PR #$$pr_number ($$branch) into main via GitHub..."; \
		gh pr merge --repo "$$repo" --merge --delete-branch "$$branch" || exit 1; \
		echo "==> Pulling main..."; \
		git -C "$$main_path" pull || exit 1; \
	else \
		echo "  [--] No open PR for branch $$branch"; \
		if [ -n "$$unmerged" ]; then \
			printf "Would you like to create a pull request first? [Y/n] "; \
			read answer; \
			if [ "$$answer" != "n" ] && [ "$$answer" != "N" ]; then \
				echo "Run 'make pull-request' to create a PR."; \
				exit 1; \
			fi; \
			printf "Remove worktree without merging? [y/N] "; \
			read answer2; \
			if [ "$$answer2" != "y" ] && [ "$$answer2" != "Y" ]; then \
				echo "Aborted."; \
				exit 1; \
			fi; \
		fi; \
	fi; \
	echo "==> Archiving completed issues to $(WF_HISTORY_DIR)/..."; \
	moved=0; \
	if [ -d "$$main_path/$(WF_ISSUES_DIR)" ]; then \
		for f in "$$main_path"/$(WF_ISSUES_DIR)/[0-9][0-9][0-9][0-9][0-9][0-9]-*.md; do \
			[ -f "$$f" ] || continue; \
			status=$$(grep -m1 '^status:' "$$f" | sed 's/^status:[[:space:]]*//'); \
			if [ "$$status" = "done" ] || [ "$$status" = "wontfix" ] || [ "$$status" = "punt" ]; then \
				mkdir -p "$$main_path/$(WF_HISTORY_DIR)"; \
				echo "  Moving $$(basename $$f) to $(WF_HISTORY_DIR)/"; \
				mv "$$f" "$$main_path/$(WF_HISTORY_DIR)/$$(basename $$f)"; \
				moved=1; \
			fi; \
		done; \
	fi; \
	if [ "$$moved" -eq 1 ]; then \
		echo "==> Committing archived history in main..."; \
		git -C "$$main_path" add $(WF_ISSUES_DIR)/ $(WF_HISTORY_DIR)/ && \
		git -C "$$main_path" commit -m "archive completed issues to history" && \
		git -C "$$main_path" push; \
	fi; \
	echo "==> Removing worktree at $$wt_path..."; \
	git -C "$$main_path" worktree remove "$$wt_path" 2>/dev/null || true; \
	git -C "$$main_path" branch -D "$$branch" 2>/dev/null || true; \
	printf '%s' "$$main_path" > "$$wt_path/.goto"; \
	echo "Done. Run: g (to cd back to main)"
endif

# Warn if any touched issue files are not marked as resolved (done/wontfix/punt).
# Usage: $(call check_undone_issues,<base-ref>,<issues-dir>)
#   base-ref:   git ref to diff against (e.g. origin/main, main)
#   issues-dir: path to issues directory (e.g. $(WF_ISSUES_DIR), $$main_path/$(WF_ISSUES_DIR))
define check_undone_issues
	not_done=""; \
	touched=$$(git diff --name-only $(1)..HEAD -- '$(WF_ISSUES_DIR)/*.md' 2>/dev/null); \
	for f in $$touched; do \
		target="$(2)/$$(basename $$f)"; \
		[ -f "$$target" ] || continue; \
		status=$$(grep -m1 '^status:' "$$target" | sed 's/^status:[[:space:]]*//'); \
		if [ "$$status" != "done" ] && [ "$$status" != "wontfix" ] && [ "$$status" != "punt" ]; then \
			not_done="$$not_done\n  $$f (status: $${status:-unset})"; \
		fi; \
	done; \
	if [ -n "$$not_done" ]; then \
		printf "⚠️  Touched issue files that are NOT done:$$not_done\n"; \
		printf "Continue anyway? [y/N] "; \
		read undone_answer; \
		if [ "$$undone_answer" != "y" ] && [ "$$undone_answer" != "Y" ]; then \
			echo "Aborted."; \
			exit 1; \
		fi; \
	fi;
endef

# ── make build ─────────────────────────────────────────────────────────────
#
# End-user-facing build verb. Convention:
#
#   - cmd/<name>/main.go  →  bin/<name>     (Go binaries; auto-discovered)
#   - Per-binary opt-out via sentinel file:
#       cmd/<name>/.skip-make-build
#     If this file exists, the scanner skips that binary. The base layer
#     doesn't know any derivative's binary names — each opted-out binary
#     drops its own sentinel and documents the rationale inside the file
#     (free-form prose, for future operators).
#
#     Sentinels exist for binaries with distribution semantics that the
#     generic scan would break — e.g., signed + notarized binaries
#     (nous, future charon/gmail) where overwriting bin/<name> with an
#     unsigned local build invalidates macOS keychain ACL grants and
#     notification capabilities. Build those via their own targets
#     (nous-build, etc.) when you need a local copy.
#
#   - For non-Go binaries (Python scripts to chmod, wheels, etc.),
#     define a `local-build` target in Makefile.local — `make build`
#     calls it after the Go build pass.
#
# Designed to be a no-op in repos that don't have a go.mod (brain
# repos without authored binaries), so it's safe to define in the
# shared base layer.
.PHONY: build local-build
# ensure-go prereq (#61): gates the go-build below so it can't race the
# toolchain install under `make -j` (both go-build targets — sdlc-build + build —
# share the one ensure-go node, so make runs it once, first). No-op recipe for
# repos without go.mod regardless.
build: ensure-go
	@if [ -f go.mod ]; then \
	    found=0; \
	    skipped=0; \
	    for d in cmd/*/; do \
	        name=$$(basename "$$d"); \
	        if [ -f "$$d/.skip-make-build" ]; then \
	            echo "  (skipping $$name — .skip-make-build sentinel present)"; \
	            skipped=$$((skipped + 1)); \
	            continue; \
	        fi; \
	        if [ -f "$$d/main.go" ]; then \
	            mkdir -p bin; \
	            echo "==> Building $$name..."; \
	            go build -o "bin/$$name" "./$$d" || exit 1; \
	            found=1; \
	        fi; \
	    done; \
	    if [ "$$found" = "0" ] && [ "$$skipped" = "0" ]; then \
	        echo "  (no cmd/*/main.go to build)"; \
	    fi; \
	fi
	@$(MAKE) --no-print-directory local-build

# local-build is the operator-extensible hook for non-Go binaries
# (shell scripts to chmod, Python wheels, anything else). Default
# no-op; override in Makefile.local. Example:
#
#   # In your Makefile.local
#   local-build:
#   	@chmod +x bin/my-script
#   	@cd python-utils && pip install --user -e .
local-build:
	@:

# ── sdlc binary ──────────────────────────────────────────────────────────────
# `sdlc` is the SDLC checkpoint binary (see workshop/issues/000031-*.md).
# Builds from cmd/sdlc/main.go, output at cmd/sdlc/bin/sdlc, symlinked
# into bin/sdlc. Mirrors ../nous's `nous-build` pattern.
#
# `make build` (the cmd/*/main.go scanner above) also picks sdlc up
# automatically — sdlc-build is the explicit dev-flow target for
# iterating just on the binary without scanning the whole cmd/ tree.
.PHONY: tools sdlc-build weave-build datatype-build vocabulary-build vocab-embed sdlc-install sdlc-bootstrap

# tools: compose all build targets for binaries this repo ships.
# Workflow ships `sdlc-build` (the canonical ariadne tool) + `build`
# (generic cmd/*/main.go scanner). Derivatives can extend additively
# in Makefile.local / Makefile.nous, e.g. `tools: nous-build`.
tools: sdlc-build build

sdlc-build: ensure-go
	@echo "==> building sdlc (build-in-owner)"
	@# Build-in-owner (#60, #95 M5): sdlc's source AND binary live ONLY in its
	@# owner (ariadne). Resolve the owner by LOCATION via dev-aliases.sh --list —
	@# immune to whether ariadne is a direct or transitive ancestor, and needs no
	@# go.mod replace — then build into the OWNER's bin/, NOT this repo's bin/.
	@# So there is exactly one sdlc on disk: $$owner/bin/sdlc. A consumer's
	@# `make tools` writes to ../ariadne/bin/ (the same place the dev-alias
	@# functions build to — the official, gitignored path); it does NOT create a
	@# duplicate consumer-local bin/sdlc. When THIS repo is the owner
	@# ($$owner is this repo's own dir), the build target is unchanged — ariadne's
	@# own bin/sdlc. Under `make bootstrap`, `bootstrap-peers` (clones ancestors) and
	@# `weave` (materializes the construct/dev-aliases.sh symlink) both precede
	@# `tools`, so the resolver and the owner are present by the time this runs.
	@owner="$$(construct/dev-aliases.sh --list 2>/dev/null | awk -F'\t' '$$1=="sdlc"{print $$2}')"; \
	if [ -z "$$owner" ]; then \
	    echo "Error: sdlc owner not found beside this repo." >&2; \
	    echo "  Run 'make bootstrap-peers' (clone ancestors) + 'make weave' first." >&2; \
	    exit 1; \
	fi; \
	mkdir -p "$$owner/bin"; \
	( cd "$$owner" && go build -o "$$owner/bin/sdlc" ./cmd/sdlc )

# weave-build: mirror of sdlc-build for cmd/weave — the intent-compiler that
# replaced construct/setup.sh (#95). weave's source AND binary live ONLY in its
# owner (ariadne); resolve the owner by LOCATION via dev-aliases.sh --list
# (immune to direct-vs-transitive ancestry, needs no go.mod replace), then build
# into the OWNER's bin/ (NOT this repo's bin/) — exactly one weave on disk at
# $$owner/bin/weave. A consumer's `make weave` builds + runs ../ariadne/bin/weave
# and produces NO consumer-local bin/weave (#95 M5, build-in-owner). When THIS
# repo is the owner ($$owner is this repo's own dir) the target is unchanged. The `weave`
# target depends on this, then runs the bare Union `$$owner/bin/weave compile`
# to compile this repo's layer composition (every harness face). Under `make bootstrap`,
# bootstrap-peers + weave's own weave-build prereq guarantee the owner + the
# dev-aliases.sh symlink are present by the time this runs.
weave-build: ensure-go
	@echo "==> building weave (build-in-owner)"
	@owner="$$(construct/dev-aliases.sh --list 2>/dev/null | awk -F'\t' '$$1=="weave"{print $$2}')"; \
	if [ -z "$$owner" ]; then \
	    echo "Error: weave owner not found beside this repo." >&2; \
	    echo "  Run 'make bootstrap-peers' (clone ancestors) first." >&2; \
	    exit 1; \
	fi; \
	mkdir -p "$$owner/bin"; \
	( cd "$$owner" && go build -o "$$owner/bin/weave" ./cmd/weave )

# datatype-build: mirror of weave-build for cmd/datatype — the DAG-aware
# datatype subsystem (#115). datatype is a PATH binary invoked by name (the
# .dynamic-skill marker runs it at weave compile time; agents run `datatype
# list` / `datatype show <name>` for apply-time access). Build-in-owner like
# weave/sdlc: resolve the owner by LOCATION (dev-aliases.sh --list, which scans
# cmd/ dirs and already reports datatype → ariadne), then build into the OWNER's
# bin/ — exactly one datatype on disk at $$owner/bin/datatype, no go.mod replace,
# no consumer-local copy. When THIS repo is the owner, it builds ariadne's own
# bin/datatype, unchanged.
datatype-build: ensure-go
	@echo "==> building datatype (build-in-owner)"
	@owner="$$(construct/dev-aliases.sh --list 2>/dev/null | awk -F'\t' '$$1=="datatype"{print $$2}')"; \
	if [ -z "$$owner" ]; then \
	    echo "Error: datatype owner not found beside this repo." >&2; \
	    echo "  Run 'make bootstrap-peers' (clone ancestors) first." >&2; \
	    exit 1; \
	fi; \
	mkdir -p "$$owner/bin"; \
	( cd "$$owner" && go build -o "$$owner/bin/datatype" ./cmd/datatype )

# vocabulary-build: mirror of datatype-build for cmd/vocabulary — the DAG-aware
# compiler for the formal vocabulary layer (#122). A PATH binary invoked by name
# by the vocabulary .dynamic-skill at weave compile (and by vocab-embed).
# Build-in-owner: resolve the owner by LOCATION (dev-aliases.sh --list) and build
# into the OWNER's bin/. When THIS repo is the owner, builds ariadne's own
# bin/vocabulary, unchanged.
vocabulary-build: ensure-go
	@echo "==> building vocabulary (build-in-owner)"
	@owner="$$(construct/dev-aliases.sh --list 2>/dev/null | awk -F'\t' '$$1=="vocabulary"{print $$2}')"; \
	if [ -z "$$owner" ]; then \
	    echo "Error: vocabulary owner not found beside this repo." >&2; \
	    echo "  Run 'make bootstrap-peers' (clone ancestors) first." >&2; \
	    exit 1; \
	fi; \
	mkdir -p "$$owner/bin"; \
	( cd "$$owner" && go build -o "$$owner/bin/vocabulary" ./cmd/vocabulary )

# vocab-embed (#122 M3): regenerate the COMMITTED Go-binding embed inputs from the
# vocabulary via `go generate`, then assert nothing drifted. The Go binding lives in
# pkg/vocab — ONE shared package every Go consumer imports (the import graph is the
# distribution; no per-consumer copy). The co-located //go:generate is the binding
# declaration, so this target is GENERIC over nouns/consumers: adding a noun is a
# go:generate line in pkg/vocab, never a new Make target (supersedes the per-entity
# issue-json-gen/check). OWNER-ONLY (pkg/vocab + cmd/vocabulary live in ariadne) —
# run from ariadne CI; deliberately NOT wired into `bootstrap`/`sdlc-build`/the
# consumer `check` (a consumer needs neither cue nor regeneration; the json is
# committed so a standalone `go build` works). The DIFFERENT cross-repo gate — has
# this repo's gitignored materialization gone stale vs the merged source — remains
# `vocabulary check --output construct/generated/vocabulary`.
vocab-embed: vocabulary-build ensure-cue
	@echo "==> regenerating pkg/vocab embed inputs (go generate)"
	@vcowner="$$(construct/dev-aliases.sh --list 2>/dev/null | awk -F'\t' '$$1=="vocabulary"{print $$2}')"; \
	PATH="$$vcowner/bin:$$PATH" go generate ./pkg/vocab/...
	@git diff --exit-code -- pkg/vocab \
	  || { echo "Error: pkg/vocab embed inputs are STALE vs construct/vocabulary/*.cue — run 'make vocab-embed' and commit (#122)." >&2; exit 1; }

# sdlc-install puts the in-tree bin/sdlc on the developer's PATH by
# appending $REPO_DIR/bin to the shell rc (zsh/bash). Idempotent; also
# prints the export line so the user can paste it manually as backup.
#
# Wired into `make bootstrap` so a single bootstrap gesture builds
# sdlc + makes it available in new shells. Mirrors nous's PATH-append
# convention; the old `~/bin` symlink approach was retired so all
# repo `bin/` dirs (ariadne, nous, …) compose uniformly on PATH.
#
# `sdlc-bootstrap` stays as a backward-compat alias for pre-rename
# muscle memory; will be removed once docs + downstream repos catch up.
sdlc-install sdlc-bootstrap:
	@scripts/sdlc-install.sh
