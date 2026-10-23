---
collaboration-mode: discussion
---

# Issue filename cap, path-length refusal and rename Implementation Plan

> **For agentic workers:** Consult AGENTS.md Section 3 (Subagent Strategy) to determine the appropriate execution approach: use superpowers-subagent-driven-development (if subagents are suitable per AGENTS.md) or superpowers-executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Issue filing can no longer publish a basename a checkout cannot hold, and an existing over-long (or otherwise
wrong) issue name can be fixed with a supported verb instead of plumbing commits.

**Architecture:** One pure slug builder (`issue.FileSlug`) caps every generated or operator-given slug at 80 bytes on a
hyphen boundary; both `issue new` paths (tracker and legacy) use it, and `flake register` reaches it through
`runIssueNew`. One pure path check (`gitx.CheckPathLimits`) refuses any written path with a component over 255 bytes
inside `TrunkFile.commitSet`, the single function that builds every tracker and main candidate tree, so the refusal
precedes any push. `sdlc issue rename` moves the card on `issue-tracker` and the details on `main` with two
idempotent compare-and-swap updates whose next step is a pure function of what both refs hold.

**Tech Stack:** Go, cobra, git plumbing through `cmd/sdlc/internal/gitx`.

---

## Core concepts

### Pure entities

| Name | Lives in | Status |
|------|----------|--------|
| `FileSlug` | `cmd/sdlc/internal/issue/scaffold.go` | new |
| `MaxSlugBytes` | `cmd/sdlc/internal/issue/scaffold.go` | new |
| `CheckPathLimits` | `cmd/sdlc/internal/gitx/pathlimit.go` | new |
| `RenameStep` / `NextRenameStep` | `cmd/sdlc/internal/tracker/rename.go` | new |
| `DetailNames` | `cmd/sdlc/internal/tracker/rename.go` | new |

- **FileSlug(title, explicit string) (string, error)** — the slug for an issue filename. `explicit != ""` is validated
  (must equal `Slugify(explicit)`, non-empty, at most `MaxSlugBytes`; refuses otherwise, naming the limit); else
  `Slugify(title)` truncated to the last `-` at or before byte 80 (hard cut at 80 when the first word is longer), trailing
  hyphens trimmed. Slugify output is ASCII, so bytes equal runes.
  - **Relationships:** consumed by `runIssueNew` (`cmd/sdlc/issue.go:401`), legacy `issue new`
    (`cmd/sdlc/legacymode.go:90`) and `issue rename`. `Slugify` itself is unchanged: `gatestate.NormalizeFamily` uses it
    and its family slugs must not change.
  - **DRY rationale:** the two copies of "explicit slug else Slugify, refuse empty" in `issue.go` and `legacymode.go`
    collapse into this one function (ARCH-DRY).
  - **Why 80:** the longest derived sibling is `<stem>-close-review.md` (+16), so the longest name the system derives is
    6+1+80+16 = 103 bytes, and the branch name (`issue.BranchName` = stem) is 87. Margin covers `.lock` and the
    255-byte macOS/Linux component limit with room for any future suffix.
