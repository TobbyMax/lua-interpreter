//go:generate mockgen -destination=./mock/parser_mock.go -package=mock lua-interpreter/internal/parser scanner
package parser

import (
	"lua-interpreter/internal/ast"
	"lua-interpreter/internal/lexer"
)

type (
	scanner interface {
		NextToken() lexer.Token
	}
	IParser interface {
		Parse() (ast.Block, error)
	}
	Parser struct {
		lexer        scanner
		currentToken lexer.Token
	}
)

func New(lexer scanner) *Parser {
	return &Parser{
		lexer: lexer,
	}
}

func (p *Parser) Parse() (ast.Block, error) {
	p.currentToken = p.lexer.NextToken()
	return p.parseBlock()
}

func (p *Parser) parseBlock() (b ast.Block, err error) {
	b.Statements, err = p.parseStatements()
	if err != nil {
		return ast.Block{}, err
	}
	b.ReturnStatement, err = p.parseReturnStatement()
	if err != nil {
		return ast.Block{}, err
	}
	return b, nil
}

func (p *Parser) parseStatements() ([]ast.Statement, error) {
	var statements []ast.Statement
	for p.currentToken.Type != lexer.TokenEOF && p.currentToken.Type != lexer.TokenKeywordReturn &&
		p.currentToken.Type != lexer.TokenKeywordEnd && p.currentToken.Type != lexer.TokenKeywordElse &&
		p.currentToken.Type != lexer.TokenKeywordElseIf && p.currentToken.Type != lexer.TokenKeywordUntil {
		stat, err := p.parseStatement()
		if err != nil {
			return nil, err
		}
		statements = append(statements, stat)
	}
	return statements, nil
}

func (p *Parser) parseReturnStatement() (*ast.ReturnStatement, error) {
	if p.currentToken.Type != lexer.TokenKeywordReturn {
		return nil, nil
	}
	p.currentToken = p.lexer.NextToken()
	expressions, err := p.parseExpressionList()
	if err != nil {
		return nil, err
	}
	return &ast.ReturnStatement{
		Expressions: expressions,
	}, nil
}

func (p *Parser) parseExpressionList() (exps []ast.Expression, err error) {
	exp, err := p.parseExpression()
	if err != nil {
		return nil, err
	}
	if exp == nil {
		return nil, nil
	}
	exps = append(exps, exp)
	for p.currentToken.Type == lexer.TokenComma {
		p.currentToken = p.lexer.NextToken()
		exp, err = p.parseExpression()
		if err != nil {
			return nil, err
		}
		exps = append(exps, exp)
	}
	return exps, nil
}
