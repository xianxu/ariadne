# `sdlc resolve` — read-only artifact-reference resolver — Implementation Plan

> **For agentic workers:** Consult AGENTS.md Section 3 (Subagent Strategy) to determine the appropriate execution approach: use superpowers-subagent-driven-development (if subagents are suitable per AGENTS.md) or superpowers-executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a read-only `sdlc resolve <ref>` (+ `sdlc open <ref>` sugar) that maps a symbolic artifact ref (`ariadne#11`, `#15 M4`, `pair#84`) to the current file path(s) of the issue and its plan/review family, correct after archiving and across sibling repos — deriving every artifact location from the vocabulary model, and single-sourcing the ref grammar as the sole parser so parley#160 and agents consume one spec.

**Architecture:** A **pure core** (a ref parser `parseRef` and a family classifier `classifyFamily`, both string→struct, no IO) sits behind a **thin IO shell** (sibling-repo directory resolution + a 3-directory glob) surfaced as two cobra commands. Locations (`workshop/issues`, `workshop/plans`, `workshop/history`) come from an extended `discovery:` block in `construct/vocabulary/issue.cue`, read through a new `pkg/vocab` accessor — so resolve hardcodes no artifact paths (ARCH-DRY / ARCH-PURPOSE). Because resolve only reads, it never takes `.git/sdlc.lock`, which is what makes it cheap enough for parley to shell to on a keypress.

**Tech Stack:** Go (cobra command tree in `cmd/sdlc`), CUE (`construct/vocabulary/issue.cue` → `pkg/vocab/issue.json` via `vocabulary export`), Go stdlib `filepath.Glob` / `os.ReadDir`.

---

## Design decisions (surfaced for operator review)

These are the judgment calls made from the #144 Spec; they shape the code below.

1. **Model extension over a new location registry (ARCH-DRY).** The issue's plan and boundary-review sidecars are the *issue's* family: they carry its 6-digit id and co-archive with it (`push.go archivePlanArtifacts`). So rather than invent a separate datatype-location registry, extend the already-consumed, already-exported `discovery:` block in `issue.cue` with two fields — `archive` (where the issue + its family move on close/merge) and `plans` (the active home of the durable plan + review sidecars). One source, already wired into `pkg/vocab` and JSON export.

2. **Grammar single-sourced *as the parser*, not as a spec doc (ARCH-DRY).** parley#160 does **not** re-implement the grammar in Lua; it hands the raw ref string to `sdlc resolve`, which owns the only parser. `sdlc resolve --help` documents the grammar for humans/agents, but the *authoritative* definition is `parseRef`. This is the DRY-correct reading of "single-sourced so parley + agents can't diverge": there is one parser, and everything delegates to it.

