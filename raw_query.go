package orm

import (
	"context"
	"database/sql"
)

type RawQuerier[T any] struct {
	// *core
	sess Session
	sql  string
	args []any
}

func RawQuery[T any](sess Session, query string, args ...any) *RawQuerier[T] {
	// c := sess.getCore()
	return &RawQuerier[T]{
		sql:  query,
		args: args,
		sess: sess,
		// core: &c,
	}
}

func (r RawQuerier[T]) Build() (*Query, error) {
	return &Query{
		SQL:  r.sql,
		Args: r.args,
	}, nil
}

func (r RawQuerier[T]) Exec(ctx context.Context) Result {
	var err error
	model, err := r.sess.getCore().r.Get(new(T))
	if err != nil {
		return Result{
			err: err,
		}
	}

	res := exec(ctx, &QueryContext{
		Type:    "RAW",
		Builder: r,
		Model:   model,
		Sess:    r.sess,
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

func (s RawQuerier[T]) Get(ctx context.Context) (*T, error) {
	var err error
	model, err := s.sess.getCore().r.Get(new(T))
	if err != nil {
		return nil, err
	}

	res := get[T](ctx, &QueryContext{
		Type:    "RAW",
		Builder: s,
		Model:   model,
		Sess:    s.sess,
	})

	if res.Result != nil {
		return res.Result.(*T), res.Err
	}
	return nil, res.Err
}

func (r RawQuerier[T]) GetMulti(ctx context.Context) ([]*T, error) {
	var err error
	model, err := r.sess.getCore().r.Get(new(T))
	if err != nil {
		return nil, err
	}

	// 这样改写是为了加入 middleware 功能
	res := getMulti[T](ctx, &QueryContext{
		Type:    "RAW",
		Builder: r,
		Model:   model,
		Sess:    r.sess,
	})
	if res.Result != nil {
		return res.Result.([]*T), res.Err
	}
	return nil, res.Err
}
