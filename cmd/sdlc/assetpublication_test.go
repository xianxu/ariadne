package main

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/issueassets"
)

func assetFixture(t *testing.T, r *trackerRepo, id, detail string) string {
	t.Helper()
	peerAdd(t, r, "workshop/asset-policy.json", `{"version":1,"policy":"allowed"}`)
	r.git("fetch", "-q", "origin")
	r.git("merge", "-q", "--no-edit", "origin/main")
	data := []byte("image\x00bytes")
	name, err := issueassets.Name(data, "png")
	if err != nil {
		t.Fatal(err)
	}
	rel := "assets/" + id + "/" + name
	writeRepoFile(t, r.root, filepath.ToSlash(filepath.Join(filepath.Dir(detail), rel)), string(data))
	appendDetail(t, r, detail, "![evidence]("+rel+")")
	return filepath.ToSlash(filepath.Join(filepath.Dir(detail), rel))
}

func TestAssetPublishBundleFromRest(t *testing.T) {
	r, _ := claimSetRepo(t)
	peerAdd(t, r, "workshop/asset-policy.json", `{"version":1,"policy":"allowed"}`)
	r.git("fetch", "-q", "origin")
	r.git("merge", "-q", "--no-edit", "origin/main")
	data := []byte("image\x00bytes")
	source := filepath.Join(t.TempDir(), "evidence.png")
	if err := os.WriteFile(source, data, 0600); err != nil {
		t.Fatal(err)
	}
	if _, stderr, err := executeSDLCTestCommand("issue", "attach", "--issue", "9", "--file", source, "--caption", "evidence"); err != nil {
		t.Fatalf("attach: %v\n%s", err, stderr)
	}
	name, err := issueassets.Name(data, "png")
	if err != nil {
		t.Fatal(err)
	}
	p := "workshop/issues/assets/000009/" + name
	if out, err := publish(t, 9); err != nil {
		t.Fatalf("%v\n%s", err, out)
	}
	if got := r.git("show", r.originMain()+":"+p); got != "image\x00bytes" {
		t.Fatalf("asset = %q", got)
	}
	if got := r.git("status", "--porcelain"); got != "" {
		t.Fatalf("checkout dirty: %s", got)
	}
}

func TestAssetPublishRejectsChangedBytesWithoutBodyEdit(t *testing.T) {
	r, _ := claimSetRepo(t)
	p := assetFixture(t, r, "000009", "workshop/issues/000009-s09.md")
	r.git("add", "workshop/issues")
	r.git("commit", "-qm", "evidence")
	r.git("push", "-q", "origin", "HEAD:main")
	writeRepoFile(t, r.root, p, "tampered bytes")
	before := r.originMain()
	if _, err := publish(t, 9); err == nil || !strings.Contains(err.Error(), "content name") {
		t.Fatalf("tampered bytes accepted: %v", err)
	}
	if r.originMain() != before {
		t.Fatal("refusal changed main")
	}
}

func TestAssetTransferGuardAndRestoreBinaryOnly(t *testing.T) {
	r := handedOff(t)
	r.git("fetch", "-q", "origin")
	r.git("merge", "-q", "--no-edit", "origin/main")
	p := assetFixture(t, r, "000008", spinOffDetails)
	r.git("add", "workshop/issues")
	r.git("commit", "-qm", "evidence")
	r.git("push", "-q", "origin", "HEAD:main")
	ownSpinOff(t, r, otherSlot)
	writeRepoFile(t, r.root, p, "tampered bytes")
	r.git("commit", "-qam", "binary only")
	if err := guardTransferredDetails(context.Background()); err == nil || !strings.Contains(err.Error(), "is owned by") {
		t.Fatalf("asset ownership escaped: %v", err)
	}
	if err := runIssueRestore(context.Background(), os.Stdout, os.Stderr, []int{8}); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(filepath.Join(r.root, p))
	if err != nil || string(got) != "image\x00bytes" {
		t.Fatalf("restore %q: %v", got, err)
	}
}

func TestAssetFirstPublicationCarriesAndRelinquishesBundle(t *testing.T) {
	r := spinOffOnCodeBranch(t)
	p := assetFixture(t, r, "000008", spinOffDetails)
	r.git("add", "workshop/issues")
	r.git("commit", "-qm", "attachment")
	if out, err := publish(t, 8); err != nil {
		t.Fatalf("%v\n%s", err, out)
	}
	if got := r.git("show", r.originMain()+":"+p); got != "image\x00bytes" {
		t.Fatalf("main asset = %q", got)
	}
	if _, err := os.Stat(filepath.Join(r.root, p)); !os.IsNotExist(err) {
		t.Fatalf("source asset not relinquished: %v", err)
	}
}

func TestAssetLegacyCommitPublication(t *testing.T) {
	r, _ := claimSetRepo(t)
	r.git("switch", "-q", "-c", "evidence")
	p := assetFixture(t, r, "000009", "workshop/issues/000009-s09.md")
	r.git("add", "workshop/issues")
	r.git("commit", "-qm", "attachment")
	if err := publishIssueCommit(os.Stdout, os.Stderr, r.root, "HEAD", "workshop/issues", "workshop/history"); err != nil {
		t.Fatal(err)
	}
	if got := r.git("show", r.originMain()+":"+p); got != "image\x00bytes" {
		t.Fatalf("asset = %q", got)
	}
}

