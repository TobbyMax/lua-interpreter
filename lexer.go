package main

import (
	"strings"
	"unicode"
)

type TokenType string

const (
	TokNumber TokenType = "Number"
	TokIdent  TokenType = "Ident"
	TokOp     TokenType = "Operator"
	TokPrint  TokenType = "Print"
	TokIf     TokenType = "If"
	TokThen   TokenType = "Then"
	TokEnd    TokenType = "End"
	TokAssign TokenType = "Assign"
	TokEOF    TokenType = "EOF"
)

type Token struct {
	Type  TokenType
	Value string
}

func Lex(input string) []Token {
	var tokens []Token
	i := 0
	for i < len(input) {
		ch := input[i]
		switch {
		case unicode.IsSpace(rune(ch)):
			i++
		case unicode.IsLetter(rune(ch)):
			start := i
			for i < len(input) && (unicode.IsLetter(rune(input[i])) || unicode.IsDigit(rune(input[i]))) {
				i++
			}
			word := input[start:i]
			switch word {
			case "print":
				tokens = append(tokens, Token{TokPrint, word})
			case "if":
				tokens = append(tokens, Token{TokIf, word})
			case "then":
				tokens = append(tokens, Token{TokThen, word})
			case "end":
				tokens = append(tokens, Token{TokEnd, word})
			default:
				tokens = append(tokens, Token{TokIdent, word})
			}
		case unicode.IsDigit(rune(ch)):
			start := i
			for i < len(input) && unicode.IsDigit(rune(input[i])) {
				i++
			}
			tokens = append(tokens, Token{TokNumber, input[start:i]})
		case ch == '=':
			tokens = append(tokens, Token{TokAssign, string(ch)})
			i++
		case strings.ContainsRune("+-*/()", rune(ch)):
			tokens = append(tokens, Token{TokOp, string(ch)})
			i++
		default:
			i++
		}
	}
	tokens = append(tokens, Token{TokEOF, ""})
	return tokens
}
