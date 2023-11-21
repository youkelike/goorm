package orm

import (
	"context"
	"strconv"

	"gitee.com/youkelike/orm/internal/errs"
)

type Selectable interface {
	selectable()
}

type Selector[T any] struct {
	builder
	db *DB

	// select 子句
	columns []Selectable
	// from 子句
	table string
	// where 子句
	// where 的数据类型只能是 Predicate，不能是 Expression，因为 Column、Value 都不能直接放到 where 中
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

func NewSelector[T any](db *DB) *Selector[T] {
	return &Selector[T]{
		builder: builder{
			dialect: db.dialect,
			quoter:  db.dialect.quoter(),
		},
		db: db,
	}
}

func (s *Selector[T]) Build() (*Query, error) {
	var err error
	s.model, err = s.db.r.Register(new(T))
	if err != nil {
		return nil, err
	}

	s.sb.WriteString("SELECT ")

	err = s.buildColumns()
	if err != nil {
		return nil, err
	}

	s.sb.WriteString(" FROM ")
	if s.table == "" {
		s.sb.WriteString(s.model.TableName)
	} else {
		s.sb.WriteString(s.table)
	}

	if len(s.where) > 0 {
		s.sb.WriteString(" WHERE ")

		// 先把切片形式的条件组装成链表
		p := s.where[0]
		for i := 1; i < len(s.where); i++ {
			p = p.And(s.where[i])
		}
		// 遍历链表
		if err := s.buildExpresssion(p); err != nil {
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
		if err := s.buildExpresssion(p); err != nil {
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

func (s *Selector[T]) buildExpresssion(expr Expression) error {
	switch p := expr.(type) {
	case nil:
	case Predicate:
		_, ok := p.left.(Predicate)
		if ok {
			s.sb.WriteString("(")
		}
		if err := s.buildExpresssion(p.left); err != nil {
			return err
		}
		if ok {
			s.sb.WriteString(")")
		}

		if p.op == opNot || p.op == opAnd || p.op == opOr {
			s.sb.WriteString(" ")
		}
		s.sb.WriteString(p.op.String())
		if p.op == opNot || p.op == opAnd || p.op == opOr {
			s.sb.WriteString(" ")
		}

		_, ok = p.right.(Predicate)
		if ok {
			s.sb.WriteString("(")
		}
		if err := s.buildExpresssion(p.right); err != nil {
			return err
		}
		if ok {
			s.sb.WriteString(")")
		}
	case Column:
		p.alias = ""
		err := s.buildColumn(p)
		if err != nil {
			return err
		}
	case value:
		s.sb.WriteString("?")
		s.addArgs(p.val)
	case RawExpr:
		s.sb.WriteString("(")
		s.sb.WriteString(p.raw)
		s.sb.WriteString(")")
		s.addArgs(p.args...)
	default:
		return errs.NewUnsupportExpression(expr)
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
			err := s.buildColumn(C(c.arg))
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

func (s *Selector[T]) From(table string) *Selector[T] {
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

func (s *Selector[T]) Get(ctx context.Context) (*T, error) {
	q, err := s.Build()
	if err != nil {
		return nil, err
	}
	db := s.db.db
	rows, err := db.QueryContext(ctx, q.SQL, q.Args...)
	if err != nil {
		return nil, err
	}

	if !rows.Next() {
		return nil, errs.ErrNoRows
	}

	tp := new(T)
	val := s.db.creator(s.model, tp)
	val.SetColumns(rows)
	return tp, err
}

func (s *Selector[T]) GetMulti(ctx context.Context) ([]*T, error) {
	panic("")
}
