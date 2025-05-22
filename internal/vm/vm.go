package vm

import (
	"errors"
	"fmt"
	"math"

	"lua-interpreter/internal/bytecode"
)

type VM struct {
	// Instruction pointer
	pc       int
	bytecode bytecode.Bytecode
	stack    []bytecode.Value
	// Stack pointer
	sp      int
	locals  []bytecode.Value
	globals map[string]bytecode.Value
	// Call frames
	frames []callFrame
	// Current frame index
	currentFrame int
}

type callFrame struct {
	fn *bytecode.Function
	// Return PC (where to return after function call)
	returnPC int
	// Base stack pointer for this frame
	basePointer int
	// Number of expected results
	expectedResults int
}

func NewVM(bc bytecode.Bytecode) *VM {
	vm := &VM{
		pc:       0,
		bytecode: bc,
		stack:    make([]bytecode.Value, 1000),
		sp:       0,
		locals:   make([]bytecode.Value, len(bc.LocalVars)),
		globals:  make(map[string]bytecode.Value),
		frames:   make([]callFrame, 100),
	}

	vm.registerBuiltins()

	return vm
}

func (vm *VM) Run() (bytecode.Value, error) {
	for vm.pc < len(vm.bytecode.Code) {
		inst := vm.bytecode.Code[vm.pc]
		err := vm.executeInstruction(inst)
		if err != nil {
			return nil, fmt.Errorf("error at pc=%d: %w", vm.pc, err)
		}
		vm.pc++
	}

	if vm.sp > 0 {
		return vm.stack[vm.sp-1], nil
	}
	return nil, nil
}

func (vm *VM) executeInstruction(inst bytecode.Instruction) error {
	switch inst.Op {
	case bytecode.OpPushNumber:
		vm.push(inst.Args[0].(float64))
	case bytecode.OpPushString:
		vm.push(inst.Args[0].(string))
	case bytecode.OpPushNil:
		vm.push(nil)
	case bytecode.OpPushBool:
		vm.push(inst.Args[0].(bool))
	case bytecode.OpPushVarArg:
		// Get varargs from current frame's locals
		if vm.currentFrame >= 0 {
			if varargs, ok := vm.locals[len(vm.locals)-1].([]bytecode.Value); ok {
				for _, arg := range varargs {
					vm.push(arg)
				}
			} else {
				vm.push(nil)
			}
		} else {
			vm.push(nil)
		}
	case bytecode.OpPushFunction:
		vm.push(inst.Args[0].(*bytecode.Function))
	case bytecode.OpReturn:
		var nResults int
		if vm.sp > 0 {
			nResults = 1 // For now, always return 1 value
		}
		err := vm.return_(nResults)
		if err != nil {
			return err
		}
	case bytecode.OpAdd:
		b := vm.pop().(float64)
		a := vm.pop().(float64)
		vm.push(a + b)
	case bytecode.OpSub:
		b := vm.pop().(float64)
		a := vm.pop().(float64)
		vm.push(a - b)
	case bytecode.OpMul:
		b := vm.pop().(float64)
		a := vm.pop().(float64)
		vm.push(a * b)
	case bytecode.OpDiv:
		b := vm.pop().(float64)
		a := vm.pop().(float64)
		vm.push(a / b)
	case bytecode.OpMod:
		b := vm.pop().(float64)
		a := vm.pop().(float64)
		vm.push(math.Mod(a, b))
	case bytecode.OpPow:
		b := vm.pop().(float64)
		a := vm.pop().(float64)
		vm.push(math.Pow(a, b))
	case bytecode.OpUnm:
		a := vm.pop().(float64)
		vm.push(-a)
	case bytecode.OpNot:
		a := vm.pop()
		vm.push(vm.isFalse(a))
	case bytecode.OpLen:
		a := vm.pop()
		switch v := a.(type) {
		case string:
			vm.push(float64(len(v)))
		case map[interface{}]bytecode.Value:
			vm.push(float64(len(v)))
		default:
			return fmt.Errorf("length operator not supported for type %T", a)
		}
	case bytecode.OpConcat:
		b := vm.pop().(string)
		a := vm.pop().(string)
		vm.push(a + b)
	case bytecode.OpEq:
		b := vm.pop()
		a := vm.pop()
		vm.push(vm.equals(a, b))
	case bytecode.OpLt:
		b := vm.pop()
		a := vm.pop()
		vm.push(vm.lessThan(a, b))
	case bytecode.OpLe:
		b := vm.pop()
		a := vm.pop()
		vm.push(vm.lessOrEqual(a, b))
	case bytecode.OpNeq:
		b := vm.pop()
		a := vm.pop()
		vm.push(!vm.equals(a, b))
	case bytecode.OpGt:
		b := vm.pop()
		a := vm.pop()
		vm.push(!vm.lessOrEqual(a, b))
	case bytecode.OpGe:
		b := vm.pop()
		a := vm.pop()
		vm.push(!vm.lessThan(a, b))
	case bytecode.OpGetLocal:
		idx := inst.Args[0].(int)
		if idx >= len(vm.locals) {
			return fmt.Errorf("local variable index out of range: %d", idx)
		}
		vm.push(vm.locals[idx])
	case bytecode.OpSetLocal:
		idx := inst.Args[0].(int)
		if idx >= len(vm.locals) {
			return fmt.Errorf("local variable index out of range: %d", idx)
		}
		vm.locals[idx] = vm.peek()
	case bytecode.OpGetGlobal:
		name := inst.Args[0].(string)
		vm.push(vm.globals[name])
	case bytecode.OpSetGlobal:
		name := inst.Args[0].(string)
		vm.globals[name] = vm.peek()
	case bytecode.OpNewTable:
		vm.push(make(map[interface{}]bytecode.Value))
	case bytecode.OpGetTable:
		key := vm.pop()
		table := vm.pop().(map[interface{}]bytecode.Value)
		vm.push(table[key])
	case bytecode.OpSetTable:
		value := vm.pop()
		key := vm.pop()
		table := vm.pop().(map[interface{}]bytecode.Value)
		table[key] = value
	case bytecode.OpCall:
		nArgs := inst.Args[0].(int)
		err := vm.call(nArgs, 1)
		if err != nil {
			return err
		}
	case bytecode.OpTailCall:
		nArgs := inst.Args[0].(int)
		return vm.tailCall(nArgs)
	case bytecode.OpTest:
		cond := vm.pop()
		vm.push(cond)
	case bytecode.OpJmp:
		cond := vm.pop()
		if vm.isFalse(cond) {
			offset := inst.Args[0].(int)
			vm.pc += offset - 1
		}
	default:
		return fmt.Errorf("unknown opcode: %v", inst.Op)
	}
	return nil
}

