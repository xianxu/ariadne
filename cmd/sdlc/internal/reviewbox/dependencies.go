package reviewbox

import (
	bytebuf "bytes"
	"encoding/hex"
	"fmt"
	"sort"
	"strconv"
	"strings"
)

// Dependency describes one pinned disposable checkout made available to a review.
type Dependency struct {
	Kind, Dir, Head, FetchAge string
}

// DependencyManifest describes the review's dependency bindings without changing
// the caller's environment or the order of its dependency slice.
func DependencyManifest(deps []Dependency, fenced bool) string {
	if len(deps) == 0 {
		return ""
	}
	ordered := append([]Dependency(nil), deps...)
	sort.Slice(ordered, func(i, j int) bool {
		a, b := ordered[i], ordered[j]
		if a.Kind != b.Kind {
			return a.Kind < b.Kind
		}
		if a.Dir != b.Dir {
			return a.Dir < b.Dir
		}
		if a.Head != b.Head {
			return a.Head < b.Head
		}
		return a.FetchAge < b.FetchAge
	})
	mode := "writable disposable clone; all writes are discarded with the review"
	if fenced {
		mode = "read-only (OS fenced) disposable clone"
	}
	var out strings.Builder
	out.WriteString("Review dependencies (pinned once; no fetch during review):\n")
	for _, dep := range ordered {
		age := dep.FetchAge
		if age == "" {
			age = "unknown"
		}
		fmt.Fprintf(&out, "SDLC_DEP_%s=%q\n  head: %s\n  fetch age: %s\n  mode: %s\n", strings.ToUpper(dep.Kind), dep.Dir, dep.Head, age, mode)
	}
	return out.String()
}

// DependencyTreeBudget validates the NUL-delimited output of git ls-tree -r -l
// --full-tree and bounds the entries and aggregate blob size before cloning.
func DependencyTreeBudget(raw []byte) (entries int, bytes int64, err error) {
	const maxMetadata = 32 << 20
	const maxEntries = 100000
	const maxBytes int64 = 512 << 20
	if len(raw) > maxMetadata {
		return 0, 0, fmt.Errorf("dependency tree metadata exceeds %d bytes", maxMetadata)
	}
	for len(raw) > 0 {
		end := bytebuf.IndexByte(raw, 0)
		if end < 0 {
			return entries, bytes, fmt.Errorf("dependency tree has an unterminated entry")
		}
		record := raw[:end]
		raw = raw[end+1:]
		tab := bytebuf.IndexByte(record, '\t')
		if tab < 0 || tab == len(record)-1 {
			return entries, bytes, fmt.Errorf("dependency tree entry has missing metadata or path")
		}
		metadata := string(record[:tab])
		if strings.ContainsAny(metadata, "\r\n\v\f") {
			return entries, bytes, fmt.Errorf("dependency tree entry has malformed metadata")
		}
		fields := strings.Fields(metadata)
		if len(fields) != 4 {
			return entries, bytes, fmt.Errorf("dependency tree entry has malformed metadata")
		}
		if fields[0] == "160000" {
			return entries, bytes, fmt.Errorf("dependency tree contains an unsupported gitlink")
		}
		if (fields[0] != "100644" && fields[0] != "100755" && fields[0] != "120000") || fields[1] != "blob" {
			return entries, bytes, fmt.Errorf("dependency tree contains an unsupported mode or object type")
		}
		oid := fields[2]
		if len(oid) != 40 && len(oid) != 64 {
			return entries, bytes, fmt.Errorf("dependency tree entry has an invalid object ID")
		}
		if _, decodeErr := hex.DecodeString(oid); decodeErr != nil {
			return entries, bytes, fmt.Errorf("dependency tree entry has an invalid object ID")
		}
		for _, c := range fields[3] {
			if c < '0' || c > '9' {
				return entries, bytes, fmt.Errorf("dependency tree entry has an invalid blob size")
			}
		}
		size, parseErr := strconv.ParseInt(fields[3], 10, 64)
		if parseErr != nil {
			return entries, bytes, fmt.Errorf("dependency tree entry has an invalid blob size")
		}
		if size > maxBytes-bytes {
			return entries, bytes, fmt.Errorf("dependency tree blobs exceed %d bytes", maxBytes)
		}
		entries++
		if entries > maxEntries {
			return entries, bytes, fmt.Errorf("dependency tree exceeds %d entries", maxEntries)
		}
		bytes += size
	}
	return entries, bytes, nil
}
