package main

import (
	"strings"
	"unicode/utf8"

	"github.com/xianxu/ariadne/pkg/sidedoc"
)

type docMarker struct {
	Line, Column int
	Kind, Text   string
}

// workingdocMarkers validates active review markup without interpreting any
// acceptance or rejection. Bare quoted anchors are references, not open items.
func workingdocMarkers(text string) []docMarker {
	markers, _ := scanWorkingdocMarkers(text)
	return markers
}

// The work counter counts scanner bytes, Markdown parser input bytes, visited
// AST nodes and literal-mask writes rather than elapsed time. Goldmark owns its
// parser internals; the counter verifies the surrounding linear traversal.
func scanWorkingdocMarkers(text string) ([]docMarker, int) {
	literal, work := sidedoc.CodeLiteralMask(text)
	var out []docMarker
	line, column := 1, 1
	for i := 0; i < len(text); {
		end := i
		kind := ""
		if !literal[i] && strings.HasPrefix(text[i:], "🤖") && i+len("🤖") < len(text) {
			opening := text[i+len("🤖")]
			if strings.ContainsRune("[<{~", rune(opening)) {
				end, kind = docMarkerEnd(text, i+len("🤖"), &work)
				if kind != "" {
					out = append(out, docMarker{line, column, kind, text[i:end]})
				}
			}
		}
		if end == i {
			_, width := utf8.DecodeRuneInString(text[i:])
			end += width
		}
		// Every byte is advanced exactly once, including consumed marker text.
		for i < end {
			_, width := utf8.DecodeRuneInString(text[i:end])
			if text[i] == '\n' {
				line++
				column = 1
			} else {
				column++
			}
			i += width
			work += width
		}
	}
	return out, work
}

// docMarkerEnd consumes one reference and the complete legacy commentary
// chain. A second reference is malformed. Delimiters can nest in turns (for
// example Markdown links); escaped delimiters remain turn content. Markers
// are line-local, so an unfinished enclosure refuses at the line boundary.
func docMarkerEnd(text string, start int, work *int) (int, string) {
	i := start
	reference := byte(0)
	turns, first := 0, byte(0)
	malformed := false
	for i < len(text) && text[i] != '\n' && text[i] != '\r' {
		open := text[i]
		if !strings.ContainsRune("[<{~", rune(open)) {
			break
		}
		isReference := open == '<' || open == '~'
		if isReference {
			if reference != 0 || turns != 0 {
				malformed = true
			}
			reference = open
		} else {
			if turns == 0 {
				first = open
			}
			turns++
		}
		close := map[byte]byte{'[': ']', '{': '}', '<': '>', '~': '~'}[open]
		i++
		depth := 1
		for i < len(text) && text[i] != '\n' && text[i] != '\r' {
			*work++
			if !isReference && text[i] == '\\' && i+1 < len(text) && text[i+1] != '\n' && text[i+1] != '\r' {
				i += 2
				continue
			}
			if !isReference && text[i] == open {
				depth++
			}
			if text[i] == close {
				depth--
			}
			i++
			if depth == 0 {
				break
			}
		}
		if depth != 0 {
			return i, "malformed"
		}
	}
	if malformed {
		return i, "malformed"
	}
	if turns > 1 || (first == '[' && reference != '~') {
		return i, "commentary"
	}
	if turns > 0 || reference == '~' {
		return i, "proposal"
	}
	return i, ""
}
