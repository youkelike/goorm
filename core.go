package orm

import (
	"context"

	"gitee.com/youkelike/orm/internal/errs"
	"gitee.com/youkelike/orm/internal/valuer"
	model "gitee.com/youkelike/orm/model"
)

type core struct {
	// core 是全局使用的，model 是个性化的，不适合放这
	// model *model.Model

	// 模型解析
	r model.Registry
	// 方言处理
	dialect Dialect
	// 结果集映射
	creator valuer.Creator
	// 中间件
	mdls []Middleware
}

// 为了支持泛型，只能用函数，不能做成绑定到对象上的方法
func get[T any](ctx context.Context, qc *QueryContext) *QueryResult {
	root := getHandler[T]
	for i := len(qc.Sess.getCore().mdls) - 1; i >= 0; i-- {
		root = qc.Sess.getCore().mdls[i](root)
	}
	return root(ctx, qc)
}

func getHandler[T any](ctx context.Context, qc *QueryContext) *QueryResult {
	q, err := qc.Builder.Build()
	if err != nil {
		return &QueryResult{
			Err: err,
		}
	}

	rows, err := qc.Sess.queryContext(ctx, q.SQL, q.Args...)
	if err != nil {
		return &QueryResult{
			Err: err,
		}
	}
	defer rows.Close()

	if !rows.Next() {
		return &QueryResult{
			Err: errs.ErrNoRows,
		}
	}

	// 在 join 查询中 select 多个表的字段时，传入的 T 必须是包含了所有 select 中字段的聚合结构体
	tp := new(T)
	val := qc.Sess.getCore().creator(qc.Model, tp)
	err = val.SetColumns(rows)
	return &QueryResult{
		Err:    err,
		Result: tp,
	}
}

func exec(ctx context.Context, qc *QueryContext) *QueryResult {
	root := execHandler
	for i := len(qc.Sess.getCore().mdls) - 1; i >= 0; i-- {
		root = qc.Sess.getCore().mdls[i](root)
	}
	return root(ctx, qc)
}

func execHandler(ctx context.Context, qc *QueryContext) *QueryResult {
	q, err := qc.Builder.Build()
	if err != nil {
		return &QueryResult{
			Err: err,
		}
	}

	res, err := qc.Sess.execContext(ctx, q.SQL, q.Args...)
	return &QueryResult{
		Err:    err,
		Result: res,
	}
}

func getMulti[T any](ctx context.Context, qc *QueryContext) *QueryResult {
	// 把业务逻辑改造成一个 handler
	root := getMultiHandler[T]
	for i := len(qc.Sess.getCore().mdls) - 1; i >= 0; i-- {
		root = qc.Sess.getCore().mdls[i](root)
	}
	return root(ctx, qc)
}

func getMultiHandler[T any](ctx context.Context, qc *QueryContext) *QueryResult {
	q, err := qc.Builder.Build()
	if err != nil {
		return &QueryResult{
			Err: err,
		}
	}

	rows, err := qc.Sess.queryContext(ctx, q.SQL, q.Args...)
	if err != nil {
		return &QueryResult{
			Err: err,
		}
	}
	defer rows.Close()

	var tps []*T
	for rows.Next() {
		tp := new(T)
		val := qc.Sess.getCore().creator(qc.Model, tp)
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
