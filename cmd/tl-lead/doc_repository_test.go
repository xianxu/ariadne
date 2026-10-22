package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/xianxu/ariadne/pkg/procgroup"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"
)

func docRepoGit(t *testing.T, root string, args ...string) string {
	t.Helper()
	c := exec.Command("git", append([]string{"-C", root}, args...)...)
	c.Env = append(os.Environ(), "GIT_CONFIG_NOSYSTEM=1", "GIT_CONFIG_GLOBAL=/dev/null")
	out, err := c.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v %s", args, err, out)
	}
	return strings.TrimSpace(string(out))
}
func docRepoFixture(t *testing.T) (string, string) {
	t.Helper()
	root := t.TempDir()
	docRepoGit(t, root, "init", "-q", "-b", "main")
	docRepoGit(t, root, "config", "user.name", "Test")
	docRepoGit(t, root, "config", "user.email", "test@example.com")
	docRepoGit(t, root, "config", "commit.gpgsign", "false")
	docRepoGit(t, root, "remote", "add", "origin", "https://github.com/xianxu/product.git")
	for p, s := range map[string]string{"workshop/workingdoc/one.md": "first\r\n\x00tail\n", "workshop/workingdoc/two.md": "second\n"} {
		if err := os.MkdirAll(filepath.Dir(filepath.Join(root, p)), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(root, p), []byte(s), 0644); err != nil {
			t.Fatal(err)
		}
	}
	docRepoGit(t, root, "add", ".")
	docRepoGit(t, root, "commit", "-qm", "seed")
	return root, docRepoGit(t, root, "rev-parse", "HEAD")
}
func docRepoWorkspaceBinary(t *testing.T, root string, mutate func(map[string]any)) string {
	t.Helper()
	id := map[string]any{"schema_version": 2, "repo": "product", "repo_identity": docRepoGit(t, root, "rev-parse", "--path-format=absolute", "--git-common-dir"), "primary_root": root, "worktree_root": root, "kind": "primary", "slot": 0, "branch": "main", "address": "product:0"}
	if mutate != nil {
		mutate(id)
	}
	raw, _ := json.Marshal(id)
	dir := t.TempDir()
	bin := filepath.Join(dir, "sdlc")
	script := "#!/bin/sh\nif test \"$1\" = workspace; then\n cat '" + filepath.Join(dir, "identity.json") + "'\nelse\n printf '%s\\n' \"$@\" > '" + filepath.Join(dir, "argv") + "'\n cat '" + filepath.Join(dir, "receipt.json") + "'\nfi\n"
	if err := os.WriteFile(bin, []byte(script), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "identity.json"), raw, 0644); err != nil {
		t.Fatal(err)
	}
	return bin
}
func TestDocRepositoryModeIdentity(t *testing.T) {
	for _, tc := range []struct {
		name, origin, branch string
		mutate               func(map[string]any)
		brain, fail          bool
		want                 docPublicationMode
	}{
		{name: "product", want: docProduct}, {name: "ops", origin: "git@github.com:xianxu/ops.git", want: docOperating},
		{name: "other owner ops is product", origin: "ssh://git@github.com/other/ops.git", want: docProduct},
		{name: "foreign host", origin: "https://example.com/xianxu/ops.git", fail: true},
		{name: "ops feature", origin: "https://github.com/xianxu/ops.git", branch: "feature", fail: true},
		{name: "wrong common", mutate: func(m map[string]any) { m["repo_identity"] = "/unknown" }, fail: true},
		{name: "wrong root", mutate: func(m map[string]any) { m["worktree_root"] = "/unknown" }, fail: true},
		{name: "wrong primary", mutate: func(m map[string]any) { m["primary_root"] = "/unknown" }, fail: true},
		{name: "wrong schema", mutate: func(m map[string]any) { m["schema_version"] = 9 }, fail: true},
		{name: "ops slot", origin: "https://github.com/xianxu/ops.git", mutate: func(m map[string]any) { m["slot"] = 1; m["kind"] = "slot" }, fail: true},
		{name: "brain", brain: true, fail: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root, _ := docRepoFixture(t)
			if tc.origin != "" {
				docRepoGit(t, root, "remote", "set-url", "origin", tc.origin)
			}
			if tc.branch != "" {
				docRepoGit(t, root, "switch", "-qc", tc.branch)
			}
			if tc.brain {
				os.MkdirAll(filepath.Join(root, ".brain"), 0755)
				os.WriteFile(filepath.Join(root, ".brain/config.md"), []byte("brain"), 0644)
			}
			bin := docRepoWorkspaceBinary(t, root, tc.mutate)
			got, err := (execDocRepository{sdlcBin: bin}).Mode(context.Background(), root)
			if tc.fail {
				if err == nil {
					t.Fatalf("accepted mode %s", got)
				}
				return
			}
			if err != nil || got != tc.want {
				t.Fatalf("mode %s: %v, want %s", got, err, tc.want)
			}
		})
	}
}
func TestDocRepositoryCommonDirAndTarget(t *testing.T) {
	root, _ := docRepoFixture(t)
	r := execDocRepository{}
	common, err := r.CommonDir(context.Background(), root)
	if err != nil {
		t.Fatal(err)
	}
	linked := filepath.Join(t.TempDir(), "linked")
	docRepoGit(t, root, "worktree", "add", "-qb", "linked", linked)
	got, err := r.CommonDir(context.Background(), linked)
	if err != nil || got != common {
		t.Fatalf("shared common dir %q %v", got, err)
	}
	os.WriteFile(filepath.Join(root, ".gitignore"), []byte("ignored/\n"), 0644)
	for _, p := range []string{"ignored/copy.md", "../outside.md", "workshop/workingdoc/one.md/child.md"} {
		if err := r.ValidateTarget(context.Background(), root, p); err == nil {
			t.Fatalf("accepted target %s", p)
		}
	}
	if err := r.ValidateTarget(context.Background(), root, "workshop/workingdoc/new.md"); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(t.TempDir(), filepath.Join(root, "link")); err != nil {
		t.Fatal(err)
	}
	if err := r.ValidateTarget(context.Background(), root, "link/file.md"); err == nil {
		t.Fatal("accepted symlink")
	}
}
func TestDocRepositoryBaselinesRawAndBatched(t *testing.T) {
	root, commit := docRepoFixture(t)
	realGit, err := exec.LookPath("git")
	if err != nil {
		t.Fatal(err)
	}
	shim := t.TempDir()
	count := filepath.Join(shim, "calls")
	script := "#!/bin/sh\nprintf '%s\\n' \"$*\" >> '" + count + "'\nexec '" + realGit + "' \"$@\"\n"
	os.WriteFile(filepath.Join(shim, "git"), []byte(script), 0755)
	t.Setenv("PATH", shim+string(os.PathListSeparator)+os.Getenv("PATH"))
	for i := 0; i < 24; i++ {
		p := fmt.Sprintf("workshop/workingdoc/copy%02d.md", i)
		body := fmt.Sprintf("raw %d\r\n\x00tail\n", i)
		if err := os.WriteFile(filepath.Join(root, p), []byte(body), 0644); err != nil {
			t.Fatal(err)
		}
	}
	docRepoGit(t, root, "add", ".")
	docRepoGit(t, root, "commit", "-qm", "distinct baselines")
	commit = docRepoGit(t, root, "rev-parse", "HEAD")
	for _, n := range []int{12, 24} {
		os.WriteFile(count, nil, 0644)
		var pointers []settlementPointer
		for i := 0; i < n; i++ {
			pointers = append(pointers, settlementPointer{fmt.Sprintf("workshop/workingdoc/copy%02d.md", i), commit})
		}
		got, err := (execDocRepository{}).ReadBaselines(context.Background(), root, pointers)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(got[settlementPointer{"workshop/workingdoc/copy00.md", commit}], []byte("raw 0\r\n\x00tail\n")) {
			t.Fatal("raw blob changed")
		}
		calls, _ := os.ReadFile(count)
		lines := strings.Split(strings.TrimSpace(string(calls)), "\n")
		if len(lines) != 1 || !strings.Contains(lines[0], "cat-file --batch") {
			t.Fatalf("%d paths did not use one batch: %s", n, calls)
		}
	}
}
func TestDocRepositoryBaselineRefusals(t *testing.T) {
	root, commit := docRepoFixture(t)
	docRepoGit(t, root, "tag", "-a", "tag", "-m", "tag")
	tag := docRepoGit(t, root, "rev-parse", "tag")
	for _, p := range []settlementPointer{{"workshop/workingdoc/missing.md", commit}, {"workshop/workingdoc/one.md", tag}, {"workshop/workingdoc/one.md", strings.Repeat("0", 40)}, {"../bad.md", commit}} {
		if _, err := (execDocRepository{}).ReadBaselines(context.Background(), root, []settlementPointer{p}); err == nil {
			t.Fatalf("accepted unavailable baseline %+v", p)
		}
	}
	big := filepath.Join(root, "workshop/workingdoc/big.md")
	os.WriteFile(big, bytes.Repeat([]byte("x"), (1<<20)+1), 0644)
	docRepoGit(t, root, "add", ".")
	docRepoGit(t, root, "commit", "-qm", "big")
	commit = docRepoGit(t, root, "rev-parse", "HEAD")
	if _, err := (execDocRepository{}).ReadBaselines(context.Background(), root, []settlementPointer{{"workshop/workingdoc/big.md", commit}}); err == nil {
		t.Fatal("accepted oversized baseline")
	}
}
func TestDocRepositoryProductReceipt(t *testing.T) {
	root, commit := docRepoFixture(t)
	bin := docRepoWorkspaceBinary(t, root, nil)
	target := "workshop/workingdoc/two.md"
	payload := []byte("second\n")
	for _, tc := range []struct {
		name, receipt string
		fail          bool
	}{
		{"valid", fmt.Sprintf(`{"schema_version":1,"path":%q,"commit":%q}`, target, commit), false},
		{"wrong path", fmt.Sprintf(`{"schema_version":1,"path":"other.md","commit":%q}`, commit), true},
		{"wrong version", fmt.Sprintf(`{"schema_version":9,"path":%q,"commit":%q}`, target, commit), true},
		{"short sha", fmt.Sprintf(`{"schema_version":1,"path":%q,"commit":"abc"}`, target), true},
		{"duplicate", fmt.Sprintf(`{"schema_version":1,"path":%q,"path":%q,"commit":%q}`, target, target, commit), true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			os.WriteFile(filepath.Join(filepath.Dir(bin), "receipt.json"), []byte(tc.receipt), 0644)
			got, err := (execDocRepository{sdlcBin: bin}).Publish(context.Background(), root, target, payload, commit)
			if tc.fail {
				if err == nil {
					t.Fatalf("accepted receipt %s", got)
				}
				return
			}
			if err != nil || got != commit {
				t.Fatalf("publish %s: %v", got, err)
			}
			argv, _ := os.ReadFile(filepath.Join(filepath.Dir(bin), "argv"))
			hash := fmt.Sprintf("%x", sha256.Sum256(payload))
			want := strings.Join([]string{"trunk", "publish", "--path", target, "--json", "--expected-content-sha256", hash, "--expected-commit", commit}, "\n") + "\n"
			if string(argv) != want {
				t.Fatalf("argv %q want %q", argv, want)
			}
		})
	}
	os.WriteFile(filepath.Join(filepath.Dir(bin), "receipt.json"), []byte(fmt.Sprintf(`{"schema_version":1,"path":%q,"commit":%q}`, target, commit)), 0644)
	if _, err := (execDocRepository{sdlcBin: bin}).Publish(context.Background(), root, target, payload, ""); err != nil {
		t.Fatalf("first publication receipt: %v", err)
	}
	argv, _ := os.ReadFile(filepath.Join(filepath.Dir(bin), "argv"))
	if !strings.Contains(string(argv), "\n--expect-absent\n") || strings.Contains(string(argv), "--expected-commit") {
		t.Fatalf("first publication lacks remote ownership guard: %q", argv)
	}
	if _, err := (execDocRepository{sdlcBin: bin}).Publish(context.Background(), root, target, []byte("wrong\n"), ""); err == nil {
		t.Fatal("accepted mismatching committed blob")
	}
}
func TestDocRepositoryCancellation(t *testing.T) {
	root, _ := docRepoFixture(t)
	dir := t.TempDir()
	ready := filepath.Join(dir, "ready")
	openedPath, releasePath := filepath.Join(dir, "opened"), filepath.Join(dir, "release")
	for _, p := range []string{openedPath, releasePath} {
		if err := syscall.Mkfifo(p, 0600); err != nil {
			t.Fatal(err)
		}
	}
	opened, err := os.OpenFile(openedPath, os.O_RDWR, 0600)
	if err != nil {
		t.Fatal(err)
	}
	defer opened.Close()
	release, err := os.OpenFile(releasePath, os.O_RDWR, 0600)
	if err != nil {
		t.Fatal(err)
	}
	defer release.Close()
	bin := filepath.Join(dir, "sdlc")
	// Atomic PID readiness: write a sibling staging file, close, then rename.
	// The FIFO barrier holds output open before the PID write, making the
	// create-before-content race deterministic instead of relying on load (#552).
	script := fmt.Sprintf("#!/bin/sh\nsleep 60 &\nexec 3> %q\nprintf r > %q\nread permit < %q\nprintf '%%s\\n' \"$$\" >&3\nexec 3>&-\nmv %q %q\nwait\n", ready+".tmp", openedPath, releasePath, ready+".tmp", ready)
	if err := os.WriteFile(bin, []byte(script), 0755); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	finished := make(chan struct{})
	go func() {
		defer close(finished)
		_, err := (execDocRepository{sdlcBin: bin}).Mode(ctx, root)
		done <- err
	}()
	defer func() {
		cancel()
		select {
		case <-finished:
		case <-time.After(3 * time.Second):
			t.Error("fixture did not stop during cleanup")
		}
	}()
	deadline := time.NewTimer(10 * time.Second)
	defer deadline.Stop()
	arrived := make(chan error, 1)
	go func() { _, err := io.ReadFull(opened, make([]byte, 1)); arrived <- err }()
	select {
	case err := <-arrived:
		if err != nil {
			t.Fatal(err)
		}
	case err := <-done:
		t.Fatalf("ended before output-open barrier: %v", err)
	case <-deadline.C:
		t.Fatal("no output-open barrier")
	}
	if _, err := os.Stat(ready); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("readiness published before PID write: %v", err)
	}
	partial, err := os.ReadFile(ready + ".tmp")
	if err != nil {
		t.Fatal(err)
	}
	_, partialErr := strconv.Atoi(strings.TrimSpace(string(partial)))
	capture, err := os.ReadFile("testdata/doc-repository-cancellation/empty-pid.log")
	if err != nil || len(partial) != 0 || partialErr == nil || !bytes.Contains(capture, []byte(partialErr.Error())) {
		t.Fatalf("partial readiness %q / %v does not replay recorded failure: %v", partial, partialErr, err)
	}
	if _, err := release.WriteString("continue\n"); err != nil {
		t.Fatal(err)
	}
	ticker := time.NewTicker(5 * time.Millisecond)
	defer ticker.Stop()
	for {
		if _, err := os.Stat(ready); err == nil {
			break
		}
		select {
		case err := <-done:
			t.Fatalf("ended before readiness: %v", err)
		case <-deadline.C:
			t.Fatal("no readiness")
		case <-ticker.C:
		}
	}
	raw, err := os.ReadFile(ready)
	if err != nil {
		t.Fatal(err)
	}
	pgid, err := strconv.Atoi(strings.TrimSpace(string(raw)))
	if err != nil || pgid <= 1 {
		t.Fatalf("invalid process group %q: %v", raw, err)
	}
	cancel()
	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("wrong cancellation: %v", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("cancelled subprocess/descendants retained pipes")
	}
	until := time.NewTimer(3 * time.Second)
	defer until.Stop()
	for {
		if errors.Is(procgroup.Kill(pgid, syscall.Signal(0)), os.ErrProcessDone) {
			break
		}
		select {
		case <-until.C:
			t.Fatal("cancelled subprocess left a runnable process group")
		case <-ticker.C:
		}
	}
}

