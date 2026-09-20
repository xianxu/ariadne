package plan

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/xianxu/ariadne/cmd/weave/internal/staging"
	"github.com/xianxu/ariadne/cmd/weave/internal/weavefs"
)

type partialPublicationFS struct {
	weavefs.OSFS
	filename string
}

func (f partialPublicationFS) WriteFile(path string, data []byte) error {
	if filepath.Base(path) == f.filename {
		if err := f.OSFS.WriteFile(path, data[:len(data)/2]); err != nil {
			return err
		}
		return errors.New("injected partial write")
	}
	return f.OSFS.WriteFile(path, data)
}
func TestManagedPartialFileWritePreservesIdentityForRetryAndRetirement(t *testing.T) {
	for _, warm := range []bool{false, true} {
		t.Run(map[bool]string{false: "cold", true: "warm"}[warm], func(t *testing.T) {
			root := t.TempDir()
			if warm {
				managedApply(t, root, []Action{WriteFile{Path: "out", Content: "old complete"}}, ScopeArtifacts)
			}
			if _, err := ApplyManaged(partialPublicationFS{filename: "out"}, root, []Action{WriteFile{Path: "out", Content: "new complete output"}}, ScopeArtifacts); err == nil {
				t.Fatal("expected partial write failure")
			}
			if warm {
				if got := managedRead(t, root, "out"); got != "old complete" {
					t.Fatalf("partial bytes published: %q", got)
				}
			} else {
				managedAbsent(t, root, "out")
			}
			managedApply(t, root, []Action{WriteFile{Path: "out", Content: "new complete output"}}, ScopeArtifacts)
			managedApply(t, root, nil, ScopeArtifacts)
			managedAbsent(t, root, "out")
		})
	}
}
func TestSeedAndIgnorePartialWritesPreservePreviousBytes(t *testing.T) {
	for _, kind := range []string{"seed", "ignore", "inventory"} {
		t.Run(kind, func(t *testing.T) {
			root := t.TempDir()
			filename := "out"
			var actions []Action
			switch kind {
			case "seed":
				source := filepath.Join(t.TempDir(), "source")
				if err := os.WriteFile(source, []byte("new complete seed"), 0755); err != nil {
					t.Fatal(err)
				}
				actions = []Action{Seed{Src: source, Dst: filename}}
			case "ignore":
				filename = ".gitignore"
				actions = []Action{EnsureGitignore{Entries: []string{"/out"}}}
			case "inventory":
				filename = "ownership.json"
				managedApply(t, root, []Action{WriteFile{Path: "out", Content: "old"}}, ScopeArtifacts)
			}
			path := filename
			if kind == "inventory" {
				path = InventoryPath
			} else {
				managedWrite(t, root, path, "old complete")
			}
			before := managedRead(t, root, path)
			fs := partialPublicationFS{filename: filename}
			var err error
			if kind == "inventory" {
				_, err = ApplyManaged(fs, root, []Action{WriteFile{Path: "out", Content: "new"}}, ScopeArtifacts)
			} else {
				err = Apply(fs, root, actions)
			}
			if err == nil {
				t.Fatal("partial write succeeded")
			}
			if got := managedRead(t, root, path); got != before {
				t.Fatalf("old %s replaced with partial bytes %q", kind, got)
			}
		})
	}
}
func TestStagedFileModesAreOwnedAndPreserved(t *testing.T) {
	root := t.TempDir()
	stage := t.TempDir()
	managedWrite(t, stage, "SKILL.md", "generated")
	managedWrite(t, stage, "run.sh", "#!/bin/sh\nexit 0\n")
	for _, mode := range []os.FileMode{0755, 0644, 0700} {
		if err := os.Chmod(filepath.Join(stage, "run.sh"), mode); err != nil {
			t.Fatal(err)
		}
		actions, err := StagedActions(weavefs.OSFS{}, root, stage, "construct/generated/demo")
		if err != nil {
			t.Fatal(err)
		}
		managedApply(t, root, actions, ScopeArtifacts)
		info, err := os.Stat(filepath.Join(root, "construct/generated/demo/run.sh"))
		if err != nil {
			t.Fatal(err)
		}
		if info.Mode().Perm() != mode {
			t.Fatalf("mode = %o, want %o", info.Mode().Perm(), mode)
		}
	}
	if err := os.Chmod(filepath.Join(root, "construct/generated/demo/run.sh"), 0600); err != nil {
		t.Fatal(err)
	}
	actions, err := StagedActions(weavefs.OSFS{}, root, stage, "construct/generated/demo")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := ApplyManaged(weavefs.OSFS{}, root, actions, ScopeArtifacts); err == nil || !strings.Contains(err.Error(), "authored") {
		t.Fatalf("authored permission edit not protected: %v", err)
	}
	managedApply(t, root, nil, ScopeArtifacts)
	if _, err := os.Stat(filepath.Join(root, "construct/generated/demo/run.sh")); err != nil {
		t.Fatal("retirement removed permission-edited output", err)
	}
}

