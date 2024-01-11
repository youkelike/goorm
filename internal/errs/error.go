package errs

import (
	"errors"
	"fmt"
)

var (
	ErrPointerOnly      = errors.New("orm: 只支持结构体和指向结构体的一级指针")
	ErrNoRows           = errors.New("orm: 没有数据")
	ErrInsertZeroRow    = errors.New("orm: 插入 0 行")
	ErrNoGroupUseHaving = errors.New("orm: having 必须配合 group 使用")
	ErrNoOrderByVerb    = errors.New("orm: order by 必须指定字段排序规则")
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

func NewUnsupportedAssignable(expr any) error {
	return fmt.Errorf("orm: 不支持的赋值表达式类型 %v", expr)
}

func NewErrFailedToRollback(bizErr, rbErr error, panicked bool) error {
	return fmt.Errorf("orm: 事务闭包回滚失败，业务错误：%w，回滚错误：%s，是否panic：%t", bizErr, rbErr, panicked)
}

func NewUnsupportTable(table any) error {
	return fmt.Errorf("orm: 不支持的TableReference类型 %v", table)
}

func NewUnknownUpdateValue() error {
	return fmt.Errorf("orm: 缺少更新数据")
}
