@echo off
echo Запуск исправленных тестов для gRPC калькулятора...
cd %~dp0

echo.
echo ====================================
echo Модульные тесты для пакета cmd/agent
echo ====================================
go test ./cmd/agent -v

echo.
echo =====================================
echo Модульные тесты для пакета calculator
echo =====================================
go test ./internal/calculator -v

echo.
echo =========================
echo Модульные тесты для models
echo =========================
go test ./internal/models -v

echo.
echo ===============================
echo Интеграционные тесты: internal/tests
echo ===============================
go test ./internal/tests -v

echo.
echo Тестирование завершено.
pause
