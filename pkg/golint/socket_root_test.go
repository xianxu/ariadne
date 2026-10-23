package golint

import (
	"fmt"
	"go/parser"
	"go/token"
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

const socketRule = "test-socket-fixture-root"
const socketHeader = `package fixture
import (
 "os"
 "net"
 "path/filepath"
 "testing"
 "github.com/xianxu/pair/cmd/internal/couchmessage"
)
func fixture(t *testing.T) {
`

func socketFindings(t *testing.T, source string) []Finding {
	t.Helper()
	got, err := Check("fixture_test.go", []byte(source), Catalog)
	if err != nil {
		t.Fatal(err)
	}
	var out []Finding
	for _, f := range got {
		if f.Rule == socketRule {
			out = append(out, f)
		}
	}
	return out
}

func TestSocketFixtureRootImportAndEngineBoundary(t *testing.T) {
	for _, tc := range []struct {
		name, source string
		want         int
	}{
		{"aliases", `package p; import(o "os"; n "net"; p "path/filepath"; tt "testing"); func f(t *tt.T){n.Listen("unix",p.Join(o.TempDir(),"rpc")); t.Setenv("COUCH_ISOLATED_ROOT",t.TempDir())}`, 2},
		{"parenthesized assignment", socketHeader + `var root string; ((root))=t.TempDir(); net.Listen("unix",root)` + "\n}", 1},
		{"parenthesized tuple", socketHeader + `root,err:=((os.MkdirTemp("","s-"))); _=err; net.Listen("unix",root)` + "\n}", 1},
		{"parenthesized tuple var", socketHeader + `var root,err=((filepath.Abs(t.TempDir()))); _=err; net.Listen("unix",root)` + "\n}", 1},
		{"parenthesized tuple error", socketHeader + `root,err:=((os.MkdirTemp("","s-"))); _=root; net.Listen("unix",err)` + "\n}", 0},
		{"parenthesized receiver", socketHeader + `(t).Setenv(("COUCH_ISOLATED_ROOT"),((t).TempDir)())` + "\n}", 1},
		{"parenthesized imported function", socketHeader + `(net.Listen)(("unix"),(os.TempDir)())` + "\n}", 1},
		{"parenthesized mkdir parent", socketHeader + `root,_:=(os.MkdirTemp((""),"s-")); net.Listen("unix",root)` + "\n}", 1},
		{"typed local TempDir", socketHeader + `var tt *testing.T=t; net.Listen("unix",tt.TempDir())` + "\n}", 1},
		{"typed local Setenv", socketHeader + `var tt *testing.T=t; tt.Setenv("COUCH_ISOLATED_ROOT",os.TempDir())` + "\n}", 1},
		{"typed local TB", socketHeader + `var tt testing.TB=t; net.Listen("unix",tt.TempDir())` + "\n}", 1},
		{"typed local value", socketHeader + `var tt testing.T; net.Listen("unix",tt.TempDir())` + "\n}", 1},
		{"typed local parenthesized type", socketHeader + `var tt (*(testing.T))=t; (tt.Setenv)("COUCH_ISOLATED_ROOT",tt.TempDir())` + "\n}", 1},
		{"typed unrelated receiver", socketHeader + `var tt Other=t; tt.Setenv("COUCH_ISOLATED_ROOT",os.TempDir()); net.Listen("unix",tt.TempDir())` + "\n}", 0},
		{"inferred receiver boundary", socketHeader + `tt:=t; tt.Setenv("COUCH_ISOLATED_ROOT",tt.TempDir()); net.Listen("unix",tt.TempDir())` + "\n}", 0},
		{"wrong package", `package p; import(n "example.net/net"; "os"); func f(){n.Listen("unix",os.TempDir())}`, 0},
		{"wrong testing", `package p; import("net"; tt "example.net/testing"); func f(t *tt.T){net.Listen("unix",t.TempDir())}`, 0},
		{"TB", `package p; import("net"; "testing"); func f(t testing.TB){net.Listen("unix",t.TempDir())}`, 1},
		{"B", `package p; import("net"; "testing"); func f(t *testing.B){net.Listen("unix",t.TempDir())}`, 1},
		{"F", `package p; import("net"; "testing"); func f(t *testing.F){net.Listen("unix",t.TempDir())}`, 1},
		{"address network unrelated", socketHeader + `net.ListenUnix("unix",&net.UnixAddr{Name:shortRoot(),Net:t.TempDir()})` + "\n}", 0},
		{"address positional", socketHeader + `net.ListenUnix("unix",&net.UnixAddr{t.TempDir(),"unix"})` + "\n}", 1},
		{"opaque function alias", socketHeader + `listen:=net.Listen; listen("unix",t.TempDir())` + "\n}", 0},
		{"dynamic env key", socketHeader + `os.Setenv(key,t.TempDir())` + "\n}", 0},
		{"unused namespace", socketHeader + `couchmessage.SocketPath(t.TempDir(),"broker"); t.Setenv("TMPDIR",t.TempDir())` + "\n}", 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := socketFindings(t, tc.source)
			if len(got) != tc.want {
				t.Fatalf("got %v want %d", got, tc.want)
			}
			all, err := CheckSources("example/p", map[string][]byte{"fixture_test.go": []byte(tc.source)}, Catalog)
			if err != nil {
				t.Fatal(err)
			}
			count := 0
			for _, f := range all["fixture_test.go"] {
				if f.Rule == socketRule {
					count++
				}
			}
			if count != tc.want {
				t.Fatalf("CheckSources %v", all)
			}
			prod, err := Check("fixture.go", []byte(tc.source), []Rule{ruleOf(t, socketRule)})
			if err != nil || len(prod) != 0 {
				t.Fatalf("test-only boundary: %v %v", prod, err)
			}
		})
	}
	for _, tc := range []struct {
		body string
		want int
	}{
		{"root:=t.TempDir(); root=shortRoot()\n//lint:allow test-socket-fixture-root — overwrite proven on this linear path\nnet.Listen(\"unix\",root)\n", 0},
		{"//lint:allow test-socket-fixture-root — stale allowance\nnet.Listen(\"unix\",shortRoot())\n", 1},
		{"//lint:allow test-socket-fixture-root\nnet.Listen(\"unix\",t.TempDir())\n", 2},
	} {
		got, err := Check("fixture_test.go", []byte(socketHeader+tc.body+"}"), []Rule{ruleOf(t, socketRule)})
		if err != nil || len(got) != tc.want {
			t.Fatalf("suppressions: %v %v want%d", got, err, tc.want)
		}
	}
}

