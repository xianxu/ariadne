package main

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
)

func stageFreshFixture(t *testing.T) ([]string, string) {
	t.Helper()
	r, _ := moveFixture(t)
	remote := strings.TrimSpace(testfix.Capture(t, r[0], "remote", "get-url", "upstream"))
	peer := testfix.Repo(t)
	testfix.Git(t, peer, "remote", "add", "origin", remote)
	testfix.Git(t, peer, "fetch", "-q", "origin")
	testfix.Git(t, peer, "checkout", "-q", "-B", "main", "origin/main")
	procedureWrite(t, peer, "new-main", "current upstream code\n")
	testfix.Git(t, peer, "add", ".")
	testfix.Git(t, peer, "commit", "-qm", "advance main")
	testfix.Git(t, peer, "push", "-q", "origin", "main")
	prev := readBuildInfo
	t.Cleanup(func() { readBuildInfo = prev })
	readBuildInfo = func(path string) (*debug.BuildInfo, error) {
		stamp, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		return &debug.BuildInfo{Settings: []debug.BuildSetting{{Key: "vcs.revision", Value: strings.TrimSpace(string(stamp))}, {Key: "vcs.modified", Value: "false"}}}, nil
	}
	procedureWrite(t, r[0], "Makefile", "build:\n\tmkdir -p bin && git rev-parse HEAD > bin/tool\n")
	return r, procedureHead(t, peer)
}

func TestStageClearFetchesMainBeforeBuild(t *testing.T) {
	for _, staged := range []bool{false, true} {
		t.Run(map[bool]string{false: "resting retry", true: "detached"}[staged], func(t *testing.T) {
			r, tip := stageFreshFixture(t)
			if staged {
				if out, err := runStageTest(t, r[1], ":0", false); err != nil {
					t.Fatalf("stage: %v\n%s", err, out)
				}
			}
			out, err := runUnstageTest(t, r[1], ":0")
			if err != nil || procedureHead(t, r[0]) != tip || !strings.Contains(out, "bin/tool vcs.revision="+tip) {
				t.Fatalf("clear must build freshly fetched main %s, got HEAD %s: %v\n%s", tip, procedureHead(t, r[0]), err, out)
			}
			if got := strings.TrimSpace(testfix.Capture(t, r[0], "branch", "--show-current")); got != "main" {
				t.Fatalf("clear left branch %q", got)
			}
		})
	}
}

func TestStageClearFreshnessRefusesBeforeEffects(t *testing.T) {
	for _, state := range []string{"ahead", "diverged", "offline", "dirty"} {
		t.Run(state, func(t *testing.T) {
			r, _ := stageFreshFixture(t)
			switch state {
			case "ahead":
				testfix.Git(t, r[0], "pull", "-q", "--ff-only", "origin", "main")
				fallthrough
			case "diverged":
				testfix.Git(t, r[0], "commit", "--allow-empty", "-qm", "local rest work")
			case "offline":
				testfix.Git(t, r[0], "remote", "set-url", "origin", filepath.Join(t.TempDir(), "missing.git"))
			}
			if out, err := runStageTest(t, r[1], ":0", false); err != nil {
				t.Fatalf("stage: %v\n%s", err, out)
			}
			if state == "dirty" {
				procedureWrite(t, r[0], "README", "operator edit\n")
			}
			before := slotState(t, r[0])
			if out, err := runUnstageTest(t, r[1], ":0"); err == nil {
				t.Fatalf("accepted %s: %s", state, out)
			}
			if slotState(t, r[0]) != before {
				t.Fatal("refusal changed target/lease")
			}
			if _, err := os.Stat(filepath.Join(r[0], "bin", "tool")); !os.IsNotExist(err) {
				t.Fatalf("refusal ran the build: %v", err)
			}
		})
	}
}

func TestStageClearDryRunReportsFetchedLag(t *testing.T) {
	for _, offline := range []bool{false, true} {
		t.Run(map[bool]string{false: "online", true: "offline"}[offline], func(t *testing.T) {
			r, _ := stageFreshFixture(t)
			if out, err := runStageTest(t, r[1], ":0", false); err != nil {
				t.Fatalf("stage: %v\n%s", err, out)
			}
			if offline {
				testfix.Git(t, r[0], "remote", "set-url", "origin", filepath.Join(t.TempDir(), "missing.git"))
			}
			before := slotState(t, r[0])
			var out bytes.Buffer
			err := runUnstage(context.Background(), r[1], ":0", "", true, &out, &out)
			want := "1 behind"
			if offline {
				want = "unverified"
			}
			if err != nil || !strings.Contains(out.String(), want) {
				t.Fatalf("dry run should report %s: %v\n%s", want, err, out.String())
			}
			if slotState(t, r[0]) != before {
				t.Fatal("dry run changed target/lease")
			}
			if _, err := os.Stat(filepath.Join(r[0], "bin", "tool")); !os.IsNotExist(err) {
				t.Fatalf("dry run built: %v", err)
			}
		})
	}
}

