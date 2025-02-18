package parser_test

import (
	"testing"

	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	"lua-interpreter/internal/ast"
	"lua-interpreter/internal/lexer"
	"lua-interpreter/internal/parser"
	"lua-interpreter/internal/parser/mock"
)

type ParserSuite struct {
	suite.Suite

	scanner *mock.Mockscanner
	parser  *parser.Parser
}

func TestParserSuite(t *testing.T) {
	suite.Run(t, new(ParserSuite))
}

func (s *ParserSuite) SetupTest() {
	ctrl := gomock.NewController(s.T())
	s.scanner = mock.NewMockscanner(ctrl)

	s.parser = parser.New(s.scanner)
}

func (s *ParserSuite) TestParseLocalVarDeclaration() {
	s.scanner.EXPECT().NextToken().Return(lexer.Token{Type: lexer.TokenKeywordLocal, Value: "local"}).Times(1)
	s.scanner.EXPECT().NextToken().Return(lexer.Token{Type: lexer.TokenIdentifier, Value: "a"}).Times(1)
	s.scanner.EXPECT().NextToken().Return(lexer.Token{Type: lexer.TokenAssign, Value: "="}).Times(1)
	s.scanner.EXPECT().NextToken().Return(lexer.Token{Type: lexer.TokenNumeral, Value: "10"}).Times(1)
	s.scanner.EXPECT().NextToken().Return(lexer.Token{Type: lexer.TokenEOF, Value: ""}).Times(1)

	block, err := s.parser.Parse()
	s.NoError(err)
	s.Len(block.Statements, 1)
	s.IsType(&ast.LocalVarDeclaration{}, block.Statements[0])
	localVarDecl := block.Statements[0].(*ast.LocalVarDeclaration)
	s.Len(localVarDecl.Vars, 1)
	s.Len(localVarDecl.Exps, 1)
	s.Equal("a", localVarDecl.Vars[0])
	s.IsType(&ast.NumeralExpression{}, localVarDecl.Exps[0])
	s.Equal(10.0, localVarDecl.Exps[0].(*ast.NumeralExpression).Value)
}

func (s *ParserSuite) TestParseFunctionCall() {
	s.scanner.EXPECT().NextToken().Return(lexer.Token{Type: lexer.TokenIdentifier, Value: "print"}).Times(1)
	s.scanner.EXPECT().NextToken().Return(lexer.Token{Type: lexer.TokenLiteralString, Value: "Hello, World!"}).Times(1)
	s.scanner.EXPECT().NextToken().Return(lexer.Token{Type: lexer.TokenEOF, Value: ""}).Times(1)

	block, err := s.parser.Parse()
	s.NoError(err)
	s.Len(block.Statements, 1)
	s.IsType(&ast.FunctionCall{}, block.Statements[0])
	funcCall := block.Statements[0].(*ast.FunctionCall)
	s.IsType(&ast.NameVar{}, funcCall.PrefixExp)
	s.Equal("print", funcCall.PrefixExp.(*ast.NameVar).Name)
	s.IsType(&ast.LiteralString{}, funcCall.Args)
	s.Equal("Hello, World!", funcCall.Args.(*ast.LiteralString).Value)
}

func (s *ParserSuite) TestParseReturnStatement() {
	s.scanner.EXPECT().NextToken().Return(lexer.Token{Type: lexer.TokenKeywordReturn, Value: "return"}).Times(1)
	s.scanner.EXPECT().NextToken().Return(lexer.Token{Type: lexer.TokenNumeral, Value: "42"}).Times(1)
	s.scanner.EXPECT().NextToken().Return(lexer.Token{Type: lexer.TokenEOF, Value: ""}).Times(1)

	block, err := s.parser.Parse()
	s.NoError(err)
	s.Len(block.Statements, 0)
	s.NotNil(block.ReturnStatement)
	s.Len(block.ReturnStatement.Expressions, 1)
	s.IsType(&ast.NumeralExpression{}, block.ReturnStatement.Expressions[0])
	s.Equal(42.0, block.ReturnStatement.Expressions[0].(*ast.NumeralExpression).Value)
}

