package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func workingdocDigest(s string) string { return fmt.Sprintf("%x", sha256.Sum256([]byte(s))) }

func workingdocReceipt(t *testing.T, output, rel, commit string) {
	t.Helper()
	var got struct {
		SchemaVersion int    `json:"schema_version"`
		Path          string `json:"path"`
		Commit        string `json:"commit"`
	}
	if err := json.Unmarshal([]byte(output), &got); err != nil {
		t.Fatalf("receipt %q: %v", output, err)
	}
	if got.SchemaVersion != 1 || got.Path != rel || got.Commit != commit {
		t.Fatalf("wrong receipt: %+v; want %s %s", got, rel, commit)
	}
}

func TestWorkingdocTrunkPublishReceiptAndIdenticalRetry(t *testing.T) {
	r, _ := claimSetRepo(t)
	r.git("switch", "-q", "-c", "000009-s09")
	rel, payload := "workshop/workingdoc/copy.md", "approved\n"
	writeRepoFile(t, r.root, rel, payload)
	args := []string{"trunk", "publish", "--path", rel, "--expected-content-sha256", workingdocDigest(payload), "--json"}
	out, diag, err := executeSDLCTestCommand(args...)
	if err != nil {
		t.Fatalf("publish: %v %s", err, diag)
	}
	published := r.originMain()
	workingdocReceipt(t, out, rel, published)
	if got := r.git("show", published+":"+rel); got != "approved" {
		t.Fatalf("blob: %q", got)
	}
	out, diag, err = executeSDLCTestCommand(args...)
	if err != nil {
		t.Fatalf("retry: %v %s", err, diag)
	}
	workingdocReceipt(t, out, rel, published)
	if r.originMain() != published {
		t.Fatal("retry duplicated publication")
	}
}

func TestWorkingdocTrunkPinnedBaselineRetainsBranchGuard(t *testing.T) {
	for _, mode := range []string{"stale", "conflict", "staged", "committed", "nonancestor"} {
		t.Run(mode, func(t *testing.T) {
			r, _ := claimSetRepo(t)
			rel := "workshop/workingdoc/pinned.md"
			writeRepoFile(t, r.root, rel, "base\n")
			r.git("add", "--", rel)
			r.git("commit", "-q", "-m", "base")
			r.git("push", "-q", "origin", "HEAD:main")
			r.git("switch", "-q", "-c", "000009-s09")
			peer := t.TempDir()
			r.git("clone", "-q", r.origin, peer)
			writeRepoFile(t, peer, rel, "previous settlement\n")
			gitInTest(t, peer, "add", "--", rel)
			gitInTest(t, peer, "commit", "-q", "-m", "previous")
			gitInTest(t, peer, "push", "-q", "origin", "HEAD:main")
			expected := r.originMain()
			if mode == "conflict" {
				writeRepoFile(t, peer, rel, "peer conflict\n")
				gitInTest(t, peer, "commit", "-qam", "conflict")
				gitInTest(t, peer, "push", "-q", "origin", "HEAD:main")
			}
			writeRepoFile(t, r.root, rel, "next settlement\n")
			if mode == "staged" || mode == "committed" || mode == "nonancestor" {
				r.git("add", "--", rel)
			}
			if mode == "committed" || mode == "nonancestor" {
				r.git("commit", "-q", "-m", "branch edit")
				writeRepoFile(t, r.root, rel, "later\n")
			}
			if mode == "nonancestor" {
				expected = r.git("rev-parse", "HEAD")
			}
			before := r.originMain()
			out, diag, err := executeSDLCTestCommand("trunk", "publish", "--path", rel, "--expected-commit", expected, "--json")
			if mode == "stale" {
				if err != nil {
					t.Fatalf("pinned publish: %v %s", err, diag)
				}
				workingdocReceipt(t, out, rel, r.originMain())
				return
			}
			if err == nil {
				t.Fatalf("accepted %s: %s", mode, out)
			}
			if r.originMain() != before {
				t.Fatal("refusal changed remote")
			}
			want := map[string]string{"conflict": "changed on main", "staged": "staged", "committed": "committed", "nonancestor": "ancestor"}[mode]
			if !strings.Contains(err.Error(), want) {
				t.Fatalf("wrong refusal: %v; want %s", err, want)
			}
		})
	}
}

