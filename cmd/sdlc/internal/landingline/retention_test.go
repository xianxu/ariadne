package landingline

import (
	"maps"
	"math/rand"
	"testing"
	"time"
)

func kept[T any](xs []T, keep []bool) []T {
	var out []T
	for i, x := range xs {
		if keep[i] {
			out = append(out, x)
		}
	}
	return out
}

func sameGrants(a, b map[int]Grant) bool {
	return maps.EqualFunc(a, b, func(x, y Grant) bool { return x.Granted.Equal(y.Granted) && x.Expires.Equal(y.Expires) })
}

// The invariant over random logs: grants and revokes for a few issues, dated
// from 40 days ago to 2 days ahead (forged or skewed), in any log order.
func TestKeepGrantsPreservesValidity(t *testing.T) {
	r := rand.New(rand.NewSource(393))
	now := t0
	for run := 0; run < 2000; run++ {
		var log []Grant
		for n := r.Intn(8); n > 0; n-- {
			granted := now.Add(time.Duration(r.Intn(42*24)-40*24) * time.Hour)
			g := grant(1+r.Intn(3), granted, 1+r.Intn(30))
			g.Revoked = r.Intn(4) == 0
			log = append(log, g)
		}
		keep := KeepGrants(log, now)
		after := kept(log, keep)
		if !sameGrants(ValidGrants(log, now), ValidGrants(after, now)) {
			t.Fatalf("run %d: sweep changed validity\nlog %v\nkept %v", run, log, after)
		}
		for _, g := range after {
			if !fresh(g.Granted, now) {
				t.Fatalf("run %d: kept an entry outside retention: %v", run, g)
			}
		}
	}
}

func TestKeepGrantsCases(t *testing.T) {
	now := t0
	live := grant(1, now.Add(-time.Hour), 24)
	revoke := live
	revoke.Granted, revoke.Revoked = now.Add(-30*time.Minute), true
	old := grant(2, now.Add(-40*24*time.Hour), 4)
	cases := []struct {
		name string
		log  []Grant
		keep []bool
	}{
		{"a grant in force stays (its expiry is ahead, which is not 'future-dated')", []Grant{live}, []bool{true}},
		{"a revoke and the grant it revokes both stay", []Grant{live, revoke}, []bool{true, true}},
		{"an old latest entry takes its issue's history with it", []Grant{old}, []bool{false}},
		{"a stale latest entry drops a fresh earlier one, which mustn't become the latest", []Grant{grant(2, now.Add(-time.Hour), 4), old}, []bool{false, false}},
		{"old history behind a fresh latest goes", []Grant{old, grant(2, now.Add(-time.Hour), 4)}, []bool{false, true}},
		{"a future-dated entry goes", []Grant{grant(3, now.Add(time.Hour), 4)}, []bool{false}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := KeepGrants(tc.log, now)
			for i := range got {
				if got[i] != tc.keep[i] {
					t.Fatalf("keep %v, want %v", got, tc.keep)
				}
			}
		})
	}
}

func TestKeepEvents(t *testing.T) {
	now := t0
	got := KeepEvents([]Event{{At: now.Add(-time.Hour)}, {At: now.Add(-31 * 24 * time.Hour)}, {At: now.Add(time.Hour)}}, now)
	if !got[0] || got[1] || got[2] {
		t.Fatalf("keep %v", got)
	}
}
