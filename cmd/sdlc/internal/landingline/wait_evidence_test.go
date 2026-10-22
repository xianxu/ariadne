package landingline

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func waitID(n string) WaitIdentity {
	return WaitIdentity{Repo: "github.com/test/repo", Root: "/fleet/" + n, Slot: "repo:" + n, Head: strings.Repeat("a", 40), OperationID: "op-" + n, PID: 123, PIDStarted: "inc-" + n, Branch: "000547-test"}
}
func TestWaitEvidenceLifecycle(t *testing.T) {
	now := time.Date(2026, 10, 22, 12, 0, 0, 0, time.UTC)
	l := openLine(t, t.TempDir(), Options{Now: func() time.Time { return now }})
	conductor, err := l.Enqueue(Ticket{Issue: 1, Branch: "conductor", WaitIdentity: ptrWait(waitID("1"))})
	if err != nil {
		t.Fatal(err)
	}
	defer conductor.Release()
	if err = l.BeginWait(conductor); err != nil {
		t.Fatal(err)
	}
	if err = l.Wait(context.Background(), conductor); err != nil {
		t.Fatal(err)
	}
	if err = l.EndWait(conductor); err != nil {
		t.Fatal(err)
	}
	car, err := l.Enqueue(Ticket{Issue: 547, Branch: "car", WaitIdentity: ptrWait(waitID("2"))})
	if err != nil {
		t.Fatal(err)
	}
	defer car.Release()
	if err = l.BeginWait(car); err != nil {
		t.Fatal(err)
	}
	now = now.Add(time.Minute)
	if err = l.BindWait(conductor, Car{Seq: car.Ticket().Seq, Issue: 547, Head: waitID("2").Head, PR: 99}); err != nil {
		t.Fatal(err)
	}
	now = now.Add(time.Minute)
	if err = l.EndWait(car); err != nil {
		t.Fatal(err)
	}
	events, _, err := l.Events()
	if err != nil {
		t.Fatal(err)
	}
	var records []WaitRecord
	for _, e := range events {
		b, _ := json.Marshal(e)
		decoded, err := DecodeWaitEvent(b)
		if err != nil {
			t.Fatal(err)
		}
		records = append(records, WaitRecord{Event: decoded, Digest: fmt.Sprintf("%x", sha256.Sum256(b))})
	}
	intervals, notes := ReduceWaitEvidence(records)
	if len(notes) != 0 || len(intervals) != 1 || intervals[0].End.Sub(intervals[0].Start) != 2*time.Minute {
		t.Fatalf("intervals=%+v notes=%+v", intervals, notes)
	}
	car.Release()
	if err = l.EndWait(car); err == nil {
		t.Fatal("end after release succeeded")
	}
}
func ptrWait(id WaitIdentity) *WaitIdentity { return &id }
func TestWaitEvidenceReducerConservative(t *testing.T) {
	at := time.Date(2026, 10, 22, 12, 0, 0, 0, time.UTC)
	id := waitID("2")
	con := waitID("1")
	start := Event{Kind: EventWaitStart, At: at, Seq: 2, Issue: 547, Wait: &WaitFact{Version: 1, Identity: id}}
	bind := Event{Kind: EventWaitBind, At: at.Add(time.Minute), Seq: 2, Issue: 547, PR: 99, Train: 1, Wait: &WaitFact{Version: 1, Identity: id, Conductor: &con}}
	end := Event{Kind: EventWaitEnd, At: at.Add(3 * time.Minute), Seq: 2, Issue: 547, Wait: &WaitFact{Version: 1, Identity: id}}
	tests := []struct {
		name                 string
		events               []Event
		complete, incomplete int
	}{
		{"reordered duplicate", []Event{end, bind, start, start}, 1, 0}, {"crash", []Event{start, bind}, 0, 1}, {"unboarded", []Event{start, end}, 0, 0}, {"missing start", []Event{bind, end}, 0, 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var rs []WaitRecord
			for _, e := range tt.events {
				b, _ := json.Marshal(e)
				rs = append(rs, WaitRecord{Event: e, Digest: fmt.Sprintf("%x", sha256.Sum256(b))})
			}
			got, notes := ReduceWaitEvidence(rs)
			if len(got) != tt.complete || len(notes) != tt.incomplete {
				t.Fatalf("%+v %+v", got, notes)
			}
		})
	}
	bad := end
	bad.At = at.Add(-time.Minute)
	got, notes := ReduceWaitEvidence([]WaitRecord{waitRecord(start), waitRecord(bind), waitRecord(bad)})
	if len(got) != 0 || len(notes) != 1 {
		t.Fatal(got, notes)
	}
	raw, _ := json.Marshal(start)
	raw = []byte(strings.Replace(string(raw), `"version":1`, `"version":1,"version":1`, 1))
	if _, err := DecodeWaitEvent(raw); err == nil {
		t.Fatal("duplicate version accepted")
	}
}
func TestWaitEvidenceInitialWriteFailure(t *testing.T) {
	l := openLine(t, t.TempDir(), Options{})
	h, err := l.Enqueue(Ticket{Issue: 547, WaitIdentity: ptrWait(waitID("2"))})
	if err != nil {
		t.Fatal(err)
	}
	defer h.Release()
	if err = os.Mkdir(filepath.Join(l.dir, EventsFile), 0755); err != nil {
		t.Fatal(err)
	}
	if err = l.BeginWait(h); err == nil {
		t.Fatal("write failure hidden")
	}
}

