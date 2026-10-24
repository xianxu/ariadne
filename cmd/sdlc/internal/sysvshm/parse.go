package sysvshm

import (
	"fmt"
	"strconv"
	"strings"
)

const MaxBytes = 1 << 20
const MaxRows = 4096

func ParseLinux(data []byte, uid int) Snapshot                { return parse(data, uid, "", false) }
func ParseDarwin(data []byte, uid int, owner string) Snapshot { return parse(data, uid, owner, true) }

// parse projects required columns while tolerating unrelated future columns.
// A malformed foreign row also invalidates the observation: partial is not empty.
func parse(data []byte, uid int, owner string, darwin bool) Snapshot {
	bad := func(reason string) Snapshot { return Snapshot{Unavailable: reason} }
	if len(data) > MaxBytes {
		return bad("SysV table exceeds byte limit")
	}
	names := []string{"key", "shmid", "size", "cpid", "nattch", "uid", "ctime"}
	if darwin {
		names = []string{"T", "ID", "KEY", "OWNER", "NATTCH", "SEGSZ", "CPID", "CTIME"}
	}
	var columns map[string]int
	var result Snapshot
	seen := map[int]bool{}
	rows := 0
	for _, line := range strings.Split(string(data), "\n") {
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}
		header := fields[0] == "key" || fields[0] == "T"
		if header {
			if columns != nil {
				return bad("repeated SysV header")
			}
			columns = map[string]int{}
			for i, name := range fields {
				if _, exists := columns[name]; exists {
					return bad("duplicate SysV column")
				}
				columns[name] = i
			}
			for _, name := range names {
				if _, ok := columns[name]; !ok {
					return bad("missing SysV column " + name)
				}
			}
			continue
		}
		if darwin && (strings.HasPrefix(line, "IPC status from ") || strings.TrimSpace(line) == "Shared Memory:") {
			continue
		}
		if columns == nil {
			return bad("missing SysV header")
		}
		rows++
		if rows > MaxRows || len(fields) != len(columns) {
			return bad("incomplete or excessive SysV rows")
		}
		at := func(name string) string { return fields[columns[name]] }
		integer := func(name string) (uint64, error) { return strconv.ParseUint(at(name), 10, 63) }
		idName, sizeName, creatorName, attachName, keyName, changeName := "shmid", "size", "cpid", "nattch", "key", "ctime"
		if darwin {
			idName, sizeName, creatorName, attachName, keyName, changeName = "ID", "SEGSZ", "CPID", "NATTCH", "KEY", "CTIME"
		}
		values := make([]uint64, 4)
		for i, name := range []string{idName, sizeName, creatorName, attachName} {
			value, err := integer(name)
			if err != nil {
				return bad(fmt.Sprintf("invalid SysV %s", name))
			}
			values[i] = value
		}
		id := int(values[0])
		if seen[id] {
			return bad("duplicate SysV ID")
		}
		seen[id] = true
		rowUID := uid
		if darwin {
			if at("T") != "m" {
				return bad("unexpected IPC type")
			}
			if _, err := strconv.ParseUint(at(keyName), 0, 64); err != nil {
				return bad("invalid SysV key")
			}
			if at("OWNER") != owner && at("OWNER") != strconv.Itoa(uid) {
				continue
			}
		} else {
			if _, err := strconv.ParseInt(at(keyName), 10, 64); err != nil {
				return bad("invalid SysV key")
			}
			value, err := integer("uid")
			if err != nil {
				return bad("invalid SysV UID")
			}
			rowUID = int(value)
			if _, err := integer("ctime"); err != nil {
				return bad("invalid SysV ctime")
			}
			if rowUID != uid {
				continue
			}
		}
		result.Segments = append(result.Segments, Segment{ID: id, Key: at(keyName), UID: rowUID, Creator: int(values[2]), Attachments: values[3], Size: values[1], Changed: at(changeName)})
	}
	if columns == nil {
		return bad("missing SysV header")
	}
	return result
}
