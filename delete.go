package orm

type Delete[T any] struct {
	builder
	table string

	where []Predicate
}

func (d *Delete[T]) Build() (*Query, error) {
	var err error
	d.model, err = parseModel(new(T))
	if err != nil {
		return nil, err
	}

	d.sb.WriteString("DELETE FROM ")
	if d.table == "" {
		d.sb.WriteString(d.model.tableName)
	} else {
		d.sb.WriteString(d.table)
	}

	if d.where != nil {
		err := d.buildWhere(d.where)
		if err != nil {
			return nil, err
		}
	}
	d.sb.WriteString(";")

	return &Query{
		SQL:  d.sb.String(),
		Args: d.args,
	}, nil
}

func (d *Delete[T]) From(table string) *Delete[T] {
	d.table = table
	return d
}

func (d *Delete[T]) Where(ps ...Predicate) *Delete[T] {
	d.where = ps
	return d
}
