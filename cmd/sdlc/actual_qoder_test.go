package main

import (
	"bytes"
	"context"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/transcripts"
)

// This fixture changes only the recorded harness. Ownership, command, close,
// prefix replay and reclaim all run through their production shared paths.
func TestQoderActualCloseOwnershipAndRetainedReplay(t *testing.T) {
	for _, reclaim := range []bool{false, true} {
		t.Run(fmt.Sprint(reclaim), func(t *testing.T) {
			r, cardPath, ownerRoot, owner, codexPath := ownedResolverFixture(t)
			if err := os.Remove(codexPath); err != nil {
				t.Fatal(err)
			}
			path := filepath.Join(os.Getenv("HOME"), ".qoder", "projects", "misleading-slug", "unrelated-filename.jsonl")
			if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
				t.Fatal(err)
			}
			data := fmt.Sprintf(`{"type":"workspace-directories","sessionId":"qoder-owned","directories":[%q]}
{"type":"runtime-config","sessionId":"qoder-owned","timestamp":1791936000000,"model":"qoder-test"}
`, owner.Worktree)
			for _, minute := range []int{2, 3, 7, 8} {
				data += fmt.Sprintf(`{"type":"user","sessionId":"qoder-owned","timestamp":"2026-10-14T00:%02d:00Z","parentUuid":"message-not-session","message":{"content":"#335"}}`+"\n", minute)
			}
			if err := os.WriteFile(path, []byte(data), 0600); err != nil {
				t.Fatal(err)
			}
			wantHours := .1
			if reclaim {
				ownedFixtureAt(t, 4)
				ownedFixtureCard(t, r, cardPath, "qoder-release", func(raw []byte) ([]byte, error) {
					return regexp.MustCompile(`(?m)^claimant:\n(?:[ \t].*\n)*`).ReplaceAll(raw, nil), nil
				})
				ownedFixtureAt(t, 6)
				ownedFixtureCard(t, r, cardPath, "qoder-reclaim", func(raw []byte) ([]byte, error) { return issue.SetCardClaimant(raw, owner) })
				wantHours = 4.0 / 60
			}
			inputs, err := resolveOwnedActualInputs(context.Background(), r.root, "", "335", "")
			if err != nil || inputs == nil || len(inputs.Sessions) != 1 || inputs.Sessions[0].Harness != "qoder" || inputs.Sessions[0].ID != "qoder-owned" || inputs.Sessions[0].ParentID != "" || inputs.Sessions[0].Spawned {
				t.Fatalf("inputs=%+v err=%v", inputs, err)
			}
			measured := computeActualFor(context.Background(), r.root, "", "335", "", "")
			if measured.Status != actualMeasured || math.Abs(measured.Hours-wantHours) > .000001 || measured.Inputs == nil || len(measured.Inputs.Sources) != 1 {
				t.Fatalf("measured=%+v want hours=%v", measured, wantHours)
			}
			// Compare the same pinned code history in another checkout.
			testfix.Git(t, ownerRoot, "merge", "--ff-only", "main")
			// A caller in another checkout measures the historical claimant's workspace.
			t.Chdir(ownerRoot)
			other := computeActualFor(context.Background(), ownerRoot, "", "335", "", "")
			if other.Status != actualMeasured || other.Hours != measured.Hours {
				t.Fatalf("checkout changed hours: %+v original=%+v", other, measured)
			}
			t.Chdir(r.root)
			var output bytes.Buffer
			command := NewActualCmd()
			command.SetOut(&output)
			command.SetErr(&output)
			command.SetArgs([]string{"--issue", "335", "--brain-dir", filepath.Join(t.TempDir(), "no-brain")})
			if err := command.Execute(); err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(output.String(), fmt.Sprintf("%.2fh", wantHours)) {
				t.Fatalf("command disagrees: %s", output.String())
			}
			flags := &closeFlags{Issue: 335, BrainDir: filepath.Join(t.TempDir(), "no-brain")}
			adopted, ok := adoptOmittedActual(&bytes.Buffer{}, flags, "335", "issue")
			if !ok || adopted.Hours != measured.Hours || flags.Actual != fmt.Sprintf("%.2f", wantHours) || flags.Inputs == nil || flags.Inputs.Sources[0].Digest != measured.Inputs.Sources[0].Digest {
				t.Fatalf("close=%+v flags=%+v", adopted, flags)
			}
			// Runtime replay and appended activity do not change retained evidence.
			f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0600)
			if err != nil {
				t.Fatal(err)
			}
			_, err = f.WriteString(`{"type":"runtime-config","sessionId":"qoder-owned","timestamp":1000,"model":"old"}` + "\n" + `{"type":"user","sessionId":"qoder-owned","timestamp":"2026-10-14T00:09:00Z","message":{"content":"#335"}}` + "\n")
			_ = f.Close()
			if err != nil {
				t.Fatal(err)
			}
			found, _, err := transcripts.DiscoverEvidence(context.Background(), []string{owner.Worktree}, transcripts.DefaultHarnesses())
			if err != nil {
				t.Fatal(err)
			}
			replay, err := ownedReplaySources(*measured.Inputs, found)
			if err != nil {
				t.Fatal(err)
			}
			replay.Scope.Issue = "335"
			again := computeOwnedActualFor(context.Background(), r.root, "335", "", replay)
			if again.Status != actualMeasured || again.Hours != measured.Hours {
				t.Fatalf("append changed replay=%+v", again)
			}
			if err := os.Truncate(path, measured.Inputs.Sources[0].Bytes-1); err != nil {
				t.Fatal(err)
			}
			if truncated := computeOwnedActualFor(context.Background(), r.root, "335", "", replay); truncated.Status != actualError {
				t.Fatalf("truncated prefix admitted: %+v", truncated)
			}
		})
	}
}

