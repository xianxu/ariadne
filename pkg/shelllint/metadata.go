package shelllint

import "strings"

// MetadataPipeFindings guards compiler identity checks in shell and workflow
// source. File-input quiet greps are safe; a pipe consumer must drain its input
// so a successful match cannot SIGPIPE the producer under pipefail (#534).
func MetadataPipeFindings(src string) []Finding {
	var out []Finding
	for i, raw := range strings.Split(src, "\n") {
		l := lex(raw)
		fields := strings.ReplaceAll(l.text, `\.`, ".")
		if !strings.Contains(fields, "vcs.revision") && !strings.Contains(fields, "vcs.modified") {
			continue
		}
		for _, at := range commandWords(l, "grep") {
			before := strings.TrimSpace(l.masked()[:at])
			if !strings.HasSuffix(before, "|") || strings.HasSuffix(before, "||") {
				continue
			}
			if quietGrep(l.arguments(at + len("grep"))) {
				out = append(out, Finding{Line: i + 1, Text: strings.TrimSpace(raw),
					Reason: "compiler metadata pipe uses quiet grep: drain the full input with grep without -q, redirecting output to /dev/null (#534)"})
			}
		}
	}
	return out
}

func quietGrep(args []string) bool {
	for i := 0; i < len(args); i++ {
		a := args[i]
		if a == "--" || !strings.HasPrefix(a, "-") {
			break
		}
		if a == "--quiet" || a == "--silent" {
			return true
		}
		if strings.HasPrefix(a, "--") {
			if a == "--regexp" || a == "--file" {
				i++ // the next argument is a pattern or filename, not flags
			}
			continue
		}
		for j, flag := range a[1:] {
			if flag == 'q' {
				return true
			}
			if strings.ContainsRune("ABCDefmd", flag) {
				if j == len(a)-2 {
					i++
				}
				break // the rest of this option is its value
			}
		}
	}
	return false
}
