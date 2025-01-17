package ast

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPrintVisitor(t *testing.T) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "", `
	package ast

	import (
		"fmt"
		"go/ast"
		"reflect"
	)
	
	type PrintVisitor struct {
	}
	
	func (p *PrintVisitor) Visit(node ast.Node) ast.Visitor {
		if node == nil {
			fmt.Println("nil")
			return p
		}
		typ := reflect.TypeOf(node)
		val := reflect.ValueOf(node)
		for typ.Kind() == reflect.Ptr {
			typ = typ.Elem()
			val = val.Elem()
		}
		fmt.Printf("val: %+v, typ: %s\n", val.Interface(), typ.Name())
		return p
	}`, parser.ParseComments)
	require.NoError(t, err)
	v := &PrintVisitor{}
	ast.Walk(v, f)

}
