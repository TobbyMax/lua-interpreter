package bytecode

type Function struct {
	Bytecode  Bytecode
	NumParams int
	IsVararg  bool
	Upvalues  []Value
}
