package orm

import (
	"context"
	"database/sql"
)

var (
	_ Session = &Tx{}
	_ Session = &DB{}
)

type Session interface {
	getCore() core
	queryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	execContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

type Tx struct {
	tx   *sql.Tx
	db   *DB
	done bool
}

func (t *Tx) getCore() core {
	return t.db.core
}

func (t *Tx) Commit() error {
	t.done = true
	return t.tx.Commit()
}

func (t *Tx) Rollback() error {
	t.done = true
	return t.tx.Rollback()
}

func (t *Tx) RollbackIfNotCommit() error {
	t.done = true
	err := t.tx.Rollback()
	if err == sql.ErrTxDone {
		return nil
	}
	return err
}

func (t *Tx) queryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return t.tx.QueryContext(ctx, query, args...)
}

func (t *Tx) execContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return t.tx.ExecContext(ctx, query, args...)
}
