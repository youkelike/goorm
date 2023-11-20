package orm

import (
	"database/sql"
	"testing"

	"gitee.com/youkelike/orm/internal/errs"
	"github.com/stretchr/testify/assert"
)

func TestSelect_Build(t *testing.T) {
	testCases := []struct {
		name      string
		builder   QueryBuilder
		wantQuery *Query
		wantErr   error
	}{
		{
			name:    "no from",
			builder: &Select[TestModel]{},
			wantQuery: &Query{
				SQL:  "SELECT * FROM test_model;",
				Args: nil,
			},
		},
		{
			name:    "use from",
			builder: (&Select[TestModel]{}).From("test_model"),
			wantQuery: &Query{
				SQL:  "SELECT * FROM test_model;",
				Args: nil,
			},
		},
		{
			name:    "empty from",
			builder: (&Select[TestModel]{}).From(""),
			wantQuery: &Query{
				SQL:  "SELECT * FROM test_model;",
				Args: nil,
			},
		},
		{
			name:    "from db",
			builder: (&Select[TestModel]{}).From("test_db.test_model"),
			wantQuery: &Query{
				SQL:  "SELECT * FROM test_db.test_model;",
				Args: nil,
			},
		},
		{
			name:    "empty where",
			builder: (&Select[TestModel]{}).Where(),
			wantQuery: &Query{
				SQL:  "SELECT * FROM test_model;",
				Args: nil,
			},
		},
		{
			name:    "where",
			builder: (&Select[TestModel]{}).Where(C("Age").Eq(18)),
			wantQuery: &Query{
				SQL:  "SELECT * FROM test_model WHERE age=?;",
				Args: []any{18},
			},
		},
		{
			name:    "not",
			builder: (&Select[TestModel]{}).Where(Not(C("Age").Eq(18))),
			wantQuery: &Query{
				SQL:  "SELECT * FROM test_model WHERE  NOT (age=?);",
				Args: []any{18},
			},
		},
		{
			name:    "and",
			builder: (&Select[TestModel]{}).Where(C("Age").Eq(18).And(C("FirstName").Eq("Tom"))),
			wantQuery: &Query{
				SQL:  "SELECT * FROM test_model WHERE (age=?) AND (first_name=?);",
				Args: []any{18, "Tom"},
			},
		},
		{
			name:    "or",
			builder: (&Select[TestModel]{}).Where(C("Age").Eq(18).Or(C("FirstName").Eq("Tom"))),
			wantQuery: &Query{
				SQL:  "SELECT * FROM test_model WHERE (age=?) OR (first_name=?);",
				Args: []any{18, "Tom"},
			},
		},
		{
			name:    "invalid column",
			builder: (&Select[TestModel]{}).Where(C("Age").Eq(18).Or(C("XXX").Eq("Tom"))),
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
