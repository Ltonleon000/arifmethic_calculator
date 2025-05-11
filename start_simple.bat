@echo off
echo Запуск калькулятора с gRPC и авторизацией...
cd %~dp0

:: Проверяем, не занят ли порт 8082
netstat -ano | findstr :8082 > nul
if %errorlevel% equ 0 (
    echo Порт 8082 уже занят. Освобождаем...
    for /f "tokens=5" %%a in ('netstat -ano ^| findstr :8082') do (
        taskkill /PID %%a /F > nul 2>&1
    )
)

:: Проверяем, не занят ли порт 8081
netstat -ano | findstr :8081 > nul
if %errorlevel% equ 0 (
    echo Порт 8081 уже занят. Освобождаем...
    for /f "tokens=5" %%a in ('netstat -ano ^| findstr :8081') do (
        taskkill /PID %%a /F > nul 2>&1
    )
)

:: Запускаем gRPC агент в отдельном окне
start "gRPC Агент" cmd /c "cd %~dp0 && go run ./cmd/simple_agent/main.go"

:: Ждем 2 секунды для запуска агента
timeout /t 2 /nobreak > nul

:: Запускаем оркестратор в отдельном окне
start "Оркестратор" cmd /c "cd %~dp0 && go run ./cmd/simple_orchestrator/main.go"

echo.
echo Система запущена!
echo Оркестратор: http://localhost:8081/
echo gRPC агент: порт 8082
echo.
echo Примеры API запросов:
echo.
echo --- Запросы авторизации ---
echo 1. Регистрация пользователя:
echo curl -X POST -H "Content-Type: application/json" -d "{\"login\":\"user1\", \"password\":\"pass123\"}" http://localhost:8081/api/v1/register
echo.
echo 2. Вход и получение JWT-токена:
echo curl -X POST -H "Content-Type: application/json" -d "{\"login\":\"user1\", \"password\":\"pass123\"}" http://localhost:8081/api/v1/login
echo.
echo --- Запросы с авторизацией ---
echo 3. Отправить выражение для вычисления (с токеном):
echo curl -X POST -H "Content-Type: application/json" -H "Authorization: Bearer [TOKEN]" -d "{\"expression\":\"2+2*3\"}" http://localhost:8081/api/v1/calculate
echo.
echo 4. Получить список выражений пользователя:
echo curl -H "Authorization: Bearer [TOKEN]" http://localhost:8081/api/v1/expressions
echo.
echo 5. Получить конкретное выражение по ID:
echo curl -H "Authorization: Bearer [TOKEN]" http://localhost:8081/api/v1/expressions/[ID]
echo.
echo --- Веб-интерфейс ---
echo Откройте simple.html в браузере для работы с калькулятором
echo.
echo Для остановки нажмите любую клавишу...
pause > nul
