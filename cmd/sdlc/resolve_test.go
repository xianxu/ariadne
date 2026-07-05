package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/xianxu/ariadne/pkg/vocab"
)

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
