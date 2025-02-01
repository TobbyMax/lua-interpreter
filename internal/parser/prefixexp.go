package parser

import (
	"errors"
	"fmt"

	"lua-interpreter/internal/ast"
	"lua-interpreter/internal/lexer"
)

var (
	functionCallPostfixTokens = []lexer.TokenType{
		lexer.TokenLeftParen,     // '(' [explist] ')'
		lexer.TokenLeftBrace,     // '{' fieldlist '}'
		lexer.TokenLiteralString, // LiteralString
		lexer.TokenColon,         // ':' Name args
	}
	variableAccessPostfixTokens = []lexer.TokenType{
		lexer.TokenLeftBracket, // '[' exp ']'
		lexer.TokenDot,         // '.' Name
	}
	prefixExpTokens = []lexer.TokenType{
		// function call
		lexer.TokenLeftParen,     // '(' exp ')'
		lexer.TokenLeftBrace,     // '{' fieldlist '}'
		lexer.TokenLiteralString, // LiteralString
		lexer.TokenColon,         // ':' Name args
		// variable access
		lexer.TokenDot,         // '.' Name
		lexer.TokenLeftBracket, // '[' exp ']'
	}
)

// prefixexp ::= var | functioncall | ‘(’ exp ‘)’
// var ::= Name | prefixexp ‘[’ exp ‘]’ | prefixexp ‘.’ Name
// functioncall ::= prefixexp args | prefixexp ‘:’ Name args
// prefixexp ::= Name
// prefixexp ::= Name [‘:’ Name] args | Name ‘[’ exp ‘]’ | Name ‘.’ Name | Name
//
//	| ‘(’ exp ‘)’
//	| prefixexp ‘[’ exp ‘]’
//	| prefixexp ‘.’ Name
//	| prefixexp [‘:’ Name] args
func (p *Parser) parsePrefixExpression() (ast.PrefixExpression, error) {
	prefix, err := p.parsePrefixExpressionHead()
	if err != nil {
		return nil, err
	}
	return p.parsePrefixExpressionTail(prefix)
}

func (p *Parser) parsePrefixExpressionTail(head ast.PrefixExpression) (prefix ast.PrefixExpression, err error) {
	prefix = head
	for isOneOfTypes(p.currentToken.Type, prefixExpTokens) {
		prefix, err = p.parsePrefixExpressionStep(prefix)
		if err != nil {
			return nil, err
		}
	}
	return prefix, nil
}

func (p *Parser) parsePrefixExpressionStep(prefix ast.PrefixExpression) (ast.PrefixExpression, error) {
	switch p.currentToken.Type {
	case lexer.TokenLeftBracket:
		p.currentToken = p.lexer.NextToken()
		exp, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		if p.currentToken.Type != lexer.TokenRightBracket {
			return nil, errors.New("missing ']'")
		}
		p.currentToken = p.lexer.NextToken()
		return &ast.IndexedVar{PrefixExp: prefix, Exp: exp}, nil
	case lexer.TokenDot:
		p.currentToken = p.lexer.NextToken()
		if p.currentToken.Type != lexer.TokenIdentifier {
			return nil, errors.New("missing identifier after '.'")
		}
		name := p.currentToken.Value
		p.currentToken = p.lexer.NextToken()
		return &ast.MemberVar{PrefixExp: prefix, Name: name}, nil
	case lexer.TokenLeftParen, lexer.TokenColon, lexer.TokenLiteralString, lexer.TokenLeftBrace:
		funcCall, err := p.parseFunctionCallPostfix(prefix)
		if err != nil {
			return nil, err
		}
		return funcCall, nil
	default:
		return prefix, nil
	}
}

