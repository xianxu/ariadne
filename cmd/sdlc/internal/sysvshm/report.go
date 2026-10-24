package sysvshm

import "strings"

// Delta reports newly observed unattached owner IDs; timing alone is not causality.
func Delta(before, after Snapshot, uid int) Report {
	r := Report{Unavailable: strings.Join(nonempty(before.Unavailable, after.Unavailable), "; ")}
	if r.Unavailable != "" {
		return r
	}
	old := make(map[int]bool, len(before.Segments))
	for _, s := range before.Segments {
		old[s.ID] = true
		if s.UID == uid {
			r.Before++
		}
	}
	for _, s := range after.Segments {
		if s.UID != uid {
			continue
		}
		r.After++
		if s.Attachments != 0 || old[s.ID] {
			continue
		}
		r.Candidates = append(r.Candidates, Candidate{Segment: s})
	}
	return r
}

func nonempty(values ...string) []string {
	var out []string
	for _, v := range values {
		if v != "" {
			out = append(out, v)
		}
	}
	return out
}
