package orm

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	"gitee.com/youkelike/go1/work/hw05/orm/internal/errs"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSelect_Build(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	db, err := OpenDB(mockDB)
	require.NoError(t, err)
	fmt.Println(mock)

	testCases := []struct {
		name      string
		builder   QueryBuilder
		wantQuery *Query
		wantErr   error
	}{
		{
			name:    "no from",
			builder: NewSelector[TestModel](db),
			wantQuery: &Query{
				SQL:  "SELECT * FROM test_model;",
				Args: nil,
			},
		},
		{
			name:    "use from",
			builder: NewSelector[TestModel](db).From("test_model"),
			wantQuery: &Query{
				SQL:  "SELECT * FROM test_model;",
				Args: nil,
			},
		},
		{
			name:    "empty from",
			builder: NewSelector[TestModel](db).From(""),
			wantQuery: &Query{
				SQL:  "SELECT * FROM test_model;",
				Args: nil,
			},
		},
		{
			name:    "from db",
			builder: NewSelector[TestModel](db).From("test_db.test_model"),
			wantQuery: &Query{
				SQL:  "SELECT * FROM test_db.test_model;",
				Args: nil,
			},
		},
		{
			name:    "empty where",
			builder: NewSelector[TestModel](db).Where(),
			wantQuery: &Query{
				SQL:  "SELECT * FROM test_model;",
				Args: nil,
			},
		},
		{
			name:    "where",
			builder: NewSelector[TestModel](db).Where(C("Age").Eq(18)),
			wantQuery: &Query{
				SQL:  "SELECT * FROM test_model WHERE age=?;",
				Args: []any{18},
			},
		},
		{
			name:    "not",
			builder: NewSelector[TestModel](db).Where(Not(C("Age").Eq(18))),
			wantQuery: &Query{
				SQL:  "SELECT * FROM test_model WHERE  NOT (age=?);",
				Args: []any{18},
			},
		},
		{
			name:    "and",
			builder: NewSelector[TestModel](db).Where(C("Age").Eq(18).And(C("FirstName").Eq("Tom"))),
			wantQuery: &Query{
				SQL:  "SELECT * FROM test_model WHERE (age=?) AND (first_name=?);",
				Args: []any{18, "Tom"},
			},
		},
		{
			name:    "or",
			builder: NewSelector[TestModel](db).Where(C("Age").Eq(18).Or(C("FirstName").Eq("Tom"))),
			wantQuery: &Query{
				SQL:  "SELECT * FROM test_model WHERE (age=?) OR (first_name=?);",
				Args: []any{18, "Tom"},
			},
		},
		{
			name:    "invalid column",
			builder: NewSelector[TestModel](db).Where(C("Age").Eq(18).Or(C("XXX").Eq("Tom"))),
			wantErr: errs.NewUnknownField("XXX"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			q, err := tc.builder.Build()
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, tc.wantQuery, q)

		})
	}

}

type TestModel struct {
	Id        int64
	FirstName string
	Age       int8
	LastName  *sql.NullString
}

func TestSelector_Get(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	db, err := OpenDB(mockDB)
	require.NoError(t, err)

	mock.ExpectQuery("SELECT .*").WillReturnError(ErrNoRows)

	rows := sqlmock.NewRows([]string{"id", "first_name", "age", "last_name"})
	mock.ExpectQuery("SELECT .*").WillReturnRows(rows)

	rows = sqlmock.NewRows([]string{"id", "first_name", "age", "last_name"})
	rows.AddRow("1", "Tom", "18", "Jerry")
	mock.ExpectQuery("SELECT .*").WillReturnRows(rows)

	// rows = sqlmock.NewRows([]string{"id", "first_name", "age", "last_name"})
	// rows.AddRow("abc", "Tom", "18", "Jerry")
	// mock.ExpectQuery("SELECT .*").WillReturnRows(rows)

	testCases := []struct {
		name    string
		s       *Selector[TestModel]
		wantRes *TestModel
		wantErr error
	}{
		{
			name:    "invalid sql",
			s:       NewSelector[TestModel](db).Where(C("XXX").Eq(1)),
			wantErr: errs.NewUnknownField("XXX"),
		},
		{
			name:    "query error",
			s:       NewSelector[TestModel](db).Where(C("Id").Eq(1)),
			wantErr: ErrNoRows,
		},
		{
			name:    "now rows",
			s:       NewSelector[TestModel](db).Where(C("Id").Eq(1)),
			wantErr: ErrNoRows,
		},
		{
			name: "data",
			s:    NewSelector[TestModel](db).Where(C("Id").Eq(1)),
			wantRes: &TestModel{
				Id:        1,
				FirstName: "Tom",
				Age:       18,
				LastName:  &sql.NullString{Valid: true, String: "Jerry"},
			},
		},
		// {
		// 	name:    "scan error",
		// 	s:       NewSelector[TestModel](db).Where(C("Id").Eq(1)),
		// 	wantErr: ErrNoRows,
		// },
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			res, err := tc.s.Get(context.Background())
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, tc.wantRes, res)
		})
	}
}
