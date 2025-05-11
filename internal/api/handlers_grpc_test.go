package api

import (
	"calculator/internal/models"
	"database/sql"
	"sync"
	"testing"
)

// Тест для обработки задач
func TestTaskProcessing(t *testing.T) {
	// Создаем тестовый обработчик
	h := NewHandler()

	// Добавляем тестовую задачу в хранилище
	task := models.Task{
		ID:            "test-task-123",
		Arg1:          10.5,
		Arg2:          20.5,
		Operation:     "+",
		OperationTime: 1000,
	}
	
	// Сохраняем задачу
	h.SaveTask(&task)

	// Проверяем, что задача есть в хранилище
	savedTask, ok := h.GetTask(task.ID)
	if !ok {
		t.Fatalf("Задача не найдена в хранилище после сохранения")
	}

	// Проверяем, что данные в задаче корректные
	if savedTask.ID != task.ID {
		t.Errorf("Неверный ID в задаче, ожидалось %s, получено %s", task.ID, savedTask.ID)
	}

	if savedTask.Arg1 != task.Arg1 {
		t.Errorf("Неверный Arg1, ожидалось %f, получено %f", task.Arg1, savedTask.Arg1)
	}

	if savedTask.Arg2 != task.Arg2 {
		t.Errorf("Неверный Arg2, ожидалось %f, получено %f", task.Arg2, savedTask.Arg2)
	}

	if savedTask.Operation != task.Operation {
		t.Errorf("Неверная операция, ожидалось %s, получено %s", task.Operation, savedTask.Operation)
	}
}

// Тест для функции SubmitAgentResult
func TestSubmitAgentResult(t *testing.T) {
	// Создаем тестовый обработчик
	h := &Handler{
		tasks:     sync.Map{},
		taskQueue: make(chan models.Task, 10),
	}

	// Добавляем тестовую задачу в хранилище
	taskID := "test-task-456"
	h.tasks.Store(taskID, models.Task{
		ID:        taskID,
		Arg1:      30.0,
		Arg2:      15.0,
		Operation: "-",
	})

	// Вызываем функцию, которую тестируем
	err := h.SubmitAgentResult(789, 15.0) // ID не важен, так как мы его не используем в нашей реализации

	// Проверяем, что функция не вернула ошибку
	if err != nil {
		t.Errorf("Ожидалось успешное выполнение, но получена ошибка: %v", err)
	}

	// Проверяем, что результат был сохранен в хранилище
	result, ok := h.tasks.Load(taskID)
	if !ok {
		t.Errorf("Задача %s не найдена в хранилище после сохранения результата", taskID)
		return
	}

	// Проверяем тип и содержимое результата
	taskResult, ok := result.(models.TaskResult)
	if !ok {
		t.Errorf("Результат имеет неверный тип: %T, ожидался models.TaskResult", result)
		return
	}

	if taskResult.ID != taskID {
		t.Errorf("Неверный ID в результате, ожидалось %s, получено %s", taskID, taskResult.ID)
	}

	if taskResult.Result != 15.0 {
		t.Errorf("Неверный результат, ожидалось 15.0, получено %f", taskResult.Result)
	}
}
