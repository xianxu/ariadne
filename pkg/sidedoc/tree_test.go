package sidedoc

import "testing"

func TestSideTree(t *testing.T) {
	if got := SideDir("workshop/issues/000424-x.md"); got != "workshop/issues/000424-x.side" {
		t.Errorf("SideDir = %s", got)
	}
	for _, c := range []struct {
		member, want string
		ok           bool
	}{
		{"w/i/000424-x.side/01-a.md", "w/h/i/000424-x.side/01-a.md", true},
		{"w/i/000424-x.side/sub/02-b.md", "w/h/i/000424-x.side/sub/02-b.md", true},
		{"w/i/000424-x.md", "", false},
		{"w/i/000424-x.sidecar/01-a.md", "", false},
		{"w/i/000424-x.side/", "", false},
	} {
		got, ok := MoveMember(c.member, "w/i/000424-x.md", "w/h/i/000424-x.md")
		if got != c.want || ok != c.ok {
			t.Errorf("MoveMember(%s) = %s, %v", c.member, got, ok)
		}
	}
	for p, want := range map[string]bool{
		"w/i/000424-x.side/01-a.md": true, "w/i/000424-x.md": false, "w/i/x.side": false, "w/.side/a.md": false,
	} {
		if InSideDir(p) != want {
			t.Errorf("InSideDir(%s) = %v", p, !want)
		}
	}
}

// Owner is a path-only projection: neither the side root nor its synthetic
// owner needs to exist, and a nested side root remains part of the first tree.
func TestOwner(t *testing.T) {
	for _, tt := range []struct {
		path, want string
		ok         bool
	}{
		{"w/i/000424-x.side", "w/i/000424-x.md", true},
		{"w/i/000424-x.side/", "w/i/000424-x.md", true},
		{"w/i/000424-x.side/01-a.md", "w/i/000424-x.md", true},
		{"w/i/000424-x.side/sub/data.json", "w/i/000424-x.md", true},
		{"w/i/000424-x.side/sub/y.side/a.md", "w/i/000424-x.md", true},
		{"w/i/000424-x.side/./sub/../a.md", "w/i/000424-x.md", true},
		{"/w/i/x.side/a.md", "/w/i/x.md", true},
		{"x.side", "x.md", true},
		{"w/i/000424-x.md", "", false},
		{"w/i/000424-x.sidecar/a.md", "", false},
		{"w/.side/a.md", "", false},
		{"w/x.md.side", "", false},
		{"w/x.md.side/a.md", "", false},
		{"w/x.md.side/y.side/a.md", "", false},
		{"", "", false},
	} {
		t.Run(tt.path, func(t *testing.T) {
			got, ok := Owner(tt.path)
			if got != tt.want || ok != tt.ok {
				t.Errorf("Owner(%q) = %q, %v; want %q, %v", tt.path, got, ok, tt.want, tt.ok)
			}
		})
	}
}

func TestOwnerPreservesFolderRootExclusion(t *testing.T) {
	if got, ok := Folder("w/x.side"); got != "" || ok || InSideDir("w/x.side") {
		t.Fatalf("side root became its own member: Folder = %q, %v", got, ok)
	}
	if got, ok := Folder("w/x.md.side/a.md"); got != "w/x.md.side" || !ok {
		t.Fatalf("noncanonical naming can no longer be detected: Folder = %q, %v", got, ok)
	}
}
