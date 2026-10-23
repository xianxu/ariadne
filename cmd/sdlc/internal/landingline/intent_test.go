package landingline

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/flock"
)

func testIntent(branch string) Intent {
	return Intent{Version: IntentVersion, Branch: branch, Issue: 565, Head: strings.Repeat("a", 40), Workspace: "ariadne:5",
		Root: "/w/slot5", PID: os.Getpid(), StartedAt: time.Date(2026, 10, 23, 9, 0, 0, 0, time.UTC)}
}

func TestIntentCodecPreservesUnknownMembersAndRefusesBadRecords(t *testing.T) {
	data, err := json.Marshal(testIntent("000565-x"))
	if err != nil {
		t.Fatal(err)
	}
	newer := strings.Replace(string(data), `"version":1`, `"version":1,"later":{"nested":[1,2]}`, 1)
	in, err := DecodeIntent([]byte(newer))
	if err != nil {
		t.Fatal(err)
	}
	again, _ := json.Marshal(in)
	if !strings.Contains(string(again), `"later":{"nested":[1,2]}`) {
		t.Fatalf("a later member was dropped on rewrite: %s", again)
	}
	for name, bad := range map[string]string{
		"duplicate member": strings.Replace(string(data), `"version":1`, `"version":1,"version":1`, 1),
		"option branch":    strings.Replace(string(data), `"000565-x"`, `"--force"`, 1),
		"shell branch":     strings.Replace(string(data), `"000565-x"`, `"x;rm"`, 1),
		"short head":       strings.Replace(string(data), strings.Repeat("a", 40), "abc", 1),
		"relative root":    strings.Replace(string(data), `"/w/slot5"`, `"w/slot5"`, 1),
		"mistyped pid":     strings.Replace(string(data), `"pid":`, `"pid":"1","x":`, 1),
		"other version":    strings.Replace(string(data), `"version":1`, `"version":2`, 1),
		"null branch":      strings.Replace(string(data), `"000565-x"`, `null`, 1),
	} {
		if _, err := DecodeIntent([]byte(bad)); err == nil {
			t.Errorf("%s accepted", name)
		}
	}
}

func TestClassifyIntent(t *testing.T) {
	for _, c := range []struct {
		locked, exists bool
		want           IntentState
	}{{true, true, IntentLive}, {true, false, IntentLive}, {false, true, IntentOrphaned}, {false, false, IntentStale}} {
		if got := ClassifyIntent(c.locked, c.exists); got != c.want {
			t.Errorf("locked=%v exists=%v: %s, want %s", c.locked, c.exists, got, c.want)
		}
	}
}

func listOne(t *testing.T, s Intents) ObservedIntent {
	t.Helper()
	got, err := s.List()
	if err != nil || len(got) != 1 {
		t.Fatalf("want one intent, got %+v %v", got, err)
	}
	return got[0]
}

func TestIntentHoldReleaseByOutcome(t *testing.T) {
	prev := IntentHoldPatience
	IntentHoldPatience = 50 * time.Millisecond
	t.Cleanup(func() { IntentHoldPatience = prev })
	s := OpenIntents(t.TempDir())
	h, err := s.Hold(testIntent("b1"))
	if err != nil {
		t.Fatal(err)
	}
	if o := listOne(t, s); !o.Locked || o.Intent.Branch != "b1" {
		t.Fatalf("a held intent reads %+v", o)
	}
	var held *IntentHeldError
	if _, err := s.Hold(testIntent("b1")); !errors.As(err, &held) || held.Intent.PID != os.Getpid() || !strings.Contains(err.Error(), "pid") {
		t.Fatalf("a second merge of a live intent: %v", err)
	}
	if err := h.Release(true); err != nil {
		t.Fatal(err)
	}
	if o := listOne(t, s); o.Locked || o.Err != nil {
		t.Fatalf("an interrupted release should leave an unlocked orphan: %+v", o)
	}
	h, err = s.Hold(testIntent("b1")) // re-entry takes the orphan over
	if err != nil {
		t.Fatal(err)
	}
	if err := h.Release(false); err != nil {
		t.Fatal(err)
	}
	if got, _ := s.List(); len(got) != 0 {
		t.Fatalf("a terminal release left %+v", got)
	}
}