func (s *ParserSuite) TestParseFunction() {
	s.scanner.EXPECT().NextToken().Return(lexer.Token{Type: lexer.TokenKeywordFunction, Value: "function"}).Times(1)
	s.scanner.EXPECT().NextToken().Return(lexer.Token{Type: lexer.TokenIdentifier, Value: "myFunc"}).Times(1)
	s.scanner.EXPECT().NextToken().Return(lexer.Token{Type: lexer.TokenLeftParen, Value: "("}).Times(1)
	s.scanner.EXPECT().NextToken().Return(lexer.Token{Type: lexer.TokenRightParen, Value: ")"}).Times(1)
	s.scanner.EXPECT().NextToken().Return(lexer.Token{Type: lexer.TokenKeywordEnd, Value: "end"}).Times(1)
	s.scanner.EXPECT().NextToken().Return(lexer.Token{Type: lexer.TokenEOF, Value: ""}).Times(1)

	block, err := s.parser.Parse()
	s.NoError(err)
	s.Len(block.Statements, 1)
	s.IsType(&ast.Function{}, block.Statements[0])
	s.Equal("myFunc", block.Statements[0].(*ast.Function).FunctionName.Name)
	s.Len(block.Statements[0].(*ast.Function).FuncBody.Block.Statements, 0)
	s.Nil(block.ReturnStatement)
}

func (s *ParserSuite) TestParseTableConstructor() {

	s.scanner.EXPECT().NextToken().Return(lexer.Token{Type: lexer.TokenIdentifier, Value: "myTable"}).Times(1)
	s.scanner.EXPECT().NextToken().Return(lexer.Token{Type: lexer.TokenAssign, Value: "="}).Times(1)
	s.scanner.EXPECT().NextToken().Return(lexer.Token{Type: lexer.TokenLeftBrace, Value: "{"}).Times(1)
	s.scanner.EXPECT().NextToken().Return(lexer.Token{Type: lexer.TokenIdentifier, Value: "key"}).Times(1)
	s.scanner.EXPECT().NextToken().Return(lexer.Token{Type: lexer.TokenAssign, Value: "="}).Times(1)
	s.scanner.EXPECT().NextToken().Return(lexer.Token{Type: lexer.TokenNumeral, Value: "42"}).Times(1)
	s.scanner.EXPECT().NextToken().Return(lexer.Token{Type: lexer.TokenRightBrace, Value: "}"}).Times(1)
	s.scanner.EXPECT().NextToken().Return(lexer.Token{Type: lexer.TokenEOF, Value: ""}).Times(1)

	block, err := s.parser.Parse()
	s.NoError(err)
	s.Len(block.Statements, 1)
	s.IsType(&ast.Assignment{}, block.Statements[0])
	assignment := block.Statements[0].(*ast.Assignment)
	s.Len(assignment.Vars, 1)
	s.IsType(&ast.NameVar{}, assignment.Vars[0])
	s.Equal("myTable", assignment.Vars[0].(*ast.NameVar).Name)
	s.Len(assignment.Exps, 1)
	s.IsType(&ast.TableConstructorExpression{}, assignment.Exps[0])
	tableExpr := assignment.Exps[0].(*ast.TableConstructorExpression)
	s.Len(tableExpr.Fields, 1)
	field := tableExpr.Fields[0].(*ast.NameField)
	s.Equal("key", field.Name)
	s.IsType(&ast.NumeralExpression{}, field.Value)
	s.Equal(42.0, field.Value.(*ast.NumeralExpression).Value)
}

