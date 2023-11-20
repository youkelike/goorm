package orm

import (
	"fmt"
	"testing"

	"gitee.com/youkelike/go1/work/hw05/orm/internal/errs"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuild(t *testing.T) {
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
			name:    "invalid column",
			builder: NewDeletor[TestModel](db).Where(C("FirstName").Eq("Tom").Or(C("XXX").Eq(1))),
			wantErr: errs.NewUnknownField("XXX"),
		},
		{
			name:    "or",
			builder: NewDeletor[TestModel](db).Where(C("FirstName").Eq("Tom").Or(C("Age").Eq(18))),
			wantQuery: &Query{
				SQL:  "DELETE FROM test_model WHERE (first_name=?) OR (age=?);",
				Args: []any{"Tom", 18},
			},
		},
		{
			name:    "and",
			builder: NewDeletor[TestModel](db).Where(C("FirstName").Eq("Tom").And(C("Age").Eq(18))),
			wantQuery: &Query{
				SQL:  "DELETE FROM test_model WHERE (first_name=?) AND (age=?);",
				Args: []any{"Tom", 18},
			},
		},
		{
			name:    "not",
			builder: NewDeletor[TestModel](db).Where(Not(C("FirstName").Eq("Tom"))),
			wantQuery: &Query{
				SQL:  "DELETE FROM test_model WHERE  NOT (first_name=?);",
				Args: []any{"Tom"},
			},
		},
		{
			name:    "where",
			builder: NewDeletor[TestModel](db).Where(C("FirstName").Eq("Tom")),
			wantQuery: &Query{
				SQL:  "DELETE FROM test_model WHERE first_name=?;",
				Args: []any{"Tom"},
			},
		},
		{
			name:    "empty where",
			builder: NewDeletor[TestModel](db).Where(),
			wantQuery: &Query{
				SQL: "DELETE FROM test_model;",
			},
		},
		{
			name:    "from db",
			builder: NewDeletor[TestModel](db).From("test_db.test_model"),
			wantQuery: &Query{
				SQL: "DELETE FROM test_db.test_model;",
			},
		},
		{
			name:    "empty from",
			builder: NewDeletor[TestModel](db).From(""),
			wantQuery: &Query{
				SQL: "DELETE FROM test_model;",
			},
		},
		{
			name:    "use from",
			builder: NewDeletor[TestModel](db).From("test_model"),
			wantQuery: &Query{
				SQL: "DELETE FROM test_model;",
			},
		},
		{
			name:    "no from",
			builder: NewDeletor[TestModel](db),
			wantQuery: &Query{
				SQL: "DELETE FROM test_model;",
			},
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
