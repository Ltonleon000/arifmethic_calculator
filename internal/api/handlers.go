package api

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"calculator/internal/calculator"
	"calculator/internal/models"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type Handler struct {
	expressions sync.Map
	tasks       sync.Map
	taskQueue   chan models.Task
}

func NewHandler() *Handler {
	return &Handler{
		taskQueue: make(chan models.Task, 100),
	}
}

func (h *Handler) CalculateHandler(w http.ResponseWriter, r *http.Request) {

	var req models.CalculationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Плохой запрос", http.StatusBadRequest)
		return
	}

	id := uuid.New().String()
	expr := &models.Expression{
		ID:         id,
		Expression: req.Expression,
		Status:     models.StatusPending,
	}
	h.expressions.Store(id, expr)

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

	var expressions []models.Expression
	h.expressions.Range(func(key, value interface{}) bool {
		expressions = append(expressions, *value.(*models.Expression))
		return true
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"expressions": expressions,
	})
}

func (h *Handler) GetExpressionHandler(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Access-Control-Allow-Origin", "*")

	vars := mux.Vars(r)
	id := vars["id"]

	expr, ok := h.expressions.Load(id)
	if !ok {
		http.Error(w, "Expression not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"expression": expr,
	})
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
