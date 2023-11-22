package orm

type Aggregate struct {
	// 函数名
	fn string
	// 列名
	arg Column
	// 别名
	alias string
}

func (a Aggregate) As(alias string) Aggregate {
	return Aggregate{
		fn:    a.fn,
		arg:   a.arg,
		alias: alias,
	}
}

func (a Aggregate) selectable() {}

func Avg(col Column) Aggregate {
	return Aggregate{
		fn:  "AVG",
		arg: col,
	}
}

func Sum(col Column) Aggregate {
	return Aggregate{
		fn:  "SUM",
		arg: col,
	}
}

func Count(col Column) Aggregate {
	return Aggregate{
		fn:  "COUNT",
		arg: col,
	}
}

func Max(col Column) Aggregate {
	return Aggregate{
		fn:  "MAX",
		arg: col,
	}
}

func Min(col Column) Aggregate {
	return Aggregate{
		fn:  "MIN",
		arg: col,
	}
}
