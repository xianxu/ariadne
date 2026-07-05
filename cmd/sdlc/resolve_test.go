package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/pkg/vocab"
)

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

func TestResolveRun_NotFound(t *testing.T) {
	root := seedTempRepo(t)
	err := runResolve(resolveOpts{ref: "#999", root: root, out: &bytes.Buffer{}})
	if err == nil || !strings.Contains(err.Error(), "no artifact resolves for #999") {
		t.Fatalf("want not-found error, got %v", err)
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
