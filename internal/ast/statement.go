package ast

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
