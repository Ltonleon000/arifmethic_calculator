package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"calculator/internal/models"
)

func TestCalculateHandler(t *testing.T) {
	handler := NewHandler()

	// Создаем тестовый запрос
	reqBody := models.CalculationRequest{
		Expression: "2+2",
	}
	body, _ := json.Marshal(reqBody)
	req, err := http.NewRequest("POST", "/api/v1/calculate", bytes.NewBuffer(body))
	if err != nil {
		t.Fatal(err)
	}

	// Создаем ResponseRecorder (реализация ResponseWriter) для записи ответа
	rr := httptest.NewRecorder()

	// Вызываем обработчик напрямую, передавая ResponseRecorder и Request
	handler.CalculateHandler(rr, req)

	// Проверяем код состояния
	if status := rr.Code; status != http.StatusCreated {
		t.Errorf("Ошибка: неверный код состояния: получено %v, ожидалось %v",
			status, http.StatusCreated)
	}

	// Проверяем тело ответа
	var response map[string]string
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	if err != nil {
		t.Fatal(err)
	}

	// Проверяем, что идентификатор выражения был возвращен
	if _, exists := response["id"]; !exists {
		t.Errorf("ошибка: ответ не содержит идентификатор")
	}
}

func TestGetTaskHandler_NoTasks(t *testing.T) {
	handler := NewHandler()

	// Создаем тестовый запрос для получения задачи, когда их нет
	req, err := http.NewRequest("GET", "/api/v1/task", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler.GetTaskHandler(rr, req)

	// Проверяем, что получаем код 404, так как задач нет
	if status := rr.Code; status != http.StatusNotFound {
		t.Errorf("ошибка: неверный код состояния: получено %v, ожидалось %v",
			status, http.StatusNotFound)
	}
}

func TestGetExpressionsHandler_Empty(t *testing.T) {
	handler := NewHandler()

	// Создаем тестовый запрос
	req, err := http.NewRequest("GET", "/api/v1/expressions", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler.GetExpressionsHandler(rr, req)

	// Проверяем код состояния
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("ошибка: неверный код состояния: получено %v, ожидалось %v",
			status, http.StatusOK)
	}

	// Выводим тело ответа для отладки
	t.Logf("Тело ответа: %s", rr.Body.String())

	// Проверяем тело ответа - объект с полем expressions
	var response map[string]interface{}
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Не удалось разобрать JSON ответ: %v. Тело ответа: %s", err, rr.Body.String())
	}

	// Проверяем, что поле expressions существует
	expressions, exists := response["expressions"]
	if !exists {
		t.Errorf("Ответ должен содержать поле 'expressions'")
		return
	}

	// Проверяем, что expressions является массивом или null
	if expressions != nil {
		exprs, ok := expressions.([]interface{})
		if !ok {
			t.Errorf("expressions должно быть массивом, получен тип: %T", expressions)
		} else if len(exprs) != 0 {
			t.Errorf("expressions должен быть пустым, получено %d элементов", len(exprs))
		}
	}
}
