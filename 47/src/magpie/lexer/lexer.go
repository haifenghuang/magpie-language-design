package lexer

import (
	"errors"
	"fmt"
	"io/ioutil"
	"magpie/token"
	"strings"
	"unicode"
)

var prevToken token.Token

// Lexer
type Lexer struct {
	Filename     string
	input        []rune
	ch           rune //current character
	position     int  //character offset
	readPosition int  //reading offset

	line int
	col  int
}

func NewFileLexer(filename string) (*Lexer, error) {
	f, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	l := NewLexer(string(f))
	l.Filename = filename
	return l, nil
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
		if l.peek() == '+' {
			tok = token.Token{Type: token.TOKEN_INCREMENT, Literal: string(l.ch) + string(l.peek())}
			l.readNext()
		} else if l.peek() == '=' {
			tok = token.Token{Type: token.TOKEN_PLUS_A, Literal: string(l.ch) + string(l.peek())}
			l.readNext()
		} else {
			tok = newToken(token.TOKEN_PLUS, l.ch)
		}
	case '-':
		if l.peek() == '-' {
			tok = token.Token{Type: token.TOKEN_DECREMENT, Literal: string(l.ch) + string(l.peek())}
			l.readNext()
		} else if l.peek() == '=' {
			tok = token.Token{Type: token.TOKEN_MINUS_A, Literal: string(l.ch) + string(l.peek())}
			l.readNext()
		} else {
			tok = newToken(token.TOKEN_MINUS, l.ch)
		}
	case '*':
		if l.peek() == '*' {
			tok = token.Token{Type: token.TOKEN_POWER, Literal: string(l.ch) + string(l.peek())}
			l.readNext()
		} else if l.peek() == '=' {
			tok = token.Token{Type: token.TOKEN_ASTERISK_A, Literal: string(l.ch) + string(l.peek())}
			l.readNext()
		} else {
			tok = newToken(token.TOKEN_MULTIPLY, l.ch)
		}
	case '/':
		if l.peek() == '/' {
			l.readNext()
			l.skipComment()
			return l.NextToken()
		} else if l.peek() == '*' {
			l.readNext()
			err := l.skipMultilineComment()
			if err == nil {
				return l.NextToken()
			} else {
				tok.Type = token.TOKEN_ILLEGAL
				tok.Pos = pos
				tok.Literal = err.Error()
				return tok
			}
		}

		// '/'通常表示除法，但是也可能是一个正则表达式
		if prevToken.Type == token.TOKEN_RPAREN || // (a+c) / b
			prevToken.Type == token.TOKEN_RBRACKET || // a[3] / b
			prevToken.Type == token.TOKEN_IDENTIFIER || // a / b
			prevToken.Type == token.TOKEN_NUMBER { // 3 / b,  3.5 / b
			if l.peek() == '=' {
				tok = token.Token{Type: token.TOKEN_SLASH_A, Literal: string(l.ch) + string(l.peek())}
				l.readNext()
			} else {
				tok = newToken(token.TOKEN_DIVIDE, l.ch)
			}
		} else { //regexp
			if regStr, err := l.readRegExLiteral(); err == nil {
				tok.Literal = regStr
				tok.Type = token.TOKEN_REGEX
				tok.Pos = pos
				return tok
			} else {
				tok.Type = token.TOKEN_ILLEGAL
				tok.Pos = pos
				tok.Literal = err.Error()
				return tok
			}
		}
	case '%':
		if l.peek() == '=' {
			tok = token.Token{Type: token.TOKEN_MOD_A, Literal: string(l.ch) + string(l.peek())}
			l.readNext()
		} else {
			tok = newToken(token.TOKEN_MOD, l.ch)
		}
	case '(':
		tok = newToken(token.TOKEN_LPAREN, l.ch)
	case ')':
		tok = newToken(token.TOKEN_RPAREN, l.ch)
	case '{':
		tok = newToken(token.TOKEN_LBRACE, l.ch)
	case '}':
		tok = newToken(token.TOKEN_RBRACE, l.ch)
	case '@':
		tok = newToken(token.TOKEN_AT, l.ch)
	case '=':
		if l.peek() == '=' {
			tok = token.Token{Type: token.TOKEN_EQ, Literal: string(l.ch) + string(l.peek())}
			l.readNext()
		} else if l.peek() == '~' {
			tok = token.Token{Type: token.TOKEN_MATCH, Literal: string(l.ch) + string(l.peek())}
			l.readNext()
		} else if l.peek() == '>' {
			tok = token.Token{Type: token.TOKEN_FATARROW, Literal: string(l.ch) + string(l.peek())}
			l.readNext()
		} else {
			tok = newToken(token.TOKEN_ASSIGN, l.ch)
		}
	case '>':
		if l.peek() == '=' {
			tok = token.Token{Type: token.TOKEN_GE, Literal: string(l.ch) + string(l.peek())}
			l.readNext()
		} else {
			tok = newToken(token.TOKEN_GT, l.ch)
		}
	case '<':
		if l.peek() == '=' {
			tok = token.Token{Type: token.TOKEN_LE, Literal: string(l.ch) + string(l.peek())}
			l.readNext()
		} else {
			tok = newToken(token.TOKEN_LT, l.ch)
		}
	case '!':
		if l.peek() == '=' {
			tok = token.Token{Type: token.TOKEN_NEQ, Literal: string(l.ch) + string(l.peek())}
			l.readNext()
		} else if l.peek() == '~' {
			tok = token.Token{Type: token.TOKEN_NOTMATCH, Literal: string(l.ch) + string(l.peek())}
			l.readNext()
		} else {
			tok = newToken(token.TOKEN_BANG, l.ch)
		}
	case ';':
		tok = newToken(token.TOKEN_SEMICOLON, l.ch)
	case ':':
		tok = newToken(token.TOKEN_COLON, l.ch)
	case ',':
		tok = newToken(token.TOKEN_COMMA, l.ch)
	case '.':
		if l.peek() == '.' {
			l.readNext()
			if l.peek() == '.' {
				tok = token.Token{Type: token.TOKEN_ELLIPSIS, Literal: "..."}
				l.readNext()
			} else {
				tok = token.Token{Type: token.TOKEN_DOTDOT, Literal: ".."}
			}
		} else {
			tok = newToken(token.TOKEN_DOT, l.ch)
		}
	case '[':
		tok = newToken(token.TOKEN_LBRACKET, l.ch)
	case ']':
		tok = newToken(token.TOKEN_RBRACKET, l.ch)
	case '&':
		if l.peek() == '&' {
			tok = token.Token{Type: token.TOKEN_AND, Literal: string(l.ch) + string(l.peek())}
			l.readNext()
		}
	case '|':
		if l.peek() == '|' {
			tok = token.Token{Type: token.TOKEN_OR, Literal: string(l.ch) + string(l.peek())}
			l.readNext()
		}
	case '#': //comment
		l.skipComment()
		return l.NextToken()
	case 0:
		tok.Literal = "<EOF>"
		tok.Type = token.TOKEN_EOF
	default:
		if isDigit(l.ch) {
			tok.Literal = l.readNumber()
			tok.Type = token.TOKEN_NUMBER
			tok.Pos = pos
			prevToken = tok
			return tok
		} else if isLetter(l.ch) {
			tok.Literal = l.readIdentifier()
			tok.Pos = pos
			tok.Type = token.LookupIdent(tok.Literal)
			prevToken = tok
			return tok
		} else if l.ch == 34 { //double quotes
			if s, err := l.readString(l.ch); err == nil {
				tok.Type = token.TOKEN_STRING
				tok.Pos = pos
				tok.Literal = s
				prevToken = tok
				return tok
			} else {
				tok.Type = token.TOKEN_ILLEGAL
				tok.Pos = pos
				tok.Literal = err.Error()
				return tok
			}
		} else {
			tok = newToken(token.TOKEN_ILLEGAL, l.ch)
		}
	}

	tok.Pos = pos
	l.readNext()
	prevToken = tok
	return tok
}

