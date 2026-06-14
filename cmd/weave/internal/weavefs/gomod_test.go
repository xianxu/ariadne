package weavefs

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// goDirective / goAtLeast124 / moduleLine-style parsing are pure — table-tested
// directly. OSGoMod.AddTool shells out to the real `go` toolchain against a
// hermetic t.TempDir go.mod (ARCH: faithful over mocked — the exec seam is
// exercised end-to-end, the same `go mod edit -tool` setup.sh runs).

func TestGoDirective(t *testing.T) {
	cases := map[string]string{
		"module x\n\ngo 1.26\n":    "1.26",
		"go 1.23.4\nmodule x\n":    "1.23.4",
		"module x\nrequire y v1\n": "",
		"":                         "",
	}
	for content, want := range cases {
		if got := goDirective(content); got != want {
			t.Errorf("goDirective(%q) = %q, want %q", content, got, want)
		}
	}
}

func TestGoAtLeast124(t *testing.T) {
	yes := []string{"1.24", "1.24.0", "1.26", "1.30", "2.0"}
	no := []string{"1.23", "1.23.9", "1.0", "1", "", "junk"}
	for _, v := range yes {
		if !goAtLeast124(v) {
			t.Errorf("goAtLeast124(%q) = false, want true", v)
		}
	}
	for _, v := range no {
		if goAtLeast124(v) {
			t.Errorf("goAtLeast124(%q) = true, want false", v)
		}
	}
}

func TestOSGoModAddToolReal(t *testing.T) {
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go toolchain not on PATH")
	}
	dir := t.TempDir()
	gomod := filepath.Join(dir, "go.mod")
	if err := os.WriteFile(gomod, []byte("module example.com/owner\n\ngo 1.26\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := (OSGoMod{}).AddTool(gomod, "example.com/owner/cmd/sdlc"); err != nil {
		t.Fatalf("AddTool: %v", err)
	}
	data, err := os.ReadFile(gomod)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "tool example.com/owner/cmd/sdlc") {
		t.Fatalf("go.mod missing tool directive after AddTool:\n%s", data)
	}
	// Idempotent: a second AddTool keeps exactly one tool directive.
	if err := (OSGoMod{}).AddTool(gomod, "example.com/owner/cmd/sdlc"); err != nil {
		t.Fatalf("AddTool (2nd): %v", err)
	}
	data2, err := os.ReadFile(gomod)
	if err != nil {
		t.Fatal(err)
	}
	if n := strings.Count(string(data2), "example.com/owner/cmd/sdlc"); n != 1 {
		t.Fatalf("tool directive appears %d times after re-AddTool, want 1:\n%s", n, data2)
	}
}

func TestOSGoModBumpsOldGoDirective(t *testing.T) {
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go toolchain not on PATH")
	}
	dir := t.TempDir()
	gomod := filepath.Join(dir, "go.mod")
	// An old go directive (< 1.24) must be bumped so `tool` is legal.
	if err := os.WriteFile(gomod, []byte("module example.com/owner\n\ngo 1.21\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := (OSGoMod{}).AddTool(gomod, "example.com/owner/cmd/sdlc"); err != nil {
		t.Fatalf("AddTool: %v", err)
	}
	data, err := os.ReadFile(gomod)
	if err != nil {
		t.Fatal(err)
	}
	if goDirective(string(data)) == "1.21" {
		t.Fatalf("go directive not bumped from 1.21:\n%s", data)
	}
	if !goAtLeast124(goDirective(string(data))) {
		t.Fatalf("go directive %q < 1.24 after bump:\n%s", goDirective(string(data)), data)
	}
}
