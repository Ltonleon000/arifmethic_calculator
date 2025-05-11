@echo off
echo Запуск оркестратора калькулятора...
cd %~dp0
go run ./cmd/orchestrator/main.go
