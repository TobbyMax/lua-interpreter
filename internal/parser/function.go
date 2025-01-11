package parser

import (
	"errors"

	"lua-interpreter/internal/lexer"
)

type (
	// FunctionBody
	// funcbody ::= ‘(’ [parlist] ‘)’ block end
	FunctionBody struct {
		ParameterList ParameterList
		Block         Block
	}
	// FunctionName
	// funcname ::= Name {‘.’ Name} [‘:’ Name]
	FunctionName struct {
		FirstName string
		Names     []string
		LastName  string
	}
	// Function
	// function funcname funcbody
	Function struct {
		FunctionName FunctionName
		FuncBody     FunctionBody
	}
)

func (p *Parser) parseFunction() (*Function, error) {
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
	return &Function{FunctionName: name, FuncBody: body}, nil
}

func (p *Parser) parseFunctionName() (FunctionName, error) {
	if p.currentToken.Type != lexer.TokenIdentifier {
		return FunctionName{}, errors.New("missing identifier")
	}
	name := p.currentToken.Value
	p.currentToken = p.lexer.NextToken()
	var names []string
	for p.currentToken.Type == lexer.TokenDot {
		p.currentToken = p.lexer.NextToken()
		if p.currentToken.Type != lexer.TokenIdentifier {
			return FunctionName{}, errors.New("missing identifier")
		}
		names = append(names, p.currentToken.Value)
		p.currentToken = p.lexer.NextToken()
	}
	var lastName string
	if p.currentToken.Type == lexer.TokenColon {
		p.currentToken = p.lexer.NextToken()
		if p.currentToken.Type != lexer.TokenIdentifier {
			return FunctionName{}, errors.New("missing identifier")
		}
		lastName = p.currentToken.Value
		p.currentToken = p.lexer.NextToken()
	}
	return FunctionName{FirstName: name, Names: names, LastName: lastName}, nil
}

func (p *Parser) parseFunctionBody() (FunctionBody, error) {
	if p.currentToken.Type != lexer.TokenLeftParen {
		return FunctionBody{}, errors.New("missing '('")
	}
	p.currentToken = p.lexer.NextToken()
	parList, err := p.parseParameterList()
	if err != nil {
		return FunctionBody{}, err
	}
	if p.currentToken.Type != lexer.TokenRightParen {
		return FunctionBody{}, errors.New("missing ')'")
	}
	p.currentToken = p.lexer.NextToken()
	block, err := p.parseBlock()
	if err != nil {
		return FunctionBody{}, err
	}
	if p.currentToken.Type != lexer.TokenKeywordEnd {
		return FunctionBody{}, errors.New("missing 'end' keyword")
	}
	p.currentToken = p.lexer.NextToken()
	return FunctionBody{ParameterList: parList, Block: block}, nil
}
