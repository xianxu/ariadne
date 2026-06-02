# AI issue-based workflow — include from your project Makefile:
#   include Makefile.workflow

# Include openshell targets if available
-include .openshell/Makefile

# Include tart targets if available (macOS VM testing — Apple Silicon)
-include .tart/Makefile
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

.PHONY: help-workflow worktree fetch push pull-request merge check pre-merge refresh issue-sync

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
	"    make refresh        Re-run $(UPSTREAM_NAME) setup (link + merge settings)" \
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

# ── Refresh / bootstrap ───────────────────────────────────────────────────────
# Two verbs, distinct concerns (per ariadne#38):
#
#   make refresh         Pure substrate-state sync. Verifies peers are present,
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
#   make bootstrap       Composition: bootstrap-peers + refresh + tools.
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

refresh:
	@if [ -x construct/setup.sh ]; then \
		construct/setup.sh; \
	else \
		echo "Error: construct/setup.sh not found in this repo."; \
		echo "  First-time bootstrap: run \`../ariadne/construct/setup.sh\` manually."; \
		echo "  After that, \`make refresh\` (and \`make bootstrap\`) will work — the script vendors itself."; \
		exit 1; \
	fi

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

# ensure-go — guarantee the Go toolchain before anything compiles sdlc. ariadne
# ships cmd/sdlc and builds it in `tools`, so go is a hard build-dependency of
# the base layer itself (pre-sdlc, ariadne needed only shell + python, so
# bootstrap never provisioned a toolchain). #61. Idempotent: no-op when go is
# present (won't fight gvm/asdf/manual installs). Auto-installs via Homebrew on
# macOS; elsewhere fails fast with guidance, before the costly peer-clone cascade.
# nous keeps its own richer toolchain (GPG/gh/…) separately; this guarantees only
# the one dep the base layer's own build needs.
.PHONY: ensure-go
ensure-go:
	@if command -v go >/dev/null 2>&1; then \
	    :; \
	elif command -v brew >/dev/null 2>&1; then \
	    echo "==> go not found — installing via Homebrew (brew install go)"; \
	    brew install go; \
	else \
	    echo "Error: ariadne ships cmd/sdlc and needs the Go toolchain to build it," >&2; \
	    echo "  but 'go' is not on PATH and Homebrew isn't available to install it." >&2; \
	    echo "  Install Go 1.26+ from https://go.dev/dl/ and re-run." >&2; \
	    exit 1; \
	fi

# Prereq-only definition — no recipe. Derivatives can `bootstrap: <my-prereq>`
# additively without colliding. Make composes the prereq list; if any
# derivative defines its own recipe for `bootstrap` (e.g. nous's existing
# GPG/install setup), that recipe is what runs after all prereqs. ensure-go is
# listed first so go is provisioned before the cascade in serial make; under
# `make -j` ordering isn't positional, but both go-build targets (sdlc-build,
# build) depend on ensure-go, so the actual compiles still wait for it (#61).
bootstrap: ensure-go bootstrap-peers refresh tools sdlc-install data-deps

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
	@if [ -x bin/sdlc ]; then \
	    bin/sdlc fetch --github-issue "$(FETCH_NUM)"; \
	    exit 0; \
	fi
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

# Push to remote, close GitHub issues for done issues, move done issues to history/.
# Works from main — the direct-on-main workflow counterpart to merge.
# Usage: make push
# Delegates to bin/sdlc push when the binary is built; falls back to
# the inline shell logic otherwise (M5 of #31).
push:
	@if [ -x bin/sdlc ]; then \
	    bin/sdlc push $(if $(YES),--yes) $(if $(NO_JUDGE),--no-judge); \
	    exit 0; \
	fi
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

# Create a GitHub pull request from the current worktree branch to main.
# Scans issues/ files touched since branch point for github_issue frontmatter.
# Must be run from inside a worktree (not from main).
# Delegates to bin/sdlc pr when the binary is built; falls back to
# the inline shell logic otherwise (M5 of #31).
pull-request:
	@if [ -x bin/sdlc ]; then \
	    bin/sdlc pr; \
	    exit 0; \
	fi
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

# Merge the current worktree branch into main (if a PR exists),
# move done issues to history/, clean up the worktree.
# Must be run from inside a worktree (not from main).
# Delegates to bin/sdlc merge when the binary is built; falls back to
# the inline shell logic otherwise (M5 of #31).
merge:
	@if [ -x bin/sdlc ]; then \
	    bin/sdlc merge $(if $(YES),--yes) $(if $(NO_JUDGE),--no-judge); \
	    exit 0; \
	fi
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
.PHONY: tools sdlc-build sdlc-install sdlc-bootstrap

# tools: compose all build targets for binaries this repo ships.
# Workflow ships `sdlc-build` (the canonical ariadne tool) + `build`
# (generic cmd/*/main.go scanner). Derivatives can extend additively
# in Makefile.local / Makefile.nous, e.g. `tools: nous-build`.
tools: sdlc-build build

sdlc-build: ensure-go
	@mkdir -p bin
	@echo "==> building bin/sdlc (build-in-owner)"
	@# Build-in-owner (#60): sdlc's source lives ONLY in its owner (ariadne).
	@# Resolve the owner by LOCATION via dev-aliases.sh --list — immune to whether
	@# ariadne is a direct or transitive ancestor, and needs no go.mod replace —
	@# then build into THIS repo's bin/ using the owner's own module. Replaces the
	@# old `cd construct && go build … through construct/go.mod`. Under `make
	@# bootstrap`, `bootstrap-peers` (clones ancestors) and `refresh` (materializes
	@# the construct/dev-aliases.sh symlink) both precede `tools`, so the resolver
	@# and the owner are present by the time this runs.
	@repo_root="$$(pwd)"; \
	owner="$$(construct/dev-aliases.sh --list 2>/dev/null | awk -F'\t' '$$1=="sdlc"{print $$2}')"; \
	if [ -z "$$owner" ]; then \
	    echo "Error: sdlc owner not found beside this repo." >&2; \
	    echo "  Run 'make bootstrap-peers' (clone ancestors) + 'make refresh' first." >&2; \
	    exit 1; \
	fi; \
	( cd "$$owner" && go build -o "$$repo_root/bin/sdlc" ./cmd/sdlc )

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
