package main

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/xianxu/ariadne/pkg/layergraph"
)

func commitDependencyFixture(t *testing.T, d localRunDeps, gh *ciFakeGH) string {
	t.Helper()
	for _, args := range [][]string{{"add", "-A"}, {"commit", "-qm", "dependency fixture"}} {
		if out, err := d.Git(args...); err != nil {
			t.Fatalf("git %v: %v %s", args, err, out)
		}
	}
	out, err := d.Git("rev-parse", "HEAD")
	if err != nil {
		t.Fatal(err)
	}
	sha := strings.TrimSpace(out)
	gh.setHead(9, sha)
	return sha
}

func TestRunLocalCheckDependencyEnvironment(t *testing.T) {
	t.Setenv("SDLC_CALLER_ROOT", "stale caller")
	t.Setenv("SDLC_DEP_SUBSTRATE", "stale substrate")
	t.Setenv("SDLC_DEP_DATA", "stale data")
	t.Setenv("SDLC_DEP_OLD", "stale undeclared kind")
	for _, resource := range []string{"suite", "runner"} {
		t.Run(resource, func(t *testing.T) {
			d, _, gh, stderr := localRunRepo(t, "acquired")
			anchor := filepath.Join(t.TempDir(), "caller space")
			base := filepath.Join(filepath.Dir(anchor), "substrate")
			writeRepoFile(t, base, "construct/base.manifest", "# substrate\n")
			writeRepoFile(t, base, "sentinel", "dependency stays unchanged\n")
			writeRepoFile(t, anchor, "construct/deps", "invalid caller edit\n")
			d.WorkflowAnchor = anchor
			writeRepoFile(t, d.Root, "construct/deps", "# tested declaration\nsubstrate ../substrate https://example.invalid/base.git\nsubstrate ../substrate/.\ndata https://example.invalid/data.git data/input\n")
			sha := commitDependencyFixture(t, d, gh)
			c := localCheck{Name: "env-proof", Resource: resource, TimeoutMinutes: 30, Command: []string{"sh", "-c", `printf 'caller=%s\nsubstrate=%s\ndata=%s\nold=%s\ncwd=%s\n' "$SDLC_CALLER_ROOT" "$SDLC_DEP_SUBSTRATE" "$SDLC_DEP_DATA" "${SDLC_DEP_OLD-unset}" "$PWD"`}}
			rec, err := runLocalCheck(context.Background(), d, c, sha)
			if err != nil || !rec.passed() {
				t.Fatalf("record %+v, err %v\n%s", rec, err, stderr)
			}
			b, err := os.ReadFile(rec.LogPath)
			want := "caller=" + anchor + "\nsubstrate=" + base + "\ndata=" + filepath.Join(anchor, "data/input") + "\nold=unset\ncwd="
			if err != nil || !strings.HasPrefix(string(b), want) || strings.Contains(string(b), "cwd="+anchor+"\n") || strings.Contains(string(b), "cwd="+d.Root+"\n") {
				t.Fatalf("want detached execution with %q, got %q (%v)", want, b, err)
			}
			if b, err := os.ReadFile(filepath.Join(base, "sentinel")); err != nil || string(b) != "dependency stays unchanged\n" {
				t.Fatalf("dependency changed: %q %v", b, err)
			}
		})
	}
}

func TestRunLocalCheckDependencyPreparationRefuses(t *testing.T) {
	for _, declaration := range []string{
		"substrate /one\nsubstrate /two\n",
		"data https://one.invalid data/one\ndata https://two.invalid data/two\n",
		"substrate\n",
	} {
		t.Run(strings.ReplaceAll(declaration, "\n", "_"), func(t *testing.T) {
			d, _, gh, stderr := localRunRepo(t, "acquired")
			writeRepoFile(t, d.Root, "construct/deps", declaration)
			sha := commitDependencyFixture(t, d, gh)
			marker := filepath.Join(t.TempDir(), "ran")
			c := localCheck{Name: "env-proof", Command: []string{"./marker-check.sh", marker}, TimeoutMinutes: 30}
			_, err := runLocalCheck(context.Background(), d, c, sha)
			if err == nil || !strings.Contains(err.Error(), "construct/deps") {
				t.Fatalf("expected declaration refusal, got %v\n%s", err, stderr)
			}
			if strings.Contains(declaration, "/two") && (!strings.Contains(err.Error(), "/one") || !strings.Contains(err.Error(), "/two")) {
				t.Fatalf("conflict must name both rows: %v", err)
			}
			if _, err := os.Stat(marker); !os.IsNotExist(err) || len(gh.statuses[sha]) != 0 {
				t.Fatalf("invalid declaration launched or recorded check: marker %v, statuses %v", err, gh.statuses)
			}
		})
	}
}

