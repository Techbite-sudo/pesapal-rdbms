package parser

import (
	"fmt"
	"strconv"
	"strings"
)

// Parser parses SQL statements
type Parser struct {
	lexer     *Lexer
	curToken  Token
	peekToken Token
	errors    []string
}

// NewParser creates a new Parser instance
func NewParser(input string) *Parser {
	p := &Parser{
		lexer:  NewLexer(input),
		errors: []string{},
	}
	// Read two tokens to initialize curToken and peekToken
	p.nextToken()
	p.nextToken()
	return p
}

// Errors returns parsing errors
func (p *Parser) Errors() []string {
	return p.errors
}

// nextToken advances to the next token
func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.lexer.NextToken()
}

// curTokenIs checks if current token is of given type
func (p *Parser) curTokenIs(t TokenType) bool {
	return p.curToken.Type == t
}

// peekTokenIs checks if peek token is of given type
func (p *Parser) peekTokenIs(t TokenType) bool {
	return p.peekToken.Type == t
}

// expectPeek checks peek token and advances if match
func (p *Parser) expectPeek(t TokenType) bool {
	if p.peekTokenIs(t) {
		p.nextToken()
		return true
	}
	p.peekError(t)
	return false
}

// peekError adds an error for unexpected peek token
func (p *Parser) peekError(t TokenType) {
	msg := fmt.Sprintf("line %d:%d: expected next token to be %s, got %s instead",
		p.peekToken.Line, p.peekToken.Column, t, p.peekToken.Type)
	p.errors = append(p.errors, msg)
}

// addError adds a custom error message
func (p *Parser) addError(msg string) {
	p.errors = append(p.errors, fmt.Sprintf("line %d:%d: %s",
		p.curToken.Line, p.curToken.Column, msg))
}

// Parse parses the SQL statement
func (p *Parser) Parse() (Statement, error) {
	var stmt Statement

	switch p.curToken.Type {
	case CREATE:
		stmt = p.parseCreateTable()
	case DROP:
		stmt = p.parseDropTable()
	case INSERT:
		stmt = p.parseInsert()
	case SELECT:
		stmt = p.parseSelect()
	case UPDATE:
		stmt = p.parseUpdate()
	case DELETE:
		stmt = p.parseDelete()
	case EOF:
		return nil, fmt.Errorf("empty statement")
	default:
		return nil, fmt.Errorf("unexpected token: %s", p.curToken.Type)
	}

	if len(p.errors) > 0 {
		return nil, fmt.Errorf("parsing errors: %s", strings.Join(p.errors, "; "))
	}

	return stmt, nil
}

// parseCreateTable parses CREATE TABLE statement
func (p *Parser) parseCreateTable() *CreateTableStmt {
	stmt := &CreateTableStmt{}

	if !p.expectPeek(TABLE) {
		return nil
	}

	if !p.expectPeek(IDENT) {
		return nil
	}
	stmt.TableName = p.curToken.Literal

	if !p.expectPeek(LPAREN) {
		return nil
	}

	stmt.Columns = p.parseColumnDefinitions()

	// parseColumnDefinitions leaves curToken at ) or at the last token before )
	// We need to ensure we're at the closing paren
	if !p.curTokenIs(RPAREN) {
		p.addError("expected ) after column definitions")
		return nil
	}

	return stmt
}