func TestSocketFixtureRootLineVisits(t *testing.T) {
	for _, n := range []int{64, 128} {
		var src strings.Builder
		src.WriteString(socketHeader + "a0:=t.TempDir()\n")
		for i := 1; i <= n; i++ {
			fmt.Fprintf(&src, "a%d:=filepath.Join(a%d,\"p\"); net.Listen(\"unix\",a%d)\n", i, i-1, i)
		}
		// Shared fan-out and a back edge must not trigger per-sink rescanning.
		for i := 1; i <= n; i++ {
			fmt.Fprintf(&src, "b%d:=a0; net.Listen(\"unix\",b%d)\n", i, i)
		}
		fmt.Fprintf(&src, "a0=a%d\n}", n)
		f, err := parser.ParseFile(token.NewFileSet(), "fixture_test.go", src.String(), 0)
		if err != nil {
			t.Fatal(err)
		}
		got, c := analyzeSocketRoots(f)
		if len(got) != 2*n {
			t.Fatalf("N%d findings%d", n, len(got))
		}
		if c.AST > 80*n || c.Nodes > 60*n || c.Edges > 60*n || c.PropagatedNodes > c.Nodes || c.PropagatedEdges > c.Edges || c.Sinks != 2*n || c.Normalized > c.AST {
			t.Fatalf("N%d exceeded one-pass node/edge/sink budgets: %+v", n, c)
		}
		t.Logf("N%d: %+v", n, c)
	}
}

