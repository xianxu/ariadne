package main

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/bench"
)

func benchObserverFixture(t *testing.T) (*benchStore, benchRunFlags, benchObserverDeps, *int) {
	t.Helper()
	task := preparedFixtureTask(t)
	s, err := prepareFixture(t, task)
	if err != nil {
		t.Fatal(err)
	}
	record, _, err := s.read()
	if err != nil {
		t.Fatal(err)
	}
	provision := benchProvision{Version: 1, RunID: record.ID, ProvidedBy: "operator", Evidence: "synthetic fixture: exact launch settings and hub claim verified", HubClaimConfirmed: true, Hub: task.TL, Worker: task.Worker}
	profilePath := filepath.Join(t.TempDir(), "provision.json")
	raw, err := json.Marshal(provision)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(profilePath, raw, 0600); err != nil {
		t.Fatal(err)
	}
	f := benchRunFlags{Hub: "bench-hub:0", Worker: "bench-work:0", Budget: time.Minute, Provision: profilePath, Ledger: filepath.Join(t.TempDir(), "ledger.jsonl")}
	if err := os.WriteFile(filepath.Join(s.root, "public", "hub", ".hub"), []byte(f.Worker+"\n"), 0600); err != nil {
		t.Fatal(err)
	}
	sends := 0
	d := benchObserverDeps{Couch: benchCouchDeps{Command: func(_ context.Context, args ...string) ([]byte, error) {
		actors := []benchActorSnapshot{}
		for i, slot := range []string{f.Hub, f.Worker} {
			name := "hub"
			if i == 1 {
				name = "work"
			}
			actors = append(actors, benchActorSnapshot{Known: true, Supported: true, Binding: benchActorBinding{Slot: slot, Repository: filepath.Join(s.root, "public", name, ".git"), Scope: name + "-scope", Tag: name + "-tag", Session: name + "-session", Nonce: name + "-nonce", Agent: task.Worker.Agent, PID: 123 + i, Start: "owned-start"}})
		}
		switch args[0] {
		case "--actors":
			return json.Marshal(benchActorsReply{Code: "ok", Actors: actors})
		case "--peek":
			var b benchActorBinding
			for _, a := range actors {
				if a.Binding.Slot == args[1] {
					b = a.Binding
				}
			}
			p := benchPeekReply{Ref: b.Slot, Agent: b.Agent, WorkingPath: filepath.Dir(b.Repository), NativeSession: b.Session + "-native", Transcripts: []string{filepath.Join(t.TempDir(), "source.jsonl")}}
			p.Session.Name = b.Session
			p.Session.Scope = b.Scope
			p.Session.Tag = b.Tag
			p.Session.Nonce = b.Nonce
			return json.Marshal(p)
		case "--send-to":
			sends++
			return nil, errors.New("connection lost after uncertain send")
		default:
			return nil, errors.New("unexpected Couch operation")
		}
	}}}
	return s, f, d, &sends
}

func TestBenchRuntimeUnconfirmedDispatchNeverResends(t *testing.T) {
	s, f, d, sends := benchObserverFixture(t)
	if err := runBenchObserver(context.Background(), s, f, d); err == nil {
		t.Fatal("uncertain dispatch reported success")
	}
	_, state, err := s.read()
	if err != nil || state.Phase() != bench.PhaseDispatchUnknown {
		t.Fatalf("unsafe state %s %v", state.Phase(), err)
	}
	if err := runBenchObserver(context.Background(), s, f, d); err == nil {
		t.Fatal("uncertain run automatically resumed dispatch")
	}
	if *sends != 1 {
		t.Fatalf("dispatched %d times", *sends)
	}
}

