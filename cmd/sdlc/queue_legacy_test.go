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

	"github.com/xianxu/ariadne/cmd/sdlc/internal/flock"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/queuegate"
)

func legacyHeldFile(t *testing.T, path, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		t.Fatal(err)
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		t.Fatal(err)
	}
	if err := flock.Exclusive(f, true); err != nil {
		f.Close()
		t.Fatal(err)
	}
	if _, err := f.WriteString(body); err != nil {
		f.Close()
		t.Fatal(err)
	}
	t.Cleanup(func() { f.Close() })
}
func legacyFixtureOptions(t *testing.T) queueLegacyOptions {
	root := t.TempDir()
	return queueLegacyOptions{FleetRoot: root, LockBase: filepath.Join(root, "suite.lock"), Slots: 2, Private: true, CommonDirs: func(context.Context, string) ([]string, error) { return []string{}, nil }, ProcessStart: func(int) (string, error) { return "incarnation", nil }}
}
func legacyKnown(root string) queuegate.Operation {
	return queuegate.Operation{SchemaVersion: 1, ID: "known", Workspace: "pair:5", Kinds: []string{"suite", "merge", "landing"}, Phase: "waiting", PID: 123, PIDStarted: "incarnation", Managed: true, StartedAt: time.Now(), Root: root}
}

func TestQueueLegacyPendingAdmissionIsIncompleteButNotActive(t *testing.T) {
	for _, resource := range []string{"token", "ticket"} {
		t.Run(resource, func(t *testing.T) {
			o := legacyFixtureOptions(t)
			path := o.LockBase + ".queue"
			if resource == "ticket" {
				path = filepath.Join(o.LockBase+".tickets", "000000000001")
			}
			raw := `{"repo":"old","worktree":"/fixture/old","pid":123,"started":"fixture"}`
			legacyHeldFile(t, path, raw)
			active, details := observeLegacyQueue(context.Background(), o, nil)
			if len(active) != 0 || len(details) == 0 || !strings.Contains(strings.Join(details, " "), path) {
				t.Fatalf("legacy pending admission proved complete or active: %+v %v", active, details)
			}
			data, err := os.ReadFile(path)
			if err != nil || string(data) != raw {
				t.Fatalf("observer mutated waiter: %q %v", data, err)
			}
		})
	}
}

func TestQueueLegacyPendingNewBindingsStayComplete(t *testing.T) {
	o := legacyFixtureOptions(t)
	op := legacyKnown(o.FleetRoot)
	raw, _ := json.Marshal(map[string]any{"queue_operation_id": op.ID, "queue_pid": op.PID, "queue_pid_started": op.PIDStarted, "queue_root": op.Root})
	legacyHeldFile(t, filepath.Join(o.LockBase+".tickets", "000000000001"), string(raw))
	legacyHeldFile(t, o.LockBase+".queue", string(raw))
	active, details := observeLegacyQueue(context.Background(), o, []queuegate.Operation{op})
	if len(active) != 0 || len(details) != 0 {
		t.Fatalf("new waiting binding became unknown: %+v %v", active, details)
	}
	o.ProcessStart = func(int) (string, error) { return "recycled", nil }
	active, details = observeLegacyQueue(context.Background(), o, []queuegate.Operation{op})
	if len(active) != 0 || len(details) == 0 {
		t.Fatalf("unproved pending binding became complete: %+v %v", active, details)
	}
}

func TestQueueLegacyPendingResidueAndAllocationAreNotWaiters(t *testing.T) {
	o := legacyFixtureOptions(t)
	dir := o.LockBase + ".tickets"
	legacyHeldFile(t, filepath.Join(dir, "alloc"), "allocation")
	if err := os.WriteFile(filepath.Join(dir, "000000000001"), []byte("old unlocked residue"), 0600); err != nil {
		t.Fatal(err)
	}
	active, details := observeLegacyQueue(context.Background(), o, nil)
	if len(active) != 0 || len(details) != 0 {
		t.Fatalf("residue or allocation counted as pending: %+v %v", active, details)
	}
}

func TestQueueLegacyPendingDirectoryFailureIsIncomplete(t *testing.T) {
	o := legacyFixtureOptions(t)
	if err := os.WriteFile(o.LockBase+".tickets", []byte("not a directory"), 0600); err != nil {
		t.Fatal(err)
	}
	active, details := observeLegacyQueue(context.Background(), o, nil)
	if len(active) != 0 || len(details) == 0 || !strings.Contains(strings.Join(details, " "), o.LockBase+".tickets") {
		t.Fatalf("unreadable ticket directory proved complete or active: %+v %v", active, details)
	}
}

