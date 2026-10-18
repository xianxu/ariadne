package golint

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
	"testing"
)

func TestWireIndexCost(t *testing.T) {
	for _, n := range []int{8, 16} {
		files := make(map[string]*ast.File, n)
		for i := 0; i < n; i++ {
			name := fmt.Sprintf("p%d/message.go", i)
			f, err := parser.ParseFile(token.NewFileSet(), name,
				"package p\n// wire: peer version is independent.\ntype Message struct{}\n", parser.ParseComments)
			if err != nil {
				t.Fatal(err)
			}
			files[name] = f
		}
		index := newWireTypeIndex(files)
		if index.fileVisits != n || index.typeVisits != n || len(index.marked) != n {
			t.Errorf("%d files: visits=(%d,%d), packages=%d; want (%d,%d,%d)",
				n, index.fileVisits, index.typeVisits, len(index.marked), n, n, n)
		}
	}
}

func TestWireAnnotationRequiresReasonAndTypeDoc(t *testing.T) {
	src := `package p
// wire:
type EmptyReason struct{}
// wire: independently versioned peer.
type Marked struct{}
// wire: a group-level note does not mark both types.
type (
	First struct{}
	Second struct{}
)
type Plain struct{} // wire: trailing text is not a type doc
`
	f, err := parser.ParseFile(token.NewFileSet(), "p.go", src, parser.ParseComments)
	if err != nil {
		t.Fatal(err)
	}
	marked, _ := markedWireTypes(f)
	if marked["Marked"] == nil || len(marked) != 1 {
		t.Errorf("marked declarations = %v; want only Marked", marked)
	}
}

func TestWireStrictDecodeSameFile(t *testing.T) {
	src := `package p
import (
	"encoding/json"
	"io"
)

// wire: peer messages may gain fields before this binary upgrades.
type PeerMessage struct { ID string }

func read(r io.Reader) {
	dec := json.NewDecoder(r)
	dec.DisallowUnknownFields()
	var message PeerMessage
	_ = dec.Decode(&message)
}`
	got, err := Check("peer.go", []byte(src), Catalog)
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range got {
		if f.Rule == "wire-strict-decode" && f.Line == 14 && strings.Contains(f.Message, "pair#594") {
			return
		}
	}
	t.Fatalf("strict decode of an annotated peer message at line 14 was not reported: %v", got)
}

func TestWireStrictDecodeReasonedSuppression(t *testing.T) {
	src := `package p
import ("encoding/json"; "io")
// wire: an older peer can send this message.
type Peer struct{}
func read(r io.Reader) {
	d := json.NewDecoder(r)
	d.DisallowUnknownFields()
	var m Peer
	//lint:allow wire-strict-decode — this fixture verifies the lint escape
	_ = d.Decode(&m)
}`
	got, err := Check("p.go", []byte(src), Catalog)
	if err != nil || len(got) != 0 {
		t.Fatalf("reasoned suppression should cover the marked decode: %v %v", got, err)
	}
}

func TestWireStrictDecodeAcrossFilesAndImport(t *testing.T) {
	sources := map[string][]byte{
		"wire/message.go": []byte(`package wire
// wire: this peer payload may gain fields between releases.
type Message struct { ID string }`),
		"wire/decode.go": []byte(`package wire
import ("encoding/json"; "io")
func read(r io.Reader) error {
	d := json.NewDecoder(r)
	d.DisallowUnknownFields()
	var m Message
	return d.Decode(&m)
}`),
		"consumer/decode.go": []byte(`package consumer
import ("encoding/json"; "io"; peer "example.test/module/wire")
func read(r io.Reader) error {
	d := json.NewDecoder(r)
	d.DisallowUnknownFields()
	var m peer.Message
	return d.Decode(&m)
}`),
	}
	got, err := CheckSources("example.test/module", sources, Catalog)
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"wire/decode.go", "consumer/decode.go"} {
		if len(got[name]) != 1 || got[name][0].Rule != "wire-strict-decode" {
			t.Errorf("%s findings = %v, want one wire-strict-decode", name, got[name])
		}
	}
}

func TestWireAnnotationDoesNotCrossTestPackage(t *testing.T) {
	sources := map[string][]byte{
		"p/peer_test.go": []byte(`package p_test
// wire: a test peer may gain fields.
type Peer struct{}`),
		"p/decode.go": []byte(`package p
import ("encoding/json"; "io")
type Peer struct{}
func read(r io.Reader) error {
	d := json.NewDecoder(r)
	d.DisallowUnknownFields()
	var m Peer
	return d.Decode(&m)
}`),
	}
	got, err := CheckSources("example.test/module", sources, Catalog)
	if err != nil {
		t.Fatal(err)
	}
	if len(got["p/decode.go"]) != 0 {
		t.Errorf("test package marked a production type: %v", got["p/decode.go"])
	}
}

