package main

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
)

func TestStartPlanNoticesStaleSettingsOnFirstCurrentRun(t *testing.T) {
	for _, current := range []bool{false, true} {
		name := "stale"
		if current {
			name = "current"
		}
		t.Run(name, func(t *testing.T) {
			f := newCatchUpFixture(t)
			previous := weaveSettingsCheck
			weaveSettingsCheck = runInstalledWeaveSettings
			t.Cleanup(func() { weaveSettingsCheck = previous })
			writeRepoFile(t, f.slot.root, ".claude/settings.json", "old\n")
			if current {
				writeRepoFile(t, f.slot.root, ".claude/settings.json", "new\n")
			}
			bin := t.TempDir()
			// Stateful fake owns only this fixture: observation reflects current bytes.
			writeRepoFile(t, bin, "weave", "#!/bin/sh\n[ \"$*\" = settings ] || exit 77\n[ \"$(cat .claude/settings.json)\" = new ] || echo 'weave: .claude/settings.json is stale: run weave refresh --substrate from your shell'\nexit 0\n")
			if err := os.Chmod(filepath.Join(bin, "weave"), 0755); err != nil {
				t.Fatal(err)
			}
			t.Setenv("PATH", testfix.PrependPath(bin, os.Getenv("PATH")))
			var out bytes.Buffer
			if err := startPlanBranch(context.Background(), &out, 9); err != nil {
				t.Fatalf("first planning failed: %v: %s", err, &out)
			}
			count := strings.Count(out.String(), ".claude/settings.json is stale: run weave refresh --substrate from your shell")
			if (!current && count != 1) || (current && count != 0) {
				t.Fatalf("current=%v notice count=%d: %s", current, count, &out)
			}
			if f.slot.git("branch", "--show-current") != "000009-nine" {
				t.Fatal("first planning did not create issue branch")
			}
		})
	}
}

func TestGateSettingsFailureRemainsVisibleWithoutRefusingPlanning(t *testing.T) {
	f := newCatchUpFixture(t)
	previous := weaveSettingsCheck
	weaveSettingsCheck = func(context.Context, string, io.Writer) error { return errors.New("settings observation failed") }
	t.Cleanup(func() { weaveSettingsCheck = previous })
	var out bytes.Buffer
	if err := startPlanBranch(context.Background(), &out, 9); err != nil {
		t.Fatalf("optional settings observation refused planning: %v", err)
	}
	if n := strings.Count(out.String(), "settings check unavailable (settings observation failed): run weave refresh --substrate from your shell"); n != 1 {
		t.Fatalf("missing precise warning: %s", &out)
	}
	if f.slot.git("branch", "--show-current") != "000009-nine" {
		t.Fatal("planning did not finish")
	}
}
