package orm

import (
	"gitee.com/youkelike/orm/internal/errs"
)

var (
	DialectMySQL  Dialect = mysqlDialect{}
	DialectSQLite Dialect = sqliteDialect{}
)

type Dialect interface {
	quoter() byte
	buildUpsert(b *builder, odk *Upsert) error
}

type standardSQL struct {
}

func (s standardSQL) quoter() byte {
	return '"'
}

func (s standardSQL) buildUpsert(b *builder, odk *Upsert) error {
	b.sb.WriteString(" ON CONFLICT(")
	for i, col := range odk.conflictColumns {
		if i > 0 {
			b.sb.WriteString(",")
		}
		err := b.buildColumn(C(col))
		if err != nil {
			return err
		}
	}
	b.sb.WriteString(") DO UPDATE SET ")
	for idx, assign := range odk.assigns {
		if idx > 0 {
			b.sb.WriteString(",")
		}
		switch a := assign.(type) {
		case Assignment:
			fd, ok := b.model.FieldMap[a.col]
			if !ok {
				return errs.NewUnknownField(a.col)
			}
			b.sb.WriteString(fd.ColName)
			b.sb.WriteString("=?")
			b.addArgs(a.val)
		case Column:
			fd, ok := b.model.FieldMap[a.name]
			if !ok {
				return errs.NewUnknownField(a.name)
			}
			b.sb.WriteString(fd.ColName)
			b.sb.WriteString("=excluded.")
			b.sb.WriteString(fd.ColName)
		default:
			return errs.NewUnsupportedAssignable(assign)
		}
	}
	return nil
}

type mysqlDialect struct {
	standardSQL
}

func (s mysqlDialect) quoter() byte {
	return '`'
}

func (s mysqlDialect) buildUpsert(b *builder, odk *Upsert) error {
	b.sb.WriteString(" ON DUPLICATE KEY UPDATE ")
	for idx, assign := range odk.assigns {
		if idx > 0 {
			b.sb.WriteString(",")
		}
		switch a := assign.(type) {
		case Assignment:
			fd, ok := b.model.FieldMap[a.col]
			if !ok {
				return errs.NewUnknownField(a.col)
			}
			b.sb.WriteString(fd.ColName)
			b.sb.WriteString("=?")
			b.addArgs(a.val)
		case Column:
			fd, ok := b.model.FieldMap[a.name]
			if !ok {
				return errs.NewUnknownField(a.name)
			}
			b.sb.WriteString(fd.ColName)
			b.sb.WriteString("=VALUES(")
			b.sb.WriteString(fd.ColName)
			b.sb.WriteString(")")
		default:
			return errs.NewUnsupportedAssignable(assign)
		}
	}
	return nil
}

type sqliteDialect struct {
	standardSQL
}

func (s sqliteDialect) quoter() byte {
	return '`'
}

type postgreDialect struct {
	standardSQL
}
