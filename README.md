# Uptime Monitor

Сервис для мониторинга доступности сайтов с уведомлениями в Telegram.

## Возможности

- ✅ Мониторинг доступности сайтов в реальном времени
- ✅ Веб-интерфейс для управления сайтами
- ✅ Автоматические проверки каждые 5 минут
- ✅ Ручное обновление статусов сайтов
- ✅ Фильтрация по статусу (все сайты / только DOWN)
- ✅ Telegram уведомления при изменении статуса
- ✅ Индивидуальные настройки уведомлений для каждого пользователя
- ✅ Система авторизации и регистрации

## Быстрый старт

### 1. Настройка окружения

Создайте файл `deployments/.env` со следующими настройками:

```env
# Database configuration
DB_HOST=localhost
DB_PORT=5432
DB_NAME=uptimedb
DB_USER=monitoruser
DB_PASSWORD=your_password_here

# Server configuration
SERVER_PORT=8080
JWT_SECRET=your-super-secret-jwt-key-change-this-in-production

# Telegram Bot configuration
TELEGRAM_TOKEN=your_telegram_bot_token_here
```

### 2. Запуск с Docker

```bash
# Запуск базы данных
docker-compose -f deployments/docker-compose.yml up -d

# Запуск сервера
go run cmd/server/main.go
```

### 3. Настройка Telegram бота

1. Создайте бота через [@BotFather](https://t.me/botfather)
2. Получите токен бота
3. Добавьте токен в файл `.env`
4. Войдите в веб-интерфейс и настройте уведомления

## Структура проекта

```
uptime-monitor/
├── cmd/server/           # Основной сервер
├── internal/             # Внутренние пакеты
│   ├── checker/          # Проверка сайтов
│   ├── config/           # Конфигурация
│   ├── handlers/         # HTTP обработчики
│   ├── middleware/       # Middleware
│   ├── models/           # Модели данных
│   ├── notifier/         # Уведомления
│   ├── storage/          # Работа с БД
│   ├── telegram/         # Telegram бот
│   └── utils/            # Утилиты
├── migrations/           # SQL миграции
├── web/                  # Веб-интерфейс
│   ├── static/           # CSS, JS файлы
│   └── templates/        # HTML шаблоны
└── deployments/          # Docker конфигурация
```

## API Endpoints

### Публичные
- `POST /api/register` - Регистрация
- `POST /api/login` - Авторизация

### Защищенные (требуют JWT)
- `GET /api/sites` - Получить список сайтов
- `POST /api/sites` - Добавить сайт
- `POST /api/sites/bulk` - Массовое добавление сайтов
- `DELETE /api/sites/{id}` - Удалить сайт
- `POST /api/sites/bulk-delete` - Массовое удаление
- `POST /api/sites/refresh` - Ручное обновление статусов
- `GET /api/verify-token` - Проверка токена
- `POST /api/telegram/link-code` - Генерация кода для Telegram

## Telegram команды

- `/start` - Начать работу с ботом
- `/link <код>` - Связать аккаунт с ботом
- `/unlink` - Отвязать аккаунт
- `/status` - Проверить статус связывания
- `/help` - Справка

## Технологии

- **Backend:** Go, PostgreSQL, JWT
- **Frontend:** HTML, CSS, JavaScript
- **Telegram:** Bot API
- **Deployment:** Docker, Docker Compose

## Разработка

```bash
# Установка зависимостей
go mod download

# Запуск в режиме разработки
go run cmd/server/main.go

# Сборка
go build -o uptime-monitor cmd/server/main.go
```

## Лицензия

MIT License