package issue

import "testing"

func TestParseFilename(t *testing.T) {
	tests := []struct {
		name     string
		wantID   string
		wantSlug string
		wantOK   bool
	}{
		{name: "000001-slug.md", wantID: "000001", wantSlug: "slug", wantOK: true},
		{name: "/tmp/000002-two-words.md", wantID: "000002", wantSlug: "two-words", wantOK: true},
		{name: "000003-.md", wantID: "000003", wantSlug: "", wantOK: true},
		{name: "00003-short.md"},
		{name: "0000003-long.md"},
		{name: "000003x-bleed.md"},
		{name: "abcdef-slug.md"},
		{name: "000003-slug.txt"},
		{name: "custom.md"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, slug, ok := ParseFilename(tt.name)
			if id != tt.wantID || slug != tt.wantSlug || ok != tt.wantOK {
				t.Fatalf("ParseFilename(%q) = %q, %q, %v; want %q, %q, %v",
					tt.name, id, slug, ok, tt.wantID, tt.wantSlug, tt.wantOK)
			}
		})
	}
}
