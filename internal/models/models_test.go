package models

import (
	"encoding/json"
	"testing"
)

func TestExpressionMarshaling(t *testing.T) {
	result := 42.0
	expr := Expression{
		ID:         "test-id",
		Expression: "2+2*20",
		Status:     StatusCompleted,
		Result:     &result,
	}

	data, err := json.Marshal(expr)
	if err != nil {
		t.Fatalf("ошибка маршалинга Expression: %v", err)
	}

	var unmarshaled Expression
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("ошибка демаршалинга Expression: %v", err)
	}

	if unmarshaled.ID != expr.ID {
		t.Errorf("ошибка: получено %s, ожидалось %s", unmarshaled.ID, expr.ID)
	}

	if unmarshaled.Expression != expr.Expression {
		t.Errorf("ошибка: получено %s, ожидалось %s", unmarshaled.Expression, expr.Expression)
	}

	if unmarshaled.Status != expr.Status {
		t.Errorf("ошибка: получено %s, ожидалось %s", unmarshaled.Status, expr.Status)
	}

	if *unmarshaled.Result != *expr.Result {
		t.Errorf("ошибка: получено %f, ожидалось %f", *unmarshaled.Result, *expr.Result)
	}
}

func TestTaskMarshaling(t *testing.T) {
	// Тест маршалинга Task
	task := Task{
		ID:            "task-id",
		Arg1:          2.5,
		Arg2:          3.5,
		Operation:     "+",
		OperationTime: 1000,
	}

	data, err := json.Marshal(task)
	if err != nil {
		t.Fatalf("ошибка маршалинга Task: %v", err)
	}

	var unmarshaled Task
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("ошибка демаршалинга Task: %v", err)
	}

	if unmarshaled.ID != task.ID {
		t.Errorf("ошибка: получено %s, ожидалось %s", unmarshaled.ID, task.ID)
	}

	if unmarshaled.Arg1 != task.Arg1 {
		t.Errorf("ошибка: получено %f, ожидалось %f", unmarshaled.Arg1, task.Arg1)
	}

	if unmarshaled.Arg2 != task.Arg2 {
		t.Errorf("ошибка: получено %f, ожидалось %f", unmarshaled.Arg2, task.Arg2)
	}

	if unmarshaled.Operation != task.Operation {
		t.Errorf("ошибка: получено %s, ожидалось %s", unmarshaled.Operation, task.Operation)
	}

	if unmarshaled.OperationTime != task.OperationTime {
		t.Errorf("Unmarshaled OperationTime = %d, want %d", unmarshaled.OperationTime, task.OperationTime)
	}
}

func TestTaskResultMarshaling(t *testing.T) {
	result := TaskResult{
		ID:     "task-id",
		Result: 6.0,
	}

	data, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("ошибка маршалинга TaskResult: %v", err)
	}

	var unmarshaled TaskResult
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("ошибка демаршалинга TaskResult: %v", err)
	}

	if unmarshaled.ID != result.ID {
		t.Errorf("ошибка: получено %s, ожидалось %s", unmarshaled.ID, result.ID)
	}

	if unmarshaled.Result != result.Result {
		t.Errorf("ошибка: получено %f, ожидалось %f", unmarshaled.Result, result.Result)
	}
}

func TestCalculationStatus(t *testing.T) {
	// Проверка констант статусов
	if StatusPending != "pending" {
		t.Errorf("статус StatusPending = %s, ожидалось 'pending'", StatusPending)
	}

	if StatusProcessing != "processing" {
		t.Errorf("статус StatusProcessing = %s, ожидалось 'processing'", StatusProcessing)
	}

	if StatusCompleted != "completed" {
		t.Errorf("статус StatusCompleted = %s, ожидалось 'completed'", StatusCompleted)
	}
}
