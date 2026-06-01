# sdlc binary

`sdlc` is the SDLC checkpoint binary — one Go binary at `cmd/sdlc/`
that collects ariadne's workflow checkpoint guards into a unified verb
namespace with embedded `--help` per subcommand. Agents should invoke
`sdlc` directly; Makefile targets are compatibility wrappers for
downstream repos that have not built the binary yet.

Design rationale: `docs/vision/2026-05-25-01-pensive-sdlc-checkpoint-binary.md`.
Build issue + plan: `workshop/history/000031-sdlc-checkpoint-binary.md`
(after archive) or `workshop/issues/000031-...` (during build).

## What it owns

The **checkpoints between SDLC stages** — not the stages themselves.
Stages stay prose; the binary refuses transitions that lack required
evidence. Subcommands are added incrementally when the same drift
recurs at a stage (not by formalizing the SDLC as a state machine).

## Verb surface

| Verb              | Replaces (Make target)      | Defends |
|-------------------|-----------------------------|---------|
| `close`           | `make close-issue`          | Issue close: actual + verified + atlas + plan ticked |
| `state`           | (new)                       | Workflow state inspection + drift detection |
| `judge`           | `make check-{dry,pure,plan,specs,lessons}` | Fresh-context LLM judge (anti-collusion) |
| `fetch`           | `make fetch N`              | **Hidden deprecated alias** for `sdlc issue new --from-github` since #56 M2 (keeps `--github-issue`) |
| `claim`           | `make issue-sync`           | Issue-file workstream-claim onto main (formerly `lock`, #39) |
| `change-code`     | `make worktree` (partial)   | Planning → implementation gate: structural + plan-quality + branching (in-place default, `--worktree=yes`/`=ask`; #39, #51) |
| `set-status`      | (new)                       | Status-transition guards. Moved under `sdlc issue set-status` (#56 M2); **hidden deprecated flat alias** kept one cycle |
| `push`            | `make push`                 | Direct-on-main ship + pre-flight judges (still available; not the default close path since #51) |
| `pr`              | `make pull-request`         | PR creation with Fixes-issue body |
| `merge`           | `make merge`                | Branch merge (in-place or worktree) via PR + cleanup + irreversible-action confirm (#51) |
| `milestone-close` | `make close-issue MILESTONE=Mx` | Milestone close + auto-dispatched milestone-review |
| `issue new`       | (new; xx-issues skill prose)| Allocate next ID + write canonical template (`--from-github N` seeds from GitHub) |
| `issue set-status`| ← flat `set-status`         | Status-transition guards (relocated #56 M2) |
| `issue list`      | (new)                       | List issues (ID/status/title), sorted by ID; `--status` filters; reuses `listIssues` |
| `issue show`      | (new)                       | Issue frontmatter + section headers, no bodies |

**Flat verbs vs the `issue` group (#56).** The flat verbs guard workflow
*transitions* (close, claim, change-code, pr, merge, …). `sdlc issue *` is the
CRUD/authoring surface for the issue *record* — the noun-grouped home for
`new` (and, post-#56-M2, `set-status`/`list`/`show`). The canonical issue-file
template lives in one place: the `Render` function in `internal/issue/scaffold.go`,
documented in prose by `sdlc issue --help`.

## Progressive disclosure

  - `sdlc --help` — the workflow contract (start-of-work runbook, conventions,
    cobra-generated verb list)
  - `sdlc <verb> --help` — per-checkpoint contract + flags + examples
  - `sdlc state` — runtime "where am I" surface for compaction recovery

`sdlc --help` is the single source of truth for the workflow contract.
`construct/local/sdlc/SKILL.md` (the `xx-sdlc` skill) is a **static pointer**
to it — it carries no copy of the contract, so it can't drift. The old
`sdlc --index` regenerator was retired once the skill stopped duplicating the
help text.

## Architecture

```
cmd/sdlc/
  main.go              cobra root + verb registration
  term.go              cinfo / cok / cwarn / die + env helpers (shared)
  runner.go            gitRunner interface + execGitRunner impl (shared)
  ghclient.go          ghCaller interface + realGH impl (shared)
  preflight.go         runPreflightJudges (push + merge pre-flight)
  close.go             ← scripts/close-issue.py
  state.go             new (read-only inspection + drift detection)
  judge.go             ← scripts/pre-merge-checks.sh
  fetch.go             thin hidden alias → runIssueNew --from-github (#56 M2)
  issue.go             new (#56): `sdlc issue` group — new / set-status / list / show
  start.go             migration stub (REMOVED in #39 — errors with
                       "use claim + change-code")
  claim.go             ← scripts/issue-sync.sh (renamed from lock.go #39)
  changecode.go        new (#39): planning → implementation gate
  branchcreate.go      new (#39): branch-creation helpers shared by
                       changecode.go (worktree + in-place paths) + the
                       name-resolution previously in start.go
  setstatus.go         new
  push.go              ← Makefile push:
  pr.go                ← Makefile pull-request:
  merge.go             ← Makefile merge:
  milestoneclose.go    composition over close + judge milestone-review
  helptext/            //go:embed *.md — one .md per verb + root
  internal/
    gitx/              git invocation seam (`run` shim, Capture, DiffBase,
                       CommitWindow, DiscoverWindowIssues, RunGit)
    issue/             frontmatter parse/edit + plan-section regexes +
                       scaffold.go (NextID/Slugify/Render — #56)
    judge/             Category enum, prompt builder, classify, dispatch
    project/           brain project-file mutation helpers
```

## Anti-collusion + form-vs-essence

Checkpoint guards defend against **omission** (claiming done without
doing) via deterministic checks (`close` refuses without `--actual` +
`--verified`). The judge subcommand defends against **theater** (form
without substance) via fresh-context LLM review — every Dispatch call
spawns a new subprocess; the agent has no doer-session state.

`push` and `merge` auto-dispatch `judge plan|specs|lessons` as pre-
flight so the checks run consistently rather than as a remembered
manual step. `milestone-close` auto-dispatches `judge milestone-review`
as a post-action.

**Judge → classifier contract.** Plan + Specs subagents must emit
`VERDICT: CLEAN | INFO | FAILURE` as line 1 of their response; the
classifier keys off that. (MilestoneReview uses the parallel
`SHIP | FIX-THEN-SHIP | REWORK`; Lessons emits a fixed REMINDER line
and skips the agent entirely.) A legacy sentinel-grep
(`no DRY/PURE violations found`, `in sync`, …) remains as a fallback
for outputs that don't carry the verdict line, but new prompts should
use the structured form — free-text approval prose otherwise scores
`FAILURE` and blocks the merge.

## Build + install

```
make sdlc-build        builds cmd/sdlc/bin/sdlc, symlinks bin/sdlc
make sdlc-bootstrap    one-shot install: verify Go, build, symlink to
                       $SDLC_INSTALL_BIN (default ~/bin)
```

`make build` also picks `sdlc` up via the cmd/*/main.go scanner.

### Downstream staleness gotcha

Downstream repos ship a *prebuilt* `bin/sdlc` (built from ariadne via the
`construct/go.mod` `replace => ../ariadne` path) — they have no `cmd/sdlc`
source of their own. That binary does **not** auto-rebuild when the base-layer
tool changes; it goes stale until the operator reruns `make sdlc-build`
(or `make sdlc-install`) in the downstream repo. A stale binary silently
lacks new behavior and can fail in confusing ways — e.g. a pre-#51 binary
hits the in-place branch flow's `git rev-parse --git-dir` → `.git` path but
still routes to the worktree topology, dying with `find main worktree: could
not find a worktree on branch 'main'`. Surfaced live by the #51 dogfood
(ariadne #53 Phase B): you-decide's binary was a month stale. **Rule:** after
any base-layer `cmd/sdlc` change reaches downstream, rerun `make sdlc-build`
there before relying on the new verb behavior.

## Makefile wrappers (transition state)

Each Make target delegates to `bin/sdlc` when built, falling back to
the original shell logic when absent:

  `make close-issue` → `sdlc close`
  `make fetch <N>`   → `sdlc fetch --github-issue N` (deprecated alias →
                       `sdlc issue new --from-github N`, #56 M2)
  `make worktree`    → `sdlc change-code --worktree=yes --no-judge --no-structural`
                       (post-#39; preserves the make target's pre-existing
                       quick-and-dirty semantics)
  `make issue-sync`  → `sdlc claim` (renamed from `sdlc lock` in #39)
  `make push`        → `sdlc push`
  `make pull-request` → `sdlc pr`
  `make merge`       → `sdlc merge`
  `make check-<cat>` → `sdlc judge <cat>`

The fallback exists so downstream repos that vendor `Makefile.workflow`
but haven't yet run `make sdlc-build` keep working. M8 (not yet started)
deprecates the shell fallbacks and removes the scripts.

## When to add a new verb

The rule from the pensive: **when the same drift gets caught at review
twice**, promote it from a `workshop/lessons.md` entry to an `sdlc <verb>`
check. The first time → prose. The second time → code.

Do not formalize the workflow into a state machine. Add checkpoint
guards for known commit moments where drift recurs; everything between
checkpoints stays prose-driven.
