package parser

import (
	"errors"
	"fmt"
	"strconv"

	"lua-interpreter/internal/ast"
	"lua-interpreter/internal/lexer"
	"lua-interpreter/internal/optimizer"
)

var (
	ErrNoFunctionCallPostfix = errors.New("expected function call arguments or ':' Name args after prefix expression")
)

var (
	UnaryOperators = []lexer.TokenType{
		lexer.TokenMinus,
		lexer.TokenNot,
		lexer.TokenTilde,
		lexer.TokenHash,
	}
	MultiplicativeOperators = []lexer.TokenType{
		lexer.TokenMult,
		lexer.TokenDiv,
		lexer.TokenIntDiv,
		lexer.TokenMod,
	}
	AdditiveOperators = []lexer.TokenType{
		lexer.TokenPlus,
		lexer.TokenMinus,
	}
	ShiftOperators = []lexer.TokenType{
		lexer.TokenShiftLeft,
		lexer.TokenShiftRight,
	}
	ComparisonOperators = []lexer.TokenType{
		lexer.TokenLess,
		lexer.TokenLessEqual,
		lexer.TokenMore,
		lexer.TokenMoreEqual,
		lexer.TokenEqual,
		lexer.TokenNotEqual,
	}
)

func (p *Parser) parseExpression() (ast.Expression, error) {
	var (
		// Sets the priority of the operators (power - highest, or - lowest)
		// https://www.lua.org/manual/5.3/manual.html#3.4.8
		// 1 - power
		parsePower = p.parseBinaryExp(p.parseExpressionBase, lexer.TokenPower)
		// 2 - unary
		parseUnary = p.parseUnaryExp(parsePower, UnaryOperators...)
		// 3 - multiplicative
		parseMulDiv = p.parseBinaryExp(parseUnary, MultiplicativeOperators...)
		// 4 - additive
		parseAddSub = p.parseBinaryExp(parseMulDiv, AdditiveOperators...)
		// 5 - concatenation
		parseConcat = p.parseBinaryExp(parseAddSub, lexer.TokenDoubleDot)
		// 6 - shift
		parseShift = p.parseBinaryExp(parseConcat, ShiftOperators...)
		// 7 - bitwise AND
		parseBinAnd = p.parseBinaryExp(parseShift, lexer.TokenBinAnd)
		// 8 - bitwise XOR
		parseBinXor = p.parseBinaryExp(parseBinAnd, lexer.TokenTilde)
		// 9 - bitwise OR
		parseBinOr = p.parseBinaryExp(parseBinXor, lexer.TokenBinOr)
		// 10 - comparison
		parseComp = p.parseBinaryExp(parseBinOr, ComparisonOperators...)
		// 11 - logical AND
		parseAnd = p.parseBinaryExp(parseComp, lexer.TokenKeywordAnd)
		// 12 - logical OR
		parseOr = p.parseBinaryExp(parseAnd, lexer.TokenKeywordOr)
	)
	return parseOr()
}

// exp ::=  nil | false | true | Numeral | LiteralString | ‘...’
//
//	| functiondef | prefixexp | tableconstructor
//	| opunary exp | exp binop exp
func (p *Parser) parseExpressionBase() (ast.Expression, error) {
	switch p.currentToken.Type {
	case lexer.TokenKeywordNil:
		p.currentToken = p.lexer.NextToken()
		return &ast.NilExpression{}, nil
	case lexer.TokenKeywordFalse:
		p.currentToken = p.lexer.NextToken()
		return &ast.BooleanExpression{Value: false}, nil
	case lexer.TokenKeywordTrue:
		p.currentToken = p.lexer.NextToken()
		return &ast.BooleanExpression{Value: true}, nil
	case lexer.TokenNumeral:
		str := p.currentToken.Value
		p.currentToken = p.lexer.NextToken()
		// todo: parse hexadecimal numbers
		num, err := strconv.ParseFloat(str, 64)
		if err != nil {
			return nil, err
		}
		return &ast.NumeralExpression{Value: num}, nil
	case lexer.TokenLiteralString:
		str := p.currentToken.Value
		p.currentToken = p.lexer.NextToken()
		return &ast.LiteralString{Value: str}, nil
	case lexer.TokenTripleDot:
		p.currentToken = p.lexer.NextToken()
		return &ast.VarArgExpression{}, nil
	case lexer.TokenKeywordFunction:
		return p.parseFunctionDefinition()
	case lexer.TokenIdentifier, lexer.TokenLeftParen:
		return p.parsePrefixExpression()
	case lexer.TokenLeftBrace:
		return p.parseTableConstructor()
	default:
		return nil, fmt.Errorf("unexpected token: [%s] %s", p.currentToken.Type, p.currentToken.Value)
	}
}

