package main

import (
	"calculator/internal"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
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

	orchestratorPort := getEnv("ORCHESTRATOR_PORT", "8082")
	orchestratorHost := getEnv("ORCHESTRATOR_HOST", "localhost")
	grpcAddr := fmt.Sprintf("%s:%s", orchestratorHost, orchestratorPort)

	client, err := internal.NewAgentGRPCClient(grpcAddr)
	if err != nil {
		log.Fatalf("Ошибка подключения к gRPC серверу оркестратора: %v", err)
	}
	defer client.Close()

	var wg sync.WaitGroup
	log.Printf("Запускаем агента с %d вычислителями\n", computingPower)

	for i := 0; i < computingPower; i++ {
		wg.Add(1)
		go worker(i, &wg, client)
	}

	wg.Wait()
}

func worker(id int, wg *sync.WaitGroup, client *internal.AgentGRPCClient) {
	defer wg.Done()

	for {
		taskMsg, ok := client.GetTask()
		if !ok || taskMsg == nil {
			time.Sleep(time.Second)
			continue
		}

		// Преобразуем аргументы из строк в float64
		arg1, err1 := strconv.ParseFloat(taskMsg.Arg1, 64)
		arg2, err2 := strconv.ParseFloat(taskMsg.Arg2, 64)
		if err1 != nil || err2 != nil {
			log.Printf("Вычислитель %d: Ошибка преобразования аргументов: %v %v", id, err1, err2)
			continue
		}

		// Парсим операцию из составного поля Operation, которое имеет формат "ID||операция"
		operationData := taskMsg.Operation
		parts := strings.Split(operationData, "||")
		var operation string
		if len(parts) > 1 {
			// Если есть разделитель "||" - берем вторую часть как операцию
			operation = parts[1]
			log.Printf("Вычислитель %d: Распознана операция %s для задачи %s", id, operation, parts[0])
		} else {
			// Запасной вариант - если формат старый, берем всё как операцию
			operation = operationData
			log.Printf("Вычислитель %d: Используем прямое значение операции: %s", id, operation)
		}

		// имитируем задержку выполнения операции типа длительная операция
		time.Sleep(time.Duration(taskMsg.OperationTime) * time.Millisecond)

		var result float64
		switch operation {
		case "+":
			result = arg1 + arg2
			log.Printf("Вычислитель %d: %f + %f = %f", id, arg1, arg2, result)
		case "-":
			result = arg1 - arg2
			log.Printf("Вычислитель %d: %f - %f = %f", id, arg1, arg2, result)
		case "*":
			result = arg1 * arg2
			log.Printf("Вычислитель %d: %f * %f = %f", id, arg1, arg2, result)
		case "/":
			if arg2 == 0 {
				log.Printf("Вычислитель %d: Деление на ноль!", id)
				continue
			}
			result = arg1 / arg2
			log.Printf("Вычислитель %d: %f / %f = %f", id, arg1, arg2, result)
		default:
			log.Printf("Вычислитель %d: Неизвестная операция: %s", id, operation)
			continue
		}

		if !client.SendResult(taskMsg.Id, result) {
			log.Printf("Вычислитель %d: Ошибка отправки результата задачи %d", id, taskMsg.Id)
		}
	}
}