func TestStageClearFailedBuildRetryFetchesAgain(t *testing.T) {
	r, tip := stageFreshFixture(t)
	if out, err := runStageTest(t, r[1], ":0", false); err != nil {
		t.Fatalf("stage: %v\n%s", err, out)
	}
	procedureWrite(t, r[0], "Makefile", "build:\n\texit 3\n")
	if out, err := runUnstageTest(t, r[1], ":0"); err == nil || !strings.Contains(err.Error(), "rerun sdlc stage --clear") || procedureHead(t, r[0]) != tip {
		t.Fatalf("failed build must retain fresh resting checkout and retry: %v\n%s", err, out)
	}
	testfix.Git(t, r[2], "switch", "-qc", "newer-main", tip)
	testfix.Git(t, r[2], "commit", "--allow-empty", "-qm", "new main after failed build")
	testfix.Git(t, r[2], "push", "-q", "origin", "HEAD:main")
	newTip := procedureHead(t, r[2])
	procedureWrite(t, r[0], "Makefile", "build:\n\tmkdir -p bin && git rev-parse HEAD > bin/tool\n")
	if out, err := runUnstageTest(t, r[1], ":0"); err != nil || procedureHead(t, r[0]) != newTip || !strings.Contains(out, "vcs.revision="+newTip) {
		t.Fatalf("retry built stale main: %v\n%s", err, out)
	}
}

func TestStageClearPreservesIncomingFileCollisions(t *testing.T) {
	for _, ignored := range []bool{false, true} {
		t.Run(map[bool]string{false: "untracked", true: "ignored"}[ignored], func(t *testing.T) {
			r, _ := stageFreshFixture(t)
			procedureWrite(t, r[0], "new-main", "operator scratch\n")
			if ignored {
				gitDir := strings.TrimSpace(testfix.Capture(t, r[0], "rev-parse", "--path-format=absolute", "--git-common-dir"))
				procedureWrite(t, gitDir, "info/exclude", "new-main\n")
			}
			if out, err := runStageTest(t, r[1], ":0", false); err != nil {
				t.Fatalf("stage: %v\n%s", err, out)
			}
			if out, err := runUnstageTest(t, r[1], ":0"); err == nil || !strings.Contains(err.Error(), "no rebuild") {
				t.Fatalf("accepted collision: %v\n%s", err, out)
			}
			if got, err := os.ReadFile(filepath.Join(r[0], "new-main")); err != nil || string(got) != "operator scratch\n" {
				t.Fatalf("lost scratch: %q, %v", got, err)
			}
			if _, err := os.Stat(filepath.Join(r[0], "bin", "tool")); !os.IsNotExist(err) {
				t.Fatalf("collision ran build: %v", err)
			}
		})
	}
}

func TestStageClearBuildUsesPinDespiteLaterRemoteAdvance(t *testing.T) {
	r, tip := stageFreshFixture(t)
	testfix.Git(t, r[2], "fetch", "-q", "origin")
	testfix.Git(t, r[2], "switch", "-qc", "later-main", tip)
	testfix.Git(t, r[2], "commit", "--allow-empty", "-qm", "later main")
	later := procedureHead(t, r[2])
	testfix.Git(t, r[2], "push", "-q", "origin", "HEAD:refs/heads/later-main")
	remote := strings.TrimSpace(testfix.Capture(t, r[0], "remote", "get-url", "origin"))
	t.Setenv("STAGE_TEST_REMOTE", remote)
	t.Setenv("STAGE_TEST_LATER", later)
	if out, err := runStageTest(t, r[1], ":0", false); err != nil {
		t.Fatalf("stage: %v\n%s", err, out)
	}
	gitDir := strings.TrimSpace(testfix.Capture(t, r[0], "rev-parse", "--path-format=absolute", "--git-common-dir"))
	hook := "hooks/post-checkout"
	procedureWrite(t, gitDir, hook, "#!/bin/sh\nexec git --git-dir=\"$STAGE_TEST_REMOTE\" update-ref refs/heads/main \"$STAGE_TEST_LATER\"\n")
	if err := os.Chmod(filepath.Join(gitDir, hook), 0755); err != nil {
		t.Fatal(err)
	}
	if out, err := runUnstageTest(t, r[1], ":0"); err != nil || procedureHead(t, r[0]) != tip || !strings.Contains(out, "vcs.revision="+tip) {
		t.Fatalf("clear did not preserve its fetched pin: %v\n%s", err, out)
	}
	if got := strings.TrimSpace(testfix.Capture(t, r[0], "ls-remote", "origin", "refs/heads/main")); !strings.HasPrefix(got, later) {
		t.Fatalf("hook did not advance the real private remote: %s", got)
	}
}

func TestStageClearNumberedRestingBranch(t *testing.T) {
	r, tip := stageFreshFixture(t)
	procedureWrite(t, r[2], "Makefile", "build:\n\tmkdir -p bin && git rev-parse HEAD > bin/tool\n")
	if out, err := runStageTest(t, r[1], ":2", false); err != nil {
		t.Fatalf("stage: %v\n%s", err, out)
	}
	if out, err := runUnstageTest(t, r[1], ":2"); err != nil || procedureHead(t, r[2]) != tip || !strings.Contains(out, "vcs.revision="+tip) {
		t.Fatalf("numbered clear did not build fetched main: %v\n%s", err, out)
	}
	if got := strings.TrimSpace(testfix.Capture(t, r[2], "branch", "--show-current")); got != "main-slot2" {
		t.Fatalf("numbered target returned to %q", got)
	}
}
