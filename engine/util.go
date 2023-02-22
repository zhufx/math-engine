package engine

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/shopspring/decimal"
)

// Top level function
// Analytical expression and execution
// err is not nil if an error occurs (including arithmetic runtime errors)
func ParseAndExec(s string) (r decimal.Decimal, err error) {
	toks, err := Parse(s)
	if err != nil {
		return decimal.Zero, err
	}
	ast := NewAST(toks, s)
	if ast.Err != nil {
		return decimal.Zero, ast.Err
	}
	ar := ast.ParseExpression()
	if ast.Err != nil {
		return decimal.Zero, ast.Err
	}
	defer func() {
		if e := recover(); e != nil {
			err = e.(error)
		}
	}()
	return ExprASTResult(ar), err
}

func ErrPos(s string, pos int) string {
	r := strings.Repeat("-", len(s)) + "\n"
	s += "\n"
	for i := 0; i < pos; i++ {
		s += " "
	}
	s += "^\n"
	return r + s + r
}

// the integer power of a number
func Pow(x float64, n float64) float64 {
	return math.Pow(x, n)
}

//
// func expr2Radian(expr ExprAST) float64 {
// 	r := ExprASTResult(expr)
// 	if TrigonometricMode == AngleMode {
// 		r = r / 180 * math.Pi
// 	}
// 	return r
// }

// Float64ToStr float64 -> string
func Float64ToStr(f float64) string {
	return strconv.FormatFloat(f, 'f', -1, 64)
}

// RegFunction is Top level function
// register a new function to use in expressions
// name: be register function name. the same function name only needs to be registered once.
// argc: this is a number of parameter signatures. should be -1, 0, or a positive integer
//       -1 variable-length argument; >=0 fixed numbers argument
// fun:  function handler
func RegFunction(name string, argc int, fun func(...ExprAST) decimal.Decimal) error {
	if len(name) == 0 {
		return errors.New("RegFunction name is not empty")
	}
	if argc < -1 {
		return errors.New("RegFunction argc should be -1, 0, or a positive integer")
	}
	if _, ok := defFunc[name]; ok {
		return errors.New("RegFunction name is already exist")
	}
	defFunc[name] = defS{argc, fun}
	return nil
}

// ExprASTResult is a Top level function
// AST traversal
// if an arithmetic runtime error occurs, a panic exception is thrown
func ExprASTResult(expr ExprAST) decimal.Decimal {
	var l, r decimal.Decimal
	switch expr.(type) {
	case BinaryExprAST:
		ast := expr.(BinaryExprAST)
		l = ExprASTResult(ast.Lhs)
		r = ExprASTResult(ast.Rhs)
		switch ast.Op {
		case "+":
			return l.Add(r)
		case "-":
			return l.Sub(r)
		case "*":
			return l.Mul(r)
		case "/":
			if r.IsZero() {
				panic(errors.New(
					fmt.Sprintf("violation of arithmetic specification: a division by zero in ExprASTResult: [%g/%g]",
						l,
						r)))
			}
			return l.Div(r)
		case "%":
			if r.IsZero() {
				panic(errors.New(
					fmt.Sprintf("violation of arithmetic specification: a division by zero in ExprASTResult: [%g%%%g]",
						l,
						r)))
			}
			return l.Mod(r)
		case "^":
			return l.Pow(r)
		default:

		}
	case NumberExprAST:
		return expr.(NumberExprAST).Val
	case FunCallerExprAST:
		f := expr.(FunCallerExprAST)
		def := defFunc[f.Name]
		return def.fun(f.Arg...)
	}

	return decimal.Zero
}