func TestQueueLegacyPendingBindingsAreIndexedLinearly(t *testing.T) {
	for _, n := range []int{8, 16} {
		t.Run(fmt.Sprint(n), func(t *testing.T) {
			o := legacyFixtureOptions(t)
			known := make([]queuegate.Operation, n)
			cost := map[string]int{}
			o.Cost = func(label string) { cost[label]++ }
			for i := range known {
				known[i] = legacyKnown(o.FleetRoot)
				known[i].ID = fmt.Sprintf("pending-%d", i)
				raw, _ := json.Marshal(map[string]any{"queue_operation_id": known[i].ID, "queue_pid": known[i].PID, "queue_pid_started": known[i].PIDStarted, "queue_root": known[i].Root})
				legacyHeldFile(t, filepath.Join(o.LockBase+".tickets", fmt.Sprintf("%012d", i+1)), string(raw))
			}
			active, details := observeLegacyQueue(context.Background(), o, known)
			if len(active) != 0 || len(details) != 0 || cost["index"] != n || cost["pending"] != n || cost["lookup"] != n {
				t.Fatalf("pending observation: %+v %v cost %v", active, details, cost)
			}
		})
	}
}

func TestQueueLegacyHeldSlotsDeduplicateFullIdentityAndNeverMutate(t *testing.T) {
	o := legacyFixtureOptions(t)
	raw := `{"repo":"pair","worktree":"/fixture/pair","pid":123,"started":"2026-10-20T01:00:00","holder":555,"slots":[0,1]}`
	legacyHeldFile(t, o.LockBase, raw)
	legacyHeldFile(t, o.LockBase+".1", raw)
	active, detail := observeLegacyQueue(context.Background(), o, nil)
	if len(active) != 1 || active[0].Kind != "suite" || active[0].State != "unknown" || len(detail) == 0 {
		t.Fatalf("active %+v detail %v", active, detail)
	}
	before := active[0].ID
	active, _ = observeLegacyQueue(context.Background(), o, nil)
	if active[0].ID != before {
		t.Fatal("unstable legacy identity")
	}
	for _, path := range []string{o.LockBase, o.LockBase + ".1"} {
		got, err := os.ReadFile(path)
		if err != nil || string(got) != raw {
			t.Fatalf("observer changed resource: %q %v", got, err)
		}
	}
}
func TestQueueLegacyKnownBindingsExcludeNewWaitersAndActiveOperations(t *testing.T) {
	o := legacyFixtureOptions(t)
	op := legacyKnown(o.FleetRoot)
	op.Phase = "running"
	raw, _ := json.Marshal(map[string]any{"queue_operation_id": op.ID, "queue_pid": op.PID, "queue_pid_started": op.PIDStarted, "queue_root": op.Root, "pid": 222, "worktree": op.Root})
	legacyHeldFile(t, o.LockBase, string(raw))
	o.Processes = func(context.Context) ([]queueLegacyProcess, error) {
		return []queueLegacyProcess{{PID: 123, Command: "sdlc merge --yes", Root: op.Root}}, nil
	}
	active, detail := observeLegacyQueue(context.Background(), o, []queuegate.Operation{op})
	if len(active) != 0 || len(detail) != 0 {
		t.Fatalf("new waiter counted active: %+v %v", active, detail)
	}
	// An identical PID with another kernel incarnation does not inherit a lease.
	o.ProcessStart = func(int) (string, error) { return "reused", nil }
	active, detail = observeLegacyQueue(context.Background(), o, []queuegate.Operation{op})
	if len(active) == 0 || len(detail) == 0 {
		t.Fatal("PID reuse hidden")
	}
}
func TestQueueLegacyUnmatchedAndUnreadableHeldLandingRemainUnknown(t *testing.T) {
	o := legacyFixtureOptions(t)
	common := filepath.Join(o.FleetRoot, "repo", ".git")
	o.CommonDirs = func(context.Context, string) ([]string, error) { return []string{common, common}, nil }
	legacyHeldFile(t, filepath.Join(common, "sdlc", "landing", "head.lock"), `{"seq":7,"issue":481,"pid":123,"branch":"000481-work"}`)
	active, detail := observeLegacyQueue(context.Background(), o, nil)
	if len(active) != 1 || active[0].Kind != "landing" || active[0].Workspace != "" || len(detail) == 0 {
		t.Fatalf("head %+v %v", active, detail)
	}
	// A held but malformed record is never treated as a free head.
	bad := filepath.Join(o.FleetRoot, "bad", ".git")
	legacyHeldFile(t, filepath.Join(bad, "sdlc", "landing", "head.lock"), `{"pid":123,"pid":124}`)
	o.CommonDirs = func(context.Context, string) ([]string, error) { return []string{bad}, nil }
	active, detail = observeLegacyQueue(context.Background(), o, nil)
	if len(active) != 1 || len(detail) == 0 {
		t.Fatalf("malformed held head %+v %v", active, detail)
	}
}
func TestQueueLegacyBoundedFailuresCannotProveAbsence(t *testing.T) {
	for _, which := range []string{"topology", "process", "incarnation", "cancelled"} {
		t.Run(which, func(t *testing.T) {
			o := legacyFixtureOptions(t)
			ctx := context.Background()
			switch which {
			case "topology":
				o.CommonDirs = func(context.Context, string) ([]string, error) { return nil, errors.New("no topology") }
			case "process":
				o.Processes = func(context.Context) ([]queueLegacyProcess, error) {
					return nil, errors.New("truncated process inventory")
				}
			case "incarnation":
				o.Processes = func(context.Context) ([]queueLegacyProcess, error) {
					return []queueLegacyProcess{{PID: 123, Command: "sdlc merge", Root: o.FleetRoot}}, nil
				}
				o.ProcessStart = func(int) (string, error) { return "", errors.New("no kernel identity") }
			case "cancelled":
				var cancel context.CancelFunc
				ctx, cancel = context.WithCancel(ctx)
				cancel()
			}
			_, detail := observeLegacyQueue(ctx, o, nil)
			if len(detail) == 0 {
				t.Fatal("incomplete observation reported complete")
			}
		})
	}
}
func TestQueueLegacyPrivateInventoryScopeAndReadOnlyAbsence(t *testing.T) {
	o := legacyFixtureOptions(t)
	o.Processes = func(context.Context) ([]queueLegacyProcess, error) {
		return []queueLegacyProcess{{PID: 123, Command: "sdlc merge", Root: filepath.Join(o.FleetRoot, "pair")}, {PID: 124, Command: "sdlc merge", Root: "/outside"}, {PID: 125, Command: "go test ./...", Root: o.FleetRoot}}, nil
	}
	active, detail := observeLegacyQueue(context.Background(), o, nil)
	if len(active) != 1 || active[0].Kind != "merge" || len(detail) == 0 {
		t.Fatalf("scoped %+v %v", active, detail)
	}
	if _, err := os.Stat(o.LockBase); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("absent raw lock created: %v", err)
	}
	if !strings.Contains(strings.Join(detail, " "), "legacy") {
		t.Fatal(detail)
	}
}

