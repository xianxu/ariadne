// Package acquire restores declared repository sources without updating existing checkouts.
package acquire

import (
	"fmt"
	"net/url"
	"path"
	"path/filepath"
	"strings"
	"unicode"
)

// Source separates transport from repository identity. GitHub's SSH and HTTPS
// forms compare equal while the user's authentication transport is retained.
type Source struct{ URL, Identity, Name string }

func NormalizeSource(raw string) (Source, error) {
	s := Source{URL: raw}
	if strings.Contains(raw, "#") || strings.IndexFunc(raw, unicode.IsSpace) >= 0 {
		return Source{}, fmt.Errorf("repository source contains whitespace or a comment marker and cannot be recorded in construct/deps")
	}
	if raw == "" || strings.HasPrefix(raw, "-") {
		return s, fmt.Errorf("invalid repository source %q", raw)
	}
	if strings.HasPrefix(raw, "github.com/") {
		s.URL = "https://" + strings.TrimSuffix(raw, ".git") + ".git"
	}
	host, p := "", ""
	if strings.HasPrefix(s.URL, "git@github.com:") {
		host = "github.com"
		p = strings.TrimPrefix(s.URL, "git@github.com:")
	} else if strings.Contains(s.URL, "://") {
		u, err := url.Parse(s.URL)
		if err != nil {
			return s, fmt.Errorf("invalid source %q: %w", raw, err)
		}
		if u.User != nil && (u.Scheme != "ssh" || u.User.Username() != "git") {
			return Source{}, fmt.Errorf("repository source must not contain credentials; use Git credential configuration")
		}
		if u.User != nil {
			if _, set := u.User.Password(); set {
				return Source{}, fmt.Errorf("repository source must not contain credentials; use Git credential configuration")
			}
		}
		host = strings.ToLower(u.Hostname())
		p = strings.TrimPrefix(u.Path, "/")
		if u.Scheme == "file" {
			p = u.Path
			s.Identity = "file:" + filepath.Clean(p)
		}
	} else if i := strings.Index(s.URL, ":"); i > 0 {
		host = s.URL[:i]
		p = s.URL[i+1:]
	} else {
		p = s.URL
		s.Identity = "file:" + filepath.Clean(p)
	}
	s.Name = strings.TrimSuffix(path.Base(p), ".git")
	if s.Name == "" || s.Name == "." || s.Name == ".." || s.Name == "/" {
		return s, fmt.Errorf("source %q has no repository name", raw)
	}
	if host == "github.com" {
		parts := strings.Split(strings.TrimSuffix(p, ".git"), "/")
		if len(parts) != 2 || parts[0] == "" || parts[0] == "." || parts[0] == ".." {
			return s, fmt.Errorf("GitHub source %q must name owner/repository", raw)
		}
		s.Identity = strings.ToLower(host + "/" + strings.TrimSuffix(p, ".git"))
	} else if s.Identity == "" {
		s.Identity = host + "/" + strings.TrimSuffix(p, ".git")
	}
	return s, nil
}
