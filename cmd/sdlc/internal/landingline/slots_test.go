package landingline

import (
	"math/rand"
	"testing"
	"time"
)

// Slots at the boundaries (a slot instant is its own Prev), then the grid's
// property over random times and grids: Prev ≤ t < Next, one interval apart,
// on the grid.
func TestSlots(t *testing.T) {
	at := func(hm string) time.Time {
		v, err := time.Parse("2006-01-02 15:04:05", "2026-10-20 "+hm)
		if err != nil {
			t.Fatal(err)
		}
		return v
	}
	for _, tc := range []struct {
		g          Slots
		now        string
		prev, next string
	}{
		{Slots{time.Hour, 0}, "10:00:00", "10:00:00", "11:00:00"},
		{Slots{time.Hour, 0}, "10:59:59", "10:00:00", "11:00:00"},
		{Slots{time.Hour, 15 * time.Minute}, "10:14:59", "09:15:00", "10:15:00"},
		{Slots{time.Hour, 15 * time.Minute}, "10:15:00", "10:15:00", "11:15:00"},
		{Slots{30 * time.Minute, 45 * time.Minute}, "10:20:00", "10:15:00", "10:45:00"}, // offset taken mod interval
		{Slots{30 * time.Minute, 0}, "10:30:00", "10:30:00", "11:00:00"},
	} {
		if p, n := tc.g.Prev(at(tc.now)), tc.g.Next(at(tc.now)); !p.Equal(at(tc.prev)) || !n.Equal(at(tc.next)) {
			t.Errorf("%+v at %s: prev %s next %s, want %s %s", tc.g, tc.now, p.Format("15:04:05"), n.Format("15:04:05"), tc.prev, tc.next)
		}
	}

	rng := rand.New(rand.NewSource(475))
	for i := 0; i < 2000; i++ {
		interval := time.Duration(1+rng.Intn(240)) * time.Minute
		g := Slots{interval, time.Duration(rng.Int63n(int64(interval)))}
		now := time.Unix(rng.Int63n(4e9)-1e9, rng.Int63n(1e9))
		p, n := g.Prev(now), g.Next(now)
		if p.After(now) || !n.After(now) || n.Sub(p) != interval || (p.Sub(epoch)-g.Offset)%interval != 0 {
			t.Fatalf("%+v at %s: prev %s next %s", g, now, p, n)
		}
	}
}
