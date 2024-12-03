package vm

import (
	"fmt"

	"lua-interpreter/internal/bytecode"
)

type VM struct {
	stack []float64
	vars  map[string]float64
}

func New() *VM {
	return &VM{
		stack: []float64{},
		vars:  make(map[string]float64),
	}
}

func (vm *VM) push(v float64) {
	vm.stack = append(vm.stack, v)
}

func (vm *VM) pop() float64 {
	if len(vm.stack) == 0 {
		panic("stack underflow")
	}
	val := vm.stack[len(vm.stack)-1]
	vm.stack = vm.stack[:len(vm.stack)-1]
	return val
}

func (vm *VM) Run(code []bytecode.Instruction) error {
	for _, instr := range code {
		switch instr.Op {
		case bytecode.OpPushConst:
			vm.push(instr.Value)
		case bytecode.OpLoadVar:
			val, ok := vm.vars[instr.Arg]
			if !ok {
				return fmt.Errorf("undefined variable: %s", instr.Arg)
			}
			vm.push(val)
		case bytecode.OpStoreVar:
			val := vm.pop()
			vm.vars[instr.Arg] = val
		case bytecode.OpPrint:
			val := vm.pop()
			fmt.Println(val)
		default:
			return fmt.Errorf("unknown opcode: %v", instr.Op)
		}
	}
	return nil
}
