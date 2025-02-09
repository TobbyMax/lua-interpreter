package ast

import "fmt"

type NativeFunction struct {
	Fn func(ctx *Context, args []Value) Value
}

func (nf *NativeFunction) Call(ctx *Context, args []Value) Value {
	return nf.Fn(ctx, args)
}

func toString(val Value) string {
	switch v := val.(type) {
	case nil:
		return "nil"
	case string:
		return v
	case float64:
		return fmt.Sprintf("%g", v)
	case bool:
		return fmt.Sprintf("%t", v)
	case *FunctionValue, *NativeFunction:
		return "function"
	case map[interface{}]Value:
		return "table"
	default:
		return fmt.Sprintf("<unknown:%T>", v)
	}
}

var printFn = &NativeFunction{
	Fn: func(ctx *Context, args []Value) Value {
		for i, arg := range args {
			if i > 0 {
				fmt.Print("\t")
			}
			fmt.Print(toString(arg))
		}
		fmt.Println()
		return nil
	},
}

var assertFn = &NativeFunction{
	Fn: func(ctx *Context, args []Value) Value {
		if len(args) == 0 {
			panic("assert: missing condition")
		}

		cond := args[0]
		if cond == false || cond == nil {
			var msg string
			if len(args) > 1 {
				msg, _ = args[1].(string)
			} else {
				msg = "assertion failed!"
			}
			panic("assert: " + msg)
		}

		return []Value{cond}
	},
}
