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
		return Source{}, fmt.Errorf("repository source must be nonempty and must not start with a dash")
	}
	if strings.HasPrefix(raw, "github.com/") {
		s.URL = "https://" + strings.TrimSuffix(raw, ".git") + ".git"
	}
	host, p := "", ""
	githubEndpoint := false
	if strings.HasPrefix(s.URL, "git@github.com:") {
		host = "github.com"
		githubEndpoint = true
		p = strings.TrimPrefix(s.URL, "git@github.com:")
	} else if strings.Contains(s.URL, "://") {
		u, err := url.Parse(s.URL)
		if err != nil {
			// URL parse errors include the original input, possibly credentials.
			return Source{}, fmt.Errorf("invalid repository URL: check its host, port, and escaping")
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
		// Equivalence is explicit for standard GitHub transports only. Other
		// endpoints retain scheme, authority (including port), path and query.
		githubEndpoint = host == "github.com" && u.RawQuery == "" &&
			((u.Scheme == "https" && (u.Port() == "" || u.Port() == "443")) ||
				(u.Scheme == "ssh" && u.User != nil && u.User.Username() == "git" && (u.Port() == "" || u.Port() == "22")))
		s.Identity = "uri:" + s.URL
		p = strings.TrimPrefix(u.Path, "/")
		if u.Scheme == "file" {
			if (u.Host != "" && u.Host != "localhost") || u.RawQuery != "" {
				return Source{}, fmt.Errorf("file source must be a local path without a query")
			}
			p = u.Path
			s.Identity = "file:" + filepath.Clean(p)
		}
	} else if i := strings.Index(s.URL, ":"); i > 0 {
		host = s.URL[:i]
		p = s.URL[i+1:]
		s.Identity = "scp:" + s.URL
	} else {
		p = s.URL
		s.Identity = "file:" + filepath.Clean(p)
	}
	s.Name = strings.TrimSuffix(path.Base(p), ".git")
	if s.Name == "" || s.Name == "." || s.Name == ".." || s.Name == "/" {
		return Source{}, fmt.Errorf("repository source has no repository name")
	}
	if githubEndpoint {
		parts := strings.Split(strings.TrimSuffix(p, ".git"), "/")
		if len(parts) != 2 || parts[0] == "" || parts[0] == "." || parts[0] == ".." {
			return Source{}, fmt.Errorf("GitHub source must name owner/repository")
		}
		s.Identity = strings.ToLower(host + "/" + strings.TrimSuffix(p, ".git"))
	}
	return s, nil
}
