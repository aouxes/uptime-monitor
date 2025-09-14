# Makefile для Uptime Monitor

.PHONY: help build run test clean docker-build docker-run deploy

# Переменные
APP_NAME=uptime-monitor
DOCKER_IMAGE=uptime-monitor:latest
BUILD_DIR=build

# Цвета для вывода
GREEN=\033[0;32m
YELLOW=\033[1;33m
RED=\033[0;31m
NC=\033[0m # No Color

help: ## Показать справку
	@echo "$(GREEN)Uptime Monitor - Команды сборки и развертывания$(NC)"
	@echo ""
	@echo "$(YELLOW)Доступные команды:$(NC)"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(GREEN)%-15s$(NC) %s\n", $$1, $$2}'

build: ## Собрать приложение для всех платформ
	@echo "$(GREEN)🚀 Сборка приложения...$(NC)"
	@./scripts/build.sh

run: ## Запустить приложение локально
	@echo "$(GREEN)🏃 Запуск приложения...$(NC)"
	@go run cmd/server/main.go

test: ## Запустить тесты
	@echo "$(GREEN)🧪 Запуск тестов...$(NC)"
	@go test ./...

clean: ## Очистить директорию сборки
	@echo "$(GREEN)🧹 Очистка директории сборки...$(NC)"
	@rm -rf $(BUILD_DIR)
	@rm -f *.tar.gz *.zip

docker-build: ## Собрать Docker образ
	@echo "$(GREEN)🐳 Сборка Docker образа...$(NC)"
	@docker build -t $(DOCKER_IMAGE) .

docker-run: ## Запустить приложение в Docker
	@echo "$(GREEN)🐳 Запуск в Docker...$(NC)"
	@docker-compose -f deployments/docker-compose.yml up -d

docker-stop: ## Остановить Docker контейнеры
	@echo "$(GREEN)🛑 Остановка Docker контейнеров...$(NC)"
	@docker-compose -f deployments/docker-compose.yml down

docker-logs: ## Показать логи Docker контейнеров
	@echo "$(GREEN)📋 Логи Docker контейнеров...$(NC)"
	@docker-compose -f deployments/docker-compose.yml logs -f

deploy: ## Развернуть на сервере (требует SERVER_IP)
	@if [ -z "$(SERVER_IP)" ]; then \
		echo "$(RED)❌ Укажите SERVER_IP: make deploy SERVER_IP=192.168.1.100$(NC)"; \
		exit 1; \
	fi
	@echo "$(GREEN)🚀 Развертывание на сервер $(SERVER_IP)...$(NC)"
	@./scripts/deploy.sh $(SERVER_IP)

dev: ## Запустить в режиме разработки
	@echo "$(GREEN)🔧 Запуск в режиме разработки...$(NC)"
	@docker-compose -f deployments/docker-compose.yml up -d postgres
	@echo "$(YELLOW)⏳ Ожидание запуска базы данных...$(NC)"
	@sleep 10
	@go run cmd/server/main.go

prod: ## Запустить в продакшен режиме
	@echo "$(GREEN)🏭 Запуск в продакшен режиме...$(NC)"
	@docker-compose -f docker-compose.prod.yml up -d

prod-stop: ## Остановить продакшен контейнеры
	@echo "$(GREEN)🛑 Остановка продакшен контейнеров...$(NC)"
	@docker-compose -f docker-compose.prod.yml down

prod-logs: ## Показать логи продакшен контейнеров
	@echo "$(GREEN)📋 Логи продакшен контейнеров...$(NC)"
	@docker-compose -f docker-compose.prod.yml logs -f

install-deps: ## Установить зависимости
	@echo "$(GREEN)📦 Установка зависимостей...$(NC)"
	@go mod download
	@go mod tidy

fmt: ## Форматировать код
	@echo "$(GREEN)🎨 Форматирование кода...$(NC)"
	@go fmt ./...

lint: ## Проверить код линтером
	@echo "$(GREEN)🔍 Проверка кода линтером...$(NC)"
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "$(YELLOW)⚠️ golangci-lint не установлен. Установите его для проверки кода.$(NC)"; \
	fi

migrate: ## Применить миграции базы данных
	@echo "$(GREEN)🗄️ Применение миграций...$(NC)"
	@if [ -f "migrations/apply.go" ]; then \
		go run migrations/apply.go; \
	else \
		echo "$(YELLOW)⚠️ Файл migrations/apply.go не найден. Создайте его для применения миграций.$(NC)"; \
	fi

backup: ## Создать backup базы данных
	@echo "$(GREEN)💾 Создание backup базы данных...$(NC)"
	@docker exec uptime-monitor-db pg_dump -U monitoruser uptimedb > backup_$(shell date +%Y%m%d_%H%M%S).sql
	@echo "$(GREEN)✅ Backup создан: backup_$(shell date +%Y%m%d_%H%M%S).sql$(NC)"

status: ## Показать статус сервисов
	@echo "$(GREEN)📊 Статус сервисов...$(NC)"
	@docker-compose -f deployments/docker-compose.yml ps 2>/dev/null || echo "$(YELLOW)Контейнеры не запущены$(NC)"

health: ## Проверить здоровье приложения
	@echo "$(GREEN)🏥 Проверка здоровья приложения...$(NC)"
	@curl -f http://localhost:8080/health 2>/dev/null && echo "$(GREEN)✅ Приложение работает$(NC)" || echo "$(RED)❌ Приложение недоступно$(NC)"

# Команды для быстрого старта
quick-start: clean install-deps build docker-run ## Быстрый старт (очистка, установка, сборка, запуск)
	@echo "$(GREEN)🎉 Быстрый старт завершен!$(NC)"
	@echo "$(YELLOW)Приложение доступно по адресу: http://localhost:8080$(NC)"

quick-deploy: build deploy ## Быстрое развертывание (сборка + развертывание)
	@echo "$(GREEN)🎉 Быстрое развертывание завершено!$(NC)"
