package sqlbuilder

import (
	"database/sql"
	"github.com/dropbox/godropbox/errors"
	"github.com/sub0zero/go-sqlbuilder/types"
)

type updateStatement interface {
	statement

	SET(values ...interface{}) updateStatement
	WHERE(expression boolExpression) updateStatement
	RETURNING(projections ...projection) updateStatement
}

func newUpdateStatement(table writableTable, columns []column) updateStatement {
	return &updateStatementImpl{
		table:   table,
		columns: columns,
	}
}

type updateStatementImpl struct {
	table        writableTable
	columns      []column
	updateValues []clause
	where        boolExpression
	returning    []projection
}

func (u *updateStatementImpl) SET(values ...interface{}) updateStatement {

	for _, value := range values {
		if clause, ok := value.(clause); ok {
			u.updateValues = append(u.updateValues, clause)
		} else {
			u.updateValues = append(u.updateValues, Literal(value))
		}
	}

	return u
}

func (u *updateStatementImpl) WHERE(expression boolExpression) updateStatement {
	u.where = expression
	return u
}

func (u *updateStatementImpl) RETURNING(projections ...projection) updateStatement {
	u.returning = defaultProjectionAliasing(projections)
	return u
}

func (u *updateStatementImpl) Sql() (sql string, args []interface{}, err error) {
	out := &queryData{}

	out.writeString("UPDATE ")

	if u.table == nil {
		return "", nil, errors.New("nil tableName.")
	}

	if err = u.table.serialize(update_statement, out); err != nil {
		return
	}

	if len(u.updateValues) == 0 {
		return "", nil, errors.New("No column updated.")
	}

	out.writeString(" SET")

	if len(u.columns) > 1 {
		out.writeString(" ( ")
	} else {
		out.writeString(" ")
	}

	err = serializeColumnList(update_statement, u.columns, out)

	if err != nil {
		return "", nil, err
	}

	if len(u.columns) > 1 {
		out.writeString(" )")
	}

	out.writeString(" =")

	if len(u.updateValues) > 1 {
		out.writeString(" (")
	}

	for i, value := range u.updateValues {
		if i > 0 {
			out.writeString(", ")
		}

		err = value.serialize(update_statement, out)

		if err != nil {
			return
		}
	}

	if len(u.updateValues) > 1 {
		out.writeString(" )")
	}

	if u.where == nil {
		return "", nil, errors.New("Updating without a WHERE clause.")
	}

	if err = out.writeWhere(update_statement, u.where); err != nil {
		return
	}

	if len(u.returning) > 0 {
		out.writeString(" RETURNING ")

		err = serializeProjectionList(update_statement, u.returning, out)

		if err != nil {
			return
		}
	}

	return out.buff.String(), out.args, nil
}

func (u *updateStatementImpl) Query(db types.Db, destination interface{}) error {
	return Query(u, db, destination)
}

func (u *updateStatementImpl) Execute(db types.Db) (res sql.Result, err error) {
	return Execute(u, db)
}