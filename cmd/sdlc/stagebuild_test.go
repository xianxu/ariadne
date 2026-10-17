package main

import (
	"bytes"
	"context"
	"errors"
	"io"
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
		return &debug.BuildInfo{Path: "example.com/test/cmd/" + command, Settings: []debug.BuildSetting{{Key: "vcs.revision", Value: strings.TrimSpace(string(stamp))}, {Key: "vcs.modified", Value: "false"}}}, nil
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

func TestSentinelBuildTargetRejectsMakeArguments(t *testing.T) {
	for _, tc := range []struct{ name, contents, want string }{
		{"option", "# stage-build: --version\n", "stage-build"},
		{"dry run", "# stage-build: -n\n", "stage-build"},
		{"assignment", "# stage-build: NAME=value\n", "stage-build"},
		{"duplicate", "# stage-build: first\n# stage-build: second\n", "stage-build"},
		{"target", "# stage-build: sdlc-build\n", ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			marker := filepath.Join(t.TempDir(), ".skip-make-build")
			if err := os.WriteFile(marker, []byte(tc.contents), 0o644); err != nil {
				t.Fatal(err)
			}
			got, err := sentinelBuildTarget(marker)
			if tc.want == "" {
				if err != nil || got != "sdlc-build" {
					t.Fatalf("target = %q, %v", got, err)
				}
			} else if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("target = %q, %v; want refusal", got, err)
			}
		})
	}
}

func TestStageBuildRejectsMissingOrInvalidCleanStamp(t *testing.T) {
	const commit = "1234567890123456789012345678901234567890"
	for _, stamp := range []string{"", "unknown", "TRUE"} {
		t.Run("stamp_"+stamp, func(t *testing.T) {
			root := t.TempDir()
			if err := os.WriteFile(filepath.Join(root, "Makefile"), []byte("build:\n\tmkdir -p bin && printf x > bin/tool\n"), 0o644); err != nil {
				t.Fatal(err)
			}
			prev := readBuildInfo
			t.Cleanup(func() { readBuildInfo = prev })
			readBuildInfo = func(string) (*debug.BuildInfo, error) {
				settings := []debug.BuildSetting{{Key: "vcs.revision", Value: commit}}
				if stamp != "" {
					settings = append(settings, debug.BuildSetting{Key: "vcs.modified", Value: stamp})
				}
				return &debug.BuildInfo{Settings: settings}, nil
			}
			var out bytes.Buffer
			if err := stageBuild(context.Background(), root, commit, &out, &out); err == nil || !strings.Contains(err.Error(), "bin/tool") {
				t.Fatalf("stageBuild = %v; want named refusal", err)
			}
		})
	}
}

func TestStageBuildRejectsNativeExecutableWithUnreadableBuildInfo(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "Makefile"), []byte("build:\n\tmkdir -p bin && printf '\\177ELFbroken' > bin/tool\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	if err := stageBuild(context.Background(), root, "1234567890123456789012345678901234567890", &out, &out); err == nil || !strings.Contains(err.Error(), "bin/tool") {
		t.Fatalf("stageBuild = %v; want named refusal", err)
	}
}

func TestStageBuildMakeBoundaryRetriesAfterOwnerFailure(t *testing.T) {
	const commit = "1234567890123456789012345678901234567890"
	root := t.TempDir()
	commandDir := filepath.Join(root, "cmd", "sdlc")
	if err := os.MkdirAll(commandDir, 0o755); err != nil {
		t.Fatal(err)
	}
	for path, data := range map[string]string{
		filepath.Join(root, "go.mod"):                 "module example.com/test\n",
		filepath.Join(commandDir, "main.go"):          "package main\n",
		filepath.Join(commandDir, ".skip-make-build"): "# stage-build: sdlc-build\n",
	} {
		if err := os.WriteFile(path, []byte(data), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	bin := filepath.Join(root, "bin", "sdlc")
	prev := readBuildInfo
	t.Cleanup(func() { readBuildInfo = prev })
	readBuildInfo = func(path string) (*debug.BuildInfo, error) {
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		return &debug.BuildInfo{Path: "example.com/test/cmd/sdlc", Settings: []debug.BuildSetting{
			{Key: "vcs.revision", Value: string(data)}, {Key: "vcs.modified", Value: "false"},
		}}, nil
	}
	var calls []string
	failOwner := true
	fakeMake := func(_ context.Context, _ string, target string, _ io.Writer) error {
		calls = append(calls, target)
		if err := os.MkdirAll(filepath.Dir(bin), 0o755); err != nil {
			return err
		}
		switch target {
		case "build":
			return os.WriteFile(bin, []byte("old"), 0o755)
		case "sdlc-build":
			if failOwner {
				return errors.New("owner failed")
			}
			return os.WriteFile(bin, []byte(commit), 0o755)
		default:
			return errors.New("unexpected target")
		}
	}
	var out bytes.Buffer
	if err := stageBuildWithMake(context.Background(), root, commit, &out, &out, fakeMake); err == nil || !strings.Contains(err.Error(), "owner failed") {
		t.Fatalf("first build = %v", err)
	}
	if data, err := os.ReadFile(bin); err != nil || string(data) != "old" {
		t.Fatalf("failed build state = %q, %v", data, err)
	}
	failOwner = false
	if err := stageBuildWithMake(context.Background(), root, commit, &out, &out, fakeMake); err != nil {
		t.Fatal(err)
	}
	if got := strings.Join(calls, ","); got != "build,sdlc-build,build,sdlc-build" {
		t.Fatalf("make calls = %s", got)
	}
	if !strings.Contains(out.String(), "bin/sdlc vcs.revision="+commit) {
		t.Fatalf("output = %s", &out)
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

func TestStageBuildIgnoresSentinelWithoutGoCommand(t *testing.T) {
	root := t.TempDir()
	commandDir := filepath.Join(root, "cmd", "shell-tool")
	if err := os.MkdirAll(commandDir, 0o755); err != nil {
		t.Fatal(err)
	}
	for path, data := range map[string]string{
		filepath.Join(root, "Makefile"):               "build:\n\ttrue\n",
		filepath.Join(commandDir, ".skip-make-build"): "# local shell packaging\n",
	} {
		if err := os.WriteFile(path, []byte(data), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	var out bytes.Buffer
	if err := stageBuild(context.Background(), root, "1234567890123456789012345678901234567890", &out, &out); err != nil {
		t.Fatalf("non-Go sentinel requires go.mod: %v\n%s", err, &out)
	}
}
