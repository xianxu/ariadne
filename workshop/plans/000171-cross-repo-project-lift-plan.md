# Cross-Repo Project Lift Implementation Plan

> **For agentic workers:** Consult AGENTS.md Section 3 (Subagent Strategy) to determine the appropriate execution approach: use superpowers-subagent-driven-development (if subagents are suitable per AGENTS.md) or superpowers-executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Lift project management into the sdlc spine so project records live in coding repos, are discovered across peers at close time, receive all-match task updates, and are committed into their peer repo only when git state makes it safe — retiring the last brain coupling (`close --brain-dir`).

**Architecture:** One pure cross-peer discovery function (`project.DiscoverByIssueRef`) replaces the single-brain-dir glob and feeds three consumers — the close gate, `sdlc resolve`, and (via a CLI surface) parley's super-repo search. Close-time updates loop the existing per-file tick/upsert helpers over every match. A pure `planPeerWrites` planner decides per-repo whether git state authorizes a scoped auto-commit, applied by a thin `gitRunner.GitInDir` shell. The project schema's `done` state stops requiring a committed baseline so pre-baseline records convert cleanly, and the four terminal legacy records migrate into their center-of-gravity repo's `workshop/history/projects/`.

**Tech Stack:** Go (`cmd/sdlc`, `pkg/vocab`, `cmd/sdlc/internal/project`), CUE (`construct/vocabulary/project.cue`), `//go:generate` + `make vocab-embed` codegen, Lua/vimscript (parley.nvim super-repo search), markdown artifacts.

**Arch principles in play** (from `sdlc arch-principles`):
- **ARCH-DRY** — reuse `resolveRepoDir`'s sibling walk, the existing tick/upsert helpers, and the `gitRunner` seam; do not introduce a second peer walker or a second project parser.
- **ARCH-PURE** — discovery is deterministic given a filesystem root (tested against temp dirs, the existing `resolve_test.go` pattern); the commit decision is a pure planner over git-state snapshots with IO pushed to a thin shell.
- **ARCH-PURPOSE** — the issue's Done-when is the whole lift: discovery + all-match update + safe peer commit + fleet navigation (resolve + parley) + residency docs + the four migrations. Delivering only the close lookup would miss the point; every piece below is in scope.

---

## Core Concepts

