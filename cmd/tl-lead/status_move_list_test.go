package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMoveRegistrationSnapshotReply(t *testing.T) {
	good := `{"Code":"ok","Registrations":[{"ID":"one","File":"/a.md","Owner":"ops:0","Future":{"x":1e400}},{"ID":"two","File":"/b.md","Owner":"ops-echo:0"}],"Future":1e400}`
	got, err := decodeMoveRegistrationSnapshot([]byte(good))
	if err != nil || len(got) != 2 || got["/a.md"] == nil || got["/b.md"] == nil || got["/a.md"].ID != "one" || got["/b.md"].Owner != "ops-echo:0" {
		t.Fatalf("snapshot = %+v, %v", got, err)
	}
	got, err = decodeMoveRegistrationSnapshot([]byte(`{"Code":"ok","Registrations":[]}`))
	if err != nil || got == nil || len(got) != 0 {
		t.Fatalf("empty snapshot = %+v, %v", got, err)
	}
}

func TestMoveRegistrationSnapshotRefusesInvalidReply(t *testing.T) {
	for name, raw := range map[string]string{
		"missing code":       `{"Registrations":[]}`,
		"wrong code":         `{"Code":"unavailable","Registrations":[]}`,
		"duplicate code":     `{"Code":"ok","code":"ok","Registrations":[]}`,
		"missing array":      `{"Code":"ok"}`,
		"null array":         `{"Code":"ok","Registrations":null}`,
		"object array":       `{"Code":"ok","Registrations":{}}`,
		"duplicate array":    `{"Code":"ok","Registrations":[],"registrations":[]}`,
		"null registration":  `{"Code":"ok","Registrations":[null]}`,
		"missing identity":   `{"Code":"ok","Registrations":[{"File":"/a.md","Owner":"ops:0"}]}`,
		"relative file":      `{"Code":"ok","Registrations":[{"ID":"one","File":"a.md","Owner":"ops:0"}]}`,
		"unclean file":       `{"Code":"ok","Registrations":[{"ID":"one","File":"/a/../b.md","Owner":"ops:0"}]}`,
		"empty file":         `{"Code":"ok","Registrations":[{"ID":"one","File":"","Owner":"ops:0"}]}`,
		"null owner":         `{"Code":"ok","Registrations":[{"ID":"one","File":"/a.md","Owner":null}]}`,
		"numeric identity":   `{"Code":"ok","Registrations":[{"ID":3,"File":"/a.md","Owner":"ops:0"}]}`,
		"duplicate field":    `{"Code":"ok","Registrations":[{"ID":"one","id":"two","File":"/a.md","Owner":"ops:0"}]}`,
		"duplicate file":     `{"Code":"ok","Registrations":[{"ID":"one","File":"/a.md","Owner":"ops:0"},{"ID":"two","File":"/a.md","Owner":"ops:0"}]}`,
		"duplicate identity": `{"Code":"ok","Registrations":[{"ID":"one","File":"/a.md","Owner":"ops:0"},{"ID":"one","File":"/b.md","Owner":"ops:0"}]}`,
		"trailing JSON":      `{"Code":"ok","Registrations":[]} {}`,
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := decodeMoveRegistrationSnapshot([]byte(raw)); err == nil {
				t.Fatalf("accepted %s", raw)
			}
		})
	}
}

func TestMoveRegistrationSnapshotUsesOneProcess(t *testing.T) {
	for _, n := range []int{5, 10} {
		t.Run(fmt.Sprint(n), func(t *testing.T) {
			dir := t.TempDir()
			bin, calls := filepath.Join(dir, "couch"), filepath.Join(dir, "calls")
			script := "#!/bin/sh\nprintf '%s\\n' \"$*\" >> \"$MOVE_LIST_CALLS\"\nprintf '%s' \"$MOVE_LIST_REPLY\"\n"
			if err := os.WriteFile(bin, []byte(script), 0755); err != nil {
				t.Fatal(err)
			}
			regs := make([]syncRegistration, n)
			for i := range regs {
				regs[i] = syncRegistration{ID: fmt.Sprint(i), File: fmt.Sprintf("/side/%d.md", i), Owner: "ops:0"}
			}
			raw, err := json.Marshal(struct {
				Code          string
				Registrations []syncRegistration
			}{"ok", regs})
			if err != nil {
				t.Fatal(err)
			}
			t.Setenv("MOVE_LIST_CALLS", calls)
			t.Setenv("MOVE_LIST_REPLY", string(raw))
			got, err := (couchSessionSync{bin: bin}).RegistrationSnapshot()
			if err != nil || len(got) != n {
				t.Fatalf("snapshot size = %d, error = %v", len(got), err)
			}
			data, err := os.ReadFile(calls)
			if err != nil {
				t.Fatal(err)
			}
			if strings.TrimSpace(string(data)) != "--session-sync list --json" {
				t.Fatalf("expected one list process for %d registrations, got %q", n, data)
			}
		})
	}
}
