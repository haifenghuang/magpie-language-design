package ast

import (
	"bytes"
	"magpie/token"
	"unicode/utf8"
)

type Node interface {
	Pos() token.Position // position of first character belonging to the node
	End() token.Position // position of first character immediately after the node

	TokenLiteral() string
	String() string
}

type Expression interface {
	Node
	expressionNode()
}

type Program struct {
	Expression Expression
}

func (p *Program) Pos() token.Position {
	return p.Expression.Pos()
}

func (p *Program) End() token.Position {
	return p.Expression.End()
}

func (p *Program) TokenLiteral() string {
	return p.Expression.TokenLiteral()
}

func (p *Program) String() string {
	var out bytes.Buffer

	out.WriteString(p.Expression.String())
	return out.String()
}

// 1 + 2 * 3
type InfixExpression struct {
	Token    token.Token
	Operator string
	Right    Expression
	Left     Expression
}

func (ie *InfixExpression) Pos() token.Position { return ie.Token.Pos }
func (ie *InfixExpression) End() token.Position { return ie.Right.End() }

func (ie *InfixExpression) expressionNode()      {}
func (ie *InfixExpression) TokenLiteral() string { return ie.Token.Literal }
func (ie *InfixExpression) String() string {
	var out bytes.Buffer

	out.WriteString("(")
	out.WriteString(ie.Left.String())
	out.WriteString(" " + ie.Operator + " ")
	out.WriteString(ie.Right.String())
	out.WriteString(")")

	return out.String()
}

// -2, -3
type PrefixExpression struct {
	Token    token.Token
	Operator string
	Right    Expression
}

func (pe *PrefixExpression) Pos() token.Position { return pe.Token.Pos }
func (pe *PrefixExpression) End() token.Position { return pe.Right.End() }

func (pe *PrefixExpression) expressionNode()      {}
func (pe *PrefixExpression) TokenLiteral() string { return pe.Token.Literal }

func (pe *PrefixExpression) String() string {
	var out bytes.Buffer

	out.WriteString("(")
	out.WriteString(pe.Operator)
	out.WriteString(pe.Right.String())
	out.WriteString(")")

	return out.String()
}

type NumberLiteral struct {
	Token token.Token
	Value float64
}

func (nl *NumberLiteral) Pos() token.Position { return nl.Token.Pos }
func (nl *NumberLiteral) End() token.Position {
	length := utf8.RuneCountInString(nl.Token.Literal)
	pos := nl.Token.Pos
	return token.Position{Filename: pos.Filename, Line: pos.Line, Col: pos.Col + length}
}

func (nl *NumberLiteral) expressionNode()      {}
func (nl *NumberLiteral) TokenLiteral() string { return nl.Token.Literal }
func (nl *NumberLiteral) String() string       { return nl.Token.Literal }

type Identifier struct {
	Token token.Token
	Value string
}

func (i *Identifier) Pos() token.Position { return i.Token.Pos }
func (i *Identifier) End() token.Position {
	length := utf8.RuneCountInString(i.Value)
	return token.Position{Filename: i.Token.Pos.Filename, Line: i.Token.Pos.Line, Col: i.Token.Pos.Col + length}
}
func (i *Identifier) expressionNode()      {}
func (i *Identifier) TokenLiteral() string { return i.Token.Literal }
func (i *Identifier) String() string       { return i.Value }

type NilLiteral struct {
	Token token.Token
}

func (n *NilLiteral) Pos() token.Position {
	return n.Token.Pos
}

func (n *NilLiteral) End() token.Position {
	length := len(n.Token.Literal)
	pos := n.Token.Pos
	return token.Position{Filename: pos.Filename, Line: pos.Line, Col: pos.Col + length}
}

func (n *NilLiteral) expressionNode()      {}
func (n *NilLiteral) TokenLiteral() string { return n.Token.Literal }
func (n *NilLiteral) String() string       { return n.Token.Literal }

type BooleanLiteral struct {
	Token token.Token
	Value bool
}

func (b *BooleanLiteral) Pos() token.Position {
	return b.Token.Pos
}

func (b *BooleanLiteral) End() token.Position {
	length := utf8.RuneCountInString(b.Token.Literal)
	pos := b.Token.Pos
	return token.Position{Filename: pos.Filename, Line: pos.Line, Col: pos.Col + length}
}

func (b *BooleanLiteral) expressionNode()      {}
func (b *BooleanLiteral) TokenLiteral() string { return b.Token.Literal }
func (b *BooleanLiteral) String() string       { return b.Token.Literal }
