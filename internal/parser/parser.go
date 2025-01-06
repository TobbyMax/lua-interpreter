package parser

import (
	"lua-interpreter/internal/lexer"
)

type (
	StatementType int

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
	// FunctionDefExpression
	// functiondef ::= function funcbody
	// funcbody ::= ‘(’ [parlist] ‘)’ block end
	FunctionDefExpression struct {
		ParameterList ParameterList
		Block         Block
	}

	// ExpressionList
	// explist ::= exp {‘,’ exp}
	ExpressionList struct {
		Expressions []Expression
	}
	// Args [ExpressionList | TableConstructorExpression | LiteralString]
	// args ::= ‘(’ [explist] ‘)’ | tableconstructor | LiteralString
	Args interface{}
	Arg  interface{}
	// FunctionCall
	// functioncall ::= prefixexp args | prefixexp ‘:’ Name args
	FunctionCall struct {
		PrefixExp PrefixExpression
		Name      string
		Args      Args
	}
	// PrefixExpression
	// prefixexp ::= var | functioncall | ‘(’ exp ‘)’
	PrefixExpression interface{}
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

	NameVar struct {
		Name string
	}
	IndexedVar struct {
		PrefixExp PrefixExpression
		Exp       Expression
	}
	MemberVar struct {
		PrefixExp PrefixExpression
		Name      string
	}
	// Var
	// var ::=  Name | prefixexp ‘[’ exp ‘]’ | prefixexp ‘.’ Name
	Var interface{}
	// Statement
	// stat ::=  ‘;’
	//       |  varlist ‘=’ explist
	//       |  functioncall
	//       |  label
	//       |  break
	//       |  goto Name
	//       |  do block end
	//       |  while exp do block end
	//       |  repeat block until exp
	//       |  if exp then block {elseif exp then block} [else block] end
	//       |  for Name ‘=’ exp ‘,’ exp [‘,’ exp] do block end
	//       |  for namelist in explist do block end
	//       |  function funcname funcbody
	//       |  local function Name funcbody
	//       |  local namelist [‘=’ explist]
	Statement interface{}
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

// stat ::=  ‘;’
//       |  varlist ‘=’ explist
//       |  functioncall
//       |  label
//       |  break
//       |  goto Name
//       |  do block end
//       |  while exp do block end
//       |  repeat block until exp
//       |  if exp then block {elseif exp then block} [else block] end
//       |  for Name ‘=’ exp ‘,’ exp [‘,’ exp] do block end
//       |  for namelist in explist do block end
//       |  function funcname funcbody
//       |  local function Name funcbody
//       |  local namelist [‘=’ explist]
//var parseFunctionMap = map[lexer.TokenType]func(p *Parser) Statement{
//	lexer.TokenSemiColon: parseEmptyStatement,
//	lexer.
//}

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
	return p.parseBlock()
}

func (p *Parser) parseBlock() (b Block, err error) {
	b.Statements, err = p.parseStatements()
	if err != nil {
		return Block{}, err
	}
	b.ReturnStatement = p.parseReturnStatement()
	return b, nil
}

func (p *Parser) parseStatements() ([]Statement, error) {
	var statements []Statement
	p.currentToken = p.lexer.NextToken()
	for p.currentToken.Type != lexer.TokenEOF && p.currentToken.Type != lexer.TokenKeywordReturn {
		stat, err := p.parseStatement()
		if err != nil {
			return nil, err
		}
		statements = append(statements, stat)
	}
	return statements, nil
}

func (p *Parser) parseStatement() (Statement, error) {
	return p.parseExpression()
}

func (p *Parser) parseReturnStatement() *ReturnStatement {
	if p.currentToken.Type != lexer.TokenKeywordReturn {
		return nil
	}
	p.currentToken = p.lexer.NextToken()
	expressions, err := p.parseExpressionList()
	if err != nil {
		return nil
	}
	return &ReturnStatement{
		Expressions: expressions,
	}
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
