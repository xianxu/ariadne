package sysvshm

import "strings"

// Delta reports new unattached owner allocations; timing alone is not causality.
func Delta(before, after Snapshot, uid int, witnesses map[int]Witness) Report {
	r := Report{Unavailable: strings.Join(nonempty(before.Unavailable, after.Unavailable), "; ")}
	if r.Unavailable != "" {
		return r
	}
	old := make(map[Segment]bool, len(before.Segments))
	for _, s := range before.Segments {
		if s.UID == uid {
			r.Before++
			s.Attachments = 0
			old[s] = true
		}
	}
	for _, s := range after.Segments {
		if s.UID != uid {
			continue
		}
		r.After++
		identity := s
		identity.Attachments = 0
		if s.Attachments != 0 || old[identity] {
			continue
		}
		w := witnesses[s.Creator]
		if w.Start == "" {
			w = Witness{}
		}
		r.Candidates = append(r.Candidates, Candidate{Segment: s, Witness: w})
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
