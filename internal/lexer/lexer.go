package lexer

import (
	"strings"
	"unicode"
)

type TokenType int

const (
	TokenNumber TokenType = iota
	TokenIdent
	TokenOp
	TokenPrint
	TokenIf
	TokenThen
	TokenEnd
	TokenAssign
	TokenEOF
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
				tokens = append(tokens, Token{TokenPrint, word})
			case "if":
				tokens = append(tokens, Token{TokenIf, word})
			case "then":
				tokens = append(tokens, Token{TokenThen, word})
			case "end":
				tokens = append(tokens, Token{TokenEnd, word})
			default:
				tokens = append(tokens, Token{TokenIdent, word})
			}
		case unicode.IsDigit(rune(ch)):
			start := i
			for i < len(input) && unicode.IsDigit(rune(input[i])) {
				i++
			}
			tokens = append(tokens, Token{TokenNumber, input[start:i]})
		case ch == '=':
			tokens = append(tokens, Token{TokenAssign, string(ch)})
			i++
		case strings.ContainsRune("+-*/()", rune(ch)):
			tokens = append(tokens, Token{TokenOp, string(ch)})
			i++
		default:
			i++
		}
	}
	tokens = append(tokens, Token{TokenEOF, ""})
	return tokens
}
