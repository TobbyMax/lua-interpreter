package parser

import (
	"errors"
	"fmt"

	"lua-interpreter/internal/lexer"
)

type (
	EmptyStatement struct{}
	NameVar        struct {
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
	Var                 interface{}
	LocalVarDeclaration struct {
		Vars []string
		Exps []Expression
	}
	Assignment struct {
		Vars []Var
		Exps []Expression
	}
	Label struct {
		Name string
	}
	Break struct{}
	Goto  struct {
		Name string
	}
	Do struct {
		Block Block
	}
	// LocalFunction
	// local function Name funcbody
	LocalFunction struct {
		Name         string
		FunctionBody FunctionBody
	}
	// Statement [LocalVarDeclaration | FunctionCall | Label | Break | Goto | Do | While | Repeat | If | ForNum | ForIn | FunctionDefinition | LocalFunctionDefExpression | LocalAssignment]
	// stat ::=  ‘;’
	//	|  varlist ‘=’ explist
	//	|  functioncall
	//	|  label
	//	|  break
	//	|  goto Name
	//	|  do block end
	//	|  while exp do block end
	//	|  repeat block until exp
	//	|  if exp then block {elseif exp then block} [else block] end
	//	|  for Name ‘=’ exp ‘,’ exp [‘,’ exp] do block end
	//	|  for namelist in explist do block end
	//	|  function funcname funcbody
	//	|  local function Name funcbody
	//	|  local namelist [‘=’ explist]
	Statement interface{}
)

func (p *Parser) parseStatement() (Statement, error) {
	switch p.currentToken.Type {
	case lexer.TokenSemiColon:
		p.currentToken = p.lexer.NextToken()
		return &EmptyStatement{}, nil
	case lexer.TokenKeywordBreak:
		p.currentToken = p.lexer.NextToken()
		return &Break{}, nil
	case lexer.TokenKeywordGoTo:
		p.currentToken = p.lexer.NextToken()
		name := p.currentToken.Value
		p.currentToken = p.lexer.NextToken()
		return &Goto{Name: name}, nil
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
func (p *Parser) parseVarList() ([]Var, error) {
	var vars []Var
	if p.currentToken.Type != lexer.TokenIdentifier {
		return nil, errors.New("missing identifier")
	}
	name := p.currentToken.Value
	p.currentToken = p.lexer.NextToken()
	vars = append(vars, NameVar{Name: name})
	for p.currentToken.Type == lexer.TokenComma {
		p.currentToken = p.lexer.NextToken()
		if p.currentToken.Type != lexer.TokenIdentifier {
			return nil, errors.New("missing identifier")
		}
		name = p.currentToken.Value
		p.currentToken = p.lexer.NextToken()
		vars = append(vars, NameVar{Name: name})
	}
	return vars, nil
}

// var ::= Name | prefixexp ‘[’ exp ‘]’ | prefixexp ‘.’ Name
// prefixexp ::= var | functioncall | ‘(’ exp ‘)’
// functioncall ::= prefixexp args | prefixexp ‘:’ Name args
// var ::= Name | prefixexp ‘[’ exp ‘]’ | prefixexp args ‘[’ exp ‘]’ | prefixexp ‘:’ Name args ‘[’ exp ‘]’
//
//	| prefixexp ‘.’ Name | prefixexp args ‘.’ Name | prefixexp ‘:’ Name args ‘.’ Name
func (p *Parser) parseVar() (Var, error) {
	name := p.currentToken.Value
	p.currentToken = p.lexer.NextToken()

	v, err := p.parsePrefixExpressionTail(&NameVar{Name: name})
	if err != nil {
		return nil, err
	}
	switch v.(type) {
	case *NameVar, *IndexedVar, *MemberVar:
		return v, nil
	default:
		return nil, fmt.Errorf("unexpected var type: %T", v)
	}
}

func (p *Parser) parseVarPostfix(prefixExp PrefixExpression) (Var, error) {
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
		return &IndexedVar{PrefixExp: prefixExp, Exp: exp}, nil
	case lexer.TokenDot:
		p.currentToken = p.lexer.NextToken()
		if p.currentToken.Type != lexer.TokenIdentifier {
			return nil, errors.New("missing identifier after '.'")
		}
		name := p.currentToken.Value
		p.currentToken = p.lexer.NextToken()
		return &MemberVar{PrefixExp: prefixExp, Name: name}, nil
	default:
		return prefixExp, nil
	}
}

func (p *Parser) parseLabel() (*Label, error) {
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
	return &Label{Name: name}, nil
}

func (p *Parser) parseLocalDeclaration() (Statement, error) {
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
		return &LocalFunction{Name: name, FunctionBody: body}, nil
	case lexer.TokenIdentifier:
		names, err := p.parseNameList()
		if err != nil {
			return nil, err
		}
		if p.currentToken.Type != lexer.TokenAssign {
			return &LocalVarDeclaration{Vars: names}, nil
		}
		p.currentToken = p.lexer.NextToken()
		exps, err := p.parseExpressionList()
		if err != nil {
			return nil, err
		}
		return &LocalVarDeclaration{Vars: names, Exps: exps}, nil
	default:
		return nil, errors.New("missing identifier or function")
	}
}

func (p *Parser) parseDoStatement() (*Do, error) {
	p.currentToken = p.lexer.NextToken()
	block, err := p.parseBlock()
	if err != nil {
		return nil, err
	}
	if p.currentToken.Type != lexer.TokenKeywordEnd {
		return nil, errors.New("missing 'end' keyword")
	}
	p.currentToken = p.lexer.NextToken()
	return &Do{Block: block}, nil
}
