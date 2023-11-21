package orm

import (
	"context"
)

type RawQuerier[T any] struct {
	core
	sess Session
	sql  string
	args []any
}

func RawQuery[T any](sess Session, query string, args ...any) *RawQuerier[T] {
	c := sess.getCore()
	return &RawQuerier[T]{
		sql:  query,
		args: args,
		sess: sess,
		core: c,
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
	r.model, err = r.r.Get(new(T))
	if err != nil {
		return Result{
			err: err,
		}
	}

	res := exec(ctx, r.sess, r.core, &QueryContext{
		Type:    "RAW",
		Builder: r,
		Model:   r.model,
	})

	if res.Result != nil {
		return res.Result.(Result)
	}

	return Result{
		err: res.Err,
	}
}

func (s RawQuerier[T]) Get(ctx context.Context) (*T, error) {
	var err error
	s.model, err = s.r.Get(new(T))
	if err != nil {
		return nil, err
	}

	res := get[T](ctx, s.sess, s.core, &QueryContext{
		Type:    "RAW",
		Builder: s,
		Model:   s.model,
	})

	if res.Result != nil {
		return res.Result.(*T), res.Err
	}
	return nil, res.Err
}

func (r RawQuerier[T]) GetMulti(ctx context.Context) ([]*T, error) {
	var err error
	r.model, err = r.r.Get(new(T))
	if err != nil {
		return nil, err
	}

	// 这样改写是为了加入 middleware 功能
	res := getMulti[T](ctx, r.sess, r.core, &QueryContext{
		Type:    "RAW",
		Builder: r,
		Model:   r.model,
	})
	if res.Result != nil {
		return res.Result.([]*T), res.Err
	}
	return nil, res.Err
}
