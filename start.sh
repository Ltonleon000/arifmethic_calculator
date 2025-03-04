#!/bin/bash

echo "===== Запуск распределенного калькулятора ====="

# Компиляция компонентов
echo "Компиляция оркестратора..."
go build -o orchestrator cmd/orchestrator/main.go
if [ $? -ne 0 ]; then
    echo "Ошибка компиляции оркестратора"
    exit 1
fi

echo "Компиляция агента..."
go build -o agent cmd/agent/main.go
if [ $? -ne 0 ]; then
    echo "Ошибка компиляции агента"
    exit 1
fi

echo "Все компоненты скомпилированы успешно!"

# Запуск компонентов
echo "Запуск оркестратора..."
./orchestrator > orchestrator.log 2>&1 &
ORCHESTRATOR_PID=$!
echo $ORCHESTRATOR_PID > orchestrator.pid
sleep 2

echo "Запуск агента..."
./agent > agent.log 2>&1 &
AGENT_PID=$!
echo $AGENT_PID > agent.pid
sleep 1

echo "Запуск фронтенда..."
go run proxy_server.go > frontend.log 2>&1 &
FRONTEND_PID=$!
echo $FRONTEND_PID > frontend.pid
sleep 2

echo "===== Все компоненты запущены успешно! ====="
echo "Калькулятор доступен по адресу: http://localhost:8081"
echo "Инструкции:"
echo "1. Откройте в браузере http://localhost:8081"
echo "2. Для остановки всех компонентов запустите ./stop.sh"