- **CheckPathLimits(paths []string) error** — refuses the first path (sorted) with a `/`-separated component over 255
  bytes, naming the path, the component length and the limit. Pure; no filesystem probe (the limit is the common
  denominator of APFS/ext4, not the local filesystem's, because the pushed tree is checked out elsewhere).
- **NextRenameStep(obs RenameObservation) (RenameStep, error)** — the rename state machine (ARCH-ORDER). Input: for
  each ref (tracker, main), whether the old path and the new path are present. Output: `RenameCard`, `RenameDetails`,
  `RenameDone`, or a refusal. Table in "State design" below; unit-tested over all 16 combinations.

### Integration points

| Name | Lives in | Status | Wraps |
|------|----------|--------|-------|
| `TrunkFile.commitSet` | `cmd/sdlc/internal/gitx/updatemany.go:188` | modified | `git hash-object/update-index/write-tree` |
| `newIssueRenameCmd` | `cmd/sdlc/issuerename.go` | new | tracker + main `TrunkFile` updates |

- **commitSet** — calls `CheckPathLimits(keys(set.Write))` before any git call. Every candidate on either ref passes
  through it: `UpdateMany`/`UpdateManyPrepared` (`updatemany.go:121`) and `PrepareCandidate` (`candidate.go:71`), i.e.
  creation's card reserve and details publication, `issue publish`, card setters, transfers and the new rename.
  `PublishCommitChecked` (`commitpublication.go`) merges a selected commit's tree through `merge-tree`, bypassing
  commitSet, so it runs the same check over `git diff --name-only base tree` before its candidate (close round 1).
  Not guarded, and why: `closetracker.go:97,263,276` re-index paths already on the issue branch (they cannot introduce a
  new name); `refbootstrap.go:166` writes the fixed tracker bootstrap files. Deletes need no check.
- **issue rename** — reads the tracker card by id (`openTracker`, the same reader `set-title` uses), checks ownership
  with the rule `issue publish` uses (the owner, or anyone while unclaimed), then applies the steps from
  `NextRenameStep` until `RenameDone`, re-observing both refs after each step. Each step is one `TrunkFile.UpdateMany`
  whose prepare callback writes the unchanged bytes at the new path and deletes the old (CAS on the fetched base, the
  existing retry loop). Tests use the existing `newTrackerRepo` fixture (`cmd/sdlc/trackerenv_test.go:25`): real git
  against a local bare origin, the stateful fake already shared by the tracker tests (ARCH-MOCK; no new external
  surface).

## Scope decisions (TL to confirm)

1. **Rename only unstarted issues** (`status: open`, no issue branch locally or on origin, no `workshop/plans/<stem>-*`
   on main). The issue branch name, plan, gate ledger and review sidecars are all derived from the stem; renaming a
   started issue would have to move a pushed branch and its open PR. The incident class (a filer-produced name) is
   always caught before `start-plan`. A started issue refuses with: "rename before start-plan; #N is <status>".
2. **`--slug` optional:** omitted, rename derives `FileSlug(card title, "")`, the one-command fix for a #421-style name.
3. **Explicit `--slug` over 80 refuses** on both `issue new` and `rename` rather than being silently truncated.
4. **Existing long names are not swept.** Today's names under 255 (e.g. #562's 130-byte basename) still work; the
   push refusal is at 255, the generation cap at 80.

## State design

Actions: `sdlc issue rename --issue N [--slug S]` (also its rerun). No cancel: each step is one atomic ref update.

| Internal state (tracker old/new, main old/new) | Visible state | Recovery action | Possible loss |
|---|---|---|---|
| old/–, old/– | not renamed | rename | none |
| –/new, old/– (process died after card step) | half-renamed: `claim` reports details not on main for the new name; ID lint flags the slug mismatch | rerun rename (→ RenameDetails) | none; bytes unchanged |
| old/–, –/new (main stepped by an older partial run in reverse, or a hand fix) | half-renamed | rerun rename (→ RenameCard) | none |
| –/new, –/new | renamed | none (`RenameDone`, rerun is a no-op) | none |
| old and new both present on a ref, or neither old nor new on a ref | corrupt | refuse, naming both paths; operator repairs | none written |
| unknown push outcome (network) | `UpdateMany`'s existing offline error | rerun rename: the fresh fetch decides | none |
| card changed between the checks and the card move (a claim, start-plan) | refused: "changed since it was checked"; nothing moved | rerun rename (it checks the new card: a started issue refuses) | none |
| card changed after the card step (a claim on the new name) | half-renamed until rerun | rerun rename: the details step completes what the card names | none |

Covered: every observable combination of the two refs, process death between steps, reruns and a concurrent card
writer. Not covered: renaming started issues (scope decision 1), renaming archived issues in `workshop/history/`.

