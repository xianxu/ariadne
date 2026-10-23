package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/bench"
)

type benchCleanupRuntime struct {
	actors  []benchActorSnapshot
	unknown bool
	calls   int
	onRead  func()
}

func (f *benchCleanupRuntime) deps() benchCouchDeps {
	return benchCouchDeps{Command: func(_ context.Context, args ...string) ([]byte, error) {
		f.calls++
		if f.onRead != nil {
			f.onRead()
		}
		if f.unknown {
			return nil, errors.New("supervisor unavailable")
		}
		if strings.Join(args, " ") != "--actors --json" {
			return nil, errors.New("unexpected runtime mutation")
		}
		actors := f.actors
		if actors == nil {
			actors = []benchActorSnapshot{}
		}
		return json.Marshal(benchActorsReply{Code: "ok", Actors: actors})
	}}
}

func benchCleanupFixture(t *testing.T, parent, name string, created time.Time) (*benchStore, *benchCleanupRuntime) {
	t.Helper()
	s, err := createBenchStore(filepath.Join(parent, name), benchFixtureTask(t))
	if err != nil {
		t.Fatal(err)
	}
	public := filepath.Join(s.root, "public")
	work, hub := filepath.Join(public, "work"), filepath.Join(public, "hub")
	for _, root := range []string{work, hub} {
		if err := os.MkdirAll(root, 0700); err != nil {
			t.Fatal(err)
		}
		prepareTestGit(t, root, "init", "-q", "-b", "main")
		prepareTestWrite(t, root, "owned.txt", "replay material", 0644)
	}
	r, _, err := s.read()
	if err != nil {
		t.Fatal(err)
	}
	h := bench.RuntimeIdentity{Slot: "ops-bench:0", Session: "hub-session", Generation: "hub-generation", Root: hub, Agent: "codex", Model: "gpt-6.1-sol", Effort: "medium"}
	w := bench.RuntimeIdentity{Slot: "work:0", Session: "worker-session", Generation: "worker-generation", Root: work, Agent: "codex", Model: "gpt-6.1-sol", Effort: "medium"}
	for i, event := range []struct {
		kind bench.EventKind
		op   string
	}{{bench.BeginPrepare, "prepare"}, {bench.Prepared, "prepare"}, {bench.BeginDispatch, "dispatch"}, {bench.DispatchConfirmed, "dispatch"}, {bench.Stop, "cancel"}, {bench.CancelConfirmed, "cancel"}, {bench.BeginRecord, "record"}, {bench.RecordConfirmed, "record"}} {
		if _, _, err := s.apply(bench.Event{ID: fmt.Sprint(i), RunID: r.ID, Kind: event.kind, OperationID: event.op, Hub: h, Worker: w}); err != nil {
			t.Fatal(err)
		}
	}
	a := func(i bench.RuntimeIdentity, pid int) benchRuntimeActor {
		return benchRuntimeActor{Identity: i, NativeSession: i.Session, Transcripts: []string{filepath.Join(s.root, "observer", i.Session+".jsonl")}, PID: pid, ProcessStart: "start", Scope: i.Generation, Tag: i.Slot}
	}
	if err := s.putData("runtime", benchRuntimeState{Version: 1, RunID: r.ID, Hub: a(h, 123), Worker: a(w, 124), Since: created.Format(time.RFC3339Nano), Deadline: created.Add(time.Hour).Format(time.RFC3339Nano), Ledger: filepath.Join(parent, "ledger.jsonl"), Provision: json.RawMessage(`{"trusted":true}`)}); err != nil {
		t.Fatal(err)
	}
	if err := s.putData("retirement", benchRetirement{Version: 1, RunID: r.ID, Hub: h, Worker: w, ConfirmedBy: "fixture operator", ConfirmedAt: created.Add(time.Hour).Format(time.RFC3339Nano), Evidence: "Exact actor generations and scorer/reviewer process trees independently confirmed gone.", ProcessesGone: true, Survivors: []string{}, ReviewerSources: []benchSourceInput{}}); err != nil {
		t.Fatal(err)
	}
	r, _, err = s.read()
	if err != nil {
		t.Fatal(err)
	}
	r.CreatedAt = created.UTC().Format(time.RFC3339Nano)
	if err := s.write(r); err != nil {
		t.Fatal(err)
	}
	return s, &benchCleanupRuntime{}
}

