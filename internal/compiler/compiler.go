package compiler

import (
	"errors"
	"fmt"

	"lua-interpreter/internal/ast"
	"lua-interpreter/internal/bytecode"
	"lua-interpreter/internal/lexer"
)

type Compiler struct {
	bytecode bytecode.Bytecode
	// Track local variables in current scope
	locals map[string]int
	// Track scope nesting level
	scopeLevel int
}

func New() *Compiler {
	return &Compiler{
		bytecode: bytecode.Bytecode{
			Code:      make([]bytecode.Instruction, 0),
			LocalVars: make([]string, 0),
			Constants: make([]interface{}, 0),
		},
		locals:     make(map[string]int),
		scopeLevel: 0,
	}
}

func (c *Compiler) Compile(block *ast.Block) (bytecode.Bytecode, error) {
	err := c.compileBlock(block)
	if err != nil {
		return bytecode.Bytecode{}, err
	}
	return c.bytecode, nil
}

func (c *Compiler) compileBlock(block *ast.Block) error {
	c.enterScope()

	for _, stmt := range block.Statements {
		err := c.compileStatement(stmt)
		if err != nil {
			return err
		}
	}

	if block.ReturnStatement != nil {
		err := c.compileReturn(block.ReturnStatement)
		if err != nil {
			return err
		}
	} else {
		// Implicit return nil if no return statement
		c.emit(bytecode.OpPushNil)
		c.emit(bytecode.OpReturn)
	}

	c.exitScope()

	return nil
}

func (c *Compiler) compileStatement(stmt ast.Statement) error {
	switch s := stmt.(type) {
	case *ast.EmptyStatement:
		return nil
	case *ast.LocalVarDeclaration:
		return c.compileLocalVarDecl(s)
	case *ast.Assignment:
		return c.compileAssignment(s)
	case *ast.FunctionCall:
		return c.compileFunctionCall(s)
	case *ast.Do:
		return c.compileDoBlock(s)
	case *ast.While:
		return c.compileWhile(s)
	case *ast.If:
		return c.compileIf(s)
	case *ast.For:
		return c.compileForNum(s)
	case *ast.ForIn:
		return c.compileForIn(s)
	case *ast.LocalFunction:
		return c.compileLocalFunction(s)
	case *ast.FunctionDefinition:
		return c.compileFunctionDefinition(s)
	default:
		return fmt.Errorf("unsupported statement type: %T", stmt)
	}
}

func (c *Compiler) compileLocalVarDecl(decl *ast.LocalVarDeclaration) error {
	for i, exp := range decl.Exps {
		err := c.compileExpression(exp)
		if err != nil {
			return err
		}

		varName := decl.Vars[i]
		c.locals[varName] = len(c.locals)
		c.bytecode.LocalVars = append(c.bytecode.LocalVars, varName)

		c.emit(bytecode.OpSetLocal, c.locals[varName])
	}

	for i := len(decl.Exps); i < len(decl.Vars); i++ {
		c.emit(bytecode.OpLoadNil)
		varName := decl.Vars[i]
		c.locals[varName] = len(c.locals)
		c.bytecode.LocalVars = append(c.bytecode.LocalVars, varName)
		c.emit(bytecode.OpSetLocal, c.locals[varName])
	}

	return nil
}

func (c *Compiler) compilePrefixExp(prefixExp ast.PrefixExpression) error {
	switch e := prefixExp.(type) {
	case *ast.NameVar:
		if idx, ok := c.locals[e.Name]; ok {
			c.emit(bytecode.OpGetLocal, idx)
		} else {
			c.emit(bytecode.OpGetGlobal, e.Name)
		}
	case *ast.IndexedVar:
		err := c.compilePrefixExp(e.PrefixExp)
		if err != nil {
			return err
		}
		err = c.compilePrefixExp(e.Exp)
		if err != nil {
			return err
		}
		c.emit(bytecode.OpGetTable)
	case *ast.MemberVar:
		err := c.compilePrefixExp(e.PrefixExp)
		if err != nil {
			return err
		}
		c.emit(bytecode.OpPushString, e.Name)
		c.emit(bytecode.OpGetTable)
	default:
		return fmt.Errorf("unsupported prefix expression type: %T", prefixExp)
	}
	return nil
}