func TestRunLocalCheckWithoutDependenciesClearsInherited(t *testing.T) {
	t.Setenv("SDLC_CALLER_ROOT", "wrong")
	t.Setenv("SDLC_DEP_SUBSTRATE", "wrong")
	t.Setenv("SDLC_DEP_DATA", "wrong")
	d, sha, _, stderr := localRunRepo(t, "acquired")
	c := localCheck{Name: "env-proof", Command: []string{"sh", "-c", `printf '%s\n' "$SDLC_CALLER_ROOT" "${SDLC_DEP_SUBSTRATE-unset}" "${SDLC_DEP_DATA-unset}"`}, TimeoutMinutes: 30}
	rec, err := runLocalCheck(context.Background(), d, c, sha)
	if err != nil || !rec.passed() {
		t.Fatalf("%+v %v\n%s", rec, err, stderr)
	}
	b, err := os.ReadFile(rec.LogPath)
	if err != nil || string(b) != d.Root+"\nunset\nunset\n" {
		t.Fatalf("environment %q, err %v", b, err)
	}
}

// Replay echo#10's wrapper selection contract without editing or reading a live
// consumer checkout. Only private fixture bytes are ever executed.
func TestRunLocalCheckDetachedSubstrateWrapper(t *testing.T) {
	d, _, gh, stderr := localRunRepo(t, "acquired")
	base := t.TempDir()
	writeRepoFile(t, base, "construct/base.manifest", "# fixture\n")
	writeRepoFile(t, base, "scripts/test/portable-ci.test.sh", "printf 'PASS substrate fixture\\n'\n")
	writeRepoFile(t, d.Root, "construct/deps", "substrate "+base+"\n")
	writeRepoFile(t, d.Root, "wrapper.sh", `set -eu
root="$(pwd -P)"
substrate="${SDLC_DEP_SUBSTRATE:-$root/../ariadne}"
if [ ! -f "$substrate/scripts/test/portable-ci.test.sh" ]; then
  printf 'SKIP: substrate not found: %s\n' "$substrate"
  exit 0
fi
bash "$substrate/scripts/test/portable-ci.test.sh"
`)
	sha := commitDependencyFixture(t, d, gh)
	c := localCheck{Name: "wrapper-proof", Command: []string{"sh", "-ec", "test ! -e ../ariadne; bash wrapper.sh"}, TimeoutMinutes: 30}
	rec, err := runLocalCheck(context.Background(), d, c, sha)
	if err != nil || !rec.passed() {
		t.Fatalf("%+v %v\n%s", rec, err, stderr)
	}
	b, err := os.ReadFile(rec.LogPath)
	if err != nil || string(b) != "PASS substrate fixture\n" {
		t.Fatalf("wrapper failed to use declared substrate without a sibling: %q %v", b, err)
	}
	missing := filepath.Join(t.TempDir(), "absent")
	writeRepoFile(t, d.Root, "construct/deps", "substrate "+missing+"\n")
	sha = commitDependencyFixture(t, d, gh)
	rec, err = runLocalCheck(context.Background(), d, c, sha)
	if err != nil || !rec.passed() {
		t.Fatalf("absent declared path: %+v %v\n%s", rec, err, stderr)
	}
	b, err = os.ReadFile(rec.LogPath)
	if err != nil || string(b) != "SKIP: substrate not found: "+missing+"\n" {
		t.Fatalf("missing path was not exported for the wrapper's skip: %q %v", b, err)
	}
	// Direct make-style invocation has no reserved environment. Preserve the
	// conventional sibling fallback and make an absent explicit path a clear skip.
	fallback := filepath.Join(filepath.Dir(d.Root), "ariadne")
	writeRepoFile(t, fallback, "scripts/test/portable-ci.test.sh", "printf 'PASS fallback fixture\\n'\n")
	for _, tc := range []struct{ substrate, want string }{
		{"", "PASS fallback fixture\n"},
		{filepath.Join(t.TempDir(), "absent"), "SKIP: substrate not found:"},
	} {
		cmd := exec.Command("bash", "wrapper.sh")
		cmd.Dir = d.Root
		cmd.Env = []string{"PATH=" + os.Getenv("PATH")}
		if tc.substrate != "" {
			cmd.Env = append(cmd.Env, "SDLC_DEP_SUBSTRATE="+tc.substrate)
		}
		if b, err := cmd.CombinedOutput(); err != nil || !strings.HasPrefix(string(b), tc.want) {
			t.Fatalf("direct wrapper %q: %q %v", tc.substrate, b, err)
		}
	}
}

