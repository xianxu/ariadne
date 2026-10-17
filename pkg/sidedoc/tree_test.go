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
