package main

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/pkg/weaveownership"
)

// faultingGit is an executable test double at the same PATH boundary used by
// every production Git interaction. Real Git backs its persistent index state;
// it can fail before an operation or after that operation has taken effect.
func faultingGit(t *testing.T, root, verb string, occurrence int, after bool) {
	t.Helper()
	realGit, err := exec.LookPath("git")
	if err != nil {
		t.Fatal(err)
	}
	state := t.TempDir()
	propagationWrite(t, state, "git", `#!/bin/sh
repo= verb= previous=
for arg in "$@"; do
  if [ "$previous" = -C ]; then repo=$arg; fi
  case "$arg" in status|ls-files|rm|add|commit) verb=$arg ;; esac
  previous=$arg
done
if [ "$repo" != "$PROP_GIT_REPO" ]; then exec "$PROP_REAL_GIT" "$@"; fi
printf '%s\n' "$verb" >> "$PROP_GIT_STATE/events"
count=0
if [ -f "$PROP_GIT_STATE/$verb" ]; then count=$(cat "$PROP_GIT_STATE/$verb"); fi
count=$((count+1))
printf '%s' "$count" > "$PROP_GIT_STATE/$verb"
if [ "$verb" = "$PROP_GIT_FAIL" ] && [ "$count" = "$PROP_GIT_OCCURRENCE" ] && [ ! -f "$PROP_GIT_STATE/fired" ]; then
  touch "$PROP_GIT_STATE/fired"
  if [ "$PROP_GIT_AFTER" = 1 ]; then "$PROP_REAL_GIT" "$@" || exit $?; fi
  echo 'injected Git failure' >&2
  exit 89
fi
exec "$PROP_REAL_GIT" "$@"
`)
	if err := os.Chmod(filepath.Join(state, "git"), 0755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PROP_REAL_GIT", realGit)
	t.Setenv("PROP_GIT_REPO", root)
	t.Setenv("PROP_GIT_STATE", state)
	t.Setenv("PROP_GIT_FAIL", verb)
	t.Setenv("PROP_GIT_OCCURRENCE", strconv.Itoa(occurrence))
	value := "0"
	if after {
		value = "1"
	}
	t.Setenv("PROP_GIT_AFTER", value)
	t.Setenv("PATH", state+string(os.PathListSeparator)+os.Getenv("PATH"))
}

func migrationFailureFixture(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	propagationRepo(t, root)
	propagationWrite(t, root, ".gitignore", "/generated-*\n/authored\n/construct/generated/weave/\n")
	for _, name := range []string{"generated-a", "generated-b", "authored"} {
		propagationWrite(t, root, name, "retained working contents\n")
	}
	git(t, root, "add", "-A")
	git(t, root, "add", "-f", "--", "generated-a", "generated-b", "authored")
	git(t, root, "commit", "-qm", "initial")
	var ids []weaveownership.Identity
	for _, name := range []string{"generated-a", "generated-b"} {
		ids = append(ids, weaveownership.Identity{Path: name, Scope: weaveownership.ScopeArtifacts, Kind: "file", Value: weaveownership.Digest([]byte("retained working contents\n"))})
	}
	data, err := json.Marshal(weaveownership.Inventory{Version: weaveownership.Version, Outputs: ids})
	if err != nil {
		t.Fatal(err)
	}
	propagationWrite(t, root, weaveownership.InventoryPath, string(data))
	return root
}

func TestMigrationGitFailureRetainsEffectsAndCanRetry(t *testing.T) {
	for _, tc := range []struct {
		verb                string
		occurrence, removed int
		after, committed    bool
	}{
		{"ls-files", 1, 0, false, false}, {"ls-files", 1, 0, true, false},
		{"rm", 1, 0, false, false}, {"rm", 2, 1, false, false}, {"rm", 2, 2, true, false},
		{"status", 1, 2, false, false}, {"status", 1, 2, true, false},
		{"add", 1, 2, false, false}, {"add", 1, 2, true, false},
		{"commit", 1, 2, false, false}, {"commit", 1, 2, true, true},
	} {
		t.Run(tc.verb+"/"+strconv.Itoa(tc.occurrence)+"/after="+strconv.FormatBool(tc.after), func(t *testing.T) {
			root := migrationFailureFixture(t)
			// A seed materialized by compile must also survive failed staging/commit.
			propagationWrite(t, root, "entrypoint.sh", "authored seed\n")
			faultingGit(t, root, tc.verb, tc.occurrence, tc.after)
			_, err := commitConsumption(root, "test#239")
			if err == nil {
				t.Fatal("Git failure hidden")
			}
			if !strings.Contains(err.Error(), "inspect git status") {
				t.Errorf("missing partial-progress guidance: %v", err)
			}
			for _, name := range []string{"generated-a", "generated-b", "authored"} {
				data, err := os.ReadFile(filepath.Join(root, name))
				if err != nil || string(data) != "retained working contents\n" {
					t.Fatalf("lost workfile %s: %q %v", name, data, err)
				}
			}
			tracked := git(t, root, "ls-files", "--", "generated-a", "generated-b")
			count := 0
			if tracked != "" {
				count = len(strings.Split(tracked, "\n"))
			}
			if count != 2-tc.removed {
				t.Fatalf("tracked=%q; want %d retained", tracked, 2-tc.removed)
			}
			if got := git(t, root, "ls-files", "--", "authored"); got != "authored" {
				t.Fatal("unrelated ignored file lost")
			}
			wantCommits := "1"
			if tc.committed {
				wantCommits = "2"
			}
			if got := git(t, root, "rev-list", "--count", "HEAD"); got != wantCommits {
				t.Fatalf("commits=%s, want %s", got, wantCommits)
			}
			// Internal retry observes actual index/HEAD state; it does not replay a
			// presumed rollback or duplicate a commit that succeeded before error.
			if _, err := commitConsumption(root, "test#239"); err != nil {
				t.Fatal(err)
			}
			if got := git(t, root, "rev-list", "--count", "HEAD"); got != "2" {
				t.Fatalf("retry lost/duplicated commit: %s", got)
			}
			if got := git(t, root, "status", "--porcelain"); got != "" {
				t.Fatalf("retry left state: %s", got)
			}
		})
	}
}

func TestPropagationAfterPartialIndexFailureRequiresOperatorResolution(t *testing.T) {
	root := migrationFailureFixture(t)
	owner := filepath.Join(filepath.Dir(root), "base")
	if err := os.MkdirAll(owner, 0755); err != nil {
		t.Fatal(err)
	}
	propagationWrite(t, root, "construct/deps", "substrate ../base\n")
	git(t, root, "add", "construct/deps")
	git(t, root, "commit", "-qm", "declare base")
	bin := t.TempDir()
	propagationWrite(t, bin, "weave", "#!/bin/sh\nprintf '%s\\n' \"$1\" >> \"$PROP_WEAVE_EVENTS\"\n")
	if err := os.Chmod(filepath.Join(bin, "weave"), 0755); err != nil {
		t.Fatal(err)
	}
	events := filepath.Join(bin, "events")
	t.Setenv("PROP_WEAVE_EVENTS", events)
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	faultingGit(t, root, "rm", 2, false)
	var out bytes.Buffer
	if err := runPropagateBase(owner, "test#239", false, &out); err == nil {
		t.Fatal("failure hidden")
	}
	first, err := os.ReadFile(events)
	if err != nil {
		t.Fatal(err)
	}
	out.Reset()
	if err := runPropagateBase(owner, "test#239", false, &out); err == nil || !strings.Contains(out.String(), "SKIPPED: dirty") {
		t.Fatalf("error=%v output=%s", err, &out)
	}
	second, err := os.ReadFile(events)
	if err != nil || !bytes.Equal(first, second) {
		t.Fatal("dirty retry executed compiler")
	}
	// Simulate the operator inspecting and committing the retained partial index.
	git(t, root, "commit", "-qm", "reviewed partial migration")
	if err := runPropagateBase(owner, "test#239", false, &out); err != nil {
		t.Fatal(err)
	}
	if got := git(t, root, "ls-files", "--", "generated-a", "generated-b"); got != "" {
		t.Fatalf("remaining migration not completed: %q", got)
	}
	if got := git(t, root, "ls-files", "--", "authored"); got != "authored" {
		t.Fatal("unrelated file lost")
	}
}

func TestPropagationStatusFailureDoesNotRunCompiler(t *testing.T) {
	root := migrationFailureFixture(t)
	owner := filepath.Join(filepath.Dir(root), "base")
	if err := os.MkdirAll(owner, 0755); err != nil {
		t.Fatal(err)
	}
	propagationWrite(t, root, "construct/deps", "substrate ../base\n")
	git(t, root, "add", "construct/deps")
	git(t, root, "commit", "-qm", "declare base")
	bin := t.TempDir()
	propagationWrite(t, bin, "weave", "#!/bin/sh\ntouch \"$PROP_WEAVE_CALLED\"\n")
	if err := os.Chmod(filepath.Join(bin, "weave"), 0755); err != nil {
		t.Fatal(err)
	}
	called := filepath.Join(bin, "called")
	t.Setenv("PROP_WEAVE_CALLED", called)
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	faultingGit(t, root, "status", 1, false)
	var out bytes.Buffer
	if err := runPropagateBase(owner, "test#239", false, &out); err == nil || !strings.Contains(out.String(), "FAILED: git status") {
		t.Fatalf("error=%v output=%s", err, &out)
	}
	if _, err := os.Stat(called); !os.IsNotExist(err) {
		t.Fatal("compiler ran after failed status probe")
	}
	if got := git(t, root, "status", "--porcelain"); got != "" {
		t.Fatalf("failed status probe mutated index: %s", got)
	}
}
