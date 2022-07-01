package token

import (
	"fmt"
)

// token
type TokenType int

const (
	TOKEN_ILLEGAL TokenType = (iota - 1) // Illegal token
	TOKEN_EOF                            //End Of File

	TOKEN_PLUS     // +
	TOKEN_MINUS    // -
	TOKEN_MULTIPLY // *
	TOKEN_DIVIDE   // '/'
	TOKEN_MOD      // '%'
	TOKEN_POWER    // **

	TOKEN_LPAREN // (
	TOKEN_RPAREN // )

	TOKEN_NUMBER     //10 or 10.1
	TOKEN_IDENTIFIER //identifier

	//reserved keywords
	TOKEN_TRUE  //true
	TOKEN_FALSE //false
	TOKEN_NIL   // nil
)

//for debug & testing
func (tt TokenType) String() string {
	switch tt {
	case TOKEN_ILLEGAL:
		return "ILLEGAL"
	case TOKEN_EOF:
		return "EOF"

	case TOKEN_PLUS:
		return "+"
	case TOKEN_MINUS:
		return "-"
	case TOKEN_MULTIPLY:
		return "*"
	case TOKEN_DIVIDE:
		return "/"
	case TOKEN_MOD:
		return "%"
	case TOKEN_POWER:
		return "**"

	case TOKEN_LPAREN:
		return "("
	case TOKEN_RPAREN:
		return ")"

	case TOKEN_NUMBER:
		return "NUMBER"
	case TOKEN_IDENTIFIER:
		return "IDENTIFIER"

	case TOKEN_TRUE:
		return "TRUE"
	case TOKEN_FALSE:
		return "FALSE"
	case TOKEN_NIL:
		return "NIL"
	default:
		return "UNKNOWN"
	}
}

var keywords = map[string]TokenType{
	"true":  TOKEN_TRUE,
	"false": TOKEN_FALSE,
	"nil":   TOKEN_NIL,
}

type Token struct {
	Pos     Position
	Type    TokenType
	Literal string
}

//Stringer method for Token
func (t Token) String() string {
	return fmt.Sprintf("Position: %s, Type: %s, Literal: %s", t.Pos, t.Type, t.Literal)
}

//Position is the location of a code point in the source
type Position struct {
	Filename string
	Offset   int //offset relative to entire file
	Line     int
	Col      int //offset relative to each line
}

//Stringer method for Position
func (p Position) String() string {
	var msg string
	if p.Filename == "" {
		msg = fmt.Sprint(" <", p.Line, ":", p.Col, "> ")
	} else {
		msg = fmt.Sprint(" <", p.Filename, ":", p.Line, ":", p.Col, "> ")
	}

	return msg
}

//We could not use `Line()` as function name, because `Line` is the struct's field
func (p Position) Sline() string { //String line
	var msg string
	if p.Filename == "" {
		msg = fmt.Sprint(p.Line)
	} else {
		msg = fmt.Sprint(" <", p.Filename, ":", p.Line, "> ")
	}
	return msg
}

func LookupIdent(ident string) TokenType {
	if tok, ok := keywords[ident]; ok {
		return tok
	}
	return TOKEN_IDENTIFIER
}