3. **Repo token → sibling dir by exact-basename-then-unique-prefix.** `pair#84` → `../pair` (exact); `parley#160` → `../parley.nvim` (unique prefix, since the token `parley` isn't a literal sibling dir); `ariadne#11` → `../ariadne` (exact). No alias registry: exact basename match wins; else a unique case-insensitive prefix match among siblings; ambiguity or no match is an error listing candidates.

4. **Family is the default output for a bare id; `Mx` narrows to the review sidecar.** `sdlc resolve #144` prints the whole family (issue → plan → reviews). `sdlc resolve #144 M2` prints the `*-m2-review.md` sidecar(s). `--json` always emits the structured object. This matches Done-when ("resolving by id returns the family") and parley's picker UX.

5. **GitHub refs are recognized but not locally resolved.** `gh#42` / `pair gh#42` classify as `kind=github` and echo the canonical ref; resolve stays read-only and offline (no `gh`, no network). parley/agents decide what to do with a github ref.

6. **Out of scope (noted so the ARCH-PURPOSE plan gate sees it was considered):** migrating the *existing* hardcoders of `workshop/plans` / `workshop/history` (`push.go`, `merge.go`, `close.go`, `state.go`) onto the new `Discovery()` accessor is a legitimate DRY follow-up that overlaps #163 (scanner consolidation) — it is **not** the point of #144. #144's purpose is the resolver + single-sourced grammar; that purpose is fully delivered by this plan (resolve itself derives from the model). A `## Log` note will point the migration at #163.

---

## Core concepts

### Pure entities (the conceptual core)

| Name | Lives in | Status |
|------|----------|--------|
| `Discovery` | `pkg/vocab/vocab.go` | modified |
| `ArtifactRef` | `cmd/sdlc/resolve.go` | new |
| `parseRef` | `cmd/sdlc/resolve.go` | new |
| `Artifact` / `artifactKind` | `cmd/sdlc/resolve.go` | new |
| `classifyFamily` | `cmd/sdlc/resolve.go` | new |

- **Discovery** — the parsed `discovery:` block of the issue noun: `Home`, `Glob`, `Archive`, `Plans`. Exposed by a new `(*IssueModel).Discovery()` accessor over a new `Discovery Discovery `json:"discovery"`` field on `IssueModel`. The `IssueModel` struct is *modified* (adds the field + accessor); the underlying data already exists in `issue.json` for `home`/`glob` and is added for `archive`/`plans`.
  - **Relationships:** 1:1 with the issue noun. Read-only; the single Go view of where issue-family artifacts live.
  - **DRY rationale:** Removes resolve's need to hardcode any of the three directories; the same accessor is the future migration target for `push`/`merge`/`state` (see Design decision 6, → #163).
  - **Future extensions:** if a fourth family member appears (e.g. a per-issue `notes/` sidecar), it's one field here, not a new lookup path in every consumer.

- **ArtifactRef** — the parsed symbolic ref: `Repo string` ("" = current repo), `ID int` (0 = none), `Milestone string` ("" = none, e.g. "M2"), `GitHub bool`.
  - **Relationships:** produced by `parseRef`, consumed by the resolve command's IO shell.
  - **DRY rationale:** first occurrence of a ref grammar that MUST NOT be re-encoded elsewhere (parley delegates to this parser).
  - **Future extensions:** a `Slug string` field if slug-keyed refs (targets by name) are added later.

- **parseRef** — `func parseRef(raw string) (ArtifactRef, error)`. Pure. Parses the grammar (below). No filesystem, no repo knowledge — a bare `#id` yields `Repo: ""`, and the caller substitutes the current repo.

- **Artifact / artifactKind** — one resolved family member: `Kind artifactKind` (`kindIssue` | `kindPlan` | `kindReview`), `Path string` (absolute), plus `Milestone string` for reviews (parsed from `-mX-review.md`; "" for the `-close-review.md`).
  - **DRY rationale:** the family classification (issue vs plan vs review, and which milestone) is defined once, in `classifyFamily`, not re-derived by each output path.

- **classifyFamily** — `func classifyFamily(id int, paths []string) []Artifact`. Pure: given the id and the union of matched file paths, classify each by filename suffix and return them ordered issue → plan → reviews (reviews sorted by milestone then name). The suffix rules: `-plan.md` → plan; `-m<N><letter?>-review.md` → review with that milestone; `-close-review.md` or any other `*-review.md` → review with milestone "". Everything else matching `NNNNNN-*.md` → issue.

### Integration points (where pure meets the world)

| Name | Lives in | Status | Wraps |
|------|----------|--------|-------|
| `resolveRepoDir` | `cmd/sdlc/resolve.go` | new | filesystem (git root + sibling scan) |
| `familyFiles` | `cmd/sdlc/resolve.go` | new | filesystem (glob 3 dirs) |
| `NewResolveCmd` / `NewOpenCmd` | `cmd/sdlc/resolve.go` | new | cobra + stdout + `$EDITOR` |

- **resolveRepoDir** — `func resolveRepoDir(ref ArtifactRef, curRoot string) (string, error)`. If `ref.Repo == ""`, returns `curRoot`. Else scans `filepath.Dir(curRoot)` for a sibling directory: exact basename match, else unique case-insensitive prefix match; ambiguous/none → error listing candidates. `curRoot` is injected (from `gitx.RepoTopLevel()`) so the sibling-scan logic is testable against a temp parent dir with fake sibling dirs.
  - **Injected into:** the command RunE; `curRoot` is the seam that keeps the sibling-match logic unit-testable.
  - **Future extensions:** honor an explicit override env (`WF_SIBLING_ROOT`) if repos ever nest differently.

- **familyFiles** — `func familyFiles(repoDir string, d vocab.Discovery, id int) ([]string, error)`. Globs `repoDir/{Home,Plans,Archive}/NNNNNN-*.md`, unions and de-dupes the matches. Pure-ish glue: directories come from the injected `Discovery`, so no path is hardcoded.
  - **Injected into:** the command RunE. `repoDir` + `d` injected → testable against a temp repo with seeded files.

- **NewResolveCmd / NewOpenCmd** — the cobra surface. `resolve` prints paths (or `--json`); `open` execs `$EDITOR` on the primary target. Registered in `buildRoot` (main.go); Long help from `helptext/resolve.md` + `helptext/open.md`.

---

## Ref grammar (authoritative; documented in `helptext/resolve.md`)

```
ref        := [repo-token] ("gh" WS)? "#" id-digits (WS milestone)?
            | [repo-token] "#" id-digits (WS milestone)?
repo-token := WORD attached directly to "#"  (e.g. "ariadne", "pair", "parley")
id-digits  := 1–6 digits (zero-padded to 6 for globbing; "#11" ⇒ 000011)
milestone  := "M" digits letter?            (e.g. "M4", "M4b")
gh marker  := a "gh" token immediately before "#" (bare `gh#42`, or `repo gh#42`)
```

Examples → parse:
| Input | Repo | ID | Milestone | GitHub |
|-------|------|----|-----------| -------|
| `#144` | "" | 144 | "" | false |
| `ariadne#11` | ariadne | 11 | "" | false |
| `#15 M4` | "" | 15 | M4 | false |
| `pair#84` | pair | 84 | "" | false |
| `parley#160 M2b` | parley | 160 | M2b | false |
| `gh#42` | "" | 42 | "" | true |
| `pair gh#42` | pair | 42 | "" | true |

Parsing strategy (pure, in `parseRef`): trim; require exactly one `#`; split into `left` (before `#`) and `right` (after). `right` = id-digits, optionally whitespace + milestone. `left` = optional repo token, with a trailing `gh` token (bare `gh` or ` gh`) marking a GitHub ref. Empty `left` ⇒ current repo, workshop ref.

---

## Chunk 1: Milestone M1 — model + pure core

Delivers the entire testable-without-a-command engine: the extended discovery model, its Go accessor, the ref parser, and the family classifier. All unit-tested, zero IO. Closes with `sdlc milestone-close --milestone M1`.

### Task 1.1: Extend the `discovery:` block in `issue.cue`

**Files:**
- Modify: `construct/vocabulary/issue.cue` (the `discovery:` block, ~line 42)

- [ ] **Step 1: Edit the discovery block** to add `archive` and `plans`, with comments:

```cue
discovery: {
	home:    "workshop/issues" // repo-relative home folder for issue instances
	glob:    "*.md"            // filename glob within home
	// archive: terminal issues AND their id-keyed plan/review family move here on
	// close/merge (ariadne#160). A resolver must search here to stay correct after
	// archiving. Repo-relative.
	archive: "workshop/history"
	// plans: active home of the issue's durable plan (NNNNNN-*-plan.md) and every
	// boundary-review sidecar (NNNNNN-*-mX-review.md / -close-review.md), same
	// 6-digit id. Co-archived to `archive` on close/merge (#136).
	plans: "workshop/plans"
}
```

- [ ] **Step 2: Vet the CUE** with the DAG-aware wrapper (verified: NOT bare `cue vet ./...`, which skips the leaf-wins merge `resolveVocab` performs — `cmd/vocabulary/cue.go`, `cmd/vocabulary/export.go` `runVet`):

Run: `make vocabulary-build && vocabulary vet`
Expected: PASS. `discovery` is an **open struct literal** (not a `#`-definition — confirmed `construct/vocabulary/issue.cue:42`), so the new `archive`/`plans` string fields vet and export cleanly.

- [ ] **Step 3: Regenerate the embedded JSON:**

Run: `make vocab-embed`
This builds the `vocabulary` binary, runs `go generate ./pkg/vocab/...` (regenerating **both** `issue.json` and `verdict.json`), then a `git diff --exit-code -- pkg/vocab` drift check.
Expected: on the first run after the model change it **reports drift on `issue.json`** — that report *is* the regeneration landing; stage + commit the regenerated `issue.json`, then re-run to confirm clean. `verdict.json` is unaffected.

- [ ] **Step 3b: Verify the JSON gained exactly the two keys:**

Run: `git diff pkg/vocab/issue.json`
Expected: only `"archive": "workshop/history"` + `"plans": "workshop/plans"` added under `discovery`.

- [ ] **Step 4: Commit**

```bash
git add construct/vocabulary/issue.cue pkg/vocab/issue.json
git commit -m "#144 M1: model the archive + plans homes in issue.cue discovery"
```

### Task 1.2: Add the `Discovery` accessor to `pkg/vocab`

**Files:**
- Modify: `pkg/vocab/vocab.go` (add type + field + accessor)
- Test: `pkg/vocab/vocab_test.go`

- [ ] **Step 1: Write the failing test** in `vocab_test.go`:

```go
func TestDiscovery(t *testing.T) {
	d := vocab.Issue().Discovery()
	if d.Home != "workshop/issues" || d.Glob != "*.md" {
		t.Fatalf("home/glob: got %+v", d)
	}
	if d.Archive != "workshop/history" {
		t.Fatalf("archive: got %q", d.Archive)
	}
	if d.Plans != "workshop/plans" {
		t.Fatalf("plans: got %q", d.Plans)
	}
}
```

- [ ] **Step 2: Run it — expect FAIL** (no `Discovery` method):

Run: `go test ./pkg/vocab/ -run TestDiscovery -v`
Expected: FAIL — `d.Discovery undefined`.

- [ ] **Step 3: Add the type, field, and accessor** in `vocab.go`:

```go
// Discovery is the parsed `discovery:` block: where instances of the issue noun
// and its id-keyed family (plan + review sidecars) live, and where they move on
// close/merge. Repo-relative; a consumer joins these to a repo root. Single source:
// construct/vocabulary/issue.cue.
type Discovery struct {
	Home    string `json:"home"`    // active issue instances
	Glob    string `json:"glob"`    // filename glob within Home
	Archive string `json:"archive"` // terminal issues + family move here
	Plans   string `json:"plans"`   // active durable plan + review sidecars
}
```

Add `Discovery Discovery `json:"discovery"`` to the `IssueModel` struct, and:

```go
// Discovery returns the issue noun's location model (home/glob/archive/plans).
func (m *IssueModel) Discovery() Discovery { return m.Discovery }
```

Note: the field and method are both named `Discovery`. If Go rejects the field/method name clash, rename the field `discovery` (unexported) and unmarshal via a custom step, OR name the field `Disc` with `json:"discovery"` and keep the accessor `Discovery()`. **Prefer** the `Disc`-field + `Discovery()`-method form to avoid the clash:

```go
type IssueModel struct {
	Categories map[string][]string `json:"categories"`
	When       map[string]string   `json:"when"`
	Disc       Discovery           `json:"discovery"`
	Lifecycle  []Transition        `json:"lifecycle"`
}
func (m *IssueModel) Discovery() Discovery { return m.Disc }
```

- [ ] **Step 4: Run the test — expect PASS:**

Run: `go test ./pkg/vocab/ -run TestDiscovery -v`
Expected: PASS.

- [ ] **Step 5: Run the vocab conformance test** (guards the embedded JSON):

Run: `go test ./pkg/vocab/...`
Expected: PASS (no regression in `conformance_test.go` / `vocab_test.go`).

- [ ] **Step 6: Commit**

```bash
git add pkg/vocab/vocab.go pkg/vocab/vocab_test.go
git commit -m "#144 M1: pkg/vocab Discovery() accessor over the discovery block"
```

### Task 1.3: Pure ref parser `parseRef`

**Files:**
- Create: `cmd/sdlc/resolve.go` (start the file with the pure types + parser)
- Test: `cmd/sdlc/resolve_test.go`

- [ ] **Step 1: Write the failing table test** in `resolve_test.go`:

```go
func TestParseRef(t *testing.T) {
	cases := []struct {
		in   string
		want ArtifactRef
		err  bool
	}{
		{"#144", ArtifactRef{ID: 144}, false},
		{"ariadne#11", ArtifactRef{Repo: "ariadne", ID: 11}, false},
		{"#15 M4", ArtifactRef{ID: 15, Milestone: "M4"}, false},
		{"pair#84", ArtifactRef{Repo: "pair", ID: 84}, false},
		{"parley#160 M2b", ArtifactRef{Repo: "parley", ID: 160, Milestone: "M2b"}, false},
		{"gh#42", ArtifactRef{ID: 42, GitHub: true}, false},
		{"pair gh#42", ArtifactRef{Repo: "pair", ID: 42, GitHub: true}, false},
		{"  ariadne#000011  ", ArtifactRef{Repo: "ariadne", ID: 11}, false},
		{"nope", ArtifactRef{}, true},       // no '#'
		{"#", ArtifactRef{}, true},          // no id
		{"#1234567", ArtifactRef{}, true},   // >6 digits
		{"a#1#2", ArtifactRef{}, true},      // two '#'
	}
	for _, c := range cases {
		got, err := parseRef(c.in)
		if (err != nil) != c.err {
			t.Fatalf("%q: err=%v want err=%v", c.in, err, c.err)
		}
		if err == nil && got != c.want {
			t.Fatalf("%q: got %+v want %+v", c.in, got, c.want)
		}
	}
}
```

- [ ] **Step 2: Run it — expect FAIL** (undefined `parseRef` / `ArtifactRef`):

Run: `go test ./cmd/sdlc/ -run TestParseRef -v`
Expected: FAIL to compile.

- [ ] **Step 3: Implement the pure types + `parseRef`** in `resolve.go`:

```go
package main

// ArtifactRef is a parsed symbolic artifact reference. Repo "" means the current
// repo. GitHub marks a GitHub-inbox ref (resolved to a label, not a local path).
type ArtifactRef struct {
	Repo      string
	ID        int
	Milestone string
	GitHub    bool
}

// parseRef parses the single-sourced ref grammar (see helptext/resolve.md). Pure.
func parseRef(raw string) (ArtifactRef, error) {
	s := strings.TrimSpace(raw)
	if strings.Count(s, "#") != 1 {
		return ArtifactRef{}, fmt.Errorf("ref %q: expected exactly one '#'", raw)
	}
	i := strings.IndexByte(s, '#')
	left, right := strings.TrimSpace(s[:i]), strings.TrimSpace(s[i+1:])

	// right: id-digits [WS milestone]
	rf := strings.Fields(right)
	if len(rf) == 0 {
		return ArtifactRef{}, fmt.Errorf("ref %q: missing id", raw)
	}
	if len(rf[0]) < 1 || len(rf[0]) > 6 {
		return ArtifactRef{}, fmt.Errorf("ref %q: id must be 1–6 digits", raw)
	}
	id, err := strconv.Atoi(rf[0])
	if err != nil || id <= 0 {
		return ArtifactRef{}, fmt.Errorf("ref %q: bad id %q", raw, rf[0])
	}
	ref := ArtifactRef{ID: id}
	if len(rf) >= 2 {
		if !milestoneRe.MatchString(rf[1]) {
			return ArtifactRef{}, fmt.Errorf("ref %q: bad milestone %q", raw, rf[1])
		}
		ref.Milestone = rf[1]
	}
	if len(rf) > 2 {
		return ArtifactRef{}, fmt.Errorf("ref %q: trailing tokens after milestone", raw)
	}

	// left: [repo] ["gh" marker]
	switch {
	case left == "gh":
		ref.GitHub = true
	case strings.HasSuffix(left, " gh"):
		ref.GitHub = true
		ref.Repo = strings.TrimSpace(strings.TrimSuffix(left, " gh"))
	default:
		ref.Repo = left
	}
	if strings.ContainsAny(ref.Repo, " \t") {
		return ArtifactRef{}, fmt.Errorf("ref %q: malformed repo token %q", raw, ref.Repo)
	}
	return ref, nil
}

var milestoneRe = regexp.MustCompile(`^M\d+[a-z]?$`)
```

Add imports: `fmt`, `regexp`, `strconv`, `strings`.

- [ ] **Step 4: Run the test — expect PASS:**

Run: `go test ./cmd/sdlc/ -run TestParseRef -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add cmd/sdlc/resolve.go cmd/sdlc/resolve_test.go
git commit -m "#144 M1: pure parseRef — the single-source ref grammar parser"
```

### Task 1.4: Pure family classifier `classifyFamily`

**Files:**
- Modify: `cmd/sdlc/resolve.go`
- Test: `cmd/sdlc/resolve_test.go`

- [ ] **Step 1: Write the failing test:**

```go
func TestClassifyFamily(t *testing.T) {
	paths := []string{
		"/r/workshop/plans/000144-foo-m2-review.md",
		"/r/workshop/issues/000144-foo.md",
		"/r/workshop/plans/000144-foo-plan.md",
		"/r/workshop/plans/000144-foo-close-review.md",
		"/r/workshop/plans/000144-foo-m1-review.md",
		"/r/workshop/plans/000999-other.md", // wrong id — ignored
	}
	got := classifyFamily(144, paths)
	// Ordered: issue, plan, then reviews (M1, M2, close).
	wantKinds := []artifactKind{kindIssue, kindPlan, kindReview, kindReview, kindReview}
	if len(got) != len(wantKinds) {
		t.Fatalf("len=%d want %d: %+v", len(got), len(wantKinds), got)
	}
	for i, k := range wantKinds {
		if got[i].Kind != k {
			t.Fatalf("pos %d: kind=%v want %v", i, got[i].Kind, k)
		}
	}
	if got[2].Milestone != "M1" || got[3].Milestone != "M2" || got[4].Milestone != "" {
		t.Fatalf("review milestones: %+v", got[2:])
	}
}
```

- [ ] **Step 2: Run — expect FAIL** (undefined `classifyFamily`):

Run: `go test ./cmd/sdlc/ -run TestClassifyFamily -v`
Expected: FAIL to compile.

- [ ] **Step 3: Implement** in `resolve.go`:

```go
type artifactKind int

const (
	kindIssue artifactKind = iota
	kindPlan
	kindReview
)

func (k artifactKind) String() string {
	switch k {
	case kindPlan:
		return "plan"
	case kindReview:
		return "review"
	default:
		return "issue"
	}
}

// Artifact is one resolved family member. Milestone is set for reviews
// (from -mX-review.md); "" for the -close-review.md and for non-reviews.
type Artifact struct {
	Kind      artifactKind
	Path      string
	Milestone string
}

var reviewMilestoneRe = regexp.MustCompile(`-m(\d+[a-z]?)-review\.md$`)

// classifyFamily classifies id NNNNNN's matched paths by filename suffix and
// returns them ordered issue → plan → reviews (reviews by milestone, close last).
// Pure: no IO. Paths not matching the id prefix are dropped defensively.
func classifyFamily(id int, paths []string) []Artifact {
	prefix := fmt.Sprintf("%06d-", id)
	var issue, plan, reviews []Artifact
	for _, p := range paths {
		base := filepath.Base(p)
		if !strings.HasPrefix(base, prefix) {
			continue
		}
		switch {
		case strings.HasSuffix(base, "-plan.md"):
			plan = append(plan, Artifact{Kind: kindPlan, Path: p})
		case strings.HasSuffix(base, "-review.md"):
			ms := ""
			if m := reviewMilestoneRe.FindStringSubmatch(base); m != nil {
				ms = "M" + m[1]
			}
			reviews = append(reviews, Artifact{Kind: kindReview, Path: p, Milestone: ms})
		default:
			issue = append(issue, Artifact{Kind: kindIssue, Path: p})
		}
	}
	sort.Slice(reviews, func(i, j int) bool {
		// milestone-tagged before close ("" sorts last); then lexically. NOTE:
		// lexical means M10 would sort before M2 — fine for the realistic M1–M9
		// range; make it numeric only if milestones ever exceed 9.
		if (reviews[i].Milestone == "") != (reviews[j].Milestone == "") {
			return reviews[j].Milestone == ""
		}
		if reviews[i].Milestone != reviews[j].Milestone {
			return reviews[i].Milestone < reviews[j].Milestone
		}
		return reviews[i].Path < reviews[j].Path
	})
	out := append(issue, plan...)
	return append(out, reviews...)
}
```

Add imports: `path/filepath`, `sort`.

- [ ] **Step 4: Run — expect PASS:**

Run: `go test ./cmd/sdlc/ -run TestClassifyFamily -v`
Expected: PASS.

- [ ] **Step 5: Run the whole sdlc package** to confirm nothing regressed:

Run: `go test ./cmd/sdlc/...`
Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add cmd/sdlc/resolve.go cmd/sdlc/resolve_test.go
git commit -m "#144 M1: pure classifyFamily — issue/plan/review classification + ordering"
```

### Task 1.5: Close M1

- [ ] **Step 1: Run the full test suite:**

Run: `go test ./...`
Expected: PASS.

- [ ] **Step 2: Update `## Log`** in the issue with M1 discoveries (model extension, pure core landed).

- [ ] **Step 3: Milestone-close** (auto-dispatches the fresh-context M1 review over the branch point → HEAD; fix Critical/Important before crossing):

Run: `sdlc milestone-close --issue 144 --milestone M1`
Expected: a `Review-Verdict:` trailer + a `closed M1` log line.

---

## Chunk 2: Milestone M2 — command + IO shell + docs

Delivers the read-only cobra surface (`sdlc resolve`, `sdlc open`), the sibling-repo + family-glob IO, the read-only-under-held-lock guarantee, the grammar help text, and the atlas update. Closes with `sdlc close --issue 144 --milestone M2`.

### Task 2.1: `resolveRepoDir` — sibling-repo directory resolution

**Files:**
- Modify: `cmd/sdlc/resolve.go`
- Test: `cmd/sdlc/resolve_test.go`

- [ ] **Step 1: Write the failing test** (against a temp parent dir with fake siblings):

```go
func TestResolveRepoDir(t *testing.T) {
	parent := t.TempDir()
	for _, name := range []string{"ariadne", "pair", "parley.nvim", "brain", "brain-family"} {
		if err := os.MkdirAll(filepath.Join(parent, name), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	cur := filepath.Join(parent, "ariadne")
	cases := []struct {
		repo    string
		wantDir string // basename, "" = expect error
	}{
		{"", "ariadne"},        // current repo
		{"pair", "pair"},        // exact
		{"parley", "parley.nvim"}, // unique prefix
		{"ariadne", "ariadne"},  // exact wins even though a prefix of nothing
		{"brain", "brain"},      // exact wins over the "brain-family" prefix sibling
		{"nope", ""},            // no match
		{"br", ""},              // ambiguous prefix (brain, brain-family)
	}
	for _, c := range cases {
		got, err := resolveRepoDir(ArtifactRef{Repo: c.repo}, cur)
		if c.wantDir == "" {
			if err == nil {
				t.Fatalf("%q: expected error, got %q", c.repo, got)
			}
			continue
		}
		if err != nil {
			t.Fatalf("%q: %v", c.repo, err)
		}
		if filepath.Base(got) != c.wantDir {
			t.Fatalf("%q: got %q want basename %q", c.repo, got, c.wantDir)
		}
	}
}
```

- [ ] **Step 2: Run — expect FAIL** (undefined `resolveRepoDir`):

Run: `go test ./cmd/sdlc/ -run TestResolveRepoDir -v`
Expected: FAIL to compile.

- [ ] **Step 3: Implement:**

```go
// resolveRepoDir maps a ref's repo token to an absolute repo directory. Empty
// token → curRoot. Else scan curRoot's parent for a sibling: exact basename
// match wins; else a unique case-insensitive prefix match; ambiguity or no
// match errors with the candidates. IO seam (reads the parent dir); curRoot is
// injected so the match logic is unit-testable.
func resolveRepoDir(ref ArtifactRef, curRoot string) (string, error) {
	if ref.Repo == "" {
		return curRoot, nil
	}
	parent := filepath.Dir(curRoot)
	entries, err := os.ReadDir(parent)
	if err != nil {
		return "", fmt.Errorf("read sibling dir %s: %w", parent, err)
	}
	var siblings []string
	for _, e := range entries {
		if e.IsDir() {
			siblings = append(siblings, e.Name())
		}
	}
	// exact basename match wins
	for _, s := range siblings {
		if s == ref.Repo {
			return filepath.Join(parent, s), nil
		}
	}
	// unique case-insensitive prefix match
	var pref []string
	low := strings.ToLower(ref.Repo)
	for _, s := range siblings {
		if strings.HasPrefix(strings.ToLower(s), low) {
			pref = append(pref, s)
		}
	}
	switch len(pref) {
	case 1:
		return filepath.Join(parent, pref[0]), nil
	case 0:
		return "", fmt.Errorf("no sibling repo matches %q under %s", ref.Repo, parent)
	default:
		sort.Strings(pref)
		return "", fmt.Errorf("ambiguous repo %q: matches %s", ref.Repo, strings.Join(pref, ", "))
	}
}
```

Add import: `os`.

- [ ] **Step 4: Run — expect PASS:**

Run: `go test ./cmd/sdlc/ -run TestResolveRepoDir -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add cmd/sdlc/resolve.go cmd/sdlc/resolve_test.go
git commit -m "#144 M2: resolveRepoDir — sibling repo dir (exact then unique-prefix)"
```

### Task 2.2: `familyFiles` — glob the three dirs

**Files:**
- Modify: `cmd/sdlc/resolve.go`
- Test: `cmd/sdlc/resolve_test.go`

- [ ] **Step 1: Write the failing test** (temp repo, files split across issues/plans/history):

```go
func TestFamilyFiles(t *testing.T) {
	root := t.TempDir()
	d := vocab.Issue().Discovery()
	seed := map[string][]string{
		d.Home:    {"000144-foo.md", "000200-bar.md"},
		d.Plans:   {"000144-foo-plan.md", "000144-foo-m1-review.md"},
		d.Archive: {"000144-foo-close-review.md"}, // partially archived
	}
	for dir, files := range seed {
		full := filepath.Join(root, dir)
		if err := os.MkdirAll(full, 0o755); err != nil {
			t.Fatal(err)
		}
		for _, f := range files {
			if err := os.WriteFile(filepath.Join(full, f), []byte("x"), 0o644); err != nil {
				t.Fatal(err)
			}
		}
	}
	got, err := familyFiles(root, d, 144)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 4 { // foo.md, plan, m1-review, close-review — NOT 000200
		t.Fatalf("got %d files: %v", len(got), got)
	}
}
```

- [ ] **Step 2: Run — expect FAIL:**

Run: `go test ./cmd/sdlc/ -run TestFamilyFiles -v`
Expected: FAIL to compile.

- [ ] **Step 3: Implement:**

```go
// familyFiles globs id NNNNNN's artifacts across the issue home, the plans home,
// and the archive — unioned and de-duped — so resolution is correct whether the
// family is active or (partially) archived. Directories come from the injected
// Discovery; nothing is hardcoded.
func familyFiles(repoDir string, d vocab.Discovery, id int) ([]string, error) {
	pat := fmt.Sprintf("%06d-*.md", id)
	seen := map[string]bool{}
	var out []string
	for _, sub := range []string{d.Home, d.Plans, d.Archive} {
		matches, err := filepath.Glob(filepath.Join(repoDir, sub, pat))
		if err != nil {
			return nil, fmt.Errorf("glob %s: %w", sub, err)
		}
		for _, m := range matches {
			if !seen[m] {
				seen[m] = true
				out = append(out, m)
			}
		}
	}
	sort.Strings(out)
	return out, nil
}
```

Add import: `github.com/xianxu/ariadne/pkg/vocab` (already imported by other files in the package; ensure it's in `resolve.go`).

- [ ] **Step 4: Run — expect PASS:**

Run: `go test ./cmd/sdlc/ -run TestFamilyFiles -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add cmd/sdlc/resolve.go cmd/sdlc/resolve_test.go
git commit -m "#144 M2: familyFiles — model-derived 3-dir glob, archive-correct"
```

### Task 2.3: The `resolve` command (wiring + output + `--json`)

**Files:**
- Modify: `cmd/sdlc/resolve.go` (add `NewResolveCmd`, output struct, run func)
- Test: `cmd/sdlc/resolve_test.go` (an end-to-end run against a temp repo)

- [ ] **Step 1: Write the failing E2E test** driving the run function against a temp repo (inject `curRoot`, capture stdout):

```go
func TestResolveRun_Family(t *testing.T) {
	root := seedTempRepo(t) // helper: writes 000144-foo.md + plan + m1/m2/close reviews
	var buf bytes.Buffer
	err := runResolve(resolveOpts{ref: "#144", root: root, out: &buf})
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 5 { // issue, plan, m1, m2, close
		t.Fatalf("got %d lines: %q", len(lines), buf.String())
	}
	if !strings.HasSuffix(lines[0], "000144-foo.md") {
		t.Fatalf("issue not first: %q", lines[0])
	}
}

func TestResolveRun_Milestone(t *testing.T) {
	root := seedTempRepo(t)
	var buf bytes.Buffer
	if err := runResolve(resolveOpts{ref: "#144 M2", root: root, out: &buf}); err != nil {
		t.Fatal(err)
	}
	got := strings.TrimSpace(buf.String())
	if !strings.HasSuffix(got, "000144-foo-m2-review.md") {
		t.Fatalf("M2 narrow failed: %q", got)
	}
}

func TestResolveRun_JSON(t *testing.T) {
	root := seedTempRepo(t) // places root at <parent>/ariadne so a sibling scan can find it
	var buf bytes.Buffer
	if err := runResolve(resolveOpts{ref: "ariadne#144", root: root, asJSON: true, out: &buf}); err != nil {
		t.Fatal(err)
	}
	var res resolveResult
	if err := json.Unmarshal(buf.Bytes(), &res); err != nil {
		t.Fatalf("json: %v (%q)", err, buf.String())
	}
	if res.ID != 144 || len(res.Files) == 0 {
		t.Fatalf("bad result: %+v", res)
	}
}

func TestResolveRun_GitHub(t *testing.T) {
	root := seedTempRepo(t)
	var buf bytes.Buffer
	if err := runResolve(resolveOpts{ref: "gh#42", root: root, out: &buf}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "github") {
		t.Fatalf("github ref not labeled: %q", buf.String())
	}
}
```

Provide the `seedTempRepo(t)` helper in the test file: it creates `<parent>/ariadne/workshop/{issues,plans,history}`, writes the family files, and returns the `ariadne` root.

- [ ] **Step 2: Run — expect FAIL:**

Run: `go test ./cmd/sdlc/ -run TestResolveRun -v`
Expected: FAIL to compile.

- [ ] **Step 3: Implement the run func + result struct + cobra command:**

```go
// resolveResult is the --json schema (field names are the JSON keys).
type resolveResult struct {
	Ref       string        `json:"ref"`
	Repo      string        `json:"repo"`
	ID        int           `json:"id"`
	Milestone string        `json:"milestone,omitempty"`
	GitHub    bool          `json:"github,omitempty"`
	Files     []resolveFile `json:"files"`
}
type resolveFile struct {
	Kind      string `json:"kind"`
	Path      string `json:"path"`
	Milestone string `json:"milestone,omitempty"`
}

type resolveOpts struct {
	ref    string
	root   string // current repo root (injected; empty ⇒ gitx.RepoTopLevel)
	asJSON bool
	out    io.Writer
}

// runResolve is the read-only engine glue: parse → resolve repo dir → glob family
// → classify → filter (milestone) → print. Takes NO lock (read-only). curRoot is
// injected for tests.
func runResolve(o resolveOpts) error {
	ref, err := parseRef(o.ref)
	if err != nil {
		return err
	}
	root := o.root
	if root == "" {
		root, err = gitx.RepoTopLevel()
		if err != nil {
			return err
		}
	}
	// GitHub refs: label, don't resolve to a file (read-only, offline).
	if ref.GitHub {
		res := resolveResult{Ref: o.ref, Repo: ref.Repo, ID: ref.ID, GitHub: true}
		if o.asJSON {
			return encodeJSON(o.out, res)
		}
		who := ref.Repo
		if who == "" {
			who = filepath.Base(root)
		}
		fmt.Fprintf(o.out, "github:%s#%d\n", who, ref.ID)
		return nil
	}
	repoDir, err := resolveRepoDir(ref, root)
	if err != nil {
		return err
	}
	files, err := familyFiles(repoDir, vocab.Issue().Discovery(), ref.ID)
	if err != nil {
		return err
	}
	fam := classifyFamily(ref.ID, files)
	// Distinguish "id not found at all" from "id found but this milestone has no
	// review sidecar" — the plan-review flagged conflating these (clarity).
	if len(fam) == 0 {
		return fmt.Errorf("no artifact resolves for #%d (searched %s)", ref.ID, repoDir)
	}
	if ref.Milestone != "" {
		fam = filterMilestone(fam, ref.Milestone) // reviews matching Mx
		if len(fam) == 0 {
			return fmt.Errorf("#%d exists but has no %s review sidecar in %s", ref.ID, ref.Milestone, repoDir)
		}
	}
	res := resolveResult{Ref: o.ref, Repo: ref.Repo, ID: ref.ID, Milestone: ref.Milestone}
	for _, a := range fam {
		res.Files = append(res.Files, resolveFile{Kind: a.Kind.String(), Path: a.Path, Milestone: a.Milestone})
	}
	if o.asJSON {
		return encodeJSON(o.out, res)
	}
	for _, a := range fam {
		fmt.Fprintln(o.out, a.Path)
	}
	return nil
}

// filterMilestone narrows the family to the review sidecar(s) for milestone ms.
// Returns the (possibly empty) hits; runResolve turns an empty result for a
// present issue into the distinct "#N exists but has no <Mx> review sidecar" error.
func filterMilestone(fam []Artifact, ms string) []Artifact {
	var hits []Artifact
	for _, a := range fam {
		if a.Kind == kindReview && a.Milestone == ms {
			hits = append(hits, a)
		}
	}
	return hits
}
```

JSON output uses the **inline** encoder idiom (verified: there is NO shared `writeJSON` helper in `cmd/sdlc`; `state.go:135-137` inlines `json.NewEncoder(stdout); enc.SetIndent("", "  "); enc.Encode(s)`). To avoid a second JSON path (ARCH-DRY) yet not invent a package-wide helper, use `json.NewEncoder(o.out)` directly in `runResolve`'s `if o.asJSON` branches:

```go
enc := json.NewEncoder(o.out)
enc.SetIndent("", "  ")
return enc.Encode(res)
```

The code block above uses a tiny file-local `encodeJSON`:

```go
func encodeJSON(w io.Writer, v any) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}
```

(Do NOT reference a package-level `writeJSON` — it does not exist. Add `encoding/json` + `io` to `resolve.go`'s imports.)

Add the cobra command:

```go
func NewResolveCmd() *cobra.Command {
	var asJSON bool
	cmd := &cobra.Command{
		Use:           "resolve <ref>",
		Args:          cobra.ExactArgs(1),
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runResolve(resolveOpts{ref: args[0], asJSON: asJSON, out: cmd.OutOrStdout()})
		},
	}
	cmd.Flags().BoolVar(&asJSON, "json", false, "emit the structured resolution as JSON")
	return cmd
}
```

- [ ] **Step 4: Run — expect PASS:**

Run: `go test ./cmd/sdlc/ -run TestResolveRun -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add cmd/sdlc/resolve.go cmd/sdlc/resolve_test.go
git commit -m "#144 M2: sdlc resolve command — family output, --json, Mx narrow, gh label"
```

### Task 2.4: `sdlc open` sugar + share the resolve engine

> **Ordering (plan-review fix):** this task does NOT register the commands in
> `buildRoot`. Registration happens in Task 2.5, *after* the helptext `.md` files
> exist — because `add(...)` calls `renderLong` → `helptext.MustGet`, which
> **panics at runtime** on a missing `.md`. Registering here would break every
> test that constructs `buildRoot()` (e.g. `TestNoCommandLongHasSurvivingPlaceholder`).

**Files:**
- Modify: `cmd/sdlc/resolve.go` (add `NewOpenCmd`)
- Modify: `cmd/sdlc/main.go` (register `resolve` + `open`, both read-only)
- Test: `cmd/sdlc/resolve_test.go` (open picks the primary target; `$EDITOR` faked)

- [ ] **Step 1: Write the failing test** — `open` selects the primary (issue for bare id, Mx review for `#id Mx`) and invokes `$EDITOR` with it. Fake the editor by setting `$EDITOR` to a script that writes its arg to a temp file, or inject an `exec` seam:

```go
func TestOpenPicksPrimary(t *testing.T) {
	root := seedTempRepo(t)
	var opened string
	openExec = func(editor, path string) error { opened = path; return nil } // injected seam
	t.Cleanup(func() { openExec = defaultOpenExec })

	if err := runOpen(openOpts{ref: "#144", root: root}); err != nil {
		t.Fatal(err)
	}
	if !strings.HasSuffix(opened, "000144-foo.md") {
		t.Fatalf("open primary: %q", opened)
	}
	if err := runOpen(openOpts{ref: "#144 M2", root: root}); err != nil {
		t.Fatal(err)
	}
	if !strings.HasSuffix(opened, "000144-foo-m2-review.md") {
		t.Fatalf("open Mx primary: %q", opened)
	}
}
```

- [ ] **Step 2: Run — expect FAIL:**

Run: `go test ./cmd/sdlc/ -run TestOpenPicksPrimary -v`
Expected: FAIL to compile.

- [ ] **Step 3: Implement** `runOpen` + `NewOpenCmd` + the injectable `openExec` seam (default execs `$EDITOR`). Primary = first family member after `filterMilestone` (so Mx → the review; bare id → the issue, which classifyFamily orders first). GitHub refs → print the label, no editor.

```go
var openExec = defaultOpenExec

func defaultOpenExec(editor, path string) error {
	c := exec.Command(editor, path)
	c.Stdin, c.Stdout, c.Stderr = os.Stdin, os.Stdout, os.Stderr
	return c.Run()
}
```

`runOpen` reuses `parseRef` / `resolveRepoDir` / `familyFiles` / `classifyFamily` / `filterMilestone` (no duplication — same engine as `runResolve`; factor the shared "ref → []Artifact" step into one helper `resolveArtifacts(o) ([]Artifact, ArtifactRef, error)` that both `runResolve` and `runOpen` call, per ARCH-DRY).

- [ ] **Step 4: Refactor `runResolve` to call the shared `resolveArtifacts`** so parse→glob→classify→filter lives once. Re-run `TestResolveRun*` to confirm no behavior change.

Run: `go test ./cmd/sdlc/ -run 'TestResolveRun|TestOpenPicksPrimary' -v`
Expected: PASS.

- [ ] **Step 5: Run the package tests + build** (commands not yet registered — that's intentional):

Run: `go build ./cmd/sdlc/ && go test ./cmd/sdlc/...`
Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add cmd/sdlc/resolve.go cmd/sdlc/resolve_test.go
git commit -m "#144 M2: sdlc open sugar + shared resolveArtifacts engine"
```

### Task 2.5: Grammar help text (single-source doc) + register in `buildRoot`

This task creates the helptext **before** wiring `add(...)` — the ordering that
avoids the `MustGet` panic (see Task 2.4's ordering note). Help text + registration
land in one commit so the tree is never in a "registered-but-no-.md" state.

**Files:**
- Create: `cmd/sdlc/helptext/resolve.md`
- Create: `cmd/sdlc/helptext/open.md`
- Modify: `cmd/sdlc/main.go` (register `resolve` + `open` via `add(...)`)
- Test: `cmd/sdlc/helptext/embed_test.go` (add resolve.md/open.md presence assertions)

- [ ] **Step 1: Write `resolve.md`** — the authoritative human/agent-facing grammar reference (the ref grammar table above, the read-only/lock-free guarantee, `--json` schema, examples incl. archived + cross-repo + Mx + gh). State explicitly: *"parley#160 and agents MUST shell to `sdlc resolve` rather than re-implement this grammar — this parser is the single source."*

- [ ] **Step 2: Write `open.md`** — brief; points at `resolve` for the grammar, documents primary-target selection and `$EDITOR`.

- [ ] **Step 3: Register in `buildRoot`** (main.go), after `state` (grouping read-only inspectors). The second arg is the longKey → must match the helptext basename so `renderLong`/`MustGet` loads it (the files now exist from Steps 1–2):

```go
add(NewResolveCmd(), "resolve", "Resolve a symbolic artifact ref (ariadne#11, #15 M4) to its current path(s) — read-only")
add(NewOpenCmd(), "open", "Resolve a ref and open the primary artifact in $EDITOR")
```

- [ ] **Step 4: Extend `embed_test.go`** with presence checks for `resolve.md` and `open.md` (mirror the existing `root.md`/`close.md` assertions).

- [ ] **Step 5: Run** — now that both `.md` exist AND the commands are registered, the `buildRoot()`-constructing tests must pass (this is where a missing `.md` would have panicked):

Run: `go build ./cmd/sdlc/ && go test ./cmd/sdlc/... && go run ./cmd/sdlc resolve --help`
Expected: help renders the grammar; full package (incl. `buildRoot()` tests) PASS.

- [ ] **Step 6: Commit**

```bash
git add cmd/sdlc/helptext/resolve.md cmd/sdlc/helptext/open.md cmd/sdlc/helptext/embed_test.go cmd/sdlc/main.go
git commit -m "#144 M2: resolve/open help text + register — single-source ref grammar for parley#160"
```

### Task 2.6: Read-only guarantee — lock-free by construction + under a held lock

The lock mechanism is **opt-in** (verified): only commands tagged
`markMutatingCommand` are wrapped by `wrapRepoLockCommands` and acquire
`.git/sdlc.lock` (`cmd/sdlc/repolock.go` — `commandNeedsRepoLock` returns false
otherwise). resolve/open are simply **never tagged**, so they're lock-free by
construction. There is **no** reusable `acquireTestLock` helper (verified); the real
API is `repolock.Acquire(ctx, repolock.Options{...})` + `(*Lock).Release()`
(`cmd/sdlc/internal/repolock/repolock.go`). Two complementary proofs:

**Files:**
- Test: `cmd/sdlc/resolve_test.go`

- [ ] **Step 1: Structural proof (cheap, exact) — the command is not in the mutating set:**

```go
func TestResolveOpenAreLockFree(t *testing.T) {
	// commandNeedsRepoLock takes *cobra.Command and reads the mutating annotation
	// that markMutatingCommand sets. resolve/open are never tagged → false.
	if commandNeedsRepoLock(NewResolveCmd()) {
		t.Fatal("resolve must not require the repo lock (read-only)")
	}
	if commandNeedsRepoLock(NewOpenCmd()) {
		t.Fatal("open must not require the repo lock (read-only)")
	}
}
```

- [ ] **Step 2: Runtime proof — resolution succeeds while a real lock is held.** Acquire a real lock against a temp `GitCommonDir` via `repolock.Acquire`, then call `runResolve` **directly** (not via `buildRoot().Execute()`, which per #149 keys the lock off cwd — calling `runResolve` sidesteps that entirely and is the honest read-only path parley uses):

```go
func TestResolveResolvesUnderHeldLock(t *testing.T) {
	root := seedTempRepo(t)
	gitCommon := filepath.Join(t.TempDir(), "gitcommon")
	if err := os.MkdirAll(gitCommon, 0o755); err != nil {
		t.Fatal(err)
	}
	lock, err := repolock.Acquire(context.Background(), repolock.Options{
		GitCommonDir: gitCommon,
		PID:          os.Getpid(),
		Hostname:     "test",
		ProcessAlive: func(int) bool { return true },
	})
	if err != nil {
		t.Fatalf("acquire: %v", err)
	}
	defer lock.Release()

	var buf bytes.Buffer
	if err := runResolve(resolveOpts{ref: "#144", root: root, out: &buf}); err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(buf.String()) == "" {
		t.Fatal("resolve produced no output under a held lock")
	}
}
```

Confirm the exact `repolock.Options` field set at implementation time (Step 1 of Task 2.6 — read `repolock.go`'s `Options` struct; adjust `PID`/`Hostname`/`ProcessAlive` to the real signature). If `Options` differs, match it — the point is a *real* held lock, then a successful `runResolve`.

- [ ] **Step 3: Run — expect PASS, promptly:**

Run: `go test ./cmd/sdlc/ -run 'TestResolveOpenAreLockFree|TestResolveResolvesUnderHeldLock' -v`
Expected: PASS. (A hang would mean resolve wrongly acquires the lock — but structurally it can't, since it's untagged.)

- [ ] **Step 4: Commit**

```bash
git add cmd/sdlc/resolve_test.go
git commit -m "#144 M2: prove resolve/open are lock-free (untagged + resolves under a held lock)"
```

### Task 2.7: Manual end-to-end verification against the real repo

- [ ] **Step 1: Build + install:**

Run: `go install ./cmd/sdlc && which sdlc`

- [ ] **Step 2: Resolve an active issue's family:**

Run: `sdlc resolve '#144'`
Expected: prints this issue's file + `000144-sdlc-resolve-plan.md` (+ any review sidecars), absolute paths.

- [ ] **Step 3: Resolve an ARCHIVED issue** (Done-when: correct after `issues/ → history/`). Pick an id that lives in `workshop/history/` (e.g. `sdlc resolve '#160'`):

Run: `sdlc resolve '#160'`
Expected: prints the `workshop/history/000160-*.md` paths (issue + plan + reviews) — proving archive-correctness.

- [ ] **Step 4: Cross-repo** (Done-when: across sibling repos):

Run: `sdlc resolve 'parley#160'`
Expected: resolves `parley` → `../parley.nvim` and prints `../parley.nvim/workshop/issues/000160-*.md`.

- [ ] **Step 5: Milestone + JSON + gh:**

Run: `sdlc resolve '#160 M2'` ; `sdlc resolve --json '#144'` ; `sdlc resolve 'gh#42'`
Expected: Mx narrows to the `-m2-review.md`; JSON parses with `files[]`; gh prints `github:...`.

- [ ] **Step 6: Record the exact commands + output** in the issue `## Log` (verification evidence for close).

### Task 2.8: Atlas update

**Files:**
- Modify: `atlas/workflow/sdlc-binary.md` (document `resolve`/`open` in the verb list + the ref-grammar single-source)
- Modify: `atlas/index.md` if a new file were added (none expected — edit in place)

- [ ] **Step 1: Add a `resolve` / `open` subsection** to `atlas/workflow/sdlc-binary.md`: read-only resolver, the grammar single-source, model-derived locations (`discovery.archive`/`plans`), the parley#160 consumer relationship, and the lock-free property.

- [ ] **Step 2: Note the DRY follow-up** (Design decision 6): existing `workshop/plans`/`workshop/history` hardcoders (`push`/`merge`/`state`) should migrate onto `vocab.Discovery()` — cross-link #163.

- [ ] **Step 3: Commit**

```bash
git add atlas/workflow/sdlc-binary.md
git commit -m "#144 M2: atlas — document sdlc resolve/open + the ref-grammar single source"
```

### Task 2.9: Close the issue

- [ ] **Step 1: Full suite green:**

Run: `go test ./...`
Expected: PASS.

- [ ] **Step 2: Tick the `## Plan` + `## Done when` boxes** in the issue; write the final `## Log` entry (evidence from Task 2.7).

- [ ] **Step 3: Compute actuals** (measured, never typed):

Run: `sdlc actual --issue 144`

- [ ] **Step 4: Close** (auto-dispatches the M2/whole-issue boundary review; fix Critical/Important first):

Run: `sdlc close --issue 144 --milestone M2 --verified '<evidence: resolve #144/#160/parley#160/gh#42 outputs + lock-free test + full suite>'`
Expected: passes the actual/verified/atlas gates; lands a `Review-Verdict:` trailer; flips to `codecomplete`.

---

## Revisions

### 2026-07-05 — collapse M1/M2 into a single close boundary

M1 and M2 remain the plan's **logical phases** (pure core, then command+IO+docs)
and the commit prefixes still read `#144 M1:` / `#144 M2:` for legibility. But
per AGENTS.md §3 (don't over-split cohesive ~sub-2h work into separate review
boundaries — it forces a redundant milestone-close + issue-close double-review),
this ships as ONE deliverable and closes at **one boundary**: a single `sdlc
close --issue 144` dispatches one fresh-context review over the whole branch
(branch-point → HEAD), covering model + parser + classifier + IO + commands +
docs together. So Task 1.5's `sdlc milestone-close --milestone M1` is dropped,
and Task 2.9 closes the issue in one pass (no `--milestone`). The atlas update
(Task 2.8) lands in-window, so the single close's atlas gate is satisfied without
a bypass. Estimate impact: the two `milestone-review` items collapse to one
boundary review — a modest reduction that `sdlc actual` will reflect at close.

## Notes on skills & conventions

- TDD throughout (superpowers-test-driven-development): red → green → commit per step.
- The pure core (Chunk 1) is unit-tested with no IO; the IO shell (Chunk 2) is tested against temp repos — the ARCH-PURE boundary is visible from the test structure (no mocks needed for the pure entities).
- Frequent commits with `#144 Mx:` prefix (AGENTS.md §12) so `git log --grep '^#144'` reads the arc.
- Both milestones close through `sdlc milestone-close` / `sdlc close`, whose auto-dispatched fresh-context review IS the mandatory boundary review (AGENTS.md §3 / #69) — do NOT run a separate `superpowers-requesting-code-review`.
