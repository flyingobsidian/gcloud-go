// Command migrateglobals rewrites a gcloud-go command group so that its
// package globals live on a per-invocation struct (#2070).
//
// It moves the group's command and flag variables onto a struct, turns its
// init functions and every function that uses the moved variables or process
// state into methods, and routes output, context, clients and global flags
// through the invocation. It prints references from outside the group that
// need fixing by hand. See doc/migrating-command-groups.md.
//
// Usage, from the repository root:
//
//	go run ./tools/migrateglobals -dir . -files storage.go,storage_hmac.go \
//		-type storageCmds -root storageCmd -ctor newStorageCmd
//
// Add -n to report without writing, or -audit to list helpers outside the
// group that still use process state.
package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"golang.org/x/tools/go/packages"
	"golang.org/x/tools/imports"
)

var (
	dir      = flag.String("dir", ".", "module root")
	filesArg = flag.String("files", "", "comma-separated group files (relative to cmd/)")
	typeName = flag.String("type", "", "state struct name, e.g. storageCmds")
	rootVar  = flag.String("root", "", "group root command var, e.g. storageCmd")
	ctorName = flag.String("ctor", "", "constructor name, e.g. newStorageCmd")
	keepArg  = flag.String("keep", "", "comma-separated funcs to keep as free functions")
	dryRun   = flag.Bool("n", false, "report only")
	fixTests = flag.Bool("fixtests", false, "rewrite references from _test.go files to use a legacyTestInvocation")
	audit    = flag.Bool("audit", false, "report calls from the group to outside helpers that use process globals")
)

// globalFlags maps legacy global flag vars to invocation fields.
var globalFlags = map[string]string{
	"flagProject": "project", "flagZone": "zone", "flagAccount": "account", "flagQuiet": "quiet",
	"flagAccessTokenFile": "accessTokenFile", "flagBillingProject": "billingProject",
	"flagConfiguration": "configuration", "flagFlagsFile": "flagsFile", "flagFilter": "filter",
	"flagFlatten": "flatten", "flagFormat": "format", "flagImpersonateSA": "impersonateSA",
	"flagLogHTTP": "logHTTP", "flagTraceToken": "traceToken", "flagUserOutputEnabled": "userOutputEnabled",
	"flagVerbosity": "verbosity",
}

// simple call renames: package-level cmd func -> replacement for the Fun part,
// and an optional first argument to insert.
type callRule struct{ fun, firstArg string }

var cmdCalls = map[string]callRule{
	"emitFormatted":         {"emitFormattedTo", "s.inv.stdout"},
	"printJSON":             {"s.inv.printJSON", ""},
	"yamlEncode":            {"s.inv.yaml", ""},
	"printTableHeading":     {"printTableHeadingTo", "s.inv.stdout"},
	"emitWithPipeline":      {"s.inv.emitWithPipeline", ""},
	"printListResults":      {"s.inv.printListResults", ""},
	"resolveProject":        {"s.inv.project", ""},
	"resolveProjectZone":    {"s.inv.projectZone", ""},
	"promptYesNo":           {"s.inv.confirm", ""},
	"promptContinue":        {"s.inv.promptContinue", ""},
	"formatOutput":          {"s.inv.formatOutput", ""},
	"IsInteractive":         {"", ""}, // whole call -> s.inv.interactive
	"resolveRegion":         {"s.inv.region", ""},
	"httpDebugWriter":       {"s.inv.httpDebugWriter", ""},
	"newRESTClient":         {"s.inv.restClient", ""},
	"resolveBillingProject": {"s.inv.billingProject", ""},
	"restUserProject":       {"s.inv.quotaProject", ""},
	"adcFilePath":           {"s.inv.adcFilePath", ""},
}

var fmtCalls = map[string]string{"Printf": "Fprintf", "Println": "Fprintln", "Print": "Fprint"}

type edit struct {
	start, end int // byte offsets
	text       string
	del        bool // inner edits inside are dropped
}

