# Uptime Monitor

Сервис для мониторинга доступности сайтов с уведомлениями в Telegram.

## Быстрый старт

1.  **Скопируйте настройки:**
    ```bash
    cp .env.example .env
    ```

2.  **Настройте подключения** в файле `.env`:
    - Пароль от базы данных
    - Токен Telegram бота

3.  **Запустите сервис:**
    ```bash
    docker-compose -f deployments/docker-compose.yml up -d
    go run ./cmd/server/main.go
    ```

## Разработка

Проект написан на Go с использованием PostgreSQL и Docker.