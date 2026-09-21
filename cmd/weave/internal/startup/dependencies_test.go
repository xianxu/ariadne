package startup

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/weave/internal/weavefs"
)

func TestDependenciesBundlesAndRepeat(t *testing.T) {
	root := t.TempDir()
	fakeBin := filepath.Join(root, "fakebin")
	os.MkdirAll(fakeBin, 0755)
	brew := `#!/bin/sh
set -eu
[ "$1" = bundle ] && [ "$2" = install ] && [ "$3" = --no-upgrade ] && [ "$4" = --file=Brewfile ]
[ ! -e fail ] || exit 7
while read package; do
  [ -n "$package" ] || continue
  touch "$PACKAGE_STATE/$package"
done < Brewfile
printf '%s\n' "$PWD" >> "$PACKAGE_STATE/visits"
`
	os.WriteFile(filepath.Join(fakeBin, "brew"), []byte(brew), 0755)
	state := filepath.Join(root, "state")
	os.MkdirAll(state, 0755)
	t.Setenv("PACKAGE_STATE", state)
	t.Setenv("PATH", fakeBin+string(os.PathListSeparator)+os.Getenv("PATH"))
	base, leaf := filepath.Join(root, "base"), filepath.Join(root, "leaf")
	os.MkdirAll(base, 0755)
	os.MkdirAll(leaf, 0755)
	os.WriteFile(filepath.Join(base, "Brewfile"), []byte("go\n"), 0644)
	os.WriteFile(filepath.Join(leaf, "Brewfile"), []byte("cue\n"), 0644)
	for range 2 {
		if err := Dependencies(weavefs.OSFS{}, []string{base, leaf}, weavefs.ExecRunner{}, false, &bytes.Buffer{}); err != nil {
			t.Fatal(err)
		}
	}
	for _, p := range []string{"go", "cue"} {
		if _, err := os.Stat(filepath.Join(state, p)); err != nil {
			t.Fatal(err)
		}
	}
	visits, _ := os.ReadFile(filepath.Join(state, "visits"))
	if string(visits) != strings.Repeat(base+"\n"+leaf+"\n", 2) {
		t.Fatalf("unexpected order %q", visits)
	}
	os.WriteFile(filepath.Join(base, "fail"), nil, 0644)
	if err := Dependencies(weavefs.OSFS{}, []string{base, leaf}, weavefs.ExecRunner{}, false, &bytes.Buffer{}); err == nil {
		t.Fatal("failed install succeeded")
	}
	after, _ := os.ReadFile(filepath.Join(state, "visits"))
	if string(after) != string(visits) {
		t.Fatal("continued after failed bundle")
	}
}

func TestDependenciesDryRunAndNoBundle(t *testing.T) {
	root := t.TempDir()
	os.WriteFile(filepath.Join(root, "Brewfile"), []byte("brew \"go\"\n"), 0644)
	var out bytes.Buffer
	if err := Dependencies(weavefs.OSFS{}, []string{root}, nil, true, &out); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "brew bundle install --no-upgrade") {
		t.Fatal(out.String())
	}
	if err := Dependencies(weavefs.OSFS{}, []string{t.TempDir()}, nil, false, &out); err != nil {
		t.Fatal(err)
	}
}