// parseColumnDefinitions parses column definitions
func (p *Parser) parseColumnDefinitions() []*ColumnDef {
	columns := []*ColumnDef{}

	p.nextToken()

	for !p.curTokenIs(RPAREN) && !p.curTokenIs(EOF) {
		col := &ColumnDef{}

		if !p.curTokenIs(IDENT) {
			p.addError("expected column name")
			return nil
		}
		col.Name = p.curToken.Literal

		p.nextToken()

		// Parse data type
		switch p.curToken.Type {
		case INTEGER:
			col.DataType = "INTEGER"
		case VARCHAR:
			col.DataType = "VARCHAR"
			if p.peekTokenIs(LPAREN) {
				p.nextToken() // consume (
				p.nextToken() // move to size
				if p.curTokenIs(INT) {
					size, _ := strconv.Atoi(p.curToken.Literal)
					col.Size = size
				}
				if !p.expectPeek(RPAREN) {
					return nil
				}
			}
		case BOOLEAN:
			col.DataType = "BOOLEAN"
		case FLOAT_TYPE:
			col.DataType = "FLOAT"
		default:
			p.addError(fmt.Sprintf("unknown data type: %s", p.curToken.Literal))
			return nil
		}

		// Parse constraints
		p.nextToken()
		for p.curTokenIs(PRIMARY) || p.curTokenIs(UNIQUE) || p.curTokenIs(NOT) {
			if p.curTokenIs(PRIMARY) {
				if !p.expectPeek(KEY) {
					return nil
				}
				col.PrimaryKey = true
			} else if p.curTokenIs(UNIQUE) {
				col.Unique = true
			} else if p.curTokenIs(NOT) {
				if !p.expectPeek(NULL) {
					return nil
				}
				col.NotNull = true
			}
			p.nextToken()
		}

		columns = append(columns, col)

		if p.curTokenIs(COMMA) {
			p.nextToken()
		} else {
			break
		}
	}

	return columns
}

// parseDropTable parses DROP TABLE statement
func (p *Parser) parseDropTable() *DropTableStmt {
	stmt := &DropTableStmt{}

	if !p.expectPeek(TABLE) {
		return nil
	}

	if !p.expectPeek(IDENT) {
		return nil
	}
	stmt.TableName = p.curToken.Literal

	return stmt
}

// parseInsert parses INSERT INTO statement
func (p *Parser) parseInsert() *InsertStmt {
	stmt := &InsertStmt{}

	if !p.expectPeek(INTO) {
		return nil
	}

	if !p.expectPeek(IDENT) {
		return nil
	}
	stmt.TableName = p.curToken.Literal

	// Parse column names (optional)
	if p.peekTokenIs(LPAREN) {
		p.nextToken()
		stmt.Columns = p.parseIdentifierList()
		if !p.expectPeek(RPAREN) {
			return nil
		}
	}

	if !p.expectPeek(VALUES) {
		return nil
	}

	// Parse values
	stmt.Values = [][]Expression{}
	for p.peekTokenIs(LPAREN) {
		p.nextToken()
		p.nextToken()
		values := p.parseExpressionList()
		stmt.Values = append(stmt.Values, values)

		if !p.expectPeek(RPAREN) {
			return nil
		}

		if p.peekTokenIs(COMMA) {
			p.nextToken()
		} else {
			break
		}
	}

	return stmt
}

// parseSelect parses SELECT statement
func (p *Parser) parseSelect() *SelectStmt {
	stmt := &SelectStmt{}

	p.nextToken()

	// Parse column list
	if p.curTokenIs(ASTERISK) {
		stmt.Columns = []string{"*"}
		p.nextToken()
	} else {
		stmt.Columns = p.parseIdentifierList()
		// parseIdentifierList leaves us at the last identifier, advance to next token
		p.nextToken()
	}

	if !p.curTokenIs(FROM) {
		p.addError("expected FROM after column list")
		return nil
	}

	if !p.expectPeek(IDENT) {
		return nil
	}
	stmt.TableName = p.curToken.Literal

	// Parse JOINs
	for p.peekTokenIs(INNER) || p.peekTokenIs(JOIN) {
		p.nextToken()
		join := &JoinClause{JoinType: "INNER"}

		if p.curTokenIs(INNER) {
			if !p.expectPeek(JOIN) {
				return nil
			}
		}

		if !p.expectPeek(IDENT) {
			return nil
		}
		join.TableName = p.curToken.Literal

		if !p.expectPeek(ON) {
			return nil
		}

		p.nextToken()
		join.On = p.parseExpression()

		stmt.Joins = append(stmt.Joins, join)
	}

	// Parse WHERE clause
	if p.peekTokenIs(WHERE) {
		p.nextToken()
		p.nextToken()
		stmt.Where = p.parseExpression()
	}

	return stmt
}

