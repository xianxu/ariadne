package affected

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"
)

// The "nothing recorded" contract (#616): only a narrowed, group-free selection is evidence, and only for a failure
// with a Go package that nothing else explains.
func TestMissed(t *testing.T) {
	const m = "example.com/m/"
	narrow := Selection{GoPkgs: []string{m + "mid", m + "top"}}
	known := func(f string) bool { return f == m+"leaf.TestKnown" }
	failures := []string{m + "leaf.TestA", m + "leaf.TestA/sub", m + "mid.TestB", "TestBare", "not a name", m + "leaf.TestKnown", m + "leaf.TestA"}
	cases := []struct {
		name string
		sel  Selection
		want []Miss
	}{
		{"a narrowed selection misses the unselected package", narrow, []Miss{{m + "leaf.TestA", m + "leaf"}, {m + "leaf.TestA/sub", m + "leaf"}}},
		{"a full selection ran everything", Selection{Full: true, Reason: "go.mod"}, nil},
		{"a group's reach is unknown", Selection{GoPkgs: narrow.GoPkgs, Groups: []string{"lua"}}, nil},
		{"a selection that ran the package misses nothing", Selection{GoPkgs: []string{m + "leaf", m + "mid"}}, nil},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := Missed(c.sel, failures, known)
			if !reflect.DeepEqual(got, c.want) {
				t.Fatalf("got %+v, want %+v", got, c.want)
			}
			selected := map[string]bool{}
			for _, p := range c.sel.GoPkgs {
				selected[p] = true
			}
			for _, x := range got {
				if selected[x.Pkg] {
					t.Fatalf("a selected package is a miss: %+v", x)
				}
			}
		})
	}
}

func ev(at int, paths ...string) MissEvidence {
	return MissEvidence{Paths: paths, Head: "h", Main: "m", Source: "merge", At: time.Unix(int64(at), 0).UTC()}
}

// A sequence of records, dismissals and prunes keeps the ledger's invariants: unique IDs, at most the cap, and each
// entry's hits equal to how often its key was recorded.
func TestMissedLedgerSequence(t *testing.T) {
	var l MissedLedger
	recorded := map[string]int{}
	miss := func(pkg string) []Miss { return []Miss{{Test: pkg + ".TestX", Pkg: pkg}} }
	check := func(step string) {
		t.Helper()
		ids := map[string]bool{}
		for _, e := range l.Entries {
			if ids[e.ID] || e.Hits != recorded[e.ID] || e.ID != MissID(e.Package, e.Paths) && e.PathsTotal == len(e.Paths) {
				t.Fatalf("%s: bad entry %+v (recorded %d)", step, e, recorded[e.ID])
			}
			ids[e.ID] = true
		}
		if len(l.Entries) > missedLedgerCap {
			t.Fatalf("%s: %d entries over the cap", step, len(l.Entries))
		}
	}
	rec := func(pkg string, at int, paths ...string) {
		l.Record(miss(pkg), ev(at, paths...))
		recorded[MissID(pkg, paths)]++
	}
	rec("a", 1, "x", "y")
	rec("a", 2, "y", "x") // the same set in another order is the same key
	rec("a", 3, "x")
	rec("b", 4, "x", "y")
	check("records")
	if len(l.Entries) != 3 || l.Entries[0].Hits != 2 || !l.Entries[0].FirstSeen.Equal(time.Unix(1, 0)) || !l.Entries[0].LastSeen.Equal(time.Unix(2, 0)) {
		t.Fatalf("dedupe: %+v", l.Entries)
	}
	gone, err := l.Dismiss(MissID("a", []string{"x"}))
	if err != nil || gone.Package != "a" {
		t.Fatalf("dismiss: %+v %v", gone, err)
	}
	delete(recorded, gone.ID)
	if _, err := l.Dismiss(gone.ID); err == nil {
		t.Fatal("a second dismiss found it")
	}
	pruned := l.Prune(func(e MissedEntry) bool { return e.Package == "b" })
	if len(pruned) != 1 || len(l.Entries) != 1 {
		t.Fatalf("prune: %+v, left %+v", pruned, l.Entries)
	}
	for _, e := range pruned {
		delete(recorded, e.ID)
	}
	check("dismiss and prune")
	// Past the cap, the oldest sightings go.
	for i := range missedLedgerCap + 5 {
		rec(fmt.Sprintf("p%03d", i), 10+i, "z")
	}
	_, dropped := l.Record(miss("last"), ev(10_000, "z"))
	recorded[MissID("last", []string{"z"})]++
	if dropped != 1 || len(l.Entries) != missedLedgerCap {
		t.Fatalf("cap: dropped %d, %d entries", dropped, len(l.Entries))
	}
	for _, e := range l.Entries {
		if e.Package == "a" || e.Package == "p000" {
			t.Fatalf("an old entry survived the cap: %+v", e)
		}
	}
	for id := range recorded {
		if !strings.Contains(fmt.Sprint(l.Entries), id) {
			delete(recorded, id)
		}
	}
	check("cap")
}

