package model

import (
	"database/sql"
	"reflect"
	"testing"

	"gitee.com/youkelike/orm/internal/errs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_registry_Register(t *testing.T) {
	testCases := []struct {
		name      string
		entity    any
		wantModel *Model
		wantErr   error
	}{
		{
			name:    "map",
			entity:  map[string]string{},
			wantErr: errs.ErrPointerOnly,
		},
		{
			name:    "slice",
			entity:  []int{},
			wantErr: errs.ErrPointerOnly,
		},
		{
			name:   "pointer",
			entity: &TestModel{},
			wantModel: &Model{
				TableName: "test_model",
				FieldMap: map[string]*Field{
					"Id": {
						ColName: "id",
						GoName:  "Id",
						Typ:     reflect.TypeOf(int64(0)),
					},
					"FirstName": {
						ColName: "first_name",
						GoName:  "FirstName",
						Typ:     reflect.TypeOf(""),
						Offset:  8,
					},
					"Age": {
						ColName: "age",
						GoName:  "Age",
						Typ:     reflect.TypeOf(int8(0)),
						Offset:  24,
					},
					"LastName": {
						ColName: "last_name",
						GoName:  "LastName",
						Typ:     reflect.TypeOf(&sql.NullString{}),
						Offset:  32,
					},
				},
				ColumnMap: map[string]*Field{
					"id": {
						ColName: "id",
						GoName:  "Id",
						Typ:     reflect.TypeOf(int64(0)),
					},
					"first_name": {
						ColName: "first_name",
						GoName:  "FirstName",
						Typ:     reflect.TypeOf(""),
						Offset:  8,
					},
					"age": {
						ColName: "age",
						GoName:  "Age",
						Typ:     reflect.TypeOf(int8(0)),
						Offset:  24,
					},
					"last_name": {
						ColName: "last_name",
						GoName:  "LastName",
						Typ:     reflect.TypeOf(&sql.NullString{}),
						Offset:  32,
					},
				},
				Fields: []*Field{
					{
						ColName: "id",
						GoName:  "Id",
						Typ:     reflect.TypeOf(int64(0)),
					},
					{
						ColName: "first_name",
						GoName:  "FirstName",
						Typ:     reflect.TypeOf(""),
						Offset:  8,
					},
					{
						ColName: "age",
						GoName:  "Age",
						Typ:     reflect.TypeOf(int8(0)),
						Offset:  24,
					},
					{
						ColName: "last_name",
						GoName:  "LastName",
						Typ:     reflect.TypeOf(&sql.NullString{}),
						Offset:  32,
					},
				},
			},
		},
		{
			name:   "struct",
			entity: TestModel{},
			wantModel: &Model{
				TableName: "test_model",
				FieldMap: map[string]*Field{
					"Id": {
						ColName: "id",
						GoName:  "Id",
						Typ:     reflect.TypeOf(int64(0)),
					},
					"FirstName": {
						ColName: "first_name",
						GoName:  "FirstName",
						Typ:     reflect.TypeOf(""),
						Offset:  8,
					},
					"Age": {
						ColName: "age",
						GoName:  "Age",
						Typ:     reflect.TypeOf(int8(0)),
						Offset:  24,
					},
					"LastName": {
						ColName: "last_name",
						GoName:  "LastName",
						Typ:     reflect.TypeOf(&sql.NullString{}),
						Offset:  32,
					},
				},
				ColumnMap: map[string]*Field{
					"id": {
						ColName: "id",
						GoName:  "Id",
						Typ:     reflect.TypeOf(int64(0)),
					},
					"first_name": {
						ColName: "first_name",
						GoName:  "FirstName",
						Typ:     reflect.TypeOf(""),
						Offset:  8,
					},
					"age": {
						ColName: "age",
						GoName:  "Age",
						Typ:     reflect.TypeOf(int8(0)),
						Offset:  24,
					},
					"last_name": {
						ColName: "last_name",
						GoName:  "LastName",
						Typ:     reflect.TypeOf(&sql.NullString{}),
						Offset:  32,
					},
				},
				Fields: []*Field{
					{
						ColName: "id",
						GoName:  "Id",
						Typ:     reflect.TypeOf(int64(0)),
					},
					{
						ColName: "first_name",
						GoName:  "FirstName",
						Typ:     reflect.TypeOf(""),
						Offset:  8,
					},
					{
						ColName: "age",
						GoName:  "Age",
						Typ:     reflect.TypeOf(int8(0)),
						Offset:  24,
					},
					{
						ColName: "last_name",
						GoName:  "LastName",
						Typ:     reflect.TypeOf(&sql.NullString{}),
						Offset:  32,
					},
				},
			},
		},
	}

	r := &registry{}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			m, err := r.Register(tc.entity)
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, tc.wantModel, m)
		})
	}
}

