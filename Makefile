.PHONY: help dev backend desktop mobile docker-up docker-down proto

help: ## Показать справку
	@echo "Backend - Команды для разработки"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

dev: ## Запустить все сервисы для разработки
	@echo "🚀 Запуск всех сервисов..."
	@make docker-up
	@echo "✅ Сервисы запущены"

docker-up: ## Запустить Docker контейнеры (PostgreSQL, Redis)
	docker-compose up -d
	@echo "✅ Docker контейнеры запущены"

docker-down: ## Остановить Docker контейнеры
	docker-compose down
	@echo "✅ Docker контейнеры остановлены"

proto: ## Сгенерировать Go код из proto файлов
	@echo "🔨 Генерация gRPC кода..."
	@protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		pkg/proto/auth/auth.proto \
		pkg/proto/file/file.proto \
		pkg/proto/device/device.proto \
		pkg/proto/transfer/transfer.proto
	@echo "✅ gRPC код сгенерирован"

backend: ## Запустить backend сервер
	@DB_PASSWORD=Backend_password go run cmd/server/main.go

backend-build: ## Собрать backend
	@cd backend && go build -o bin/server cmd/server/main.go

backend-test: ## Запустить тесты backend
	@cd backend && go test ./...

desktop-dev: ## Запустить desktop приложение в режиме разработки
	@cd desktop && npm run dev

desktop-build: ## Собрать desktop приложение
	@cd desktop && npm run build

mobile-dev: ## Запустить mobile приложение в режиме разработки
	@cd mobile && npm run dev

mobile-build: ## Собрать mobile приложение
	@cd mobile && npm run build

install: ## Установить все зависимости
	@echo "📦 Установка зависимостей..."
	@cd backend && go mod download
	@cd desktop && npm install
	@cd mobile && npm install
	@echo "✅ Зависимости установлены"

clean: ## Очистить временные файлы
	@echo "🧹 Очистка..."
	@rm -rf backend/bin backend/dist
	@rm -rf desktop/dist desktop/build
	@rm -rf mobile/dist mobile/build
	@echo "✅ Очистка завершена"
