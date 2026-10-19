package main

import (
	"bytes"
	"errors"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"
	"testing"
)

func TestBinaryVersionReportsIdentity(t *testing.T) {
	for _, rev := range []string{"0123456789", ""} {
		fakeRunningSDLC(t, binaryBuild{Path: "/private/copy/sdlc", Revision: rev, Modified: "false"})
		root := buildRoot()
		var out bytes.Buffer
		root.SetOut(&out)
		root.SetArgs([]string{"--version"})
		if err := root.Execute(); err != nil {
			t.Fatal(err)
		}
		for _, want := range []string{"/private/copy/sdlc", "revision=" + orNone(rev), "modified=false"} {
			if !strings.Contains(out.String(), want) {
				t.Errorf("want %q in %q", want, out.String())
			}
		}
	}
}

func TestReadBinaryBuildResolvesExecutableAndUnknownMetadata(t *testing.T) {
	root := testfix.Repo(t, testfix.InitialCommit())
	target := filepath.Join(root, "bin", "sdlc")
	if err := os.Mkdir(filepath.Dir(target), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(target, nil, 0755); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(root, "alias")
	if err := os.Symlink(target, link); err != nil {
		t.Fatal(err)
	}
	b := readBinaryBuild(func() (string, error) { return link, nil }, func() (*debug.BuildInfo, bool) { return nil, false })
	resolved, _ := filepath.EvalSymlinks(target)
	if b.Path != resolved || b.Revision != "" || b.Source != filepath.Dir(filepath.Dir(resolved)) {
		t.Fatalf("unexpected build: %+v", b)
	}
	b = readBinaryBuild(func() (string, error) { return "", errors.New("unavailable") }, func() (*debug.BuildInfo, bool) { return nil, false })
	if b.Path != "" || b.Source != "" {
		t.Fatalf("fabricated identity: %+v", b)
	}
}

func TestFleetBinaryWarningIsAdvisoryAndNamesBothPaths(t *testing.T) {
	t.Setenv("COUCH_TREE", "test")
	fakeRunningSDLC(t, binaryBuild{Path: "/private/old/sdlc", Revision: "old"})
	old := expectedFleetSDLC
	expectedFleetSDLC = func() (string, error) { return "/fleet/ariadne/bin/sdlc", nil }
	t.Cleanup(func() { expectedFleetSDLC = old })
	var out bytes.Buffer
	warnFleetBinary(&out)
	for _, want := range []string{"WARNING", "/private/old/sdlc", "/fleet/ariadne/bin/sdlc", "old"} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("want %q in %q", want, out.String())
		}
	}
	out.Reset()
	fakeRunningSDLC(t, binaryBuild{Path: "/fleet/ariadne/bin/sdlc"})
	warnFleetBinary(&out)
	if out.Len() != 0 {
		t.Fatalf("correct fleet binary warned: %s", out.String())
	}
}
