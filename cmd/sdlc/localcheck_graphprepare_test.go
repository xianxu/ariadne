package main

import (
	"bytes"
	"context"
	"fmt"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/affected"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// #517 / pair#738: native go list sees only a fresh same-head checkout.
// Main's preparation declaration wins even when the candidate edits its copy.
func TestAffectedGraphPreparationFromPinnedMain(t *testing.T) {
	events := filepath.Join(t.TempDir(), "prepare-events")
	selected := filepath.Join(t.TempDir(), "selected")
	candidate := filepath.Join(t.TempDir(), "candidate-command")
	t.Setenv("PREP_EVENTS", events)
	t.Setenv("SELECTED", selected)
	t.Setenv("CANDIDATE_COMMAND", candidate)
	const cfg = `{"version":1,"command":["make","test-affected"],"graph_prepare":{"command":["sh","prepare.sh"]},"groups":[{"name":"fixture","paths":["feature.txt",".sdlc/**"]}]}`
	dir := tempRepoWith(t, map[string]string{
		affectedConfigPath:     cfg,
		localChecksPath:        `{"version":1,"checks":[{"name":"suite","command":["sh","-c","echo suite"],"required":true}]}`,
		"Makefile.workflow":    "# fake\n",
		"scripts/test-lock.py": fakeSuiteLockHelper("print('acquired')\n"),
		"go.mod":               "module example.com/m\n\ngo 1.22\n",
		".gitignore":           "leaf/assets/\n",
		"leaf/leaf.go":         "package leaf\nimport _ \"embed\"\n//go:embed assets/data.txt\nvar Data string\n",
		"leaf/version.txt":     "main\n",
		"mid/mid.go":           "package mid\nimport \"example.com/m/leaf\"\nvar Data = leaf.Data\n",
		"prepare.sh":           "test ! -e leaf/assets/data.txt || exit 40\nmkdir -p leaf/assets\ncat leaf/version.txt > leaf/assets/data.txt\nprintf '%s|%s|%s\\n' \"$PWD\" \"$(git rev-parse HEAD)\" \"$(cat leaf/assets/data.txt)\" >> \"$PREP_EVENTS\"\n",
		"Makefile":             "test-affected:\n\t@printf '%s' \"$$SDLC_AFFECTED_GO_PKGS\" > \"$$SELECTED\"\n",
	})
	// Poison the initiating checkout's ignored output; neither graph nor prepare
	// may use it. Change the requested head's input and its untrusted declaration.
	writeRepoFile(t, dir, "leaf/version.txt", "requested-head\n")
	writeRepoFile(t, dir, affectedConfigPath, strings.Replace(cfg, `"sh","prepare.sh"`, `"sh","-c","touch $CANDIDATE_COMMAND"`, 1))
	git(t, dir, "add", "-A")
	git(t, dir, "commit", "-qm", "candidate inputs and untrusted preparation")
	git(t, dir, "push", "-q", "origin", "feature")
	head := strings.TrimSpace(git(t, dir, "rev-parse", "HEAD"))
	writeRepoFile(t, dir, "leaf/assets/data.txt", "caller-poison\n")
	ci := newCIFakeGH()
	ci.setHead(5, head)
	swapMergeDeps(t, &e2eGH{openPR: "5", ciSeamGH: ciSeamGH{ci: ci}}, nil)
	// An external fetch can move origin/main after it was observed. Advance it
	// to the hostile candidate as soon as the command resolves its main pin.
	advanced := false
	mergeRunner = &graphPrepareAdvanceMainRunner{gitRunner: mergeRunner, advance: func() {
		git(t, dir, "update-ref", "refs/remotes/origin/main", head)
		advanced = true
	}}
	var stderr bytes.Buffer
	err := runLocalCheckCmd(context.Background(), &stderr, localCheckFlags{Names: []string{affectedCheckName}, Rev: "HEAD"})
	if err != nil {
		t.Fatalf("affected failed: %v\n%s", err, &stderr)
	}
	if !advanced {
		t.Fatal("main was never pinned before reading declarations")
	}
	if ci.statuses[head]["sdlc-local/affected"] != "success" || ci.statuses[head]["sdlc-local/suite"] != "" {
		t.Fatalf("statuses %v\n%s", ci.statuses, &stderr)
	}
	got, err := os.ReadFile(selected)
	if err != nil || string(got) != "example.com/m/leaf example.com/m/mid" {
		t.Fatalf("selection %q: %v", got, err)
	}
	raw, err := os.ReadFile(events)
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(string(raw)), "\n")
	if len(lines) != 1 {
		t.Fatalf("prepare ran %d times: %q", len(lines), raw)
	}
	fields := strings.Split(lines[0], "|")
	if len(fields) != 3 || fields[0] == dir || fields[1] != head || fields[2] != "requested-head" {
		t.Fatalf("prepare provenance %q; head %s", raw, head)
	}
	if exists(fields[0]) || exists(candidate) {
		t.Fatalf("checkout left behind or candidate declaration ran: %q", raw)
	}
	poison, err := os.ReadFile(filepath.Join(dir, "leaf/assets/data.txt"))
	if err != nil || string(poison) != "caller-poison\n" {
		t.Fatalf("caller changed: %q %v", poison, err)
	}
}

