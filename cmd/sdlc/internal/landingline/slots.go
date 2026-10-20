package landingline

import "time"

// Slots is a repository's departure grid (#475): wall-clock times
// epoch + Offset + k·Interval, anchored at the Unix epoch (an exact UTC hour),
// so every process on the machine computes the same grid, and an interval
// that divides an hour gives hh:00 + offset. The grid never shifts: an early
// departure (a full train, a car waited too long) leaves later slots where
// they are, so the machine plan's spacing between repositories holds.
type Slots struct {
	Interval time.Duration
	Offset   time.Duration // in [0, Interval)
}

var epoch = time.Unix(0, 0)

// Prev is the latest slot at or before t.
func (g Slots) Prev(t time.Time) time.Time {
	if g.Interval <= 0 {
		return t
	}
	since := t.Sub(epoch) - g.Offset%g.Interval
	k := since / g.Interval
	if since < 0 && since%g.Interval != 0 {
		k-- // floor, for a time before the epoch's first slot
	}
	return epoch.Add(g.Offset%g.Interval + k*g.Interval).In(t.Location())
}

// Next is the first slot after t.
func (g Slots) Next(t time.Time) time.Time {
	if g.Interval <= 0 {
		return t
	}
	return g.Prev(t).Add(g.Interval)
}
