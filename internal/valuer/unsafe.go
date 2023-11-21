package valuer

import (
	"database/sql"
	"reflect"
	"unsafe"

	"gitee.com/youkelike/orm/internal/errs"
	"gitee.com/youkelike/orm/model"
)

type unsafeValue struct {
	model   *model.Model
	address unsafe.Pointer
}

var _ Creator = NewUnsafeValue

func NewUnsafeValue(model *model.Model, val any) Value {
	return unsafeValue{
		model:   model,
		address: reflect.ValueOf(val).UnsafePointer(),
	}
}

func (r unsafeValue) Field(name string) (any, error) {
	fd, ok := r.model.FieldMap[name]
	if !ok {
		return nil, errs.NewUnknownColumn(name)
	}
	fdAddress := unsafe.Pointer(uintptr(r.address) + fd.Offset)

	// 把一段内存地址以指定的数据类型进行解读
	val := reflect.NewAt(fd.Typ, fdAddress)
	// 获取解析后的值
	return val.Elem().Interface(), nil
}

func (r unsafeValue) SetColumns(rows *sql.Rows) error {
	cs, err := rows.Columns()
	if err != nil {
		return err
	}

	var vals []any
	// 结构体起始地址
	// address := reflect.ValueOf(r.val).UnsafePointer()
	for _, c := range cs {
		fd, ok := r.model.ColumnMap[c]
		if !ok {
			return errs.NewUnknownColumn(c)
		}
		// 结构体中字段的地址
		fdAddress := unsafe.Pointer(uintptr(r.address) + fd.Offset)
		// 在一段指定的地址上，创建一个特定类型的指针，指向这段地址
		// 后续修改这个指针的值就是在修改这段内存地址上的数据
		val := reflect.NewAt(fd.Typ, fdAddress)
		vals = append(vals, val.Interface())
	}
	err = rows.Scan(vals...)
	return err
}
