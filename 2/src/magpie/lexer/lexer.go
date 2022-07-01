package lexer

import (
	"magpie/token"
	"unicode"
)

// Lexer
type Lexer struct {
	filename     string
	input        []rune
	ch           rune //current character
	position     int  //character offset
	readPosition int  //reading offset

	line int
	col  int
}

func NewLexer(input string) *Lexer {
	l := &Lexer{input: []rune(input)}
	l.ch = ' '
	l.position = 0
	l.readPosition = 0

	l.line = 1
	l.col = 0

	l.readNext()
	//0xFEFF: BOM(byte order mark), only permitted as very first character
	if l.ch == 0xFEFF {
		l.readNext() //ignore BOM at file beginning
	}

	return l
}

func (l *Lexer) readNext() {
	if l.readPosition >= len(l.input) {
		l.ch = 0
	} else {
		l.ch = l.input[l.readPosition]
		if l.ch == '\n' {
			l.col = 0
			l.line++
		} else {
			l.col += 1
		}
	}

	l.position = l.readPosition
	l.readPosition++
}

func (l *Lexer) peek() rune {
	if l.readPosition >= len(l.input) {
		return 0
	}
	return l.input[l.readPosition]
}

func (l *Lexer) NextToken() token.Token {
	var tok token.Token
	l.skipWhitespace()

	pos := l.getPos()

	switch l.ch {
	case '+':
		tok = newToken(token.TOKEN_PLUS, l.ch)
	case '-':
		tok = newToken(token.TOKEN_MINUS, l.ch)
	case '*':
		if l.peek() == '*' {
			tok = token.Token{Type: token.TOKEN_POWER, Literal: string(l.ch) + string(l.peek())}
			l.readNext()
		} else {
			tok = newToken(token.TOKEN_MULTIPLY, l.ch)
		}
	case '/':
		tok = newToken(token.TOKEN_DIVIDE, l.ch)
	case '%':
		tok = newToken(token.TOKEN_MOD, l.ch)
	case '(':
		tok = newToken(token.TOKEN_LPAREN, l.ch)
	case ')':
		tok = newToken(token.TOKEN_RPAREN, l.ch)
	case 0:
		tok.Literal = "<EOF>"
		tok.Type = token.TOKEN_EOF
	default:
		if isDigit(l.ch) {
			tok.Literal = l.readNumber()
			tok.Type = token.TOKEN_NUMBER
			tok.Pos = pos
			return tok
		} else if isLetter(l.ch) {
			tok.Literal = l.readIdentifier()
			tok.Pos = pos
			tok.Type = token.LookupIdent(tok.Literal)
			return tok
		} else {
			tok = newToken(token.TOKEN_ILLEGAL, l.ch)
		}
	}

	tok.Pos = pos
	l.readNext()
	return tok
}

func (l *Lexer) readNumber() string {
	var ret []rune

	ch := l.ch
	ret = append(ret, ch)
	l.readNext()

	for isDigit(l.ch) || l.ch == '.' {
		ret = append(ret, l.ch)
		l.readNext()
	}

	return string(ret)
}

func (l *Lexer) readIdentifier() string {
	position := l.position
	for isLetter(l.ch) || isDigit(l.ch) {
		l.readNext()
	}
	return string(l.input[position:l.position])
}

func (l *Lexer) skipWhitespace() {
	for unicode.IsSpace(l.ch) {
		l.readNext()
	}
}

func (l *Lexer) getPos() token.Position {
	return token.Position{
		Filename: l.filename,
		Offset:   l.position,
		Line:     l.line,
		Col:      l.col,
	}
}

func newToken(tokenType token.TokenType, ch rune) token.Token {
	return token.Token{Type: tokenType, Literal: string(ch)}
}

func isDigit(ch rune) bool {
	return '0' <= ch && ch <= '9'
}

func isLetter(ch rune) bool {
	return unicode.IsLetter(ch) || ch == '_'
}
