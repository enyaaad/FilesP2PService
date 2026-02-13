# Swagger API Documentation

Полная документация API для backend доступна через Swagger UI.

## Доступ к Swagger UI

После запуска сервера, Swagger UI доступен по адресу:

```
http://localhost:8080/swagger/index.html
```

## Генерация документации

Для обновления Swagger документации после изменений в handlers:

```bash
cd backend
swag init -g cmd/server/main.go -o docs --parseDependency --parseInternal
```

Или используя go run:

```bash
cd backend
go run github.com/swaggo/swag/cmd/swag@latest init -g cmd/server/main.go -o docs --parseDependency --parseInternal
```

## Структура документации

### Endpoints по категориям:

#### Auth (Аутентификация)
- `POST /api/v1/auth/register` - Регистрация пользователя
- `POST /api/v1/auth/login` - Вход пользователя
- `POST /api/v1/auth/refresh` - Обновление токенов

#### Devices (Устройства)
- `POST /api/v1/devices` - Регистрация устройства
- `GET /api/v1/devices` - Список устройств
- `GET /api/v1/devices/{id}` - Получение устройства
- `PUT /api/v1/devices/{id}` - Обновление устройства
- `DELETE /api/v1/devices/{id}` - Удаление устройства
- `POST /api/v1/devices/{id}/last-seen` - Обновление времени активности

#### Files (Файлы)
- `POST /api/v1/files` - Загрузка файла
- `GET /api/v1/files` - Список файлов (с пагинацией)
- `GET /api/v1/files/{id}` - Метаданные файла
- `GET /api/v1/files/{id}/download` - Скачивание файла (Range requests)
- `DELETE /api/v1/files/{id}` - Удаление файла

#### WebRTC
- `GET /api/v1/webrtc/turn-credentials` - Получение TURN credentials

## Аутентификация

Большинство endpoints требуют JWT токен в заголовке:

```
Authorization: Bearer {access_token}
```

Для получения токена:
1. Зарегистрируйтесь через `POST /api/v1/auth/register`
2. Или войдите через `POST /api/v1/auth/login`
3. Используйте `access_token` из ответа

В Swagger UI:
1. Нажмите кнопку "Authorize" вверху страницы
2. Введите: `Bearer {ваш_access_token}`
3. Нажмите "Authorize"

## Примеры использования

### Регистрация пользователя

```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "securePassword123"
  }'
```

### Вход

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "securePassword123"
  }'
```

### Загрузка файла

```bash
curl -X POST http://localhost:8080/api/v1/files \
  -H "Authorization: Bearer {access_token}" \
  -F "file=@/path/to/file.pdf"
```

### Список файлов

```bash
curl -X GET "http://localhost:8080/api/v1/files?limit=10&offset=0" \
  -H "Authorization: Bearer {access_token}"
```

## Форматы ответов

Все ответы в формате JSON, кроме:
- `GET /api/v1/files/{id}/download` - возвращает бинарные данные файла

## Коды ошибок

- `200` - Успешно
- `201` - Создано
- `400` - Неверный запрос
- `401` - Не авторизован
- `403` - Нет доступа
- `404` - Не найдено
- `409` - Конфликт (например, пользователь уже существует)
- `500` - Внутренняя ошибка сервера

## Swagger файлы

Документация генерируется в следующие файлы:
- `docs/swagger.yaml` - YAML формат
- `docs/swagger.json` - JSON формат
- `docs/docs.go` - Go код для интеграции

Эти файлы автоматически обновляются при генерации документации.
