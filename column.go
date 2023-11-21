package orm

type Column struct {
	table TableReference
	name  string
	alias string
	order string
}

// C("name")
func C(name string) Column {
	return Column{name: name}
}

func (c Column) assign() {}

func (c Column) As(alias string) Column {
	return Column{
		name:  c.name,
		alias: alias,
		table: c.table,
	}
}

func (c Column) Desc() Column {
	return Column{
		name:  c.name,
		order: "DESC",
	}
}
func (c Column) Asc() Column {
	return Column{
		name:  c.name,
		order: "ASC",
	}
}

// C("name").Eq("Tom")
func (c Column) Eq(arg any) Predicate {
	return Predicate{
		left:  c,
		op:    opEq,
		right: valueOf(arg),
	}
}
func (c Column) Gt(arg any) Predicate {
	return Predicate{
		left:  c,
		op:    opGt,
		right: valueOf(arg),
	}
}
func (c Column) Lt(arg any) Predicate {
	return Predicate{
		left:  c,
		op:    opLt,
		right: valueOf(arg),
	}
}

func valueOf(arg any) Expression {
	switch val := arg.(type) {
	case Expression:
		return val
	default:
		return value{val: val}
	}
}

func (c Column) expr() {}

func (c Column) selectable() {}