func TestLocalCheckDependencyPrimaryFallback(t *testing.T) {
	d, sha, _, _ := localRunRepo(t, "acquired")
	base := filepath.Join(filepath.Dir(d.Root), "provider")
	writeRepoFile(t, base, "construct/base.manifest", "# provider\n")
	caller := filepath.Join(t.TempDir(), "linked", "caller")
	if out, err := d.Git("worktree", "add", "--detach", caller, sha); err != nil {
		t.Fatalf("create caller: %v %s", err, out)
	}
	writeRepoFile(t, d.Root, "construct/deps", "substrate ../provider\n")
	env, err := localCheckEnvironment(caller, d.Root)
	if err != nil || !slices.Contains(env, "SDLC_DEP_SUBSTRATE="+base) {
		t.Fatalf("primary fallback %q not exported: %v", base, err)
	}
}

func TestDependencyEnvironmentProjection(t *testing.T) {
	base := []string{"KEEP=a=b", "SDLC_CALLER_ROOT=old", "SDLC_DEP_OLD=old", "SDLC_CALLER_ROOT_EXTRA=keep"}
	rows := []layergraph.Dependency{{Kind: "data", Path: "/data"}, {Kind: "data", Path: "/data/."}}
	env, err := dependencyEnvironment(base, "/caller", rows)
	want := "KEEP=a=b\nSDLC_CALLER_ROOT_EXTRA=keep\nSDLC_CALLER_ROOT=/caller\nSDLC_DEP_DATA=/data"
	if err != nil || strings.Join(env, "\n") != want || base[1] != "SDLC_CALLER_ROOT=old" {
		t.Fatalf("projection %v, err %v, inherited %v", env, err, base)
	}
}

func TestLocalCheckDependencyBatchEnvironment(t *testing.T) {
	d, sha, _, stderr := localRunRepo(t, "acquired")
	writeRepoFile(t, d.Root, "construct/deps", "data https://example.invalid/data.git data/input\n")
	log := filepath.Join(t.TempDir(), "batch.log")
	argv := []string{"sh", "-c", `printf '%s\n' "$SDLC_CALLER_ROOT" "$SDLC_DEP_DATA"`}
	run, err := runUnderSlotBatch(context.Background(), d, d.Root, filepath.Join(d.Root, "scripts/test-lock.py"), filepath.Dir(log), "suite-rerun", sha, 1, suitePrioritySubject{}, [][]string{argv, argv}, log, time.Minute)
	if err != nil || run.Exit != 0 {
		t.Fatalf("batch %+v %v\n%s", run, err, stderr)
	}
	b, err := os.ReadFile(log)
	want := strings.Repeat(d.Root+"\n"+filepath.Join(d.Root, "data/input")+"\n", 2)
	if err != nil || string(b) != want {
		t.Fatalf("batch env %q, want %q (%v)", b, want, err)
	}
}