func (s *ParserSuite) TestParseArithmeticOptimizations() {
	s.scanner.EXPECT().NextToken().Return(lexer.Token{Type: lexer.TokenIdentifier, Value: "result"}).Times(1)
	s.scanner.EXPECT().NextToken().Return(lexer.Token{Type: lexer.TokenAssign, Value: "="}).Times(1)
	s.scanner.EXPECT().NextToken().Return(lexer.Token{Type: lexer.TokenNumeral, Value: "5"}).Times(1)
	s.scanner.EXPECT().NextToken().Return(lexer.Token{Type: lexer.TokenPlus, Value: "+"}).Times(1)
	s.scanner.EXPECT().NextToken().Return(lexer.Token{Type: lexer.TokenNumeral, Value: "3"}).Times(1)
	s.scanner.EXPECT().NextToken().Return(lexer.Token{Type: lexer.TokenMult, Value: "*"}).Times(1)
	s.scanner.EXPECT().NextToken().Return(lexer.Token{Type: lexer.TokenNumeral, Value: "2"}).Times(1)
	s.scanner.EXPECT().NextToken().Return(lexer.Token{Type: lexer.TokenEOF, Value: ""}).Times(1)

	block, err := s.parser.Parse()
	s.NoError(err)
	s.Len(block.Statements, 1)
	s.IsType(&ast.Assignment{}, block.Statements[0])
	assignment := block.Statements[0].(*ast.Assignment)
	s.Len(assignment.Vars, 1)
	s.IsType(&ast.NameVar{}, assignment.Vars[0])
	s.Equal("result", assignment.Vars[0].(*ast.NameVar).Name)
	s.Len(assignment.Exps, 1)
	s.IsType(&ast.NumeralExpression{}, assignment.Exps[0])
	s.Equal(11.0, assignment.Exps[0].(*ast.NumeralExpression).Value)
}

func (s *ParserSuite) TestParseIfStatement() {
	s.scanner.EXPECT().NextToken().Return(lexer.Token{Type: lexer.TokenKeywordIf, Value: "if"}).Times(1)
	s.scanner.EXPECT().NextToken().Return(lexer.Token{Type: lexer.TokenIdentifier, Value: "x"}).Times(1)
	s.scanner.EXPECT().NextToken().Return(lexer.Token{Type: lexer.TokenEqual, Value: "=="}).Times(1)
	s.scanner.EXPECT().NextToken().Return(lexer.Token{Type: lexer.TokenNumeral, Value: "10"}).Times(1)
	s.scanner.EXPECT().NextToken().Return(lexer.Token{Type: lexer.TokenKeywordThen, Value: "then"}).Times(1)
	s.scanner.EXPECT().NextToken().Return(lexer.Token{Type: lexer.TokenIdentifier, Value: "print"}).Times(1)
	s.scanner.EXPECT().NextToken().Return(lexer.Token{Type: lexer.TokenLeftParen, Value: "("}).Times(1)
	s.scanner.EXPECT().NextToken().Return(lexer.Token{Type: lexer.TokenLiteralString, Value: "x is 10"}).Times(1)
	s.scanner.EXPECT().NextToken().Return(lexer.Token{Type: lexer.TokenRightParen, Value: ")"}).Times(1)
	s.scanner.EXPECT().NextToken().Return(lexer.Token{Type: lexer.TokenKeywordEnd, Value: "end"}).Times(1)
	s.scanner.EXPECT().NextToken().Return(lexer.Token{Type: lexer.TokenEOF, Value: ""}).Times(1)

	block, err := s.parser.Parse()
	s.NoError(err)
	s.Len(block.Statements, 1)
	s.IsType(&ast.If{}, block.Statements[0])
	ifStmt := block.Statements[0].(*ast.If)
	s.Len(ifStmt.Exps, 1)
	s.Len(ifStmt.Blocks, 1)
	cond := ifStmt.Exps[0]
	s.IsType(&ast.BinaryOperatorExpression{}, cond)
	binaryExpr := cond.(*ast.BinaryOperatorExpression)
	s.Equal("x", binaryExpr.Left.(*ast.NameVar).Name)
	s.Equal("==", binaryExpr.Operator.Value)
	s.IsType(&ast.NumeralExpression{}, binaryExpr.Right)
	s.Equal(float64(10), binaryExpr.Right.(*ast.NumeralExpression).Value)
	s.Len(ifStmt.Blocks[0].Statements, 1)
}