func TestIntentReentryKeepsLaterMembers(t *testing.T) {
	s := OpenIntents(t.TempDir())
	h, err := s.Hold(testIntent("b1"))
	if err != nil {
		t.Fatal(err)
	}
	h.Release(true)
	path := listOne(t, s).Path
	data, _ := os.ReadFile(path)
	os.WriteFile(path, []byte(strings.Replace(string(data), `"version":1`, `"version":1,"later":true`, 1)), 0o644)
	next := testIntent("b1")
	next.Head = strings.Repeat("b", 40)
	if h, err = s.Hold(next); err != nil {
		t.Fatal(err)
	}
	defer h.Release(false)
	data, _ = os.ReadFile(path)
	if !strings.Contains(string(data), `"later":true`) || !strings.Contains(string(data), strings.Repeat("b", 40)) {
		t.Fatalf("re-entry record: %s", data)
	}
}

func TestIntentSweep(t *testing.T) {
	s := OpenIntents(t.TempDir())
	now := time.Date(2026, 10, 23, 12, 0, 0, 0, time.UTC)
	for _, b := range []string{"gone", "kept", "old", "live"} {
		in := testIntent(b)
		if b == "old" {
			in.StartedAt = now.Add(-IntentMaxAge - time.Hour)
		}
		h, err := s.Hold(in)
		if err != nil {
			t.Fatal(err)
		}
		if b == "live" {
			defer h.Release(false)
			continue
		}
		h.Release(true)
	}
	os.WriteFile(s.dir+"/garbage", []byte("{"), 0o644)
	exists := func(b string) (bool, error) { return b != "gone" && b != "live", nil }
	removed, err := s.Sweep(exists, now, IntentMaxAge)
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, r := range removed {
		names = append(names, r.Branch)
	}
	if strings.Join(names, ",") != "gone,old" && strings.Join(names, ",") != "old,gone" {
		t.Fatalf("swept %q", names)
	}
	left, _ := s.List()
	var kept []string
	for _, o := range left {
		if o.Err != nil {
			kept = append(kept, "unreadable")
		} else {
			kept = append(kept, o.Intent.Branch)
		}
	}
	if len(kept) != 3 || !strings.Contains(strings.Join(kept, ","), "kept") || !strings.Contains(strings.Join(kept, ","), "live") || !strings.Contains(strings.Join(kept, ","), "unreadable") {
		t.Fatalf("left %q", kept)
	}
}

// The helper process holds an intent until it is killed, as a merge runner
// does when a full shutdown kills it.
func TestHelperHoldIntent(t *testing.T) {
	dir := os.Getenv("LANDINGLINE_INTENT_DIR")
	if dir == "" {
		t.Skip("helper process only")
	}
	if _, err := OpenIntents(dir).Hold(testIntent("killed")); err != nil {
		os.Exit(2)
	}
	os.Stdout.WriteString("ready\n")
	time.Sleep(time.Hour)
}

func TestIntentOfAKilledRunnerIsOrphaned(t *testing.T) {
	dir := t.TempDir()
	cmd := exec.Command(os.Args[0], "-test.run=^TestHelperHoldIntent$")
	cmd.Env = append(os.Environ(), "LANDINGLINE_INTENT_DIR="+dir)
	out, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = cmd.Process.Kill(); _ = cmd.Wait() })
	if line, err := bufio.NewReader(out).ReadString('\n'); err != nil || line != "ready\n" {
		t.Fatalf("helper not ready: %q %v", line, err)
	}
	s := OpenIntents(dir)
	if o := listOne(t, s); !o.Locked {
		t.Fatal("a running merge's intent reads unlocked")
	}
	cmd.Process.Signal(syscall.SIGKILL)
	cmd.Wait()
	o := listOne(t, s)
	if o.Locked || o.Err != nil || o.Intent.Branch != "killed" || o.Intent.PID != cmd.Process.Pid {
		t.Fatalf("after the kill: %+v", o)
	}
	if ClassifyIntent(o.Locked, true) != IntentOrphaned {
		t.Fatal("not orphaned")
	}
}

