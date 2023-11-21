package orm

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	"gitee.com/youkelike/orm/internal/errs"
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
			name:    "offset and limit",
			builder: NewSelector[TestModel](db).Offset(1).Limit(10),
			wantQuery: &Query{
				SQL: "SELECT * FROM test_model OFFSET 1, LIMIT 10;",
			},
		},
		{
			name:    "offset",
			builder: NewSelector[TestModel](db).Offset(10),
			wantQuery: &Query{
				SQL: "SELECT * FROM test_model OFFSET 10;",
			},
		},
		{
			name:    "limit",
			builder: NewSelector[TestModel](db).Limit(10),
			wantQuery: &Query{
				SQL: "SELECT * FROM test_model LIMIT 10;",
			},
		},
		{
			name:    "order by wrong",
			builder: NewSelector[TestModel](db).OrderBy(C("Id")),
			wantErr: errs.ErrNoOrderByVerb,
		},
		{
			name:    "order by 2",
			builder: NewSelector[TestModel](db).OrderBy(C("Id").Desc(), C("Age").Asc()),
			wantQuery: &Query{
				SQL: "SELECT * FROM test_model ORDER BY id DESC,age ASC;",
			},
		},
		{
			name:    "order by DESC",
			builder: NewSelector[TestModel](db).OrderBy(C("Id").Desc()),
			wantQuery: &Query{
				SQL: "SELECT * FROM test_model ORDER BY id DESC;",
			},
		},
		{
			name:    "order by ASC",
			builder: NewSelector[TestModel](db).OrderBy(C("Id").Asc()),
			wantQuery: &Query{
				SQL: "SELECT * FROM test_model ORDER BY id ASC;",
			},
		},
		{
			name:    "having without group by",
			builder: NewSelector[TestModel](db).Having(C("XXX").Eq(1)),
			wantErr: errs.ErrNoGroupUseHaving,
		},
		{
			name:    "having  invalid column",
			builder: NewSelector[TestModel](db).GroupBy(C("Id")).Having(C("XXX").Eq(1)),
			wantErr: errs.NewUnknownField("XXX"),
		},
		{
			name:    "having 2",
			builder: NewSelector[TestModel](db).GroupBy(C("Id")).Having(C("Id").Gt(1).And(C("Age").Lt(20))),
			wantQuery: &Query{
				SQL:  "SELECT * FROM test_model GROUP BY id HAVING (id>?) AND (age<?);",
				Args: []any{1, 20},
			},
		},
		{
			name:    "having",
			builder: NewSelector[TestModel](db).GroupBy(C("Id")).Having(C("Id").Gt(1)),
			wantQuery: &Query{
				SQL:  "SELECT * FROM test_model GROUP BY id HAVING id>?;",
				Args: []any{1},
			},
		},
		{
			name:    "group by invalid column",
			builder: NewSelector[TestModel](db).GroupBy(C("Id"), C("XXX")),
			wantErr: errs.NewUnknownField("XXX"),
		},
		{
			name:    "group by 2",
			builder: NewSelector[TestModel](db).GroupBy(C("Id"), C("Age")),
			wantQuery: &Query{
				SQL:  "SELECT * FROM test_model GROUP BY id,age;",
				Args: nil,
			},
		},
		{
			name:    "group by",
			builder: NewSelector[TestModel](db).GroupBy(C("Id")),
			wantQuery: &Query{
				SQL:  "SELECT * FROM test_model GROUP BY id;",
				Args: nil,
			},
		},
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

	mock.ExpectQuery("SELECT .*").WillReturnError(errs.ErrNoRows)

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
			wantErr: errs.ErrNoRows,
		},
		{
			name:    "now rows",
			s:       NewSelector[TestModel](db).Where(C("Id").Eq(1)),
			wantErr: errs.ErrNoRows,
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

func TestSelector_Select(t *testing.T) {
	mockDB, _, err := sqlmock.New()
	require.NoError(t, err)
	db, err := OpenDB(mockDB)
	require.NoError(t, err)
	defer mockDB.Close()

	testCases := []struct {
		name      string
		s         QueryBuilder
		wantQuery *Query
		wantErr   error
	}{
		{
			name: "alias in where",
			s:    NewSelector[TestModel](db).Where(C("Age").As("ag").Eq(18)),
			wantQuery: &Query{
				SQL:  "SELECT * FROM test_model WHERE age=?;",
				Args: []any{18},
			},
		},
		{
			name: "Avg alias",
			s:    NewSelector[TestModel](db).Select(Avg("Age").As("ag")),
			wantQuery: &Query{
				SQL: "SELECT AVG(age) AS ag FROM test_model;",
			},
		},
		{
			name: "alias columns",
			s:    NewSelector[TestModel](db).Select(C("FirstName").As("fname"), C("Age")),
			wantQuery: &Query{
				SQL: "SELECT first_name AS fname,age FROM test_model;",
			},
		},
		{
			name:    "invalid columns",
			s:       NewSelector[TestModel](db).Select(C("XXX")),
			wantErr: errs.NewUnknownField("XXX"),
		},
		{
			name: "multiple columns",
			s:    NewSelector[TestModel](db).Select(C("FirstName"), C("Age")),
			wantQuery: &Query{
				SQL: "SELECT first_name,age FROM test_model;",
			},
		},
		{
			name: "Avg",
			s:    NewSelector[TestModel](db).Select(Avg("Age")),
			wantQuery: &Query{
				SQL: "SELECT AVG(age) FROM test_model;",
			},
		},
		{
			name: "Sum",
			s:    NewSelector[TestModel](db).Select(Sum("Age")),
			wantQuery: &Query{
				SQL: "SELECT SUM(age) FROM test_model;",
			},
		},
		{
			name: "multiple aggregate",
			s:    NewSelector[TestModel](db).Select(Sum("Age"), Count("FirstName")),
			wantQuery: &Query{
				SQL: "SELECT SUM(age),COUNT(first_name) FROM test_model;",
			},
		},
		{
			name:    "Sum invalid",
			s:       NewSelector[TestModel](db).Select(Sum("XXX")),
			wantErr: errs.NewUnknownField("XXX"),
		},
		{
			name: "raw expression",
			s:    NewSelector[TestModel](db).Select(Raw("COUNT(DISTINCT first_name)")),
			wantQuery: &Query{
				SQL: "SELECT COUNT(DISTINCT first_name) FROM test_model;",
			},
		},
		{
			name: "raw expression as predicate",
			s:    NewSelector[TestModel](db).Where(Raw("age>?", 18).AsPredicate()),
			wantQuery: &Query{
				SQL:  "SELECT * FROM test_model WHERE (age>?);",
				Args: []any{18},
			},
		},
		{
			name: "raw expression used in predicate",
			s:    NewSelector[TestModel](db).Where(C("Id").Eq(Raw("age+?", 1))),
			wantQuery: &Query{
				SQL:  "SELECT * FROM test_model WHERE id=(age+?);",
				Args: []any{1},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			q, err := tc.s.Build()
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, tc.wantQuery, q)
		})
	}
}
