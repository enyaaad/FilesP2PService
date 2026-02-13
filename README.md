# Backend

Go backend для файлообмена между устройствами с поддержкой P2P через WebRTC.

## Быстрый старт

### 1. Запуск инфраструктуры (PostgreSQL и Redis)

```bash
# Из корня проекта
docker-compose up -d
```

### 2. Запуск backend сервера

```bash
# Вариант 1: Через Makefile (из корня проекта)
make backend

# Вариант 2: Напрямую
cd backend
go run cmd/server/main.go
```

### 3. Проверка

- Health check: http://localhost:8080/health
- Swagger UI: http://localhost:8080/swagger/index.html

## Переменные окружения

Все переменные окружения имеют значения по умолчанию, но для production рекомендуется создать `.env` файл:

```bash
cp .env.example .env
# Отредактируйте .env при необходимости
```

## Основные компоненты

- **HTTP API Server** (Gin) - порт 8080 - REST API для фронтенда
- **gRPC Server** - порт 9090 - Внутренние сервисы
- **WebSocket Signaling Server** - порт 8081 - WebRTC signaling
- **TURN Server** - порт 3478 - Ретрансляция WebRTC трафика

## API Endpoints

### Аутентификация (публичные)
- `POST /api/v1/auth/register` - Регистрация пользователя
- `POST /api/v1/auth/login` - Вход
- `POST /api/v1/auth/refresh` - Обновление токенов

### Устройства (требуют аутентификации)
- `POST /api/v1/devices` - Регистрация устройства
- `GET /api/v1/devices` - Список устройств
- `GET /api/v1/devices/{id}` - Получение устройства
- `PUT /api/v1/devices/{id}` - Обновление устройства
- `DELETE /api/v1/devices/{id}` - Удаление устройства
- `POST /api/v1/devices/{id}/last-seen` - Обновление активности

### Файлы (требуют аутентификации)
- `POST /api/v1/files` - Загрузка файла
- `GET /api/v1/files` - Список файлов (пагинация)
- `GET /api/v1/files/{id}` - Метаданные файла
- `GET /api/v1/files/{id}/download` - Скачивание (Range requests)
- `DELETE /api/v1/files/{id}` - Удаление файла

### WebRTC (требуют аутентификации)
- `GET /api/v1/webrtc/turn-credentials` - TURN credentials

## Swagger документация

Полная документация API доступна через Swagger UI:

```
http://localhost:8080/swagger/index.html
```

## Производительность

- Zero-Copy передача файлов
- Параллельная обработка чанков
- Кэширование в Redis для быстрого доступа к метаданным
- P2P передача через WebRTC в локальной сети

## Структура проекта

```
backend/
├── cmd/
│   └── server/        # Точка входа
├── internal/
│   ├── api/           # HTTP handlers и сервер
│   ├── grpc/           # gRPC сервер и сервисы
│   ├── websocket/      # WebSocket signaling сервер
│   ├── webrtc/         # TURN сервер
│   ├── storage/        # Хранилище файлов
│   ├── models/         # Модели данных
│   ├── repository/     # Репозитории для БД
│   └── database/       # Подключение к БД и миграции
├── pkg/
│   ├── config/         # Конфигурация
│   ├── jwt/            # JWT утилиты
│   ├── logger/         # Логирование
│   └── proto/          # Protobuf определения
└── docs/               # Swagger документация
```

## Полезные команды

```bash
# Показать все команды
make help

# Запустить Docker контейнеры
make docker-up

# Остановить Docker контейнеры
make docker-down

# Собрать backend
make backend-build

# Запустить тесты
make backend-test

# Сгенерировать Swagger документацию
cd backend
swag init -g cmd/server/main.go -o docs --parseDependency --parseInternal
```

## Документация

- [QUICK_START.md](./QUICK_START.md) - Подробная инструкция по запуску
- [docs/SWAGGER.md](./docs/SWAGGER.md) - Документация по Swagger
