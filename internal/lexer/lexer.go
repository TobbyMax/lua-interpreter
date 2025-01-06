package lexer

import (
	"strings"
	"unicode"
)

type TokenType int

// https://www.lua.org/manual/5.3/manual.html#lua_load:~:text=3.1%20%E2%80%93%20Lexical%20Conventions
const (
	TokenEOF TokenType = iota
	// Operators and delimiters
	TokenPlus
	TokenMinus
	TokenMult
	TokenDiv
	TokenMod
	TokenPower
	TokenHash
	TokenBinAnd
	TokenBinOr
	TokenTilde
	TokenShiftLeft
	TokenShiftRight
	TokenIntDiv
	TokenEqual
	TokenNotEqual
	TokenMoreEqual
	TokenLessEqual
	TokenMore
	TokenLess
	TokenAssign
	TokenLeftParen
	TokenRightParen
	TokenLeftBrace
	TokenRightBrace
	TokenLeftBracket
	TokenRightBracket
	TokenDoubleColon
	TokenColon
	TokenSemiColon
	TokenComma
	TokenDot
	TokenDoubleDot
	TokenTripleDot
	TokenNot
	// Keywords
	TokenKeywordAnd
	TokenKeywordBreak
	TokenKeywordDo
	TokenKeywordElse
	TokenKeywordElseIf
	TokenKeywordEnd
	TokenKeywordFalse
	TokenKeywordFor
	TokenKeywordFunction
	TokenKeywordGoTo
	TokenKeywordIf
	TokenKeywordIn
	TokenKeywordLocal
	TokenKeywordNil
	TokenKeywordNot
	TokenKeywordOr
	TokenKeywordRepeat
	TokenKeywordReturn
	TokenKeywordThen
	TokenKeywordTrue
	TokenKeywordUntil
	TokenKeywordWhile
	// Other
	TokenIdentifier
	TokenLiteralString
	TokenNumeral
	TokenComment
	TokenError
)

const (
	IdentifierStartSymbols = "_abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	IdentifierSymbols      = "_abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	HexNumberSymbols       = ".0123456789abcdefABCDEF"
)

// for debugging
func (t TokenType) String() string {
	switch t {
	case TokenEOF:
		return "EOF"
	case TokenPlus:
		return "+"
	case TokenMinus:
		return "-"
	case TokenMult:
		return "*"
	case TokenDiv:
		return "/"
	case TokenMod:
		return "%"
	case TokenPower:
		return "^"
	case TokenHash:
		return "#"
	case TokenBinAnd:
		return "&"
	case TokenBinOr:
		return "|"
	case TokenTilde:
		return "~"
	case TokenShiftLeft:
		return "<<"
	case TokenShiftRight:
		return ">>"
	case TokenIntDiv:
		return "//"
	case TokenEqual:
		return "=="
	case TokenNotEqual:
		return "~="
	case TokenMoreEqual:
		return ">="
	case TokenLessEqual:
		return "<="
	case TokenMore:
		return ">"
	case TokenLess:
		return "<"
	case TokenAssign:
		return "="
	case TokenLeftParen:
		return "("
	case TokenRightParen:
		return ")"
	case TokenLeftBrace:
		return "{"
	case TokenRightBrace:
		return "}"
	case TokenLeftBracket:
		return "["
	case TokenRightBracket:
		return "]"
	case TokenDoubleColon:
		return "::"
	case TokenColon:
		return ":"
	case TokenSemiColon:
		return ";"
	case TokenComma:
		return ","
	case TokenDot:
		return "."
	case TokenDoubleDot:
		return ".."
	case TokenTripleDot:
		return "..."
	case TokenKeywordAnd:
		return "and"
	case TokenKeywordBreak:
		return "break"
	case TokenKeywordDo:
		return "do"
	case TokenKeywordElse:
		return "else"
	case TokenKeywordElseIf:
		return "elseif"
	case TokenKeywordEnd:
		return "end"
	case TokenKeywordFalse:
		return "false"
	case TokenKeywordFor:
		return "for"
	case TokenKeywordFunction:
		return "function"
	case TokenKeywordGoTo:
		return "goto"
	case TokenKeywordIf:
		return "if"
	case TokenKeywordIn:
		return "in"
	case TokenKeywordLocal:
		return "local"
	case TokenKeywordNil:
		return "nil"
	case TokenKeywordNot:
		return "not"
	case TokenKeywordOr:
		return "or"
	case TokenKeywordRepeat:
		return "repeat"
	case TokenKeywordReturn:
		return "return"
	case TokenKeywordThen:
		return "then"
	case TokenKeywordTrue:
		return "true"
	case TokenKeywordUntil:
		return "until"
	case TokenKeywordWhile:
		return "while"
	case TokenIdentifier:
		return "identifier"
	case TokenLiteralString:
		return "string"
	case TokenNumeral:
		return "number"
	case TokenComment:
		return "comment"
	case TokenError:
		return "error"
	default:
		return "unknown"
	}
}

