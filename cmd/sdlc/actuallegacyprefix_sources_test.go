package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/activetime"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/gatestate"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/transcripts"
)

func legacySourceSnapshot() gatestate.MeasurementInputs {
	return gatestate.MeasurementInputs{Version: 1, Method: gatestate.MeasurementMethodLegacyV1, Repository: "file:/repo", CodeHead: strings.Repeat("a", 40), Since: "2026-10-01T01:00:00Z", Until: "2026-10-01T02:00:00Z", GapMinutes: 15, CommitWeight: 1}
}

func TestLegacyPrefixSourcesOneDiscoveryAndLinearBinding(t *testing.T) {
	for _, n := range []int{129, 258} {
		t.Run(fmt.Sprint(n), func(t *testing.T) {
			snapshot := legacySourceSnapshot()
			since, _ := time.Parse(time.RFC3339, snapshot.Since)
			until, _ := time.Parse(time.RFC3339, snapshot.Until)
			sessions := make([]transcripts.SessionEvidence, n)
			for i := range sessions {
				sessions[i] = transcripts.SessionEvidence{Harness: []string{"codex", "claude", "qoder"}[i%3], ID: fmt.Sprint(i), File: fmt.Sprintf("/files/%d", i), CWD: "/repo", Digest: strings.Repeat("b", 64), Length: int64(i + 1), Started: since.Add(-time.Hour)}
			}
			sessions[1].Started = since.Add(time.Minute)
			sessions[n-1].Started = until.Add(time.Nanosecond)
			calls := 0
			out, err := discoverLegacyPrefixSources(context.Background(), []string{"/repo", "/brain"}, snapshot, func(ctx context.Context, roots []string, hs []transcripts.Harness) ([]transcripts.SessionEvidence, []string, error) {
				calls++
				if !reflect.DeepEqual(roots, []string{"/repo", "/brain"}) || len(hs) == 0 {
					t.Fatalf("roots/harnesses: %v %v", roots, hs)
				}
				return sessions, []string{"discovery warning"}, nil
			})
			if err != nil {
				t.Fatal(err)
			}
			if calls != 1 || len(out.Files) != n-1 || len(out.Sessions) != n-1 || len(out.Snapshot.Sources) != n-1 || len(out.Prefixes) != n-1 || len(out.Windows) != n-1 {
				t.Fatalf("discovery=%d, bound counts: files=%d sessions=%d sources=%d prefixes=%d windows=%d", calls, len(out.Files), len(out.Sessions), len(out.Snapshot.Sources), len(out.Prefixes), len(out.Windows))
			}
			for i, s := range out.Sessions {
				expected := since
				if s.Started.After(expected) {
					expected = s.Started
				}
				if got := out.Windows[s.File]; len(got) != 1 || !got[0].Start.Equal(expected) || !got[0].End.Equal(until) || !got[0].Began.IsZero() {
					t.Fatalf("windows[%d]=%v", i, got)
				}
				if out.Prefixes[s.File] != (activetime.SourcePrefix{Bytes: s.Length, Digest: s.Digest}) {
					t.Fatalf("pin[%d] lost", i)
				}
				if !reflect.DeepEqual(out.Snapshot.Sources[i], ownedMeasurementSource(s, out.Windows[s.File])) {
					t.Fatal("source pin lost")
				}
			}
			if out.Scope.Head != snapshot.CodeHead || out.Scope.Issue != "" || out.Scope.BranchPoint != "" || len(out.SourceSlots) != 0 || len(out.TrackerRefs) != 0 || len(out.Protected) != 0 || len(out.Snapshot.Engagements) != 0 || len(out.Snapshot.QueueWaits) != 0 {
				t.Fatalf("manufactured ownership: %+v", out)
			}
			if !reflect.DeepEqual(out.Warnings, []string{"discovery warning"}) {
				t.Fatalf("lost warning: %v", out.Warnings)
			}
		})
	}
}

func TestLegacyPrefixSourcesErrors(t *testing.T) {
	snapshot := legacySourceSnapshot()
	sentinel := errors.New("discovery failed")
	_, err := discoverLegacyPrefixSources(context.Background(), nil, snapshot, func(context.Context, []string, []transcripts.Harness) ([]transcripts.SessionEvidence, []string, error) {
		return nil, []string{"warning"}, sentinel
	})
	if !errors.Is(err, sentinel) {
		t.Fatalf("discovery failure lost: %v", err)
	}
	if _, err := bindLegacyPrefixSources(snapshot, nil, nil); err == nil {
		t.Fatal("missing prefix sources admitted")
	}
	for _, bad := range []string{"since", "until", "reverse"} {
		t.Run(bad, func(t *testing.T) {
			m := snapshot
			switch bad {
			case "since":
				m.Since = "bad"
			case "until":
				m.Until = "bad"
			case "reverse":
				m.Until = m.Since
			}
			if _, err := bindLegacyPrefixSources(m, nil, nil); err == nil {
				t.Fatal("invalid interval admitted")
			}
		})
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err = discoverLegacyPrefixSources(ctx, nil, snapshot, func(context.Context, []string, []transcripts.Harness) ([]transcripts.SessionEvidence, []string, error) {
		t.Fatal("cancelled discovery called")
		return nil, nil, nil
	})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("cancellation lost: %v", err)
	}
}

