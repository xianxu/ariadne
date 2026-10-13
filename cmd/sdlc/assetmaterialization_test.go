package main

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/issueassets"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/tracker"
)

func TestFastForwardRestRefusesAssetConversionBeforeEffects(t *testing.T) {
	root := testfix.Repo(t, testfix.InitialCommit())
	original := strings.TrimSpace(testfix.Capture(t, root, "rev-parse", "HEAD"))
	testfix.Git(t, root, "switch", "-qc", "incoming")
	asset := "workshop/history/issues/assets/000001/evidence.png"
	procedureWrite(t, root, asset, "one\ntwo")
	procedureWrite(t, root, ".gitattributes", "workshop/history/issues/assets/** text eol=crlf\n")
	testfix.Git(t, root, "add", ".")
	testfix.Git(t, root, "commit", "-qm", "incoming evidence")
	target := strings.TrimSpace(testfix.Capture(t, root, "rev-parse", "HEAD"))
	testfix.Git(t, root, "switch", "-q", "main")
	reader, err := gitx.NewTrunkFileContext(context.Background(), root, "origin", "main")
	if err != nil {
		t.Fatal(err)
	}
	env := &trackerEnv{ctx: context.Background(), root: root, main: reader, resting: "main"}
	warn := fastForwardRest(env, target)
	if !strings.Contains(warn, asset) || !strings.Contains(warn, ".gitattributes") {
		t.Fatalf("conversion accepted: %q", warn)
	}
	if strings.TrimSpace(testfix.Capture(t, root, "rev-parse", "HEAD")) != original {
		t.Fatal("refusal fast forwarded HEAD")
	}
}

func TestRestoreToHeadRefusesMaterializedAssetTampering(t *testing.T) {
	root := testfix.Repo(t, testfix.InitialCommit())
	p := "workshop/issues/assets/000001/evidence.png"
	procedureWrite(t, root, p, "evidence")
	testfix.Git(t, root, "add", ".")
	testfix.Git(t, root, "commit", "-qm", "evidence")
	hooks := t.TempDir()
	procedureWrite(t, hooks, "post-checkout", "#!/bin/sh\nprintf hook-tampered > '"+p+"'\n")
	if err := os.Chmod(filepath.Join(hooks, "post-checkout"), 0755); err != nil {
		t.Fatal(err)
	}
	testfix.Git(t, root, "config", "core.hooksPath", hooks)
	reader, err := gitx.NewTrunkFileContext(context.Background(), root, "origin", "main")
	if err != nil {
		t.Fatal(err)
	}
	env := &trackerEnv{ctx: context.Background(), root: root, main: reader}
	err = restoreToHead(env, p)
	if err == nil || !strings.Contains(err.Error(), p) || !strings.Contains(err.Error(), "restore") {
		t.Fatalf("tampered materialization accepted: %v", err)
	}
}

func TestFinishPublishedVerifiesAssetBytesBeforeCommit(t *testing.T) {
	r, _, _ := startedHere(t)
	p := assetFixture(t, r, "000009", handoffDetail)
	other := "workshop/issues/000008-other.md"
	writeRepoFile(t, r.root, other, "other issue\n")
	r.git("add", ".")
	r.git("commit", "-qm", "local evidence")
	head := r.git("rev-parse", "HEAD")
	raw, err := os.ReadFile(filepath.Join(r.root, handoffDetail))
	if err != nil {
		t.Fatal(err)
	}
	files, err := readAssetFiles(r.root, handoffDetail, "000009")
	if err != nil {
		t.Fatal(err)
	}
	bundle, err := issueassets.BuildBundle("000009", issueAssetBody(raw), files)
	if err != nil {
		t.Fatal(err)
	}
	env, err := openTracker(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	original := restoreToHead
	t.Cleanup(func() { restoreToHead = original })
	restoreToHead = func(env *trackerEnv, rel string) error {
		if rel == other {
			return os.WriteFile(filepath.Join(r.root, p), []byte("changed by earlier checkout"), 0644)
		}
		return original(env, rel)
	}
	var warnings bytes.Buffer
	finishPublished(env, &warnings, []republishItem{
		{id: "000008", path: other, read: []byte("other issue\n")},
		{id: "000009", path: handoffDetail, read: raw, publish: append(append([]byte(nil), raw...), []byte("\npublished\n")...), assets: bundle, inventory: files},
	})
	if !strings.Contains(warnings.String(), "content name") {
		t.Fatalf("missing changed-byte refusal: %s", warnings.String())
	}
	if r.git("rev-parse", "HEAD") != head {
		t.Fatal("published-copy commit accepted changed bytes")
	}
}

func TestFastForwardRestVerifiesAlreadyMaterializedAssets(t *testing.T) {
	root := testfix.Repo(t, testfix.InitialCommit())
	p := "workshop/issues/assets/000001/evidence.png"
	procedureWrite(t, root, p, "evidence")
	testfix.Git(t, root, "add", ".")
	testfix.Git(t, root, "commit", "-qm", "evidence")
	head := strings.TrimSpace(testfix.Capture(t, root, "rev-parse", "HEAD"))
	procedureWrite(t, root, p, "left by interrupted materialization")
	reader, err := gitx.NewTrunkFileContext(context.Background(), root, "origin", "main")
	if err != nil {
		t.Fatal(err)
	}
	env := &trackerEnv{ctx: context.Background(), root: root, main: reader, resting: "main"}
	if warn := fastForwardRest(env, head); !strings.Contains(warn, p) || !strings.Contains(warn, "restore") {
		t.Fatalf("no-op retry accepted changed bytes: %q", warn)
	}
}

func TestMoveDetailRemoverVerifiesRetainedAssets(t *testing.T) {
	root := testfix.Repo(t, testfix.InitialCommit())
	p := "workshop/issues/assets/000001/evidence.png"
	procedureWrite(t, root, p, "published evidence")
	testfix.Git(t, root, "add", ".")
	testfix.Git(t, root, "commit", "-qm", "published")
	published := strings.TrimSpace(testfix.Capture(t, root, "rev-parse", "HEAD"))
	procedureWrite(t, root, p, "later committed evidence")
	testfix.Git(t, root, "add", ".")
	testfix.Git(t, root, "commit", "-qm", "advanced main")
	reader, err := gitx.NewTrunkFileContext(context.Background(), root, "origin", "main")
	if err != nil {
		t.Fatal(err)
	}
	env := &trackerEnv{ctx: context.Background(), root: root, main: reader, resting: "main", branch: "main"}
	spec := tracker.ReceiptSpec{DestinationPath: "workshop/issues/000001-x.md"}
	if err := moveDetailRemover(env)(spec, published); err != nil {
		t.Fatalf("advanced valid HEAD refused: %v", err)
	}
	procedureWrite(t, root, p, "left by interrupted materialization")
	if err := moveDetailRemover(env)(spec, published); err == nil || !strings.Contains(err.Error(), p) {
		t.Fatalf("retry accepted changed bytes: %v", err)
	}
}
