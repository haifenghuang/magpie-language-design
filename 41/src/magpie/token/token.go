package token

import (
	"fmt"
)

// token
type TokenType int

const (
	TOKEN_ILLEGAL TokenType = (iota - 1) // Illegal token
	TOKEN_EOF                            //End Of File

	TOKEN_PLUS      // +
	TOKEN_MINUS     // -
	TOKEN_MULTIPLY  // *
	TOKEN_DIVIDE    // '/'
	TOKEN_MOD       // '%'
	TOKEN_POWER     // **
	TOKEN_INCREMENT // ++
	TOKEN_DECREMENT // --

	TOKEN_LPAREN    // (
	TOKEN_RPAREN    // )
	TOKEN_ASSIGN    // =
	TOKEN_SEMICOLON //;
	TOKEN_COLON     //:
	TOKEN_COMMA     //,
	TOKEN_DOT       //.
	TOKEN_DOTDOT    //..
	TOKEN_LBRACE    // {
	TOKEN_RBRACE    // }
	TOKEN_BANG      // !
	TOKEN_LBRACKET  // [
	TOKEN_RBRACKET  // ]
	TOKEN_COMMENT   // #

	TOKEN_LT       // <
	TOKEN_LE       // <=
	TOKEN_GT       // >
	TOKEN_GE       // >=
	TOKEN_EQ       // ==
	TOKEN_NEQ      // !=
	TOKEN_MATCH    // =~
	TOKEN_NOTMATCH // !~
	TOKEN_FATARROW // =>

	TOKEN_AND // &&
	TOKEN_OR  // ||

	TOKEN_NUMBER     //10 or 10.1
	TOKEN_IDENTIFIER //identifier
	TOKEN_STRING     //""

	//reserved keywords
	TOKEN_TRUE        //true
	TOKEN_FALSE       //false
	TOKEN_NIL         // nil
	TOKEN_LET         //let
	TOKEN_RETURN      //return
	TOKEN_FUNCTION    //fn
	TOKEN_IF          //if
	TOKEN_ELSE        //else
	TOKEN_WHILE       //while
	TOKEN_DO          //do
	TOKEN_FOR         //for
	TOKEN_IN          //in
	TOKEN_BREAK       //break
	TOKEN_CONTINUE    //continue
	TOKEN_IMPORT      //import
	TOKEN_STRUCT      //struct
	TOKEN_SWITCH      //switch
	TOKEN_CASE        //case
	TOKEN_DEFAULT     //default
	TOKEN_FALLTHROUGH //fallthrough
	TOKEN_TRY         //try
	TOKEN_CATCH       //catch
	TOKEN_FINALLY     //finally
	TOKEN_THROW       //throw

	TOKEN_REGEX // regular expression
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
	case TOKEN_INCREMENT:
		return "++"
	case TOKEN_DECREMENT:
		return "--"
	case TOKEN_LPAREN:
		return "("
	case TOKEN_RPAREN:
		return ")"
	case TOKEN_ASSIGN:
		return "="
	case TOKEN_SEMICOLON:
		return ";"
	case TOKEN_COLON:
		return ":"
	case TOKEN_COMMA:
		return ","
	case TOKEN_DOT:
		return "."
	case TOKEN_DOTDOT:
		return ".."
	case TOKEN_LBRACE:
		return "{"
	case TOKEN_RBRACE:
		return "}"
	case TOKEN_BANG:
		return "!"
	case TOKEN_LBRACKET:
		return "["
	case TOKEN_RBRACKET:
		return "]"
	case TOKEN_COMMENT:
		return "#"

	case TOKEN_LT:
		return "<"
	case TOKEN_LE:
		return "<="
	case TOKEN_GT:
		return ">"
	case TOKEN_GE:
		return ">="
	case TOKEN_EQ:
		return "=="
	case TOKEN_NEQ:
		return "!="
	case TOKEN_MATCH:
		return "=~"
	case TOKEN_NOTMATCH:
		return "!~"
	case TOKEN_FATARROW:
		return "=>"

	case TOKEN_AND:
		return "&&"
	case TOKEN_OR:
		return "||"

	case TOKEN_NUMBER:
		return "NUMBER"
	case TOKEN_IDENTIFIER:
		return "IDENTIFIER"
	case TOKEN_STRING:
		return "STRING"

	case TOKEN_TRUE:
		return "TRUE"
	case TOKEN_FALSE:
		return "FALSE"
	case TOKEN_NIL:
		return "NIL"
	case TOKEN_LET:
		return "LET"
	case TOKEN_RETURN:
		return "RETURN"
	case TOKEN_FUNCTION:
		return "FUNCTION"
	case TOKEN_IF:
		return "IF"
	case TOKEN_ELSE:
		return "ELSE"
	case TOKEN_WHILE:
		return "WHILE"
	case TOKEN_DO:
		return "DO"
	case TOKEN_FOR:
		return "FOR"
	case TOKEN_IN:
		return "IN"
	case TOKEN_BREAK:
		return "BREAK"
	case TOKEN_CONTINUE:
		return "CONTINUE"
	case TOKEN_IMPORT:
		return "IMPORT"
	case TOKEN_STRUCT:
		return "STRUCT"
	case TOKEN_SWITCH:
		return "SWITCH"
	case TOKEN_CASE:
		return "CASE"
	case TOKEN_DEFAULT:
		return "DEFAULT"
	case TOKEN_FALLTHROUGH:
		return "FALLTHROUGH"
	case TOKEN_TRY:
		return "TRY"
	case TOKEN_CATCH:
		return "CATCH"
	case TOKEN_FINALLY:
		return "FINALLY"
	case TOKEN_THROW:
		return "THROW"

	case TOKEN_REGEX:
		return "<REGEX>"
	default:
		return "UNKNOWN"
	}
}

var keywords = map[string]TokenType{
	"true":        TOKEN_TRUE,
	"false":       TOKEN_FALSE,
	"nil":         TOKEN_NIL,
	"let":         TOKEN_LET,
	"return":      TOKEN_RETURN,
	"fn":          TOKEN_FUNCTION,
	"if":          TOKEN_IF,
	"else":        TOKEN_ELSE,
	"while":       TOKEN_WHILE,
	"do":          TOKEN_DO,
	"for":         TOKEN_FOR,
	"in":          TOKEN_IN,
	"break":       TOKEN_BREAK,
	"continue":    TOKEN_CONTINUE,
	"import":      TOKEN_IMPORT,
	"struct":      TOKEN_STRUCT,
	"switch":      TOKEN_SWITCH,
	"case":        TOKEN_CASE,
	"default":     TOKEN_DEFAULT,
	"fallthrough": TOKEN_FALLTHROUGH,
	"try":         TOKEN_TRY,
	"catch":       TOKEN_CATCH,
	"finally":     TOKEN_FINALLY,
	"throw":       TOKEN_THROW,
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