ARCH-SECURE: the card and detail bytes are copied, never parsed and rewritten, so the mirror OID stays valid; the new
slug is validated by `FileSlug` before any path is built (no `/`, `..` or uppercase can reach a tree). ARCH-FUNERAL:
creates nothing durable beyond the renamed files; the old paths are deleted in the same commit. ARCH-EVOLVE: no
persisted format changes. ARCH-CONSTRAINTS: interactive CLI, two pushes, N/A otherwise. ARCH-PURPOSE: all three
filers (`issue new` tracker/legacy, `flake register`) derive from `FileSlug`; every candidate writer derives from the
one guard.

## Tasks

### Task 1: FileSlug

**Files:** Modify `cmd/sdlc/internal/issue/scaffold.go`; Test `cmd/sdlc/internal/issue/scaffold_test.go`

- [ ] **Step 1: failing tests**

```go
func TestFileSlugCapsGeneratedSlugAtWordBoundary(t *testing.T) {
	title := strings.Repeat("parley lua test name that is long ", 9) // ~300 chars
	got, err := FileSlug(title, "")
	if err != nil { t.Fatal(err) }
	if len(got) > MaxSlugBytes || strings.HasSuffix(got, "-") || !strings.HasPrefix(Slugify(title), got) {
		t.Fatalf("FileSlug = %q (%d bytes)", got, len(got))
	}
	if next := Slugify(title)[len(got)]; next != '-' { t.Fatalf("cut mid-word at %q", got) }
}

func TestFileSlugHardCutsOneLongWord(t *testing.T) {
	got, _ := FileSlug(strings.Repeat("x", 300), "")
	if got != strings.Repeat("x", MaxSlugBytes) { t.Fatalf("got %d bytes", len(got)) }
}

func TestFileSlugExplicit(t *testing.T) {
	for _, bad := range []string{strings.Repeat("a", MaxSlugBytes+1), "Upper", "a/b", "../x", "-a"} {
		if _, err := FileSlug("t", bad); err == nil { t.Errorf("%q accepted", bad) }
	}
	if got, err := FileSlug("ignored", "short-name"); err != nil || got != "short-name" { t.Fatal(got, err) }
	if _, err := FileSlug("!!!", ""); err == nil { t.Fatal("empty slug accepted") }
}

func TestFileSlugLeavesShortTitlesAsSlugify(t *testing.T) {
	if got, _ := FileSlug("Fix the thing", ""); got != "fix-the-thing" { t.Fatal(got) }
}
```

- [ ] **Step 2:** `go test ./cmd/sdlc/internal/issue -run TestFileSlug` → FAIL (undefined).
- [ ] **Step 3: implement**

```go
// MaxSlugBytes caps an issue filename slug (#566). The longest derived sibling,
// <stem>-close-review.md, then stays near 100 bytes, far inside the 255-byte
// component limit every checkout's filesystem enforces.
const MaxSlugBytes = 80

// FileSlug is the slug an issue's files are named with: the operator's explicit
// slug, validated, or the title's Slugify cut at the last word boundary within
// MaxSlugBytes. Every filer uses it (#566).
func FileSlug(title, explicit string) (string, error) {
	if explicit != "" {
		if explicit != Slugify(explicit) || len(explicit) > MaxSlugBytes {
			return "", fmt.Errorf("slug %q must be lowercase words joined by hyphens, at most %d bytes", explicit, MaxSlugBytes)
		}
		return explicit, nil
	}
	s := Slugify(title)
	if len(s) > MaxSlugBytes {
		cut := s[:MaxSlugBytes]
		if s[MaxSlugBytes] != '-' {
			if i := strings.LastIndexByte(cut, '-'); i > 0 {
				cut = cut[:i]
			}
		}
		s = strings.TrimRight(cut, "-")
	}
	if s == "" {
		return "", fmt.Errorf("title %q produced an empty slug; pass --slug", title)
	}
	return s, nil
}
```

