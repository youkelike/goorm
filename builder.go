package orm

import (
	"strings"

	"gitee.com/youkelike/orm/internal/errs"
	"gitee.com/youkelike/orm/model"
)

type builder struct {
	sb    strings.Builder
	args  []any
	model *model.Model

	dialect Dialect
	quoter  byte
}

func (b *builder) quote(name string) {
	b.sb.WriteByte(b.quoter)
	b.sb.WriteString(name)
	b.sb.WriteByte(b.quoter)
}

func (b *builder) buildColumn(c Column) error {
	// fd, ok := b.model.FieldMap[name]
	// if !ok {
	// 	return errs.NewUnknownField(name)
	// }
	// b.quote(fd.ColName)
	// return nil

	fd, ok := b.model.FieldMap[c.name]
	if !ok {
		return errs.NewUnknownField(c.name)
	}
	b.sb.WriteString(fd.ColName)
	if c.alias != "" {
		b.sb.WriteString(" AS ")
		b.sb.WriteString(c.alias)
	}
	if c.order != "" {
		b.sb.WriteString(" ")
		b.sb.WriteString(c.order)
	}
	return nil
}

func (b *builder) addArgs(vals ...any) error {
	if len(vals) == 0 {
		return nil
	}
	b.args = append(b.args, vals...)
	return nil
}
