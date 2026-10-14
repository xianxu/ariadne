package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/issueassets"
)

func migrationAssetFixture(t *testing.T, conversion bool) (*trackerRepo, string) {
	t.Helper()
	r := legacyRepo(t)
	detail := "workshop/history/issues/000003-gone.md"
	data := "line one\nline two\n"
	name, _ := issueassets.Name([]byte(data), "png")
	asset := "workshop/history/issues/assets/000003/" + name
	if !conversion {
		appendDetail(t, r, detail, "![evidence](assets/000003/"+name+")")
		writeRepoFile(t, r.root, asset, data)
		r.git("add", "workshop/history")
		r.git("commit", "-qm", "existing evidence")
		r.git("push", "-q", "origin", "main")
	}
	r.git("switch", "-q", "-c", "000002-two")
	writeRepoFile(t, r.root, "branch.txt", "branch\n")
	r.git("add", "branch.txt")
	r.git("commit", "-qm", "branch work")
	if !conversion {
		writeRepoFile(t, r.root, asset, "tampered under original content name\n")
		r.git("commit", "-qam", "changed evidence")
	}
	r.git("switch", "-q", "main")
	if conversion {
		appendDetail(t, r, detail, "![evidence](assets/000003/"+name+")")
		writeRepoFile(t, r.root, asset, data)
		writeRepoFile(t, r.root, ".gitattributes", "workshop/history/issues/assets/** text eol=crlf\n")
		r.git("add", "workshop/history", ".gitattributes")
		r.git("commit", "-qm", "new archived evidence with conversion")
		r.git("push", "-q", "origin", "main")
	}
	digest, out, err := migrateDryRun(t)
	if err != nil {
		t.Fatalf("migration plan: %v\n%s", err, out)
	}
	if _, stderr, err := executeSDLCTestCommand("issue", "migrate", "--apply", "--expect", digest); err != nil {
		t.Fatalf("migration apply: %v\n%s", err, stderr)
	}
	r.git("switch", "-q", "000002-two")
	return r, asset
}

func TestAssetMigrateReconcileRefusesConversionBeforeMerge(t *testing.T) {
	r, asset := migrationAssetFixture(t, true)
	head := r.git("rev-parse", "HEAD")
	_, stderr, err := executeSDLCTestCommand("issue", "migrate", "--reconcile")
	if err == nil || !strings.Contains(err.Error(), "convert asset") {
		t.Fatalf("cutover materialized converting asset: %v\n%s", err, stderr)
	}
	if r.git("rev-parse", "HEAD") != head || r.git("status", "--porcelain") != "" {
		t.Fatal("conversion refusal changed checkout")
	}
	if _, err := os.Stat(filepath.Join(r.root, asset)); !os.IsNotExist(err) {
		t.Fatalf("conversion refusal materialized asset: %v", err)
	}
}

func TestAssetMigrateReconcileRefusesInvalidIndexedBundleBeforeCommit(t *testing.T) {
	r, asset := migrationAssetFixture(t, false)
	head := r.git("rev-parse", "HEAD")
	before, err := os.ReadFile(filepath.Join(r.root, asset))
	if err != nil {
		t.Fatal(err)
	}
	_, stderr, err := executeSDLCTestCommand("issue", "migrate", "--reconcile")
	if err == nil || !strings.Contains(err.Error(), "content name") {
		t.Fatalf("cutover committed invalid asset bundle: %v\n%s", err, stderr)
	}
	if r.git("rev-parse", "HEAD") != head {
		t.Fatal("refusal committed the merged invalid evidence")
	}
	after, err := os.ReadFile(filepath.Join(r.root, asset))
	if err != nil || string(after) != string(before) {
		t.Fatalf("refusal lost branch bytes: %v", err)
	}
}
