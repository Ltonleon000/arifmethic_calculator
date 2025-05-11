@echo off
echo Запуск калькулятора с gRPC (оркестратор + агент)...
cd %~dp0

:: Проверяем, не занят ли порт 8081
netstat -ano | findstr :8081 > nul
if %errorlevel% equ 0 (
    echo Порт 8081 уже занят. Освобождаем...
    for /f "tokens=5" %%a in ('netstat -ano ^| findstr :8081') do (
        taskkill /PID %%a /F > nul 2>&1
    )
)

:: Проверяем, не занят ли порт 8082
netstat -ano | findstr :8082 > nul
if %errorlevel% equ 0 (
    echo Порт 8082 уже занят. Освобождаем...
    for /f "tokens=5" %%a in ('netstat -ano ^| findstr :8082') do (
        taskkill /PID %%a /F > nul 2>&1
    )
)

:: Запускаем оркестратор в отдельном окне
start "Оркестратор" cmd /c "cd %~dp0 && go run ./cmd/orchestrator/main.go"

:: Ждем 2 секунды для запуска оркестратора
timeout /t 2 /nobreak > nul

:: Запускаем агента в отдельном окне
start "Агент" cmd /c "cd %~dp0 && go run ./cmd/agent/main.go"

:: Ждем еще 2 секунды для запуска всех сервисов
timeout /t 2 /nobreak > nul
:: Не открываем браузер автоматически, чтобы пользователь мог сам выбрать файл

echo.
echo Система запущена!
echo Оркестратор: http://localhost:8081/
echo gRPC сервер: порт 8082
echo.
echo Для остановки нажмите любую клавишу...
pause > nul
