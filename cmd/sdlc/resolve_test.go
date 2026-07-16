package main

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/helptext"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/repolock"
	"github.com/xianxu/ariadne/pkg/vocab"
)

// TestResolveDocExamplesParse binds the human-facing grammar doc (resolve.md /
// open.md) to the parser: every single-quoted ref shown in an EXAMPLES block
// must parse. So the documented table can't silently drift from parseRef (the
// single source) — the plan-quality judge's suggested guard (#144).
func TestResolveDocExamplesParse(t *testing.T) {
	exampleRef := regexp.MustCompile(`sdlc (?:resolve|open)(?: --json)? '([^']+)'`)
	found := 0
	for _, name := range []string{"resolve", "open"} {
		doc, ok := helptext.Get(name)
		if !ok {
			t.Fatalf("%s.md not embedded", name)
		}
		for _, m := range exampleRef.FindAllStringSubmatch(doc, -1) {
			found++
			if _, err := parseRef(m[1]); err != nil {
				t.Errorf("%s.md example %q does not parse: %v", name, m[1], err)
			}
		}
	}
	if found == 0 {
		t.Fatal("no example refs found in resolve.md/open.md — regex or docs changed")
	}
}

// seedTempRepo lays out <tmp>/ariadne/workshop/{issues,plans,history} with a
// #144 family and returns the ariadne repo root. Placing the root at
// <parent>/ariadne lets both `#144` (current repo) and `ariadne#144` (sibling
// scan from within ariadne) resolve. The M1/M2/close reviews live in plans/.
func seedTempRepo(t *testing.T) string {
	t.Helper()
	parent := t.TempDir()
	root := filepath.Join(parent, "ariadne")
	d := vocab.Issue().Discovery()
	seed := map[string][]string{
		d.Home: {"000144-foo.md"},
		d.Plans: {
			"000144-foo-plan.md",
			"000144-foo-m1-review.md",
			"000144-foo-m2-review.md",
			"000144-foo-close-review.md",
		},
	}
	for sub, files := range seed {
		full := filepath.Join(root, sub)
		if err := os.MkdirAll(full, 0o755); err != nil {
			t.Fatal(err)
		}
		for _, f := range files {
			if err := os.WriteFile(filepath.Join(full, f), []byte("x"), 0o644); err != nil {
				t.Fatal(err)
			}
		}
	}
	// history/ exists but empty (nothing archived yet).
	if err := os.MkdirAll(filepath.Join(root, d.Archive), 0o755); err != nil {
		t.Fatal(err)
	}
	return root
}

func TestResolveRun_Family(t *testing.T) {
	root := seedTempRepo(t)
	var buf bytes.Buffer
	if err := runResolve(resolveOpts{ref: "#144", root: root, out: &buf}); err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 5 { // issue, plan, m1, m2, close
		t.Fatalf("got %d lines: %q", len(lines), buf.String())
	}
	if !strings.HasSuffix(lines[0], "000144-foo.md") {
		t.Fatalf("issue not first: %q", lines[0])
	}
	if !strings.HasSuffix(lines[1], "000144-foo-plan.md") {
		t.Fatalf("plan not second: %q", lines[1])
	}
}

func TestResolveRun_Milestone(t *testing.T) {
	root := seedTempRepo(t)
	var buf bytes.Buffer
	if err := runResolve(resolveOpts{ref: "#144 M2", root: root, out: &buf}); err != nil {
		t.Fatal(err)
	}
	got := strings.TrimSpace(buf.String())
	if !strings.HasSuffix(got, "000144-foo-m2-review.md") || strings.Contains(got, "\n") {
		t.Fatalf("M2 narrow failed: %q", got)
	}
}

func TestResolveRun_MilestoneMissing(t *testing.T) {
	root := seedTempRepo(t)
	err := runResolve(resolveOpts{ref: "#144 M9", root: root, out: &bytes.Buffer{}})
	if err == nil || !strings.Contains(err.Error(), "no M9 review sidecar") {
		t.Fatalf("want distinct milestone-missing error, got %v", err)
	}
}

func TestResolveRun_JSON(t *testing.T) {
	root := seedTempRepo(t) // root is <parent>/ariadne so a sibling scan finds it
	var buf bytes.Buffer
	if err := runResolve(resolveOpts{ref: "ariadne#144", root: root, asJSON: true, out: &buf}); err != nil {
		t.Fatal(err)
	}
	var res resolveResult
	if err := json.Unmarshal(buf.Bytes(), &res); err != nil {
		t.Fatalf("json: %v (%q)", err, buf.String())
	}
	if res.ID != 144 || res.Repo != "ariadne" || len(res.Files) != 5 {
		t.Fatalf("bad result: %+v", res)
	}
	if res.Files[0].Kind != "issue" {
		t.Fatalf("first file kind: %q", res.Files[0].Kind)
	}
}

