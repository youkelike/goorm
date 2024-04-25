package orm

import (
	"context"
	"database/sql"
	"reflect"

	"gitee.com/youkelike/orm/internal/errs"
	"gitee.com/youkelike/orm/model"
)

type UpsertBuilder[T any] struct {
	i               *Inserter[T]
	conflictColumns []string
}

type Upsert struct {
	assigns         []Assignable
	conflictColumns []string
}

func (o *UpsertBuilder[T]) Update(assigns ...Assignable) *Inserter[T] {
	o.i.upsert = &Upsert{assigns: assigns, conflictColumns: o.conflictColumns}
	return o.i
}

func (o *UpsertBuilder[T]) ConflictColumns(cols ...string) *UpsertBuilder[T] {
	o.conflictColumns = cols
	return o
}

type Assignable interface {
	assign()
}

type Inserter[T any] struct {
	// db      *DB
	builder
	columns []string
	values  []*T
	upsert  *Upsert

	sess Session
}

func NewInserter[T any](sess Session) *Inserter[T] {
	c := sess.getCore()
	return &Inserter[T]{
		builder: builder{
			r:      c.r,
			quoter: c.dialect.quoter(),
		},
		sess: sess,
	}
}

func (i *Inserter[T]) Upsert() *UpsertBuilder[T] {
	return &UpsertBuilder[T]{
		i: i,
	}
}

func (i *Inserter[T]) Columns(cols ...string) *Inserter[T] {
	i.columns = cols
	return i
}

func (i *Inserter[T]) Values(vals ...*T) *Inserter[T] {
	i.values = vals
	return i
}

func (i *Inserter[T]) Build() (*Query, error) {
	if len(i.values) == 0 {
		return nil, errs.ErrInsertZeroRow
	}

	i.sb.WriteString("INSERT INTO ")
	if i.model == nil {
		var err error
		i.model, err = i.r.Get(i.values[0])
		if err != nil {
			return nil, err
		}
	}

	i.sb.WriteString(i.model.TableName)
	// i.quote(m.TableName)
	i.sb.WriteString(" (")

	fields := i.model.Fields
	if len(i.columns) > 0 {
		fields = make([]*model.Field, 0, len(i.columns))
		for _, fd := range i.columns {
			fdMeta, ok := i.model.FieldMap[fd]
			if !ok {
				return nil, errs.NewUnknownField(fd)
			}
			fields = append(fields, fdMeta)
		}
	}

	for idx, field := range fields {
		if idx > 0 {
			i.sb.WriteString(",")
		}
		i.sb.WriteString(field.ColName)
		// i.quote(field.ColName)
	}
	i.sb.WriteString(")")
	i.sb.WriteString(" VALUES ")
	i.args = make([]any, 0, len(i.values)*len(fields))
	for r := range i.values {
		if r > 0 {
			i.sb.WriteString(",")
		}
		i.sb.WriteString("(")
		for c, field := range fields {
			if c > 0 {
				i.sb.WriteString(",")
			}
			i.sb.WriteString("?")
			val := reflect.ValueOf(i.values[r]).Elem().FieldByName(field.GoName).Interface()
			i.addArgs(val)
		}
		i.sb.WriteString(")")
	}

	if i.upsert != nil {
		err := i.sess.getCore().dialect.buildUpsert(&i.builder, i.upsert)
		if err != nil {
			return nil, err
		}
	}

	i.sb.WriteString(";")
	return &Query{
		SQL:  i.sb.String(),
		Args: i.args,
	}, nil
}

func (i *Inserter[T]) Exec(ctx context.Context) Result {
	var err error
	i.model, err = i.r.Get(new(T))
	if err != nil {
		return Result{
			err: err,
		}
	}

	res := exec(ctx, &QueryContext{
		Type:    "INSERT",
		Builder: i,
		Model:   i.model,
		Sess:    i.sess,
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

// var _ Handler = (&Inserter[any]{}).execHandler

// func (i *Inserter[T]) execHandler(ctx context.Context, qc *QueryContext) *QueryResult {
// 	q, err := i.Build()
// 	if err != nil {
// 		return &QueryResult{
// 			Err: err,
// 		}
// 	}

// 	res, err := i.sess.execContext(ctx, q.SQL, q.Args...)
// 	return &QueryResult{
// 		Err:    err,
// 		Result: res,
// 	}
// }
