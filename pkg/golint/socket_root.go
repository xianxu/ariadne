package golint

import (
	"go/ast"
	"go/token"
)

// Syntax-only provenance, not a path-length or whole-program safety proof.
// Every visible write contributes; opaque calls and container mutation do not.
// In particular, Pair's SocketPath namespace is hashed, so it is NOT a sink.
func socketFixtureRoots(f *ast.File, _ *CheckContext) []ast.Node {
	out, _ := analyzeSocketRoots(f)
	return out
}

type socketRootNode struct {
	tainted bool
	next    []*socketRootNode
}

type socketRootCounts struct {
	AST, Nodes, Edges, PropagatedNodes, PropagatedEdges, Sinks int
}

type socketRootGraph struct {
	nodes  map[any]*socketRootNode
	queue  []*socketRootNode
	counts socketRootCounts
}

func (g *socketRootGraph) node(key any) *socketRootNode {
	if n := g.nodes[key]; n != nil {
		return n
	}
	n := &socketRootNode{}
	g.nodes[key] = n
	g.counts.Nodes++
	return n
}

func (g *socketRootGraph) edge(from, to any) {
	a, b := g.node(from), g.node(to)
	a.next = append(a.next, b)
	g.counts.Edges++
}

func (g *socketRootGraph) mark(n *socketRootNode) {
	if !n.tainted {
		n.tainted = true
		g.queue = append(g.queue, n)
	}
}

func (g *socketRootGraph) propagateInheritedRoots() {
	for i := 0; i < len(g.queue); i++ {
		g.counts.PropagatedNodes++
		for _, next := range g.queue[i].next {
			g.counts.PropagatedEdges++
			g.mark(next)
		}
	}
}

type socketRootAPIs struct{ os, net, filepath, testing string }

func (a socketRootAPIs) imported(expr ast.Expr, pkg, name string) bool {
	sel, ok := expr.(*ast.SelectorExpr)
	if !ok || sel.Sel.Name != name || pkg == "" || pkg == "." || pkg == "_" {
		return false
	}
	id, ok := sel.X.(*ast.Ident)
	return ok && id.Obj == nil && id.Name == pkg
}

func (a socketRootAPIs) testingMethod(expr ast.Expr, name string) bool {
	sel, ok := expr.(*ast.SelectorExpr)
	if !ok || sel.Sel.Name != name {
		return false
	}
	id, ok := sel.X.(*ast.Ident)
	if !ok || id.Obj == nil {
		return false
	}
	field, ok := id.Obj.Decl.(*ast.Field)
	if !ok {
		return false
	}
	typ := field.Type
	if p, ok := typ.(*ast.StarExpr); ok {
		typ = p.X
	}
	for _, name := range []string{"T", "B", "F", "TB"} {
		if a.imported(typ, a.testing, name) {
			return true
		}
	}
	return false
}

func (a socketRootAPIs) firstPathResult(call *ast.CallExpr) bool {
	return a.imported(call.Fun, a.os, "MkdirTemp") || a.imported(call.Fun, a.filepath, "Abs") || a.imported(call.Fun, a.filepath, "EvalSymlinks")
}

// socketRootSink selects only length-sensitive paths/addresses. Literal network
// and environment keys are deliberate analysis boundaries, not runtime guesses.
func socketRootSink(c *ast.CallExpr, a socketRootAPIs) []ast.Expr {
	if len(c.Args) < 2 {
		return nil
	}
	key, literal := stringLit(c.Args[0])
	if !literal {
		return nil
	}
	if key == "COUCH_ISOLATED_ROOT" && (a.imported(c.Fun, a.os, "Setenv") || a.testingMethod(c.Fun, "Setenv")) {
		return c.Args[1:2]
	}
	if key != "unix" && key != "unixgram" && key != "unixpacket" {
		return nil
	}
	for _, name := range []string{"Listen", "ListenPacket", "Dial", "ResolveUnixAddr", "ListenUnix", "ListenUnixgram"} {
		if a.imported(c.Fun, a.net, name) {
			return c.Args[1:2]
		}
	}
	if a.imported(c.Fun, a.net, "DialUnix") && len(c.Args) == 3 {
		return c.Args[1:3]
	}
	return nil
}

