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
| `close`           | `make close-issue`          | Issue close: actual + verified + atlas + plan ticked; on full-issue close auto-dispatches the one boundary review (#69, `--no-judge` to skip) |
| `actual`          | (new #68)                   | Compute an issue's focused dev-hours (runs active-time-v3 with brain+repo transcript dirs) |
| `state`           | (new)                       | Workflow state inspection + drift detection |
| `judge`           | `make check-{dry,pure,plan,specs,lessons}` | Fresh-context LLM judge (anti-collusion) |
| `fetch`           | `make fetch N`              | **Hidden deprecated alias** for `sdlc issue new --from-github` since #56 M2 (keeps `--github-issue`) |
| `claim`           | `make issue-sync`           | Issue-file workstream-claim onto main (formerly `lock`, #39) |
| `start-plan`      | (new #75)                   | Planning-entry transition: delivers the `at-plan` architecture lens to design against |
| `change-code`     | `make worktree` (partial)   | Planning → implementation gate: structural + plan-quality + branching (in-place default, `--worktree=yes`/`=ask`; #39, #51) |
| `set-status`      | (new)                       | Status-transition guards. Moved under `sdlc issue set-status` (#56 M2); **hidden deprecated flat alias** kept one cycle |
| `push`            | `make push`                 | Direct-on-main ship + pre-flight judges (still available; not the default close path since #51) |
| `pr`              | `make pull-request`         | PR creation with Fixes-issue body |
| `merge`           | `make merge`                | Branch merge (in-place or worktree) via PR + cleanup + irreversible-action confirm (#51) |
| `milestone-close` | `make close-issue MILESTONE=Mx` | Milestone close + auto-dispatched boundary review (the one reviewer, per-milestone window; #69) |
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
  actual.go            new (#68): runs active-time-v3 → suggested --actual
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

**Per-gate bypass (#67).** `close` has 8 gates (actual, verified, atlas,
milestone-verdict, plan-unchecked, project, re-close, and the #69 boundary
review), each with its own `--no-<gate>` flag (`--no-actual`, `--no-verified`,
`--no-atlas`, `--no-verdict`, `--no-plan-check`, `--no-project`,
`--no-reclose-guard`, `--no-judge`);
`closeFlags.skip(gate)` is the single arbiter (`Force || the field`). A
per-gate flag is an *acknowledgment* that one guard doesn't apply (e.g. a
pure bugfix → `--no-atlas`); it logs an audit `[!]` line and only fires
when the gate would actually have refused. `--force` waives all at once.
`milestone-close` forwards the same flags into its delegated `runClose`.
The convention generalizes `merge`'s pre-existing `--no-judge`.

**Measured actuals (#68).** `--actual` is computed, not hand-typed. `sdlc actual
--issue N` (engine in `actual.go`, shared with close's missing-`--actual`
explainer) runs `construct/local/issues/active-time-v3.py` over the issue's
`CommitWindow` + `DiscoverWindowIssues` peers, feeding it **brain + the issue's
repo** transcript dirs (`~/.claude/projects/<cwd-encoded>`) — the validated
heuristic (events come only from transcripts; the wrong/missing dirs were why
actuals read 0 and got faked). v3's exit codes are the contract: **2** = no
`--dir` (misinvocation), **3** = commits-but-0-events (telemetry gap → labeled
judgment), **0 + a `#N: h.hh hr` line** = measured. Dir-selection is deliberately
narrow (NOT all folders) — an unrelated concurrently-edited repo inflates the
count. `WindowCapDays` is 61 (was 31) so month-long issues keep their window.

`push` and `merge` auto-dispatch `judge plan|specs|lessons` as pre-
flight so the checks run consistently rather than as a remembered
manual step. `milestone-close` auto-dispatches `judge milestone-review`
as a post-action.

**Judges are read-only (#62).** A judge is a reviewer, not a doer — all
categories run with a read-only tool allowlist (`Read,Grep,Glob,Bash`); they
report findings and the main agent (full context) applies fixes. The `specs`
judge used to auto-edit stale docs (`Edit,Write`), which let a *passing* gate
leave the tree dirty and strand the subsequent merge. `merge` now also (a)
re-asserts a clean tree immediately before the irreversible `gh pr merge`
(refuse, don't strand), and (b) resumes an interrupted merge — a re-run detects
an already-merged PR and finishes the local cleanup instead of erroring.
Both behaviors have e2e regression coverage (#63, `merge_e2e_test.go`): a
`tempRepo(t)` harness runs `runMerge` against a real throwaway repo + local bare
origin (in-place topology), so switch/pull/archive/branch-delete execute for
real. The unlock is a trio of `func`→`var` test seams — `die` (term.go,
swapped for panic+recover via `expectDie`), `detectRepo` (fetch.go), and
`runPreflightJudgesFn` (merge.go step 5, used to inject a tree-dirtying
"judge"). `expectDie` is the reusable pattern for testing any `run*` verb's
refusal path.

**Judge → classifier contract (#70).** One contract, both sides reference it:
the human mirror is `construct/judge-output-contract.md`; the Go source of truth
is `internal/judge/contract.go` (`ContractTokens`, `Blocking()`, and
`ContractPreamble` — the format snippet every agent prompt embeds). Every
agent-emitting judge leads its response with `VERDICT: <TOKEN> (confidence: …)`;
the classifier (`ParseVerdictToken`) scans for that token **anywhere** (tolerating
a preamble — the `VERDICT:` prefix + a trailing precision guard make prose
false-positives near-impossible) and gates on the token's **blocking-ness**, never
on prose presence. Tokens: `CLEAN INFO SHIP FIX-THEN-SHIP` pass; `FAILURE REWORK
BLOCK` block. Lessons is the exception — a fixed `REMINDER:` line, no agent. This
killed the bug where a `VERDICT: CLEAN` behind a title scored `FAILURE` and blocked
the merge (and the milestone `unknown`). A thin legacy sentinel-grep remains for
un-migrated/foreign outputs; a `judge_test.go` drift test keeps the doc + Go
tokens in sync.

**Architecture principles (#75).** Injected architectural taste the model lacks
(payoff shows months downstream → no training signal). `internal/judge/
architecture.md` is the single source — markered `ARCH-*` entries, each with a
`principle` / `at-plan` / `at-review` lens — `//go:embed`'d as
`ArchitectureRegistry` and delivered verbatim into the prompts that need it (one
file, embedded per fresh context). Today: the **plan-quality** judge renders the
`at-plan` lens (highest leverage — the design is still changeable), the
**milestone-review** judge renders `at-review` (backstop), and the standalone
**dry/pure** judges render their principle from the registry (authored once).
Cite the marker (`ARCH-DRY`) in plans/Logs/findings. Adding an `ARCH-*` entry
flows into every consumer with no other edit. **`sdlc start-plan`** (#75 M2)
delivers the `at-plan` lens to the main thread at design time — the forward
counterpart to `change-code`'s plan-quality review (`claim → start-plan →
change-code`); a drift test keeps AGENTS.md's narrative in sync with the markers.
#71 adds `ARCH-SHIM`.

**One boundary review, binary-owned (#69).** The *procedure* and the *principles*
are separate embedded sources: `internal/judge/code-review.md` (`//go:embed`'d as
`codeReviewTemplate`, rendered by `CodeReviewBody`) is the **one reviewer prompt**
— the superpowers quality/testing/readiness checklist reconciled with ariadne's
Core-concepts cross-check, Atlas gate, severity buckets, and the
`SHIP|FIX-THEN-SHIP|REWORK` verdict. It *refers* to the ARCH-* markers (the
`{{ARCH_STAR}}` token expands to the live marker list via `ArchitectureMarkers()`,
the single extraction site shared with the AGENTS.md drift test); the principle
*definitions* arrive co-present from `ArchitectureBlock("at-review")` at dispatch.
The procedure must not inline principle bodies (a guardrail test pins this). Both
`milestone-close` (per-milestone window) and `close` (whole-issue / end-of-issue
window) dispatch this same review — so the agent does **not** run a separate
`superpowers-requesting-code-review` pass at a boundary (AGENTS.md §3); that skill
remains for ad-hoc/in-session reviews. The double-review #69 removed was the
agent's superpowers pass *plus* the binary's auto-dispatch on the same diff.

**Window base — prior review boundary (#58).** `boundaryWindowBase`
(`milestoneclose.go`) is the single source for *both* the atlas-coverage gate
(`runClose`) and the boundary review's window, so they provably cover the same
commits (ARCH-DRY). A milestone window bases on the **previous review boundary**
— the most recent prior commit touching the issue file that carries a
`Review-Verdict:` trailer (the prior milestone close), found by
`previousReviewBoundary` — not on the first `#N Mx` commit. This closes a gap
where an inter-milestone `#N`-but-not-`Mx` commit (a `side-quest:`, a fix) landed
between M(x-1)'s close and Mx's first commit would slip *both* windows and escape
review. The first milestone (no prior boundary) and the whole-issue close fall
back to the branch start (parent of the first `#N` commit). If a prior close's
trailer was never pasted, the lookup finds nothing and falls back to branch start
— over-covering rather than under-covering, the safe direction.

## Build + install

```
make ensure-go         guarantee the Go toolchain (#61): no-op if present,
                       brew-installs on macOS, else fails fast with go.dev/dl
make sdlc-build        builds bin/sdlc (build-in-owner since #60 — see gotcha);
                       depends on ensure-go
make sdlc-install      build + append the repo's bin/ to the shell PATH
                       (`sdlc-bootstrap` is a back-compat alias)
```

`make build` also picks `sdlc` up via the cmd/*/main.go scanner.

**Go is a base-layer build dependency (#61).** ariadne ships `cmd/sdlc` and
compiles it in `tools`, so `bootstrap` provisions Go up front (`ensure-go` is its
first prerequisite) — before the peer-clone cascade and the recursive ariadne
bootstrap's tool build. Pre-sdlc, ariadne needed only shell + python (always
present), so bootstrap never provisioned a toolchain; #61 closed that gap. nous
owns its richer toolchain (Homebrew/GPG/gh/…) separately.

### Downstream staleness gotcha

Downstream repos ship a *prebuilt* `bin/sdlc` — they have no `cmd/sdlc` source
of their own. As of #60 `make sdlc-build` resolves sdlc's **owner** (ariadne) by
location via `construct/dev-aliases.sh --list`, then builds in the owner's own
module into `<repo>/bin/sdlc` (build-in-owner) — no `construct/go.mod`
`replace` needed. That binary does **not** auto-rebuild when the base-layer
tool changes; it goes stale until the operator reruns `make sdlc-build`
(or `make sdlc-install`) in the downstream repo. A stale binary silently
lacks new behavior and can fail in confusing ways — e.g. a pre-#51 binary
hits the in-place branch flow's `git rev-parse --git-dir` → `.git` path but
still routes to the worktree topology, dying with `find main worktree: could
not find a worktree on branch 'main'`. Surfaced live by the #51 dogfood
(ariadne #53 Phase B): you-decide's binary was a month stale. **Rule:** after
any base-layer `cmd/sdlc` change reaches downstream, rerun `make sdlc-build`
there before relying on the new verb behavior.

### `sdlc push` archive recovery

`sdlc push` has a recovery preflight before its normal untracked-file guard.
If a previous push already moved terminal-status issue files from
`workshop/issues/` to `workshop/history/` but failed before the archive commit,
the working tree contains deleted issue files plus untracked or staged history
files. That state is otherwise indistinguishable from forbidden untracked files
to the generic guard, so `push` recognizes exact prepared archive pairs,
verifies the history copy still has a terminal status, stages `issues/` and
`history/`, commits "archive completed issues to history", pushes, and exits
without rerunning judges against an archive-only retry. Any unrelated dirty file
keeps the refusal path and tells the operator to clear that unrelated work before
rerunning `sdlc push --yes`.

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
