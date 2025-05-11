@echo off
echo Запуск тестов для калькулятора с gRPC...
cd %~dp0

echo.
echo ====================================
echo Модульные тесты: internal/calculator
echo ====================================
go test ./internal/calculator -v

echo.
echo ================================
echo Модульные тесты: internal/models
echo ================================
go test ./internal/models -v

echo.
echo ==============================
echo Модульные тесты: internal/api
echo ==============================
go test ./internal/api -v

echo.
echo =========================
echo Модульные тесты: cmd/agent
echo =========================
go test ./cmd/agent -v

echo.
echo ================================
echo Модульные тесты: cmd/orchestrator
echo ================================
go test ./cmd/orchestrator -v

echo.
echo ===============================
echo Интеграционные тесты: internal/tests
echo ===============================
go test ./internal/tests -v

echo.
echo ======================================
echo Полное покрытие всех тестов проекта
echo ======================================
go test ./... -v

echo.
echo Тестирование завершено.
pause
