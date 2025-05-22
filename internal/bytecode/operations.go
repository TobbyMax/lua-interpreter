package bytecode

type OpCode int

const (
	OpPushNumber OpCode = iota
	OpPushString
	OpPushNil
	OpPushBool
	OpPushVarArg
	OpPushFunction
	OpReturn
	OpAdd
	OpSub
	OpMul
	OpDiv
	OpMod
	OpPow
	OpUnm
	OpNot
	OpLen
	OpConcat
	OpEq
	OpLt
	OpLe
	OpNeq
	OpGt
	OpGe
	OpTest
	OpTestSet
	OpCall
	OpTailCall
	OpGetGlobal
	OpSetGlobal
	OpGetLocal
	OpSetLocal
	OpNewTable
	OpSetTable
	OpGetTable
	OpForPrep
	OpForLoop
	OpJmp
	OpLoadBool
	OpLoadNil
)
