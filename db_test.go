package orm

import (
	"context"
	"database/sql"
	"testing"

	"gitee.com/youkelike/orm/internal/errs"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDB_DoTx(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	db, err := OpenDB(mockDB)
	require.NoError(t, err)

	res := sqlmock.NewResult(0, 1)
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT *").WillReturnError(errs.ErrNoRows)
	mock.ExpectExec("UPDATE *").WillReturnResult(res)
	mock.ExpectExec("INSERT *").WillReturnResult(res)
	mock.ExpectExec("UPDATE *").WillReturnResult(res)
	mock.ExpectCommit()
	tm := &TestModel{
		Id:        1,
		FirstName: "zs",
		Age:       18,
		// LastName:  &sql.NullString{String: "mark", Valid: true},
	}

	err = db.DoTx(context.Background(), func(ctx context.Context, tx *Tx) error {
		_, err := NewSelector[TestModel](tx).Where(C("Id").Eq(1)).Get(ctx)
		assert.Equal(t, errs.ErrNoRows, err)

		ret := NewUpdater[TestModel](tx).Value(tm).Updates(C("Age"), C("FirstName")).Where(C("FirstName").Eq("Tom")).Exec(ctx)
		lastInsertId, err := ret.LastInsertId()
		require.NoError(t, err)
		assert.Equal(t, int64(0), lastInsertId)
		rowsAffected, err := ret.RowsAffected()
		require.NoError(t, err)
		assert.Equal(t, int64(1), rowsAffected)

		ret = NewInserter[TestModel](db).Values(&TestModel{
			Id:        1,
			FirstName: "Tom",
			Age:       18,
			LastName:  &sql.NullString{Valid: true, String: "Jerry"},
		}).Upsert().ConflictColumns("Id").Update(Assign("Age", 10), Assign("FirstName", "Bob")).Exec(ctx)
		lastInsertId, err = ret.LastInsertId()
		require.NoError(t, err)
		assert.Equal(t, int64(0), lastInsertId)
		rowsAffected, err = ret.RowsAffected()
		require.NoError(t, err)
		assert.Equal(t, int64(1), rowsAffected)

		ret = RawQuery[TestModel](tx, "UPDATE * test_model WHERE id = ?", -1).Exec(ctx)
		lastInsertId, err = ret.LastInsertId()
		require.NoError(t, err)
		assert.Equal(t, int64(0), lastInsertId)
		rowsAffected, err = ret.RowsAffected()
		require.NoError(t, err)
		assert.Equal(t, int64(1), rowsAffected)

		return nil
	}, &sql.TxOptions{})
	require.NoError(t, err)
}
