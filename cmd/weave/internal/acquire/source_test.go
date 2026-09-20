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
	for _, raw := range []string{"https://user:secret@github.com/org/repo.git", "https://token@github.com/org/repo.git", "https://github.com/org/repo#fragment", "/path with spaces/repo.git", "/path\nrepo.git", "file:///tmp/base.git?other", "file://remote/tmp/base.git"} {
		if _, err := NormalizeSource(raw); err == nil {
			t.Fatalf("accepted %q", raw)
		}
	}
}

func TestSourceIdentityPreservesEndpointDifferences(t *testing.T) {
	for _, pair := range [][2]string{
		{"file:repo.git", "repo.git"},
		{"ssh://git@example.com:2222/team/base.git", "ssh://git@example.com:3333/team/base.git"},
		{"https://example.com/team/base.git", "http://example.com/team/base.git"},
		{"https://example.com/team/base.git?tenant=a", "https://example.com/team/base.git?tenant=b"},
		{"ssh://git@github.com:2222/org/base.git", "git@github.com:org/base.git"},
		{"https://github.com:8443/org/base.git", "https://github.com/org/base.git"},
		{"git@example.com:team/base.git", "ssh://git@example.com/team/base.git"},
		{"https://example.com/team/base.git", "https://example.com/team/base"},
	} {
		a, err := NormalizeSource(pair[0])
		if err != nil {
			t.Fatal(err)
		}
		b, err := NormalizeSource(pair[1])
		if err != nil {
			t.Fatal(err)
		}
		if a.Identity == b.Identity {
			t.Fatalf("distinct endpoints compare equal: %q and %q", pair[0], pair[1])
		}
	}
	for _, raw := range []string{"https://github.com:443/org/base.git", "ssh://git@github.com:22/org/base.git"} {
		s, err := NormalizeSource(raw)
		if err != nil || s.Identity != "github.com/org/base" {
			t.Fatalf("standard GitHub endpoint %q: %+v %v", raw, s, err)
		}
	}
}
