package gomodx

import "testing"

func TestModuleLine(t *testing.T) {
	cases := map[string]string{
		"module example.com/owner\n\ngo 1.26\n": "example.com/owner",
		"go 1.23\nmodule x\n":                   "x",
		"require example.com/y v1.0.0\n":        "", // no module line
		"module\n":                              "", // bare keyword, no path
		"":                                      "",
	}
	for content, want := range cases {
		if got := ModuleLine(content); got != want {
			t.Errorf("ModuleLine(%q) = %q, want %q", content, got, want)
		}
	}
}

func TestGoDirective(t *testing.T) {
	cases := map[string]string{
		"module x\n\ngo 1.26\n":    "1.26",
		"go 1.23.4\nmodule x\n":    "1.23.4",
		"module x\nrequire y v1\n": "",
		"":                         "",
	}
	for content, want := range cases {
		if got := GoDirective(content); got != want {
			t.Errorf("GoDirective(%q) = %q, want %q", content, got, want)
		}
	}
}

func TestHasTool(t *testing.T) {
	const path = "example.com/owner/cmd/sdlc"

	// Single-line directive form.
	single := "module example.com/owner\n\ngo 1.26\n\ntool example.com/owner/cmd/sdlc\n"
	if !HasTool(single, path) {
		t.Errorf("HasTool(single-line) = false, want true")
	}

	// Block form: `tool ( … )` with the row indented inside the block. This
	// form was previously untested.
	block := "module example.com/owner\n\ngo 1.26\n\ntool (\n\texample.com/owner/cmd/sdlc\n\texample.com/owner/cmd/weave\n)\n"
	if !HasTool(block, path) {
		t.Errorf("HasTool(block form) = false, want true")
	}
	if !HasTool(block, "example.com/owner/cmd/weave") {
		t.Errorf("HasTool(block form, second row) = false, want true")
	}

	// Absent.
	if HasTool("module example.com/owner\n\ngo 1.26\n", path) {
		t.Errorf("HasTool(no tool directive) = true, want false")
	}

	// require-line false-positive guard: the import path appearing on a require
	// line must NOT count as a tool directive.
	requireOnly := "module example.com/owner\n\ngo 1.26\n\nrequire example.com/owner/cmd/sdlc v1.0.0\n"
	if HasTool(requireOnly, path) {
		t.Errorf("HasTool(require-line false positive) = true, want false")
	}

	// A tool block that closes must not let a later non-block line match by the
	// block branch; an unrelated single-line tool must not match a different path.
	other := "tool example.com/owner/cmd/other\n"
	if HasTool(other, path) {
		t.Errorf("HasTool(different single-line tool) = true, want false")
	}
}
