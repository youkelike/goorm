package orm

import (
	"context"

	"gitee.com/youkelike/orm/internal/errs"
	"gitee.com/youkelike/orm/model"
)

type Upsert struct {
	assigns         []Assignable
	conflictColumns []string
}

type UpsertBuilder[T any] struct {
	i               *Inserter[T]
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
	builder
	columns []string
	values  []*T
	db      *DB
	upsert  *Upsert
}

func NewInserter[T any](db *DB) *Inserter[T] {
	return &Inserter[T]{
		builder: builder{
			dialect: db.dialect,
			quoter:  db.dialect.quoter(),
		},
		db: db,
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
	m, err := i.db.r.Get(i.values[0])
	if err != nil {
		return nil, err
	}
	i.model = m

	i.sb.WriteString(m.TableName)
	// i.quote(m.TableName)
	i.sb.WriteString(" (")

	fields := m.Fields
	if len(i.columns) > 0 {
		fields = make([]*model.Field, 0, len(i.columns))
		for _, fd := range i.columns {
			fdMeta, ok := m.FieldMap[fd]
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
	for r, v := range i.values {
		if r > 0 {
			i.sb.WriteString(",")
		}
		i.sb.WriteString("(")

		// 结构体读取字段值只能用点号操作符，想要通过结构体字段名来读取结构体字段的值，只能用 reflect 或者 unsafe
		val := i.db.creator(i.model, v)

		for c, field := range fields {
			if c > 0 {
				i.sb.WriteString(",")
			}
			i.sb.WriteString("?")

			// val := reflect.ValueOf(i.values[r]).Elem().FieldByName(field.GoName).Interface()
			arg, err := val.Field(field.GoName)
			if err != nil {
				return nil, err
			}

			i.addArgs(arg)
		}
		i.sb.WriteString(")")
	}

	if i.upsert != nil {
		err := i.dialect.buildUpsert(&i.builder, i.upsert)
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
	q, err := i.Build()
	if err != nil {
		return Result{
			err: err,
		}
	}

	res, err := i.db.db.ExecContext(ctx, q.SQL, q.Args...)
	return Result{
		res: res,
		err: err,
	}
}