func (c *Compiler) compileExpression(exp ast.Expression) error {
	switch e := exp.(type) {
	case *ast.NumeralExpression:
		c.emit(bytecode.OpPushNumber, e.Value)
	case *ast.LiteralString:
		c.emit(bytecode.OpPushString, e.Value)
	case *ast.BooleanExpression:
		c.emit(bytecode.OpPushBool, e.Value)
	case *ast.NilExpression:
		c.emit(bytecode.OpPushNil)
	case *ast.BinaryOperatorExpression:
		err := c.compileBinaryOp(e)
		if err != nil {
			return err
		}
	case *ast.UnaryOperatorExpression:
		err := c.compileUnaryOp(e)
		if err != nil {
			return err
		}
	case *ast.VarArgExpression:
		c.emit(bytecode.OpPushVarArg)
	case *ast.FunctionCall:
		err := c.compileFunctionCall(e)
		if err != nil {
			return err
		}
	default:
		err := c.compilePrefixExp(exp)
		if err != nil {
			return fmt.Errorf("unsupported expression type: %T", exp)
		}
	}
	return nil
}

func (c *Compiler) compileBinaryOp(exp *ast.BinaryOperatorExpression) error {
	err := c.compileExpression(exp.Left)
	if err != nil {
		return err
	}

	err = c.compileExpression(exp.Right)
	if err != nil {
		return err
	}

	switch exp.Operator.Type {
	case lexer.TokenPlus:
		c.emit(bytecode.OpAdd)
	case lexer.TokenMinus:
		c.emit(bytecode.OpSub)
	case lexer.TokenMult:
		c.emit(bytecode.OpMul)
	case lexer.TokenDiv:
		c.emit(bytecode.OpDiv)
	case lexer.TokenMod:
		c.emit(bytecode.OpMod)
	case lexer.TokenPower:
		c.emit(bytecode.OpPow)
	case lexer.TokenEqual:
		c.emit(bytecode.OpEq)
	case lexer.TokenLess:
		c.emit(bytecode.OpLt)
	case lexer.TokenLessEqual:
		c.emit(bytecode.OpLe)
	case lexer.TokenDoubleDot:
		c.emit(bytecode.OpConcat)
	case lexer.TokenNotEqual:
		c.emit(bytecode.OpNeq)
	case lexer.TokenMore:
		c.emit(bytecode.OpGt)
	case lexer.TokenMoreEqual:
		c.emit(bytecode.OpGe)
	default:
		return fmt.Errorf("unsupported binary operator: %s", exp.Operator.Type)
	}

	return nil
}

func (c *Compiler) compileUnaryOp(exp *ast.UnaryOperatorExpression) error {
	err := c.compileExpression(exp.Expression)
	if err != nil {
		return err
	}

	switch exp.Operator.Type {
	case lexer.TokenMinus:
		c.emit(bytecode.OpUnm)
	case lexer.TokenHash:
		c.emit(bytecode.OpLen)
	case lexer.TokenNot:
		c.emit(bytecode.OpNot)
	default:
		return fmt.Errorf("unsupported unary operator: %s", exp.Operator.Type)
	}

	return nil
}

func (c *Compiler) compileReturn(ret *ast.ReturnStatement) error {
	if ret == nil {
		c.emit(bytecode.OpPushNil)
		c.emit(bytecode.OpReturn)
		return nil
	}

	for _, exp := range ret.Expressions {
		err := c.compileExpression(exp)
		if err != nil {
			return err
		}
	}
	c.emit(bytecode.OpReturn)
	return nil
}

func (c *Compiler) emit(op bytecode.OpCode, args ...interface{}) {
	c.bytecode.Code = append(c.bytecode.Code, bytecode.Instruction{Op: op, Args: args})
}