func TestAssetPublishRemovesUnlinkedMember(t *testing.T) {
	r, _ := claimSetRepo(t)
	detail := "workshop/issues/000009-s09.md"
	p := assetFixture(t, r, "000009", detail)
	r.git("add", "workshop/issues")
	r.git("commit", "-qm", "evidence")
	r.git("push", "-q", "origin", "HEAD:main")
	raw := r.git("show", "HEAD:"+detail)
	writeRepoFile(t, r.root, detail, strings.Split(raw, "![evidence]")[0])
	if out, err := publish(t, 9); err != nil {
		t.Fatalf("%v\n%s", err, out)
	}
	if strings.Contains(r.git("ls-tree", "-r", "--name-only", r.originMain()), p) {
		t.Fatal("unlinked asset retained on main")
	}
	if _, err := os.Stat(filepath.Join(r.root, p)); !os.IsNotExist(err) {
		t.Fatalf("local asset retained: %v", err)
	}
}

func TestAssetFirstPublicationRecoveryKeepsPinnedBytes(t *testing.T) {
	r := spinOffOnCodeBranch(t)
	p := assetFixture(t, r, "000008", spinOffDetails)
	r.git("add", "workshop/issues")
	r.git("commit", "-qm", "attachment")
	block := blockTrackerAfterFirstPush(t, r)
	if _, err := publish(t, 8); err == nil {
		t.Fatal("fixture did not interrupt")
	}
	writeRepoFile(t, r.root, p, "later bytes")
	if err := os.Remove(block); err != nil {
		t.Fatal(err)
	}
	if err := runRecoveryReconcile(context.Background(), os.Stdout, os.Stderr, 8); err == nil || !strings.Contains(err.Error(), "changed after it was published") {
		t.Fatalf("changed asset lost: %v", err)
	}
	if got := r.git("show", r.originMain()+":"+p); got != "image\x00bytes" {
		t.Fatalf("pinned bytes = %q", got)
	}
	if _, err := os.Stat(filepath.Join(r.root, spinOffDetails)); err != nil {
		t.Fatal("removed body before validating assets")
	}
	writeRepoFile(t, r.root, p, "image\x00bytes")
	if err := runRecoveryReconcile(context.Background(), os.Stdout, os.Stderr, 8); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(r.root, p)); !os.IsNotExist(err) {
		t.Fatalf("asset remains after reconcile: %v", err)
	}
}

func TestAssetPublishLostResponseFinishesWholeBundle(t *testing.T) {
	r, _ := claimSetRepo(t)
	p := assetFixture(t, r, "000009", "workshop/issues/000009-s09.md")
	env, err := openTracker(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	var out, errs bytes.Buffer
	if err := republishOwnedLosing(t, env, &out, &errs); err == nil {
		t.Fatal("missing lost response")
	}
	if _, err := publish(t, 9); err != nil {
		t.Fatal(err)
	}
	if got := r.git("show", r.originMain()+":"+p); got != "image\x00bytes" {
		t.Fatalf("asset = %q", got)
	}
	if r.git("status", "--porcelain") != "" {
		t.Fatal("retry left local bundle dirty")
	}
}

func TestAssetPublishMainMovedMergesWholeBundle(t *testing.T) {
	r, _ := claimSetRepo(t)
	p := assetFixture(t, r, "000009", "workshop/issues/000009-s09.md")
	peerAdd(t, r, p, "image\x00bytes")
	peerAdd(t, r, "workshop/issues/000009-s09.md", strings.Replace(r.git("show", "HEAD:workshop/issues/000009-s09.md")+"\n", "## Problem", "## Problem\n\nMain changed independently. ![main]("+strings.TrimPrefix(p, "workshop/issues/")+")", 1))
	if out, err := publish(t, 9); err != nil {
		t.Fatalf("%v\n%s", err, out)
	}
	if got := r.git("show", r.originMain()+":"+p); got != "image\x00bytes" {
		t.Fatalf("asset = %q", got)
	}
}

func TestAssetTransferGuardPolicy(t *testing.T) {
	r, _ := claimSetRepo(t)
	r.git("switch", "-q", "-c", "000009-s09")
	assetFixture(t, r, "000009", "workshop/issues/000009-s09.md")
	r.git("add", "workshop/issues")
	r.git("commit", "-qm", "attachment")
	writeRepoFile(t, r.root, "workshop/asset-policy.json", `{"version":1,"policy":"text-only"}`)
	if err := guardTransferredDetails(context.Background()); err == nil || !strings.Contains(err.Error(), "text-only") {
		t.Fatalf("privacy gate escaped: %v", err)
	}
}

func TestAssetLegacyGuardPolicy(t *testing.T) {
	r := legacyRepo(t)
	r.git("switch", "-q", "-c", "evidence")
	assetFixture(t, r, "000001", "workshop/issues/000001-one.md")
	r.git("add", "workshop/issues")
	r.git("commit", "-qm", "attachment")
	writeRepoFile(t, r.root, "workshop/asset-policy.json", `{"version":1,"policy":"text-only"}`)
	if err := guardTransferredDetails(context.Background()); err == nil || !strings.Contains(err.Error(), "text-only") {
		t.Fatalf("legacy privacy gate escaped: %v", err)
	}
}

func TestAssetPublicationRejectsDirectorySymlink(t *testing.T) {
	for _, rel := range []string{"workshop/issues/assets", "workshop/issues/assets/000009"} {
		t.Run(filepath.Base(rel), func(t *testing.T) {
			r, _ := claimSetRepo(t)
			r.git("switch", "-q", "-c", "evidence")
			if err := os.MkdirAll(filepath.Dir(filepath.Join(r.root, rel)), 0755); err != nil {
				t.Fatal(err)
			}
			if err := os.Symlink(t.TempDir(), filepath.Join(r.root, rel)); err != nil {
				t.Fatal(err)
			}
			r.git("add", "--", rel)
			r.git("commit", "-qm", "unsafe directory")
			if err := guardTransferredDetails(context.Background()); err == nil {
				t.Fatal("symlink asset directory escaped publication validation")
			}
		})
	}
}
