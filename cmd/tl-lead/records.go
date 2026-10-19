package main

import (
	"fmt"
	"slices"
	"strings"
)

// editRecord is one of couch's `apply` records: replace the occurrence-th
// plain match of Old with New, giving the operator Explain.
type editRecord struct {
	Old        string `json:"old"`
	Occurrence int    `json:"occurrence"`
	New        string `json:"new"`
	Explain    string `json:"explain"`
}

// applyRecords is couch's apply over text (pair sessionsync's batch rule):
// every record resolves against text as it stands, each occurrence counted
// from the end of the previous match, none overlapping, all or none. The
// migration checks its own records with it before sending them, and the
// fake applies with it.
func applyRecords(text string, recs []editRecord) (string, error) {
	type span struct {
		at, end int
		new     string
	}
	spans := make([]span, 0, len(recs))
	for i, r := range recs {
		at, from := -1, 0
		for n := 0; n < r.Occurrence && r.Old != ""; n++ {
			j := strings.Index(text[from:], r.Old)
			if j < 0 {
				at = -1
				break
			}
			at, from = from+j, from+j+len(r.Old)
		}
		if at < 0 || r.Occurrence < 1 {
			return "", fmt.Errorf("record %d does not resolve", i+1)
		}
		spans = append(spans, span{at, at + len(r.Old), r.New})
	}
	slices.SortFunc(spans, func(a, b span) int { return a.at - b.at })
	var b strings.Builder
	last := 0
	for _, sp := range spans {
		if sp.at < last {
			return "", fmt.Errorf("records overlap at byte %d", sp.at)
		}
		b.WriteString(text[last:sp.at] + sp.new)
		last = sp.end
	}
	b.WriteString(text[last:])
	return b.String(), nil
}

// lineRecords turns a line-for-line rewrite of text into apply records, one
// per changed line, and proves them: applied to before they must give after
// exactly, or they are refused (a line repeated inside a longer one could
// shift an occurrence count).
func lineRecords(before, after, explain string) ([]editRecord, error) {
	bl, al := strings.Split(before, "\n"), strings.Split(after, "\n")
	if len(bl) != len(al) {
		return nil, fmt.Errorf("the rewrite changed the line count (%d → %d)", len(bl), len(al))
	}
	var recs []editRecord
	offset := 0
	for i, l := range bl {
		if l != al[i] {
			recs = append(recs, editRecord{Old: l, Occurrence: occurrenceAt(before, l, offset), New: al[i], Explain: explain})
		}
		offset += len(l) + 1
	}
	if len(recs) == 0 {
		return nil, nil
	}
	got, err := applyRecords(before, recs)
	if err != nil {
		return nil, err
	}
	if got != after {
		return nil, fmt.Errorf("the edit records do not reproduce the rewrite")
	}
	return recs, nil
}

// occurrenceAt is the 1-based count, as applyRecords counts, of the match of
// s that starts at offset.
func occurrenceAt(text, s string, offset int) int {
	n, from := 0, 0
	for {
		j := strings.Index(text[from:], s)
		if j < 0 || from+j > offset {
			return n
		}
		n++
		if from+j == offset {
			return n
		}
		from += j + len(s)
	}
}
