package parser

// TokenType represents the type of token
type TokenType int

const (
	// Special tokens
	ILLEGAL TokenType = iota
	EOF
	WHITESPACE

	// Literals
	IDENT  // table names, column names
	INT    // 123
	STRING // "hello" or 'hello'
	FLOAT  // 123.45

	// Keywords
	SELECT
	FROM
	WHERE
	INSERT
	INTO
	VALUES
	UPDATE
	SET
	DELETE
	CREATE
	TABLE
	DROP
	PRIMARY
	KEY
	UNIQUE
	JOIN
	INNER
	ON
	AND
	OR
	NOT
	NULL

	// Data types
	INTEGER
	VARCHAR
	BOOLEAN
	FLOAT_TYPE

	// Operators
	ASTERISK  // *
	COMMA     // ,
	SEMICOLON // ;
	LPAREN    // (
	RPAREN    // )
	EQ        // =
	NEQ       // !=
	LT        // <
	GT        // >
	LTE       // <=
	GTE       // >=
)

// Token represents a lexical token
type Token struct {
	Type    TokenType
	Literal string
	Line    int
	Column  int
}

// Keywords maps string literals to their token types
var keywords = map[string]TokenType{
	"SELECT":  SELECT,
	"FROM":    FROM,
	"WHERE":   WHERE,
	"INSERT":  INSERT,
	"INTO":    INTO,
	"VALUES":  VALUES,
	"UPDATE":  UPDATE,
	"SET":     SET,
	"DELETE":  DELETE,
	"CREATE":  CREATE,
	"TABLE":   TABLE,
	"DROP":    DROP,
	"PRIMARY": PRIMARY,
	"KEY":     KEY,
	"UNIQUE":  UNIQUE,
	"JOIN":    JOIN,
	"INNER":   INNER,
	"ON":      ON,
	"AND":     AND,
	"OR":      OR,
	"NOT":     NOT,
	"NULL":    NULL,
	"INTEGER": INTEGER,
	"VARCHAR": VARCHAR,
	"BOOLEAN": BOOLEAN,
	"FLOAT":   FLOAT_TYPE,
}

// LookupIdent checks if an identifier is a keyword
func LookupIdent(ident string) TokenType {
	if tok, ok := keywords[ident]; ok {
		return tok
	}
	return IDENT
}

// String returns a string representation of the token type
func (t TokenType) String() string {
	switch t {
	case ILLEGAL:
		return "ILLEGAL"
	case EOF:
		return "EOF"
	case IDENT:
		return "IDENT"
	case INT:
		return "INT"
	case STRING:
		return "STRING"
	case FLOAT:
		return "FLOAT"
	case SELECT:
		return "SELECT"
	case FROM:
		return "FROM"
	case WHERE:
		return "WHERE"
	case INSERT:
		return "INSERT"
	case INTO:
		return "INTO"
	case VALUES:
		return "VALUES"
	case UPDATE:
		return "UPDATE"
	case SET:
		return "SET"
	case DELETE:
		return "DELETE"
	case CREATE:
		return "CREATE"
	case TABLE:
		return "TABLE"
	case DROP:
		return "DROP"
	case PRIMARY:
		return "PRIMARY"
	case KEY:
		return "KEY"
	case UNIQUE:
		return "UNIQUE"
	case JOIN:
		return "JOIN"
	case INNER:
		return "INNER"
	case ON:
		return "ON"
	case AND:
		return "AND"
	case OR:
		return "OR"
	case NOT:
		return "NOT"
	case NULL:
		return "NULL"
	case INTEGER:
		return "INTEGER"
	case VARCHAR:
		return "VARCHAR"
	case BOOLEAN:
		return "BOOLEAN"
	case FLOAT_TYPE:
		return "FLOAT_TYPE"
	case ASTERISK:
		return "*"
	case COMMA:
		return ","
	case SEMICOLON:
		return ";"
	case LPAREN:
		return "("
	case RPAREN:
		return ")"
	case EQ:
		return "="
	case NEQ:
		return "!="
	case LT:
		return "<"
	case GT:
		return ">"
	case LTE:
		return "<="
	case GTE:
		return ">="
	default:
		return "UNKNOWN"
	}
}
