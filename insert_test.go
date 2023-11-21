package orm

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"testing"

	"gitee.com/youkelike/orm/internal/errs"
	"github.com/DATA-DOG/go-sqlmock"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInserter_Build(t *testing.T) {
	// mockDB, _, err := sqlmock.New()
	// require.NoError(t, err)
	// db, err := OpenDB(mockDB)
	// require.NoError(t, err)
	// defer mockDB.Close()

	db := memoryDB(t, DBWithDialect(DialectMySQL))

	testCases := []struct {
		name      string
		i         QueryBuilder
		wantErr   error
		wantQuery *Query
	}{
		{
			name: "upsert",
			i: NewInserter[TestModel](db).Values(&TestModel{
				Id:        1,
				FirstName: "Tom",
				Age:       18,
				LastName:  &sql.NullString{Valid: true, String: "Jerry"},
			}).Upsert().Update(Assign("Age", 10), Assign("FirstName", "Bob")),
			wantQuery: &Query{
				SQL:  "INSERT INTO test_model (id,first_name,age,last_name) VALUES (?,?,?,?) ON DUPLICATE KEY UPDATE age=?,first_name=?;",
				Args: []any{int64(1), "Tom", int8(18), &sql.NullString{Valid: true, String: "Jerry"}, 10, "Bob"},
			},
		},
		{
			name: "upsert update column",
			i: NewInserter[TestModel](db).Values(&TestModel{
				Id:        1,
				FirstName: "Tom",
				Age:       18,
				LastName:  &sql.NullString{Valid: true, String: "Jerry"},
			}).Upsert().Update(C("Age"), C("FirstName")),
			wantQuery: &Query{
				SQL:  "INSERT INTO test_model (id,first_name,age,last_name) VALUES (?,?,?,?) ON DUPLICATE KEY UPDATE age=VALUES(age),first_name=VALUES(first_name);",
				Args: []any{int64(1), "Tom", int8(18), &sql.NullString{Valid: true, String: "Jerry"}},
			},
		},
		{
			name:    "no row",
			i:       NewInserter[TestModel](db).Values(),
			wantErr: errs.ErrInsertZeroRow,
		},
		{
			name: "signle row",
			i: NewInserter[TestModel](db).Values(&TestModel{
				Id:        1,
				FirstName: "Tom",
				Age:       18,
				LastName:  &sql.NullString{Valid: true, String: "Jerry"},
			}),
			wantQuery: &Query{
				SQL:  "INSERT INTO test_model (id,first_name,age,last_name) VALUES (?,?,?,?);",
				Args: []any{int64(1), "Tom", int8(18), &sql.NullString{Valid: true, String: "Jerry"}},
			},
		},
		{
			name: "many row",
			i: NewInserter[TestModel](db).Values(&TestModel{
				Id:        1,
				FirstName: "Tom",
				Age:       18,
				LastName:  &sql.NullString{Valid: true, String: "Jerry"},
			}, &TestModel{
				Id:        2,
				FirstName: "Tom2",
				Age:       19,
				LastName:  &sql.NullString{Valid: true, String: "Jerry2"},
			}),
			wantQuery: &Query{
				SQL:  "INSERT INTO test_model (id,first_name,age,last_name) VALUES (?,?,?,?),(?,?,?,?);",
				Args: []any{int64(1), "Tom", int8(18), &sql.NullString{Valid: true, String: "Jerry"}, int64(2), "Tom2", int8(19), &sql.NullString{Valid: true, String: "Jerry2"}},
			},
		},
		{
			name: "partial column",
			i: NewInserter[TestModel](db).Columns("Id", "FirstName").Values(&TestModel{
				Id:        1,
				FirstName: "Tom",
				Age:       18,
				LastName:  &sql.NullString{Valid: true, String: "Jerry"},
			}),
			wantQuery: &Query{
				SQL:  "INSERT INTO test_model (id,first_name) VALUES (?,?);",
				Args: []any{int64(1), "Tom"},
			},
		},
		{
			name: "partial column many rows",
			i: NewInserter[TestModel](db).Columns("Id", "FirstName").Values(&TestModel{
				Id:        1,
				FirstName: "Tom",
			}, &TestModel{
				Id:        2,
				FirstName: "Tom2",
			}),
			wantQuery: &Query{
				SQL:  "INSERT INTO test_model (id,first_name) VALUES (?,?),(?,?);",
				Args: []any{int64(1), "Tom", int64(2), "Tom2"},
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			q, err := tc.i.Build()
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, tc.wantQuery, q)
		})
	}
}

func TestSQLite_Inserter_Build(t *testing.T) {
	db := memoryDB(t, DBWithDialect(DialectSQLite))

	testCases := []struct {
		name      string
		i         QueryBuilder
		wantErr   error
		wantQuery *Query
	}{
		{
			name: "upsert",
			i: NewInserter[TestModel](db).Values(&TestModel{
				Id:        1,
				FirstName: "Tom",
				Age:       18,
				LastName:  &sql.NullString{Valid: true, String: "Jerry"},
			}).Upsert().ConflictColumns("Id").Update(Assign("Age", 10), Assign("FirstName", "Bob")),
			wantQuery: &Query{
				SQL:  "INSERT INTO test_model (id,first_name,age,last_name) VALUES (?,?,?,?) ON CONFLICT(id) DO UPDATE SET age=?,first_name=?;",
				Args: []any{int64(1), "Tom", int8(18), &sql.NullString{Valid: true, String: "Jerry"}, 10, "Bob"},
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			q, err := tc.i.Build()
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, tc.wantQuery, q)
		})
	}
}

func TestInserter_Exec(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	db, err := OpenDB(mockDB)
	require.NoError(t, err)

	testCases := []struct {
		name     string
		i        *Inserter[TestModel]
		wantErr  error
		affected int64
	}{
		{
			name: "db error",
			i: func() *Inserter[TestModel] {
				mock.ExpectExec("INSERT INTO .*").WillReturnError(errors.New("db error"))

				return NewInserter[TestModel](db).Values(&TestModel{})
			}(),
			wantErr: errors.New("db error"),
		},
		{
			name: "query error",
			i: func() *Inserter[TestModel] {
				return NewInserter[TestModel](db).Values(&TestModel{}).Columns("Invalid")
			}(),
			wantErr: errs.NewUnknownField("Invalid"),
		},
		{
			name: "exec",
			i: func() *Inserter[TestModel] {
				res := driver.RowsAffected(1)
				mock.ExpectExec("INSERT INTO .*").WillReturnResult(res)

				return NewInserter[TestModel](db).Values(&TestModel{})
			}(),
			affected: 1,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			res := tc.i.Exec(context.Background())
			affected, err := res.RowsAffected()
			require.Equal(t, err, tc.wantErr)
			if err != nil {
				return
			}
			assert.Equal(t, affected, tc.affected)
		})
	}
}

func memoryDB(t *testing.T, opts ...DBOption) *DB {
	db, err := Open("sqlite3", "file:test.db?cache=shared&mode=memory", opts...)
	require.NoError(t, err)
	return db
}
