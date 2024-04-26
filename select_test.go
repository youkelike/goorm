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
			name:    "no from",
			builder: NewSelector[TestModel](db),
			wantQuery: &Query{
				SQL:  "SELECT * FROM test_model;",
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
			name:    "group by",
			builder: NewSelector[TestModel](db).Where(C("Age").Eq(18).Or(C("FirstName").Eq("Tom"))).GroupBy(C("Age")),
			wantQuery: &Query{
				SQL:  "SELECT * FROM test_model WHERE (age=?) OR (first_name=?) GROUP BY age;",
				Args: []any{18, "Tom"},
			},
		},
		{
			name:    "order by",
			builder: NewSelector[TestModel](db).Where(C("Age").Eq(18).Or(C("FirstName").Eq("Tom"))).OrderBy(C("Age").Asc()),
			wantQuery: &Query{
				SQL:  "SELECT * FROM test_model WHERE (age=?) OR (first_name=?) ORDER BY age ASC;",
				Args: []any{18, "Tom"},
			},
		},
		{
			name:    "offset and limit",
			builder: NewSelector[TestModel](db).Where(C("Age").Eq(18).Or(C("FirstName").Eq("Tom"))).Offset(10).Limit(10),
			wantQuery: &Query{
				SQL:  "SELECT * FROM test_model WHERE (age=?) OR (first_name=?) OFFSET 10, LIMIT 10;",
				Args: []any{18, "Tom"},
			},
		},
		{
			name:    "all",
			builder: NewSelector[TestModel](db).Where(C("Age").Eq(18).Or(C("FirstName").Eq("Tom"))).GroupBy(C("Age")).OrderBy(C("Age").Asc()).Offset(10).Limit(10),
			wantQuery: &Query{
				SQL:  "SELECT * FROM test_model WHERE (age=?) OR (first_name=?) GROUP BY age ORDER BY age ASC OFFSET 10, LIMIT 10;",
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
			s:    NewSelector[TestModel](db).Select(Avg(C("Age")).As("ag")),
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
			s:    NewSelector[TestModel](db).Select(Avg(C("Age"))),
			wantQuery: &Query{
				SQL: "SELECT AVG(age) FROM test_model;",
			},
		},
		{
			name: "Sum",
			s:    NewSelector[TestModel](db).Select(Sum(C("Age"))),
			wantQuery: &Query{
				SQL: "SELECT SUM(age) FROM test_model;",
			},
		},
		{
			name: "Sum with table",
			s:    NewSelector[TestModel](db).Select(Sum(TableOf(new(TestModel)).C("Age"))),
			wantQuery: &Query{
				SQL: "SELECT SUM(test_model.age) FROM test_model;",
			},
		},
		{
			name: "Sum with table alias",
			s:    NewSelector[TestModel](db).Select(Sum(TableOf(new(TestModel)).As("t").C("Age"))),
			wantQuery: &Query{
				SQL: "SELECT SUM(t.age) FROM test_model;",
			},
		},
		{
			name: "multiple aggregate",
			s:    NewSelector[TestModel](db).Select(Sum(C("Age")), Count(C("FirstName"))),
			wantQuery: &Query{
				SQL: "SELECT SUM(age),COUNT(first_name) FROM test_model;",
			},
		},
		{
			name:    "Sum invalid",
			s:       NewSelector[TestModel](db).Select(Sum(C("XXX"))),
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

func TestSelector_Join(t *testing.T) {
	db := memoryDB(t)
	type Order struct {
		Id        int
		UsingCol1 string
		UsingCol2 string
	}

	type OrderDetail struct {
		OrderId int
		ItemId  int
		Address string
		Price   int
	}

	type Item struct {
		Id int
	}

	testCases := []struct {
		name      string
		s         QueryBuilder
		wantQuery *Query
		wantErr   error
	}{
		{
			name: "join subquery",
			s: func() QueryBuilder {
				t1 := SubqueryOf(NewSelector[Order](db).Where(C("Id").Gt(1))).As("t1")
				t2 := TableOf(&OrderDetail{})
				t3 := t2.Join(t1).Using("UsingCol1", "UsingCol2")

				return NewSelector[Order](db).From(t3)
			}(),
			wantQuery: &Query{
				SQL:  "SELECT * FROM (order_detail JOIN (SELECT * FROM order WHERE id>?) AS t1 USING (using_col1,using_col2));",
				Args: []interface{}{1},
			},
		},
		{
			name: "subquery",
			s: func() QueryBuilder {
				t1 := SubqueryOf(NewSelector[Order](db).Where(C("Id").Gt(1))).As("t1")

				return NewSelector[Order](db).From(t1)
			}(),
			wantQuery: &Query{
				SQL:  "SELECT * FROM (SELECT * FROM order WHERE id>?) AS t1;",
				Args: []interface{}{1},
			},
		},
		{
			name: "specify table",
			s:    NewSelector[Order](db).From(TableOf(&OrderDetail{})),
			wantQuery: &Query{
				SQL: "SELECT * FROM order_detail;",
			},
		},
		{
			name: "join using",
			s: func() QueryBuilder {
				t1 := TableOf(&Order{})
				t2 := TableOf(&OrderDetail{})
				t3 := t1.Join(t2).Using("UsingCol1", "UsingCol2")

				return NewSelector[Order](db).From(t3)
			}(),
			wantQuery: &Query{
				SQL: "SELECT * FROM (order JOIN order_detail USING (using_col1,using_col2));",
			},
		},
		{
			name: "join on",
			s: func() QueryBuilder {
				t1 := TableOf(&Order{}).As("t1")
				t2 := TableOf(&OrderDetail{}).As("t2")
				t3 := t1.Join(t2).On(t1.C("Id").Eq(t2.C("OrderId")))

				return NewSelector[Order](db).From(t3)
			}(),
			wantQuery: &Query{
				SQL: "SELECT * FROM (order AS t1 JOIN order_detail AS t2 ON t1.id=t2.order_id);",
			},
		},
		{
			name: "left join",
			s: func() QueryBuilder {
				t1 := TableOf(&Order{}).As("t1")
				t2 := TableOf(&OrderDetail{}).As("t2")
				t3 := t1.LeftJoin(t2).On(t1.C("Id").Eq(t2.C("OrderId")))

				return NewSelector[Order](db).From(t3)
			}(),
			wantQuery: &Query{
				SQL: "SELECT * FROM (order AS t1 LEFT JOIN order_detail AS t2 ON t1.id=t2.order_id);",
			},
		},
		{
			name: "right join",
			s: func() QueryBuilder {
				t1 := TableOf(&Order{}).As("t1")
				t2 := TableOf(&OrderDetail{}).As("t2")
				t3 := t1.RightJoin(t2).On(t1.C("Id").Eq(t2.C("OrderId")))

				return NewSelector[Order](db).From(t3)
			}(),
			wantQuery: &Query{
				SQL: "SELECT * FROM (order AS t1 RIGHT JOIN order_detail AS t2 ON t1.id=t2.order_id);",
			},
		},
		{
			name: "join table",
			s: func() QueryBuilder {
				t1 := TableOf(&Order{}).As("t1")
				t2 := TableOf(&OrderDetail{}).As("t2")
				t3 := t1.Join(t2).On(t1.C("Id").Eq(t2.C("OrderId")))
				t4 := TableOf(&Item{}).As("t4")
				t5 := t3.Join(t4).On(t2.C("ItemId").Eq(t4.C("Id")))

				return NewSelector[Order](db).From(t5)
			}(),
			wantQuery: &Query{
				SQL: "SELECT * FROM ((order AS t1 JOIN order_detail AS t2 ON t1.id=t2.order_id) JOIN item AS t4 ON t2.item_id=t4.id);",
			},
		},
		{
			name: "table join",
			s: func() QueryBuilder {
				t1 := TableOf(&Order{}).As("t1")
				t2 := TableOf(&OrderDetail{}).As("t2")
				t3 := t1.Join(t2).On(t1.C("Id").Eq(t2.C("OrderId")))
				t4 := TableOf(&Item{}).As("t4")
				t5 := t4.Join(t3).On(t2.C("ItemId").Eq(t4.C("Id")))

				return NewSelector[Order](db).From(t5)
			}(),
			wantQuery: &Query{
				SQL: "SELECT * FROM (item AS t4 JOIN (order AS t1 JOIN order_detail AS t2 ON t1.id=t2.order_id) ON t2.item_id=t4.id);",
			},
		},
		{
			name: "right join with fields select",
			s: func() QueryBuilder {
				t1 := TableOf(&Order{}).As("t1")
				t2 := TableOf(&OrderDetail{}).As("t2")
				t3 := t1.RightJoin(t2).On(t1.C("Id").Eq(t2.C("OrderId")))

				return NewSelector[Order](db).Select(t1.C("Id"), t2.C("ItemId"), t2.C("Price"), t2.C("Address").As("addr")).From(t3)
			}(),
			wantQuery: &Query{
				SQL: "SELECT t1.id,t2.item_id,t2.price,t2.address AS addr FROM (order AS t1 RIGHT JOIN order_detail AS t2 ON t1.id=t2.order_id);",
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

func TestSelector_Scan(t *testing.T) {
	type Order struct {
		Id        int
		UserName  string
		UsingCol1 string
		UsingCol2 string
	}

	type OrderDetail struct {
		OrderId int
		ItemId  int
		Address string
		Price   int
	}

	type Result struct {
		Id      int
		ItemId  int
		Address string
		Price   int
	}

	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	db, err := OpenDB(mockDB)
	require.NoError(t, err)

	rows := sqlmock.NewRows([]string{"id", "item_id", "address"})
	rows.AddRow(1, 1, "guangzhou")
	mock.ExpectQuery("SELECT .*").WillReturnRows(rows)

	rows = sqlmock.NewRows([]string{"id", "item_id", "address", "price", "user_name"})
	rows.AddRow(1, 1, "guangzhou", 100, "alice")
	mock.ExpectQuery("SELECT .*").WillReturnRows(rows)

	rows = sqlmock.NewRows([]string{"id", "item_id", "address", "price"})
	rows.AddRow(1, 1, "guangzhou", 100)
	mock.ExpectQuery("SELECT .*").WillReturnRows(rows)

	testCases := []struct {
		name    string
		entity  any
		s       *Selector[Order]
		wantErr error
		wantRes []any
	}{
		{
			name:    "no pointer",
			entity:  Result{},
			wantErr: errs.ErrScanEntityValid,
			wantRes: nil,
		},
		{
			name: "struct fields more than query results",
			s: func() *Selector[Order] {
				t1 := TableOf(&Order{}).As("t1")
				t2 := TableOf(&OrderDetail{}).As("t2")
				t3 := t1.RightJoin(t2).On(t1.C("Id").Eq(t2.C("OrderId")))
				return NewSelector[Order](db).
					Select(t1.C("Id"), t2.C("ItemId"), t2.C("Price")).
					From(t3)
			}(),
			entity:  &Result{},
			wantErr: nil,
			wantRes: []any{
				&Result{
					Id:      1,
					ItemId:  1,
					Address: "guangzhou",
					Price:   0,
				},
			},
		},
		{
			name: "struct fields less than query results",
			s: func() *Selector[Order] {
				t1 := TableOf(&Order{}).As("t1")
				t2 := TableOf(&OrderDetail{}).As("t2")
				t3 := t1.RightJoin(t2).On(t1.C("Id").Eq(t2.C("OrderId")))
				return NewSelector[Order](db).
					Select(t1.C("Id"), t2.C("ItemId"), t2.C("Price"), t2.C("Address"), t1.C("UserName")).
					From(t3)
			}(),
			entity:  &Result{},
			wantErr: errs.NewUnknownColumn("user_name"),
		},
		{
			name: "default",
			s: func() *Selector[Order] {
				t1 := TableOf(&Order{}).As("t1")
				t2 := TableOf(&OrderDetail{}).As("t2")
				t3 := t1.RightJoin(t2).On(t1.C("Id").Eq(t2.C("OrderId")))
				return NewSelector[Order](db).
					Select(t1.C("Id"), t2.C("ItemId"), t2.C("Price"), t2.C("Address")).
					From(t3)
			}(),
			entity:  &Result{},
			wantErr: nil,
			wantRes: []any{
				&Result{
					Id:      1,
					ItemId:  1,
					Address: "guangzhou",
					Price:   100,
				},
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			res, err := tc.s.Scan(tc.entity)
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, tc.wantRes, res)
		})
	}
}
