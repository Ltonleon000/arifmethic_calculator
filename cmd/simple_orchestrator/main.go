package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"calculator/internal/auth"
	"calculator/internal/calculatorpb"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Простая структура для хранения выражения
type Expression struct {
	ID         string  `json:"id"`
	Expression string  `json:"expression"`
	Status     string  `json:"status"`
	Result     float64 `json:"result,omitempty"`
}

// Простой API-сервер с поддержкой CORS и авторизацией
func main() {
	log.Println("Запуск упрощенного оркестратора с авторизацией")

	// Создаем подключение к базе данных SQLite
	db, err := sql.Open("sqlite3", "./simple.db")
	if err != nil {
		log.Fatalf("Ошибка открытия БД: %v", err)
	}
	defer db.Close()

	// Создаем таблицу пользователей
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS users (
		id TEXT PRIMARY KEY,
		login TEXT UNIQUE NOT NULL,
		password TEXT NOT NULL
	)`)
	if err != nil {
		log.Fatalf("Ошибка создания таблицы пользователей: %v", err)
	}

	// Создаем таблицу выражений
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS expressions (
		id TEXT PRIMARY KEY,
		user_id TEXT NOT NULL,
		expression TEXT NOT NULL,
		status TEXT NOT NULL,
		result REAL NULL,
		FOREIGN KEY (user_id) REFERENCES users(id)
	)`)
	if err != nil {
		log.Fatalf("Ошибка создания таблицы выражений: %v", err)
	}

	r := mux.NewRouter()

	// Настройка CORS
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	})

	// Публичные маршруты (API без авторизации)
	public := r.PathPrefix("/api/v1").Subrouter()
	
	// Регистрация пользователя
	public.HandleFunc("/register", auth.RegisterHandler(db)).Methods("POST")
	
	// Вход пользователя
	public.HandleFunc("/login", auth.LoginHandler(db)).Methods("POST")

	// Защищенные маршруты (требуют JWT-токен)
	protected := r.PathPrefix("/api/v1").Subrouter()
	protected.Use(auth.JWTMiddleware)

	// Эндпоинт для отправки выражения (требует авторизацию)
	protected.HandleFunc("/calculate", func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Expression string `json:"expression"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Неверный запрос", http.StatusBadRequest)
			return
		}

		// Получаем ID пользователя из токена
		userID := auth.GetUserID(r)
		if userID == "" {
			http.Error(w, "Не удалось получить ID пользователя", http.StatusUnauthorized)
			return
		}

		// Создаем новое выражение
		id := uuid.New().String()
		expr := Expression{
			ID:         id,
			Expression: req.Expression,
			Status:     "pending",
		}

		// Сохраняем в БД с привязкой к пользователю
		_, err := db.Exec("INSERT INTO expressions (id, user_id, expression, status) VALUES (?, ?, ?, ?)",
			expr.ID, userID, expr.Expression, expr.Status)
		if err != nil {
			log.Printf("Ошибка сохранения: %v", err)
			http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
			return
		}

		// Запускаем вычисление в отдельной горутине
		go calculateExpression(expr.ID, expr.Expression, db)

		// Возвращаем ID выражения
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"id": expr.ID})
	}).Methods("POST")

	// Получить выражения текущего пользователя
	protected.HandleFunc("/expressions", func(w http.ResponseWriter, r *http.Request) {
		// Получаем ID пользователя из токена
		userID := auth.GetUserID(r)
		if userID == "" {
			http.Error(w, "Не удалось получить ID пользователя", http.StatusUnauthorized)
			return
		}

		// Получаем выражения только для текущего пользователя
		rows, err := db.Query("SELECT id, expression, status, result FROM expressions WHERE user_id = ?", userID)
		if err != nil {
			http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var expressions []Expression
		for rows.Next() {
			var expr Expression
			var result sql.NullFloat64
			if err := rows.Scan(&expr.ID, &expr.Expression, &expr.Status, &result); err != nil {
				log.Printf("Ошибка сканирования: %v", err)
				continue
			}
			if result.Valid {
				expr.Result = result.Float64
			}
			expressions = append(expressions, expr)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string][]Expression{"expressions": expressions})
	}).Methods("GET")

	// Получить конкретное выражение (требует авторизацию)
	protected.HandleFunc("/expressions/{id}", func(w http.ResponseWriter, r *http.Request) {
		// Получаем ID пользователя из токена
		userID := auth.GetUserID(r)
		if userID == "" {
			http.Error(w, "Не удалось получить ID пользователя", http.StatusUnauthorized)
			return
		}

		id := mux.Vars(r)["id"]
		if id == "" {
			http.Error(w, "Не указан ID", http.StatusBadRequest)
			return
		}

		var expr Expression
		var result sql.NullFloat64

		// Проверяем, что выражение принадлежит текущему пользователю
		err := db.QueryRow("SELECT id, expression, status, result FROM expressions WHERE id = ? AND user_id = ?", id, userID).
			Scan(&expr.ID, &expr.Expression, &expr.Status, &result)
		if err != nil {
			if err == sql.ErrNoRows {
				http.Error(w, "Выражение не найдено", http.StatusNotFound)
			} else {
				http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
			}
			return
		}

		if result.Valid {
			expr.Result = result.Float64
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]Expression{"expression": expr})
	}).Methods("GET")

	// Запускаем HTTP сервер
	port := 8081
	log.Printf("HTTP сервер запущен на порту %d", port)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), r); err != nil {
		log.Fatalf("Ошибка запуска HTTP сервера: %v", err)
	}
}

// Функция для вычисления выражения через gRPC
func calculateExpression(id, expression string, db *sql.DB) {
	log.Printf("Отправка выражения '%s' на вычисление (ID: %s)", expression, id)

	// Устанавливаем соединение с gRPC-сервером
	conn, err := grpc.Dial("localhost:8082", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Printf("Ошибка подключения к gRPC: %v", err)
		updateStatus(id, "error", 0, db)
		return
	}
	defer conn.Close()

	// Создаем клиент
	client := calculatorpb.NewCalculatorServiceClient(conn)

	// Подготавливаем запрос
	req := &calculatorpb.CalculateRequest{
		Expression: expression,
	}

	// Устанавливаем таймаут 10 секунд
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Отправляем запрос
	resp, err := client.Calculate(ctx, req)
	if err != nil {
		log.Printf("Ошибка вызова gRPC: %v", err)
		updateStatus(id, "error", 0, db)
		return
	}

	// Получаем результат
	result, err := strconv.ParseFloat(resp.Result, 64)
	if err != nil {
		log.Printf("Ошибка парсинга результата: %v", err)
		updateStatus(id, "error", 0, db)
		return
	}

	// Обновляем статус в БД
	updateStatus(id, "completed", result, db)
	log.Printf("Вычисление завершено: %s = %f", expression, result)
}

// Обновить статус выражения в БД
func updateStatus(id, status string, result float64, db *sql.DB) {
	_, err := db.Exec("UPDATE expressions SET status = ?, result = ? WHERE id = ?", status, result, id)
	if err != nil {
		log.Printf("Ошибка обновления статуса: %v", err)
	} else {
		log.Printf("Статус выражения %s обновлен на %s, результат: %f", id, status, result)
	}
}
