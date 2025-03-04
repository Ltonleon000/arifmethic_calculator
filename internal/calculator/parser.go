package calculator

import (
	"fmt"
	"strconv"
	"strings"
)

type Operator struct {
	Type  string
	Value string
}

type Parser struct {
	operators []Operator
	pos       int
}

func NewParser(expression string) *Parser {
	return &Parser{
		operators: operators(expression),
		pos:       0,
	}
}

// разбиваем строку выражения на ператоры например:
//
//	"2+2*2" -> [{number "2"} {operator "+"} {number "2"} {operator "*"} {number "2"}]
func operators(expression string) []Operator {
	expression = strings.ReplaceAll(expression, " ", "")
	var operators []Operator
	var num strings.Builder

	for i := 0; i < len(expression); i++ {
		ch := expression[i]
		if isDigit(ch) || ch == '.' {
			num.WriteByte(ch)
		} else {
			if num.Len() > 0 {
				operators = append(operators, Operator{Type: "number", Value: num.String()})
				num.Reset()
			}
			if isOperator(ch) {
				operators = append(operators, Operator{Type: "operator", Value: string(ch)})
			}
		}
	}
	if num.Len() > 0 {
		operators = append(operators, Operator{Type: "number", Value: num.String()})
	}
	return operators
}

// Parse разбирает выражение и возвращает список операций с учётом приоритета
// Сначала выполняются умножение и деление, потом сложение и вычитание
func (p *Parser) Parse() ([]Operation, error) {
	var firstPass []Operator
	i := 0
	for i < len(p.operators) {
		if i+2 < len(p.operators) &&
			p.operators[i].Type == "number" &&
			(p.operators[i+1].Value == "*" || p.operators[i+1].Value == "/") &&
			p.operators[i+2].Type == "number" {
			firstPass = append(firstPass, Operator{
				Type:  "number",
				Value: fmt.Sprintf("(%s%s%s)", p.operators[i].Value, p.operators[i+1].Value, p.operators[i+2].Value),
			})
			i += 3
		} else {
			firstPass = append(firstPass, p.operators[i])
			i++
		}
	}

	var operations []Operation
	i = 0
	for i < len(firstPass) {
		if firstPass[i].Type == "number" {
			val := firstPass[i].Value
			if strings.HasPrefix(val, "(") && strings.HasSuffix(val, ")") {
				val = val[1 : len(val)-1]
				ops := strings.Split(val, "")
				num1, _ := strconv.ParseFloat(ops[0], 64)
				num2, _ := strconv.ParseFloat(ops[2], 64)
				operations = append(operations, Operation{
					Operator: ops[1],
					Operand1: num1,
					Operand2: num2,
				})
			} else {
				if i+2 < len(firstPass) &&
					firstPass[i+1].Type == "operator" &&
					firstPass[i+2].Type == "number" {
					num1, _ := strconv.ParseFloat(val, 64)
					num2, _ := strconv.ParseFloat(firstPass[i+2].Value, 64)
					operations = append(operations, Operation{
						Operator: firstPass[i+1].Value,
						Operand1: num1,
						Operand2: num2,
					})
					i += 2
				}
			}
		}
		i++
	}
	return operations, nil
}

type Operation struct {
	Operator string
	Operand1 float64
	Operand2 float64
}

func isDigit(ch byte) bool {
	return ch >= '0' && ch <= '9'
}

func isOperator(ch byte) bool {
	return ch == '+' || ch == '-' || ch == '*' || ch == '/'
}
