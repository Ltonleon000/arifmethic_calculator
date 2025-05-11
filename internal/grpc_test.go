package internal

import (
	"calculator/calculatorpb"
	"context"
	"fmt"
	"strings"
	"testing"
)

// Тестовый сервер для эмуляции оркестратора
type mockAgentServer struct {
	calculatorpb.UnimplementedAgentServiceServer
	getTaskFunc    func(context.Context, *calculatorpb.GetTaskRequest) (*calculatorpb.GetTaskResponse, error)
	sendResultFunc func(context.Context, *calculatorpb.SendResultRequest) (*calculatorpb.SendResultResponse, error)
}

func (s *mockAgentServer) GetTask(ctx context.Context, req *calculatorpb.GetTaskRequest) (*calculatorpb.GetTaskResponse, error) {
	return s.getTaskFunc(ctx, req)
}

func (s *mockAgentServer) SendResult(ctx context.Context, req *calculatorpb.SendResultRequest) (*calculatorpb.SendResultResponse, error) {
	return s.sendResultFunc(ctx, req)
}

// Тест для проверки создания gRPC клиента агента
func TestAgentGRPCClient(t *testing.T) {
	// Тест создания клиента с неверным адресом
	_, err := NewAgentGRPCClient("invalid:address")
	if err == nil {
		t.Error("Ожидалась ошибка при подключении к некорректному адресу, но ошибки не было")
	}
}

// Тест для передачи задачи от клиента агенту
func TestTaskDataConversion(t *testing.T) {
	// Создаем тестовую задачу
	task := &calculatorpb.Task{
		Id:            123,
		Arg1:          "10.5",
		Arg2:          "20.5",
		Operation:     "abcd1234||+",
		OperationTime: 1000,
	}

	// Проверяем форматирование аргументов
	arg1, err1 := parseTaskArg(task.Arg1)
	arg2, err2 := parseTaskArg(task.Arg2)
	
	if err1 != nil {
		t.Errorf("Ошибка парсинга первого аргумента: %v", err1)
	}
	
	if err2 != nil {
		t.Errorf("Ошибка парсинга второго аргумента: %v", err2)
	}
	
	if arg1 != 10.5 {
		t.Errorf("Неправильное значение первого аргумента: ожидалось 10.5, получено %f", arg1)
	}
	
	if arg2 != 20.5 {
		t.Errorf("Неправильное значение второго аргумента: ожидалось 20.5, получено %f", arg2)
	}
	
	// Проверяем парсинг составной операции
	parts := parseOperationParts(task.Operation)
	if len(parts) < 2 {
		t.Error("Не удалось распарсить составную строку операции")
	} else {
		if parts[0] != "abcd1234" {
			t.Errorf("Неправильный идентификатор задачи: ожидался abcd1234, получено %s", parts[0])
		}
		
		if parts[1] != "+" {
			t.Errorf("Неправильная операция: ожидался +, получено %s", parts[1])
		}
	}
}

// Вспомогательные функции для тестов
func parseTaskArg(arg string) (float64, error) {
	var result float64
	_, err := fmt.Sscanf(arg, "%f", &result)
	return result, err
}

func parseOperationParts(operation string) []string {
	return strings.Split(operation, "||")
}