func TestQueueLegacyProcessParserRejectsPartialAndBounds(t *testing.T) {
	good := []byte("123 501 /tools/sdlc merge --yes\n124 999 /tools/sdlc merge\n125 501 zsh -c sdlc merge\n126 501 /tools/sdlc-418 --quiet merge --yes\n")
	rows, err := queueLegacyParseProcesses(good, 501)
	if err != nil || len(rows) != 2 || rows[0].PID != 123 || rows[1].PID != 126 {
		t.Fatalf("rows %+v %v", rows, err)
	}
	for _, raw := range [][]byte{[]byte("123 501 sdlc merge"), []byte("not-a-pid 501 sdlc merge\n"), []byte("123 not-a-user sdlc merge\n"), []byte("123 501\n"), []byte(strings.Repeat("x", queuegate.MaxJSON+1))} {
		if _, err := queueLegacyParseProcesses(raw, 501); err == nil {
			t.Fatalf("accepted partial/invalid inventory len=%d", len(raw))
		}
	}
}

func TestQueueLegacyHeadBindingRequiresCompleteIncarnation(t *testing.T) {
	o := legacyFixtureOptions(t)
	op := legacyKnown(o.FleetRoot)
	op.Phase = "running"
	common := filepath.Join(o.FleetRoot, "repo", ".git")
	o.CommonDirs = func(context.Context, string) ([]string, error) { return []string{common}, nil }
	raw, _ := json.Marshal(map[string]any{"seq": 7, "queue_operation_id": op.ID, "queue_pid": op.PID, "queue_pid_started": op.PIDStarted, "queue_root": op.Root, "pid": op.PID})
	legacyHeldFile(t, filepath.Join(common, "sdlc", "landing", "head.lock"), string(raw))
	active, detail := observeLegacyQueue(context.Background(), o, []queuegate.Operation{op})
	if len(active) != 0 || len(detail) != 0 {
		t.Fatalf("bound head duplicated %+v %v", active, detail)
	}
	op.Root += "/different"
	active, detail = observeLegacyQueue(context.Background(), o, []queuegate.Operation{op})
	if len(active) != 1 || len(detail) == 0 {
		t.Fatal("wrong-root binding accepted")
	}
}

