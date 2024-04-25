package orm

import (
	"bytes"

	"gitee.com/youkelike/orm/internal/errs"
	"gitee.com/youkelike/orm/model"
)

type builder struct {
	model *model.Model
	r     model.Registry
	// dialect Dialect
	// core

	sb   bytes.Buffer
	args []any

	quoter byte
}

func (b *builder) quote(name string) {
	b.sb.WriteByte(b.quoter)
	b.sb.WriteString(name)
	b.sb.WriteByte(b.quoter)
}

func (b *builder) buildColumn(c Column) error {
	switch table := c.table.(type) {
	case nil:
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
	case Table:
		m, err := b.r.Get(table.entity)
		if err != nil {
			return err
		}
		fd, ok := m.FieldMap[c.name]
		if !ok {
			return errs.NewUnknownField(c.name)
		}
		if table.alias != "" {
			b.sb.WriteString(table.alias)
			b.sb.WriteString(".")
		} else {
			b.sb.WriteString(m.TableName)
			b.sb.WriteString(".")
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
	default:
		return errs.NewUnsupportTable(table)
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
