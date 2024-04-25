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

func TestRawQuerier_Get(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	db, err := OpenDB(mockDB)
	require.NoError(t, err)

	mock.ExpectQuery("SELECT .*").WillReturnError(errs.ErrNoRows)

	mock.ExpectQuery("SELECT .*").WillReturnError(errs.ErrNoRows)

	rows := sqlmock.NewRows([]string{"id", "first_name", "age", "last_name"})
	rows.AddRow("1", "Tom", "18", "Jerry")
	mock.ExpectQuery("SELECT .*").WillReturnRows(rows)

	testCases := []struct {
		name    string
		r       *RawQuerier[TestModel]
		wantRes *TestModel
		wantErr error
	}{
		{
			name:    "query error",
			r:       RawQuery[TestModel](db, "SELECT * FROM test_model"),
			wantErr: errs.ErrNoRows,
		},
		{
			name:    "now rows",
			r:       RawQuery[TestModel](db, "SELECT * FROM test_model WHERE id = ?", -1),
			wantErr: errs.ErrNoRows,
		},
		{
			name: "data",
			r:    RawQuery[TestModel](db, "SELECT * FROM test_model WHERE id = ?", 1),
			wantRes: &TestModel{
				Id:        1,
				FirstName: "Tom",
				Age:       18,
				LastName:  &sql.NullString{Valid: true, String: "Jerry"},
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			res, err := tc.r.Get(context.Background())
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, tc.wantRes, res)
		})
	}
}