type killedPublicationFS struct{ weavefs.OSFS }

func (f killedPublicationFS) WriteFile(path string, data []byte) error {
	if filepath.Base(path) == "out" {
		if err := f.OSFS.WriteFile(path, data[:len(data)/2]); err != nil {
			return err
		}
		fmt.Println("partial-stage-ready")
		for {
			time.Sleep(time.Hour)
		}
	}
	return f.OSFS.WriteFile(path, data)
}
func TestPublicationWriterProcess(t *testing.T) {
	root := os.Getenv("WEAVE_PUBLICATION_TEST_ROOT")
	if root == "" {
		t.Skip("publication subprocess only")
	}
	if _, err := ApplyManaged(killedPublicationFS{}, root, []Action{WriteFile{Path: "out", Content: "new complete"}}, ScopeArtifacts); err != nil {
		t.Fatal(err)
	}
}
func TestKilledPartialPublicationRecoversWithAndWithoutActions(t *testing.T) {
	for _, retry := range []bool{false, true} {
		t.Run(map[bool]string{false: "retire", true: "retry"}[retry], func(t *testing.T) {
			root := t.TempDir()
			managedApply(t, root, []Action{WriteFile{Path: "out", Content: "old complete"}}, ScopeArtifacts)
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestPublicationWriterProcess$")
			cmd.Env = append(os.Environ(), "WEAVE_PUBLICATION_TEST_ROOT="+root)
			pipe, err := cmd.StdoutPipe()
			if err != nil {
				t.Fatal(err)
			}
			if err := cmd.Start(); err != nil {
				t.Fatal(err)
			}
			defer func() { _ = cmd.Process.Kill(); _ = cmd.Wait() }()
			scanner := bufio.NewScanner(pipe)
			if !scanner.Scan() || scanner.Text() != "partial-stage-ready" {
				t.Fatal("writer did not stage partial output", scanner.Text(), scanner.Err())
			}
			if got := managedRead(t, root, "out"); got != "old complete" {
				t.Fatalf("partial final content: %q", got)
			}
			if err := cmd.Process.Kill(); err != nil {
				t.Fatal(err)
			}
			_ = cmd.Wait()
			if retry {
				managedApply(t, root, []Action{WriteFile{Path: "out", Content: "new complete"}}, ScopeArtifacts)
			}
			managedApply(t, root, nil, ScopeArtifacts)
			managedAbsent(t, root, "out")
			entries, err := os.ReadDir(filepath.Join(root, staging.RootRel))
			if err != nil {
				t.Fatal(err)
			}
			for _, entry := range entries {
				if strings.HasPrefix(entry.Name(), ".publication-weave-") {
					t.Fatalf("dead publication stage remains: %s", entry.Name())
				}
			}
		})
	}
}
func TestLegacyInventoryWithoutModeRemainsReadable(t *testing.T) {
	root := t.TempDir()
	managedApply(t, root, []Action{WriteFile{Path: "out", Content: "generated"}}, ScopeArtifacts)
	if strings.Contains(managedRead(t, root, InventoryPath), `"mode"`) {
		t.Fatal("ordinary file unexpectedly mode-owned")
	}
	mode := os.FileMode(0755)
	managedApply(t, root, []Action{WriteFile{Path: "out", Content: "generated", Mode: &mode}}, ScopeArtifacts)
	managedApply(t, root, nil, ScopeArtifacts)
	managedAbsent(t, root, "out")
}

func TestManagedRejectsReservedSeedBeforePublication(t *testing.T) {
	for _, path := range []string{InventoryPath, ".gitignore"} {
		t.Run(path, func(t *testing.T) {
			root := t.TempDir()
			managedApply(t, root, []Action{WriteFile{Path: "existing", Content: "owned"}}, ScopeArtifacts)
			before := managedRead(t, root, path)
			source := filepath.Join(t.TempDir(), "source")
			if err := os.WriteFile(source, []byte("not inventory or managed ignores"), 0644); err != nil {
				t.Fatal(err)
			}
			actions := []Action{Seed{Src: source, Dst: path}, WriteFile{Path: "out", Content: "later failure"}}
			if _, err := ApplyManaged(partialPublicationFS{filename: "out"}, root, actions, ScopeArtifacts); err == nil {
				t.Fatal("accepted reserved seed")
			}
			if got := managedRead(t, root, path); got != before {
				t.Fatalf("reserved state changed before failure: %q", got)
			}
			managedAbsent(t, root, "out")
		})
	}
}