func TestQoderActualCloseRefusesUnresolvedSourceFailures(t *testing.T) {
	for _, scenario := range []string{"malformed-before-owned", "malformed-before-foreign", "malformed-without-workspace", "oversized-before-owned", "unreadable-before-workspace"} {
		t.Run(scenario, func(t *testing.T) {
			r, _, _, owner, _ := ownedResolverFixture(t)
			path := filepath.Join(os.Getenv("HOME"), ".qoder/projects/fixture/broken.jsonl")
			if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
				t.Fatal(err)
			}
			good := fmt.Sprintf(`{"type":"workspace-directories","sessionId":"good-qoder","directories":[%q]}`+"\n", owner.Worktree)
			for _, minute := range []int{2, 3, 7, 8} {
				good += fmt.Sprintf(`{"type":"user","sessionId":"good-qoder","timestamp":"2026-10-14T00:%02d:00Z","message":{"content":"#335"}}`+"\n", minute)
			}
			if err := os.WriteFile(filepath.Join(filepath.Dir(path), "good.jsonl"), []byte(good), 0600); err != nil {
				t.Fatal(err)
			}
			baseline := computeActualFor(context.Background(), r.root, "", "335", "", "")
			if baseline.Status != actualMeasured || baseline.Hours <= 0 {
				t.Fatalf("fixture lacks good owned source: %+v", baseline)
			}
			workspace := owner.Worktree
			if scenario == "malformed-before-foreign" {
				workspace = filepath.Join(t.TempDir(), "foreign")
			}
			rows := fmt.Sprintf(`{"type":"workspace-directories","sessionId":"broken-qoder","directories":[%q]}
{"type":"user","sessionId":"broken-qoder","timestamp":"2026-10-14T00:03:00Z"}
`, workspace)
			switch scenario {
			case "malformed-before-owned", "malformed-before-foreign":
				rows = "not-json\n" + rows
			case "malformed-without-workspace":
				rows = "not-json\n"
			case "oversized-before-owned":
				rows = strings.Repeat(" ", transcripts.UsageLineLimit+1) + "\n" + rows
			case "unreadable-before-workspace":
				if err := os.Symlink(filepath.Join(t.TempDir(), "missing.jsonl"), path); err != nil {
					t.Fatal(err)
				}
			}
			if scenario != "unreadable-before-workspace" {
				if err := os.WriteFile(path, []byte(rows), 0600); err != nil {
					t.Fatal(err)
				}
			}
			measured := computeActualFor(context.Background(), r.root, "", "335", "", "")
			if measured.Status != actualError {
				t.Errorf("omitted unresolved source and measured partial actual: %+v", measured)
			}
			flags := &closeFlags{Issue: 335, BrainDir: filepath.Join(t.TempDir(), "no-brain")}
			var diagnostics bytes.Buffer
			_, adopted := adoptOmittedActual(&diagnostics, flags, "335", "issue")
			if adopted || flags.Actual != "" || flags.Inputs != nil {
				t.Fatalf("close adopted partial evidence: adopted=%v actual=%q inputs=%+v diagnostics=%s", adopted, flags.Actual, flags.Inputs, diagnostics.String())
			}
		})
	}
}

