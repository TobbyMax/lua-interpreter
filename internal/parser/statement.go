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
	While struct {
		Exp   Expression
		Block Block
	}
	Repeat struct {
		Block Block
		Exp   Expression
	}
	If struct {
		Exps   []Expression
		Blocks []Block
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
	// FunctionBody
	// funcbody ::= ‘(’ [parlist] ‘)’ block end
	FunctionBody struct {
		ParameterList ParameterList
		Block         Block
	}
	// LocalFunctionDefinition
	// functiondef ::= function funcbody
	LocalFunctionDefinition struct {
		FunctionBody FunctionBody
	}
	// FunctionName
	// funcname ::= Name {‘.’ Name} [‘:’ Name]
	FunctionName struct {
		FirstName string
		Names     []string
		LastName  string
	}
	// FunctionDefinition
	// function funcname funcbody
	FunctionDefinition struct {
		FunctionName FunctionName
		FuncBody     FunctionBody
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
		name := p.currentToken.Value
		p.currentToken = p.lexer.NextToken()
		return &Goto{Name: name}, nil
	case lexer.TokenKeywordDo:
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
	case lexer.TokenKeywordWhile:
		return p.parseWhileStatement()
	case lexer.TokenKeywordRepeat:
		return p.parseRepeatStatement()
	case lexer.TokenKeywordIf:
		p.currentToken = p.lexer.NextToken()
		var (
			exps   []Expression
			blocks []Block
		)
		for {
			exp, err := p.parseExpression()
			if err != nil {
				return nil, err
			}
			if exp == nil {
				return nil, errors.New("missing expression")
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
				break
			} else if p.currentToken.Type == lexer.TokenKeywordEnd {
				p.currentToken = p.lexer.NextToken()
				break
			} else {
				return nil, errors.New("missing 'elseif', 'else' or 'end' keyword")
			}
		}
		return &If{Exps: exps, Blocks: blocks}, nil
	case lexer.TokenKeywordFor:
		p.currentToken = p.lexer.NextToken()
		if p.currentToken.Type != lexer.TokenIdentifier {
			return nil, errors.New("missing identifier")
		}
		name := p.currentToken.Value
		// TODO: add for-in support
		p.currentToken = p.lexer.NextToken()
		if p.currentToken.Type != lexer.TokenEqual {
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
	case lexer.TokenKeywordFunction:
		return p.parseFunctionDefinition()
	case lexer.TokenKeywordLocal:
		p.currentToken = p.lexer.NextToken()
		switch p.currentToken.Type {
		case lexer.TokenKeywordFunction:
			p.currentToken = p.lexer.NextToken()
			body, err := p.parseFunctionBody()
			if err != nil {
				return nil, err
			}
			return &LocalFunctionDefinition{FunctionBody: body}, nil
		case lexer.TokenIdentifier:
			names, err := p.parseNameList()
			if err != nil {
				return nil, err
			}
			if p.currentToken.Type != lexer.TokenEqual {
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
	case lexer.TokenDoubleColon:
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
	case lexer.TokenIdentifier:
		vars, err := p.parseVarList()
		if err != nil {
			return nil, err
		}
		if p.currentToken.Type != lexer.TokenEqual {
			return nil, errors.New("missing '='")
		}
		p.currentToken = p.lexer.NextToken()
		exps, err := p.parseExpressionList()
		if err != nil {
			return nil, err
		}
		return &Assignment{Vars: vars, Exps: exps}, nil
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
func (p *Parser) parseVar() (Var, error) {
	switch p.currentToken.Type {
	case lexer.TokenIdentifier:
		name := p.currentToken.Value
		p.currentToken = p.lexer.NextToken()
		return NameVar{Name: name}, nil
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
		prefixExp, err := p.parsePrefixExpression()
		if err != nil {
			return nil, err
		}
		return IndexedVar{PrefixExp: prefixExp.(PrefixExpression), Exp: exp}, nil
	case lexer.TokenDot:
		p.currentToken = p.lexer.NextToken()
		if p.currentToken.Type != lexer.TokenIdentifier {
			return nil, errors.New("missing identifier")
		}
		name := p.currentToken.Value
		p.currentToken = p.lexer.NextToken()
		prefixExp, err := p.parsePrefixExpression()
		if err != nil {
			return nil, err
		}
		return MemberVar{PrefixExp: prefixExp.(PrefixExpression), Name: name}, nil
	default:
		return nil, fmt.Errorf("unexpected token: %s", p.currentToken.Type)
	}
}

// prefixexp ::= var | functioncall | ‘(’ exp ‘)’
func (p *Parser) parsePrefixExpression() (PrefixExpression, error) {
	switch p.currentToken.Type {
	// todo: add support for function calls
	case lexer.TokenIdentifier:
		name := p.currentToken.Value
		p.currentToken = p.lexer.NextToken()
		return NameVar{Name: name}, nil
	case lexer.TokenLeftParen:
		p.currentToken = p.lexer.NextToken()
		exp, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		if p.currentToken.Type != lexer.TokenRightParen {
			return nil, errors.New("missing ')'")
		}
		p.currentToken = p.lexer.NextToken()
		return exp.(PrefixExpression), nil
	default:
		return nil, fmt.Errorf("unexpected token: %s", p.currentToken.Type)
	}
}

// parlist ::= namelist [‘,’ ‘...’] | ‘...’
func (p *Parser) parseParameterList() (ParameterList, error) {
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
	return ParameterList{Names: names, IsVarArg: isVararg}, nil
}

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
