package parser

import (
	"errors"

	"lua-interpreter/internal/ast"
	"lua-interpreter/internal/lexer"
)

func (p *Parser) parseIfStatement() (*ast.If, error) {
	p.currentToken = p.lexer.NextToken()
	var (
		exps   []ast.Expression
		blocks []ast.Block
	)
	for {
		exp, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		exps = append(exps, exp)
		if p.currentToken.Type != lexer.TokenKeywordThen {
			return nil, errors.New("missing 'then' keyword")
		}
		p.currentToken = p.lexer.NextToken()
		block, err := p.parseBlock()
		if err != nil {
			return nil, err
		}
		blocks = append(blocks, block)
		if p.currentToken.Type == lexer.TokenKeywordElseIf {
			p.currentToken = p.lexer.NextToken()
			continue
		} else if p.currentToken.Type == lexer.TokenKeywordElse {
			p.currentToken = p.lexer.NextToken()
			block, err = p.parseBlock()
			if err != nil {
				return nil, err
			}
			blocks = append(blocks, block)
			if p.currentToken.Type != lexer.TokenKeywordEnd {
				return nil, errors.New("missing 'end' keyword")
			}
			break
		} else if p.currentToken.Type == lexer.TokenKeywordEnd {
			p.currentToken = p.lexer.NextToken()
			break
		} else {
			return nil, errors.New("missing 'elseif', 'else' or 'end' keyword")
		}
	}
	return &ast.If{Exps: exps, Blocks: blocks}, nil
}