// #565 BR: a reader's lock is not a running merge. Hold waits out a reader
// that holds the record while Hold starts; readers do not see each other as
// live; a sweep that decided before a re-entry does not delete it.
func TestIntentReadersAreNotOwners(t *testing.T) {
	s := OpenIntents(t.TempDir())
	h, err := s.Hold(testIntent("b1"))
	if err != nil {
		t.Fatal(err)
	}
	h.Release(true)
	path := listOne(t, s).Path

	reader, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := flock.Shared(reader, true); err != nil {
		t.Fatal(err)
	}
	if o := listOne(t, s); o.Locked {
		t.Fatal("a second reader saw the first as a running merge")
	}
	waited := make(chan struct{})
	var once sync.Once
	onHoldWait = func() { once.Do(func() { close(waited) }) }
	t.Cleanup(func() { onHoldWait = func() {} })
	done := make(chan error, 1)
	go func() {
		h, err := s.Hold(testIntent("b1"))
		if err == nil {
			h.Release(true)
		}
		done <- err
	}()
	<-waited // Hold found the reader's lock and is waiting it out
	reader.Close()
	if err := <-done; err != nil {
		t.Fatalf("re-entry refused because of a reader: %v", err)
	}

	reentered := false
	removed, err := s.Sweep(func(string) (bool, error) {
		// A re-entry lands between the sweep's read and its removal.
		next := testIntent("b1")
		next.StartedAt = next.StartedAt.Add(time.Hour)
		h, err := s.Hold(next)
		if err != nil {
			return false, err
		}
		reentered = true
		return false, h.Release(true)
	}, time.Date(2026, 10, 23, 12, 0, 0, 0, time.UTC), IntentMaxAge)
	if err != nil || !reentered || len(removed) != 0 {
		t.Fatalf("swept a re-entered intent: removed %+v, err %v", removed, err)
	}
	if o := listOne(t, s); o.Err != nil || !o.Intent.StartedAt.Equal(testIntent("b1").StartedAt.Add(time.Hour)) {
		t.Fatalf("after the sweep: %+v", o)
	}
}

func TestIntentHoldNeverOverwritesAnUnreadableRecord(t *testing.T) {
	prev := IntentHoldPatience
	IntentHoldPatience = 50 * time.Millisecond
	t.Cleanup(func() { IntentHoldPatience = prev })
	for name, body := range map[string]string{
		"corrupt":       "{",
		"later version": `{"version":2,"branch":"b1"}`,
	} {
		s := OpenIntents(t.TempDir())
		os.MkdirAll(s.dir, 0o755)
		path := s.path("b1")
		os.WriteFile(path, []byte(body), 0o644)
		if _, err := s.Hold(testIntent("b1")); err == nil || !strings.Contains(err.Error(), fmt.Sprintf("unreadable merge intent %q", path)) {
			t.Errorf("%s: %v", name, err)
		}
		if got, _ := os.ReadFile(path); string(got) != body {
			t.Errorf("%s: overwritten to %s", name, got)
		}
	}
}

// A first record is linked into place already locked and complete, so no
// reader can find it empty; a writer that died before linking leaves a
// temporary file the sweep collects, while a live writer's stays.
func TestIntentCreationLeavesNoWindowOrResidue(t *testing.T) {
	s := OpenIntents(t.TempDir())
	h, err := s.Hold(testIntent("b1"))
	if err != nil {
		t.Fatal(err)
	}
	defer h.Release(false)
	if o := listOne(t, s); !o.Locked || o.Err != nil || o.Intent.Branch != "b1" {
		t.Fatalf("a new record reads %+v", o)
	}
	entries, _ := os.ReadDir(s.dir)
	if len(entries) != 1 {
		t.Fatalf("creation left residue: %v", entries)
	}
	dead := s.dir + "/.tmp-dead"
	os.WriteFile(dead, []byte("{"), 0o644)
	live, err := os.CreateTemp(s.dir, ".tmp-")
	if err != nil {
		t.Fatal(err)
	}
	defer live.Close()
	if err := flock.Exclusive(live, true); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Sweep(func(string) (bool, error) { return true, nil }, time.Now(), IntentMaxAge); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(dead); !os.IsNotExist(err) {
		t.Fatal("a dead writer's temporary record survived the sweep")
	}
	if _, err := os.Stat(live.Name()); err != nil {
		t.Fatal("the sweep removed a live writer's temporary record")
	}
}
