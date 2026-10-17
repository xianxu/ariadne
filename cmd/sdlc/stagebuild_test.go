package main

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"
	"testing"
)

func TestStageBuildSentinelTargetAndRenamedOutput(t *testing.T) {
	const commit = "1234567890123456789012345678901234567890"
	prev := readBuildInfo
	t.Cleanup(func() { readBuildInfo = prev })
	readBuildInfo = func(path string) (*debug.BuildInfo, error) {
		stamp, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		command := filepath.Base(path)
		if command == "pair" {
			command = "pair-go"
		}
		return &debug.BuildInfo{Path: "example.com/test/cmd/" + command, Settings: []debug.BuildSetting{{Key: "vcs.revision", Value: strings.TrimSpace(string(stamp))}}}, nil
	}
	for _, tc := range []struct {
		name, command, binary, marker, makefile string
		wantErr                                 string
	}{
		{"declared target", "sdlc", "sdlc", "# stage-build: sdlc-build\n", "build:\n\tmkdir -p bin && printf old > bin/sdlc\nsdlc-build:\n\tprintf " + commit + " > bin/sdlc\n", ""},
		{"renamed generic output", "pair-go", "pair", "# built by Makefile.local\n", "build:\n\tmkdir -p bin && printf " + commit + " > bin/pair\n", ""},
		{"missing output", "pair-go", "pair", "# built by Makefile.local\n", "build:\n\ttrue\n", "cmd/pair-go"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			commandDir := filepath.Join(root, "cmd", tc.command)
			if err := os.MkdirAll(commandDir, 0o755); err != nil {
				t.Fatal(err)
			}
			for name, data := range map[string]string{
				filepath.Join(root, "go.mod"):                 "module example.com/test\n",
				filepath.Join(root, "Makefile"):               tc.makefile,
				filepath.Join(commandDir, "main.go"):          "package main\n",
				filepath.Join(commandDir, ".skip-make-build"): tc.marker,
			} {
				if err := os.WriteFile(name, []byte(data), 0o644); err != nil {
					t.Fatal(err)
				}
			}
			var out bytes.Buffer
			err := stageBuild(context.Background(), root, commit, &out, &out)
			if tc.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
					t.Fatalf("stageBuild error = %v, want %q; output:\n%s", err, tc.wantErr, &out)
				}
				return
			}
			if err != nil || !strings.Contains(out.String(), "bin/"+tc.binary+" vcs.revision="+commit) {
				t.Fatalf("stageBuild = %v; output:\n%s", err, &out)
			}
		})
	}
}

func TestStageBuildRefusesUnverifiableBinary(t *testing.T) {
	const commit = "1234567890123456789012345678901234567890"
	for _, tc := range []struct {
		name, revision string
		modified       bool
		readErr        error
	}{
		{name: "missing revision"},
		{name: "modified", revision: commit, modified: true},
		{name: "unreadable", readErr: errors.New("corrupt Go build info")},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			if err := os.WriteFile(filepath.Join(root, "Makefile"), []byte("build:\n\tmkdir -p bin && printf x > bin/tool\n"), 0o644); err != nil {
				t.Fatal(err)
			}
			prev := readBuildInfo
			t.Cleanup(func() { readBuildInfo = prev })
			readBuildInfo = func(string) (*debug.BuildInfo, error) {
				return &debug.BuildInfo{Settings: []debug.BuildSetting{
					{Key: "vcs.revision", Value: tc.revision},
					{Key: "vcs.modified", Value: map[bool]string{true: "true", false: "false"}[tc.modified]},
				}}, tc.readErr
			}
			var out bytes.Buffer
			if err := stageBuild(context.Background(), root, commit, &out, &out); err == nil || !strings.Contains(err.Error(), "tool") {
				t.Fatalf("stageBuild error = %v, want named refusal; output:\n%s", err, &out)
			}
		})
	}
}
