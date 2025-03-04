@echo off
echo ===== Остановка распределенного калькулятора =====

echo Остановка оркестратора...
taskkill /f /im orchestrator.exe 2>nul
if %ERRORLEVEL% equ 0 (
    echo Оркестратор остановлен
) else (
    echo Оркестратор не был запущен
)

echo Остановка агента...
taskkill /f /im agent.exe 2>nul
if %ERRORLEVEL% equ 0 (
    echo Агент остановлен
) else (
    echo Агент не был запущен
)

echo Остановка фронтенда (go run)...
taskkill /f /fi "WINDOWTITLE eq Фронтенд*" 2>nul
