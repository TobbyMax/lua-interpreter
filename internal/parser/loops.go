package parser

import (
	"errors"

	"lua-interpreter/internal/lexer"
)

type (
	While struct {
		Exp   Expression
		Block Block
	}
	Repeat struct {
		Block Block
		Exp   Expression
	}
	For struct {
		Name  string
		Init  Expression
		Limit Expression
		Step  *Expression
		Block Block
	}
	ForIn struct {
		Names []string
		Exps  []Expression
		Block Block
	}
)

func (p *Parser) parseWhileStatement() (*While, error) {
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
	return &While{Exp: exp, Block: block}, nil
}

func (p *Parser) parseRepeatStatement() (*Repeat, error) {
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
	return &Repeat{Block: block, Exp: exp}, nil
}

func (p *Parser) parseForStatement() (*For, error) {
	p.currentToken = p.lexer.NextToken()
	if p.currentToken.Type != lexer.TokenIdentifier {
		return nil, errors.New("missing identifier")
	}
	name := p.currentToken.Value
	// TODO: add for-in support
	p.currentToken = p.lexer.NextToken()
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
	var step *Expression
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
	return &For{Name: name, Init: init, Limit: limit, Step: step, Block: block}, nil
}
