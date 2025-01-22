package ast

type (
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
