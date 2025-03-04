@echo off
echo ===== Запуск распределенного калькулятора =====

echo Компиляция оркестратора...
go build -o orchestrator.exe cmd/orchestrator/main.go
if %ERRORLEVEL% neq 0 (
    echo Ошибка компиляции оркестратора
    exit /b %ERRORLEVEL%
)
echo Компиляция агента...
go build -o agent.exe cmd/agent/main.go
if %ERRORLEVEL% neq 0 (
    echo Ошибка компиляции агента
    exit /b %ERRORLEVEL%
)
echo Все компоненты скомпилированы успешно!

echo Запуск оркестратора...
start "Оркестратор" cmd /c "orchestrator.exe"
timeout /t 2 /nobreak > nul

echo Запуск агента...
start "Агент" cmd /c "agent.exe"
timeout /t 1 /nobreak > nul

echo Запуск фронтенда...
start "Фронтенд" cmd /c "go run proxy_server.go"
timeout /t 2 /nobreak > nul

echo Открытие браузера...
start http://localhost:8081