var keywords = map[string]TokenType{
	"and":      TokenKeywordAnd,
	"break":    TokenKeywordBreak,
	"do":       TokenKeywordDo,
	"else":     TokenKeywordElse,
	"elseif":   TokenKeywordElseIf,
	"end":      TokenKeywordEnd,
	"false":    TokenKeywordFalse,
	"for":      TokenKeywordFor,
	"function": TokenKeywordFunction,
	"goto":     TokenKeywordGoTo,
	"if":       TokenKeywordIf,
	"in":       TokenKeywordIn,
	"local":    TokenKeywordLocal,
	"nil":      TokenKeywordNil,
	"not":      TokenKeywordNot,
	"or":       TokenKeywordOr,
	"repeat":   TokenKeywordRepeat,
	"return":   TokenKeywordReturn,
	"then":     TokenKeywordThen,
	"true":     TokenKeywordTrue,
	"until":    TokenKeywordUntil,
	"while":    TokenKeywordWhile,
}

type Token struct {
	Type  TokenType
	Value string
}

type Lexer struct {
	input    string
	position int
}

func NewLexer(input string) *Lexer {
	return &Lexer{
		input:    input,
		position: 0,
	}
}

func (l *Lexer) HasNext() bool {
	return l.position < len(l.input)
}

func (l *Lexer) NextToken() Token {
	var token Token
	input := l.input
	i := l.position
	for i < len(input) && unicode.IsSpace(rune(input[i])) {
		i++
	}
	if i < len(input) {
		ch := input[i]
		switch {
		case strings.ContainsRune(IdentifierStartSymbols, rune(ch)):
			start := i
			for i < len(input) && strings.ContainsRune(IdentifierSymbols, rune(input[i])) {
				i++
			}
			word := input[start:i]
			if _, ok := keywords[word]; ok {
				token = Token{keywords[word], word}
			} else {
				token = Token{TokenIdentifier, word}
			}
		case unicode.IsDigit(rune(ch)):
			start := i
			isHex := false
			isFloat := false
			isFailed := false
			if i+1 < len(input) && input[i] == '0' && input[i+1] == 'x' {
				isHex = true
				i += 2
			}
			for i < len(input) {
				if input[i] == '.' {
					if isFloat {
						isFailed = true
						break
					}
					isFloat = true
					i++
					continue
				}
				if isHex && strings.ContainsRune(HexNumberSymbols, rune(input[i])) {
					i++
				} else if unicode.IsDigit(rune(input[i])) {
					i++
				} else {
					break
				}
			}
			if isFailed {
				token = Token{TokenError, input[start:i]}
			} else {
				token = Token{TokenNumeral, input[start:i]}
			}
		case ch == '"':
			start := i
			i++
			for i < len(input) && input[i] != '"' {
				if input[i] == '\\' {
					if i+1 < len(input) && input[i+1] == '"' {
						i++
					}
				}
				i++
			}
			if i < len(input) {
				i++
				token = Token{TokenLiteralString, input[start:i]}
			} else {
				token = Token{TokenError, input[start:i]}
			}
		case ch == '\'':
			start := i
			i++
			for i < len(input) && input[i] != '\'' {
				if i+1 < len(input) && input[i+1] == '"' {
					i++
				}
				i++
			}
			if i < len(input) {
				i++
				token = Token{TokenLiteralString, input[start:i]}
			} else {
				token = Token{TokenError, input[start:i]}
			}
		case ch == '+':
			token = Token{TokenPlus, string(ch)}
			i++
		case ch == '-':
			if i+1 < len(input) && input[i+1] == '-' {
				start := i
				if i+2 < len(input) && input[i+2] == '[' {
					j := i + 3
					equalCount := 0
					for j < len(input) && input[j] == '=' {
						equalCount++
						j++
					}
					if j < len(input) && input[j] == '[' {
						i = j
						for i < len(input) {
							if input[i] == ']' {
								count := 0
								for i+1 < len(input) && input[i+1] == '=' {
									count++
									i++
								}
								if count == equalCount && i+1 < len(input) && input[i+1] == ']' {
									i += 2
									break
								}
							}
							i++
						}
					}
				}
				if i >= len(input) {
					token = Token{TokenError, input[start:i]}
				} else if i > start {
					token = Token{TokenComment, input[start:i]}
				} else {
					for i+1 < len(input) && input[i+1] != '\n' {
						i++
					}
					token = Token{TokenComment, input[start:i]}
				}
			}
			token = Token{TokenMinus, string(ch)}
			i++
		case ch == '*':
			token = Token{TokenMult, string(ch)}
			i++
		case ch == '%':
			token = Token{TokenMod, string(ch)}
			i++
		case ch == '^':
			token = Token{TokenPower, string(ch)}
			i++
		case ch == '#':
			token = Token{TokenHash, string(ch)}
			i++
		case ch == '&':
			token = Token{TokenBinAnd, string(ch)}
			i++
		case ch == '|':
			token = Token{TokenBinOr, string(ch)}
			i++
		case ch == '/':
			if i+1 < len(input) && input[i+1] == '/' {
				token = Token{TokenIntDiv, "//"}
				i += 2
			}
			token = Token{TokenDiv, string(ch)}
			i++
		case ch == '=':
			if i+1 < len(input) && input[i+1] == '=' {
				token = Token{TokenEqual, "=="}
				i += 2
			}
			token = Token{TokenAssign, string(ch)}
			i++
		case ch == '<':
			if i+1 < len(input) && input[i+1] == '=' {
				token = Token{TokenLessEqual, "<="}
				i += 2
			}
			if i+1 < len(input) && input[i+1] == '<' {
				token = Token{TokenShiftLeft, "<<"}
				i += 2
			}
			token = Token{TokenLess, string(ch)}
			i++
		case ch == '>':
			if i+1 < len(input) && input[i+1] == '=' {
				token = Token{TokenMoreEqual, ">="}
				i += 2
			}
			if i+1 < len(input) && input[i+1] == '>' {
				token = Token{TokenShiftRight, ">>"}
				i += 2
			}
			token = Token{TokenMore, string(ch)}
			i++
		case ch == '~':
			if i+1 < len(input) && input[i+1] == '=' {
				token = Token{TokenNotEqual, "~="}
				i += 2
			}
			token = Token{TokenTilde, string(ch)}
			i++
		case ch == ':':
			if i+1 < len(input) && input[i+1] == ':' {
				token = Token{TokenDoubleColon, "::"}
				i += 2
			}
			token = Token{TokenColon, string(ch)}
			i++
		case ch == ';':
			token = Token{TokenSemiColon, string(ch)}
			i++
		case ch == ',':
			token = Token{TokenComma, string(ch)}
			i++
		case ch == '.':
			if i+1 < len(input) && input[i+1] == '.' {
				if i+2 < len(input) && input[i+2] == '.' {
					token = Token{TokenTripleDot, "..."}
					i += 3
				}
				token = Token{TokenDoubleDot, ".."}
				i += 2
			}
			if i+1 < len(input) && unicode.IsDigit(rune(input[i+1])) {
				start := i
				i++
				for i < len(input) && unicode.IsDigit(rune(input[i])) {
					i++
				}
				token = Token{TokenNumeral, input[start:i]}
			}
			token = Token{TokenDot, string(ch)}
			i++
		case ch == '(':
			token = Token{TokenLeftParen, string(ch)}
			i++
		case ch == ')':
			token = Token{TokenRightParen, string(ch)}
			i++
		case ch == '{':
			token = Token{TokenLeftBrace, string(ch)}
			i++
		case ch == '}':
			token = Token{TokenRightBrace, string(ch)}
			i++
		case ch == '[':
			token = Token{TokenLeftBracket, string(ch)}
			i++
		case ch == ']':
			token = Token{TokenRightBracket, string(ch)}
			i++
		default:
			token = Token{TokenError, string(ch)}
			i++
		}
	} else {
		token = Token{TokenEOF, ""}
	}
	l.position = i
	return token
}