func TestWorkingdocTrunkHashRefusesBeforePublication(t *testing.T) {
	for _, hash := range []string{"abc", strings.Repeat("g", 64), workingdocDigest("older\n")} {
		t.Run(hash[:3], func(t *testing.T) {
			r, _ := claimSetRepo(t)
			rel := "workshop/workingdoc/hash.md"
			writeRepoFile(t, r.root, rel, "newer\n")
			before := r.originMain()
			out, _, err := executeSDLCTestCommand("trunk", "publish", "--path", rel, "--expected-content-sha256", hash, "--json")
			if err == nil || !strings.Contains(err.Error(), "sha256") {
				t.Fatalf("hash accepted or wrong error: %v", err)
			}
			if r.originMain() != before || out != "" {
				t.Fatalf("refusal produced publication/receipt: %q", out)
			}
			got, _ := os.ReadFile(filepath.Join(r.root, rel))
			if string(got) != "newer\n" {
				t.Fatal("local edit lost")
			}
		})
	}
}

// A real pre-push hook advances an unrelated remote path, forcing the CAS to
// retry. The hook also changes the live target after capture; neither retry nor
// receipt may substitute those newer bytes for the validated payload.
func TestWorkingdocTrunkCaptureSurvivesCASRetry(t *testing.T) {
	r, _ := claimSetRepo(t)
	rel, payload := "workshop/workingdoc/race.md", "captured\n"
	writeRepoFile(t, r.root, rel, payload)
	peer := t.TempDir()
	r.git("clone", "-q", r.origin, peer)
	writeRepoFile(t, peer, "peer.txt", "peer\n")
	gitInTest(t, peer, "add", "peer.txt")
	gitInTest(t, peer, "commit", "-q", "-m", "peer")
	r.git("config", "core.hooksPath", filepath.Join(r.root, ".git/hooks"))
	hook := "#!/bin/sh\nif test ! -f .git/raced; then\n touch .git/raced\n git -C '" + peer + "' push -q origin HEAD:main || exit 1\n printf 'newer live edit\\n' > '" + rel + "'\nfi\n"
	if err := os.WriteFile(filepath.Join(r.root, ".git/hooks/pre-push"), []byte(hook), 0755); err != nil {
		t.Fatal(err)
	}
	out, diag, err := executeSDLCTestCommand("trunk", "publish", "--path", rel, "--expected-content-sha256", workingdocDigest(payload), "--json")
	if err != nil {
		t.Fatalf("publish: %v %s", err, diag)
	}
	workingdocReceipt(t, out, rel, r.originMain())
	if got := r.git("show", r.originMain()+":"+rel); got != "captured" {
		t.Fatalf("remote blob %q", got)
	}
	if got := r.git("show", r.originMain()+":peer.txt"); got != "peer" {
		t.Fatal("retry lost peer")
	}
	got, _ := os.ReadFile(filepath.Join(r.root, rel))
	if string(got) != "newer live edit\n" {
		t.Fatalf("live edit %q", got)
	}
}

func TestWorkingdocTrunkMutationBeforeCaptureRefuses(t *testing.T) {
	r, _ := claimSetRepo(t)
	rel := "workshop/workingdoc/before.md"
	writeRepoFile(t, r.root, rel, "approved\n")
	old := trunkProseOpen
	defer func() { trunkProseOpen = old }()
	trunkProseOpen = func(name string) (io.ReadCloser, error) {
		if err := os.WriteFile(name, []byte("changed\n"), 0644); err != nil {
			return nil, err
		}
		return os.Open(name)
	}
	before := r.originMain()
	_, _, err := executeSDLCTestCommand("trunk", "publish", "--path", rel, "--expected-content-sha256", workingdocDigest("approved\n"), "--json")
	if err == nil || !strings.Contains(err.Error(), "sha256") || r.originMain() != before {
		t.Fatalf("capture mismatch: %v", err)
	}
}

