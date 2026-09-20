package startup

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/weave/internal/weavefs"
)

func TestToolsOwnerOrderAndOptionalTarget(t *testing.T) {
	root := t.TempDir()
	base, leaf := filepath.Join(root, "base"), filepath.Join(root, "leaf")
	for _, dir := range []string{base, leaf} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatal(err)
		}
	}
	body := "tools:\n\tmkdir -p bin\n\tprintf '#!/bin/sh\\necho foundation\\n' > bin/base-tool\n\tchmod +x bin/base-tool\n"
	os.WriteFile(filepath.Join(base, "Makefile"), []byte(body), 0644)
	os.WriteFile(filepath.Join(leaf, "Makefile"), []byte("tools:\n\tbase-tool > built\n"), 0644)
	var out bytes.Buffer
	env := ToolEnvironment(os.Environ(), []string{base, leaf})
	if err := Tools(weavefs.OSFS{}, []string{base, leaf}, weavefs.ExecRunner{Context: context.Background(), Env: env, Stdout: &out, Stderr: &out}, false, &out); err != nil {
		t.Fatal(err, out.String())
	}
	got, err := os.ReadFile(filepath.Join(leaf, "built"))
	if err != nil || string(got) != "foundation\n" {
		t.Fatalf("build = %q, %v", got, err)
	}
	os.WriteFile(filepath.Join(leaf, "Makefile"), []byte("other:\n\tfalse\n"), 0644)
	if err := Tools(weavefs.OSFS{}, []string{leaf}, weavefs.ExecRunner{Stdout: &out, Stderr: &out}, false, &out); err != nil {
		t.Fatal("absent tools target:", err)
	}
}

func TestToolsFailuresStopAndDryRunDoesNotBuild(t *testing.T) {
	root := t.TempDir()
	os.WriteFile(filepath.Join(root, "Makefile"), []byte("tools:\n\tfalse\n"), 0644)
	var out bytes.Buffer
	if err := Tools(weavefs.OSFS{}, []string{root}, weavefs.ExecRunner{Stdout: &out, Stderr: &out}, true, &out); err != nil {
		t.Fatal(err)
	}
	if err := Tools(weavefs.OSFS{}, []string{root}, weavefs.ExecRunner{Stdout: &out, Stderr: &out}, false, &out); err == nil || !strings.Contains(err.Error(), root) {
		t.Fatalf("want owner build error, got %v", err)
	}
	os.WriteFile(filepath.Join(root, "Makefile"), []byte("invalid syntax\n"), 0644)
	if err := Tools(weavefs.OSFS{}, []string{root}, weavefs.ExecRunner{Stdout: &out, Stderr: &out}, false, &out); err == nil {
		t.Fatal("parse error swallowed")
	}
}