func TestBenchRuntimeConfirmedReceiptWaitsThenRequiresRetirement(t *testing.T) {
	s, f, d, _ := benchObserverFixture(t)
	base := d.Couch.Command
	receiptFile := filepath.Join(t.TempDir(), "receipt.json")
	var target benchActorBinding
	raw, err := base(context.Background(), "--actors", "--json")
	if err != nil {
		t.Fatal(err)
	}
	var census benchActorsReply
	if err := json.Unmarshal(raw, &census); err != nil {
		t.Fatal(err)
	}
	for _, a := range census.Actors {
		if a.Binding.Slot == f.Hub {
			target = a.Binding
		}
	}
	dispatches := 0
	d.Couch.Command = func(ctx context.Context, args ...string) ([]byte, error) {
		switch args[0] {
		case "--send-to":
			dispatches++
			reply := map[string]any{"Code": "ok", "Error": "", "Receipt": map[string]any{"Status": "queued", "Message": map[string]any{"ID": "receipt-1", "To": target, "Body": args[3]}}}
			raw, err := json.Marshal(reply)
			if err != nil {
				return nil, err
			}
			if err := os.WriteFile(receiptFile, raw, 0600); err != nil {
				return nil, err
			}
			return raw, nil
		case "--message-status":
			raw, err := os.ReadFile(receiptFile)
			if err != nil {
				return nil, err
			}
			var reply map[string]any
			if err := json.Unmarshal(raw, &reply); err != nil {
				return nil, err
			}
			reply["Receipt"].(map[string]any)["Status"] = "submitted"
			return json.Marshal(reply)
		default:
			return base(ctx, args...)
		}
	}
	now := time.Now().UTC()
	d.Now = func() time.Time { return now }
	d.Wait = func(context.Context, time.Duration) error { now = now.Add(2 * time.Minute); return nil }
	if err := runBenchObserver(context.Background(), s, f, d); err == nil {
		t.Fatal("turn completion invented process retirement")
	}
	r, state, err := s.read()
	if err != nil || state.Phase() != bench.PhaseReconciliation || state.RuntimeStopped() {
		t.Fatalf("receipt wait/timeout state %s quiet=%v %v", state.Phase(), state.RuntimeStopped(), err)
	}
	confirmed := false
	for _, raw := range r.Events {
		var e bench.Event
		if err := json.Unmarshal(raw, &e); err != nil {
			t.Fatal(err)
		}
		if e.Kind == bench.DispatchConfirmed {
			confirmed = true
		}
	}
	if !confirmed || dispatches < 1 || dispatches > 2 {
		t.Fatalf("missing receipt confirmation or repeated dispatch: %v sends %d", confirmed, dispatches)
	}
}

func TestBenchRuntimeRefusesChangedNativeConversation(t *testing.T) {
	s, f, d, _ := benchObserverFixture(t)
	_ = runBenchObserver(context.Background(), s, f, d)
	r, state, err := s.read()
	if err != nil {
		t.Fatal(err)
	}
	runtime, err := loadBenchRuntime(r, state)
	if err != nil {
		t.Fatal(err)
	}
	if err := verifyBenchConversations(context.Background(), d.Couch, runtime); err != nil {
		t.Fatal(err)
	}
	base := d.Couch.Command
	d.Couch.Command = func(ctx context.Context, args ...string) ([]byte, error) {
		raw, err := base(ctx, args...)
		if err == nil && args[0] == "--peek" {
			var p benchPeekReply
			if err := json.Unmarshal(raw, &p); err != nil {
				return nil, err
			}
			p.NativeSession = "new-conversation-in-same-ui"
			return json.Marshal(p)
		}
		return raw, err
	}
	if err := verifyBenchConversations(context.Background(), d.Couch, runtime); err == nil {
		t.Fatal("new conversation inherited prior candidate authority")
	}
}

func TestBenchRuntimePartialBindingNamesSafeRetirement(t *testing.T) {
	s, f, d, sends := benchObserverFixture(t)
	if err := s.putData("runtime", map[string]bool{"partial": true}); err != nil {
		t.Fatal(err)
	}
	err := runBenchObserver(context.Background(), s, f, d)
	if err == nil || !strings.Contains(err.Error(), "bench cancel") || *sends != 0 {
		t.Fatal("partial pre-dispatch binding has no safe recovery", err, *sends)
	}
}
