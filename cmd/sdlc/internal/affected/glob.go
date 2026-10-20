package affected

import (
	"path"
	"strings"
)

// Match reports whether name matches pattern (#473). Both are '/'-separated; each pattern segment matches one name
// segment by path.Match, except a whole "**" segment, which matches zero or more segments. A malformed segment is an
// error whether or not the name would reach it.
func Match(pattern, name string) (bool, error) {
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
	// memo[i][j]: pat[i:] vs segs[j:]; 0 unknown, 1 match, 2 no match.
	memo := make([][]byte, len(pat)+1)
	for i := range memo {
		memo[i] = make([]byte, len(segs)+1)
	}
	var match func(i, j int) bool
	match = func(i, j int) bool {
		if memo[i][j] != 0 {
			return memo[i][j] == 1
		}
		var ok bool
		switch {
		case i == len(pat):
			ok = j == len(segs)
		case pat[i] == "**":
			ok = match(i+1, j) || (j < len(segs) && match(i, j+1))
		case j < len(segs):
			m, _ := path.Match(pat[i], segs[j])
			ok = m && match(i+1, j+1)
		}
		memo[i][j] = 2
		if ok {
			memo[i][j] = 1
		}
		return ok
	}
	return match(0, 0), nil
}

// validGlob reports whether pattern is non-empty and well formed for Match.
func validGlob(pattern string) bool {
	if pattern == "" {
		return false
	}
	_, err := Match(pattern, "")
	return err == nil
}