func TestSocketFixtureRootSharedTypeNormalization(t *testing.T) {
	for _, n := range []int{64, 128} {
		var src strings.Builder
		fmt.Fprintf(&src, "package p; import(\"testing\";\"net\"); func f(t *%stesting.T%s){\n", strings.Repeat("(", n), strings.Repeat(")", n))
		for i := 0; i < n; i++ {
			src.WriteString("net.Listen(\"unix\",(t).TempDir())\n")
		}
		src.WriteString("}")
		f, err := parser.ParseFile(token.NewFileSet(), "fixture_test.go", src.String(), 0)
		if err != nil {
			t.Fatal(err)
		}
		got, c := analyzeSocketRoots(f)
		if len(got) != n || c.Normalized > c.AST {
			t.Fatalf("N%d repeated shared-type normalization: findings%d counts%+v", n, len(got), c)
		}
		t.Logf("N%d shared type: %+v", n, c)
	}
}

// This is the approved short-root fixture idiom, not a runtime allocator:
// ordinary files retain TMPDIR; only length-sensitive native sockets use it.
func socketFixtureParent(scratchUsable bool) string {
	if scratchUsable {
		return "/Volumes/TestScratch"
	}
	return "/tmp"
}

func TestSocketFixtureShortRootWithLongTMPDIR(t *testing.T) {
	long := filepath.Join(t.TempDir(), strings.Repeat("long", 40))
	if err := os.MkdirAll(long, 0700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("TMPDIR", long)
	if os.TempDir() != long {
		t.Fatal("long inherited TMPDIR not established")
	}
	if socketFixtureParent(true) != "/Volumes/TestScratch" || socketFixtureParent(false) != "/tmp" {
		t.Fatal("short-root preference/fallback")
	}
	usable := false
	if info, err := os.Stat("/Volumes/TestScratch"); err == nil && info.IsDir() {
		usable = true
	}
	parent := socketFixtureParent(usable)
	var root string
	t.Run("owned listener", func(t *testing.T) {
		var err error
		root, err = os.MkdirTemp(parent, "sr-")
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() {
			if err := os.RemoveAll(root); err != nil {
				t.Error(err)
			}
		})
		physical, err := filepath.EvalSymlinks(root)
		if err != nil {
			t.Fatal(err)
		}
		socket := filepath.Join(physical, "m", strings.Repeat("a", 40)+".sock")
		if len(socket) > 103 {
			t.Fatalf("short fixture path has %d bytes", len(socket))
		}
		if err := os.Mkdir(filepath.Dir(socket), 0700); err != nil {
			t.Fatal(err)
		}
		listener, err := net.ListenUnix("unix", &net.UnixAddr{Name: socket, Net: "unix"})
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() {
			if err := listener.Close(); err != nil {
				t.Error(err)
			}
		})
		if err := listener.SetDeadline(time.Now().Add(5 * time.Second)); err != nil {
			t.Fatal(err)
		}
		conn, err := net.DialTimeout("unix", socket, 5*time.Second)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { conn.Close() })
		accepted, err := listener.AcceptUnix()
		if err != nil {
			t.Fatal(err)
		}
		if err := accepted.Close(); err != nil {
			t.Fatal(err)
		}
		t.Logf("private socket under %s fits at %d bytes despite TMPDIR length%d", parent, len(socket), len(long))
	})
	if root == "" {
		t.Fatal("fixture did not create its private root")
	}
	if _, err := os.Lstat(root); !os.IsNotExist(err) {
		t.Fatalf("private root cleanup: %v", err)
	}
}

func TestSocketFixtureRootRecordedPairSource(t *testing.T) {
	for _, tc := range []struct {
		name string
		want int
	}{{"unsafe", 1}, {"safe", 0}} {
		t.Run(tc.name, func(t *testing.T) {
			src, err := os.ReadFile("testdata/socket-root/pair-" + tc.name + ".go.txt")
			if err != nil {
				t.Fatal(err)
			}
			got := socketFindings(t, string(src))
			if len(got) != tc.want {
				t.Fatalf("got %v, want %d socket-root findings", got, tc.want)
			}
			for _, f := range got {
				line := strings.Split(string(src), "\n")[f.Line-1]
				if !strings.Contains(line, `t.Setenv("COUCH_ISOLATED_ROOT", root)`) {
					t.Fatalf("wrong sink: %v: %s", f, line)
				}
			}
		})
	}
}