func TestQoderLegacyActualCloseRefusesUnresolvedSource(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	ownedFixtureAt(t, 0)
	root := testfix.Repo(t, testfix.InitialCommit(), testfix.Chdir())
	physical, err := filepath.EvalSymlinks(root)
	if err != nil {
		t.Fatal(err)
	}
	root = physical
	writeRepoFile(t, root, "workshop/issues/000335-legacy.md", "---\nid: 335\nstatus: working\nstarted: 2026-10-14T00:01:00Z\n---\n# Legacy actual\n")
	ownedFixtureAt(t, 1)
	testfix.Git(t, root, "add", "workshop")
	testfix.Git(t, root, "commit", "-qm", "#335: plan: legacy fixture")
	ownedFixtureAt(t, 10)
	testfix.Git(t, root, "commit", "--allow-empty", "-qm", "#335: work")
	path := filepath.Join(os.Getenv("HOME"), ".qoder/projects/fixture/good.jsonl")
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		t.Fatal(err)
	}
	good := fmt.Sprintf(`{"type":"workspace-directories","sessionId":"good-qoder","directories":[%q]}`+"\n", root)
	for _, minute := range []int{2, 3, 7, 8} {
		good += fmt.Sprintf(`{"type":"user","sessionId":"good-qoder","timestamp":"2026-10-14T00:%02d:00Z","message":{"content":"#335"}}`+"\n", minute)
	}
	if err := os.WriteFile(path, []byte(good), 0600); err != nil {
		t.Fatal(err)
	}
	baseline := computeActualFor(context.Background(), root, "", "335", "", "")
	if baseline.Status != actualMeasured || baseline.Hours <= 0 || !strings.Contains(strings.Join(baseline.Warnings, " "), "legacy no-tracker") {
		t.Fatalf("legacy fixture lacks positive labelled baseline: %+v", baseline)
	}
	foreign := `{"type":"workspace-directories","sessionId":"foreign-qoder","directories":["/foreign"]}
{"type":"user","sessionId":"foreign-qoder","timestamp":"2026-10-14T00:03:00Z","message":{"content":"foreign"}}
{"type":"workspace-directories","sessionId":"foreign-qoder","directories":["/another-foreign"]}
`
	if err := os.WriteFile(filepath.Join(filepath.Dir(path), "foreign.jsonl"), []byte(foreign), 0600); err != nil {
		t.Fatal(err)
	}
	excluded := computeActualFor(context.Background(), root, "", "335", "", "")
	if excluded.Status != actualMeasured || excluded.Hours != baseline.Hours || !strings.Contains(strings.Join(excluded.Warnings, " "), "skipped source") {
		t.Fatalf("proven foreign exclusion lost diagnostic: %+v", excluded)
	}
	bad := "not-json\n" + strings.ReplaceAll(good, "good-qoder", "broken-qoder")
	if err := os.WriteFile(filepath.Join(filepath.Dir(path), "broken.jsonl"), []byte(bad), 0600); err != nil {
		t.Fatal(err)
	}
	measured := computeActualFor(context.Background(), root, "", "335", "", "")
	if measured.Status != actualError {
		t.Errorf("legacy actual silently omitted unresolved Qoder source: %+v", measured)
	}
	flags := &closeFlags{Issue: 335, BrainDir: filepath.Join(t.TempDir(), "no-brain")}
	var diagnostics bytes.Buffer
	_, adopted := adoptOmittedActual(&diagnostics, flags, "335", "issue")
	if adopted || flags.Actual != "" {
		t.Fatalf("legacy close adopted partial evidence: adopted=%v actual=%q diagnostic=%s", adopted, flags.Actual, diagnostics.String())
	}
}
