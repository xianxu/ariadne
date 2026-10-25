package sysvshm

import (
	"reflect"
	"testing"
)

func TestDeltaOnlyNewUnattachedOwnerIdentities(t *testing.T) {
	old := Segment{ID: 1, Key: "0x1", UID: 501, Creator: 20, Size: 56, Changed: "01:00:00"}
	attached := old
	attached.ID, attached.Attachments = 2, 1
	replacement := old
	replacement.Creator = 21
	foreign := old
	foreign.ID, foreign.UID = 3, 502
	newSegment := old
	newSegment.ID = 4
	for _, tc := range []struct {
		name          string
		before, after []Segment
		want          []int
	}{
		{"preexisting", []Segment{old}, []Segment{old}, nil},
		{"attached", nil, []Segment{attached}, nil},
		{"removed", []Segment{old}, nil, nil},
		{"foreign", nil, []Segment{foreign}, nil},
		{"new", []Segment{old}, []Segment{old, newSegment}, []int{4}},
		{"same ID changed creator is ambiguous", []Segment{old}, []Segment{replacement}, nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := Delta(Snapshot{Segments: tc.before}, Snapshot{Segments: tc.after}, 501)
			var ids []int
			for _, c := range r.Candidates {
				ids = append(ids, c.Segment.ID)
			}
			if !reflect.DeepEqual(ids, tc.want) {
				t.Fatalf("IDs %v, want %v", ids, tc.want)
			}
		})
	}
}

func TestDeltaUnavailableCannotClaimClean(t *testing.T) {
	r := Delta(Snapshot{Unavailable: "truncated"}, Snapshot{}, 501)
	if r.Unavailable == "" {
		t.Fatal("invalid before snapshot claimed clean observation")
	}
}

func TestDeltaMetadataChangesAreNotNewIDs(t *testing.T) {
	before := Segment{ID: 7, Key: "old", UID: 501, Creator: 42, Size: 56, Changed: "old"}
	for _, mutate := range []func(*Segment){
		func(s *Segment) { s.Changed = "later" },
		func(s *Segment) { s.Size = 4096 },
		func(s *Segment) { s.UID = 502 },
		func(s *Segment) { s.Attachments = 4 },
		func(s *Segment) { s.Creator = 99; s.Key = "changed"; s.Changed = "later" },
	} {
		after := before
		mutate(&after)
		if r := Delta(Snapshot{Segments: []Segment{before}}, Snapshot{Segments: []Segment{after}}, 501); len(r.Candidates) != 0 {
			t.Fatalf("same ID metadata became a new segment: %+v", r)
		}
	}
}
