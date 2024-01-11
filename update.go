package orm

import (
	"context"
	"database/sql"
	"reflect"

	"gitee.com/youkelike/orm/internal/errs"
	"gitee.com/youkelike/orm/model"
)

type Updater[T any] struct {
	builder
	sess Session

	table string

	value   *T
	updates []Column
	where   []Predicate
}

func NewUpdater[T any](sess Session) *Updater[T] {
	c := sess.getCore()
	return &Updater[T]{
		builder: builder{
			core:   c,
			quoter: c.dialect.quoter(),
		},
		sess: sess,
	}
}

func (d *Updater[T]) Build() (*Query, error) {
	var err error
	d.model, err = d.r.Register(new(T))
	if err != nil {
		return nil, err
	}

	d.sb.WriteString("UPDATE ")
	if d.table == "" {
		d.sb.WriteString(d.model.TableName)
	} else {
		d.sb.WriteString(d.table)
	}

	d.sb.WriteString(" SET ")
	var fdList []*model.Field
	if len(d.updates) > 0 { // 更新指定列
		for _, col := range d.updates {
			fd, ok := d.model.FieldMap[col.name]
			if !ok {
				return nil, errs.NewUnknownField(col.name)
			}
			fdList = append(fdList, fd)
		}
	} else { // 更新所有列
		fdList = d.model.Fields
	}
	if d.value == nil {
		return nil, errs.NewUnknownUpdateValue()
	}
	for i, fd := range fdList {
		if i > 0 {
			d.sb.WriteString(",")
		}
		val := reflect.ValueOf(d.value).Elem().FieldByName(fd.GoName).Interface()
		d.sb.WriteString(fd.ColName)
		d.sb.WriteString("=?")
		// if v, ok := val.(*sql.NullString); ok {
		// 	d.addArgs(v.String)
		// } else {
		// 	d.addArgs(val)
		// }
		d.addArgs(val)
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

func (d *Updater[T]) buildExpression(expr Expression) error {
	switch p := expr.(type) {
	case nil:
	case Predicate:
		_, ok := p.left.(Predicate)
		if ok {
			d.sb.WriteString("(")
		}
		if err := d.buildExpression(p.left); err != nil {
			return err
		}
		if ok {
			d.sb.WriteString(")")
		}

		if p.op == opNot || p.op == opAnd || p.op == opOr {
			d.sb.WriteString(" ")
		}
		d.sb.WriteString(p.op.String())
		if p.op == opNot || p.op == opAnd || p.op == opOr {
			d.sb.WriteString(" ")
		}

		_, ok = p.right.(Predicate)
		if ok {
			d.sb.WriteString("(")
		}
		if err := d.buildExpression(p.right); err != nil {
			return err
		}
		if ok {
			d.sb.WriteString(")")
		}
	case Column:
		p.alias = ""
		err := d.buildColumn(p)
		if err != nil {
			return err
		}
	case value:
		d.sb.WriteString("?")
		d.addArgs(p.val)
	case RawExpr:
		d.sb.WriteString("(")
		d.sb.WriteString(p.raw)
		d.sb.WriteString(")")
		d.addArgs(p.args...)
	default:
		return errs.NewUnsupportExpression(expr)
	}
	return nil
}

func (d *Updater[T]) From(table string) *Updater[T] {
	d.table = table
	return d
}

func (d *Updater[T]) Value(val *T) *Updater[T] {
	d.value = val
	return d
}

func (d *Updater[T]) Updates(columns ...Column) *Updater[T] {
	d.updates = columns
	return d
}

func (d *Updater[T]) Where(ps ...Predicate) *Updater[T] {
	d.where = ps
	return d
}

func (d *Updater[T]) Exec(ctx context.Context) Result {
	var err error
	d.model, err = d.r.Get(new(T))
	if err != nil {
		return Result{
			err: err,
		}
	}

	res := exec(ctx, d.sess, d.core, &QueryContext{
		Type:    "UPDATE",
		Builder: d,
		Model:   d.model,
	})

	var sqlRes sql.Result
	if res.Result != nil {
		sqlRes = res.Result.(sql.Result)
	}

	return Result{
		err: res.Err,
		res: sqlRes,
	}
}