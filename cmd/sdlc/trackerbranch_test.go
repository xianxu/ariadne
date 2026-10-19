package main

import (
	"strings"
	"testing"
)

func TestRequireIssueBranch(t *testing.T) {
	const path = "workshop/issues/000009-s09.md"
	for _, tc := range []struct {
		name, current, named string
		wantError             bool
	}{
		{"correct", "000009-s09", "", false},
		{"other issue", "000010-other", "000010-other", true},
		{"rest", "main-slot1", "main-slot1", true},
		{"detached", "", "a detached HEAD", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := requireIssueBranch("000009", path, tc.current)
			if !tc.wantError {
				if err != nil {
					t.Fatal(err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), "000009-s09") || !strings.Contains(err.Error(), tc.named) {
				t.Fatalf("wrong branch error = %v", err)
			}
		})
	}
}
