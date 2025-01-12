package parser

import (
	"errors"
	"fmt"
	"strconv"

	"lua-interpreter/internal/lexer"
)

var (
	ErrNoFunctionCallPostfix = errors.New("expected function call arguments or ':' Name args after prefix expression")
)

type (
	// FunctionDefinition
	// functiondef ::= function funcbody
	FunctionDefinition struct {
		FunctionBody FunctionBody
	}
	// ExpToExpField
	// ‘[’ exp ‘]’ ‘=’ exp
	ExpToExpField struct {
		Key   Expression
		Value Expression
	}
	// NameField
	// Name ‘=’ exp
	NameField struct {
		Name  string
		Value Expression
	}
	// ExpressionField
	// exp
	ExpressionField struct {
		Value Expression
	}
	// Field
	// field ::= ‘[’ exp ‘]’ ‘=’ exp | Name ‘=’ exp | exp
	Field interface{}
	// TableConstructorExpression
	// tableconstructor ::= ‘{’ [fieldlist] ‘}’
	TableConstructorExpression struct {
		Fields []Field
	}
	// ExpressionList
	// explist ::= exp {‘,’ exp}
	ExpressionList struct {
		Expressions []Expression
	}
	// Args [ExpressionList | TableConstructorExpression | LiteralString]
	// args ::= ‘(’ [explist] ‘)’ | tableconstructor | LiteralString
	Args interface{}
	// FunctionCall
	// functioncall ::= prefixexp args | prefixexp ‘:’ Name args
	FunctionCall struct {
		PrefixExp PrefixExpression
		Name      string
		Args      Args
	}
	// PrefixExpression
	// prefixexp ::= var | functioncall | ‘(’ exp ‘)’
	PrefixExpression        interface{}
	UnaryOperatorExpression struct {
		Operator   lexer.Token
		Expression Expression
	}
	BinaryOperatorExpression struct {
		Operator lexer.Token
		Left     Expression
		Right    Expression
	}
	// Expression
	// exp ::=  nil | false | true | Numeral | LiteralString | ‘...’
	//       | functiondef | prefixexp | tableconstructor | opunary exp
	//       | exp binop exp
	Expression interface{}
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

func (p *Parser) parseExpression() (Expression, error) {
	var (
		// Sets the priority of the operators (power - highest, or - lowest)
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
		// todo: parse hexadecimal numbers
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
	case lexer.TokenKeywordFunction:
		p.currentToken = p.lexer.NextToken()
		return p.parseFunctionDefinition()
	case lexer.TokenIdentifier, lexer.TokenLeftParen:
		return p.parsePrefixExpression()
	case lexer.TokenLeftBrace:
		return p.parseTableConstructor()
	default:
		return nil, errors.New("unexpected token: " + p.currentToken.Type.String())
	}
}

func (p *Parser) parseTableConstructor() (*TableConstructorExpression, error) {
	p.currentToken = p.lexer.NextToken()
	var fields []Field
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
	return &TableConstructorExpression{Fields: fields}, nil
}

// field ::= ‘[’ exp ‘]’ ‘=’ exp | Name ‘=’ exp | exp
func (p *Parser) parseField() (Field, error) {
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
		return &ExpToExpField{Key: key, Value: value}, nil
	case lexer.TokenIdentifier:
		name := p.currentToken.Value
		p.currentToken = p.lexer.NextToken()
		if p.currentToken.Type != lexer.TokenAssign {
			return &ExpressionField{Value: &NameVar{Name: name}}, nil
		}
		p.currentToken = p.lexer.NextToken()
		value, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		return &NameField{Name: name, Value: value}, nil
	default:
		exp, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		return &ExpressionField{Value: exp}, nil
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

// prefixexp ::= var | functioncall | ‘(’ exp ‘)’
func (p *Parser) parsePrefixExpression() (PrefixExpression, error) {
	return p.parseFunctionCall()
}

// functioncall ::= prefixexp args | prefixexp ‘:’ Name args
func (p *Parser) parseFunctionCall() (PrefixExpression, error) {
	prefixExp, err := p.parsePrefixExpressionBase()
	if err != nil {
		return nil, err
	}
	funcCall, err := p.parseFunctionCallPostfix(prefixExp)
	if errors.Is(err, ErrNoFunctionCallPostfix) {
		// If we didn't find a function call postfix, we return the prefix expression
		// as it is a valid prefix expression (e.g. a variable or a function definition).
		return prefixExp, nil
	}
	if err != nil {
		return nil, err
	}
	return funcCall, nil
}

// parseFunctionCallPostfix
// This function is called when we already have a prefix expression
// and we want to parse the function call postfix (args or ':' Name args)
func (p *Parser) parseFunctionCallPostfix(prefixExp PrefixExpression) (*FunctionCall, error) {
	if p.currentToken.Type == lexer.TokenLiteralString {
		str := p.currentToken.Value
		p.currentToken = p.lexer.NextToken()
		return &FunctionCall{
			PrefixExp: prefixExp,
			Name:      "",
			Args:      &LiteralString{Value: str},
		}, nil
	} else if p.currentToken.Type == lexer.TokenLeftBrace {
		args, err := p.parseTableConstructor()
		if err != nil {
			return nil, err
		}
		return &FunctionCall{
			PrefixExp: prefixExp,
			Name:      "",
			Args:      args,
		}, nil
	} else if p.currentToken.Type == lexer.TokenColon {
		p.currentToken = p.lexer.NextToken()
		if p.currentToken.Type != lexer.TokenIdentifier {
			return nil, errors.New("missing identifier after ':'")
		}
		name := p.currentToken.Value
		p.currentToken = p.lexer.NextToken()
		args, err := p.parseArgs()
		if err != nil {
			return nil, err
		}
		return &FunctionCall{
			PrefixExp: prefixExp,
			Name:      name,
			Args:      args,
		}, nil
	} else if p.currentToken.Type == lexer.TokenColon {
		p.currentToken = p.lexer.NextToken()
		if p.currentToken.Type != lexer.TokenIdentifier {
			return nil, errors.New("missing identifier after ':'")
		}
		name := p.currentToken.Value
		p.currentToken = p.lexer.NextToken()
		args, err := p.parseArgs()
		if err != nil {
			return nil, err
		}
		return &FunctionCall{
			PrefixExp: prefixExp,
			Name:      name,
			Args:      args,
		}, nil
	} else if p.currentToken.Type == lexer.TokenLeftParen {
		args, err := p.parseArgs()
		if err != nil {
			return nil, err
		}
		return &FunctionCall{
			PrefixExp: prefixExp,
			Name:      "",
			Args:      args,
		}, nil
	}
	return nil, ErrNoFunctionCallPostfix
}

// prefixexp ::= var | ‘(’ exp ‘)’
func (p *Parser) parsePrefixExpressionBase() (PrefixExpression, error) {
	switch p.currentToken.Type {
	case lexer.TokenIdentifier:
		return p.parseVar()
	case lexer.TokenLeftParen:
		p.currentToken = p.lexer.NextToken()
		exp, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		if p.currentToken.Type != lexer.TokenRightParen {
			return nil, errors.New("missing ')'")
		}
		p.currentToken = p.lexer.NextToken()
		return exp.(PrefixExpression), nil
	default:
		return nil, fmt.Errorf("unexpected token: %s", p.currentToken.Type)
	}
}

// args ::= ‘(’ [explist] ‘)’ | tableconstructor | LiteralString
func (p *Parser) parseArgs() (Args, error) {
	if p.currentToken.Type == lexer.TokenLeftParen {
		p.currentToken = p.lexer.NextToken()
		if p.currentToken.Type == lexer.TokenRightParen {
			p.currentToken = p.lexer.NextToken()
			return &ExpressionList{Expressions: nil}, nil
		}
		explist, err := p.parseExpressionList()
		if err != nil {
			return nil, err
		}
		if p.currentToken.Type != lexer.TokenRightParen {
			return nil, errors.New("missing ')'")
		}
		p.currentToken = p.lexer.NextToken()
		return explist, nil
	} else if p.currentToken.Type == lexer.TokenLeftBrace {
		return p.parseTableConstructor()
	} else if p.currentToken.Type == lexer.TokenLiteralString {
		str := p.currentToken.Value
		p.currentToken = p.lexer.NextToken()
		return &LiteralString{Value: str}, nil
	}
	return nil, fmt.Errorf("unexpected token: %s", p.currentToken.Type)
}
