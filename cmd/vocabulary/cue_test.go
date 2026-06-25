package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestOSCue_VetAndExport exercises the real `cue` binary (the IO seam). Skipped
// when cue isn't on PATH (e.g. a CI image without it).
func TestOSCue_VetAndExport(t *testing.T) {
	if _, err := exec.LookPath("cue"); err != nil {
		t.Skip("cue not on PATH")
	}
	good := filepath.Join(t.TempDir(), "x.cue")
	if err := os.WriteFile(good, []byte("package x\nfoo: \"bar\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	var c CueRunner = osCue{}
	if err := c.Vet(good); err != nil {
		t.Fatalf("vet good model: %v", err)
	}
	js, err := c.Export(good)
	if err != nil {
		t.Fatalf("export: %v", err)
	}
	if !strings.Contains(string(js), `"bar"`) {
		t.Fatalf("export missing value:\n%s", js)
	}
}
