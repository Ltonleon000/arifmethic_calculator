package calculator

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// Calculator представляет собой калькулятор для арифметических выражений
type Calculator struct {
	parser *Parser
}

// NewCalculator создает новый экземпляр калькулятора
func NewCalculator() *Calculator {
	return &Calculator{}
}

// Parse разбирает выражение на аргументы и операцию
func (c *Calculator) Parse(expression string) (float64, float64, string, error) {
	// Удаляем лишние пробелы
	expression = strings.TrimSpace(expression)
	if expression == "" {
		return 0, 0, "", errors.New("пустое выражение")
	}

	// Ищем оператор
	var operator string
	var operatorIndex int = -1

	for i, char := range expression {
		if char == '+' || char == '-' || char == '*' || char == '/' {
			operator = string(char)
			operatorIndex = i
			break
		}
	}

	if operatorIndex == -1 {
		return 0, 0, "", errors.New("оператор не найден")
	}

	// Получаем аргументы
	arg1Str := strings.TrimSpace(expression[:operatorIndex])
	arg2Str := strings.TrimSpace(expression[operatorIndex+1:])

	if arg1Str == "" || arg2Str == "" {
		return 0, 0, "", errors.New("неполное выражение")
	}

	// Конвертируем строки в числа
	arg1, err := strconv.ParseFloat(arg1Str, 64)
	if err != nil {
		return 0, 0, "", fmt.Errorf("ошибка преобразования первого аргумента: %v", err)
	}

	arg2, err := strconv.ParseFloat(arg2Str, 64)
	if err != nil {
		return 0, 0, "", fmt.Errorf("ошибка преобразования второго аргумента: %v", err)
	}

	return arg1, arg2, operator, nil
}

// Calculate выполняет вычисление на основе двух аргументов и оператора
func (c *Calculator) Calculate(arg1, arg2 float64, operator string) (float64, error) {
	switch operator {
	case "+":
		return arg1 + arg2, nil
	case "-":
		return arg1 - arg2, nil
	case "*":
		return arg1 * arg2, nil
	case "/":
		if arg2 == 0 {
			return 0, errors.New("деление на ноль")
		}
		return arg1 / arg2, nil
	default:
		return 0, fmt.Errorf("неподдерживаемый оператор: %s", operator)
	}
}

// EvaluateExpression анализирует выражение и вычисляет результат
func (c *Calculator) EvaluateExpression(expression string) (float64, error) {
	arg1, arg2, operator, err := c.Parse(expression)
	if err != nil {
		return 0, err
	}

	return c.Calculate(arg1, arg2, operator)
}