- [ ] **Step 4:** rerun → PASS.
- [ ] **Step 5: wire both filers.** Replace `cmd/sdlc/issue.go:401-407` with `slug, err := issue.FileSlug(title, f.Slug); if err != nil { return err }`, and the same in `cmd/sdlc/legacymode.go:90-96` with `die(stderr, err.Error())`.
- [ ] **Step 6: filer test** in `cmd/sdlc/issue_test.go`: on a `newTrackerRepo`, run `issue new` with a 300-char title; assert the card path on `issue-tracker` and the details path on `main` share a basename of at most 6+1+80+3 bytes. Same through `flake register`'s `fileFlakeIssue` with two long test names (the #421 shape), asserting the printed path. Legacy: one case in `legacymode_test.go`.
- [ ] **Step 7:** `go test ./cmd/sdlc -run 'TestIssueNew|TestFlake|TestLegacy'` PASS; commit `#566: cap issue filename slugs in one builder`.

### Task 2: path-length refusal before push

**Files:** Create `cmd/sdlc/internal/gitx/pathlimit.go`, `pathlimit_test.go`; Modify `cmd/sdlc/internal/gitx/updatemany.go:188`

- [ ] **Step 1: failing tests**

```go
func TestCheckPathLimits(t *testing.T) {
	ok := "workshop/issues/000001-" + strings.Repeat("a", 240) + ".md" // component 251
	if err := CheckPathLimits([]string{ok}); err != nil { t.Fatal(err) }
	long := "workshop/issues/000001-" + strings.Repeat("a", 300) + ".md"
	err := CheckPathLimits([]string{ok, long})
	if err == nil || !strings.Contains(err.Error(), "310 bytes") || !strings.Contains(err.Error(), "255") { t.Fatal(err) }
}
```

and in an existing `TrunkFile` test file, a `UpdateMany` whose prepare writes a 331-byte basename: assert the error,
and that `git ls-remote` shows the remote tip unchanged (nothing pushed).

- [ ] **Step 2:** `go test ./cmd/sdlc/internal/gitx -run 'TestCheckPathLimits|TestUpdateManyRefusesLongPath'` → FAIL.
- [ ] **Step 3: implement**

```go
// MaxPathComponentBytes is the file-name limit of APFS and ext4. A pushed tree
// is checked out on other machines, so the limit is the common one, not this
// filesystem's (#566: a 331-byte name broke every parley checkout).
const MaxPathComponentBytes = 255

// CheckPathLimits refuses a path that some checkout could not create.
func CheckPathLimits(paths []string) error {
	sorted := append([]string(nil), paths...)
	sort.Strings(sorted)
	for _, p := range sorted {
		for _, c := range strings.Split(p, "/") {
			if len(c) > MaxPathComponentBytes {
				return fmt.Errorf("refusing to write %s: its name %q is %d bytes, over the %d-byte file-name limit of a checkout", p, c, len(c), MaxPathComponentBytes)
			}
		}
	}
	return nil
}
```

First lines of `commitSet`: `if err := CheckPathLimits(slices.Collect(maps.Keys(set.Write))); err != nil { return "", err }`.

- [ ] **Step 4:** PASS; commit `#566: refuse over-long paths before any candidate is built`.

### Task 3: rename state machine

**Files:** Create `cmd/sdlc/internal/tracker/rename.go`, `rename_test.go`

- [ ] **Step 1: failing table test** over all 16 `{TrackerOld, TrackerNew, MainOld, MainNew bool}` combinations,
  asserting the State design table: (T,F,T,F)→RenameCard; (F,T,T,F)→RenameDetails; (T,F,F,T)→RenameCard;
  (F,T,F,T)→RenameDone; every other combination → error naming both paths.
- [ ] **Step 2:** FAIL; **Step 3:** implement

