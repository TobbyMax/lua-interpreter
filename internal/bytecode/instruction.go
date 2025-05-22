package bytecode

type Instruction struct {
	Op   OpCode
	Args []interface{}
}
