package reviewbox

import (
	"reflect"
	"testing"
)

func TestWrapPerFence(t *testing.T) {
	protect := []string{"/src", "/src/.git"}
	cases := []struct {
		fence Fence
		name  string
		args  []string
	}{
		{FenceSeatbelt, "sandbox-exec", []string{"-p", `(version 1)(allow default)(deny file-write* (subpath "/src") (subpath "/src/.git"))`, "env", "-C", "/box", "claude", "-p", "x"}},
		{FenceBwrap, "bwrap", []string{"--dev-bind", "/", "/", "--ro-bind", "/src", "/src", "--ro-bind", "/src/.git", "/src/.git", "--chdir", "/box", "claude", "-p", "x"}},
		{FenceNone, "env", []string{"-C", "/box", "claude", "-p", "x"}},
	}
	for _, c := range cases {
		iso := Isolation{Fence: c.fence, Dir: "/box", Protect: protect}
		name, args := iso.Wrap("claude", []string{"-p", "x"})
		if name != c.name || !reflect.DeepEqual(args, c.args) {
			t.Errorf("%s: got %s %q, want %s %q", c.fence, name, args, c.name, c.args)
		}
	}
}

func TestWrapDoesNotAliasCallerArgs(t *testing.T) {
	args := make([]string, 1, 8)
	args[0] = "x"
	_, out := Isolation{Fence: FenceNone, Dir: "/box"}.Wrap("claude", args)
	out[len(out)-1] = "mutated"
	if args[0] != "x" {
		t.Fatalf("Wrap aliased the caller's slice: %q", args)
	}
}

func TestSeatbeltProfileEscapes(t *testing.T) {
	got := seatbeltProfile([]string{`/a "b"\c`})
	want := `(version 1)(allow default)(deny file-write* (subpath "/a \"b\"\\c"))`
	if got != want {
		t.Fatalf("got %s, want %s", got, want)
	}
}

func TestLevel(t *testing.T) {
	for fence, want := range map[Fence]string{FenceSeatbelt: LevelFenced, FenceBwrap: LevelFenced, FenceNone: LevelCheckout, "": LevelCheckout} {
		if got := (Isolation{Fence: fence}).Level(); got != want {
			t.Errorf("%q: got %s, want %s", fence, got, want)
		}
	}
}
