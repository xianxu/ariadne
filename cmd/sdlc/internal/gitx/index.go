package gitx

import (
	"bytes"
	"fmt"
	"path"
	"strconv"
	"strings"
)

// IndexEntry preserves the flags and conflict stage that an ordinary status
// query can conceal. Path is byte-exact, including tabs and newlines.
type IndexEntry struct {
	Path, Tag, Mode, OID string
	Stage                int
}

// ParseIndexStageZ parses ls-files --stage -v -z. Unknown or truncated input
// is an observation error, never evidence that a checkout has no issue edits.
func ParseIndexStageZ(raw []byte) ([]IndexEntry, error) {
	if len(raw) == 0 {
		return nil, nil
	}
	if raw[len(raw)-1] != 0 {
		return nil, fmt.Errorf("truncated index inventory")
	}
	rows := bytes.Split(raw[:len(raw)-1], []byte{0})
	if len(rows) > SnapshotEntryLimit {
		return nil, fmt.Errorf("index inventory exceeds entry budget")
	}
	entries := make([]IndexEntry, 0, len(rows))
	for _, row := range rows {
		header, name, ok := bytes.Cut(row, []byte{'\t'})
		fields := strings.Fields(string(header))
		p := string(name)
		if !ok || len(fields) != 4 || len(fields[0]) != 1 || !strings.Contains("HSMRCK?hsmrck", fields[0]) || p == "" || path.IsAbs(p) || path.Clean(p) != p || strings.HasPrefix(p, "../") {
			return nil, fmt.Errorf("invalid index inventory entry")
		}
		switch fields[1] {
		case "100644", "100755", "120000", "160000":
		default:
			return nil, fmt.Errorf("invalid index mode")
		}
		oid, err := parseObjectID([]byte(fields[2]))
		if err != nil {
			return nil, err
		}
		stage, err := strconv.Atoi(fields[3])
		if err != nil || stage < 0 || stage > 3 {
			return nil, fmt.Errorf("invalid index conflict stage")
		}
		entries = append(entries, IndexEntry{Path: p, Tag: fields[0], Mode: fields[1], OID: oid, Stage: stage})
	}
	return entries, nil
}
