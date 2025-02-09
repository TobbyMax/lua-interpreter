package ast

import (
	"fmt"
	"math"
	"strings"

	"lua-interpreter/internal/lexer"
)

type Value interface{}

type Context struct {
	Parent    *Context
	Return    Value // для возврата из функций
	Variables map[string]Value
	globals   map[string]Value // только в корне
}

func NewRootContext() *Context {
	ctx := &Context{
		Variables: make(map[string]Value),
		globals:   make(map[string]Value),
	}
	ctx.Set("print", printFn)
	ctx.Set("assert", assertFn)

	return ctx
}

func (ctx *Context) NewChild() *Context {
	return &Context{
		Parent:    ctx,
		Variables: make(map[string]Value),
		globals:   ctx.globals,
	}
}

func (ctx *Context) SetLocal(name string, val Value) {
	ctx.Variables[name] = val
}

func (ctx *Context) Get(name string) Value {
	for c := ctx; c != nil; c = c.Parent {
		if val, ok := c.Variables[name]; ok {
			return val
		}
	}
	if val, ok := ctx.globals[name]; ok {
		return val
	}
	panic("undefined variable: " + name)
}

func (ctx *Context) Set(name string, val Value) {
	for c := ctx; c != nil; c = c.Parent {
		if _, ok := c.Variables[name]; ok {
			c.Variables[name] = val
			return
		}
	}
	ctx.globals[name] = val
}

type Evaluable interface {
	Eval(ctx *Context) Value
}

func (n *NumeralExpression) Eval(ctx *Context) Value {
	return n.Value
}

func (s *LiteralString) Eval(ctx *Context) Value {
	return s.Value
}

func (b *BooleanExpression) Eval(ctx *Context) Value {
	return b.Value
}

func (n *NilExpression) Eval(ctx *Context) Value {
	return nil
}

func (v *VarArgExpression) Eval(ctx *Context) Value {
	panic("vararg not implemented in this context")
}

func (b *BinaryOperatorExpression) Eval(ctx *Context) Value {
	left := b.Left.Eval(ctx)
	right := b.Right.Eval(ctx)

	switch b.Operator.Type {
	// Arithmetic operations
	case lexer.TokenPlus:
		return left.(float64) + right.(float64)
	case lexer.TokenMinus:
		return left.(float64) - right.(float64)
	case lexer.TokenMult:
		return left.(float64) * right.(float64)
	case lexer.TokenDiv:
		return left.(float64) / right.(float64)
	case lexer.TokenIntDiv:
		return float64(int(left.(float64)) / int(right.(float64)))
	case lexer.TokenMod:
		return float64(int(left.(float64)) % int(right.(float64)))
	case lexer.TokenPower:
		return math.Pow(left.(float64), right.(float64))
	case lexer.TokenDoubleDot:
		return left.(string) + right.(string)
	// Comparison operations
	case lexer.TokenEqual:
		return left == right
	case lexer.TokenNotEqual:
		return left != right
	case lexer.TokenLess:
		return left.(float64) < right.(float64)
	case lexer.TokenLessEqual:
		return left.(float64) <= right.(float64)
	case lexer.TokenMore:
		return left.(float64) > right.(float64)
	case lexer.TokenMoreEqual:
		return left.(float64) >= right.(float64)
	// Logical operations
	case lexer.TokenKeywordAnd:
		return left.(bool) && right.(bool)
	case lexer.TokenKeywordOr:
		return left.(bool) || right.(bool)
	// Bitwise operations
	case lexer.TokenBinAnd:
		if numLeft, ok := left.(float64); ok {
			if numRight, ok := right.(float64); ok {
				if math.Trunc(numLeft) != numLeft || math.Trunc(numRight) != numRight {
					panic("bitwise AND can only be applied to integers")
				}
				return float64(int64(numLeft) & int64(numRight))
			}
		}
		panic("bitwise AND can only be applied to integers")
	case lexer.TokenBinOr:
		if numLeft, ok := left.(float64); ok {
			if numRight, ok := right.(float64); ok {
				if math.Trunc(numLeft) != numLeft || math.Trunc(numRight) != numRight {
					panic("bitwise OR can only be applied to integers")
				}
				return float64(int64(numLeft) | int64(numRight))
			}
		}
		panic("bitwise OR can only be applied to integers")
	default:
		panic("unknown binary operator: " + b.Operator.Type.String())
	}
}

func (u *UnaryOperatorExpression) Eval(ctx *Context) Value {
	val := u.Expression.Eval(ctx)

	switch u.Operator.Type {
	case lexer.TokenNot:
		if val.(bool) == false {
			return true
		}
		return false
	case lexer.TokenMinus:
		if num, ok := val.(float64); ok {
			return -num
		}
		panic("unary minus can only be applied to numbers")
	case lexer.TokenTilde:
		if num, ok := val.(float64); ok {
			if math.Trunc(num) != num {
				panic("operand has a non-integer value in bitwise NOT operation")
			}
			return float64(^int64(num))
		}
		panic("bitwise NOT can only be applied to integers")
	case lexer.TokenHash:
		if str, ok := val.(string); ok {
			return float64(len(str))
		} else if tbl, ok := val.(map[string]Value); ok {
			return float64(len(tbl))
		} else if arr, ok := val.([]Value); ok {
			return float64(len(arr))
		} else {
			panic("invalid operand for # operator")
		}
	default:
		panic("unknown unary operator: " + u.Operator.Type.String())
	}
}