// Paths are stored capped, but the key is the full set: two changes that share their first paths stay distinct.
func TestMissedLedgerPathCap(t *testing.T) {
	var many, more []string
	for i := range 30 {
		many = append(many, fmt.Sprintf("f%02d", i))
	}
	more = append(append(more, many...), "zz")
	var l MissedLedger
	l.Record([]Miss{{"p.TestX", "p"}}, ev(1, many...))
	l.Record([]Miss{{"p.TestX", "p"}}, ev(2, more...))
	if len(l.Entries) != 2 || len(l.Entries[0].Paths) != missedPathCap || l.Entries[0].PathsTotal != 30 || l.Entries[1].PathsTotal != 31 {
		t.Fatalf("%+v", l.Entries)
	}
}

// An older-shaped writer rewriting a newer record keeps its unknown members at both levels; known members stay strict.
func TestMissedLedgerCodec(t *testing.T) {
	newer := `{"version":1,"curator":{"team":"x"},"entries":[{"id":"0123456789ab","package":"p","test":"p.TestX",` +
		`"paths":["a"],"paths_total":1,"head":"h","main":"m","source":"merge","first_seen":"2026-10-24T00:00:00Z",` +
		`"last_seen":"2026-10-24T00:00:00Z","hits":1,"triage":{"owner":"y","notes":[1,2]}}]}`
	l, err := DecodeMissedLedger([]byte(newer))
	if err != nil {
		t.Fatal(err)
	}
	l.Record([]Miss{{"q.TestY", "q"}}, ev(5, "b"))
	out, err := json.Marshal(l)
	if err != nil {
		t.Fatal(err)
	}
	back, err := DecodeMissedLedger(out)
	if err != nil {
		t.Fatalf("%v\n%s", err, out)
	}
	if string(back.Extra["curator"]) != `{"team":"x"}` || string(back.Entries[0].Extra["triage"]) != `{"owner":"y","notes":[1,2]}` || len(back.Entries) != 2 {
		t.Fatalf("unknown members lost:\n%s", out)
	}
	for name, doc := range map[string]string{
		"version 2":       strings.Replace(newer, `"version":1`, `"version":2`, 1),
		"duplicate key":   strings.Replace(newer, `"hits":1`, `"hits":1,"hits":2`, 1),
		"wrong type":      strings.Replace(newer, `"hits":1`, `"hits":"1"`, 1),
		"null":            strings.Replace(newer, `"paths":["a"]`, `"paths":null`, 1),
		"bad source":      strings.Replace(newer, `"source":"merge"`, `"source":"ci"`, 1),
		"bad id":          strings.Replace(newer, `0123456789ab`, `XYZ`, 1),
		"no hits":         strings.Replace(newer, `"hits":1`, `"hits":0`, 1),
		"duplicate entry": newer[:len(newer)-2] + "," + newer[strings.Index(newer, `{"id"`):len(newer)-2] + "]}",
		"trailing data":   newer + "x",
	} {
		if _, err := DecodeMissedLedger([]byte(doc)); err == nil {
			t.Errorf("%s: accepted", name)
		}
	}
}

// FuzzDecodeMissedLedger: a decoded ledger re-encodes to one that decodes identically.
func FuzzDecodeMissedLedger(f *testing.F) {
	f.Add([]byte(`{"version":1,"entries":[]}`))
	f.Add([]byte(`{"version":1,"x":1,"entries":[{"id":"0123456789ab","package":"p","test":"p.T","paths":[],"paths_total":0,"head":"","main":"","source":"train","first_seen":"2026-10-24T00:00:00Z","last_seen":"2026-10-24T00:00:00Z","hits":1,"y":[]}]}`))
	f.Fuzz(func(t *testing.T, data []byte) {
		l, err := DecodeMissedLedger(data)
		if err != nil {
			return
		}
		out, err := json.Marshal(l)
		if err != nil {
			t.Fatal(err)
		}
		again, err := DecodeMissedLedger(out)
		if err != nil {
			t.Fatalf("re-encoded ledger refused: %v\n%s", err, out)
		}
		out2, _ := json.Marshal(again)
		if string(out) != string(out2) {
			t.Fatalf("not stable:\n%s\n%s", out, out2)
		}
	})
}
