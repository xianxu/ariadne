package gitx

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// PublicationTarget pins effective transport configuration, not a checkout
// identity. Repository is stable across clones; private recovery also binds its
// local worktree separately. Callers re-resolve before effects after a pause.
type PublicationTarget struct {
	Remote, FetchURL, PushURL, Repository string
}

var publicationRemoteName = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._/-]*$`)
var githubRepositoryPath = regexp.MustCompile(`^[A-Za-z0-9_.-]+/[A-Za-z0-9_.-]+$`)

// ResolvePublicationTarget uses the specified resting branch's explicit
// upstream. An ordinary feature worktree supplies main; a slot supplies its
// own resting branch. It never guesses origin, performs network IO, or mutates
// refs/index/worktree. Multiple destinations and transport ambiguity refuse.
func ResolvePublicationTarget(ctx context.Context, root, restingBranch string) (PublicationTarget, error) {
	if ctx == nil || root == "" || restingBranch == "" {
		return PublicationTarget{}, errors.New("publication target requires context, repository and resting branch")
	}
	query := func(args ...string) (string, error) {
		out, diag, err := runGitBoundedInputContext(ctx, root, nil, nil, diagnosticOutputLimit, args...)
		if err != nil {
			return "", fmt.Errorf("resolve publication target: git %s: %w: %s", args[0], err, diag)
		}
		value := strings.TrimSuffix(string(out), "\n")
		if value == "" || strings.ContainsAny(value, "\r\n\x00") {
			return "", errors.New("publication target requires exactly one configured value")
		}
		return value, nil
	}
	// check-ref-format prints nothing on success, unlike the configuration reads.
	if _, diag, err := runGitBoundedInputContext(ctx, root, nil, nil, diagnosticOutputLimit, "check-ref-format", "refs/heads/"+restingBranch); err != nil {
		return PublicationTarget{}, fmt.Errorf("invalid resting branch: %w: %s", err, diag)
	}
	var t PublicationTarget
	var err error
	t.Remote, err = query("config", "--get-all", "branch."+restingBranch+".remote")
	if err != nil {
		return t, fmt.Errorf("configure %s to track one named remote/main: %w", restingBranch, err)
	}
	if !publicationRemoteName.MatchString(t.Remote) {
		return t, errors.New("publication upstream must use one named remote")
	}
	if _, diag, err := runGitBoundedInputContext(ctx, root, nil, nil, diagnosticOutputLimit, "check-ref-format", "refs/remotes/"+t.Remote+"/main"); err != nil {
		return t, fmt.Errorf("invalid publication remote: %w: %s", err, diag)
	}
	base, err := query("config", "--get-all", "branch."+restingBranch+".merge")
	if err != nil || base != "refs/heads/main" {
		return t, errors.New("publication upstream must be exactly remote/main")
	}
	if t.FetchURL, err = query("remote", "get-url", "--all", t.Remote); err != nil {
		return t, err
	}
	if t.PushURL, err = query("remote", "get-url", "--push", "--all", t.Remote); err != nil {
		return t, err
	}
	if t.Repository, err = publicationRepository(root, t.FetchURL); err != nil {
		return t, err
	}
	push, err := publicationRepository(root, t.PushURL)
	if err != nil {
		return t, err
	}
	if t.Repository != push {
		return t, errors.New("fetch and push must target the same repository")
	}
	return t, nil
}

// PublicationRepository is the stable identity of a remote URL as seen from
// root (GitHub HTTPS and SSH spell one repository; local paths canonicalize).
func PublicationRepository(root, raw string) (string, error) { return publicationRepository(root, raw) }

func publicationRepository(root, raw string) (string, error) {
	// GitHub's owner/repo identity is independent of HTTPS versus SSH transport
	// and is case-insensitive. Do not apply that assumption to arbitrary hosts.
	githubPath := ""
	if strings.HasPrefix(raw, "git@github.com:") {
		githubPath = strings.TrimPrefix(raw, "git@github.com:")
	} else if u, err := url.Parse(raw); err == nil && u.Host == "github.com" {
		if u.RawQuery != "" || u.Fragment != "" || (u.Scheme != "https" && u.Scheme != "ssh") ||
			(u.User != nil && (u.Scheme != "ssh" || u.User.String() != "git")) {
			return "", errors.New("unsupported GitHub publication transport")
		}
		githubPath = strings.TrimPrefix(u.Path, "/")
	}
	if githubPath != "" {
		githubPath = strings.TrimSuffix(githubPath, ".git")
		if !githubRepositoryPath.MatchString(githubPath) || strings.Contains(githubPath, "..") {
			return "", errors.New("invalid GitHub publication repository")
		}
		return "github.com/" + strings.ToLower(githubPath), nil
	}
	// Local bare remotes support hermetic conformance and offline repositories.
	// Canonicalize symlinks so two checkout-relative URLs identify the same repo.
	local := raw
	if strings.HasPrefix(raw, "file://") {
		u, err := url.Parse(raw)
		if err != nil || u.Host != "" || u.User != nil || u.RawQuery != "" || u.Fragment != "" {
			return "", errors.New("unsupported file publication URL")
		}
		local = u.Path
	} else if strings.Contains(raw, ":") {
		// Other transports retain their exact effective identity; no speculative
		// host/path normalization that could equate different repositories.
		return "remote:" + raw, nil
	}
	if !filepath.IsAbs(local) {
		local = filepath.Join(root, local)
	}
	local, err := filepath.Abs(local)
	if err != nil {
		return "", err
	}
	// Identity belongs to the configured location, not its reachability: an
	// unreachable (absent) repository keeps its cleaned path, so an offline read
	// fails at fetch, where it is labelled, instead of here.
	resolved, err := filepath.EvalSymlinks(local)
	switch {
	case err == nil:
		local = resolved
	case errors.Is(err, os.ErrNotExist):
		local = filepath.Clean(local)
	default:
		return "", fmt.Errorf("resolve local publication repository: %w", err)
	}
	return "file:" + local, nil
}
