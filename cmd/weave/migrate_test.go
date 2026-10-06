package main

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/weave/internal/acquire"
	"github.com/xianxu/ariadne/cmd/weave/internal/weavefs"
)

const sourcelessDeps = "substrate ../base # the base layer\nsubstrate ../local\n"

// migrationFleet is a primary checkout declaring two sourceless substrates:
// base, whose checkout has a remote origin, and local, which has none (#296).
func migrationFleet(t *testing.T) string {
	t.Helper()
	fleet := t.TempDir()
	root := filepath.Join(fleet, "host")
	for _, dir := range []string{root, filepath.Join(fleet, "base"), filepath.Join(fleet, "local")} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatal(err)
		}
		setupGit(t, dir, "init", "-q")
	}
	setupGit(t, filepath.Join(fleet, "base"), "remote", "add", "origin", "https://example.test/base.git")
	writeDeps(t, root, sourcelessDeps)
	setupGit(t, root, "add", ".")
	setupGit(t, root, "commit", "-qm", "init")
	return root
}

func writeDeps(t *testing.T, root, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(root, "construct"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "construct", "deps"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

func readDeps(t *testing.T, root string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(root, "construct", "deps"))
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func TestMigrateSourcelessRecordsOriginInPrimary(t *testing.T) {
	root := migrationFleet(t)
	var out bytes.Buffer
	if err := migrateSourceless(context.Background(), weavefs.OSFS{}, root, &out); err != nil {
		t.Fatal(err)
	}
	want := "substrate ../base https://example.test/base.git # the base layer\nsubstrate ../local\n"
	if got := readDeps(t, root); got != want {
		t.Fatalf("deps:\n%s\nwant:\n%s", got, want)
	}
	if !strings.Contains(out.String(), "recorded source https://example.test/base.git for substrate ../base") ||
		!strings.Contains(out.String(), "warning: substrate ../local has no source") {
		t.Fatalf("output: %s", out.String())
	}

	out.Reset()
	if err := migrateSourceless(context.Background(), weavefs.OSFS{}, root, &out); err != nil {
		t.Fatal(err)
	}
	if got := readDeps(t, root); got != want || strings.Contains(out.String(), "recorded") {
		t.Fatalf("second run changed deps or claimed a write: %s / %s", got, out.String())
	}
}

func TestMigrateSourcelessNeverWritesOutsidePrimary(t *testing.T) {
	root := migrationFleet(t)
	linked := filepath.Join(filepath.Dir(root), "linked")
	setupGit(t, root, "worktree", "add", "-q", "-b", "feature", linked)
	if err := migrateSourceless(context.Background(), weavefs.OSFS{}, linked, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	if got := readDeps(t, linked); got != sourcelessDeps {
		t.Fatalf("linked worktree rewritten: %s", got)
	}

	recovered := acquire.Result{Recovered: []acquire.Recovery{{Owner: root, Path: "../base", URL: "https://example.test/base.git", Sibling: "/fleet/base"}}}
	for name, tc := range map[string]struct {
		client acquire.Client
		dryRun bool
		err    error
	}{
		"slot":    {client: acquire.Client{Policy: &acquire.Policy{}}},
		"dry run": {dryRun: true},
		"failed":  {err: errors.New("restore failed")},
	} {
		var out bytes.Buffer
		err := finishRestore(context.Background(), weavefs.OSFS{}, root, tc.client, recovered, tc.err, tc.dryRun, &out)
		if !errors.Is(err, tc.err) {
			t.Fatalf("%s: error %v", name, err)
		}
		if got := readDeps(t, root); got != sourcelessDeps {
			t.Fatalf("%s rewrote deps: %s", name, got)
		}
		if !strings.Contains(out.String(), "recovered https://example.test/base.git from /fleet/base. Record it in construct/deps: substrate ../base https://example.test/base.git") {
			t.Fatalf("%s: recovery notice missing: %s", name, out.String())
		}
	}
	// Control: the same call in the primary, after a real restore, migrates.
	if err := finishRestore(context.Background(), weavefs.OSFS{}, root, acquire.Client{}, recovered, nil, false, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	if got := readDeps(t, root); got == sourcelessDeps {
		t.Fatal("primary restore did not migrate")
	}
}

// A slot of a repository with a sourceless substrate row restores it from the
// primary-side sibling's origin, reports it, and leaves construct/deps alone.
func TestNumberedSetupRecoversSourcelessSubstrate(t *testing.T) {
	fleet, env, host := numberedSetupFixture(t)
	work := filepath.Join(t.TempDir(), "base")
	mkfile(t, filepath.Join(work, "construct/base.manifest"), "# base\n")
	setupGit(t, work, "init", "-q", "-b", "main")
	setupGit(t, work, "add", ".")
	setupGit(t, work, "commit", "-qm", "base")
	// The declared URL stays the identity; Git fetches it from the fixture.
	t.Setenv("GIT_CONFIG_COUNT", "1")
	t.Setenv("GIT_CONFIG_KEY_0", "url."+work+".insteadOf")
	t.Setenv("GIT_CONFIG_VALUE_0", "https://example.test/base.git")
	sibling := filepath.Join(fleet, "base")
	if err := os.MkdirAll(sibling, 0755); err != nil {
		t.Fatal(err)
	}
	setupGit(t, sibling, "init", "-q")
	setupGit(t, sibling, "remote", "add", "origin", "https://example.test/base.git")
	mkfile(t, filepath.Join(host, "construct/deps"), "substrate ../base\n")
	t.Chdir(host)

	out, err := executeSetup(t, "dependencies", false)
	if err != nil {
		t.Fatal(err, out)
	}
	if !strings.Contains(out, "recovered https://example.test/base.git from "+sibling) {
		t.Fatalf("no recovery notice: %s", out)
	}
	dest := filepath.Join(env, "base")
	if got := setupGit(t, dest, "config", "remote.origin.url"); got != "https://example.test/base.git" {
		t.Fatalf("restored origin %s", got)
	}
	if got := readDeps(t, host); got != "substrate ../base\n" {
		t.Fatalf("slot rewrote construct/deps: %s", got)
	}
}
