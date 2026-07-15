package vocab

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestInitialStatus(t *testing.T) {
	if got := Issue().InitialStatus(); got != "open" {
		t.Errorf("InitialStatus() = %q, want open (categories.open[0])", got)
	}
}

func TestSections(t *testing.T) {
	secs := Issue().Sections()
	// Order + seeds must match the cue model exactly (creation-template shape).
	want := []struct{ name, seed string }{
		{"Problem", ""},
		{"Spec", ""},
		{"Done when", "-"},
		{"Plan", "- [ ]"},
		{"Log", ""},
	}
	if len(secs) != len(want) {
		t.Fatalf("Sections() len = %d, want %d: %+v", len(secs), len(want), secs)
	}
	for i, w := range want {
		if secs[i].Name != w.name || secs[i].Seed != w.seed {
			t.Errorf("Sections()[%d] = {%q,%q}, want {%q,%q}", i, secs[i].Name, secs[i].Seed, w.name, w.seed)
		}
	}
}

func TestIssuePredicates(t *testing.T) {
	m := Issue()
	cases := []struct {
		name string
		got  bool
		want bool
	}{
		{"IsTerminal(done)", m.IsTerminal("done"), true},
		{"IsTerminal(working)", m.IsTerminal("working"), false},
		{"IsActive(blocked)", m.IsActive("blocked"), true},
		{"IsActive(done)", m.IsActive("done"), false},
		{"IsOpen(open)", m.IsOpen("open"), true},
		{"IsOpen(working)", m.IsOpen("working"), false},
		{"CanTransition(open,working)", m.CanTransition("open", "working"), true},
		{"CanTransition(open,done)", m.CanTransition("open", "done"), false},
		{"CanTransition(done,working)", m.CanTransition("done", "working"), true},
		// #160: codecomplete is an active (not terminal) status.
		{"IsActive(codecomplete)", m.IsActive("codecomplete"), true},
		{"IsTerminal(codecomplete)", m.IsTerminal("codecomplete"), false},
		// #160 close edges: close writes codecomplete unconditionally, so BOTH
		// working→codecomplete and blocked→codecomplete are model-legal.
		{"CanTransition(working,codecomplete)", m.CanTransition("working", "codecomplete"), true},
		{"CanTransition(blocked,codecomplete)", m.CanTransition("blocked", "codecomplete"), true},
		// #160 publish edge + rework/abandon.
		{"CanTransition(codecomplete,done)", m.CanTransition("codecomplete", "done"), true},
		{"CanTransition(codecomplete,working)", m.CanTransition("codecomplete", "working"), true},
		// working no longer closes straight to done (it routes through codecomplete).
		{"CanTransition(working,done)", m.CanTransition("working", "done"), false},
	}
	for _, c := range cases {
		if c.got != c.want {
			t.Errorf("%s = %v, want %v", c.name, c.got, c.want)
		}
	}
}

// TestDiscovery pins the discovery: block accessor (ariadne#144): the location
// model must expose home/glob (existing) plus archive/plans (added for the
// artifact-family resolver). Derived from issue.cue, never hand-set.
func TestDiscovery(t *testing.T) {
	d := Issue().Discovery()
	if d.Home != "workshop/issues" || d.Glob != "*.md" {
		t.Fatalf("home/glob: got %+v", d)
	}
	if d.Archive != "workshop/history" {
		t.Fatalf("archive: got %q, want workshop/history", d.Archive)
	}
	if d.Plans != "workshop/plans" {
		t.Fatalf("plans: got %q, want workshop/plans", d.Plans)
	}
}

func TestAllStatuses(t *testing.T) {
	// Ordered open → active → terminal; must match the legacy validStatuses set.
	want := []string{"open", "working", "blocked", "codecomplete", "done", "wontfix", "punt"}
	if got := Issue().AllStatuses(); !reflect.DeepEqual(got, want) {
		t.Errorf("AllStatuses() = %v, want %v", got, want)
	}
}

