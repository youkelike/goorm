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

	// 多个遍历器合作
	s := &FileVisitor{}
	// 一个遍历器完成所有节点查找
	// s := &FileEntryVisitor{}

	// Walk 相当于多叉树中的 dfs，还是前序遍历（先 Visit 当前节点），
	// 首先，会调用 ss := s.Visit(f)，这个调用返回的 ss 也是一个 Visitor，可以是新的，也可以是原来的，
	// 如果 ss 为空，递归就返回了，
	// 否则，会用返回的 ss 来 Walk 所有 f 的子节点（f 有各种不同类型，子节点都不一样），类似这样：
	// for 子节点 in f {
	//     ast.Walk(ss, 子节点)
	// }
	//
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