func TestWaitEvidenceBindingRequiresLiveUnendedExactCar(t *testing.T) {
	l := openLine(t, t.TempDir(), Options{})
	conductor, err := l.Enqueue(Ticket{Issue: 1, WaitIdentity: ptrWait(waitID("1"))})
	if err != nil {
		t.Fatal(err)
	}
	defer conductor.Release()
	if err = l.Wait(context.Background(), conductor); err != nil {
		t.Fatal(err)
	}
	h, err := l.Enqueue(Ticket{Issue: 547, WaitIdentity: ptrWait(waitID("2"))})
	if err != nil {
		t.Fatal(err)
	}
	defer h.Release()
	if err = l.BeginWait(h); err != nil {
		t.Fatal(err)
	}
	car := Car{Seq: h.Ticket().Seq, Issue: 547, Head: strings.Repeat("b", 40), PR: 99}
	if err = l.BindWait(conductor, car); err == nil {
		t.Fatal("foreign head bound")
	}
	car.Head = waitID("2").Head
	if err = l.EndWait(h); err != nil {
		t.Fatal(err)
	}
	if err = l.BindWait(conductor, car); err == nil {
		t.Fatal("ended waiter bound")
	}
	h.Release()
	if err = l.BindWait(conductor, car); err == nil {
		t.Fatal("released waiter bound")
	}
}
func TestWaitEvidenceLateWriteFailurePreservesPartialFacts(t *testing.T) {
	l := openLine(t, t.TempDir(), Options{})
	h, err := l.Enqueue(Ticket{Issue: 547, WaitIdentity: ptrWait(waitID("2"))})
	if err != nil {
		t.Fatal(err)
	}
	defer h.Release()
	if err = l.BeginWait(h); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(l.dir, EventsFile)
	original, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err = os.Rename(path, path+".saved"); err != nil {
		t.Fatal(err)
	}
	if err = os.Mkdir(path, 0755); err != nil {
		t.Fatal(err)
	}
	if err = l.EndWait(h); err == nil {
		t.Fatal("late failure hidden")
	}
	if !h.waiting() {
		t.Fatal("evidence failure released owner")
	}
	var start Event
	if err = json.Unmarshal(original, &start); err != nil {
		t.Fatal(err)
	}
	intervals, notes := ReduceWaitEvidence([]WaitRecord{waitRecord(start)})
	if len(intervals) != 0 || len(notes) != 1 {
		t.Fatal(intervals, notes)
	}
}

func TestWaitEvidenceRejectsNoncanonicalIdentity(t *testing.T) {
	for _, mutate := range []func(*WaitIdentity){func(id *WaitIdentity) { id.Slot = "malformed" }, func(id *WaitIdentity) { id.Head = strings.Repeat("A", 40) }, func(id *WaitIdentity) { id.PIDStarted = strings.Repeat("x", 4097) }} {
		id := waitID("2")
		mutate(&id)
		e := Event{Kind: EventWaitStart, At: time.Now(), Seq: 2, Issue: 547, Wait: &WaitFact{Version: 1, Identity: id}}
		raw, _ := json.Marshal(e)
		if _, err := DecodeWaitEvent(raw); err == nil {
			t.Fatalf("invalid identity accepted: %+v", id)
		}
	}
}