func analyzeSocketRoots(f *ast.File) ([]ast.Node, socketRootCounts) {
	a := socketRootAPIs{localName(f, "os"), localName(f, "net"), localName(f, "path/filepath"), localName(f, "testing")}
	g := socketRootGraph{nodes: map[any]*socketRootNode{}}
	type sink struct {
		call *ast.CallExpr
		args []ast.Expr
	}
	var sinks []sink
	// Bindings and expression edges are indexed once; worklist propagation later
	// handles forward references, branching writes, shared aliases and cycles.
	bind := func(names []ast.Expr, values []ast.Expr) {
		for i, lhs := range names {
			id, ok := lhs.(*ast.Ident)
			if !ok || id.Obj == nil || id.Name == "_" {
				continue
			}
			if len(names) == len(values) {
				g.edge(values[i], id.Obj)
			} else if i == 0 && len(values) == 1 {
				if call, ok := values[0].(*ast.CallExpr); ok && a.firstPathResult(call) {
					g.edge(call, id.Obj)
				}
			}
		}
	}
	ast.Inspect(f, func(n ast.Node) bool {
		if n == nil {
			return false
		}
		g.counts.AST++
		switch x := n.(type) {
		case *ast.Ident:
			if x.Obj != nil {
				g.edge(x.Obj, x)
			}
		case *ast.AssignStmt:
			bind(x.Lhs, x.Rhs)
		case *ast.ValueSpec:
			names := make([]ast.Expr, len(x.Names))
			for i, id := range x.Names {
				names[i] = id
			}
			bind(names, x.Values)
		case *ast.ParenExpr:
			g.edge(x.X, x)
		case *ast.BinaryExpr:
			if x.Op == token.ADD {
				g.edge(x.X, x)
				g.edge(x.Y, x)
			}
		case *ast.UnaryExpr:
			if x.Op == token.AND {
				g.edge(x.X, x)
			}
		case *ast.CompositeLit:
			if a.imported(x.Type, a.net, "UnixAddr") {
				for i, elt := range x.Elts {
					if kv, ok := elt.(*ast.KeyValueExpr); ok {
						if id, ok := kv.Key.(*ast.Ident); ok && id.Name == "Name" {
							g.edge(kv.Value, x)
						}
					} else if i == 0 {
						g.edge(elt, x)
					} // net.UnixAddr{Name, Net}
				}
			}
		case *ast.CallExpr:
			if (a.imported(x.Fun, a.os, "TempDir") || a.testingMethod(x.Fun, "TempDir")) && len(x.Args) == 0 {
				g.mark(g.node(x))
			}
			if a.imported(x.Fun, a.os, "MkdirTemp") && len(x.Args) == 2 {
				if parent, ok := stringLit(x.Args[0]); ok && parent == "" {
					g.mark(g.node(x))
				}
				g.edge(x.Args[0], x)
			}
			for _, name := range []string{"Join", "Clean", "Abs", "EvalSymlinks"} {
				if a.imported(x.Fun, a.filepath, name) {
					for _, arg := range x.Args {
						g.edge(arg, x)
					}
				}
			}
			if args := socketRootSink(x, a); len(args) > 0 {
				sinks = append(sinks, sink{x, args})
			}
		}
		return true
	})
	g.propagateInheritedRoots()
	var out []ast.Node
	for _, s := range sinks {
		g.counts.Sinks++
		for _, arg := range s.args {
			if n := g.nodes[arg]; n != nil && n.tainted {
				out = append(out, s.call)
				break
			}
		}
	}
	return out, g.counts
}