// The external publisher owns effects. This executable fixture writes a real
// bare remote, loses its first reply, and confirms the same immutable candidate
// on retry; the adapter must never invent a receipt from a failed invocation.
func TestDocRepositoryProductLostAcknowledgment(t *testing.T) {
	root, base := docRepoFixture(t)
	bare := filepath.Join(t.TempDir(), "remote.git")
	docRepoGit(t, root, "init", "--bare", "-q", bare)
	docRepoGit(t, root, "push", "-q", bare, "HEAD:main")
	target := "workshop/workingdoc/two.md"
	payload := []byte("new captured bytes\n")
	if err := os.WriteFile(filepath.Join(root, target), payload, 0644); err != nil {
		t.Fatal(err)
	}
	docRepoGit(t, root, "add", "--", target)
	tree := docRepoGit(t, root, "write-tree")
	candidate := docRepoGit(t, root, "commit-tree", tree, "-p", base, "-m", "publish selected")
	docRepoGit(t, root, "reset", "-q", "HEAD", "--", target)
	bin := docRepoWorkspaceBinary(t, root, nil)
	dir := filepath.Dir(bin)
	receipt, _ := json.Marshal(map[string]any{"schema_version": 1, "path": target, "commit": candidate})
	os.WriteFile(filepath.Join(dir, "receipt.json"), receipt, 0644)
	raw, err := os.ReadFile(bin)
	if err != nil {
		t.Fatal(err)
	}
	insert := " git push -q '" + bare + "' '" + candidate + ":refs/heads/main' || exit 1\n if test ! -f '" + filepath.Join(dir, "lost") + "'; then\n touch '" + filepath.Join(dir, "lost") + "'\n echo 'acknowledgment lost' >&2\n exit 1\n fi\n"
	script := strings.Replace(string(raw), "else\n", "else\n"+insert, 1)
	os.WriteFile(bin, []byte(script), 0755)
	repo := execDocRepository{sdlcBin: bin}
	if commit, err := repo.Publish(context.Background(), root, target, payload, base); err == nil || commit != "" {
		t.Fatalf("lost acknowledgment conferred authority: %q %v", commit, err)
	}
	if actual := docRepoGit(t, root, "--git-dir", bare, "rev-parse", "main"); actual != candidate {
		t.Fatalf("fake did not publish: %s", actual)
	}
	commit, err := repo.Publish(context.Background(), root, target, payload, base)
	if err != nil || commit != candidate {
		t.Fatalf("retry %s: %v", commit, err)
	}
	if actual := docRepoGit(t, root, "--git-dir", bare, "rev-parse", "main"); actual != candidate {
		t.Fatal("retry duplicated publication")
	}
	if actual := docRepoGit(t, root, "rev-parse", "HEAD"); actual != base {
		t.Fatal("adapter moved HEAD")
	}
	if actual := docRepoGit(t, root, "diff-tree", "--no-commit-id", "--name-only", "-r", base, candidate); actual != target {
		t.Fatalf("publication selected extra paths %q", actual)
	}
}