// functioncall ::= prefixexp args | prefixexp ‘:’ Name args
// functioncall ::= Name args | prefixexp ‘[’ exp ‘]’ args |  prefixexp ‘.’ Name | prefixexp ‘:’ Name args
func (p *Parser) parseFunctionCall() (ast.PrefixExpression, error) {
	prefixExp, err := p.parsePrefixExpressionHead()
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
func (p *Parser) parseFunctionCallPostfix(prefixExp ast.PrefixExpression) (*ast.FunctionCall, error) {
	if p.currentToken.Type == lexer.TokenLiteralString {
		str := p.currentToken.Value
		p.currentToken = p.lexer.NextToken()
		return &ast.FunctionCall{
			PrefixExp: prefixExp,
			Name:      "",
			Args:      &ast.LiteralString{Value: str},
		}, nil
	} else if p.currentToken.Type == lexer.TokenLeftBrace {
		args, err := p.parseTableConstructor()
		if err != nil {
			return nil, err
		}
		return &ast.FunctionCall{
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
		return &ast.FunctionCall{
			PrefixExp: prefixExp,
			Name:      name,
			Args:      args,
		}, nil
	} else if p.currentToken.Type == lexer.TokenLeftParen {
		args, err := p.parseArgs()
		if err != nil {
			return nil, err
		}
		return &ast.FunctionCall{
			PrefixExp: prefixExp,
			Name:      "",
			Args:      args,
		}, nil
	}
	return nil, ErrNoFunctionCallPostfix
}

// prefixexp ::= var | ‘(’ exp ‘)’
func (p *Parser) parsePrefixExpressionHead() (ast.PrefixExpression, error) {
	switch p.currentToken.Type {
	case lexer.TokenIdentifier:
		name := p.currentToken.Value
		p.currentToken = p.lexer.NextToken()
		return &ast.NameVar{Name: name}, nil
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
		return exp.(ast.PrefixExpression), nil
	default:
		return nil, fmt.Errorf("unexpected token: %s", p.currentToken.Type)
	}
}

// args ::= ‘(’ [explist] ‘)’ | tableconstructor | LiteralString
func (p *Parser) parseArgs() (ast.Args, error) {
	if p.currentToken.Type == lexer.TokenLeftParen {
		p.currentToken = p.lexer.NextToken()
		if p.currentToken.Type == lexer.TokenRightParen {
			p.currentToken = p.lexer.NextToken()
			return &ast.ExpressionList{Expressions: nil}, nil
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
		return &ast.LiteralString{Value: str}, nil
	}
	return nil, fmt.Errorf("unexpected token: %s", p.currentToken.Type)
}

// varlist ‘=’ explist
// varlist ::= var { ',' var }
// functioncall
// functioncall ::= prefixexp args | prefixexp ‘:’ Name args
// prefixexp ::= var  // for this case prefixexp is a Var
// args ::= ‘(’ [explist] ‘)’ | tableconstructor | LiteralString
func (p *Parser) parseAssignmentOrFunctionCall() (ast.Statement, error) {
	name := p.currentToken.Value
	p.currentToken = p.lexer.NextToken()

	prefix, err := p.parsePrefixExpressionTail(&ast.NameVar{Name: name})
	if err != nil {
		return nil, err
	}
	switch prefix.(type) {
	case *ast.FunctionCall:
		return prefix, nil
	case *ast.NameVar, *ast.MemberVar, *ast.IndexedVar:
		vars := []ast.Var{prefix.(ast.Var)}
		if p.currentToken.Type == lexer.TokenComma {
			p.currentToken = p.lexer.NextToken()
			vl, err := p.parseVarList()
			if err != nil {
				return nil, err
			}
			vars = append(vars, vl...)
		}
		if p.currentToken.Type != lexer.TokenAssign {
			return nil, errors.New("missing '='")
		}
		p.currentToken = p.lexer.NextToken()
		exps, err := p.parseExpressionList()
		if err != nil {
			return nil, err
		}
		return &ast.Assignment{Vars: vars, Exps: exps}, nil
	default:
		return nil, fmt.Errorf("unexpected type: %T", prefix)
	}
}
