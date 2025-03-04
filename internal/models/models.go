package models

type CalculationStatus string

const (
	StatusPending    CalculationStatus = "pending"    // ждем выполнения
	StatusProcessing CalculationStatus = "processing" // выполняется
	StatusCompleted  CalculationStatus = "completed"  // ура, готово
)

type Expression struct {
	ID         string            `json:"id"`
	Expression string            `json:"expression,omitempty"`
	Status     CalculationStatus `json:"status"`
	Result     *float64          `json:"result,omitempty"`
}

type CalculationRequest struct {
	Expression string `json:"expression"`
}

type Task struct {
	ID            string  `json:"id"`
	Arg1          float64 `json:"arg1"`
	Arg2          float64 `json:"arg2"`
	Operation     string  `json:"operation"`
	OperationTime int64   `json:"operation_time"`
}

type TaskResult struct {
	ID     string  `json:"id"`
	Result float64 `json:"result"`
}
