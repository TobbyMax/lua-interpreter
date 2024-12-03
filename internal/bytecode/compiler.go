package bytecode

import (
	"errors"
	"strconv"

	"lua-interpreter/internal/lexer"
)

func Compile(tokens []lexer.Token) ([]Instruction, error) {
	pos := 0
	next := func() lexer.Token {
		if pos >= len(tokens) {
			return lexer.Token{Type: lexer.TokEOF}
		}
		tok := tokens[pos]
		pos++
		return tok
	}
	peek := func() lexer.Token {
		if pos >= len(tokens) {
			return lexer.Token{Type: lexer.TokEOF}
		}
		return tokens[pos]
	}

	var instrs []Instruction

	tok := next()
	switch tok.Type {
	case lexer.TokIdent:
		if peek().Type == lexer.TokAssign {
			next()
			expr := next()
			switch expr.Type {
			case lexer.TokNumber:
				val, _ := strconv.ParseFloat(expr.Value, 64)
				instrs = append(instrs, Instruction{Op: OpPushConst, Value: val})
			case lexer.TokIdent:
				instrs = append(instrs, Instruction{Op: OpLoadVar, Arg: expr.Value})
			default:
				return nil, errors.New("unexpected expression in assignment")
			}
			instrs = append(instrs, Instruction{Op: OpStoreVar, Arg: tok.Value})
		} else {
			return nil, errors.New("unexpected identifier without assignment")
		}
	case lexer.TokPrint:
		expr := next()
		switch expr.Type {
		case lexer.TokNumber:
			val, _ := strconv.ParseFloat(expr.Value, 64)
			instrs = append(instrs, Instruction{Op: OpPushConst, Value: val})
		case lexer.TokIdent:
			instrs = append(instrs, Instruction{Op: OpLoadVar, Arg: expr.Value})
		default:
			return nil, errors.New("invalid print argument")
		}
		instrs = append(instrs, Instruction{Op: OpPrint})
	default:
		return nil, errors.New("unexpected token: " + tok.Value)
	}

	return instrs, nil
}
