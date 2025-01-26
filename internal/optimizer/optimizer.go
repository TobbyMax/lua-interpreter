package optimizer

import (
	"math"

	"lua-interpreter/internal/ast"
	"lua-interpreter/internal/lexer"
)

type (
	BinaryOptimizer func(exp *ast.BinaryOperatorExpression) ast.Expression
	UnaryOptimizer  func(exp *ast.UnaryOperatorExpression) ast.Expression
)

func isTrue(exp ast.Expression) bool {
	switch exp.(type) {
	case *ast.NumeralExpression, *ast.LiteralString:
		return true
	case *ast.BooleanExpression:
		return exp.(*ast.BooleanExpression).Value
	default:
		return false
	}
}

func isFalse(exp ast.Expression) bool {
	switch exp.(type) {
	case *ast.NilExpression:
		return !isTrue(exp)
	case *ast.BooleanExpression:
		return !exp.(*ast.BooleanExpression).Value
	default:
		return false
	}
}

func noBinaryOptimizer(exp *ast.BinaryOperatorExpression) ast.Expression {
	return exp
}

func OptimizeBinary(opType lexer.TokenType) BinaryOptimizer {
	switch opType {
	case lexer.TokenKeywordOr:
		return optimizeOr
	case lexer.TokenKeywordAnd:
		return optimizeAnd
	case lexer.TokenPlus, lexer.TokenMinus, lexer.TokenMult, lexer.TokenDiv,
		lexer.TokenIntDiv, lexer.TokenMod, lexer.TokenPower:
		return optimizeArithmetic(opType)
	case lexer.TokenBinAnd, lexer.TokenBinOr, lexer.TokenTilde,
		lexer.TokenShiftRight, lexer.TokenShiftLeft:
		return optimizeBitOperations(opType)
	default:
		return noBinaryOptimizer
	}
}

func noUnaryOptimizer(exp *ast.UnaryOperatorExpression) ast.Expression {
	return exp
}

func OptimizeUnary(opType lexer.TokenType) UnaryOptimizer {
	switch opType {
	case lexer.TokenMinus:
		return optimizeUnaryMinus
	case lexer.TokenNot:
		return optimizeNot
	case lexer.TokenTilde:
		return optimizeBitNot
	default:
		return noUnaryOptimizer
	}
}

func optimizeBitOperations(opType lexer.TokenType) BinaryOptimizer {
	return func(exp *ast.BinaryOperatorExpression) ast.Expression {
		first, okFst := exp.Left.(*ast.NumeralExpression)
		second, okSnd := exp.Right.(*ast.NumeralExpression)
		if okFst && okSnd {
			if math.Trunc(first.Value) != first.Value || math.Trunc(second.Value) != second.Value {
				// todo: return an error instead of panic
				panic("one of the operands has a non-integer value in bit operation")
			}
			a := int64(first.Value)
			b := int64(second.Value)
			switch opType {
			case lexer.TokenBinAnd:
				return &ast.NumeralExpression{Value: float64(a & b)}
			case lexer.TokenBinOr:
				return &ast.NumeralExpression{Value: float64(a | b)}
			case lexer.TokenShiftLeft:
				if b >= 0 {
					return &ast.NumeralExpression{Value: float64(a << uint64(b))}
				} else {
					return &ast.NumeralExpression{Value: float64(a >> uint64(-b))}
				}
			case lexer.TokenShiftRight:
				if b >= 0 {
					return &ast.NumeralExpression{Value: float64(a >> uint64(b))}
				} else {
					return &ast.NumeralExpression{Value: float64(a << uint64(-b))}
				}
			default:
				return exp
			}
		}
		return exp
	}
}

// https://www.lua.org/manual/5.3/manual.html#3.4.8
func optimizeOr(expr *ast.BinaryOperatorExpression) ast.Expression {
	if isTrue(expr.Left) {
		return expr.Left
	}
	if isFalse(expr.Left) {
		return expr.Right
	}
	return expr
}

// https://www.lua.org/manual/5.3/manual.html#3.4.8
func optimizeAnd(expr *ast.BinaryOperatorExpression) ast.Expression {
	if isFalse(expr.Left) {
		return expr.Left
	}
	if isTrue(expr.Left) {
		return expr.Right
	}
	return expr
}

func optimizeArithmetic(opType lexer.TokenType) BinaryOptimizer {
	return func(exp *ast.BinaryOperatorExpression) ast.Expression {
		first, okFst := exp.Left.(*ast.NumeralExpression)
		second, okSnd := exp.Right.(*ast.NumeralExpression)
		if okFst && okSnd {
			a := first.Value
			b := second.Value
			switch opType {
			case lexer.TokenPlus:
				return &ast.NumeralExpression{Value: a + b}
			case lexer.TokenMinus:
				return &ast.NumeralExpression{Value: a - b}
			case lexer.TokenMult:
				return &ast.NumeralExpression{Value: a * b}
			case lexer.TokenDiv:
				if b != 0 {
					return &ast.NumeralExpression{Value: a / b}
				}
			case lexer.TokenIntDiv:
				if b != 0 {
					return &ast.NumeralExpression{Value: math.Floor(a / b)}
				}
			case lexer.TokenMod:
				if b != 0 {
					return &ast.NumeralExpression{Value: a - math.Floor(a/b)*b}
				}
			case lexer.TokenPower:
				return &ast.NumeralExpression{Value: math.Pow(a, b)}
			default:
				return exp
			}
		}
		return exp
	}
}

func optimizeNot(exp *ast.UnaryOperatorExpression) ast.Expression {
	if isTrue(exp.Expression) {
		return &ast.BooleanExpression{Value: false}
	} else if isFalse(exp.Expression) {
		return &ast.BooleanExpression{Value: true}
	}
	return exp
}

func optimizeBitNot(exp *ast.UnaryOperatorExpression) ast.Expression {
	switch e := exp.Expression.(type) {
	case *ast.NumeralExpression:
		if math.Trunc(e.Value) != e.Value {
			// todo: return an error instead of panic
			panic("operand has a non-integer value in bitwise NOT operation")
		}
		n := int64(e.Value)
		return &ast.NumeralExpression{Value: float64(^n)}
	default:
		return exp
	}
	return exp
}

func optimizeUnaryMinus(exp *ast.UnaryOperatorExpression) ast.Expression {
	switch e := exp.Expression.(type) {
	case *ast.NumeralExpression:
		return &ast.NumeralExpression{Value: -e.Value}
	default:
		return exp
	}
}
