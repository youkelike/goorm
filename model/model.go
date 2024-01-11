package model

import (
	"reflect"
	"strings"
	"sync"
	"unicode"

	"gitee.com/youkelike/orm/internal/errs"
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
	TableName string
	// 字段名到字段的映射
	FieldMap map[string]*Field
	// 列名到字段的映射
	ColumnMap map[string]*Field
	Fields    []*Field
}

type Field struct {
	GoName  string
	ColName string
	Typ     reflect.Type
	Offset  uintptr
}

// 用接口的方式提供自定义表名的途径
type TableName interface {
	TableName() string
}

type ModelOption func(*Model) error

func WithTableName(TableName string) ModelOption {
	return func(m *Model) error {
		m.TableName = TableName
		return nil
	}
}

func WithColumnName(Field, colname string) ModelOption {
	return func(m *Model) error {
		fd, ok := m.FieldMap[Field]
		if !ok {
			return errs.NewUnknownField(Field)
		}
		// fd 的类型是 *Field，这里改了，m.ColumnMap 和 m.Fields 中都会改
		fd.ColName = colname
		return nil
	}
}

type registry struct {
	models sync.Map
}

func NewRegistry() Registry {
	return &registry{}
}

func (r *registry) Get(val any) (*Model, error) {
	Typ := reflect.TypeOf(val)
	m, ok := r.models.Load(Typ)
	if ok {
		return m.(*Model), nil
	}

	mm, err := r.Register(val)
	if err != nil {
		return nil, err
	}
	// r.models.Store(Typ, mm)
	return mm, nil
}

// func (r *registry) get(val any) (*model, error) {
// 	Typ := reflect.TypeOf(val)
// 	m, ok := r.models[Typ]
// 	if ok {
// 		return m, nil
// 	}

// 	m, err := r.parseModel(val)
// 	if err != nil {
// 		return nil, err
// 	}
// 	r.models[Typ] = m
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
	FieldMap := make(map[string]*Field, numField)
	ColumnMap := make(map[string]*Field, numField)
	Fields := make([]*Field, 0, numField)
	for i := 0; i < numField; i++ {
		fd := typ.Field(i)

		// 从 struct 字段名、tag 解析出表字段名
		pair, err := r.parseTag(fd.Tag)
		if err != nil {
			return nil, err
		}
		ColName := pair[tagColumn]
		if ColName == "" {
			ColName = underscoreName(fd.Name)
		}

		f := &Field{
			GoName:  fd.Name,
			ColName: ColName,
			Typ:     fd.Type,
			Offset:  fd.Offset,
		}
		FieldMap[fd.Name] = f
		ColumnMap[ColName] = f
		Fields = append(Fields, f)
	}

	var tableName string
	if tbl, ok := entity.(TableName); ok {
		tableName = tbl.TableName()
	}
	if tableName == "" {
		tableName = underscoreName(typ.Name())
	}

	m := &Model{
		TableName: tableName,
		FieldMap:  FieldMap,
		ColumnMap: ColumnMap,
		Fields:    Fields,
	}
	for _, opt := range opts {
		err := opt(m)
		if err != nil {
			return nil, err
		}
	}
	r.models.Store(reflect.TypeOf(entity), m)

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

func underscoreName(TableName string) string {
	var buf []byte
	for i, v := range TableName {
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
