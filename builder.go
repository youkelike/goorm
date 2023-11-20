package orm

import (
	"strings"

	"gitee.com/youkelike/orm/internal/errs"
)

type builder struct {
	sb    strings.Builder
	model *model
	args  []any
}

func (d *builder) buildWhere(ps []Predicate) error {
	d.sb.WriteString(" WHERE ")
	p := ps[0]
	for i := 1; i < len(ps); i++ {
		p = p.And(ps[i])
	}

	if err := d.buildExpression(p); err != nil {
		return err
	}
	return nil
}

func (d *builder) buildExpression(expr Expression) error {
	switch exp := expr.(type) {
	case Predicate:
		_, ok := exp.left.(Predicate)
		if ok {
			d.sb.WriteString("(")
		}
		if err := d.buildExpression(exp.left); err != nil {
			return err
		}
		if ok {
			d.sb.WriteString(")")
		}

		if exp.op == opAnd || exp.op == opOr || exp.op == opNot {
			d.sb.WriteString(" ")
		}
		d.sb.WriteString(exp.op.String())
		if exp.op == opAnd || exp.op == opOr || exp.op == opNot {
			d.sb.WriteString(" ")
		}

		_, ok = exp.right.(Predicate)
		if ok {
			d.sb.WriteString("(")
		}
		if err := d.buildExpression(exp.right); err != nil {
			return err
		}
		if ok {
			d.sb.WriteString(")")
		}
	case Column:
		fd, ok := d.model.fields[exp.name]
		if !ok {
			return errs.NewUnknownField(exp.name)
		}
		d.sb.WriteString(fd.colName)
	case value:
		d.sb.WriteString("?")
		d.addArg(exp.val)
	case nil:
	default:
		return errs.NewUnsupportExpression(exp)
	}
	return nil
}

func (d *builder) addArg(val any) {
	if d.args == nil {
		d.args = make([]any, 0, 8)
	}
	d.args = append(d.args, val)
}