// parseUpdate parses UPDATE statement
func (p *Parser) parseUpdate() *UpdateStmt {
	stmt := &UpdateStmt{Set: make(map[string]Expression)}

	if !p.expectPeek(IDENT) {
		return nil
	}
	stmt.TableName = p.curToken.Literal

	if !p.expectPeek(SET) {
		return nil
	}

	// Parse SET clause
	p.nextToken()
	for !p.curTokenIs(WHERE) && !p.curTokenIs(EOF) && !p.curTokenIs(SEMICOLON) {
		if !p.curTokenIs(IDENT) {
			p.addError("expected column name in SET clause")
			return nil
		}
		colName := p.curToken.Literal

		if !p.expectPeek(EQ) {
			return nil
		}

		p.nextToken()
		stmt.Set[colName] = p.parseExpression()

		if p.peekTokenIs(COMMA) {
			p.nextToken()
			p.nextToken()
		} else {
			break
		}
	}

	// Parse WHERE clause
	if p.peekTokenIs(WHERE) {
		p.nextToken()
		p.nextToken()
		stmt.Where = p.parseExpression()
	}

	return stmt
}

// parseDelete parses DELETE statement
func (p *Parser) parseDelete() *DeleteStmt {
	stmt := &DeleteStmt{}

	if !p.expectPeek(FROM) {
		return nil
	}

	if !p.expectPeek(IDENT) {
		return nil
	}
	stmt.TableName = p.curToken.Literal

	// Parse WHERE clause
	if p.peekTokenIs(WHERE) {
		p.nextToken()
		p.nextToken()
		stmt.Where = p.parseExpression()
	}

	return stmt
}

// parseIdentifierList parses a comma-separated list of identifiers
func (p *Parser) parseIdentifierList() []string {
	list := []string{}

	if p.curTokenIs(IDENT) {
		list = append(list, p.curToken.Literal)
	}

	for p.peekTokenIs(COMMA) {
		p.nextToken()
		p.nextToken()
		if p.curTokenIs(IDENT) {
			list = append(list, p.curToken.Literal)
		}
	}

	// Don't advance - let the caller decide what to do next
	return list
}

// parseExpressionList parses a comma-separated list of expressions
func (p *Parser) parseExpressionList() []Expression {
	list := []Expression{}

	list = append(list, p.parseExpression())

	for p.peekTokenIs(COMMA) {
		p.nextToken()
		p.nextToken()
		list = append(list, p.parseExpression())
	}

	return list
}

// parseExpression parses an expression
func (p *Parser) parseExpression() Expression {
	left := p.parsePrimary()

	// Check for binary operators
	if p.peekTokenIs(EQ) || p.peekTokenIs(NEQ) || p.peekTokenIs(LT) ||
		p.peekTokenIs(GT) || p.peekTokenIs(LTE) || p.peekTokenIs(GTE) ||
		p.peekTokenIs(AND) || p.peekTokenIs(OR) {
		p.nextToken()
		operator := p.curToken.Literal
		p.nextToken()
		right := p.parseExpression()
		return &BinaryExpr{
			Left:     left,
			Operator: operator,
			Right:    right,
		}
	}

	return left
}

// parsePrimary parses a primary expression (literal or identifier)
func (p *Parser) parsePrimary() Expression {
	switch p.curToken.Type {
	case IDENT:
		return &Identifier{Value: p.curToken.Literal}
	case INT:
		val, _ := strconv.Atoi(p.curToken.Literal)
		return &Literal{Value: val}
	case FLOAT:
		val, _ := strconv.ParseFloat(p.curToken.Literal, 64)
		return &Literal{Value: val}
	case STRING:
		return &Literal{Value: p.curToken.Literal}
	case NULL:
		return &NullLiteral{}
	default:
		p.addError(fmt.Sprintf("unexpected token in expression: %s", p.curToken.Type))
		return nil
	}
}