func TestWorkingdocTrunkReceiptRequiresReadback(t *testing.T) {
	for _, mode := range []string{"interrupted", "changed", "clean"} {
		t.Run(mode, func(t *testing.T) {
			r, _ := claimSetRepo(t)
			rel := "workshop/workingdoc/readback.md"
			writeRepoFile(t, r.root, rel, "approved\n")
			if mode == "clean" {
				r.git("add", "--", rel)
				r.git("commit", "-q", "-m", "clean approved")
				r.git("push", "-q", "origin", "HEAD:main")
			}
			old := trunkPublishAfterPush
			defer func() { trunkPublishAfterPush = old }()
			trunkPublishAfterPush = func() error {
				if mode == "interrupted" {
					return fmt.Errorf("lost acknowledgment")
				}
				if mode == "changed" {
					peer := t.TempDir()
					r.git("clone", "-q", r.origin, peer)
					writeRepoFile(t, peer, rel, "other\n")
					gitInTest(t, peer, "commit", "-qam", "other")
					gitInTest(t, peer, "push", "-q", "origin", "HEAD:main")
				}
				return nil
			}
			out, _, err := executeSDLCTestCommand("trunk", "publish", "--path", rel, "--json")
			if mode == "clean" {
				if err != nil {
					t.Fatal(err)
				}
				workingdocReceipt(t, out, rel, r.originMain())
				return
			}
			if err == nil || out != "" {
				t.Fatalf("unconfirmed receipt: %q %v", out, err)
			}
			if mode == "interrupted" {
				trunkPublishAfterPush = old
				published := r.originMain()
				out, _, err = executeSDLCTestCommand("trunk", "publish", "--path", rel, "--json")
				if err != nil {
					t.Fatal(err)
				}
				workingdocReceipt(t, out, rel, published)
				if r.originMain() != published {
					t.Fatal("retry duplicated uncertain publication")
				}
			}
		})
	}
}

