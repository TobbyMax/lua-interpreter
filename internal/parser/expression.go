package parser

import (
	"strconv"

	"lua-interpreter/internal/lexer"
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
	}
)

func (p *Parser) parseExpression() (Expression, error) {
	var (
		// Sets the priority of the operators (power - highest, or - lowest)
		//parsePower  = p.parseBinaryExp(p.parseExpressionBase, lexer.TokenPower)
		//parseUnary  = p.parseUnaryExp(parsePower, UnaryOperators...)
		//parseMulDiv = p.parseBinaryExp(parseUnary, MultiplicativeOperators...)
		parseAddSub = p.parseBinaryExp(p.parseExpressionBase, AdditiveOperators...)
		//parseConcat = p.parseBinaryExp(parseAddSub, lexer.TokenDoubleDot)
		//parseShift  = p.parseBinaryExp(parseConcat, ShiftOperators...)
		//parseBinAnd = p.parseBinaryExp(parseShift, lexer.TokenBinAnd)
		//parseBinXor = p.parseBinaryExp(parseBinAnd, lexer.TokenTilde)
		//parseBinOr  = p.parseBinaryExp(parseBinXor, lexer.TokenBinOr)
		//parseComp   = p.parseBinaryExp(parseBinOr, ComparisonOperators...)
		//parseAnd    = p.parseBinaryExp(parseComp, lexer.TokenKeywordAnd)
		//parseOr     = p.parseBinaryExp(parseAnd, lexer.TokenKeywordOr)
	)
	return parseAddSub()
}

// exp ::=  nil | false | true | Numeral | LiteralString | ‘...’
//
//	| functiondef | prefixexp | tableconstructor | opunary exp
//	| exp binop exp
func (p *Parser) parseExpressionBase() (Expression, error) {
	switch p.currentToken.Type {
	case lexer.TokenKeywordNil:
		p.currentToken = p.lexer.NextToken()
		return &NilExpression{}, nil
	case lexer.TokenKeywordFalse:
		p.currentToken = p.lexer.NextToken()
		return &BooleanExpression{Value: false}, nil
	case lexer.TokenKeywordTrue:
		p.currentToken = p.lexer.NextToken()
		return &BooleanExpression{Value: true}, nil
	case lexer.TokenNumeral:
		str := p.currentToken.Value
		p.currentToken = p.lexer.NextToken()
		num, err := strconv.ParseFloat(str, 64)
		if err != nil {
			return nil, err
		}
		return &NumeralExpression{Value: num}, nil
	case lexer.TokenLiteralString:
		str := p.currentToken.Value
		p.currentToken = p.lexer.NextToken()
		return &LiteralString{Value: str}, nil
	case lexer.TokenTripleDot:
		p.currentToken = p.lexer.NextToken()
		return &VarArgExpression{}, nil
	//case lexer.TokenKeywordFunction:
	//	p.currentToken = p.lexer.NextToken()
	//	return p.parseFunctionDef()
	//case lexer.TokenLeftBrace:
	//	p.currentToken = p.lexer.NextToken()
	//	return p.parseTableConstructor()
	//case lexer.TokenIdentifier, lexer.TokenLeftParen:
	//	return p.parsePrefixExpression()
	default:
		return nil, nil
	}
}

func (p *Parser) parseBinaryExp(next func() (Expression, error), tokenTypes ...lexer.TokenType) func() (Expression, error) {
	return func() (Expression, error) {
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
			exp = &BinaryOperatorExpression{
				Operator: op,
				Left:     exp,
				Right:    right,
			}
		}
		return exp, nil
	}
}

func (p *Parser) parseUnaryExp(next func() (Expression, error), tokenTypes ...lexer.TokenType) func() (Expression, error) {
	return func() (Expression, error) {
		if isOneOfTypes(p.currentToken.Type, tokenTypes) {
			op := p.currentToken
			p.currentToken = p.lexer.NextToken()
			exp, err := next()
			if err != nil {
				return nil, err
			}
			return &UnaryOperatorExpression{
				Operator:   op,
				Expression: exp,
			}, nil
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
