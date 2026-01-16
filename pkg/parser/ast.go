package parser

// Statement represents any SQL statement
type Statement interface {
	statementNode()
}

// Expression represents any SQL expression
type Expression interface {
	expressionNode()
}

// ColumnDef represents a column definition in CREATE TABLE
type ColumnDef struct {
	Name       string
	DataType   string
	Size       int  // for VARCHAR(size)
	PrimaryKey bool
	Unique     bool
	NotNull    bool
}

func (c *ColumnDef) statementNode() {}

// CreateTableStmt represents CREATE TABLE statement
type CreateTableStmt struct {
	TableName string
	Columns   []*ColumnDef
}

func (c *CreateTableStmt) statementNode() {}

// DropTableStmt represents DROP TABLE statement
type DropTableStmt struct {
	TableName string
}

func (d *DropTableStmt) statementNode() {}

// InsertStmt represents INSERT INTO statement
type InsertStmt struct {
	TableName string
	Columns   []string
	Values    [][]Expression
}

func (i *InsertStmt) statementNode() {}

// SelectStmt represents SELECT statement
type SelectStmt struct {
	Columns   []string // column names or "*"
	TableName string
	Joins     []*JoinClause
	Where     Expression
}

func (s *SelectStmt) statementNode() {}

// UpdateStmt represents UPDATE statement
type UpdateStmt struct {
	TableName string
	Set       map[string]Expression
	Where     Expression
}

func (u *UpdateStmt) statementNode() {}

// DeleteStmt represents DELETE statement
type DeleteStmt struct {
	TableName string
	Where     Expression
}

func (d *DeleteStmt) statementNode() {}

// JoinClause represents a JOIN clause
type JoinClause struct {
	JoinType  string // "INNER", "LEFT", "RIGHT"
	TableName string
	On        Expression
}

// BinaryExpr represents a binary expression (e.g., a = b, a > 5)
type BinaryExpr struct {
	Left     Expression
	Operator string
	Right    Expression
}

func (b *BinaryExpr) expressionNode() {}

// Identifier represents a column or table name
type Identifier struct {
	Value string
}

func (i *Identifier) expressionNode() {}

// Literal represents a literal value (string, number, etc.)
type Literal struct {
	Value interface{}
}

func (l *Literal) expressionNode() {}

// NullLiteral represents a NULL value
type NullLiteral struct{}

func (n *NullLiteral) expressionNode() {}