The lift operates on **project records** (typed markdown under a repo's `workshop/projects/`) addressed by **qualified issue refs** (`repo#id`). The close gate must answer "which project records, anywhere in the fleet, reference this closing issue?" and then "for each, is it safe to commit the update into its repo?" Those two questions are the two pure cores below; everything else is a thin seam or a reuse of an existing helper.

### Pure entities (the conceptual core)

| Name | Lives in | Status |
|------|----------|--------|
| `ProjectMatch` | `cmd/sdlc/internal/project/discover.go` | new |
| `DiscoverByIssueRef` | `cmd/sdlc/internal/project/discover.go` | new |
| `RepoGitState` | `cmd/sdlc/peerwrite.go` | new |
| `PeerWriteDecision` | `cmd/sdlc/peerwrite.go` | new |
| `planPeerWrites` | `cmd/sdlc/peerwrite.go` | new |
| terminal baseline guard | `construct/vocabulary/project.cue` | modified |

- **ProjectMatch** — one discovered project record referencing the issue: its absolute path, owning repo dir, repo basename, and whether it was found in a peer's `workshop/projects/` (the canonical home) or the deprecated `brain/data/project/` legacy home.
  - **Relationships:** N:1 with the closing issue (many projects may legitimately reference one issue — multiple matches are membership, not ambiguity); N:1 with a repo dir.
  - **DRY rationale:** Replaces `FindByIssueRef`'s single-path return and its refuse-on-multiple error. One match struct is consumed by the close gate, `sdlc resolve`, and the `sdlc project find` CLI surface parley calls — no per-consumer re-globbing.
  - **Future extensions:** Add a `Status`/`FM` field if a consumer needs to filter matches by lifecycle state without re-parsing.

- **DiscoverByIssueRef** — `DiscoverByIssueRef(parentDir, repoName, issueID string, scope DiscoverScope) ([]ProjectMatch, error)`: enumerate sibling repos under `parentDir`, glob each one's `workshop/projects/*.md` (always) and — when `scope == ActiveAndArchive` — also each one's `workshop/history/projects/*.md`, for the marker `[<repoName>#<issueID>`; also scan the legacy `<parentDir>/brain/data/project/*.md`. Return every match, deduped by resolved path. Deterministic given the filesystem; tested against temp dirs.
  - **The scope parameter is load-bearing — the three consumers differ:** the **close gate** passes `ActiveOnly` (it must not re-tick a `done` project already archived in `workshop/history/projects/`); **`sdlc project find` / `resolve` / parley** pass `ActiveAndArchive` (navigation must find a record regardless of which repo *or lifecycle location* holds it — matching how issue resolution already unions home+plans+archive in `familyFiles`, `resolve.go:248-271`). `DiscoverScope` is a small enum (`ActiveOnly`, `ActiveAndArchive`) in the same file.
  - **Archive dir is derived, not hardcoded:** the `workshop/history/projects` path comes from `vocab.ArchiveSubdir(vocab.Project().Discovery().Archive, vocab.ArchiveProjects)` (`vocab.go:115`), not a string literal — ARCH-DRY with the #181 archive layout.
  - **Relationships:** Produces `[]ProjectMatch`. Reuses the parent-dir enumeration logic already in `resolveRepoDir` (`resolve.go:181`) — extract a shared `siblingRepoDirs(parentDir)` helper (placed in `internal/project`, see Task 2.1) rather than duplicating the walk.
  - **DRY rationale:** The single walker for "find project records referencing X across the fleet," shared by three consumers that differ only in scope.
  - **Future extensions:** A predicate form (`DiscoverBy(pred func(ProjectMatch) bool)`) if future artifacts (roadmaps, targets) need the same cross-peer glob; the sibling-walk helper is already the reusable half.
  - **Legacy-home note:** A match found under `brain/data/project/` is returned with `Legacy: true` so the caller emits a deprecation warning nudging final migration. This keeps the still-active `metis-v2-experiment-algebra` (deliberately left in brain) working until it too migrates.
  - **Spurious-sibling guard:** the parent dir holds non-fleet siblings (`metis.bak/`, `worktree/`, `temp/`, dot-dirs, …). `resolveRepoDir` is immune because it exact-matches one basename, but a fleet-wide glob is not — a stale `metis.bak/workshop/projects/` copy would yield a duplicate match → spurious tick + (M3) spurious peer-commit. `siblingRepoDirs` skips obvious non-fleet dirs: names ending `.bak`, `worktree`, and dot-prefixed dirs. (Latent today — only `ariadne/` has `workshop/projects/` — but M6 creates `metis/workshop/history/projects/`, which a later `.bak` snapshot would duplicate once archive scanning is on.)

- **RepoGitState** — snapshot of one peer repo's git state relevant to the auto-commit decision: `Branch string`, `HasStagedChanges bool`. Read once per repo by the thin shell, passed by value into the planner.
  - **Relationships:** 1:1 with a repo dir that owns at least one `ProjectMatch` edited this close.
  - **DRY rationale:** Isolates the two git facts the decision needs so the planner takes data, not a git client.

- **PeerWriteDecision** — the planner's verdict for one repo's edits: `RepoDir string`, `Files []string`, `Commit bool`, `Reason string` (why not, when `Commit` is false), `NextAction string` (the exact manual command to finish, when not committed).
  - **Relationships:** 1:1 with a repo dir touched this close.
  - **Future extensions:** A `DryRun` field if `sdlc close --dry-run` later wants to preview peer commits.

- **planPeerWrites** — `planPeerWrites(edits map[string][]string, states map[string]RepoGitState, curRepoDir string) []PeerWriteDecision`: pure. For each repo with edits: if it is the current repo, defer to the normal close commit (no separate peer commit); else if `Branch == "main" && !HasStagedChanges`, decide `Commit: true` scoped to exactly `Files`; else `Commit: false` with `Reason` + `NextAction`. No IO.
  - **Relationships:** Consumes `RepoGitState`, produces `PeerWriteDecision`. Mirrors the `computeClose`/`applyClose` pure-plan-then-apply split (`close.go:334`/`close.go:696`).
  - **DRY rationale:** The two independent safety decisions (should the semantic update happen — always yes; does git state authorize a commit — this planner) are separated so a valid issue close is never blocked by a peer's branch state.
  - **Future extensions:** A policy struct if the on-main / clean-index rule ever needs per-repo overrides.

- **terminal baseline guard (modified)** — `construct/vocabulary/project.cue`'s compiled guard currently requires `deadline!`/`planned_finish!` for `committed || executing || paused || done`. Drop `|| status == "done"`: a `done` record archived from the pre-baseline era honestly has no committed baseline, while a properly-run new project still carries one because it passed through `executing` (which still requires it). Requirement removed, presence preserved.
  - **Relationships:** The schema authority; `pkg/vocab/project.json` is regenerated from it, and `construct/vocabulary/vet_test.sh` binds the two.
  - **DRY rationale:** One-line model change at the single source; no consumer restates the guard.

### Integration points (where pure meets the world)

| Name | Lives in | Status | Wraps |
|------|----------|--------|-------|
| sibling-repo FS enumeration | `cmd/sdlc/internal/project/discover.go` (or a shared helper) | new | filesystem |
| peer git-state reader + scoped committer | `cmd/sdlc/peerwrite.go` | new | `gitRunner.GitInDir` |
| close-gate wiring | `cmd/sdlc/close.go:565-650` | modified | discovery + planner |
| `sdlc project find` subcommand | `cmd/sdlc/project.go` | new | `DiscoverByIssueRef` |
| `sdlc resolve` project kind | `cmd/sdlc/resolve.go` | modified | `DiscoverByIssueRef` |
| parley `project` artifact class | `parley.nvim` (super-repo search) | new | `sdlc project find` / `sdlc resolve` |

- **sibling-repo FS enumeration** — `siblingRepoDirs(parentDir)` reads `filepath.Dir(curRoot)` and lists sibling directories, applying only the spurious-sibling skip-list (`.bak` suffix, `worktree`, dot-dirs). It does **not** filter by "has `workshop/projects/`" — that would change `resolveRepoDir`'s behavior (which enumerates all dirs). The per-dir glob naturally skips dirs lacking the home. Injected into `DiscoverByIssueRef`.
  - **Injected into:** `DiscoverByIssueRef`. Tests build a temp parent dir with sibling repos (the `resolve_test.go` / `migrate_test.go` fixture pattern) — no mocks.
  - **Future extensions:** Honor an explicit fleet manifest if implicit sibling discovery ever proves too broad.

- **peer git-state reader + scoped committer** — reads `RepoGitState` per repo (`git -C <dir> rev-parse --abbrev-ref HEAD`, `git -C <dir> diff --cached --quiet`) and, when the plan says commit, stages exactly the decided files and commits them via `r.GitInDir(repoDir, ...)`. Uses the `gitRunner` interface (`runner.go:24`) already used for cross-repo commits in `claim.go:311`, `merge.go:550`, and the two-repo `migrate.go:336`. Confirmed: `execGitRunner.GitInDir` uses `CombinedOutput()`, so `git diff --cached --quiet`'s non-zero exit surfaces as a Go error — `err != nil` correctly means "staged changes present."
  - **Injected into:** `applyClose` — but note the close path does **not** currently construct a `gitRunner` (unlike claim/changecode/merge/pr/push, each of which declares its own `var xRunner gitRunner = execGitRunner{}`). M3 introduces one into the close path and **changes `applyClose`'s signature** from `(stderr io.Writer, f *closeFlags, r closeResult)` to add a `gitRunner` and a `stdout io.Writer` (for the "committed X in Y" report — `applyClose` currently gets only `stderr`), threading it through every caller. The staged/compensating pattern from `projectclose.go:237` (`commitProjectClose`) is the model for not half-writing.
  - **Future extensions:** Batched multi-file commit if a single close ever edits several project files in the same peer.

- **close-gate wiring** — `computeClose` calls `DiscoverByIssueRef` instead of `FindByIssueRef`, loops the existing `TickMilestoneTaskRow`/`TickAllTaskRowsForIssue`/`UpsertDetailBlockFields` over each match, and records per-repo edits; `applyClose` reads git state, runs `planPeerWrites`, applies decisions, and prints the report.
  - **Injected into:** n/a — this is the seam itself.

- **`sdlc project find` subcommand** — `sdlc project find --issue <repo>#<id>` prints matching project paths (one per line, with repo + legacy flag). The stable CLI surface parley shells out to, so parley need not import Go.
  - **Future extensions:** `--json` output if parley wants structured fields.

- **`sdlc resolve` project kind** — teach `resolve` that a `project` ref (or a `--kind project` filter) resolves across the fleet via `DiscoverByIssueRef`, so project refs are addressable fleet-wide like issues.

- **parley `project` artifact class** — parley's super-repo search treats `project` as an always-cross-repo class: searching/jumping finds project records regardless of which repo holds the file, by calling `sdlc project find` / `sdlc resolve`.

---

## Milestone map (review boundaries)

Six milestones, each its own `sdlc milestone-close` boundary. Order reflects dependencies: M1 unblocks M6's conversion; M2 produces the discovery M3/M4 consume; M3 is the risky peer-git half (dedicated adversarial review); M4/M5 are navigation/docs; M6 is the one-time data migration.

- **M1** — Relax the terminal baseline guard (schema + codegen + conformance).
- **M2** — Cross-repo project discovery + all-match close update.
- **M3** — Safe peer-write commit mechanics.
- **M4** — Fleet navigation: `sdlc project find` / `resolve` + parley `project` class.
- **M5** — Residency documentation + base-layer ripple + propagate.
- **M6** — Migrate the four terminal legacy records.

Plan checkbox rows (also mirrored into the issue's `## Plan`):

- [x] M1 — relax `done` baseline guard; regenerate `project.json`; vet + conformance green
- [x] M2 — `DiscoverByIssueRef` + shared sibling walk; wire close gate to all-match update; brain legacy warning
- [ ] M3 — `RepoGitState`/`PeerWriteDecision`/`planPeerWrites` + thin git shell; off-main/staged report path; process-level multi-repo test
- [ ] M4 — `sdlc project find`; `sdlc resolve` project kind; parley `project` artifact class
- [ ] M5 — AGENTS.base §8 + brain-peer line, atlas, project datatype doc; `sdlc propagate-base`
- [ ] M6 — migrate charon-launch-push + shared-brain → nous, kaggle-ml-base-layer → kbench, metis-v1 → metis (history/projects/, schema-converted, validated, committed both sides)

---

## Chunk 1: M1 — Relax the terminal baseline guard

### Task 1.1: Change the compiled guard in the schema

**Files:**
- Modify: `construct/vocabulary/project.cue:88` (the `if status == ...` guard)
- Test: `construct/vocabulary/vet_test.sh` (existing; add/confirm a done-without-baseline case)

- [x] **Step 1: Write the failing conformance case**

Add to the project section of `construct/vocabulary/vet_test.sh` (after the existing `pjson` export ~line 34) a check that a minimal `done` record **without** `deadline`/`planned_finish` unifies against `#Project`. Concretely, export a fixture and assert `cue vet` passes:

```sh
# done record predating baseline discipline must validate (ariadne#171 M1)
# NOTE: vet_test.sh defines only $dir today — add a temp dir for the fixture.
tmp="$(mktemp -d)"; trap 'rm -rf "$tmp"' EXIT
cat > "$tmp/legacy_done.json" <<'EOF'
{"type":"project","name":"x","goal":"g","done_when":"w","status":"done"}
EOF
cue vet "$dir/project.cue" "$tmp/legacy_done.json" -d '#Project' \
  || { echo "FAIL: done record without baseline should validate"; exit 1; }
```

- [x] **Step 2: Run it to verify it fails**

Run: `bash construct/vocabulary/vet_test.sh`
Expected: FAIL — the current guard requires `deadline!` for `done`, so `cue vet` errors with an incomplete-value / missing-field message.

- [x] **Step 3: Make the minimal change**

In `construct/vocabulary/project.cue`, change the guard condition (currently line ~88):

```cue
	// compiled guard: every post-commit status except dropped carries the
	// baseline (a dropped project may have died pre-commit).
	if status == "committed" || status == "executing" || status == "paused" || status == "done" {
		deadline!:       string
		planned_finish!: string
	}
```

to:

```cue
	// compiled guard: a live post-commit project (committed/executing/paused)
	// carries the baseline. `done` is exempt: a properly-run project reaches
	// done via executing (so it still has one), but a record archived from the
	// pre-baseline era honestly has none — requiring it would force fabricated
	// dates on legacy migrations (ariadne#171). A dropped project may have died
	// pre-commit, so it was never required.
	if status == "committed" || status == "executing" || status == "paused" {
		deadline!:       string
		planned_finish!: string
	}
```

- [x] **Step 4: Run the vet test to verify it passes**

Run: `bash construct/vocabulary/vet_test.sh`
Expected: PASS.

### Task 1.2: Regenerate the derived faces and confirm the binding

**Files:**
- Regenerate: `pkg/vocab/project.json` (embed input; `//go:generate` at `pkg/vocab/project.go:10`)
- Regenerate: `construct/generated/vocabulary/` served face + `.source-sha` stamp (gitignored local build artifact; via `make weave`)

- [x] **Step 1: Regenerate the embed + served face**

Run `make vocab-embed` (the embed input) **and** `make weave` (the served face). **Correction (M1 execution, 2026-07-17):** `#Project` is a CUE `#`-definition and does NOT `cue export`, so the guard change does NOT alter the exported concrete blocks — `pkg/vocab/project.json` is a **byte-identical no-op** (there is no guard-condition diff to look for; the earlier expectation of one was wrong). The real regeneration deliverable at any cue-touching boundary is `make weave`: it rewrites `construct/generated/vocabulary/.source-sha` (a sha256 over the raw `.cue` text, so any edit invalidates it even when the export is identical). Skipping it leaves `vocabulary check` STALE — the M1 boundary review's Important finding. `construct/generated/` is gitignored, so this produces no committable diff, but it must be run so `vocabulary check` / `make check` gates go green.
Expected: `git diff pkg/vocab/project.json` is empty; `./bin/vocabulary check --output construct/generated/vocabulary` exits 0.

- [x] **Step 2: Run the vocab package tests**

Run: `go test ./pkg/vocab/...`
Expected: PASS (the drift test binding `project.cue` ↔ `project.json` is green).

- [x] **Step 3: Confirm the prose-drift binding + fix the prose guard restatement**

The prose-derives-from-model drift test is `pkg/vocab/prose_drift_test.go` — already covered by Step 2's `go test ./pkg/vocab/...`; no separate target needed. But `construct/datatype/project.md:29-30` restates the guard as `deadline | required after commit` / `planned_finish | required after commit`. "after commit" is now imprecise for `done` (a `done` record archived from the pre-baseline era has neither). Update those two rows to: `required for committed/executing/paused (a done record archived pre-baseline may lack it)`.
Run: `go test ./pkg/vocab/...`
Expected: PASS (and the prose no longer over-claims the requirement).

- [x] **Step 4: Commit**

```bash
git add construct/vocabulary/project.cue construct/vocabulary/vet_test.sh pkg/vocab/project.json construct/datatype/project.md
git commit -m "#171 M1: relax project baseline guard — done exempt from deadline requirement"
```

- [x] **Step 5: Milestone-close**

Run: `sdlc milestone-close --issue 171 --milestone M1`
Fix any Critical/Important findings before crossing; log the `Review-Verdict:` outcome in `## Log`.

---

## Chunk 2: M2 — Cross-repo project discovery + all-match close update

### Task 2.1: Extract the shared sibling-repo walk

**Files:**
- Create: `cmd/sdlc/internal/project/discover.go` (holds `siblingRepoDirs` + the skip-list)
- Modify: `cmd/sdlc/resolve.go:181` (`resolveRepoDir`) — call the extracted helper
- Test: `cmd/sdlc/resolve_test.go` (confirm `resolveRepoDir` still passes after extraction)

**Placement decision (resolves the layering the design flagged):** `DiscoverByIssueRef` lives in `internal/project`, which **cannot** import package `main`. So the shared walk goes **into `internal/project`** (`siblingRepoDirs`), and `resolveRepoDir` (package `main`) calls it — `main` already imports `internal/project` (see `close.go`'s `project.FindByIssueRef`). This keeps the lower package import-clean.

- [x] **Step 1: Extract the reusable half of `resolveRepoDir`, behavior-identical**

Read `resolve.go:181-224`. Today it lists **all** sibling directories under `filepath.Dir(curRoot)` (`os.ReadDir` + `IsDir`, resolve.go:191-195 — no git filter) and then exact/prefix-matches one basename. Extract `func siblingRepoDirs(parentDir string) ([]string, error)` in `internal/project`, returning absolute sibling dir paths. **Keep it behavior-identical for `resolveRepoDir`** — do NOT add a git-repo filter (that would change `resolveRepoDir`'s enumeration and break the "no regression" claim). The only filtering `siblingRepoDirs` applies is the spurious-sibling skip-list (names ending `.bak`, `worktree`, dot-prefixed) — verify none of the real fleet repos match those (they don't: ariadne, nous, metis, kbench, brain, charon, …), so `resolveRepoDir`'s behavior is preserved for every real lookup. Rebuild `resolveRepoDir`'s matching (exact > unique prefix > error) on top.

- [x] **Step 2: Run the resolve tests to verify no regression**

Run: `go test ./cmd/sdlc/ -run 'Resolve|RepoDir' -v`
Expected: PASS (behavior unchanged for all real-fleet lookups; only stale `.bak`/`worktree`/dot siblings are newly skipped).

### Task 2.2: `DiscoverByIssueRef` (all-match, cross-peer, brain-legacy-aware)

**Files:**
- Modify: `cmd/sdlc/internal/project/discover.go` (created in Task 2.1 for `siblingRepoDirs`)
- Test: `cmd/sdlc/internal/project/discover_test.go`

- [x] **Step 1: Write the failing test**

```go
func TestDiscoverByIssueRef_AllMatchesAcrossPeers(t *testing.T) {
	parent := t.TempDir()
	// peer repos with projects referencing metis#18, a non-matching one, a
	// brain legacy record, and an ARCHIVED match (history/projects).
	writeProject(t, parent, "metis", "workshop/projects", "p1.md", "[metis#18 M1]")
	writeProject(t, parent, "kbench", "workshop/projects", "p2.md", "[metis#18]")
	writeProject(t, parent, "nous", "workshop/projects", "p3.md", "[nous#9]")
	writeProject(t, parent, "brain", "data/project", "legacy.md", "[metis#18]")
	writeProject(t, parent, "metis", "workshop/history/projects", "old.md", "[metis#18]")

	// ActiveOnly (close gate): archived record NOT returned.
	act, err := DiscoverByIssueRef(parent, "metis", "18", ActiveOnly)
	if err != nil { t.Fatal(err) }
	if len(act) != 3 { t.Fatalf("ActiveOnly: want 3 (2 active + 1 legacy), got %d: %+v", len(act), act) }
	legacy := 0
	for _, m := range act { if m.Legacy { legacy++ } }
	if legacy != 1 { t.Fatalf("want 1 legacy match, got %d", legacy) }

	// ActiveAndArchive (find/resolve/parley): archived record IS returned.
	all, err := DiscoverByIssueRef(parent, "metis", "18", ActiveAndArchive)
	if err != nil { t.Fatal(err) }
	if len(all) != 4 { t.Fatalf("ActiveAndArchive: want 4 (incl. archived), got %d: %+v", len(all), all) }
}

func TestDiscoverByIssueRef_SkipsStaleSiblings(t *testing.T) {
	parent := t.TempDir()
	writeProject(t, parent, "metis", "workshop/projects", "p1.md", "[metis#18]")
	writeProject(t, parent, "metis.bak", "workshop/projects", "p1.md", "[metis#18]") // stale copy
	writeProject(t, parent, "worktree", "workshop/projects", "x.md", "[metis#18]")    // worktree tree
	got, _ := DiscoverByIssueRef(parent, "metis", "18", ActiveOnly)
	if len(got) != 1 { t.Fatalf("stale siblings must be skipped; want 1, got %d: %+v", len(got), got) }
}
```

(Provide a `writeProject` helper that `os.MkdirAll`s `<parent>/<repo>/<subdir>` and writes the file with a minimal project frontmatter + the marker in the body.)

- [x] **Step 2: Run it to verify it fails**

Run: `go test ./cmd/sdlc/internal/project/ -run TestDiscoverByIssueRef -v`
Expected: FAIL — `DiscoverByIssueRef` undefined.

- [x] **Step 3: Implement `DiscoverByIssueRef` + `ProjectMatch` + `DiscoverScope`**

```go
type DiscoverScope int

const (
	ActiveOnly       DiscoverScope = iota // workshop/projects only (close gate)
	ActiveAndArchive                      // + workshop/history/projects (find/resolve/parley)
)

type ProjectMatch struct {
	Path    string // absolute
	RepoDir string // absolute repo root that owns the match
	Repo    string // repo basename
	Legacy  bool   // found under brain/data/project (deprecated home)
}

// DiscoverByIssueRef returns every project record across the fleet whose body
// contains the marker "[<repoName>#<issueID>" (open-bracket form matches both
// "[metis#18]" and "[metis#18 M1]"). It always scans each sibling repo's
// canonical home (vocab.Project().Discovery().Home) and, when scope is
// ActiveAndArchive, the derived archive home (vocab.ArchiveSubdir(...)); it
// also scans the deprecated brain/data/project/*.md legacy home. Deterministic
// given parentDir; multiple matches are legitimate membership, not an error.
func DiscoverByIssueRef(parentDir, repoName, issueID string, scope DiscoverScope) ([]ProjectMatch, error) {
	marker := "[" + repoName + "#" + issueID
	disc := vocab.Project().Discovery()
	home := disc.Home // "workshop/projects"
	archive := vocab.ArchiveSubdir(disc.Archive, vocab.ArchiveProjects) // "workshop/history/projects"

	var out []ProjectMatch
	seen := map[string]bool{}
	scan := func(repoDir, relDir string, legacy bool) {
		files, _ := filepath.Glob(filepath.Join(repoDir, relDir, "*.md"))
		for _, f := range files {
			real, _ := filepath.EvalSymlinks(f)
			if real == "" { real = f }
			if seen[real] { continue }
			data, rerr := os.ReadFile(f)
			if rerr != nil { continue } // best-effort, matches FindByIssueRef
			if strings.Contains(string(data), marker) {
				seen[real] = true
				out = append(out, ProjectMatch{
					Path: f, RepoDir: repoDir,
					Repo: filepath.Base(repoDir), Legacy: legacy,
				})
			}
		}
	}

	siblings, err := siblingRepoDirs(parentDir) // skip-list applied inside
	if err != nil { return nil, err }
	for _, repoDir := range siblings {
		if filepath.Base(repoDir) == "brain" {
			// Legacy brain home. Under ActiveOnly (the close gate), skip a
			// legacy record whose status is terminal — the four migratable
			// records are `done` and must not be re-ticked during the M2→M6
			// window; only a still-active brain record (metis-v2) should tick.
			// Under ActiveAndArchive (navigation), include all legacy records.
			scan(repoDir, filepath.Join("data", "project"), true)
			continue
		}
		scan(repoDir, home, false)
		if scope == ActiveAndArchive {
			scan(repoDir, archive, false)
		}
	}
	if scope == ActiveOnly {
		out = dropTerminalLegacy(out) // reads status: front-matter of Legacy matches; drops done/dropped
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Path < out[j].Path })
	return out, nil
}
```

Notes: the `workshop/projects` / `workshop/history/projects` strings come from `vocab`, not literals (ARCH-DRY with #181). `siblingRepoDirs` (Task 2.1, same package) already applies the spurious-sibling skip-list, so a stale `metis.bak/` copy is never scanned. `dropTerminalLegacy` is a tiny helper that parses each `Legacy` match's `status:` front-matter (via the existing `projectdoc`/metadata decode) and drops `done`/`dropped` — closing the close-gate re-tick hazard the scope parameter exists for (plan-quality review finding, 2026-07-17). Add a test: a `done` legacy record is absent under `ActiveOnly` but present under `ActiveAndArchive`; an `active` legacy record (metis-v2 shape) is present under both.

- [x] **Step 4: Run the test to verify it passes**

Run: `go test ./cmd/sdlc/internal/project/ -run TestDiscoverByIssueRef -v`
Expected: PASS.

- [x] **Step 5: Add edge-case tests**

Cover: zero matches → empty slice, nil error; a symlinked duplicate path counted once; an unreadable file skipped; the same issue referenced by two projects in the *same* repo both returned.

Run: `go test ./cmd/sdlc/internal/project/ -v`
Expected: PASS.

### Task 2.3: Wire the close gate to all-match update

**Files:**
- Modify: `cmd/sdlc/close.go:565-650` (`computeClose` project-update block)
- Modify: `cmd/sdlc/close.go` `closeResult` — change `projectEditPath/projectText/projectEditText` (single) to a slice of per-file edits
- Test: `cmd/sdlc/close_test.go` (add a two-project all-match case)

- [x] **Step 1: Write the failing test**

In `close_test.go`, build a temp fleet where the closing issue (`ariadne#31 M1`) is referenced by **two** project files in two peer repos. Assert that after `computeClose`, both files have the milestone row ticked and both appear in `closeResult`'s edit list, and that the applied-messages mention both.

- [x] **Step 2: Run it to verify it fails**

Run: `go test ./cmd/sdlc/ -run TestClose.*Project.*AllMatch -v`
Expected: FAIL — current code finds one file and refuses on multiple.

- [x] **Step 3: Replace `FindByIssueRef` with `DiscoverByIssueRef` and loop**

In `computeClose`, replace the `project.FindByIssueRef(f.BrainDir, ...)` block (`close.go:569`) with (note `repoTop`/`repoName` already exist at `close.go:406`/`410`):

```go
parentDir := filepath.Dir(repoTop) // repoTop = gitx.RepoTopLevel(), close.go:406
matches, derr := project.DiscoverByIssueRef(parentDir, repoName, issueStr, project.ActiveOnly)
if derr != nil {
	cwarn(stderr, derr.Error()+" — skipping project update")
} else if len(matches) == 0 {
	cwarn(stderr, fmt.Sprintf("no project across the fleet references %s#%s — skipping project update", repoName, issueStr))
}
for _, m := range matches {
	if m.Legacy {
		cwarn(stderr, fmt.Sprintf("project %s is in the deprecated brain/data/project home — migrate it to <repo>/workshop/projects (ariadne#171)", filepath.Base(m.Path)))
	}
	// ... existing per-file logic: ReadFile(m.Path), TickMilestoneTaskRow /
	// TickAllTaskRowsForIssue, UpsertDetailBlockFields, shouldNudgeProjectRetro
	// ... append a projectEdit{path, repoDir, oldText, newText} to closeResult
}
```

The close gate passes `ActiveOnly` — it must not re-tick a `done` project already archived in `workshop/history/projects/`. Keep the milestone-mode detail-block skeleton refusal (`close.go:600-626`) per match. Retain the `--no-project` skip semantics.

**`--brain-dir` is NOT removed and is NOT a no-op.** It has a second, live consumer: `estimate.VelocityPath(f.BrainDir, "calibration-ledger.tsv")` (`close.go:758-762`) — the velocity/calibration ledger the issue deliberately keeps in brain (measurement, not coordination). This change removes only the flag's **project-discovery** use (`close.go:569`); the flag stays fully functional for the ledger. Update the flag's help text (`close.go:140`) to say it now locates only the calibration ledger, no longer project files. (The same `--brain-dir` on `milestone-close`, `actual`, `estimate-source`, and `project close` is untouched — those are all measurement paths.)

- [x] **Step 4: Update `applyClose` to write every edit**

`applyClose` (`close.go:696-706`) currently writes one project file. Loop the edit slice and write each (the peer-commit decision comes in M3 — for now, M2 writes all matched files in place, exactly as the single-file path did, and reports paths).

- [x] **Step 5: Run the tests to verify they pass**

Run: `go test ./cmd/sdlc/ -run TestClose -v`
Expected: PASS (single-match cases unchanged; the new all-match case green).

- [x] **Step 6: Commit**

```bash
git add cmd/sdlc/resolve.go cmd/sdlc/close.go cmd/sdlc/internal/project/discover.go cmd/sdlc/internal/project/discover_test.go cmd/sdlc/close_test.go cmd/sdlc/resolve_test.go
git commit -m "#171 M2: cross-repo project discovery + all-match close update"
```

- [x] **Step 7: Milestone-close**

Run: `sdlc milestone-close --issue 171 --milestone M2`
Fix Critical/Important; log the verdict.

---

## Chunk 3: M3 — Safe peer-write commit mechanics

### Task 3.1: The pure planner

**Files:**
- Create: `cmd/sdlc/peerwrite.go`
- Test: `cmd/sdlc/peerwrite_test.go`

- [ ] **Step 1: Write the failing test (the decision table)**

```go
func TestPlanPeerWrites(t *testing.T) {
	cur := "/fleet/ariadne"
	edits := map[string][]string{
		"/fleet/ariadne": {"workshop/projects/a.md"},   // current repo → deferred to normal close
		"/fleet/nous":    {"workshop/projects/b.md"},    // peer on main, clean index → commit
		"/fleet/metis":   {"workshop/projects/c.md"},    // peer off-main → report-only
		"/fleet/kbench":   {"workshop/projects/d.md"},   // peer on main, staged changes → report-only
	}
	states := map[string]RepoGitState{
		"/fleet/nous":  {Branch: "main", HasStagedChanges: false},
		"/fleet/metis": {Branch: "feature-x", HasStagedChanges: false},
		"/fleet/kbench": {Branch: "main", HasStagedChanges: true},
	}
	got := planPeerWrites(edits, states, cur)
	// assert: nous → Commit true, files=[b.md]; metis → Commit false, reason mentions branch;
	// kbench → Commit false, reason mentions staged; ariadne → not in the peer decisions
	// (current repo handled by the normal close commit).
}
```

- [ ] **Step 2: Run it to verify it fails**

Run: `go test ./cmd/sdlc/ -run TestPlanPeerWrites -v`
Expected: FAIL — undefined.

- [ ] **Step 3: Implement the planner**

```go
type RepoGitState struct {
	Branch           string
	HasStagedChanges bool
}

type PeerWriteDecision struct {
	RepoDir    string
	Files      []string
	Commit     bool
	Reason     string // why not committed (empty when Commit)
	NextAction string // exact manual command to finish (empty when Commit)
}

// planPeerWrites decides, per peer repo, whether git state authorizes an
// automatic scoped commit of the project-file edits. The current repo is
// omitted — its project edit rides the normal close commit. A peer commits
// only on main with a clean index; otherwise the file is left updated but
// uncommitted and the operator gets the reason + exact next action. The
// issue close succeeds regardless (caller never fails on a report-only).
func planPeerWrites(edits map[string][]string, states map[string]RepoGitState, curRepoDir string) []PeerWriteDecision {
	var out []PeerWriteDecision
	for repoDir, files := range edits {
		if repoDir == curRepoDir {
			continue
		}
		st := states[repoDir]
		d := PeerWriteDecision{RepoDir: repoDir, Files: files}
		switch {
		case st.Branch != "main":
			d.Reason = fmt.Sprintf("%s is on branch %q, not main", filepath.Base(repoDir), st.Branch)
			d.NextAction = fmt.Sprintf("cd %s && git add %s && git commit -m 'project: close-time update (ariadne#171)'", repoDir, strings.Join(files, " "))
		case st.HasStagedChanges:
			d.Reason = fmt.Sprintf("%s has pre-existing staged changes — refusing to absorb another session's index", filepath.Base(repoDir))
			d.NextAction = fmt.Sprintf("cd %s && git add %s && git commit  # after handling your staged changes", repoDir, strings.Join(files, " "))
		default:
			d.Commit = true
		}
		out = append(out, d)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].RepoDir < out[j].RepoDir })
	return out
}
```

- [ ] **Step 4: Run the test to verify it passes**

Run: `go test ./cmd/sdlc/ -run TestPlanPeerWrites -v`
Expected: PASS. Add cases: repo with edits but no state entry (treat as unknown → report-only, not commit); empty edits → empty decisions.

### Task 3.2: The thin git shell (state read + scoped commit)

**Files:**
- Modify: `cmd/sdlc/peerwrite.go` (add `readRepoGitState` + `applyPeerWrites`)
- Modify: `cmd/sdlc/close.go` (`applyClose` calls the shell)
- Test: `cmd/sdlc/close_finalize_test.go` or a new `peerwrite_apply_test.go` with a real multi-repo git fixture

- [ ] **Step 1: Write the failing process-level test**

Build a temp parent dir with two real git repos (init, commit an initial project file) — reuse the `hermeticRepo`/sibling-fixture pattern from `resolve_test.go`/`migrate_test.go`. Peer on `main` clean → after `applyPeerWrites`, assert `git -C <peer> log -1` shows a new commit touching exactly the project file and the working tree is clean. Peer with a staged unrelated change → assert the project file is written but **not** committed and the staged change is untouched.

- [ ] **Step 2: Run it to verify it fails**

Run: `go test ./cmd/sdlc/ -run TestApplyPeerWrites -v`
Expected: FAIL.

- [ ] **Step 3: Implement the shell**

```go
func readRepoGitState(r gitRunner, repoDir string) RepoGitState {
	branch, _ := r.GitInDir(repoDir, "rev-parse", "--abbrev-ref", "HEAD")
	// `diff --cached --quiet` exits non-zero iff there are staged changes.
	_, err := r.GitInDir(repoDir, "diff", "--cached", "--quiet")
	return RepoGitState{Branch: strings.TrimSpace(branch), HasStagedChanges: err != nil}
}

// applyPeerWrites commits each committing decision (staging exactly its
// Files) and reports each report-only decision. File writes already happened
// in applyClose, so this takes only the decisions — never fails the close.
func applyPeerWrites(r gitRunner, decisions []PeerWriteDecision, stdout, stderr io.Writer) {
	// write all files first (idempotent), then commit the committing ones
	for _, d := range decisions {
		if d.Commit {
			args := append([]string{"add"}, d.Files...)
			if _, err := r.GitInDir(d.RepoDir, args...); err != nil { /* warn, continue */ }
			if _, err := r.GitInDir(d.RepoDir, "commit", "-m", "project: close-time update (ariadne#171)"); err != nil { /* warn */ }
			cinfo(stdout, fmt.Sprintf("committed %s in %s", strings.Join(d.Files, ", "), filepath.Base(d.RepoDir)))
		} else {
			cwarn(stderr, fmt.Sprintf("project updated but NOT committed in %s: %s\n  finish: %s", filepath.Base(d.RepoDir), d.Reason, d.NextAction))
		}
	}
}
```

(File writes happen in `applyClose` as in M2; `applyPeerWrites` only handles the commit decision. `diff --cached --quiet` semantics: exit 0 = no staged changes; the `gitRunner` impl must surface the non-zero exit as an error — verify `execGitRunner` does, and if it swallows exit codes, read `git status --porcelain` for staged entries instead. Add a unit test pinning the staged-detection path.)

- [ ] **Step 4: Introduce a gitRunner into the close path and change `applyClose`'s signature**

The close path constructs **no** `gitRunner` today (unlike claim/changecode/merge/pr/push). Add `var closeRunner gitRunner = execGitRunner{}` at the close path's entry, and change `applyClose` from `(stderr io.Writer, f *closeFlags, r closeResult)` to `(stdout, stderr io.Writer, r gitRunner, f *closeFlags, res closeResult)` — it needs the runner for peer commits and a `stdout` writer for the "committed X in Y" report (it currently gets only `stderr`). Thread the new args through every caller: `reviewThenFinalize` → `finalizeBoundaryReview`, `runClose`, and the `--no-judge` branch in `runCloseWithReview` (grep `applyClose(` for the full call set). Then, after writing the local issue + current-repo project file, build `edits map[string][]string` from the M2 edit slice grouped by `RepoDir`, read `RepoGitState` per peer, run `planPeerWrites`, then `applyPeerWrites`. The current repo's project file (if any) commits with the issue as today.

- [ ] **Step 5: Run all close + peerwrite tests**

Run: `go test ./cmd/sdlc/ -run 'Close|PeerWrite' -v`
Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add cmd/sdlc/peerwrite.go cmd/sdlc/peerwrite_test.go cmd/sdlc/close.go cmd/sdlc/close_finalize_test.go
git commit -m "#171 M3: safe peer-write commit mechanics — scoped commit on main+clean, report-only otherwise"
```

- [ ] **Step 7: Milestone-close**

Run: `sdlc milestone-close --issue 171 --milestone M3`
This is the risky git-mutation milestone — read the fresh review carefully; fix Critical/Important; log the verdict.

---

## Chunk 4: M4 — Fleet navigation (resolve + parley)

### Task 4.1: `sdlc project find`

**Files:**
- Modify: `cmd/sdlc/project.go` (register `find` subcommand)
- Test: `cmd/sdlc/project_crud_test.go` or a new `projectfind_test.go`

- [ ] **Step 1: Write the failing test** — `sdlc project find --issue metis#18` over a temp fleet prints the matching project paths (one per line), legacy matches flagged.
- [ ] **Step 2: Run to verify it fails.** `go test ./cmd/sdlc/ -run TestProjectFind -v`
- [ ] **Step 3: Implement** — parse `--issue <repo>#<id>` via `parseRef` (`resolve.go:56`), call `project.DiscoverByIssueRef(filepath.Dir(repoTop), ref.Repo, ref.ID, project.ActiveAndArchive)` (navigation includes archived records), print `m.Path` (+ ` (legacy)` when `m.Legacy`). Add help text under `cmd/sdlc/helptext/`.
- [ ] **Step 4: Run to verify it passes.**

### Task 4.2: `sdlc resolve` project kind

**Files:**
- Modify: `cmd/sdlc/resolve.go`
- Test: `cmd/sdlc/resolve_test.go`

- [ ] **Step 1: Write the failing test** — resolving a `project` ref (or `--kind project`) returns project records fleet-wide via `DiscoverByIssueRef(..., project.ActiveAndArchive)` (archived records resolve too).
- [ ] **Step 2–4:** implement the smallest wiring that routes a project-kind resolution through the shared discovery with `ActiveAndArchive` scope; keep issue/other-kind resolution unchanged. Run `go test ./cmd/sdlc/ -run Resolve -v`.

### Task 4.3: parley `project` artifact class

**Files:**
- Modify: `parley.nvim` super-repo search config (find the artifact-class registry; grep for where `issue`/`plan` classes are declared)
- Test: parley's own test harness if present; else a manual step in this plan's Manual Verification

- [ ] **Step 1:** Read parley.nvim's super-repo search to locate how artifact classes are declared and how a class maps to a search root.
- [ ] **Step 2:** Add `project` as an always-cross-repo class: its search shells out to `sdlc project find` / `sdlc resolve` (or globs every peer's `workshop/projects/`), so a project jump works regardless of which repo holds the file.
- [ ] **Step 3:** Verify in parley (Manual Verification below).
- [ ] **Step 4: Commit** (parley.nvim is a peer repo — commit there per its own conventions; note the cross-repo edit in this issue's `## Log` per AGENTS §Peer Repo).

- [ ] **Step 5: Milestone-close**

Run: `sdlc milestone-close --issue 171 --milestone M4`

---

## Chunk 5: M5 — Residency documentation + base-layer ripple

**Files:**
- Modify: `AGENTS.base.md:62` ("Project files are usually in `brain`") and `AGENTS.base.md:20` ("brain = special peer holding cross-cutting state (`project`, `roadmap`)")
- Modify: `atlas/` (the residency/workflow map file; `grep -rl "data/project\|project.*brain" atlas/`)
- Modify: `construct/datatype/project.md` (residency + move procedure section, if not already covered)
- Regenerate: woven `AGENTS.md`/`CLAUDE.md`/`GEMINI.md` via `sdlc propagate-base`

- [ ] **Step 1:** Rewrite `AGENTS.base.md:62` to state projects live in each project's center-of-gravity repo under `workshop/projects/` (default: top product; movable via `sdlc migrate`; refs are `repo#id`), archived under `workshop/history/projects/`. Update `:20` so brain is described as capture/measurement only — no SDLC process artifacts.
- [ ] **Step 2:** Update the atlas residency/workflow map to the new home and the cross-peer close-gate discovery; keep `atlas/index.md` linking every file.
- [ ] **Step 3:** Ensure `construct/datatype/project.md` documents the residency default + move procedure (center of gravity; ref-rewrite on move via `sdlc migrate`).
- [ ] **Step 4:** Run `sdlc propagate-base` to re-weave downstream constitution files; run `go test ./cmd/sdlc/ -run Propagate` if present.
- [ ] **Step 5: Commit**

```bash
git add AGENTS.base.md atlas/ construct/datatype/project.md AGENTS.md CLAUDE.md GEMINI.md
git commit -m "#171 M5: project residency docs — coding-repo workshop/projects, brain holds no SDLC artifacts"
```

- [ ] **Step 6: Milestone-close** — `sdlc milestone-close --issue 171 --milestone M5`

---

## Chunk 6: M6 — Migrate the four terminal legacy records

**Data operation (one-time), not code.** All four are `status: done`; each converts to the current schema (validating under M1's relaxed guard) and lands in its center-of-gravity repo's `workshop/history/projects/`. Refs inside are already qualified peer refs; keep them fully qualified (unambiguous from any vantage) rather than localizing — these are frozen historical records. `metis-v2-experiment-algebra` stays in brain (active, legacy `status: active`) and is explicitly NOT migrated here — see the issue Revision reconciling this with the Done-when's "ends empty" (its migration is deferred until it closes, so it isn't relocated to `history/projects` mid-flight).

**Why manual, not `sdlc migrate` (#179):** #179 was filed as the relocation tool, so the divergence is deliberate. `migrate`'s core job is rewriting refs (dest-qualified → bare; `migrate.go rewriteRefs`), which would *localize* exactly the qualified refs these frozen records should keep, and its dest-vantage verification (`migrate.go:300`) fails closed if any historical ref no longer resolves (a real risk for old records citing since-archived/renumbered issues). A plain move preserves the record verbatim.

Destinations (operator-confirmed 2026-07-17):

| Record | Destination |
|---|---|
| `charon-launch-push` | `nous/workshop/history/projects/` |
| `shared-brain` | `nous/workshop/history/projects/` |
| `kaggle-ml-base-layer` | `kbench/workshop/history/projects/` |
| `metis-v1` | `metis/workshop/history/projects/` |

For each record:

- [ ] **Step 1: Create the destination dir** — `mkdir -p <dest>/workshop/history/projects` (none exist yet).
- [ ] **Step 2: Copy + schema-convert** — copy the file in; confirm frontmatter conforms to `#Project` under the relaxed guard. Add `closed:` (ISO) if absent, using the record's own stated close date from its body/Log. Do NOT fabricate `deadline`/`planned_finish` (M1 makes them optional for `done`). Leave `mvp_scope`/`explicitly_out` refs qualified.
- [ ] **Step 3: Validate** — `sdlc project validate <dest>/workshop/history/projects/<name>.md` (or `bin/vocabulary`/the conformance check the project verbs use). Expected: valid.
- [ ] **Step 4: Commit in the destination peer** — from the dest repo, `git add workshop/history/projects/<name>.md && git commit -m "history: archive <name> project (migrated from brain, ariadne#171)"`. (These are ariadne-styled peers on `main`; if a dest is off-main, follow the same report-and-hand-off discipline as the close gate.)
- [ ] **Step 5: Remove from brain** — `git rm data/project/<name>.md` in brain (nous auto-commits brain). Confirm the four are gone and only `metis-v2-experiment-algebra.md` remains under `brain/data/project/`.

- [ ] **Step 6: Verify discovery finds the migrated records** — from a dest repo, `sdlc project find --issue <a-ref-from-the-record>` returns the archived path. This exercises the `ActiveAndArchive` scope end-to-end (the record now lives in `workshop/history/projects/`, which only the archive-inclusive scope scans — the reason M2's `DiscoverByIssueRef` took a scope parameter). Note: closing already-closed issues won't re-tick; this is a read-path verification.

- [ ] **Step 7: Milestone-close** — `sdlc milestone-close --issue 171 --milestone M6`

---

## Manual Verification

Automated tests cover discovery, the planner, and the multi-repo commit shell. These require a human/live-fleet check:

1. **parley project jump (M4):** In parley super-repo mode, search/jump to a `project` and confirm it resolves a record in a *different* repo than the current one.
2. **Off-main report path (M3):** Check out a feature branch in a peer that owns a project referencing an open issue, run `sdlc close` on that issue from its repo, and confirm: the issue closes, the peer's project file is updated but uncommitted, and the operator sees the branch reason + exact `cd … && git add … && git commit` next action.
3. **On-main auto-commit (M3):** Same but peer on `main`, clean index → the project file is committed with a scoped commit touching only that file; a pre-existing *unstaged* unrelated change in the peer is left alone; a pre-existing *staged* change flips it to report-only.
4. **Legacy brain warning (M2):** Close an issue referenced by the still-resident `metis-v2` (brain) and confirm the deprecation warning fires while the tick still applies.

## Issue close (after all milestones)

`sdlc close --issue 171 --verified '<evidence: test names + manual-verification results + migration commits>'`. Let close measure + adopt `--actual` (do not hand-type). Update the atlas at close if any surface changed since M5. The mandatory fresh review runs at this boundary.

## Revisions

### 2026-07-17 — M1 executed; two corrections + a discovery refinement

Milestone M1 shipped (commits `4c6922a` guard + `50c9e3f` side-quest test fix;
boundary review verdict FIX-THEN-SHIP, no Critical, one Important resolved).
Corrections folded back into the plan:

1. **`pkg/vocab/project.json` regeneration is a byte-identical no-op.** `#Project`
   is a CUE `#`-definition and doesn't `cue export`, so relaxing its guard does
   not change the exported concrete blocks. Task 1.2 Step 1's expected diff does
   not exist. The real regeneration deliverable at a cue-touching boundary is
   **`make weave`** — it rewrites the `construct/generated/vocabulary/.source-sha`
   stamp (sha256 over raw `.cue` text; invalidated by any edit). The plan
   omitted it; now added explicitly. (`construct/generated/` is gitignored →
   no committable diff, but `vocabulary check` goes STALE without it. This is
   the third recurrence of the class — a follow-up to wire `vocabulary check`
   into the push/merge gate is warranted, tracked separately, not M1 scope.)
2. **Fixtures landed durably.** Task 1.1 Step 1's mktemp sketch was superseded by
   committed `construct/vocabulary/testdata/project_{done,executing}_no_baseline.json`
   with a negative control (executing-without-baseline must still be rejected).

Discovery refinement (from the change-code plan-quality finding, applied in M2's
spec): under `ActiveOnly`, `DiscoverByIssueRef` drops terminal-status
brain-legacy matches (`dropTerminalLegacy`) so the close gate can't re-tick a
`done` legacy record during the M2→M6 window.