func TestResolveRun_GitHub(t *testing.T) {
	root := seedTempRepo(t)
	var buf bytes.Buffer
	if err := runResolve(resolveOpts{ref: "gh#42", root: root, out: &buf}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "github:ariadne#42") {
		t.Fatalf("github ref not labeled: %q", buf.String())
	}
}

// A GitHub ref's --json must emit files as an empty array, never null, so a
// consumer (parley#160) can iterate .files unconditionally.
func TestResolveRun_GitHubJSONFilesNotNull(t *testing.T) {
	root := seedTempRepo(t)
	var buf bytes.Buffer
	if err := runResolve(resolveOpts{ref: "gh#42", root: root, asJSON: true, out: &buf}); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(buf.String(), `"files": null`) {
		t.Fatalf("github --json emitted null files: %q", buf.String())
	}
	var res resolveResult
	if err := json.Unmarshal(buf.Bytes(), &res); err != nil {
		t.Fatal(err)
	}
	if res.Files == nil || len(res.Files) != 0 || !res.GitHub {
		t.Fatalf("want empty non-nil files + github=true: %+v", res)
	}
}

// Cross-repo where the resolved repo dir differs from the current root: run
// pair#50 from within ariadne and confirm it reads pair's own family. Exercises
// the full runResolve → resolveRepoDir(sibling) → familyFiles path end-to-end.
func TestResolveRun_CrossRepo(t *testing.T) {
	root := seedTempRepo(t) // <parent>/ariadne
	parent := filepath.Dir(root)
	d := vocab.Issue().Discovery()
	pairIssues := filepath.Join(parent, "pair", d.Home)
	if err := os.MkdirAll(pairIssues, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(pairIssues, "000050-pair-thing.md"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	if err := runResolve(resolveOpts{ref: "pair#50", root: root, out: &buf}); err != nil {
		t.Fatal(err)
	}
	got := strings.TrimSpace(buf.String())
	if !strings.HasSuffix(got, filepath.Join("pair", d.Home, "000050-pair-thing.md")) {
		t.Fatalf("cross-repo pair#50 did not resolve pair's file: %q", got)
	}
}

func TestResolveRun_NotFound(t *testing.T) {
	root := seedTempRepo(t)
	err := runResolve(resolveOpts{ref: "#999", root: root, out: &bytes.Buffer{}})
	if err == nil || !strings.Contains(err.Error(), "no artifact resolves for #999") {
		t.Fatalf("want not-found error, got %v", err)
	}
}

func TestOpenPicksPrimary(t *testing.T) {
	root := seedTempRepo(t)
	var opened string
	prev := openExec
	openExec = func(editor, path string) error { opened = path; return nil }
	t.Cleanup(func() { openExec = prev })

	if err := runOpen(openOpts{ref: "#144", root: root, out: &bytes.Buffer{}}); err != nil {
		t.Fatal(err)
	}
	if !strings.HasSuffix(opened, "000144-foo.md") {
		t.Fatalf("open primary (bare id → issue): %q", opened)
	}
	if err := runOpen(openOpts{ref: "#144 M2", root: root, out: &bytes.Buffer{}}); err != nil {
		t.Fatal(err)
	}
	if !strings.HasSuffix(opened, "000144-foo-m2-review.md") {
		t.Fatalf("open Mx primary (→ review): %q", opened)
	}
}

// TestResolveOpenAreLockFree is the structural proof (#144 Done-when): resolve
// and open are never tagged markMutatingCommand, so wrapRepoLockCommands skips
// them and they never acquire .git/sdlc.lock. Lock-free by construction.
func TestResolveOpenAreLockFree(t *testing.T) {
	if commandNeedsRepoLock(NewResolveCmd()) {
		t.Fatal("resolve must not require the repo lock (read-only)")
	}
	if commandNeedsRepoLock(NewOpenCmd()) {
		t.Fatal("open must not require the repo lock (read-only)")
	}
}

// TestResolveResolvesUnderHeldLock is the runtime proof: hold a real
// .git/sdlc.lock (via repolock.Acquire against a temp GitCommonDir), then call
// runResolve directly (parley's read-only path) and confirm it returns the
// family without blocking on the lock.
func TestResolveResolvesUnderHeldLock(t *testing.T) {
	root := seedTempRepo(t)
	gitCommon := t.TempDir()
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

func TestOpenGitHubNotOpened(t *testing.T) {
	root := seedTempRepo(t)
	called := false
	prev := openExec
	openExec = func(editor, path string) error { called = true; return nil }
	t.Cleanup(func() { openExec = prev })

	var buf bytes.Buffer
	if err := runOpen(openOpts{ref: "gh#42", root: root, out: &buf}); err != nil {
		t.Fatal(err)
	}
	if called {
		t.Fatal("github ref must not invoke the editor")
	}
	if !strings.Contains(buf.String(), "github:ariadne#42") {
		t.Fatalf("github label missing: %q", buf.String())
	}
}

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
		{"", "ariadne"},           // current repo
		{"pair", "pair"},          // exact
		{"parley", "parley.nvim"}, // unique prefix
		{"ariadne", "ariadne"},    // exact
		{"brain", "brain"},        // exact wins over the brain-family prefix sibling
		{"nope", ""},              // no match
		{"br", ""},                // ambiguous prefix (brain, brain-family)
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

func TestClassifyFamily(t *testing.T) {
	paths := []string{
		"/r/workshop/plans/000144-foo-m2-review.md",
		"/r/workshop/issues/000144-foo.md",
		"/r/workshop/plans/000144-foo-plan.md",
		"/r/workshop/plans/000144-foo-close-review.md",
		"/r/workshop/plans/000144-foo-m1-review.md",
		"/r/workshop/plans/000999-other.md", // wrong id — dropped
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
	if got[0].Kind.String() != "issue" || got[1].Kind.String() != "plan" || got[2].Kind.String() != "review" {
		t.Fatalf("kind String() mismatch: %+v", got[:3])
	}
}

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
		{"nope", ArtifactRef{}, true},     // no '#'
		{"#", ArtifactRef{}, true},        // no id
		{"#1234567", ArtifactRef{}, true}, // >6 digits
		{"a#1#2", ArtifactRef{}, true},    // two '#'
		{"#0", ArtifactRef{}, true},       // zero id
		{"#12 M", ArtifactRef{}, true},    // malformed milestone
		{"#12 M4 X", ArtifactRef{}, true}, // trailing token
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

// #181: resolution finds archived families in BOTH layouts — the per-kind
// subfolders (history/issues + history/plans) and the pre-#181 flat root
// (un-migrated + downstream repos) — as one complete, ordered family each.
func TestResolveRun_SubfolderedAndFlatArchive(t *testing.T) {
	root := seedTempRepo(t)
	d := vocab.Issue().Discovery()
	issuesSub := vocab.ArchiveSubdir(d.Archive, vocab.ArchiveIssues)
	plansSub := vocab.ArchiveSubdir(d.Archive, vocab.ArchivePlans)
	seed := map[string][]string{
		issuesSub: {"000031-sub.md"},                              // subfoldered issue
		plansSub:  {"000031-sub-plan.md", "000031-sub-close-review.md"}, // + family
		d.Archive: {"000032-flat.md", "000032-flat-plan.md"},      // legacy flat family
	}
	for sub, files := range seed {
		full := filepath.Join(root, sub)
		if err := os.MkdirAll(full, 0o755); err != nil {
			t.Fatal(err)
		}
		for _, f := range files {
			if err := os.WriteFile(filepath.Join(full, f), []byte("x"), 0o644); err != nil {
				t.Fatal(err)
			}
		}
	}
	for _, tc := range []struct {
		ref   string
		wants []string
	}{
		{"#31", []string{"000031-sub.md", "000031-sub-plan.md", "000031-sub-close-review.md"}},
		{"#32", []string{"000032-flat.md", "000032-flat-plan.md"}},
	} {
		var buf bytes.Buffer
		if err := runResolve(resolveOpts{ref: tc.ref, root: root, out: &buf}); err != nil {
			t.Fatalf("%s: %v", tc.ref, err)
		}
		lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
		if len(lines) != len(tc.wants) {
			t.Fatalf("%s: got %d lines, want %d: %q", tc.ref, len(lines), len(tc.wants), buf.String())
		}
		for i, w := range tc.wants {
			if !strings.HasSuffix(lines[i], w) {
				t.Errorf("%s line %d = %q, want suffix %q (ordering: issue → plan → reviews)", tc.ref, i, lines[i], w)
			}
		}
	}
}
