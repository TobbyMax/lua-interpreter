package lexer_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"lua-interpreter/internal/lexer"
)

type LexerSuite struct {
	suite.Suite

	lexer *lexer.Lexer
}

func TestLexerSuite(t *testing.T) {
	suite.Run(t, new(LexerSuite))
}

func (s *LexerSuite) TestNextToken() {
	tests := []struct {
		input    string
		expected []lexer.Token
	}{
		{
			input: "local a = 10",
			expected: []lexer.Token{
				{Type: lexer.TokenKeywordLocal, Value: "local"},
				{Type: lexer.TokenIdentifier, Value: "a"},
				{Type: lexer.TokenAssign, Value: "="},
				{Type: lexer.TokenNumeral, Value: "10"},
				{Type: lexer.TokenEOF, Value: ""},
			},
		},
		{
			input: "print('Hello, World!')",
			expected: []lexer.Token{
				{Type: lexer.TokenIdentifier, Value: "print"},
				{Type: lexer.TokenLeftParen, Value: "("},
				{Type: lexer.TokenLiteralString, Value: "Hello, World!"},
				{Type: lexer.TokenRightParen, Value: ")"},
				{Type: lexer.TokenEOF, Value: ""},
			},
		},
	}

	for _, test := range tests {
		s.lexer = lexer.NewLexer(test.input)
		for _, expectedToken := range test.expected {
			token := s.lexer.NextToken()
			s.Equal(expectedToken.Type, token.Type)
			s.Equal(expectedToken.Value, token.Value)
		}
	}
}

