package main

import (
	"strings"
	"unicode/utf8"
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

// The work counter counts examined/written bytes and delimiter runs, rather
// than elapsed time, so the linear scanning contract can be tested reliably.
func scanWorkingdocMarkers(text string) ([]docMarker, int) {
	literal, work := docMarkerLiterals(text)
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

// docMarkerLiterals marks CommonMark-style backtick/tilde fenced blocks and
// exact-run backtick spans. Unmatched backticks are ordinary text. Indexing
// the next equal-length run avoids rescanning every unmatched opener; inline
// spans may cross newlines, but never cross a fenced block.
func docMarkerLiterals(text string) ([]bool, int) {
	literal := make([]bool, len(text))
	work := 0
	fenceChar, fenceWidth := byte(0), 0
	for start := 0; start < len(text); {
		end := start + strings.IndexByte(text[start:], '\n')
		if end < start {
			end = len(text)
		} else {
			end++
		}
		line := text[start:end]
		indent := 0
		for indent < len(line) && line[indent] == ' ' {
			indent++
		}
		width, char := 0, byte(0)
		if indent <= 3 && indent < len(line) && (line[indent] == '`' || line[indent] == '~') {
			char = line[indent]
			for indent+width < len(line) && line[indent+width] == char {
				width++
			}
		}
		inside := fenceWidth != 0
		if inside && char == fenceChar && width >= fenceWidth && strings.TrimSpace(line[indent+width:]) == "" {
			fenceWidth = 0
		} else if !inside && width >= 3 && (char != '`' || !strings.ContainsRune(line[indent+width:], '`')) {
			fenceChar, fenceWidth = char, width
			inside = true
		}
		if inside {
			for j := start; j < end; j++ {
				literal[j] = true
			}
		}
		work += 3 * len(line) // line lookup, fence recognition and mask writes
		start = end
	}
	type run struct{ start, end, next int }
	for start := 0; start < len(text); {
		if literal[start] {
			start++
			work++
			continue
		}
		end := start
		var runs []run
		for end < len(text) && !literal[end] {
			work++
			if text[end] != '`' {
				end++
				continue
			}
			runStart := end
			for end < len(text) && text[end] == '`' && !literal[end] {
				end++
				work++
			}
			// A backslash escapes a potential opener, but a closer inside a
			// span still matches literally, as CommonMark requires.
			runs = append(runs, run{runStart, end, -1})
		}
		next := make(map[int]int)
		for j := len(runs) - 1; j >= 0; j-- {
			work++
			width := runs[j].end - runs[j].start
			if k, ok := next[width]; ok {
				runs[j].next = k
			}
			next[width] = j
		}
		for j := 0; j < len(runs); j++ {
			work++
			backslashes := 0
			for k := runs[j].start - 1; k >= start && text[k] == '\\'; k-- {
				backslashes++
				work++
			}
			if backslashes%2 != 0 || runs[j].next < 0 {
				continue
			}
			close := runs[j].next
			for k := runs[j].start; k < runs[close].end; k++ {
				literal[k] = true
				work++
			}
			j = close
		}
		start = end
	}
	return literal, work
}
