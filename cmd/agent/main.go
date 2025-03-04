package main

import (
	"bytes"
	"calculator/internal/models"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"
)

var (
	computingPower  int
	orchestratorURL string
)

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists && value != "" {
		return value
	}
	return defaultValue
}

func main() {
	var err error
	computingPowerStr := getEnv("COMPUTING_POWER", "2")
	computingPower, err = strconv.Atoi(computingPowerStr)
	if err != nil {
		log.Printf("Ошибка при чтении COMPUTING_POWER: %v. Используем значение по умолчанию: 2", err)
		computingPower = 2
	}

	orchestratorPort := getEnv("ORCHESTRATOR_PORT", "8080")
	orchestratorHost := getEnv("ORCHESTRATOR_HOST", "localhost")
	orchestratorURL = getEnv("ORCHESTRATOR_URL", fmt.Sprintf("http://%s:%s", orchestratorHost, orchestratorPort))

	var wg sync.WaitGroup
	log.Printf("Запускаем агента с %d вычислителями\n", computingPower)

	for i := 0; i < computingPower; i++ {
		wg.Add(1)
		go worker(i, &wg)
	}

	wg.Wait()
}

func worker(id int, wg *sync.WaitGroup) {

	defer wg.Done()
	client := &http.Client{}
	taskURL := fmt.Sprintf("%s/internal/task", orchestratorURL)

	for {
		resp, err := http.Get(taskURL)
		if err != nil {
			log.Printf("Вычислитель %d: Ошибка получения задачи: %v", id, err)
			time.Sleep(time.Second)
			continue
		}

		if resp.StatusCode == http.StatusNotFound {
			time.Sleep(time.Second)
			continue
		}

		var taskResp struct {
			Task models.Task `json:"task"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&taskResp); err != nil {
			log.Printf("Вычислитель %d: Ошибка декодирования задачи: %v", id, err)
			resp.Body.Close()
			continue
		}
		resp.Body.Close()

		// имитируем задержку выполнения операции типа длительная операция
		time.Sleep(time.Duration(taskResp.Task.OperationTime) * time.Millisecond)

		var result float64
		switch taskResp.Task.Operation {
		case "+":
			result = taskResp.Task.Arg1 + taskResp.Task.Arg2
			log.Printf("Вычислитель %d: %f + %f = %f", id, taskResp.Task.Arg1, taskResp.Task.Arg2, result)
		case "-":
			result = taskResp.Task.Arg1 - taskResp.Task.Arg2
			log.Printf("Вычислитель %d: %f - %f = %f", id, taskResp.Task.Arg1, taskResp.Task.Arg2, result)
		case "*":
			result = taskResp.Task.Arg1 * taskResp.Task.Arg2
			log.Printf("Вычислитель %d: %f * %f = %f", id, taskResp.Task.Arg1, taskResp.Task.Arg2, result)
		case "/":
			if taskResp.Task.Arg2 == 0 {
				log.Printf("Вычислитель %d: Деление на ноль!", id)
				continue
			}
			result = taskResp.Task.Arg1 / taskResp.Task.Arg2
			log.Printf("Вычислитель %d: %f / %f = %f", id, taskResp.Task.Arg1, taskResp.Task.Arg2, result)
		}

		taskResult := models.TaskResult{
			ID:     taskResp.Task.ID,
			Result: result,
		}

		resultJSON, err := json.Marshal(taskResult)
		if err != nil {
			log.Printf("Вычислитель %d: Ошибка маршалинга результата: %v", id, err)
			continue
		}

		_, err = client.Post(
			taskURL,
			"application/json",
			bytes.NewBuffer(resultJSON),
		)
		if err != nil {
			log.Printf("Вычислитель %d: Ошибка отправки результата: %v", id, err)
			continue
		}
	}
}
