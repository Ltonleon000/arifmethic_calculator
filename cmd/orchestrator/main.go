package main

import (
	"calculator/internal/api"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

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

		// Проверяем, является ли запрос к публичному API и в этом случае добавляем CORS-заголовки
		if r.URL.Path == "/api/v1/calculate" ||
			r.URL.Path == "/api/v1/expressions" ||
			r.URL.Path == "/api/v1/expressions/" ||
			len(r.URL.Path) > 18 && r.URL.Path[:18] == "/api/v1/expressions/" {

			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}

func main() {
	r := mux.NewRouter()
	handler := api.NewHandler()

	// Публичные API endpoints
	r.HandleFunc("/api/v1/calculate", handler.CalculateHandler).Methods("POST", "OPTIONS")
	r.HandleFunc("/api/v1/expressions", handler.GetExpressionsHandler).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/v1/expressions/{id}", handler.GetExpressionHandler).Methods("GET", "OPTIONS")

	// Внутренние API endpoints для агентов
	r.HandleFunc("/internal/task", handler.GetTaskHandler).Methods("GET")
	r.HandleFunc("/internal/task", handler.SubmitTaskResultHandler).Methods("POST")

	// Обработчик для калькулятора на корневом пути
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "calculator.html")
	}).Methods("GET")

	// Применяем CORS middleware ко всем маршрутам
	corsRouter := corsMiddleware(r)

	// Формируем адрес для прослушивания
	listenAddr := fmt.Sprintf(":%s", port)
	log.Printf("Запускаем оркестратор на порту %s", listenAddr)

	if err := http.ListenAndServe(listenAddr, corsRouter); err != nil {
		log.Fatal(err)
	}
}
