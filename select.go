package orm

import (
	"context"
	"reflect"
	"strconv"
	"strings"

	"gitee.com/youkelike/orm/internal/errs"
)

type Selectable interface {
	selectable()
}

type Selector[T any] struct {
	builder
	sess Session

	// select 子句
	columns []Selectable
	// from 子句
	table TableReference
	// where 子句
	// where 的数据类型只能是 Predicate，不能是 Expression，因为 Column、Value 都不能直接放到 where 中，
	// 而且还要处理多个 Predicate 之间的组合问题，也就是多个 where 条件的组合，
	where []Predicate
	// group 子句
	groupBy []Column
	// having 子句
	having []Predicate
	// order 子句
	orderBy []Column
	// offset 子句
	offset int
	// limit 子句
	limit int
}

func NewSelector[T any](sess Session) *Selector[T] {
	c := sess.getCore()
	return &Selector[T]{
		builder: builder{
			r:      c.r,
			quoter: c.dialect.quoter(),
		},
		sess: sess,
	}
}

func (s *Selector[T]) Build() (*Query, error) {
	if s.model == nil {
		var err error
		s.model, err = s.r.Get(new(T))
		if err != nil {
			return nil, err
		}
	}

	s.sb.WriteString("SELECT ")
	err := s.buildColumns()
	if err != nil {
		return nil, err
	}
	s.sb.WriteString(" FROM ")

	err = s.buildTable(s.table)
	if err != nil {
		return nil, err
	}
	// if s.table == "" {
	// 	s.sb.WriteString(s.model.TableName)
	// } else {
	// 	s.sb.WriteString(s.table)
	// }

	if len(s.where) > 0 {
		s.sb.WriteString(" WHERE ")

		// 先把切片形式的条件组装成链表
		p := s.where[0]
		for i := 1; i < len(s.where); i++ {
			p = p.And(s.where[i])
		}
		// 遍历链表
		if err := s.buildPredicate(p); err != nil {
			return nil, err
		}
	}

	if len(s.groupBy) > 0 {
		s.sb.WriteString(" GROUP BY ")
		for i, col := range s.groupBy {
			if i > 0 {
				s.sb.WriteString(",")
			}
			err := s.buildColumn(col)
			if err != nil {
				return nil, err
			}
		}
	}

	if len(s.having) > 0 {
		if len(s.groupBy) == 0 {
			return nil, errs.ErrNoGroupUseHaving
		}
		s.sb.WriteString(" HAVING ")
		// 先把切片形式的条件组装成链表
		p := s.having[0]
		for i := 1; i < len(s.having); i++ {
			p = p.And(s.having[i])
		}
		// 遍历链表
		if err := s.buildPredicate(p); err != nil {
			return nil, err
		}
	}

	if len(s.orderBy) > 0 {
		s.sb.WriteString(" ORDER BY ")
		for i, col := range s.orderBy {
			if col.order == "" {
				return nil, errs.ErrNoOrderByVerb
			}
			if i > 0 {
				s.sb.WriteString(",")
			}
			err := s.buildColumn(col)
			if err != nil {
				return nil, err
			}
		}
	}

	if s.offset > 0 {
		s.sb.WriteString(" OFFSET ")
		s.sb.WriteString(strconv.Itoa(s.offset))
		if s.limit > 0 {
			s.sb.WriteString(",")
		}
	}
	if s.limit > 0 {
		s.sb.WriteString(" LIMIT ")
		s.sb.WriteString(strconv.Itoa(s.limit))
	}

	s.sb.WriteString(";")

	return &Query{
		SQL:  s.sb.String(),
		Args: s.args,
	}, nil
}

