package orm

import (
	"context"
)

type Select[T any] struct {
	builder
	table string
	where []Predicate
}

func (s *Select[T]) Build() (*Query, error) {
	var err error
	s.model, err = parseModel(new(T))
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
		err := s.buildWhere(s.where)
		if err != nil {
			return nil, err
		}
	}

	s.sb.WriteString(";")

	return &Query{
		SQL:  s.sb.String(),
		Args: s.args,
	}, nil
}

func (s *Select[T]) From(table string) *Select[T] {
	s.table = table
	return s
}

func (s *Select[T]) Where(ps ...Predicate) *Select[T] {
	s.where = ps
	return s
}

func (s *Select[T]) Get(ctx context.Context) (*T, error) {
	panic("")
}

func (s *Select[T]) GetMulti(ctx context.Context) ([]*T, error) {
	panic("")
}
