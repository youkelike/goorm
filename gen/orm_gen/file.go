package main

import (
	"go/ast"
)

type File struct {
	Package string
	Imports []string
	Types   []Type
}

// 用于拿到 package、import、type 定义
type FileVisitor struct {
	Package string
	Imports []string
	Types   []*TypeVisitor
}

func (f *FileVisitor) Get() *File {
	types := make([]Type, 0, len(f.Types))
	for _, typ := range f.Types {
		types = append(types, Type{
			Name:   typ.name,
			Fields: typ.fields,
		})
	}
	return &File{
		Package: f.Package,
		Imports: f.Imports,
		Types:   types,
	}
}

func (f *FileVisitor) Visit(node ast.Node) ast.Visitor {
	switch n := node.(type) {
	case *ast.File:
		f.Package = n.Name.String()
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

// 用于拿到结构体定义的字段
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
