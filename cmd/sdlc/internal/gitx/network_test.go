package gitx

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestNetworkClassification(t *testing.T) {
	cause := errors.New("exit status 128")
	for _, tc := range []struct{ diagnostic, host string }{
		{"fatal: unable to access 'https://user:SECRET@git.example/repo?SECRET': Could not resolve host: git.example", "git.example"},
		{"ssh: Could not resolve hostname my-alias: nodename nor servname provided, or not known", "my-alias"},
		{"ssh: connect to host git.example port 22: Connection refused", "git.example"},
		{"fatal: unable to access 'https://git.example/repo': Failed to connect to git.example port 443 after 5 ms: Couldn't connect to server", "git.example"},
		{"fatal: unable to access 'https://git.example/repo': Connection timed out after 1000 milliseconds", "git.example"},
		{"fatal: unable to access 'https://git.example/repo': Operation timed out after 1000 milliseconds with 0 bytes received", "git.example"},
		{"fatal: unable to access 'https://git.example/repo': Could not resolve proxy: proxy.example", "proxy.example"},
		{"ssh: connect to host 2001:db8::1 port 22: Connection timed out", "2001:db8::1"},
		{"fatal: unable to access 'https://SECRET@bad_host/path': Could not resolve host: bad_host", ""},
	} {
		for _, verb := range []string{"fetch", "push", "pull", "clone", "ls-remote"} {
			err := ClassifyNetworkFailure([]string{"-C", "repo", verb, "origin"}, []byte(tc.diagnostic), cause)
			var n *GitNetworkFailure
			if !errors.As(err, &n) || !errors.Is(err, cause) {
				t.Fatalf("%s: missing typed error/cause: %v", tc.diagnostic, err)
			}
			hint := "sdlc net check"
			if tc.host != "" {
				hint += " --host " + tc.host
			}
			if !strings.HasPrefix(err.Error(), "infrastructure: network") || !strings.Contains(err.Error(), hint) || strings.Contains(err.Error(), "SECRET") {
				t.Fatalf("unsafe/incomplete report: %v", err)
			}
			if tc.host == "" && strings.Contains(err.Error(), "--host") {
				t.Fatalf("unsafe host: %v", err)
			}
			if ClassifyNetworkFailure([]string{verb}, []byte(tc.diagnostic), err) != err {
				t.Fatal("not idempotent")
			}
		}
	}
}

func TestNetworkClassificationNegative(t *testing.T) {
	cause := errors.New("exit 1")
	for _, d := range []string{
		"fatal: Authentication failed for 'https://git.example/repo'", "Permission denied (publickey).", "! [rejected] main -> main (stale info)",
		"remote: error: Could not resolve host: hook.example", "hook: Connection timed out", "fatal: cannot run hook: Connection refused", "fatal: could not read Username", "Operation timed out", "fatal: unable to access local file: Connection timed out",
	} {
		if got := ClassifyNetworkFailure([]string{"push"}, []byte(d), cause); got != cause {
			t.Errorf("classified %q: %v", d, got)
		}
	}
	for _, verb := range []string{"status", "commit", "log", "config"} {
		if got := ClassifyNetworkFailure([]string{verb}, []byte("ssh: Could not resolve hostname git.example: unknown"), cause); got != cause {
			t.Errorf("classified local %s", verb)
		}
	}
	for _, cause := range []error{context.Canceled, context.DeadlineExceeded, nil} {
		if got := ClassifyNetworkFailure([]string{"fetch"}, nil, cause); got != cause {
			t.Errorf("classified bare context/nil: %v", got)
		}
	}
}

