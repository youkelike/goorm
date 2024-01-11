package orm

import (
	"context"

	"gitee.com/youkelike/orm/internal/errs"
	"gitee.com/youkelike/orm/internal/valuer"
	model "gitee.com/youkelike/orm/model"
)

type core struct {
	model   *model.Model
	dialect Dialect
	creator valuer.Creator
	r       model.Registry
	mdls    []Middleware
}

func get[T any](ctx context.Context, sess Session, c core, qc *QueryContext) *QueryResult {
	var root Handler = func(ctx context.Context, qc *QueryContext) *QueryResult {
		return getHandler[T](ctx, sess, c, qc)
	}
	for i := len(c.mdls) - 1; i >= 0; i-- {
		root = c.mdls[i](root)
	}
	return root(ctx, qc)
}

func getHandler[T any](ctx context.Context, sess Session, c core, qc *QueryContext) *QueryResult {
	q, err := qc.Builder.Build()
	if err != nil {
		return &QueryResult{
			Err: err,
		}
	}

	// db := s.db.db
	// rows, err := db.QueryContext(ctx, q.SQL, q.Args...)
	rows, err := sess.queryContext(ctx, q.SQL, q.Args...)
	if err != nil {
		return &QueryResult{
			Err: err,
		}
	}

	if !rows.Next() {
		return &QueryResult{
			Err: errs.ErrNoRows,
		}
	}

	// 在 join 查询中 select 多个表的字段时，传入的 T 必须是包含了所有 select 中字段的聚合结构体
	tp := new(T)
	val := c.creator(c.model, tp)
	err = val.SetColumns(rows)
	return &QueryResult{
		Err:    err,
		Result: tp,
	}
}

func exec(ctx context.Context, sess Session, c core, qc *QueryContext) *QueryResult {
	var root Handler = func(ctx context.Context, qc *QueryContext) *QueryResult {
		return execHandler(ctx, sess, c, qc)
	}
	for i := len(c.mdls) - 1; i >= 0; i-- {
		root = c.mdls[i](root)
	}
	return root(ctx, qc)
}

func execHandler(ctx context.Context, sess Session, c core, qc *QueryContext) *QueryResult {
	q, err := qc.Builder.Build()
	if err != nil {
		return &QueryResult{
			Err: err,
		}
	}

	res, err := sess.execContext(ctx, q.SQL, q.Args...)
	return &QueryResult{
		Err:    err,
		Result: res,
	}
}

func getMulti[T any](ctx context.Context, sess Session, c core, qc *QueryContext) *QueryResult {
	// 把业务逻辑改造成一个 handler
	var root Handler = func(ctx context.Context, qc *QueryContext) *QueryResult {
		return getMultiHandler[T](ctx, sess, c, qc)
	}
	for i := len(c.mdls) - 1; i >= 0; i-- {
		root = c.mdls[i](root)
	}
	return root(ctx, qc)
}

func getMultiHandler[T any](ctx context.Context, sess Session, c core, qc *QueryContext) *QueryResult {
	q, err := qc.Builder.Build()
	if err != nil {
		return &QueryResult{
			Err: err,
		}
	}

	rows, err := sess.queryContext(ctx, q.SQL, q.Args...)
	if err != nil {
		return &QueryResult{
			Err: err,
		}
	}

	var tps []*T
	for rows.Next() {
		tp := new(T)
		val := c.creator(c.model, tp)
		err = val.SetColumns(rows)
		tps = append(tps, tp)
	}

	if tps == nil {
		return &QueryResult{
			Err: errs.ErrNoRows,
		}
	}

	return &QueryResult{
		Err:    err,
		Result: tps,
	}
}
