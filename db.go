package orm

import (
	"database/sql"

	"gitee.com/youkelike/orm/internal/valuer"
	model "gitee.com/youkelike/orm/model"
)

// DB 是连接框架和 sql.DB 包的结合点
type DB struct {
	r       model.Registry
	db      *sql.DB
	creator valuer.Creator
	dialect Dialect
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
		db: db,
		// 这里出现了 3 种不同的接口使用方法：
		// 1、使用一个构造方法，返回一个实现了接口的结构体
		r: model.NewRegistry(),
		// 2、用一个包变量保存实现了接口的结构体，直接引用这个包变量
		dialect: DialectMySQL,
		// 3、类似 build 模式，不直接用接口，而是用一个方法类型，只有在后续调用方法才会返回一个实现了接口的结构体
		creator: valuer.NewUnsafeValue,
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
