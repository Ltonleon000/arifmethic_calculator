# Примеры API запросов

В этом файле представлены примеры использования API распределенного калькулятора с помощью curl.

## Регистрация пользователя

```bash
curl --location 'http://localhost:8081/api/v1/register' \
--header 'Content-Type: application/json' \
--data '{
  "login": "testuser",
  "password": "testpassword"
}'
```

**Успешный ответ**: HTTP 200 OK

**Ошибки**:
- HTTP 400 Bad Request - если логин или пароль не предоставлены
- HTTP 409 Conflict - если пользователь с таким логином уже существует

## Вход пользователя и получение JWT токена

```bash
curl --location 'http://localhost:8081/api/v1/login' \
--header 'Content-Type: application/json' \
--data '{
  "login": "testuser",
  "password": "testpassword"
}'
```

**Успешный ответ**: HTTP 200 OK
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**Ошибки**:
- HTTP 400 Bad Request - если логин или пароль не предоставлены
- HTTP 401 Unauthorized - если предоставлены неверные учетные данные

## Отправка выражения на вычисление (требует JWT)

```bash
curl --location 'http://localhost:8081/api/v1/calculate' \
--header 'Content-Type: application/json' \
--header 'Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...' \
--data '{
  "expression": "2+2*2"
}'
```

**Успешный ответ**: HTTP 200 OK
```json
{
  "id": "a8f5e6c3-1d2b-4a3c-9e8f-7d6c5b4a3e2d",
  "expression": "2+2*2",
  "status": "pending"
}
```

**Ошибки**:
- HTTP 400 Bad Request - если выражение некорректно
- HTTP 401 Unauthorized - если JWT токен не предоставлен или недействителен

## Получение результата вычисления (требует JWT)

```bash
curl --location 'http://localhost:8081/api/v1/expressions/a8f5e6c3-1d2b-4a3c-9e8f-7d6c5b4a3e2d' \
--header 'Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...'
```

**Успешный ответ**: HTTP 200 OK
```json
{
  "id": "a8f5e6c3-1d2b-4a3c-9e8f-7d6c5b4a3e2d",
  "expression": "2+2*2",
  "status": "completed",
  "result": 6
}
```

**Ошибки**:
- HTTP 401 Unauthorized - если JWT токен не предоставлен или недействителен
- HTTP 404 Not Found - если выражение с указанным ID не найдено

## Получение списка всех выражений пользователя (требует JWT)

```bash
curl --location 'http://localhost:8081/api/v1/expressions' \
--header 'Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...'
```

**Успешный ответ**: HTTP 200 OK
```json
[
  {
    "id": "a8f5e6c3-1d2b-4a3c-9e8f-7d6c5b4a3e2d",
    "expression": "2+2*2",
    "status": "completed",
    "result": 6
  },
  {
    "id": "b9e8d7c6-5f4e-3d2c-1b0a-9e8d7f6a5b4c",
    "expression": "10-5/2",
    "status": "completed",
    "result": 7.5
  }
]
```

**Ошибки**:
- HTTP 401 Unauthorized - если JWT токен не предоставлен или недействителен

## Примеры последовательного использования

### Полный цикл работы с калькулятором

1. Регистрация пользователя
2. Вход и получение токена
3. Отправка выражения на вычисление
4. Проверка статуса вычисления
5. Получение результата

```bash
# 1. Регистрация
curl --location 'http://localhost:8081/api/v1/register' \
--header 'Content-Type: application/json' \
--data '{
  "login": "newuser",
  "password": "password123"
}'

# 2. Вход
TOKEN=$(curl --location 'http://localhost:8081/api/v1/login' \
--header 'Content-Type: application/json' \
--data '{
  "login": "newuser",
  "password": "password123"
}' | jq -r '.token')

# 3. Отправка выражения
EXPR_ID=$(curl --location 'http://localhost:8081/api/v1/calculate' \
--header 'Content-Type: application/json' \
--header "Authorization: Bearer $TOKEN" \
--data '{
  "expression": "10+20*3"
}' | jq -r '.id')

# 4. Ждем и проверяем результат
sleep 5
curl --location "http://localhost:8081/api/v1/expressions/$EXPR_ID" \
--header "Authorization: Bearer $TOKEN"
```
