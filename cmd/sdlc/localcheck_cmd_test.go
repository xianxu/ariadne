package main

import (
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSelectLocalChecks(t *testing.T) {
	decl := localCheckDecl{Version: 1, Checks: []localCheck{{Name: "a", Command: []string{"x"}}, {Name: "b", Command: []string{"y"}}}}
	if got, err := selectLocalChecks(decl, nil, true); err != nil || len(got) != 2 || got[0].Name != "a" {
		t.Fatalf("--all: %v %v", got, err)
	}
	if got, err := selectLocalChecks(decl, []string{"b"}, false); err != nil || len(got) != 1 || got[0].Name != "b" {
		t.Fatalf("--name b: %v %v", got, err)
	}
	if _, err := selectLocalChecks(decl, []string{"zz"}, false); err == nil || !strings.Contains(err.Error(), "declared: a, b") {
		t.Fatalf("unknown: %v", err)
	}
	for _, tc := range []struct {
		names []string
		all   bool
	}{{nil, false}, {[]string{"a"}, true}} {
		if _, err := selectLocalChecks(decl, tc.names, tc.all); err == nil {
			t.Fatalf("%v %v accepted", tc.names, tc.all)
		}
	}
}

func TestLocalCheckCmd(t *testing.T) {
	decl := `{"version":1,"checks":[{"name":"ok","command":["sh","-c","echo fine"]},{"name":"bad","command":["sh","-c","exit 4"]}]}`
	dir := tempRepoWith(t, map[string]string{
		localChecksPath:        decl,
		"Makefile.workflow":    "# fake\n",
		"scripts/test-lock.py": fakeSuiteLockHelper("print('acquired')\n"),
	})
	ci := newCIFakeGH()
	head := strings.TrimSpace(git(t, dir, "rev-parse", "HEAD"))
	ci.setHead(5, head)
	swapMergeDeps(t, &e2eGH{openPR: "5", ciSeamGH: ciSeamGH{ci: ci}}, nil)
	ctx := context.Background()

	if err := runLocalCheckCmd(ctx, io.Discard, localCheckFlags{All: true, Rev: "HEAD"}); err == nil || !strings.Contains(err.Error(), "bad") || strings.Contains(err.Error(), "ok (") {
		t.Fatalf("--all: %v", err)
	}
	if ci.statuses[head]["sdlc-local/ok"] != "success" || ci.statuses[head]["sdlc-local/bad"] != "failure" || len(ci.comments[5]) != 1 {
		t.Fatalf("statuses %v comments %d", ci.statuses, len(ci.comments[5]))
	}

	// A pushed commit that is not a branch tip (an older commit of the PR) is
	// refused: merge reads the statuses on the pushed head only.
	if err := os.WriteFile(filepath.Join(dir, "next.txt"), []byte("y\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	git(t, dir, "add", "-A")
	git(t, dir, "commit", "-qm", "next")
	git(t, dir, "push", "-q", "origin", "feature")
	if err := runLocalCheckCmd(ctx, io.Discard, localCheckFlags{Names: []string{"ok"}, Rev: "HEAD~1"}); err == nil || !strings.Contains(err.Error(), "not the tip of a branch on origin") {
		t.Fatalf("pushed non-tip: %v", err)
	}

	// An unpushed commit is refused before anything runs.
	if err := os.WriteFile(filepath.Join(dir, "local.txt"), []byte("x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	git(t, dir, "add", "-A")
	git(t, dir, "commit", "-qm", "local only")
	var stderr bytes.Buffer
	if err := runLocalCheckCmd(ctx, &stderr, localCheckFlags{Names: []string{"ok"}, Rev: "HEAD"}); err == nil || !strings.Contains(err.Error(), "not the tip of a branch on origin") || stderr.Len() != 0 {
		t.Fatalf("unpushed: %v %q", err, stderr.String())
	}
}

func TestLocalCheckHelp(t *testing.T) {
	h := renderLong("local-check")
	for _, want := range []string{localChecksPath, "test-lock.py", "last 60 lines", "required", "before any runner took it"} {
		if !strings.Contains(strings.ReplaceAll(h, "\n", " "), want) && !strings.Contains(h, want) {
			t.Errorf("local-check help does not mention %q", want)
		}
	}
}