// TestRenderLifecycleHelp_DerivesFromModel pins #125's render: every status appears
// with its When semantics, the legal edges are present, and the output is byte-stable
// (two calls identical) — so the help text can't drift from the model.
func TestRenderLifecycleHelp_DerivesFromModel(t *testing.T) {
	m := Issue()
	out := m.RenderLifecycleHelp()
	if out != m.RenderLifecycleHelp() {
		t.Fatal("RenderLifecycleHelp is not byte-stable across calls")
	}
	for _, s := range m.AllStatuses() {
		if !strings.Contains(out, s) {
			t.Errorf("render missing status %q", s)
		}
		if w := m.When[s]; w != "" && !strings.Contains(out, w) {
			t.Errorf("render missing When semantics for %q (%q)", s, w)
		}
	}
	// A known legal edge renders (open → working).
	if !strings.Contains(out, "STATUSES") || !strings.Contains(out, "LEGAL TRANSITIONS") {
		t.Errorf("render missing a section header:\n%s", out)
	}
	if !strings.Contains(out, "working") {
		t.Errorf("render missing the open→working legal target:\n%s", out)
	}
}

func TestStatusNamesAndGloss(t *testing.T) {
	m := Issue()
	names := m.StatusNames(" | ")
	for _, s := range m.AllStatuses() {
		if !strings.Contains(names, s) {
			t.Errorf("StatusNames missing %q: %q", s, names)
		}
	}
	gloss := m.StatusGloss()
	if m.StatusGloss() != gloss {
		t.Fatal("StatusGloss not byte-stable")
	}
	for _, s := range m.AllStatuses() {
		if w := m.When[s]; w != "" && !strings.Contains(gloss, w) {
			t.Errorf("StatusGloss missing When for %q", s)
		}
	}
}

// TestArchiveSubdirs pins the per-kind archive layout convention (#181):
// derived from an arbitrary root (writers take --history-dir overrides), one
// function, every consumer routes through it.
func TestArchiveSubdirs(t *testing.T) {
	issues, plans := ArchiveSubdirs("workshop/history")
	if issues != "workshop/history/issues" {
		t.Errorf("issues subdir = %q", issues)
	}
	if plans != "workshop/history/plans" {
		t.Errorf("plans subdir = %q", plans)
	}
	// Arbitrary root (flag override) derives consistently.
	if i2, p2 := ArchiveSubdirs("/tmp/hx"); i2 != "/tmp/hx/issues" || p2 != "/tmp/hx/plans" {
		t.Errorf("override root: got (%q, %q)", i2, p2)
	}
}

// TestArchiveSubdirs_SingleDerivationPoint is the #163-pattern source guard
// (#181): no non-test Go file may concatenate the archive subdir literals
// itself — every consumer derives through ArchiveSubdirs, so a future layout
// change (e.g. history/projects/, #180) is one function edit. Scans the
// repo's Go source for the tell-tale literals.
func TestArchiveSubdirs_SingleDerivationPoint(t *testing.T) {
	root, err := repoRoot()
	if err != nil {
		t.Skipf("repo root not found: %v", err)
	}
	offenders := []string{}
	for _, dir := range []string{"cmd", "pkg"} {
		filepath.WalkDir(filepath.Join(root, dir), func(path string, d os.DirEntry, err error) error {
			if err != nil || d.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
				return nil
			}
			data, rerr := os.ReadFile(path)
			if rerr != nil {
				return nil
			}
			for _, lit := range []string{`"history/issues"`, `"history/plans"`, `history + "/issues"`, `history + "/plans"`} {
				if strings.Contains(string(data), lit) {
					offenders = append(offenders, path+": "+lit)
				}
			}
			// The subdir names may only appear as filepath.Join(root, "issues"/"plans")
			// inside ArchiveSubdirs itself (vocab.go).
			if filepath.Base(path) != "vocab.go" && strings.Contains(string(data), `filepath.Join(historyDir, "issues")`) {
				offenders = append(offenders, path+`: filepath.Join(historyDir, "issues")`)
			}
			return nil
		})
	}
	if len(offenders) > 0 {
		t.Errorf("archive subdir literals outside vocab.ArchiveSubdirs (derive, don't concatenate):\n  %s",
			strings.Join(offenders, "\n  "))
	}
}

// repoRoot walks up from cwd to the go.mod that declares module ariadne.
func repoRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("no go.mod above cwd")
		}
		dir = parent
	}
}
