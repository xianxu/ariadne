package main

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAssetVisibilityHostResponseFailsClosed(t *testing.T) {
	for _, tc := range []struct{ name, raw, want string }{
		{"private", `{"visibility":"PRIVATE"}`, "PRIVATE"},
		{"public", `{"visibility":"PUBLIC"}`, "PUBLIC"},
		{"duplicate", `{"visibility":"PUBLIC","visibility":"PRIVATE"}`, ""},
		{"unknown", `{"visibility":"OTHER"}`, ""},
		{"extra", `{"visibility":"PRIVATE","unexpected":true}`, ""},
		{"trailing", `{"visibility":"PRIVATE"} {}`, ""},
		{"oversize", strings.Repeat(" ", 4096) + `{"visibility":"PRIVATE"}`, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			response := filepath.Join(t.TempDir(), "response")
			if err := os.WriteFile(response, []byte(tc.raw), 0600); err != nil {
				t.Fatal(err)
			}
			t.Setenv("ASSET_VISIBILITY_RESPONSE", response)
			fakeLandingGH(t, "cat \"$ASSET_VISIBILITY_RESPONSE\"\n")
			got, err := (realGH{}).RepoVisibility(context.Background(), "example/repo")
			if tc.want == "" {
				if err == nil {
					t.Fatalf("accepted malformed host visibility: %q", got)
				}
			} else if err != nil || got != tc.want {
				t.Fatalf("got %q, %v", got, err)
			}
		})
	}
}