func (t *TableConstructorExpression) Eval(ctx *Context) Value {
	table := map[interface{}]Value{}
	// Lua tables are 1-indexed by default
	var index float64 = 1

	for _, field := range t.Fields {
		switch f := field.(type) {
		case *ExpToExpField:
			key := f.Key.Eval(ctx)
			val := f.Value.Eval(ctx)
			table[key] = val
		case *NameField:
			val := f.Value.Eval(ctx)
			table[f.Name] = val
		case *ExpressionField:
			val := f.Value.Eval(ctx)
			table[index] = val
			index++
		}
	}

	return table
}

func (b *Block) Eval(ctx *Context) Value {
	for _, stmt := range b.Statements {
		stmt.Eval(ctx)
		if ctx.Return != nil {
			return ctx.Return
		}
	}
	if b.ReturnStatement != nil {
		var vals []Value
		for _, exp := range b.ReturnStatement.Expressions {
			vals = append(vals, exp.Eval(ctx))
		}
		if len(vals) == 1 {
			ctx.Return = vals[0]
		} else {
			ctx.Return = vals // Можно сделать многозначный return как в Lua
		}
		return ctx.Return
	}
	return nil
}

func (f *FunctionDefinition) Eval(ctx *Context) Value {
	return &FunctionValue{
		Params:   f.FunctionBody.ParameterList.Names,
		IsVarArg: f.FunctionBody.ParameterList.IsVarArg,
		Body:     f.FunctionBody.Block,
	}
}

func (fc *FunctionCall) Eval(ctx *Context) Value {
	prefixVal := fc.PrefixExp.Eval(ctx)
	var (
		fn           *FunctionValue
		fnNative     *NativeFunction
		table        map[interface{}]Value
		ok           bool
		isMethodCall bool
	)
	if fc.Name != "" {
		isMethodCall = true
		if table, ok = prefixVal.(map[interface{}]Value); ok {
			field, ok := table[fc.Name]
			if !ok {
				panic("undefined method: " + fc.Name)
			}
			fn, ok = field.(*FunctionValue)
			if !ok {
				panic("expected function for method: " + fc.Name)
			}
		} else {
			panic("prefix expression is not a table for method call")
		}
	} else {
		switch val := prefixVal.(type) {
		case *NativeFunction:
			fnNative = val
		case *FunctionValue:
			fn = val
		default:
			panic("expected function or native function for function call, got: " + fmt.Sprintf("%T", prefixVal))
		}
	}

	fnCtx := ctx.NewChild()

	var args []Value
	switch a := fc.Args.(type) {
	case []Expression:
		for _, exp := range a {
			args = append(args, exp.Eval(ctx))
		}
	case *TableConstructorExpression:
		args = append(args, a.Eval(ctx))
	case *LiteralString:
		args = append(args, a.Value)
	}

	if fnNative != nil {
		return fnNative.Call(fnCtx, args)
	}

	params := fn.Params
	if isMethodCall {
		fnCtx.Variables[params[0]] = table
		params = params[1:]
	}
	for i, name := range params {
		if i < len(args) {
			fnCtx.Variables[name] = args[i]
		} else {
			fnCtx.Variables[name] = nil
		}
	}
	// vararg может быть сохранён в `_VARARG` или аналоге

	return fn.Body.Eval(fnCtx)
}

func (s *EmptyStatement) Eval(ctx *Context) Value {
	return nil
}

func (v *NameVar) Eval(ctx *Context) Value {
	return ctx.Get(v.Name)
}

func (v *IndexedVar) Eval(ctx *Context) Value {
	table := v.PrefixExp.Eval(ctx).(map[interface{}]Value)
	key := v.Exp.Eval(ctx)
	val, ok := table[key]
	if !ok {
		return nil
	}
	return val
}

func (v *MemberVar) Eval(ctx *Context) Value {
	table := v.PrefixExp.Eval(ctx).(map[interface{}]Value)
	return table[v.Name]
}

func (s *LocalVarDeclaration) Eval(ctx *Context) Value {
	for i, name := range s.Vars {
		if i < len(s.Exps) {
			ctx.SetLocal(name, s.Exps[i].Eval(ctx))
		} else {
			ctx.SetLocal(name, nil)
		}
	}
	return nil
}

