package orm

import "database/sql"

// DB 是连接框架和 sql.DB 包的结合点
type DB struct {
	r  *registry
	db *sql.DB
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
		r:  NewRegistry(),
		db: db,
	}
	for _, op := range opts {
		op(res)
	}
	return res, nil
}
