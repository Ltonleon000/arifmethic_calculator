#!/bin/bash

echo "===== Остановка распределенного калькулятора ====="

# Остановка оркестратора
if [ -f orchestrator.pid ]; then
    PID=$(cat orchestrator.pid)
    echo "Остановка оркестратора (PID: $PID)..."
    kill -9 $PID 2>/dev/null
    if [ $? -eq 0 ]; then
        echo "Оркестратор остановлен"
    else
        echo "Не удалось остановить оркестратор"
    fi
    rm orchestrator.pid
else
    echo "Оркестратор не был запущен"
fi

# Остановка агента
if [ -f agent.pid ]; then
    PID=$(cat agent.pid)
    echo "Остановка агента (PID: $PID)..."
    kill -9 $PID 2>/dev/null
    if [ $? -eq 0 ]; then
        echo "Агент остановлен"
    else
        echo "Не удалось остановить агент"
    fi
    rm agent.pid
else
    echo "Агент не был запущен"
fi

# Остановка фронтенда
if [ -f frontend.pid ]; then
    PID=$(cat frontend.pid)
    echo "Остановка фронтенда (PID: $PID)..."
    kill -9 $PID 2>/dev/null
    if [ $? -eq 0 ]; then
        echo "Фронтенд остановлен"
    else
        echo "Не удалось остановить фронтенд"
    fi
    rm frontend.pid
else
    echo "Фронтенд не был запущен"
fi

echo "===== Все компоненты остановлены ====="
