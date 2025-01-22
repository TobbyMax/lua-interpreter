package ast

import (
	"lua-interpreter/internal/lexer"
)

type (
	// FunctionDefinition
	// functiondef ::= function funcbody
	FunctionDefinition struct {
		FunctionBody FunctionBody
	}
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
	// ExpressionList
	// explist ::= exp {‘,’ exp}
	ExpressionList struct {
		Expressions []Expression
	}
	// Args [ExpressionList | TableConstructorExpression | LiteralString]
	// args ::= ‘(’ [explist] ‘)’ | tableconstructor | LiteralString
	Args interface{}
	// FunctionCall
	// functioncall ::= prefixexp args | prefixexp ‘:’ Name args
	FunctionCall struct {
		PrefixExp PrefixExpression
		Name      string
		Args      Args
	}
	// PrefixExpression
	// prefixexp ::= var | functioncall | ‘(’ exp ‘)’
	PrefixExpression        interface{}
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
)

type (
	NilExpression     struct{}
	BooleanExpression struct {
		Value bool
	}

	// ParameterList
	// parlist ::= namelist [‘,’ ‘...’] | ‘...’

	// ReturnStatement ::= return [explist] [‘;’]
	// retstat ::= return [explist] [‘;’]

	// Block
	// block ::= { stat } [ retstat ]

)