func TestRegistry_get(t *testing.T) {
	testCases := []struct {
		name      string
		entity    any
		wantModel *Model
		wantErr   error
	}{
		{
			name:   "pointer",
			entity: &TestModel{},
			wantModel: &Model{
				TableName: "test_model",
				FieldMap: map[string]*Field{
					"Id": {
						ColName: "id",
						GoName:  "Id",
						Typ:     reflect.TypeOf(int64(0)),
					},
					"FirstName": {
						ColName: "first_name",
						GoName:  "FirstName",
						Typ:     reflect.TypeOf(""),
						Offset:  8,
					},
					"Age": {
						ColName: "age",
						GoName:  "Age",
						Typ:     reflect.TypeOf(int8(0)),
						Offset:  24,
					},
					"LastName": {
						ColName: "last_name",
						GoName:  "LastName",
						Typ:     reflect.TypeOf(&sql.NullString{}),
						Offset:  32,
					},
				},
				ColumnMap: map[string]*Field{
					"id": {
						ColName: "id",
						GoName:  "Id",
						Typ:     reflect.TypeOf(int64(0)),
					},
					"first_name": {
						ColName: "first_name",
						GoName:  "FirstName",
						Typ:     reflect.TypeOf(""),
						Offset:  8,
					},
					"age": {
						ColName: "age",
						GoName:  "Age",
						Typ:     reflect.TypeOf(int8(0)),
						Offset:  24,
					},
					"last_name": {
						ColName: "last_name",
						GoName:  "LastName",
						Typ:     reflect.TypeOf(&sql.NullString{}),
						Offset:  32,
					},
				},
				Fields: []*Field{
					{
						ColName: "id",
						GoName:  "Id",
						Typ:     reflect.TypeOf(int64(0)),
					},
					{
						ColName: "first_name",
						GoName:  "FirstName",
						Typ:     reflect.TypeOf(""),
						Offset:  8,
					},
					{
						ColName: "age",
						GoName:  "Age",
						Typ:     reflect.TypeOf(int8(0)),
						Offset:  24,
					},
					{
						ColName: "last_name",
						GoName:  "LastName",
						Typ:     reflect.TypeOf(&sql.NullString{}),
						Offset:  32,
					},
				},
			},
		},
		{
			name: "tag",
			entity: func() any {
				type TagTable struct {
					FirstName string `orm:"column=first_name_t"`
				}
				return &TagTable{}
			}(),
			wantModel: &Model{
				TableName: "tag_table",
				FieldMap: map[string]*Field{
					"FirstName": {
						ColName: "first_name_t",
						GoName:  "FirstName",
						Typ:     reflect.TypeOf(""),
					},
				},
				ColumnMap: map[string]*Field{
					"first_name_t": {
						ColName: "first_name_t",
						GoName:  "FirstName",
						Typ:     reflect.TypeOf(""),
					},
				},
				Fields: []*Field{
					{
						ColName: "first_name_t",
						GoName:  "FirstName",
						Typ:     reflect.TypeOf(""),
					},
				},
			},
		},
		{
			name: "empty tag",
			entity: func() any {
				type TagTable struct {
					FirstName string `orm:"column="`
				}
				return &TagTable{}
			}(),
			wantModel: &Model{
				TableName: "tag_table",
				FieldMap: map[string]*Field{
					"FirstName": {
						ColName: "first_name",
						GoName:  "FirstName",
						Typ:     reflect.TypeOf(""),
					},
				},
				ColumnMap: map[string]*Field{
					"first_name": {
						ColName: "first_name",
						GoName:  "FirstName",
						Typ:     reflect.TypeOf(""),
					},
				},
				Fields: []*Field{
					{
						ColName: "first_name",
						GoName:  "FirstName",
						Typ:     reflect.TypeOf(""),
					},
				},
			},
		},
		{
			name: "wrong tag",
			entity: func() any {
				type TagTable struct {
					FirstName string `orm:"column"`
				}
				return &TagTable{}
			}(),
			wantErr: errs.NewInvalidTagContent("column"),
		},
		{
			name: "ignore tag",
			entity: func() any {
				type TagTable struct {
					FirstName string `orm:"xx=abc"`
				}
				return &TagTable{}
			}(),
			wantModel: &Model{
				TableName: "tag_table",
				FieldMap: map[string]*Field{
					"FirstName": {
						ColName: "first_name",
						GoName:  "FirstName",
						Typ:     reflect.TypeOf(""),
					},
				},
				ColumnMap: map[string]*Field{
					"first_name": {
						ColName: "first_name",
						GoName:  "FirstName",
						Typ:     reflect.TypeOf(""),
					},
				},
				Fields: []*Field{
					{
						ColName: "first_name",
						GoName:  "FirstName",
						Typ:     reflect.TypeOf(""),
					},
				},
			},
		},
		{
			name:   "table name",
			entity: CustomTableName{},
			wantModel: &Model{
				TableName: "custom_table_name",
				FieldMap: map[string]*Field{
					"FirstName": {
						ColName: "first_name",
						GoName:  "FirstName",
						Typ:     reflect.TypeOf(""),
					},
				},
				ColumnMap: map[string]*Field{
					"first_name": {
						ColName: "first_name",
						GoName:  "FirstName",
						Typ:     reflect.TypeOf(""),
					},
				},
				Fields: []*Field{
					{
						ColName: "first_name",
						GoName:  "FirstName",
						Typ:     reflect.TypeOf(""),
					},
				},
			},
		},
		{
			name:   "table name ptr",
			entity: &CustomTableNamePtr{},
			wantModel: &Model{
				TableName: "custom_table_name_ptr",
				FieldMap: map[string]*Field{
					"FirstName": {
						ColName: "first_name",
						GoName:  "FirstName",
						Typ:     reflect.TypeOf(""),
					},
				},
				ColumnMap: map[string]*Field{
					"first_name": {
						ColName: "first_name",
						GoName:  "FirstName",
						Typ:     reflect.TypeOf(""),
					},
				},
				Fields: []*Field{
					{
						ColName: "first_name",
						GoName:  "FirstName",
						Typ:     reflect.TypeOf(""),
					},
				},
			},
		},
		{
			name:   "empty table name",
			entity: &EmptyTableName{},
			wantModel: &Model{
				TableName: "empty_table_name",
				FieldMap: map[string]*Field{
					"FirstName": {
						ColName: "first_name",
						GoName:  "FirstName",
						Typ:     reflect.TypeOf(""),
					},
				},
				ColumnMap: map[string]*Field{
					"first_name": {
						ColName: "first_name",
						GoName:  "FirstName",
						Typ:     reflect.TypeOf(""),
					},
				},
				Fields: []*Field{
					{
						ColName: "first_name",
						GoName:  "FirstName",
						Typ:     reflect.TypeOf(""),
					},
				},
			},
		},
	}

	r := NewRegistry()
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			m, err := r.Get(tc.entity)
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, tc.wantModel, m)
			typ := reflect.TypeOf(tc.entity)
			cache, ok := r.(*registry).models.Load(typ)
			assert.True(t, ok)
			assert.Equal(t, tc.wantModel, cache)
		})
	}
}