type graphPrepareAdvanceMainRunner struct {
	gitRunner
	advance func()
}

func (r *graphPrepareAdvanceMainRunner) GitInDir(dir string, args ...string) ([]byte, error) {
	out, err := r.gitRunner.GitInDir(dir, args...)
	if err == nil && strings.Join(args, " ") == "rev-parse --verify origin/main^{commit}" && r.advance != nil {
		advance := r.advance
		r.advance = nil
		advance()
	}
	return out, err
}

func TestAffectedGraphPreparationFallbacks(t *testing.T) {
	for _, tc := range []struct {
		name    string
		command []string
		broken  bool
		want    string
	}{
		{"omitted", nil, false, ""},
		{"nonzero", []string{"sh", "-c", "exit 7"}, false, "graph preparation: exit status 7"},
		{"launch failure", []string{"./missing-generator"}, false, "graph preparation:"},
		{"timeout", []string{"sh", "-c", "exec sleep 60"}, false, "graph preparation: context deadline exceeded"},
		{"still broken graph", []string{"true"}, true, "go list"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			d, gh, change, out, marker := affectedFixture(t)
			files := map[string]string{"leaf/leaf.go": "package leaf\nfunc F() int { return 2 }\n"}
			if tc.broken {
				files["leaf/leaf.go"] = "not Go\n"
			}
			sha := change(files)
			cfg := affectedFixtureConfig(t)
			if tc.command != nil {
				cfg.GraphPrepare = &affected.Preparation{Command: tc.command, TimeoutMinutes: 30}
			}
			if tc.name == "timeout" {
				cfg.GraphPrepare.TimeoutMinutes = 1
				d.minute = 100 * time.Millisecond
			}
			rec, err := runAffectedCheck(context.Background(), d, productionAffectedDeps(d, "main", cfg), cfg, affectedFixtureSuite, sha, affectedOrdinary)
			if err != nil || !rec.passed() {
				t.Fatalf("rec %+v err %v\n%s", rec, err, d.Stderr)
			}
			if tc.want == "" {
				if rec.Name != affectedCheckName || !exists(out) || exists(marker) {
					t.Fatalf("omission changed behavior: %+v\n%s", rec, d.Stderr)
				}
			} else if rec.Name != suiteCheckName || !exists(marker) || exists(out) || gh.statuses[sha]["sdlc-local/affected"] != "" || !strings.Contains(fmt.Sprint(d.Stderr), tc.want) {
				t.Fatalf("unsafe or unexplained fallback: %+v statuses %v\n%s", rec, gh.statuses, d.Stderr)
			}
		})
	}
}
