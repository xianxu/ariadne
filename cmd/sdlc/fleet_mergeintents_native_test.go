package main

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/fleet"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/landingline"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/queuegate"
)

func TestFleetMergeIntentsLegacyV1Fixture(t *testing.T) {
	path := filepath.Join("testdata", "merge-intents-v1", "legacy.json")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	in, err := landingline.DecodeIntent(raw)
	if err != nil {
		t.Fatal(err)
	}
	if in.Workspace != "" || string(in.Extra["future"]) != `{"nested":[1,2]}` {
		t.Fatalf("legacy decode: %+v", in)
	}
	report := reportMergeIntent(landingline.ObservedIntent{Intent: in, Path: "/legacy/.git/sdlc/landing/intents/fixture"}, true, nil)
	if report.State != "orphaned" || report.Root != in.Root || report.Reentry != "sdlc merge --branch "+in.Branch+" --yes" {
		t.Fatalf("legacy observation: %+v", report)
	}
	rewritten, err := json.Marshal(in)
	if err != nil {
		t.Fatal(err)
	}
	again, err := landingline.DecodeIntent(rewritten)
	if err != nil || !bytes.Equal(again.Extra["future"], in.Extra["future"]) {
		t.Fatalf("unknown metadata lost: %s %v", rewritten, err)
	}
	after, err := os.ReadFile(path)
	if err != nil || !bytes.Equal(after, raw) {
		t.Fatal("legacy reader rewrote the fixture")
	}
}

func snapshotFleetIntentFixture(t *testing.T, root string) map[string]string {
	t.Helper()
	out := map[string]string{}
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			out[path] = "directory"
			return nil
		}
		raw, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		out[path] = string(raw)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return out
}

