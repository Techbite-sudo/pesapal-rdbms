package parser

import (
	"strings"
	"unicode"
)

// Lexer tokenizes SQL input
type Lexer struct {
	input        string
	position     int  // current position in input (points to current char)
	readPosition int  // current reading position in input (after current char)
	ch           byte // current char under examination
	line         int
	column       int
}

// NewLexer creates a new Lexer instance
func NewLexer(input string) *Lexer {
	l := &Lexer{
		input:  input,
		line:   1,
		column: 0,
	}
	l.readChar()
	return l
}

// readChar reads the next character and advances position
func (l *Lexer) readChar() {
	if l.readPosition >= len(l.input) {
		l.ch = 0 // ASCII code for "NUL"
	} else {
		l.ch = l.input[l.readPosition]
	}
	l.position = l.readPosition
	l.readPosition++
	l.column++
}

// peekChar looks at the next character without advancing
func (l *Lexer) peekChar() byte {
	if l.readPosition >= len(l.input) {
		return 0
	}
	return l.input[l.readPosition]
}

// NextToken returns the next token from the input
func (l *Lexer) NextToken() Token {
	var tok Token

	l.skipWhitespace()

	tok.Line = l.line
	tok.Column = l.column

	switch l.ch {
	case '*':
		tok = Token{Type: ASTERISK, Literal: string(l.ch), Line: l.line, Column: l.column}
	case ',':
		tok = Token{Type: COMMA, Literal: string(l.ch), Line: l.line, Column: l.column}
	case ';':
		tok = Token{Type: SEMICOLON, Literal: string(l.ch), Line: l.line, Column: l.column}
	case '(':
		tok = Token{Type: LPAREN, Literal: string(l.ch), Line: l.line, Column: l.column}
	case ')':
		tok = Token{Type: RPAREN, Literal: string(l.ch), Line: l.line, Column: l.column}
	case '=':
		tok = Token{Type: EQ, Literal: string(l.ch), Line: l.line, Column: l.column}
	case '!':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = Token{Type: NEQ, Literal: string(ch) + string(l.ch), Line: l.line, Column: l.column}
		} else {
			tok = Token{Type: ILLEGAL, Literal: string(l.ch), Line: l.line, Column: l.column}
		}
	case '<':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = Token{Type: LTE, Literal: string(ch) + string(l.ch), Line: l.line, Column: l.column}
		} else {
			tok = Token{Type: LT, Literal: string(l.ch), Line: l.line, Column: l.column}
		}
	case '>':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = Token{Type: GTE, Literal: string(ch) + string(l.ch), Line: l.line, Column: l.column}
		} else {
			tok = Token{Type: GT, Literal: string(l.ch), Line: l.line, Column: l.column}
		}
	case '"', '\'':
		tok.Type = STRING
		tok.Literal = l.readString(l.ch)
	case 0:
		tok.Literal = ""
		tok.Type = EOF
	default:
		if isLetter(l.ch) {
			tok.Literal = l.readIdentifier()
			tok.Type = LookupIdent(strings.ToUpper(tok.Literal))
			return tok
		} else if isDigit(l.ch) {
			literal, isFloat := l.readNumber()
			if isFloat {
				tok.Type = FLOAT
			} else {
				tok.Type = INT
			}
			tok.Literal = literal
			return tok
		} else {
			tok = Token{Type: ILLEGAL, Literal: string(l.ch), Line: l.line, Column: l.column}
		}
	}

	l.readChar()
	return tok
}

// readIdentifier reads an identifier (table name, column name, keyword)
func (l *Lexer) readIdentifier() string {
	position := l.position
	for isLetter(l.ch) || isDigit(l.ch) || l.ch == '_' {
		l.readChar()
	}
	return l.input[position:l.position]
}

// readNumber reads a number (integer or float)
func (l *Lexer) readNumber() (string, bool) {
	position := l.position
	isFloat := false

	for isDigit(l.ch) {
		l.readChar()
	}

	// Check for decimal point
	if l.ch == '.' && isDigit(l.peekChar()) {
		isFloat = true
		l.readChar() // consume '.'
		for isDigit(l.ch) {
			l.readChar()
		}
	}

	return l.input[position:l.position], isFloat
}

// readString reads a string literal enclosed in quotes
func (l *Lexer) readString(quote byte) string {
	position := l.position + 1
	for {
		l.readChar()
		if l.ch == quote || l.ch == 0 {
			break
		}
		if l.ch == '\n' {
			l.line++
			l.column = 0
		}
	}
	return l.input[position:l.position]
}

// skipWhitespace skips whitespace characters
func (l *Lexer) skipWhitespace() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\n' || l.ch == '\r' {
		if l.ch == '\n' {
			l.line++
			l.column = 0
		}
		l.readChar()
	}
}

// isLetter checks if a character is a letter
func isLetter(ch byte) bool {
	return unicode.IsLetter(rune(ch))
}

// isDigit checks if a character is a digit
func isDigit(ch byte) bool {
	return unicode.IsDigit(rune(ch))
}
