package bytecode

type OpCode int

const (
	OpPushConst OpCode = iota
	OpLoadVar
	OpStoreVar
	OpPrint
)

type Instruction struct {
	Op    OpCode
	Arg   string  // используется для переменных или констант
	Value float64 // используется для чисел
}
