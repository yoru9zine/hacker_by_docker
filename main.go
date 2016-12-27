package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"strings"
)

func unquote(s string) string {
	s = strings.TrimPrefix(s, `"`)
	s = strings.TrimSuffix(s, `"`)
	return s
}

type hacker struct {
	Name    string
	Comment string
}

type hackerFinder struct {
	variablePos token.Pos
	names       []ast.Expr
	comments    []*ast.CommentGroup
}

func (hf *hackerFinder) Visit(node ast.Node) ast.Visitor {
	switch v := node.(type) {
	case *ast.ValueSpec:
		name := v.Names[0].Name
		if name == "right" {
			hf.variablePos = v.Names[0].NamePos
			hf.names = v.Values[0].(*ast.CompositeLit).Elts
		}
	case *ast.File:
		hf.comments = v.Comments
	}
	return hf
}

func getHackers() []*hacker {
	f, err := os.Open(fmt.Sprintf("%s/src/github.com/docker/docker/pkg/namesgenerator/names-generator.go", os.Getenv("GOPATH")))
	if err != nil {
		log.Fatalf("failed to open names-generator.go")
	}
	fset := token.NewFileSet()
	ff, err := parser.ParseFile(fset, "", f, parser.ParseComments)
	if err != nil {
		log.Fatalf("failed to names-generator.go")
	}
	hf := &hackerFinder{}
	ast.Walk(hf, ff)

	hackers := []*hacker{}
	lastLine := fset.File(ff.Package).Line(hf.variablePos)
	for _, n := range hf.names {
		l := fset.File(ff.Package).Line(n.Pos())
		var commentLines []string
		for _, cg := range hf.comments {
			for _, c := range cg.List {
				cl := fset.File(ff.Package).Line(c.Slash)
				if cl > lastLine && cl < l {
					commentLines = append(commentLines, strings.TrimPrefix(c.Text, `// `))
				}
			}
		}
		hackers = append(hackers, &hacker{Name: unquote(n.(*ast.BasicLit).Value), Comment: strings.Join(commentLines, "\n")})
		lastLine = l
	}
	return hackers
}

var (
	list = flag.Bool("l", false, "list all hackers")
)

func main() {
	flag.Parse()
	hackers := getHackers()
	if *list {
		for _, h := range hackers {
			fmt.Println(h.Name)
		}
		return
	}
	if len(flag.Args()) == 0 {
		return
	}
	for _, name := range flag.Args() {
		for _, h := range hackers {
			if strings.HasSuffix(name, h.Name) {
				println(h.Comment)
			}
		}
	}
}
