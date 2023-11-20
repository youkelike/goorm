package reflect

import (
	"errors"
	"reflect"
)

func IterateFields(entity any) (map[string]any, error) {
	if entity == nil {
		return nil, errors.New("不支持 nil")
	}

	typ := reflect.TypeOf(entity)
	val := reflect.ValueOf(entity)
	if val.IsZero() {
		return nil, errors.New("不支持零值")
	}

	// 指针类型处理
	for typ.Kind() == reflect.Pointer {
		typ = typ.Elem()
		val = val.Elem()
	}

	// 非结构体类型处理
	if typ.Kind() != reflect.Struct {
		return nil, errors.New("不支持的类型")
	}

	// 遍历结构体的字段
	nf := typ.NumField()
	res := make(map[string]any, nf)
	for i := 0; i < nf; i++ {
		fieldType := typ.Field(i)
		fieldValue := val.Field(i)
		// 处理非导出字段，给类型零值
		if fieldType.IsExported() {
			res[fieldType.Name] = fieldValue.Interface()
		} else {
			res[fieldType.Name] = reflect.Zero(fieldType.Type).Interface()
		}

	}
	return res, nil
}

func SetField(entity any, field string, newVal any) error {
	val := reflect.ValueOf(entity)
	for val.Kind() == reflect.Pointer {
		val = val.Elem()
	}

	fieldVal := val.FieldByName(field)
	if !fieldVal.CanSet() {
		return errors.New("字段不可修改")
	}
	fieldVal.Set(reflect.ValueOf(newVal))

	return nil
}
