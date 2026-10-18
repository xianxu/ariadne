package golint

import (
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// The annotation is the opt-in boundary. Keep these current tolerant inputs
// marked so the rule can protect later decoder changes in their packages.
func TestCurrentWireMessagesAnnotated(t *testing.T) {
	root := moduleRoot(t)
	for _, tc := range []struct{ file, name string }{
		{"cmd/sdlc/internal/judge/stream.go", "streamEvent"},
		{"cmd/sdlc/teststats_parse.go", "goTestJSONEvent"},
		{"cmd/sdlc/internal/usage/read.go", "record"},
		{"cmd/sdlc/internal/usage/codex.go", "codexUsage"},
	} {
		f, err := parser.ParseFile(token.NewFileSet(), filepath.Join(root, tc.file), nil, parser.ParseComments)
		if err != nil {
			t.Fatal(err)
		}
		marked, _ := markedWireTypes(f)
		if marked[tc.name] == nil {
			t.Errorf("%s: %s lost its wire annotation", tc.file, tc.name)
		}
	}
}

// TestTrackedGoFilesPassTheCatalog runs every catalogued rule over every
// tracked Go file (#407). A finding is fixed, or suppressed on its line with
// the reason it does not apply.
func TestTrackedGoFilesPassTheCatalog(t *testing.T) {
	root := moduleRoot(t)
	out, err := exec.Command("git", "-C", root, "ls-files", "-z", "--", "*.go").Output()
	if err != nil {
		t.Fatalf("git ls-files: %v", err)
	}
	start := time.Now()
	sources := map[string][]byte{}
	for _, rel := range strings.Split(strings.TrimRight(string(out), "\x00"), "\x00") {
		if rel == "" || strings.HasPrefix(rel, "workshop/") || strings.Contains("/"+rel, "/testdata/") {
			continue // archived prose and fixtures are not the module's code
		}
		src, err := os.ReadFile(filepath.Join(root, rel))
		if os.IsNotExist(err) {
			continue // deleted in the working tree
		}
		if err != nil {
			t.Fatal(err)
		}
		sources[rel] = src
	}
	if len(sources) == 0 {
		t.Fatal("no tracked Go files found; the file filter is broken")
	}
	goMod, err := os.ReadFile(filepath.Join(root, "go.mod"))
	if err != nil {
		t.Fatal(err)
	}
	module, ok := strings.CutPrefix(strings.SplitN(string(goMod), "\n", 2)[0], "module ")
	if !ok || module == "" {
		t.Fatalf("go.mod lacks a module line")
	}
	findings, err := CheckSources(module, sources, Catalog)
	if err != nil {
		t.Fatal(err)
	}
	for rel, list := range findings {
		for _, f := range list {
			t.Errorf("%s:%s", rel, f)
		}
	}
	t.Logf("checked %d Go files against %d rules in %s", len(sources), len(Catalog), time.Since(start).Round(time.Millisecond))
}

func moduleRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("no go.mod above the test directory")
		}
		dir = parent
	}
}
