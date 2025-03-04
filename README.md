# Распределённый калькулятор

Привет! Это необычный калькулятор, который умеет считать арифметические выражения в распределённом режиме. 
Представьте, что каждая математическая операция настолько сложная, что её нужно выполнять отдельно и очень долго. 
Именно для этого и создан наш калькулятор!

# Быстрый старт
Если нет желания читать документация, то информация по быстрому запуску находится в файле [Быстрый старт](https://github.com/Ltonleon000/arifmethic_calculator/blob/master/QUICK_START.md)

## Содержание
- [Архитектура системы](#архитектура-системы)
- [Установка и настройка](#установка-и-настройка)
- [Компоненты системы](#компоненты-системы)
- [Конфигурация](#конфигурация)
- [Запуск и остановка](#запуск-и-остановка)
- [API](#api)
- [Запуск тестов](#запуск-тестов)
- [Запуск проекта](#запуск-проекта)
- [Как это работает?](#как-это-работает)

## Архитектура системы

Распределенный калькулятор имеет трехкомпонентную архитектуру:

```
┌─────────────┐      ┌───────────────┐      ┌───────────────┐
│             │      │               │      │               │
│  Фронтенд   │◄────►│  Оркестратор  │◄────►│    Агенты     │
│ (Браузер)   │      │   (Go API)    │      │ (Вычислители) │
│             │      │               │      │               │
└─────────────┘      └───────────────┘      └───────────────┘
      │                     │                      │
      │                     │                      │
      ▼                     ▼                      ▼
  Порт 8081             Порт 8080          Подключаются к 
(Прокси-сервер)        (REST API)          оркестратору
```

### Процесс вычисления

```
┌────────────┐     ┌────────────┐     ┌────────────┐     ┌────────────┐
│            │     │            │     │            │     │            │
│ Ввод       │────►│ Парсинг    │────►│ Выполнение │────►│ Отображение│
│ выражения  │     │ выражения  │     │ операций   │     │ результата │
│            │     │            │     │            │     │            │
└────────────┘     └────────────┘     └────────────┘     └────────────┘
                         │                  ▲
                         │                  │
                         ▼                  │
                   ┌────────────┐     ┌────────────┐
                   │            │     │            │
                   │ Создание   │────►│ Выполнение │
                   │ задач      │     │ задач      │
                   │            │     │            │
                   └────────────┘     └────────────┘
```

## Установка и настройка

### Требования
- Go 1.16 или новее
- Веб-браузер
- Windows или Linux

### Установка

1. Клонируйте репозиторий:
   ```
   git clone https://github.com/Ltonleon000/arifmethic_calculator.git
   cd distributed-calculator
   ```

2. Скомпилируйте исполняемые файлы:
   ```
   go build -o orchestrator.exe cmd/orchestrator/main.go
   go build -o agent.exe cmd/agent/main.go
   ```

## Компоненты системы

### 1. Оркестратор
Центральный компонент системы, выполняющий следующие функции:
- Принимает выражения от пользователя
- Разбивает их на элементарные операции
- Создает задачи для агентов
- Координирует процесс вычисления
- Собирает результаты и формирует итоговый ответ

### 2. Агенты
Выполняют элементарные арифметические операции:
- Получают задачи от оркестратора
- Выполняют указанную операцию (сложение, вычитание, умножение, деление)
- Возвращают результат оркестратору
- Могут эмулировать задержку для демонстрации распределенных вычислений

## Конфигурация

Система использует следующие переменные окружения для настройки:

| Параметр | Описание | Значение по умолчанию |
|----------|----------|------------------------|
| ORCHESTRATOR_PORT | Порт, на котором работает оркестратор | 8080 |
| FRONTEND_PORT | Порт, на котором работает фронтенд | 8081 |
| ORCHESTRATOR_HOST | Хост оркестратора | localhost |
| FRONTEND_HOST | Хост фронтенда | localhost |
| COMPUTING_POWER | Количество потоков вычислителей | 2 |
| TIME_ADDITION_MS | Время выполнения сложения (мс) | 1000 |
| TIME_SUBTRACTION_MS | Время выполнения вычитания (мс) | 1000 |
| TIME_MULTIPLICATIONS_MS | Время выполнения умножения (мс) | 2000 |
| TIME_DIVISIONS_MS | Время выполнения деления (мс) | 2000 |
| LOG_LEVEL | Уровень логирования | info |
| API_TIMEOUT | Таймаут ожидания ответа API (сек) | 30 |

## Запуск и остановка

### Запуск на Windows
Для запуска всей системы используйте скрипт:
```
start.bat
```

Скрипт выполнит следующие действия:
1. Скомпилирует все компоненты (оркестратор и агент)
2. Запустит оркестратор на порту ORCHESTRATOR_PORT (по умолчанию 8080)
3. Запустит агент с настроенным количеством вычислителей (COMPUTING_POWER)
4. Запустит прокси-сервер для обслуживания фронтенда на порту FRONTEND_PORT (по умолчанию 8081)
5. Откроет браузер с интерфейсом калькулятора

### Запуск на Linux/Mac
Для запуска всей системы используйте скрипт:
```
chmod +x start.sh  # Только при первом запуске
./start.sh
```

### Остановка на Windows
```
stop.bat
```

### Остановка на Linux/Mac
```
chmod +x stop.sh  # Только при первом запуске
./stop.sh
```

### Настройка параметров

Вы можете настраивать параметры запуска через переменные окружения:

#### Windows
```
set ORCHESTRATOR_PORT=9090
set FRONTEND_PORT=9091
set COMPUTING_POWER=4
start.bat
```

#### Linux/Mac
```
export ORCHESTRATOR_PORT=9090
export FRONTEND_PORT=9091
export COMPUTING_POWER=4
./start.sh
```

## API

### Основные эндпоинты

| Метод | Эндпоинт | Описание |
|-------|----------|----------|
| POST | /api/v1/calculate | Создание нового вычисления |
| GET | /api/v1/expressions/:id | Получение статуса и результата вычисления |
| GET | /api/v1/expressions | Получение списка всех выражений |
| GET | /api/v1/task | Получение задачи агентом |
| POST | /api/v1/task/result | Отправка результата задачи агентом |

### Примеры запросов

#### Создание вычисления
```
POST /api/v1/calculate
Content-Type: application/json

{
  "expression": "2+2*3"
}
```

#### Получение результата
```
GET /api/v1/expressions/8f7d9c3e-1a2b-3c4d-5e6f-7g8h9i0j1k2l
```

## Поиск и устранение неисправностей

### Порты заняты
Если порты 8080 или 8081 уже используются другими приложениями:
1. Измените значения ORCHESTRATOR_PORT и FRONTEND_PORT:
   ```
   set ORCHESTRATOR_PORT=9090
   set FRONTEND_PORT=9091
   start.bat
   ```

### CORS-ошибки
Если в консоли браузера появляются ошибки CORS:
1. Убедитесь, что все компоненты запущены
2. Убедитесь, что заголовки CORS корректно установлены в ответах API

### Ошибки при запросах к API
Если запросы к API возвращают ошибки:
1. Проверьте, что оркестратор запущен и доступен по указанному URL
2. Убедитесь, что формат данных в запросах соответствует ожидаемому
3. Проверьте логи оркестратора на наличие ошибок

## Запуск тестов

Для запуска всех тестов можно использовать скрипты:

### Windows
```
test.bat
```

### Linux/Mac
```
./test.sh
```

### Структура тестов

Тесты организованы по пакетам:

- `internal/calculator` - тесты для парсера и вычислений
- `internal/models` - тесты для моделей данных 
- `internal/api` - тесты для API-обработчиков


## Как это работает?

1. Вы отправляете выражение, например "2 + 2 * 2"
2. Наш умный оркестратор разбивает его на отдельные операции:
   - Сначала посчитает 2 * 2 = 4
   - Потом прибавит к результату 2
3. Агенты-вычислители берут эти операции и старательно их считают
4. В конце вы получаете результат: 6!

## Запуск проекта

### Шаг 1: Запускаем оркестратор (главный сервер)
```bash
go run ./cmd/orchestrator/main.go
```

### Шаг 2: Запускаем агентов-вычислителей
```bash
go run ./cmd/agent/main.go
```
Можно запустить несколько агентов в разных терминалах - чем больше агентов, тем быстрее считаем!

## Настройка

### Для оркестратора
Можно настроить, сколько времени занимает каждая операция (в миллисекундах):
- `TIME_ADDITION_MS` - время сложения
- `TIME_SUBTRACTION_MS` - время вычитания
- `TIME_MULTIPLICATIONS_MS` - время умножения
- `TIME_DIVISIONS_MS` - время деления

### Для агента
- `COMPUTING_POWER` - сколько вычислителей запустить внутри одного агента (по умолчанию 2)

## Примеры использования

### Отправляем выражение на вычисление
```bash
curl --location 'localhost/api/v1/calculate' \
--header 'Content-Type: application/json' \
--data '{
  "expression": "2+2*2"
}'
```
В ответ получим ID нашего выражения:
```json
{
    "id": "12345"
}
```

### Проверяем статус вычисления
```bash
curl --location 'localhost/api/v1/expressions/12345'
```
Ответ будет таким:
```json
{
    "expression": {
        "id": "12345",
        "status": "completed",
        "result": 6
    }
}
```

## Как это устроено внутри?

Проект состоит из двух главных частей:

1. **Оркестратор** - это наш главный сервер, который:
   - Принимает выражения от пользователей
   - Разбивает их на простые операции
   - Раздаёт задачи агентам
   - Собирает результаты

2. **Агент** - это трудяга-вычислитель, который:
   - Постоянно просит у оркестратора новые задачи
   - Выполняет математические операции
   - Отправляет результаты обратно

## Структура проекта

- `cmd/` - здесь живут наши программы
  - `orchestrator/` - код оркестратора
  - `agent/` - код агента
- `internal/` - внутренние библиотеки
  - `models/` - описания данных
  - `calculator/` - логика калькулятора
  - `api/` - API-хендлеры