func main() {
	flag.Parse()
	files := map[string]bool{}
	var order []string
	for _, f := range strings.Split(*filesArg, ",") {
		p, _ := filepath.Abs(filepath.Join(*dir, "cmd", f))
		files[p] = true
		order = append(order, p)
	}
	sort.Strings(order)
	keep := map[string]bool{}
	for _, k := range strings.Split(*keepArg, ",") {
		if k != "" {
			keep[k] = true
		}
	}

	cfg := &packages.Config{Mode: packages.NeedName | packages.NeedSyntax | packages.NeedTypes | packages.NeedTypesInfo | packages.NeedFiles, Dir: *dir, Tests: true}
	pkgs, err := packages.Load(cfg, "./cmd")
	if err != nil {
		log.Fatal(err)
	}
	// Pick the test variant (has all files incl. _test.go) for reporting, the
	// plain package for rewriting.
	var pkg, tpkg *packages.Package
	for _, p := range pkgs {
		if p.ID == "github.com/flyingobsidian/gcloud-go/cmd" {
			pkg = p
		} else if strings.HasPrefix(p.ID, "github.com/flyingobsidian/gcloud-go/cmd [") {
			tpkg = p
		}
	}
	if len(pkg.Errors) > 0 {
		log.Fatal(pkg.Errors)
	}
	fset := pkg.Fset
	info := pkg.TypesInfo
	scope := pkg.Types.Scope()
	legacy := legacyServiceMap(*dir)

	fileOf := func(pos token.Pos) string { return fset.Position(pos).Filename }
	inGroup := func(pos token.Pos) bool { return files[fileOf(pos)] }

	// 1. Package-level vars declared in the group.
	type mvar struct {
		obj  *types.Var
		spec *ast.ValueSpec
		decl *ast.GenDecl
		idx  int
		init ast.Expr
		doc  string
		// dependent vars are initialized from other moved vars, so they
		// are set after them.
		dependent bool
	}
	moved := map[types.Object]*mvar{}
	var movedOrder []*mvar
	// address-taken or assigned package vars
	mutated := map[types.Object]bool{}
	for _, f := range pkg.Syntax {
		ast.Inspect(f, func(n ast.Node) bool {
			switch n := n.(type) {
			case *ast.UnaryExpr:
				if n.Op == token.AND {
					if id, ok := n.X.(*ast.Ident); ok {
						mutated[info.Uses[id]] = true
					}
				}
			case *ast.AssignStmt:
				for _, l := range n.Lhs {
					if id, ok := l.(*ast.Ident); ok {
						if o := info.Uses[id]; o != nil {
							mutated[o] = true
						}
					}
				}
			case *ast.IncDecStmt:
				if id, ok := n.X.(*ast.Ident); ok {
					mutated[info.Uses[id]] = true
				}
			}
			return true
		})
	}
	var initFuncs []*ast.FuncDecl
	var others []*mvar // group vars not moved (yet)
	funcs := map[types.Object]*ast.FuncDecl{}
	for _, f := range pkg.Syntax {
		if !inGroup(f.Pos()) {
			continue
		}
		for _, d := range f.Decls {
			switch d := d.(type) {
			case *ast.GenDecl:
				if d.Tok != token.VAR {
					continue
				}
				for _, sp := range d.Specs {
					vs := sp.(*ast.ValueSpec)
					for i, name := range vs.Names {
						obj := info.Defs[name].(*types.Var)
						var init ast.Expr
						if i < len(vs.Values) {
							init = vs.Values[i]
						}
						isCmd := false
						if u, ok := init.(*ast.UnaryExpr); ok && u.Op == token.AND {
							if cl, ok := u.X.(*ast.CompositeLit); ok {
								if se, ok := cl.Type.(*ast.SelectorExpr); ok && se.Sel.Name == "Command" {
									isCmd = true
								}
							}
						}
						if isCmd || mutated[obj] {
							if len(vs.Values) > 0 && len(vs.Values) != len(vs.Names) {
								log.Fatalf("multi-value var spec for %s", name.Name)
							}
							doc := ""
							if vs.Doc != nil {
								doc = vs.Doc.Text()
							} else if len(d.Specs) == 1 && d.Doc != nil {
								doc = d.Doc.Text()
							}
							mv := &mvar{obj: obj, spec: vs, decl: d, idx: i, init: init, doc: doc}
							moved[obj] = mv
							movedOrder = append(movedOrder, mv)
						} else if init != nil && len(vs.Values) == len(vs.Names) {
							doc := ""
							if vs.Doc != nil {
								doc = vs.Doc.Text()
							} else if len(d.Specs) == 1 && d.Doc != nil {
								doc = d.Doc.Text()
							}
							others = append(others, &mvar{obj: obj, spec: vs, decl: d, idx: i, init: init, doc: doc})
						}
					}
				}
			case *ast.FuncDecl:
				if d.Recv != nil {
					continue
				}
				if d.Name.Name == "init" {
					initFuncs = append(initFuncs, d)
					continue
				}
				funcs[info.Defs[d.Name]] = d
			}
		}
	}

	// Vars initialized from moved vars move too.
	for changed := true; changed; {
		changed = false
		for _, mv := range others {
			if moved[mv.obj] != nil {
				continue
			}
			uses := false
			ast.Inspect(mv.init, func(n ast.Node) bool {
				if id, ok := n.(*ast.Ident); ok && moved[info.Uses[id]] != nil {
					uses = true
				}
				return true
			})
			if uses {
				mv.dependent = true
				moved[mv.obj] = mv
				movedOrder = append(movedOrder, mv)
				changed = true
			}
		}
	}

	if *audit {
		// all package funcs (any file) and whether they touch process state
		all := map[types.Object]*ast.FuncDecl{}
		for _, f := range pkg.Syntax {
			for _, d := range f.Decls {
				if fd, ok := d.(*ast.FuncDecl); ok && fd.Recv == nil && fd.Name.Name != "init" {
					all[info.Defs[fd.Name]] = fd
				}
			}
		}
		why := map[types.Object]string{}
		for o, fd := range all {
			ast.Inspect(fd, func(n ast.Node) bool {
				switch n := n.(type) {
				case *ast.Ident:
					if u := info.Uses[n]; u != nil && u.Parent() == scope {
						if _, ok := globalFlags[n.Name]; ok {
							why[o] = n.Name
						}
						if _, ok := u.(*types.Var); ok && !strings.HasPrefix(n.Name, "err") && mutated[u] && !files[fileOf(u.Pos())] {
							if why[o] == "" {
								why[o] = "var " + n.Name
							}
						}
					}
				case *ast.SelectorExpr:
					if pk, ok := n.X.(*ast.Ident); ok {
						if pn, ok := info.Uses[pk].(*types.PkgName); ok {
							p, sel := pn.Imported().Path(), n.Sel.Name
							if p == "fmt" && fmtCalls[sel] != "" || p == "os" && (sel == "Stdout" || sel == "Stderr" || sel == "Stdin" || sel == "Getenv") || p == "context" && sel == "Background" {
								why[o] = p + "." + sel
							}
						}
					}
				}
				return true
			})
		}
		for changed := true; changed; {
			changed = false
			for o, fd := range all {
				if why[o] != "" {
					continue
				}
				ast.Inspect(fd.Body, func(n ast.Node) bool {
					if id, ok := n.(*ast.Ident); ok {
						if u := info.Uses[id]; u != nil && why[u] != "" && why[o] == "" {
							why[o] = "calls " + u.Name()
						}
					}
					return true
				})
				if why[o] != "" {
					changed = true
				}
			}
		}
		seen := map[string]bool{}
		for _, f := range pkg.Syntax {
			if !inGroup(f.Pos()) {
				continue
			}
			ast.Inspect(f, func(n ast.Node) bool {
				if id, ok := n.(*ast.Ident); ok {
					u := info.Uses[id]
					if u == nil || all[u] == nil || inGroup(u.Pos()) || why[u] == "" {
						return true
					}
					if _, ok := cmdCalls[u.Name()]; ok {
						return true
					}
					k := u.Name()
					if !seen[k] {
						seen[k] = true
						fmt.Printf("AUDIT %s (%s) used at %s\n", k, why[u], fset.Position(id.Pos()))
					}
				}
				return true
			})
		}
		return
	}

	// 2. Which funcs need the state: transitive closure.
	needs := map[types.Object]bool{}
	direct := func(fd *ast.FuncDecl) bool {
		found := false
		ast.Inspect(fd, func(n ast.Node) bool {
			switch n := n.(type) {
			case *ast.Ident:
				o := info.Uses[n]
				if o == nil {
					return true
				}
				if moved[o] != nil || isRestClientVar(o, scope) {
					found = true
				}
				if _, ok := globalFlags[n.Name]; ok && o.Parent() == scope {
					found = true
				}
				if (n.Name == "rootCmd") && o.Parent() == scope {
					found = true
				}
				if _, ok := cmdCalls[n.Name]; ok && o.Parent() == scope {
					found = true
				}
			case *ast.SelectorExpr:
				if pk, ok := n.X.(*ast.Ident); ok {
					if pn, ok := info.Uses[pk].(*types.PkgName); ok {
						p, s := pn.Imported().Path(), n.Sel.Name
						if p == "fmt" && fmtCalls[s] != "" || p == "os" && (s == "Stdout" || s == "Stderr" || s == "Stdin" || s == "Getenv") ||
							p == "context" && s == "Background" || strings.HasSuffix(p, "internal/gcp") && legacy[s] != "" ||
							strings.HasSuffix(p, "internal/compute") && s == "NewService" ||
							strings.HasSuffix(p, "internal/config") && (s == "Load" || s == "Property" || s == "Resolve") {
							found = true
						}
					}
				}
			}
			return true
		})
		return found
	}
	for o, fd := range funcs {
		if direct(fd) {
			needs[o] = true
		}
	}
	for changed := true; changed; {
		changed = false
		for o, fd := range funcs {
			if needs[o] {
				continue
			}
			ast.Inspect(fd.Body, func(n ast.Node) bool {
				if id, ok := n.(*ast.Ident); ok && needs[info.Uses[id]] {
					needs[o] = true
				}
				return true
			})
			if needs[o] {
				changed = true
			}
		}
	}
	for o := range needs {
		if keep[o.Name()] {
			delete(needs, o)
		}
		if fd := funcs[o]; fd != nil && fd.Type.TypeParams != nil {
			fmt.Printf("GENERIC %s left as a function: give it an *invocation by hand\n", o.Name())
			delete(needs, o)
		}
	}

	// 3. Report outside references.
	report := func(p *packages.Package, tests bool) {
		for _, f := range p.Syntax {
			fn := fileOf(f.Pos())
			if files[fn] || strings.HasSuffix(fn, "_test.go") != tests {
				continue
			}
			ast.Inspect(f, func(n ast.Node) bool {
				if id, ok := n.(*ast.Ident); ok {
					o := p.TypesInfo.Uses[id]
					if o == nil {
						return true
					}
					for mo := range moved {
						if mo.Name() == o.Name() && mo.Pos() == o.Pos() {
							fmt.Printf("REF var %s at %s\n", o.Name(), fset.Position(id.Pos()))
						}
					}
					for fo := range needs {
						if fo.Name() == o.Name() && fo.Pos() == o.Pos() {
							fmt.Printf("REF func %s at %s\n", o.Name(), fset.Position(id.Pos()))
						}
					}
				}
				return true
			})
		}
	}
	report(pkg, false)
	if tpkg != nil {
		report(tpkg, true)
	}
	fmt.Printf("moved vars: %d, methods: %d, inits: %d\n", len(moved), len(needs), len(initFuncs))
	if *dryRun {
		return
	}
	if *fixTests && tpkg != nil {
		groupObjs := map[token.Pos]bool{}
		for o := range moved {
			groupObjs[o.Pos()] = true
		}
		for o := range needs {
			groupObjs[o.Pos()] = true
		}
		fixTestFiles(tpkg, groupObjs)
	}

	// Receiver name: the first candidate no converted function uses as a local.
	R := ""
	for _, cand := range []string{"s", "g", "grp", "cmds"} {
		used := false
		for o, fd := range funcs {
			if !needs[o] {
				continue
			}
			ast.Inspect(fd, func(n ast.Node) bool {
				if id, ok := n.(*ast.Ident); ok && id.Name == cand && info.Defs[id] != nil {
					used = true
				}
				return true
			})
		}
		for _, fd := range initFuncs {
			ast.Inspect(fd, func(n ast.Node) bool {
				if id, ok := n.(*ast.Ident); ok && id.Name == cand && info.Defs[id] != nil {
					used = true
				}
				return true
			})
		}
		if !used {
			R = cand
			break
		}
	}
	fmt.Printf("receiver: %s\n", R)

	// 4. Edits per file.
	edits := map[string][]edit{}
	src := map[string][]byte{}
	for _, p := range order {
		b, err := os.ReadFile(p)
		if err != nil {
			log.Fatal(err)
		}
		src[p] = b
	}
	off := func(pos token.Pos) int { return fset.Position(pos).Offset }
	add := func(pos token.Pos, e edit) { fn := fileOf(pos); edits[fn] = append(edits[fn], e) }

	// enclosing cobra param lookup
	cobraParam := func(stack []ast.Node) string {
		for i := len(stack) - 1; i >= 0; i-- {
			var ft *ast.FuncType
			switch f := stack[i].(type) {
			case *ast.FuncDecl:
				ft = f.Type
			case *ast.FuncLit:
				ft = f.Type
			default:
				continue
			}
			for _, fld := range ft.Params.List {
				if t, ok := fld.Type.(*ast.StarExpr); ok {
					if se, ok := t.X.(*ast.SelectorExpr); ok && se.Sel.Name == "Command" && len(fld.Names) == 1 && fld.Names[0].Name != "_" {
						return fld.Names[0].Name
					}
				}
			}
			if _, ok := stack[i].(*ast.FuncDecl); ok {
				return ""
			}
		}
		return ""
	}

	rewriteBody := func(root ast.Node, isState bool) {
		var stack []ast.Node
		ast.Inspect(root, func(n ast.Node) bool {
			if n == nil {
				stack = stack[:len(stack)-1]
				return true
			}
			stack = append(stack, n)
			switch n := n.(type) {
			case *ast.ExprStmt:
				// rootCmd.AddCommand(x) -> delete
				if c, ok := n.X.(*ast.CallExpr); ok {
					if se, ok := c.Fun.(*ast.SelectorExpr); ok && se.Sel.Name == "AddCommand" {
						if id, ok := se.X.(*ast.Ident); ok && id.Name == "rootCmd" {
							add(n.Pos(), edit{off(n.Pos()), off(n.End()), "", true})
						}
					}
				}
			case *ast.CallExpr:
				switch fun := n.Fun.(type) {
				case *ast.Ident:
					o := info.Uses[fun]
					if o == nil || o.Parent() != scope || needs[o] {
						break
					}
					r, ok := cmdCalls[fun.Name]
					if !ok || !isState {
						break
					}
					if fun.Name == "IsInteractive" {
						add(n.Pos(), edit{off(n.Pos()), off(n.End()), R + ".inv.interactive", true})
						break
					}
					add(fun.Pos(), edit{off(fun.Pos()), off(fun.End()), strings.Replace(r.fun, "s.", R+".", 1), false})
					if r.firstArg != "" {
						t := strings.Replace(r.firstArg, "s.", R+".", 1) + ", "
						if len(n.Args) == 0 {
							t = strings.Replace(r.firstArg, "s.", R+".", 1)
						}
						add(n.Lparen, edit{off(n.Lparen) + 1, off(n.Lparen) + 1, t, false})
					}
				case *ast.SelectorExpr:
					pk, ok := fun.X.(*ast.Ident)
					if !ok || !isState {
						break
					}
					pn, ok := info.Uses[pk].(*types.PkgName)
					if !ok {
						break
					}
					p, s := pn.Imported().Path(), fun.Sel.Name
					switch {
					case p == "fmt" && fmtCalls[s] != "":
						add(fun.Pos(), edit{off(fun.Pos()), off(fun.End()), "fmt." + fmtCalls[s], false})
						t := R + ".inv.stdout, "
						if len(n.Args) == 0 {
							t = R + ".inv.stdout"
						}
						add(n.Lparen, edit{off(n.Lparen) + 1, off(n.Lparen) + 1, t, false})
					case p == "context" && s == "Background":
						t := R + ".inv.context()"
						if c := cobraParam(stack); c != "" {
							t = c + ".Context()"
						}
						add(n.Pos(), edit{off(n.Pos()), off(n.End()), t, true})
					case (strings.HasSuffix(p, "internal/gcp") && legacy[s] != "" || strings.HasSuffix(p, "internal/compute") && s == "NewService") && len(n.Args) >= 2:
						m := legacy[s]
						if s == "NewService" {
							m = "Compute"
						}
						add(fun.Pos(), edit{off(fun.Pos()), off(fun.End()), R + ".inv.clients()." + m, false})
						if id, ok := n.Args[1].(*ast.Ident); ok && id.Name == "flagAccount" {
							add(n.Args[0].End(), edit{off(n.Args[0].End()), off(n.Args[1].End()), "", true})
						} else {
							fmt.Printf("WARN non-flagAccount account arg at %s\n", fset.Position(n.Pos()))
						}
					case strings.HasSuffix(p, "internal/config") && s == "Load" && len(n.Args) == 0:
						add(n.Pos(), edit{off(n.Pos()), off(n.End()), R + ".inv.properties()", true})
					case strings.HasSuffix(p, "internal/config") && s == "Property":
						add(fun.Pos(), edit{off(fun.Pos()), off(fun.End()), R + ".inv.env.Property", false})
					case strings.HasSuffix(p, "internal/config") && s == "Resolve":
						add(fun.Pos(), edit{off(fun.Pos()), off(fun.End()), R + ".inv.resolve", false})
					case p == "os" && s == "Getenv":
						add(fun.Pos(), edit{off(fun.Pos()), off(fun.End()), R + ".inv.env.Getenv", false})
					}
				}
			case *ast.SelectorExpr:
				if pk, ok := n.X.(*ast.Ident); ok && isState {
					if pn, ok := info.Uses[pk].(*types.PkgName); ok && pn.Imported().Path() == "os" {
						switch n.Sel.Name {
						case "Stdout", "Stderr", "Stdin":
							add(n.Pos(), edit{off(n.Pos()), off(n.End()), R + ".inv." + strings.ToLower(n.Sel.Name), true})
						}
					}
				}
			case *ast.Ident:
				o := info.Uses[n]
				if o == nil || o.Parent() != scope {
					break
				}
				if moved[o] != nil || needs[o] {
					add(n.Pos(), edit{off(n.Pos()), off(n.Pos()), R + ".", false})
				} else if isState && isRestClientVar(o, scope) {
					// a package-level REST client: bind it to the invocation
					add(n.Pos(), edit{off(n.Pos()), off(n.End()), R + ".inv.rest(" + n.Name + ")", true})
				} else if f, ok := globalFlags[n.Name]; ok && isState {
					add(n.Pos(), edit{off(n.Pos()), off(n.End()), R + ".inv.flags." + f, false})
				}
			}
			return true
		})
	}

	for o, fd := range funcs {
		rewriteBody(fd, needs[o])
		if needs[o] {
			add(fd.Name.Pos(), edit{off(fd.Pos()), off(fd.Pos()) + len("func "), "func (" + R + " *" + *typeName + ") ", false})
		}
	}
	initNames := make([]string, len(initFuncs))
	for i, fd := range initFuncs {
		rewriteBody(fd, true)
		base := strings.TrimSuffix(filepath.Base(fileOf(fd.Pos())), ".go")
		name := "init"
		for _, part := range strings.Split(base, "_") {
			name += strings.ToUpper(part[:1]) + part[1:]
		}
		for j := 0; j < i; j++ {
			if initNames[j] == name {
				name += fmt.Sprint(i)
			}
		}
		initNames[i] = name
		add(fd.Name.Pos(), edit{off(fd.Name.Pos()) - len("func "), off(fd.Name.End()), fmt.Sprintf("func (%s *%s) %s", R, *typeName, name), false})
	}
	// initializers of moved vars: rewrite in place first, then extract
	initText := map[*mvar]string{}
	for _, mv := range movedOrder {
		if mv.init != nil {
			rewriteBody(mv.init, true)
		}
	}
	// apply edits per file, extracting init texts
	result := map[string]string{}
	for _, p := range order {
		es := edits[p]
		sort.SliceStable(es, func(i, j int) bool {
			if es[i].start != es[j].start {
				return es[i].start < es[j].start
			}
			return es[i].end > es[j].end
		})
		// drop edits inside deletions
		var kept []edit
		for _, e := range es {
			inside := false
			for _, d := range es {
				if d.del && d != e && d.start <= e.start && e.end <= d.end && !(d.start == e.start && d.end == e.end) {
					inside = true
				}
			}
			if !inside {
				kept = append(kept, e)
			}
		}
		apply := func(a, b int) string {
			var buf bytes.Buffer
			pos := a
			for _, e := range kept {
				if e.start < a || e.end > b {
					continue
				}
				if e.start < pos {
					log.Fatalf("overlapping edit in %s at %d", p, e.start)
				}
				buf.Write(src[p][pos:e.start])
				buf.WriteString(e.text)
				pos = e.end
			}
			buf.Write(src[p][pos:b])
			return buf.String()
		}
		for _, mv := range movedOrder {
			if fileOf(mv.obj.Pos()) == p && mv.init != nil {
				initText[mv] = apply(off(mv.init.Pos()), off(mv.init.End()))
			}
		}
		// spec removal edits: whole GenDecl if all its names move, else spec lines
		type rng struct{ a, b int }
		var removes []rng
		seenDecl := map[*ast.GenDecl]bool{}
		for _, mv := range movedOrder {
			if fileOf(mv.obj.Pos()) != p || seenDecl[mv.decl] {
				continue
			}
			seenDecl[mv.decl] = true
			all := true
			for _, sp := range mv.decl.Specs {
				for _, nm := range sp.(*ast.ValueSpec).Names {
					if moved[info.Defs[nm]] == nil {
						all = false
					}
				}
			}
			if all {
				a := off(mv.decl.Pos())
				if mv.decl.Doc != nil {
					a = off(mv.decl.Doc.Pos())
				}
				removes = append(removes, rng{lineStart(src[p], a), lineEnd(src[p], off(mv.decl.End()))})
			} else {
				for _, sp := range mv.decl.Specs {
					vs := sp.(*ast.ValueSpec)
					allS := true
					for _, nm := range vs.Names {
						if moved[info.Defs[nm]] == nil {
							allS = false
						}
					}
					if !allS {
						continue
					}
					a := off(vs.Pos())
					if vs.Doc != nil {
						a = off(vs.Doc.Pos())
					}
					removes = append(removes, rng{lineStart(src[p], a), lineEnd(src[p], off(vs.End()))})
				}
			}
		}
		for _, r := range removes {
			kept = append(kept, edit{r.a, r.b, "", true})
		}
		sort.SliceStable(kept, func(i, j int) bool { return kept[i].start < kept[j].start })
		// drop edits inside removals
		var final []edit
		for _, e := range kept {
			inside := false
			for _, r := range removes {
				if r.a <= e.start && e.end <= r.b && !(e.start == r.a && e.end == r.b && e.del) {
					inside = true
				}
			}
			if !inside {
				final = append(final, e)
			}
		}
		kept = final
		result[p] = apply(0, len(src[p]))
	}

	// 5. Generate struct + constructor into the root var's file.
	rootFile := ""
	var fields, ctor bytes.Buffer
	qual := func(p *types.Package) string {
		if p == pkg.Types {
			return ""
		}
		return p.Name()
	}
	for _, mv := range movedOrder {
		if mv.obj.Name() == *rootVar {
			rootFile = fileOf(mv.obj.Pos())
		}
		if mv.doc != "" {
			for _, l := range strings.Split(strings.TrimRight(mv.doc, "\n"), "\n") {
				fmt.Fprintf(&fields, "\t// %s\n", l)
			}
		}
		fmt.Fprintf(&fields, "\t%s %s\n", mv.obj.Name(), types.TypeString(mv.obj.Type(), qual))
	}
	for _, dependent := range []bool{false, true} {
		for _, p := range order {
			for _, mv := range movedOrder {
				if fileOf(mv.obj.Pos()) == p && mv.init != nil && mv.dependent == dependent {
					fmt.Fprintf(&ctor, "\t%s.%s = %s\n", R, mv.obj.Name(), initText[mv])
				}
			}
		}
	}
	// Package var initialization runs before init functions, which run in
	// file name order.
	idx := make([]int, len(initFuncs))
	for i := range idx {
		idx[i] = i
	}
	sort.SliceStable(idx, func(a, b int) bool { return fileOf(initFuncs[idx[a]].Pos()) < fileOf(initFuncs[idx[b]].Pos()) })
	for _, i := range idx {
		fmt.Fprintf(&ctor, "\t%s.%s()\n", R, initNames[i])
	}
	gen := fmt.Sprintf(`
// %[1]s holds the commands and flag values of one invocation's %[4]q
// command group, so invocations don't share state (#2070).
type %[1]s struct {
	inv *invocation

%[2]s}

// %[3]s returns the %[4]q command group for inv.
func %[3]s(inv *invocation) *cobra.Command {
	return %[3]ss(inv).%[6]s
}

// %[3]ss builds the %[4]q command group for inv and returns its state.
func %[3]ss(inv *invocation) *%[1]s {
	%[7]s := &%[1]s{inv: inv}
%[5]s	return %[7]s
}
`, *typeName, fields.String(), *ctorName, strings.TrimSuffix(*rootVar, "Cmd"), ctor.String(), *rootVar, R)
	if rootFile == "" {
		log.Fatal("root var not found")
	}
	result[rootFile] += gen
	registerGroup(filepath.Join(*dir, "cmd", "inprocess.go"), *ctorName)

	for _, p := range order {
		out, err := imports.Process(p, []byte(result[p]), nil)
		if err != nil {
			os.WriteFile(p, []byte(result[p]), 0644)
			fmt.Printf("FORMAT ERROR %s: %v\n", p, err)
			continue
		}
		if err := os.WriteFile(p, out, 0644); err != nil {
			log.Fatal(err)
		}
	}
}