func TestFleetMergeIntentsPrivateExecutableConformance(t *testing.T) {
	isolateFleetIntegrationGit(t)
	queueRoot := canonicalRoot(t.TempDir())
	t.Setenv("WF_SDLC_QUEUE_ROOT", queueRoot)
	t.Setenv("COUCH_TREE", "")
	binary := buildFleetE2EBinary(t)
	base := canonicalRoot(t.TempDir())
	primary := newFleetE2ERepo(t, base, "sample", "")
	slot := filepath.Join(base, "worktree", "sample-slot1", "sample")
	if err := os.MkdirAll(filepath.Dir(slot), 0700); err != nil {
		t.Fatal(err)
	}
	runFleetIntegrationGit(t, primary, "worktree", "add", "-b", "000624-native", slot)
	dependency := newFleetE2ERepo(t, filepath.Dir(slot), "base", "")
	if err := os.MkdirAll(filepath.Join(slot, "construct"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(slot, "construct", "deps"), []byte("substrate ../base\n"), 0600); err != nil {
		t.Fatal(err)
	}
	runFleetIntegrationGit(t, dependency, "branch", "000624-native")
	makeIntent := func(root string) (landingline.Intents, *landingline.IntentHold) {
		t.Helper()
		line, err := landingLineDir(execGitRunner{}, root)
		if err != nil {
			t.Fatal(err)
		}
		store := landingline.OpenIntents(line)
		head := strings.TrimSpace(string(runFleetIntegrationGit(t, root, "rev-parse", "HEAD")))
		hold, err := store.Hold(landingline.Intent{Branch: "000624-native", Head: head, Root: canonicalRoot(root), PID: 4242, StartedAt: time.Date(2026, 10, 25, 0, 0, 0, 0, time.UTC)})
		if err != nil {
			t.Fatal(err)
		}
		return store, hold
	}
	store, hold := makeIntent(slot)
	t.Cleanup(func() {
		if hold != nil {
			_ = hold.Release(false)
		}
	})
	_, depHold := makeIntent(dependency)
	if err := depHold.Release(true); err != nil {
		t.Fatal(err)
	}
	queue := &queuegate.Store{Root: queueRoot}
	if err := queue.Pause(context.Background(), "fixture-owner", "literal reason ; $()", false); err != nil {
		t.Fatal(err)
	}
	read := func() fleet.Inventory {
		t.Helper()
		beforeQueue := snapshotFleetIntentFixture(t, queueRoot)
		line, err := landingLineDir(execGitRunner{}, slot)
		if err != nil {
			t.Fatal(err)
		}
		beforeStore := snapshotFleetIntentFixture(t, filepath.Join(line, "intents"))
		cmd := exec.CommandContext(t.Context(), binary, "fleet", "inventory", "--json", "--path", slot)
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		if err := cmd.Run(); err != nil {
			t.Fatalf("private inventory: %v\n%s", err, &stderr)
		}
		if stderr.Len() != 0 {
			t.Fatalf("non-JSON diagnostics on stderr: %s", &stderr)
		}
		var got fleet.Inventory
		if err := json.Unmarshal(stdout.Bytes(), &got); err != nil {
			t.Fatal(err)
		}
		if got.Capabilities.MergeIntents != 1 {
			t.Fatal("missing producer capability")
		}
		if !reflect.DeepEqual(beforeStore, snapshotFleetIntentFixture(t, filepath.Join(line, "intents"))) || !reflect.DeepEqual(beforeQueue, snapshotFleetIntentFixture(t, queueRoot)) {
			t.Fatal("read-only command changed store or queue")
		}
		return got
	}
	first := read()
	host := findFleetIntegrationRow(t, first, slot)
	dep := findFleetIntegrationRow(t, first, dependency)
	if len(host.MergeIntents) != 1 || host.MergeIntents[0].State != "live" || host.MergeIntents[0].Reentry != "" {
		t.Fatalf("real held owner not live: %+v", host)
	}
	if len(dep.MergeIntents) != 1 || dep.MergeIntents[0].State != "orphaned" || !strings.Contains(dep.MergeIntents[0].Note, "literal reason ; $()") {
		t.Fatalf("dependency/pause evidence: %+v", dep)
	}
	if err := hold.Release(true); err != nil {
		t.Fatal(err)
	}
	hold = nil
	second := read()
	host = findFleetIntegrationRow(t, second, slot)
	if len(host.MergeIntents) != 1 || host.MergeIntents[0].State != "orphaned" {
		t.Fatalf("released owner not orphaned: %+v", host)
	}
	observed, err := store.List()
	if err != nil || len(observed) != 1 {
		t.Fatal("private store fixture")
	}
	corrupt := filepath.Join(filepath.Dir(observed[0].Path), "corrupt")
	if err := os.WriteFile(corrupt, []byte("not json"), 0600); err != nil {
		t.Fatal(err)
	}
	mixed := read()
	host = findFleetIntegrationRow(t, mixed, slot)
	if host.MergeIntentsState != "partial" || len(host.MergeIntents) != 1 {
		t.Fatalf("real mixed read lost evidence: %+v", host)
	}
	diagnostics := 0
	for _, d := range mixed.Diagnostics {
		if d.Stage == "merge_intents" {
			diagnostics++
			if !strings.Contains(d.Message, corrupt) {
				t.Fatal("missing corrupt record path")
			}
		}
	}
	if diagnostics != 1 {
		t.Fatalf("corrupt record reported %d times", diagnostics)
	}
	if err := os.Remove(corrupt); err != nil {
		t.Fatal(err)
	}
	// Terminal removal follows the real store API, not a fabricated tombstone.
	in := observed[0].Intent
	terminal, err := store.Hold(in)
	if err != nil {
		t.Fatal(err)
	}
	if err := terminal.Release(false); err != nil {
		t.Fatal(err)
	}
	runFleetIntegrationGit(t, dependency, "branch", "-D", "000624-native")
	last := read()
	host = findFleetIntegrationRow(t, last, slot)
	dep = findFleetIntegrationRow(t, last, dependency)
	if host.MergeIntentsState != "complete" || len(host.MergeIntents) != 0 {
		t.Fatalf("terminal removal: %+v", host)
	}
	if len(dep.MergeIntents) != 1 || dep.MergeIntents[0].State != "stale" || dep.MergeIntents[0].Reentry != "" {
		t.Fatalf("stale branch: %+v", dep)
	}
}
