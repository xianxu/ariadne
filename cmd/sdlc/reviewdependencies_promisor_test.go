package main

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
)

// A local metadata query is not automatically read-only: Git's promisor
// machinery can lazily fetch missing blobs while ls-tree computes sizes.
func TestReviewDependenciesMissingPromisorBlobNeverFetches(t *testing.T) {
	source, _ := reviewDependencySourceFixture(t)
	blob := strings.TrimSpace(testfix.Capture(t, source, "rev-parse", "HEAD:fixture"))
	object := filepath.Join(source, ".git", "objects", blob[:2], blob[2:])
	if err := os.Remove(object); err != nil {
		t.Fatal(err)
	}
	tools := t.TempDir()
	marker := filepath.Join(tools, "upload-pack-called")
	helper := filepath.Join(tools, "upload-pack")
	if err := os.WriteFile(helper, []byte("#!/bin/sh\nprintf called > \"$REVIEW_PROMISOR_MARKER\"\nexit 1\n"), 0755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("REVIEW_PROMISOR_MARKER", marker)
	testfix.Git(t, source, "config", "remote.origin.url", "file://"+source)
	testfix.Git(t, source, "config", "remote.origin.uploadpack", helper)
	testfix.Git(t, source, "config", "remote.origin.promisor", "true")
	testfix.Git(t, source, "config", "extensions.partialclone", "origin")
	manifest := reviewDependencyConsumerFixture(t, "substrate "+source+"\n")
	before := reviewPromisorSnapshot(t, source)
	root := t.TempDir()
	t.Setenv("WF_REVIEW_CHECKOUT_ROOT", root)
	owner, iso, err := openReviewIsolationContext(t.Context(), new(bytes.Buffer), manifest)
	if owner != nil {
		if closeErr := owner.Close(); closeErr != nil {
			t.Error(closeErr)
		}
	}
	if err == nil || len(iso.Dependencies) != 0 {
		t.Errorf("accepted an incomplete dependency: bindings=%+v error=%v", iso.Dependencies, err)
	}
	if _, err := os.Lstat(marker); !os.IsNotExist(err) {
		t.Errorf("dependency observation invoked remote upload-pack: marker error=%v", err)
	}
	if _, err := os.Lstat(object); !os.IsNotExist(err) {
		t.Errorf("missing source blob was acquired or replaced: %v", err)
	}
	after := reviewPromisorSnapshot(t, source)
	var changed []string
	for path, value := range before {
		if after[path] != value {
			changed = append(changed, path)
		}
	}
	for path := range after {
		if _, present := before[path]; !present {
			changed = append(changed, path)
		}
	}
	if len(changed) != 0 {
		sort.Strings(changed)
		t.Errorf("read-only dependency preparation changed source paths: %v", changed)
	}
	if entries, err := os.ReadDir(root); err != nil || len(entries) != 0 {
		t.Errorf("failed preparation leaked checkouts or locks: %v %v", entries, err)
	}
}

func reviewPromisorSnapshot(t *testing.T, root string) map[string]string {
	t.Helper()
	snapshot := map[string]string{}
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		value := info.Mode().String()
		switch {
		case info.Mode().IsRegular():
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			value += fmt.Sprintf(" %x", sha256.Sum256(data))
		case info.Mode()&os.ModeSymlink != 0:
			target, err := os.Readlink(path)
			if err != nil {
				return err
			}
			value += " " + target
		}
		snapshot[rel] = value
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return snapshot
}
