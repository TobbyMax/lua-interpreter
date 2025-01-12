package parser

import (
	"lua-interpreter/internal/lexer"
)

type (
	NilExpression     struct{}
	BooleanExpression struct {
		Value bool
	}
	NumeralExpression struct {
		Value float64
	}
	LiteralString struct {
		Value string
	}
	VarArgExpression struct{}
	// ParameterList
	// parlist ::= namelist [‘,’ ‘...’] | ‘...’
	ParameterList struct {
		Names    []string
		IsVarArg bool
	}
	// ReturnStatement ::= return [explist] [‘;’]
	// retstat ::= return [explist] [‘;’]
	ReturnStatement struct {
		Expressions []Expression
	}
	// Block
	// block ::= { stat } [ retstat ]
	Block struct {
		Statements      []Statement
		ReturnStatement *ReturnStatement
	}
)

type Parser struct {
	lexer        *lexer.Lexer
	currentToken lexer.Token
}

func NewParser(lexer *lexer.Lexer) *Parser {
	return &Parser{
		lexer: lexer,
	}
}

func (p *Parser) Parse() (Block, error) {
	p.currentToken = p.lexer.NextToken()
	return p.parseBlock()
}

func (p *Parser) parseBlock() (b Block, err error) {
	b.Statements, err = p.parseStatements()
	if err != nil {
		return Block{}, err
	}
	b.ReturnStatement, err = p.parseReturnStatement()
	if err != nil {
		return Block{}, err
	}
	return b, nil
}

func (p *Parser) parseStatements() ([]Statement, error) {
	var statements []Statement
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

func (p *Parser) parseReturnStatement() (*ReturnStatement, error) {
	if p.currentToken.Type != lexer.TokenKeywordReturn {
		return nil, nil
	}
	p.currentToken = p.lexer.NextToken()
	expressions, err := p.parseExpressionList()
	if err != nil {
		return nil, err
	}
	return &ReturnStatement{
		Expressions: expressions,
	}, nil
}

func (p *Parser) parseExpressionList() (exps []Expression, err error) {
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