func TestQueueLegacyWaitingLeaseCannotHideActuallyHeldResource(t *testing.T) {
	o := legacyFixtureOptions(t)
	op := legacyKnown(o.FleetRoot)
	raw, _ := json.Marshal(map[string]any{"queue_operation_id": op.ID, "queue_pid": op.PID, "queue_pid_started": op.PIDStarted, "queue_root": op.Root})
	legacyHeldFile(t, o.LockBase, string(raw))
	active, detail := observeLegacyQueue(context.Background(), o, []queuegate.Operation{op})
	if len(active) != 1 || len(detail) == 0 {
		t.Fatal("waiting lease concealed held raw suite")
	}
}

func TestQueueLegacyDifferentIncarnationsWithSamePIDStayDistinct(t *testing.T) {
	o := legacyFixtureOptions(t)
	for i, started := range []string{"first", "second"} {
		path := o.LockBase
		if i > 0 {
			path += ".1"
		}
		raw, _ := json.Marshal(map[string]any{"pid": 123, "worktree": "/fixture/pair", "started": started})
		legacyHeldFile(t, path, string(raw))
	}
	active, _ := observeLegacyQueue(context.Background(), o, nil)
	if len(active) != 2 || active[0].ID == active[1].ID {
		t.Fatalf("PID-only dedup %+v", active)
	}
}

func TestQueueLegacyObserverHonorsCallerDeadline(t *testing.T) {
	o := legacyFixtureOptions(t)
	o.Processes = func(ctx context.Context) ([]queueLegacyProcess, error) { <-ctx.Done(); return nil, ctx.Err() }
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	_, detail := observeLegacyQueue(ctx, o, nil)
	if ctx.Err() == nil || len(detail) == 0 {
		t.Fatal("caller deadline did not produce uncertainty")
	}
}

func TestQueueLegacyBindingsIndexOnceAndLookupLinearly(t *testing.T) {
	for _, n := range []int{8, 16} {
		t.Run(fmt.Sprint(n), func(t *testing.T) {
			o := legacyFixtureOptions(t)
			known := make([]queuegate.Operation, n)
			processes := make([]queueLegacyProcess, n)
			cost := map[string]int{}
			o.Cost = func(label string) { cost[label]++ }
			for i := range known {
				known[i] = legacyKnown(o.FleetRoot)
				known[i].ID = fmt.Sprintf("known-%d", i)
				known[i].PID = 100 + i
				processes[i] = queueLegacyProcess{PID: known[i].PID, Command: "sdlc merge", Root: o.FleetRoot}
			}
			o.Processes = func(context.Context) ([]queueLegacyProcess, error) { return processes, nil }
			active, detail := observeLegacyQueue(context.Background(), o, known)
			if len(active) != 0 || len(detail) != 0 {
				t.Fatalf("known processes %+v %v", active, detail)
			}
			if cost["index"] != n || cost["lookup"] != n {
				t.Fatalf("wanted one indexing visit and lookup per operation, got %v", cost)
			}
		})
	}
}

func TestQueueLegacyCommonDirectoriesUseOnlyLocalGitTopology(t *testing.T) {
	root := t.TempDir()
	repo := filepath.Join(root, "repo")
	git(t, "", "init", "-q", repo)
	if err := os.Mkdir(filepath.Join(root, "unrelated"), 0700); err != nil {
		t.Fatal(err)
	}
	dirs, err := queueLegacyCommonDirs(context.Background(), root)
	if err != nil || len(dirs) != 1 || dirs[0] != filepath.Join(repo, ".git") {
		t.Fatalf("topology %v %v", dirs, err)
	}
	broken := filepath.Join(root, "broken")
	if err := os.Mkdir(broken, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(broken, ".git"), []byte("not a git marker"), 0600); err != nil {
		t.Fatal(err)
	}
	dirs, err = queueLegacyCommonDirs(context.Background(), root)
	if err == nil || len(dirs) != 1 {
		t.Fatalf("partial topology lost uncertainty or known directory: %v %v", dirs, err)
	}
}

func TestQueueLegacyInvalidUTF8CannotBindOrProveProcessAbsence(t *testing.T) {
	bad := append([]byte(`{"queue_operation_id":"known","future":"`), 0xff)
	bad = append(bad, []byte(`"}`)...)
	if _, err := queueLegacyFields(bad); err == nil {
		t.Fatal("invalid UTF8 held record accepted")
	}
	process := append([]byte("123 501 sdlc merge "), 0xff)
	process = append(process, '\n')
	if _, err := queueLegacyParseProcesses(process, 501); err == nil {
		t.Fatal("invalid UTF8 process output accepted")
	}
}