func lineStart(b []byte, i int) int {
	for i > 0 && b[i-1] != '\n' {
		i--
	}
	return i
}

func lineEnd(b []byte, i int) int {
	for i < len(b) && b[i] != '\n' {
		i++
	}
	if i < len(b) {
		i++
	}
	return i
}

var legacyRE = regexp.MustCompile(`func (\w+)\(ctx context\.Context, account(?: string)?(?:, \w+ string)?\) \([^)]*\) \{\n\treturn legacyClients\(account\)\.(\w+)\(ctx[,)]`)

func legacyServiceMap(dir string) map[string]string {
	b, err := os.ReadFile(filepath.Join(dir, "internal/gcp/legacy.go"))
	if err != nil {
		log.Fatal(err)
	}
	m := map[string]string{}
	for _, s := range legacyRE.FindAllStringSubmatch(string(b), -1) {
		m[s[1]] = s[2]
	}
	return m
}

// isRestClientVar reports whether o is a package-level *restClient variable.
func isRestClientVar(o types.Object, scope *types.Scope) bool {
	v, ok := o.(*types.Var)
	if !ok || v.Parent() != scope {
		return false
	}
	p, ok := v.Type().(*types.Pointer)
	if !ok {
		return false
	}
	n, ok := p.Elem().(*types.Named)
	return ok && n.Obj().Name() == "restClient"
}

