package main

import (
	"bytes"
	"path/filepath"
	"testing"
)

func TestDependenciesCommandAndDryRun(t *testing.T) {
	root := t.TempDir()
	t.Chdir(root)
	mkfile(t, filepath.Join(root, "construct", "base.manifest"), "# local\n")
	var out bytes.Buffer
	cmd := buildRoot()
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"dependencies", "--dry-run"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
}