// read a regexp, including flags.
func (l *Lexer) readRegExLiteral() (string, error) {
	out := ""

	for {
		l.readNext()

		if l.ch == 0 {
			return "unterminated regular expression", fmt.Errorf("unterminated regular expression")
		}

		if l.ch == '/' { // consume the terminating "/".
			l.readNext()
			flags := "" // regexp flags(or modifier)

			// Below flags are supported:
			//   i -> case-insensitive
			//   m -> multi-line mode
			//   s -> let . match \n
			//   U -> ungreedy
			for l.ch == 'i' || l.ch == 'm' || l.ch == 's' || l.ch == 'U' {
				// save the char - unless it is a repeat
				if !strings.Contains(flags, string(l.ch)) {
					// we're going to sort the flags
					tmp := strings.Split(flags, "")
					tmp = append(tmp, string(l.ch))
					flags = strings.Join(tmp, "")
				}

				l.readNext()
			}

			// convert the regexp to go-lang
			if len(flags) > 0 {
				out = "(?" + flags + ")" + out
			}
			break
		}
		out = out + string(l.ch)
	}

	return out, nil
}

func (l *Lexer) readNumber() string {
	var ret []rune

	ch := l.ch
	ret = append(ret, ch)
	l.readNext()

	for isDigit(l.ch) || l.ch == '.' {
		if l.ch == '.' {
			if !isDigit(l.peek()) { //should be a method calling, e.g. 10.2.floor()
				return string(ret)
			}
		} //end if
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

func (l *Lexer) readString(r rune) (string, error) {
	var ret []rune
eos:
	for {
		l.readNext()
		switch l.ch {
		case '\n':
			return "", errors.New("unexpected EOL")
		case 0:
			return "", errors.New("unexpected EOF")
		case r:
			l.readNext()
			break eos //eos:end of string
		case '\\':
			l.readNext()
			switch l.ch {
			case 'b':
				ret = append(ret, '\b')
				continue
			case 'f':
				ret = append(ret, '\f')
				continue
			case 'r':
				ret = append(ret, '\r')
				continue
			case 'n':
				ret = append(ret, '\n')
				continue
			case 't':
				ret = append(ret, '\t')
				continue
			}
			ret = append(ret, l.ch)
			continue
		default:
			ret = append(ret, l.ch)
		}
	}

	return string(ret), nil
}

func (l *Lexer) skipWhitespace() {
	for unicode.IsSpace(l.ch) {
		l.readNext()
	}
}

func (l *Lexer) skipComment() {
	for l.ch != '\n' && l.ch != 0 {
		l.readNext()
	}
}

func (l *Lexer) skipMultilineComment() error {
	var err error = nil
loop:
	for {
		l.readNext()
		switch l.ch {
		case '*':
			switch l.peek() {
			case '/': // '*/'
				l.readNext() //skip the '*'
				l.readNext() //skip the '/'
				break loop
			}
		case 0: // Got EOF, which means unterminated multiline comment.
			err = errors.New("Unterminated multiline comment, GOT EOF!")
			break loop
		}
	}
	return err
}

func (l *Lexer) getPos() token.Position {
	return token.Position{
		Filename: l.Filename,
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
	return unicode.IsLetter(ch) || ch == '_' || ch == '$'
}