func TestLegacyPrefixSourcesLinkedWorktreeDiscovery(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	root := testfix.Repo(t, testfix.InitialCommit())
	root, err := filepath.EvalSymlinks(root)
	if err != nil {
		t.Fatal(err)
	}
	caller := filepath.Join(t.TempDir(), "caller")
	worker := filepath.Join(t.TempDir(), "worker")
	brain := t.TempDir()
	testfix.Git(t, root, "worktree", "add", "-q", "-b", "prefix-caller", caller)
	testfix.Git(t, root, "worktree", "add", "-q", "-b", "other-worker", worker)
	caller, err = filepath.EvalSymlinks(caller)
	if err != nil {
		t.Fatal(err)
	}
	worker, err = filepath.EvalSymlinks(worker)
	if err != nil {
		t.Fatal(err)
	}
	brain, err = filepath.EvalSymlinks(brain)
	if err != nil {
		t.Fatal(err)
	}
	write := func(path, body string) {
		t.Helper()
		if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(body), 0600); err != nil {
			t.Fatal(err)
		}
	}
	for id, cwd := range map[string]string{"primary": root, "caller": caller, "worker": worker, "brain": brain, "future": root} {
		stamp := "2026-10-01T01:10:00Z"
		if id == "future" {
			stamp = "2026-10-01T02:00:01Z"
		}
		write(filepath.Join(os.Getenv("HOME"), ".codex/sessions", id+".jsonl"), fmt.Sprintf(`{"type":"session_meta","timestamp":%q,"payload":{"id":%q,"cwd":%q}}`+"\n", stamp, id, cwd))
	}
	write(filepath.Join(os.Getenv("HOME"), ".qoder/projects/fixture/qoder.jsonl"), fmt.Sprintf(`{"type":"workspace-directories","sessionId":"qoder","directories":[%q]}
{"type":"user","sessionId":"qoder","timestamp":"2026-10-01T01:20:00Z","message":{"content":"work"}}
`, root))
	claudeDir := strings.NewReplacer("/", "-", ".", "-").Replace(root)
	write(filepath.Join(os.Getenv("HOME"), ".claude/projects", claudeDir, "claude.jsonl"), fmt.Sprintf(`{"type":"user","sessionId":"claude","cwd":%q,"timestamp":"2026-10-01T01:30:00Z","message":{"content":"work"}}`+"\n", root))
	snapshot := legacySourceSnapshot()
	snapshot.CodeHead = strings.TrimSpace(testfix.Git(t, root, "rev-parse", "HEAD"))
	out, err := discoverLegacyPrefixInputs(context.Background(), caller, brain, snapshot)
	if err != nil {
		t.Fatal(err)
	}
	ids := map[string]bool{}
	for _, s := range out.Sessions {
		ids[s.ID] = true
		if s.CWD == worker {
			t.Fatal("unrelated linked worker admitted")
		}
	}
	for _, id := range []string{"primary", "caller", "brain", "claude", "qoder"} {
		if !ids[id] {
			t.Fatalf("missing %s session: %v", id, ids)
		}
	}
	if ids["future"] || ids["worker"] || len(out.Files) != 5 {
		t.Fatalf("unexpected discovered sessions: %v", ids)
	}
}

func TestLegacyPrefixSourcesTemporalBoundaryAndCanonicalRoots(t *testing.T) {
	snapshot := legacySourceSnapshot()
	until, _ := time.Parse(time.RFC3339, snapshot.Until)
	future := transcripts.SessionEvidence{Harness: "codex", ID: "future", File: "/future", Started: until.Add(time.Nanosecond)}
	if _, err := bindLegacyPrefixSources(snapshot, []transcripts.SessionEvidence{future}, nil); err == nil {
		t.Fatal("all-future evidence silently omitted prefix")
	}
	exact := future
	exact.ID = "boundary"
	exact.File = "/boundary"
	exact.Started = until
	if _, err := bindLegacyPrefixSources(snapshot, []transcripts.SessionEvidence{future, exact}, nil); err == nil {
		t.Fatal("zero-width cutover session admitted as prefix evidence")
	}
	root := testfix.Repo(t, testfix.InitialCommit())
	canonical, err := filepath.EvalSymlinks(root)
	if err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(t.TempDir(), "alias")
	if err := os.Symlink(canonical, link); err != nil {
		t.Fatal(err)
	}
	roots, err := legacyPrefixDiscoveryRoots(context.Background(), link, canonical)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(roots, []string{canonical}) {
		t.Fatalf("canonical duplicate roots: %v", roots)
	}
	roots, err = legacyPrefixDiscoveryRoots(context.Background(), canonical, filepath.Join(t.TempDir(), "no-brain"))
	if err != nil || !reflect.DeepEqual(roots, []string{canonical}) {
		t.Fatalf("missing optional brain: %v %v", roots, err)
	}
}
