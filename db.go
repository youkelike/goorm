package orm

import (
	"context"
	"database/sql"

	"gitee.com/youkelike/orm/internal/errs"
	"gitee.com/youkelike/orm/internal/valuer"
	model "gitee.com/youkelike/orm/model"
)

// 主要用于集中管理、注册解析好的模型
type DB struct {
	core
	db *sql.DB
}

func (db *DB) getCore() core {
	return db.core
}

func (db *DB) DoTx(ctx context.Context,
	// 这里的参数必须是 *Tx，不能是 Session 接口，因为必须明确这里是在调用事件下的方法
	fn func(ctx context.Context, tx *Tx) error,
	opts *sql.TxOptions) (err error) {

	tx, err := db.BeginTx(ctx, opts)
	if err != nil {
		return err
	}
	panicked := true
	defer func() {
		if panicked || err != nil {
			e := tx.Rollback()
			err = errs.NewErrFailedToRollback(err, e, true)
		} else {
			err = tx.Commit()
		}
	}()
	err = fn(ctx, tx)
	panicked = false
	return err
}

func (db *DB) BeginTx(ctx context.Context, opts *sql.TxOptions) (*Tx, error) {
	tx, err := db.db.BeginTx(ctx, opts)
	if err != nil {
		return nil, err
	}
	return &Tx{tx: tx, db: db}, nil
}

// 事务扩散方案
// type txKey struct{}
// func (db *DB) BeginTxV2(ctx context.Context, opts *sql.TxOptions) (context.Context, *Tx, error) {
// 	val := ctx.Value(txKey{})
// 	tx, ok := val.(*Tx)
// 	if ok && !tx.done {
// 		return ctx, tx, nil
// 	}
// 	tx, err := db.BeginTx(ctx, opts)
// 	if err != nil {
// 		return nil, nil, err
// 	}
// 	ctx = context.WithValue(ctx, txKey{}, tx)
// 	return ctx, tx, nil
// }

func (db *DB) queryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return db.db.QueryContext(ctx, query, args...)
}

func (db *DB) execContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return db.db.ExecContext(ctx, query, args...)
}

type DBOption func(*DB)

func Open(driver string, dataSourceName string, opts ...DBOption) (*DB, error) {
	db, err := sql.Open(driver, dataSourceName)
	if err != nil {
		return nil, err
	}

	return OpenDB(db, opts...)
}

func MustOpen(driver string, dataSourceName string, opts ...DBOption) *DB {
	res, err := Open(driver, dataSourceName, opts...)
	if err != nil {
		panic(err)
	}
	return res
}

// 拎出这个方法在测试时很有用
func OpenDB(db *sql.DB, opts ...DBOption) (*DB, error) {
	res := &DB{
		core: core{
			r:       model.NewRegistry(),
			creator: valuer.NewUnsafeValue,
			dialect: DialectMySQL,
		},
		db: db,
	}
	for _, op := range opts {
		op(res)
	}
	return res, nil
}

func DBUseReflect() DBOption {
	return func(d *DB) {
		d.creator = valuer.NewReflectValue
	}
}

func DBWithRegistry(registry model.Registry) DBOption {
	return func(d *DB) {
		d.r = registry
	}
}

func DBWithDialect(dialect Dialect) DBOption {
	return func(d *DB) {
		d.dialect = dialect
	}
}

func DBWithMiddlewares(mdls ...Middleware) DBOption {
	return func(d *DB) {
		d.mdls = mdls
	}
}
