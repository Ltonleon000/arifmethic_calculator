package main

import (
	"calculator/internal"
	"calculator/internal/api"
	"calculator/internal/models"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
)

var (
	timeAdditionMS       int64
	timeSubtractionMS    int64
	timeMultiplicationMS int64
	timeDivisionMS       int64
	port                 string
)

// getEnv возвращает значение переменной окружения или значение по умолчанию
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists && value != "" {
		return value
	}
	return defaultValue
}

func init() {
	var err error

	// Инициализация порогов времени выполнения операций
	timeAdditionMS, err = strconv.ParseInt(getEnv("TIME_ADDITION_MS", "1000"), 10, 64) // время выполнения сложения
	if err != nil {
		timeAdditionMS = 1000 // по умолчанию 1 секунда
	}

	timeSubtractionMS, err = strconv.ParseInt(getEnv("TIME_SUBTRACTION_MS", "1000"), 10, 64) // время выполнения вычитания
	if err != nil {
		timeSubtractionMS = 1000
	}

	timeMultiplicationMS, err = strconv.ParseInt(getEnv("TIME_MULTIPLICATIONS_MS", "2000"), 10, 64) // время выполнения умножения
	if err != nil {
		timeMultiplicationMS = 2000
	}

	timeDivisionMS, err = strconv.ParseInt(getEnv("TIME_DIVISIONS_MS", "2000"), 10, 64) // время выполнения деления
	if err != nil {
		timeDivisionMS = 2000
	}

	// Получаем порт из переменной файла настроек
	port = getEnv("ORCHESTRATOR_PORT", "8080")
}

// добавим поддержку CORS чтобы браузер мог обращаться к оркестратору
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// Добавляем CORS-заголовки только для публичных API, для агентов cors не нужен
		if r.URL.Path == "/" || r.URL.Path == "/calculator" || r.URL.Path == "/favicon.ico" {
			next.ServeHTTP(w, r)
			return
		}

		// Упрощаем проверку - добавляем CORS заголовки ко всем запросам к /api/
		if strings.HasPrefix(r.URL.Path, "/api/") {

			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}

func main() {

	db, err := internal.OpenDB("arifmethic.db")
	if err != nil {
		log.Fatalf("Ошибка открытия БД: %v", err)
	}
	err = models.Migrate(db)
	if err != nil {
		log.Fatalf("Ошибка миграции БД: %v", err)
	}

	r := mux.NewRouter()
	handler := api.NewHandler(db)

	// Запуск gRPC сервера для агентов на другом порту, чтобы избежать конфликта
	grpcPort := "8082" // Используем порт 8082 для gRPC
	go internal.StartGRPCServer(
		handler.GetTaskForAgent,   // функция получения задачи
		handler.SubmitAgentResult, // функция отправки результата
		grpcPort,                  // порт для gRPC сервера
	)
	log.Printf("Запускаем gRPC сервер на порту %s", grpcPort)

	// Регистрация и логин
	r.HandleFunc("/api/v1/register", api.RegisterHandler(db)).Methods("POST")
	r.HandleFunc("/api/v1/login", api.LoginHandler(db)).Methods("POST")

	// Публичные API endpoints (требуют JWT)
	r.Handle("/api/v1/calculate", api.JWTMiddleware(http.HandlerFunc(handler.CalculateHandler))).Methods("POST", "OPTIONS")
	r.Handle("/api/v1/expressions", api.JWTMiddleware(http.HandlerFunc(handler.GetExpressionsHandler))).Methods("GET", "OPTIONS")
	r.Handle("/api/v1/expressions/{id}", api.JWTMiddleware(http.HandlerFunc(handler.GetExpressionHandler))).Methods("GET", "OPTIONS")

	// Внутренние API endpoints для агентов
	r.HandleFunc("/internal/task", handler.GetTaskHandler).Methods("GET")
	r.HandleFunc("/internal/task", handler.SubmitTaskResultHandler).Methods("POST")

	// Обработчик для калькулятора на корневом пути
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "calculator.html")
	}).Methods("GET")

	// Применяем CORS middleware ко всем маршрутам
	corsRouter := corsMiddleware(r)

	// Формируем адрес для прослушивания, HTTP сервер на порту 8081, поскольку gRPC уже использует порт 8080
	httpPort := "8081"
	listenAddr := fmt.Sprintf(":%s", httpPort)
	log.Printf("Запускаем HTTP сервер оркестратора на порту %s", listenAddr)

	if err := http.ListenAndServe(listenAddr, corsRouter); err != nil {
		log.Fatal(err)
	}
}