// Lose the push acknowledgment after the real remote accepted the commit. The
// current invocation must refuse a receipt; the next invocation reconciles bytes.
func TestWorkingdocTrunkUnknownPushRetry(t *testing.T) {
	r, _ := claimSetRepo(t)
	rel := "workshop/workingdoc/unknown.md"
	writeRepoFile(t, r.root, rel, "approved\n")
	realGit, err := exec.LookPath("git")
	if err != nil {
		t.Fatal(err)
	}
	shimDir := t.TempDir()
	marker := filepath.Join(shimDir, "lost")
	script := "#!/bin/sh\nif test \"$1\" = push && test ! -f '" + marker + "'; then\n '" + realGit + "' \"$@\" || exit $?\n touch '" + marker + "'\n echo 'fatal: connection reset after write' >&2\n exit 1\nfi\nexec '" + realGit + "' \"$@\"\n"
	if err := os.WriteFile(filepath.Join(shimDir, "git"), []byte(script), 0755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", shimDir+string(os.PathListSeparator)+os.Getenv("PATH"))
	out, _, err := executeSDLCTestCommand("trunk", "publish", "--path", rel, "--json")
	if err == nil || out != "" {
		t.Fatalf("unknown push receipt: %q %v", out, err)
	}
	published := r.originMain()
	if got := r.git("show", published+":"+rel); got != "approved" {
		t.Fatalf("unknown outcome did not publish: %q", got)
	}
	out, diag, err := executeSDLCTestCommand("trunk", "publish", "--path", rel, "--json")
	if err != nil {
		t.Fatalf("retry: %v %s", err, diag)
	}
	workingdocReceipt(t, out, rel, published)
	if r.originMain() != published {
		t.Fatal("unknown push retry duplicated commit")
	}
}

func TestWorkingdocTrunkExpectedCommitRejectsTagObject(t *testing.T) {
	r, _ := claimSetRepo(t)
	r.git("tag", "-a", "baseline", "-m", "annotated tag")
	tag := r.git("rev-parse", "baseline")
	rel := "workshop/workingdoc/tag.md"
	writeRepoFile(t, r.root, rel, "approved\n")
	before := r.originMain()
	out, _, err := executeSDLCTestCommand("trunk", "publish", "--path", rel, "--expected-commit", tag, "--json")
	if err == nil || !strings.Contains(err.Error(), "commit object") {
		t.Fatalf("tag accepted/wrong error: %q %v", out, err)
	}
	if r.originMain() != before || out != "" {
		t.Fatal("tag precondition produced publication")
	}
}

func TestWorkingdocTrunkPinnedAncestorRecheckedOnRetry(t *testing.T) {
	r, _ := claimSetRepo(t)
	rel := "workshop/workingdoc/rewritten.md"
	writeRepoFile(t, r.root, rel, "base\n")
	r.git("add", "--", rel)
	r.git("commit", "-q", "-m", "base")
	r.git("push", "-q", "origin", "HEAD:main")
	expected := r.originMain()
	writeRepoFile(t, r.root, rel, "approved\n")
	peer := t.TempDir()
	r.git("clone", "-q", r.origin, peer)
	gitInTest(t, peer, "switch", "--orphan", "rewritten")
	writeRepoFile(t, peer, rel, "base\n")
	gitInTest(t, peer, "add", "--", rel)
	gitInTest(t, peer, "commit", "-q", "-m", "rewrite")
	rewritten := gitInTest(t, peer, "rev-parse", "HEAD")
	r.git("config", "core.hooksPath", filepath.Join(r.root, ".git/hooks"))
	hook := "#!/bin/sh\nif test ! -f .git/rewritten; then\n touch .git/rewritten\n git -C '" + peer + "' push -q --force origin HEAD:main || exit 1\nfi\n"
	if err := os.WriteFile(filepath.Join(r.root, ".git/hooks/pre-push"), []byte(hook), 0755); err != nil {
		t.Fatal(err)
	}
	out, _, err := executeSDLCTestCommand("trunk", "publish", "--path", rel, "--expected-commit", expected, "--json")
	if err == nil || !strings.Contains(err.Error(), "ancestor") {
		t.Fatalf("retry accepted rewritten history: %q %v", out, err)
	}
	if r.originMain() != rewritten || out != "" {
		t.Fatal("retry wrote beyond rewritten remote")
	}
}

func TestWorkingdocTrunkFirstPublicationOwnership(t *testing.T) {
	for _, mode := range []string{"absent", "identical", "different", "race", "retry-base"} {
		t.Run(mode, func(t *testing.T) {
			r, _ := claimSetRepo(t)
			rel, payload := "workshop/workingdoc/first.md", "approved\n"
			if mode == "identical" || mode == "different" || mode == "retry-base" {
				original := payload
				if mode != "identical" {
					original = "unrelated committed design\n"
				}
				writeRepoFile(t, r.root, rel, original)
				r.git("add", "--", rel)
				r.git("commit", "-qm", "existing destination")
				r.git("push", "-q", "origin", "HEAD:main")
			}
			r.git("switch", "-q", "-c", "000009-s09")
			writeRepoFile(t, r.root, rel, payload)
			before := r.originMain()
			if mode == "race" || mode == "retry-base" {
				peer := t.TempDir()
				r.git("clone", "-q", r.origin, peer)
				if mode == "retry-base" {
					gitInTest(t, peer, "rm", "--", rel)
					gitInTest(t, peer, "commit", "-qm", "destination removed before snapshot")
					gitInTest(t, peer, "push", "-q", "origin", "HEAD:main")
				}
				racing := "racing owner's content\n"
				if mode == "retry-base" {
					racing = "unrelated committed design\n"
				}
				writeRepoFile(t, peer, rel, racing)
				gitInTest(t, peer, "add", "--", rel)
				gitInTest(t, peer, "commit", "-qm", "racing destination")
				r.git("config", "core.hooksPath", filepath.Join(r.root, ".git/hooks"))
				hook := "#!/bin/sh\nif test ! -f .git/first-raced; then\n touch .git/first-raced\n git -C '" + peer + "' push -q origin HEAD:main || exit 1\nfi\n"
				if err := os.WriteFile(filepath.Join(r.root, ".git/hooks/pre-push"), []byte(hook), 0755); err != nil {
					t.Fatal(err)
				}
			}
			out, diag, err := executeSDLCTestCommand("trunk", "publish", "--path", rel, "--expect-absent", "--json")
			if mode == "different" || mode == "race" || mode == "retry-base" {
				if err == nil || !strings.Contains(err.Error(), "first publication conflicts") {
					t.Fatalf("missing ownership refusal: %v %s %s", err, out, diag)
				}
				if mode == "different" && r.originMain() != before {
					t.Fatal("refusal changed remote")
				}
				want := "unrelated committed design"
				if mode == "race" {
					want = "racing owner's content"
				}
				if got := r.git("show", r.originMain()+":"+rel); got != want {
					t.Fatalf("overwrote committed owner: %q", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("safe first publication: %v %s", err, diag)
			}
			workingdocReceipt(t, out, rel, r.originMain())
			if mode == "identical" && r.originMain() != before {
				t.Fatal("identical retry duplicated commit")
			}
		})
	}
}