func (s *LexerSuite) TestNextTokenSingleTokens() {
	tests := []struct {
		input    string
		expected []lexer.TokenType
	}{
		{"?", []lexer.TokenType{lexer.TokenError, lexer.TokenEOF}},
		{"--[=[long comment]=]", []lexer.TokenType{lexer.TokenEOF}},
		{"::", []lexer.TokenType{lexer.TokenDoubleColon, lexer.TokenEOF}},
		{"...", []lexer.TokenType{lexer.TokenTripleDot, lexer.TokenEOF}},
		{`'a\"b\'c'`, []lexer.TokenType{lexer.TokenLiteralString, lexer.TokenEOF}},
		{`"a\"b\'c"`, []lexer.TokenType{lexer.TokenLiteralString, lexer.TokenEOF}},
		{`"a\"b\nc`, []lexer.TokenType{lexer.TokenError, lexer.TokenEOF}},
		{`"abc`, []lexer.TokenType{lexer.TokenError, lexer.TokenEOF}},
		{`'abc`, []lexer.TokenType{lexer.TokenError, lexer.TokenEOF}},
		{"--comment", []lexer.TokenType{lexer.TokenEOF}},
		{"--[fake]comment", []lexer.TokenType{lexer.TokenEOF}},
		{"--[]]comment", []lexer.TokenType{lexer.TokenEOF}},
		{"--[[long comment]====]abc]]", []lexer.TokenType{lexer.TokenEOF}},
		{"--[=[long\ncomment\n]]rfc]=]", []lexer.TokenType{lexer.TokenEOF}},
		{"--[=[broken\ncomment\n]]", []lexer.TokenType{lexer.TokenError}},
		{"0x1A", []lexer.TokenType{lexer.TokenNumeral, lexer.TokenEOF}},
		{"0x1A.2", []lexer.TokenType{lexer.TokenNumeral, lexer.TokenEOF}},
		{"123.456", []lexer.TokenType{lexer.TokenNumeral, lexer.TokenEOF}},
		{"1.2.3", []lexer.TokenType{lexer.TokenError, lexer.TokenNumeral, lexer.TokenEOF}},
		{"1.2tr", []lexer.TokenType{lexer.TokenNumeral, lexer.TokenIdentifier, lexer.TokenEOF}},
		{"[", []lexer.TokenType{lexer.TokenLeftBracket, lexer.TokenEOF}},
		{"]", []lexer.TokenType{lexer.TokenRightBracket, lexer.TokenEOF}},
		{"{", []lexer.TokenType{lexer.TokenLeftBrace, lexer.TokenEOF}},
		{"}", []lexer.TokenType{lexer.TokenRightBrace, lexer.TokenEOF}},
		{"(", []lexer.TokenType{lexer.TokenLeftParen, lexer.TokenEOF}},
		{")", []lexer.TokenType{lexer.TokenRightParen, lexer.TokenEOF}},
		{":", []lexer.TokenType{lexer.TokenColon, lexer.TokenEOF}},
		{".", []lexer.TokenType{lexer.TokenDot, lexer.TokenEOF}},
		{",", []lexer.TokenType{lexer.TokenComma, lexer.TokenEOF}},
		{";", []lexer.TokenType{lexer.TokenSemiColon, lexer.TokenEOF}},
		{"=", []lexer.TokenType{lexer.TokenAssign, lexer.TokenEOF}},
		{"==", []lexer.TokenType{lexer.TokenEqual, lexer.TokenEOF}},
		{"<", []lexer.TokenType{lexer.TokenLess, lexer.TokenEOF}},
		{"<=", []lexer.TokenType{lexer.TokenLessEqual, lexer.TokenEOF}},
		{">", []lexer.TokenType{lexer.TokenMore, lexer.TokenEOF}},
		{">=", []lexer.TokenType{lexer.TokenMoreEqual, lexer.TokenEOF}},
		{"#", []lexer.TokenType{lexer.TokenHash, lexer.TokenEOF}},
		{"~=", []lexer.TokenType{lexer.TokenNotEqual, lexer.TokenEOF}},
		{"not", []lexer.TokenType{lexer.TokenKeywordNot, lexer.TokenEOF}},
		{"goto", []lexer.TokenType{lexer.TokenKeywordGoTo, lexer.TokenEOF}},
		{"repeat", []lexer.TokenType{lexer.TokenKeywordRepeat, lexer.TokenEOF}},
		{"until", []lexer.TokenType{lexer.TokenKeywordUntil, lexer.TokenEOF}},
		{"then", []lexer.TokenType{lexer.TokenKeywordThen, lexer.TokenEOF}},
		{"elseif", []lexer.TokenType{lexer.TokenKeywordElseIf, lexer.TokenEOF}},
		{"else", []lexer.TokenType{lexer.TokenKeywordElse, lexer.TokenEOF}},
		{"end", []lexer.TokenType{lexer.TokenKeywordEnd, lexer.TokenEOF}},
		{"if", []lexer.TokenType{lexer.TokenKeywordIf, lexer.TokenEOF}},
		{"while", []lexer.TokenType{lexer.TokenKeywordWhile, lexer.TokenEOF}},
		{"for", []lexer.TokenType{lexer.TokenKeywordFor, lexer.TokenEOF}},
		{"do", []lexer.TokenType{lexer.TokenKeywordDo, lexer.TokenEOF}},
		{"return", []lexer.TokenType{lexer.TokenKeywordReturn, lexer.TokenEOF}},
		{"break", []lexer.TokenType{lexer.TokenKeywordBreak, lexer.TokenEOF}},
		{"and", []lexer.TokenType{lexer.TokenKeywordAnd, lexer.TokenEOF}},
		{"or", []lexer.TokenType{lexer.TokenKeywordOr, lexer.TokenEOF}},
		{"function", []lexer.TokenType{lexer.TokenKeywordFunction, lexer.TokenEOF}},
		{"true", []lexer.TokenType{lexer.TokenKeywordTrue, lexer.TokenEOF}},
		{"false", []lexer.TokenType{lexer.TokenKeywordFalse, lexer.TokenEOF}},
		{"nil", []lexer.TokenType{lexer.TokenKeywordNil, lexer.TokenEOF}},
		{"..", []lexer.TokenType{lexer.TokenDoubleDot, lexer.TokenEOF}},
		{"+", []lexer.TokenType{lexer.TokenPlus, lexer.TokenEOF}},
		{"-", []lexer.TokenType{lexer.TokenMinus, lexer.TokenEOF}},
		{"*", []lexer.TokenType{lexer.TokenMult, lexer.TokenEOF}},
		{"/", []lexer.TokenType{lexer.TokenDiv, lexer.TokenEOF}},
		{"//", []lexer.TokenType{lexer.TokenIntDiv, lexer.TokenEOF}},
		{"%", []lexer.TokenType{lexer.TokenMod, lexer.TokenEOF}},
		{"^", []lexer.TokenType{lexer.TokenPower, lexer.TokenEOF}},
		{"&", []lexer.TokenType{lexer.TokenBinAnd, lexer.TokenEOF}},
		{"|", []lexer.TokenType{lexer.TokenBinOr, lexer.TokenEOF}},
		{"~", []lexer.TokenType{lexer.TokenTilde, lexer.TokenEOF}},
		{"<<", []lexer.TokenType{lexer.TokenShiftLeft, lexer.TokenEOF}},
		{">>", []lexer.TokenType{lexer.TokenShiftRight, lexer.TokenEOF}},
		{"a", []lexer.TokenType{lexer.TokenIdentifier, lexer.TokenEOF}},
		{" a ", []lexer.TokenType{lexer.TokenIdentifier, lexer.TokenEOF}},
	}

	for _, test := range tests {
		s.lexer = lexer.NewLexer(test.input)
		for _, expected := range test.expected {
			token := s.lexer.NextToken()
			s.Equal(expected, token.Type, "input: %q", test.input)
		}
	}
}
