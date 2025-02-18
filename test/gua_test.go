package test

import (
	_ "embed"
	"testing"

	"github.com/stretchr/testify/suite"

	"lua-interpreter/internal/interpreter"
)

type ParserSuite struct {
	suite.Suite
}

var (
	//go:embed "testdata/loops.lua"
	loopsLua string
	//go:embed "testdata/if.lua"
	ifLua string
	//go:embed "testdata/factorial.lua"
	factorialLua string
)

func TestParserSuite(t *testing.T) {
	suite.Run(t, new(ParserSuite))
}

func (s *ParserSuite) TestLoopStatements() {
	v, err := interpreter.Eval(loopsLua)
	s.NoError(err, "should not return an error")
	s.IsType(float64(0), v, "should return a float64 value")
	s.Equal(float64(88), v, "should return the expected value")
}

func (s *ParserSuite) TestIfStatement() {
	v, err := interpreter.Eval(ifLua)
	s.NoError(err, "should not return an error")
	s.IsType(float64(0), v, "should return a float64 value")
	s.Equal(float64(41), v, "should return the expected value")
}

func (s *ParserSuite) TestRecursiveFactorial() {
	v, err := interpreter.Eval(factorialLua)
	s.NoError(err, "should not return an error")
	s.IsType(float64(0), v, "should return a float64 value")
	s.Equal(float64(120), v, "should return the expected value")
}
