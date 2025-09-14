#!/bin/bash

# Скрипт для развертывания на сервере

set -e

# Цвета для вывода
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Функция для вывода сообщений
log() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Проверка аргументов
if [ $# -eq 0 ]; then
    echo "Использование: $0 <server_ip> [ssh_key]"
    echo "Пример: $0 192.168.1.100"
    echo "Пример: $0 192.168.1.100 ~/.ssh/id_rsa"
    exit 1
fi

SERVER_IP=$1
SSH_KEY=${2:-~/.ssh/id_rsa}
APP_NAME="uptime-monitor"
REMOTE_USER="root"

log "🚀 Развертывание Uptime Monitor на сервер $SERVER_IP"

# Проверка SSH ключа
if [ ! -f "$SSH_KEY" ]; then
    error "SSH ключ не найден: $SSH_KEY"
    exit 1
fi

# Проверка подключения к серверу
log "🔍 Проверка подключения к серверу..."
if ! ssh -i "$SSH_KEY" -o ConnectTimeout=10 -o BatchMode=yes "$REMOTE_USER@$SERVER_IP" "echo 'Подключение успешно'" &> /dev/null; then
    error "Не удается подключиться к серверу $SERVER_IP"
    exit 1
fi

log "✅ Подключение к серверу установлено"

# Создание директории на сервере
log "📁 Создание директории приложения на сервере..."
ssh -i "$SSH_KEY" "$REMOTE_USER@$SERVER_IP" "mkdir -p /opt/$APP_NAME"

# Копирование файлов
log "📤 Копирование файлов на сервер..."
scp -i "$SSH_KEY" -r build/* "$REMOTE_USER@$SERVER_IP:/opt/$APP_NAME/"

# Копирование конфигурационных файлов
log "📋 Копирование конфигурационных файлов..."
scp -i "$SSH_KEY" docker-compose.prod.yml "$REMOTE_USER@$SERVER_IP:/opt/$APP_NAME/docker-compose.yml"
scp -i "$SSH_KEY" nginx.conf "$REMOTE_USER@$SERVER_IP:/opt/$APP_NAME/"
scp -i "$SSH_KEY" env.example "$REMOTE_USER@$SERVER_IP:/opt/$APP_NAME/.env.example"

# Установка Docker на сервере
log "🐳 Проверка и установка Docker..."
ssh -i "$SSH_KEY" "$REMOTE_USER@$SERVER_IP" "
    if ! command -v docker &> /dev/null; then
        echo 'Установка Docker...'
        curl -fsSL https://get.docker.com -o get-docker.sh
        sh get-docker.sh
        systemctl enable docker
        systemctl start docker
    else
        echo 'Docker уже установлен'
    fi
"

# Установка Docker Compose
log "🔧 Проверка и установка Docker Compose..."
ssh -i "$SSH_KEY" "$REMOTE_USER@$SERVER_IP" "
    if ! command -v docker-compose &> /dev/null; then
        echo 'Установка Docker Compose...'
        curl -L \"https://github.com/docker/compose/releases/latest/download/docker-compose-\$(uname -s)-\$(uname -m)\" -o /usr/local/bin/docker-compose
        chmod +x /usr/local/bin/docker-compose
    else
        echo 'Docker Compose уже установлен'
    fi
"

# Создание .env файла
log "⚙️ Настройка переменных окружения..."
ssh -i "$SSH_KEY" "$REMOTE_USER@$SERVER_IP" "
    cd /opt/$APP_NAME
    if [ ! -f .env ]; then
        cp .env.example .env
        echo 'Создан файл .env из примера. Отредактируйте его с вашими настройками.'
    fi
"

# Запуск приложения
log "🚀 Запуск приложения..."
ssh -i "$SSH_KEY" "$REMOTE_USER@$SERVER_IP" "
    cd /opt/$APP_NAME
    docker-compose down 2>/dev/null || true
    docker-compose up -d
"

# Проверка статуса
log "🔍 Проверка статуса приложения..."
ssh -i "$SSH_KEY" "$REMOTE_USER@$SERVER_IP" "
    cd /opt/$APP_NAME
    docker-compose ps
"

# Проверка доступности
log "🌐 Проверка доступности приложения..."
sleep 10
if curl -f "http://$SERVER_IP:8080/health" &> /dev/null; then
    log "✅ Приложение успешно развернуто и доступно на http://$SERVER_IP:8080"
else
    warn "⚠️ Приложение может быть недоступно. Проверьте логи:"
    echo "ssh -i $SSH_KEY $REMOTE_USER@$SERVER_IP 'cd /opt/$APP_NAME && docker-compose logs'"
fi

log "🎉 Развертывание завершено!"
echo ""
echo "📋 Следующие шаги:"
echo "1. Отредактируйте файл .env на сервере:"
echo "   ssh -i $SSH_KEY $REMOTE_USER@$SERVER_IP 'nano /opt/$APP_NAME/.env'"
echo ""
echo "2. Настройте домен и SSL сертификат (опционально)"
echo ""
echo "3. Проверьте логи приложения:"
echo "   ssh -i $SSH_KEY $REMOTE_USER@$SERVER_IP 'cd /opt/$APP_NAME && docker-compose logs -f'"
echo ""
echo "4. Перезапустите приложение после изменения .env:"
echo "   ssh -i $SSH_KEY $REMOTE_USER@$SERVER_IP 'cd /opt/$APP_NAME && docker-compose restart'"