func (s *ParserSuite) TestParseWhileLoop() {
	s.scanner.EXPECT().NextToken().Return(lexer.Token{Type: lexer.TokenKeywordWhile, Value: "while"}).Times(1)
	s.scanner.EXPECT().NextToken().Return(lexer.Token{Type: lexer.TokenIdentifier, Value: "x"}).Times(1)
	s.scanner.EXPECT().NextToken().Return(lexer.Token{Type: lexer.TokenLess, Value: "<"}).Times(1)
	s.scanner.EXPECT().NextToken().Return(lexer.Token{Type: lexer.TokenNumeral, Value: "10"}).Times(1)
	s.scanner.EXPECT().NextToken().Return(lexer.Token{Type: lexer.TokenKeywordDo, Value: "do"}).Times(1)
	s.scanner.EXPECT().NextToken().Return(lexer.Token{Type: lexer.TokenIdentifier, Value: "x"}).Times(1)
	s.scanner.EXPECT().NextToken().Return(lexer.Token{Type: lexer.TokenAssign, Value: "="}).Times(1)
	s.scanner.EXPECT().NextToken().Return(lexer.Token{Type: lexer.TokenIdentifier, Value: "x"}).Times(1)
	s.scanner.EXPECT().NextToken().Return(lexer.Token{Type: lexer.TokenPlus, Value: "+"}).Times(1)
	s.scanner.EXPECT().NextToken().Return(lexer.Token{Type: lexer.TokenNumeral, Value: "1"}).Times(1)
	s.scanner.EXPECT().NextToken().Return(lexer.Token{Type: lexer.TokenKeywordEnd, Value: "end"}).Times(1)
	s.scanner.EXPECT().NextToken().Return(lexer.Token{Type: lexer.TokenEOF, Value: ""}).Times(1)

	block, err := s.parser.Parse()
	s.NoError(err)
	s.Len(block.Statements, 1)
	s.IsType(&ast.While{}, block.Statements[0])
	whileStmt := block.Statements[0].(*ast.While)
	s.IsType(&ast.BinaryOperatorExpression{}, whileStmt.Exp)
	binaryExpr := whileStmt.Exp.(*ast.BinaryOperatorExpression)
	s.Equal("x", binaryExpr.Left.(*ast.NameVar).Name)
	s.Equal("<", binaryExpr.Operator.Value)
	s.IsType(&ast.NumeralExpression{}, binaryExpr.Right)
	s.Equal(float64(10), binaryExpr.Right.(*ast.NumeralExpression).Value)
	s.Len(whileStmt.Block.Statements, 1)
	assignment := whileStmt.Block.Statements[0]
	s.IsType(&ast.Assignment{}, assignment)
	assignmentStmt := assignment.(*ast.Assignment)
	s.Len(assignmentStmt.Vars, 1)
	s.IsType(&ast.NameVar{}, assignmentStmt.Vars[0])
	s.Equal("x", assignmentStmt.Vars[0].(*ast.NameVar).Name)
	s.Len(assignmentStmt.Exps, 1)
	s.IsType(&ast.BinaryOperatorExpression{}, assignmentStmt.Exps[0])
	binaryExp := assignmentStmt.Exps[0].(*ast.BinaryOperatorExpression)
	s.Equal("x", binaryExp.Left.(*ast.NameVar).Name)
	s.Equal("+", binaryExp.Operator.Value)
	s.IsType(&ast.NumeralExpression{}, binaryExp.Right)
	s.Equal(float64(1), binaryExp.Right.(*ast.NumeralExpression).Value)
}