func TestWaitEvidenceFactGrowthIgnoresPolls(t *testing.T) {
	for _, n := range []int{8, 16} {
		t.Run(fmt.Sprint(n), func(t *testing.T) {
			l := openLine(t, t.TempDir(), Options{})
			head, err := l.Enqueue(Ticket{Issue: 1, WaitIdentity: ptrWait(waitID("0"))})
			if err != nil {
				t.Fatal(err)
			}
			defer head.Release()
			if err = l.Wait(context.Background(), head); err != nil {
				t.Fatal(err)
			}
			var hs []*Handle
			for i := 0; i < n; i++ {
				id := waitID(fmt.Sprint(i + 1))
				h, err := l.Enqueue(Ticket{Issue: 547, WaitIdentity: &id})
				if err != nil {
					t.Fatal(err)
				}
				defer h.Release()
				hs = append(hs, h)
				if err = l.BeginWait(h); err != nil {
					t.Fatal(err)
				}
				if err = l.BindWait(head, Car{Seq: h.Ticket().Seq, Issue: 547, Head: id.Head, PR: i + 1}); err != nil {
					t.Fatal(err)
				}
			}
			before, _, err := l.Events()
			if err != nil {
				t.Fatal(err)
			}
			for i := 0; i < 2*n; i++ {
				if _, _, _, err = l.tryAdmit(hs[i%n]); err != nil {
					t.Fatal(err)
				}
			}
			after, _, err := l.Events()
			if err != nil {
				t.Fatal(err)
			}
			if len(after) != len(before) || len(after) != 2*n {
				t.Fatalf("polls produced facts: before=%d after=%d", len(before), len(after))
			}
			for _, h := range hs {
				if err = l.EndWait(h); err != nil {
					t.Fatal(err)
				}
			}
			events, _, err := l.Events()
			if err != nil {
				t.Fatal(err)
			}
			if len(events) != 3*n {
				t.Fatalf("facts=%d want=%d", len(events), 3*n)
			}
			var records []WaitRecord
			for _, e := range events {
				raw, _ := json.Marshal(e)
				records = append(records, WaitRecord{Event: e, Digest: fmt.Sprintf("%x", sha256.Sum256(raw))})
			}
			intervals, notes := ReduceWaitEvidence(records)
			if len(intervals) != n || len(notes) != 0 {
				t.Fatal(intervals, notes)
			}
		})
	}
}

func TestWaitEvidenceContradictoryFactStreams(t *testing.T) {
	at := time.Date(2026, 10, 22, 12, 0, 0, 0, time.UTC)
	id, con := waitID("2"), waitID("1")
	start := Event{Kind: EventWaitStart, At: at, Seq: 2, Issue: 547, Wait: &WaitFact{Version: 1, Identity: id}}
	bind := Event{Kind: EventWaitBind, At: at.Add(time.Minute), Seq: 2, Issue: 547, PR: 99, Train: 1, Wait: &WaitFact{Version: 1, Identity: id, Conductor: &con}}
	end := Event{Kind: EventWaitEnd, At: at.Add(3 * time.Minute), Seq: 2, Issue: 547, Wait: &WaitFact{Version: 1, Identity: id}}
	for _, name := range []string{"start", "end", "head", "root", "slot", "foreign conductor", "self conductor", "binding before", "binding after", "binding timestamp"} {
		t.Run(name, func(t *testing.T) {
			extra := bind
			wf := *bind.Wait
			extra.Wait = &wf
			switch name {
			case "start":
				extra = start
				extra.At = extra.At.Add(time.Second)
			case "end":
				extra = end
				extra.At = extra.At.Add(time.Second)
			case "head":
				extra.Wait.Identity.Head = strings.Repeat("b", 40)
			case "root":
				extra.Wait.Identity.Root = "/foreign"
			case "slot":
				extra.Wait.Identity.Slot = "repo:9"
			case "foreign conductor":
				foreign := con
				foreign.Repo = "github.com/foreign/repo"
				extra.Wait.Conductor = &foreign
			case "self conductor":
				extra.Wait.Conductor = &id
			case "binding before":
				extra.At = at.Add(-time.Second)
			case "binding after":
				extra.At = end.At.Add(time.Second)
			case "binding timestamp":
				extra.At = extra.At.Add(time.Second)
			}
			intervals, notes := ReduceWaitEvidence([]WaitRecord{waitRecord(start), waitRecord(bind), waitRecord(end), waitRecord(extra)})
			if len(intervals) != 0 || len(notes) != 1 {
				t.Fatal(intervals, notes)
			}
		})
	}
}

func TestWaitEvidenceRejectsMissingRawDigest(t *testing.T) {
	at := time.Now()
	id, con := waitID("2"), waitID("1")
	events := []Event{{Kind: EventWaitStart, At: at, Seq: 2, Issue: 547, Wait: &WaitFact{Version: 1, Identity: id}}, {Kind: EventWaitBind, At: at, Seq: 2, Issue: 547, PR: 99, Train: 1, Wait: &WaitFact{Version: 1, Identity: id, Conductor: &con}}, {Kind: EventWaitEnd, At: at, Seq: 2, Issue: 547, Wait: &WaitFact{Version: 1, Identity: id}}}
	var records []WaitRecord
	for _, event := range events {
		records = append(records, waitRecord(event))
	}
	records[1].Digest = ""
	intervals, notes := ReduceWaitEvidence(records)
	if len(intervals) != 0 || len(notes) != 1 {
		t.Fatal(intervals, notes)
	}
}
func waitRecord(event Event) WaitRecord {
	raw, _ := json.Marshal(event)
	return WaitRecord{Event: event, Digest: fmt.Sprintf("%x", sha256.Sum256(raw))}
}
