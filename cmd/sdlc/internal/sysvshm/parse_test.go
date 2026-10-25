package sysvshm

import (
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"
)

const linuxHeader = "key shmid perms size cpid lpid nattch uid gid cuid cgid atime dtime ctime rss swap\n"
const linuxRow = "123 7 600 56 42 0 0 501 20 501 20 0 0 12345 0 0\n"
const darwinHeader = "IPC status from <running system> as of Sat Oct 24\nT ID KEY MODE OWNER GROUP CREATOR CGROUP NATTCH SEGSZ CPID LPID ATIME DTIME CTIME\nShared Memory:\n"
const darwinRow = "m 7 0x7b --rw------- xianxu staff xianxu staff 0 56 42 0 no-entry no-entry 00:43:21\n"

func TestParseNativeSysVTables(t *testing.T) {
	linux := ParseLinux([]byte(linuxHeader+linuxRow), 501)
	want := Segment{ID: 7, Key: "123", UID: 501, Creator: 42, Size: 56, Changed: "12345"}
	if linux.Unavailable != "" || !reflect.DeepEqual(linux.Segments, []Segment{want}) {
		t.Fatalf("%+v", linux)
	}
	darwin := ParseDarwin([]byte(darwinHeader+darwinRow), 501, "xianxu")
	want.Key, want.Changed = "0x7b", "00:43:21"
	if darwin.Unavailable != "" || !reflect.DeepEqual(darwin.Segments, []Segment{want}) {
		t.Fatalf("%+v", darwin)
	}
	if got := ParseLinux([]byte(linuxHeader+linuxRow), 502); got.Unavailable != "" || len(got.Segments) != 0 {
		t.Fatalf("foreign: %+v", got)
	}
	if got := ParseDarwin([]byte(darwinHeader+darwinRow), 502, "other"); got.Unavailable != "" || len(got.Segments) != 0 {
		t.Fatalf("foreign: %+v", got)
	}
}

func TestParseSysVRejectsIncompleteAndAmbiguousTables(t *testing.T) {
	for _, tc := range []struct{ name, input string }{
		{"empty bytes", ""},
		{"missing column", strings.Replace(linuxHeader, " cpid", "", 1) + linuxRow},
		{"duplicate column", strings.Replace(linuxHeader, " cpid", " cpid cpid", 1) + linuxRow},
		{"duplicate id", linuxHeader + linuxRow + linuxRow},
		{"repeated header", linuxHeader + linuxHeader},
		{"truncated", linuxHeader + "123 7 600\n"},
		{"negative", linuxHeader + strings.Replace(linuxRow, "56", "-1", 1)},
		{"bad type", linuxHeader + strings.Replace(linuxRow, "42", "oops", 1)},
		{"oversize", strings.Repeat(" ", (1<<20)+1)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := ParseLinux([]byte(tc.input), 501); got.Unavailable == "" {
				t.Fatalf("accepted %+v", got)
			}
		})
	}
	if got := ParseDarwin([]byte(darwinHeader+strings.Replace(darwinRow, "m ", "q ", 1)), 501, "xianxu"); got.Unavailable == "" {
		t.Fatal("accepted foreign IPC type")
	}
}

func TestParseSysVEmptyAndFutureColumns(t *testing.T) {
	for _, input := range []string{linuxHeader, strings.TrimSpace(linuxHeader) + " future\n" + strings.TrimSpace(linuxRow) + " ignored\n"} {
		if got := ParseLinux([]byte(input), 501); got.Unavailable != "" {
			t.Fatalf("%+v", got)
		}
	}
	if got := ParseDarwin([]byte(darwinHeader), 501, "xianxu"); got.Unavailable != "" {
		t.Fatalf("%+v", got)
	}
}

func TestDarwinIncidentCaptureAndTableLimits(t *testing.T) {
	raw, err := os.ReadFile("testdata/darwin-incident.txt")
	if err != nil {
		t.Fatal(err)
	}
	snapshot := ParseDarwin(raw, 501, "xianxu")
	if snapshot.Unavailable != "" || len(snapshot.Segments) != 32 {
		t.Fatalf("incident %+v", snapshot)
	}
	for _, s := range snapshot.Segments {
		if s.Attachments != 0 || s.Size != 56 || s.Creator <= 0 {
			t.Fatalf("incident projection %+v", s)
		}
	}
	var oversized strings.Builder
	oversized.WriteString(linuxHeader)
	for i := 0; i <= MaxRows; i++ {
		fmt.Fprintf(&oversized, "123 %d 600 56 42 0 0 501 20 501 20 0 0 12345 0 0\n", i)
	}
	if got := ParseLinux([]byte(oversized.String()), 501); got.Unavailable == "" {
		t.Fatal("accepted excessive rows")
	}
}

func TestDeltaCountedProjectionAtEightAndSixteenRows(t *testing.T) {
	for _, size := range []int{8, 16} {
		before, after := Snapshot{}, Snapshot{}
		for i := 0; i < size; i++ {
			before.Segments = append(before.Segments, Segment{ID: i, UID: 501})
			after.Segments = append(after.Segments, Segment{ID: i + size, UID: 501})
		}
		got := Delta(before, after, 501)
		if got.Before != size || got.After != size || len(got.Candidates) != size {
			t.Fatalf("size=%d %+v", size, got)
		}
	}
}
