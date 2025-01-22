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
