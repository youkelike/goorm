package errs

import (
	"errors"
	"fmt"
)

var (
	ErrPointerOnly = errors.New("orm: 只支持结构体和指向结构体的一级指针")
	ErrNoRows      = errors.New("orm: 没有数据")
)

func NewUnknownField(name string) error {
	return fmt.Errorf("orm: 未知字段 %s", name)
}

func NewUnknownColumn(name string) error {
	return fmt.Errorf("orm: 未知列名 %s", name)
}

func NewUnsupportExpression(expr any) error {
	return fmt.Errorf("orm: 不支持的表达式类型 %v", expr)
}

func NewInvalidTagContent(tag string) error {
	return fmt.Errorf("orm: 不支持的标签 %s", tag)
}
