---
collaboration-mode: discussion
---

# tl-lead teams Implementation Plan

> **For agentic workers:** Consult AGENTS.md Section 3 (Subagent Strategy) to determine the appropriate execution approach: use superpowers-subagent-driven-development (if subagents are suitable per AGENTS.md) or superpowers-executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make a team a first-class, addressable unit for tl-lead: a `teams/<name>/` folder with `.hub` and `team.md`, a `--team` that every command resolves, a scaffold verb, path-limited team commits, a view of new unclaimed issues in the repos a team owns, and help that states the team, routing and steward rules.

**Architecture:** A pure `team` model (name grammar, `team.md` frontmatter, `.hub` claim lines, scaffold contents, the commit plan, the issue view's filter) sits under a thin shell. The shell resolves the team from the working directory, writes the scaffold files, runs git for the commit, and runs sdlc through the existing `SDLC` seam (extended by one read). Couch is not touched: it reads only `.hub` (pair#694), and this issue only writes claim lines that pair's grammar accepts.

**Tech Stack:** Go (cobra), `gopkg.in/yaml.v3` (already a dependency), git and sdlc as subprocesses.

**Source design:** ops `teams/devinfra/projects/introduce-team/design.md`, sections 1, 2 and 4. Section 2 is the contract shared with pair#694 (pair:1).

---

## Scope

In:
1. **Team addressing:** `team:<name>`. The name is the folder name. tl-lead accepts `--team devinfra`, `--team team:devinfra` or a path to the folder.
2. **`--team` on every command,** as a persistent root flag. The default is the team folder you are in, found by walking up to a directory that holds `.hub` and whose parent is named `teams`. The status verbs' `FILE` becomes optional, defaulting to the team's single `tl-status-*.md`.
3. **`tl-lead team new NAME --owns R[,R…] [--claim C]… [--steward]`:** scaffolds `teams/NAME/` with `.hub`, `team.md` and `projects/`. It can also adopt an existing folder (devinfra today), writing only the two missing files.
4. **`tl-lead team list`:** every team in the ops checkout, with its address, owned repos and steward flag. It flags two teams owning one repo, and zero or several stewards. `team list REPO` prints the owning team's address, for a heads-up.
5. **`tl-lead team commit -m MSG`:** a path-limited commit of the team folder plus any `workshop/projects/*.md` link that points into it, then `git pull --rebase --autostash`, then `git push`. It is resumable.
6. **`tl-lead team issues [--limit N]`:** for each repo in `team.md`'s `owns:`, the open issues nobody has claimed, newest first.
7. **Help:**
   - tl-lead `help.md`: a TEAMS section; ROLES now says one TL per team; the DISPATCH filing line; and the session file's `couch-owner`, which becomes the team.
   - tl-worker `help.md`: FILING ISSUES now says to file directly where the problem lives and tell your hub in one line; the hub never relays.
   - `atlas/workflow/tl-skills.md`.

Out (whose they are):
- **Couch startup, lifecycle, switcher, `team:` routing and `.hub` `key=value` directives** belong to pair#694. tl-lead never writes a directive line.
- **The live proof** (Done-when 2: `teams/echo` running with its own session file, plus a live cross-team exchange) needs pair#694 as well. The hub runs it after both land (design §5.3). This issue's part is that `tl-lead team new echo` produces the folder that proof starts from.
- **The TL skill stub** `xx-tl-lead` stays as it is: it points at `tl-lead --help`, which is where the content goes.

## Decisions taken here (flag any you'd change)

- **D1. `team.md` carries no `hub:` field.** The folder name is the team's name, and the address is `team:<folder>`. A `hub:` value could only drift from it, which is the same reason `.hub` has no name line (ARCH-DRY). If the design meant something else by "hub name" (the agent or model?), say so and it becomes a plain informational field.
- **D2. `.hub` defaults to one repo-claim line per `--owns` repo.** `--claim echo:2` adds slot claims. tl-lead writes nothing else: no comments, no directives.
- **D3. Session FILE becomes optional.** When it's omitted, tl-lead uses the team's one `tl-status-*.md`. If the team has none or several, it refuses and names them. An explicit FILE keeps today's meaning.
- **D4. The team commit is a verb, not prose**, so "path-limited, pull before push" is enforced rather than remembered (ARCH-PURPOSE).
- **D5. Steward checks warn and never refuse.** Couch's command-line order decides the steward for Couch. `team.md`'s `steward:` is what TLs read, and `team list` warns when the teams disagree (no steward, or more than one).

## Core concepts

### Pure entities

| Name | Lives in | Status |
|------|----------|--------|
| `TeamName` / `parseTeamRef` | `cmd/tl-lead/team.go` | new |
| `TeamDoc` / `parseTeamDoc` | `cmd/tl-lead/team.go` | new |
| `hubClaimLine` (validate one claim) | `cmd/tl-lead/team.go` | new |
| `scaffold` (files for `team new`) | `cmd/tl-lead/team.go` | new |
| `teamTable` (list + ownership/steward checks) | `cmd/tl-lead/team.go` | new |
| `commitPaths` (pathspec for a team commit) | `cmd/tl-lead/team_commit.go` | new |
| `unclaimedIssues` (filter + order) | `cmd/tl-lead/team_issues.go` | new |

- **TeamName / parseTeamRef:** `^[a-z0-9][a-z0-9-]{0,39}$`. `parseTeamRef` accepts `name`, `team:name` or a path and returns the folder. A name that fails the grammar refuses, so `--team ../x` can never escape `teams/` (ARCH-SECURE).
- **TeamDoc:** the YAML frontmatter of `team.md`, with `type: team`, `owns: [repo…]` and `steward: bool`. It is decoded into a typed struct; unknown keys are tolerated. tl-lead never rewrites `team.md` after creating it, so no read-modify-write path exists to lose fields (ARCH-EVOLVE). Missing or broken frontmatter, or `type` other than `team`, is an error that names the file. An entry in `owns` must be a bare repo name (pair's repo grammar, below).
- **hubClaimLine:** mirrors pair's `couchcore.parseHubClaim`:
  - bare repo: relative, clean, not `.`/`..`, no `\:!#`, no whitespace or control characters;
  - or `repo:N` with canonical N ≥ 0.

  tl-lead can't import pair's internal package. The copy is about 15 lines, guarded by a test that runs pair's own accept and reject vectors from `hub_claim_test.go` (copied as a table with a pointer to the source). A live conformance check asks the real Couch to parse the scaffolded `.hub` once pair#694 gives it a parse verb. Until then the vectors are the contract, and the plan says so.
- **scaffold(name, owns, claims, steward) → map[path]bytes**, with two files:
  - `.hub`: the claim lines, each validated.
  - `team.md`: frontmatter, then `# team:<name>`, then a short "Asks other teams for" stub.

  It is pure, and a golden test covers it.
- **teamTable(teams) → rows + warnings:**
  - each row: address, owns, steward.
  - warnings: a repo owned by two teams; zero or several stewards; an unreadable `team.md`. An unreadable `team.md` is reported, not fatal.
- **commitPaths(root, team, links) → []string:** the team folder plus each `workshop/projects/*.md` symlink whose target resolves inside it. It is pure over a list of (link, target) pairs.
- **unclaimedIssues(states, limit):** keeps `status == open` with no owner and no `owner_error`, sorted by ID descending and capped at `limit`. An issue with an `owner_error` is listed apart as `owner unreadable`, never as unclaimed. The same rule as sdlc state: a failed read is never absence.

### Integration points

| Name | Lives in | Status | Wraps |
|------|----------|--------|-------|
| `resolveTeam` | `cmd/tl-lead/team.go` | new | filesystem walk, `git rev-parse --show-toplevel` |
| `writeScaffold` | `cmd/tl-lead/team.go` | new | filesystem |
| `teamGit` | `cmd/tl-lead/team_commit.go` | new | `git` |
| `SDLC.RepoIssues` | `cmd/tl-lead/sdlc.go` | modified | `sdlc state --json` |

- **resolveTeam(cwd, flag):**
  - With a flag: parse the reference, then `<ops toplevel>/teams/<name>`, which must hold `.hub`.
  - Without one: walk up from cwd to the first directory holding `.hub` whose parent is `teams`. Today's Couch also reads a `.hub` at a hub worktree's root; that isn't a team, and the parent check excludes it.
  - The result is a `Team{Root, Name, Dir}`.

  Injected into every verb through `deps.team` (a func), so tests point it at a temp tree.
- **writeScaffold:**
  - Each file is created with `O_EXCL`.
  - If the file exists with the same bytes, it counts as already written. This makes an interrupted `team new` resume on a rerun.
  - If it exists with different bytes, the command refuses before writing anything. Both files are checked first.
  - `projects/` is created with `MkdirAll`.
- **teamGit:** runs `git -C root …` and returns stdout or an error carrying stderr. Following the existing `gitNewestLanding` and `gitignoreGap`, tests run the real git against a temp repo with a bare `origin`, so production and test cross the same boundary and the test is the conformance check (ARCH-MOCK). No fake is needed: git is local and deterministic.
- **SDLC.RepoIssues(repoRoot):** runs `sdlc state --json` in the repo's primary checkout (`RepoRoot` already resolves it) and decodes `issues[]` into `{id, status, title, owner?, owner_error?}`, tolerating unknown keys.
  - The fake adds a `states map[repoPath][]IssueState`.
  - A live check (`TL_LEAD_LIVE=1`) decodes the real output in this repo and asserts at least one issue and a known status.

### Principles

- **ARCH-DRY:** `RepoRoot` and the `SDLC` seam are reused. The `.hub` grammar is a deliberate, guarded copy, as stated above. The session-file default goes through one helper, `sessionFile(team, args)`, used by every status verb.
- **ARCH-PURE:** every decision is a pure function. The shell only reads, writes, and runs git and sdlc.
- **ARCH-SECURE:**
  - Inputs: `team.md` and `.hub` are hand-edited, and `--team` comes from the user.
  - Each is parsed into typed values at the boundary.
  - A bad name or claim refuses with the line. A broken `team.md` is reported by `team list` and refused by `team issues`.
  - No credentials are involved; `git push` uses the operator's existing remote.
- **ARCH-ORDER:**
  - **`team new`:** holds no state between runs. Each file is idempotent by content.
  - **`team commit`:** a sequence of add, commit (skipped when nothing is staged under the paths), pull with rebase and autostash, then push. A stop at any step resumes on a rerun: the commit already exists, so it only pulls and pushes.
  - **The event most likely to be mishandled:** a concurrent push by another team between the pull and the push. Push is retried after a fresh pull, at most 3 times, then it refuses and names the last error.
  - **A rebase conflict:** `git rebase --abort`, keeping the local commit, then refuse with "resolve by merging both sides" (help: TWO MODES).
  - **A test drives each stop point.**
- **ARCH-CONSTRAINTS:** these are interactive CLI verbs.
  - `team issues` costs, per owned repo, one `sdlc workspace` and one `sdlc state --json`: 1–3 s each, sequential. A team owns 1 to 5 repos, so it takes under about 15 s. The existing 60 s per-call timeout bounds it.
  - `team list` reads one small file per team.
  - Other categories don't apply: nothing runs per keystroke or in the background.
- **ARCH-FUNERAL:**
  - **A team folder** lasts as long as the team. It ends when the operator retires the team with `git rm -r teams/<name>`; the help says so, and no verb exists for it.
  - **`team commit`** creates commits, which are history.
  - **Autostash entries:** git drops them on a clean apply. A conflicting apply leaves one stash entry, which the refusal names.
  - **Nothing else durable is created.**
- **ARCH-EVOLVE:**
  - `team.md` is read-only after creation, and unknown keys are ignored on read.
  - `.hub` is written once, with claim lines only. Its directives belong to pair#694.
  - The `sdlc state` decode ignores unknown keys.
  - Test: decoding a `team.md` that has extra keys and a nested map succeeds, and the file is never rewritten.
- **ARCH-MOCK:** covered under Integration points.

### State design table (tl-worker rule; scaffold and commit)

Actions: rerun the verb; edit the file by hand.

| Internal state / failure mode | Visible state | Recovery | Possible loss |
|---|---|---|---|
| `team new` stops after `.hub` | "team.md missing" from `team list` | rerun `team new` (same args) | none |
| `team new` rerun with different args | refusal naming the differing file | edit by hand or remove the file | none |
| `team commit` stops after commit, before push | `git status` ahead of origin | rerun `team commit` | none |
| push races another team | retried; after 3, refusal | rerun | none |
| rebase conflict | refusal, rebase aborted, local commit kept | merge both sides by hand, rerun | none |
| autostash apply conflict | refusal naming the stash | `git stash pop` by hand after resolving | another team's uncommitted edits stay in the stash until popped |

Covered: every interruption of both verbs. Not covered: another team's TL committing the same file. One writer per file is the rule (help: THE PROJECT DIRECTORY), and the path limit makes it hold across teams.

## Tasks

> **Superseded (2026-10-19, Revisions):** Tasks 1–7 below are replaced by R1–R7 in the Revisions section. Kept for history; do not execute them.

### Task 1: team model and resolution

**Files:** create `cmd/tl-lead/team.go` and `cmd/tl-lead/team_test.go`. Modify `cmd/tl-lead/main.go`: add `deps.team`, the persistent `--team` flag, and register `newTeamCmd(d)`.

- [ ] Write the failing tests:
  - `TestParseTeamRef`: `echo`, `team:echo` and a path are accepted; `../x`, `Echo`, `team:` and a 41-character name are refused.
  - `TestHubClaimVectors`: pair's accept and reject vectors.
  - `TestParseTeamDoc`: good, with unknown and nested keys; no frontmatter; `type: project`; `owns: [../x]`.
  - `TestResolveTeam`, on a temp tree:
    - nested cwd → team;
    - root `.hub` (not under `teams/`) → no team;
    - `--team` naming a folder without `.hub` → refusal.
- [ ] Run `go test ./cmd/tl-lead -run 'TeamRef|HubClaim|TeamDoc|ResolveTeam'` and confirm it fails.
- [ ] Implement `TeamName`, `parseTeamRef`, `hubClaimLine`, `TeamDoc`/`parseTeamDoc` and `resolveTeam`.
- [ ] Run the tests again and confirm they pass.
- [ ] Commit `#456: team model and --team resolution`.

### Task 2: `team new` and `team list`

**Files:** extend `team.go` and `team_test.go`; add golden `testdata/team/new-echo/{.hub,team.md}`.

- [ ] Write the failing tests:
  - `TestScaffoldGolden`: `echo --owns echo --claim pair:5`.
  - `TestTeamNewAdoptsExisting`: a folder with other files gains the two files, and the other files are untouched.
  - `TestTeamNewResumes`: `.hub` already present with the same bytes → it writes `team.md`.
  - `TestTeamNewRefusesDiffering`: nothing written.
  - `TestTeamNewWarnsUntracked`: the gitignore check uses real git in a temp repo, and a session-file path that git would track warns. This reuses `gitignoreGap`'s check-ignore idiom, extracted to `gitIgnored(root, rel)`.
  - `TestTeamTable`: two teams own `pair` → a warning; no steward → a warning; a broken `team.md` → a row marked unreadable.
  - `TestTeamListOwner`: `team list pair` prints `team:devinfra`.
- [ ] Run them and confirm they fail.
- [ ] Implement `scaffold`, `writeScaffold`, `teamTable` and the cobra verbs `team new` and `team list [REPO]`.
- [ ] Run them and confirm they pass.
- [ ] Commit.

### Task 3: session FILE default

**Files:** `status.go`, `instructions.go`, `index.go`, `migrate_apply.go` and their tests.

- [ ] Write the failing tests:
  - `TestSessionFileDefault`: one `tl-status-*.md` → that file; none → a refusal; two → a refusal naming both; an explicit FILE wins.
  - One status verb (`check`) run with no FILE from inside a temp team.
- [ ] Run them and confirm they fail.
- [ ] Implement `sessionFile(d, args, n)`, and turn each status verb's `ExactArgs(k)` into `RangeArgs(k-1, k)`. FILE stays first: when the count is k-1, FILE is the default. Update the `Use:` strings to `[FILE]`.
- [ ] Run them and confirm they pass. Run `go test ./cmd/tl-lead` and confirm the whole package passes.
- [ ] Commit.

### Task 4: `team commit`

**Files:** create `team_commit.go` and `team_commit_test.go`.

- [ ] Write the failing tests. Each uses real git with a temp bare origin and two clones:
  - `TestCommitPaths`: pure; one link points into the team, one points elsewhere.
  - `TestTeamCommitPathLimited`: dirty files in `teams/echo` and `teams/devinfra` → only echo's files are committed; devinfra's stay dirty after the pull.
  - `TestTeamCommitIncludesProjectLink`.
  - `TestTeamCommitResumesAfterCommit`: commit and stop (an injected push failure), then rerun → pushed, with one commit.
  - `TestTeamCommitRacesPush`: the other clone pushes first → a pull, then success.
  - `TestTeamCommitRebaseConflict`: refusal, no rebase left in progress, the local commit kept.
  - `TestTeamCommitNothing`: a clean tree → "nothing to commit", no error.
- [ ] Run them and confirm they fail.
- [ ] Implement `commitPaths` and `teamCommit(git, team, msg)`:
  1. `git add -A -- paths`.
  2. `git diff --cached --quiet -- paths`: if anything changed, `git commit -m msg -- paths`.
  3. Loop, at most 3 times: `git pull --rebase --autostash`, then `git push`.

  Inject the step runner so a test can fail at any step.
- [ ] Run them and confirm they pass.
- [ ] Commit.

### Task 5: `team issues`

**Files:** `sdlc.go` (`RepoIssues`); create `team_issues.go` and `team_issues_test.go`; the `fakeSDLC` in `trains_test.go`; `conformance_test.go`.

- [ ] Write the failing tests:
  - `TestUnclaimedIssues`: pure; mixed statuses, owners and an owner error; checks the limit and order.
  - `TestTeamIssuesCmd`: the fake `RepoRoot` and `RepoIssues` for two owned repos. Output is grouped by repo, `pair#702  title`, with `owner unreadable` rows apart. A repo whose `RepoRoot` fails prints `pair: ? (error)` and the others still print.
  - `TestRepoIssuesLive`: requires `TL_LEAD_LIVE`.
- [ ] Run them and confirm they fail.
- [ ] Implement.
- [ ] Run them and confirm they pass.
- [ ] Commit.

### Task 6: help, worker rule, atlas

**Files:**
- `cmd/tl-lead/help.md`;
- `cmd/tl-worker/help.md` and `cmd/tl-worker/testdata/rules.tsv`;
- `cmd/tl-lead/main_test.go` (G3 topics and the subcommand names);
- `atlas/workflow/tl-skills.md`.

- [ ] Update the help tests first:
  - tl-worker `rules.tsv`: drop `cross-repo … one-line proposal to the TL`; add `filing-direct own sdlc issue new` (there) and `filing-hub own tell your hub in one line`.
  - tl-lead: `TestLeadHelpNamesSubcommands` gains `team new`, `team list`, `team commit` and `team issues`. A G3 topic, `TEAMS`, pins `team:<name>`, `.hub` and `team.md`, path-limited, never relay, and steward.
- [ ] Run them and confirm they fail.
- [ ] Write the prose:
  - **tl-lead TEAMS section:** the folder contract (§2), addressing, `--team`, the four verbs, routing, the steward's machine-wide matters, and retiring a team.
  - **ROLES:** "One TL per team".
  - **DISPATCH:** replace "The TL files or folds cross-repo proposals" with the direct-filing rule.
  - **Status file:** `couch-owner: team:<name>` once pair#694 lands. Until then, the hub slot.
  - **tl-worker FILING ISSUES:** file in the repo where the problem lives, from that repo's checkout (`sdlc issue new` there, then `sdlc unclaim` unless you will work it), and tell your hub in one line. The hub never relays a filing. "Report to `hub`" now means the team that claims your slot.
  - **atlas:** a "Teams (ariadne#456)" entry.
- [ ] Run `go test ./cmd/tl-lead ./cmd/tl-worker` and confirm it passes.
- [ ] Commit.

### Task 7: verify

- [ ] `make test`, in the background, before the close.
- [ ] **Manual check in ops, read-only except the dry adopt:**
  1. `tl-lead team list` from `teams/devinfra` reports that devinfra has no `.hub` yet.
  2. `tl-lead team new devinfra --owns ariadne,pair,parley.nvim --steward` in a scratch copy of the ops tree (not the live one) shows the two files.
  3. Send the output to the hub. Adopting the real devinfra folder, and `team new echo`, are the hub's runs after landing (it owns ops).

## Done when (this issue's share)

> **Superseded (2026-10-19):** see "Done when (revised)" in the Revisions section.

- `tl-lead team new|list|commit|issues` exist, with tests:
  - `team new` scaffolds a folder pair#694's `.hub` grammar accepts;
  - `--team` resolves from inside a team folder;
  - status verbs default FILE to the team's session file.
- The tl-lead help and the tl-worker help state teams, `team:<name>`, direct filing with a one-line note to your own hub, path-limited commits and the steward. The help tests pin these.
- The live proof (echo team, cross-team exchange) is handed to the hub, to run once pair#694 lands.

## Revisions

### 2026-10-19: one team per repo (operator decision, design §7–§9)

**Reason:** the operator chose one team per ops repo. A team's TL is the `:0` slot of its own ops repo: `ops:0` (devinfra) and `ops-echo:0` (echo). Couch's existing multi-hub support carries it (`couch ~/workspace/ops ~/workspace/ops-echo`), so there is no team concept in Couch and pair#694 is parked. The steward is the first hub on Couch's command line; it is not a field in `team.md`. A TL never lives inside a product repo (branch protection).

**Dropped** from the plan above:
- the `team:<name>` address;
- `--team` and the walk up to a `.hub` (Task 1's resolution; D3's default FILE);
- `team commit` and path-limited commits (Task 4; D4);
- `team.md`'s `owns:` and `steward:` fields (D1, D2 and D5 change as below).

**The revised scope:**

1. **Hub contract (help only).**
   - A slot whose checkout has `.hub` at its root is a hub and runs `tl-lead`; every other slot runs `tl-worker`. No new address kind.
   - `tl-lead` help: ROLES becomes one TL per team, the team being its ops repo. A new TEAMS section covers the steward duties, cross-team routing and `team new`.
   - `tl-worker` help, FILING ISSUES: file in the repo where the problem lives (`sdlc issue new` there, then `sdlc unclaim` unless you'll work it) and tell your hub in one line. Hubs never relay a filing. "Report to `hub`" means the hub that claims your slot, which is how Couch already works.
2. **Ownership has one source, `.hub`** (ARCH-DRY, design §9).
   - Claim lines only: usually whole repos, rarely a slot. tl-lead's grammar mirrors pair's, checked against pair's test vectors (`hubClaimLine`, unchanged from Task 1).
   - `team.md` at the ops repo root holds frontmatter `type: team`, a pointer to `.hub` and prose ("Asks other teams for"). It carries no repo list and no steward flag.
3. **`tl-lead team list [REPO]`** finds teams as the sibling checkouts in the workspace (the ops repo's parent folder) that have `.hub` at their root.
   - Without REPO, it prints each team's name (the repo's folder name), its hub address `<name>:0` and its claims.
   - With REPO, it prints the owning hub's address. Owner resolution follows Couch's rule (`couchcore/hub_ownership.go`): an exact slot claim beats a repo claim; between equal claims, the first hub on the command line wins, and tl-lead can't see that order, so it lists every claimant and says so. A repo nobody claims belongs to the first hub (the steward).
   - It warns when one repo is claimed by two hubs.
4. **`tl-lead team issues`:** new issues nobody has claimed in the repos this hub's `.hub` claims, newest first, through `sdlc state --json` (Task 5, unchanged).
5. **`tl-lead team new PATH [--claim C]...`:** creates or adopts an ops repo.
   - It writes `.hub` (claim lines only) and `team.md`. A missing folder is created and `git init` is run in it. An existing git checkout is adopted: only the two missing files are written.
   - Each file is written exclusively. Identical bytes count as already written, which makes the command resumable. Any differing file refuses before anything is written.
   - It never commits or pushes, and never creates the GitHub repo: those are the operator's and the hub's steps. It prints the next steps.
6. **`tl-lead fleet resume [--apply]`:** run only when the operator asks to resume the fleet; never at startup (operator, §9).
   - **The join.** For each slot of the repos this hub claims (`sdlc fleet inventory` through the existing `SDLC` seam), it reads three things:
     - its liveness, from `couch --actors` (existing `CouchActors` seam);
     - its in-flight work: a non-terminal issue whose owner workspace is that slot, from `sdlc state --json` (`RepoIssues`, Task 5);
     - when it is down with work in flight, `couch --resume-candidates repo:N --json` (a new `Couch` read seam with a stateful fake, plus a `TL_LEAD_LIVE` conformance decode).
   - **The classes:**
     - *running*: left alone;
     - *idle*: down with no work, left down;
     - *resume*: down, with work and a proven conversation;
     - *ambiguous*: down with work but no proven conversation, or a conversation but no claim. These are listed for Needs you, never guessed.
   - **The run.** A dry run prints the classes. `--apply` runs `couch --resume repo:N` one slot at a time, since the call is heavy and Couch says one at a time. Each outcome is reported as resumed, refused (with Couch's reason), or uncertain.
   - **Order and retries (ARCH-ORDER).** Each resume is an external effect whose outcome may be uncertain. A rerun re-reads liveness first, so a slot already resumed is *running* and is skipped, and retries never duplicate. A refusal stops nothing: the next slot proceeds. An uncertain outcome is reported, not retried, and the TL peeks before acting.
   - **Pure planner.** `planResume(slots, actors, issues, candidates)` returns the classes, with a table test that covers every class and the stale-read case.

**Decisions now:**
- D1: `team.md` points at `.hub` and holds no list.
- D2: `.hub` claims come from `--claim` (repo or `repo:N`).
- D3 is withdrawn: FILE stays required, because `--team` is gone.
- D4 is withdrawn.
- D5: the steward is the first hub on the command line. tl-lead only documents the role, and `team list` notes that it can't see the order.

**Revised tasks** (each test-first, as in the original tasks):
- **R1:** `hubClaimLine`, a `.hub` reader and `ownerOf` (see the plan-quality round-1 revision; the `team.md` parse is dropped).
- **R2:** `team new PATH`, covering create, adopt, resume and refuse.
- **R3:** `team list [REPO]`, over a temp workspace of sibling repos.
- **R4:** `RepoIssues` and `team issues`, from Task 5 with `owns` replaced by `.hub` repo claims.
- **R5:** the `Couch` resume seam plus its fake, `planResume` (targets come from `ownerOf`), and `fleet resume [--apply]`.
- **R6:** help for tl-lead and tl-worker, plus atlas.
- **R7:** the full suite. Live: a dry run of `team list`, `team issues` and `fleet resume` from ops (all reads). `team new` and the live proof (`ops-echo`, a cross-team exchange) are the hub's runs.

**State design table (`fleet resume`).** Actions: dry run, `--apply`, rerun.

| Internal state / failure mode | Visible state | Recovery | Possible loss |
|---|---|---|---|
| slot down, claim, proven conversation | resume | `--apply` | none |
| slot down, claim, no proven conversation | ambiguous (Needs you) | the operator picks: `--resume --candidate`, or `--reboot` | none (tl-lead never chooses) |
| slot down, no claim, conversation exists | ambiguous | the operator decides | none |
| `--resume` refused (busy, lock, stale binary) | refused, with the reason | fix per the reason, then rerun | none |
| `--resume` outcome uncertain (timeout) | uncertain | peek the slot, then rerun (it is skipped if running) | none |
| a Couch or sdlc read fails | `?` for that repo or slot, others proceed | rerun | none |

Covered: every class and every outcome of the effect. Not covered: Couch's own recovery confirmations (PID approval). Those are the hub's, through `--recovery-confirmation`, and tl-lead never approves them.

### 2026-10-19: plan-quality round 1 (PQ-1, PQ-2 and Minors)

**PQ-2, one ownership resolver (ARCH-DRY).** `ownerOf(slot, hubs)` mirrors Couch's rule:
- An exact `repo:N` claim beats a repo claim.
- Several hubs at the same rank are a tie. tl-lead can't see the command-line order, so it returns every claimant and the result is not this hub's alone.
- A slot nobody claims belongs to the first hub on the command line, the steward. tl-lead can't tell which hub that is, so it reports *unclaimed*.

`team list REPO` and `fleet resume` both call it, and `fleet resume` acts only on slots that resolve to this hub alone. A tie or an unclaimed slot is listed as ambiguous and never resumed. The inventory covers every repo in the fleet, not just those this hub claims, so another hub's exact slot claim always counts.

**Minor: failed reads are unknown.**
- A failed `couch --actors` read makes every slot's liveness unknown. Nothing is classed as down, and nothing is resumed.
- A failed `sdlc state` read for a repo makes that repo's claims unknown.
- An owner workspace that isn't in the inventory is listed as unknown, never matched.

**Minor: the filing rule** in tl-worker help uses `sdlc issue new --no-claim` (verified in `sdlc issue new --help`) for a filing you won't work, instead of `new` followed by `unclaim`.

**Minor: YAGNI.** R1 drops the `team.md` parse. Nothing reads `team.md`: `team new` writes it from a template, and routing reads only `.hub`. R1 is now just `hubClaimLine` plus a `.hub` reader that keeps comment lines and skips them.

**Minor:** the original Tasks 1–7 and the original Done when are marked superseded.

**Done when (revised).** This replaces the issue's and the plan's earlier Done when.
- `tl-lead team new PATH [--claim C]...` creates or adopts an ops repo with `.hub` (claim lines only, accepted by pair's grammar vectors) and `team.md` (pointing at `.hub`). It is resumable and refuses on any differing file. Tested.
- `tl-lead team list [REPO]` lists sibling ops repos with `.hub`, and with REPO names the owner through `ownerOf`, listing every claimant on a tie. Tested against a temp workspace.
- `tl-lead team issues` lists open issues nobody has claimed in the repos this hub claims. Tested against the `SDLC` fake; the live decode is behind `TL_LEAD_LIVE`.
- `tl-lead fleet resume` (dry run) classifies each slot that resolves to this hub as running, idle, resume or ambiguous. A failed read is unknown, never down. `--apply` resumes the *resume* class one slot at a time and reports each outcome. Tested over the stateful Couch fake, including the rerun-skips-running case; the live decode of `--resume-candidates` is behind `TL_LEAD_LIVE`.
- The tl-lead help states the hub contract, the steward duties (the first hub on Couch's command line), routing by `.hub`, and the four verbs. The tl-worker help states direct filing (`--no-claim` when you won't work it) with a one-line note to your own hub, and that hubs never relay. Help tests pin both.
- **Live:** read-only dry runs of `team list`, `team issues` and `fleet resume` from ops are sent to the hub. The live proof (`ops-echo` plus a cross-team exchange) is no longer blocked on pair#694, which is parked. The steps that need it (a new GitHub repo, and a Couch restart with two hubs) are the operator's and the hub's, run after this lands, starting from `tl-lead team new ~/workspace/ops-echo`.

### 2026-10-19: plan-quality round 2 (PQ-7, the layers tl-lead cannot see)

**The rule for this family: tl-lead's ownership copy covers only what Couch persists.** Couch's `Resolve` (`couchcore/hub_ownership.go`) has three layers above the `.hub` claim ranks, and here is what `ownerOf` does with each:
1. **A hub's own slot.** Mirrored: a repo with `.hub` at its root owns its `:0`, and `ownerOf` checks this first.
2. **Runtime moves** (`couch --move-slot`, the `overrides` map). Not on disk, and Couch exposes no JSON read of a slot's live owner today: ownership shows only in its menu. They are ephemeral: pair's `TestHubMoveAuthorityRevisionAndDurability` shows a restarted Couch resolving from `.hub` alone. So on the cold start that `fleet resume` exists for, `.hub` is the whole truth.
3. **Command-line order.** It breaks ties and owns unclaimed slots. Not visible, so ties and unclaimed slots are ambiguous and never resumed (round 1).

**What `fleet resume` does about layer 2:**
- The dry run and `--apply` both say "Couch runtime moves are not visible; a slot moved with --move-slot since Couch started is classed by .hub".
- `--apply` lists the exact slots it will resume before resuming any.
- **Consequence bound:** resuming restarts a slot's own conversation. It changes no ownership, and the resumed worker still reports to the hub that owns it (Couch resolves `hub` at send time). So the worst case of a stale read is another hub's slot resumed early: no lost work and no misrouted report.

**Follow-up (proposed to the hub, for pair):** a JSON owner read, for example an `owner` field on `couch --actors --json` rows and on `--resume-candidates`. Once it exists, `fleet resume` takes Couch's answer, and `ownerOf` becomes only the fallback for slots that are down. Until then the caveat above stands.

**Test:** `TestPlanResumeNamesInvisibleLayers` checks that the dry run carries the caveat and that a tied or unclaimed slot is never in the *resume* class.
