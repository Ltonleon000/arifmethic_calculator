package api

import (
	"calculator/internal/calculator"
	"calculator/internal/models"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
)

func setupTestHandler() (*Handler, *mux.Router) {
	// Создаем тестовый обработчик без подключения к БД
	h := &Handler{
		calculator: calculator.NewCalculator(),
		JWTSecret:  []byte("test-secret"),
	}

	// Настраиваем роутер
	r := mux.NewRouter()
	h.RegisterRoutes(r.PathPrefix("/api").Subrouter())
	return h, r
}

// Тестирование аутентификации
func TestAuthentication(t *testing.T) {
	h, r := setupTestHandler()

	// Тест регистрации пользователя
	t.Run("RegisterUser", func(t *testing.T) {
		reqBody := `{"username":"testuser","password":"testpass"}`
		req, _ := http.NewRequest("POST", "/api/register", strings.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("Неверный код ответа: получено %v, ожидалось %v", status, http.StatusOK)
		}

		var resp struct {
			Token string `json:"token"`
		}
		if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
			t.Errorf("Не удалось распарсить JSON ответ: %v", err)
		}

		if resp.Token == "" {
			t.Error("Токен не получен в ответе")
		}
	})

	// Тест авторизации пользователя
	t.Run("LoginUser", func(t *testing.T) {
		// Сначала создаем пользователя
		username := "logintest"
		password := "loginpass"
		h.users[username] = models.User{
			Username: username,
			Password: password,
		}

		reqBody := fmt.Sprintf(`{"username":"%s","password":"%s"}`, username, password)
		req, _ := http.NewRequest("POST", "/api/login", strings.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("Неверный код ответа: получено %v, ожидалось %v", status, http.StatusOK)
		}

		var resp struct {
			Token string `json:"token"`
		}
		if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
			t.Errorf("Не удалось распарсить JSON ответ: %v", err)
		}

		if resp.Token == "" {
			t.Error("Токен не получен в ответе")
		}

		// Проверяем, что токен валидный
		token, err := jwt.Parse(resp.Token, func(token *jwt.Token) (interface{}, error) {
			return h.JWTSecret, nil
		})
		if err != nil {
			t.Errorf("Ошибка при парсинге токена: %v", err)
		}

		if !token.Valid {
			t.Error("Полученный токен не валидный")
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			t.Error("Не удалось получить claims из токена")
		}

		if claims["username"] != username {
			t.Errorf("Неверное имя пользователя в токене: ожидалось %s, получено %v", username, claims["username"])
		}
	})
}

// Тестирование API для вычисления выражений
func TestCalculate(t *testing.T) {
	h, r := setupTestHandler()

	// Создаем тестового пользователя и генерируем токен
	username := "calcuser"
	h.users[username] = models.User{
		Username: username,
		Password: "calcpass",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": username,
		"exp":      time.Now().Add(time.Hour * 24).Unix(),
	})
	tokenString, _ := token.SignedString(h.JWTSecret)

	// Тест вычисления выражения
	t.Run("CalculateExpression", func(t *testing.T) {
		reqBody := `{"expression":"5 + 3"}`
		req, _ := http.NewRequest("POST", "/api/calculate", strings.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+tokenString)

		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("Неверный код ответа: получено %v, ожидалось %v", status, http.StatusOK)
		}

		var resp models.Expression
		if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
			t.Errorf("Не удалось распарсить JSON ответ: %v", err)
		}

		if resp.Expression != "5 + 3" {
			t.Errorf("Неверное выражение в ответе: ожидалось '5 + 3', получено '%s'", resp.Expression)
		}

		if resp.Status != models.StatusPending && resp.Status != models.StatusCompleted {
			t.Errorf("Неверный статус: ожидался 'pending' или 'completed', получен '%s'", resp.Status)
		}
	})

	// Тест получения результатов вычисления
	t.Run("GetResults", func(t *testing.T) {
		// Добавляем тестовый результат
		result := 42.0
		testExpr := models.Expression{
			ID:         "test-expr-id",
			Expression: "6 * 7",
			Status:     models.StatusCompleted,
			Result:     &result,
			UserID:     username,
		}
		h.expressions[testExpr.ID] = testExpr

		req, _ := http.NewRequest("GET", "/api/results", nil)
		req.Header.Set("Authorization", "Bearer "+tokenString)

		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("Неверный код ответа: получено %v, ожидалось %v", status, http.StatusOK)
		}

		var resp []models.Expression
		if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
			t.Errorf("Не удалось распарсить JSON ответ: %v", err)
		}

		if len(resp) < 1 {
			t.Error("Не получены результаты выражений")
			return
		}

		found := false
		for _, expr := range resp {
			if expr.ID == testExpr.ID {
				found = true
				if expr.Expression != testExpr.Expression {
					t.Errorf("Неверное выражение: ожидалось '%s', получено '%s'", testExpr.Expression, expr.Expression)
				}
				if expr.Status != models.StatusCompleted {
					t.Errorf("Неверный статус: ожидался 'completed', получен '%s'", expr.Status)
				}
				if *expr.Result != *testExpr.Result {
					t.Errorf("Неверный результат: ожидалось %f, получено %f", *testExpr.Result, *expr.Result)
				}
				break
			}
		}

		if !found {
			t.Errorf("Не найдено тестовое выражение с ID '%s'", testExpr.ID)
		}
	})
}

// Тестирование API с неверными данными
func TestInvalidInput(t *testing.T) {
	_, r := setupTestHandler()

	// Тест неверного выражения
	t.Run("InvalidExpression", func(t *testing.T) {
		reqBody := `{"expression":"5 + "}`
		req, _ := http.NewRequest("POST", "/api/calculate", strings.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer validtoken")

		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusBadRequest {
			t.Errorf("Неверный код ответа: получено %v, ожидалось %v", status, http.StatusBadRequest)
		}
	})

	// Тест отсутствия авторизации
	t.Run("Unauthorized", func(t *testing.T) {
		reqBody := `{"expression":"5 + 3"}`
		req, _ := http.NewRequest("POST", "/api/calculate", strings.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")
		// Не устанавливаем токен авторизации

		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusUnauthorized {
			t.Errorf("Неверный код ответа: получено %v, ожидалось %v", status, http.StatusUnauthorized)
		}
	})
}
