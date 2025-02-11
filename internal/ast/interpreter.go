package ast

import (
	"errors"
	"fmt"
	"math"
	"strings"

	"lua-interpreter/internal/lexer"
)

var (
	ErrBreak                = errors.New("break statement outside of loop")
	ErrVarArgNotDefined     = errors.New("cannot use '...' outside a vararg function ")
	ErrBitwiseAndOnlyInt    = errors.New("bitwise AND can only be applied to integers")
	ErrBitwiseOrOnlyInt     = errors.New("bitwise OR can only be applied to integers")
	ErrUnaryMinusOnlyNum    = errors.New("unary minus can only be applied to numbers")
	ErrBitwiseNotOnlyInt    = errors.New("bitwise NOT can only be applied to integers")
	ErrInvalidOperandLength = errors.New("invalid operand for # operator")
)

type GotoError struct {
	Label string
}

func (e *GotoError) Error() string {
	return fmt.Sprintf("no visible label '%s' for <goto>", e.Label)
}

type Value interface{}

type Context struct {
	Parent     *Context
	Return     Value // для возврата из функций
	isReturned bool
	Variables  map[string]Value
	globals    map[string]Value // только в корне
	labels     map[string]int   // для меток goto
}

func NewRootContext() *Context {
	ctx := &Context{
		Variables: make(map[string]Value),
		globals:   make(map[string]Value),
		labels:    make(map[string]int),
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
		labels:    make(map[string]int),
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
	return nil
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
	Eval(ctx *Context) (Value, error)
}

func (n *NumeralExpression) Eval(_ *Context) (Value, error) {
	return n.Value, nil
}

func (s *LiteralString) Eval(_ *Context) (Value, error) {
	return s.Value, nil
}

func (b *BooleanExpression) Eval(_ *Context) (Value, error) {
	return b.Value, nil
}

func (n *NilExpression) Eval(_ *Context) (Value, error) {
	return nil, nil
}

func (v *VarArgExpression) Eval(ctx *Context) (Value, error) {
	if ctx.Get("...") != nil {
		return ctx.Get("..."), nil
	}
	return nil, ErrVarArgNotDefined
}

func (b *BinaryOperatorExpression) Eval(ctx *Context) (Value, error) {
	left, _ := b.Left.Eval(ctx)
	right, _ := b.Right.Eval(ctx)

	switch b.Operator.Type {
	// Arithmetic operations
	case lexer.TokenPlus:
		return left.(float64) + right.(float64), nil
	case lexer.TokenMinus:
		return left.(float64) - right.(float64), nil
	case lexer.TokenMult:
		return left.(float64) * right.(float64), nil
	case lexer.TokenDiv:
		return left.(float64) / right.(float64), nil
	case lexer.TokenIntDiv:
		return float64(int(left.(float64)) / int(right.(float64))), nil
	case lexer.TokenMod:
		return float64(int(left.(float64)) % int(right.(float64))), nil
	case lexer.TokenPower:
		return math.Pow(left.(float64), right.(float64)), nil
	case lexer.TokenDoubleDot:
		return left.(string) + right.(string), nil
	// Comparison operations
	case lexer.TokenEqual:
		return left == right, nil
	case lexer.TokenNotEqual:
		return left != right, nil
	case lexer.TokenLess:
		return left.(float64) < right.(float64), nil
	case lexer.TokenLessEqual:
		return left.(float64) <= right.(float64), nil
	case lexer.TokenMore:
		return left.(float64) > right.(float64), nil
	case lexer.TokenMoreEqual:
		return left.(float64) >= right.(float64), nil
	// Logical operations
	case lexer.TokenKeywordAnd:
		return left.(bool) && right.(bool), nil
	case lexer.TokenKeywordOr:
		return left.(bool) || right.(bool), nil
	// Bitwise operations
	case lexer.TokenBinAnd:
		if numLeft, ok := left.(float64); ok {
			if numRight, ok := right.(float64); ok {
				if math.Trunc(numLeft) != numLeft || math.Trunc(numRight) != numRight {
					return nil, ErrBitwiseAndOnlyInt
				}
				return float64(int64(numLeft) & int64(numRight)), nil
			}
		}
		return nil, ErrBitwiseAndOnlyInt
	case lexer.TokenBinOr:
		if numLeft, ok := left.(float64); ok {
			if numRight, ok := right.(float64); ok {
				if math.Trunc(numLeft) != numLeft || math.Trunc(numRight) != numRight {
					return nil, ErrBitwiseOrOnlyInt
				}
				return float64(int64(numLeft) | int64(numRight)), nil
			}
		}
		return nil, ErrBitwiseOrOnlyInt
	default:
		return nil, fmt.Errorf("unknown binary operator: %s", b.Operator.Type.String())
	}
}

func (u *UnaryOperatorExpression) Eval(ctx *Context) (Value, error) {
	val, _ := u.Expression.Eval(ctx)

	switch u.Operator.Type {
	case lexer.TokenNot:
		if val.(bool) == false {
			return true, nil
		}
		return false, nil
	case lexer.TokenMinus:
		if num, ok := val.(float64); ok {
			return -num, nil
		}
		return nil, ErrUnaryMinusOnlyNum
	case lexer.TokenTilde:
		if num, ok := val.(float64); ok {
			if math.Trunc(num) != num {
				return nil, ErrBitwiseNotOnlyInt
			}
			return float64(^int64(num)), nil
		}
		return nil, ErrBitwiseNotOnlyInt
	case lexer.TokenHash:
		if str, ok := val.(string); ok {
			return float64(len(str)), nil
		} else if tbl, ok := val.(map[string]Value); ok {
			return float64(len(tbl)), nil
		} else if arr, ok := val.([]Value); ok {
			return float64(len(arr)), nil
		} else {
			return nil, ErrInvalidOperandLength
		}
	default:
		return nil, fmt.Errorf("unknown unary operator: %s", u.Operator.Type.String())
	}
}

func (t *TableConstructorExpression) Eval(ctx *Context) (Value, error) {
	table := map[interface{}]Value{}
	// Lua tables are 1-indexed by default
	var index float64 = 1

	for i, field := range t.Fields {
		switch f := field.(type) {
		case *ExpToExpField:
			key, _ := f.Key.Eval(ctx)
			val, _ := f.Value.Eval(ctx)
			table[key] = val
		case *NameField:
			val, _ := f.Value.Eval(ctx)
			table[f.Name] = val
		case *ExpressionField:
			val, _ := f.Value.Eval(ctx)
			vararg, ok := val.([]Value)
			if ok && len(vararg) > 0 {
				if i == len(t.Fields)-1 {
					for _, v := range vararg {
						table[index] = v
						index++
					}
				} else {
					table[index] = vararg[0]
				}
			} else {
				table[index] = val
				index++
			}
		}
	}

	return table, nil
}

func (b *Block) Eval(ctx *Context) (Value, error) {
	for i, stmt := range b.Statements {
		if l, ok := stmt.(*Label); ok {
			ctx.labels[l.Name] = i
		}
	}
	for i := 0; i < len(b.Statements); i++ {
		stmt := b.Statements[i]
		_, err := stmt.Eval(ctx)
		if err != nil {
			var gotoErr *GotoError
			if errors.As(err, &gotoErr) {
				if labelIndex, ok := ctx.labels[gotoErr.Label]; ok {
					i = labelIndex // Переход к метке
					continue
				} else {
					return nil, err
				}
			}
			return nil, fmt.Errorf("error evaluating statement: %w", err)
		}

		if ctx.isReturned {
			if ctx.Parent != nil {
				ctx.Parent.isReturned = true
				ctx.Parent.Return = ctx.Return
			}
			return ctx.Return, nil
		}
	}
	if b.ReturnStatement != nil {
		var vals []Value
		for _, exp := range b.ReturnStatement.Expressions {
			val, err := exp.Eval(ctx)
			if err != nil {
				return nil, fmt.Errorf("error evaluating return expression: %w", err)
			}
			vals = append(vals, val)
		}
		ctx.isReturned = true
		if len(vals) == 1 {
			ctx.Return = vals[0]
		} else {
			ctx.Return = vals // Можно сделать многозначный return как в Lua
		}
		if ctx.Parent != nil {
			ctx.Parent.isReturned = true
			ctx.Parent.Return = ctx.Return
		}
		return ctx.Return, nil
	}
	return nil, nil
}

func (f *FunctionDefinition) Eval(ctx *Context) (Value, error) {
	return &FunctionValue{
		Params:   f.FunctionBody.ParameterList.Names,
		IsVarArg: f.FunctionBody.ParameterList.IsVarArg,
		Body:     f.FunctionBody.Block,
	}, nil
}

func (fc *FunctionCall) Eval(ctx *Context) (Value, error) {
	prefixVal, _ := fc.PrefixExp.Eval(ctx)
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
				return nil, fmt.Errorf("undefined method '%s' for table", fc.Name)
			}
			fn, ok = field.(*FunctionValue)
			if !ok {
				return nil, fmt.Errorf("expected function for method '%s', got: %T", fc.Name, field)
			}
		} else {
			return nil, fmt.Errorf("prefix expression is not a table for method call: %T", prefixVal)
		}
	} else {
		switch val := prefixVal.(type) {
		case *NativeFunction:
			fnNative = val
		case *FunctionValue:
			fn = val
		default:
			return nil, fmt.Errorf("expected function or native function for function call, got: %T", prefixVal)
		}
	}

	fnCtx := ctx.NewChild()

	var args []Value
	switch a := fc.Args.(type) {
	case []Expression:
		for _, exp := range a {
			val, err := exp.Eval(ctx)
			if err != nil {
				return nil, fmt.Errorf("error evaluating argument: %w", err)
			}
			args = append(args, val)
		}
	case *TableConstructorExpression:
		t, err := a.Eval(ctx)
		if err != nil {
			return nil, fmt.Errorf("error evaluating table constructor: %w", err)
		}
		args = append(args, t)
	case *LiteralString:
		args = append(args, a.Value)
	}

	if fnNative != nil {
		return fnNative.Call(fnCtx, args), nil
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
	if fn.IsVarArg {
		var varargs []Value
		if len(args) > len(params) {
			varargs = args[len(params):]
		}
		fnCtx.Variables["..."] = varargs
	}

	res, err := fn.Body.Eval(fnCtx)
	ctx.isReturned = false
	return res, err
}

