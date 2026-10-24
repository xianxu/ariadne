package main

import (
	"archive/tar"
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
)

func TestBenchTemplateRoundTrip(t *testing.T) {
	files := []gitx.TreeFile{{Path: "a", Mode: "100644", Content: []byte{0, 1, 255}}, {Path: "bin/run", Mode: "100755", Content: []byte("#!/bin/sh\n")}, {Path: "link", Mode: "120000", Content: []byte("bin/run")}}
	p := filepath.Join(t.TempDir(), "template.tar")
	if err := writeBenchTemplate(context.Background(), p, files); err != nil {
		t.Fatal(err)
	}
	got, err := readBenchTemplate(context.Background(), p)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, files) {
		t.Fatalf("roundtrip differs: %#v", got)
	}
	before, _ := os.ReadFile(p)
	if err := writeBenchTemplate(context.Background(), p, files); err == nil {
		t.Fatal("overwrote immutable template")
	}
	q := filepath.Join(t.TempDir(), "template.tar")
	reverse := []gitx.TreeFile{files[2], files[1], files[0]}
	if err := writeBenchTemplate(context.Background(), q, reverse); err != nil {
		t.Fatal(err)
	}
	after, _ := os.ReadFile(q)
	if !bytes.Equal(before, after) {
		t.Fatal("archive is not deterministic")
	}
}

func TestBenchTemplateRejectsMalformed(t *testing.T) {
	for _, tc := range []struct {
		name    string
		headers []*tar.Header
		body    string
	}{
		{"traversal", []*tar.Header{{Name: "../bad", Mode: 0644, Size: 1}}, "x"},
		{"absolute", []*tar.Header{{Name: "/bad", Mode: 0644, Size: 1}}, "x"},
		{"duplicate", []*tar.Header{{Name: "a", Mode: 0644}, {Name: "a", Mode: 0644}}, ""},
		{"directory", []*tar.Header{{Name: "dir", Typeflag: tar.TypeDir, Mode: 0755}}, ""},
		{"escaping-link", []*tar.Header{{Name: "link", Typeflag: tar.TypeSymlink, Linkname: "../escape", Mode: 0777}}, ""},
		{"symlink-ancestor", []*tar.Header{{Name: "link", Typeflag: tar.TypeSymlink, Linkname: "inside", Mode: 0777}, {Name: "link/file", Mode: 0644}}, ""},
		{"mode", []*tar.Header{{Name: "a", Mode: 0600}}, ""},
		{"oversized", []*tar.Header{{Name: "a", Mode: 0644, Size: (1 << 20) + 1}}, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var b bytes.Buffer
			w := tar.NewWriter(&b)
			for _, h := range tc.headers {
				if err := w.WriteHeader(h); err != nil {
					t.Fatal(err)
				}
				if tc.body != "" {
					_, _ = w.Write([]byte(tc.body))
				}
			}
			_ = w.Close()
			p := filepath.Join(t.TempDir(), "bad.tar")
			if err := os.WriteFile(p, b.Bytes(), 0600); err != nil {
				t.Fatal(err)
			}
			if _, err := readBenchTemplate(context.Background(), p); err == nil {
				t.Fatal("accepted unsafe archive")
			}
		})
	}
	p := filepath.Join(t.TempDir(), "truncated.tar")
	_ = os.WriteFile(p, []byte("not a tar"), 0600)
	if _, err := readBenchTemplate(context.Background(), p); err == nil {
		t.Fatal("accepted truncation")
	}
}

func TestBenchTemplateBoundsAndCancellation(t *testing.T) {
	for _, files := range [][]gitx.TreeFile{
		{{Path: "large", Mode: "100644", Content: []byte(strings.Repeat("x", (1<<20)+1))}},
		make([]gitx.TreeFile, 10001),
	} {
		p := filepath.Join(t.TempDir(), "template.tar")
		if err := writeBenchTemplate(context.Background(), p, files); err == nil {
			t.Fatal("accepted over limit")
		}
		if _, err := os.Lstat(p); !os.IsNotExist(err) {
			t.Fatal("failed write published output")
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	p := filepath.Join(t.TempDir(), "template.tar")
	if err := writeBenchTemplate(ctx, p, nil); err == nil {
		t.Fatal("ignored cancellation")
	}
	if _, err := os.Lstat(p); !os.IsNotExist(err) {
		t.Fatal("cancel published output")
	}
}

func TestBenchTemplateRefusesSymlinksAndAppendedData(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "target")
	if err := os.WriteFile(target, []byte("unchanged"), 0600); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(dir, "link")
	if err := os.Symlink(target, link); err != nil {
		t.Fatal(err)
	}
	if err := writeBenchTemplate(context.Background(), link, nil); err == nil {
		t.Fatal("followed output symlink")
	}
	if _, err := readBenchTemplate(context.Background(), link); err == nil {
		t.Fatal("followed input symlink")
	}
	got, _ := os.ReadFile(target)
	if string(got) != "unchanged" {
		t.Fatal("modified symlink target")
	}
	p := filepath.Join(dir, "good.tar")
	if err := writeBenchTemplate(context.Background(), p, nil); err != nil {
		t.Fatal(err)
	}
	raw, _ := os.ReadFile(p)
	raw = append(raw, append([]byte("hidden"), make([]byte, 2048-6)...)...)
	if err := os.WriteFile(p, raw, 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := readBenchTemplate(context.Background(), p); err == nil {
		t.Fatal("ignored appended payload")
	}
}

func TestBenchTemplateTreeAndDecodedEntryBounds(t *testing.T) {
	contents := make([]byte, 1<<20)
	files := make([]gitx.TreeFile, 65)
	for i := range files {
		files[i] = gitx.TreeFile{Path: fmt.Sprintf("file-%d", i), Mode: "100644", Content: contents}
	}
	p := filepath.Join(t.TempDir(), "template.tar")
	if err := writeBenchTemplate(context.Background(), p, files); err == nil {
		t.Fatal("accepted tree above 64 MiB")
	}
	var raw bytes.Buffer
	w := tar.NewWriter(&raw)
	for i := 0; i < 10001; i++ {
		if err := w.WriteHeader(&tar.Header{Name: fmt.Sprintf("file-%d", i), Mode: 0644}); err != nil {
			t.Fatal(err)
		}
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, raw.Bytes(), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := readBenchTemplate(context.Background(), p); err == nil {
		t.Fatal("accepted more than 10000 decoded entries")
	}
}
