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
		{"replacement", []Segment{old}, []Segment{replacement}, []int{1}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := Delta(Snapshot{Segments: tc.before}, Snapshot{Segments: tc.after}, 501, nil)
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
	r := Delta(Snapshot{Unavailable: "truncated"}, Snapshot{}, 501, nil)
	if r.Unavailable == "" {
		t.Fatal("invalid before snapshot claimed clean observation")
	}
}

func TestDeltaCreatorRequiresMatchingWitness(t *testing.T) {
	s := Segment{ID: 8, UID: 501, Creator: 42}
	for _, tc := range []struct {
		name      string
		witnesses map[int]Witness
		want      string
	}{
		{"unwitnessed", nil, ""},
		{"different creator", map[int]Witness{43: {Name: "postgres", Start: "123"}}, ""},
		{"uncertain incarnation", map[int]Witness{42: {Name: "postgres"}}, ""},
		{"witnessed", map[int]Witness{42: {Name: "postgres", Start: "123"}}, "postgres"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := Delta(Snapshot{}, Snapshot{Segments: []Segment{s}}, 501, tc.witnesses)
			if len(r.Candidates) != 1 || r.Candidates[0].Witness.Name != tc.want {
				t.Fatalf("%+v", r)
			}
		})
	}
}
