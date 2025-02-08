package ast

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
		PrefixNames []string
		Name        string
		IsMethod    bool // true if it has a colon
	}
	// Function
	// function funcname funcbody
	Function struct {
		FunctionName FunctionName
		FuncBody     FunctionBody
	}
)
