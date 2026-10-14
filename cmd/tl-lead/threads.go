package main

import (
	"errors"
	"strings"
)

// threadEntries splits a `## Threads` body into its top-level `- ` entries;
// continuation lines stay with their entry.
func threadEntries(body string) []string {
	var out []string
	for _, line := range strings.Split(strings.Trim(body, "\n"), "\n") {
		if strings.HasPrefix(line, "- ") || len(out) == 0 {
			out = append(out, line)
		} else {
			out[len(out)-1] += "\n" + line
		}
	}
	kept := out[:0]
	for _, e := range out {
		if strings.TrimSpace(e) != "" {
			kept = append(kept, e)
		}
	}
	return kept
}

// prependThread puts entry first and drops every existing entry that starts
// with one of the prefixes.
func prependThread(body, entry string, drop []string) (string, error) {
	entry = strings.Trim(entry, "\n")
	if !strings.HasPrefix(entry, "- ") {
		return "", errors.New("a Threads entry starts with '- '")
	}
	out := []string{entry}
	for _, e := range threadEntries(body) {
		dropped := false
		for _, p := range drop {
			if strings.HasPrefix(e, p) {
				dropped = true
				break
			}
		}
		if !dropped {
			out = append(out, e)
		}
	}
	return strings.Join(out, "\n"), nil
}