func (s *Selector[T]) buildTable(table TableReference) error {
	switch t := table.(type) {
	case nil:
		s.sb.WriteString(s.model.TableName)
	case Table:
		m, err := s.r.Get(t.entity)
		if err != nil {
			return err
		}
		s.sb.WriteString(m.TableName)
		if t.alias != "" {
			s.sb.WriteString(" AS ")
			s.sb.WriteString(t.alias)
		}
	case Join:
		s.sb.WriteString("(")
		err := s.buildTable(t.left)
		if err != nil {
			return err
		}
		s.sb.WriteString(" " + t.typ + " ")
		err = s.buildTable(t.right)
		if err != nil {
			return err
		}

		if len(t.using) > 0 {
			s.sb.WriteString(" USING (")
			for i, col := range t.using {
				if i > 0 {
					s.sb.WriteString(",")
				}
				err := s.buildColumn(Column{name: col})
				if err != nil {
					return err
				}
			}
			s.sb.WriteString(")")
		}

		if len(t.on) > 0 {
			s.sb.WriteString(" ON ")
			p := t.on[0]
			for i := 1; i < len(t.on); i++ {
				p = p.And(t.on[i])
			}
			if err := s.buildPredicate(p); err != nil {
				return err
			}
		}
		s.sb.WriteString(")")
	case Subquery[T]:
		res, err := t.builder.Build()
		if err != nil {
			return err
		}

		s.addArgs(res.Args...)
		s.sb.WriteString("(")
		s.sb.WriteString(strings.Trim(res.SQL, ";"))
		s.sb.WriteString(")")
		s.sb.WriteString(" AS ")
		s.sb.WriteString(t.as)
	default:
		return errs.NewUnsupportTable(table)
	}
	return nil
}

func (s *Selector[T]) buildColumns() error {
	if len(s.columns) == 0 {
		s.sb.WriteString("*")
		return nil
	}

	for i, col := range s.columns {
		if i > 0 {
			s.sb.WriteString(",")
		}
		switch c := col.(type) {
		case Column:
			err := s.buildColumn(col.(Column))
			if err != nil {
				return err
			}
		case Aggregate:
			s.sb.WriteString(c.fn)
			s.sb.WriteString("(")
			err := s.buildColumn(c.arg)
			if err != nil {
				return err
			}
			s.sb.WriteString(")")
			if c.alias != "" {
				s.sb.WriteString(" AS ")
				s.sb.WriteString(c.alias)
			}
		case RawExpr:
			s.sb.WriteString(c.raw)
			s.addArgs(c.args...)
		}
	}
	return nil
}

func (s *Selector[T]) Select(cols ...Selectable) *Selector[T] {
	s.columns = cols
	return s
}

func (s *Selector[T]) From(table TableReference) *Selector[T] {
	s.table = table
	return s
}

func (s *Selector[T]) Where(ps ...Predicate) *Selector[T] {
	s.where = ps
	return s
}

func (s *Selector[T]) GroupBy(cols ...Column) *Selector[T] {
	s.groupBy = cols
	return s
}

func (s *Selector[T]) Having(ps ...Predicate) *Selector[T] {
	s.having = ps
	return s
}

func (s *Selector[T]) OrderBy(cols ...Column) *Selector[T] {
	s.orderBy = cols
	return s
}
func (s *Selector[T]) Limit(val int) *Selector[T] {
	s.limit = val
	return s
}
func (s *Selector[T]) Offset(val int) *Selector[T] {
	s.offset = val
	return s
}

// join 查询的结果处理
func (s *Selector[T]) Scan(entity any) (ret []any, err error) {
	typ := reflect.TypeOf(entity)
	if typ.Kind() != reflect.Pointer || typ.Elem().Kind() != reflect.Struct {
		return nil, errs.ErrScanEntityValid
	}

	model, err := s.r.Get(entity)
	if err != nil {
		return nil, err
	}

	q, err := s.Build()
	if err != nil {
		return nil, err
	}

	rows, err := s.sess.queryContext(context.Background(), q.SQL, q.Args...)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		val := s.sess.getCore().creator(model, entity)
		err = val.SetColumns(rows)
		ret = append(ret, entity)
	}

	return ret, err
}

func (s *Selector[T]) Get(ctx context.Context) (*T, error) {
	var err error
	s.model, err = s.r.Get(new(T))
	if err != nil {
		return nil, err
	}

	res := get[T](ctx, &QueryContext{
		Type:    "SELECT",
		Builder: s,
		Model:   s.model,
		Sess:    s.sess,
	})
	if res.Result != nil {
		return res.Result.(*T), res.Err
	}
	return nil, res.Err
}

func (s *Selector[T]) GetMulti(ctx context.Context) ([]*T, error) {
	var err error
	s.model, err = s.r.Get(new(T))
	if err != nil {
		return nil, err
	}

	// 这样改写是为了加入 middleware 功能
	res := getMulti[T](ctx, &QueryContext{
		Type:    "SELECT",
		Builder: s,
		Model:   s.model,
		Sess:    s.sess,
	})
	if res.Result != nil {
		return res.Result.([]*T), res.Err
	}
	return nil, res.Err
}
