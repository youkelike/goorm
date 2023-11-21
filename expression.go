package orm

type Expression interface {
	expr()
}

type RawExpr struct {
	raw  string
	args []any
}

func Raw(expr string, args ...any) RawExpr {
	return RawExpr{
		raw:  expr,
		args: args,
	}
}

// 标识在 select 子句可以用
func (r RawExpr) selectable() {}

// 标识在 where 子句可以用
func (r RawExpr) expr() {}

func (r RawExpr) AsPredicate() Predicate {
	return Predicate{
		left: r,
	}
}
