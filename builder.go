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

func (b *builder) buildPredicate(p Predicate) error {
	left, ok := p.left.(Predicate)
	if ok {
		b.sb.WriteString("(")
		err := b.buildPredicate(left)
		if err != nil {
			return err
		}
		b.sb.WriteString(")")
	} else {
		err := b.buildExpresssion(p.left)
		if err != nil {
			return err
		}
	}

	if p.op == opNot || p.op == opAnd || p.op == opOr {
		b.sb.WriteString(" ")
	}
	b.sb.WriteString(p.op.String())
	if p.op == opNot || p.op == opAnd || p.op == opOr {
		b.sb.WriteString(" ")
	}

	right, ok := p.right.(Predicate)
	if ok {
		b.sb.WriteString("(")
		err := b.buildPredicate(right)
		if err != nil {
			return err
		}
		b.sb.WriteString(")")
	} else {
		err := b.buildExpresssion(p.right)
		if err != nil {
			return err
		}
	}
	return nil
}

func (b *builder) buildExpresssion(expr Expression) error {
	switch p := expr.(type) {
	case nil:
	// case Predicate:
	// 	return s.buildPredicate(p)
	case Column:
		p.alias = ""
		err := b.buildColumn(p)
		if err != nil {
			return err
		}
	case value:
		b.sb.WriteString("?")
		b.addArgs(p.val)
	case RawExpr:
		b.sb.WriteString("(")
		b.sb.WriteString(p.raw)
		b.sb.WriteString(")")
		b.addArgs(p.args...)
	default:
		return errs.NewUnsupportExpression(expr)
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
