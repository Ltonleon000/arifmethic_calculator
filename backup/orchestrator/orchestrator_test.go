package main

import (
	"calculator/internal/calculator"
	"testing"
)

// Тест для проверки парсинга математических выражений
func TestCalculatorParsing(t *testing.T) {
	testCases := []struct {
		expression      string
		expectedArg1    float64
		expectedArg2    float64
		expectedOp      string
		shouldParseSuccessfully bool
	}{
		// Корректные выражения
		{"5 + 3", 5, 3, "+", true},
		{"10 - 7", 10, 7, "-", true},
		{"3.5 * 2", 3.5, 2, "*", true},
		{"15 / 3", 15, 3, "/", true},
		
		// Некорректные выражения
		{"5 + ", 0, 0, "", false},
		{"+ 3", 0, 0, "", false},
		{"5 ? 3", 0, 0, "", false},
		{"abc", 0, 0, "", false},
	}

	calc := calculator.NewCalculator()

	for i, tc := range testCases {
		arg1, arg2, op, err := calc.Parse(tc.expression)
		
		if tc.shouldParseSuccessfully && err != nil {
			t.Errorf("Тест #%d: ожидался успешный парсинг выражения '%s', но получена ошибка: %v", 
				i, tc.expression, err)
			continue
		}
		
		if !tc.shouldParseSuccessfully && err == nil {
			t.Errorf("Тест #%d: ожидалась ошибка при парсинге выражения '%s', но парсинг успешен", 
				i, tc.expression)
			continue
		}
		
		// Если выражение должно было распарситься успешно, проверяем результаты
		if tc.shouldParseSuccessfully {
			if arg1 != tc.expectedArg1 {
				t.Errorf("Тест #%d: для выражения '%s' ожидалось значение arg1 = %f, получено %f", 
					i, tc.expression, tc.expectedArg1, arg1)
			}
			
			if arg2 != tc.expectedArg2 {
				t.Errorf("Тест #%d: для выражения '%s' ожидалось значение arg2 = %f, получено %f", 
					i, tc.expression, tc.expectedArg2, arg2)
			}
			
			if op != tc.expectedOp {
				t.Errorf("Тест #%d: для выражения '%s' ожидалась операция '%s', получена '%s'", 
					i, tc.expression, tc.expectedOp, op)
			}
		}
	}
}

// Тест для проверки вычисления результатов
func TestCalculatorCalculate(t *testing.T) {
	testCases := []struct {
		arg1          float64
		arg2          float64
		operation     string
		expectedResult float64
		shouldCalcSuccessfully bool
	}{
		// Корректные операции
		{10, 5, "+", 15, true},
		{10, 5, "-", 5, true},
		{10, 5, "*", 50, true},
		{10, 5, "/", 2, true},
		
		// Некорректные операции
		{10, 0, "/", 0, false},  // Деление на ноль
		{10, 5, "?", 0, false},  // Неизвестная операция
	}

	calc := calculator.NewCalculator()

	for i, tc := range testCases {
		result, err := calc.Calculate(tc.arg1, tc.arg2, tc.operation)
		
		if tc.shouldCalcSuccessfully && err != nil {
			t.Errorf("Тест #%d: ожидалось успешное вычисление %f %s %f, но получена ошибка: %v", 
				i, tc.arg1, tc.operation, tc.arg2, err)
			continue
		}
		
		if !tc.shouldCalcSuccessfully && err == nil {
			t.Errorf("Тест #%d: ожидалась ошибка при вычислении %f %s %f, но вычисление успешно", 
				i, tc.arg1, tc.operation, tc.arg2)
			continue
		}
		
		// Если вычисление должно было быть успешным, проверяем результат
		if tc.shouldCalcSuccessfully && result != tc.expectedResult {
			t.Errorf("Тест #%d: для операции %f %s %f ожидался результат %f, получен %f", 
				i, tc.arg1, tc.operation, tc.arg2, tc.expectedResult, result)
		}
	}
}
