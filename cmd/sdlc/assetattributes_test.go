package main

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
)

func TestAssetAttributesRefuseConversions(t *testing.T) {
	root := testfix.Repo(t, testfix.InitialCommit())
	dst := "workshop/issues/assets/000143/evidence.bin"
	for _, attr := range []string{"text", "text=auto", "eol=crlf", "working-tree-encoding=UTF-16", "ident", "filter=evidence"} {
		t.Run(attr, func(t *testing.T) {
			mkArtifact(t, filepath.Join(root, ".gitattributes"), "*.bin "+attr+"\n")
			err := verifyAssetAttributes(context.Background(), root, dst)
			if err == nil || !strings.Contains(err.Error(), dst) || !strings.Contains(err.Error(), ".gitattributes") {
				t.Fatalf("conversion not refused actionably: %v", err)
			}
		})
	}
	testfix.Git(t, root, "config", "core.autocrlf", "true")
	mkArtifact(t, filepath.Join(root, ".gitattributes"), "*.bin !text !filter !eol !working-tree-encoding !ident\n")
	if err := verifyAssetAttributes(context.Background(), root, dst); err == nil || !strings.Contains(err.Error(), "core.autocrlf") {
		t.Fatalf("implicit text conversion accepted: %v", err)
	}
	mkArtifact(t, filepath.Join(root, ".gitattributes"), "*.bin -text -filter !eol -working-tree-encoding -ident\n")
	if err := verifyAssetAttributes(context.Background(), root, dst); err != nil {
		t.Fatal(err)
	}
}

func TestAssetAttributesPinnedSnapshotAndNoFilterExecution(t *testing.T) {
	root := testfix.Repo(t, testfix.InitialCommit())
	dst := "workshop/issues/assets/000143/evidence.bin"
	marker := filepath.Join(root, "filter-executed")
	testfix.Git(t, root, "config", "filter.evidence.smudge", "touch "+marker)
	mkArtifact(t, filepath.Join(root, ".gitattributes"), "*.bin filter=evidence\n")
	testfix.Git(t, root, "add", ".gitattributes")
	testfix.Git(t, root, "commit", "-qm", "converting attributes")
	ref := strings.TrimSpace(testfix.Capture(t, root, "rev-parse", "HEAD"))
	mkArtifact(t, filepath.Join(root, ".gitattributes"), "*.bin -text -filter !eol -working-tree-encoding -ident\n")
	if err := verifyAssetAttributes(context.Background(), root, dst); err != nil {
		t.Fatal(err)
	}
	if err := verifyAssetAttributesAt(context.Background(), root, dst, ref); err == nil {
		t.Fatal("pinned incoming conversion accepted using local attributes")
	}
	if _, err := os.Stat(marker); !os.IsNotExist(err) {
		t.Fatal("attribute inspection executed a filter")
	}
}

func TestAssetAttributesRefuseBareAutocrlf(t *testing.T) {
	root := testfix.Repo(t, testfix.InitialCommit())
	config := filepath.Join(root, ".git", "config")
	f, err := os.OpenFile(config, os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = f.WriteString("\n[core]\n\tautocrlf\n"); err != nil {
		t.Fatal(err)
	}
	if err = f.Close(); err != nil {
		t.Fatal(err)
	}
	if got := strings.TrimSpace(testfix.Capture(t, root, "config", "--bool", "--get", "core.autocrlf")); got != "true" {
		t.Fatalf("fixture is not implicit true: %q", got)
	}
	if err := verifyAssetAttributes(context.Background(), root, "workshop/issues/assets/000143/evidence.bin"); err == nil {
		t.Fatal("bare autocrlf=true accepted")
	}
}