func (c *Compiler) enterScope() {
	c.scopeLevel++
}

func (c *Compiler) exitScope() {
	for name, idx := range c.locals {
		if idx >= len(c.bytecode.LocalVars)-c.scopeLevel {
			delete(c.locals, name)
		}
	}
	c.scopeLevel--
}

func (c *Compiler) compileAssignment(assign *ast.Assignment) error {
	for _, exp := range assign.Exps {
		err := c.compileExpression(exp)
		if err != nil {
			return err
		}
	}

	for i := len(assign.Vars) - 1; i >= 0; i-- {
		v := assign.Vars[i]
		switch vr := v.(type) {
		case *ast.NameVar:
			if idx, ok := c.locals[vr.Name]; ok {
				c.emit(bytecode.OpSetLocal, idx)
			} else {
				c.emit(bytecode.OpSetGlobal, vr.Name)
			}
		case *ast.IndexedVar:
			err := c.compileExpression(vr.PrefixExp)
			if err != nil {
				return err
			}
			err = c.compileExpression(vr.Exp)
			if err != nil {
				return err
			}
			c.emit(bytecode.OpSetTable)
		case *ast.MemberVar:
			err := c.compileExpression(vr.PrefixExp)
			if err != nil {
				return err
			}
			c.emit(bytecode.OpPushString, vr.Name)
			c.emit(bytecode.OpSetTable)
		}
	}

	return nil
}

func (c *Compiler) compileFunctionCall(call *ast.FunctionCall) error {
	err := c.compilePrefixExp(call.PrefixExp)
	if err != nil {
		return err
	}

	if call.Name != "" {
		c.emit(bytecode.OpPushString, call.Name)
		c.emit(bytecode.OpGetTable)
	}

	nArgs := 0
	switch args := call.Args.(type) {
	case []ast.Expression:
		for _, arg := range args {
			err := c.compileExpression(arg)
			if err != nil {
				return err
			}
			nArgs++
		}
	case *ast.TableConstructorExpression:
		err := c.compileExpression(args)
		if err != nil {
			return err
		}
		nArgs = 1
	case *ast.LiteralString:
		c.emit(bytecode.OpPushString, args.Value)
		nArgs = 1
	}

	c.emit(bytecode.OpCall, nArgs)
	return nil
}

func (c *Compiler) compileDoBlock(do *ast.Do) error {
	return c.compileBlock(&do.Block)
}

func (c *Compiler) compileWhile(while *ast.While) error {
	conditionPos := len(c.bytecode.Code)

	err := c.compileExpression(while.Exp)
	if err != nil {
		return err
	}

	c.emit(bytecode.OpTest)
	jumpPos := len(c.bytecode.Code)
	c.emit(bytecode.OpJmp, 0) // Placeholder jump offset

	err = c.compileBlock(&while.Block)
	if err != nil {
		return err
	}

	c.emit(bytecode.OpJmp, conditionPos-len(c.bytecode.Code)-1)

	c.bytecode.Code[jumpPos].Args[0] = len(c.bytecode.Code) - jumpPos - 1

	return nil
}

func (c *Compiler) compileIf(ifStmt *ast.If) error {
	endJumps := make([]int, 0)

	for i, exp := range ifStmt.Exps {
		err := c.compileExpression(exp)
		if err != nil {
			return err
		}

		c.emit(bytecode.OpTest)
		jumpPos := len(c.bytecode.Code)
		c.emit(bytecode.OpJmp, 0)

		err = c.compileBlock(&ifStmt.Blocks[i])
		if err != nil {
			return err
		}

		if i < len(ifStmt.Exps)-1 {
			endJumps = append(endJumps, len(c.bytecode.Code))
			c.emit(bytecode.OpJmp, 0) // Placeholder jump offset
		}

		c.bytecode.Code[jumpPos].Args[0] = len(c.bytecode.Code) - jumpPos - 1
	}

	if len(ifStmt.Blocks) > len(ifStmt.Exps) {
		err := c.compileBlock(&ifStmt.Blocks[len(ifStmt.Blocks)-1])
		if err != nil {
			return err
		}
	}

	for _, pos := range endJumps {
		c.bytecode.Code[pos].Args[0] = len(c.bytecode.Code) - pos - 1
	}

	return nil
}

