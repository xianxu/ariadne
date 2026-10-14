package issueassets

import (
	"bytes"
	"fmt"
	"reflect"
	"strings"
	"testing"
)

func member(t *testing.T, b []byte) string {
	t.Helper()
	n, e := Name(b, ".PNG")
	if e != nil {
		t.Fatal(e)
	}
	return "assets/000337/" + n
}
func TestParseRefsMarkdown(t *testing.T) {
	p := member(t, []byte("x"))
	body := []byte("![one](" + p + ")\n[x][ref]\n\n[ref]: " + p + "\n\n`[bad](assets/000338/x.png)`\n```\n![bad](assets/000338/x.png)\n```\n")
	got, e := ParseRefs("000337", body)
	if e != nil || !reflect.DeepEqual(got, []string{p}) {
		t.Fatalf("%v %v", got, e)
	}
	for _, p := range []string{"assets/000338/x.png", "assets/000337/../x.png", "assets%2f000337%2fx.png", "./assets/000337/x.png", "assets/000337/x.png?x=1", "assets/000337/x.png#x", "assets/000337/x.png%00", "assets&#47;000337&#47;x.png", "assets&sol;000337&sol;x.png"} {
		_, e := ParseRefs("000337", []byte("[x]("+p+")"))
		if e == nil || !strings.Contains(e.Error(), "link") {
			t.Errorf("unsafe %q: %v", p, e)
		}
	}
}
func TestBuildBundleIntegrityAndDelta(t *testing.T) {
	data := []byte("evidence")
	p := member(t, data)
	body := []byte("[evidence](" + p + ")")
	b, e := BuildBundle("000337", body, map[string][]byte{p: data})
	if e != nil {
		t.Fatal(e)
	}
	data[0] = 'X'
	if string(b.Files[p]) != "evidence" {
		t.Fatal("bundle aliases source")
	}
	for _, files := range []map[string][]byte{nil, {p: []byte("wrong")}, {p: []byte("evidence"), "assets/000337/orphan.png": []byte("x")}} {
		if _, e := BuildBundle("000337", body, files); e == nil {
			t.Fatal("accepted invalid bundle")
		}
	}
	d := BundleDelta(b, Bundle{})
	if !reflect.DeepEqual(d.Deletes, []string{p}) {
		t.Fatal(d)
	}
	d = BundleDelta(Bundle{}, b)
	if !bytes.Equal(d.Writes[p], []byte("evidence")) {
		t.Fatal(d)
	}
}
func TestBundleBudgets(t *testing.T) {
	for _, count := range []int{9, 33} {
		files := map[string][]byte{}
		var body strings.Builder
		for i := 0; i < count; i++ {
			data := []byte(fmt.Sprint(i))
			if count == 9 {
				data = bytes.Repeat([]byte{byte(i)}, MaxFileBytes)
			}
			p := member(t, data)
			files[p] = data
			fmt.Fprintf(&body, "[x](%s)\n", p)
		}
		if _, e := BuildBundle("000337", []byte(body.String()), files); e == nil || !strings.Contains(e.Error(), "remove") {
			t.Fatalf("count %d: %v", count, e)
		}
	}
	if _, e := Name(bytes.Repeat([]byte("x"), MaxFileBytes+1), "png"); e == nil {
		t.Fatal("oversized accepted")
	}
	for _, ext := range []string{"../png", "a/b", "", "tar.gz"} {
		if _, e := Name(nil, ext); e == nil {
			t.Fatalf("accepted %q", ext)
		}
	}
}
func FuzzParseRefs(f *testing.F) {
	f.Add("![x](assets/000337/x.png)")
	f.Add("`[x](assets/../x)`")
	f.Fuzz(func(t *testing.T, s string) {
		refs, e := ParseRefs("000337", []byte(s))
		if e != nil {
			return
		}
		for _, p := range refs {
			if !strings.HasPrefix(p, "assets/000337/") || strings.Contains(p, "..") {
				t.Fatal(p)
			}
		}
	})
}
func FuzzBundleDelta(f *testing.F) {
	f.Add([]byte("old"), []byte("new"))
	f.Fuzz(func(t *testing.T, a, b []byte) {
		if len(a) > MaxFileBytes || len(b) > MaxFileBytes {
			t.Skip()
		}
		pa := member(t, a)
		pb := member(t, b)
		old, e := BuildBundle("000337", []byte("[x]("+pa+")"), map[string][]byte{pa: a})
		if e != nil {
			t.Fatal(e)
		}
		next, e := BuildBundle("000337", []byte("[x]("+pb+")"), map[string][]byte{pb: b})
		if e != nil {
			t.Fatal(e)
		}
		d := BundleDelta(old, next)
		for _, p := range d.Deletes {
			delete(old.Files, p)
		}
		for p, b := range d.Writes {
			old.Files[p] = b
		}
		if !reflect.DeepEqual(old.Files, next.Files) {
			t.Fatal("delta does not converge")
		}
	})
}