func TestWireImportUsesDeclaredPackageName(t *testing.T) {
	sources := map[string][]byte{
		"messages/v2/message.go": []byte(`package wire
// wire: peer revisions are independent.
type Message struct{}`),
		"reader/decode.go": []byte(`package reader
import ("encoding/json"; "io"; "example.test/module/messages/v2")
func read(r io.Reader) error {
	d := json.NewDecoder(r)
	d.DisallowUnknownFields()
	var m wire.Message
	return d.Decode(&m)
}`),
	}
	got, err := CheckSources("example.test/module", sources, Catalog)
	if err != nil {
		t.Fatal(err)
	}
	if len(got["reader/decode.go"]) != 1 || got["reader/decode.go"][0].Rule != "wire-strict-decode" {
		t.Errorf("declared package name was not resolved: %v", got["reader/decode.go"])
	}
}

func TestWireGeneratedTypeIsNotIndexed(t *testing.T) {
	sources := map[string][]byte{
		"p/generated.go": []byte(`// Code generated by fixture. DO NOT EDIT.
package p
// wire: generated producer shape.
type Peer struct{}`),
		"p/decode.go": []byte(`package p
import ("encoding/json"; "io")
func read(r io.Reader) error {
	d := json.NewDecoder(r)
	d.DisallowUnknownFields()
	var m Peer
	return d.Decode(&m)
}`),
	}
	got, err := CheckSources("example.test/module", sources, Catalog)
	if err != nil {
		t.Fatal(err)
	}
	if len(got["p/decode.go"]) != 0 {
		t.Errorf("generated declaration marked a source type: %v", got["p/decode.go"])
	}
}

func TestWireStrictDecodeFollowsJSONImportAlias(t *testing.T) {
	src := `package p
import (j "encoding/json"; "io")
// wire: peer evolves separately.
type Peer struct{}
func read(r io.Reader) error {
	d := j.NewDecoder(r)
	d.DisallowUnknownFields()
	var m Peer
	return d.Decode(&m)
}`
	got, err := Check("p.go", []byte(src), Catalog)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Rule != "wire-strict-decode" {
		t.Errorf("aliased encoding/json import: findings %v", got)
	}
}

