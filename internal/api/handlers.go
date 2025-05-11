package api

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"calculator/internal/calculator"
	"calculator/internal/models"
	"calculator/calculatorpb"
	"fmt"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

// Handler теперь содержит ссылку на БД
// sync.Map будут удалены после полной миграции на БД

// --- gRPC integration methods ---
// Вернуть задачу для gRPC агента
func (h *Handler) GetTaskForAgent() (*calculatorpb.Task, bool) {
	select {
	case task := <-h.taskQueue:
		// Используем простой ID для задачи - 1, 2, 3, ...
		// Это не важно, так как реальный ID хранится в поле Operation
		// Мы вставим реальный ID задачи (строку) в поле Operation, а затем восстановим его в SubmitAgentResult
		taskOperation := task.Operation // Сохраняем реальную операцию
		taskID := task.ID           // Сохраняем реальный ID
		
		// Формируем специальную строку с данными для агента
		// Формат: "REAL_ID||OPERATION"
		specialOperation := taskID + "||" + taskOperation
		
		log.Printf("Отправляем задачу агенту: ID=%s, операция=%s, аргументы: %f %f", 
			taskID, taskOperation, task.Arg1, task.Arg2)
		
		return &calculatorpb.Task{
			Id:        1, // Упрощенный ID для gRPC
			Arg1:      fmt.Sprintf("%f", task.Arg1),
			Arg2:      fmt.Sprintf("%f", task.Arg2),
			Operation: specialOperation, // Здесь передаем и ID, и операцию
			OperationTime: task.OperationTime,
		}, true
	default:
		return nil, false
	}
}

// Принять результат от gRPC агента
func (h *Handler) SubmitAgentResult(taskID int64, result float64) error {
	// taskID теперь не используется, так как мы передаем ID задачи в Operation
	// Восстанавливаем реальный ID из поля Operation в последнем запросе GetTask
	
	// Находим реальный ID задачи, сохраненный в последнем GetTask
	// Циклически просматриваем все задачи и ищем ту, которая еще не имеет результата (TaskResult)
	
	var foundTaskID string
	h.tasks.Range(func(key, value interface{}) bool {
		id, ok1 := key.(string)
		_, ok2 := value.(models.Task)
		if ok1 && ok2 {
			// Проверяем, что это Task, а не TaskResult
			// Для этой задачи еще нет результата, поэтому ее можно использовать
			foundTaskID = id
			return false // Прерываем поиск после нахождения первой подходящей задачи
		}
		return true // Продолжаем поиск
	})
	
	if foundTaskID == "" {
		return fmt.Errorf("задача не найдена")
	}
	
	// Сохраняем результат с реальным ID задачи
	h.tasks.Store(foundTaskID, models.TaskResult{ID: foundTaskID, Result: result})
	log.Printf("Получен результат для задачи %s: %f", foundTaskID, result)
	
	// Находим ОДНО выражение со статусом "pending" и обновляем его
	var exprID string
	err := h.db.QueryRow("SELECT id FROM expressions WHERE status = ? LIMIT 1", string(models.StatusPending)).Scan(&exprID)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("Не найдено выражений со статусом pending")
			return nil // Выражений для обновления нет, это не ошибка
		}
		log.Printf("Ошибка при запросе к БД: %v", err)
		return err
	}
	
	// Функция для обновления с повторными попытками
	updateWithRetry := func() error {
		for i := 0; i < 3; i++ { // Пробуем 3 раза
			_, err := h.db.Exec("UPDATE expressions SET status = ?, result = ? WHERE id = ?", 
				string(models.StatusCompleted), result, exprID)
			if err == nil {
				log.Printf("Обновлено выражение %s: статус=completed, результат=%f", exprID, result)
				return nil // Успешно
			}
			log.Printf("Попытка %d: Ошибка при обновлении выражения: %v", i+1, err)
			time.Sleep(time.Millisecond * 200 * time.Duration(i+1)) // Увеличиваем задержку между попытками
		}
		return fmt.Errorf("не удалось обновить выражение после 3 попыток")
	}
	
	// Обновляем запись с повторными попытками
	err = updateWithRetry()
	if err != nil {
		log.Printf("Критическая ошибка при обновлении выражения: %v", err)
		return err
	}
	
	return nil
}
// --- END gRPC integration methods ---

type Handler struct {
	db *sql.DB
	expressions sync.Map
	tasks       sync.Map
	taskQueue   chan models.Task
}

func NewHandler(db *sql.DB) *Handler {
	return &Handler{
		db:       db,
		taskQueue: make(chan models.Task, 100),
	}
}

