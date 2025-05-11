@echo off
echo Запуск агента калькулятора...
cd %~dp0
go run ./cmd/agent/main.go
