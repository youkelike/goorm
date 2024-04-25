package main

import "go/ast"

type FileEntryVisitor struct {
	Package string
	Imports []string
	Types   []Type
}

// 一个方法完成所有节点查找
func (f *FileEntryVisitor) Visit(node ast.Node) ast.Visitor {
	switch n := node.(type) {
	case *ast.File:
		f.Package = n.Name.String()
	case *ast.ImportSpec:
		path := n.Path.Value
		if n.Name != nil && n.Name.String() != "" {
			path = n.Name.String() + " " + path
		}
		f.Imports = append(f.Imports, path)
	case *ast.TypeSpec:
		if s, ok := n.Type.(*ast.StructType); ok {
			typ := Type{Name: n.Name.String(), Fields: make([]Field, 0, len(s.Fields.List))}
			var fdTyp string
			for _, fd := range s.Fields.List {
				switch nt := fd.Type.(type) {
				case *ast.Ident:
					fdTyp = nt.String()
				case *ast.StarExpr:
					switch xt := nt.X.(type) {
					case *ast.Ident:
						fdTyp = "*" + xt.String()
					case *ast.SelectorExpr:
						fdTyp = "*" + xt.X.(*ast.Ident).String() + "." + xt.Sel.String()
					}
				case *ast.ArrayType:
					fdTyp = "[]byte"
				default:
					panic("不支持的类型")
				}
				for _, name := range fd.Names {
					typ.Fields = append(typ.Fields, Field{
						Name: name.String(),
						Type: fdTyp,
					})
				}
			}
			f.Types = append(f.Types, typ)
		}
	}
	return f
}

func (f *FileEntryVisitor) Get() *File {
	return &File{
		Package: f.Package,
		Imports: f.Imports,
		Types:   f.Types,
	}
}