type CustomTableName struct {
	FirstName string
}

func (c CustomTableName) TableName() string {
	return "custom_table_name"
}

type CustomTableNamePtr struct {
	FirstName string
}

func (c *CustomTableNamePtr) TableName() string {
	return "custom_table_name_ptr"
}

type EmptyTableName struct {
	FirstName string
}

func (c EmptyTableName) TableName() string {
	return ""
}

func TestModelWithTableName(t *testing.T) {
	r := NewRegistry()
	m, err := r.Register(&TestModel{}, WithTableName("test_model_xxx"))
	require.NoError(t, err)
	assert.Equal(t, "test_model_xxx", m.TableName)
}

func TestModelWithColumnName(t *testing.T) {
	testCases := []struct {
		name        string
		Field       string
		ColName     string
		wantColName string
		wantErr     error
	}{
		{
			name:        "column name",
			Field:       "FirstName",
			ColName:     "first_name_xxx",
			wantColName: "first_name_xxx",
		},
		{
			name:    "invalid column name",
			Field:   "XXX",
			ColName: "first_name_xxx",
			wantErr: errs.NewUnknownField("XXX"),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			r := NewRegistry()
			m, err := r.Register(&TestModel{}, WithColumnName(tc.Field, tc.ColName))
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			fd, ok := m.FieldMap[tc.Field]
			assert.True(t, ok)
			assert.Equal(t, tc.ColName, fd.ColName)
		})
	}
}

type TestModel struct {
	Id        int64
	FirstName string
	Age       int8
	LastName  *sql.NullString
}
