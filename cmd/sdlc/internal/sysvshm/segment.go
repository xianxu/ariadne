// Package sysvshm describes bounded host observations, never cleanup authority.
package sysvshm

// Segment is one observed allocation. Attachments is not its identity.
type Segment struct {
	ID          int
	Key         string
	UID         int
	Creator     int
	Attachments uint64
	Size        uint64
	Changed     string
}

// Snapshot distinguishes complete emptiness from unavailable observation.
type Snapshot struct {
	Segments    []Segment
	Unavailable string
}

// Witness is a creator observed in the owned process group, with incarnation.
type Witness struct{ Name, Start string }
type Candidate struct {
	Segment Segment
	Witness Witness
}
type Report struct {
	Before, After int
	Candidates    []Candidate
	Unavailable   string
}