func (s *ParserSuite) TestParseForLoop() {
	s.scanner.EXPECT().NextToken().Return(lexer.Token{Type: lexer.TokenKeywordFor, Value: "for"}).Times(1)
	s.scanner.EXPECT().NextToken().Return(lexer.Token{Type: lexer.TokenIdentifier, Value: "i"}).Times(1)
	s.scanner.EXPECT().NextToken().Return(lexer.Token{Type: lexer.TokenAssign, Value: "="}).Times(1)
	s.scanner.EXPECT().NextToken().Return(lexer.Token{Type: lexer.TokenNumeral, Value: "1"}).Times(1)
	s.scanner.EXPECT().NextToken().Return(lexer.Token{Type: lexer.TokenComma, Value: ","}).Times(1)
	s.scanner.EXPECT().NextToken().Return(lexer.Token{Type: lexer.TokenNumeral, Value: "10"}).Times(1)
	s.scanner.EXPECT().NextToken().Return(lexer.Token{Type: lexer.TokenKeywordDo, Value: "do"}).Times(1)
	s.scanner.EXPECT().NextToken().Return(lexer.Token{Type: lexer.TokenIdentifier, Value: "print"}).Times(1)
	s.scanner.EXPECT().NextToken().Return(lexer.Token{Type: lexer.TokenLeftParen, Value: "("}).Times(1)
	s.scanner.EXPECT().NextToken().Return(lexer.Token{Type: lexer.TokenIdentifier, Value: "i"}).Times(1)
	s.scanner.EXPECT().NextToken().Return(lexer.Token{Type: lexer.TokenRightParen, Value: ")"}).Times(1)
	s.scanner.EXPECT().NextToken().Return(lexer.Token{Type: lexer.TokenKeywordEnd, Value: "end"}).Times(1)
	s.scanner.EXPECT().NextToken().Return(lexer.Token{Type: lexer.TokenEOF, Value: ""}).Times(1)

	block, err := s.parser.Parse()
	s.NoError(err)
	s.Len(block.Statements, 1)
	s.IsType(&ast.For{}, block.Statements[0])
	forStmt := block.Statements[0].(*ast.For)
	s.Equal("i", forStmt.Name)
	s.IsType(&ast.NumeralExpression{}, forStmt.Init)
	s.Equal(float64(1), forStmt.Init.(*ast.NumeralExpression).Value)
	s.IsType(&ast.NumeralExpression{}, forStmt.Limit)
	s.Equal(float64(10), forStmt.Limit.(*ast.NumeralExpression).Value)
	stat := forStmt.Block.Statements[0]
	s.IsType(&ast.FunctionCall{}, stat)
	funcCall := stat.(*ast.FunctionCall)
	s.IsType(&ast.NameVar{}, funcCall.PrefixExp)
	s.Equal("print", funcCall.PrefixExp.(*ast.NameVar).Name)
	s.IsType([]ast.Expression{}, funcCall.Args)
	s.Len(funcCall.Args, 1)
	s.IsType(&ast.NameVar{}, funcCall.Args.([]ast.Expression)[0])
	s.Equal("i", funcCall.Args.([]ast.Expression)[0].(*ast.NameVar).Name)
}

func (s *ParserSuite) TestParseErrorHandling() {
	s.scanner.EXPECT().NextToken().Return(lexer.Token{Type: lexer.TokenKeywordLocal, Value: "local"}).Times(1)
	s.scanner.EXPECT().NextToken().Return(lexer.Token{Type: lexer.TokenIdentifier, Value: "a"}).Times(1)
	s.scanner.EXPECT().NextToken().Return(lexer.Token{Type: lexer.TokenEqual, Value: "=="}).Times(1)

	_, err := s.parser.Parse()
	s.Error(err)
	s.Contains(err.Error(), "unexpected token: ==")
}

