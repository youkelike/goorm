package orm

import (
	"context"
	"database/sql"
)

// 执行查询
type Querier[T any] interface {
	Get(ctx context.Context) (*T, error)
	GetMulti(ctx context.Context) ([]*T, error)
}

// 执行增删改
type Executor interface {
	Exec(ctx context.Context) (sql.Result, error)
}

// 生成 sql 语句
type QueryBuilder interface {
	Build() (*Query, error)
}

// sql 语句包含了语法和参数两部分
type Query struct {
	SQL  string
	Args []any
}

// 用接口的方式提供自定义表名的途径
type TableName interface {
	TableName() string
}
