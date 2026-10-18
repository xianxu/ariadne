package reviewbox

import (
	"encoding/hex"
	"fmt"
	"slices"
	"strings"
)

// authorityKeys selects, from the source's config, the keys main authority
// needs: main's upstream and each remote's fetch refspec (#447). Never a URL,
// a credential, an insteadOf or a push setting.
const authorityKeys = `^(branch\.main\.(remote|merge)|remote\..+\.fetch)$`

// inertURL is every copied remote's URL and push URL: git has no helper for
// it, so a fetch or push from the review checkout fails, and nothing in the
// checkout names the source or its real remote.
const inertURL = "readonly://review-checkout"

// authority is the slot's main authority as a review checkout copies it
// (#447): the config keys and refs a repository test reads to find main and
// the merge base, so a test that runs in the slot runs the same way in
// review. Remotes are the remotes it names; each gets the inert URL.
type authority struct {
	config  [][2]string // key, value; a key may repeat (fetch refspecs)
	refs    [][2]string // ref, object ID
	remotes []string
}

// mainAuthority parses the source's `git config --get-regexp authorityKeys`
// output and its `for-each-ref --format='%(objectname) %(refname)'` rows into
// the authority to copy. Only refs/heads/main and refs/remotes/<R>/main pass,
// and only the fetch refspecs of a remote that branch.main or such a ref
// names. A malformed row refuses rather than copying a guess.
func mainAuthority(config, refs string) (authority, error) {
	var a authority
	addRemote := func(r string) {
		if r != "" && r != "." && !slices.Contains(a.remotes, r) {
			a.remotes = append(a.remotes, r)
		}
	}
	for _, row := range lines(refs) {
		oid, ref, ok := strings.Cut(row, " ")
		if !ok || !objectID(oid) {
			return authority{}, fmt.Errorf("review checkout: unreadable ref row %q", row)
		}
		switch {
		case ref == "refs/heads/main":
		case strings.HasPrefix(ref, "refs/remotes/") && strings.HasSuffix(ref, "/main") && len(ref) > len("refs/remotes//main"):
			addRemote(strings.TrimSuffix(strings.TrimPrefix(ref, "refs/remotes/"), "/main"))
		default:
			continue
		}
		a.refs = append(a.refs, [2]string{ref, oid})
	}
	var fetches [][2]string
	for _, row := range lines(config) {
		key, value, ok := strings.Cut(row, " ")
		if !ok || value == "" {
			return authority{}, fmt.Errorf("review checkout: unreadable config row %q", row)
		}
		switch {
		case key == "branch.main.remote":
			addRemote(value)
			a.config = append(a.config, [2]string{key, value})
		case key == "branch.main.merge":
			a.config = append(a.config, [2]string{key, value})
		case strings.HasPrefix(key, "remote.") && strings.HasSuffix(key, ".fetch"):
			fetches = append(fetches, [2]string{key, value})
		default:
			return authority{}, fmt.Errorf("review checkout: config key %q is not main authority", key)
		}
	}
	for _, f := range fetches {
		if slices.Contains(a.remotes, strings.TrimSuffix(strings.TrimPrefix(f[0], "remote."), ".fetch")) {
			a.config = append(a.config, f)
		}
	}
	return a, nil
}

func lines(s string) []string {
	var out []string
	for _, l := range strings.Split(s, "\n") {
		if l = strings.TrimSpace(l); l != "" {
			out = append(out, l)
		}
	}
	return out
}

func objectID(s string) bool {
	_, err := hex.DecodeString(s)
	return err == nil && (len(s) == 40 || len(s) == 64)
}