func TestBenchCleanupKeepsRawAndSourceAndRequiresRetirement(t *testing.T) {
	s, f := benchCleanupFixture(t, t.TempDir(), "run", time.Now().UTC())
	r, _, err := s.read()
	if err != nil {
		t.Fatal(err)
	}
	prepareTestWrite(t, r.Task.SourceRepo, "source-sentinel", "keep", 0644)
	if err := os.Symlink(r.Task.SourceRepo, filepath.Join(s.root, "public", "source-link")); err != nil {
		t.Fatal(err)
	}
	f.unknown = true
	if err := cleanBenchRun(context.Background(), s, f.deps()); err == nil {
		t.Fatal("uncertain supervisor allowed cleanup")
	}
	_, state, err := s.read()
	if err != nil || !state.CanClean() {
		t.Fatal("failed precondition changed state", err)
	}
	f.unknown = false
	if err := cleanBenchRun(context.Background(), s, f.deps()); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Lstat(filepath.Join(s.root, "public")); !os.IsNotExist(err) {
		t.Fatal("public tree retained", err)
	}
	if _, state, err := s.read(); err != nil || state.Phase() != bench.PhaseCleaned {
		t.Fatal("raw record removed or phase wrong", err)
	}
	if raw, err := os.ReadFile(filepath.Join(r.Task.SourceRepo, "source-sentinel")); err != nil || string(raw) != "keep" {
		t.Fatal("cleanup followed source symlink", err)
	}
	if err := cleanBenchRun(context.Background(), s, f.deps()); err != nil {
		t.Fatal("repeat clean failed", err)
	}
}

func TestBenchCleanupRefusesBusyReplacementAndRecreatedPublic(t *testing.T) {
	for _, kind := range []string{"busy", "replacement", "symlink", "swapped", "scorer"} {
		t.Run(kind, func(t *testing.T) {
			s, f := benchCleanupFixture(t, t.TempDir(), "run", time.Now().UTC())
			public := filepath.Join(s.root, "public")
			switch kind {
			case "busy":
				release, err := s.acquireRun()
				if err != nil {
					t.Fatal(err)
				}
				defer release()
			case "replacement":
				f.actors = []benchActorSnapshot{{Binding: benchActorBinding{Slot: "new:0", Repository: filepath.Join(public, "work", ".git"), Session: "new", Nonce: "new"}, Known: true, Supported: true}}
			case "symlink":
				if err := os.Rename(public, public+"-original"); err != nil {
					t.Fatal(err)
				}
				if err := os.Symlink(t.TempDir(), public); err != nil {
					t.Fatal(err)
				}
			case "swapped":
				f.onRead = func() {
					f.onRead = nil
					if err := os.Rename(public, public+"-original"); err != nil {
						t.Fatal(err)
					}
					if err := os.Mkdir(public, 0700); err != nil {
						t.Fatal(err)
					}
				}
			case "scorer":
				if err := os.MkdirAll(filepath.Join(s.root, "observer", "score-unknown", "parent", "anchor"), 0700); err != nil {
					t.Fatal(err)
				}
			}
			if err := cleanBenchRun(context.Background(), s, f.deps()); err == nil {
				t.Fatal("unsafe cleanup accepted")
			}
			if _, err := os.Lstat(public); err != nil {
				t.Fatal("refused cleanup removed public", err)
			}
		})
	}
}

