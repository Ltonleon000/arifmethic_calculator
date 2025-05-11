package tests

import (
	"calculator/calculatorpb"
	"context"
	"net"
	"strconv"
	"strings"
	"sync"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

const bufSize = 1024 * 1024

// Настраиваем буфер для локального gRPC-сервера в тестах
var lis *bufconn.Listener

func init() {
	lis = bufconn.Listen(bufSize)
}

// Вспомогательная функция для создания диалера для буферизованного соединения
func bufDialer(context.Context, string) (net.Conn, error) {
	return lis.Dial()
}

// MockAgentService - имитирует сервис агента для тестирования
type MockAgentService struct {
	calculatorpb.UnimplementedAgentServiceServer
	tasks         []*calculatorpb.Task
	results       map[int64]float64
	tasksMutex    sync.Mutex
	resultsMutex  sync.Mutex
}

func NewMockAgentService() *MockAgentService {
	return &MockAgentService{
		tasks:      make([]*calculatorpb.Task, 0),
		results:    make(map[int64]float64),
	}
}

// GetTask - имитирует получение задачи агентом
func (s *MockAgentService) GetTask(ctx context.Context, req *calculatorpb.GetTaskRequest) (*calculatorpb.GetTaskResponse, error) {
	s.tasksMutex.Lock()
	defer s.tasksMutex.Unlock()

	if len(s.tasks) == 0 {
		return &calculatorpb.GetTaskResponse{
			HasTask: false,
			Task:    nil,
		}, nil
	}

	task := s.tasks[0]
	s.tasks = s.tasks[1:] // Удаляем задачу из очереди

	return &calculatorpb.GetTaskResponse{
		HasTask: true,
		Task:    task,
	}, nil
}

// SendResult - имитирует отправку результата агентом
func (s *MockAgentService) SendResult(ctx context.Context, req *calculatorpb.SendResultRequest) (*calculatorpb.SendResultResponse, error) {
	s.resultsMutex.Lock()
	defer s.resultsMutex.Unlock()

	s.results[req.TaskId] = req.Result
	return &calculatorpb.SendResultResponse{
		Ok: true,
	}, nil
}

// Добавление задачи в очередь мок-сервиса
func (s *MockAgentService) AddTask(task *calculatorpb.Task) {
	s.tasksMutex.Lock()
	defer s.tasksMutex.Unlock()

	s.tasks = append(s.tasks, task)
}

// Получение результата из мок-сервиса
func (s *MockAgentService) GetResult(taskId int64) (float64, bool) {
	s.resultsMutex.Lock()
	defer s.resultsMutex.Unlock()

	result, exists := s.results[taskId]
	return result, exists
}

// Тест интеграции между оркестратором и агентом через gRPC
func TestGRPCIntegration(t *testing.T) {
	// Создаем мок-сервис агента
	mockService := NewMockAgentService()

	// Запускаем gRPC сервер
	server := grpc.NewServer()
	calculatorpb.RegisterAgentServiceServer(server, mockService)
	
	// Запускаем сервер в отдельной горутине
	go func() {
		if err := server.Serve(lis); err != nil {
			t.Logf("Ошибка запуска gRPC сервера: %v", err)
		}
	}()
	defer server.Stop()

	// Создаем клиента для тестов
	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "bufnet", 
		grpc.WithContextDialer(bufDialer),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("Не удалось создать соединение: %v", err)
	}
	defer conn.Close()

	client := calculatorpb.NewAgentServiceClient(conn)

	// Тест 1: Агент запрашивает задачу, когда очередь пуста
	t.Run("EmptyTaskQueue", func(t *testing.T) {
		resp, err := client.GetTask(ctx, &calculatorpb.GetTaskRequest{})
		if err != nil {
			t.Fatalf("Ошибка запроса задачи: %v", err)
		}

		if resp.HasTask {
			t.Error("Ожидалось отсутствие задач в очереди, но получена задача")
		}
	})

	// Тест 2: Агент запрашивает задачу, когда в очереди есть задача
	t.Run("TaskAvailable", func(t *testing.T) {
		// Добавляем задачу
		task := &calculatorpb.Task{
			Id:        123,
			Arg1:      "10.5",
			Arg2:      "5.5",
			Operation: "task-123||+",
		}
		mockService.AddTask(task)

		// Запрашиваем задачу
		resp, err := client.GetTask(ctx, &calculatorpb.GetTaskRequest{})
		if err != nil {
			t.Fatalf("Ошибка запроса задачи: %v", err)
		}

		if !resp.HasTask {
			t.Fatal("Ожидалось наличие задачи, но задача не получена")
		}

		if resp.Task.Id != task.Id {
			t.Errorf("Неверный ID задачи: ожидалось %d, получено %d", task.Id, resp.Task.Id)
		}

		if resp.Task.Arg1 != task.Arg1 {
			t.Errorf("Неверный Arg1: ожидалось %s, получено %s", task.Arg1, resp.Task.Arg1)
		}

		if resp.Task.Arg2 != task.Arg2 {
			t.Errorf("Неверный Arg2: ожидалось %s, получено %s", task.Arg2, resp.Task.Arg2)
		}

		if resp.Task.Operation != task.Operation {
			t.Errorf("Неверная операция: ожидалось %s, получено %s", task.Operation, resp.Task.Operation)
		}
	})

	// Тест 3: Агент отправляет результат вычисления
	t.Run("SendResult", func(t *testing.T) {
		taskId := int64(456)
		result := 16.0

		// Отправляем результат
		resp, err := client.SendResult(ctx, &calculatorpb.SendResultRequest{
			TaskId: taskId,
			Result: result,
		})
		if err != nil {
			t.Fatalf("Ошибка отправки результата: %v", err)
		}

		if !resp.Ok {
			t.Error("Ожидалось успешное сохранение результата, но получен отказ")
		}

		// Проверяем, что результат был сохранен
		savedResult, exists := mockService.GetResult(taskId)
		if !exists {
			t.Fatal("Результат не сохранен в хранилище")
		}

		if savedResult != result {
			t.Errorf("Неверный сохраненный результат: ожидалось %f, получено %f", result, savedResult)
		}
	})

	// Тест 4: Полный жизненный цикл задачи (добавление -> получение -> вычисление -> отправка результата)
	t.Run("FullTaskLifecycle", func(t *testing.T) {
		// Добавляем задачу
		task := &calculatorpb.Task{
			Id:        789,
			Arg1:      "20.0",
			Arg2:      "5.0",
			Operation: "task-789||-",
		}
		mockService.AddTask(task)

		// Агент запрашивает задачу
		getResp, err := client.GetTask(ctx, &calculatorpb.GetTaskRequest{})
		if err != nil {
			t.Fatalf("Ошибка запроса задачи: %v", err)
		}

		if !getResp.HasTask {
			t.Fatal("Ожидалось наличие задачи, но задача не получена")
		}

		// Агент обрабатывает задачу
		parts := parseOperationString(getResp.Task.Operation)
		if len(parts) < 2 {
			t.Fatalf("Неверный формат операции: %s", getResp.Task.Operation)
		}

		operationType := parts[1]
		arg1, err := parseFloat(getResp.Task.Arg1)
		if err != nil {
			t.Fatalf("Ошибка парсинга Arg1: %v", err)
		}

		arg2, err := parseFloat(getResp.Task.Arg2)
		if err != nil {
			t.Fatalf("Ошибка парсинга Arg2: %v", err)
		}

		// Вычисляем результат
		calculatedResult := 0.0
		switch operationType {
		case "+":
			calculatedResult = arg1 + arg2
		case "-":
			calculatedResult = arg1 - arg2
		case "*":
			calculatedResult = arg1 * arg2
		case "/":
			if arg2 == 0 {
				t.Fatalf("Ошибка: деление на ноль")
			}
			calculatedResult = arg1 / arg2
		default:
			t.Fatalf("Неизвестная операция: %s", operationType)
		}

		// Агент отправляет результат
		sendResp, err := client.SendResult(ctx, &calculatorpb.SendResultRequest{
			TaskId: getResp.Task.Id,
			Result: calculatedResult,
		})
		if err != nil {
			t.Fatalf("Ошибка отправки результата: %v", err)
		}

		if !sendResp.Ok {
			t.Error("Ожидалось успешное сохранение результата, но получен отказ")
		}

		// Проверяем, что результат был сохранен правильно
		savedResult, exists := mockService.GetResult(getResp.Task.Id)
		if !exists {
			t.Fatal("Результат не сохранен в хранилище")
		}

		expectedResult := 15.0 // 20.0 - 5.0 = 15.0
		if savedResult != expectedResult {
			t.Errorf("Неверный сохраненный результат: ожидалось %f, получено %f", expectedResult, savedResult)
		}
	})
}

// Вспомогательные функции для парсинга
func parseOperationString(operation string) []string {
	return strings.Split(operation, "||")
}

func parseFloat(s string) (float64, error) {
	return strconv.ParseFloat(s, 64)
}
