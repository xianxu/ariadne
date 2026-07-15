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
| `actual`          | (new #68)                   | Compute an issue's focused dev-hours (in-binary active-time-v3 engine over brain+repo transcript sources) |
| `active-time`     | (new #110; was active-time-v3.py) | Standalone CLI over the same engine — the per-segment attribution table for manual inspection; preserves the 2/3/0 loud-fail exit codes |
| `state`           | (new)                       | Workflow state inspection + drift detection |
| `resolve`         | (new #144)                  | **Read-only** symbolic-ref → current path(s): the issue + its plan/review family, archive-correct + cross-repo. Locations from the `discovery:` model; grammar single-sourced as the parser. No lock (see below) |
| `open`            | (new #144)                  | Sugar over `resolve`: open the primary artifact in `$EDITOR` |
| `propagate-base`  | (new #106; precheck #109)   | Re-weave every recursive DEPENDENT of this repo (downstream counterpart to `substrateChain`): discover dependents (Makefile.workflow + substrate chain), order foundation-first, then per repo a clean-tree precheck → `make weave` + verify-complete + commit (untracking now-generated files). A dependent with a DIRTY working tree (pre-existing uncommitted work — e.g. a concurrent session) is SKIPPED untouched (never `git add -A`'d) and the run exits non-zero. `--dry-run`/`--ref`. |
| `judge`           | `make check-{dry,pure,plan,specs,lessons}` | Fresh-context LLM judge (anti-collusion) |
| `fetch`           | `make fetch N`              | **Hidden deprecated alias** for `sdlc issue new --from-github` since #56 M2 (keeps `--github-issue`) |
| `claim`           | `make issue-sync`           | Issue-file workstream-claim onto main (formerly `lock`, #39) |
| `start-plan`      | (new #75)                   | Planning-entry transition: delivers the `at-plan` architecture lens + the durable-plan pointer (`superpowers-writing-plans` → `workshop/plans/`, #72) to design against |
| `change-code`     | `make worktree` (partial)   | Planning → implementation gate: structural + estimate (#113) + **estimate-reconciliation + estimate-quality (#117)** + plan-quality + branching (in-place default, `--worktree=yes`/`=ask`; #39, #51) |
| `set-status`      | (new)                       | Status-transition guards. Moved under `sdlc issue set-status` (#56 M2); **hidden deprecated flat alias** kept one cycle |
| `push`            | `make push`                 | Direct-on-main ship + the #124 instance-conformance gate (`--no-validate`) + pre-flight judges (still available; not the default close path since #51) |
| `pr`              | `make pull-request`         | PR creation with Fixes-issue body |
| `merge`           | `make merge`                | Branch merge (in-place or worktree) via PR + the #124 instance-conformance gate (`--no-validate`) + cleanup + irreversible-action confirm (#51) |
| `milestone-close` | `make close-issue MILESTONE=Mx` | Milestone close + auto-dispatched boundary review (the one reviewer, per-milestone window; #69). THE milestone-close path — `close` refuses `--milestone` (#146); `--no-judge` here is the labeled skip-review escape. |
| `issue new`       | (new; xx-issues skill prose)| Allocate next ID + write canonical template (`--from-github N` seeds from GitHub) |
| `issue set-status`| ← flat `set-status`         | Status-transition guards (relocated #56 M2) |
| `issue list`      | (new)                       | List issues (ID/status/title), sorted by ID; `--status` filters; reuses `listIssues` |
| `issue show`      | (new)                       | Issue frontmatter + section headers, no bodies |
| `issue validate`  | (new #124)                  | Validate issue file(s) against `#Issue` — frontmatter cue-vet (via `vocabulary validate-instance`) + section presence; multi-target (#133): `<file>...` / `--issue N[,N...]` / `--all` (mutually exclusive). The on-demand surface of the instance-conformance loop |

**Flat verbs vs the `issue` group (#56).** The flat verbs guard workflow
*transitions* (close, claim, change-code, pr, merge, …). `sdlc issue *` is the
CRUD/authoring surface for the issue *record* — the noun-grouped home for
`new` (and, post-#56-M2, `set-status`/`list`/`show`). The canonical issue-file
template lives in one place: the `Render` function in `internal/issue/scaffold.go`,
documented in prose by `sdlc issue --help`.

## Repo transaction lock (#132)

Mutating `sdlc` verbs are serialized by an SDLC-owned local transaction lock at
`.git/sdlc.lock`. Most mutating verbs hold the lock for the whole command
transaction, not just individual Git calls: issue ID allocation, issue/status
file writes, commits, branch changes, local archive moves, and pushes all run
under the same holder.
The lock directory is created atomically with `mkdir`; holder metadata lives in
`meta.json` inside the directory and records pid, hostname, cwd, command, argv,
and start time.

**Test hermeticity (#149/#165).** Because the lock path (and other repo state)
resolves from cwd via `git rev-parse`, a `cmd/sdlc` test that drives a mutating verb
through `buildRoot().Execute()` without chdir-ing into a temp git repo would grab
the developer's REAL `.git/sdlc.lock` (hanging `go test` under a live holder) and
could mutate the real tree (a stray test sequence corrupted `main` in the #148
session). Two guards: command-tree tests chdir into an isolated repo via
`hermeticRepo(t)` (so the lock resolves to the temp `.git`); and a package `TestMain`
snapshots the real repo (HEAD/branch/porcelain/`.git/sdlc.lock`) before+after the
run and FAILS a passing run that left durable damage (`snapshotDiff`, pure) — the
backstop that catches any test that still leaks.

The lock path is resolved from `git rev-parse --git-common-dir`, so linked
worktrees for one repo share the same lock. That is intentional: worktrees share
the issue namespace, object store, and remote refs that the motivating races
touched. The lock does not serialize another clone or machine, so remote
push/ref races still surface through the existing push/merge retry guidance.

`close` and `milestone-close` are narrower: they lock the compute phase, release
the lock while the external boundary review runs, then reacquire before
finalization and refuse to write if HEAD, the issue file, or any prepared project
file edit changed while the lock was released. `change-code`, `merge`, and `push` may still hold the lock while
synchronous judges run. Their wait/timeout messages call this out as a
long-running review/ship transaction; quick commands should wait or retry
instead of deleting a live lock. Recovery is conservative but not wedging:
`die()` drains the active lock cleanup registry before `os.Exit`, missing
`meta.json` during the tiny mkdir-before-write window is treated as holder
initialization and polled through, and a confirmed-dead same-host holder is
reclaimed by atomically renaming the stale lock directory before removal.
Cross-host or over-age uncertainty still produces operator-facing recovery
guidance rather than silent deletion; a live same-host pid overrides the age
ceiling.

## Artifact-reference resolution (`sdlc resolve` / `open`, #144)

ariadne artifacts cross-reference each other with **symbolic** refs
(`ariadne#11`, `#15 M4`, `pair#84`). The id is stable but the path is not — slugs
get renamed, and `sdlc merge`/`push` move an issue and its whole plan/review
family `issues|plans/ → history/` (#160). So refs stay symbolic and resolve at
**read time**; nothing is stored, nothing rots. `sdlc resolve <ref>` is that
resolver (the parley#160 editor UX shells to it).

- **Grammar single-sourced as the parser.** `parseRef` (`cmd/sdlc/resolve.go`) is
  the *only* implementation of the ref grammar. parley#160 and agents shell to
  `sdlc resolve` rather than re-encoding it in Lua — so the grammar can't diverge.
  `helptext/resolve.md` documents it for humans; a test (`TestResolveDocExamplesParse`)
  binds every documented example back to `parseRef` so the doc can't drift.
- **Locations derive from the model, not hardcoded.** The issue's `discovery:`
  block (`construct/vocabulary/issue.cue`, read via `pkg/vocab` `Discovery()`) now
  carries `home` + `glob` + `archive` (`workshop/history`) + `plans`
  (`workshop/plans`). `familyFiles` globs those three dirs for `NNNNNN-*.md` and
  `classifyFamily` sorts issue → plan → reviews. A 6-digit id resolves the whole
  family; `#id Mx` narrows to the `-mX-review.md` sidecar; `gh#id` labels a GitHub
  ref without resolving a local file (read-only + offline).
- **Cross-repo** by scanning the current repo's parent for a sibling: exact
  basename wins, else a unique case-insensitive prefix (`parley` → `parley.nvim`).
- **Read-only ⟹ lock-free by construction.** `resolve`/`open` are never tagged
  `markMutatingCommand`, so `wrapRepoLockCommands` skips them and they never touch
  `.git/sdlc.lock` (proven structurally + under a held lock). That's what makes it
  cheap enough (~process spawn) for parley to shell to on a keypress.

Pure core (`parseRef`, `classifyFamily`) is unit-tested with no IO; the IO seams
(`resolveRepoDir`, `familyFiles`) test against temp repos (ARCH-PURE). **Follow-up
(#163):** the existing `workshop/plans`/`workshop/history` hardcoders in
`push`/`merge`/`state` archive logic should migrate onto the same `Discovery()`
accessor — a DRY consolidation, separate from this resolver.

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
  repolock.go          root-level Cobra wrapper for mutating commands:
                       metadata annotation, command-context re-entrancy, and
                       Git-common-dir lock acquisition
  ghclient.go          ghCaller interface + realGH impl (shared)
  preflight.go         runPreflightJudges (push + merge pre-flight)
  close.go             ← scripts/close-issue.py
  actual.go            new (#68): computeActual → internal/activetime → measured --actual (adopted at close when omitted, #178)
  activetime.go        new (#110): `sdlc active-time` CLI (runActiveTime + table renderer)
  internal/activetime/ new (#110): native v3 engine ported from active-time-v3.py
                       (event/commit/segment loaders + Compute; pure core + thin IO seam)
  internal/transcripts/ new (#134): transcript-source harness registry — a Harness
                       per agent CLI (claude.go / codex.go), pure Select aggregator
                       feeding actual.go; adding a harness = one entry
  state.go             new (read-only inspection + drift detection; see "Drift checks")
  resolve.go           new (#144): `sdlc resolve`/`open` — pure parseRef + classifyFamily
                       behind the resolveRepoDir/familyFiles IO seams; read-only, lock-free
  judge.go             ← scripts/pre-merge-checks.sh
  fetch.go             thin hidden alias → runIssueNew --from-github (#56 M2)
  issue.go             new (#56): `sdlc issue` group — new / set-status / list / show / validate (#124)
  validategate.go      new (#124): the deterministic instance-conformance gate run
                       by push+merge before the irreversible action, independent of
                       the LLM judges (frontmatter on every changed issue; sections
                       added-only); shells `vocabulary validate-instance`. `--no-validate`
  start.go             migration stub (REMOVED in #39 — errors with
                       "use claim + change-code")
  claim.go             ← scripts/issue-sync.sh (renamed from lock.go #39)
  changecode.go        new (#39): planning → implementation gate
  branchcreate.go      new (#39): branch-creation helpers shared by
                       changecode.go (worktree + in-place paths) + the
                       name-resolution previously in start.go. #156: both
                       paths are IDEMPOTENT for milestone re-runs — pure
                       deciders `decideInPlaceBranch` (onTarget/switch/create)
                       + `decideWorktreeBranch` (reuse/addExisting/addNew)
                       over `currentBranch`/`branchExists`/`worktreeForBranch`
                       probes, so re-running change-code on an existing branch
                       switches/reuses instead of dying `checkout -b: already
                       exists`. `worktreeForBranch` filters the single-source
                       `parseWorktrees` (extracted from state.go's listWorktrees;
                       findMainWorktree refolded onto it — one porcelain grammar).
  setstatus.go         new
  push.go              ← Makefile push:
  pr.go                ← Makefile pull-request:
  merge.go             ← Makefile merge:
  milestoneclose.go    composition over close + judge milestone-review
  estimatesource.go    new (#134): `sdlc estimate-source` pull — names the shared
                       method + repo-local calibration doc (estimateSourceStatus
                       seam over estimate.SourceGuidance); start-plan/change-code push it
  helptext/            //go:embed *.md — one .md per verb + root
  internal/
    repolock/          local repo transaction lock: pure holder metadata /
                       stale-observation decisions + thin mkdir/meta.json
                       acquire/release IO shell
    gitx/              git invocation seam (`run` shim, Capture, DiffBase,
                       MainRef, CommitWindow, WorkingTransitionISO (#113 claim
                       anchor), DiscoverWindowIssues, RunGit,
                       IsShippedWorkSubject/ShippedWorkOnMain — #76 ship probe)
    issue/             frontmatter parse/edit + plan-section regexes +
                       scaffold.go (NextID/Slugify/Render — #56)
    judge/             Category enum, prompt builder, classify, dispatch
    project/           brain project-file mutation helpers
```

## Drift checks (`sdlc state`)

`state` gates nothing — it reports. `detectDrift` (state.go) surfaces warn-only
inconsistencies so an agent recovering after compaction sees what's off:
missing-frontmatter, `done`/`wontfix`/`punt` still in `workshop/issues/` (should
be archived), and `working` with **zero** plan items ticked (the no-progress
end). #76 added the inverse — the **close-off candidate**: an `open`/`working`
issue whose plan is all (or all-but-one) ticked *and* whose work has shipped to
main, i.e. done work that never got formally closed. The "shipped" signal is
gh-free by design: a subject-anchored `git log <MainRef>` scan
(`gitx.ShippedWorkOnMain` → pure `IsShippedWorkSubject`) distinguishes a real
`#N Mx:` work commit from bookkeeping (`file issue`/`ticket`/`claim`/`close`),
so a merged PR's work commits *are* the signal — no network, degrades to
"not-shipped" when there's no main ref. Warn-only on purpose: it points at `sdlc
close --issue N` for a human glance, never auto-closes (closing carries
actual/verified judgment a heuristic can't supply). The probe is injected into
`detectDrift` as a `shipProbe` func so the drift logic stays testable without
git.

## Friction audit (`process-manual --friction-report`, #172)

The per-gate `--no-<gate>` bypass design (below) was built to make bypasses
*explicit and measurable*; `process-manual --friction-report` closes that loop. It
measures, across the **whole Claude corpus** (all repos), where the spine creates
friction — per-gate bypass rates, and (M2+) refusal→retry loops + firing-order
anomalies — as a clean, command-anchored measure that replaces a contaminated grep.

The core problem is **discrimination, not capture**: this repo *develops* sdlc, so
`close.go`'s source and cat-n log reads spray every `--no-<gate>` string into tool
output. So the instrument (1) **anchors to `Bash(sdlc <verb>)` invocations** (drops
the source/edit/log-read noise — the dominant contamination), joined to their
`tool_use_id`-linked tool_result content-block; (2) classifies each output line
against a **per-gate signature catalog** (`internal/processmanual/gatesig.go` —
12 gates / 16 sigs / 3 ACK grammars: G1 close·mclose `--no-X (or --force): …`,
cinfo no-judge, G2 change-code `X gate bypassed (--force: …)` (silent alone), G3
merge/push `--no-X: …`), requiring the runtime `\x1b[0m` reset for a bypass ACK and
grammar+digit-anchored patterns for refusals (so the `printSemanticWarmup` success
line and source restatements are rejected); and (3) states **observability limits**
honestly (change-code silent-alone bypasses countable only via `--force`; merge/push
refusals that don't name the flag). A **cross-command drift guard** (`gates_test.go`)
asserts the catalog matches each command's registered `--no-*` flags, so a new gate
can't be added without the audit noticing (ARCH-DRY). Codex coverage (M3) derives a
net-new Go parser from `atlas/workflow/introspect.md` → "Codex transcript format" —
the two can't share code (Python vs Go), only the spec.

M1 headline over the real corpus: `--no-judge` dominant, `--no-verified` = 0 (the
gate design works), and bypasses concentrate in **peer repos, not ariadne** — the
substrate repo follows its own gates; lighter repos route around them. The raw grep
over-counted ~4× (unlinked echoes: process-manual outputs, transcript reads).

The M2 detectors ride the same invocation stream (`buildFrictionReport` is the
composition seam; gate events are deduped per invocation — one no-validate refusal
prints two matching lines):

- **Refusal→retry** (`detectRefusalRetries`) pairs each gate refusal with the next
  same-verb+same-issue invocation in the same transcript; `resolved` distinguishes
  satisfying the gate from routing around it (`via bypass`). merge/push refusals
  never name their flag → paired by verb+context, labeled `flag-omitted`.
- **Firing-order** (`detectFiringOrder`) walks each (repo, issue)'s invocations
  across transcripts against the AGENTS.md §2 ladder (claim ≺ start-plan ≺
  change-code ≺ milestone-close ≺ close ≺ merge/push). Only an observed **order
  inversion** flags (change-code after a clean close/merge); legal loops —
  mclose→change-code, start-plan re-runs, close→change-code after a REWORK verdict
  (recovered via `judge.ParseVerdict`) — stay silent, and absent early stages are
  partial observation, not anomalies (precision over recall). merge/push carry no
  `--issue` → attributed from segment context or counted unattributed. The
  `skill-late` arm flags a plan/TDD Skill load after a non-`.md` file edit in the
  same segment+issue (Edit/Write/MultiEdit → `KindFileEdit`, excluded from
  `--session` reports).

M2 headline: refusals mostly resolve by **satisfying** the gate (no-estimate-recon
19/19 satisfied; via-bypass is rare — no-verdict 2, no-atlas 1), i.e. refusals do
their job; firing-order found change-code-after-close 16 / skill-late 2.

**M3 — codex coverage (both agents).** The walk also enumerates
`~/.codex/sessions/**/rollout-*.jsonl` via `codex.go`, whose format knowledge
derives entirely from `atlas/workflow/introspect.md` → "Codex transcript format"
(the spec is the DRY point with Python introspect; a cross-language golden at
`internal/processmanual/testdata/codex-golden/` pins the shared keep/skip +
classification decisions). Fork-replay rollouts (`forked_from_id` on the FIRST
`session_meta`) are skipped — the real corpus skips exactly the spec's 40 —
while sub-agent threads are processed; repo labels come from `session_meta.cwd`
(worktree-normalized). sdlc's ANSI survives codex's exec_command wrapper, so the
SAME `classifyOutputLine` serves both agents. `SdlcInvocation.Failed` (Claude's
`is_error` flag / codex's non-zero `Process exited with code N` — deliberately
NOT the spec's hint-gated taste-friction `is_error`) keeps failed invocations
from raising the firing-order ladder. Report adds the per-agent bypass split +
forks-skipped counter; skill-late is Claude-only (codex has no Skill tool).

M3 headline: codex 43 vs claude 37 bypasses; codex's signature move is the
**re-close** (no-reclose-guard: 25 bypasses, and its 3 refusals all resolved
**via bypass** — the one gate agents route around after refusal); no-actual
refusals jump to 38 corpus-wide (35 satisfied). Firing-order anomalies unchanged
by codex (16 + 2).

## Anti-collusion + form-vs-essence

Checkpoint guards defend against **omission** (claiming done without
doing) via deterministic checks (`close` refuses without `--actual` +
`--verified`). The judge subcommand defends against **theater** (form
without substance) via fresh-context LLM review — every Dispatch call
spawns a new subprocess; the agent has no doer-session state.

**Judge agent defaults (#129).** Every judge dispatch (`sdlc judge`,
`change-code` plan/estimate quality, close and milestone boundary reviews, plus
push/merge preflight) resolves through `internal/judge.ResolveAgentCLI`. The
precedence is: explicit `--agent`, then `AGENT_CMD` (operator/script override,
including invalid values so dispatch validation still fails loudly), then
`PAIR_AGENT`, then conservative process/session signals such as `CODEX_CI` /
`CODEX_THREAD_ID`, then `claude`. Command flags stay empty by default so
environment defaults live in one place; call sites pass an explicit-source bit
from Cobra's `Flags().Changed("agent")`. Close dry-run uses the same boundary
review prompt/options builder as real dispatch and prints the would-be command
line, so `PAIR_AGENT=codex` is inspectable as `codex exec` before any subprocess
runs.

**Dispatch progress heartbeat (#140).** A boundary review can run silently for
minutes; `internal/judge.Dispatch` now emits a heartbeat to `opts.Stderr` every
`heartbeatInterval` (30s) while the agent subprocess runs — elapsed + agent +
child PID (`… still working — 1m0s elapsed via claude (pid N; inspect: ps -p N)`).
It is harness-agnostic by construction: all three fields come from `sdlc` wrapping
the child (`Run` now Start→`onStart(pid)`→Wait into one combined buffer, exposing
the PID), not from child output, so it reads identically for claude/codex/gemini.
Gated on `opts.Stderr != nil`, so the fast path (unit tests, quick dispatches)
stays synchronous and silent; the captured output + exit-code policy
(`classifyRunResult`) are unchanged, so `Classify`/`ParseVerdict`/the sidecar are
untouched. Deliberately no byte-count/log-tail signal — `claude -p` buffers to the
end (a live counter would read "0 B" and look stalled) and no agent exposes a
reliably-locatable per-invocation log; the PID is the automated form of the
operator's manual `ps` inspection.

**The estimate shell (#117)** applies this same form/essence split to
`estimate_hours`, plus a feedback arm — the one forecast in the system with a
deterministic ground-truth measurement (`sdlc actual`). *Form:* the change-code
**estimate-reconciliation** gate parses the issue's `## Estimate` fenced block
(`internal/estimate`, pure) and refuses unless `estimate_hours` reconciles with
an itemized v2-lineage primitive derivation (`Σdesign×(1+buffer)+Σimpl×familiarity`). No
unitemized estimate — a fabricated number can't pass. *Essence:* the
**estimate-quality** judge checks the derivation was applied, not back-fit.
*Feedback:* on a full-issue close `sdlc close` appends every numeric estimate↔actual pair
to `brain/.../velocity/calibration-ledger.tsv` (`$WF_CALIB_LEDGER` override) and
flags >2× same-direction drift over the last N unique **window-trusted** rows for
the latest recognized model revision — pre-#116 rows are `window-trusted=no` and
excluded (a truncated actual isn't a clean point; `actual_hours: N/A` rows are
excluded; duplicate append rows for the same issue/model count once;
skip-with-warning when no brain/ ledger dir exists downstream). This
closes the loop the hand-kept validation log never did. Grammar + closed
vocabulary: `helptext/estimate.md` (canonical slugs in
`internal/estimate/vocab.go`, a bidirectional drift test guards the mirror).

**Estimator-source discovery (#134).** The shared method is single-sourced in
`sdlc` (the grammar + `Models()`/`Primitives()`), but the *repo-local
calibration* — the actual per-primitive hour ranges, which drift as closes accrue
(#127) — lives in a brain artifact. So an agent could satisfy the block grammar
while picking hours from memory. `sdlc estimate-source` (`estimatesource.go`, the
pull; mirrors `arch-principles`) names BOTH in one output: the shared-method
pointer + the resolved calibration doc for `estimate.CurrentModel()` by default
(`estimate.SourcePath` →
`$WF_ESTIMATOR_SRC` override, else `estimate.VelocityPath(brainDir,
<model>.md)` — the same `data/life/42shots/velocity/` builder `close.go`'s ledger
path now derives from, ARCH-DRY), tagged `ok | stale | MISSING` (stale = sibling
ledger newer than the doc). The pure renderers (`SourceGuidance` full / `SourceLine`
one-line) live in `internal/estimate/source.go`; `estimateSourceStatus` is the
thin stat/mtime seam. Fail-loud is asymmetric on purpose: the PULL exits non-zero
when the source is missing, while the PUSHes (`start-plan` emits `SourceLine` after
the estimate nudge; `change-code`'s missing-block error points at the command)
warn-and-continue so a brain-less downstream repo never breaks the gates.

**Per-gate bypass (#67).** `close` has 8 gates (actual, verified, atlas,
milestone-verdict, plan-unchecked, project, re-close, and the #69 boundary
review), each with its own `--no-<gate>` flag (`--no-actual`, `--no-verified`,
`--no-atlas`, `--no-verdict`, `--no-plan-check`, `--no-project`,
`--no-reclose-guard`, `--no-judge`);
`closeFlags.skip(gate)` is the single arbiter (`Force || the field`). A
per-gate flag is an *acknowledgment* that one guard doesn't apply (e.g. a
pure bugfix → `--no-atlas`); it logs an audit `[!]` line and only fires
when the gate would actually have refused. `--force` waives all at once.
The atlas gate additionally auto-satisfies (info line, no flag needed) when
the window contains no code surface — `hasCodePath`, the single docs
classifier: `*.md` / `workshop/` / `atlas/` / `docs/` are documentation,
everything else is surface (#177; docs-only closes had an incoherent demand).

The milestone-verdict gate demands per-milestone `Review-Verdict:` evidence,
EXCEPT for **trailing** unclosed milestones (#175): Mx rows after the last
verdict-carrying one — or all rows, when none carries a verdict (the
single-pass over-split shape) — are accepted with a loud info line, because
the issue-close boundary review's window (branch-point→HEAD) covers their
work and the close *is* their first boundary. `partitionMissingVerdicts`
(pure, `close.go`) does the split; **midstream** misses (a later milestone
closed WITH review) still refuse — that boundary was genuinely crossed
unreviewed — and `--no-judge` turns trailing acceptance back into a refusal,
since it skips the very review the acceptance leans on. The plan-quality
judge flags over-split Mx plans at design time (the forward fix).

**Off-workflow guards (#176, `repoguard.go`).** The 7 lifecycle verbs (=
`processmanual.WorkflowVerbs()`, enforced by a drift test) refuse up front in
a brain repo (`.brain/config.md` — the #172 audit found brain concentrating
bypasses because its own merged constitution invited sdlc) and in repos
without `workshop/issues/`; reads (estimate-source, actual, state,
process-manual, issue …) are unguarded by construction since sdlc legitimately
reads brain. `WF_SPINE_GUARD=off` is the single emergency hatch and cwarn-ACKs — greppable
in transcripts (GateCatalog wiring for this env-gate family is deliberate
follow-up work; the instrument today derives only from catalog rows). `guardIssueNotDone` closes the front
door on terminal issues: start-plan/change-code refuse on `status: done`
(re-close was already guarded at the back).
`milestone-close` forwards the same flags into its delegated `computeClose`
(the #139 compute→review→finalize; `runClose` is now test-only, #146).
The convention generalizes `merge`'s pre-existing `--no-judge`.

**Measured actuals (#68, #110).** `--actual` is computed, not hand-typed. `sdlc
actual --issue N` (`actual.go`'s `computeActual`, shared with close's
missing-`--actual` explainer) runs the native **`internal/activetime`** engine
(`activetime.Compute`, in-process — no python3) over the issue's `CommitWindow` +
`DiscoverWindowIssues` peers, feeding it **brain + the issue's repo** transcript
sources. Source selection is a **harness abstraction** (`internal/transcripts`,
#134), not a Claude-only path convention: each agent CLI implements a `Harness`
(`Name()` + `Sources(cwds) → Sources{Dirs,Files}`); `DefaultHarnesses()` is the
registry; a pure `Select` merges their contributions for the brain+repo cwds.
Two harnesses ship — **Claude** (one `~/.claude/projects/<cwd-encoded>` dir per
cwd → `Dirs`, engine globs them) and **Codex** (date-sharded
`~/.codex/sessions/YYYY/MM/DD/*.jsonl` filtered by `session_meta.cwd` → `Files`,
since Codex stores cwd inside the file, not the path). A third agent CLI is one
new `<harness>.go` + one registry entry — `actual.go` and the engine never change
(ARCH-PURPOSE). Pure encoders/parsers (`cwdToClaudeDir`, `codexCWDFromBytes`) are
unit-tested directly; each `Sources` method is the thin IO seam, tolerant of
malformed / empty / no-`session_meta` Codex files. This is the validated heuristic (events come only from transcripts;
the wrong/missing sources were why actuals read 0 and got faked). The engine
returns a structured `Result.Status`: `TelemetryGap`
(commits-but-0-events → labeled judgment), `EmptyWindow` (nothing to measure), or
`Measured` (`PerIssue[N]` → hours). Attribution is global-boundary based (#92):
source-scoped activity runs are claimed by nearby issue-referenced commits;
overlaps collapse within one transcript source but not across parallel sources;
all commit-subject issue refs become claimants, while no-ref commits are neutral
time boundaries. Suspicious attribution is surfaced as `Result.Warnings` and
rendered by `actual` / `active-time`. Dir-selection is deliberately narrow (NOT all
folders/sessions) — an unrelated concurrently-edited repo inflates the count.
`WindowCapDays` is 61 (was 31) so month-long issues keep their window. The
window-**start** is the *earlier* of `CommitWindow`'s parent-of-first-`#N`-commit
and the **engagement anchor** (`resolveWindowStart`), anchoring at the cheap early
`claim` so DESIGN attention (brainstorm / spec / plan / reviews) before the first
code commit is in-window instead of cut off; gap-truncation keeps a dormant
claim→work gap from inflating the actual. The anchor is resolved in robustness
order (#116): the explicit `started:` frontmatter stamp (written once at the
open→working flip in `applyStatus`, local-offset RFC3339 to match `%aI`) →
`gitx.WorkingTransitionISO` (the #113 git-log heuristic, now the legacy fallback)
→ commit-parent. The explicit stamp survives rebases/moves where the heuristic's
"best-effort" history scan could silently miss and drop design time.

`sdlc active-time` (#110) is the standalone CLI over the same engine — the
manual-inspection sibling that prints the full attribution table and warnings. It preserves
the #68 loud-fail exit-code contract: **2** = misinvoke (no `--dir`/`--issue`/
`--git-repo`), **3** = telemetry gap (commits but 0 events), **0** =
measured-or-empty. (Before #110 the engine was `active-time-v3.py`, a python3
subprocess whose human-formatted stdout `actual.go` regex-scraped; #110 ported it
into the binary and deleted the script.)

**Passed-`--actual` backstop (#87).** When `--actual` *is* given, close still
runs the engine and compares (`actualDeviation`, the pure comparator in
`close.go`): ratio ≥3× → warn, ≥10× → **refuse** (bypass with `--force`), with a
0.5h absolute floor so small gaps don't trip. Skips silently when the engine
can't measure. Closes the hole where a hand-typed value (the failure #86's docs
prime against) was trusted blindly — the doc fix removes the priming, this
removes the blind trust. `milestone-close` inherits it (computes via `computeClose`).

`push` and `merge` auto-dispatch `judge plan|specs|lessons` as pre-
flight so the checks run consistently rather than as a remembered
manual step. `milestone-close` auto-dispatches `judge milestone-review`
as a post-action.

**Judges are read-only (#62).** A judge is a reviewer, not a doer — all
categories run with a read-only tool allowlist (`Read,Grep,Glob,Bash`); they
report findings and the main agent (full context) applies fixes. The `specs`
judge used to auto-edit stale docs (`Edit,Write`), which let a *passing* gate
leave the tree dirty and strand the subsequent merge. `merge` now also (a)
re-asserts no **tracked** dirt immediately before the irreversible `gh pr merge`
(refuse, don't strand), and (b) resumes an interrupted merge — a re-run detects
an already-merged PR and finishes the local cleanup instead of erroring. The
resume path is guarded (#148): before cleanup it fetches `origin/main` and counts
`origin/main..origin/<branch>` (`countUnmerged`, fakeable seam); a nonzero count
means a **reused branch name** (its old work shipped via that PR, new work piled on)
→ `decideMergeAction` returns `actionResumeBlocked` and merge refuses *before* any
switch/delete/archive, rather than silently stranding the new commits.
The clean-tree guards (step 2 and the 9b re-assert) refuse only on tracked
**code** changes via one pure `assessDirty(...).Refuse()` decision (#78);
`assessDirty` buckets each porcelain line into Blocking / Untracked / Tracker.
**Untracked** files survive `git switch main`, so they're surfaced as a warning,
not a blocker — unrelated local WIP no longer forces a stash-around-the-merge
detour. **Tracker** files (`workshop/issues|history/NNNNNN-*.md`) are likewise
never blocking, tracked-modified *or* untracked (#82 M2): they're append-only
shared state synced to main out-of-band (#82 M1), not code contention, so a dirty
issue file never gates a merge. (Path matching reuses push.go's
`isIssuePath`/`isHistoryPath`; the path is pulled by field-split, not column
slice, since `worktreeDirty` whole-trims and strips the first line's leading
status space.)
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
BLOCK` block. **Boundary review (#147): the handoff is now block-first.** The
reviewer emits a fenced ```` ```verdict ```` block validated against `verdict.cue`
(`ParseVerdictBlock`, the authoritative structured handoff — see [Vocabulary](vocabulary.md)
+ the `agent-binary-handoff-schema` target); the prose `VERDICT:` line is the
logged fallback. So a prose-narrated verdict no longer degrades to `unknown`. Lessons is the exception — a fixed `REMINDER:` line, no agent. This
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
The **estimate-quality** judge (#117) is a change-code-time-only sibling of
plan-quality — deliberately NOT in `AllCategories()` so it never enters push/merge
bulk dispatch; it checks the `## Estimate` derivation was *applied*, not back-fit.
Cite the marker (`ARCH-DRY`) in plans/Logs/findings. Adding an `ARCH-*` entry
flows into every consumer with no other edit. **`sdlc start-plan`** (#75 M2)
delivers the `at-plan` lens to the main thread at design time — the forward
counterpart to `change-code`'s plan-quality review (`claim → start-plan →
change-code`). #71 adds `ARCH-SHIM`. #126 landed `ARCH-PURPOSE` (serve the issue's
actual purpose; single-source ⇒ every consumer *derives* — the registry's 3rd
marker, disambiguated from Simplicity-First/YAGNI). **#128** added
**`sdlc arch-principles`** — a standalone command that prints the registry (the
same `ArchitectureBlock` primitive start-plan calls), the *pull* path for non-gate
work (autonomous fixes, quick edits, Q&A) that never hits start-plan. AGENTS.md's
narrative collapsed from a per-marker restatement to a *route* to that command
(definitions single-sourced; the drift test — `TestArchitecture_NarrativeRoutesToArchPrinciples`
— now guards the route + marker awareness, not an enumeration). The start-plan
*push* stays: a gate-time push beats a model-dependent pull (the same asymmetry
that keeps the judges' inline embed). **#82 M3 / #83** also have `start-plan` print non-blocking
**dependency-path contention** — one line per repo this one reads live. The
symlink model means a repo reads ALL its transitive upstreams' working trees, so
the "moving ground" is the whole dependency chain, not a single base.
`substrateChain(root)` walks `construct/deps` transitively (resolving each
`substrate <path>` against its *declaring* root; `parseSubstrateTargets` is the
Go port of `lib-deps.sh deps_substrate_targets`, kept in lockstep); per upstream,
`gatherBaseContention(root, …)` reads branch + dirty *code* count (tracker files
excluded, reusing M2's `assessDirty.Blocking`) + other `status: working` issues,
and the pure `baseContentionSummary` renders the line (clean → green "clear to
plan"). The **root** (ariadne — no upstream) reports its own concurrent work; a
**derivative** reports its upstream(s) (`base (ariadne): …` from a nous session),
which is the Spec's primary case. (#83 replaced #82 M3's broken `isBaseRepo`
heuristic — `construct/` is a real dir in *every* repo, so it fired everywhere.)
It never refuses. **#72** adds a third payload between the architecture block and
the contention line: the pure `planPointer(issue)` durable-plan reminder — author
the plan via the `superpowers-writing-plans` skill into
`workshop/plans/NNNNNN-slug-plan.md`, never the harness builtin's ephemeral
`~/.claude/plans/`. So start-plan emits, in order, *what to design against*
(architecture) → *how/where to capture it* (the skill pointer) → *the moving
ground* (contention).

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

**Repo orientation (#137).** The review prompt orients the fresh reviewer to the
**actual repo under review**, derived from the live git context — not a hardcoded
`ariadne`. `boundaryOrientation` (`cmd/sdlc/orientation.go`) computes the repo name
(git-root basename), root path, the `<repo>#N` issue ref (so a `pair` review reads
`pair#72`, never `ariadne#72`), issue file, boundary kind, and a base-vs-downstream
note (base detected via `construct/base.manifest`); these are passed as plain
strings into the pure `internal/judge` layer (`PromptInput` → `code-review.md`
header), keeping git IO at the cmd boundary (ARCH-PURE). Computed once in the
shared `boundaryReviewDispatchOptions` (ARCH-DRY); the same derivation feeds the
sidecar H1.

**Two-phase close, finalize-after-verdict (#139).** `sdlc close` and
`sdlc milestone-close` no longer write before the review. `runClose` splits into a
read-only `computeClose` (all gates + composes the new issue/project text → a
`closeResult`, writing nothing) and `applyClose` (the writes + calibration ledger).
Full-issue close and milestone-close both **compute → review → finalize**: the
boundary review runs against the *un-mutated* working tree (the reviewer reads the
honest `status: working` issue), and `applyClose` fires only on a **finalizing**
verdict via the shared finalization helper. The command path releases
`.git/sdlc.lock` while the external review subprocess runs, then reacquires and
checks that HEAD, the issue file, and any prepared project-file edit still match
the reviewed snapshot before writing. `closeVerdictOutcome` derives from
`vocab.Verdict()` (#147): finalizing (SHIP/FIX-THEN-SHIP) → finalize; blocking
(REWORK) → **not finalized**, issue left `working`, non-zero exit, "fix + re-run"
(no `--no-reclose-guard` needed on the rerun since it never went `done`);
unknown / dispatch-error → **halt**: don't finalize an ambiguous gate — stop and
consult a human. `--no-judge` finalizes (explicit operator skip, handled before
dispatch). The success messages ("flipped → done") print only from `applyClose`, so
a REWORK never claims a write that didn't happen.

**Subprocess PATH (#138).** The agent subprocess `sdlc` spawns for a review (and
for `sdlc judge`) gets the owner `bin/` prepended to its `PATH`, so a fresh
reviewer can resolve `sdlc` (and sibling tools) even when the spawning shell's
startup files never put `ariadne/bin` on `PATH`. The dir is `dir(os.Executable())`
(single-sourced via `ownerBinDir`, in `internal/judge/dispatch.go`), injected at
the one launch seam (`Run`) via the pure `binAugmentedEnv` — so it works from
downstream repos (the binary is `…/ariadne/bin/sdlc` regardless of cwd) with no
dependence on the user's `~/.zshenv`/`~/.bash_profile`. Launch-failure errors name
the attempted agent + that bin dir.

**Review sidecar (#136).** The boundary review is no longer a transient terminal
artifact: every actually-dispatched review writes its full transcript to a durable
sidecar under `workshop/plans/` — `NNNNNN-slug-close-review.md` for a whole-issue
close, `NNNNNN-slug-m<x>-review.md` for milestone `Mx`. The write lives in the
single shared `dispatchBoundaryReview` (`reviewsidecar.go`: pure `sidecarMeta` +
`renderReviewEntry` + `sidecarPath` behind a thin atomic-write seam — ARCH-PURE),
so both close paths inherit it for free (ARCH-DRY). Each file carries a metadata
header (issue id/title, repo, issue file, boundary kind, milestone, base..head
window, command, reviewer, timestamp, verdict) plus the body. A re-run of the same
boundary **appends** a timestamped `## Re-review` section rather than overwriting
(the §1 revision convention). The terminal still prints the full body + the
`Review-Verdict:` trailer; the sidecar adds a durable surface an agent can reopen
after scrollback loss or compaction (the path is echoed as `review sidecar: …`).
`--no-judge`/`--dry-run`/not-run boundaries write nothing — no body to persist.

**Window base — prior review boundary (#58).** `boundaryWindowBase`
(`milestoneclose.go`) is the single source for *both* the atlas-coverage gate
(`computeClose`) and the boundary review's window, so they provably cover the same
commits (ARCH-DRY). A milestone window bases on the **previous review boundary**
— the most recent prior commit touching the issue file that carries a
`Review-Verdict:` trailer (the prior milestone close), found by
`previousReviewBoundary` — not on the first `#N Mx` commit. This closes a gap
where an inter-milestone `#N`-but-not-`Mx` commit (a `side-quest:`, a fix) landed
between M(x-1)'s close and Mx's first commit would slip *both* windows and escape
review. The first milestone (no prior boundary) falls back to the branch start
(parent of the first `#N` commit); if a prior close's trailer was never pasted,
the lookup finds nothing and falls back the same way — over-covering rather than
under-covering, the safe direction.

The **whole-issue** close (the end-of-issue integration review) bases on the
**branch point** — `gitx.MergeBaseWithMain()` = `merge-base(main, HEAD)` — so it
windows exactly this branch's commits, not unrelated history merged onto main
before the issue's first commit (#77; an issue filed early but implemented late
otherwise over-captured everything since it was *filed*). On `main` (the direct
`sdlc push` flow, no divergence) merge-base == HEAD, so it falls back to the
issue's branch start. `MergeBaseWithMain` is deliberately separate from
`DiffBase` (the `sdlc judge` window): same merge-base core, but it returns `""`
on no-divergence so the caller can pick the issue-specific fallback.

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
verifies the history copy still has a terminal status, stages **only those exact
moved paths** (`git add -- <src> <dst>` via the shared `archiveAddArgs` helper —
never a directory-wide `git add issues/ history/`, which would also sweep
unrelated untracked WIP onto main, #80), commits "archive completed issues to
history", pushes, and exits without rerunning judges against an archive-only
retry. The same precise-staging helper backs the non-recovery archive commit in
both `push` and `merge`. The archive sweep also moves the issue's
`workshop/plans/NNNNNN-*` artifacts (durable plan + review sidecars) into history
alongside it (`archivePlanArtifacts`, #143); recovery reconstructs those plan
moves too, but — since plan artifacts carry no terminal frontmatter — gates them
on their id-prefixed plans-dir source rather than the terminal-status check used
for issue files. A review sidecar is frequently **untracked** at ship time
(`sdlc close` writes it after the implementer's last commit, and a FIX-THEN-SHIP
fixup stages explicit paths, not the sidecar), so `archivePlanArtifacts` probes
each source with `git ls-files` (the injected `gitSrcUntracked` seam) and marks
the move `SourceUntracked`; `archiveAddArgs` then stages **only the history dest**
for an untracked source — adding its vanished pre-rename path would abort the
whole archive commit with `pathspec did not match` (#154). Tracked sources (issue
files, committed durable plans, recovery moves) still stage source-deletion +
history-addition exactly as before. Any unrelated dirty file keeps the refusal
path and tells the operator to clear that unrelated work before rerunning `sdlc
push --yes`.

## Makefile wrappers (transition state)

Each Make target delegates to `bin/sdlc` when built, falling back to
the original shell logic when absent:

  `make close-issue` → `sdlc close`
  `make fetch <N>`   → `sdlc fetch --github-issue N` (deprecated alias →
                       `sdlc issue new --from-github N`, #56 M2)
  `make worktree`    → `sdlc change-code --worktree=yes --no-judge --no-structural
                       --no-estimate` (post-#39, #113; preserves the make target's
                       pre-existing quick-and-dirty gate-free semantics)
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
