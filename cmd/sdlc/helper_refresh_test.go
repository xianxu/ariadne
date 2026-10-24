package main

import (
	"bytes"
	"context"
	"errors"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/flakes"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestHelperRefreshPrecedesSuitePreflight(t *testing.T) {
	for _, baseline := range []bool{false, true} {
		t.Run(map[bool]string{false: "recorded", true: "baseline"}[baseline], func(t *testing.T) {
			d, sha, _, out := localRunRepo(t, "acquired")
			d.SkipRecord = baseline
			helper := filepath.Join(d.Root, "scripts/test-lock.py")
			if err := os.WriteFile(helper, []byte("raise Exception('stale preflight')\n"), 0600); err != nil {
				t.Fatal(err)
			}
			calls := 0
			d.refreshSuiteHelper = func(ctx context.Context, root string, w io.Writer) error {
				calls++
				if root != d.Root {
					t.Fatalf("wrong caller %s", root)
				}
				return os.WriteFile(helper, []byte(fakeSuiteLockHelper("print('acquired')\n")), 0600)
			}
			r, err := runLocalCheck(context.Background(), d, localCheck{Name: "suite", Command: []string{"true"}, TimeoutMinutes: 1}, sha)
			if err != nil || !r.passed() || calls != 1 {
				t.Fatalf("refresh not before preflight: %+v %v calls=%d\n%s", r, err, calls, out)
			}
		})
	}
}

func TestHelperRefreshUsesTrainAnchor(t *testing.T) {
	d, _, _, _ := localRunRepo(t, "acquired")
	d.WorkflowAnchor = d.Root
	d.Root = t.TempDir()
	calls := 0
	d.refreshSuiteHelper = func(_ context.Context, root string, _ io.Writer) error {
		calls++
		if root != d.WorkflowAnchor {
			t.Fatalf("refresh of detached contents: %s", root)
		}
		return nil
	}
	if _, err := d.prepareSuiteHelper(context.Background()); err != nil || calls != 1 {
		t.Fatalf("anchor setup: %v calls=%d", err, calls)
	}
}

func TestHelperRefreshCancellationStopsBeforePreflight(t *testing.T) {
	d, sha, gh, _ := localRunRepo(t, "acquired")
	d.refreshSuiteHelper = func(context.Context, string, io.Writer) error { return context.Canceled }
	_, err := runLocalCheck(context.Background(), d, localCheck{Name: "suite", Command: []string{"true"}, TimeoutMinutes: 1}, sha)
	if err == nil || !strings.Contains(err.Error(), "canceled") || len(gh.statuses[sha]) != 0 {
		t.Fatalf("ran after canceled refresh: %v", err)
	}
}

func TestHelperRefreshRunnerResourceExcluded(t *testing.T) {
	d, sha, _, _ := localRunRepo(t, "acquired")
	d.refreshSuiteHelper = func(context.Context, string, io.Writer) error { t.Fatal("runner touched suite supplier"); return nil }
	r, err := runLocalCheck(context.Background(), d, localCheck{Name: "runner", Resource: "runner", Command: []string{"true"}, TimeoutMinutes: 1}, sha)
	if err != nil || !r.passed() {
		t.Fatalf("runner: %+v %v", r, err)
	}
}

func privateHelperRefreshHost(t *testing.T) string {
	t.Helper()
	fleet := t.TempDir()
	primary := testfix.Repo(t, testfix.At(fleet, "host"), testfix.InitialCommit())
	host := filepath.Join(fleet, "worktree", "host-slot1", "host")
	testfix.Git(t, primary, "worktree", "add", "-qb", "main-slot1", host)
	supplier := filepath.Join(filepath.Dir(host), "ariadne")
	writeRepoFile(t, supplier, "Makefile.workflow", "# private supplier\n")
	writeRepoFile(t, supplier, "scripts/test-lock.py", "# fixture\n")
	if err := os.Symlink("../ariadne/Makefile.workflow", filepath.Join(host, "Makefile.workflow")); err != nil {
		t.Fatal(err)
	}
	return host
}