func (s *EmptyStatement) Eval(_ *Context) (Value, error) {
	return nil, nil
}

func (v *NameVar) Eval(ctx *Context) (Value, error) {
	return ctx.Get(v.Name), nil
}

func (v *IndexedVar) Eval(ctx *Context) (Value, error) {
	var table map[interface{}]Value
	prefix, err := v.PrefixExp.Eval(ctx)
	if err != nil {
		return nil, fmt.Errorf("error evaluating prefix expression: %w", err)
	}
	if tbl, ok := prefix.(map[interface{}]Value); !ok {
		return nil, fmt.Errorf("expected table for indexed variable, got: %T", table)
	} else {
		table = tbl
	}
	key, _ := v.Exp.Eval(ctx)
	val, ok := table[key]
	if !ok {
		return nil, nil
	}
	return val, nil
}

func (v *MemberVar) Eval(ctx *Context) (Value, error) {
	var table map[interface{}]Value
	prefix, err := v.PrefixExp.Eval(ctx)
	if err != nil {
		return nil, fmt.Errorf("error evaluating prefix expression: %w", err)
	}
	if tbl, ok := prefix.(map[interface{}]Value); !ok {
		return nil, fmt.Errorf("expected table for member variable, got: %T", table)
	} else {
		table = tbl
	}
	return table[v.Name], nil
}