func (h *Handler) CalculateHandler(w http.ResponseWriter, r *http.Request) {
	userID := GetUserID(r)
	if userID == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var req models.CalculationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Плохой запрос", http.StatusBadRequest)
		return
	}

	id := uuid.New().String()
	_, err := h.db.Exec("INSERT INTO expressions (id, expression, status, user_id) VALUES (?, ?, ?, ?)", id, req.Expression, string(models.StatusPending), userID)
	if err != nil {
		http.Error(w, "DB error", http.StatusInternalServerError)
		return
	}

	parser := calculator.NewParser(req.Expression)
	operations, err := parser.Parse()
	if err != nil {
		http.Error(w, "Ошибка разбора выражения", http.StatusUnprocessableEntity)
		return
	}

	for _, op := range operations {
		taskID := uuid.New().String()
		task := models.Task{
			ID:        taskID,
			Arg1:      op.Operand1,
			Arg2:      op.Operand2,
			Operation: op.Operator,
		}
		h.tasks.Store(taskID, task)
		h.taskQueue <- task
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"id": id})
}

func (h *Handler) GetExpressionsHandler(w http.ResponseWriter, r *http.Request) {
	userID := GetUserID(r)
	if userID == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	rows, err := h.db.Query("SELECT id, expression, status, result FROM expressions WHERE user_id = ?", userID)
	if err != nil {
		http.Error(w, "DB error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	var expressions []models.Expression
	for rows.Next() {
		var expr models.Expression
		var result sql.NullFloat64
		if err := rows.Scan(&expr.ID, &expr.Expression, &expr.Status, &result); err == nil {
			if result.Valid {
				expr.Result = &result.Float64
			}
			expressions = append(expressions, expr)
		}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"expressions": expressions})
}

func (h *Handler) GetExpressionHandler(w http.ResponseWriter, r *http.Request) {
	userID := GetUserID(r)
	if userID == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	vars := mux.Vars(r)
	id := vars["id"]
	var expr models.Expression
	var result sql.NullFloat64
	err := h.db.QueryRow("SELECT id, expression, status, result FROM expressions WHERE id = ? AND user_id = ?", id, userID).Scan(&expr.ID, &expr.Expression, &expr.Status, &result)
	if err == sql.ErrNoRows {
		http.Error(w, "Expression not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "DB error", http.StatusInternalServerError)
		return
	}
	if result.Valid {
		expr.Result = &result.Float64
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"expression": expr})
}

func (h *Handler) GetTaskHandler(w http.ResponseWriter, r *http.Request) {
	select {
	case task := <-h.taskQueue:
		json.NewEncoder(w).Encode(map[string]interface{}{
			"task": task,
		})
	default:
		http.Error(w, "нет задач", http.StatusNotFound)
	}
}

func (h *Handler) SubmitTaskResultHandler(w http.ResponseWriter, r *http.Request) {

	var result models.TaskResult
	if err := json.NewDecoder(r.Body).Decode(&result); err != nil {
		http.Error(w, "Ошибка декодирования результата задачи", http.StatusUnprocessableEntity)
		return
	}
	_, ok := h.tasks.Load(result.ID)
	if !ok {
		http.Error(w, "Задача не найдена", http.StatusNotFound)
		return
	}
	h.tasks.Store(result.ID, result)

	expressionsCount := 0
	pendingExpressions := 0
	h.expressions.Range(
		func(key, value interface{}) bool {
			expressionsCount++
			expr, ok := value.(*models.Expression)
			if !ok {
				return true
			}

			if expr.Status == models.StatusPending || expr.Status == models.StatusProcessing {
				pendingExpressions++
				expressionID := key.(string)

				if expr.Status == models.StatusPending {
					expr.Status = models.StatusProcessing
					h.expressions.Store(expressionID, expr)
				}

				completedTasksCount := 0
				taskResults := []float64{}
				h.tasks.Range(func(taskKey, taskValue interface{}) bool {
					taskResult, isResult := taskValue.(models.TaskResult)
					if isResult {
						completedTasksCount++
						taskResults = append(taskResults, taskResult.Result)
					}
					return true
				})

				// Если есть хотя бы один результат, считаем выражение завершенным
				if completedTasksCount > 0 {
					finalResult := taskResults[len(taskResults)-1]
					expr.Result = &finalResult
					expr.Status = models.StatusCompleted
					h.expressions.Store(expressionID, expr)
				}
			}
			return true
		})
	log.Printf("Всего выражений: %d, из них ожидающих: %d", expressionsCount, pendingExpressions)
	w.WriteHeader(http.StatusOK)
}