// fixTestFiles rewrites test functions that use the group's moved variables
// or converted functions (groupObjs holds their positions): each such
// function gets its own group state, built for a legacyTestInvocation, and
// its references, and the global flags it uses, go through that state.
// References outside functions are reported for fixing by hand.
func fixTestFiles(tpkg *packages.Package, groupObjs map[token.Pos]bool) {
	fset, info := tpkg.Fset, tpkg.TypesInfo
	scope := tpkg.Types.Scope()
	isGroup := func(id *ast.Ident) bool {
		o := info.Uses[id]
		return o != nil && o.Parent() == scope && groupObjs[o.Pos()]
	}
	for _, f := range tpkg.Syntax {
		fn := fset.Position(f.Pos()).Filename
		if !strings.HasSuffix(fn, "_test.go") {
			continue
		}
		var edits []edit
		off := func(pos token.Pos) int { return fset.Position(pos).Offset }
		for _, d := range f.Decls {
			fd, ok := d.(*ast.FuncDecl)
			if !ok || fd.Body == nil {
				if gd, ok := d.(*ast.GenDecl); ok {
					ast.Inspect(gd, func(n ast.Node) bool {
						if id, ok := n.(*ast.Ident); ok && isGroup(id) {
							fmt.Printf("MANUAL %s used outside a function at %s\n", id.Name, fset.Position(id.Pos()))
						}
						return true
					})
				}
				continue
			}
			var refs []*ast.Ident
			defined := map[string]bool{}
			ast.Inspect(fd, func(n ast.Node) bool {
				if id, ok := n.(*ast.Ident); ok {
					if isGroup(id) {
						refs = append(refs, id)
					}
					if info.Defs[id] != nil {
						defined[id.Name] = true
					}
				}
				return true
			})
			if len(refs) == 0 {
				continue
			}
			g := "grp"
			for i := 2; defined[g]; i++ {
				g = fmt.Sprintf("grp%d", i)
			}
			for _, id := range refs {
				edits = append(edits, edit{off(id.Pos()), off(id.Pos()), g + ".", false})
			}
			ast.Inspect(fd.Body, func(n ast.Node) bool {
				if id, ok := n.(*ast.Ident); ok {
					if o := info.Uses[id]; o != nil && o.Parent() == scope {
						if fl, ok := globalFlags[id.Name]; ok {
							edits = append(edits, edit{off(id.Pos()), off(id.End()), g + ".inv.flags." + fl, false})
						}
					}
				}
				return true
			})
			edits = append(edits, edit{off(fd.Body.Lbrace) + 1, off(fd.Body.Lbrace) + 1,
				fmt.Sprintf("\n\t%s := %ss(legacyTestInvocation())", g, *ctorName), false})
		}
		if len(edits) == 0 {
			continue
		}
		src, err := os.ReadFile(fn)
		if err != nil {
			log.Fatal(err)
		}
		sort.SliceStable(edits, func(i, j int) bool { return edits[i].start > edits[j].start })
		for _, e := range edits {
			src = append(src[:e.start], append([]byte(e.text), src[e.end:]...)...)
		}
		out, err := imports.Process(fn, src, nil)
		if err != nil {
			out = src
			fmt.Printf("FORMAT ERROR %s: %v\n", fn, err)
		}
		if err := os.WriteFile(fn, out, 0644); err != nil {
			log.Fatal(err)
		}
		fmt.Printf("FIXED %s\n", fn)
	}
}

// registerGroup adds ctor(inv) to the list in migratedGroups in path.
func registerGroup(path, ctor string) {
	b, err := os.ReadFile(path)
	if err != nil {
		log.Fatal(err)
	}
	s := string(b)
	i := strings.Index(s, "func migratedGroups(")
	if i < 0 {
		log.Fatal("migratedGroups not found")
	}
	j := strings.Index(s[i:], "\n\t}\n}")
	if j < 0 {
		log.Fatal("end of migratedGroups list not found")
	}
	j += i + 1
	s = s[:j] + "\t\t" + ctor + "(inv),\n" + s[j:]
	if err := os.WriteFile(path, []byte(s), 0644); err != nil {
		log.Fatal(err)
	}
}
