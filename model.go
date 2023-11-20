package orm

import (
	"fmt"
	"reflect"
	"strings"
	"sync"
	"unicode"

	"gitee.com/youkelike/go1/work/hw05/orm/internal/errs"
)

var (
	tagColumn = "column"
)

type Registry interface {
	// 获取已解析好的映射
	Get(val any) (*Model, error)
	// 解析、获得结构体到表的映射
	Register(val any, opts ...ModelOption) (*Model, error)
}

// 通过反射解析出来的结构体到表的映射
type Model struct {
	// 表名
	tableName string
	// 字段名到字段的映射
	fieldMap map[string]*field
	// 列名到字段的映射
	columnMap map[string]*field
}

type field struct {
	goName  string
	colName string
	typ     reflect.Type
}

type ModelOption func(*Model) error

func ModelWithTableName(tableName string) ModelOption {
	return func(m *Model) error {
		m.tableName = tableName
		return nil
	}
}

func ModelWithColumnName(field, colname string) ModelOption {
	return func(m *Model) error {
		fd, ok := m.fieldMap[field]
		if !ok {
			return errs.NewUnknownField(field)
		}
		fd.colName = colname
		return nil
	}
}

// 主要用于集中管理、注册解析好的模型
type registry struct {
	models sync.Map
}

func NewRegistry() *registry {
	return &registry{}
}

func (r *registry) Get(val any) (*Model, error) {
	typ := reflect.TypeOf(val)
	for typ.Kind() == reflect.Pointer {
		typ = typ.Elem()
	}
	m, ok := r.models.Load(typ)
	if ok {
		return m.(*Model), nil
	}

	fmt.Println("register")
	mm, err := r.Register(val)
	if err != nil {
		return nil, err
	}
	// r.models.Store(typ, mm)
	return mm, nil
}

// func (r *registry) get(val any) (*model, error) {
// 	typ := reflect.TypeOf(val)
// 	m, ok := r.models[typ]
// 	if ok {
// 		return m, nil
// 	}

// 	m, err := r.parseModel(val)
// 	if err != nil {
// 		return nil, err
// 	}
// 	r.models[typ] = m
// 	return m, nil
// }

// 对于表名的解析顺序：结构体名、表名接口、表名 option
// 对于表中列名的解析顺序：字段名、字段 tag、字段名 option
func (r *registry) Register(entity any, opts ...ModelOption) (*Model, error) {
	typ := reflect.TypeOf(entity)
	for typ.Kind() == reflect.Pointer {
		typ = typ.Elem()
	}

	if typ.Kind() != reflect.Struct {
		return nil, errs.ErrPointerOnly
	}

	numField := typ.NumField()
	fieldMap := make(map[string]*field, numField)
	columnMap := make(map[string]*field, numField)
	for i := 0; i < numField; i++ {
		fd := typ.Field(i)

		// 从 tag 解析出表中列名
		pair, err := r.parseTag(fd.Tag)
		if err != nil {
			return nil, err
		}
		colName := pair[tagColumn]
		if colName == "" {
			colName = underscoreName(fd.Name)
		}

		f := &field{
			goName:  fd.Name,
			colName: colName,
			typ:     fd.Type,
		}
		fieldMap[fd.Name] = f
		columnMap[colName] = f
	}

	var tableName string
	if tbl, ok := entity.(TableName); ok {
		tableName = tbl.TableName()
	}
	if tableName == "" {
		tableName = underscoreName(typ.Name())
	}

	m := &Model{
		tableName: tableName,
		fieldMap:  fieldMap,
		columnMap: columnMap,
	}
	for _, opt := range opts {
		err := opt(m)
		if err != nil {
			return nil, err
		}
	}
	r.models.Store(typ, m)
	// r.models.Store(reflect.TypeOf(entity), m)

	return m, nil
}

// tag 是这个格式的：`orm:"column=id,xx=xx" xxx:"xx"`
func (r *registry) parseTag(tag reflect.StructTag) (map[string]string, error) {
	ormTag, ok := tag.Lookup("orm")
	if !ok {
		return map[string]string{}, nil
	}
	pairs := strings.Split(ormTag, ",")
	res := make(map[string]string, len(pairs))
	for _, pair := range pairs {
		segs := strings.Split(pair, "=")
		if len(segs) != 2 {
			return nil, errs.NewInvalidTagContent(pair)
		}
		key := segs[0]
		val := segs[1]
		res[key] = val
	}
	return res, nil
}

func underscoreName(tableName string) string {
	var buf []byte
	for i, v := range tableName {
		if unicode.IsUpper(v) {
			if i != 0 {
				buf = append(buf, '_')
			}
			buf = append(buf, byte(unicode.ToLower(v)))
		} else {
			buf = append(buf, byte(v))
		}
	}
	return string(buf)
}