func (s *LocalVarDeclaration) Eval(ctx *Context) (Value, error) {
	for i := 0; i < len(s.Vars); i++ {
		name := s.Vars[i]
		if i < len(s.Exps)-1 {
			val, err := s.Exps[i].Eval(ctx)
			if err != nil {
				return nil, fmt.Errorf("error evaluating local variable %s: %w", name, err)
			}
			vals, ok := val.([]Value)
			if ok && len(vals) > 0 {
				val = vals[0]
			}
			ctx.SetLocal(name, val)
		} else if i == len(s.Exps)-1 {
			val, err := s.Exps[i].Eval(ctx)
			if err != nil {
				return nil, fmt.Errorf("error evaluating local variable %s: %w", name, err)
			}
			vals, ok := val.([]Value)
			if ok {
				for _, v := range vals {
					if i < len(s.Vars) {
						ctx.SetLocal(s.Vars[i], v)
						i++
					} else {
						break
					}
				}
			} else {
				ctx.SetLocal(name, val)
			}
		} else {
			ctx.SetLocal(name, nil)
		}
	}
	return nil, nil
}

func (s *Assignment) Eval(ctx *Context) (Value, error) {
	for i := 0; i < len(s.Vars); i++ {
		v := s.Vars[i]
		if i < len(s.Exps)-1 {
			val, err := s.Exps[i].Eval(ctx)
			if err != nil {
				return nil, fmt.Errorf("error evaluating local variable %s: %w", v, err)
			}
			vals, ok := val.([]Value)
			if ok && len(vals) > 0 {
				val = vals[0]
			}
			err = v.Set(ctx, val)
			if err != nil {
				return nil, fmt.Errorf("error setting value for variable %s: %w", v, err)
			}
		} else if i == len(s.Exps)-1 {
			val, err := s.Exps[i].Eval(ctx)
			if err != nil {
				return nil, fmt.Errorf("error evaluating local variable %s: %w", v, err)
			}
			vals, ok := val.([]Value)
			if ok {
				for _, vl := range vals {
					if i < len(s.Vars) {
						err = s.Vars[i].Set(ctx, vl)
						if err != nil {
							return nil, fmt.Errorf("error setting value for variable %s: %w", v, err)
						}
						i++
					} else {
						break
					}
				}
			} else {
				err = v.Set(ctx, val)
				if err != nil {
					return nil, fmt.Errorf("error setting value for variable %s: %w", v, err)
				}
			}
		} else {
			err := v.Set(ctx, nil)
			if err != nil {
				return nil, fmt.Errorf("error setting value for variable %s: %w", v, err)
			}
		}
	}
	return nil, nil
}

