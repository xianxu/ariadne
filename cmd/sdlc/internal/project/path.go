package project

import (
	"fmt"
	"path/filepath"
	"regexp"
)

var slugRE = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

// ResolvePath is the sole project slug-to-file boundary. It accepts canonical
// kebab slugs only, so every caller remains contained beneath projectsDir.
func ResolvePath(projectsDir, slug string) (string, error) {
	if !slugRE.MatchString(slug) {
		return "", fmt.Errorf("invalid project slug %q (want lowercase kebab-case)", slug)
	}
	return filepath.Join(projectsDir, slug+".md"), nil
}
