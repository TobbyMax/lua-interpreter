package main

import (
	"errors"
	"fmt"
	"strconv"
)

type Node interface {
	Eval(env *Env) error
}

type NumberNode struct {
	Value float64
}

func (n *NumberNode) Eval(env *Env) error {
	env.Push(n.Value)
	return nil
}

type VarNode struct {
	Name string
}

func (n *VarNode) Eval(env *Env) error {
	val, ok := env.vars[n.Name]
	if !ok {
		return errors.New("undefined variable: " + n.Name)
	}
	env.Push(val)
	return nil
}

type AssignNode struct {
	Name string
	Expr Node
}

func (n *AssignNode) Eval(env *Env) error {
	err := n.Expr.Eval(env)
	if err != nil {
		return err
	}
	val := env.Pop()
	env.vars[n.Name] = val
	return nil
}

type PrintNode struct {
	Expr Node
}

func (n *PrintNode) Eval(env *Env) error {
	err := n.Expr.Eval(env)
	if err != nil {
		return err
	}
	fmt.Println(env.Pop())
	return nil
}

type BinOpNode struct {
	Left  Node
	Op    string
	Right Node
}

func (n *BinOpNode) Eval(env *Env) error {
	if err := n.Left.Eval(env); err != nil {
		return err
	}
	left := env.Pop()
	if err := n.Right.Eval(env); err != nil {
		return err
	}
	right := env.Pop()

	var result float64
	switch n.Op {
	case "+":
		result = left + right
	case "-":
		result = left - right
	case "*":
		result = left * right
	case "/":
		result = left / right
	}
	env.Push(result)
	return nil
}

func Parse(tokens []Token) (Node, error) {
	pos := 0
	next := func() Token {
		if pos >= len(tokens) {
			return Token{TokEOF, ""}
		}
		tok := tokens[pos]
		pos++
		return tok
	}
	peek := func() Token {
		if pos >= len(tokens) {
			return Token{TokEOF, ""}
		}
		return tokens[pos]
	}

	parseExpr := func() (Node, error) {
		tok := next()
		switch tok.Type {
		case TokNumber:
			val, _ := strconv.ParseFloat(tok.Value, 64)
			return &NumberNode{val}, nil
		case TokIdent:
			return &VarNode{tok.Value}, nil
		default:
			return nil, errors.New("unexpected token: " + tok.Value)
		}
	}

	parse := func() (Node, error) {
		tok := next()
		switch tok.Type {
		case TokIdent:
			if peek().Type == TokAssign {
				next() // consume '='
				expr, err := parseExpr()
				if err != nil {
					return nil, err
				}
				return &AssignNode{tok.Value, expr}, nil
			}
		case TokPrint:
			expr, err := parseExpr()
			if err != nil {
				return nil, err
			}
			return &PrintNode{expr}, nil
		}
		return nil, errors.New("unknown statement")
	}

	return parse()
}