func TestNetworkHosts(t *testing.T) {
	for _, host := range []string{"github.com", "my-alias", "192.0.2.1", "2001:db8::1", "[2001:db8::1]"} {
		if _, err := NormalizeNetworkHost(host); err != nil {
			t.Errorf("%q: %v", host, err)
		}
	}
	for _, host := range []string{"", "-flag", "user@host", "https://host", "host:443", "host/path", "host?SECRET", "host\nSECRET", "host_name", "127.0.0.999"} {
		if got, err := NormalizeNetworkHost(host); err == nil {
			t.Errorf("accepted %q as %q", host, got)
		}
	}
	for _, tc := range []struct{ raw, host string }{
		{"https://user:SECRET@git.example:443/repo?q=SECRET#SECRET", "git.example"}, {"ssh://user@git.example:22/repo", "git.example"}, {"git@my-alias:repo.git", "my-alias"}, {"git@[2001:db8::1]:repo", "2001:db8::1"}, {"git://192.0.2.1/repo", "192.0.2.1"},
	} {
		if got, err := RemoteNetworkHost(tc.raw); err != nil || got != tc.host {
			t.Errorf("%q = %q, %v", tc.raw, got, err)
		}
	}
	for _, raw := range []string{"/tmp/repo", "../repo", "./host:repo", "file:///repo", "ext::SECRET", "helper::https://host/repo", "host", "https://-flag/repo"} {
		if got, err := RemoteNetworkHost(raw); err == nil {
			t.Errorf("accepted %q as %q", raw, got)
		}
	}
}

func FuzzNetworkSafety(f *testing.F) {
	for _, s := range []string{"https://user:SECRET@example.org/repo", "ssh: Could not resolve hostname example.org: unknown", "host\nSECRET"} {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, raw string) {
		if host, err := RemoteNetworkHost(raw); err == nil {
			if _, err := NormalizeNetworkHost(host); err != nil {
				t.Fatalf("invalid returned host %q", host)
			}
		}
		cause := errors.New("cause credential SECRET/hidden")
		err := ClassifyNetworkFailure([]string{"fetch"}, []byte(raw), cause)
		if err != cause && (strings.Contains(err.Error(), "SECRET/hidden") || strings.ContainsAny(err.Error(), "\n\r")) {
			t.Fatalf("unsafe output %q", err)
		}
		if got := ClassifyNetworkFailure([]string{"commit"}, []byte(raw), cause); got != cause {
			t.Fatal("local command classified")
		}
	})
}

func FuzzNetworkRemoteCredentials(f *testing.F) {
	for _, host := range []string{"example.org", "[2001:db8::1]", "evil@bad_host", "host\nSECRET"} {
		f.Add(host)
	}
	f.Fuzz(func(t *testing.T, raw string) {
		remote := "https://user:SECRET%2Fcredential@" + raw + "/repo?q=SECRET%2Fcredential#SECRET%2Fcredential"
		host, err := RemoteNetworkHost(remote)
		if err == nil {
			if strings.Contains(host, "secret%2fcredential") || strings.ContainsAny(host, "/@?#\r\n") {
				t.Fatalf("credential/config escaped: %q", host)
			}
			if _, err := NormalizeNetworkHost(host); err != nil {
				t.Fatal(err)
			}
		}
		d := []byte("fatal: unable to access '" + remote + "': Connection timed out after 1000 milliseconds")
		classified := ClassifyNetworkFailure([]string{"fetch"}, d, errors.New("SECRET/credential"))
		var n *GitNetworkFailure
		if errors.As(classified, &n) && strings.Contains(classified.Error(), "SECRET") {
			t.Fatalf("credential escaped: %v", classified)
		}
	})
}

func TestNetworkCancellationAndCanonicalHost(t *testing.T) {
	cause := errors.Join(context.Canceled, errors.New("exit 128"))
	if got := ClassifyNetworkFailure([]string{"fetch"}, []byte("ssh: Could not resolve hostname git.example: unknown"), cause); got != cause {
		t.Fatalf("classified cancellation: %v", got)
	}
	if host, err := NormalizeNetworkHost("GitHub.COM."); err != nil || host != "github.com" {
		t.Fatalf("canonical host=%q,%v", host, err)
	}
}