func (p *Parser) parseTableConstructor() (*ast.TableConstructorExpression, error) {
	p.currentToken = p.lexer.NextToken()
	var fields []ast.Field
	for p.currentToken.Type != lexer.TokenRightBrace {
		if p.currentToken.Type == lexer.TokenComma {
			p.currentToken = p.lexer.NextToken()
			continue
		}
		field, err := p.parseField()
		if err != nil {
			return nil, err
		}
		fields = append(fields, field)
		if p.currentToken.Type == lexer.TokenRightBrace {
			p.currentToken = p.lexer.NextToken()
			break
		} else if p.currentToken.Type != lexer.TokenComma {
			return nil, errors.New("expected ',' or '}'")
		}
		p.currentToken = p.lexer.NextToken()
	}
	return &ast.TableConstructorExpression{Fields: fields}, nil
}

// field ::= ‘[’ exp ‘]’ ‘=’ exp | Name ‘=’ exp | exp
func (p *Parser) parseField() (ast.Field, error) {
	switch p.currentToken.Type {
	case lexer.TokenLeftBracket:
		p.currentToken = p.lexer.NextToken()
		key, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		if p.currentToken.Type != lexer.TokenRightBracket {
			return nil, errors.New("missing ']'")
		}
		p.currentToken = p.lexer.NextToken()
		if p.currentToken.Type != lexer.TokenAssign {
			return nil, errors.New("missing '='")
		}
		p.currentToken = p.lexer.NextToken()
		value, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		return &ast.ExpToExpField{Key: key, Value: value}, nil
	case lexer.TokenIdentifier:
		name := p.currentToken.Value
		p.currentToken = p.lexer.NextToken()
		if p.currentToken.Type != lexer.TokenAssign {
			exp, err := p.parsePrefixExpressionTail(&ast.NameVar{Name: name})
			if err != nil {
				return nil, err
			}
			return &ast.ExpressionField{
				Value: exp,
			}, nil
		}
		p.currentToken = p.lexer.NextToken()
		value, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		return &ast.NameField{Name: name, Value: value}, nil
	default:
		exp, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		return &ast.ExpressionField{Value: exp}, nil
	}
}

func (p *Parser) parseBinaryExp(next func() (ast.Expression, error), tokenTypes ...lexer.TokenType) func() (ast.Expression, error) {
	return func() (ast.Expression, error) {
		exp, err := next()
		if err != nil {
			return nil, err
		}
		for isOneOfTypes(p.currentToken.Type, tokenTypes) {
			op := p.currentToken
			p.currentToken = p.lexer.NextToken()
			right, err := next()
			if err != nil {
				return nil, err
			}
			exp = optimizer.OptimizeBinary(op.Type)(
				&ast.BinaryOperatorExpression{
					Operator: op,
					Left:     exp,
					Right:    right,
				},
			)
		}
		return exp, nil
	}
}

func (p *Parser) parseUnaryExp(next func() (ast.Expression, error), tokenTypes ...lexer.TokenType) func() (ast.Expression, error) {
	return func() (ast.Expression, error) {
		if isOneOfTypes(p.currentToken.Type, tokenTypes) {
			op := p.currentToken
			p.currentToken = p.lexer.NextToken()
			exp, err := next()
			if err != nil {
				return nil, err
			}
			return optimizer.OptimizeUnary(op.Type)(
				&ast.UnaryOperatorExpression{
					Operator:   op,
					Expression: exp,
				},
			), nil
		}
		return next()
	}
}

func isOneOfTypes(tokenType lexer.TokenType, tokenTypes []lexer.TokenType) bool {
	for _, t := range tokenTypes {
		if tokenType == t {
			return true
		}
	}
	return false
}
