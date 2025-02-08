package parser

import (
	"errors"

	"lua-interpreter/internal/ast"
	"lua-interpreter/internal/lexer"
)

func (p *Parser) parseFunction() (*ast.Function, error) {
	p.currentToken = p.lexer.NextToken()
	if p.currentToken.Type != lexer.TokenIdentifier {
		return nil, errors.New("missing identifier")
	}
	name, err := p.parseFunctionName()
	if err != nil {
		return nil, err
	}
	body, err := p.parseFunctionBody()
	if err != nil {
		return nil, err
	}
	return &ast.Function{FunctionName: name, FuncBody: body}, nil
}

func (p *Parser) parseFunctionName() (ast.FunctionName, error) {
	if p.currentToken.Type != lexer.TokenIdentifier {
		return ast.FunctionName{}, errors.New("missing identifier")
	}
	isMethod := false
	lastName := p.currentToken.Value
	p.currentToken = p.lexer.NextToken()
	var names []string
	for p.currentToken.Type == lexer.TokenDot {
		names = append(names, lastName)
		p.currentToken = p.lexer.NextToken()
		if p.currentToken.Type != lexer.TokenIdentifier {
			return ast.FunctionName{}, errors.New("missing identifier")
		}
		lastName = p.currentToken.Value
		p.currentToken = p.lexer.NextToken()
	}
	if p.currentToken.Type == lexer.TokenColon {
		isMethod = true
		names = append(names, lastName)
		p.currentToken = p.lexer.NextToken()
		if p.currentToken.Type != lexer.TokenIdentifier {
			return ast.FunctionName{}, errors.New("missing identifier")
		}
		lastName = p.currentToken.Value
		p.currentToken = p.lexer.NextToken()
	}
	return ast.FunctionName{PrefixNames: names, Name: lastName, IsMethod: isMethod}, nil
}

func (p *Parser) parseFunctionBody() (ast.FunctionBody, error) {
	if p.currentToken.Type != lexer.TokenLeftParen {
		return ast.FunctionBody{}, errors.New("missing '('")
	}
	p.currentToken = p.lexer.NextToken()
	parList, err := p.parseParameterList()
	if err != nil {
		return ast.FunctionBody{}, err
	}
	if p.currentToken.Type != lexer.TokenRightParen {
		return ast.FunctionBody{}, errors.New("missing ')'")
	}
	p.currentToken = p.lexer.NextToken()
	block, err := p.parseBlock()
	if err != nil {
		return ast.FunctionBody{}, err
	}
	if p.currentToken.Type != lexer.TokenKeywordEnd {
		return ast.FunctionBody{}, errors.New("missing 'end' keyword")
	}
	p.currentToken = p.lexer.NextToken()
	return ast.FunctionBody{ParameterList: parList, Block: block}, nil
}

func (p *Parser) parseFunctionDefinition() (*ast.FunctionDefinition, error) {
	if p.currentToken.Type != lexer.TokenKeywordFunction {
		return nil, errors.New("missing 'function' keyword")
	}
	p.currentToken = p.lexer.NextToken()
	body, err := p.parseFunctionBody()
	if err != nil {
		return nil, err
	}
	return &ast.FunctionDefinition{FunctionBody: body}, nil
}

// parlist ::= namelist [‘,’ ‘...’] | ‘...’
func (p *Parser) parseParameterList() (ast.ParameterList, error) {
	var names []string
	var isVararg bool
	if p.currentToken.Type == lexer.TokenTripleDot {
		isVararg = true
		p.currentToken = p.lexer.NextToken()
	} else {
		for p.currentToken.Type == lexer.TokenIdentifier {
			names = append(names, p.currentToken.Value)
			p.currentToken = p.lexer.NextToken()
			if p.currentToken.Type == lexer.TokenComma {
				p.currentToken = p.lexer.NextToken()
			} else {
				break
			}
		}
		if p.currentToken.Type == lexer.TokenTripleDot {
			isVararg = true
			p.currentToken = p.lexer.NextToken()
		}
	}
	return ast.ParameterList{Names: names, IsVarArg: isVararg}, nil
}
