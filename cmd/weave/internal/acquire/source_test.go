package acquire

import "testing"

func TestNormalizeSource(t *testing.T) {
	for _, raw := range []string{"github.com/Org/base", "https://github.com/Org/base.git", "git@github.com:Org/base.git", "ssh://git@github.com/Org/base.git"} {
		got, err := NormalizeSource(raw)
		if err != nil || got.Identity != "github.com/org/base" || got.Name != "base" {
			t.Fatalf("%s: %+v %v", raw, got, err)
		}
		if raw == "github.com/Org/base" && got.URL != "https://github.com/Org/base.git" {
			t.Fatal(got)
		}
		if raw != "github.com/Org/base" && got.URL != raw {
			t.Fatal("changed transport", got)
		}
	}
	for _, raw := range []string{"", "https://github.com/org", "--upload-pack=evil", "https://github.com/org/../bad"} {
		if _, err := NormalizeSource(raw); err == nil {
			t.Fatalf("accepted %q", raw)
		}
	}
}

func TestNormalizeSourceRejectsUnrecordableAndCredentials(t *testing.T) {
	for _, raw := range []string{"https://user:secret@github.com/org/repo.git", "https://token@github.com/org/repo.git", "https://github.com/org/repo#fragment", "/path with spaces/repo.git", "/path\nrepo.git"} {
		if _, err := NormalizeSource(raw); err == nil {
			t.Fatalf("accepted %q", raw)
		}
	}
}
