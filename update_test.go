package orm

import (
	"database/sql"
	"fmt"
	"testing"

	"gitee.com/youkelike/orm/internal/errs"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildUpdater(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	db, err := OpenDB(mockDB)
	require.NoError(t, err)
	fmt.Println(mock)

	tm := &TestModel{
		Id:        1,
		FirstName: "zs",
		Age:       18,
		// LastName:  &sql.NullString{String: "mark", Valid: true},
	}

	testCases := []struct {
		name      string
		builder   QueryBuilder
		wantQuery *Query
		wantErr   error
	}{
		{
			name:    "invalid value",
			builder: NewUpdater[TestModel](db).Where(C("FirstName").Eq("Tom").Or(C("XXX").Eq(1))),
			wantErr: errs.NewUnknownUpdateValue(),
		},
		{
			name:    "value",
			builder: NewUpdater[TestModel](db).Value(tm),
			wantQuery: &Query{
				SQL:  "UPDATE test_model SET id=?,first_name=?,age=?,last_name=?;",
				Args: []any{int64(1), "zs", int8(18), (*sql.NullString)(nil)},
			},
		},
		{
			name:    "where",
			builder: NewUpdater[TestModel](db).Value(tm).Where(C("FirstName").Eq("Tom")),
			wantQuery: &Query{
				SQL:  "UPDATE test_model SET id=?,first_name=?,age=?,last_name=? WHERE first_name=?;",
				Args: []any{int64(1), "zs", int8(18), (*sql.NullString)(nil), "Tom"},
			},
		},
		{
			name:    "updates",
			builder: NewUpdater[TestModel](db).Value(tm).Updates(C("Age"), C("FirstName")).Where(C("FirstName").Eq("Tom")),
			wantQuery: &Query{
				SQL:  "UPDATE test_model SET age=?,first_name=? WHERE first_name=?;",
				Args: []any{int8(18), "zs", "Tom"},
			},
		},
		{
			name:    "updates invalid",
			builder: NewUpdater[TestModel](db).Value(tm).Updates(C("XXX")).Where(C("FirstName").Eq("Tom")),
			wantErr: errs.NewUnknownField("XXX"),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			res, err := tc.builder.Build()
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, tc.wantQuery, res)
		})
	}
}