func TestSocketFixtureRootFlows(t *testing.T) {
	for _, tc := range []struct {
		name, body string
		want       int
	}{
		{"direct", `net.Listen("unix",os.TempDir())`, 1},
		{"testing", `net.Listen("unix",filepath.Join(t.TempDir(),"rpc"))`, 1},
		{"mkdir tuple", `root,err:=os.MkdirTemp("","ro-"); _=err; t.Setenv("COUCH_ISOLATED_ROOT",root)`, 1},
		{"inherited parent", `root,_:=os.MkdirTemp(os.TempDir(),"ro-"); net.Listen("unix",root)`, 1},
		{"alias join", `root:=t.TempDir(); a:=root; b:=filepath.Join(a,"store"); os.Setenv("COUCH_ISOLATED_ROOT",b)`, 1},
		{"canonical tuple", `root,_:=filepath.EvalSymlinks(t.TempDir()); root,_=filepath.Abs(root); net.Listen("unix",filepath.Clean(root))`, 1},
		{"var declaration", `var root= os.TempDir(); net.Dial("unix",root+"/rpc")`, 1},
		{"branch write", `root:=shortRoot(); if flag {root=t.TempDir()}; net.Listen("unix",root)`, 1},
		{"cycle", `a,b:=shortRoot(),shortRoot(); a=b; b=a; b=t.TempDir(); net.Listen("unix",a)`, 1},
		{"address", `addr:=&net.UnixAddr{Net:"unix",Name:filepath.Join(t.TempDir(),"rpc")}; alias:=addr; net.ListenUnix("unix",alias)`, 1},
		{"dial remote", `addr:=&net.UnixAddr{Net:"unix",Name:os.TempDir()}; net.DialUnix("unix",nil,addr)`, 1},
		{"dial local", `addr:=&net.UnixAddr{Net:"unix",Name:os.TempDir()}; net.DialUnix("unix",addr,nil)`, 1},
		{"packet", `net.ListenPacket("unixgram",t.TempDir())`, 1},
		{"resolve", `net.ResolveUnixAddr("unix",t.TempDir())`, 1},
		{"unixgram", `net.ListenUnixgram("unixgram",&net.UnixAddr{Name:t.TempDir()})`, 1},
		{"parens", `a:=(t.TempDir()); net.Listen("unix",(a))`, 1},
		{"short root", `root:=shortRoot(); _=t.TempDir(); net.Listen("unix",filepath.Join(root,"rpc"))`, 0},
		{"unrelated temp", `long:=filepath.Join(t.TempDir(),"long"); t.Setenv("TMPDIR",long); net.Listen("unix",shortRoot())`, 0},
		{"hashed namespace", `root:=t.TempDir(); couchmessage.SocketPath(filepath.Join(root,"store"),"broker")`, 0},
		{"bounded parent hashed namespace", `t.Setenv("COUCH_ISOLATED_ROOT",shortRoot()); couchmessage.SocketPath(t.TempDir(),"broker")`, 0},
		{"tcp", `net.Listen("tcp",os.TempDir()); net.Dial("tcp4",t.TempDir())`, 0},
		{"unknown network", `net.Listen(network,t.TempDir())`, 0},
		{"shadow local", `root:=shortRoot(); {root:=t.TempDir(); _=root}; net.Listen("unix",root)`, 0},
		{"shadow os", `os:=other; net.Listen("unix",os.TempDir())`, 0},
		{"shadow net", `net:=other; net.Listen("unix",t.TempDir())`, 0},
		{"not testing", `other.TempDir(); other.Setenv("COUCH_ISOLATED_ROOT",t.TempDir())`, 0},
		{"opaque helper", `root:=opaque(t.TempDir()); net.Listen("unix",root)`, 0},
		{"error result", `root,err:=os.MkdirTemp("","ro-"); _=root; net.Listen("unix",err)`, 0},
		{"tuple error result", `root,err:=filepath.Abs(t.TempDir()); _=root; net.Listen("unix",err)`, 0},
		{"literal variable shadow", `root:=t.TempDir(); {root:=shortRoot(); net.Listen("unix",root)}`, 0},
		{"comment string", `_="net.Listen(\"unix\",t.TempDir())" // net.Listen("unix",t.TempDir())`, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := socketFindings(t, socketHeader+tc.body+"\n}")
			if len(got) != tc.want {
				t.Fatalf("%s: got %v, want %d", tc.body, got, tc.want)
			}
		})
	}
}