func (vm *VM) push(v bytecode.Value) {
	vm.stack[vm.sp] = v
	vm.sp++
}

func (vm *VM) pop() bytecode.Value {
	vm.sp--
	return vm.stack[vm.sp]
}

func (vm *VM) peek() bytecode.Value {
	return vm.stack[vm.sp-1]
}

func (vm *VM) isFalse(v bytecode.Value) bool {
	switch v := v.(type) {
	case nil:
		return true
	case bool:
		return !v
	default:
		return false
	}
}

func (vm *VM) equals(a, b bytecode.Value) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return a == b
}

func (vm *VM) lessThan(a, b bytecode.Value) bool {
	switch va := a.(type) {
	case float64:
		if vb, ok := b.(float64); ok {
			return va < vb
		}
	case string:
		if vb, ok := b.(string); ok {
			return va < vb
		}
	}
	return false
}

func (vm *VM) lessOrEqual(a, b bytecode.Value) bool {
	return vm.lessThan(a, b) || vm.equals(a, b)
}

func (vm *VM) call(nArgs int, nResults int) error {
	fn := vm.stack[vm.sp-nArgs-1]

	switch f := fn.(type) {
	case *bytecode.Function:
		// Create new call frame
		frame := &vm.frames[vm.currentFrame+1]
		frame.fn = f
		frame.returnPC = vm.pc
		frame.basePointer = vm.sp - nArgs - 1
		frame.expectedResults = nResults

		// Set up new execution environment
		vm.currentFrame++
		vm.pc = 0
		vm.bytecode = f.Bytecode

		vm.locals = make([]bytecode.Value, len(f.Bytecode.LocalVars))
		for i := 0; i < f.NumParams && i < nArgs; i++ {
			vm.locals[i] = vm.stack[frame.basePointer+1+i]
		}

		if f.IsVararg && nArgs > f.NumParams {
			varargs := make([]bytecode.Value, nArgs-f.NumParams)
			copy(varargs, vm.stack[frame.basePointer+1+f.NumParams:frame.basePointer+1+nArgs])
			vm.locals[f.NumParams] = varargs
		}

		vm.sp = frame.basePointer + 1

	case func([]bytecode.Value) (bytecode.Value, error):
		args := make([]bytecode.Value, nArgs)
		copy(args, vm.stack[vm.sp-nArgs:vm.sp])

		result, err := f(args)
		if err != nil {
			return err
		}

		vm.sp -= nArgs + 1
		vm.push(result)

	default:
		return fmt.Errorf("attempt to call a %T value", fn)
	}

	return nil
}

func (vm *VM) tailCall(nArgs int) error {
	// Similar to regular call but reuses current frame
	return vm.call(nArgs, vm.frames[vm.currentFrame].expectedResults)
}

func (vm *VM) return_(nResults int) error {
	if vm.currentFrame < 0 {
		return errors.New("no function to return from")
	}

	frame := &vm.frames[vm.currentFrame]

	for i := 0; i < nResults && i < frame.expectedResults; i++ {
		vm.stack[frame.basePointer+i] = vm.stack[vm.sp-nResults+i]
	}

	// Restore previous execution environment
	vm.currentFrame--
	if vm.currentFrame >= 0 {
		prevFrame := &vm.frames[vm.currentFrame]
		vm.pc = prevFrame.returnPC
		vm.bytecode = prevFrame.fn.Bytecode
	}

	// Adjust stack pointer
	vm.sp = frame.basePointer + frame.expectedResults

	return nil
}

func (vm *VM) registerBuiltins() {
	vm.globals["print"] = func(args []bytecode.Value) (bytecode.Value, error) {
		for i, arg := range args {
			if i > 0 {
				fmt.Print("\t")
			}
			fmt.Print(arg)
		}
		fmt.Println()
		return nil, nil
	}
}
