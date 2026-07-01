// Package helptext exposes //go:embed-backed help texts so cobra
// commands can populate their Long descriptions.
//
// Convention: each subcommand gets one Markdown file here, named by the
// command's stem (close.md, state.md, ...). The root narrative — the
// single workflow contract emitted by `sdlc --help` — lives in root.md.
//
// Why embed instead of inline strings: the prose grows beyond one
// paragraph per command; keeping it in Markdown files (rather than Go
// string literals) keeps it editable and diffable.
package helptext

import (
	"embed"
	"fmt"
	"strings"
)

//go:embed *.md
var fs embed.FS

// FS exposes the embedded help-text filesystem for enumeration (issue #153,
// `sdlc process-manual`). Returns the concrete embed.FS, which satisfies io/fs.FS where
// callers need it — so we avoid importing io/fs here (it would shadow `fs`).
func FS() embed.FS { return fs }

// Get returns the content of <name>.md with trailing whitespace
// trimmed to one terminating newline. Returns ok=false if absent.
func Get(name string) (string, bool) {
	b, err := fs.ReadFile(name + ".md")
	if err != nil {
		return "", false
	}
	return strings.TrimRight(string(b), "\n") + "\n", true
}

// MustGet returns the content of <name>.md, panicking if absent. Use
// for help texts that ship with the binary — a missing entry is a
// build-time bug, not a runtime condition.
func MustGet(name string) string {
	s, ok := Get(name)
	if !ok {
		panic(fmt.Sprintf("helptext: %s.md not embedded", name))
	}
	return s
}
