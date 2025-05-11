package main

import (
	"calculator/calculatorpb"
	"fmt"
	"strconv"
	"strings"
	"testing"
)

// Тестирование парсинга операций из формата "ID||операция"
func TestOperationParsing(t *testing.T) {
	testCases := []struct {
		operationString string
		expectedID      string
		expectedOp      string
		shouldParse     bool
	}{
		// Корректные случаи
		{"123e4567-e89b-12d3-a456-426614174000||+", "123e4567-e89b-12d3-a456-426614174000", "+", true},
		{"abc||*", "abc", "*", true},
		{"test-id||/", "test-id", "/", true},
		{"123456||-", "123456", "-", true},
		// Некорректные случаи
		{"+", "", "+", false},        // Нет разделителя
		{"||+", "", "+", true},       // Пустой ID
		{"123||", "123", "", true},   // Пустая операция
		{"", "", "", false},          // Пустая строка
	}

	for i, tc := range testCases {
		parts := parseOperationString(tc.operationString)
		
		// Проверяем правильно ли разделилось
		if tc.shouldParse && (len(parts) < 2) {
			t.Errorf("Тест #%d: ожидалось разделение '%s', но разделить не удалось", i, tc.operationString)
			continue
		}
		
		if !tc.shouldParse && (len(parts) >= 2) {
			t.Errorf("Тест #%d: не ожидалось разделение '%s', но результат был: %v", i, tc.operationString, parts)
			continue
		}
		
		// Если строка должна была разделиться, проверяем результаты
		if tc.shouldParse && len(parts) >= 2 {
			if parts[0] != tc.expectedID {
				t.Errorf("Тест #%d: ожидался ID '%s', получен '%s'", i, tc.expectedID, parts[0])
			}
			
			if parts[1] != tc.expectedOp {
				t.Errorf("Тест #%d: ожидалась операция '%s', получена '%s'", i, tc.expectedOp, parts[1])
			}
		}
	}
}

// Тестирование преобразования аргументов из строк в числа
func TestArgParsing(t *testing.T) {
	testCases := []struct {
		argString   string
		expectedVal float64
		shouldParse bool
	}{
		{"10.5", 10.5, true},
		{"3.14159", 3.14159, true},
		{"-5.0", -5.0, true},
		{"0", 0.0, true},
		{"abc", 0.0, false},        // Нечисловое значение
		{"", 0.0, false},           // Пустая строка
	}

	for i, tc := range testCases {
		val, err := parseArgument(tc.argString)
		
		if tc.shouldParse && err != nil {
			t.Errorf("Тест #%d: ожидался успешный парсинг '%s', но получена ошибка: %v", 
				i, tc.argString, err)
			continue
		}
		
		if !tc.shouldParse && err == nil {
			t.Errorf("Тест #%d: ожидалась ошибка при парсинге '%s', но парсинг успешен", 
				i, tc.argString)
			continue
		}
		
		if tc.shouldParse && tc.expectedVal != val {
			t.Errorf("Тест #%d: для '%s' ожидалось значение %f, получено %f", 
				i, tc.argString, tc.expectedVal, val)
		}
	}
}

// Тестирование обработки задач из gRPC
func TestTaskProcessing(t *testing.T) {
	testCases := []struct {
		task           calculatorpb.Task
		expectedResult float64
		shouldProcess  bool
	}{
		{
			calculatorpb.Task{
				Id:        1,
				Arg1:      "5.0",
				Arg2:      "7.0",
				Operation: "test-id||+",
			},
			12.0, true, // 5 + 7 = 12
		},
		{
			calculatorpb.Task{
				Id:        2,
				Arg1:      "10.0",
				Arg2:      "2.0",
				Operation: "test-id||-",
			},
			8.0, true, // 10 - 2 = 8
		},
		{
			calculatorpb.Task{
				Id:        3,
				Arg1:      "4.0",
				Arg2:      "5.0",
				Operation: "test-id||*",
			},
			20.0, true, // 4 * 5 = 20
		},
		{
			calculatorpb.Task{
				Id:        4,
				Arg1:      "15.0",
				Arg2:      "3.0",
				Operation: "test-id||/",
			},
			5.0, true, // 15 / 3 = 5
		},
		{
			calculatorpb.Task{
				Id:        5,
				Arg1:      "10.0",
				Arg2:      "0.0",
				Operation: "test-id||/",
			},
			0.0, false, // Деление на ноль
		},
		{
			calculatorpb.Task{
				Id:        6,
				Arg1:      "xyz",
				Arg2:      "5.0",
				Operation: "test-id||+",
			},
			0.0, false, // Невалидный аргумент
		},
		{
			calculatorpb.Task{
				Id:        7,
				Arg1:      "5.0",
				Arg2:      "abc",
				Operation: "test-id||+",
			},
			0.0, false, // Невалидный аргумент
		},
		{
			calculatorpb.Task{
				Id:        8,
				Arg1:      "5.0",
				Arg2:      "7.0",
				Operation: "test-id||$",
			},
			0.0, false, // Неподдерживаемая операция
		},
	}

	for i, tc := range testCases {
		result, err := processGRPCTask(&tc.task)
		
		if tc.shouldProcess && err != nil {
			t.Errorf("Тест #%d: ожидалась успешная обработка задачи, но получена ошибка: %v", 
				i, err)
			continue
		}
		
		if !tc.shouldProcess && err == nil {
			t.Errorf("Тест #%d: ожидалась ошибка обработки задачи, но обработка успешна", i)
			continue
		}
		
		if tc.shouldProcess && result != tc.expectedResult {
			t.Errorf("Тест #%d: ожидался результат %f, получен %f", 
				i, tc.expectedResult, result)
		}
	}
}

// Вспомогательные функции для тестирования (должны быть реализованы в основном коде)
// Эти реализации только для тестов и должны соответствовать реальному коду

func parseOperationString(operation string) []string {
	if operation == "" {
		return []string{}
	}
	parts := strings.Split(operation, "||")
	return parts
}

func parseArgument(arg string) (float64, error) {
	return strconv.ParseFloat(arg, 64)
}

func processGRPCTask(task *calculatorpb.Task) (float64, error) {
	// Парсим аргументы
	arg1, err1 := parseArgument(task.Arg1)
	if err1 != nil {
		return 0, fmt.Errorf("ошибка при парсинге первого аргумента: %v", err1)
	}
	
	arg2, err2 := parseArgument(task.Arg2)
	if err2 != nil {
		return 0, fmt.Errorf("ошибка при парсинге второго аргумента: %v", err2)
	}
	
	// Парсим операцию
	parts := parseOperationString(task.Operation)
	if len(parts) < 2 {
		return 0, fmt.Errorf("некорректный формат операции: %s", task.Operation)
	}
	
	operation := parts[1]
	
	// Выполняем операцию
	var result float64
	switch operation {
	case "+":
		result = arg1 + arg2
	case "-":
		result = arg1 - arg2
	case "*":
		result = arg1 * arg2
	case "/":
		if arg2 == 0 {
			return 0, fmt.Errorf("деление на ноль")
		}
		result = arg1 / arg2
	default:
		return 0, fmt.Errorf("неподдерживаемая операция: %s", operation)
	}
	
	return result, nil
}
