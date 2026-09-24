package acquire

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
)

func TestRestoreLayerLimitBeforeExpansion(t *testing.T) {
	base := t.TempDir()
	root := filepath.Join(base, "root")
	put(t, filepath.Join(root, "construct/deps"), "substrate ../a\nsubstrate ../b\n")
	put(t, filepath.Join(base, "a/construct/base.manifest"), "")
	// b is absent: the limit must refuse before attempting inspection/acquisition.
	_, err := (Client{MaxLayers: 2}).Restore(context.Background(), root, true)
	if err == nil || !strings.Contains(err.Error(), "layer limit 2") {
		t.Fatalf("got %v", err)
	}
	put(t, filepath.Join(base, "b/construct/base.manifest"), "")
	result, err := (Client{}).Restore(context.Background(), root, true)
	if err != nil || len(result.Layers) != 3 {
		t.Fatalf("default: %+v %v", result, err)
	}
}

func TestRestoreBoundedDeclarations(t *testing.T) {
	for _, kind := range []string{"oversize", "symlink", "dangling-symlink", "directory", "fifo"} {
		t.Run(kind, func(t *testing.T) {
			root := t.TempDir()
			path := filepath.Join(root, "construct/deps")
			if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
				t.Fatal(err)
			}
			switch kind {
			case "oversize":
				put(t, path, strings.Repeat("#", 65))
			case "symlink":
				put(t, filepath.Join(root, "elsewhere"), "")
				if err := os.Symlink("../elsewhere", path); err != nil {
					t.Fatal(err)
				}
			case "dangling-symlink":
				if err := os.Symlink("missing", path); err != nil {
					t.Fatal(err)
				}
			case "directory":
				if err := os.Mkdir(path, 0755); err != nil {
					t.Fatal(err)
				}
			case "fifo":
				if err := syscall.Mkfifo(path, 0600); err != nil {
					t.Fatal(err)
				}
			}
			_, err := (Client{MaxDeclarationBytes: 64}).Restore(context.Background(), root, true)
			if err == nil {
				t.Fatal("accepted nonordinary or oversized declaration")
			}
		})
	}
}

func TestRestoreLayerLimitDeduplicatesQueue(t *testing.T) {
	base := t.TempDir()
	root := filepath.Join(base, "root")
	put(t, filepath.Join(root, "construct/deps"), "substrate ../a\nsubstrate ../b\n")
	for _, name := range []string{"a", "b"} {
		put(t, filepath.Join(base, name, "construct/base.manifest"), "")
		put(t, filepath.Join(base, name, "construct/deps"), "substrate ../leaf\n")
	}
	put(t, filepath.Join(base, "leaf/construct/base.manifest"), "")
	result, err := (Client{MaxLayers: 4}).Restore(context.Background(), root, true)
	if err != nil || len(result.Layers) != 4 {
		t.Fatalf("got %+v %v", result, err)
	}
}

func TestAcquireDiagnosticsHideSourceCredentials(t *testing.T) {
	const secret = "SENTINEL_PRIVATE_TOKEN"
	declared := "https://example.test/layer.git?token=" + secret
	for _, kind := range []string{"origin-failure", "origin-missing", "origin-conflict", "origin-invalid"} {
		t.Run(kind, func(t *testing.T) {
			dest := canonical(t.TempDir())
			source, err := ResolveSource(declared, dest)
			if err != nil {
				t.Fatal(err)
			}
			g := &stateGit{run: func(_ int, _ string, args []string) (string, error) {
				if args[0] == "rev-parse" {
					return dest, nil
				}
				if args[len(args)-1] == "--list" {
					if kind == "origin-failure" {
						return "", exitFailure(128)
					}
					return "", nil
				}
				switch kind {
				case "origin-missing":
					return "", exitFailure(1)
				case "origin-invalid":
					return "https://example.test:bad" + secret + "/layer.git", nil
				default:
					return "https://example.test/other.git?token=" + secret, nil
				}
			}}
			err = (Client{Git: g}).checkExisting(context.Background(), dest, source, false)
			if err == nil || strings.Contains(err.Error(), secret) || !strings.Contains(err.Error(), dest) {
				t.Fatalf("unsafe or unhelpful error: %v", err)
			}
			if !strings.Contains(err.Error(), "check") && !strings.Contains(err.Error(), "configure") && !strings.Contains(err.Error(), "fix") {
				t.Fatalf("missing recovery instruction: %v", err)
			}
		})
	}
	for _, kind := range []string{"missing", "conflict", "invalid-source"} {
		t.Run(kind, func(t *testing.T) {
			root := canonical(t.TempDir())
			dest := filepath.Join(root, "peer")
			rows := "substrate ./peer " + declared + "\n"
			if kind == "conflict" {
				rows += "substrate ./peer https://example.test/other.git?token=" + secret + "\n"
			}
			if kind == "invalid-source" {
				rows = "substrate ./peer https://example.test:bad" + secret + "/layer.git\n"
			}
			put(t, filepath.Join(root, "construct/deps"), rows)
			result, err := (Client{}).Restore(context.Background(), root, true)
			if err == nil || strings.Contains(err.Error(), secret) || strings.Contains(strings.Join(result.Missing, " "), secret) {
				t.Fatalf("unsafe or missing failure: %+v %v", result, err)
			}
			if kind != "invalid-source" && !strings.Contains(err.Error(), dest) {
				t.Fatalf("missing destination: %v", err)
			}
			if !strings.Contains(err.Error(), "check") && !strings.Contains(err.Error(), "restore") && !strings.Contains(err.Error(), "fix") {
				t.Fatalf("missing recovery instruction: %v", err)
			}
		})
	}
}
