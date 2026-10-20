package affected

import (
	"path"
	"strings"
	"testing"
)

const exampleConfig = `{"version": 1,
 "command": ["make", "test-affected"],
 "weight": 1, "timeout_minutes": 30,
 "full":   ["Makefile*", "go.mod", "go.sum", ".sdlc/**", "tests/lib/**"],
 "groups": [
   {"name": "lua",      "paths": ["**/*.lua", "tests/lua/**"]},
   {"name": "pair-cli", "paths": ["tests/pair-*.sh"], "go": ["./cmd/pair"]}
 ]}`

func TestParseConfigExample(t *testing.T) {
	cfg, err := ParseConfig([]byte(exampleConfig))
	if err != nil {
		t.Fatalf("ParseConfig: %v", err)
	}
	if cfg.Version != 1 || len(cfg.Command) != 2 || cfg.Weight != 1 || cfg.TimeoutMinutes != 30 {
		t.Fatalf("cfg = %+v", cfg)
	}
	if len(cfg.Groups) != 2 || cfg.Groups[1].Go[0] != "./cmd/pair" {
		t.Fatalf("groups = %+v", cfg.Groups)
	}
}

func TestParseConfigDefaults(t *testing.T) {
	cfg, err := ParseConfig([]byte(`{"version":1,"command":["true"]}`))
	if err != nil {
		t.Fatalf("ParseConfig: %v", err)
	}
	if cfg.Weight != 1 || cfg.TimeoutMinutes != 30 {
		t.Fatalf("defaults: weight=%d timeout=%d", cfg.Weight, cfg.TimeoutMinutes)
	}
}

func TestParseConfigRefuses(t *testing.T) {
	cases := []struct{ name, json, want string }{
		{"unknown field", `{"version":1,"command":["x"],"extra":1}`, "unknown field"},
		{"duplicate key", `{"version":1,"command":["x"],"command":["y"]}`, "duplicate object key"},
		{"duplicate group key", `{"version":1,"command":["x"],"groups":[{"name":"a","name":"b","paths":["x"]}]}`, "duplicate object key"},
		{"trailing data", `{"version":1,"command":["x"]} {}`, "trailing"},
		{"trailing garbage", `{"version":1,"command":["x"]} x`, "invalid character"},
		{"not an object", `[]`, ""},
		{"empty", ``, ""},
		{"version 0", `{"command":["x"]}`, "version"},
		{"version 2", `{"version":2,"command":["x"]}`, "version"},
		{"no command", `{"version":1}`, "command"},
		{"empty argv element", `{"version":1,"command":["make",""]}`, "command"},
		{"weight too high", `{"version":1,"command":["x"],"weight":9}`, "weight"},
		{"weight negative", `{"version":1,"command":["x"],"weight":-1}`, "weight"},
		{"timeout too high", `{"version":1,"command":["x"],"timeout_minutes":121}`, "timeout_minutes"},
		{"timeout negative", `{"version":1,"command":["x"],"timeout_minutes":-5}`, "timeout_minutes"},
		{"bad full glob", `{"version":1,"command":["x"],"full":["a/["]}`, "full"},
		{"empty full glob", `{"version":1,"command":["x"],"full":[""]}`, "full"},
		{"bad group name", `{"version":1,"command":["x"],"groups":[{"name":"Lua","paths":["*.lua"]}]}`, "group name"},
		{"empty group name", `{"version":1,"command":["x"],"groups":[{"paths":["*.lua"]}]}`, "group name"},
		{"duplicate group", `{"version":1,"command":["x"],"groups":[{"name":"a","paths":["x"]},{"name":"a","paths":["y"]}]}`, "duplicate"},
		{"empty group", `{"version":1,"command":["x"],"groups":[{"name":"a"}]}`, "paths or go"},
		{"bad group glob", `{"version":1,"command":["x"],"groups":[{"name":"a","paths":["x/\\"]}]}`, "glob"},
		{"bad go root", `{"version":1,"command":["x"],"groups":[{"name":"a","go":["cmd/pair"]}]}`, "go root"},
		{"group unknown field", `{"version":1,"command":["x"],"groups":[{"name":"a","paths":["x"],"env":"y"}]}`, "unknown field"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, err := ParseConfig([]byte(c.json))
			if err == nil {
				t.Fatalf("ParseConfig(%s) = nil error", c.json)
			}
			if !strings.Contains(err.Error(), c.want) {
				t.Fatalf("error %q lacks %q", err, c.want)
			}
		})
	}
}

