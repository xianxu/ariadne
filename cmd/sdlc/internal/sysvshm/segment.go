// Package sysvshm describes bounded host observations, never cleanup authority.
package sysvshm

// Segment is one observed row. Mutable metadata never establishes allocation identity.
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

// Candidate is a newly observed ID, never process attribution.
type Candidate struct{ Segment Segment }
type Report struct {
	Before, After int
	Candidates    []Candidate
	Unavailable   string
}