```go
type RenameObservation struct{ TrackerOld, TrackerNew, MainOld, MainNew bool }
type RenameStep int
const (RenameCard RenameStep = iota + 1; RenameDetails; RenameDone)

// NextRenameStep decides a rename's next ref update from what both refs hold
// now, so an interrupted rename resumes by running it again (#566).
func NextRenameStep(o RenameObservation) (RenameStep, error) {
	if o.TrackerOld == o.TrackerNew || o.MainOld == o.MainNew {
		return 0, errors.New("each ref must hold exactly one of the old and new names")
	}
	switch {
	case o.TrackerOld:
		return RenameCard, nil
	case o.MainOld:
		return RenameDetails, nil
	}
	return RenameDone, nil
}
```

  plus `MovePath(old, new string) func(*gitx.TrunkView) (gitx.TrunkWrite, error)`: reads `old` from the view (refusing
  when absent), returns `TrunkWrite{Write: {new: bytes}, Delete: {old}, ExactBytes: true}`.
- [ ] **Step 4:** PASS; commit.

### Task 4: `sdlc issue rename`

**Files:** Create `cmd/sdlc/issuerename.go`, `issuerename_test.go`; Modify `cmd/sdlc/issue.go:60` (register), the
`issue` help contract text, `atlas/workflow/sdlc-binary.md` (issue verbs).

- [ ] **Step 1: failing integration tests** on `newTrackerRepo`, each asserting both refs afterwards via `git ls-tree`:
  - rename `--slug short` of an open unclaimed issue: card and details at the new path, old absent, blobs byte-identical.
  - no `--slug`: a card whose title is 300 chars and whose files carry the full slug → renamed to `FileSlug(title)`.
  - resume: pre-apply only the card step by hand, rerun → details follow; rerun again → "already named …", exit 0, no commit.
  - refusals, each with both refs unchanged: status `working`; an issue branch on origin; another slot's claim;
    `--slug` over 80; new name already used by another file.
- [ ] **Step 2:** FAIL; **Step 3:** implement the command: resolve the card (`openTracker`), refuse per scope decision 1
  and the publish ownership rule, compute the new slug with `issue.FileSlug(card.Title, f.Slug)`, then loop:
  observe both refs (`TrunkView.Exists` on fresh tips), `NextRenameStep`, apply `UpdateMany(msg, tracker.MovePath(…))`
  on the tracker or main `TrunkFile`, with messages `#N: issue: rename card` / `#N: issue: rename details`. Print the
  new details path. Bound the loop at 3 iterations (two steps plus the Done observation).
- [ ] **Step 4:** PASS; update help/atlas; commit.

### Task 5: verification

- [ ] `go test ./cmd/sdlc/internal/issue ./cmd/sdlc/internal/gitx ./cmd/sdlc/internal/tracker` and
  `go test ./cmd/sdlc -run 'Issue|Flake|Legacy|Rename'`.
- [ ] Fresh-reader check of `sdlc issue rename --help`: can a reader with no context fix a #421-style name and resume
  a half-done rename?
- [ ] Recorded suite: `sdlc local-check run --name suite` on the pushed head, then `sdlc close`.

## Revisions

2026-10-23 (close round 1): Dropped the "no `workshop/plans/<stem>-*` on main" rename precondition; `status: open`
plus no local or remote issue branch covers the unstarted case. Added `tracker.DetailNames` (pure): after the card step
the card no longer names the old details file, so main's details are found by ID. The permission check (`renamable`)
now runs before any moving step, including the details step of a half-done rename (BR-1). `PublishCommitChecked` joins
the guarded writers. A legacy filer test pins the cap there (BR-2).

2026-10-23 (close round 2, BR-5): the card move pins the bytes its checks read (`MovePath` expect, `MoveCard` takes the Record); a change in between refuses and a rerun re-checks. The details step stays unpinned: ownership lives on the card, which already has the new name. State design rows corrected.
