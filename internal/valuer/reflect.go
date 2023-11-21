package valuer

import (
	"database/sql"
	"reflect"

	"gitee.com/youkelike/orm/internal/errs"
	"gitee.com/youkelike/orm/model"
)

type reflectValue struct {
	model *model.Model
	// val 是 new(T)
	// val any
	val reflect.Value
}

var _ Creator = NewReflectValue

func NewReflectValue(model *model.Model, val any) Value {
	return reflectValue{
		model: model,
		val:   reflect.ValueOf(val).Elem(),
	}
}

func (r reflectValue) Field(name string) (any, error) {
	return r.val.FieldByName(name).Interface(), nil
}

func (r reflectValue) SetColumns(rows *sql.Rows) error {
	cs, err := rows.Columns()
	if err != nil {
		return err
	}

	vals := make([]any, 0, len(cs))
	for _, c := range cs {
		fd, ok := r.model.ColumnMap[c]
		if !ok {
			return errs.NewUnknownColumn(c)
		}
		// 根据字段类型创建一个指针类型的值
		val := reflect.New(fd.Typ)
		// 不能这样写，因为后面要对它赋值
		// val := reflect.Zero(fd.typ)

		// vals 里接收的是 any 类型，需要转换一下
		vals = append(vals, val.Interface())
	}

	err = rows.Scan(vals...)
	if err != nil {
		return err
	}

	// 这里操作的是具体结构体的指针
	tpValueElem := r.val
	for i, c := range cs {
		fd := r.model.ColumnMap[c]

		// 结构体指针必须转成结构体，才能给其字段赋值
		tpValueElem.FieldByName(fd.GoName).Set(reflect.ValueOf(vals[i]).Elem())
	}

	return err
}