func TestBenchSweepOnlyPrunesProvenCleanedRuns(t *testing.T) {
	now := time.Date(2026, 11, 30, 0, 0, 0, 0, time.UTC)
	parent := t.TempDir()
	old, f := benchCleanupFixture(t, parent, "old", now.Add(-31*24*time.Hour))
	if err := cleanBenchRun(context.Background(), old, f.deps()); err != nil {
		t.Fatal(err)
	}
	recent, _ := benchCleanupFixture(t, parent, "recent", now.Add(-time.Hour))
	unknown := filepath.Join(parent, "unrelated")
	if err := os.Mkdir(unknown, 0700); err != nil {
		t.Fatal(err)
	}
	prepareTestWrite(t, unknown, "keep", "unrelated", 0600)
	ledger := filepath.Join(parent, "ledger.jsonl")
	if err := os.WriteFile(ledger, []byte("compact external row\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := sweepBenchRunsAt(context.Background(), parent, f.deps(), now); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(old.root); !os.IsNotExist(err) {
		t.Fatal("expired cleaned record retained", err)
	}
	for _, p := range []string{recent.root, unknown, ledger} {
		if _, err := os.Stat(p); err != nil {
			t.Fatal("retention removed protected/unknown material", p, err)
		}
	}
}

func TestBenchSweepRechecksCustodyAfterFreshRetirementRead(t *testing.T) {
	for _, kind := range []string{"public", "scorer"} {
		t.Run(kind, func(t *testing.T) {
			now := time.Now().UTC()
			s, f := benchCleanupFixture(t, t.TempDir(), "old", now.Add(-31*24*time.Hour))
			if err := cleanBenchRun(context.Background(), s, f.deps()); err != nil {
				t.Fatal(err)
			}
			f.onRead = func() {
				f.onRead = nil
				target := filepath.Join(s.root, "public")
				if kind == "scorer" {
					target = filepath.Join(s.root, "observer", "score-new", "candidate", "anchor")
				}
				if err := os.MkdirAll(target, 0700); err != nil {
					t.Fatal(err)
				}
			}
			if err := sweepBenchRunsAt(context.Background(), filepath.Dir(s.root), f.deps(), now); err == nil {
				t.Fatal("new custody removed during retention")
			}
			if _, err := os.Stat(s.root); err != nil {
				t.Fatal("recreated run material deleted", err)
			}
		})
	}
}

func TestBenchSweepReservesAdmissionAndNamesBusyOwner(t *testing.T) {
	for _, busy := range []bool{false, true} {
		t.Run(fmt.Sprint(busy), func(t *testing.T) {
			now := time.Now().UTC()
			parent := t.TempDir()
			s, f := benchCleanupFixture(t, parent, "oldest-cleaned", now.Add(-time.Hour))
			if err := cleanBenchRun(context.Background(), s, f.deps()); err != nil {
				t.Fatal(err)
			}
			task := benchFixtureTask(t)
			for n := 0; n < 49; n++ {
				if _, err := createBenchStore(filepath.Join(parent, fmt.Sprint(n)), task); err != nil {
					t.Fatal(err)
				}
			}
			if busy {
				release, err := s.acquireRun()
				if err != nil {
					t.Fatal(err)
				}
				defer release()
			}
			err := sweepBenchRunsAt(context.Background(), parent, f.deps(), now)
			if busy {
				if err == nil || !strings.Contains(err.Error(), s.root) {
					t.Fatalf("busy eligible owner not named: %v", err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if _, err := os.Stat(s.root); !os.IsNotExist(err) {
				t.Fatal("cap pressure did not retire oldest cleaned run", err)
			}
		})
	}
}

func TestBenchSweepBoundAndAdmissionPressure(t *testing.T) {
	now := time.Now().UTC()
	t.Run("unknown child bound", func(t *testing.T) {
		parent := t.TempDir()
		for n := 0; n < 1001; n++ {
			if err := os.Mkdir(filepath.Join(parent, fmt.Sprint(n)), 0700); err != nil {
				t.Fatal(err)
			}
		}
		if err := sweepBenchRunsAt(context.Background(), parent, (&benchCleanupRuntime{}).deps(), now); err == nil {
			t.Fatal("unbounded parent inventory scanned")
		}
	})
	t.Run("uncertain records count but never prune", func(t *testing.T) {
		parent := t.TempDir()
		task := benchFixtureTask(t)
		for n := 0; n < 50; n++ {
			if _, err := createBenchStore(filepath.Join(parent, fmt.Sprint(n)), task); err != nil {
				t.Fatal(err)
			}
		}
		if err := sweepBenchRunsAt(context.Background(), parent, (&benchCleanupRuntime{}).deps(), now); err == nil || !strings.Contains(err.Error(), "50") {
			t.Fatalf("admission did not explain retained records: %v", err)
		}
	})
}
