package main

import (
	_ "embed"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"text/template"
)

// func main() {
// 	f, err := os.OpenFile("testdata/user.gen.go")
// 	if err != nil {
// 		panic(err)
// 	}
// 	gen(f, "testdata/user.go")
// }

//go:embed tpl.gohtml
var genOrm string

func gen(w io.Writer, srcFile string) error {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, srcFile, nil, parser.ParseComments)
	if err != nil {
		return err
	}
	s := &SingleFileEntryVisitor{}
	ast.Walk(s, f)
	file := s.Get()
	fmt.Println(file.Package)

	tpl := template.New("gen-orm")
	tpl, err = tpl.Parse(genOrm)
	if err != nil {
		return err
	}
	return tpl.Execute(w, Data{
		File: file,
		Ops:  []string{"LT", "GT", "EQ"},
	})
}

type Data struct {
	*File
	Ops []string
}
