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
		fields:  fields,
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
	fdAddress := unsafe.Pointer(uintptr(a.address) + fd.Offset)
	// *(*int)(fdAddress) = val.(int)
	reflect.NewAt(fd.Typ, fdAddress).Elem().Set(reflect.ValueOf(val))
	return nil
}

type FieldMeta struct {
	Offset uintptr
	Typ    reflect.Type
}