func TestHelperRefreshInstalledExecutableProtocolAndFallback(t *testing.T) {
	for _, mode := range []string{"success", "old", "missing", "canceled"} {
		t.Run(mode, func(t *testing.T) {
			host := privateHelperRefreshHost(t)
			bin := t.TempDir()
			record := filepath.Join(t.TempDir(), "calls")
			t.Setenv("HELPER_REFRESH_CALLS", record)
			if mode != "missing" {
				body := "#!/bin/sh\nprintf '%s\\n' \"$PWD\" \"$@\" \"$WEAVE_DEPENDENCIES_CHECK_ONLY\" >> \"$HELPER_REFRESH_CALLS\"\n"
				switch mode {
				case "success":
					body += "printf 'suite helper refreshed private-fixture\\n'\n"
				case "old":
					body += "printf 'unknown flag\\nusage should be captured\\n' >&2\nexit 2\n"
				case "canceled":
					body += "sleep 60\n"
				}
				if err := os.WriteFile(filepath.Join(bin, "weave"), []byte(body), 0700); err != nil {
					t.Fatal(err)
				}
			}
			if mode == "missing" {
				gitPath, err := exec.LookPath("git")
				if err != nil {
					t.Fatal(err)
				}
				if err := os.Symlink(gitPath, filepath.Join(bin, "git")); err != nil {
					t.Fatal(err)
				}
				t.Setenv("PATH", bin)
			} else {
				t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
			}
			var out bytes.Buffer
			ctx := context.Background()
			var cancel context.CancelFunc
			if mode == "canceled" {
				ctx, cancel = context.WithTimeout(ctx, 100*time.Millisecond)
				defer cancel()
			}
			err := runInstalledHelperRefresh(ctx, host, &out)
			if mode == "canceled" {
				if err == nil {
					t.Fatal("caller cancellation ignored")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if mode == "success" {
				raw, err := os.ReadFile(record)
				want := host + "\nrefresh\n--suite-helper\n1\n"
				if err != nil || string(raw) != want || out.String() != "suite helper refreshed private-fixture\n" {
					t.Fatalf("protocol: %q %v output=%s", raw, err, &out)
				}
			} else if strings.Count(strings.TrimSpace(out.String()), "\n") != 0 || !strings.Contains(out.String(), "repair:") || !strings.Contains(out.String(), "skipped") {
				t.Fatalf("fallback: %s", &out)
			}
		})
	}
}

func TestHelperRefreshSavedResumeLaunchAndProofOnlyRepost(t *testing.T) {
	t.Run("launch", func(t *testing.T) {
		d, sha, _, _ := localRunRepo(t, "skipped")
		dir, err := localCheckLogDir(d)
		if err != nil {
			t.Fatal(err)
		}
		in := resumeCoreFixture(t)
		in.Parent.Repo, in.Parent.SHA, in.Parent.LogPath = d.Repo, sha, "failed/aa.log"
		log := []byte("--- FAIL: TestOne (1.00s)\nFAIL\texample.org/p\t1.00s\n")
		in.Parent.LogSHA256 = resumeRecordDigest(log)
		resumeSavedWrite(t, suiteStoreRoot(dir), in.Parent, log)
		d.CurrentSuitePolicy = func(context.Context) (flakes.Policy, error) { return flakes.PolicyAccept, nil }
		d.FlakeRegistry = gateRegistry(t, "example.org/p.TestOne")
		calls := 0
		d.refreshSuiteHelper = func(context.Context, string, io.Writer) error { calls++; return context.Canceled }
		_, err = resumeSaved(context.Background(), d, suiteCheck("false"), nil, sha, "aa")
		if !errors.Is(err, context.Canceled) || calls != 1 {
			t.Fatalf("resume omitted prelaunch refresh: %v calls=%d", err, calls)
		}
	})
	t.Run("proof-repost", func(t *testing.T) {
		d, sha, _, _ := resumeSavedFixture(t, true)
		d.refreshSuiteHelper = func(context.Context, string, io.Writer) error {
			t.Fatal("proof-only repost refreshed supplier")
			return nil
		}
		r, err := resumeSaved(context.Background(), d, suiteCheck("false"), nil, sha, "aa")
		if err != nil || !r.passed() {
			t.Fatalf("proof-only repost: %+v %v", r, err)
		}
	})
}