func (s *Label) Eval(_ *Context) (Value, error) {
	return nil, nil
}

func (s *Goto) Eval(_ *Context) (Value, error) {
	return nil, &GotoError{Label: s.Name}
}

func (s *Break) Eval(_ *Context) (Value, error) {
	return nil, ErrBreak
}

func (s *Do) Eval(ctx *Context) (Value, error) {
	newCtx := ctx.NewChild()
	return s.Block.Eval(newCtx)
}

func (s *LocalFunction) Eval(ctx *Context) (Value, error) {
	fn := &FunctionValue{
		Params:   s.FunctionBody.ParameterList.Names,
		IsVarArg: s.FunctionBody.ParameterList.IsVarArg,
		Body:     s.FunctionBody.Block,
	}
	ctx.SetLocal(s.Name, fn)
	return nil, nil
}

func (s *While) Eval(ctx *Context) (Value, error) {
	for {
		cond, err := s.Exp.Eval(ctx)
		if err != nil {
			return nil, fmt.Errorf("error evaluating while condition: %w", err)
		}
		if !isTruthy(cond) {
			break
		}
		_, err = s.Block.Eval(ctx.NewChild())
		if errors.Is(err, ErrBreak) {
			break // прерывание цикла
		}
		if err != nil {
			return nil, fmt.Errorf("error evaluating while block: %w", err)
		}
	}
	return nil, nil
}

func (s *Repeat) Eval(ctx *Context) (Value, error) {
	for {
		_, err := s.Block.Eval(ctx.NewChild())
		if errors.Is(err, ErrBreak) {
			break // прерывание цикла
		}
		if err != nil {
			return nil, fmt.Errorf("error evaluating repeat block: %w", err)
		}
		cond, err := s.Exp.Eval(ctx)
		if err != nil {
			return nil, fmt.Errorf("error evaluating repeat condition: %w", err)
		}
		if isTruthy(cond) {
			break
		}
	}
	return nil, nil
}

func (s *For) Eval(ctx *Context) (Value, error) {
	val, err := s.Init.Eval(ctx)
	if err != nil {
		return nil, fmt.Errorf("error evaluating for loop init: %w", err)
	}
	init, ok := val.(float64)
	if !ok {
		return nil, fmt.Errorf("expected numeric value for for loop init, got: %T", val)
	}
	val, err = s.Limit.Eval(ctx)
	if err != nil {
		return nil, fmt.Errorf("error evaluating for loop limit: %w", err)
	}
	limit, ok := val.(float64)
	if !ok {
		return nil, fmt.Errorf("expected numeric value for for loop limit, got: %T", val)
	}
	step := 1.0
	if s.Step != nil {
		val, err = s.Step.Eval(ctx)
		if err != nil {
			return nil, fmt.Errorf("error evaluating for loop step: %w", err)
		}
		step, ok = val.(float64)
		if !ok {
			return nil, fmt.Errorf("expected numeric value for for loop step, got: %T", val)
		}
	}
	for i := init; (step > 0 && i <= limit) || (step < 0 && i >= limit); i += step {
		loopCtx := ctx.NewChild()
		loopCtx.SetLocal(s.Name, i)
		_, err = s.Block.Eval(loopCtx)
		if errors.Is(err, ErrBreak) {
			break // прерывание цикла
		}
		if err != nil {
			return nil, fmt.Errorf("error evaluating for loop body: %w", err)
		}
	}
	return nil, nil
}

