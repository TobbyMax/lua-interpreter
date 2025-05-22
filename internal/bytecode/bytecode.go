package bytecode

type Bytecode struct {
	Code      []Instruction
	LocalVars []string
	Constants []interface{}
}