func TestWireStrictDecodeDoesNotGuessFlowOrTarget(t *testing.T) {
	for _, tc := range []struct {
		name, body string
		want       bool
	}{
		{"direct", `d := json.NewDecoder(r); d.DisallowUnknownFields(); var m Peer; _ = d.Decode(&m)`, true},
		{"var decoder", `var d = json.NewDecoder(r); d.DisallowUnknownFields(); var m Peer; _ = d.Decode(&m)`, true},
		{"composite target", `d := json.NewDecoder(r); d.DisallowUnknownFields(); _ = d.Decode(&Peer{})`, true},
		{"inferred target", `d := json.NewDecoder(r); d.DisallowUnknownFields(); m := Peer{}; _ = d.Decode(&m)`, true},
		{"new target", `d := json.NewDecoder(r); d.DisallowUnknownFields(); m := new(Peer); _ = d.Decode(m)`, true},
		{"if init decode", `d := json.NewDecoder(r); d.DisallowUnknownFields(); var m Peer; if err := d.Decode(&m); err != nil { return }`, true},
		{"fresh decoder after alias remains unproven", `d := json.NewDecoder(r); alias := d; _ = alias; d = json.NewDecoder(r); d.DisallowUnknownFields(); var m Peer; _ = d.Decode(&m)`, false},
		{"no strict flag", `d := json.NewDecoder(r); var m Peer; _ = d.Decode(&m)`, false},
		{"local type", `d := json.NewDecoder(r); d.DisallowUnknownFields(); var m Local; _ = d.Decode(&m)`, false},
		{"different decoder", `d := json.NewDecoder(r); other := json.NewDecoder(r); d.DisallowUnknownFields(); var m Peer; _ = other.Decode(&m)`, false},
		{"shadowed json import", `json := factory{}; d := json.NewDecoder(r); d.DisallowUnknownFields(); var m Peer; _ = d.Decode(&m)`, false},
		{"rebound decoder", `d := json.NewDecoder(r); d.DisallowUnknownFields(); d = json.NewDecoder(r); var m Peer; _ = d.Decode(&m)`, false},
		{"indirect decoder overwrite", `d := json.NewDecoder(r); d.DisallowUnknownFields(); *d = *json.NewDecoder(r); var m Peer; _ = d.Decode(&m)`, false},
		{"conditional indirect decoder overwrite", `d := json.NewDecoder(r); d.DisallowUnknownFields(); if enabled { *d = *json.NewDecoder(r) }; var m Peer; _ = d.Decode(&m)`, false},
		{"alias before strict flag", `d := json.NewDecoder(r); alias := d; d.DisallowUnknownFields(); *alias = *json.NewDecoder(r); var m Peer; _ = d.Decode(&m)`, false},
		{"alias without overwrite is still unproven", `d := json.NewDecoder(r); alias := d; _ = alias; d.DisallowUnknownFields(); var m Peer; _ = d.Decode(&m)`, false},
		{"alias after strict flag", `d := json.NewDecoder(r); d.DisallowUnknownFields(); alias := d; *alias = *json.NewDecoder(r); var m Peer; _ = d.Decode(&m)`, false},
		{"var alias after strict flag", `d := json.NewDecoder(r); d.DisallowUnknownFields(); var alias = d; *alias = *json.NewDecoder(r); var m Peer; _ = d.Decode(&m)`, false},
		{"helper escape", `d := json.NewDecoder(r); d.DisallowUnknownFields(); consume := func(*json.Decoder) {}; consume(d); var m Peer; _ = d.Decode(&m)`, false},
		{"closure capture", `d := json.NewDecoder(r); d.DisallowUnknownFields(); overwrite := func() { *d = *json.NewDecoder(r) }; overwrite(); var m Peer; _ = d.Decode(&m)`, false},
		{"address alias survives fresh assignment", `d := json.NewDecoder(r); ref := &d; d = json.NewDecoder(r); d.DisallowUnknownFields(); *ref = json.NewDecoder(r); _ = d.Decode(&Peer{})`, false},
		{"address escaped before construction", `var d *json.Decoder; ref := &d; d = json.NewDecoder(r); d.DisallowUnknownFields(); *ref = json.NewDecoder(r); _ = d.Decode(&Peer{})`, false},
		{"closure captured before construction", `var d *json.Decoder; overwrite := func() { d = json.NewDecoder(r) }; d = json.NewDecoder(r); d.DisallowUnknownFields(); overwrite(); _ = d.Decode(&Peer{})`, false},
		{"second constructor disqualifies name", `d := json.NewDecoder(r); d = json.NewDecoder(r); d.DisallowUnknownFields(); _ = d.Decode(&Peer{})`, false},
		{"nonreceiver after decode disqualifies name", `d := json.NewDecoder(r); d.DisallowUnknownFields(); _ = d.Decode(&Peer{}); _ = d`, false},
		{"range decoder rebind", `d := json.NewDecoder(r); d.DisallowUnknownFields(); var i int; for i, d = range []*json.Decoder{json.NewDecoder(r)} {}; _ = i; var m Peer; _ = d.Decode(&m)`, false},
		{"conditional decoder rebind", `d := json.NewDecoder(r); d.DisallowUnknownFields(); if enabled { d = json.NewDecoder(r) }; var m Peer; _ = d.Decode(&m)`, false},
		{"multiple assignment rebind", `d := json.NewDecoder(r); d.DisallowUnknownFields(); var other interface{}; d, other = json.NewDecoder(r), nil; var m Peer; _ = d.Decode(&m)`, false},
		{"interface value target", `d := json.NewDecoder(r); d.DisallowUnknownFields(); var m any = Peer{}; _ = d.Decode(&m)`, false},
		{"interface pointer target", `d := json.NewDecoder(r); d.DisallowUnknownFields(); var m any = &Peer{}; _ = d.Decode(m)`, false},
		{"direct target rebind", `d := json.NewDecoder(r); d.DisallowUnknownFields(); m := new(Peer); m = new(Peer); _ = d.Decode(m)`, false},
		{"multiple target rebind", `d := json.NewDecoder(r); d.DisallowUnknownFields(); m := new(Peer); var other *Peer; m, other = other, m; _ = d.Decode(m)`, false},
		{"conditional target rebind", `d := json.NewDecoder(r); d.DisallowUnknownFields(); m := new(Peer); if enabled { m = new(Peer) }; _ = d.Decode(m)`, false},
		{"range target rebind", `d := json.NewDecoder(r); d.DisallowUnknownFields(); m := new(Peer); var i int; for i, m = range []*Peer{new(Peer)} {}; _ = i; _ = d.Decode(m)`, false},
		{"conditional flag", `d := json.NewDecoder(r); if enabled { d.DisallowUnknownFields() }; var m Peer; _ = d.Decode(&m)`, false},
		{"shadowed local type", `d := json.NewDecoder(r); d.DisallowUnknownFields(); type Peer struct{}; var m Peer; _ = d.Decode(&m)`, false},
		{"shadowed target", `d := json.NewDecoder(r); d.DisallowUnknownFields(); var m Peer; { var m Local; _ = d.Decode(&m) }`, false},
		{"flag after decode", `d := json.NewDecoder(r); var m Peer; _ = d.Decode(&m); d.DisallowUnknownFields()`, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			src := `package p
import ("encoding/json"; "io")
// wire: a peer can add optional fields.
type Peer struct { ID string }
type Local struct { ID string }
func read(r io.Reader, enabled bool) { ` + tc.body + ` }`
			got, err := Check("p.go", []byte(src), Catalog)
			if err != nil {
				t.Fatal(err)
			}
			hits := 0
			for _, f := range got {
				if f.Rule == "wire-strict-decode" {
					hits++
				}
			}
			if (hits == 1) != tc.want || hits > 1 {
				t.Errorf("findings = %v; want wire-strict-decode=%v", got, tc.want)
			}
		})
	}
}