func (s *ForIn) Eval(ctx *Context) (Value, error) {
	// todo:
	val, err := s.Exps[0].Eval(ctx)
	if err != nil {
		return nil, fmt.Errorf("error evaluating for-in iterator: %w", err)
	}
	iter, ok := val.(func() (map[string]Value, bool))
	if !ok {
		return nil, fmt.Errorf("expected iterator function for for-in loop, got: %T", val)
	}
	for {
		val, ok := iter()
		if !ok {
			break
		}
		loopCtx := ctx.NewChild()
		for _, name := range s.Names {
			loopCtx.SetLocal(name, val[name]) // упрощённо
		}
		_, err := s.Block.Eval(loopCtx)
		if errors.Is(err, ErrBreak) {
			break // прерывание цикла
		}
		if err != nil {
			return nil, fmt.Errorf("error evaluating for-in loop body: %w", err)
		}
	}
	return nil, nil
}

func (s *If) Eval(ctx *Context) (Value, error) {
	for i, cond := range s.Exps {
		res, err := cond.Eval(ctx)
		if err != nil {
			return nil, fmt.Errorf("error evaluating if condition: %w", err)
		}
		if isTruthy(res) {
			return s.Blocks[i].Eval(ctx.NewChild())
		}
	}
	// else-блок, если он есть
	if len(s.Blocks) > len(s.Exps) {
		return s.Blocks[len(s.Blocks)-1].Eval(ctx.NewChild())
	}
	return nil, nil
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

func (fb *FunctionBody) Eval(ctx *Context) (Value, error) {
	return &FunctionValue{
		Params:   fb.ParameterList.Names,
		IsVarArg: fb.ParameterList.IsVarArg,
		Body:     fb.Block,
	}, nil
}

func (f *Function) Eval(ctx *Context) (Value, error) {
	body, err := f.FuncBody.Eval(ctx)
	if err != nil {
		return nil, fmt.Errorf("error evaluating function body: %w", err)
	}
	fnVal, ok := body.(*FunctionValue)
	if !ok {
		return nil, fmt.Errorf("expected function value, got: %T", fnVal)
	}

	if len(f.FunctionName.PrefixNames) > 0 {
		table := ctx.Get(f.FunctionName.PrefixNames[0]).(map[interface{}]Value)
		for i, name := range f.FunctionName.PrefixNames[1:] {
			if field, ok := table[name]; ok {
				if innerTable, ok := field.(map[interface{}]Value); ok {
					table = innerTable
				} else {
					return nil, fmt.Errorf("expected table for prefix name '%s', got: %T", name, field)
				}
			} else {
				return nil, fmt.Errorf(
					"undefined table name '%s' in function definition",
					strings.Join(f.FunctionName.PrefixNames[:i+2], "."),
				)
			}
		}
		if f.FunctionName.IsMethod {
			fnVal.Params = append([]string{"self"}, fnVal.Params...)
		}
		table[f.FunctionName.Name] = fnVal
	} else {
		ctx.Set(f.FunctionName.Name, fnVal)
	}

	return nil, nil
}

type Settable interface {
	Set(ctx *Context, val Value) error
}

func (v *NameVar) Set(ctx *Context, val Value) error {
	ctx.Set(v.Name, val)
	return nil
}

func (v *IndexedVar) Set(ctx *Context, val Value) error {
	prefix, err := v.PrefixExp.Eval(ctx)
	if err != nil {
		return fmt.Errorf("error evaluating indexed variable prefix: %w", err)
	}
	table, ok := prefix.(map[interface{}]Value)
	if !ok {
		return fmt.Errorf("expected table for indexed variable, got: %T", prefix)
	}
	key, _ := v.Exp.Eval(ctx)
	table[key] = val
	return nil
}

func (v *MemberVar) Set(ctx *Context, val Value) error {
	prefix, err := v.PrefixExp.Eval(ctx)
	if err != nil {
		return fmt.Errorf("error evaluating member variable prefix: %w", err)
	}
	table, ok := prefix.(map[interface{}]Value)
	if !ok {
		return fmt.Errorf("expected table for member variable, got: %T", prefix)
	}
	table[v.Name] = val
	return nil
}
