package valuer

import (
	"database/sql"

	"gitee.com/youkelike/orm/model"
)

// 这里的设计有两种方式：
// 1、用包方法，reflect 和 unsafe 各一个包方法
// 2、用接口，提供 reflect 和 unsafe 各自的实现
type Value interface {
	// 这里的入参设计不需要传入待解析的结构体，
	// 因为结构体需要被反射解析才能用，反射解析结构体这一步在框架的 model 对象中已经做了，只要在创建实现了接口的结构体时传入 model 对象就好
	SetColumns(rows *sql.Rows) error
	// 根据结构体字段名获取字段值
	Field(name string) (any, error)
}

// 这里为什么不定义成接口而是定义了一个方法类型？
// 可以看作 build 模式的变体，这里相当于 build 模式中的 Build 方法
type Creator func(model *model.Model, entity any) Value
