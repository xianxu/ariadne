package main

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
)

func TestPreCloseCheck(t *testing.T) {
	for _, tc := range []struct {
		trains, affected bool
		want             string
	}{
		{true, true, affectedCheckName},
		{true, false, suiteCheckName},
		{false, true, suiteCheckName},
		{false, false, suiteCheckName},
	} {
		got, why := preCloseCheck(tc.trains, tc.affected)
		if got != tc.want || why == "" {
			t.Errorf("preCloseCheck(trains %v, affected %v) = %q (%q), want %q with a reason", tc.trains, tc.affected, got, why, tc.want)
		}
	}
}

// pair's shape resolves to affected; ariadne's (no landing.json), a serial
// line, and an unreadable file all resolve to suite.
func TestResolvePreCloseCheck(t *testing.T) {
	for _, tc := range []struct {
		name, landing, affectedJSON, want, why string
	}{
		{"pair", `{"train_interval_minutes": 60}`, `{"version":1,"command":["make","test-affected"]}`, affectedCheckName, "the train runs the full suite"},
		{"ariadne", "", "", suiteCheckName, "trains are off"},
		{"serial with affected", `{"train_interval_minutes": 0}`, `{"version":1,"command":["make","test-affected"]}`, suiteCheckName, "trains are off"},
		{"trains without affected", `{"train_interval_minutes": 60}`, "", suiteCheckName, "declares no"},
		{"unreadable landing", `{"train_interval_minutes": "soon"}`, `{"version":1,"command":["make","test-affected"]}`, suiteCheckName, "unreadable"},
		{"unreadable affected", `{"train_interval_minutes": 60}`, `{"version":`, suiteCheckName, "unreadable"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root := testfix.Repo(t, testfix.InitialCommit())
			if tc.landing != "" {
				procedureWrite(t, root, landingConfigPath, tc.landing+"\n")
			}
			if tc.affectedJSON != "" {
				procedureWrite(t, root, affectedConfigPath, tc.affectedJSON+"\n")
			}
			git(t, root, "add", "-A")
			git(t, root, "commit", "-qm", "config", "--allow-empty")
			gitCall := func(args ...string) (string, error) { return landingGit(execGitRunner{}, root, args...) }
			got, why := resolvePreCloseCheck(gitCall, "HEAD")
			if got != tc.want || !strings.Contains(why, tc.why) {
				t.Fatalf("resolved %q (%q), want %q with %q", got, why, tc.want, tc.why)
			}
		})
	}
}

// --pre-close chooses the check itself, so it never combines with a choice.
func TestPreCloseFlagExcludesAChosenCheck(t *testing.T) {
	for _, extra := range [][]string{{"--name", "suite"}, {"--all"}, {"--boarding"}} {
		cmd := NewLocalCheckCmd()
		cmd.SetArgs(append([]string{"run", "--pre-close"}, extra...))
		cmd.SetOut(&strings.Builder{})
		cmd.SetErr(&strings.Builder{})
		if err := cmd.Execute(); err == nil || !strings.Contains(err.Error(), "pre-close") {
			t.Errorf("--pre-close %v: %v", extra, err)
		}
	}
}

// The CLI runs what the resolver picks (#578): pair's shape runs and records
// affected with no suite, and the same head without trains runs the suite as
// before.
func TestLocalCheckCmdPreClose(t *testing.T) {
	for _, trains := range []bool{true, false} {
		files := map[string]string{
			affectedConfigPath:     `{"version":1,"command":["make","test-affected"],"groups":[{"name":"fixture","paths":["feature.txt"]}]}`,
			localChecksPath:        `{"version":1,"checks":[{"name":"suite","command":["sh","-c","echo suite"],"required":true}]}`,
			"Makefile":             "test-affected:\n\t@echo affected\n",
			"Makefile.workflow":    "# fake\n",
			"scripts/test-lock.py": fakeSuiteLockHelper("print('acquired')\n"),
			"go.mod":               "module example.com/m\n\ngo 1.22\n",
		}
		if trains {
			files[landingConfigPath] = `{"train_interval_minutes": 60}`
		}
		dir := tempRepoWith(t, files)
		head := strings.TrimSpace(git(t, dir, "rev-parse", "HEAD"))
		ci := newCIFakeGH()
		ci.setHead(5, head)
		swapMergeDeps(t, &e2eGH{openPR: "5", ciSeamGH: ciSeamGH{ci: ci}}, nil)
		var stderr bytes.Buffer
		if err := runLocalCheckCmd(context.Background(), &stderr, localCheckFlags{PreClose: true, Rev: "HEAD"}); err != nil {
			t.Fatalf("trains %v: %v\n%s", trains, err, &stderr)
		}
		want, other := suiteCheckName, affectedCheckName
		if trains {
			want, other = affectedCheckName, suiteCheckName
		}
		if ci.statuses[head]["sdlc-local/"+want] != "success" || ci.statuses[head]["sdlc-local/"+other] != "" ||
			!strings.Contains(stderr.String(), "--pre-close: running "+want) {
			t.Fatalf("trains %v: statuses %v\n%s", trains, ci.statuses[head], &stderr)
		}
	}
}
