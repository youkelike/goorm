package orm

import (
	"context"
	"database/sql"
)

type Deletor[T any] struct {
	builder
	sess Session

	table string

	where []Predicate
}

func NewDeletor[T any](sess Session) *Deletor[T] {
	c := sess.getCore()
	return &Deletor[T]{
		builder: builder{
			r:      c.r,
			quoter: c.dialect.quoter(),
		},
		sess: sess,
	}
}

func (d *Deletor[T]) Build() (*Query, error) {
	if d.model == nil {
		var err error
		d.model, err = d.r.Get(new(T))
		if err != nil {
			return nil, err
		}
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

		if err := d.buildPredicate(p); err != nil {
			return nil, err
		}
	}
	d.sb.WriteString(";")

	return &Query{
		SQL:  d.sb.String(),
		Args: d.args,
	}, nil
}

func (d *Deletor[T]) From(table string) *Deletor[T] {
	d.table = table
	return d
}

func (d *Deletor[T]) Where(ps ...Predicate) *Deletor[T] {
	d.where = ps
	return d
}

func (d *Deletor[T]) Exec(ctx context.Context) Result {
	var err error
	d.model, err = d.r.Get(new(T))
	if err != nil {
		return Result{
			err: err,
		}
	}

	res := exec(ctx, &QueryContext{
		Type:    "DELETE",
		Builder: d,
		Model:   d.model,
		Sess:    d.sess,
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