func (s *ParserSuite) TestParseEmptyBlock() {
	s.scanner.EXPECT().NextToken().Return(lexer.Token{Type: lexer.TokenEOF, Value: ""}).Times(1)

	block, err := s.parser.Parse()
	s.NoError(err)
	s.Len(block.Statements, 0)
	s.Nil(block.ReturnStatement)
}

func (s *ParserSuite) TestParseRepeatLoop() {
	s.scanner.EXPECT().NextToken().Return(lexer.Token{Type: lexer.TokenKeywordRepeat, Value: "repeat"}).Times(1)
	s.scanner.EXPECT().NextToken().Return(lexer.Token{Type: lexer.TokenIdentifier, Value: "x"}).Times(1)
	s.scanner.EXPECT().NextToken().Return(lexer.Token{Type: lexer.TokenAssign, Value: "="}).Times(1)
	s.scanner.EXPECT().NextToken().Return(lexer.Token{Type: lexer.TokenNumeral, Value: "0"}).Times(1)
	s.scanner.EXPECT().NextToken().Return(lexer.Token{Type: lexer.TokenKeywordUntil, Value: "until"}).Times(1)
	s.scanner.EXPECT().NextToken().Return(lexer.Token{Type: lexer.TokenIdentifier, Value: "x"}).Times(1)
	s.scanner.EXPECT().NextToken().Return(lexer.Token{Type: lexer.TokenEqual, Value: "=="}).Times(1)
	s.scanner.EXPECT().NextToken().Return(lexer.Token{Type: lexer.TokenNumeral, Value: "10"}).Times(1)
	s.scanner.EXPECT().NextToken().Return(lexer.Token{Type: lexer.TokenEOF, Value: ""}).Times(1)

	block, err := s.parser.Parse()
	s.NoError(err)
	s.Len(block.Statements, 1)
	s.IsType(&ast.Repeat{}, block.Statements[0])
	repeatStmt := block.Statements[0].(*ast.Repeat)
	s.Len(repeatStmt.Block.Statements, 1)
	stat := repeatStmt.Block.Statements[0]
	s.IsType(&ast.Assignment{}, stat)
	assignment := stat.(*ast.Assignment)
	s.Len(assignment.Vars, 1)
}

func (s *ParserSuite) TestParseLabel() {
	s.scanner.EXPECT().NextToken().Return(lexer.Token{Type: lexer.TokenDoubleColon, Value: "::"}).Times(1)
	s.scanner.EXPECT().NextToken().Return(lexer.Token{Type: lexer.TokenIdentifier, Value: "myLabel"}).Times(1)
	s.scanner.EXPECT().NextToken().Return(lexer.Token{Type: lexer.TokenDoubleColon, Value: "::"}).Times(1)
	s.scanner.EXPECT().NextToken().Return(lexer.Token{Type: lexer.TokenEOF, Value: ""}).Times(1)

	block, err := s.parser.Parse()
	s.NoError(err)
	s.Len(block.Statements, 1)
	s.IsType(&ast.Label{}, block.Statements[0])
	label := block.Statements[0].(*ast.Label)
	s.Equal("myLabel", label.Name)
}

func (s *ParserSuite) TestParseGoto() {
	s.scanner.EXPECT().NextToken().Return(lexer.Token{Type: lexer.TokenKeywordGoTo, Value: "goto"}).Times(1)
	s.scanner.EXPECT().NextToken().Return(lexer.Token{Type: lexer.TokenIdentifier, Value: "myLabel"}).Times(1)
	s.scanner.EXPECT().NextToken().Return(lexer.Token{Type: lexer.TokenEOF, Value: ""}).Times(1)

	block, err := s.parser.Parse()
	s.NoError(err)
	s.Len(block.Statements, 1)
	s.IsType(&ast.Goto{}, block.Statements[0])
	gotoStmt := block.Statements[0].(*ast.Goto)
	s.Equal("myLabel", gotoStmt.Name)
}
