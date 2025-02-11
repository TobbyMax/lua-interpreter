package parser

import (
	"errors"
	"fmt"

	"lua-interpreter/internal/ast"
	"lua-interpreter/internal/lexer"
)

func (p *Parser) parseStatement() (ast.Statement, error) {
	switch p.currentToken.Type {
	case lexer.TokenSemiColon:
		p.currentToken = p.lexer.NextToken()
		return &ast.EmptyStatement{}, nil
	case lexer.TokenKeywordBreak:
		p.currentToken = p.lexer.NextToken()
		return &ast.Break{}, nil
	case lexer.TokenKeywordGoTo:
		p.currentToken = p.lexer.NextToken()
		name := p.currentToken.Value
		p.currentToken = p.lexer.NextToken()
		return &ast.Goto{Name: name}, nil
	case lexer.TokenKeywordDo:
		return p.parseDoStatement()
	case lexer.TokenKeywordWhile:
		return p.parseWhileStatement()
	case lexer.TokenKeywordRepeat:
		return p.parseRepeatStatement()
	case lexer.TokenKeywordIf:
		return p.parseIfStatement()
	case lexer.TokenKeywordFor:
		return p.parseForStatement()
	case lexer.TokenKeywordFunction:
		return p.parseFunction()
	case lexer.TokenKeywordLocal:
		return p.parseLocalDeclaration()
	case lexer.TokenDoubleColon:
		return p.parseLabel()
	case lexer.TokenLeftParen:
		return p.parseFunctionCall()
	case lexer.TokenIdentifier:
		return p.parseAssignmentOrFunctionCall()
	default:
		return nil, fmt.Errorf("unexpected token: %s", p.currentToken.Type)
	}
}

func (p *Parser) parseNameList() ([]string, error) {
	var names []string
	if p.currentToken.Type != lexer.TokenIdentifier {
		return nil, errors.New("missing identifier")
	}
	names = append(names, p.currentToken.Value)
	p.currentToken = p.lexer.NextToken()
	for p.currentToken.Type == lexer.TokenComma {
		p.currentToken = p.lexer.NextToken()
		if p.currentToken.Type != lexer.TokenIdentifier {
			return nil, errors.New("missing identifier")
		}
		names = append(names, p.currentToken.Value)
		p.currentToken = p.lexer.NextToken()
	}
	return names, nil
}

// varlist ::= var { ',' var }
func (p *Parser) parseVarList() ([]ast.Var, error) {
	var vars []ast.Var
	if p.currentToken.Type != lexer.TokenIdentifier {
		return nil, errors.New("missing identifier")
	}
	name := p.currentToken.Value
	p.currentToken = p.lexer.NextToken()
	vars = append(vars, &ast.NameVar{Name: name})
	for p.currentToken.Type == lexer.TokenComma {
		p.currentToken = p.lexer.NextToken()
		if p.currentToken.Type != lexer.TokenIdentifier {
			return nil, errors.New("missing identifier")
		}
		name = p.currentToken.Value
		p.currentToken = p.lexer.NextToken()
		vars = append(vars, &ast.NameVar{Name: name})
	}
	return vars, nil
}

func (p *Parser) parseLabel() (*ast.Label, error) {
	p.currentToken = p.lexer.NextToken()
	if p.currentToken.Type != lexer.TokenIdentifier {
		return nil, errors.New("missing identifier")
	}
	name := p.currentToken.Value
	p.currentToken = p.lexer.NextToken()
	if p.currentToken.Type != lexer.TokenDoubleColon {
		return nil, errors.New("missing '::'")
	}
	p.currentToken = p.lexer.NextToken()
	return &ast.Label{Name: name}, nil
}

func (p *Parser) parseLocalDeclaration() (ast.Statement, error) {
	p.currentToken = p.lexer.NextToken()
	switch p.currentToken.Type {
	case lexer.TokenKeywordFunction:
		p.currentToken = p.lexer.NextToken()
		if p.currentToken.Type != lexer.TokenIdentifier {
			return nil, errors.New("missing Name for local function")
		}
		name := p.currentToken.Value
		p.currentToken = p.lexer.NextToken()
		body, err := p.parseFunctionBody()
		if err != nil {
			return nil, err
		}
		return &ast.LocalFunction{Name: name, FunctionBody: body}, nil
	case lexer.TokenIdentifier:
		names, err := p.parseNameList()
		if err != nil {
			return nil, err
		}
		if p.currentToken.Type != lexer.TokenAssign {
			return &ast.LocalVarDeclaration{Vars: names}, nil
		}
		p.currentToken = p.lexer.NextToken()
		exps, err := p.parseExpressionList()
		if err != nil {
			return nil, err
		}
		return &ast.LocalVarDeclaration{Vars: names, Exps: exps}, nil
	default:
		return nil, errors.New("missing identifier or function")
	}
}

func (p *Parser) parseDoStatement() (*ast.Do, error) {
	p.currentToken = p.lexer.NextToken()
	block, err := p.parseBlock()
	if err != nil {
		return nil, err
	}
	if p.currentToken.Type != lexer.TokenKeywordEnd {
		return nil, errors.New("missing 'end' keyword")
	}
	p.currentToken = p.lexer.NextToken()
	return &ast.Do{Block: block}, nil
}
