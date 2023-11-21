package main

import (
	"go/ast"
)

type SingleFileEntryVisitor struct {
	file *FileVisitor
}

func (s *SingleFileEntryVisitor) Get() *File {
	types := make([]Type, 0, len(s.file.Types))
	for _, typ := range s.file.Types {
		types = append(types, Type{
			Name:   typ.name,
			Fields: typ.fields,
		})
	}
	return &File{
		Package: s.file.Package,
		Imports: s.file.Imports,
		Types:   types,
	}
}

// 要先检查传入的 node 是否是要处理的类型，符合要求再处理
// 对于返回值，有 3 中情况：
// 1、返回 nil 结束遍历
// 2、返回自己，后续会继续遍历 node 的子节点
// 3、返回一个其他的 visitor，会用它来遍历 node 的子节点
func (s *SingleFileEntryVisitor) Visit(node ast.Node) ast.Visitor {
	fn, ok := node.(*ast.File)
	if !ok {
		// 意思是没找到，继续在子节点里找
		return s
	}

	// 意思是用一个新的 visitor 去遍历子节点
	s.file = &FileVisitor{
		Package: fn.Name.String(),
	}
	return s.file
}

type File struct {
	Package string
	Imports []string
	Types   []Type
}

type FileVisitor struct {
	Package string
	Imports []string
	Types   []*TypeVisitor
}

func (f *FileVisitor) Visit(node ast.Node) ast.Visitor {
	switch n := node.(type) {
	case *ast.ImportSpec:
		path := n.Path.Value
		if n.Name != nil && n.Name.String() != "" {
			path = n.Name.String() + " " + path
		}
		f.Imports = append(f.Imports, path)
		// case *ast.GenDecl:
		// 	if n.Tok == token.IMPORT {
		// 		for _, spec := range n.Specs {
		// 			f.Imports = append(f.Imports, spec.(*ast.ImportSpec).Path.Value)
		// 		}
		// 	}
	case *ast.TypeSpec:
		v := &TypeVisitor{name: n.Name.String()}
		f.Types = append(f.Types, v)
		return v
	}
	return f
}

type TypeVisitor struct {
	name   string
	fields []Field
}

func (t *TypeVisitor) Visit(node ast.Node) ast.Visitor {
	n, ok := node.(*ast.Field)
	if !ok {
		return t
	}
	var typ string
	switch nt := n.Type.(type) {
	case *ast.Ident:
		typ = nt.String()
	case *ast.StarExpr:
		switch xt := nt.X.(type) {
		case *ast.Ident:
			typ = "*" + xt.String()
		case *ast.SelectorExpr:
			typ = "*" + xt.X.(*ast.Ident).String() + "." + xt.Sel.String()
		}
	case *ast.ArrayType:
		typ = "[]byte"
	default:
		panic("不支持的类型")
	}
	for _, name := range n.Names {
		t.fields = append(t.fields, Field{
			Name: name.String(),
			Type: typ,
		})
	}
	return t
}

type Type struct {
	Name   string
	Fields []Field
}

type Field struct {
	Name string
	Type string
}
