package gitx

import (
	"context"
	"errors"
	"fmt"
	"net/netip"
	"net/url"
	"regexp"
	"strings"
)

// GitNetworkFailure is transport evidence, not a judgment about whether a
// remote write took effect. The original process/context error remains available.
type GitNetworkFailure struct {
	Host  string
	Proxy bool
	cause error
}

func (e *GitNetworkFailure) Error() string {
	target, hint := "", "sdlc net check"
	if e.Host != "" {
		label := "host"
		if e.Proxy {
			label = "proxy"
		}
		target = " (" + label + " " + e.Host + ")"
		hint += " --host " + e.Host
	}
	return "infrastructure: network" + target + "; diagnose with " + hint
}
func (e *GitNetworkFailure) Unwrap() error { return e.cause }

// NormalizeNetworkHost accepts only a DNS name/SSH alias or an IP literal.
// Its errors deliberately omit the untrusted input, which may contain secrets.
func NormalizeNetworkHost(raw string) (string, error) {
	invalid := func() (string, error) { return "", fmt.Errorf("invalid network host: use a DNS name or IP address") }
	if raw == "" || strings.TrimSpace(raw) != raw {
		return invalid()
	}
	host := raw
	if strings.HasPrefix(host, "[") && strings.HasSuffix(host, "]") {
		host = host[1 : len(host)-1]
	}
	if ip, err := netip.ParseAddr(host); err == nil {
		if ip.Zone() != "" {
			return invalid()
		}
		return ip.String(), nil
	}
	if strings.ContainsAny(host, ":/\\@?#[]\r\n\t ") || len(host) > 253 {
		return invalid()
	}
	host = strings.ToLower(host)
	name := strings.TrimSuffix(host, ".")
	numeric := true
	for _, c := range name {
		if (c < '0' || c > '9') && c != '.' {
			numeric = false
		}
	}
	if numeric {
		return invalid()
	}
	for _, label := range strings.Split(name, ".") {
		if len(label) == 0 || len(label) > 63 || label[0] == '-' || label[len(label)-1] == '-' {
			return invalid()
		}
		for _, c := range label {
			if !(c >= 'a' && c <= 'z' || c >= '0' && c <= '9' || c == '-') {
				return invalid()
			}
		}
	}
	return name, nil
}

// RemoteNetworkHost extracts a host from Git's effective URL (get-url), without
// retaining userinfo, paths, queries or helper/local transport arguments.
func RemoteNetworkHost(raw string) (string, error) {
	invalid := func() (string, error) {
		return "", fmt.Errorf("remote has no network host; use --host with a DNS name or IP address")
	}
	if strings.Contains(raw, "://") {
		u, err := url.Parse(raw)
		if err != nil {
			return invalid()
		}
		switch u.Scheme {
		case "http", "https", "ssh", "git":
		default:
			return invalid()
		}
		return NormalizeNetworkHost(u.Hostname())
	}
	if strings.Contains(raw, "::") && !strings.Contains(raw, "]:") {
		return invalid()
	}
	hostPart, path, found := strings.Cut(raw, ":")
	// Bracketed IPv6 contains colons before the scp separator.
	if i := strings.Index(raw, "]:"); i >= 0 {
		hostPart, path, found = raw[:i+1], raw[i+2:], true
	}
	if !found || path == "" || strings.ContainsAny(hostPart, "/\\") {
		return invalid()
	}
	if at := strings.LastIndex(hostPart, "@"); at >= 0 {
		hostPart = hostPart[at+1:]
	}
	return NormalizeNetworkHost(hostPart)
}

var sshResolveNetwork = regexp.MustCompile(`^ssh: Could not resolve hostname ([^\s:]+): .+`)
var sshConnectNetwork = regexp.MustCompile(`^ssh: connect to host (\S+) port [0-9]+: (?:Connection refused|Connection timed out|Operation timed out)\s*$`)
var curlResolveNetwork = regexp.MustCompile(`^Could not resolve (host|proxy):\s*(\S+)\s*$`)
var curlConnectNetwork = regexp.MustCompile(`^Failed to connect to (\S+) port [0-9]+(?: after [0-9]+ ms)?: (?:Couldn't connect to server|Connection refused|Connection timed out)\s*$`)
var curlTimeoutNetwork = regexp.MustCompile(`^(?:Connection timed out after [0-9]+ milliseconds|Operation timed out after [0-9]+ milliseconds(?: with [0-9]+(?: out of [0-9]+)? bytes received)?|Failed to connect to \S+ port [0-9]+ after [0-9]+ ms: Timeout was reached)\s*$`)

func networkVerb(args []string) bool {
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch arg {
		case "-C", "-c", "--git-dir", "--work-tree", "--namespace", "--config-env":
			i++
			continue
		}
		if strings.HasPrefix(arg, "-") {
			continue
		}
		switch arg {
		case "fetch", "push", "pull", "clone", "ls-remote":
			return true
		}
		return false
	}
	return false
}

// ClassifyNetworkFailure is pure and immediate: only Git transport diagnostics
// on remote-capable commands are evidence. No probes, retries or config reads.
func ClassifyNetworkFailure(args []string, diagnostic []byte, cause error) error {
	if cause == nil {
		return nil
	}
	if errors.Is(cause, context.Canceled) {
		return cause
	}
	var existing *GitNetworkFailure
	if errors.As(cause, &existing) || !networkVerb(args) {
		return cause
	}
	// Transport signatures must begin their own line; hook/remote prose does not
	// become a network failure merely by mentioning a resolver or timeout.
	for _, line := range strings.Split(string(diagnostic), "\n") {
		line = strings.TrimSuffix(line, "\r")
		host, proxy, recognized := "", false, false
		if m := sshResolveNetwork.FindStringSubmatch(line); m != nil {
			host, recognized = m[1], true
		} else if m := sshConnectNetwork.FindStringSubmatch(line); m != nil {
			host, recognized = m[1], true
		} else if strings.HasPrefix(line, "fatal: unable to access '") {
			rest := strings.TrimPrefix(line, "fatal: unable to access '")
			// Git encloses the URL in quotes; the last delimiter also permits an
			// apostrophe in URL userinfo without ever rendering that userinfo.
			if i := strings.LastIndex(rest, "': "); i >= 0 {
				remote, detail := rest[:i], rest[i+3:]
				if strings.HasPrefix(remote, "http://") || strings.HasPrefix(remote, "https://") {
					if m := curlResolveNetwork.FindStringSubmatch(detail); m != nil {
						host, proxy, recognized = m[2], m[1] == "proxy", true
					} else if m := curlConnectNetwork.FindStringSubmatch(detail); m != nil {
						host, recognized = m[1], true
					} else if curlTimeoutNetwork.MatchString(detail) {
						host, _ = RemoteNetworkHost(remote)
						recognized = true
					}
				}
			}
		}
		if recognized {
			safe, _ := NormalizeNetworkHost(host)
			return &GitNetworkFailure{Host: safe, Proxy: proxy, cause: cause}
		}
	}
	return cause
}

// networkDiagnostic suppresses redundant raw transport prose when the safe
// typed error already describes it. Ordinary Git failures retain diagnostics.
func networkDiagnostic(err error, diagnostic []byte) []byte {
	var network *GitNetworkFailure
	if errors.As(err, &network) {
		return nil
	}
	return diagnostic
}