func (c *Compiler) compileForNum(forStmt *ast.For) error {
	err := c.compileExpression(forStmt.Init)
	if err != nil {
		return err
	}

	err = c.compileExpression(forStmt.Limit)
	if err != nil {
		return err
	}

	// Compile step expression (default to 1 if not provided)
	if forStmt.Step != nil {
		err = c.compileExpression(forStmt.Step)
		if err != nil {
			return err
		}
	} else {
		c.emit(bytecode.OpPushNumber, float64(1))
	}

	loopStartPos := len(c.bytecode.Code)
	c.emit(bytecode.OpForPrep, 0) // Placeholder jump offset

	c.locals[forStmt.Name] = len(c.locals)
	c.bytecode.LocalVars = append(c.bytecode.LocalVars, forStmt.Name)

	err = c.compileBlock(&forStmt.Block)
	if err != nil {
		return err
	}

	c.emit(bytecode.OpForLoop, loopStartPos-len(c.bytecode.Code)-1)

	c.bytecode.Code[loopStartPos].Args[0] = len(c.bytecode.Code) - loopStartPos - 1

	return nil
}

func (c *Compiler) compileForIn(forStmt *ast.ForIn) error {
	// TODO: Implement generic for loop
	return errors.New("generic for loop not implemented yet")
}

func (c *Compiler) compileLocalFunction(fn *ast.LocalFunction) error {
	function := &bytecode.Function{
		NumParams: len(fn.FunctionBody.ParameterList.Names),
		IsVararg:  fn.FunctionBody.ParameterList.IsVarArg,
	}

	funcCompiler := New()

	for _, param := range fn.FunctionBody.ParameterList.Names {
		funcCompiler.locals[param] = len(funcCompiler.locals)
		funcCompiler.bytecode.LocalVars = append(funcCompiler.bytecode.LocalVars, param)
	}

	if fn.FunctionBody.ParameterList.IsVarArg {
		funcCompiler.locals["..."] = len(funcCompiler.locals)
		funcCompiler.bytecode.LocalVars = append(funcCompiler.bytecode.LocalVars, "...")
	}

	bc, err := funcCompiler.Compile(&fn.FunctionBody.Block)
	if err != nil {
		return err
	}

	function.Bytecode = bc

	c.locals[fn.Name] = len(c.locals)
	c.bytecode.LocalVars = append(c.bytecode.LocalVars, fn.Name)

	c.emit(bytecode.OpPushFunction, function)
	c.emit(bytecode.OpSetLocal, c.locals[fn.Name])

	return nil
}

func (c *Compiler) compileFunctionDefinition(fn *ast.FunctionDefinition) error {
	function := &bytecode.Function{
		NumParams: len(fn.FunctionBody.ParameterList.Names),
		IsVararg:  fn.FunctionBody.ParameterList.IsVarArg,
	}

	funcCompiler := New()

	for _, param := range fn.FunctionBody.ParameterList.Names {
		funcCompiler.locals[param] = len(funcCompiler.locals)
		funcCompiler.bytecode.LocalVars = append(funcCompiler.bytecode.LocalVars, param)
	}

	if fn.FunctionBody.ParameterList.IsVarArg {
		funcCompiler.locals["..."] = len(funcCompiler.locals)
		funcCompiler.bytecode.LocalVars = append(funcCompiler.bytecode.LocalVars, "...")
	}

	bc, err := funcCompiler.Compile(&fn.FunctionBody.Block)
	if err != nil {
		return err
	}

	function.Bytecode = bc

	c.emit(bytecode.OpPushFunction, function)

	return nil
}
