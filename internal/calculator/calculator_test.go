package calculator

import (
	"testing"
)

func TestCalculatorParse(t *testing.T) {
	testCases := []struct {
		name          string
		expression    string
		expectedArg1  float64
		expectedArg2  float64
		expectedOp    string
		expectError   bool
	}{
		{
			name:         "Addition",
			expression:   "5 + 3",
			expectedArg1: 5,
			expectedArg2: 3,
			expectedOp:   "+",
			expectError:  false,
		},
		{
			name:         "Subtraction",
			expression:   "10 - 7",
			expectedArg1: 10,
			expectedArg2: 7,
			expectedOp:   "-",
			expectError:  false,
		},
		{
			name:         "Multiplication",
			expression:   "3.5 * 2",
			expectedArg1: 3.5,
			expectedArg2: 2,
			expectedOp:   "*",
			expectError:  false,
		},
		{
			name:         "Division",
			expression:   "15 / 3",
			expectedArg1: 15,
			expectedArg2: 3,
			expectedOp:   "/",
			expectError:  false,
		},
		{
			name:        "Missing Second Argument",
			expression:  "5 + ",
			expectError: true,
		},
		{
			name:        "Missing First Argument",
			expression:  "+ 3",
			expectError: true,
		},
		{
			name:        "Invalid Operation",
			expression:  "5 ? 3",
			expectError: true,
		},
		{
			name:        "Non-Numeric Value",
			expression:  "abc + 3",
			expectError: true,
		},
		{
			name:        "Empty String",
			expression:  "",
			expectError: true,
		},
	}

	calculator := NewCalculator()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			arg1, arg2, op, err := calculator.Parse(tc.expression)

			if tc.expectError {
				if err == nil {
					t.Errorf("ожидалась ошибка при парсинге '%s', но парсинг успешен", tc.expression)
				}
				return
			}

			if err != nil {
				t.Errorf("неожиданная ошибка при парсинге '%s': %v", tc.expression, err)
				return
			}

			if arg1 != tc.expectedArg1 {
				t.Errorf("неверное значение arg1: ожидалось %f, получено %f", tc.expectedArg1, arg1)
			}

			if arg2 != tc.expectedArg2 {
				t.Errorf("неверное значение arg2: ожидалось %f, получено %f", tc.expectedArg2, arg2)
			}

			if op != tc.expectedOp {
				t.Errorf("неверная операция: ожидалось '%s', получено '%s'", tc.expectedOp, op)
			}
		})
	}
}

func TestCalculate(t *testing.T) {
	testCases := []struct {
		name           string
		arg1           float64
		arg2           float64
		operation      string
		expectedResult float64
		expectError    bool
	}{
		{
			name:           "Addition",
			arg1:           10,
			arg2:           5,
			operation:      "+",
			expectedResult: 15,
			expectError:    false,
		},
		{
			name:           "Subtraction",
			arg1:           10,
			arg2:           5,
			operation:      "-",
			expectedResult: 5,
			expectError:    false,
		},
		{
			name:           "Multiplication",
			arg1:           10,
			arg2:           5,
			operation:      "*",
			expectedResult: 50,
			expectError:    false,
		},
		{
			name:           "Division",
			arg1:           10,
			arg2:           5,
			operation:      "/",
			expectedResult: 2,
			expectError:    false,
		},
		{
			name:        "Division by Zero",
			arg1:        10,
			arg2:        0,
			operation:   "/",
			expectError: true,
		},
		{
			name:        "Invalid Operation",
			arg1:        10,
			arg2:        5,
			operation:   "%",
			expectError: true,
		},
	}

	calculator := NewCalculator()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := calculator.Calculate(tc.arg1, tc.arg2, tc.operation)

			if tc.expectError {
				if err == nil {
					t.Errorf("ожидалась ошибка при вычислении '%f %s %f', но вычисление успешно", 
						tc.arg1, tc.operation, tc.arg2)
				}
				return
			}

			if err != nil {
				t.Errorf("неожиданная ошибка при вычислении '%f %s %f': %v", 
					tc.arg1, tc.operation, tc.arg2, err)
				return
			}

			if result != tc.expectedResult {
				t.Errorf("неверный результат: ожидалось %f, получено %f", tc.expectedResult, result)
			}
		})
	}
}

func TestIntegrationParseAndCalculate(t *testing.T) {
	testCases := []struct {
		name           string
		expression     string
		expectedResult float64
		expectError    bool
	}{
		{
			name:           "Addition",
			expression:     "10 + 5",
			expectedResult: 15,
			expectError:    false,
		},
		{
			name:           "Subtraction",
			expression:     "10 - 5",
			expectedResult: 5,
			expectError:    false,
		},
		{
			name:           "Multiplication",
			expression:     "10 * 5",
			expectedResult: 50,
			expectError:    false,
		},
		{
			name:           "Division",
			expression:     "10 / 5",
			expectedResult: 2,
			expectError:    false,
		},
		{
			name:        "Division by Zero",
			expression:  "10 / 0",
			expectError: true,
		},
		{
			name:        "Invalid Expression",
			expression:  "10 % 5",
			expectError: true,
		},
		{
			name:        "Non-Numeric Value",
			expression:  "abc + 3",
			expectError: true,
		},
	}

	calculator := NewCalculator()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Парсим выражение
			arg1, arg2, op, err := calculator.Parse(tc.expression)
			if err != nil {
				if !tc.expectError {
					t.Errorf("неожиданная ошибка при парсинге '%s': %v", tc.expression, err)
				}
				return
			}

			// Вычисляем результат
			result, err := calculator.Calculate(arg1, arg2, op)
			if err != nil {
				if !tc.expectError {
					t.Errorf("неожиданная ошибка при вычислении '%s': %v", tc.expression, err)
				}
				return
			}

			if tc.expectError {
				t.Errorf("ожидалась ошибка при обработке '%s', но обработка успешна", tc.expression)
				return
			}

			if result != tc.expectedResult {
				t.Errorf("неверный результат: ожидалось %f, получено %f", tc.expectedResult, result)
			}
		})
	}
}
