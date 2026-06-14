// Package gomodx is the ONE home for weave's pure go.mod text parsing
// (ARCH-DRY, ARCH-PURE). go.mod reasoning was spread across three spots —
// plan/apply.go and golden/golden.go each carried a verbatim copy of the
// module-line parser, and weavefs/gomod.go read the `go` directive — so this
// package collects the pure text helpers in one importable, dependency-free
// place. It sits below plan, golden, and weavefs (no internal imports), so all
// three can import it without a cycle. The IO/exec seams stay in weavefs; this
// is text-in, value-out only.
package gomodx

import "strings"

// ModuleLine extracts the module path from go.mod content (awk '/^module /
// {print $2; exit}'). Pure. "" when absent.
func ModuleLine(content string) string {
	for _, line := range strings.Split(content, "\n") {
		if strings.HasPrefix(line, "module ") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				return fields[1]
			}
		}
	}
	return ""
}

// GoDirective extracts the version from the first `go ` line of go.mod content
// (awk '/^go / {print $2; exit}'). Pure. "" when absent.
func GoDirective(content string) string {
	for _, line := range strings.Split(content, "\n") {
		if strings.HasPrefix(line, "go ") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				return fields[1]
			}
		}
	}
	return ""
}

// HasTool reports whether go.mod content declares `tool <importPath>` — either
// a single-line directive (`tool <path>`) or a row inside a `tool ( … )` block.
// A plain substring check on the import path would false-positive on a require
// line, so match the path as a standalone field on a tool-bearing line. Pure.
func HasTool(content, importPath string) bool {
	inBlock := false
	for _, line := range strings.Split(content, "\n") {
		trimmed := strings.TrimSpace(line)
		switch {
		case trimmed == "tool (":
			inBlock = true
			continue
		case inBlock && trimmed == ")":
			inBlock = false
			continue
		case inBlock:
			if trimmed == importPath {
				return true
			}
		case strings.HasPrefix(trimmed, "tool "):
			fields := strings.Fields(trimmed)
			if len(fields) == 2 && fields[1] == importPath {
				return true
			}
		}
	}
	return false
}
