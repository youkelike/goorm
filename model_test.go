package orm

import (
	"database/sql"
	"reflect"
	"testing"

	"gitee.com/youkelike/go1/work/hw05/orm/internal/errs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_registry_Register(t *testing.T) {
	testCases := []struct {
		name      string
		entity    any
		fields    []*field
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
			fields: []*field{
				{
					colName: "id",
					goName:  "Id",
					typ:     reflect.TypeOf(int64(0)),
				},
				{
					colName: "first_name",
					goName:  "FirstName",
					typ:     reflect.TypeOf(""),
				},
				{
					colName: "age",
					goName:  "Age",
					typ:     reflect.TypeOf(int8(0)),
				},
				{
					colName: "last_name",
					goName:  "LastName",
					typ:     reflect.TypeOf(&sql.NullString{}),
				},
			},
			wantModel: &Model{
				tableName: "test_model",
			},
		},
		{
			name:   "struct",
			entity: TestModel{},
			fields: []*field{
				{
					colName: "id",
					goName:  "Id",
					typ:     reflect.TypeOf(int64(0)),
				},
				{
					colName: "first_name",
					goName:  "FirstName",
					typ:     reflect.TypeOf(""),
				},
				{
					colName: "age",
					goName:  "Age",
					typ:     reflect.TypeOf(int8(0)),
				},
				{
					colName: "last_name",
					goName:  "LastName",
					typ:     reflect.TypeOf(&sql.NullString{}),
				},
			},
			wantModel: &Model{
				tableName: "test_model",
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

			fieldMap := make(map[string]*field)
			columnMap := make(map[string]*field)
			for _, f := range tc.fields {
				fieldMap[f.goName] = f
				columnMap[f.colName] = f
			}
			tc.wantModel.fieldMap = fieldMap
			tc.wantModel.columnMap = columnMap
			assert.Equal(t, tc.wantModel, m)
		})
	}
}

func TestRegistry_get(t *testing.T) {
	testCases := []struct {
		name      string
		entity    any
		fields    []*field
		wantModel *Model
		wantErr   error
	}{
		{
			name:   "pointer",
			entity: &TestModel{},
			fields: []*field{
				{
					colName: "id",
					goName:  "Id",
					typ:     reflect.TypeOf(int64(0)),
				},
				{
					colName: "first_name",
					goName:  "FirstName",
					typ:     reflect.TypeOf(""),
				},
				{
					colName: "age",
					goName:  "Age",
					typ:     reflect.TypeOf(int8(0)),
				},
				{
					colName: "last_name",
					goName:  "LastName",
					typ:     reflect.TypeOf(&sql.NullString{}),
				},
			},
			wantModel: &Model{
				tableName: "test_model",
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
			fields: []*field{
				{
					colName: "first_name_t",
					goName:  "FirstName",
					typ:     reflect.TypeOf(""),
				},
			},
			wantModel: &Model{
				tableName: "tag_table",
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
			fields: []*field{
				{
					colName: "first_name",
					goName:  "FirstName",
					typ:     reflect.TypeOf(""),
				},
			},
			wantModel: &Model{
				tableName: "tag_table",
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
			fields: []*field{
				{
					colName: "first_name",
					goName:  "FirstName",
					typ:     reflect.TypeOf(""),
				},
			},
			wantModel: &Model{
				tableName: "tag_table",
			},
		},
		{
			name:   "table name",
			entity: CustomTableName{},
			fields: []*field{
				{
					colName: "first_name",
					goName:  "FirstName",
					typ:     reflect.TypeOf(""),
				},
			},
			wantModel: &Model{
				tableName: "custom_table_name",
			},
		},
		{
			name:   "table name ptr",
			entity: &CustomTableNamePtr{},
			fields: []*field{
				{
					colName: "first_name",
					goName:  "FirstName",
					typ:     reflect.TypeOf(""),
				},
			},
			wantModel: &Model{
				tableName: "custom_table_name_ptr",
			},
		},
		{
			name:   "empty table name",
			entity: &EmptyTableName{},
			fields: []*field{
				{
					colName: "first_name",
					goName:  "FirstName",
					typ:     reflect.TypeOf(""),
				},
			},
			wantModel: &Model{
				tableName: "empty_table_name",
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

			fieldMap := make(map[string]*field)
			columnMap := make(map[string]*field)
			for _, f := range tc.fields {
				fieldMap[f.goName] = f
				columnMap[f.colName] = f
			}
			tc.wantModel.fieldMap = fieldMap
			tc.wantModel.columnMap = columnMap

			assert.Equal(t, tc.wantModel, m)
			typ := reflect.TypeOf(tc.entity)
			for typ.Kind() == reflect.Pointer {
				typ = typ.Elem()
			}
			cache, ok := r.models.Load(typ)
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
	m, err := r.Register(&TestModel{}, ModelWithTableName("test_model_xxx"))
	require.NoError(t, err)
	assert.Equal(t, "test_model_xxx", m.tableName)
}

func TestModelWithColumnName(t *testing.T) {
	testCases := []struct {
		name        string
		field       string
		colName     string
		wantColName string
		wantErr     error
	}{
		{
			name:        "column name",
			field:       "FirstName",
			colName:     "first_name_xxx",
			wantColName: "first_name_xxx",
		},
		{
			name:    "invalid column name",
			field:   "XXX",
			colName: "first_name_xxx",
			wantErr: errs.NewUnknownField("XXX"),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			r := NewRegistry()
			m, err := r.Register(&TestModel{}, ModelWithColumnName(tc.field, tc.colName))
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			fd, ok := m.fieldMap[tc.field]
			assert.True(t, ok)
			assert.Equal(t, tc.colName, fd.colName)
		})
	}
}
