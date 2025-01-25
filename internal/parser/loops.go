package parser

import (
	"errors"

	"lua-interpreter/internal/ast"
	"lua-interpreter/internal/lexer"
)

func (p *Parser) parseWhileStatement() (*ast.While, error) {
	p.currentToken = p.lexer.NextToken()
	exp, err := p.parseExpression()
	if err != nil {
		return nil, err
	}
	if p.currentToken.Type != lexer.TokenKeywordDo {
		return nil, errors.New("missing 'do' keyword")
	}
	p.currentToken = p.lexer.NextToken()
	block, err := p.parseBlock()
	if err != nil {
		return nil, err
	}
	if p.currentToken.Type != lexer.TokenKeywordEnd {
		return nil, errors.New("missing 'end' keyword")
	}
	p.currentToken = p.lexer.NextToken()
	return &ast.While{Exp: exp, Block: block}, nil
}

func (p *Parser) parseRepeatStatement() (*ast.Repeat, error) {
	p.currentToken = p.lexer.NextToken()
	block, err := p.parseBlock()
	if err != nil {
		return nil, err
	}
	if p.currentToken.Type != lexer.TokenKeywordUntil {
		return nil, errors.New("missing 'until' keyword")
	}
	p.currentToken = p.lexer.NextToken()
	exp, err := p.parseExpression()
	if err != nil {
		return nil, err
	}
	return &ast.Repeat{Block: block, Exp: exp}, nil
}

func (p *Parser) parseForStatement() (ast.Statement, error) {
	p.currentToken = p.lexer.NextToken()
	if p.currentToken.Type != lexer.TokenIdentifier {
		return nil, errors.New("missing identifier")
	}
	name := p.currentToken.Value
	p.currentToken = p.lexer.NextToken()
	names := []string{name}
	switch p.currentToken.Type {
	case lexer.TokenAssign:
		return p.parseForStatementWithName(name)
	case lexer.TokenComma:
		p.currentToken = p.lexer.NextToken()
		nl, err := p.parseNameList()
		if err != nil {
			return nil, err
		}
		names = append(names, nl...)
		fallthrough
	case lexer.TokenKeywordIn:
		return p.parseForInWithNames(names)
	default:
		return nil, errors.New("expected '=' or 'in' keyword")
	}
}

func (p *Parser) parseForStatementWithName(name string) (*ast.For, error) {
	if p.currentToken.Type != lexer.TokenAssign {
		return nil, errors.New("missing '='")
	}
	p.currentToken = p.lexer.NextToken()
	init, err := p.parseExpression()
	if err != nil {
		return nil, err
	}
	if p.currentToken.Type != lexer.TokenComma {
		return nil, errors.New("missing ','")
	}
	p.currentToken = p.lexer.NextToken()
	limit, err := p.parseExpression()
	if err != nil {
		return nil, err
	}
	var step *ast.Expression
	if p.currentToken.Type == lexer.TokenComma {
		p.currentToken = p.lexer.NextToken()
		stepExp, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		step = &stepExp
	}
	if p.currentToken.Type != lexer.TokenKeywordDo {
		return nil, errors.New("missing 'do' keyword")
	}
	p.currentToken = p.lexer.NextToken()
	block, err := p.parseBlock()
	if err != nil {
		return nil, err
	}
	if p.currentToken.Type != lexer.TokenKeywordEnd {
		return nil, errors.New("missing 'end' keyword")
	}
	p.currentToken = p.lexer.NextToken()
	return &ast.For{Name: name, Init: init, Limit: limit, Step: step, Block: block}, nil
}

func (p *Parser) parseForInWithNames(names []string) (*ast.ForIn, error) {
	if p.currentToken.Type != lexer.TokenKeywordIn {
		return nil, errors.New("missing 'in' keyword")
	}
	p.currentToken = p.lexer.NextToken()
	exps, err := p.parseExpressionList()
	if err != nil {
		return nil, err
	}
	if p.currentToken.Type != lexer.TokenKeywordDo {
		return nil, errors.New("missing 'do' keyword")
	}
	p.currentToken = p.lexer.NextToken()
	block, err := p.parseBlock()
	if err != nil {
		return nil, err
	}
	if p.currentToken.Type != lexer.TokenKeywordEnd {
		return nil, errors.New("missing 'end' keyword")
	}
	p.currentToken = p.lexer.NextToken()
	return &ast.ForIn{Names: names, Exps: exps, Block: block}, nil
}