func (s *Assignment) Eval(ctx *Context) Value {
	for i, v := range s.Vars {
		var val Value
		if i < len(s.Exps) {
			val = s.Exps[i].Eval(ctx)
		}
		switch varExpr := v.(type) {
		case *NameVar:
			ctx.Set(varExpr.Name, val)
		case *IndexedVar:
			table := varExpr.PrefixExp.Eval(ctx).(map[interface{}]Value)
			key := varExpr.Exp.Eval(ctx)
			table[key] = val
		case *MemberVar:
			table := varExpr.PrefixExp.Eval(ctx).(map[interface{}]Value)
			table[varExpr.Name] = val
		default:
			panic("unsupported assignment target")
		}
	}
	return nil
}

func (s *Label) Eval(ctx *Context) Value {
	return nil
}

func (s *Goto) Eval(ctx *Context) Value {
	panic("goto not implemented")
}

func (s *Break) Eval(ctx *Context) Value {
	panic("break")
}

func (s *Do) Eval(ctx *Context) Value {
	newCtx := ctx.NewChild()
	return s.Block.Eval(newCtx)
}

func (s *LocalFunction) Eval(ctx *Context) Value {
	fn := &FunctionValue{
		Params:   s.FunctionBody.ParameterList.Names,
		IsVarArg: s.FunctionBody.ParameterList.IsVarArg,
		Body:     s.FunctionBody.Block,
	}
	ctx.SetLocal(s.Name, fn)
	return nil
}

func (s *While) Eval(ctx *Context) Value {
	for {
		cond := s.Exp.Eval(ctx).(bool)
		if !cond {
			break
		}
		func() {
			defer func() {
				if r := recover(); r != nil && r != "break" {
					panic(r)
				}
			}()
			s.Block.Eval(ctx.NewChild())
		}()
	}
	return nil
}

func (s *Repeat) Eval(ctx *Context) Value {
	for {
		func() {
			defer func() {
				if r := recover(); r != nil && r != "break" {
					panic(r)
				}
			}()
			s.Block.Eval(ctx.NewChild())
		}()
		cond := s.Exp.Eval(ctx).(bool)
		if cond {
			break
		}
	}
	return nil
}

func (s *For) Eval(ctx *Context) Value {
	init := s.Init.Eval(ctx).(float64)
	limit := s.Limit.Eval(ctx).(float64)
	step := 1.0
	if s.Step != nil {
		step = (*s.Step).Eval(ctx).(float64)
	}
	for i := init; (step > 0 && i <= limit) || (step < 0 && i >= limit); i += step {
		loopCtx := ctx.NewChild()
		loopCtx.SetLocal(s.Name, i)
		func() {
			defer func() {
				if r := recover(); r != nil && r != "break" {
					panic(r)
				}
			}()
			s.Block.Eval(loopCtx)
		}()
	}
	return nil
}

func (s *ForIn) Eval(ctx *Context) Value {
	// todo:
	iter := s.Exps[0].Eval(ctx).(func() (map[string]Value, bool))
	for {
		val, ok := iter()
		if !ok {
			break
		}
		loopCtx := ctx.NewChild()
		for _, name := range s.Names {
			loopCtx.SetLocal(name, val[name]) // упрощённо
		}
		func() {
			defer func() {
				if r := recover(); r != nil && r != "break" {
					panic(r)
				}
			}()
			s.Block.Eval(loopCtx)
		}()
	}
	return nil
}

func (s *If) Eval(ctx *Context) Value {
	for i, cond := range s.Exps {
		if isTruthy(cond.Eval(ctx)) {
			return s.Blocks[i].Eval(ctx.NewChild())
		}
	}
	// else-блок, если он есть
	if len(s.Blocks) > len(s.Exps) {
		return s.Blocks[len(s.Blocks)-1].Eval(ctx.NewChild())
	}
	return nil
}

func isTruthy(val Value) bool {
	switch v := val.(type) {
	case nil:
		return false
	case bool:
		return v
	default:
		return true
	}
}

type FunctionValue struct {
	Params   []string
	IsVarArg bool
	Body     Block
}

func (fb *FunctionBody) Eval(ctx *Context) Value {
	return &FunctionValue{
		Params:   fb.ParameterList.Names,
		IsVarArg: fb.ParameterList.IsVarArg,
		Body:     fb.Block,
	}
}
func (f *Function) Eval(ctx *Context) Value {
	fnVal := f.FuncBody.Eval(ctx).(*FunctionValue)

	if len(f.FunctionName.PrefixNames) > 0 {
		table := ctx.Get(f.FunctionName.PrefixNames[0]).(map[interface{}]Value)
		for _, name := range f.FunctionName.PrefixNames[1:] {
			if field, ok := table[name]; ok {
				if innerTable, ok := field.(map[interface{}]Value); ok {
					table = innerTable
				} else {
					panic("expected table for function prefix")
				}
			} else {
				panic("undefined prefix in function name: " + strings.Join(f.FunctionName.PrefixNames, "."))
			}
		}
		if f.FunctionName.IsMethod {
			fnVal.Params = append([]string{"self"}, fnVal.Params...)
		}
		table[f.FunctionName.Name] = fnVal
	} else {
		ctx.Set(f.FunctionName.Name, fnVal)
	}

	return nil
}
