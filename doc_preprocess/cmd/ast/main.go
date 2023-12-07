package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"reflect"
)

// EvalVisitor 用于遍历AST并计算表达式的值
type EvalVisitor struct {
	result int
}

func (ev *EvalVisitor) Visit(node ast.Node) ast.Visitor {
	switch n := node.(type) {
	case *ast.BinaryExpr:
		switch n.Op {
		case token.ADD:
			ev.result = ev.eval(n.X) + ev.eval(n.Y)
		case token.SUB:
			ev.result = ev.eval(n.X) - ev.eval(n.Y)
		case token.MUL:
			ev.result = ev.eval(n.X) * ev.eval(n.Y)
		case token.QUO:
			ev.result = ev.eval(n.X) / ev.eval(n.Y)
		}
	}
	return ev
}

func (ev *EvalVisitor) eval(expr ast.Expr) int {
	switch n := expr.(type) {
	case *ast.BinaryExpr:
		return ev.eval(n)
	case *ast.BasicLit:
		value, _ := evalBasicLit(n)
		return value
	}
	return 0
}

func evalBasicLit(lit *ast.BasicLit) (int, error) {
	return fmt.Sscanf(lit.Value, "%d")
}

func (data *LogData) AstBasicExpr() {
	expr, _ := parser.ParseExpr(`request.body.messages[0].id.name == 7 || 7==7`)
	ast.Print(nil, expr)
	fmt.Println(data.Eval(data, expr))
}

var kindMapping = map[token.Token]reflect.Kind{
	token.INT:    reflect.Int,
	token.FLOAT:  reflect.Float64,
	token.STRING: reflect.String,
}

func (data *LogData) Eval(d Filterable, exp ast.Expr) (interface{}, reflect.Kind) {
	switch exp := exp.(type) {
	case *ast.BinaryExpr: //如果是二元表达式类型，调用EvalBinaryExpr进行解析
		return data.EvalBinaryExpr(exp)
	case *ast.BasicLit: //如果是基础面值类型
		return exp.Value, kindMapping[exp.Kind]
	case *ast.SelectorExpr: //如果是选择器表达式类型
		// Handle selector expressions (assumed to be struct.field)
		objValue := reflect.ValueOf(d)
		// Ensure objValue is a struct
		if objValue.Kind() == reflect.Struct {
			// Evaluate X part of the selector (e.g., struct)
			structObj, _ := data.Eval(objValue, exp.X)
			// Evaluate Sel part of the selector (e.g., field)
			return data.Eval(structObj, exp.Sel)
		}
	case *ast.Ident: //如果是标识符类型
		// Handle identifiers (assumed to be struct field names)
		value := reflect.ValueOf(d)
		field := value.FieldByName(exp.Name)
		if field.IsValid() {
			v := field.Interface()
			return v, field.Kind()
		}
	case *ast.IndexExpr: //如果是索引表达式类型
	case *ast.IndexListExpr: //如果是索引列表表达式类型
	case *ast.SliceExpr: //如果是切片表达式类型
	}

	return 0, reflect.Invalid
}

func (data *LogData) EvalBinaryExpr(exp *ast.BinaryExpr) (interface{}, reflect.Kind) { //这里仅实现了+和*
	x, kind := data.Eval(data, exp.X)
	y, _ := data.Eval(data, exp.Y)
	switch exp.Op {
	case token.EQL:
		return reflect.DeepEqual(x, y), kind
	case token.LOR:
		return x.(bool) || y.(bool), kind
		// TODO: support
	}
	return false, reflect.Bool
}

func main() {

	data := LogData{
		Request: Request{
			Body: Body{
				Source: 7,
			},
		},
	}
	// field := getField(data, "request.body.source")

	data.AstBasicExpr()

	// ev := &EvalVisitor{}
	// ast.Walk(ev, expr)

	// fmt.Printf("Result of expression %s: %d\n", exprStr, ev.result)
}