func TestMatch(t *testing.T) {
	cases := []struct {
		pattern, path string
		want          bool
	}{
		{"**/*.lua", "a.lua", true},
		{"**/*.lua", "a/b/c.lua", true},
		{"**/*.lua", "a/b/c.go", false},
		{".sdlc/**", ".sdlc", true},
		{".sdlc/**", ".sdlc/local-checks.json", true},
		{".sdlc/**", ".sdlcx/a", false},
		{"Makefile*", "Makefile", true},
		{"Makefile*", "sub/Makefile", false},
		{"a/**/b", "a/b", true},
		{"a/**/b", "a/x/y/b", true},
		{"a/**/**/b", "a/x/b", true},
		{"tests/pair-*.sh", "tests/pair-cli.sh", true},
		{"*", "a/b", false},
		{"**", "a/b", true},
	}
	for _, c := range cases {
		got, err := Match(c.pattern, c.path)
		if err != nil || got != c.want {
			t.Errorf("Match(%q, %q) = %v, %v; want %v", c.pattern, c.path, got, err, c.want)
		}
	}
	for _, bad := range []string{"[", "a/[", "x\\", "**/[", "a/**/\\"} {
		if _, err := Match(bad, "a/b"); err == nil {
			t.Errorf("Match(%q) did not error", bad)
		}
	}
}

// refMatch is the reference for FuzzMatch: every "**" segment expands to k "*" segments (k = 0..len(path segments)),
// then each segment pair goes through path.Match.
func refMatch(pattern, name string) (bool, error) {
	pat := strings.Split(pattern, "/")
	for _, seg := range pat {
		if seg == "**" {
			continue
		}
		if _, err := path.Match(seg, ""); err != nil {
			return false, err
		}
	}
	segs := strings.Split(name, "/")
	var try func(i int, acc []string) bool
	try = func(i int, acc []string) bool {
		if len(acc) > len(segs) {
			return false
		}
		if i == len(pat) {
			if len(acc) != len(segs) {
				return false
			}
			for j := range acc {
				if ok, _ := path.Match(acc[j], segs[j]); !ok {
					return false
				}
			}
			return true
		}
		if pat[i] != "**" {
			return try(i+1, append(acc[:len(acc):len(acc)], pat[i]))
		}
		for k := 0; k <= len(segs)-len(acc); k++ {
			next := acc[:len(acc):len(acc)]
			for range k {
				next = append(next, "*")
			}
			if try(i+1, next) {
				return true
			}
		}
		return false
	}
	return try(0, nil), nil
}

func FuzzMatch(f *testing.F) {
	seeds := [][2]string{
		{"**", ""}, {"**", "a"}, {"a//b", "a//b"}, {"a/**", "a"}, {"**/a", "a"}, {"**/**", "x/y"},
		{"a/**/**/b", "a/b"}, {"/**", "/x"}, {"**/", "x/"}, {"[", "a"}, {"\\", "a"}, {"a/[", "a/b"},
		{"*/*.go", "cmd/x.go"}, {"a**", "abc"}, {"[a-", "a"}, {"[]", "a"},
	}
	for _, s := range seeds {
		f.Add(s[0], s[1])
	}
	f.Fuzz(func(t *testing.T, pattern, name string) {
		if strings.Count(pattern, "**") > 4 || strings.Count(name, "/") > 8 || len(pattern) > 64 || len(name) > 64 {
			t.Skip()
		}
		got, gerr := Match(pattern, name)
		want, werr := refMatch(pattern, name)
		if (gerr != nil) != (werr != nil) {
			t.Fatalf("Match(%q, %q) err=%v, reference err=%v", pattern, name, gerr, werr)
		}
		if gerr == nil && got != want {
			t.Fatalf("Match(%q, %q) = %v, reference %v", pattern, name, got, want)
		}
		if gerr != nil && got {
			t.Fatalf("Match(%q, %q) = true with error", pattern, name)
		}
	})
}
