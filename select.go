package orm

import (
	"context"
	"reflect"
	"strings"

	"gitee.com/youkelike/orm/internal/errs"
)

type Selector[T any] struct {
	table string
	where []Predicate

	sb   *strings.Builder
	args []any

	model *Model
	db    *DB
}

func NewSelector[T any](db *DB) *Selector[T] {
	return &Selector[T]{
		sb: &strings.Builder{},
		db: db,
	}
}

func (s *Selector[T]) Build() (*Query, error) {
	var err error
	// 这里用 Get 比 Register 好
	s.model, err = s.db.r.Get(new(T))
	// s.model, err = s.db.r.Register(new(T))
	if err != nil {
		return nil, err
	}

	s.sb.WriteString("SELECT * FROM ")

	if s.table == "" {
		s.sb.WriteString(s.model.tableName)
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
		fd, ok := s.model.fieldMap[p.name]
		if !ok {
			return errs.NewUnknownField(p.name)
		}
		s.sb.WriteString(fd.colName)
	case value:
		s.sb.WriteString("?")
		s.addArg(p.val)
	default:
		return errs.NewUnsupportExpression(expr)
	}
	return nil
}

func (s *Selector[T]) addArg(val any) *Selector[T] {
	if s.args == nil {
		s.args = make([]any, 0, 8)
	}
	s.args = append(s.args, val)
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
		return nil, ErrNoRows
	}

	// 查询到记录的所有列名
	cs, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	// 挖坑，指定每一个坑的数据类型
	vals := make([]any, 0, len(cs))
	for _, c := range cs {
		fd, ok := s.model.columnMap[c]
		if !ok {
			return nil, errs.NewUnknownColumn(c)
		}
		// 根据字段类型创建一个指针类型的值
		val := reflect.New(fd.typ)
		// 不能这样写，因为后面要对它赋值
		// val := reflect.Zero(fd.typ)

		// vals 里接收的是 any 类型，需要转换一下
		vals = append(vals, val.Interface())
	}

	// 利用查询的返回值，往坑里填具体值
	err = rows.Scan(vals...)
	if err != nil {
		return nil, err
	}

	// 把填在坑里的值填到 struct 中
	// 这里操作的是具体结构体的指针
	tp := new(T)
	tpValue := reflect.ValueOf(tp)
	for i, c := range cs {
		fd := s.model.columnMap[c]

		// 结构体指针必须转成结构体，才能给其字段赋值
		tpValue.Elem().FieldByName(fd.goName).Set(reflect.ValueOf(vals[i]).Elem())
	}

	return tp, err
}

func (s *Selector[T]) GetMulti(ctx context.Context) ([]*T, error) {
	panic("")
}