func TestDocRepositoryPrepareFetchesWithoutTargetWrite(t *testing.T) {
	root, head := docRepoFixture(t)
	bare := filepath.Join(t.TempDir(), "remote.git")
	docRepoGit(t, root, "init", "--bare", "-q", bare)
	docRepoGit(t, root, "push", "-q", bare, "HEAD:main")
	peer := filepath.Join(t.TempDir(), "peer")
	docRepoGit(t, root, "clone", "-q", "-b", "main", bare, peer)
	docRepoGit(t, peer, "config", "user.name", "Test")
	docRepoGit(t, peer, "config", "user.email", "test@example.com")
	docRepoGit(t, peer, "config", "commit.gpgsign", "false")
	target := "workshop/workingdoc/two.md"
	payload := []byte("published by another checkout\n")
	if err := os.WriteFile(filepath.Join(peer, target), payload, 0644); err != nil {
		t.Fatal(err)
	}
	docRepoGit(t, peer, "commit", "-qam", "peer publication")
	docRepoGit(t, peer, "push", "-q", "origin", "HEAD:main")
	published := docRepoGit(t, peer, "rev-parse", "HEAD")
	repo := execDocRepository{sdlcBin: docRepoWorkspaceBinary(t, root, nil)}
	pointer := settlementPointer{target, published}
	if _, err := repo.ReadBaselines(context.Background(), root, []settlementPointer{pointer}); err == nil {
		t.Fatal("fixture baseline already available")
	}
	before, err := os.ReadFile(filepath.Join(root, target))
	if err != nil {
		t.Fatal(err)
	}
	index := docRepoGit(t, root, "write-tree")
	realGit, err := exec.LookPath("git")
	if err != nil {
		t.Fatal(err)
	}
	shim := t.TempDir()
	script := "#!/bin/sh\nif test \"$1\" = fetch; then\n exec '" + realGit + "' -c 'url." + bare + ".insteadOf=https://github.com/xianxu/product.git' \"$@\"\nfi\nexec '" + realGit + "' \"$@\"\n"
	os.WriteFile(filepath.Join(shim, "git"), []byte(script), 0755)
	t.Setenv("PATH", shim+string(os.PathListSeparator)+os.Getenv("PATH"))
	if err := repo.Prepare(context.Background(), root, target); err != nil {
		t.Fatal(err)
	}
	blobs, err := repo.ReadBaselines(context.Background(), root, []settlementPointer{pointer})
	if err != nil || !bytes.Equal(blobs[pointer], payload) {
		t.Fatalf("baseline unavailable after prepare: %q %v", blobs[pointer], err)
	}
	after, err := os.ReadFile(filepath.Join(root, target))
	if err != nil || !bytes.Equal(before, after) {
		t.Fatal("prepare wrote target")
	}
	if docRepoGit(t, root, "rev-parse", "HEAD") != head || docRepoGit(t, root, "write-tree") != index {
		t.Fatal("prepare changed HEAD/index")
	}
	// A remote failure is unknown evidence, never an invitation to create a copy.
	os.WriteFile(filepath.Join(shim, "git"), []byte("#!/bin/sh\nif test \"$1\" = fetch; then echo 'remote unavailable' >&2; exit 55; fi\nexec '"+realGit+"' \"$@\"\n"), 0755)
	if err := repo.Prepare(context.Background(), root, target); err == nil || !strings.Contains(err.Error(), "remote unavailable") {
		t.Fatalf("lost fetch error: %v", err)
	}
	after, err = os.ReadFile(filepath.Join(root, target))
	if err != nil || !bytes.Equal(before, after) {
		t.Fatal("failed prepare wrote target")
	}
}
