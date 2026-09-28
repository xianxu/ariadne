package gitx

import (
	"bytes"
	"fmt"
)

// StatusEntry is one `git status --porcelain=v1 -z` entry. Path is exactly as
// git wrote it (spaces and all); Orig is a rename's or copy's source, else "".
type StatusEntry struct {
	XY   string
	Path string
	Orig string
}

// ParseStatusZ parses `git status --porcelain=v1 -z` output, byte-exact. It is
// the one reader of status paths: fixed-column slicing of trimmed, newline
// output loses the first entry's leading status space (" M path" → "M path"),
// which cost a close its gate ledger (#259). Output must be the raw stream,
// NUL-terminated; any malformed entry is an error, never a skipped path.
func ParseStatusZ(porcelain []byte) ([]StatusEntry, error) {
	if len(porcelain) == 0 {
		return nil, nil
	}
	if porcelain[len(porcelain)-1] != 0 {
		return nil, fmt.Errorf("status stream is missing final NUL terminator")
	}
	fields := bytes.Split(porcelain[:len(porcelain)-1], []byte{0})
	var entries []StatusEntry
	for i := 0; i < len(fields); i++ {
		field := fields[i]
		if len(field) < 4 || field[2] != ' ' || !ValidStatusCode(string(field[:2])) {
			return nil, fmt.Errorf("field %d is not XY+path status", i+1)
		}
		e := StatusEntry{XY: string(field[:2]), Path: string(field[3:])}
		if field[0] == 'R' || field[0] == 'C' || field[1] == 'R' || field[1] == 'C' {
			i++
			if i >= len(fields) || len(fields[i]) == 0 {
				return nil, fmt.Errorf("field %d rename/copy is missing source path", i)
			}
			e.Orig = string(fields[i])
		}
		entries = append(entries, e)
	}
	return entries, nil
}

// ValidStatusCode reports whether code is a porcelain v1 XY status.
func ValidStatusCode(code string) bool {
	switch code {
	case " A", " M", " T", " D",
		"M ", "MM", "MT", "MD",
		"T ", "TM", "TT", "TD",
		"A ", "AM", "AT", "AD",
		"D ",
		"R ", "RM", "RT", "RD",
		"C ", "CM", "CT", "CD",
		" R", " C",
		"DD", "AU", "UD", "UA", "DU", "AA", "UU",
		"??":
		return true
	default:
		return false
	}
}
