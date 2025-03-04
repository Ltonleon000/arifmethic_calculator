@echo off
echo ===== Запуск тестов распределенного калькулятора =====

echo Запуск тестов для пакета calculator (парсер)...
cd internal\calculator
go test -v
if %errorlevel% neq 0 (
    echo Тесты пакета calculator не прошли
    exit /b 1
)
cd ..\..

echo Запуск тестов для пакета models (структуры данных)...
cd internal\models
go test -v
if %errorlevel% neq 0 (
    echo Тесты пакета models не прошли
    exit /b 1
)
cd ..\..

echo Запуск тестов для пакета api (обработчики API)...
cd internal\api
go test -v
if %errorlevel% neq 0 (
    echo Тесты пакета api не прошли
    exit /b 1
)
cd ..\..

echo Запуск тестов для cmd пакетов...
go test ./cmd/...
