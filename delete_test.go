package orm

import (
	"testing"

	"gitee.com/youkelike/orm/internal/errs"
	"github.com/stretchr/testify/assert"
)

func TestBuild(t *testing.T) {
	testCases := []struct {
		name      string
		builder   QueryBuilder
		wantQuery *Query
		wantErr   error
	}{
		{
			name:    "invalid column",
			builder: (&Delete[TestModel]{}).Where(C("FirstName").Eq("Tom").Or(C("XXX").Eq(1))),
			wantErr: errs.NewUnknownField("XXX"),
		},
		{
			name:    "or",
			builder: (&Delete[TestModel]{}).Where(C("FirstName").Eq("Tom").Or(C("Age").Eq(18))),
			wantQuery: &Query{
				SQL:  "DELETE FROM test_model WHERE (first_name=?) OR (age=?);",
				Args: []any{"Tom", 18},
			},
		},
		{
			name:    "and",
			builder: (&Delete[TestModel]{}).Where(C("FirstName").Eq("Tom").And(C("Age").Eq(18))),
			wantQuery: &Query{
				SQL:  "DELETE FROM test_model WHERE (first_name=?) AND (age=?);",
				Args: []any{"Tom", 18},
			},
		},
		{
			name:    "not",
			builder: (&Delete[TestModel]{}).Where(Not(C("FirstName").Eq("Tom"))),
			wantQuery: &Query{
				SQL:  "DELETE FROM test_model WHERE  NOT (first_name=?);",
				Args: []any{"Tom"},
			},
		},
		{
			name:    "where",
			builder: (&Delete[TestModel]{}).Where(C("FirstName").Eq("Tom")),
			wantQuery: &Query{
				SQL:  "DELETE FROM test_model WHERE first_name=?;",
				Args: []any{"Tom"},
			},
		},
		{
			name:    "empty where",
			builder: (&Delete[TestModel]{}).Where(),
			wantQuery: &Query{
				SQL: "DELETE FROM test_model;",
			},
		},
		{
			name:    "from db",
			builder: (&Delete[TestModel]{}).From("test_db.test_model"),
			wantQuery: &Query{
				SQL: "DELETE FROM test_db.test_model;",
			},
		},
		{
			name:    "empty from",
			builder: (&Delete[TestModel]{}).From(""),
			wantQuery: &Query{
				SQL: "DELETE FROM test_model;",
			},
		},
		{
			name:    "use from",
			builder: (&Delete[TestModel]{}).From("test_model"),
			wantQuery: &Query{
				SQL: "DELETE FROM test_model;",
			},
		},
		{
			name:    "no from",
			builder: &Delete[TestModel]{},
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
