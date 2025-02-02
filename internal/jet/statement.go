package jet

import (
	"context"
	"database/sql"
	"github.com/go-jet/jet/qrm"
)

//Statement is common interface for all statements(SELECT, INSERT, UPDATE, DELETE, LOCK)
type Statement interface {
	// Sql returns parametrized sql query with list of arguments.
	Sql() (query string, args []interface{})
	// DebugSql returns debug query where every parametrized placeholder is replaced with its argument.
	// Do not use it in production. Use it only for debug purposes.
	DebugSql() (query string)

	// Query executes statement over database connection db and stores row result in destination.
	// Destination can be either pointer to struct or pointer to a slice.
	// If destination is pointer to struct and query result set is empty, method returns qrm.ErrNoRows.
	Query(db qrm.DB, destination interface{}) error
	// QueryContext executes statement with a context over database connection db and stores row result in destination.
	// Destination can be either pointer to struct or pointer to a slice.
	// If destination is pointer to struct and query result set is empty, method returns qrm.ErrNoRows.
	QueryContext(context context.Context, db qrm.DB, destination interface{}) error

	//Exec executes statement over db connection without returning any rows.
	Exec(db qrm.DB) (sql.Result, error)
	//Exec executes statement with context over db connection without returning any rows.
	ExecContext(context context.Context, db qrm.DB) (sql.Result, error)
}

// SerializerStatement interface
type SerializerStatement interface {
	Serializer
	Statement
}

// StatementWithProjections interface
type StatementWithProjections interface {
	Statement
	HasProjections
	Serializer
}

// HasProjections interface
type HasProjections interface {
	projections() ProjectionList
}

// serializerStatementInterfaceImpl struct
type serializerStatementInterfaceImpl struct {
	dialect       Dialect
	statementType StatementType
	parent        SerializerStatement
}

func (s *serializerStatementInterfaceImpl) Sql() (query string, args []interface{}) {

	queryData := &SQLBuilder{Dialect: s.dialect}

	s.parent.serialize(s.statementType, queryData, noWrap)

	query, args = queryData.finalize()
	return
}

func (s *serializerStatementInterfaceImpl) DebugSql() (query string) {
	sqlBuilder := &SQLBuilder{Dialect: s.dialect, debug: true}

	s.parent.serialize(s.statementType, sqlBuilder, noWrap)

	query, _ = sqlBuilder.finalize()
	return
}

func (s *serializerStatementInterfaceImpl) Query(db qrm.DB, destination interface{}) error {
	query, args := s.Sql()

	return qrm.Query(context.Background(), db, query, args, destination)
}

func (s *serializerStatementInterfaceImpl) QueryContext(context context.Context, db qrm.DB, destination interface{}) error {
	query, args := s.Sql()

	return qrm.Query(context, db, query, args, destination)
}

func (s *serializerStatementInterfaceImpl) Exec(db qrm.DB) (res sql.Result, err error) {
	query, args := s.Sql()
	return db.Exec(query, args...)
}

func (s *serializerStatementInterfaceImpl) ExecContext(context context.Context, db qrm.DB) (res sql.Result, err error) {
	query, args := s.Sql()

	return db.ExecContext(context, query, args...)
}

// ExpressionStatement interfacess
type ExpressionStatement interface {
	Expression
	Statement
	HasProjections
}

// NewExpressionStatementImpl creates new expression statement
func NewExpressionStatementImpl(Dialect Dialect, statementType StatementType, parent ExpressionStatement, clauses ...Clause) ExpressionStatement {
	return &expressionStatementImpl{
		expressionInterfaceImpl{Parent: parent},
		statementImpl{
			serializerStatementInterfaceImpl: serializerStatementInterfaceImpl{
				parent:        parent,
				dialect:       Dialect,
				statementType: statementType,
			},
			Clauses: clauses,
		},
	}
}

type expressionStatementImpl struct {
	expressionInterfaceImpl
	statementImpl
}

func (s *expressionStatementImpl) serializeForProjection(statement StatementType, out *SQLBuilder) {
	s.serialize(statement, out)
}

// NewStatementImpl creates new statementImpl
func NewStatementImpl(Dialect Dialect, statementType StatementType, parent SerializerStatement, clauses ...Clause) SerializerStatement {
	return &statementImpl{
		serializerStatementInterfaceImpl: serializerStatementInterfaceImpl{
			parent:        parent,
			dialect:       Dialect,
			statementType: statementType,
		},
		Clauses: clauses,
	}
}

type statementImpl struct {
	serializerStatementInterfaceImpl

	Clauses []Clause
}

func (s *statementImpl) projections() ProjectionList {
	for _, clause := range s.Clauses {
		if selectClause, ok := clause.(ClauseWithProjections); ok {
			return selectClause.projections()
		}
	}

	return nil
}

func (s *statementImpl) serialize(statement StatementType, out *SQLBuilder, options ...SerializeOption) {

	if !contains(options, noWrap) {
		out.WriteString("(")
		out.IncreaseIdent()
	}

	for _, clause := range s.Clauses {
		clause.Serialize(statement, out)
	}

	if !contains(options, noWrap) {
		out.DecreaseIdent()
		out.NewLine()
		out.WriteString(")")
	}
}
