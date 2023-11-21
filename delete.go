package orm

import (
	"strings"

	"gitee.com/youkelike/orm/internal/errs"
	model "gitee.com/youkelike/orm/model"
)

type Deletor[T any] struct {
	table string
	model *model.Model
	where []Predicate
	args  []any
	sb    *strings.Builder
	db    *DB
}

func NewDeletor[T any](db *DB) *Deletor[T] {
	return &Deletor[T]{
		sb: &strings.Builder{},
		db: db,
	}
}

func (d *Deletor[T]) Build() (*Query, error) {
	var err error
	d.model, err = d.db.r.Register(new(T))
	if err != nil {
		return nil, err
	}

	d.sb.WriteString("DELETE FROM ")
	if d.table == "" {
		d.sb.WriteString(d.model.TableName)
	} else {
		d.sb.WriteString(d.table)
	}

	if d.where != nil {
		d.sb.WriteString(" WHERE ")
		p := d.where[0]
		for i := 1; i < len(d.where); i++ {
			p = p.And(d.where[i])
		}

		if err := d.buildExpression(p); err != nil {
			return nil, err
		}
	}
	d.sb.WriteString(";")

	return &Query{
		SQL:  d.sb.String(),
		Args: d.args,
	}, nil
}

func (d *Deletor[T]) buildExpression(expr Expression) error {
	switch exp := expr.(type) {
	case Predicate:
		_, ok := exp.left.(Predicate)
		if ok {
			d.sb.WriteString("(")
		}
		if err := d.buildExpression(exp.left); err != nil {
			return err
		}
		if ok {
			d.sb.WriteString(")")
		}

		if exp.op == opAnd || exp.op == opOr || exp.op == opNot {
			d.sb.WriteString(" ")
		}
		d.sb.WriteString(exp.op.String())
		if exp.op == opAnd || exp.op == opOr || exp.op == opNot {
			d.sb.WriteString(" ")
		}

		_, ok = exp.right.(Predicate)
		if ok {
			d.sb.WriteString("(")
		}
		if err := d.buildExpression(exp.right); err != nil {
			return err
		}
		if ok {
			d.sb.WriteString(")")
		}
	case Column:
		fd, ok := d.model.FieldMap[exp.name]
		if !ok {
			return errs.NewUnknownField(exp.name)
		}
		d.sb.WriteString(fd.ColName)
	case value:
		d.sb.WriteString("?")
		d.addArg(exp.val)
	case nil:
	default:
		return errs.NewUnsupportExpression(exp)
	}
	return nil
}

func (d *Deletor[T]) From(table string) *Deletor[T] {
	d.table = table
	return d
}

func (d *Deletor[T]) Where(ps ...Predicate) *Deletor[T] {
	d.where = ps
	return d
}

func (d *Deletor[T]) addArg(val any) *Deletor[T] {
	if d.args == nil {
		d.args = make([]any, 0, 8)
	}
	d.args = append(d.args, val)
	return d
}
