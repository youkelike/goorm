package unsafe

import (
	"errors"
	"reflect"
	"unsafe"
)

type UnsafeAccessor struct {
	// 结构体内字段偏移量
	fields map[string]FieldMeta
	// 结构体起始地址
	address unsafe.Pointer
}

func NewUnsafeAccessor(entity any) *UnsafeAccessor {
	// 反射要结合 reflect，通过它拿到对象的起始地址和字段的偏移量、数据类型
	typ := reflect.TypeOf(entity)
	typ = typ.Elem()
	numField := typ.NumField()
	fields := make(map[string]FieldMeta, numField)
	for i := 0; i < numField; i++ {
		fd := typ.Field(i)
		fields[fd.Name] = FieldMeta{
			Offset: fd.Offset,
			Typ:    fd.Type,
		}
	}
	val := reflect.ValueOf(entity)
	return &UnsafeAccessor{
		fields: fields,
		// 对象的起始地址
		address: val.UnsafePointer(),
	}
}

func (a *UnsafeAccessor) Field(field string) (any, error) {
	fd, ok := a.fields[field]
	if !ok {
		return nil, errors.New("非法字段")
	}
	fdAddress := unsafe.Pointer(uintptr(a.address) + fd.Offset)
	// return *(*int)(fdAddress), nil
	return reflect.NewAt(fd.Typ, fdAddress).Elem().Interface(), nil
}

func (a *UnsafeAccessor) SetField(field string, val any) error {
	fd, ok := a.fields[field]
	if !ok {
		return errors.New("非法字段")
	}
	// 地址运算的结果最好转换成 unsafe.Pointer，因为垃圾回收有可能改变对象的地址，但 unsafe.Pointer 始终指向目标对象
	fdAddress := unsafe.Pointer(uintptr(a.address) + fd.Offset)
	// *(*int)(fdAddress) = val.(int)
	reflect.NewAt(fd.Typ, fdAddress).Elem().Set(reflect.ValueOf(val))
	return nil
}

type FieldMeta struct {
	Offset uintptr
	Typ    reflect.Type
}
